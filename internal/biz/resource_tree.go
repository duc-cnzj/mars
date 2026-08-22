package biz

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kmetatypes "k8s.io/apimachinery/pkg/types"
)

// ResourceTree 是项目资源拓扑树的领域返回:整体部署状态 + 节点 + 边。
type ResourceTree struct {
	Status types.Deploy
	Nodes  []*ResourceTreeNode
	Edges  []*ResourceTreeEdge
}

// ResourceTreeNode 是拓扑树节点。kind/status 与前端拓扑渲染器对齐:
// kind ∈ Application|Deployment|ReplicaSet|Pod|Service|StatefulSet|DaemonSet,
// status ∈ healthy|degraded|progressing|unknown。
// Old 标记滚动升级中的旧版本 ReplicaSet/Pod,前端可据此弱化或打标签。
type ResourceTreeNode struct {
	ID        string
	Kind      string
	Name      string
	Namespace string
	Status    string
	Labels    map[string]string
	Old       bool
}

// ResourceTreeEdge 是拓扑树边:Type=owner 属主引用 / selector 标签选择器。
type ResourceTreeEdge struct {
	ID     string
	Type   string
	Source string
	Target string
}

// buildResourceTree 构建项目资源拓扑树:Application → Deployment → ReplicaSet → Pod /
// StatefulSet / DaemonSet,以及 Service 的 selector 边。完整资源列表(区别于
// buildStateContainers 的活跃容器平铺),与 AllContainers 一致不裁剪 workload 类型。
//
// Deployment/StatefulSet/DaemonSet 以 manifest 声明的名字为准逐个读实时对象
// (拿 UID 与滚动状态,未创建占位 progressing);
// ReplicaSet 从命名空间列表里筛属主是本项目 Deployment 的,按 revision 注解识别旧版本;
// pod 按其属主归到对应 workload 下:RS 属主归 Deployment 子树,sts/ds 属主归其子树
// (old 复用 collectWorkloadOldPods 的 hash 判定),其余裸 pod 直挂 Application;
// Service 用 selector 与项目 pod labels 的交集判定归属。
func buildResourceTree(ctx context.Context, k8sRepo K8sRepo, proj *Project) (*ResourceTree, error) {
	ns := proj.Namespace.Name
	appID := "application-" + strconv.Itoa(proj.ID)
	tree := &ResourceTree{
		Status: proj.DeployStatus,
		Nodes: []*ResourceTreeNode{{
			ID: appID, Kind: "Application", Name: proj.Name,
			Namespace: ns, Status: deployStatusString(proj.DeployStatus), Labels: map[string]string{},
		}},
	}
	// 无 PodSelectors = 从未部署,只回 Application 根节点
	if len(proj.PodSelectors) == 0 {
		return tree, nil
	}

	pods, err := k8sRepo.ListPodsBySelectors(ns, proj.PodSelectors)
	if err != nil {
		return nil, err
	}
	deployments, statefulSets, daemonSets := k8sRepo.GetWorkloadsByManifest(proj.Manifest)
	rss, err := k8sRepo.ListReplicaSets(ns)
	if err != nil {
		return nil, err
	}

	// 1. Deployment 节点:按 manifest 名字读实时对象判滚动状态,并收 UID 供 RS 筛选。
	//    尚未创建(部署刚发起)时占位为 progressing 节点。
	depByUID := make(map[kmetatypes.UID]*appsv1.Deployment)
	depIDByUID := make(map[kmetatypes.UID]string)
	for _, wl := range deployments {
		live, err := k8sRepo.GetDeployment(ns, wl.Name)
		if err != nil {
			if errs.IsNotFound(err) {
				tree.addNode(ns, "Deployment", wl.Name, "progressing", nil, false)
				tree.addEdge("owner", appID, kindID("Deployment", wl.Name))
				continue
			}
			return nil, err
		}
		depByUID[live.UID] = live
		depIDByUID[live.UID] = kindID("Deployment", live.Name)
		st, _, _ := judgeDeploymentRollout(live, rss, pods)
		tree.addNode(ns, "Deployment", live.Name, deployStatusString(st), nil, false)
		tree.addEdge("owner", appID, kindID("Deployment", live.Name))
	}

	// 2. StatefulSet / DaemonSet 节点(与 Deployment 同级挂 Application):与 AllContainers
	//    一致不做类型裁剪,sts/ds pod 完整入图。未创建占位 progressing;属主 pod 聚合到其下
	//    (old 标记复用 collectWorkloadOldPods 的 hash 判定),聚合过的 pod 不再直挂 Application。
	oldPodNames := collectWorkloadOldPods(k8sRepo, pods)
	ownedWorkloadPods := make(map[string]struct{})
	for _, wl := range statefulSets {
		live, err := k8sRepo.GetStatefulSet(ns, wl.Name)
		if err != nil {
			if errs.IsNotFound(err) {
				tree.addNode(ns, "StatefulSet", wl.Name, "progressing", nil, false)
				tree.addEdge("owner", appID, kindID("StatefulSet", wl.Name))
				continue
			}
			return nil, err
		}
		owned := podsOwnedBy(pods, live.UID)
		st, _, _ := judgeStatefulSetRollout(live, owned)
		for name := range tree.addWorkloadSubtree(ns, appID, "StatefulSet", live.Name, owned, oldPodNames, st) {
			ownedWorkloadPods[name] = struct{}{}
		}
	}
	for _, wl := range daemonSets {
		live, err := k8sRepo.GetDaemonSet(ns, wl.Name)
		if err != nil {
			if errs.IsNotFound(err) {
				tree.addNode(ns, "DaemonSet", wl.Name, "progressing", nil, false)
				tree.addEdge("owner", appID, kindID("DaemonSet", wl.Name))
				continue
			}
			return nil, err
		}
		owned := podsOwnedBy(pods, live.UID)
		st, _, _ := judgeDaemonSetRollout(live, owned)
		for name := range tree.addWorkloadSubtree(ns, appID, "DaemonSet", live.Name, owned, oldPodNames, st) {
			ownedWorkloadPods[name] = struct{}{}
		}
	}

	// 3. ReplicaSet:筛属主为本项目 Deployment 的 RS,按 revision 标旧,聚合其 pod。
	rsByUID := make(map[kmetatypes.UID]*appsv1.ReplicaSet)
	for _, rs := range rss {
		var ownerUID kmetatypes.UID
		for _, re := range rs.OwnerReferences {
			if re.Kind == "Deployment" {
				ownerUID = re.UID
				break
			}
		}
		if _, ok := depByUID[ownerUID]; !ok {
			continue // 非本项目 Deployment 的 RS
		}
		rsByUID[rs.UID] = rs
	}
	// 每个 Deployment 下最新 revision:低于它的 RS 标为旧版本
	latestRev := make(map[kmetatypes.UID]int)
	for uid := range depByUID {
		for _, rs := range rss {
			if ownedBy(rs, uid) {
				if r := revisionOf(rs); r > latestRev[uid] {
					latestRev[uid] = r
				}
			}
		}
	}

	// pod 归类:属项目 RS → 聚合到 RS;属 sts/ds → 已在其子树;否则直挂 Application
	rsAgg := make(map[kmetatypes.UID][]*corev1.Pod)
	for _, pod := range pods {
		var rsUID kmetatypes.UID
		var controller string
		for _, ref := range pod.OwnerReferences {
			if ref.Kind == "ReplicaSet" {
				rsUID = ref.UID
				break
			}
			// 非 RS 属主(sts/ds/裸 pod)记录其控制者,直挂 Application 时打标签
			controller = strings.ToLower(ref.Kind) + "/" + ref.Name
		}
		if _, ok := rsByUID[rsUID]; ok {
			rsAgg[rsUID] = append(rsAgg[rsUID], pod)
			continue
		}
		if _, ok := ownedWorkloadPods[pod.Name]; ok {
			continue // 已归入 StatefulSet/DaemonSet 子树
		}
		old := false
		if _, ok := oldPodNames[pod.Name]; ok {
			old = true
		}
		tree.addNode(ns, "Pod", pod.Name, podStatus(pod), podLabels(pod, old, controller), old)
		tree.addEdge("owner", appID, kindID("Pod", pod.Name))
	}
	// 无 pod 但存在的项目 RS(缩容到 0)也入图
	for uid := range rsByUID {
		if _, ok := rsAgg[uid]; !ok {
			rsAgg[uid] = nil
		}
	}
	// 按 RS 名排序输出,保证确定性布局（RS 名在命名空间内唯一,直接以名建表）
	rsNames := make([]string, 0, len(rsByUID))
	rsByName := make(map[string]*appsv1.ReplicaSet, len(rsByUID))
	for _, rs := range rsByUID {
		rsByName[rs.Name] = rs
		rsNames = append(rsNames, rs.Name)
	}
	sort.Strings(rsNames)
	for _, name := range rsNames {
		rs := rsByName[name]
		rsID := kindID("ReplicaSet", name)
		var depUID kmetatypes.UID
		for _, re := range rs.OwnerReferences {
			if re.Kind == "Deployment" {
				depUID = re.UID
				break
			}
		}
		old := latestRev[depUID] > 0 && revisionOf(rs) < latestRev[depUID]
		tree.addNode(ns, "ReplicaSet", name, rsStatus(rsAgg[rs.UID]), nil, old)
		tree.addEdge("owner", depIDByUID[depUID], rsID)
		for _, pod := range sortPods(rsAgg[rs.UID]) {
			podID := kindID("Pod", pod.Name)
			tree.addNode(ns, "Pod", pod.Name, podStatus(pod), podLabels(pod, old, ""), old)
			tree.addEdge("owner", rsID, podID)
		}
	}

	// 4. Service:selector 与项目 pod labels 有交集才算本项目服务,按名排序输出。
	services, err := k8sRepo.ListServices(ns)
	if err != nil {
		return nil, err
	}
	var svcNames []string
	svcMatched := make(map[string][]*corev1.Pod)
	for _, svc := range services {
		if len(svc.Spec.Selector) == 0 {
			continue
		}
		var matched []*corev1.Pod
		for _, pod := range pods {
			if selectorMatches(svc.Spec.Selector, pod.Labels) {
				matched = append(matched, pod)
			}
		}
		if len(matched) == 0 {
			continue
		}
		svcNames = append(svcNames, svc.Name)
		svcMatched[svc.Name] = matched
	}
	sort.Strings(svcNames)
	for _, name := range svcNames {
		svcID := kindID("Service", name)
		tree.addNode(ns, "Service", name, serviceStatus(svcMatched[name]), nil, false)
		for _, pod := range sortPods(svcMatched[name]) {
			tree.addEdge("selector", svcID, kindID("Pod", pod.Name))
		}
	}

	// 5. 聚合整体状态:任一 degraded→Failed,任一 progressing→Deploying,全健康→Deployed。
	//    无任何工作负载时回退项目记录的部署状态(避免新项目误报 unknown)。
	tree.Status = tree.aggregate()
	if len(tree.Nodes) == 1 {
		tree.Status = proj.DeployStatus
	}
	tree.Nodes[0].Status = deployStatusString(tree.Status)
	return tree, nil
}

// addNode 追加拓扑节点(ID 由 kind+name 推导,与前端 id 规则一致)。
func (t *ResourceTree) addNode(ns, kind, name, status string, labels map[string]string, old bool) {
	t.Nodes = append(t.Nodes, &ResourceTreeNode{
		ID: kindID(kind, name), Kind: kind, Name: name, Namespace: ns,
		Status: status, Labels: labels, Old: old,
	})
}

// addEdge 追加拓扑边(ID = type-source-target,天然唯一)。
func (t *ResourceTree) addEdge(typ, source, target string) {
	t.Edges = append(t.Edges, &ResourceTreeEdge{ID: typ + "-" + source + "-" + target, Type: typ, Source: source, Target: target})
}

// addWorkloadSubtree 为单个 StatefulSet/DaemonSet 建节点并聚合其属主 pod:
// workload 节点状态由调用方传入(复用 judge 滚动判定),owned 是其属主 pod
// (已按 UID 过滤);pod 挂 workload 下并建 owner 边,controller 标签省略
// (属主关系由边表达)。返回被聚合的 pod 名集合,调用方据此避免重复直挂 Application。
func (t *ResourceTree) addWorkloadSubtree(ns, appID, kind, name string, owned []*corev1.Pod, oldPodNames map[string]struct{}, status types.Deploy) map[string]struct{} {
	t.addNode(ns, kind, name, deployStatusString(status), nil, false)
	t.addEdge("owner", appID, kindID(kind, name))
	names := make(map[string]struct{}, len(owned))
	for _, pod := range sortPods(owned) {
		old := false
		if _, ok := oldPodNames[pod.Name]; ok {
			old = true
		}
		t.addNode(ns, "Pod", pod.Name, podStatus(pod), podLabels(pod, old, ""), old)
		t.addEdge("owner", kindID(kind, name), kindID("Pod", pod.Name))
		names[pod.Name] = struct{}{}
	}
	return names
}

// kindID 由资源 kind 与名称推导节点 id(全小写 + 连字符,如 deployment-demo-app)。
func kindID(kind, name string) string {
	return strings.ToLower(kind) + "-" + name
}

// aggregate 聚合非根节点状态为整体部署判定:任一 degraded→Failed,任一
// progressing/unknown→Deploying,全 healthy→Deployed,无子节点→Unknown。
func (t *ResourceTree) aggregate() types.Deploy {
	var children []string
	for i, n := range t.Nodes {
		if i > 0 {
			children = append(children, n.Status)
		}
	}
	if len(children) == 0 {
		return types.Deploy_StatusUnknown
	}
	agg := types.Deploy_StatusDeployed
	for _, s := range children {
		if s == "degraded" {
			return types.Deploy_StatusFailed
		}
	}
	for _, s := range children {
		if s == "progressing" || s == "unknown" {
			agg = types.Deploy_StatusDeploying
		}
	}
	return agg
}

// deployStatusString 把 types.Deploy 映射为前端 NodeStatus 字符串。
func deployStatusString(st types.Deploy) string {
	switch st {
	case types.Deploy_StatusDeployed:
		return "healthy"
	case types.Deploy_StatusDeploying:
		return "progressing"
	case types.Deploy_StatusFailed:
		return "degraded"
	default:
		return "unknown"
	}
}

// rsStatus 由 RS 名下 pod 状态聚合 RS 节点状态:任一 degraded→degraded,任一
// progressing→progressing,无 pod(已缩容到 0)视为 healthy,否则 healthy。
func rsStatus(pods []*corev1.Pod) string {
	if len(pods) == 0 {
		return "healthy"
	}
	return aggregatePodStatus(pods)
}

// serviceStatus 由匹配 pod 就绪情况聚合 Service 状态:有 ready→healthy,
// 有 degraded→degraded,否则(仅进行中)progressing。
func serviceStatus(pods []*corev1.Pod) string {
	ready, degraded := false, false
	for _, pod := range pods {
		switch podStatus(pod) {
		case "healthy":
			ready = true
		case "degraded":
			degraded = true
		}
	}
	switch {
	case ready:
		return "healthy"
	case degraded:
		return "degraded"
	default:
		return "progressing"
	}
}

// aggregatePodStatus 聚合一组 pod 状态:degraded 优先,其次 progressing,全健康才 healthy。
func aggregatePodStatus(pods []*corev1.Pod) string {
	degraded, progressing := false, false
	for _, pod := range pods {
		switch podStatus(pod) {
		case "degraded":
			degraded = true
		case "progressing":
			progressing = true
		}
	}
	switch {
	case degraded:
		return "degraded"
	case progressing:
		return "progressing"
	default:
		return "healthy"
	}
}

// podStatus 推导单个 pod 节点状态:终止中/等待就绪→progressing,Failed 或容器
// 稳定失败 reason→degraded,全部容器 Ready→healthy,否则 progressing。
func podStatus(pod *corev1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "progressing"
	}
	if pod.Status.Phase == corev1.PodFailed {
		return "degraded"
	}
	if pod.Status.Phase == corev1.PodPending {
		return "progressing"
	}
	if len(pod.Status.ContainerStatuses) == 0 {
		return "progressing"
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if w := cs.State.Waiting; w != nil {
			if _, ok := fatalWaitingReasons[w.Reason]; ok {
				return "degraded"
			}
		}
		if t := cs.State.Terminated; t != nil {
			if _, ok := fatalTerminatedReasons[t.Reason]; ok {
				return "degraded"
			}
		}
	}
	if podAllContainersReady(pod) {
		return "healthy"
	}
	return "progressing"
}

// podLabels 构造 pod 节点标签:记录阶段、旧版本标记与非 RS 属主(sts/ds 直挂场景)。
func podLabels(pod *corev1.Pod, old bool, controller string) map[string]string {
	labels := map[string]string{"phase": string(pod.Status.Phase)}
	if old {
		labels["old"] = "true"
	}
	if controller != "" {
		labels["controller"] = controller
	}
	return labels
}

// selectorMatches 判断 k8s Service selector 与 pod labels 是否完全匹配(空 selector 不匹配,
// 避免把命名空间内无选择器服务误纳入)。
func selectorMatches(selector, labels map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

// sortPods 按创建时间升序、名称兜底排序,保证树输出与前端布局确定性。
func sortPods(pods []*corev1.Pod) []*corev1.Pod {
	sorted := append([]*corev1.Pod(nil), pods...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if !sorted[i].CreationTimestamp.Time.Equal(sorted[j].CreationTimestamp.Time) {
			return sorted[i].CreationTimestamp.Time.Before(sorted[j].CreationTimestamp.Time)
		}
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}
