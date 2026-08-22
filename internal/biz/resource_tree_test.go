package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmetatypes "k8s.io/apimachinery/pkg/types"
)

// fakeTreeK8sRepo 是 buildResourceTree 的 K8sRepo 替身，覆写资源树推导用到的全部原语
// （ListPodsBySelectors/ListReplicaSets/GetDeployment/GetWorkloadsByManifest/ListServices/
// GetStatefulSet/GetDaemonSet），collectWorkloadOldPods 分类 StatefulSet 旧副本用 GetStatefulSet。
type fakeTreeK8sRepo struct {
	K8sRepo
	pods         []*corev1.Pod
	listPodsErr  error
	rss          []*appsv1.ReplicaSet
	listRSErr    error
	deployments  map[string]*appsv1.Deployment
	getDepErr    error
	manifestDeps []*appsv1.Deployment
	services     []*corev1.Service
	listSvcErr   error
	statefulSets map[string]*appsv1.StatefulSet
	daemonSets   map[string]*appsv1.DaemonSet
	manifestSts  []*appsv1.StatefulSet
	manifestDs   []*appsv1.DaemonSet
	getStsErr    error
	getDsErr     error
}

func (f *fakeTreeK8sRepo) ListPodsBySelectors(namespace string, selectors []string) ([]*corev1.Pod, error) {
	return f.pods, f.listPodsErr
}

func (f *fakeTreeK8sRepo) ListReplicaSets(namespace string) ([]*appsv1.ReplicaSet, error) {
	return f.rss, f.listRSErr
}

func (f *fakeTreeK8sRepo) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	if f.getDepErr != nil {
		return nil, f.getDepErr
	}
	return f.deployments[name], nil
}

func (f *fakeTreeK8sRepo) GetWorkloadsByManifest(manifests []string) ([]*appsv1.Deployment, []*appsv1.StatefulSet, []*appsv1.DaemonSet) {
	return f.manifestDeps, f.manifestSts, f.manifestDs
}

func (f *fakeTreeK8sRepo) ListServices(namespace string) ([]*corev1.Service, error) {
	return f.services, f.listSvcErr
}

func (f *fakeTreeK8sRepo) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	if f.getStsErr != nil {
		return nil, f.getStsErr
	}
	return f.statefulSets[name], nil
}

func (f *fakeTreeK8sRepo) GetDaemonSet(namespace, name string) (*appsv1.DaemonSet, error) {
	if f.getDsErr != nil {
		return nil, f.getDsErr
	}
	return f.daemonSets[name], nil
}

// readyPod 构造一个 Running + 全部容器 Ready 的 pod 骨架。
func readyPod(name, rsName string, rsUID kmetatypes.UID) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "ns",
			Labels:    map[string]string{"app": "demo"},
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: rsName, UID: rsUID},
			},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status: corev1.PodStatus{
			Phase:             corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{{Name: "web", Ready: true}},
		},
	}
}

// readyDeployment 构造一个已收敛到 Deployed 的 Deployment 骨架（spec/status 计数一致）。
func readyDeployment(name string, uid kmetatypes.UID) *appsv1.Deployment {
	replicas := int32(2)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "ns", UID: uid, Generation: 1},
		Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 1,
			UpdatedReplicas:    replicas,
			AvailableReplicas:  replicas,
		},
	}
}

// treeNode 在树中按 id 查找节点，未找到返回 nil。
func treeNode(t *testing.T, tree *ResourceTree, id string) *ResourceTreeNode {
	t.Helper()
	for _, n := range tree.Nodes {
		if n.ID == id {
			return n
		}
	}
	return nil
}

// hasTreeEdge 判断树中是否存在指定 type/source/target 的边。
func hasTreeEdge(tree *ResourceTree, typ, source, target string) bool {
	for _, e := range tree.Edges {
		if e.Type == typ && e.Source == source && e.Target == target {
			return true
		}
	}
	return false
}

// TestBuildResourceTree_EmptySelectors 空 PodSelectors（从未部署）只回 Application 根节点，
// status 跟随项目记录的部署状态，不产生任何 k8s 调用。
func TestBuildResourceTree_EmptySelectors(t *testing.T) {
	k := &fakeTreeK8sRepo{}
	proj := &Project{
		ID: 1, Name: "demo-app", Namespace: &Namespace{Name: "ns"},
		DeployStatus: types.Deploy_StatusDeployed,
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.NoError(t, err)
	if assert.Len(t, got.Nodes, 1) {
		assert.Equal(t, "application-1", got.Nodes[0].ID)
		assert.Equal(t, "Application", got.Nodes[0].Kind)
		assert.Equal(t, "healthy", got.Nodes[0].Status)
	}
	assert.Empty(t, got.Edges)
	assert.Equal(t, types.Deploy_StatusDeployed, got.Status)
}

// TestBuildResourceTree_HappyPath 覆盖完整滚动发布树：Deployment → 新/旧 RS → Pod 属主链、
// revision 新旧判定、StatefulSet 旧副本子树、Service selector 边与整体聚合。
func TestBuildResourceTree_HappyPath(t *testing.T) {
	oldRS := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-app-7c3e9d1f5", UID: "rs-old-uid",
			Annotations:     map[string]string{RevisionAnnotation: "1"},
			OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", UID: "dep-1"}},
		},
	}
	newRS := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-app-8b2f4e7d", UID: "rs-new-uid",
			Annotations:     map[string]string{RevisionAnnotation: "2"},
			OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", UID: "dep-1"}},
		},
	}
	podNew1 := readyPod("pod-new-1", "demo-app-8b2f4e7d", "rs-new-uid")
	podNew2 := readyPod("pod-new-2", "demo-app-8b2f4e7d", "rs-new-uid")
	podOld1 := readyPod("pod-old-1", "demo-app-7c3e9d1f5", "rs-old-uid")
	// StatefulSet 旧副本：hash 不匹配 updateRevision 标旧，挂 sts 子树下
	podSts := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod-sts", Namespace: "ns",
			Labels:          map[string]string{"app": "demo", appsv1.ControllerRevisionHashLabelKey: "rev1"},
			OwnerReferences: []metav1.OwnerReference{{Kind: "StatefulSet", Name: "sts", UID: "sts-uid"}},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status: corev1.PodStatus{
			Phase:             corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{{Name: "web", Ready: true}},
		},
	}
	k := &fakeTreeK8sRepo{
		pods: []*corev1.Pod{podNew1, podNew2, podOld1, podSts},
		rss:  []*appsv1.ReplicaSet{oldRS, newRS},
		deployments: map[string]*appsv1.Deployment{
			"demo-app": readyDeployment("demo-app", "dep-1"),
		},
		manifestDeps: []*appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "demo-app"}}},
		services: []*corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "demo-app-svc", Namespace: "ns"},
			Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "demo"}},
		}},
		manifestSts: []*appsv1.StatefulSet{{ObjectMeta: metav1.ObjectMeta{Name: "sts"}}},
		statefulSets: map[string]*appsv1.StatefulSet{"sts": {
			ObjectMeta: metav1.ObjectMeta{Name: "sts", Namespace: "ns", UID: "sts-uid"},
			Status:     appsv1.StatefulSetStatus{UpdateRevision: "rev2"},
		}},
	}
	proj := &Project{
		ID: 1, Name: "demo-app", Namespace: &Namespace{Name: "ns"},
		PodSelectors: []string{"app=demo"}, Manifest: []string{"demo.yaml"},
		DeployStatus: types.Deploy_StatusDeployed,
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.NoError(t, err)

	// 根节点 + 聚合状态（全健康 → Deployed）
	assert.Equal(t, "application-1", got.Nodes[0].ID)
	assert.Equal(t, "healthy", got.Nodes[0].Status)
	assert.Equal(t, types.Deploy_StatusDeployed, got.Status)

	// Deployment：状态 healthy，Application→Deployment owner 边
	dep := treeNode(t, got, "deployment-demo-app")
	if assert.NotNil(t, dep) {
		assert.Equal(t, "Deployment", dep.Kind)
		assert.Equal(t, "healthy", dep.Status)
	}
	assert.True(t, hasTreeEdge(got, "owner", "application-1", "deployment-demo-app"))

	// 新/旧 RS：revision 2 为新、revision 1 标旧；各自挂在 Deployment 下
	rsNew := treeNode(t, got, "replicaset-demo-app-8b2f4e7d")
	rsOld := treeNode(t, got, "replicaset-demo-app-7c3e9d1f5")
	if assert.NotNil(t, rsNew) && assert.NotNil(t, rsOld) {
		assert.False(t, rsNew.Old)
		assert.True(t, rsOld.Old)
		assert.Equal(t, "healthy", rsNew.Status)
		assert.Equal(t, "healthy", rsOld.Status)
	}
	assert.True(t, hasTreeEdge(got, "owner", "deployment-demo-app", "replicaset-demo-app-8b2f4e7d"))
	assert.True(t, hasTreeEdge(got, "owner", "deployment-demo-app", "replicaset-demo-app-7c3e9d1f5"))

	// RS 下 pod：新 RS 两个、旧 RS 一个，旧 pod 标旧
	for _, id := range []string{"pod-pod-new-1", "pod-pod-new-2"} {
		if n := treeNode(t, got, id); assert.NotNil(t, n) {
			assert.False(t, n.Old)
			assert.True(t, hasTreeEdge(got, "owner", "replicaset-demo-app-8b2f4e7d", id))
		}
	}
	oldPod := treeNode(t, got, "pod-pod-old-1")
	if assert.NotNil(t, oldPod) {
		assert.True(t, oldPod.Old)
		assert.True(t, hasTreeEdge(got, "owner", "replicaset-demo-app-7c3e9d1f5", "pod-pod-old-1"))
	}

	// StatefulSet 节点 + 旧副本 pod：sts 子树挂 Application，pod 挂 sts 下，old 标记
	stsNode := treeNode(t, got, "statefulset-sts")
	if assert.NotNil(t, stsNode) {
		assert.Equal(t, "StatefulSet", stsNode.Kind)
		assert.Equal(t, "healthy", stsNode.Status)
	}
	assert.True(t, hasTreeEdge(got, "owner", "application-1", "statefulset-sts"))
	stsPod := treeNode(t, got, "pod-pod-sts")
	if assert.NotNil(t, stsPod) {
		assert.Equal(t, "true", stsPod.Labels["old"])
		assert.Empty(t, stsPod.Labels["controller"])
	}
	assert.True(t, hasTreeEdge(got, "owner", "statefulset-sts", "pod-pod-sts"))

	// Service：健康 + 到全部 4 个 pod 的 selector 边
	svc := treeNode(t, got, "service-demo-app-svc")
	if assert.NotNil(t, svc) {
		assert.Equal(t, "Service", svc.Kind)
		assert.Equal(t, "healthy", svc.Status)
	}
	for _, id := range []string{"pod-pod-new-1", "pod-pod-new-2", "pod-pod-old-1", "pod-pod-sts"} {
		assert.True(t, hasTreeEdge(got, "selector", "service-demo-app-svc", id), "缺少 selector 边: "+id)
	}
}

// TestBuildResourceTree_Degraded 覆盖容器稳定失败（CrashLoopBackOff）沿 Pod→RS→Application
// 逐级降级，整体聚合为 Failed。
func TestBuildResourceTree_Degraded(t *testing.T) {
	rs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-app-8b2f4e7d", UID: "rs-new-uid",
			Annotations:     map[string]string{RevisionAnnotation: "2"},
			OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", UID: "dep-1"}},
		},
	}
	bad := readyPod("pod-bad", "demo-app-8b2f4e7d", "rs-new-uid")
	bad.Status.ContainerStatuses[0].State.Waiting = &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff", Message: "back-off"}
	k := &fakeTreeK8sRepo{
		pods: []*corev1.Pod{bad},
		rss:  []*appsv1.ReplicaSet{rs},
		deployments: map[string]*appsv1.Deployment{
			"demo-app": readyDeployment("demo-app", "dep-1"),
		},
		manifestDeps: []*appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "demo-app"}}},
	}
	proj := &Project{
		ID: 1, Name: "demo-app", Namespace: &Namespace{Name: "ns"},
		PodSelectors: []string{"app=demo"}, Manifest: []string{"demo.yaml"},
		DeployStatus: types.Deploy_StatusDeployed,
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.NoError(t, err)

	assert.Equal(t, "degraded", treeNode(t, got, "pod-pod-bad").Status)
	assert.Equal(t, "degraded", treeNode(t, got, "replicaset-demo-app-8b2f4e7d").Status)
	assert.Equal(t, "degraded", treeNode(t, got, "deployment-demo-app").Status)
	assert.Equal(t, "degraded", got.Nodes[0].Status)
	assert.Equal(t, types.Deploy_StatusFailed, got.Status)
}

// TestBuildResourceTree_FilterFailedPod 覆盖 Failed 阶段 pod 被剔除:与 AllContainers 一致,
// Failed pod 不入图(无 pod 节点与 owner 边),不参与聚合;非 Failed pod 正常入图。
func TestBuildResourceTree_FilterFailedPod(t *testing.T) {
	rs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-app-8b2f4e7d", UID: "rs-new-uid",
			Annotations:     map[string]string{RevisionAnnotation: "2"},
			OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", UID: "dep-1"}},
		},
	}
	ready := readyPod("pod-ready", "demo-app-8b2f4e7d", "rs-new-uid")
	failed := readyPod("pod-failed", "demo-app-8b2f4e7d", "rs-new-uid")
	failed.Status.Phase = corev1.PodFailed
	replicas := int32(1)
	k := &fakeTreeK8sRepo{
		pods: []*corev1.Pod{ready, failed},
		rss:  []*appsv1.ReplicaSet{rs},
		deployments: map[string]*appsv1.Deployment{
			"demo-app": {
				ObjectMeta: metav1.ObjectMeta{Name: "demo-app", Namespace: "ns", UID: "dep-1", Generation: 1},
				Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
				Status: appsv1.DeploymentStatus{
					ObservedGeneration: 1, UpdatedReplicas: 1, AvailableReplicas: 1,
				},
			},
		},
		manifestDeps: []*appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "demo-app"}}},
	}
	proj := &Project{
		ID: 1, Name: "demo-app", Namespace: &Namespace{Name: "ns"},
		PodSelectors: []string{"app=demo"}, Manifest: []string{"demo.yaml"},
		DeployStatus: types.Deploy_StatusDeployed,
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.NoError(t, err)

	// Failed pod 无节点、无 owner 边;ready pod 正常挂 RS 下
	assert.Nil(t, treeNode(t, got, "pod-pod-failed"))
	assert.False(t, hasTreeEdge(got, "owner", "replicaset-demo-app-8b2f4e7d", "pod-pod-failed"))
	if n := treeNode(t, got, "pod-pod-ready"); assert.NotNil(t, n) {
		assert.Equal(t, "healthy", n.Status)
	}
	assert.True(t, hasTreeEdge(got, "owner", "replicaset-demo-app-8b2f4e7d", "pod-pod-ready"))
	// 全健康 → Deployed,根节点 healthy
	assert.Equal(t, types.Deploy_StatusDeployed, got.Status)
	assert.Equal(t, "healthy", got.Nodes[0].Status)
}

// TestBuildResourceTree_DeploymentNotCreated 覆盖部署刚发起、Deployment 尚未创建的场景：
// 占位为 progressing 节点，仍挂 Application 下，聚合为 Deploying。
func TestBuildResourceTree_DeploymentNotCreated(t *testing.T) {
	k := &fakeTreeK8sRepo{
		deployments:  map[string]*appsv1.Deployment{},
		getDepErr:    errs.NotFound("deployment not found"),
		manifestDeps: []*appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "demo-app"}}},
	}
	proj := &Project{
		ID: 1, Name: "demo-app", Namespace: &Namespace{Name: "ns"},
		PodSelectors: []string{"app=demo"}, Manifest: []string{"demo.yaml"},
		DeployStatus: types.Deploy_StatusDeployed,
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.NoError(t, err)

	dep := treeNode(t, got, "deployment-demo-app")
	if assert.NotNil(t, dep) {
		assert.Equal(t, "progressing", dep.Status)
	}
	assert.True(t, hasTreeEdge(got, "owner", "application-1", "deployment-demo-app"))
	assert.Equal(t, "progressing", got.Nodes[0].Status)
	assert.Equal(t, types.Deploy_StatusDeploying, got.Status)
}

// TestBuildResourceTree_ListPodsError 上抛 ListPodsBySelectors 错误。
func TestBuildResourceTree_ListPodsError(t *testing.T) {
	k := &fakeTreeK8sRepo{listPodsErr: errors.New("list down")}
	proj := &Project{
		ID: 1, Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=demo"},
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "list down")
}

// workloadPod 构造一个 Running + 全部容器 Ready、属主为指定 workload 的 pod 骨架。
// rev 写入 controller-revision-hash 标签，供 statefulSetNewPods/daemonSetNewPods 新旧判定。
func workloadPod(name, kind, owner string, uid kmetatypes.UID, rev string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: "ns",
			Labels:          map[string]string{"app": "demo", appsv1.ControllerRevisionHashLabelKey: rev},
			OwnerReferences: []metav1.OwnerReference{{Kind: kind, Name: owner, UID: uid}},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status: corev1.PodStatus{
			Phase:             corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{{Name: "web", Ready: true}},
		},
	}
}

// TestBuildResourceTree_StatefulSetDaemonSet 覆盖 StatefulSet/DaemonSet 与裸 pod 同图：
// sts/ds 各成子树挂 Application 下（属主 pod 聚合其下、owner 边齐全），
// 裸 pod 仍直挂 Application，Service selector 边覆盖全部 pod。
func TestBuildResourceTree_StatefulSetDaemonSet(t *testing.T) {
	stsReplicas := int32(2)
	stsPod1 := workloadPod("pod-sts-1", "StatefulSet", "sts", "sts-uid", "rev2")
	stsPod2 := workloadPod("pod-sts-2", "StatefulSet", "sts", "sts-uid", "rev2")
	dsPod := workloadPod("pod-ds-1", "DaemonSet", "ds", "ds-uid", "rev1")
	bare := readyPod("pod-bare", "", "")
	bare.ObjectMeta.OwnerReferences = nil

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "sts", Namespace: "ns", UID: "sts-uid", Generation: 1},
		Spec:       appsv1.StatefulSetSpec{Replicas: &stsReplicas},
		Status: appsv1.StatefulSetStatus{
			ObservedGeneration: 1, UpdatedReplicas: 2, AvailableReplicas: 2, UpdateRevision: "rev2",
		},
	}
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "ds", Namespace: "ns", UID: "ds-uid", Generation: 1},
		Status: appsv1.DaemonSetStatus{
			ObservedGeneration: 1, DesiredNumberScheduled: 1, UpdatedNumberScheduled: 1, NumberAvailable: 1,
		},
	}
	k := &fakeTreeK8sRepo{
		pods:         []*corev1.Pod{stsPod1, stsPod2, dsPod, bare},
		manifestSts:  []*appsv1.StatefulSet{{ObjectMeta: metav1.ObjectMeta{Name: "sts"}}},
		manifestDs:   []*appsv1.DaemonSet{{ObjectMeta: metav1.ObjectMeta{Name: "ds"}}},
		statefulSets: map[string]*appsv1.StatefulSet{"sts": sts},
		daemonSets:   map[string]*appsv1.DaemonSet{"ds": ds},
		services: []*corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "app-svc", Namespace: "ns"},
			Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "demo"}},
		}},
	}
	proj := &Project{
		ID: 1, Name: "demo-app", Namespace: &Namespace{Name: "ns"},
		PodSelectors: []string{"app=demo"}, Manifest: []string{"sts.yaml", "ds.yaml"},
		DeployStatus: types.Deploy_StatusDeployed,
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.NoError(t, err)

	// 整体聚合：全健康 → Deployed
	assert.Equal(t, "healthy", got.Nodes[0].Status)
	assert.Equal(t, types.Deploy_StatusDeployed, got.Status)

	// StatefulSet 子树：节点 + 两个属主 pod + Application/sts 与 sts/pod 两级 owner 边
	stsNode := treeNode(t, got, "statefulset-sts")
	if assert.NotNil(t, stsNode) {
		assert.Equal(t, "StatefulSet", stsNode.Kind)
		assert.Equal(t, "healthy", stsNode.Status)
		assert.False(t, stsNode.Old)
	}
	assert.True(t, hasTreeEdge(got, "owner", "application-1", "statefulset-sts"))
	for _, id := range []string{"pod-pod-sts-1", "pod-pod-sts-2"} {
		if n := treeNode(t, got, id); assert.NotNil(t, n) {
			assert.False(t, n.Old)
		}
		assert.True(t, hasTreeEdge(got, "owner", "statefulset-sts", id))
	}

	// DaemonSet 子树：节点 + 一个属主 pod
	dsNode := treeNode(t, got, "daemonset-ds")
	if assert.NotNil(t, dsNode) {
		assert.Equal(t, "DaemonSet", dsNode.Kind)
		assert.Equal(t, "healthy", dsNode.Status)
	}
	assert.True(t, hasTreeEdge(got, "owner", "application-1", "daemonset-ds"))
	assert.True(t, hasTreeEdge(got, "owner", "daemonset-ds", "pod-pod-ds-1"))

	// 裸 pod：仍直挂 Application，无 controller 标签、不标旧
	bareNode := treeNode(t, got, "pod-pod-bare")
	if assert.NotNil(t, bareNode) {
		assert.False(t, bareNode.Old)
		assert.Empty(t, bareNode.Labels["controller"])
	}
	assert.True(t, hasTreeEdge(got, "owner", "application-1", "pod-pod-bare"))

	// Service selector 边覆盖 sts/ds/bare 全部 4 个 pod
	for _, id := range []string{"pod-pod-sts-1", "pod-pod-sts-2", "pod-pod-ds-1", "pod-pod-bare"} {
		assert.True(t, hasTreeEdge(got, "selector", "service-app-svc", id), "缺少 selector 边: "+id)
	}
}

// TestBuildResourceTree_StatefulSetDaemonSetNotCreated 覆盖 manifest 声明但尚未创建的
// sts/ds：占位 progressing 节点 + Application 属主边，整体聚合为 Deploying。
func TestBuildResourceTree_StatefulSetDaemonSetNotCreated(t *testing.T) {
	k := &fakeTreeK8sRepo{
		statefulSets: map[string]*appsv1.StatefulSet{},
		daemonSets:   map[string]*appsv1.DaemonSet{},
		manifestSts:  []*appsv1.StatefulSet{{ObjectMeta: metav1.ObjectMeta{Name: "sts"}}},
		manifestDs:   []*appsv1.DaemonSet{{ObjectMeta: metav1.ObjectMeta{Name: "ds"}}},
		getStsErr:    errs.NotFound("sts not found"),
		getDsErr:     errs.NotFound("ds not found"),
	}
	proj := &Project{
		ID: 1, Name: "demo-app", Namespace: &Namespace{Name: "ns"},
		PodSelectors: []string{"app=demo"}, Manifest: []string{"sts.yaml", "ds.yaml"},
		DeployStatus: types.Deploy_StatusDeployed,
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.NoError(t, err)

	for _, id := range []string{"statefulset-sts", "daemonset-ds"} {
		if n := treeNode(t, got, id); assert.NotNil(t, n, "缺少占位节点: "+id) {
			assert.Equal(t, "progressing", n.Status)
		}
	}
	assert.True(t, hasTreeEdge(got, "owner", "application-1", "statefulset-sts"))
	assert.True(t, hasTreeEdge(got, "owner", "application-1", "daemonset-ds"))
	assert.Equal(t, "progressing", got.Nodes[0].Status)
	assert.Equal(t, types.Deploy_StatusDeploying, got.Status)
}

// TestBuildResourceTree_StatefulSetDegraded 覆盖 sts 容器稳定失败：pod 与 sts 节点逐级降级，
// 整体聚合为 Failed。
func TestBuildResourceTree_StatefulSetDegraded(t *testing.T) {
	replicas := int32(1)
	bad := workloadPod("pod-sts-bad", "StatefulSet", "sts", "sts-uid", "rev1")
	bad.Status.ContainerStatuses[0].State.Waiting = &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff", Message: "back-off"}
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "sts", Namespace: "ns", UID: "sts-uid", Generation: 1},
		Spec:       appsv1.StatefulSetSpec{Replicas: &replicas},
		Status: appsv1.StatefulSetStatus{
			ObservedGeneration: 1, UpdatedReplicas: 1, AvailableReplicas: 1, UpdateRevision: "rev1",
		},
	}
	k := &fakeTreeK8sRepo{
		pods:         []*corev1.Pod{bad},
		manifestSts:  []*appsv1.StatefulSet{{ObjectMeta: metav1.ObjectMeta{Name: "sts"}}},
		statefulSets: map[string]*appsv1.StatefulSet{"sts": sts},
	}
	proj := &Project{
		ID: 1, Name: "demo-app", Namespace: &Namespace{Name: "ns"},
		PodSelectors: []string{"app=demo"}, Manifest: []string{"sts.yaml"},
		DeployStatus: types.Deploy_StatusDeployed,
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.NoError(t, err)

	assert.Equal(t, "degraded", treeNode(t, got, "pod-pod-sts-bad").Status)
	assert.Equal(t, "degraded", treeNode(t, got, "statefulset-sts").Status)
	assert.Equal(t, "degraded", got.Nodes[0].Status)
	assert.Equal(t, types.Deploy_StatusFailed, got.Status)
}

// TestBuildResourceTree_WorkloadReadError 覆盖 GetStatefulSet/GetDaemonSet 读取失败
// （非 NotFound，如 API 抖动）时上抛，不误判为占位节点。
func TestBuildResourceTree_WorkloadReadError(t *testing.T) {
	for _, tc := range []struct {
		name string
		k    *fakeTreeK8sRepo
	}{
		{
			name: "StatefulSet",
			k: &fakeTreeK8sRepo{
				manifestSts:  []*appsv1.StatefulSet{{ObjectMeta: metav1.ObjectMeta{Name: "sts"}}},
				statefulSets: map[string]*appsv1.StatefulSet{},
				getStsErr:    errors.New("api down"),
			},
		},
		{
			name: "DaemonSet",
			k: &fakeTreeK8sRepo{
				manifestDs: []*appsv1.DaemonSet{{ObjectMeta: metav1.ObjectMeta{Name: "ds"}}},
				daemonSets: map[string]*appsv1.DaemonSet{},
				getDsErr:   errors.New("api down"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proj := &Project{
				ID: 1, Namespace: &Namespace{Name: "ns"},
				PodSelectors: []string{"app=demo"}, Manifest: []string{"x.yaml"},
			}
			got, err := buildResourceTree(context.TODO(), tc.k, proj)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, "api down")
		})
	}
}

// TestBuildResourceTree_OrphanWorkloadPod 覆盖属主为 sts 但该 sts 不在 manifest 里的 pod：
// 无对应 workload 节点，降级直挂 Application，controller 打标 + old 标记（hash 判定仍生效）。
func TestBuildResourceTree_OrphanWorkloadPod(t *testing.T) {
	orphan := workloadPod("pod-orphan", "StatefulSet", "orphan", "orphan-uid", "rev1")
	k := &fakeTreeK8sRepo{
		pods:        []*corev1.Pod{orphan},
		manifestSts: []*appsv1.StatefulSet{},
		statefulSets: map[string]*appsv1.StatefulSet{"orphan": {
			ObjectMeta: metav1.ObjectMeta{Name: "orphan", Namespace: "ns", UID: "orphan-uid"},
			Status:     appsv1.StatefulSetStatus{UpdateRevision: "rev2"},
		}},
	}
	proj := &Project{
		ID: 1, Name: "demo-app", Namespace: &Namespace{Name: "ns"},
		PodSelectors: []string{"app=demo"}, Manifest: []string{"only-deploy.yaml"},
		DeployStatus: types.Deploy_StatusDeployed,
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.NoError(t, err)

	assert.Nil(t, treeNode(t, got, "statefulset-orphan"))
	orphanNode := treeNode(t, got, "pod-pod-orphan")
	if assert.NotNil(t, orphanNode) {
		assert.Equal(t, "true", orphanNode.Labels["old"])
		assert.Equal(t, "statefulset/orphan", orphanNode.Labels["controller"])
	}
	assert.True(t, hasTreeEdge(got, "owner", "application-1", "pod-pod-orphan"))
}

// TestBuildResourceTree_Errors 覆盖资源树推导路径上的读取失败（非 NotFound）全部上抛：
// ListReplicaSets / GetDeployment / ListServices。
func TestBuildResourceTree_Errors(t *testing.T) {
	tests := []struct {
		name string
		k    *fakeTreeK8sRepo
		want string
	}{
		{"ListReplicaSets", &fakeTreeK8sRepo{listRSErr: errors.New("rs down")}, "rs down"},
		{
			"GetDeployment",
			&fakeTreeK8sRepo{
				manifestDeps: []*appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "demo-app"}}},
				getDepErr:    errors.New("api down"),
			},
			"api down",
		},
		{"ListServices", &fakeTreeK8sRepo{listSvcErr: errors.New("svc down")}, "svc down"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := &Project{
				ID: 1, Namespace: &Namespace{Name: "ns"},
				PodSelectors: []string{"app=demo"}, Manifest: []string{"demo.yaml"},
			}
			got, err := buildResourceTree(context.TODO(), tt.k, proj)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.want)
		})
	}
}

// TestBuildResourceTree_ServiceAndRSNoMatch 覆盖：非本项目 Deployment 的 RS 不入图、
// 缩容到 0 的项目 RS 仍入图且 healthy、空 selector 与无匹配 pod 的 Service 均被跳过。
func TestBuildResourceTree_ServiceAndRSNoMatch(t *testing.T) {
	rs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-app-8b2f4e7d", UID: "rs-new-uid",
			Annotations:     map[string]string{RevisionAnnotation: "1"},
			OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", UID: "dep-1"}},
		},
	}
	orphanRS := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "other-app-xxxx", UID: "orphan-rs-uid",
			Annotations:     map[string]string{RevisionAnnotation: "1"},
			OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", UID: "other-dep"}},
		},
	}
	dep := readyDeployment("demo-app", "dep-1")
	dep.Status = appsv1.DeploymentStatus{ObservedGeneration: 1} // 无 pod，Deployment 判 deploying
	k := &fakeTreeK8sRepo{
		rss:          []*appsv1.ReplicaSet{rs, orphanRS},
		deployments:  map[string]*appsv1.Deployment{"demo-app": dep},
		manifestDeps: []*appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "demo-app"}}},
		services: []*corev1.Service{
			{ObjectMeta: metav1.ObjectMeta{Name: "no-selector", Namespace: "ns"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "no-match", Namespace: "ns"}, Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "nope"}}},
		},
	}
	proj := &Project{
		ID: 1, Name: "demo-app", Namespace: &Namespace{Name: "ns"},
		PodSelectors: []string{"app=demo"}, Manifest: []string{"demo.yaml"},
		DeployStatus: types.Deploy_StatusDeployed,
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.NoError(t, err)

	// 项目 RS 缩容到 0 仍入图且 healthy；孤儿 RS 不入图
	rsNode := treeNode(t, got, "replicaset-demo-app-8b2f4e7d")
	if assert.NotNil(t, rsNode) {
		assert.Equal(t, "healthy", rsNode.Status)
	}
	assert.Nil(t, treeNode(t, got, "replicaset-other-app-xxxx"))
	// 两个 Service 都被跳过
	assert.Nil(t, treeNode(t, got, "service-no-selector"))
	assert.Nil(t, treeNode(t, got, "service-no-match"))
	// Deployment 无 pod → deploying，聚合为 Deploying
	assert.Equal(t, "progressing", got.Nodes[0].Status)
	assert.Equal(t, types.Deploy_StatusDeploying, got.Status)
}

// TestBuildResourceTree_NoWorkloadsFallback 覆盖 PodSelectors 命中但命名空间无任何
// 工作负载：仅回 Application 根节点，整体状态回退项目记录的部署状态（不误报 unknown）。
func TestBuildResourceTree_NoWorkloadsFallback(t *testing.T) {
	k := &fakeTreeK8sRepo{}
	proj := &Project{
		ID: 1, Name: "demo-app", Namespace: &Namespace{Name: "ns"},
		PodSelectors: []string{"app=demo"}, Manifest: []string{"demo.yaml"},
		DeployStatus: types.Deploy_StatusDeployed,
	}
	got, err := buildResourceTree(context.TODO(), k, proj)
	assert.NoError(t, err)

	assert.Len(t, got.Nodes, 1)
	assert.Equal(t, "healthy", got.Nodes[0].Status)
	assert.Equal(t, types.Deploy_StatusDeployed, got.Status)
}

// Test_aggregate 覆盖整体状态聚合的各级优先级：无子节点→Unknown、degraded→Failed、
// progressing/unknown→Deploying、全 healthy→Deployed。
func Test_aggregate(t *testing.T) {
	with := func(statuses ...string) *ResourceTree {
		tree := &ResourceTree{Nodes: []*ResourceTreeNode{{ID: "application-1"}}}
		for _, s := range statuses {
			tree.Nodes = append(tree.Nodes, &ResourceTreeNode{Status: s})
		}
		return tree
	}
	assert.Equal(t, types.Deploy_StatusUnknown, with().aggregate())
	assert.Equal(t, types.Deploy_StatusDeployed, with("healthy", "healthy").aggregate())
	assert.Equal(t, types.Deploy_StatusFailed, with("healthy", "degraded").aggregate())
	assert.Equal(t, types.Deploy_StatusDeploying, with("healthy", "progressing").aggregate())
	assert.Equal(t, types.Deploy_StatusDeploying, with("healthy", "unknown").aggregate())
}

// Test_podStatus 覆盖单 pod 节点状态的各级判定分支。
func Test_podStatus(t *testing.T) {
	base := func() *corev1.Pod { return readyPod("p", "rs", "rs-uid") }
	tests := []struct {
		name string
		mod  func(p *corev1.Pod)
		want string
	}{
		{"healthy", func(p *corev1.Pod) {}, "healthy"},
		{"deleting", func(p *corev1.Pod) { ts := metav1.Now(); p.DeletionTimestamp = &ts }, "progressing"},
		{"pending", func(p *corev1.Pod) { p.Status.Phase = corev1.PodPending }, "progressing"},
		{"no container statuses", func(p *corev1.Pod) { p.Status.ContainerStatuses = nil }, "progressing"},
		{"waiting fatal", func(p *corev1.Pod) {
			p.Status.ContainerStatuses[0] = corev1.ContainerStatus{Name: "web", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}}}
		}, "degraded"},
		{"waiting transient", func(p *corev1.Pod) {
			p.Status.ContainerStatuses[0] = corev1.ContainerStatus{Name: "web", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ContainerCreating"}}}
		}, "progressing"},
		{"terminated fatal", func(p *corev1.Pod) {
			p.Status.ContainerStatuses[0] = corev1.ContainerStatus{Name: "web", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "Error"}}}
		}, "degraded"},
		{"terminated transient", func(p *corev1.Pod) {
			p.Status.ContainerStatuses[0] = corev1.ContainerStatus{Name: "web", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "Completed"}}}
		}, "progressing"},
		{"not ready", func(p *corev1.Pod) { p.Status.ContainerStatuses[0].Ready = false }, "progressing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := base()
			tt.mod(p)
			assert.Equal(t, tt.want, podStatus(p))
		})
	}
}

// Test_aggregatePodStatus 覆盖 pod 集合聚合：degraded 优先、progressing 次之、全健康才 healthy。
func Test_aggregatePodStatus(t *testing.T) {
	healthy := readyPod("p1", "rs", "r1")
	bad := readyPod("p2", "rs", "r1")
	bad.Status.ContainerStatuses[0] = corev1.ContainerStatus{Name: "web", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}}}
	progressing := readyPod("p3", "rs", "r1")
	progressing.Status.ContainerStatuses[0].Ready = false

	assert.Equal(t, "healthy", aggregatePodStatus([]*corev1.Pod{healthy}))
	assert.Equal(t, "degraded", aggregatePodStatus([]*corev1.Pod{healthy, bad}))
	assert.Equal(t, "progressing", aggregatePodStatus([]*corev1.Pod{healthy, progressing}))
}

// Test_serviceStatus 覆盖 Service 状态：有 ready→healthy、有 degraded→degraded、否则 progressing。
func Test_serviceStatus(t *testing.T) {
	healthy := readyPod("p1", "rs", "r1")
	bad := readyPod("p2", "rs", "r1")
	bad.Status.ContainerStatuses[0] = corev1.ContainerStatus{Name: "web", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}}}
	progressing := readyPod("p3", "rs", "r1")
	progressing.Status.ContainerStatuses[0].Ready = false

	assert.Equal(t, "healthy", serviceStatus([]*corev1.Pod{healthy}))
	assert.Equal(t, "degraded", serviceStatus([]*corev1.Pod{bad}))
	assert.Equal(t, "progressing", serviceStatus([]*corev1.Pod{progressing}))
	// ready 优先：只要有一个 healthy 即 healthy，degraded 不覆盖
	assert.Equal(t, "healthy", serviceStatus([]*corev1.Pod{healthy, bad}))
}

// Test_rsStatus 覆盖 RS 状态：无 pod（缩容到 0）→ healthy，有 pod → 聚合判定。
func Test_rsStatus(t *testing.T) {
	assert.Equal(t, "healthy", rsStatus(nil))
	assert.Equal(t, "healthy", rsStatus([]*corev1.Pod{readyPod("p1", "rs", "r1")}))
	bad := readyPod("p2", "rs", "r1")
	bad.Status.ContainerStatuses[0].Ready = false
	assert.Equal(t, "progressing", rsStatus([]*corev1.Pod{bad}))
}

// Test_selectorMatches 覆盖 Service selector 判定：空 selector 不匹配、键值需完全一致。
func Test_selectorMatches(t *testing.T) {
	labels := map[string]string{"app": "demo", "tier": "web"}
	assert.False(t, selectorMatches(nil, labels))
	assert.False(t, selectorMatches(map[string]string{"app": "nope"}, labels))
	assert.True(t, selectorMatches(map[string]string{"app": "demo", "tier": "web"}, labels))
}

// Test_sortPods 覆盖 pod 排序：创建时间升序，时间相同时按名兜底，保证输出确定性。
func Test_sortPods(t *testing.T) {
	now := time.Now()
	old := readyPod("old", "rs", "r1")
	old.CreationTimestamp = metav1.NewTime(now.Add(-time.Minute))
	mid := readyPod("mid", "rs", "r1")
	mid.CreationTimestamp = metav1.NewTime(now)
	newer := readyPod("newer", "rs", "r1")
	newer.CreationTimestamp = metav1.NewTime(now.Add(time.Minute))
	// b 与 mid 时间相同 → 按名排（b < mid）
	b := readyPod("b", "rs", "r1")
	b.CreationTimestamp = metav1.NewTime(now)

	got := sortPods([]*corev1.Pod{newer, mid, old, b})
	var names []string
	for _, p := range got {
		names = append(names, p.Name)
	}
	assert.Equal(t, []string{"old", "b", "mid", "newer"}, names)
}
