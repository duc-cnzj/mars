package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmetatypes "k8s.io/apimachinery/pkg/types"
)

// ---------- 构造器 ----------

// newDepForTest 构造一个 observed 已收敛、指定 updated/available/desired 的 Deployment。
func newDepForTest(updated, available, desired int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "dep", Namespace: "ns", Generation: 2},
		Spec:       appsv1.DeploymentSpec{Replicas: lo.ToPtr(desired)},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 2, UpdatedReplicas: updated, AvailableReplicas: available,
		},
	}
}

// newStsForTest 构造一个 observed 已收敛、指定 updated/available/desired 与 updateRevision 的 StatefulSet。
func newStsForTest(updated, available, desired int32, updateRevision string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "sts", Namespace: "ns", Generation: 2},
		Spec:       appsv1.StatefulSetSpec{Replicas: lo.ToPtr(desired)},
		Status: appsv1.StatefulSetStatus{
			ObservedGeneration: 2, UpdatedReplicas: updated, AvailableReplicas: available,
			UpdateRevision: updateRevision,
		},
	}
}

// newDsForTest 构造一个 observed 已收敛、指定 updated/available/desired 的 DaemonSet。
func newDsForTest(updated, available, desired int32) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "ds", Namespace: "ns", Generation: 2},
		Status: appsv1.DaemonSetStatus{
			ObservedGeneration: 2, UpdatedNumberScheduled: updated,
			NumberAvailable: available, DesiredNumberScheduled: desired,
		},
	}
}

// newRSForTest 构造被 depUID 拥有的 ReplicaSet，带 revision 注解。
func newRSForTest(uid, revision, depUID string) *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rs-" + uid, Namespace: "ns", UID: kmetatypes.UID(uid),
			Annotations:     map[string]string{deploymentRevisionAnnotation: revision},
			OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", UID: kmetatypes.UID(depUID)}},
		},
	}
}

// rsPodForTest 构造被指定 ReplicaSet UID 拥有、容器处于 waitingReason 等待态的 pod。
func rsPodForTest(name, rsUID, waitingReason string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: "ns",
			OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", UID: kmetatypes.UID(rsUID)}},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
	}
	if waitingReason != "" {
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{{
			Name:  "web",
			State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: waitingReason}},
		}}
	} else {
		pod.Status.Phase = corev1.PodRunning
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{{Name: "web", Ready: true}}
	}
	return pod
}

// ---------- judgeDeploymentRollout ----------

func TestJudgeDeploymentRollout_DesiredZero(t *testing.T) {
	st, reason, fails := judgeDeploymentRollout(newDepForTest(0, 0, 0), nil, nil)
	assert.Equal(t, types.Deploy_StatusDeployed, st)
	assert.Contains(t, reason, "副本数为 0")
	assert.Empty(t, fails)
}

func TestJudgeDeploymentRollout_NotObserved(t *testing.T) {
	dep := newDepForTest(1, 1, 1)
	dep.Status.ObservedGeneration = 1 // < Generation(2)
	st, reason, fails := judgeDeploymentRollout(dep, nil, nil)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "控制器尚未完成对账")
	assert.Empty(t, fails)
}

func TestJudgeDeploymentRollout_NoNewPods(t *testing.T) {
	st, reason, fails := judgeDeploymentRollout(newDepForTest(0, 5, 5), nil, nil)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "新版本 pod 尚未创建")
	assert.Empty(t, fails)
}

func TestJudgeDeploymentRollout_NewPodFailed(t *testing.T) {
	dep := newDepForTest(1, 0, 1)
	dep.UID = "dep"
	rss := []*appsv1.ReplicaSet{newRSForTest("rs-new", "2", "dep")}
	pods := []*corev1.Pod{rsPodForTest("pod-new", "rs-new", "CrashLoopBackOff")}
	st, reason, fails := judgeDeploymentRollout(dep, rss, pods)
	assert.Equal(t, types.Deploy_StatusFailed, st)
	assert.Contains(t, reason, "CrashLoopBackOff")
	if assert.Len(t, fails, 1) {
		assert.Equal(t, "pod-new", fails[0].Pod)
		assert.Equal(t, "web", fails[0].Container)
		assert.Equal(t, "CrashLoopBackOff", fails[0].Reason)
	}
}

// TestJudgeDeploymentRollout_Progress 覆盖用户关切场景：5 个副本滚动中 3 个就绪，
// 不应判 Deployed，而应返回 Deploying。
func TestJudgeDeploymentRollout_Progress(t *testing.T) {
	st, reason, fails := judgeDeploymentRollout(newDepForTest(3, 3, 5), nil, nil)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "3/5")
	assert.Empty(t, fails)
}

func TestJudgeDeploymentRollout_Deployed(t *testing.T) {
	st, reason, fails := judgeDeploymentRollout(newDepForTest(3, 3, 3), nil, nil)
	assert.Equal(t, types.Deploy_StatusDeployed, st)
	assert.Empty(t, reason)
	assert.Empty(t, fails)
}

// TestJudgeDeploymentRollout_NewPodNotReady 覆盖用户上报场景：滚动窗口期新 pod 已创建但
// 未就绪（ContainerCreating），旧 pod 仍在运行撑起 AvailableReplicas——Deployment 计数
// 满足 updated=available=desired 但新 pod 未 Ready，不得误判 Deployed。
func TestJudgeDeploymentRollout_NewPodNotReady(t *testing.T) {
	dep := newDepForTest(1, 1, 1) // UpdatedReplicas=1, AvailableReplicas=1, desired=1
	dep.UID = "dep"
	rss := []*appsv1.ReplicaSet{
		newRSForTest("rs-old", "1", "dep"),
		newRSForTest("rs-new", "2", "dep"),
	}
	newPod := rsPodForTest("pod-new", "rs-new", "")
	newPod.Status.ContainerStatuses[0].Ready = false // 新 pod 容器未就绪
	oldPod := rsPodForTest("pod-old", "rs-old", "")  // 旧 pod 仍 Ready
	st, reason, fails := judgeDeploymentRollout(dep, rss, []*corev1.Pod{newPod, oldPod})
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "0/1")
	assert.Empty(t, fails)
}

// TestJudgeDeploymentRollout_NewPodReady 覆盖滚动窗口收尾：新 pod 已 Ready 才判 Deployed。
func TestJudgeDeploymentRollout_NewPodReady(t *testing.T) {
	dep := newDepForTest(1, 1, 1)
	dep.UID = "dep"
	rss := []*appsv1.ReplicaSet{
		newRSForTest("rs-old", "1", "dep"),
		newRSForTest("rs-new", "2", "dep"),
	}
	newPod := rsPodForTest("pod-new", "rs-new", "") // Ready=true
	oldPod := rsPodForTest("pod-old", "rs-old", "")
	st, reason, fails := judgeDeploymentRollout(dep, rss, []*corev1.Pod{newPod, oldPod})
	assert.Equal(t, types.Deploy_StatusDeployed, st)
	assert.Empty(t, reason)
	assert.Empty(t, fails)
}

// ---------- judgeStatefulSetRollout ----------

func TestJudgeStatefulSetRollout_DesiredZero(t *testing.T) {
	st, reason, fails := judgeStatefulSetRollout(newStsForTest(0, 0, 0, "rev2"), nil)
	assert.Equal(t, types.Deploy_StatusDeployed, st)
	assert.Contains(t, reason, "副本数为 0")
	assert.Empty(t, fails)
}

func TestJudgeStatefulSetRollout_NotObserved(t *testing.T) {
	sts := newStsForTest(1, 1, 1, "rev2")
	sts.Status.ObservedGeneration = 1
	st, reason, fails := judgeStatefulSetRollout(sts, nil)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "控制器尚未完成对账")
	assert.Empty(t, fails)
}

func TestJudgeStatefulSetRollout_NoNewPods(t *testing.T) {
	st, reason, fails := judgeStatefulSetRollout(newStsForTest(0, 3, 3, "rev2"), nil)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "新版本 pod 尚未创建")
	assert.Empty(t, fails)
}

// TestJudgeStatefulSetRollout_NewPodFailed 覆盖最新版本 pod（hash==updateRevision）崩溃 → Failed。
func TestJudgeStatefulSetRollout_NewPodFailed(t *testing.T) {
	sts := newStsForTest(1, 0, 1, "rev2")
	pods := []*corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod-sts", Namespace: "ns",
			Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "rev2"},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
			Name:  "web",
			State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff"}},
		}}},
	}}
	st, reason, fails := judgeStatefulSetRollout(sts, pods)
	assert.Equal(t, types.Deploy_StatusFailed, st)
	assert.Contains(t, reason, "ImagePullBackOff")
	if assert.Len(t, fails, 1) {
		assert.Equal(t, "pod-sts", fails[0].Pod)
	}
}

func TestJudgeStatefulSetRollout_Progress(t *testing.T) {
	st, reason, fails := judgeStatefulSetRollout(newStsForTest(2, 2, 3, "rev2"), nil)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "2/3")
	assert.Empty(t, fails)
}

func TestJudgeStatefulSetRollout_Deployed(t *testing.T) {
	st, reason, fails := judgeStatefulSetRollout(newStsForTest(2, 2, 2, "rev2"), nil)
	assert.Equal(t, types.Deploy_StatusDeployed, st)
	assert.Empty(t, reason)
	assert.Empty(t, fails)
}

// TestJudgeStatefulSetRollout_NewPodNotReady 覆盖 STS 滚动窗口期最新版本 pod 未 Ready：
// 计数满足但新 pod 未就绪，不得误判 Deployed。
func TestJudgeStatefulSetRollout_NewPodNotReady(t *testing.T) {
	sts := newStsForTest(1, 1, 1, "rev2")
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "sts-new", Namespace: "ns", Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "rev2"}},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status:     corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "web", Ready: false}}},
	}
	st, reason, fails := judgeStatefulSetRollout(sts, []*corev1.Pod{newPod})
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "0/1")
	assert.Empty(t, fails)
}

// ---------- judgeDaemonSetRollout ----------

func TestJudgeDaemonSetRollout_DesiredZero(t *testing.T) {
	st, reason, fails := judgeDaemonSetRollout(newDsForTest(0, 0, 0), nil)
	assert.Equal(t, types.Deploy_StatusDeployed, st)
	assert.Contains(t, reason, "无节点需调度")
	assert.Empty(t, fails)
}

func TestJudgeDaemonSetRollout_NotObserved(t *testing.T) {
	ds := newDsForTest(2, 2, 2)
	ds.Status.ObservedGeneration = 1
	st, reason, fails := judgeDaemonSetRollout(ds, nil)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "控制器尚未完成对账")
	assert.Empty(t, fails)
}

func TestJudgeDaemonSetRollout_NoNewPods(t *testing.T) {
	st, reason, fails := judgeDaemonSetRollout(newDsForTest(0, 3, 3), nil)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "新版本 pod 尚未创建")
	assert.Empty(t, fails)
}

// TestJudgeDaemonSetRollout_NewPodFailed 覆盖最新 hash 组 pod 崩溃 → Failed。
func TestJudgeDaemonSetRollout_NewPodFailed(t *testing.T) {
	ds := newDsForTest(1, 0, 1)
	pods := []*corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod-ds", Namespace: "ns",
			Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "hash-new"},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
			Name:  "web",
			State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "OOMKilled"}},
		}}},
	}}
	st, reason, fails := judgeDaemonSetRollout(ds, pods)
	assert.Equal(t, types.Deploy_StatusFailed, st)
	assert.Contains(t, reason, "OOMKilled")
	if assert.Len(t, fails, 1) {
		assert.Equal(t, "pod-ds", fails[0].Pod)
	}
}

func TestJudgeDaemonSetRollout_Progress(t *testing.T) {
	st, reason, fails := judgeDaemonSetRollout(newDsForTest(1, 1, 3), nil)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "1/3")
	assert.Empty(t, fails)
}

func TestJudgeDaemonSetRollout_Deployed(t *testing.T) {
	st, reason, fails := judgeDaemonSetRollout(newDsForTest(3, 3, 3), nil)
	assert.Equal(t, types.Deploy_StatusDeployed, st)
	assert.Empty(t, reason)
	assert.Empty(t, fails)
}

// TestJudgeDaemonSetRollout_NewPodNotReady 覆盖 DS 滚动窗口期最新 hash 组 pod 未 Ready：
// 计数满足但新 pod 未就绪，不得误判 Deployed。
func TestJudgeDaemonSetRollout_NewPodNotReady(t *testing.T) {
	ds := newDsForTest(1, 1, 1)
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "ds-new", Namespace: "ns", CreationTimestamp: metav1.Time{Time: time.Now()}, Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "hash-new"}},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status:     corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "web", Ready: false}}},
	}
	st, reason, fails := judgeDaemonSetRollout(ds, []*corev1.Pod{newPod})
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "0/1")
	assert.Empty(t, fails)
}

// ---------- collectPodFailures ----------

func TestCollectPodFailures_PodFailed(t *testing.T) {
	pods := []*corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-failed"},
		Status:     corev1.PodStatus{Phase: corev1.PodFailed, Reason: "Evicted", Message: "node full"},
	}}
	refs := collectPodFailures(pods)
	if assert.Len(t, refs, 1) {
		assert.Equal(t, "pod-failed", refs[0].Pod)
		assert.Empty(t, refs[0].Container)
		assert.Equal(t, "Evicted", refs[0].Reason)
		assert.Equal(t, "node full", refs[0].Message)
	}
}

func TestCollectPodFailures_PodFailedNoReason(t *testing.T) {
	refs := collectPodFailures([]*corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-failed"},
		Status:     corev1.PodStatus{Phase: corev1.PodFailed},
	}})
	if assert.Len(t, refs, 1) {
		assert.Equal(t, "PodFailed", refs[0].Reason)
	}
}

func TestCollectPodFailures_FatalWaiting(t *testing.T) {
	for _, reason := range []string{"CrashLoopBackOff", "ImagePullBackOff", "CreateContainerConfigError", "InvalidImageName", "RunContainerError"} {
		refs := collectPodFailures([]*corev1.Pod{rsPodForTest("p", "rs", reason)})
		if assert.Len(t, refs, 1, "reason=%s", reason) {
			assert.Equal(t, reason, refs[0].Reason)
		}
	}
}

func TestCollectPodFailures_FatalTerminated(t *testing.T) {
	for _, reason := range []string{"Error", "OOMKilled", "ContainerCannotRun"} {
		pods := []*corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "p"},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
				Name:  "web",
				State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: reason}},
			}}},
		}}
		refs := collectPodFailures(pods)
		if assert.Len(t, refs, 1, "reason=%s", reason) {
			assert.Equal(t, reason, refs[0].Reason)
		}
	}
}

// TestCollectPodFailures_TransientIgnored 覆盖瞬时态不误判：ContainerCreating/ErrImagePull
// 属拉镜像/创建中的过渡状态，不应报失败。
func TestCollectPodFailures_TransientIgnored(t *testing.T) {
	pods := []*corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "p"},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
				{Name: "web", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ContainerCreating"}}},
				{Name: "sidecar", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ErrImagePull"}}},
				{Name: "ok", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "Completed", ExitCode: 0}}},
			}},
		},
	}
	assert.Empty(t, collectPodFailures(pods))
}

// TestPodAllContainersReady 覆盖无容器状态/任一容器未就绪/全部就绪三条分支。
func TestPodAllContainersReady(t *testing.T) {
	assert.False(t, podAllContainersReady(&corev1.Pod{})) // 无容器状态 → 未就绪
	assert.False(t, podAllContainersReady(&corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
		{Name: "web", Ready: false}, {Name: "sidecar", Ready: true},
	}}})) // 任一容器未就绪 → 未就绪
	assert.True(t, podAllContainersReady(&corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
		{Name: "web", Ready: true},
	}}})) // 全部就绪
}

// TestPodReadyCount 覆盖空列表与混合就绪态的计数。
func TestPodReadyCount(t *testing.T) {
	pods := []*corev1.Pod{
		{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "web", Ready: true}}}},
		{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "web", Ready: false}}}},
		{},
	}
	assert.Equal(t, 1, podReadyCount(pods))
	assert.Equal(t, 0, podReadyCount(nil))
}

// ---------- 新版本 pod 定位辅助 ----------

func TestLatestReplicaSet_PicksMaxRevision(t *testing.T) {
	dep := newDepForTest(2, 2, 2)
	dep.UID = "dep-uid"
	rss := []*appsv1.ReplicaSet{
		newRSForTest("rs-1", "1", "dep-uid"),
		newRSForTest("rs-2", "2", "dep-uid"),
	}
	latest := latestReplicaSet(dep, rss)
	assert.NotNil(t, latest)
	assert.Equal(t, kmetatypes.UID("rs-2"), latest.UID)
}

func TestLatestReplicaSet_IgnoresForeign(t *testing.T) {
	dep := newDepForTest(1, 1, 1)
	dep.UID = "dep-uid"
	rss := []*appsv1.ReplicaSet{newRSForTest("rs-other", "99", "other-dep")}
	assert.Nil(t, latestReplicaSet(dep, rss))
}

func TestLatestReplicaSet_Empty(t *testing.T) {
	assert.Nil(t, latestReplicaSet(newDepForTest(0, 0, 0), nil))
}

func TestRevisionOf(t *testing.T) {
	rs := newRSForTest("rs", "7", "dep")
	assert.Equal(t, 7, revisionOf(rs))
	rs.Annotations[deploymentRevisionAnnotation] = "abc"
	assert.Equal(t, 0, revisionOf(rs))
	assert.Equal(t, 0, revisionOf(&appsv1.ReplicaSet{}))
}

func TestDeploymentNewPods(t *testing.T) {
	dep := newDepForTest(2, 2, 2)
	dep.UID = "dep-uid"
	rss := []*appsv1.ReplicaSet{
		newRSForTest("rs-old", "1", "dep-uid"),
		newRSForTest("rs-new", "2", "dep-uid"),
	}
	pods := []*corev1.Pod{
		rsPodForTest("old", "rs-old", ""),
		rsPodForTest("new", "rs-new", ""),
		rsPodForTest("other", "rs-other", ""),
	}
	got := deploymentNewPods(dep, rss, pods)
	assert.Len(t, got, 1)
	assert.Equal(t, "new", got[0].Name)
}

func TestDeploymentNewPods_NoLatestRS(t *testing.T) {
	dep := newDepForTest(1, 1, 1)
	dep.UID = "dep-uid"
	assert.Nil(t, deploymentNewPods(dep, []*appsv1.ReplicaSet{newRSForTest("rs", "1", "other")}, []*corev1.Pod{rsPodForTest("p", "rs", "")}))
}

func TestStatefulSetNewPods(t *testing.T) {
	sts := newStsForTest(2, 2, 2, "rev2")
	pods := []*corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "old", Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "rev1"}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "new", Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "rev2"}}},
	}
	got := statefulSetNewPods(sts, pods)
	assert.Len(t, got, 1)
	assert.Equal(t, "new", got[0].Name)
}

func TestStatefulSetNewPods_EmptyRevision(t *testing.T) {
	sts := newStsForTest(0, 0, 0, "")
	assert.Nil(t, statefulSetNewPods(sts, nil))
}

func TestDaemonSetNewPods_NewestGroup(t *testing.T) {
	now := time.Now()
	pods := []*corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "old", CreationTimestamp: metav1.Time{Time: now.Add(-time.Hour)}, Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "old-hash"}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "new", CreationTimestamp: metav1.Time{Time: now}, Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "new-hash"}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "new2", CreationTimestamp: metav1.Time{Time: now.Add(time.Minute)}, Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "new-hash"}}},
	}
	got := daemonSetNewPods(pods)
	assert.ElementsMatch(t, []string{"new", "new2"}, lo.Map(got, func(p *corev1.Pod, _ int) string { return p.Name }))
}

func TestDaemonSetNewPods_NoHashFallback(t *testing.T) {
	pods := []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "p1"}}, {ObjectMeta: metav1.ObjectMeta{Name: "p2"}}}
	got := daemonSetNewPods(pods)
	assert.Len(t, got, 2)
}

func TestOwnedBy(t *testing.T) {
	obj := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{OwnerReferences: []metav1.OwnerReference{{UID: "a"}, {UID: "b"}}}}
	assert.True(t, ownedBy(obj, "a"))
	assert.True(t, ownedBy(obj, "b"))
	assert.False(t, ownedBy(obj, "c"))
	assert.False(t, ownedBy(&corev1.Pod{}, "a"))
}

func TestPodsOwnedBy(t *testing.T) {
	pods := []*corev1.Pod{
		rsPodForTest("m1", "rs", ""),
		rsPodForTest("m2", "rs", ""),
		rsPodForTest("other", "rs-other", ""),
	}
	got := podsOwnedBy(pods, "rs")
	assert.ElementsMatch(t, []string{"m1", "m2"}, lo.Map(got, func(p *corev1.Pod, _ int) string { return p.Name }))
}

// ---------- formatReason / toDomainFailures ----------

func TestFormatReason(t *testing.T) {
	got := formatReason([]failedContainerRef{
		{Pod: "p1", Container: "web", Reason: "CrashLoopBackOff"},
		{Pod: "p2", Reason: "Evicted"},
	})
	assert.Equal(t, "容器 web 失败: CrashLoopBackOff; pod p2: Evicted", got)
}

func TestToDomainFailures(t *testing.T) {
	got := toDomainFailures([]failedContainerRef{{Pod: "p", Container: "web", Reason: "Error", Message: "m"}}, "Deployment", "dep")
	if assert.Len(t, got, 1) {
		assert.Equal(t, "Deployment", got[0].Kind)
		assert.Equal(t, "dep", got[0].Workload)
		assert.Equal(t, "p", got[0].Pod)
		assert.Equal(t, "web", got[0].Container)
		assert.Equal(t, "Error", got[0].Reason)
		assert.Equal(t, "m", got[0].Message)
	}
}

// ---------- collectWorkloadOldPods（STS/DS 旧副本分类） ----------

// fakeStsK8sRepo 是 collectWorkloadOldPods 的 K8sRepo 替身，只覆写 GetStatefulSet。
type fakeStsK8sRepo struct {
	K8sRepo
	sts *appsv1.StatefulSet
	err error
}

func (f *fakeStsK8sRepo) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	return f.sts, f.err
}

func TestCollectWorkloadOldPods_StatefulSet(t *testing.T) {
	pods := []*corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "sts-new", Namespace: "ns", Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "rev2"}, OwnerReferences: []metav1.OwnerReference{{Kind: "StatefulSet", Name: "sts", UID: "u"}}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "sts-old", Namespace: "ns", Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "rev1"}, OwnerReferences: []metav1.OwnerReference{{Kind: "StatefulSet", Name: "sts", UID: "u"}}}},
	}
	k := &fakeStsK8sRepo{sts: newStsForTest(1, 1, 2, "rev2")}
	got := collectWorkloadOldPods(k, pods)
	if assert.Len(t, got, 1) {
		assert.Contains(t, got, "sts-old")
	}
}

func TestCollectWorkloadOldPods_StatefulSet_ReadFail(t *testing.T) {
	pods := []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "sts-pod", Namespace: "ns", Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "rev1"}, OwnerReferences: []metav1.OwnerReference{{Kind: "StatefulSet", Name: "sts"}}}}}
	got := collectWorkloadOldPods(&fakeStsK8sRepo{err: errors.New("boom")}, pods)
	assert.Empty(t, got)
}

func TestCollectWorkloadOldPods_StatefulSet_NoUpdateRevision(t *testing.T) {
	pods := []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "sts-pod", Namespace: "ns", Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "rev1"}, OwnerReferences: []metav1.OwnerReference{{Kind: "StatefulSet", Name: "sts"}}}}}
	got := collectWorkloadOldPods(&fakeStsK8sRepo{sts: newStsForTest(1, 1, 1, "")}, pods)
	assert.Empty(t, got)
}

func TestCollectWorkloadOldPods_DaemonSet(t *testing.T) {
	now := time.Now()
	pods := []*corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "ds-new", Namespace: "ns", CreationTimestamp: metav1.Time{Time: now}, Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "hash-new"}, OwnerReferences: []metav1.OwnerReference{{Kind: "DaemonSet", Name: "ds", UID: "u"}}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "ds-old", Namespace: "ns", CreationTimestamp: metav1.Time{Time: now.Add(-time.Hour)}, Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "hash-old"}, OwnerReferences: []metav1.OwnerReference{{Kind: "DaemonSet", Name: "ds", UID: "u"}}}},
	}
	got := collectWorkloadOldPods(&fakeStsK8sRepo{}, pods)
	if assert.Len(t, got, 1) {
		assert.Contains(t, got, "ds-old")
	}
}

func TestCollectWorkloadOldPods_NoStsDs(t *testing.T) {
	pods := []*corev1.Pod{rsPodForTest("dep-pod", "rs", "")}
	assert.Empty(t, collectWorkloadOldPods(&fakeStsK8sRepo{}, pods))
}

// ---------- CheckApplyStatus / judgeWorkloads 集成 ----------

// fakeStatusK8sRepo 是 CheckApplyStatus 判定链路的 K8sRepo 替身：覆写 manifest 解析、
// pod 列表与三类工作负载实时读取，并记录 GetPodLogs 收到的 Previous 标志。
type fakeStatusK8sRepo struct {
	K8sRepo
	pods         []*corev1.Pod
	listPodsErr  error
	depWorkloads []*appsv1.Deployment
	stsWorkloads []*appsv1.StatefulSet
	dsWorkloads  []*appsv1.DaemonSet
	deployments  map[string]*appsv1.Deployment
	statefulSets map[string]*appsv1.StatefulSet
	daemonSets   map[string]*appsv1.DaemonSet
	replicaSets  []*appsv1.ReplicaSet
	listRSErr    error
	getDeployErr error
	getStsErr    error
	getDsErr     error
	logs         map[string]string
	logErr       error
}

func (f *fakeStatusK8sRepo) ListPodsBySelectors(namespace string, selectors []string) ([]*corev1.Pod, error) {
	return f.pods, f.listPodsErr
}

func (f *fakeStatusK8sRepo) GetWorkloadsByManifest(manifests []string) ([]*appsv1.Deployment, []*appsv1.StatefulSet, []*appsv1.DaemonSet) {
	return f.depWorkloads, f.stsWorkloads, f.dsWorkloads
}

func (f *fakeStatusK8sRepo) ListReplicaSets(namespace string) ([]*appsv1.ReplicaSet, error) {
	return f.replicaSets, f.listRSErr
}

func (f *fakeStatusK8sRepo) GetReplicaSet(namespace, name string) (*appsv1.ReplicaSet, error) {
	for _, rs := range f.replicaSets {
		if rs.Name == name {
			return rs, nil
		}
	}
	return nil, errs.NotFound("replicaset not found")
}

func (f *fakeStatusK8sRepo) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	if f.getDeployErr != nil {
		return nil, f.getDeployErr
	}
	if dep, ok := f.deployments[name]; ok {
		return dep, nil
	}
	return nil, errs.NotFound("deployment not found")
}

func (f *fakeStatusK8sRepo) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	if f.getStsErr != nil {
		return nil, f.getStsErr
	}
	if sts, ok := f.statefulSets[name]; ok {
		return sts, nil
	}
	return nil, errs.NotFound("statefulset not found")
}

func (f *fakeStatusK8sRepo) GetDaemonSet(namespace, name string) (*appsv1.DaemonSet, error) {
	if f.getDsErr != nil {
		return nil, f.getDsErr
	}
	if ds, ok := f.daemonSets[name]; ok {
		return ds, nil
	}
	return nil, errs.NotFound("daemonset not found")
}

func (f *fakeStatusK8sRepo) GetPodLogs(ctx context.Context, namespace, name string, opts *corev1.PodLogOptions) (string, error) {
	if f.logErr != nil {
		return "", f.logErr
	}
	return f.logs[name], nil
}

// fakeStatusProjectRepo 是 CheckApplyStatus 的 ProjectRepo 替身，只覆写 Show。
type fakeStatusProjectRepo struct {
	ProjectRepo
	project *Project
	err     error
}

func (f *fakeStatusProjectRepo) Show(ctx context.Context, id int) (*Project, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.project, nil
}

func newStatusProjectBiz(repo ProjectRepo, k8s K8sRepo) *projectBiz {
	return &projectBiz{logger: mlog.NewForConfig(nil), projRepo: repo, k8sRepo: k8s}
}

func TestCheckApplyStatus_ShowError(t *testing.T) {
	k := &fakeStatusK8sRepo{}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{err: errors.New("show down")}, k)
	got, err := b.CheckApplyStatus(context.TODO(), 1)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "show down")
}

// TestCheckApplyStatus_NoWorkloads 覆盖无 Deployment/StatefulSet/DaemonSet 时返回 UNKNOWN。
func TestCheckApplyStatus_NoWorkloads(t *testing.T) {
	k := &fakeStatusK8sRepo{}
	proj := &Project{Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=a"}}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{project: proj}, k)
	got, err := b.CheckApplyStatus(context.TODO(), 1)
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusUnknown, got.Status)
	assert.Contains(t, got.Reason, "未发现")
}

// TestCheckApplyStatus_DeploymentNotFound 覆盖 Deployment 尚未创建 → Deploying。
func TestCheckApplyStatus_DeploymentNotFound(t *testing.T) {
	k := &fakeStatusK8sRepo{depWorkloads: []*appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "web"}}}}
	proj := &Project{Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=a"}, Manifest: []string{"deploy"}}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{project: proj}, k)
	got, err := b.CheckApplyStatus(context.TODO(), 1)
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusDeploying, got.Status)
	assert.Contains(t, got.Reason, "尚未创建")
}

func TestCheckApplyStatus_DeploymentError(t *testing.T) {
	k := &fakeStatusK8sRepo{
		depWorkloads: []*appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "web"}}},
		getDeployErr: errors.New("boom"),
	}
	proj := &Project{Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=a"}, Manifest: []string{"deploy"}}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{project: proj}, k)
	got, err := b.CheckApplyStatus(context.TODO(), 1)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "boom")
}

// TestCheckApplyStatus_BuildContainersError 覆盖 buildStateContainers 失败时 CheckApplyStatus 上抛。
func TestCheckApplyStatus_BuildContainersError(t *testing.T) {
	k := &fakeStatusK8sRepo{listPodsErr: errors.New("list down")}
	proj := &Project{Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=a"}, Manifest: []string{"deploy"}}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{project: proj}, k)
	got, err := b.CheckApplyStatus(context.TODO(), 1)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "list down")
}

// TestCheckApplyStatus_NoPodSelectors 覆盖项目无 PodSelectors 时 buildStateContainers
// 提前返回（不触发 k8s 调用），判定仍走工作负载状态。
func TestCheckApplyStatus_NoPodSelectors(t *testing.T) {
	k := &fakeStatusK8sRepo{}
	proj := &Project{Namespace: &Namespace{Name: "ns"}, Manifest: []string{"deploy"}}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{project: proj}, k)
	got, err := b.CheckApplyStatus(context.TODO(), 1)
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusUnknown, got.Status)
}

// TestCheckApplyStatus_FailedChain 覆盖整条失败链路：STS 最新版本 pod CrashLoopBackOff →
// Failed + failures 明细 + fillFailureLogs 拉取日志尾部。
func TestCheckApplyStatus_FailedChain(t *testing.T) {
	sts := newStsForTest(1, 0, 1, "rev2")
	sts.Name = "sts"
	failPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "sts-pod", Namespace: "ns", Labels: map[string]string{appsv1.ControllerRevisionHashLabelKey: "rev2"}},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
			Name:  "web",
			State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}},
		}}},
	}
	k := &fakeStatusK8sRepo{
		pods:         []*corev1.Pod{failPod},
		stsWorkloads: []*appsv1.StatefulSet{sts},
		statefulSets: map[string]*appsv1.StatefulSet{"sts": sts},
		logs:         map[string]string{"sts-pod": "crash trace"},
	}
	proj := &Project{Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=a"}, Manifest: []string{"sts yaml"}}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{project: proj}, k)
	got, err := b.CheckApplyStatus(context.TODO(), 1)
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusFailed, got.Status)
	assert.Contains(t, got.Reason, "CrashLoopBackOff")
	if assert.Len(t, got.Failures, 1) {
		assert.Equal(t, "StatefulSet", got.Failures[0].Kind)
		assert.Equal(t, "sts", got.Failures[0].Workload)
		assert.Equal(t, "sts-pod", got.Failures[0].Pod)
		assert.Equal(t, "web", got.Failures[0].Container)
		assert.Equal(t, "CrashLoopBackOff", got.Failures[0].Reason)
		assert.Equal(t, "crash trace", got.Failures[0].Logs)
	}
}

// TestCheckApplyStatus_TransitionWindowNewPodNotReady 复现用户上报场景：滚动窗口期
// Deployment 计数 updated=available=desired 已满足，但新版本 web pod 未 Ready、旧 pod
// 仍 Ready（撑起 AvailableReplicas），整体必须判 Deploying 而非 Deployed。
func TestCheckApplyStatus_TransitionWindowNewPodNotReady(t *testing.T) {
	dep := newDepForTest(1, 1, 1)
	dep.Name = "web"
	dep.UID = "dep"
	oldRS := newRSForTest("rs-old", "1", "dep")
	newRS := newRSForTest("rs-new", "2", "dep")
	newPod := rsPodForTest("pod-new", "rs-new", "")
	newPod.Status.ContainerStatuses[0].Ready = false
	oldPod := rsPodForTest("pod-old", "rs-old", "")
	k := &fakeStatusK8sRepo{
		depWorkloads: []*appsv1.Deployment{dep},
		deployments:  map[string]*appsv1.Deployment{"web": dep},
		replicaSets:  []*appsv1.ReplicaSet{oldRS, newRS},
		pods:         []*corev1.Pod{newPod, oldPod},
	}
	proj := &Project{Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=web"}, Manifest: []string{"web deploy"}}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{project: proj}, k)
	got, err := b.CheckApplyStatus(context.TODO(), 1)
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusDeploying, got.Status)
	assert.Contains(t, got.Reason, "web")
}

// ---------- judgeWorkloads 聚合 ----------

func TestJudgeWorkloads_Deployed(t *testing.T) {
	dep := newDepForTest(2, 2, 2)
	dep.Name = "web"
	k := &fakeStatusK8sRepo{
		deployments: map[string]*appsv1.Deployment{"web": dep},
		replicaSets: []*appsv1.ReplicaSet{newRSForTest("rs-new", "2", "dep")},
		pods:        []*corev1.Pod{rsPodForTest("pod-new", "rs-new", "")},
	}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	st, reason, failures, err := b.judgeWorkloads(context.TODO(), "ns", []*appsv1.Deployment{dep}, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusDeployed, st)
	assert.Empty(t, reason)
	assert.Empty(t, failures)
}

// TestJudgeWorkloads_FailedPriority 覆盖聚合优先级：一个失败 + 一个进行中 → 整体 Failed。
func TestJudgeWorkloads_FailedPriority(t *testing.T) {
	depFailed := newDepForTest(1, 0, 1)
	depFailed.Name = "web"
	depFailed.UID = "dep"
	depDeploying := newDepForTest(1, 1, 2)
	depDeploying.Name = "api"
	depDeploying.UID = "other-dep"
	k := &fakeStatusK8sRepo{
		deployments: map[string]*appsv1.Deployment{"web": depFailed, "api": depDeploying},
		replicaSets: []*appsv1.ReplicaSet{newRSForTest("rs-fail", "1", "dep")},
		pods:        []*corev1.Pod{rsPodForTest("pod-fail", "rs-fail", "CrashLoopBackOff")},
	}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	st, reason, failures, err := b.judgeWorkloads(context.TODO(), "ns", []*appsv1.Deployment{depFailed, depDeploying}, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusFailed, st)
	assert.Contains(t, reason, "web")
	assert.NotEmpty(t, failures)
}

// TestJudgeWorkloads_DeployingAggregation 覆盖全部进行中且无失败 → 整体 Deploying。
func TestJudgeWorkloads_DeployingAggregation(t *testing.T) {
	dep := newDepForTest(1, 1, 2)
	dep.Name = "web"
	k := &fakeStatusK8sRepo{
		deployments: map[string]*appsv1.Deployment{"web": dep},
		replicaSets: []*appsv1.ReplicaSet{newRSForTest("rs-new", "2", "dep")},
		pods:        []*corev1.Pod{rsPodForTest("pod-new", "rs-new", "")},
	}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	st, reason, failures, err := b.judgeWorkloads(context.TODO(), "ns", []*appsv1.Deployment{dep}, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "滚动进行中")
	assert.Empty(t, failures)
}

func TestJudgeWorkloads_ListReplicaSetsError(t *testing.T) {
	dep := newDepForTest(1, 1, 1)
	dep.Name = "web"
	k := &fakeStatusK8sRepo{depWorkloads: []*appsv1.Deployment{dep}, listRSErr: errors.New("rs down")}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	_, _, _, err := b.judgeWorkloads(context.TODO(), "ns", []*appsv1.Deployment{dep}, nil, nil)
	assert.ErrorContains(t, err, "rs down")
}

func TestJudgeWorkloads_StatefulSetNotFound(t *testing.T) {
	sts := newStsForTest(1, 1, 1, "rev2")
	sts.Name = "sts"
	k := &fakeStatusK8sRepo{stsWorkloads: []*appsv1.StatefulSet{sts}} // statefulSets 空 → 未创建
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	st, reason, failures, err := b.judgeWorkloads(context.TODO(), "ns", nil, []*appsv1.StatefulSet{sts}, nil)
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "尚未创建")
	assert.Empty(t, failures)
}

func TestJudgeWorkloads_StatefulSetError(t *testing.T) {
	sts := newStsForTest(1, 1, 1, "rev2")
	sts.Name = "sts"
	k := &fakeStatusK8sRepo{stsWorkloads: []*appsv1.StatefulSet{sts}, getStsErr: errors.New("boom")}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	_, _, _, err := b.judgeWorkloads(context.TODO(), "ns", nil, []*appsv1.StatefulSet{sts}, nil)
	assert.ErrorContains(t, err, "boom")
}

func TestJudgeWorkloads_DaemonSetNotFound(t *testing.T) {
	ds := newDsForTest(2, 2, 2)
	ds.Name = "ds"
	k := &fakeStatusK8sRepo{dsWorkloads: []*appsv1.DaemonSet{ds}} // daemonSets 空 → 未创建
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	st, reason, failures, err := b.judgeWorkloads(context.TODO(), "ns", nil, nil, []*appsv1.DaemonSet{ds})
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "尚未创建")
	assert.Empty(t, failures)
}

func TestJudgeWorkloads_DaemonSetError(t *testing.T) {
	ds := newDsForTest(2, 2, 2)
	ds.Name = "ds"
	k := &fakeStatusK8sRepo{dsWorkloads: []*appsv1.DaemonSet{ds}, getDsErr: errors.New("boom")}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	_, _, _, err := b.judgeWorkloads(context.TODO(), "ns", nil, nil, []*appsv1.DaemonSet{ds})
	assert.ErrorContains(t, err, "boom")
}

// TestPodsByWorkload_InvalidSelector 覆盖 selector 解析失败返回 nil 的防御分支
// （manifest 中合法对象 selector 必可解析，此为对异常输入的兜底）。
func TestPodsByWorkload_InvalidSelector(t *testing.T) {
	k := &fakeStatusK8sRepo{}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	sel := &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
		{Key: "k", Operator: metav1.LabelSelectorOperator("InvalidOp"), Values: []string{"v"}},
	}}
	assert.Nil(t, b.podsByWorkload("ns", sel))
}

// TestJudgeWorkloads_MixedWorkloads 覆盖三类工作负载混合：Deployment Deployed、
// StatefulSet Deploying、DaemonSet Deployed → 整体 Deploying。
func TestJudgeWorkloads_MixedWorkloads(t *testing.T) {
	dep := newDepForTest(1, 1, 1)
	dep.Name = "web"
	sts := newStsForTest(1, 0, 2, "rev2")
	sts.Name = "sts"
	ds := newDsForTest(2, 2, 2)
	ds.Name = "ds"
	k := &fakeStatusK8sRepo{
		deployments:  map[string]*appsv1.Deployment{"web": dep},
		statefulSets: map[string]*appsv1.StatefulSet{"sts": sts},
		daemonSets:   map[string]*appsv1.DaemonSet{"ds": ds},
		replicaSets:  []*appsv1.ReplicaSet{newRSForTest("rs-new", "2", "dep")},
		pods:         []*corev1.Pod{rsPodForTest("pod-new", "rs-new", "")},
	}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	st, reason, failures, err := b.judgeWorkloads(context.TODO(), "ns", []*appsv1.Deployment{dep}, []*appsv1.StatefulSet{sts}, []*appsv1.DaemonSet{ds})
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusDeploying, st)
	assert.Contains(t, reason, "sts")
	assert.Empty(t, failures)
}

// ---------- fillFailureLogs ----------

func TestFillFailureLogs_CrashLoopPrevious(t *testing.T) {
	k := &fakeStatusK8sRepo{logs: map[string]string{"p": "  log line  \n"}}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	failures := []*ContainerFailure{{Pod: "p", Container: "web", Reason: "CrashLoopBackOff"}}
	b.fillFailureLogs(context.TODO(), "ns", failures)
	assert.Equal(t, "log line", failures[0].Logs)
}

func TestFillFailureLogs_LogErrorSwallowed(t *testing.T) {
	k := &fakeStatusK8sRepo{logErr: errors.New("no logs")}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	failures := []*ContainerFailure{{Pod: "p", Container: "web", Reason: "ImagePullBackOff"}}
	b.fillFailureLogs(context.TODO(), "ns", failures)
	assert.Empty(t, failures[0].Logs)
}

func TestFillFailureLogs_SkipEmptyContainer(t *testing.T) {
	k := &fakeStatusK8sRepo{logs: map[string]string{"p": "x"}}
	b := newStatusProjectBiz(&fakeStatusProjectRepo{}, k)
	failures := []*ContainerFailure{{Pod: "p", Reason: "Evicted"}}
	b.fillFailureLogs(context.TODO(), "ns", failures)
	assert.Empty(t, failures[0].Logs)
}
