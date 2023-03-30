package services

import (
	"context"
	"testing"
	"time"

	"github.com/duc-cnzj/mars-client/v4/cluster"
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

func TestClusterSvc_AuthFuncOverride(t *testing.T) {
	c := new(clusterSvc)
	_, err := c.AuthFuncOverride(context.TODO(), "")
	assert.Nil(t, err)
}

func TestClusterSvc_ClusterInfo(t *testing.T) {
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
	c := new(clusterSvc)
	info, err := c.ClusterInfo(context.TODO(), &cluster.InfoRequest{})
	assert.Nil(t, err)
	// 2 个 node(10G memory 3 core) 20G 6 core
	// 使用 1 core 1 G
	// request 3core 2G memory
	assert.Equal(t, &cluster.InfoResponse{
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
