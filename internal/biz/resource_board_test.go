package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// resourceBoardRepoStub 是 K8sBiz.ResourceBoard 测试用的 K8sRepo 假实现：
// 内嵌接口继承其余方法，只覆盖 ResourceSnapshot。
type resourceBoardRepoStub struct {
	K8sRepo
	snapshot *ResourceSnapshotData
	err      error
}

func (s *resourceBoardRepoStub) ResourceSnapshot(ctx context.Context) (*ResourceSnapshotData, error) {
	return s.snapshot, s.err
}

// TestK8sBiz_ResourceBoard_Success 成功路径：快照 + 管理集合 + 项目归属聚合成空间板。
func TestK8sBiz_ResourceBoard_Success(t *testing.T) {
	stub := &resourceBoardRepoStub{snapshot: &ResourceSnapshotData{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a"},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
			}}}},
		}},
	}}
	b := NewK8sBiz(stub)

	got, err := b.ResourceBoard(context.TODO(), []string{"ns-a"}, []*Project{{Name: "proj1", Namespace: &Namespace{Name: "ns-a"}, PodSelectors: []string{"app=a"}}})
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Len(t, got.Namespaces, 1)
	assert.Equal(t, "ns-a", got.Namespaces[0].Name)
	assert.Equal(t, int32(1), got.Namespaces[0].PodCount)
}

// TestK8sBiz_ResourceBoard_RepoError 失败路径：repo 拉取快照失败时整体上抛。
func TestK8sBiz_ResourceBoard_RepoError(t *testing.T) {
	stub := &resourceBoardRepoStub{err: errors.New("snapshot boom")}
	b := NewK8sBiz(stub)

	got, err := b.ResourceBoard(context.TODO(), []string{"ns-a"}, nil)
	assert.Nil(t, got)
	assert.Error(t, err)
}

// TestBuildResourceBoard_NilData 防御：快照为 nil 时返回空板，不 panic。
func TestBuildResourceBoard_NilData(t *testing.T) {
	board := buildResourceBoard(nil, map[string]bool{"ns-a": true}, nil)
	assert.NotNil(t, board)
	assert.Empty(t, board.Namespaces)
}

// TestBuildResourceNamespaces 命名空间聚合：requests 累加 pod spec 容器 Requests、
// 实际用量累加 PodMetrics，只保留管理集合内的空间，按名排序保证确定性输出。
func TestBuildResourceNamespaces(t *testing.T) {
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a"},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
			}}}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "p2", Namespace: "ns-a"},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			}}}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "p3", Namespace: "kube-system"},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1500m")},
			}}}},
		},
	}
	podMetrics := []v1beta1.PodMetrics{
		{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("64Mi"),
			}},
		}},
		{ObjectMeta: metav1.ObjectMeta{Name: "p2", Namespace: "ns-a"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("300m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			}},
		}},
	}

	got := buildResourceNamespaces(map[string]bool{"ns-a": true}, pods, podMetrics)
	assert.Len(t, got, 1)
	ns := got[0]
	assert.Equal(t, "ns-a", ns.Name)
	assert.Equal(t, int32(2), ns.PodCount)
	assert.Equal(t, int64(1500), ns.CpuRequestMilli)
	assert.Equal(t, int64(500), ns.CpuUsageMilli)
	assert.Equal(t, int64(1342177280), ns.MemRequestBytes)
	assert.Equal(t, int64(201326592), ns.MemUsageBytes)
}

// TestBuildResourceNamespaces_MetricOnlyAndUnmanaged 指标环边：只有指标无 Pod 的
// 管理空间也产出记录（PodCount=0）；非管理空间的指标被跳过（不入板）。
func TestBuildResourceNamespaces_MetricOnlyAndUnmanaged(t *testing.T) {
	pods := []corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "ns-a"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
		}}}}},
	}
	podMetrics := []v1beta1.PodMetrics{
		// 管理空间 ns-b：无 Pod 但有指标 → 仍产出记录（PodCount=0）。
		{ObjectMeta: metav1.ObjectMeta{Name: "m-b", Namespace: "ns-b"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")}},
		}},
		// 非管理空间 kube-system：指标被跳过。
		{ObjectMeta: metav1.ObjectMeta{Name: "m-sys", Namespace: "kube-system"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1500m")}},
		}},
	}

	got := buildResourceNamespaces(map[string]bool{"ns-a": true, "ns-b": true}, pods, podMetrics)
	if assert.Len(t, got, 2) {
		assert.Equal(t, "ns-a", got[0].Name)
		assert.Equal(t, int32(1), got[0].PodCount)
		assert.Equal(t, int64(500), got[0].CpuRequestMilli)
		assert.Equal(t, "ns-b", got[1].Name)
		assert.Zero(t, got[1].PodCount)
		assert.Equal(t, int64(100), got[1].CpuUsageMilli)
	}
}

// TestBuildResourceNamespaces_Sorted 排序：多个管理空间按名升序输出。
func TestBuildResourceNamespaces_Sorted(t *testing.T) {
	pods := []corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "ns-b"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "ns-a"}},
	}
	got := buildResourceNamespaces(map[string]bool{"ns-a": true, "ns-b": true}, pods, nil)
	assert.Len(t, got, 2)
	assert.Equal(t, "ns-a", got[0].Name)
	assert.Equal(t, "ns-b", got[1].Name)
}

// TestAttachResourceProjects 项目拆分：按 PodSelectors 匹配同空间 Pod，requests/用量
// 各自累加；一个 pod 命中多个项目时每个项目都计入（selectors 重叠场景）。
func TestAttachResourceProjects(t *testing.T) {
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a", Labels: map[string]string{"app": "a"}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
			}}}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "p2", Namespace: "ns-a", Labels: map[string]string{"app": "b"}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
			}}}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "p3", Namespace: "ns-a", Labels: map[string]string{"app": "c"}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
			}}}},
		},
	}
	podMetrics := []v1beta1.PodMetrics{
		{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("200m")}},
		}},
		{ObjectMeta: metav1.ObjectMeta{Name: "p2", Namespace: "ns-a"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("300m")}},
		}},
	}
	projects := []*Project{
		{Name: "proj-b", Namespace: &Namespace{Name: "ns-a"}, PodSelectors: []string{"app=b"}},
		{Name: "proj-a", Namespace: &Namespace{Name: "ns-a"}, PodSelectors: []string{"app=a"}},
		// 项目归属其他空间：不匹配任何 pod，也不入本空间板。
		{Name: "proj-other", Namespace: &Namespace{Name: "ns-other"}, PodSelectors: []string{"app=a"}},
		// 非法 selector：解析失败跳过，不 panic。
		{Name: "proj-bad", Namespace: &Namespace{Name: "ns-a"}, PodSelectors: []string{"==="}},
	}
	namespaces := []*ResourceNamespace{{Name: "ns-a"}}

	attachResourceProjects(namespaces, projects, pods, nil, podMetrics)

	assert.Len(t, namespaces[0].Projects, 2) // proj-other 不入板，proj-bad 无命中
	// 项目按名排序：proj-a 在前。
	assert.Equal(t, "proj-a", namespaces[0].Projects[0].Name)
	assert.Equal(t, int32(1), namespaces[0].Projects[0].PodCount)
	assert.Equal(t, int64(500), namespaces[0].Projects[0].CpuRequestMilli)
	assert.Equal(t, int64(200), namespaces[0].Projects[0].CpuUsageMilli)
	assert.Equal(t, "proj-b", namespaces[0].Projects[1].Name)
	assert.Equal(t, int64(1000), namespaces[0].Projects[1].CpuRequestMilli)
	assert.Equal(t, int64(300), namespaces[0].Projects[1].CpuUsageMilli)
}

// TestAttachResourceProjects_MultiMatch 重叠 selector：同一 pod 命中两个项目时各自计数。
func TestAttachResourceProjects_MultiMatch(t *testing.T) {
	pods := []corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a", Labels: map[string]string{"app": "a", "tier": "web"}},
	}}
	projects := []*Project{
		{Name: "proj-x", Namespace: &Namespace{Name: "ns-a"}, PodSelectors: []string{"app=a"}},
		{Name: "proj-y", Namespace: &Namespace{Name: "ns-a"}, PodSelectors: []string{"tier=web"}},
	}
	namespaces := []*ResourceNamespace{{Name: "ns-a"}}

	attachResourceProjects(namespaces, projects, pods, nil, nil)

	assert.Len(t, namespaces[0].Projects, 2)
	assert.Equal(t, int32(1), namespaces[0].Projects[0].PodCount)
	assert.Equal(t, int32(1), namespaces[0].Projects[1].PodCount)
}

// TestAttachResourceProjects_NilAndCrossNamespace 防御：nil 项目/无命名空间边的项目
// 直接跳过；与本空间项目命名空间不同的 pod 不计入（namespace 不匹配 continue）。
func TestAttachResourceProjects_NilAndCrossNamespace(t *testing.T) {
	pods := []corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "p-a", Namespace: "ns-a", Labels: map[string]string{"app": "a"}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "p-b", Namespace: "ns-b", Labels: map[string]string{"app": "a"}}},
	}
	projects := []*Project{
		nil, // 空指针跳过
		{Name: "proj-no-ns", PodSelectors: []string{"app=a"}},                                  // 无命名空间边跳过
		{Name: "proj-a", Namespace: &Namespace{Name: "ns-a"}, PodSelectors: []string{"app=a"}}, // 只匹配 ns-a 内的 pod
	}
	namespaces := []*ResourceNamespace{{Name: "ns-a"}}

	attachResourceProjects(namespaces, projects, pods, nil, nil)

	if assert.Len(t, namespaces[0].Projects, 1) {
		assert.Equal(t, "proj-a", namespaces[0].Projects[0].Name)
		assert.Equal(t, int32(1), namespaces[0].Projects[0].PodCount) // ns-b 的 p-b 不计入
	}
}

// TestParsedPodSelectors 解析：合法 selector 保留，坏语法（如 "==="）跳过。
// 注意 labels.Parse("") 返回合法的空 selector（匹配所有），不属于非法输入。
func TestParsedPodSelectors(t *testing.T) {
	selectors := parsedPodSelectors([]string{"app=a", "tier in (web,db)", "==="})
	assert.Len(t, selectors, 2)
	assert.Empty(t, parsedPodSelectors(nil))
}

// TestMatchAnySelector 匹配：selector 命中/未命中/空列表恒不命中。
func TestMatchAnySelector(t *testing.T) {
	selectors := parsedPodSelectors([]string{"app=a"})
	assert.True(t, matchAnySelector(selectors, map[string]string{"app": "a"}))
	assert.False(t, matchAnySelector(selectors, map[string]string{"app": "b"}))
	assert.False(t, matchAnySelector(nil, map[string]string{"app": "a"}))
}

// TestAttachResourceProjects_DeploymentChain 项目内 Deployment 属主链分组：
// pod → RS → Deployment 逐段解析，requests/usage/PodCount 按工作负载正确累加。
func TestAttachResourceProjects_DeploymentChain(t *testing.T) {
	depUID := types.UID("deploy-1")
	rs1UID := types.UID("rs-1")
	rs2UID := types.UID("rs-2")
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "web-x", Namespace: "ns-a", Labels: map[string]string{"app": "a"},
				OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: "web-x-rs", UID: rs1UID}}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("400m"),
					corev1.ResourceMemory: resource.MustParse("512Mi"),
				}}}}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "web-y", Namespace: "ns-a", Labels: map[string]string{"app": "a"},
				OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: "web-y-rs", UID: rs2UID}}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("300m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				}}}}},
		},
	}
	replicaSets := []appsv1.ReplicaSet{
		{ObjectMeta: metav1.ObjectMeta{Name: "web-x-rs", UID: rs1UID, OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", Name: "web-api", UID: depUID}}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "web-y-rs", UID: rs2UID, OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", Name: "worker", UID: depUID}}}},
	}
	podMetrics := []v1beta1.PodMetrics{
		{ObjectMeta: metav1.ObjectMeta{Name: "web-x", Namespace: "ns-a"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")}},
		}},
		{ObjectMeta: metav1.ObjectMeta{Name: "web-y", Namespace: "ns-a"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("50m")}},
		}},
	}
	projects := []*Project{{Name: "proj-a", Namespace: &Namespace{Name: "ns-a"}, PodSelectors: []string{"app=a"}}}
	namespaces := []*ResourceNamespace{{Name: "ns-a"}}

	attachResourceProjects(namespaces, projects, pods, replicaSets, podMetrics)

	if assert.Len(t, namespaces[0].Projects, 1) {
		p := namespaces[0].Projects[0]
		assert.Equal(t, int32(2), p.PodCount)
		assert.Equal(t, int64(700), p.CpuRequestMilli, "项目总量 = 两个 Deployment 之和")
		assert.Equal(t, int64(150), p.CpuUsageMilli)
		if assert.Len(t, p.Workloads, 2) {
			// 按 kind 再 name 排序确定性：web-api 在 worker 前
			assert.Equal(t, "Deployment", p.Workloads[0].Kind)
			assert.Equal(t, "web-api", p.Workloads[0].Name)
			assert.Equal(t, int32(1), p.Workloads[0].PodCount)
			assert.Equal(t, int64(400), p.Workloads[0].CpuRequestMilli)
			assert.Equal(t, int64(100), p.Workloads[0].CpuUsageMilli)
			assert.Equal(t, int64(536870912), p.Workloads[0].MemRequestBytes)
			assert.Equal(t, "worker", p.Workloads[1].Name)
			assert.Equal(t, int64(300), p.Workloads[1].CpuRequestMilli)
			assert.Equal(t, int64(50), p.Workloads[1].CpuUsageMilli)
		}
	}
}

// TestAttachResourceProjects_StatefulSetAndDaemonSet 直接属主分组：pod 属主为
// StatefulSet/DaemonSet 时不经 RS，直接归入对应工作负载。
func TestAttachResourceProjects_StatefulSetAndDaemonSet(t *testing.T) {
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "sts-0", Namespace: "ns-a", Labels: map[string]string{"app": "a"},
				OwnerReferences: []metav1.OwnerReference{{Kind: "StatefulSet", Name: "db"}}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("200m")}}}}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "ds-x", Namespace: "ns-a", Labels: map[string]string{"app": "a"},
				OwnerReferences: []metav1.OwnerReference{{Kind: "DaemonSet", Name: "agent"}}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")}}}}},
		},
	}
	projects := []*Project{{Name: "proj-a", Namespace: &Namespace{Name: "ns-a"}, PodSelectors: []string{"app=a"}}}
	namespaces := []*ResourceNamespace{{Name: "ns-a"}}

	attachResourceProjects(namespaces, projects, pods, nil, nil)

	if assert.Len(t, namespaces[0].Projects, 1) {
		workloads := namespaces[0].Projects[0].Workloads
		if assert.Len(t, workloads, 2) {
			// kind 排序：DaemonSet < StatefulSet
			assert.Equal(t, "DaemonSet", workloads[0].Kind)
			assert.Equal(t, "agent", workloads[0].Name)
			assert.Equal(t, "StatefulSet", workloads[1].Kind)
			assert.Equal(t, "db", workloads[1].Name)
			assert.Equal(t, int64(200), workloads[1].CpuRequestMilli)
		}
	}
}

// TestAttachResourceProjects_BarePodAndMissingRS 裸 pod 与 RS 缺失兜底：
// 无 workload 属主 / 属主 RS 不在快照内的 pod 都计入项目总量但不单列工作负载。
func TestAttachResourceProjects_BarePodAndMissingRS(t *testing.T) {
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "bare", Namespace: "ns-a", Labels: map[string]string{"app": "a"}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("300m")}}}}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "orphan", Namespace: "ns-a", Labels: map[string]string{"app": "a"},
				OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: "ghost-rs", UID: types.UID("ghost")}}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")}}}}},
		},
	}
	projects := []*Project{{Name: "proj-a", Namespace: &Namespace{Name: "ns-a"}, PodSelectors: []string{"app=a"}}}
	namespaces := []*ResourceNamespace{{Name: "ns-a"}}

	// 快照内无该 RS → orphan 兜底为裸 pod
	attachResourceProjects(namespaces, projects, pods, nil, nil)

	if assert.Len(t, namespaces[0].Projects, 1) {
		p := namespaces[0].Projects[0]
		assert.Equal(t, int32(2), p.PodCount, "裸 pod 计入项目总量")
		assert.Equal(t, int64(800), p.CpuRequestMilli, "总量含裸 pod 的 requests")
		assert.Empty(t, p.Workloads, "无属主 workload 的 pod 不单列")
	}
}

// TestWorkloadOf 属主链解析：Deployment 经 RS 属主链、STS/DS 直接属主、
// Job/裸 pod/RS 缺失均返回空键。
func TestWorkloadOf(t *testing.T) {
	rsByUID := rsByUIDIndex([]appsv1.ReplicaSet{
		{ObjectMeta: metav1.ObjectMeta{Name: "r1", UID: "u1", OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", Name: "web"}}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "r2", UID: "u2"}}, // 无 Deployment 属主
	})
	cases := []struct {
		name string
		pod  corev1.Pod
		kind string
		wn   string
	}{
		{"deployment chain", corev1.Pod{ObjectMeta: metav1.ObjectMeta{OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: "r1-x", UID: "u1"}}}}, "Deployment", "web"},
		{"statefulset direct", corev1.Pod{ObjectMeta: metav1.ObjectMeta{OwnerReferences: []metav1.OwnerReference{{Kind: "StatefulSet", Name: "db"}}}}, "StatefulSet", "db"},
		{"daemonset direct", corev1.Pod{ObjectMeta: metav1.ObjectMeta{OwnerReferences: []metav1.OwnerReference{{Kind: "DaemonSet", Name: "agent"}}}}, "DaemonSet", "agent"},
		{"rs missing in index", corev1.Pod{ObjectMeta: metav1.ObjectMeta{OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: "ghost", UID: "nope"}}}}, "", ""},
		{"rs without deployment owner", corev1.Pod{ObjectMeta: metav1.ObjectMeta{OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: "r2-x", UID: "u2"}}}}, "", ""},
		{"job bare", corev1.Pod{ObjectMeta: metav1.ObjectMeta{OwnerReferences: []metav1.OwnerReference{{Kind: "Job", Name: "j1"}}}}, "", ""},
		{"no owner", corev1.Pod{}, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kind, name := workloadOf(&tc.pod, rsByUID)
			assert.Equal(t, tc.kind, kind)
			assert.Equal(t, tc.wn, name)
		})
	}
}

// TestRsByUIDIndex RS 按 UID 建索引：命中/未命中，索引指向原切片元素。
func TestRsByUIDIndex(t *testing.T) {
	rss := []appsv1.ReplicaSet{
		{ObjectMeta: metav1.ObjectMeta{Name: "a", UID: "u-a"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "b", UID: "u-b"}},
	}
	idx := rsByUIDIndex(rss)
	assert.Same(t, &rss[0], idx["u-a"])
	assert.Same(t, &rss[1], idx["u-b"])
	assert.Nil(t, idx["missing"])
	assert.Empty(t, rsByUIDIndex(nil))
}
