package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsContainerReady(t *testing.T) {
	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "testContainer",
					Ready: true,
				},
			},
		},
	}

	assert.True(t, isContainerReady(pod, "testContainer"))

	pod.Status.ContainerStatuses[0].Ready = false
	assert.False(t, isContainerReady(pod, "testContainer"))

	assert.False(t, isContainerReady(pod, "nonExistingContainer"))
}

func TestSortStatePod_Len(t *testing.T) {
	pods := SortStatePod{
		{Pod: &corev1.Pod{}},
		{Pod: &corev1.Pod{}},
		{Pod: &corev1.Pod{}},
	}
	assert.Equal(t, 3, pods.Len())
}

func TestSortStatePod_Swap(t *testing.T) {
	pods := SortStatePod{
		{Pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "Pod1",
			},
		}},
		{Pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "Pod2",
			},
		}},
	}
	pods.Swap(0, 1)
	assert.Equal(t, "Pod2", pods[0].Pod.Name)
	assert.Equal(t, "Pod1", pods[1].Pod.Name)
}

func TestSortStatePod_Less(t *testing.T) {
	pods := SortStatePod{
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}},
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}, Status: corev1.PodStatus{Phase: corev1.PodPending}}},
	}
	assert.True(t, pods.Less(0, 1))
	pods = SortStatePod{
		{OrderIndex: 2, Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}},
		{OrderIndex: 1, Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}, Status: corev1.PodStatus{Phase: corev1.PodPending}}},
	}
	assert.True(t, pods.Less(0, 1))

	pods = SortStatePod{
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: true},
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: false},
	}
	assert.True(t, pods.Less(1, 0))

	pods = SortStatePod{
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: true, Terminating: true},
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: true},
	}
	assert.True(t, pods.Less(1, 0))

	pods = SortStatePod{
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1", CreationTimestamp: metav1.Time{Time: time.Now()}}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: true},
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2", CreationTimestamp: metav1.Time{Time: time.Now().Add(-1 * time.Hour)}}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: true},
	}
	assert.True(t, pods.Less(1, 0))

	// 同阶段 + 同 IsOld 时，OrderIndex 大的排后。
	pods = SortStatePod{
		{OrderIndex: 2, IsOld: true, Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}},
		{OrderIndex: 1, IsOld: true, Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}},
	}
	assert.True(t, pods.Less(0, 1))
}

// fakePodK8sRepo 是 buildStateContainers 的 K8sRepo 替身，只覆写容器拓扑推导用到的两个原语。
type fakePodK8sRepo struct {
	K8sRepo
	pods             []*corev1.Pod
	listPodsErr      error
	replicaSets      map[string]*appsv1.ReplicaSet
	getReplicaSetErr error
}

func (f *fakePodK8sRepo) ListPodsBySelectors(namespace string, selectors []string) ([]*corev1.Pod, error) {
	return f.pods, f.listPodsErr
}

func (f *fakePodK8sRepo) GetReplicaSet(namespace, name string) (*appsv1.ReplicaSet, error) {
	if f.getReplicaSetErr != nil {
		return nil, f.getReplicaSetErr
	}
	return f.replicaSets[name], nil
}

func TestBuildStateContainers_EmptySelectors(t *testing.T) {
	k := &fakePodK8sRepo{}
	proj := &Project{Namespace: &Namespace{Name: "ns"}}
	got, err := buildStateContainers(context.TODO(), k, proj)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestBuildStateContainers_ListPodsError(t *testing.T) {
	k := &fakePodK8sRepo{listPodsErr: errors.New("list down")}
	proj := &Project{Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=a"}}
	got, err := buildStateContainers(context.TODO(), k, proj)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "list down")
}

// TestBuildStateContainers_HappyPath 覆盖滚动发布新旧副本识别：rs-old 的 revision 注解
// （"1"）小于 rs-new（"2"），故 rs-old 名下 pod 标记 IsOld；同时验证 Failed pod 被过滤、
// IgnoreContainerNames 侧车容器被剔除、Terminating/Pending 标志与容器 Ready 判定。
func TestBuildStateContainers_HappyPath(t *testing.T) {
	oldRS := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "rs-old",
			UID:         "rs-old-uid",
			Annotations: map[string]string{RevisionAnnotation: "1"},
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "Deployment", UID: "dep-1"},
			},
		},
	}
	newRS := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "rs-new",
			UID:         "rs-new-uid",
			Annotations: map[string]string{RevisionAnnotation: "2"},
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "Deployment", UID: "dep-1"},
			},
		},
	}
	podNew := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "pod-new",
			Namespace:   "ns",
			Annotations: map[string]string{PodOrderIndex: "1", IgnoreContainerNames: " sidecar "},
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: "rs-new", UID: "rs-new-uid"},
			},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}, {Name: "sidecar"}}},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "web", Ready: true},
				{Name: "sidecar", Ready: false},
			},
		},
	}
	podOld := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "pod-old",
			Namespace:   "ns",
			Annotations: map[string]string{PodOrderIndex: "2"},
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: "rs-old", UID: "rs-old-uid"},
			},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "web", Ready: true},
			},
		},
	}
	podFailed := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-failed", Namespace: "ns"},
		Status:     corev1.PodStatus{Phase: corev1.PodFailed},
	}
	k := &fakePodK8sRepo{
		pods: []*corev1.Pod{podNew, podOld, podFailed},
		replicaSets: map[string]*appsv1.ReplicaSet{
			"rs-old": oldRS,
			"rs-new": newRS,
		},
	}
	proj := &Project{Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=a"}}
	got, err := buildStateContainers(context.TODO(), k, proj)
	assert.NoError(t, err)
	if assert.Len(t, got, 2) {
		// 新 pod 在前（IsOld=false），旧 pod 在后（IsOld=true），Failed pod 已过滤。
		assert.Equal(t, "pod-new", got[0].Pod)
		assert.Equal(t, "web", got[0].Container)
		assert.False(t, got[0].IsOld)
		assert.True(t, got[0].Ready)
		assert.Equal(t, "pod-old", got[1].Pod)
		assert.True(t, got[1].IsOld)
		assert.True(t, got[1].Ready)
	}
}

func TestBuildStateContainers_GetReplicaSetError(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "ns",
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: "rs", UID: "rs-uid"},
			},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "web", Ready: true},
			},
		},
	}
	k := &fakePodK8sRepo{pods: []*corev1.Pod{pod}, getReplicaSetErr: errors.New("rs down")}
	proj := &Project{Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=a"}}
	got, err := buildStateContainers(context.TODO(), k, proj)
	assert.NoError(t, err)
	// GetReplicaSet 失败时跳过该 pod 的旧副本判定，但 pod 本身仍进入结果（IsOld=false）。
	if assert.Len(t, got, 1) {
		assert.False(t, got[0].IsOld)
	}
}

// TestBuildStateContainers_RSWithoutDeploymentOwner 覆盖 ReplicaSet 无 Deployment owner 时
// 不进 objectMap，pod 不标记 IsOld 的正常路径。
func TestBuildStateContainers_RSWithoutDeploymentOwner(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "ns",
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: "rs", UID: "rs-uid"},
			},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "web", Ready: true},
			},
		},
	}
	k := &fakePodK8sRepo{
		pods: []*corev1.Pod{pod},
		replicaSets: map[string]*appsv1.ReplicaSet{
			"rs": {ObjectMeta: metav1.ObjectMeta{Name: "rs", UID: "rs-uid"}},
		},
	}
	proj := &Project{Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=a"}}
	got, err := buildStateContainers(context.TODO(), k, proj)
	assert.NoError(t, err)
	if assert.Len(t, got, 1) {
		assert.False(t, got[0].IsOld)
		assert.True(t, got[0].Ready)
	}
}

// TestBuildStateContainers_RevisionCompareBothBranches 覆盖新旧副本判定（older 为真/假）的
// 两条分支：list 是 map，迭代顺序随机，故同一场景多次调用后两种顺序都出现，两条分支
// 跨调用累计均被执行；无论顺序如何，旧副本 pod 恒标记 IsOld=true（结果确定）。
func TestBuildStateContainers_RevisionCompareBothBranches(t *testing.T) {
	for i := 0; i < 30; i++ {
		rsOld := &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "rs-old",
				UID:         "rs-old-uid",
				Annotations: map[string]string{RevisionAnnotation: "1"},
				OwnerReferences: []metav1.OwnerReference{
					{Kind: "Deployment", UID: "dep-1"},
				},
			},
		}
		rsNew := &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "rs-new",
				UID:         "rs-new-uid",
				Annotations: map[string]string{RevisionAnnotation: "2"},
				OwnerReferences: []metav1.OwnerReference{
					{Kind: "Deployment", UID: "dep-1"},
				},
			},
		}
		podNew := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-new",
				Namespace: "ns",
				OwnerReferences: []metav1.OwnerReference{
					{Kind: "ReplicaSet", Name: "rs-new", UID: "rs-new-uid"},
				},
			},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
			Status: corev1.PodStatus{
				Phase:             corev1.PodRunning,
				ContainerStatuses: []corev1.ContainerStatus{{Name: "web", Ready: true}},
			},
		}
		podOld := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-old",
				Namespace: "ns",
				OwnerReferences: []metav1.OwnerReference{
					{Kind: "ReplicaSet", Name: "rs-old", UID: "rs-old-uid"},
				},
			},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
			Status: corev1.PodStatus{
				Phase:             corev1.PodRunning,
				ContainerStatuses: []corev1.ContainerStatus{{Name: "web", Ready: true}},
			},
		}
		k := &fakePodK8sRepo{
			pods: []*corev1.Pod{podNew, podOld},
			replicaSets: map[string]*appsv1.ReplicaSet{
				"rs-old": rsOld,
				"rs-new": rsNew,
			},
		}
		got, err := buildStateContainers(context.TODO(), k, &Project{Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=a"}})
		assert.NoError(t, err)
		if assert.Len(t, got, 2) {
			// 排序后新副本 pod 恒在前，旧副本 pod 恒标记 IsOld=true。
			assert.False(t, got[0].IsOld)
			assert.True(t, got[1].IsOld)
		}
	}
}
