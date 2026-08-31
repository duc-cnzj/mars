package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// clusterBoardRepoStub 是 K8sBiz.ClusterBoard 测试用的 K8sRepo 假实现：
// 内嵌接口继承其余方法，只覆盖看板相关的 ClusterBoard/ClusterInfo。
type clusterBoardRepoStub struct {
	K8sRepo
	board    *ClusterBoardData
	info     *ClusterInfo
	boardErr error
}

func (s *clusterBoardRepoStub) ClusterBoard(ctx context.Context, force bool) (*ClusterBoardData, error) {
	return s.board, s.boardErr
}

func (s *clusterBoardRepoStub) ClusterInfo() *ClusterInfo {
	return s.info
}

// TestK8sBiz_ClusterBoard_Success 成功路径：repo 快照 + 总览聚合成看板返回，
// 管理命名空间集合 nil 时排行/Top Pod 为空（无 mars 空间可展示）。
func TestK8sBiz_ClusterBoard_Success(t *testing.T) {
	stub := &clusterBoardRepoStub{
		board: &ClusterBoardData{Nodes: []*BoardNode{
			{Name: "node01"},
		}},
		info: &ClusterInfo{Status: StatusHealth},
	}
	b := NewK8sBiz(stub)

	got, err := b.ClusterBoard(context.TODO(), nil, "")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, StatusHealth, got.Overview.Status)
	assert.Len(t, got.Nodes, 1)
	assert.Empty(t, got.Namespaces)
	assert.Empty(t, got.Pods)
}

// TestK8sBiz_ClusterBoard_RepoError 失败路径：repo 拉取快照失败时整体上抛，不组装看板。
func TestK8sBiz_ClusterBoard_RepoError(t *testing.T) {
	stub := &clusterBoardRepoStub{boardErr: errors.New("snapshot boom")}
	b := NewK8sBiz(stub)

	got, err := b.ClusterBoard(context.TODO(), nil, "")
	assert.Nil(t, got)
	assert.Error(t, err)
}

// TestNodeRole 角色派生：master/control-plane 角色标签归 master，其余归 worker。
func TestNodeRole(t *testing.T) {
	masterLabels := []map[string]string{
		{"node-role.kubernetes.io/master": ""},
		{"node-role.kubernetes.io/control-plane": ""},
	}
	for _, labels := range masterLabels {
		assert.Equal(t, "master", nodeRole(&BoardNode{Labels: labels}))
	}
	assert.Equal(t, "worker", nodeRole(&BoardNode{Labels: map[string]string{"node-role.kubernetes.io/worker": ""}}))
	assert.Equal(t, "worker", nodeRole(&BoardNode{}))
}

// TestNodeStatus 状态派生：不可调度优先标 SchedulingDisabled，再按 Ready 条件定级，
// 无 Ready 条件兜底 NotReady。
func TestNodeStatus(t *testing.T) {
	assert.Equal(t, "SchedulingDisabled", nodeStatus(&BoardNode{Unschedulable: true, ReadyStatus: "True"}))
	assert.Equal(t, "Ready", nodeStatus(&BoardNode{ReadyStatus: "True"}))
	assert.Equal(t, "NotReady", nodeStatus(&BoardNode{ReadyStatus: "False"}))
	// 无 Ready 条件（如刚注册的节点）ReadyStatus 为空，兜底 NotReady。
	assert.Equal(t, "NotReady", nodeStatus(&BoardNode{}))
}

// TestNodeUsage 用量匹配：按节点名命中 NodeMetrics 返回用量，未命中返回 0。
func TestNodeUsage(t *testing.T) {
	metrics := []*BoardNodeMetric{
		{Name: "node01", CpuUsageMilli: 1000, MemUsageBytes: 1073741824},
	}
	cpu, memory := nodeUsage("node01", metrics)
	assert.Equal(t, int64(1000), cpu)
	assert.Equal(t, int64(1073741824), memory)

	cpu, memory = nodeUsage("missing", metrics)
	assert.Zero(t, cpu)
	assert.Zero(t, memory)
}

// TestNodeRequests 请求聚合：只累加落在目标节点上的 Pod 容器 Requests，
// 其他节点 Pod 被忽略；无匹配节点返回 0。
func TestNodeRequests(t *testing.T) {
	pods := []*BoardPod{
		{NodeName: "node01", CpuRequestMilli: 500, MemRequestBytes: 268435456},
		{NodeName: "node02", CpuRequestMilli: 1000, MemRequestBytes: 1073741824},
	}
	cpu, memory := nodeRequests(pods, "node01")
	assert.Equal(t, int64(500), cpu)
	assert.Equal(t, int64(268435456), memory)

	cpu, memory = nodeRequests(pods, "node03")
	assert.Zero(t, cpu)
	assert.Zero(t, memory)
}

// TestPodMetricsUsage 容器用量累加：快照已把容器用量聚合成整 Pod 用量，直接读取。
func TestPodMetricsUsage(t *testing.T) {
	m := &BoardPodMetric{CpuMilli: 500, MemBytes: 134217728}
	cpu, memory := podMetricsUsage(m)
	assert.Equal(t, int64(500), cpu)
	assert.Equal(t, int64(134217728), memory)

	cpu, memory = podMetricsUsage(&BoardPodMetric{})
	assert.Zero(t, cpu)
	assert.Zero(t, memory)
}

// TestBuildBoardNode 节点明细集成：容量/用量/请求/角色/状态一次性装配正确。
func TestBuildBoardNode(t *testing.T) {
	node := &BoardNode{
		Name:             "node01",
		Labels:           map[string]string{"node-role.kubernetes.io/master": ""},
		ReadyStatus:      "True",
		CpuCapacityMilli: 4000,
		MemCapacityBytes: 8589934592,
	}
	metrics := []*BoardNodeMetric{
		{Name: "node01", CpuUsageMilli: 1000, MemUsageBytes: 1073741824},
	}
	pods := []*BoardPod{
		{NodeName: "node01", CpuRequestMilli: 500, MemRequestBytes: 268435456},
	}

	got := buildBoardNode(node, metrics, pods)
	assert.Equal(t, "node01", got.Name)
	assert.Equal(t, "master", got.Role)
	assert.Equal(t, "Ready", got.Status)
	assert.Equal(t, int64(4000), got.CpuCapacity)
	assert.Equal(t, int64(1000), got.CpuUsage)
	assert.Equal(t, int64(500), got.CpuRequest)
	assert.Equal(t, int64(8589934592), got.MemCapacity)
	assert.Equal(t, int64(1073741824), got.MemUsage)
	assert.Equal(t, int64(268435456), got.MemRequest)
}

// TestBuildBoardNamespaces 命名空间聚合：Pod 数与用量按命名空间归属统计互不串扰。
func TestBuildBoardNamespaces(t *testing.T) {
	namespaces := []*BoardNamespace{
		{Name: "ns-a"},
		{Name: "ns-b"},
	}
	pods := []*BoardPod{
		{Namespace: "ns-a"},
		{Namespace: "ns-a"},
		{Namespace: "ns-b"},
	}
	podMetrics := []*BoardPodMetric{
		{Namespace: "ns-a", CpuMilli: 500},
		{Namespace: "ns-b", CpuMilli: 250},
	}

	got := buildBoardNamespaces(namespaces, pods, podMetrics)
	assert.Len(t, got, 2)
	assert.Equal(t, int32(2), got[0].PodCount)
	assert.Equal(t, int64(500), got[0].CpuMilli)
	assert.Equal(t, int32(1), got[1].PodCount)
	assert.Equal(t, int64(250), got[1].CpuMilli)
}

// TestBuildBoardTopPods 排序与截断：默认按 CPU 降序取前 topN，topSort="mem" 按内存降序；
// 空输入返回空切片。
func TestBuildBoardTopPods(t *testing.T) {
	metrics := []*BoardPodMetric{
		{Name: "low", Namespace: "ns-a", CpuMilli: 100, MemBytes: 4 << 30},
		{Name: "high", Namespace: "ns-b", CpuMilli: 2000, MemBytes: 256 << 20},
	}
	// 默认（空 = cpu）按 CPU 降序：high(2 core) 在前
	got := buildBoardTopPods(metrics, 1, "")
	assert.Len(t, got, 1)
	assert.Equal(t, "high", got[0].Pod)
	assert.Equal(t, int64(2000), got[0].CpuMilli)

	// "mem" 按内存降序：low(4Gi) 反超 high(256Mi)
	gotMem := buildBoardTopPods(metrics, 1, "mem")
	assert.Len(t, gotMem, 1)
	assert.Equal(t, "low", gotMem[0].Pod)
	assert.Equal(t, int64(4<<30), gotMem[0].MemoryBytes)

	assert.Empty(t, buildBoardTopPods(nil, 5, ""))
}

// TestBuildClusterBoard 看板组装：快照为 nil 时只保留总览（防御），正常快照全量组装。
func TestBuildClusterBoard(t *testing.T) {
	overview := &ClusterInfo{Status: StatusHealth}
	board := buildClusterBoard(nil, overview, nil, "")
	assert.Equal(t, overview, board.Overview)
	assert.Empty(t, board.Nodes)

	data := &ClusterBoardData{
		Nodes: []*BoardNode{{Name: "n"}},
	}
	board = buildClusterBoard(data, overview, nil, "")
	assert.Equal(t, overview, board.Overview)
	assert.Len(t, board.Nodes, 1)
	assert.Empty(t, board.Namespaces)
	assert.Empty(t, board.Pods)
}

// TestBuildClusterBoard_ManagedFilter 管理空间过滤：命名空间排行与 Top Pod 只保留
// mars 管理集合内的空间及其 Pod；节点表仍用全量 Pod（请求聚合需全量，不被过滤）。
func TestBuildClusterBoard_ManagedFilter(t *testing.T) {
	managed := map[string]bool{"ns-a": true}
	data := &ClusterBoardData{
		Nodes: []*BoardNode{{Name: "n"}},
		Namespaces: []*BoardNamespace{
			{Name: "ns-a"},
			{Name: "kube-system"},
		},
		Pods: []*BoardPod{
			{Namespace: "ns-a", NodeName: "n", CpuRequestMilli: 500},
			{Namespace: "kube-system", NodeName: "n", CpuRequestMilli: 1500},
		},
		PodMetrics: []*BoardPodMetric{
			{Name: "p-a", Namespace: "ns-a", CpuMilli: 500},
			{Name: "p-sys", Namespace: "kube-system", CpuMilli: 2000},
		},
	}
	board := buildClusterBoard(data, &ClusterInfo{}, managed, "")
	assert.Len(t, board.Namespaces, 1)
	assert.Equal(t, "ns-a", board.Namespaces[0].Name)
	assert.Len(t, board.Pods, 1)
	assert.Equal(t, "p-a", board.Pods[0].Pod)
	assert.Equal(t, int64(500), board.Pods[0].CpuMilli)
	// 节点请求聚合仍覆盖全量 Pod（含系统 Pod），节点表不受过滤影响
	cpuReq, _ := nodeRequests(data.Pods, "n")
	assert.Equal(t, int64(2000), cpuReq)
}

// TestFilterNamespaces 命名空间过滤：只保留 mars 管理集合内的空间。
func TestFilterNamespaces(t *testing.T) {
	nss := []*BoardNamespace{
		{Name: "ns-a"},
		{Name: "kube-system"},
	}
	got := filterNamespaces(nss, map[string]bool{"ns-a": true})
	assert.Len(t, got, 1)
	assert.Equal(t, "ns-a", got[0].Name)
	assert.Empty(t, filterNamespaces(nss, nil))
}

// TestFilterPodMetrics Pod 指标过滤：只保留落在 mars 管理命名空间内的指标。
func TestFilterPodMetrics(t *testing.T) {
	ms := []*BoardPodMetric{
		{Name: "p-a", Namespace: "ns-a"},
		{Name: "p-sys", Namespace: "kube-system"},
	}
	got := filterPodMetrics(ms, map[string]bool{"ns-a": true})
	assert.Len(t, got, 1)
	assert.Equal(t, "p-a", got[0].Name)
	assert.Empty(t, filterPodMetrics(ms, nil))
}
