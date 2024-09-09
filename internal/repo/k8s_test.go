package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/uploader"
	"github.com/duc-cnzj/mars/v5/internal/util/k8s"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	eventsv1lister "k8s.io/client-go/listers/events/v1"
	restclient "k8s.io/client-go/rest"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	fake2 "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

func TestNewK8sRepo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	fileRepo := NewMockFileRepo(m)
	mockUploader := uploader.NewMockUploader(m)
	mockData.EXPECT().Config().Return(&config.Config{})
	repo := NewK8sRepo(
		mlog.NewForConfig(nil),
		timer.NewRealTimer(),
		mockData,
		fileRepo,
		mockUploader,
		NewDefaultArchiver(),
		NewExecutorManager(mockData, mlog.NewForConfig(nil)),
	).(*k8sRepo)
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.logger)
	assert.NotNil(t, repo.timer)
	assert.NotNil(t, repo.data)
	assert.NotNil(t, repo.fileRepo)
	assert.NotNil(t, repo.uploader)
	assert.NotNil(t, repo.archiver)
	assert.NotNil(t, repo.executor)
}

func TestSplitManifests(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	fileRepo := NewMockFileRepo(m)
	mockUploader := uploader.NewMockUploader(m)
	mockData.EXPECT().Config().Return(&config.Config{})
	repo := NewK8sRepo(
		mlog.NewForConfig(nil),
		timer.NewRealTimer(),
		mockData,
		fileRepo,
		mockUploader,
		NewDefaultArchiver(),
		NewExecutorManager(mockData, mlog.NewForConfig(nil)),
	).(*k8sRepo)

	t.Run("should split manifest string correctly", func(t *testing.T) {
		manifest := "manifest1\n---\nmanifest2\n---\nmanifest3"
		expected := []string{"manifest1", "manifest2", "manifest3"}

		result := repo.SplitManifests(manifest)

		assert.Equal(t, expected, result)
	})

	t.Run("should return single manifest when no delimiters", func(t *testing.T) {
		manifest := "manifest1"
		expected := []string{"manifest1"}

		result := repo.SplitManifests(manifest)

		assert.Equal(t, expected, result)
	})

	t.Run("should handle empty manifest string", func(t *testing.T) {
		manifest := ""
		expected := []string{}

		result := repo.SplitManifests(manifest)

		assert.Equal(t, expected, result)
	})
}

func Test_k8sRepo_CreateDockerSecrets(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	mockData.EXPECT().Config().Return(&config.Config{
		ImagePullSecrets: config.DockerAuths{
			{
				Username: "a",
				Password: "b",
				Email:    "cc",
				Server:   "d",
			},
		},
	})
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{Client: clientset})
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	secret, err := kr.CreateDockerSecret(context.TODO(), "a")
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(secret.Name, "mars-"))
	assert.Equal(t, corev1.SecretTypeDockerConfigJson, secret.Type)
	d := DockerConfigJSON{}
	json.Unmarshal(secret.Data[corev1.DockerConfigJsonKey], &d)
	assert.Equal(t, "a", d.Auths["d"].Username)
	assert.Equal(t, "b", d.Auths["d"].Password)
	assert.Equal(t, "cc", d.Auths["d"].Email)
}

func Test_k8sRepo_GetNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{Client: clientset}).AnyTimes()
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	namespace, err := kr.GetNamespace(context.TODO(), "a")
	assert.Error(t, err)
	assert.Nil(t, namespace)
	_, err = kr.CreateNamespace(context.TODO(), "a")
	assert.Nil(t, err)
	namespace, err = kr.GetNamespace(context.TODO(), "a")
	assert.Nil(t, err)
	assert.Equal(t, "a", namespace.Name)
}

func NewEventLister(events ...*eventsv1.Event) eventsv1lister.EventLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range events {
		idxer.Add(po)
	}
	return eventsv1lister.NewEventLister(idxer)
}

func Test_k8sRepo_ListEvents(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{EventLister: NewEventLister(&eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ev",
			Namespace: "a",
		},
	})}).AnyTimes()
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	events, err := kr.ListEvents("a")
	assert.Nil(t, err)
	assert.Len(t, events, 1)
}

func TestGetPod(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{PodLister: NewPodLister(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "po",
			Namespace: "a",
		},
	})}).AnyTimes()
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	pod, err := kr.GetPod("a", "po")
	assert.Nil(t, err)
	assert.Equal(t, "po", pod.Name)
}

func TestFindDefaultContainer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{PodLister: NewPodLister(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				defaultContainerAnnotationName: "second-container",
			},
			Name:      "pod",
			Namespace: "a",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "first-container",
				},
				{
					Name: "second-container",
				},
			},
		},
	}, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-b",
			Namespace: "a",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "first-container",
				},
				{
					Name: "second-container",
				},
			},
		},
	})}).AnyTimes()

	_, err := kr.FindDefaultContainer(context.TODO(), "a", "c")
	assert.Error(t, err)

	container, err := kr.FindDefaultContainer(context.TODO(), "a", "pod")
	assert.Nil(t, err)
	assert.Equal(t, "second-container", container)

	container, err = kr.FindDefaultContainer(context.TODO(), "a", "pod-b")
	assert.Nil(t, err)
	assert.Equal(t, "first-container", container)
}

func TestIsPodRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{PodLister: NewPodLister(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					defaultContainerAnnotationName: "second-container",
				},
				Name:      "pod",
				Namespace: "a",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "first-container",
					},
					{
						Name: "second-container",
					},
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-c",
				Namespace: "a",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "first-container",
					},
					{
						Name: "second-container",
					},
				},
			},
			Status: corev1.PodStatus{
				Phase:  corev1.PodFailed,
				Reason: "Evicted",
			},
		},
	)}).AnyTimes()

	running, reason := kr.IsPodRunning("a", "pod")
	assert.True(t, running)
	assert.Empty(t, reason)
	running, reason = kr.IsPodRunning("a", "pod-b")
	assert.False(t, running)
	assert.NotEmpty(t, reason)
	running, reason = kr.IsPodRunning("a", "pod-c")
	assert.False(t, running)
	assert.Equal(t, "po pod-c already evicted in namespace a!", reason)
}

func TestGetCpuAndMemoryQuantity(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}

	t.Run("should return correct cpu and memory quantities", func(t *testing.T) {
		podMetrics := v1beta1.PodMetrics{
			Containers: []v1beta1.ContainerMetrics{
				{
					Usage: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
				},
				{
					Usage: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("200Mi"),
					},
				},
			},
		}

		cpu, memory := kr.GetCpuAndMemoryQuantity(podMetrics)

		assert.Equal(t, "300m", cpu.String())
		assert.Equal(t, "300Mi", memory.String())
	})

	t.Run("should return zero cpu and memory quantities when no containers", func(t *testing.T) {
		podMetrics := v1beta1.PodMetrics{
			Containers: []v1beta1.ContainerMetrics{},
		}

		cpu, memory := kr.GetCpuAndMemoryQuantity(podMetrics)

		assert.Equal(t, "<nil>", cpu.String())
		assert.Equal(t, "<nil>", memory.String())
	})
}

func TestAnalyseMetricsToCpuAndMemory(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}

	t.Run("should return correct cpu and memory when list is not empty", func(t *testing.T) {
		list := []v1beta1.PodMetrics{
			{
				Containers: []v1beta1.ContainerMetrics{
					{
						Usage: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("100M"),
						},
					},
					{
						Usage: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
		}

		cpuStr, memoryStr := kr.GetCpuAndMemory(context.TODO(), list)

		assert.Equal(t, "300 m", cpuStr)
		assert.Equal(t, "300 MB", memoryStr)
	})

	t.Run("should return zero cpu and memory when list is empty", func(t *testing.T) {
		list := []v1beta1.PodMetrics{}

		cpuStr, memoryStr := kr.GetCpuAndMemory(context.TODO(), list)

		assert.Equal(t, "0 m", cpuStr)
		assert.Equal(t, "0 MB", memoryStr)
	})

	t.Run("should return zero cpu and memory when containers are empty", func(t *testing.T) {
		list := []v1beta1.PodMetrics{
			{
				Containers: []v1beta1.ContainerMetrics{},
			},
		}

		cpuStr, memoryStr := kr.GetCpuAndMemory(context.TODO(), list)

		assert.Equal(t, "0 m", cpuStr)
		assert.Equal(t, "0 MB", memoryStr)
	})
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
			assert.Equal(t, test.Wants, (&k8sRepo{}).getStatus(test.MemoryRate, test.CpuRate))
		})
	}
}

func TestClusterInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cpu := &resource.Quantity{}
	cpu.Add(resource.MustParse("3"))
	memory := &resource.Quantity{}
	memory.Add(resource.MustParse("10Gi"))
	fc := fake.NewSimpleClientset(
		&corev1.PodList{
			TypeMeta: metav1.TypeMeta{},
			ListMeta: metav1.ListMeta{},
			Items: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
					Spec: corev1.PodSpec{
						NodeName: "node01",
						Containers: []corev1.Container{
							{
								Name:  "app",
								Image: "xxx:v1",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										// 3core cpu request
										corev1.ResourceCPU: *resource.NewMilliQuantity(3000, resource.DecimalSI),
										// 2G memory request
										corev1.ResourceMemory: *resource.NewQuantity(2*(1024*1024*1024), resource.DecimalSI),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
				// FIXME: fake 客户端不能做 fieldSelector 过滤
			},
		},
		&corev1.NodeList{
			Items: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node01"},
					Status: corev1.NodeStatus{
						Capacity: corev1.ResourceList{
							corev1.ResourceCPU:    cpu.DeepCopy(),
							corev1.ResourceMemory: memory.DeepCopy(),
						},
						Allocatable: corev1.ResourceList{
							corev1.ResourceCPU:    cpu.DeepCopy(),
							corev1.ResourceMemory: memory.DeepCopy(),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node02"},
					Status: corev1.NodeStatus{
						Capacity: corev1.ResourceList{
							corev1.ResourceCPU:    cpu.DeepCopy(),
							corev1.ResourceMemory: memory.DeepCopy(),
						},
						Allocatable: corev1.ResourceList{
							corev1.ResourceCPU:    cpu.DeepCopy(),
							corev1.ResourceMemory: memory.DeepCopy(),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node03"},
					Spec: corev1.NodeSpec{
						Taints: []corev1.Taint{
							{
								Key:    "",
								Value:  "",
								Effect: "NoExecute",
							},
						},
					},
					Status: corev1.NodeStatus{
						Capacity: corev1.ResourceList{
							corev1.ResourceCPU:    cpu.DeepCopy(),
							corev1.ResourceMemory: memory.DeepCopy(),
						},
						Allocatable: corev1.ResourceList{
							corev1.ResourceCPU:    cpu.DeepCopy(),
							corev1.ResourceMemory: memory.DeepCopy(),
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
					Usage: corev1.ResourceList{
						corev1.ResourceCPU:    cpuUsage.DeepCopy(),
						corev1.ResourceMemory: memoryUsage.DeepCopy(),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node02"},
					Window:     metav1.Duration{Duration: time.Minute},
					Usage:      corev1.ResourceList{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node03"},
					Window:     metav1.Duration{Duration: time.Minute},
					Usage: corev1.ResourceList{
						corev1.ResourceCPU:    cpuUsage.DeepCopy(),
						corev1.ResourceMemory: memoryUsage.DeepCopy(),
					},
				},
			},
		}, nil
	})
	mockData := data.NewMockData(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{Client: fc, MetricsClient: fcm}).AnyTimes()
	info := kr.ClusterInfo()
	assert.Equal(t, &ClusterInfo{
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
	m := gomock.NewController(t)
	defer m.Finish()
	cpu := &resource.Quantity{}
	cpu.Add(resource.MustParse("3"))
	memory := &resource.Quantity{}
	memory.Add(resource.MustParse("10Gi"))
	fc := fake.NewSimpleClientset(
		&corev1.PodList{
			TypeMeta: metav1.TypeMeta{},
			ListMeta: metav1.ListMeta{},
			Items: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
					Spec: corev1.PodSpec{
						NodeName: "node01",
						Containers: []corev1.Container{
							{
								Name:  "app",
								Image: "xxx:corev1",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										// 3core cpu request
										corev1.ResourceCPU: *resource.NewMilliQuantity(3000, resource.DecimalSI),
										// 2G memory request
										corev1.ResourceMemory: *resource.NewQuantity(2*(1024*1024*1024), resource.DecimalSI),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod2"},
					Spec: corev1.PodSpec{
						NodeName: "node02",
						Containers: []corev1.Container{
							{
								Name:  "app",
								Image: "xxx:v2",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										// 3core cpu request
										corev1.ResourceCPU: *resource.NewMilliQuantity(3000, resource.DecimalSI),
										// 2G memory request
										corev1.ResourceMemory: *resource.NewQuantity(2*(1024*1024*1024), resource.DecimalSI),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
		},
	)

	mockData := data.NewMockData(m)
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{Client: fc}).AnyTimes()
	// FIXME: fake client 没办法过滤 node
	c, mem := (&k8sRepo{
		data: mockData,
	}).getNodeRequestCpuAndMemory([]corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node3"},
		},
	})
	assert.Equal(t, "6", c.String())
	assert.Equal(t, fmt.Sprintf("%d", 4*(1024*1024*1024)), mem.String())
}

func Test_getPodSelectorsInDeploymentAndStatefulSetByManifest(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{
			in: dedent.Dedent(`
				apiVersion: apps/v1
				kind: Deployment
				metadata:
				  annotations:
				    meta.helm.sh/release-name: mars
				  generation: 56
				  labels:
				    app.kubernetes.io/name: mars
				  name: mars
				  namespace: default
				spec:
				  selector:
				    matchLabels:
				      app.kubernetes.io/instance: mars
				      app.kubernetes.io/name: mars
				`),
			out: "app.kubernetes.io/instance=mars,app.kubernetes.io/name=mars",
		},
		{
			in: dedent.Dedent(`
				apiVersion: apps/v1
				kind: Deployment
				metadata:
				  annotations:
				    meta.helm.sh/release-name: mars
				  generation: 56
				  labels:
				    app.kubernetes.io/name: mars
				  name: mars
				  namespace: default
				spec:
				  selector:
				    matchLabels:
				      app.kubernetes.io/instance: abc
				      app.kubernetes.io/name: abc
				`),
			out: "app.kubernetes.io/instance=abc,app.kubernetes.io/name=abc",
		},
		{
			in: dedent.Dedent(`
				apiVersion: apps/v1
				kind: StatefulSet
				metadata:
				  labels:
				    app.kubernetes.io/component: primary
				    app.kubernetes.io/instance: mars-db
				  name: mars-db-mysql-primary
				  namespace: default
				spec:
				  selector:
				    matchLabels:
				      app.kubernetes.io/component: primary
				      app.kubernetes.io/instance: mars-db
				      app.kubernetes.io/name: mysql
				`),
			out: "app.kubernetes.io/component=primary,app.kubernetes.io/instance=mars-db,app.kubernetes.io/name=mysql",
		},
		{
			in: dedent.Dedent(`
				W0509 17:36:48.835823   98185 helpers.go:555] --dry-run is deprecated and can be replaced with --dry-run=client.
				apiVersion: v1
				kind: Pod
				metadata:
				  creationTimestamp: null
				  labels:
				    run: nginx
				  name: nginx
				spec:
				  containers:
				  - image: nginx
				    name: nginx
				    resources: {}
				  dnsPolicy: ClusterFirst
				  restartPolicy: Always
				status: {}
				`),
			out: "",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1
				kind: Job
				metadata:
				  name: pi
				spec:
				  template:
				    spec:
				      containers:
				      - name: pi
				        image: perl:5.34.0
				        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
				      restartPolicy: Never
				  backoffLimit: 4
				`),
			out: "",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1
				kind: Job
				metadata:
				  name: pi
				spec:
				  template:
				    metadata:
				      labels:
				        app: jobRunner-one
				    spec:
				      containers:
				      - name: pi
				        image: perl:5.34.0
				        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
				      restartPolicy: Never
				  backoffLimit: 4
				`),
			out: "app=jobRunner-one",
		},
		{
			in: dedent.Dedent(`
				apiVersion: apps/v1
				kind: DaemonSet
				metadata:
				  name: fluentd-elasticsearch
				spec:
				  selector:
				    matchLabels:
				      name: fluentd-elasticsearch
				  template:
				    metadata:
				      labels:
				        name: fluentd-elasticsearch
				    spec:
				      containers:
				      - name: fluentd-elasticsearch
				        image: quay.io/fluentd_elasticsearch/fluentd:v2.5.2
				        volumeMounts:
				        - name: varlog
				          mountPath: /var/log
				`),
			out: "name=fluentd-elasticsearch",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1
				kind: CronJob
				metadata:
				  name: hello
				spec:
				  schedule: "* * * * *"
				  jobTemplate:
				    spec:
				      template:
				        metadata:
				          labels:
				            app: cronjob
				        spec:
				          containers:
				          - name: hello
				            image: busybox:1.28
				            imagePullPolicy: IfNotPresent
				            command:
				            - /bin/sh
				            - -c
				            - date; echo Hello from the Kubernetes cluster
				          restartPolicy: OnFailure
				`),
			out: "app=cronjob",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1beta1
				kind: CronJob
				metadata:
				  name: hello
				spec:
				  schedule: "* * * * *"
				  jobTemplate:
				    spec:
				      template:
				        metadata:
				          labels:
				            app: cronjob-v1beta1
				        spec:
				          containers:
				          - name: hello
				            image: busybox:1.28
				            imagePullPolicy: IfNotPresent
				            command:
				            - /bin/sh
				            - -c
				            - date; echo Hello from the Kubernetes cluster
				          restartPolicy: OnFailure
				`),
			out: "app=cronjob-v1beta1",
		},
	}

	for _, test := range tests {
		tt := test
		t.Run("", func(t *testing.T) {
			labels := (&k8sRepo{}).GetPodSelectorsByManifest([]string{tt.in})
			if len(labels) > 0 {
				assert.Equal(t, tt.out, labels[0])
			} else {
				assert.Equal(t, tt.out, "")
			}
		})
	}
}

func TestDeleteNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{Client: clientset}).AnyTimes()
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	err := kr.DeleteNamespace(context.TODO(), "a")
	assert.Error(t, err)
	kr.CreateNamespace(context.TODO(), "a")
	err = kr.DeleteNamespace(context.TODO(), "a")
	assert.Nil(t, err)
}

func TestDeleteSecret(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{Client: clientset}).AnyTimes()
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	err := kr.DeleteSecret(context.TODO(), "a", "s")
	assert.Error(t, err)
	clientset.CoreV1().Secrets("a").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "s",
			Namespace: "a",
		},
	}, metav1.CreateOptions{})
	err = kr.DeleteSecret(context.TODO(), "a", "s")
	assert.Nil(t, err)
}

func TestExecutor(t *testing.T) {
	ex := &executor{}
	ex.WithMethod("GET")
	assert.Equal(t, "GET", ex.method)
	ex.WithContainer("a", "b", "c")
	assert.Equal(t, "a", ex.namespace)
	assert.Equal(t, "b", ex.pod)
	assert.Equal(t, "c", ex.container)
	ex.WithCommand([]string{"ls"})
	assert.Equal(t, []string{"ls"}, ex.cmd)

	option := ex.newOption(nil, nil, nil, true)
	assert.False(t, option.Stdin)
	assert.False(t, option.Stdout)
	assert.False(t, option.Stderr)
	assert.True(t, option.TTY)
	assert.Equal(t, "c", option.Container)
	assert.Equal(t, []string{"ls"}, option.Command)

	bf := &bytes.Buffer{}
	option = ex.newOption(bf, bf, bf, false)
	assert.True(t, option.Stdin)
	assert.True(t, option.Stdout)
	assert.True(t, option.Stderr)
}

func Test_defaultRemoteExecutor_New(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{}).Times(2)
	v := &defaultRemoteExecutor{data: mockData}
	v.New()
	assert.NotNil(t, v)
}

func Test_k8sRepo_Execute(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	manager := NewMockExecutorManager(m)
	r := &k8sRepo{
		executor: manager,
	}
	ec := NewMockExecutor(m)
	manager.EXPECT().New().Return(ec)
	c := &Container{
		Namespace: "a",
		Pod:       "v",
		Container: "c",
	}
	ec.EXPECT().WithContainer(c.Namespace, c.Pod, c.Container).Return(ec)
	ec.EXPECT().WithMethod("POST").Return(ec)
	ec.EXPECT().WithCommand([]string{"ls"}).Return(ec)
	input := &ExecuteInput{
		Cmd:               []string{"ls"},
		TerminalSizeQueue: nil,
	}
	ec.EXPECT().Execute(gomock.Any(), input)
	assert.Nil(t, r.Execute(context.TODO(), c, input))
}

func Test_defaultRemoteExecutor_NewFileCopy(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	d := &defaultRemoteExecutor{
		data:   mockData,
		logger: mlog.NewForConfig(nil),
	}
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{
		RestConfig: &restclient.Config{},
	}).Times(2)
	fileCopy := d.NewFileCopy(1, &bytes.Buffer{})
	assert.NotNil(t, fileCopy)
}
func TestGetPodLogs(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8sClient().Return(&data.K8sClient{Client: clientset}).AnyTimes()
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}

	t.Run("should return logs when pod exists", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
			},
		}
		clientset.CoreV1().Pods("test-namespace").Create(context.TODO(), pod, metav1.CreateOptions{})
		clientset.CoreV1().Pods("test-namespace").UpdateStatus(context.TODO(), pod, metav1.UpdateOptions{})
		clientset.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
		_, err := kr.GetPodLogs(context.TODO(), "test-namespace", "test-pod", &corev1.PodLogOptions{})
		assert.Nil(t, err)
	})
}
func TestCopyFromPod(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	manager := NewMockExecutorManager(m)
	mockExecutor := NewMockExecutor(m)
	manager.EXPECT().New().Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithCommand(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithContainer("test-namespace", "test-pod", "").Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithMethod(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*ExecuteInput)
		input.Stderr.Write([]byte("xxx"))
		return slices.Equal(input.Cmd, []string{"sh", "-c", "ls -l " + "/test/file/path"})
	})).Return(nil).AnyTimes()
	kr := &k8sRepo{
		logger:   mlog.NewForConfig(nil),
		data:     mockData,
		uploader: mockUploader,
		fileRepo: mockFileRepo,
		executor: manager,
	}

	_, err := kr.CopyFromPod(context.TODO(), &CopyFromPodInput{
		Namespace: "test-namespace",
		Pod:       "test-pod",
		FilePath:  "/test/file/path",
		UserName:  "test-user",
	})
	s, _ := status.FromError(err)
	assert.Equal(t, "xxx", s.Message())
	assert.Equal(t, codes.InvalidArgument, s.Code())
}

func TestCopyFromPod1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	manager := NewMockExecutorManager(m)
	mockExecutor := NewMockExecutor(m)
	manager.EXPECT().New().Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithCommand(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithContainer("test-namespace", "test-pod", "").Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithMethod(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*ExecuteInput)
		return slices.Equal(input.Cmd, []string{"sh", "-c", "ls -l " + "/test/file/path"})
	})).Return(nil)
	kr := &k8sRepo{
		logger:   mlog.NewForConfig(nil),
		data:     mockData,
		uploader: mockUploader,
		timer:    timer.NewRealTimer(),
		fileRepo: mockFileRepo,
		executor: manager,
	}

	manager.EXPECT().NewFileCopy(5, gomock.Any()).Return(nil)
	mockUploader.EXPECT().Disk("podfile").Return(mockUploader)
	mockUploader.EXPECT().NewFile(gomock.Any()).Return(nil, errors.New("x"))

	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*ExecuteInput)
		input.Stdout.Write([]byte("/"))
		return slices.Equal(input.Cmd, []string{"sh", "-c", "pwd"})
	})).Return(nil)

	_, err := kr.CopyFromPod(context.TODO(), &CopyFromPodInput{
		Namespace: "test-namespace",
		Pod:       "test-pod",
		FilePath:  "/test/file/path",
		UserName:  "test-user",
	})
	assert.Equal(t, "x", err.Error())
}

func TestCopyFromPod_success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	manager := NewMockExecutorManager(m)
	mockExecutor := NewMockExecutor(m)
	manager.EXPECT().New().Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithCommand(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithContainer("test-namespace", "test-pod", "").Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithMethod(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*ExecuteInput)
		return slices.Equal(input.Cmd, []string{"sh", "-c", "ls -l " + "/test/file/path"})
	})).Return(nil)
	kr := &k8sRepo{
		logger:   mlog.NewForConfig(nil),
		data:     mockData,
		uploader: mockUploader,
		timer:    timer.NewRealTimer(),
		fileRepo: mockFileRepo,
		executor: manager,
	}

	manager.EXPECT().NewFileCopy(5, gomock.Any()).Return(&mockFileCopy{})
	mockUploader.EXPECT().Disk("podfile").Return(mockUploader)
	file := uploader.NewMockFile(m)
	mockUploader.EXPECT().NewFile(gomock.Any()).Return(file, nil)

	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*ExecuteInput)
		input.Stdout.Write([]byte("/"))
		return slices.Equal(input.Cmd, []string{"sh", "-c", "pwd"})
	})).Return(nil)

	info := &mockFileInfo{
		size: 1,
	}
	file.EXPECT().Stat().Return(info, nil)
	file.EXPECT().Close()
	file.EXPECT().Name().Return("fname")

	mockFileRepo.EXPECT().Create(gomock.Any(), &CreateFileInput{
		Path:       "fname",
		Username:   "test-user",
		Size:       uint64(1),
		UploadType: schematype.Local,
		Namespace:  "test-namespace",
		Pod:        "test-pod",
		Container:  "",
	})

	mockUploader.EXPECT().Type().Return(schematype.Local)
	_, err := kr.CopyFromPod(context.TODO(), &CopyFromPodInput{
		Namespace: "test-namespace",
		Pod:       "test-pod",
		FilePath:  "/test/file/path",
		UserName:  "test-user",
	})
	assert.Nil(t, err)
}

var _ k8s.FileCopy = (*mockFileCopy)(nil)

type mockFileCopy struct{}

func (m *mockFileCopy) CopyFromPod(ctx context.Context, src k8s.CopyFileSpec, file uploader.File) error {
	return nil
}
