package biz

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmetatypes "k8s.io/apimachinery/pkg/types"
)

// deploymentRevisionAnnotation 是 ReplicaSet 上记录其 revision 的注解键。
// k8s.io/api/apps/v1 未导出该常量，故本地定义（值不可变）。
const deploymentRevisionAnnotation = "deployment.kubernetes.io/revision"

// tailLogLines 拉取失败容器日志的尾部行数。
const tailLogLines = 200

// ContainerFailure 描述部署判定 FAILED 时失败容器的诊断信息（含日志尾部）。
type ContainerFailure struct {
	Kind      string // 工作负载类型: Deployment/StatefulSet/DaemonSet
	Workload  string // 工作负载名
	Pod       string // 失败 pod 名
	Container string // 失败容器名（pod 级失败如 Evicted 时为空）
	Reason    string // 失败原因: CrashLoopBackOff/ImagePullBackOff/Error/OOMKilled/Evicted 等
	Message   string // 容器状态 message（容器从未启动无日志时兜底展示）
	Logs      string // 容器日志尾部（CrashLoopBackOff 取上一次崩溃日志）
}

// ApplyStatus 是 CheckApplyStatus 的领域返回：聚合判定 + 原因 + 容器明细 + 失败诊断。
type ApplyStatus struct {
	Status     types.Deploy
	Reason     string
	Containers []*types.StateContainer
	Failures   []*ContainerFailure
}

// failedContainerRef 是纯扫描产物：失败容器定位与原因（日志另由上层拉取，保持纯函数可测）。
type failedContainerRef struct {
	Kind, Workload, Pod, Container, Reason, Message string
}

// fatalWaitingReasons 是容器进入等待态后需要判定为失败的稳定 reason。
// ContainerCreating/ErrImagePull 等瞬时态不计入，避免把拉镜像的瞬时状态误报成失败。
var fatalWaitingReasons = map[string]struct{}{
	"CrashLoopBackOff":           {},
	"ImagePullBackOff":           {},
	"CreateContainerConfigError": {},
	"InvalidImageName":           {},
	"RunContainerError":          {},
}

// fatalTerminatedReasons 是容器终止态需要判定为失败的 reason。
var fatalTerminatedReasons = map[string]struct{}{
	"Error":              {},
	"OOMKilled":          {},
	"ContainerCannotRun": {},
}

// CheckApplyStatus 判定项目最近一次部署后新版本容器是否正常运行。
// 解析项目 manifests 里的 Deployment/StatefulSet/DaemonSet，逐个读取实时状态与
// 最新版本 pod 容器状态，聚合出 Deployed/Deploying/Failed/Unknown 判定：
// 任一工作负载 Failed → 整体 Failed，否则任一 Deploying → 整体 Deploying，全正常才 Deployed。
// 旧版本 pod 异常（Evicted/终止中）属滚动清理过程，不影响新版本判定。
func (p *projectBiz) CheckApplyStatus(ctx context.Context, id int) (*ApplyStatus, error) {
	proj, err := p.projRepo.Show(ctx, id)
	if err != nil {
		return nil, err
	}
	containers, err := buildStateContainers(ctx, p.k8sRepo, proj)
	if err != nil {
		return nil, err
	}
	deployments, statefulSets, daemonSets := p.k8sRepo.GetWorkloadsByManifest(proj.Manifest)
	if len(deployments)+len(statefulSets)+len(daemonSets) == 0 {
		return &ApplyStatus{Status: types.Deploy_StatusUnknown, Reason: "未发现 Deployment/StatefulSet/DaemonSet 工作负载", Containers: containers}, nil
	}
	status, reason, failures, err := p.judgeWorkloads(ctx, proj.Namespace.Name, deployments, statefulSets, daemonSets)
	if err != nil {
		return nil, err
	}
	return &ApplyStatus{Status: status, Reason: reason, Containers: containers, Failures: failures}, nil
}

// judgeWorkloads 逐个判定三类工作负载的滚动状态并聚合：
// Failed 优先，其次 Deploying；全部 Deployed 才返回 Deployed。
func (p *projectBiz) judgeWorkloads(ctx context.Context, ns string, deployments []*appsv1.Deployment, statefulSets []*appsv1.StatefulSet, daemonSets []*appsv1.DaemonSet) (types.Deploy, string, []*ContainerFailure, error) {
	status := types.Deploy_StatusDeployed
	var reasons []string
	var failures []*ContainerFailure

	// mark 聚合单个工作负载的判定：失败优先、进行中次之、Deployed 不改变整体。
	mark := func(st types.Deploy, reason string, fails []*ContainerFailure) {
		switch st {
		case types.Deploy_StatusFailed:
			status = types.Deploy_StatusFailed
			reasons = append(reasons, reason)
			failures = append(failures, fails...)
		case types.Deploy_StatusDeploying:
			if status != types.Deploy_StatusFailed {
				status = types.Deploy_StatusDeploying
				reasons = append(reasons, reason)
			}
		}
	}

	// Deployment 需最新 ReplicaSet 定位新版本 pod，RS 列表统一取一次。
	var (
		rss []*appsv1.ReplicaSet
		err error
	)
	if len(deployments) > 0 {
		rss, err = p.k8sRepo.ListReplicaSets(ns)
		if err != nil {
			return 0, "", nil, err
		}
	}

	for _, wl := range deployments {
		live, err := p.k8sRepo.GetDeployment(ns, wl.Name)
		if err != nil {
			if errs.IsNotFound(err) {
				mark(types.Deploy_StatusDeploying, "Deployment "+wl.Name+" 尚未创建", nil)
				continue
			}
			return 0, "", nil, err
		}
		st, reason, fails := judgeDeploymentRollout(live, rss, p.podsByWorkload(ns, wl.Spec.Selector))
		mark(st, "Deployment "+wl.Name+": "+reason, toDomainFailures(fails, "Deployment", wl.Name))
	}

	for _, wl := range statefulSets {
		live, err := p.k8sRepo.GetStatefulSet(ns, wl.Name)
		if err != nil {
			if errs.IsNotFound(err) {
				mark(types.Deploy_StatusDeploying, "StatefulSet "+wl.Name+" 尚未创建", nil)
				continue
			}
			return 0, "", nil, err
		}
		st, reason, fails := judgeStatefulSetRollout(live, p.podsByWorkload(ns, wl.Spec.Selector))
		mark(st, "StatefulSet "+wl.Name+": "+reason, toDomainFailures(fails, "StatefulSet", wl.Name))
	}

	for _, wl := range daemonSets {
		live, err := p.k8sRepo.GetDaemonSet(ns, wl.Name)
		if err != nil {
			if errs.IsNotFound(err) {
				mark(types.Deploy_StatusDeploying, "DaemonSet "+wl.Name+" 尚未创建", nil)
				continue
			}
			return 0, "", nil, err
		}
		st, reason, fails := judgeDaemonSetRollout(live, p.podsByWorkload(ns, wl.Spec.Selector))
		mark(st, "DaemonSet "+wl.Name+": "+reason, toDomainFailures(fails, "DaemonSet", wl.Name))
	}

	if status == types.Deploy_StatusFailed {
		p.fillFailureLogs(ctx, ns, failures)
	}
	return status, strings.Join(reasons, "; "), failures, nil
}

// judgeDeploymentRollout 依据 Deployment 实时状态与最新版本 pod 判定滚动结果。
// 旧版本 pod 异常不影响判定；新版本 pod 失败优先于计数判定，避免计数一致但容器故障时误报成功。
func judgeDeploymentRollout(dep *appsv1.Deployment, rss []*appsv1.ReplicaSet, pods []*corev1.Pod) (types.Deploy, string, []failedContainerRef) {
	desired := int32(0)
	if dep.Spec.Replicas != nil {
		desired = *dep.Spec.Replicas
	}
	if desired == 0 {
		return types.Deploy_StatusDeployed, "副本数为 0，无容器需验证", nil
	}
	// 控制器尚未收敛到最新 spec，直接判进行中
	if dep.Status.ObservedGeneration < dep.Generation {
		return types.Deploy_StatusDeploying, "控制器尚未完成对账", nil
	}
	// 过渡窗口：最新 spec 已提交但新版本 pod 一个都没创建，绝不算成功
	if dep.Status.UpdatedReplicas == 0 {
		return types.Deploy_StatusDeploying, "新版本 pod 尚未创建", nil
	}
	if fails := collectPodFailures(deploymentNewPods(dep, rss, pods)); len(fails) > 0 {
		return types.Deploy_StatusFailed, formatReason(fails), fails
	}
	if dep.Status.UpdatedReplicas == desired && dep.Status.AvailableReplicas == desired {
		return types.Deploy_StatusDeployed, "", nil
	}
	return types.Deploy_StatusDeploying, fmt.Sprintf("滚动进行中 %d/%d 就绪", dep.Status.AvailableReplicas, desired), nil
}

// judgeStatefulSetRollout 依据 StatefulSet 实时状态与最新版本 pod 判定滚动结果。
func judgeStatefulSetRollout(sts *appsv1.StatefulSet, pods []*corev1.Pod) (types.Deploy, string, []failedContainerRef) {
	desired := int32(0)
	if sts.Spec.Replicas != nil {
		desired = *sts.Spec.Replicas
	}
	if desired == 0 {
		return types.Deploy_StatusDeployed, "副本数为 0，无容器需验证", nil
	}
	if sts.Status.ObservedGeneration < sts.Generation {
		return types.Deploy_StatusDeploying, "控制器尚未完成对账", nil
	}
	if sts.Status.UpdatedReplicas == 0 {
		return types.Deploy_StatusDeploying, "新版本 pod 尚未创建", nil
	}
	if fails := collectPodFailures(statefulSetNewPods(sts, pods)); len(fails) > 0 {
		return types.Deploy_StatusFailed, formatReason(fails), fails
	}
	if sts.Status.UpdatedReplicas == desired && sts.Status.AvailableReplicas == desired {
		return types.Deploy_StatusDeployed, "", nil
	}
	return types.Deploy_StatusDeploying, fmt.Sprintf("滚动进行中 %d/%d 就绪", sts.Status.AvailableReplicas, desired), nil
}

// judgeDaemonSetRollout 依据 DaemonSet 实时状态与最新版本 pod 判定滚动结果。
func judgeDaemonSetRollout(ds *appsv1.DaemonSet, pods []*corev1.Pod) (types.Deploy, string, []failedContainerRef) {
	desired := ds.Status.DesiredNumberScheduled
	if desired == 0 {
		return types.Deploy_StatusDeployed, "无节点需调度，无容器需验证", nil
	}
	if ds.Status.ObservedGeneration < ds.Generation {
		return types.Deploy_StatusDeploying, "控制器尚未完成对账", nil
	}
	if ds.Status.UpdatedNumberScheduled == 0 {
		return types.Deploy_StatusDeploying, "新版本 pod 尚未创建", nil
	}
	if fails := collectPodFailures(daemonSetNewPods(pods)); len(fails) > 0 {
		return types.Deploy_StatusFailed, formatReason(fails), fails
	}
	if ds.Status.UpdatedNumberScheduled == desired && ds.Status.NumberAvailable == desired {
		return types.Deploy_StatusDeployed, "", nil
	}
	return types.Deploy_StatusDeploying, fmt.Sprintf("滚动进行中 %d/%d 就绪", ds.Status.NumberAvailable, desired), nil
}

// deploymentNewPods 返回 Deployment 下最新 revision ReplicaSet 所管理的 pod（新版本 pod）。
func deploymentNewPods(dep *appsv1.Deployment, rss []*appsv1.ReplicaSet, pods []*corev1.Pod) []*corev1.Pod {
	latest := latestReplicaSet(dep, rss)
	if latest == nil {
		return nil
	}
	return ownedByRS(pods, latest.UID)
}

// latestReplicaSet 返回 Deployment 下 revision 注解最大的 ReplicaSet；无则返回 nil。
func latestReplicaSet(dep *appsv1.Deployment, rss []*appsv1.ReplicaSet) *appsv1.ReplicaSet {
	var latest *appsv1.ReplicaSet
	for _, rs := range rss {
		if !ownedBy(rs, dep.UID) {
			continue
		}
		if latest == nil || revisionOf(rs) > revisionOf(latest) {
			latest = rs
		}
	}
	return latest
}

// revisionOf 解析 ReplicaSet 的 deployment.kubernetes.io/revision 注解为整数，解析失败返回 0。
func revisionOf(rs *appsv1.ReplicaSet) int {
	rev, err := strconv.Atoi(rs.Annotations[deploymentRevisionAnnotation])
	if err != nil {
		return 0
	}
	return rev
}

// statefulSetNewPods 返回 StatefulSet 当前更新版本
// （controller-revision-hash == status.updateRevision）的 pod。
func statefulSetNewPods(sts *appsv1.StatefulSet, pods []*corev1.Pod) []*corev1.Pod {
	rev := sts.Status.UpdateRevision
	if rev == "" {
		return nil
	}
	var res []*corev1.Pod
	for _, pod := range pods {
		if pod.Labels[appsv1.ControllerRevisionHashLabelKey] == rev {
			res = append(res, pod)
		}
	}
	return res
}

// daemonSetNewPods 以 controller-revision-hash 分组、按创建时间最新的一组作为新版本 pod。
// DaemonSet 状态无 revision 字段，滚动中最新创建的 pod 必属新模板，属合理近似。
func daemonSetNewPods(pods []*corev1.Pod) []*corev1.Pod {
	byHash := make(map[string][]*corev1.Pod)
	var newestHash string
	var newest time.Time
	for _, pod := range pods {
		hash := pod.Labels[appsv1.ControllerRevisionHashLabelKey]
		if hash == "" {
			continue
		}
		byHash[hash] = append(byHash[hash], pod)
		created := pod.CreationTimestamp.Time
		if newest.IsZero() || created.After(newest) {
			newest = created
			newestHash = hash
		}
	}
	if newestHash == "" {
		return pods
	}
	return byHash[newestHash]
}

// ownedBy 判断 obj 是否被指定 UID 的属主直接拥有。
func ownedBy(obj metav1.Object, uid kmetatypes.UID) bool {
	for _, ref := range obj.GetOwnerReferences() {
		if ref.UID == uid {
			return true
		}
	}
	return false
}

// ownedByRS 过滤出 OwnerReference 包含指定 ReplicaSet UID 的 pod。
func ownedByRS(pods []*corev1.Pod, uid kmetatypes.UID) []*corev1.Pod {
	var res []*corev1.Pod
	for _, pod := range pods {
		if ownedBy(pod, uid) {
			res = append(res, pod)
		}
	}
	return res
}

// collectPodFailures 纯函数：从 pod 状态推断失败容器（不触发任何 k8s 调用）。
// pod 级失败（Failed/Evicted）记一条无容器名的失败；容器级失败按稳定 reason 判定。
func collectPodFailures(pods []*corev1.Pod) []failedContainerRef {
	var refs []failedContainerRef
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodFailed {
			reason := pod.Status.Reason
			if reason == "" {
				reason = "PodFailed"
			}
			refs = append(refs, failedContainerRef{Pod: pod.Name, Reason: reason, Message: pod.Status.Message})
			continue
		}
		for _, cs := range pod.Status.ContainerStatuses {
			if w := cs.State.Waiting; w != nil && w.Reason != "" {
				if _, ok := fatalWaitingReasons[w.Reason]; ok {
					refs = append(refs, failedContainerRef{Pod: pod.Name, Container: cs.Name, Reason: w.Reason, Message: w.Message})
					continue
				}
			}
			if t := cs.State.Terminated; t != nil && t.Reason != "" {
				if _, ok := fatalTerminatedReasons[t.Reason]; ok {
					refs = append(refs, failedContainerRef{Pod: pod.Name, Container: cs.Name, Reason: t.Reason, Message: t.Message})
				}
			}
		}
	}
	return refs
}

// formatReason 把失败容器汇总为人类可读原因（逐条列出，分号分隔）。
func formatReason(fails []failedContainerRef) string {
	var parts []string
	for _, f := range fails {
		if f.Container != "" {
			parts = append(parts, fmt.Sprintf("容器 %s 失败: %s", f.Container, f.Reason))
		} else {
			parts = append(parts, fmt.Sprintf("pod %s: %s", f.Pod, f.Reason))
		}
	}
	return strings.Join(parts, "; ")
}

// toDomainFailures 把纯扫描得到的失败容器转成领域类型并补全工作负载信息。
func toDomainFailures(fails []failedContainerRef, kind, workload string) []*ContainerFailure {
	out := make([]*ContainerFailure, 0, len(fails))
	for _, f := range fails {
		out = append(out, &ContainerFailure{
			Kind: kind, Workload: workload, Pod: f.Pod, Container: f.Container,
			Reason: f.Reason, Message: f.Message,
		})
	}
	return out
}

// fillFailureLogs 为失败容器补拉日志尾部：
// CrashLoopBackOff 取上一次崩溃实例的日志；容器从未启动（如 ImagePullBackOff）时
// 日志拉取失败仅降级为 message，不改变判定。
func (p *projectBiz) fillFailureLogs(ctx context.Context, ns string, failures []*ContainerFailure) {
	for _, f := range failures {
		if f.Container == "" {
			continue
		}
		opts := &corev1.PodLogOptions{
			TailLines: lo.ToPtr(int64(tailLogLines)),
			Container: f.Container,
			Previous:  f.Reason == "CrashLoopBackOff",
		}
		logs, err := p.k8sRepo.GetPodLogs(ctx, ns, f.Pod, opts)
		if err != nil {
			continue
		}
		f.Logs = strings.TrimSpace(logs)
	}
}

// podsByWorkload 按工作负载 selector 列出命名空间内 pod；selector 解析失败返回 nil
// （manifest 中合法对象的 selector 必然可解析，此为防御分支）。
func (p *projectBiz) podsByWorkload(ns string, sel *metav1.LabelSelector) []*corev1.Pod {
	parsed, err := metav1.LabelSelectorAsSelector(sel)
	if err != nil {
		return nil
	}
	pods, _ := p.k8sRepo.ListPodsBySelectors(ns, []string{parsed.String()})
	return pods
}
