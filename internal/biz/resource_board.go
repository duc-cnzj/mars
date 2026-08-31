package biz

import (
	"context"
	"fmt"
	"sort"

	"github.com/dustin/go-humanize"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	kmetatypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// ResourceSnapshotData 空间资源聚合的一次原始快照：全集群 Running Pod、ReplicaSet
// 与其指标。requests 取自 pod spec 容器 Resources.Requests，实际用量取自 PodMetrics
// 容器 Usage；ReplicaSet 供项目 pod → Deployment 属主链解析（pod 直接属主是 RS，
// RS 属主是 Deployment；StatefulSet/DaemonSet pod 直接属主即 workload）。
type ResourceSnapshotData struct {
	Pods        []corev1.Pod
	ReplicaSets []appsv1.ReplicaSet
	PodMetrics  []v1beta1.PodMetrics
}

// ResourceProject 单个命名空间内一个项目的资源申请/实际用量聚合（Pod selectors 匹配）。
type ResourceProject struct {
	Name            string
	PodCount        int32
	CpuRequestMilli int64
	CpuUsageMilli   int64
	MemRequestBytes int64
	MemUsageBytes   int64
	// Workloads 项目内各工作负载（Deployment/StatefulSet/DaemonSet）的细分聚合，
	// 按 pod 属主链匹配；无 workload 属主的裸 pod 计入项目总量但不单列。
	Workloads []*ResourceProjectWorkload
}

// ResourceProjectWorkload 项目内单个工作负载（Deployment/StatefulSet/DaemonSet）的
// 资源申请/实际用量聚合（pod 属主链匹配）。
type ResourceProjectWorkload struct {
	Kind            string
	Name            string
	PodCount        int32
	CpuRequestMilli int64
	CpuUsageMilli   int64
	MemRequestBytes int64
	MemUsageBytes   int64
}

// ResourceNamespace 单个命名空间的资源申请/实际用量聚合（全部 Running Pod）+ 项目拆分。
// namespace 总量 = 空间内所有 Pod，不受项目归属影响（一个 pod 可命中多个项目或零命中）。
type ResourceNamespace struct {
	Name            string
	PodCount        int32
	CpuRequestMilli int64
	CpuUsageMilli   int64
	MemRequestBytes int64
	MemUsageBytes   int64
	Projects        []*ResourceProject
}

// ResourceBoard 空间资源管理聚合：全部管理命名空间（含项目明细）。
type ResourceBoard struct {
	Namespaces []*ResourceNamespace
}

// FormatResourceUsage 把资源用量（CPU 毫核 / 内存字节）格式化为 "N m" / 人类可读字节：
// CPU 对齐 GetCpuAndMemory 的 "%d m" 契约，内存走 humanize.Bytes（十进制基数，
// 0 字节输出 "0 B"）。作为全仓资源用量展示的唯一格式来源，命名空间管理列表、
// 空间资源板与集群用量共用，保证各页面展示口径一致。
func FormatResourceUsage(cpuMilli, memBytes int64) (string, string) {
	return fmt.Sprintf("%d m", cpuMilli), humanize.Bytes(uint64(memBytes))
}

// ResourceBoard 聚合管理命名空间的资源申请/实际用量占比数据：
// requests 取 Running Pod spec、实际用量取 PodMetrics；项目拆分按项目 PodSelectors
// 匹配 pod。命名空间集合由快照中落在 managedNames 内的命名空间派生
// （无 Pod 也无指标的空间不产出记录——无资源可管理即不展示）。
func (k *k8sBiz) ResourceBoard(ctx context.Context, managedNames []string, projects []*Project) (*ResourceBoard, error) {
	data, err := k.k8sRepo.ResourceSnapshot(ctx, false)
	if err != nil {
		return nil, err
	}
	return buildResourceBoard(data, toNamespaceSet(managedNames), projects), nil
}

// buildResourceBoard 把原始快照聚合为空间资源板：命名空间总量按全部 Running Pod
// 累加 requests/用量，项目拆分按 selectors 匹配同名空间内的 pod。快照为 nil 时
// 返回空板（防御未知 nil，对齐 buildClusterBoard 的防御语义）。
func buildResourceBoard(data *ResourceSnapshotData, managed map[string]bool, projects []*Project) *ResourceBoard {
	if data == nil {
		return &ResourceBoard{}
	}
	namespaces := buildResourceNamespaces(managed, data.Pods, data.PodMetrics)
	attachResourceProjects(namespaces, projects, data.Pods, data.ReplicaSets, data.PodMetrics)
	return &ResourceBoard{Namespaces: namespaces}
}

// buildResourceNamespaces 从快照派生命名空间板：每个空间累加全部 Running Pod 的
// requests（pod spec）与实际用量（PodMetrics），Pod 数取空间内 Running Pod 数量。
// 只保留 managed 集合内的空间；命名空间按名排序保证确定性输出。
func buildResourceNamespaces(managed map[string]bool, pods []corev1.Pod, podMetrics []v1beta1.PodMetrics) []*ResourceNamespace {
	byName := make(map[string]*ResourceNamespace)
	for _, pod := range pods {
		if !managed[pod.Namespace] {
			continue
		}
		ns := byName[pod.Namespace]
		if ns == nil {
			ns = &ResourceNamespace{Name: pod.Namespace}
			byName[pod.Namespace] = ns
		}
		ns.PodCount++
		for _, c := range pod.Spec.Containers {
			ns.CpuRequestMilli += c.Resources.Requests.Cpu().MilliValue()
			ns.MemRequestBytes += c.Resources.Requests.Memory().Value()
		}
	}
	for _, m := range podMetrics {
		if !managed[m.Namespace] {
			continue
		}
		ns := byName[m.Namespace]
		if ns == nil {
			ns = &ResourceNamespace{Name: m.Namespace}
			byName[m.Namespace] = ns
		}
		cpu, memory := podMetricsUsage(m)
		ns.CpuUsageMilli += cpu
		ns.MemUsageBytes += memory
	}
	result := make([]*ResourceNamespace, 0, len(byName))
	for _, ns := range byName {
		result = append(result, ns)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

// podUsage 单个 Pod 的实际用量（CPU 毫核 / 内存字节），来自 PodMetrics 汇总。
type podUsage struct {
	cpuMilli int64
	memBytes int64
}

// attachResourceProjects 按项目 PodSelectors 把项目拆分配进对应命名空间板：
// requests 从 pod spec 累加，实际用量按 (namespace, name) 从指标索引取；
// 命中 selectors 任意一个即归该项目，并按其属主链拆进工作负载细分。
// 项目按名排序、工作负载按 kind 再 name 排序保证确定性输出。
func attachResourceProjects(namespaces []*ResourceNamespace, projects []*Project, pods []corev1.Pod, replicaSets []appsv1.ReplicaSet, podMetrics []v1beta1.PodMetrics) {
	nsByName := make(map[string]*ResourceNamespace, len(namespaces))
	for _, ns := range namespaces {
		nsByName[ns.Name] = ns
	}
	usageByKey := make(map[string]podUsage, len(podMetrics))
	for _, m := range podMetrics {
		cpu, memory := podMetricsUsage(m)
		usageByKey[m.Namespace+"/"+m.Name] = podUsage{cpuMilli: cpu, memBytes: memory}
	}
	rsByUID := rsByUIDIndex(replicaSets)
	for _, proj := range projects {
		if proj == nil || proj.Namespace == nil {
			continue
		}
		ns := nsByName[proj.Namespace.Name]
		if ns == nil {
			continue // 非管理空间内的项目不入板
		}
		selectors := parsedPodSelectors(proj.PodSelectors)
		item := &ResourceProject{Name: proj.Name}
		workloads := make(map[string]*ResourceProjectWorkload)
		for _, pod := range pods {
			if pod.Namespace != proj.Namespace.Name || !matchAnySelector(selectors, pod.Labels) {
				continue
			}
			item.PodCount++
			for _, c := range pod.Spec.Containers {
				item.CpuRequestMilli += c.Resources.Requests.Cpu().MilliValue()
				item.MemRequestBytes += c.Resources.Requests.Memory().Value()
			}
			var usage podUsage
			if u, ok := usageByKey[pod.Namespace+"/"+pod.Name]; ok {
				usage = u
				item.CpuUsageMilli += u.cpuMilli
				item.MemUsageBytes += u.memBytes
			}
			// 无 workload 属主的裸 pod（Job 等）只计入项目总量，不单列工作负载
			if kind, name := workloadOf(&pod, rsByUID); kind != "" {
				wk := workloads[kind+"/"+name]
				if wk == nil {
					wk = &ResourceProjectWorkload{Kind: kind, Name: name}
					workloads[kind+"/"+name] = wk
				}
				wk.PodCount++
				for _, c := range pod.Spec.Containers {
					wk.CpuRequestMilli += c.Resources.Requests.Cpu().MilliValue()
					wk.MemRequestBytes += c.Resources.Requests.Memory().Value()
				}
				wk.CpuUsageMilli += usage.cpuMilli
				wk.MemUsageBytes += usage.memBytes
			}
		}
		item.Workloads = sortedWorkloads(workloads)
		// 命中 0 个 Running Pod 的项目不产出记录：资源明细表只展示实际有资源可管理的
		// 项目（对齐命名空间级「无资源即不展示」语义），也避免前端 0/0 占比除零。
		if item.PodCount > 0 {
			ns.Projects = append(ns.Projects, item)
		}
	}
	for _, ns := range namespaces {
		sort.Slice(ns.Projects, func(i, j int) bool { return ns.Projects[i].Name < ns.Projects[j].Name })
	}
}

// rsByUIDIndex 把 ReplicaSet 列表按 UID 建索引，供项目 pod → Deployment 属主链解析。
func rsByUIDIndex(rss []appsv1.ReplicaSet) map[kmetatypes.UID]*appsv1.ReplicaSet {
	index := make(map[kmetatypes.UID]*appsv1.ReplicaSet, len(rss))
	for i := range rss {
		index[rss[i].UID] = &rss[i]
	}
	return index
}

// workloadOf 解析 pod 属主的工作负载分组键（kind,name）：Deployment 经 RS 属主链
// （pod 直接属主 ReplicaSet → RS 属主 Deployment），StatefulSet/DaemonSet pod 直接
// 属主即 workload；无 workload 属主的裸 pod（Job 等）或属主 RS 不在快照内返回空键
// （计入项目总量但不单列）。
func workloadOf(pod *corev1.Pod, rsByUID map[kmetatypes.UID]*appsv1.ReplicaSet) (kind, name string) {
	for _, ref := range pod.OwnerReferences {
		switch ref.Kind {
		case "StatefulSet":
			return "StatefulSet", ref.Name
		case "DaemonSet":
			return "DaemonSet", ref.Name
		case "ReplicaSet":
			if rs := rsByUID[ref.UID]; rs != nil {
				for _, rref := range rs.OwnerReferences {
					if rref.Kind == "Deployment" {
						return "Deployment", rref.Name
					}
				}
			}
		}
	}
	return "", ""
}

// sortedWorkloads 把按 (kind,name) 键聚合的工作负载去 map 化并按 kind 再 name 排序，
// 保证输出确定性（对齐项目/命名空间级排序约定）。
func sortedWorkloads(workloads map[string]*ResourceProjectWorkload) []*ResourceProjectWorkload {
	result := make([]*ResourceProjectWorkload, 0, len(workloads))
	for _, wk := range workloads {
		result = append(result, wk)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Kind != result[j].Kind {
			return result[i].Kind < result[j].Kind
		}
		return result[i].Name < result[j].Name
	})
	return result
}

// parsedPodSelectors 把项目 PodSelectors 字符串解析为 labels.Selector，非法 selector 跳过。
func parsedPodSelectors(raw []string) []labels.Selector {
	selectors := make([]labels.Selector, 0, len(raw))
	for _, s := range raw {
		if sel, err := labels.Parse(s); err == nil {
			selectors = append(selectors, sel)
		}
	}
	return selectors
}

// matchAnySelector 判断 pod labels 是否命中 selector 列表中的任意一个（空列表恒不命中）。
func matchAnySelector(selectors []labels.Selector, podLabels map[string]string) bool {
	for _, sel := range selectors {
		if sel.Matches(labels.Set(podLabels)) {
			return true
		}
	}
	return false
}
