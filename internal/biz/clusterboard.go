package biz

import (
	"context"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// clusterBoardTopN 集群看板 Top Pod 的条数上限（按 top_sort 维度降序取前 N，默认 CPU）。
const clusterBoardTopN = 20

// ClusterBoardData 是集群看板的一次原始快照：由 data 层一次性拉取节点/命名空间/
// 全集群 Pod 与其指标，biz 层负责聚合成展示模型。保持 data 层纯数据获取，
// TopN 排序与角色/状态派生等展示逻辑收敛在 biz 纯函数中，便于测试。
type ClusterBoardData struct {
	Nodes       []corev1.Node
	NodeMetrics []v1beta1.NodeMetrics
	Namespaces  []corev1.Namespace
	Pods        []corev1.Pod
	PodMetrics  []v1beta1.PodMetrics
}

// ClusterBoardNode 是集群看板的单节点明细：角色/状态 + 容量/用量/请求（CPU 毫核、内存字节）。
type ClusterBoardNode struct {
	Name                              string
	Role                              string
	Status                            string
	CpuCapacity, CpuUsage, CpuRequest int64
	MemCapacity, MemUsage, MemRequest int64
}

// ClusterBoardNamespace 是集群看板的单命名空间聚合：CPU/内存用量 + Pod 数。
type ClusterBoardNamespace struct {
	Name        string
	CpuMilli    int64
	MemoryBytes int64
	PodCount    int32
}

// ClusterBoardPod 是集群看板的 Top Pod（按 top_sort 维度降序，默认 CPU）。
type ClusterBoardPod struct {
	Namespace   string
	Pod         string
	CpuMilli    int64
	MemoryBytes int64
}

// ClusterBoard 是集群看板聚合快照：总览 + 节点 + 命名空间 + Top Pod。
type ClusterBoard struct {
	Overview   *ClusterInfo
	Nodes      []*ClusterBoardNode
	Namespaces []*ClusterBoardNamespace
	Pods       []*ClusterBoardPod
}

// ClusterBoard 聚合集群看板快照：总览复用 ClusterInfo，节点/命名空间/Top Pod
// 从 data 快照派生（透传 repo）。命名空间排行与 Top Pod 只保留 mars 自己管理的
// 命名空间（managedNames 集合）及其 Pod，排除 kube-system/calico 等系统组件；
// 节点表不参与过滤（节点请求聚合需全量 Pod 才算得准）。
func (k *k8sBiz) ClusterBoard(ctx context.Context, managedNames []string, topSort string) (*ClusterBoard, error) {
	data, err := k.k8sRepo.ClusterBoard(ctx, false)
	if err != nil {
		return nil, err
	}
	return buildClusterBoard(data, k.ClusterInfo(), toNamespaceSet(managedNames), topSort), nil
}

// toNamespaceSet 把命名空间名列表转成过滤用集合；nil/空列表返回空集合（什么都不留）。
func toNamespaceSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return set
}

// filterNamespaces 只保留 mars 管理集合内的命名空间（排行仅展示 mars 管理空间）。
func filterNamespaces(namespaces []corev1.Namespace, managed map[string]bool) []corev1.Namespace {
	filtered := make([]corev1.Namespace, 0, len(namespaces))
	for _, ns := range namespaces {
		if managed[ns.Name] {
			filtered = append(filtered, ns)
		}
	}
	return filtered
}

// filterPodMetrics 只保留落在 mars 管理命名空间内的 Pod 指标（Top Pod 排除系统 Pod）。
func filterPodMetrics(metrics []v1beta1.PodMetrics, managed map[string]bool) []v1beta1.PodMetrics {
	filtered := make([]v1beta1.PodMetrics, 0, len(metrics))
	for _, m := range metrics {
		if managed[m.Namespace] {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// buildClusterBoard 把原始快照聚合成看板模型：Overview 直接引用传入总览，
// 节点/命名空间/Top Pod 分别派生；命名空间与 Top Pod 按 mars 管理集合过滤；
// topSort 控制 Top Pod 排行维度（"cpu"/"mem"，空=CPU）。快照为 nil 时只保留
// 总览，防御未知 nil。
func buildClusterBoard(data *ClusterBoardData, overview *ClusterInfo, managed map[string]bool, topSort string) *ClusterBoard {
	if data == nil {
		return &ClusterBoard{Overview: overview}
	}
	managedPods := filterPodMetrics(data.PodMetrics, managed)
	return &ClusterBoard{
		Overview:   overview,
		Nodes:      buildBoardNodes(data.Nodes, data.NodeMetrics, data.Pods),
		Namespaces: buildBoardNamespaces(filterNamespaces(data.Namespaces, managed), data.Pods, managedPods),
		Pods:       buildBoardTopPods(managedPods, clusterBoardTopN, topSort),
	}
}

// buildBoardNodes 逐节点派生看板明细（角色/状态/容量/用量/请求）。
func buildBoardNodes(nodes []corev1.Node, nodeMetrics []v1beta1.NodeMetrics, pods []corev1.Pod) []*ClusterBoardNode {
	result := make([]*ClusterBoardNode, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, buildBoardNode(node, nodeMetrics, pods))
	}
	return result
}

// buildBoardNode 派生单个节点明细：容量取 Status.Capacity，用量匹配同名 NodeMetrics，
// 请求累加落在该节点上的 Running Pod 的容器 Requests。
func buildBoardNode(node corev1.Node, nodeMetrics []v1beta1.NodeMetrics, pods []corev1.Pod) *ClusterBoardNode {
	cpuUsage, memoryUsage := nodeUsage(node.Name, nodeMetrics)
	cpuRequest, memoryRequest := nodeRequests(pods, node.Name)
	return &ClusterBoardNode{
		Name:        node.Name,
		Role:        nodeRole(node),
		Status:      nodeStatus(node),
		CpuCapacity: node.Status.Capacity.Cpu().MilliValue(),
		CpuUsage:    cpuUsage,
		CpuRequest:  cpuRequest,
		MemCapacity: node.Status.Capacity.Memory().Value(),
		MemUsage:    memoryUsage,
		MemRequest:  memoryRequest,
	}
}

// nodeRole 按节点标签判定角色：存在 master/control-plane 角色标签即 master，否则 worker。
func nodeRole(node corev1.Node) string {
	for _, key := range []string{
		"node-role.kubernetes.io/master",
		"node-role.kubernetes.io/control-plane",
	} {
		if _, ok := node.Labels[key]; ok {
			return "master"
		}
	}
	return "worker"
}

// nodeStatus 按节点状态派生看板状态：不可调度优先标 SchedulingDisabled，
// 否则按 Ready 条件判 Ready/NotReady。
func nodeStatus(node corev1.Node) string {
	if node.Spec.Unschedulable {
		return "SchedulingDisabled"
	}
	for _, cond := range node.Status.Conditions {
		if cond.Type != corev1.NodeReady {
			continue
		}
		if cond.Status == corev1.ConditionTrue {
			return "Ready"
		}
		return "NotReady"
	}
	return "NotReady"
}

// nodeUsage 从节点指标列表匹配同名单节点的 CPU/内存用量（毫核/字节），未匹配返回 0。
func nodeUsage(name string, nodeMetrics []v1beta1.NodeMetrics) (int64, int64) {
	for _, m := range nodeMetrics {
		if m.Name != name {
			continue
		}
		return m.Usage.Cpu().MilliValue(), m.Usage.Memory().Value()
	}
	return 0, 0
}

// nodeRequests 累加落在指定节点上的全部 Pod 的容器 CPU/内存 Requests（毫核/字节）。
func nodeRequests(pods []corev1.Pod, nodeName string) (int64, int64) {
	var cpu, memory int64
	for _, pod := range pods {
		if pod.Spec.NodeName != nodeName {
			continue
		}
		for _, container := range pod.Spec.Containers {
			cpu += container.Resources.Requests.Cpu().MilliValue()
			memory += container.Resources.Requests.Memory().Value()
		}
	}
	return cpu, memory
}

// podMetricsUsage 累加单个 Pod 指标下全部容器的 CPU/内存用量（毫核/字节）：
// PodMetrics 无顶层 Usage，用量分散在 Containers 列表里。
func podMetricsUsage(m v1beta1.PodMetrics) (int64, int64) {
	var cpu, memory int64
	for _, c := range m.Containers {
		cpu += c.Usage.Cpu().MilliValue()
		memory += c.Usage.Memory().Value()
	}
	return cpu, memory
}

// buildBoardNamespaces 逐命名空间聚合：Pod 数取落在该命名空间的 Running Pod 数量，
// CPU/内存用量累加该命名空间下全部 Pod 指标。
func buildBoardNamespaces(namespaces []corev1.Namespace, pods []corev1.Pod, podMetrics []v1beta1.PodMetrics) []*ClusterBoardNamespace {
	result := make([]*ClusterBoardNamespace, 0, len(namespaces))
	for _, ns := range namespaces {
		item := &ClusterBoardNamespace{Name: ns.Name}
		for _, pod := range pods {
			if pod.Namespace == ns.Name {
				item.PodCount++
			}
		}
		for _, m := range podMetrics {
			if m.Namespace != ns.Name {
				continue
			}
			cpu, memory := podMetricsUsage(m)
			item.CpuMilli += cpu
			item.MemoryBytes += memory
		}
		result = append(result, item)
	}
	return result
}

// buildBoardTopPods 把全部 Pod 指标转成看板项，按 topSort 维度降序取前 topN 条：
// "mem" 按内存用量排序，其余（含空/"cpu"）按 CPU 用量排序。
func buildBoardTopPods(podMetrics []v1beta1.PodMetrics, topN int, topSort string) []*ClusterBoardPod {
	pods := make([]*ClusterBoardPod, 0, len(podMetrics))
	for _, m := range podMetrics {
		cpu, memory := podMetricsUsage(m)
		pods = append(pods, &ClusterBoardPod{
			Namespace:   m.Namespace,
			Pod:         m.Name,
			CpuMilli:    cpu,
			MemoryBytes: memory,
		})
	}
	sort.Slice(pods, func(i, j int) bool {
		if topSort == "mem" {
			return pods[i].MemoryBytes > pods[j].MemoryBytes
		}
		return pods[i].CpuMilli > pods[j].CpuMilli
	})
	if len(pods) > topN {
		pods = pods[:topN]
	}
	return pods
}
