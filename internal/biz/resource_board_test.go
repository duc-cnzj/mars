package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resourceBoardRepoStub 是 K8sBiz.ResourceBoard 测试用的 K8sRepo 假实现：
// 内嵌接口继承其余方法，只覆盖 ResourceSnapshot。
type resourceBoardRepoStub struct {
	K8sRepo
	snapshot *ResourceSnapshotData
	err      error
}

func (s *resourceBoardRepoStub) ResourceSnapshot(ctx context.Context, force bool) (*ResourceSnapshotData, error) {
	return s.snapshot, s.err
}

// TestK8sBiz_ResourceBoard_Success 成功路径：快照 + 管理集合 + 项目归属聚合成空间板。
func TestK8sBiz_ResourceBoard_Success(t *testing.T) {
	stub := &resourceBoardRepoStub{snapshot: &ResourceSnapshotData{
		Pods: []*ResourcePod{
			{Name: "p1", Namespace: "ns-a", CpuRequestMilli: 500},
		},
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

// TestBuildResourceNamespaces 命名空间聚合：requests 累加快照 Pod 聚合值、
// 实际用量累加 PodMetrics 聚合值，只保留管理集合内的空间，按名排序保证确定性输出。
func TestBuildResourceNamespaces(t *testing.T) {
	pods := []*ResourcePod{
		{Name: "p1", Namespace: "ns-a", CpuRequestMilli: 500, MemRequestBytes: 268435456},
		{Name: "p2", Namespace: "ns-a", CpuRequestMilli: 1000, MemRequestBytes: 1073741824},
		{Name: "p3", Namespace: "kube-system", CpuRequestMilli: 1500},
	}
	podMetrics := []*ResourcePodMetric{
		{Name: "p1", Namespace: "ns-a", CpuMilli: 200, MemBytes: 67108864},
		{Name: "p2", Namespace: "ns-a", CpuMilli: 300, MemBytes: 134217728},
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
	pods := []*ResourcePod{
		{Name: "p", Namespace: "ns-a", CpuRequestMilli: 500},
	}
	podMetrics := []*ResourcePodMetric{
		// 管理空间 ns-b：无 Pod 但有指标 → 仍产出记录（PodCount=0）。
		{Name: "m-b", Namespace: "ns-b", CpuMilli: 100},
		// 非管理空间 kube-system：指标被跳过。
		{Name: "m-sys", Namespace: "kube-system", CpuMilli: 1500},
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
	pods := []*ResourcePod{
		{Namespace: "ns-b"},
		{Namespace: "ns-a"},
	}
	got := buildResourceNamespaces(map[string]bool{"ns-a": true, "ns-b": true}, pods, nil)
	assert.Len(t, got, 2)
	assert.Equal(t, "ns-a", got[0].Name)
	assert.Equal(t, "ns-b", got[1].Name)
}

// TestAttachResourceProjects 项目拆分：按 PodSelectors 匹配同空间 Pod，requests/用量
// 各自累加；一个 pod 命中多个项目时每个项目都计入（selectors 重叠场景）。
func TestAttachResourceProjects(t *testing.T) {
	pods := []*ResourcePod{
		{Name: "p1", Namespace: "ns-a", Labels: map[string]string{"app": "a"}, CpuRequestMilli: 500, MemRequestBytes: 268435456},
		{Name: "p2", Namespace: "ns-a", Labels: map[string]string{"app": "b"}, CpuRequestMilli: 1000},
		{Name: "p3", Namespace: "ns-a", Labels: map[string]string{"app": "c"}, CpuRequestMilli: 2000},
	}
	podMetrics := []*ResourcePodMetric{
		{Name: "p1", Namespace: "ns-a", CpuMilli: 200},
		{Name: "p2", Namespace: "ns-a", CpuMilli: 300},
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
	pods := []*ResourcePod{{
		Name:      "p1",
		Namespace: "ns-a",
		Labels:    map[string]string{"app": "a", "tier": "web"},
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
	pods := []*ResourcePod{
		{Name: "p-a", Namespace: "ns-a", Labels: map[string]string{"app": "a"}},
		{Name: "p-b", Namespace: "ns-b", Labels: map[string]string{"app": "a"}},
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
	pods := []*ResourcePod{
		{
			Name: "web-x", Namespace: "ns-a", Labels: map[string]string{"app": "a"},
			Owners:          []*ResourceOwner{{Kind: "ReplicaSet", Name: "web-x-rs", UID: "rs-1"}},
			CpuRequestMilli: 400, MemRequestBytes: 536870912,
		},
		{
			Name: "web-y", Namespace: "ns-a", Labels: map[string]string{"app": "a"},
			Owners:          []*ResourceOwner{{Kind: "ReplicaSet", Name: "web-y-rs", UID: "rs-2"}},
			CpuRequestMilli: 300, MemRequestBytes: 268435456,
		},
	}
	replicaSets := []*ResourceReplicaSet{
		{UID: "rs-1", Owners: []*ResourceOwner{{Kind: "Deployment", Name: "web-api"}}},
		{UID: "rs-2", Owners: []*ResourceOwner{{Kind: "Deployment", Name: "worker"}}},
	}
	podMetrics := []*ResourcePodMetric{
		{Name: "web-x", Namespace: "ns-a", CpuMilli: 100},
		{Name: "web-y", Namespace: "ns-a", CpuMilli: 50},
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
	pods := []*ResourcePod{
		{
			Name: "sts-0", Namespace: "ns-a", Labels: map[string]string{"app": "a"},
			Owners:          []*ResourceOwner{{Kind: "StatefulSet", Name: "db"}},
			CpuRequestMilli: 200,
		},
		{
			Name: "ds-x", Namespace: "ns-a", Labels: map[string]string{"app": "a"},
			Owners:          []*ResourceOwner{{Kind: "DaemonSet", Name: "agent"}},
			CpuRequestMilli: 100,
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
	pods := []*ResourcePod{
		{
			Name: "bare", Namespace: "ns-a", Labels: map[string]string{"app": "a"},
			CpuRequestMilli: 300,
		},
		{
			Name: "orphan", Namespace: "ns-a", Labels: map[string]string{"app": "a"},
			Owners:          []*ResourceOwner{{Kind: "ReplicaSet", Name: "ghost-rs", UID: "ghost"}},
			CpuRequestMilli: 500,
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
	rsByUID := rsByUIDIndex([]*ResourceReplicaSet{
		{UID: "u1", Owners: []*ResourceOwner{{Kind: "Deployment", Name: "web"}}},
		{UID: "u2"}, // 无 Deployment 属主
	})
	cases := []struct {
		name string
		pod  *ResourcePod
		kind string
		wn   string
	}{
		{"deployment chain", &ResourcePod{Owners: []*ResourceOwner{{Kind: "ReplicaSet", Name: "r1-x", UID: "u1"}}}, "Deployment", "web"},
		{"statefulset direct", &ResourcePod{Owners: []*ResourceOwner{{Kind: "StatefulSet", Name: "db"}}}, "StatefulSet", "db"},
		{"daemonset direct", &ResourcePod{Owners: []*ResourceOwner{{Kind: "DaemonSet", Name: "agent"}}}, "DaemonSet", "agent"},
		{"rs missing in index", &ResourcePod{Owners: []*ResourceOwner{{Kind: "ReplicaSet", Name: "ghost", UID: "nope"}}}, "", ""},
		{"rs without deployment owner", &ResourcePod{Owners: []*ResourceOwner{{Kind: "ReplicaSet", Name: "r2-x", UID: "u2"}}}, "", ""},
		{"job bare", &ResourcePod{Owners: []*ResourceOwner{{Kind: "Job", Name: "j1"}}}, "", ""},
		{"no owner", &ResourcePod{}, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kind, name := workloadOf(tc.pod, rsByUID)
			assert.Equal(t, tc.kind, kind)
			assert.Equal(t, tc.wn, name)
		})
	}
}

// TestRsByUIDIndex RS 按 UID 建索引：命中/未命中，索引指向原切片元素。
func TestRsByUIDIndex(t *testing.T) {
	rss := []*ResourceReplicaSet{
		{UID: "u-a"},
		{UID: "u-b"},
	}
	idx := rsByUIDIndex(rss)
	assert.Same(t, rss[0], idx["u-a"])
	assert.Same(t, rss[1], idx["u-b"])
	assert.Nil(t, idx["missing"])
	assert.Empty(t, rsByUIDIndex(nil))
}
