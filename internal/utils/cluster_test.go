package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	fake2 "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

func TestClusterInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	cpu := &resource.Quantity{}
	cpu.Add(resource.MustParse("3"))
	memory := &resource.Quantity{}
	memory.Add(resource.MustParse("10Gi"))
	fc := fake.NewSimpleClientset(
		&v1.PodList{
			TypeMeta: metav1.TypeMeta{},
			ListMeta: metav1.ListMeta{},
			Items: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
					Spec: v1.PodSpec{
						NodeName: "node01",
						Containers: []v1.Container{
							{
								Name:  "app",
								Image: "xxx:v1",
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										// 3core cpu request
										v1.ResourceCPU: *resource.NewMilliQuantity(3000, resource.DecimalSI),
										// 2G memory request
										v1.ResourceMemory: *resource.NewQuantity(2*(1024*1024*1024), resource.DecimalSI),
									},
								},
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
				// FIXME: fake 客户端不能做 fieldSelector 过滤
			},
		},
		&v1.NodeList{
			Items: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node01"},
					Status: v1.NodeStatus{
						Capacity: v1.ResourceList{
							v1.ResourceCPU:    cpu.DeepCopy(),
							v1.ResourceMemory: memory.DeepCopy(),
						},
						Allocatable: v1.ResourceList{
							v1.ResourceCPU:    cpu.DeepCopy(),
							v1.ResourceMemory: memory.DeepCopy(),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node02"},
					Status: v1.NodeStatus{
						Capacity: v1.ResourceList{
							v1.ResourceCPU:    cpu.DeepCopy(),
							v1.ResourceMemory: memory.DeepCopy(),
						},
						Allocatable: v1.ResourceList{
							v1.ResourceCPU:    cpu.DeepCopy(),
							v1.ResourceMemory: memory.DeepCopy(),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node03"},
					Spec: v1.NodeSpec{
						Taints: []v1.Taint{
							{
								Key:    "",
								Value:  "",
								Effect: "NoExecute",
							},
						},
					},
					Status: v1.NodeStatus{
						Capacity: v1.ResourceList{
							v1.ResourceCPU:    cpu.DeepCopy(),
							v1.ResourceMemory: memory.DeepCopy(),
						},
						Allocatable: v1.ResourceList{
							v1.ResourceCPU:    cpu.DeepCopy(),
							v1.ResourceMemory: memory.DeepCopy(),
						},
					},
				},
			},
		})
	cpuUsage := &resource.Quantity{}
	cpuUsage.Add(resource.MustParse("1"))
	memoryUsage := &resource.Quantity{}
	memoryUsage.Add(resource.MustParse("1Gi"))
	fcm := &fake2.Clientset{}
	fcm.AddReactor("list", "nodes", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1beta1.NodeMetricsList{
			ListMeta: metav1.ListMeta{
				ResourceVersion: "1",
			},
			Items: []v1beta1.NodeMetrics{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node01"},
					Window:     metav1.Duration{Duration: time.Minute},
					Usage: v1.ResourceList{
						v1.ResourceCPU:    cpuUsage.DeepCopy(),
						v1.ResourceMemory: memoryUsage.DeepCopy(),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node02"},
					Window:     metav1.Duration{Duration: time.Minute},
					Usage:      v1.ResourceList{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node03"},
					Window:     metav1.Duration{Duration: time.Minute},
					Usage: v1.ResourceList{
						v1.ResourceCPU:    cpuUsage.DeepCopy(),
						v1.ResourceMemory: memoryUsage.DeepCopy(),
					},
				},
			},
		}, nil
	})
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{Client: fc, MetricsClient: fcm})
	info := ClusterInfo()
	assert.Equal(t, &InfoResponse{
		Status:            "health",
		FreeMemory:        "19 GiB",
		FreeCpu:           "5.00 core",
		FreeRequestMemory: "18 GiB",
		FreeRequestCpu:    "3.00 core",
		TotalMemory:       "20 GiB",
		TotalCpu:          "6.00 core",
		UsageMemoryRate:   "5.0%",
		UsageCpuRate:      "16.7%",
		RequestMemoryRate: "10.0%",
		RequestCpuRate:    "50.0%",
	}, info)
}

func Test_getNodeRequestCpuAndMemory(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	cpu := &resource.Quantity{}
	cpu.Add(resource.MustParse("3"))
	memory := &resource.Quantity{}
	memory.Add(resource.MustParse("10Gi"))
	fc := fake.NewSimpleClientset(
		&v1.PodList{
			TypeMeta: metav1.TypeMeta{},
			ListMeta: metav1.ListMeta{},
			Items: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
					Spec: v1.PodSpec{
						NodeName: "node01",
						Containers: []v1.Container{
							{
								Name:  "app",
								Image: "xxx:v1",
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										// 3core cpu request
										v1.ResourceCPU: *resource.NewMilliQuantity(3000, resource.DecimalSI),
										// 2G memory request
										v1.ResourceMemory: *resource.NewQuantity(2*(1024*1024*1024), resource.DecimalSI),
									},
								},
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod2"},
					Spec: v1.PodSpec{
						NodeName: "node02",
						Containers: []v1.Container{
							{
								Name:  "app",
								Image: "xxx:v2",
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										// 3core cpu request
										v1.ResourceCPU: *resource.NewMilliQuantity(3000, resource.DecimalSI),
										// 2G memory request
										v1.ResourceMemory: *resource.NewQuantity(2*(1024*1024*1024), resource.DecimalSI),
									},
								},
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
			},
		},
	)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{Client: fc})

	// FIXME: fake client 没办法过滤 node
	c, m := getNodeRequestCpuAndMemory([]v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node3"},
		},
	})
	assert.Equal(t, "6", c.String())
	assert.Equal(t, fmt.Sprintf("%d", 4*(1024*1024*1024)), m.String())
}

func Test_getStatus(t *testing.T) {
	var tests = []struct {
		CpuRate    float64
		MemoryRate float64
		Wants      ClusterStatus
	}{
		{
			CpuRate:    60,
			MemoryRate: 60,
			Wants:      StatusHealth,
		},
		{
			CpuRate:    61,
			MemoryRate: 61,
			Wants:      StatusNotGood,
		},
		{
			CpuRate:    61,
			MemoryRate: 10,
			Wants:      StatusNotGood,
		},
		{
			CpuRate:    10,
			MemoryRate: 60,
			Wants:      StatusHealth,
		},
		{
			CpuRate:    81,
			MemoryRate: 81,
			Wants:      StatusBad,
		},
		{
			CpuRate:    10,
			MemoryRate: 81,
			Wants:      StatusBad,
		},
		{
			CpuRate:    81,
			MemoryRate: 1,
			Wants:      StatusBad,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("cpu:%.0f-memory:%.0f-%s", test.CpuRate, test.MemoryRate, test.Wants), func(t *testing.T) {
			assert.Equal(t, test.Wants, getStatus(test.MemoryRate, test.CpuRate))
		})
	}
}
