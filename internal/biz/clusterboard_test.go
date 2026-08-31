package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// clusterBoardRepoStub 是 K8sBiz.ClusterBoard 测试用的 K8sRepo 假实现：
// 内嵌接口继承其余方法，只覆盖看板相关的 ClusterBoard/ClusterInfo。
type clusterBoardRepoStub struct {
	K8sRepo
	board    *ClusterBoardData
	info     *ClusterInfo
	boardErr error
}

func (s *clusterBoardRepoStub) ClusterBoard(ctx context.Context) (*ClusterBoardData, error) {
	return s.board, s.boardErr
}

func (s *clusterBoardRepoStub) ClusterInfo() *ClusterInfo {
	return s.info
}

// TestK8sBiz_ClusterBoard_Success 成功路径：repo 快照 + 总览聚合成看板返回，
// 管理命名空间集合 nil 时排行/Top Pod 为空（无 mars 空间可展示）。
func TestK8sBiz_ClusterBoard_Success(t *testing.T) {
	stub := &clusterBoardRepoStub{
		board: &ClusterBoardData{Nodes: []corev1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "node01"}},
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
		assert.Equal(t, "master", nodeRole(corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: labels}}))
	}
	assert.Equal(t, "worker", nodeRole(corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/worker": ""}}}))
	assert.Equal(t, "worker", nodeRole(corev1.Node{}))
}

// TestNodeStatus 状态派生：不可调度优先标 SchedulingDisabled，再按 Ready 条件定级，
// 无 Ready 条件兜底 NotReady。
func TestNodeStatus(t *testing.T) {
	assert.Equal(t, "SchedulingDisabled", nodeStatus(corev1.Node{
		Spec: corev1.NodeSpec{Unschedulable: true},
		Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
			{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
		}},
	}))
	assert.Equal(t, "Ready", nodeStatus(corev1.Node{
		Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
			{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
		}},
	}))
	assert.Equal(t, "NotReady", nodeStatus(corev1.Node{
		Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
			{Type: corev1.NodeReady, Status: corev1.ConditionFalse},
		}},
	}))
	// 无 Ready 条件（如刚注册的节点）兜底 NotReady。
	assert.Equal(t, "NotReady", nodeStatus(corev1.Node{Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
		{Type: corev1.NodeMemoryPressure, Status: corev1.ConditionTrue},
	}}}))
}

// TestNodeUsage 用量匹配：按节点名命中 NodeMetrics 返回用量，未命中返回 0。
func TestNodeUsage(t *testing.T) {
	metrics := []v1beta1.NodeMetrics{
		{ObjectMeta: metav1.ObjectMeta{Name: "node01"}, Usage: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		}},
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
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "p1"},
			Spec: corev1.PodSpec{
				NodeName: "node01",
				Containers: []corev1.Container{
					{Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("500m"),
						corev1.ResourceMemory: resource.MustParse("256Mi"),
					}}},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "p2"},
			Spec: corev1.PodSpec{
				NodeName: "node02",
				Containers: []corev1.Container{
					{Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					}}},
				},
			},
		},
	}
	cpu, memory := nodeRequests(pods, "node01")
	assert.Equal(t, int64(500), cpu)
	assert.Equal(t, int64(268435456), memory)

	cpu, memory = nodeRequests(pods, "node03")
	assert.Zero(t, cpu)
	assert.Zero(t, memory)
}

// TestPodMetricsUsage 容器用量累加：Pod 指标用量分散在 Containers 列表，逐容器求和。
func TestPodMetricsUsage(t *testing.T) {
	m := v1beta1.PodMetrics{Containers: []v1beta1.ContainerMetrics{
		{Usage: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("64Mi"),
		}},
		{Usage: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("300m"),
			corev1.ResourceMemory: resource.MustParse("64Mi"),
		}},
	}}
	cpu, memory := podMetricsUsage(m)
	assert.Equal(t, int64(500), cpu)
	assert.Equal(t, int64(134217728), memory)

	cpu, memory = podMetricsUsage(v1beta1.PodMetrics{})
	assert.Zero(t, cpu)
	assert.Zero(t, memory)
}

// TestBuildBoardNode 节点明细集成：容量/用量/请求/角色/状态一次性装配正确。
func TestBuildBoardNode(t *testing.T) {
	node := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node01",
			Labels: map[string]string{"node-role.kubernetes.io/master": ""},
		},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
			Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
		},
	}
	metrics := []v1beta1.NodeMetrics{
		{ObjectMeta: metav1.ObjectMeta{Name: "node01"}, Usage: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		}},
	}
	pods := []corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "p1"},
		Spec: corev1.PodSpec{
			NodeName: "node01",
			Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			}}}},
		},
	}}

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
	namespaces := []corev1.Namespace{
		{ObjectMeta: metav1.ObjectMeta{Name: "ns-a"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "ns-b"}},
	}
	pods := []corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "p2", Namespace: "ns-a"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "p3", Namespace: "ns-b"}},
	}
	podMetrics := []v1beta1.PodMetrics{
		{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")}},
		}},
		{ObjectMeta: metav1.ObjectMeta{Name: "p3", Namespace: "ns-b"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("250m")}},
		}},
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
	metrics := []v1beta1.PodMetrics{
		{ObjectMeta: metav1.ObjectMeta{Name: "low", Namespace: "ns-a"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("4Gi")}},
		}},
		{ObjectMeta: metav1.ObjectMeta{Name: "high", Namespace: "ns-b"}, Containers: []v1beta1.ContainerMetrics{
			{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2"), corev1.ResourceMemory: resource.MustParse("256Mi")}},
		}},
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
		Nodes: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "n"}}},
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
		Nodes: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "n"}}},
		Namespaces: []corev1.Namespace{
			{ObjectMeta: metav1.ObjectMeta{Name: "ns-a"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
		},
		Pods: []corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "p-a", Namespace: "ns-a"}, Spec: corev1.PodSpec{
				NodeName: "n",
				Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("500m"),
				}}}},
			}},
			{ObjectMeta: metav1.ObjectMeta{Name: "p-sys", Namespace: "kube-system"}, Spec: corev1.PodSpec{
				NodeName: "n",
				Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("1500m"),
				}}}},
			}},
		},
		PodMetrics: []v1beta1.PodMetrics{
			{ObjectMeta: metav1.ObjectMeta{Name: "p-a", Namespace: "ns-a"}, Containers: []v1beta1.ContainerMetrics{
				{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")}},
			}},
			{ObjectMeta: metav1.ObjectMeta{Name: "p-sys", Namespace: "kube-system"}, Containers: []v1beta1.ContainerMetrics{
				{Usage: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")}},
			}},
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
	nss := []corev1.Namespace{
		{ObjectMeta: metav1.ObjectMeta{Name: "ns-a"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
	}
	got := filterNamespaces(nss, map[string]bool{"ns-a": true})
	assert.Len(t, got, 1)
	assert.Equal(t, "ns-a", got[0].Name)
	assert.Empty(t, filterNamespaces(nss, nil))
}

// TestFilterPodMetrics Pod 指标过滤：只保留落在 mars 管理命名空间内的指标。
func TestFilterPodMetrics(t *testing.T) {
	ms := []v1beta1.PodMetrics{
		{ObjectMeta: metav1.ObjectMeta{Name: "p-a", Namespace: "ns-a"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "p-sys", Namespace: "kube-system"}},
	}
	got := filterPodMetrics(ms, map[string]bool{"ns-a": true})
	assert.Len(t, got, 1)
	assert.Equal(t, "p-a", got[0].Name)
	assert.Empty(t, filterPodMetrics(ms, nil))
}
