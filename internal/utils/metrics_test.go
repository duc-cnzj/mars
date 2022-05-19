package utils

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

func TestGetCpuAndMemory(t *testing.T) {
	cpu, memory := GetCpuAndMemory([]v1beta1.PodMetrics{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod",
				Namespace: "ns",
			},
			Timestamp: metav1.Time{},
			Window:    metav1.Duration{},
			Containers: []v1beta1.ContainerMetrics{
				{
					Name: "container1",
					Usage: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(4, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
					},
				},
				{
					Name: "container2",
					Usage: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(10, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
					},
				},
			},
		},
	})
	assert.Equal(t, "14 m", cpu)
	assert.Equal(t, "10 MB", memory)
}

func TestGetCpuAndMemoryInNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	fm := &fake.Clientset{}
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{MetricsClient: fm})
	fm.AddReactor("list", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		res := &v1beta1.PodMetricsList{
			ListMeta: metav1.ListMeta{
				ResourceVersion: "2",
			},
			Items: []v1beta1.PodMetrics{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "ns",
					},
					Timestamp: metav1.Time{},
					Window:    metav1.Duration{},
					Containers: []v1beta1.ContainerMetrics{
						{
							Name: "container1",
							Usage: v1.ResourceList{
								v1.ResourceCPU:    *resource.NewMilliQuantity(5, resource.DecimalSI),
								v1.ResourceMemory: *resource.NewQuantity(6*(1000*1000), resource.DecimalSI),
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: "ns",
					},
					Timestamp: metav1.Time{},
					Window:    metav1.Duration{},
					Containers: []v1beta1.ContainerMetrics{
						{
							Name: "container1",
							Usage: v1.ResourceList{
								v1.ResourceCPU:    *resource.NewMilliQuantity(4, resource.DecimalSI),
								v1.ResourceMemory: *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
							},
						},
					},
				},
			},
		}
		return true, res, nil
	})
	cpu, memory := GetCpuAndMemoryInNamespace("ns")
	assert.Equal(t, "9 m", cpu)
	assert.Equal(t, "11 MB", memory)
}

func TestGetCpuAndMemoryQuantity(t *testing.T) {
	cpu, memory := GetCpuAndMemoryQuantity(v1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "ns",
		},
		Timestamp: metav1.Time{},
		Window:    metav1.Duration{},
		Containers: []v1beta1.ContainerMetrics{
			{
				Name: "container1",
				Usage: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(4, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
				},
			},
		},
	})
	assert.Equal(t, resource.NewMilliQuantity(4, resource.DecimalSI).String(), cpu.String())
	assert.Equal(t, resource.NewQuantity(5*(1000*1000), resource.DecimalSI).String(), memory.String())
}

func Test_analyseMetricsToCpuAndMemory(t *testing.T) {
	cpu, memory := analyseMetricsToCpuAndMemory([]v1beta1.PodMetrics{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod",
				Namespace: "ns",
			},
			Timestamp: metav1.Time{},
			Window:    metav1.Duration{},
			Containers: []v1beta1.ContainerMetrics{
				{
					Name: "container1",
					Usage: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(4, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
					},
				},
				{
					Name: "container2",
					Usage: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(10, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "ns",
			},
			Timestamp: metav1.Time{},
			Window:    metav1.Duration{},
			Containers: []v1beta1.ContainerMetrics{
				{
					Name: "container1",
					Usage: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(4, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
					},
				},
				{
					Name: "container2",
					Usage: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(10, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
					},
				},
			},
		},
	})
	assert.Equal(t, "28 m", cpu)
	assert.Equal(t, "20 MB", memory)
}
