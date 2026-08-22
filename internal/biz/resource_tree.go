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
	networkingv1 "k8s.io/api/networking/v1"
	kmetatypes "k8s.io/apimachinery/pkg/types"
)

// ResourceTree 是项目资源拓扑树的领域返回:整体部署状态 + 节点 + 边。
type ResourceTree struct {
	Status types.Deploy
	Nodes  []*ResourceTreeNode
	Edges  []*ResourceTreeEdge
}

// ResourceTreeNode 是拓扑树节点。kind/status 与前端拓扑渲染器对齐:
// kind ∈ Application|Ingress|Service|Deployment|ReplicaSet|Pod|StatefulSet|DaemonSet,
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

// ResourceTreeEdge 是拓扑树边:Type=owner 属主引用 / selector 标签选择器 / route 路由后端。
type ResourceTreeEdge struct {
	ID     string
	Type   string
	Source string
	Target string
}

// buildResourceTree 构建项目资源拓扑树，按访问链路分层：
// Application → Ingress → Service → Deployment/StatefulSet/DaemonSet → ReplicaSet → Pod。
//
// Ingress 从命名空间列表筛 backend 命中项目 Service 的（route 边连到该 svc，
// backend 不指向项目 svc 的不入图）；Service 用 selector 与项目 pod labels 的交集
// 判定归属，并据此覆盖其下的 workload；workload（deploy/sts/ds）挂在覆盖它的
// Service 下，无 Service 覆盖时兜底直挂 Application；无 Ingress 覆盖的 Service
// 同样兜底直挂
// Application。ReplicaSet 从命名空间列表里筛属主是本项目 Deployment 的，按 revision
// 注解识别旧版本；pod 按其属主归到对应 workload 下（RS 属主归 Deployment 子树，
// sts/ds 属主归其子树，old 复用 collectWorkloadOldPods 的 hash 判定），其余裸 pod
// 直挂 Application。与 AllContainers 一致：不裁剪 workload 类型，且剔除 Failed 阶段
// pod（视为非活跃资源）。
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
	// 与 AllContainers 一致：Failed 阶段 pod 视为非活跃，不参与拓扑与状态判定
	pods = activePods(pods)
	deployments, statefulSets, daemonSets := k8sRepo.GetWorkloadsByManifest(proj.Manifest)
	rss, err := k8sRepo.ListReplicaSets(ns)
	if err != nil {
		return nil, err
	}

	// podWorkload 记录 pod → 属主 workload 节点 id（无属主 workload 的裸 pod 不在其中），
	// 供 Service selector 覆盖 workload 的判定。
	podWorkload := make(map[string]string)

	// 1. Deployment 节点：按 manifest 名字读实时对象判滚动状态（父边延后到 svc 覆盖确定后挂载）。
	//    尚未创建(部署刚发起)时占位为 progressing 节点。
	depByUID := make(map[kmetatypes.UID]*appsv1.Deployment)
	depIDByUID := make(map[kmetatypes.UID]string)
	for _, wl := range deployments {
		live, err := k8sRepo.GetDeployment(ns, wl.Name)
		if err != nil {
			if errs.IsNotFound(err) {
				tree.addNode(ns, "Deployment", wl.Name, "progressing", nil, false)
				continue
			}
			return nil, err
		}
		depByUID[live.UID] = live
		depIDByUID[live.UID] = kindID("Deployment", live.Name)
		st, _, _ := judgeDeploymentRollout(live, rss, pods)
		tree.addNode(ns, "Deployment", live.Name, deployStatusString(st), nil, false)
	}

	// 2. StatefulSet / DaemonSet 节点与属主 pod 子树（父边延后）：与 AllContainers 一致
	//    不做类型裁剪，sts/ds pod 完整入图。未创建占位 progressing；old 标记复用
	//    collectWorkloadOldPods 的 hash 判定。
	oldPodNames := collectWorkloadOldPods(k8sRepo, pods)
	ownedWorkloadPods := make(map[string]struct{}) // 已归入 sts/ds 子树的 pod，避免重复直挂 Application
	for _, wl := range statefulSets {
		live, err := k8sRepo.GetStatefulSet(ns, wl.Name)
		if err != nil {
			if errs.IsNotFound(err) {
				tree.addNode(ns, "StatefulSet", wl.Name, "progressing", nil, false)
				continue
			}
			return nil, err
		}
		owned := podsOwnedBy(pods, live.UID)
		st, _, _ := judgeStatefulSetRollout(live, owned)
		wkID := tree.addWorkloadSubtree(ns, "StatefulSet", live.Name, owned, oldPodNames, st)
		for _, p := range owned {
			podWorkload[p.Name] = wkID
			ownedWorkloadPods[p.Name] = struct{}{}
		}
	}
	for _, wl := range daemonSets {
		live, err := k8sRepo.GetDaemonSet(ns, wl.Name)
		if err != nil {
			if errs.IsNotFound(err) {
				tree.addNode(ns, "DaemonSet", wl.Name, "progressing", nil, false)
				continue
			}
			return nil, err
		}
		owned := podsOwnedBy(pods, live.UID)
		st, _, _ := judgeDaemonSetRollout(live, owned)
		wkID := tree.addWorkloadSubtree(ns, "DaemonSet", live.Name, owned, oldPodNames, st)
		for _, p := range owned {
			podWorkload[p.Name] = wkID
			ownedWorkloadPods[p.Name] = struct{}{}
		}
	}

	// 3. ReplicaSet：筛属主为本项目 Deployment 的 RS，按 revision 标旧，聚合其 pod。
	rsByUID := make(map[kmetatypes.UID]*appsv1.ReplicaSet)
	for _, rs := range rss {
		if _, ok := depByUID[deploymentOwnerUID(rs)]; !ok {
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

	// pod 归类:属项目 RS → 聚合到 RS 并记录属主 Deployment；属 sts/ds → 已在其子树；
	// 否则裸 pod 直挂 Application（controller 打标，old 标记 hash 判定仍生效）。
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
		if rs, ok := rsByUID[rsUID]; ok {
			rsAgg[rsUID] = append(rsAgg[rsUID], pod)
			if depID, ok := depIDByUID[deploymentOwnerUID(rs)]; ok {
				podWorkload[pod.Name] = depID
			}
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
		depUID := deploymentOwnerUID(rs)
		depID := depIDByUID[depUID]
		old := latestRev[depUID] > 0 && revisionOf(rs) < latestRev[depUID]
		tree.addNode(ns, "ReplicaSet", name, rsStatus(rsAgg[rs.UID]), nil, old)
		tree.addEdge("owner", depID, rsID)
		for _, pod := range sortPods(rsAgg[rs.UID]) {
			podID := kindID("Pod", pod.Name)
			tree.addNode(ns, "Pod", pod.Name, podStatus(pod), podLabels(pod, old, ""), old)
			tree.addEdge("owner", rsID, podID)
		}
	}

	// 4. Service：selector 与项目 pod labels 有交集才算项目服务；同时算出它覆盖的
	//    workload（selector 命中的 pod 归属哪个 workload 节点）。
	services, err := k8sRepo.ListServices(ns)
	if err != nil {
		return nil, err
	}
	var svcNames []string
	svcMatched := make(map[string][]*corev1.Pod) // svc 名 → 命中的项目 pod
	svcWorkloads := make(map[string][]string)    // svc 名 → 覆盖的 workload 节点 id（按 id 排序）
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
		covered := make(map[string]struct{})
		for _, pod := range matched {
			if wkID, ok := podWorkload[pod.Name]; ok {
				covered[wkID] = struct{}{}
			}
		}
		ids := make([]string, 0, len(covered))
		for id := range covered {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		svcWorkloads[svc.Name] = ids
	}
	sort.Strings(svcNames)

	// 5. Ingress：backend 命中项目 Service 才算项目 ingress（route 边连到该 svc）。
	ingresses, err := k8sRepo.ListIngresses(ns)
	if err != nil {
		return nil, err
	}
	var ingressNames []string
	ingressSvc := make(map[string][]string) // ingress 名 → backend 命中的 svc 名
	for _, ing := range ingresses {
		var covered []string
		for _, svcName := range ingressBackendServiceNames(ing) {
			if _, ok := svcMatched[svcName]; ok {
				covered = append(covered, svcName)
			}
		}
		if len(covered) == 0 {
			continue // backend 不指向项目 svc 的 ingress 不入图
		}
		ingressNames = append(ingressNames, ing.Name)
		ingressSvc[ing.Name] = covered
	}
	sort.Strings(ingressNames)

	// 6. 挂父边：Ingress 恒挂 Application；Service 有 ingress 覆盖走 route 边否则兜底挂
	//    Application；workload 有 svc 覆盖走 selector 边否则兜底挂 Application。
	ingressCoveredSvc := make(map[string]struct{})
	for _, svcs := range ingressSvc {
		for _, svcName := range svcs {
			ingressCoveredSvc[svcName] = struct{}{}
		}
	}
	workloadSvcs := make(map[string][]string)
	for svcName, ids := range svcWorkloads {
		for _, id := range ids {
			workloadSvcs[id] = append(workloadSvcs[id], svcName)
		}
	}
	for id := range workloadSvcs {
		sort.Strings(workloadSvcs[id])
	}

	for _, name := range ingressNames {
		ingID := kindID("Ingress", name)
		var statuses []string
		for _, svcName := range ingressSvc[name] {
			statuses = append(statuses, serviceStatus(svcMatched[svcName]))
		}
		tree.addNode(ns, "Ingress", name, ingressStatus(statuses), nil, false)
		tree.addEdge("owner", appID, ingID)
		for _, svcName := range ingressSvc[name] {
			tree.addEdge("route", ingID, kindID("Service", svcName))
		}
	}
	for _, name := range svcNames {
		tree.addNode(ns, "Service", name, serviceStatus(svcMatched[name]), nil, false)
		if _, ok := ingressCoveredSvc[name]; ok {
			continue // route 边已由覆盖它的 ingress 挂出
		}
		tree.addEdge("owner", appID, kindID("Service", name))
	}
	// workload 父边按节点 id 排序保证确定性输出
	var workloadIDs []string
	for _, n := range tree.Nodes {
		switch n.Kind {
		case "Deployment", "StatefulSet", "DaemonSet":
			workloadIDs = append(workloadIDs, n.ID)
		}
	}
	sort.Strings(workloadIDs)
	for _, id := range workloadIDs {
		svcs := workloadSvcs[id]
		if len(svcs) == 0 {
			tree.addEdge("owner", appID, id)
			continue
		}
		for _, svcName := range svcs {
			tree.addEdge("selector", kindID("Service", svcName), id)
		}
	}

	// 7. 聚合整体状态:任一 degraded→Failed,任一 progressing→Deploying,全健康→Deployed。
	//    无任何资源时回退项目记录的部署状态(避免新项目误报 unknown)。
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

// addWorkloadSubtree 为单个 StatefulSet/DaemonSet 建节点并聚合其属主 pod：
// workload 节点状态由调用方传入(复用 judge 滚动判定),owned 是其属主 pod
// (已按 UID 过滤);pod 挂 workload 下并建 owner 边,controller 标签省略
// (属主关系由边表达)。返回 workload 节点 id；父边(svc selector / app owner)
// 由调用方在 svc 覆盖确定后统一挂载。
func (t *ResourceTree) addWorkloadSubtree(ns, kind, name string, owned []*corev1.Pod, oldPodNames map[string]struct{}, status types.Deploy) string {
	wkID := kindID(kind, name)
	t.addNode(ns, kind, name, deployStatusString(status), nil, false)
	for _, pod := range sortPods(owned) {
		old := false
		if _, ok := oldPodNames[pod.Name]; ok {
			old = true
		}
		t.addNode(ns, "Pod", pod.Name, podStatus(pod), podLabels(pod, old, ""), old)
		t.addEdge("owner", wkID, kindID("Pod", pod.Name))
	}
	return wkID
}

// kindID 由资源 kind 与名称推导节点 id(全小写 + 连字符,如 deployment-demo-app)。
func kindID(kind, name string) string {
	return strings.ToLower(kind) + "-" + name
}

// deploymentOwnerUID 返回 ReplicaSet 属主 Deployment 的 UID；无 Deployment 属主时返回空 UID。
func deploymentOwnerUID(rs *appsv1.ReplicaSet) kmetatypes.UID {
	for _, re := range rs.OwnerReferences {
		if re.Kind == "Deployment" {
			return re.UID
		}
	}
	return ""
}

// ingressBackendServiceNames 提取 Ingress 全部 backend 引用的 Service 名
// （默认 backend + 各 rule 的 path backend，去重），供 route 边与项目归属判定。
func ingressBackendServiceNames(ing *networkingv1.Ingress) []string {
	seen := make(map[string]struct{})
	var names []string
	add := func(b *networkingv1.IngressBackend) {
		if b == nil || b.Service == nil {
			return
		}
		if _, ok := seen[b.Service.Name]; ok {
			return
		}
		seen[b.Service.Name] = struct{}{}
		names = append(names, b.Service.Name)
	}
	add(ing.Spec.DefaultBackend)
	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for _, path := range rule.HTTP.Paths {
			add(&path.Backend)
		}
	}
	return names
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

// ingressStatus 由后端 svc 状态聚合 Ingress 状态:任一 degraded→degraded,任一
// progressing→progressing,全 healthy→healthy。ingress 仅在覆盖项目 svc 时入图,
// 因此不会出现无后端的情况。
func ingressStatus(statuses []string) string {
	degraded, progressing := false, false
	for _, s := range statuses {
		switch s {
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

// podStatus 推导单个 pod 节点状态:终止中/等待就绪→progressing,容器稳定失败
// reason→degraded,全部容器 Ready→healthy,否则 progressing。
// 入参约定为非 Failed 阶段 pod(调用方已过 activePods,Failed 不入图)。
func podStatus(pod *corev1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "progressing"
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
