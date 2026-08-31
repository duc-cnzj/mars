package data

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/k8sutil"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	appsv1lister "k8s.io/client-go/listers/apps/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	eventsv1lister "k8s.io/client-go/listers/events/v1"
	restclient "k8s.io/client-go/rest"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/remotecommand"
	clientgoexec "k8s.io/client-go/util/exec"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	fake2 "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

func TestNewK8sRepo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	fileRepo := NewMockFileRepo(m)
	mockUploader := uploader.NewMockUploader(m)
	mockData.EXPECT().Config().Return(&config.Config{})
	repo := NewK8sRepo(
		mlog.NewForConfig(nil),
		timer.NewReal(),
		mockData,
		fileRepo,
		mockUploader,
		NewDefaultArchiver(),
		NewExecutorManager(mockData, mlog.NewForConfig(nil)),
		NewCacheImpl(&config.Config{}, nil, mlog.NewForConfig(nil)),
	).(*k8sRepo)
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.logger)
	assert.NotNil(t, repo.timer)
	assert.NotNil(t, repo.data)
	assert.NotNil(t, repo.fileRepo)
	assert.NotNil(t, repo.uploader)
	assert.NotNil(t, repo.archiver)
	assert.NotNil(t, repo.executor)
	assert.NotNil(t, repo.cache)
}

// GetSecret 通过 K8sClient 实时读取命名空间下的 secret：
// 存在返回 secret，不存在返回 k8s NotFound 错误。
func TestK8sRepo_GetSecret(t *testing.T) {
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-tls", Namespace: "default"},
		Type:       corev1.SecretTypeTLS,
	}
	d := NewDataImpl(&NewDataParams{
		Cfg: &config.Config{},
		K8sClient: &K8sClient{
			Client: fake.NewSimpleClientset(sec),
		},
	})
	repo := &k8sRepo{data: d}

	got, err := repo.GetSecret(context.TODO(), "default", "my-tls")
	assert.NoError(t, err)
	assert.Equal(t, sec.Name, got.Name)

	_, err = repo.GetSecret(context.TODO(), "default", "missing")
	assert.Error(t, err)
}

func TestSplitManifests(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	fileRepo := NewMockFileRepo(m)
	mockUploader := uploader.NewMockUploader(m)
	mockData.EXPECT().Config().Return(&config.Config{})
	repo := NewK8sRepo(
		mlog.NewForConfig(nil),
		timer.NewReal(),
		mockData,
		fileRepo,
		mockUploader,
		NewDefaultArchiver(),
		NewExecutorManager(mockData, mlog.NewForConfig(nil)),
		NewCacheImpl(&config.Config{}, nil, mlog.NewForConfig(nil)),
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
	mockData := NewMockDataStore(m)
	// CreateDockerSecret 委托 CreateDockerSecrets：收集 servers 读一次 + 内部过滤读一次。
	mockData.EXPECT().Config().Return(&config.Config{
		ImagePullSecrets: config.DockerAuths{
			{
				Username: "a",
				Password: "b",
				Email:    "cc",
				Server:   "d",
			},
		},
	}).Times(2)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8s().Return(&K8sClient{Client: clientset})
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	secret, err := kr.CreateDockerSecret(context.TODO(), "a")
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(secret.Name, "mars-"))
	assert.Equal(t, corev1.SecretTypeDockerConfigJson, secret.Type)
	d := biz.DockerConfigJSON{}
	json.Unmarshal(secret.Data[corev1.DockerConfigJsonKey], &d)
	assert.Equal(t, "a", d.Auths["d"].Username)
	assert.Equal(t, "b", d.Auths["d"].Password)
	assert.Equal(t, "cc", d.Auths["d"].Email)

	// 序列化契约锁定：docker config.json 必须是小写键，kubelet 解析依赖此格式。
	// 曾因 biz.DockerConfigEntry 缺 json tag 序列化出大写键导致认证失效（P0）。
	raw := map[string]map[string]map[string]string{}
	require.NoError(t, json.Unmarshal(secret.Data[corev1.DockerConfigJsonKey], &raw))
	assert.Contains(t, raw, "auths")
	entry := raw["auths"]["d"]
	assert.Equal(t, "a", entry["username"])
	assert.Equal(t, "b", entry["password"])
	assert.Equal(t, "cc", entry["email"])
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("a:b")), entry["auth"])
	assert.NotContains(t, entry, "Username")
}

func Test_k8sRepo_GetNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8s().Return(&K8sClient{Client: clientset}).AnyTimes()
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
	mockData := NewMockDataStore(m)
	mockData.EXPECT().K8s().Return(&K8sClient{EventLister: NewEventLister(&eventsv1.Event{
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
	mockData := NewMockDataStore(m)
	mockData.EXPECT().K8s().Return(&K8sClient{PodLister: NewPodLister(&corev1.Pod{
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
	mockData := NewMockDataStore(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	mockData.EXPECT().K8s().Return(&K8sClient{PodLister: NewPodLister(&corev1.Pod{
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
	mockData := NewMockDataStore(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	mockData.EXPECT().K8s().Return(&K8sClient{PodLister: NewPodLister(
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
	mockData := NewMockDataStore(m)
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
	mockData := NewMockDataStore(m)
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

	t.Run("should sum multiple pods", func(t *testing.T) {
		// 多 Pod 命中时走 cpu.Add/memory.Add 累加分支（k8s.go 531-537）。
		list := []v1beta1.PodMetrics{
			{Containers: []v1beta1.ContainerMetrics{{Usage: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("100M"),
			}}}},
			{Containers: []v1beta1.ContainerMetrics{{Usage: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("200M"),
			}}}},
		}

		cpuStr, memoryStr := kr.GetCpuAndMemory(context.TODO(), list)

		assert.Equal(t, "300 m", cpuStr)
		assert.Equal(t, "300 MB", memoryStr)
	})
}

func Test_getStatus(t *testing.T) {
	var tests = []struct {
		CpuRate    float64
		MemoryRate float64
		Wants      biz.ClusterStatus
	}{
		{
			CpuRate:    80,
			MemoryRate: 80,
			Wants:      StatusHealth,
		},
		{
			CpuRate:    81,
			MemoryRate: 81,
			Wants:      StatusNotGood,
		},
		{
			CpuRate:    81,
			MemoryRate: 10,
			Wants:      StatusNotGood,
		},
		{
			CpuRate:    10,
			MemoryRate: 80,
			Wants:      StatusHealth,
		},
		{
			CpuRate:    95,
			MemoryRate: 95,
			Wants:      StatusNotGood,
		},
		{
			CpuRate:    96,
			MemoryRate: 96,
			Wants:      StatusBad,
		},
		{
			CpuRate:    10,
			MemoryRate: 96,
			Wants:      StatusBad,
		},
		{
			CpuRate:    96,
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
	mockData := NewMockDataStore(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	mockData.EXPECT().K8s().Return(&K8sClient{Client: fc, MetricsClient: fcm}).AnyTimes()
	info := kr.ClusterInfo()
	assert.Equal(t, &biz.ClusterInfo{
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

// TestClusterInfo_CacheHit 覆盖 30s 缓存的合并读：首次调用实时 List 一次并回填，TTL 内
// 第二次调用命中缓存不重复 List nodes（用 fake reactor 计数验证），且两次结果一致。
func TestClusterInfo_CacheHit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cpu := &resource.Quantity{}
	cpu.Add(resource.MustParse("3"))
	memory := &resource.Quantity{}
	memory.Add(resource.MustParse("10Gi"))
	fc := fake.NewSimpleClientset(&corev1.NodeList{Items: []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node01"},
			Status: corev1.NodeStatus{
				Allocatable: corev1.ResourceList{
					corev1.ResourceCPU:    cpu.DeepCopy(),
					corev1.ResourceMemory: memory.DeepCopy(),
				},
			},
		},
	}})
	// 前置 reactor 只计数不接管，交给内置 ObjectTracker 返回已注册的 NodeList。
	nodeListCalls := 0
	fc.PrependReactor("list", "nodes", func(action testing2.Action) (bool, runtime.Object, error) {
		nodeListCalls++
		return false, nil, nil
	})
	fcm := &fake2.Clientset{}
	fcm.AddReactor("list", "nodes", func(action testing2.Action) (bool, runtime.Object, error) {
		return true, &v1beta1.NodeMetricsList{
			Items: []v1beta1.NodeMetrics{
				{ObjectMeta: metav1.ObjectMeta{Name: "node01"}, Usage: corev1.ResourceList{}},
			},
		}, nil
	})
	mockData := NewMockDataStore(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
		cache:  NewCacheImpl(&config.Config{CacheDriver: "memory"}, nil, mlog.NewForConfig(nil)),
	}
	mockData.EXPECT().K8s().Return(&K8sClient{Client: fc, MetricsClient: fcm}).AnyTimes()

	first := kr.ClusterInfo()
	second := kr.ClusterInfo()
	assert.Equal(t, 1, nodeListCalls, "TTL 内第二次调用应命中缓存，不重复 List nodes")
	assert.Equal(t, first, second)
}

// Nodes().List 失败时不能 panic（nil 解引用），应返回空 biz.ClusterInfo。
func TestClusterInfo_NodesListError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	fc := fake.NewSimpleClientset()
	fc.PrependReactor("list", "nodes", func(action testing2.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("nodes list boom")
	})
	fcm := &fake2.Clientset{}
	mockData := NewMockDataStore(m)
	kr := &k8sRepo{logger: mlog.NewForConfig(nil), data: mockData, cache: NewCacheImpl(&config.Config{}, nil, mlog.NewForConfig(nil))}
	mockData.EXPECT().K8s().Return(&K8sClient{Client: fc, MetricsClient: fcm}).AnyTimes()
	info := kr.ClusterInfo()
	assert.Equal(t, &biz.ClusterInfo{}, info)
}

// NodeMetricses().List 失败时不能 panic，应按空用量继续统计。
func TestClusterInfo_MetricsListError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cpu := &resource.Quantity{}
	cpu.Add(resource.MustParse("3"))
	memory := &resource.Quantity{}
	memory.Add(resource.MustParse("10Gi"))
	fc := fake.NewSimpleClientset(
		&corev1.NodeList{Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "node01"},
				Status: corev1.NodeStatus{
					Allocatable: corev1.ResourceList{
						corev1.ResourceCPU:    cpu.DeepCopy(),
						corev1.ResourceMemory: memory.DeepCopy(),
					},
				},
			},
		}},
	)
	fcm := &fake2.Clientset{}
	fcm.AddReactor("list", "nodes", func(action testing2.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("metrics list boom")
	})
	mockData := NewMockDataStore(m)
	kr := &k8sRepo{logger: mlog.NewForConfig(nil), data: mockData, cache: NewCacheImpl(&config.Config{}, nil, mlog.NewForConfig(nil))}
	mockData.EXPECT().K8s().Return(&K8sClient{Client: fc, MetricsClient: fcm}).AnyTimes()
	info := kr.ClusterInfo()
	// 不 panic，节点信息仍可统计：metrics 失败时 TotalMemory 回退到节点 allocatable（单节点 10Gi）。
	assert.Equal(t, "10 GiB", info.TotalMemory)
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

	mockData := NewMockDataStore(m)
	mockData.EXPECT().K8s().Return(&K8sClient{Client: fc}).AnyTimes()
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
			labels := (&k8sRepo{
				logger: mlog.NewForConfig(nil),
			}).GetPodSelectorsByManifest([]string{tt.in})
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
	mockData := NewMockDataStore(m)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8s().Return(&K8sClient{Client: clientset}).AnyTimes()
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

func TestDeletePod(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8s().Return(&K8sClient{Client: clientset}).AnyTimes()
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	// 捕获实际下发的 DeleteOptions，验证 opts 透传（强制删除策略由 biz 层决定）
	var gotOpts metav1.DeleteOptions
	clientset.PrependReactor("delete", "pods", func(action testing2.Action) (bool, runtime.Object, error) {
		if del, ok := action.(testing2.DeleteAction); ok {
			gotOpts = del.GetDeleteOptions()
		}
		return false, nil, nil
	})
	// 删除不存在的 pod → errs.Wrap 归类为 NotFound
	err := kr.DeletePod(context.TODO(), "a", "p", metav1.DeleteOptions{})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
	// 创建后强制删除 → 成功，GracePeriodSeconds/PropagationPolicy 透传
	clientset.CoreV1().Pods("a").Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "a"},
	}, metav1.CreateOptions{})
	zero, bg := int64(0), metav1.DeletePropagationBackground
	err = kr.DeletePod(context.TODO(), "a", "p", metav1.DeleteOptions{
		GracePeriodSeconds: &zero,
		PropagationPolicy:  &bg,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(0), *gotOpts.GracePeriodSeconds)
	assert.Equal(t, metav1.DeletePropagationBackground, *gotOpts.PropagationPolicy)
	// 删除后 pod 已从 apiserver 移除
	_, err = clientset.CoreV1().Pods("a").Get(context.TODO(), "p", metav1.GetOptions{})
	assert.Error(t, err)
}

func TestDeleteSecret(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8s().Return(&K8sClient{Client: clientset}).AnyTimes()
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

// fakeTerminalSizeQueue 是 TerminalSizeQueue 的桩：第一次 Next 返回固定尺寸，之后返回 nil（流结束）。
type fakeTerminalSizeQueue struct {
	count int
}

// Next 第一次返回固定尺寸，后续返回 nil。
func (f *fakeTerminalSizeQueue) Next() *biz.TerminalSize {
	f.count++
	if f.count == 1 {
		return &biz.TerminalSize{Width: 10, Height: 20}
	}
	return nil
}

// Test_translateExecError 覆盖 translateExecError 三分支：nil 透传、非退出码错误透传、
// CodeExitError 翻译为领域 ExecExitError。
func Test_translateExecError(t *testing.T) {
	assert.Nil(t, translateExecError(nil))

	generic := errors.New("boom")
	assert.Same(t, generic, translateExecError(generic))

	exited := &clientgoexec.CodeExitError{Err: errors.New("boom"), Code: 2}
	translated := translateExecError(exited)
	require.NotNil(t, translated)
	got, ok := translated.(*biz.ExecExitError)
	require.True(t, ok)
	assert.Equal(t, 2, got.Code)
	assert.Equal(t, exited.Error(), got.Message)
}

// Test_toRemotecommandTerminalSizeQueue 覆盖尺寸队列适配：nil 输入返回 nil、
// 领域尺寸转换为 remotecommand 尺寸、领域队列返回 nil 时透传 nil。
func Test_toRemotecommandTerminalSizeQueue(t *testing.T) {
	assert.Nil(t, toRemotecommandTerminalSizeQueue(nil))

	adapter := toRemotecommandTerminalSizeQueue(&fakeTerminalSizeQueue{})
	require.NotNil(t, adapter)
	assert.Equal(t, &remotecommand.TerminalSize{Width: 10, Height: 20}, adapter.Next())
	assert.Nil(t, adapter.Next())
}

func Test_defaultRemoteExecutor_New(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	mockData.EXPECT().K8s().Return(&K8sClient{}).Times(2)
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
	c := &biz.Container{
		Namespace: "a",
		Pod:       "v",
		Container: "c",
	}
	ec.EXPECT().WithContainer(c.Namespace, c.Pod, c.Container).Return(ec)
	ec.EXPECT().WithMethod("POST").Return(ec)
	ec.EXPECT().WithCommand([]string{"ls"}).Return(ec)
	input := &biz.ExecuteInput{
		Cmd:               []string{"ls"},
		TerminalSizeQueue: nil,
	}
	ec.EXPECT().Execute(gomock.Any(), input)
	assert.Nil(t, r.Execute(context.TODO(), c, input))
}

func Test_defaultRemoteExecutor_NewFileCopy(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	d := &defaultRemoteExecutor{
		data:   mockData,
		logger: mlog.NewForConfig(nil),
	}
	shared := &restclient.Config{}
	mockData.EXPECT().K8s().Return(&K8sClient{
		RestConfig: shared,
	}).Times(2)
	fileCopy := d.NewFileCopy(1, &bytes.Buffer{})
	assert.NotNil(t, fileCopy)
	// 共享 config 不能被改写：旧实现直接在原指针上改 APIPath/GroupVersion/
	// NegotiatedSerializer，会污染后续所有 executor 的请求路径。
	assert.Nil(t, shared.GroupVersion)
	assert.Empty(t, shared.APIPath)
	assert.Nil(t, shared.NegotiatedSerializer)
}
func TestGetPodLogs(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8s().Return(&K8sClient{Client: clientset}).AnyTimes()
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
	mockData := NewMockDataStore(m)
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	manager := NewMockExecutorManager(m)
	mockExecutor := NewMockExecutor(m)
	manager.EXPECT().New().Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithCommand(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithContainer("test-namespace", "test-pod", "").Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithMethod(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*biz.ExecuteInput)
		input.Stdout.Write([]byte("0"))
		return slices.Equal(input.Cmd, []string{"sh", "-c", "test -f /test/file/path && echo 1 || echo 0"})
	})).Return(nil).AnyTimes()
	kr := &k8sRepo{
		logger:   mlog.NewForConfig(nil),
		data:     mockData,
		uploader: mockUploader,
		fileRepo: mockFileRepo,
		executor: manager,
	}

	_, err := kr.CopyFromPod(context.TODO(), &biz.CopyFromPodInput{
		Namespace: "test-namespace",
		Pod:       "test-pod",
		FilePath:  "/test/file/path",
		UserName:  "test-user",
	})
	s, _ := status.FromError(err)
	assert.Equal(t, "下载内容必须是文件: test-pod /test/file/path", s.Message())
	assert.Equal(t, codes.InvalidArgument, s.Code())
}

func TestCopyFromPod1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	manager := NewMockExecutorManager(m)
	mockExecutor := NewMockExecutor(m)
	manager.EXPECT().New().Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithCommand(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithContainer("test-namespace", "test-pod", "").Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithMethod(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*biz.ExecuteInput)
		input.Stdout.Write([]byte("1"))
		return slices.Equal(input.Cmd, []string{"sh", "-c", "test -f /test/file/path && echo 1 || echo 0"})
	})).Return(nil)
	kr := &k8sRepo{
		logger:   mlog.NewForConfig(nil),
		data:     mockData,
		uploader: mockUploader,
		timer:    timer.NewReal(),
		fileRepo: mockFileRepo,
		executor: manager,
	}

	manager.EXPECT().NewFileCopy(5, gomock.Any()).Return(nil)
	mockUploader.EXPECT().Disk("podfile").Return(mockUploader)
	mockUploader.EXPECT().NewFile(gomock.Any()).Return(nil, errors.New("x"))

	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*biz.ExecuteInput)
		input.Stdout.Write([]byte("/"))
		return slices.Equal(input.Cmd, []string{"sh", "-c", "pwd"})
	})).Return(nil)

	_, err := kr.CopyFromPod(context.TODO(), &biz.CopyFromPodInput{
		Namespace: "test-namespace",
		Pod:       "test-pod",
		FilePath:  "/test/file/path",
		UserName:  "test-user",
	})
	assert.Contains(t, err.Error(), "x")
}

func TestCopyFromPod_success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	manager := NewMockExecutorManager(m)
	mockExecutor := NewMockExecutor(m)
	manager.EXPECT().New().Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithCommand(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithContainer("test-namespace", "test-pod", "").Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithMethod(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*biz.ExecuteInput)
		input.Stdout.Write([]byte("1"))
		return slices.Equal(input.Cmd, []string{"sh", "-c", "test -f /test/file/path && echo 1 || echo 0"})
	})).Return(nil)
	kr := &k8sRepo{
		logger:   mlog.NewForConfig(nil),
		data:     mockData,
		uploader: mockUploader,
		timer:    timer.NewReal(),
		fileRepo: mockFileRepo,
		executor: manager,
	}

	manager.EXPECT().NewFileCopy(5, gomock.Any()).Return(&mockFileCopy{})
	mockUploader.EXPECT().Disk("podfile").Return(mockUploader)
	file := uploader.NewMockFile(m)
	mockUploader.EXPECT().NewFile(gomock.Any()).Return(file, nil)

	mockExecutor.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*biz.ExecuteInput)
		input.Stdout.Write([]byte("/"))
		return slices.Equal(input.Cmd, []string{"sh", "-c", "pwd"})
	})).Return(nil)

	info := &mockFileInfo{
		size: 1,
	}
	file.EXPECT().Stat().Return(info, nil)
	file.EXPECT().Close()
	file.EXPECT().Name().Return("fname")

	mockFileRepo.EXPECT().Create(gomock.Any(), &biz.CreateFileInput{
		Path:       "fname",
		Username:   "test-user",
		Size:       uint64(1),
		UploadType: schematype.Local,
		Namespace:  "test-namespace",
		Pod:        "test-pod",
		Container:  "",
	})

	mockUploader.EXPECT().Type().Return(schematype.Local)
	_, err := kr.CopyFromPod(context.TODO(), &biz.CopyFromPodInput{
		Namespace: "test-namespace",
		Pod:       "test-pod",
		FilePath:  "/test/file/path",
		UserName:  "test-user",
	})
	assert.Nil(t, err)
}

var _ k8sutil.FileCopy = (*mockFileCopy)(nil)

type mockFileInfo struct {
	size int64
}

func (m *mockFileInfo) Name() string {
	return ""
}

func (m *mockFileInfo) Size() int64 {
	return m.size
}

func (m *mockFileInfo) Mode() fs.FileMode {
	return fs.FileMode(0644)
}

func (m *mockFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (m *mockFileInfo) IsDir() bool {
	return false
}

func (m *mockFileInfo) Sys() any {
	return nil
}

type mockFileCopy struct{}

func (m *mockFileCopy) CopyFromPod(ctx context.Context, src k8sutil.CopyFileSpec, file uploader.File) error {
	return nil
}

// TestK8sRepo_UpdateSecret 覆盖更新指定 secret 内容的端口（cron TLS 同步用）。
func TestK8sRepo_UpdateSecret(t *testing.T) {
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "s", Namespace: "default"},
		Data:       map[string][]byte{"tls.crt": []byte("old")},
	}
	d := NewDataImpl(&NewDataParams{
		Cfg:       &config.Config{},
		K8sClient: &K8sClient{Client: fake.NewSimpleClientset(sec)},
	})
	repo := &k8sRepo{data: d}

	sec.Data["tls.crt"] = []byte("new")
	got, err := repo.UpdateSecret(context.TODO(), "default", "s", sec)
	assert.NoError(t, err)
	assert.Equal(t, "new", string(got.Data["tls.crt"]))
}

// TestK8sRepo_CreateDockerSecrets 覆盖按 servers 子集创建 docker secret 的端口：
// 只含命中 server 的凭据，config 不泄漏进业务层。
func TestK8sRepo_CreateDockerSecrets(t *testing.T) {
	d := NewDataImpl(&NewDataParams{
		Cfg: &config.Config{ImagePullSecrets: config.DockerAuths{
			{Server: "reg.io", Username: "u", Password: "p", Email: "e"},
			{Server: "other.io", Username: "u2", Password: "p2", Email: "e2"},
		}},
		K8sClient: &K8sClient{Client: fake.NewSimpleClientset()},
	})
	repo := &k8sRepo{data: d}

	secret, err := repo.CreateDockerSecrets(context.TODO(), "default", []string{"reg.io"})
	assert.NoError(t, err)
	assert.Equal(t, corev1.SecretTypeDockerConfigJson, secret.Type)

	var dc biz.DockerConfigJSON
	assert.NoError(t, json.Unmarshal(secret.Data[corev1.DockerConfigJsonKey], &dc))
	assert.Len(t, dc.Auths, 1)
	assert.Contains(t, dc.Auths, "reg.io")
	assert.NotContains(t, dc.Auths, "other.io")
}

// TestK8sRepo_SubscribePodEvents 覆盖 Pod 事件订阅端口的转换链路：
// informer fanout 的 Obj 泛型转换为领域 PodEvent，取消订阅关闭事件通道。
func TestK8sRepo_SubscribePodEvents(t *testing.T) {
	input := make(chan Obj[*corev1.Pod], 4)
	fan := newFanOut[*corev1.Pod](mlog.NewForConfig(nil), "pod", input, map[string]chan<- Obj[*corev1.Pod]{})
	d := NewDataImpl(&NewDataParams{
		Cfg:       &config.Config{},
		K8sClient: &K8sClient{podFanOut: fan},
	})
	repo := &k8sRepo{data: d}

	ch, unsubscribe := repo.SubscribePodEvents("pod-watcher")
	done := make(chan struct{})
	defer close(done)
	go fan.Distribute(done)

	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "default"}}
	input <- newObj[*corev1.Pod](nil, pod, Update)

	ev := <-ch
	assert.Equal(t, biz.PodEventUpdate, ev.Type)
	assert.Equal(t, "p", ev.Current.Name)
	assert.Nil(t, ev.Old)

	// 取消订阅关闭事件通道，消费方 range 幂等退出。
	unsubscribe()
	_, ok := <-ch
	assert.False(t, ok)
}

// TestIsPodRunning_WaitingAndNotRunning 补齐 IsPodRunning 的两个剩余分支：
// 容器处于 Waiting 态（取 Waiting.Reason/Message）与无容器状态（"pod not running."）。
func TestIsPodRunning_WaitingAndNotRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	mockData.EXPECT().K8s().Return(&K8sClient{PodLister: NewPodLister(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "waiting", Namespace: "a"},
			Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c"}}},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
				ContainerStatuses: []corev1.ContainerStatus{{
					State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff", Message: "backoff"}},
				}},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "nostatus", Namespace: "a"},
			Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c"}}},
			Status:     corev1.PodStatus{Phase: corev1.PodPending},
		},
	)}).AnyTimes()

	running, reason := kr.IsPodRunning("a", "waiting")
	assert.False(t, running)
	assert.Equal(t, "CrashLoopBackOff backoff", reason)

	running, reason = kr.IsPodRunning("a", "nostatus")
	assert.False(t, running)
	assert.Equal(t, "pod not running.", reason)
}

// TestFindDefaultContainer_NoContainers 补齐零容器时返回 "未找到容器" 的分支。
func TestFindDefaultContainer_NoContainers(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	mockData.EXPECT().K8s().Return(&K8sClient{PodLister: NewPodLister(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "nocontainers", Namespace: "a"},
		},
	)}).AnyTimes()

	_, err := kr.FindDefaultContainer(context.TODO(), "a", "nocontainers")
	assert.Error(t, err)
	// errs.NotFound 映射为 grpc NotFound：断言状态码与消息，而非 Error() 全串（带 grpc 前缀）。
	assert.Equal(t, codes.NotFound, status.Code(err))
	assert.Equal(t, "未找到容器: a/nocontainers", status.Convert(err).Message())
}

// TestListPodsBySelectors 覆盖按 selector 列 Pod：命中、跨 selector 去重、
// 非法 selector 解析跳过、informer List 失败跳过。LabelSelectorAsSelector
// 错误分支不可达（入参恒来自 ParseToLabelSelector，见实现处注释）。
func TestListPodsBySelectors(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	mockData.EXPECT().K8s().Return(&K8sClient{PodLister: NewPodLister(
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "web-1", Namespace: "a", Labels: map[string]string{"app": "web"}}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "db-1", Namespace: "a", Labels: map[string]string{"role": "db"}}},
	)}).AnyTimes()

	// 单 selector 命中。
	pods, err := kr.ListPodsBySelectors("a", []string{"app=web"})
	assert.NoError(t, err)
	assert.Len(t, pods, 1)
	assert.Equal(t, "web-1", pods[0].Name)

	// 两个 selector 命中同一 Pod → 去重。
	pods, err = kr.ListPodsBySelectors("a", []string{"app=web", "app in (web, other)"})
	assert.NoError(t, err)
	assert.Len(t, pods, 1)

	// 非法 selector 解析失败 → 跳过，不报错。
	pods, err = kr.ListPodsBySelectors("a", []string{"%%%bad", "app=web"})
	assert.NoError(t, err)
	assert.Len(t, pods, 1)

	// 无匹配 → 空结果不报错。
	pods, err = kr.ListPodsBySelectors("a", []string{"none=missing"})
	assert.NoError(t, err)
	assert.Len(t, pods, 0)

	// informer List 返回错误 → 该 selector 跳过，整体不报错。
	mockErr := NewMockDataStore(m)
	mockErr.EXPECT().K8s().Return(&K8sClient{PodLister: errPodLister{}}).AnyTimes()
	krErr := &k8sRepo{logger: mlog.NewForConfig(nil), data: mockErr}
	pods, err = krErr.ListPodsBySelectors("a", []string{"app=web", "app in (web, other)"})
	assert.NoError(t, err)
	assert.Len(t, pods, 0)
}

// errPodNamespaceLister 是 List 恒报错的 PodNamespaceLister 替身，
// 用于覆盖 ListPodsBySelectors 的 informer 查询失败分支。
type errPodNamespaceLister struct{}

func (errPodNamespaceLister) List(_ labels.Selector) ([]*corev1.Pod, error) {
	return nil, errors.New("list boom")
}
func (errPodNamespaceLister) Get(string) (*corev1.Pod, error) {
	return nil, errors.New("get boom")
}

// errPodLister 是 PodLister 替身，Pods 返回恒报错的命名空间级 lister。
type errPodLister struct{}

func (errPodLister) List(_ labels.Selector) ([]*corev1.Pod, error) {
	return nil, errors.New("list boom")
}
func (errPodLister) Pods(string) corev1lister.PodNamespaceLister {
	return errPodNamespaceLister{}
}

// newMetricsFake 构造带 get/list 反应器的 metrics fake。
// fake 的 ObjectTracker 以 scheme 默认资源 "podmetrics" 建 key，而 typed client
// 查询的是 "pods.metrics.k8s.io"，两者错位导致 "not found"；用 reactor 直供数据。
func newMetricsFake(pms ...*v1beta1.PodMetrics) *fake2.Clientset {
	mc := fake2.NewSimpleClientset()
	mc.PrependReactor("get", "pods", func(action testing2.Action) (bool, runtime.Object, error) {
		ga := action.(testing2.GetAction)
		for _, pm := range pms {
			if pm.Name == ga.GetName() {
				return true, pm, nil
			}
		}
		return true, nil, errors.New("pods.metrics.k8s.io not found")
	})
	mc.PrependReactor("list", "pods", func(action testing2.Action) (bool, runtime.Object, error) {
		items := make([]v1beta1.PodMetrics, 0, len(pms))
		for _, pm := range pms {
			items = append(items, *pm)
		}
		return true, &v1beta1.PodMetricsList{Items: items}, nil
	})
	return mc
}

// TestK8sRepo_AddTlsSecret 覆盖创建 TLS secret 的端口。
func TestK8sRepo_AddTlsSecret(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	clientset := fake.NewSimpleClientset()
	mockData.EXPECT().K8s().Return(&K8sClient{Client: clientset}).AnyTimes()
	kr := &k8sRepo{data: mockData}

	sec, err := kr.AddTlsSecret("ns", "tls-name", "keydata", "crtdata")
	require.NoError(t, err)
	assert.Equal(t, "tls-name", sec.Name)
	assert.Equal(t, corev1.SecretTypeTLS, sec.Type)
	assert.Equal(t, "mars", sec.Annotations["created-by"])

	got, err := clientset.CoreV1().Secrets("ns").Get(context.TODO(), "tls-name", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "keydata", got.StringData["tls.key"])
	assert.Equal(t, "crtdata", got.StringData["tls.crt"])
}

// TestK8sRepo_GetPodMetrics 覆盖单 Pod 指标查询端口。
func TestK8sRepo_GetPodMetrics(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	pm := &v1beta1.PodMetrics{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns"}}
	mc := newMetricsFake(pm)
	mockData.EXPECT().K8s().Return(&K8sClient{MetricsClient: mc}).AnyTimes()
	kr := &k8sRepo{data: mockData}

	got, err := kr.GetPodMetrics(context.TODO(), "ns", "p1")
	require.NoError(t, err)
	assert.Equal(t, "p1", got.Name)

	_, err = kr.GetPodMetrics(context.TODO(), "ns", "missing")
	assert.Error(t, err)
}

// TestK8sRepo_GetAllPodMetrics 覆盖按 PodSelectors 聚合集群指标：
// 空选择器返回 nil、命中聚合、selector 无匹配跳过。
func TestK8sRepo_GetAllPodMetrics(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "web-1", Namespace: "ns", Labels: map[string]string{"app": "web"}}}
	mc := newMetricsFake(&v1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{Name: "web-1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		Containers: []v1beta1.ContainerMetrics{{Name: "c", Usage: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("100m"),
		}}},
	})
	mockData.EXPECT().K8s().Return(&K8sClient{
		MetricsClient: mc,
		PodLister:     NewPodLister(pod),
	}).AnyTimes()
	kr := &k8sRepo{data: mockData, logger: mlog.NewForConfig(nil)}

	// 无 PodSelectors → nil。
	assert.Nil(t, kr.GetAllPodMetrics(context.TODO(), &biz.Project{Namespace: &biz.Namespace{Name: "ns"}}))

	// 命中 selector → 聚合出 1 条。
	list := kr.GetAllPodMetrics(context.TODO(), &biz.Project{
		Namespace:    &biz.Namespace{Name: "ns"},
		PodSelectors: []string{"app=web"},
	})
	assert.Len(t, list, 1)
	assert.Equal(t, "web-1", list[0].Name)

	// selector 无匹配 → 跳过，空结果。
	list = kr.GetAllPodMetrics(context.TODO(), &biz.Project{
		Namespace:    &biz.Namespace{Name: "ns"},
		PodSelectors: []string{"app=missing"},
	})
	assert.Len(t, list, 0)
}

// TestK8sRepo_GetAllPodMetrics_SelectorError 覆盖 selector 解析失败跳过分支（k8s.go 257-260）。
func TestK8sRepo_GetAllPodMetrics_SelectorError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	mockData.EXPECT().K8s().Return(&K8sClient{
		MetricsClient: newMetricsFake(),
		PodLister:     NewPodLister(),
	}).AnyTimes()
	kr := &k8sRepo{data: mockData, logger: mlog.NewForConfig(nil)}

	list := kr.GetAllPodMetrics(context.TODO(), &biz.Project{
		Namespace:    &biz.Namespace{Name: "ns"},
		PodSelectors: []string{"invalid"},
	})
	assert.Len(t, list, 0)
}

// TestK8sRepo_GetAllPodMetrics_ListerError 覆盖 PodLister 查询失败跳过分支（k8s.go 263-265）。
func TestK8sRepo_GetAllPodMetrics_ListerError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	mockData.EXPECT().K8s().Return(&K8sClient{
		MetricsClient: newMetricsFake(),
		PodLister:     errPodLister{},
	}).AnyTimes()
	kr := &k8sRepo{data: mockData, logger: mlog.NewForConfig(nil)}

	list := kr.GetAllPodMetrics(context.TODO(), &biz.Project{
		Namespace:    &biz.Namespace{Name: "ns"},
		PodSelectors: []string{"app=web"},
	})
	assert.Len(t, list, 0)
}

// TestK8sRepo_GetAllPodMetrics_MetricsError 覆盖 metrics 查询失败跳过分支（k8s.go 273-275）。
func TestK8sRepo_GetAllPodMetrics_MetricsError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "web-1", Namespace: "ns", Labels: map[string]string{"app": "web"}}}
	failing := fake2.NewSimpleClientset()
	failing.PrependReactor("list", "pods", func(action testing2.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("metrics list boom")
	})
	mockData.EXPECT().K8s().Return(&K8sClient{
		MetricsClient: failing,
		PodLister:     NewPodLister(pod),
	}).AnyTimes()
	kr := &k8sRepo{data: mockData, logger: mlog.NewForConfig(nil)}

	list := kr.GetAllPodMetrics(context.TODO(), &biz.Project{
		Namespace:    &biz.Namespace{Name: "ns"},
		PodSelectors: []string{"app=web"},
	})
	assert.Len(t, list, 0)
}

// TestK8sRepo_GetCpuAndMemoryInNamespace 覆盖命名空间 CPU/内存汇总端口。
func TestK8sRepo_GetCpuAndMemoryInNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	mc := newMetricsFake(&v1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "ns"},
		Containers: []v1beta1.ContainerMetrics{{Name: "c", Usage: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("100Mi"),
		}}},
	})
	mockData.EXPECT().K8s().Return(&K8sClient{MetricsClient: mc}).AnyTimes()
	kr := &k8sRepo{data: mockData}

	cpu, mem := kr.GetCpuAndMemoryInNamespace(context.TODO(), "ns")
	assert.Equal(t, "100 m", cpu)
	assert.NotEmpty(t, mem)
}

// TestDefaultArchiver 覆盖归档/打开/删除三个端口的真实文件操作。
func TestDefaultArchiver(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	require.NoError(t, os.WriteFile(src, []byte("hi"), 0o644))
	dst := filepath.Join(dir, "a.tar.gz")

	ar := NewDefaultArchiver()
	require.NoError(t, ar.Archive([]string{src}, dst))
	rc, err := ar.Open(dst)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, rc)
	require.NoError(t, err)
	require.NoError(t, rc.Close())
	require.NoError(t, ar.Remove(dst))
	_, err = os.Stat(dst)
	assert.Error(t, err)
}

// errFileCopy 是 CopyFromPod 恒报错的 FileCopy 替身，
// 用于覆盖 CopyFromPod 的归档复制失败分支。
type errFileCopy struct{ err error }

func (e *errFileCopy) CopyFromPod(_ context.Context, _ k8sutil.CopyFileSpec, _ uploader.File) error {
	return e.err
}

// errArchiver 是 copyToPod 归档接口的故障替身：archiveErr/openErr 非空即报错，
// 用于覆盖归档/打开失败分支。
type errArchiver struct {
	archiveErr error
	openErr    error
}

func (e *errArchiver) Archive(_ []string, _ string) error { return e.archiveErr }
func (e *errArchiver) Open(_ string) (io.ReadCloser, error) {
	if e.openErr != nil {
		return nil, e.openErr
	}
	return io.NopCloser(&errReader{}), nil
}
func (e *errArchiver) Remove(_ string) error { return nil }

// errReader 是读即报错的 io.Reader，供 errArchiver.Open 在 io.Copy 分支
// 触发源读取失败。
type errReader struct{}

func (e *errReader) Read(_ []byte) (int, error) { return 0, errors.New("read boom") }

// copyFromPodK8sRepo 构造覆盖 CopyFromPod 错误分支所需的 k8sRepo。
// 三个 Execute 调用统一用 AnyTimes 的 executor mock，由 gomock.Cond 区分命令。
func copyFromPodK8sRepo(t *testing.T, m *gomock.Controller, mockUploader *uploader.MockUploader, mockFileRepo *MockFileRepo) (*k8sRepo, *MockExecutorManager, *MockExecutor) {
	t.Helper()
	mockData := NewMockDataStore(m)
	manager := NewMockExecutorManager(m)
	mockExecutor := NewMockExecutor(m)
	mockData.EXPECT().K8s().Return(&K8sClient{}).AnyTimes()
	manager.EXPECT().New().Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithCommand(gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockExecutor).AnyTimes()
	mockExecutor.EXPECT().WithMethod(gomock.Any()).Return(mockExecutor).AnyTimes()
	return &k8sRepo{
		logger:   mlog.NewForConfig(nil),
		data:     mockData,
		uploader: mockUploader,
		fileRepo: mockFileRepo,
		executor: manager,
		timer:    timer.NewReal(),
	}, manager, mockExecutor
}

// execOutput 构造一个 Execute 期望：向 input.Stdout 写入 out 并返回 err，
// cmd 匹配指定命令序列。用于区分 CopyFromPod 内 test -f / pwd 两次 Execute。
func execOutput(ec *MockExecutor, cmd []string, out string, err error) {
	ec.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*biz.ExecuteInput)
		input.Stdout.Write([]byte(out))
		return slices.Equal(input.Cmd, cmd)
	})).Return(err)
}

// execDrain 构造 copyToPod 使用的 Execute 期望：copyToPod 用 io.Pipe 把归档流喂给
// Execute 的 Stdin，mock 必须消费 Stdin，否则 goroutine 的 io.Copy 阻塞在 pipe 写端，
// defer wg.Wait() 死等挂起测试。input.Cmd 恒空（命令存 executor 内部字段），不匹配。
func execDrain(ec *MockExecutor) {
	ec.EXPECT().Execute(gomock.Any(), gomock.Cond(func(x any) bool {
		input := x.(*biz.ExecuteInput)
		if input.Stdin != nil {
			_, _ = io.Copy(io.Discard, input.Stdin)
		}
		return true
	})).Return(nil)
}

// TestCopyFromPod_pwdError 覆盖 pwd Execute 失败的错误分支。
func TestCopyFromPod_pwdError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, ec := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)

	execOutput(ec, []string{"sh", "-c", "test -f /p && echo 1 || echo 0"}, "1", nil)
	execOutput(ec, []string{"sh", "-c", "pwd"}, "", errors.New("pwd boom"))

	_, err := kr.CopyFromPod(context.TODO(), &biz.CopyFromPodInput{
		Namespace: "ns", Pod: "p", FilePath: "/p", UserName: "u",
	})
	assert.ErrorContains(t, err, "pwd boom")
}

// TestCopyFromPod_invalidPath 覆盖 FilePath 不在 pwd 前缀之下的错误分支
// pwd 返回相对路径 "a"，绝对路径 /p 前缀不匹配。
func TestCopyFromPod_invalidPath(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, ec := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)

	execOutput(ec, []string{"sh", "-c", "test -f /p && echo 1 || echo 0"}, "1", nil)
	execOutput(ec, []string{"sh", "-c", "pwd"}, "a", nil)

	_, err := kr.CopyFromPod(context.TODO(), &biz.CopyFromPodInput{
		Namespace: "ns", Pod: "p", FilePath: "/p", UserName: "u",
	})
	s, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, s.Code())
	assert.Equal(t, "非法文件路径: p /p", s.Message())
}

// TestCopyFromPod_fileCopyError 覆盖归档复制失败分支：err 非 nil 触发
// defer 里的 up.Delete(file.Name())并返回错误。
func TestCopyFromPod_fileCopyError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, manager, ec := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)

	execOutput(ec, []string{"sh", "-c", "test -f /p && echo 1 || echo 0"}, "1", nil)
	execOutput(ec, []string{"sh", "-c", "pwd"}, "/", nil)

	file := uploader.NewMockFile(m)
	manager.EXPECT().NewFileCopy(5, gomock.Any()).Return(&errFileCopy{err: errors.New("copy boom")})
	mockUploader.EXPECT().Disk("podfile").Return(mockUploader)
	mockUploader.EXPECT().NewFile(gomock.Any()).Return(file, nil)
	// defer：err != nil → up.Delete(file.Name())。
	file.EXPECT().Close()
	file.EXPECT().Name().Return("/tmp/out.tar")
	mockUploader.EXPECT().Delete("/tmp/out.tar").Return(nil)

	_, err := kr.CopyFromPod(context.TODO(), &biz.CopyFromPodInput{
		Namespace: "ns", Pod: "p", FilePath: "/p", UserName: "u",
	})
	assert.ErrorContains(t, err, "copy from pod error")
	assert.ErrorContains(t, err, "copy boom")
}

// TestCopyFromPod_statError 覆盖归档成功后 file.Stat 失败的错误分支。
func TestCopyFromPod_statError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, manager, ec := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)

	execOutput(ec, []string{"sh", "-c", "test -f /p && echo 1 || echo 0"}, "1", nil)
	execOutput(ec, []string{"sh", "-c", "pwd"}, "/", nil)

	file := uploader.NewMockFile(m)
	manager.EXPECT().NewFileCopy(5, gomock.Any()).Return(&mockFileCopy{})
	mockUploader.EXPECT().Disk("podfile").Return(mockUploader)
	mockUploader.EXPECT().NewFile(gomock.Any()).Return(file, nil)
	file.EXPECT().Stat().Return(nil, errors.New("stat boom"))
	// defer：err != nil → up.Delete(file.Name())。
	file.EXPECT().Close()
	file.EXPECT().Name().Return("/tmp/out.tar")
	mockUploader.EXPECT().Delete("/tmp/out.tar").Return(nil)

	_, err := kr.CopyFromPod(context.TODO(), &biz.CopyFromPodInput{
		Namespace: "ns", Pod: "p", FilePath: "/p", UserName: "u",
	})
	assert.ErrorContains(t, err, "stat boom")
}

// TestK8sRepo_CopyFileToPod_GetError 覆盖 CopyFileToPod 先查文件记录失败的分支
func TestK8sRepo_CopyFileToPod_GetError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	fileRepo := NewMockFileRepo(m)
	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(nil, errors.New("x"))
	kr := &k8sRepo{data: mockData, fileRepo: fileRepo}
	_, err := kr.CopyFileToPod(context.TODO(), &biz.CopyFileToPodInput{FileId: 1})
	assert.EqualError(t, err, "x")
}

// TestK8sRepo_CopyFileToPod 覆盖 CopyFileToPod 全链路：真实归档 + mock
// executor/uploader，落到 fileRepo.Update。
func TestK8sRepo_CopyFileToPod(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, ec := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)
	kr.archiver = NewDefaultArchiver()
	kr.maxUploadSize = 1 << 20

	// copyToPod：uploader.Stat 返回小文件信息，Local 类型跳过下载分支。
	// 注：LocalUploader()无条件调用，Local 分支实际不消费其返回值，
	// 但 gomock 仍须给出期望。
	info := uploader.NewMockFileInfo(m)
	info.EXPECT().Size().Return(uint64(8)).AnyTimes()
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "a.txt")
	require.NoError(t, os.WriteFile(srcPath, []byte("hi"), 0o644))
	mockUploader.EXPECT().LocalUploader().Return(uploader.NewMockUploader(m))
	mockUploader.EXPECT().Stat(srcPath).Return(info, nil)
	mockUploader.EXPECT().Type().Return(schematype.Local)

	execDrain(ec)

	mockFileRepo.EXPECT().GetByID(gomock.Any(), 7).Return(&biz.File{Path: srcPath}, nil)
	mockFileRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(&biz.File{}, nil)

	got, err := kr.CopyFileToPod(context.TODO(), &biz.CopyFileToPodInput{
		FileId: 7, Namespace: "ns", Pod: "p", Container: "c",
	})
	require.NoError(t, err)
	assert.NotNil(t, got)
}

// TestK8sRepo_CopyToPod_SizeExceeded 覆盖 copyToPod 超过 maxUploadSize 的错误分支
// 文件大小超限直接返回，不进入下载/打包。
func TestK8sRepo_CopyToPod_SizeExceeded(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, _ := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)
	kr.maxUploadSize = 100

	info := uploader.NewMockFileInfo(m)
	info.EXPECT().Size().Return(uint64(1000)).AnyTimes()
	srcPath := "/some/big/file"
	mockUploader.EXPECT().LocalUploader().Return(mockUploader)
	mockUploader.EXPECT().Stat(srcPath).Return(info, nil)
	mockFileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{Path: srcPath}, nil)

	_, err := kr.CopyFileToPod(context.TODO(), &biz.CopyFileToPodInput{FileId: 1})
	assert.ErrorContains(t, err, "最大不得超过")
}

// TestK8sRepo_CopyToPod_NonLocalReadError 覆盖非 Local 类型下载远程文件失败的错误分支
// uploader.Read 返回错误，拷贝在进入打包前终止。
func TestK8sRepo_CopyToPod_NonLocalReadError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, _ := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)
	kr.maxUploadSize = 1 << 20

	info := uploader.NewMockFileInfo(m)
	info.EXPECT().Size().Return(uint64(8)).AnyTimes()
	srcPath := "/remote/file"
	mockUploader.EXPECT().LocalUploader().Return(mockUploader)
	mockUploader.EXPECT().Stat(srcPath).Return(info, nil)
	mockUploader.EXPECT().Type().Return(schematype.S3)
	mockUploader.EXPECT().Read(srcPath).Return(nil, errors.New("read boom"))
	mockFileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{Path: srcPath}, nil)

	_, err := kr.CopyFileToPod(context.TODO(), &biz.CopyFileToPodInput{FileId: 1})
	assert.ErrorContains(t, err, "read boom")
}

// TestK8sRepo_CopyToPod_NonLocalPutError 覆盖非 Local 类型 Put 到本地失败的错误分支
// Read 成功但 Exists 不存在，Put 报错即返回。
func TestK8sRepo_CopyToPod_NonLocalPutError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, _ := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)
	kr.maxUploadSize = 1 << 20

	info := uploader.NewMockFileInfo(m)
	info.EXPECT().Size().Return(uint64(8)).AnyTimes()
	srcPath := "/remote/file"
	read := uploader.NewMockFile(m)
	read.EXPECT().Close()
	gomock.InOrder(
		mockUploader.EXPECT().LocalUploader().Return(mockUploader),
		mockUploader.EXPECT().Stat(srcPath).Return(info, nil),
		mockUploader.EXPECT().Type().Return(schematype.S3),
		mockUploader.EXPECT().Read(srcPath).Return(read, nil),
		mockUploader.EXPECT().Exists(srcPath).Return(false),
		mockUploader.EXPECT().Put(srcPath, read).Return(nil, errors.New("put boom")),
	)
	mockFileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{Path: srcPath}, nil)

	_, err := kr.CopyFileToPod(context.TODO(), &biz.CopyFileToPodInput{FileId: 1})
	assert.ErrorContains(t, err, "put boom")
}

// TestK8sRepo_CopyToPod_NonLocal 覆盖非 Local 类型下载到本地的完整链路
// Read → Exists→Delete → Put → 本地打包 → 上传容器。
// localUploader 与 uploader 同 mock：本地落地路径需真实文件供归档读取。
func TestK8sRepo_CopyToPod_NonLocal(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, ec := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)
	kr.archiver = NewDefaultArchiver()
	kr.maxUploadSize = 1 << 20

	info := uploader.NewMockFileInfo(m)
	info.EXPECT().Size().Return(uint64(8)).AnyTimes()
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "a.txt")
	require.NoError(t, os.WriteFile(srcPath, []byte("hi"), 0o644))
	localPath := filepath.Join(dir, "local.txt")
	require.NoError(t, os.WriteFile(localPath, []byte("hi"), 0o644))

	read := uploader.NewMockFile(m)
	read.EXPECT().Close()
	put := uploader.NewMockFileInfo(m)
	put.EXPECT().Path().Return(localPath).AnyTimes()

	// copyToPod 顺序：LocalUploader → Stat → Size → Type != Local → Read →
	// Exists→Delete(旧) → Put → localPath=put.Path() → defer Delete(新) →
	// 归档 → executor。
	gomock.InOrder(
		mockUploader.EXPECT().LocalUploader().Return(mockUploader),
		mockUploader.EXPECT().Stat(srcPath).Return(info, nil),
		mockUploader.EXPECT().Type().Return(schematype.S3),
		mockUploader.EXPECT().Read(srcPath).Return(read, nil),
		mockUploader.EXPECT().Exists(srcPath).Return(true),
		mockUploader.EXPECT().Delete(srcPath).Return(nil),
		mockUploader.EXPECT().Put(srcPath, read).Return(put, nil),
		mockUploader.EXPECT().Delete(localPath).Return(nil),
	)
	execDrain(ec)
	mockFileRepo.EXPECT().GetByID(gomock.Any(), 7).Return(&biz.File{Path: srcPath}, nil)
	mockFileRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(&biz.File{}, nil)

	got, err := kr.CopyFileToPod(context.TODO(), &biz.CopyFileToPodInput{
		FileId: 7, Namespace: "ns", Pod: "p", Container: "c",
	})
	require.NoError(t, err)
	assert.NotNil(t, got)
}

// TestK8sRepo_CopyToPod_StatError 覆盖 uploader.Stat 失败的错误分支。
func TestK8sRepo_CopyToPod_StatError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, _ := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)
	kr.maxUploadSize = 1 << 20

	srcPath := "/some/file"
	mockUploader.EXPECT().LocalUploader().Return(mockUploader)
	mockUploader.EXPECT().Stat(srcPath).Return(nil, errors.New("stat boom"))
	mockFileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{Path: srcPath}, nil)

	_, err := kr.CopyFileToPod(context.TODO(), &biz.CopyFileToPodInput{FileId: 1})
	assert.ErrorContains(t, err, "stat boom")
}

// TestK8sRepo_CopyToPod_ArchiveError 覆盖归档失败的错误分支。
func TestK8sRepo_CopyToPod_ArchiveError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, _ := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)
	kr.archiver = &errArchiver{archiveErr: errors.New("archive boom")}
	kr.maxUploadSize = 1 << 20

	info := uploader.NewMockFileInfo(m)
	info.EXPECT().Size().Return(uint64(8)).AnyTimes()
	srcPath := "/some/file"
	mockUploader.EXPECT().LocalUploader().Return(mockUploader)
	mockUploader.EXPECT().Stat(srcPath).Return(info, nil)
	mockUploader.EXPECT().Type().Return(schematype.Local)
	mockFileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{Path: srcPath}, nil)

	_, err := kr.CopyFileToPod(context.TODO(), &biz.CopyFileToPodInput{FileId: 1})
	assert.ErrorContains(t, err, "archive boom")
}

// TestK8sRepo_CopyToPod_OpenError 覆盖归档打开失败的错误分支。
func TestK8sRepo_CopyToPod_OpenError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, _ := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)
	kr.archiver = &errArchiver{openErr: errors.New("open boom")}
	kr.maxUploadSize = 1 << 20

	info := uploader.NewMockFileInfo(m)
	info.EXPECT().Size().Return(uint64(8)).AnyTimes()
	srcPath := "/some/file"
	mockUploader.EXPECT().LocalUploader().Return(mockUploader)
	mockUploader.EXPECT().Stat(srcPath).Return(info, nil)
	mockUploader.EXPECT().Type().Return(schematype.Local)
	mockFileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{Path: srcPath}, nil)

	_, err := kr.CopyFileToPod(context.TODO(), &biz.CopyFileToPodInput{FileId: 1})
	assert.ErrorContains(t, err, "open boom")
}

// TestK8sRepo_CopyToPod_IOCopyError 覆盖归档流向容器时 io.Copy 读源失败的分支
// Open 返回读即报错的 reader，goroutine 内 io.Copy 出错仅打日志，
// Execute 仍照常执行（mock 消费 Stdin）。
func TestK8sRepo_CopyToPod_IOCopyError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, ec := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)
	kr.archiver = &errArchiver{}
	kr.maxUploadSize = 1 << 20

	info := uploader.NewMockFileInfo(m)
	info.EXPECT().Size().Return(uint64(8)).AnyTimes()
	srcPath := "/some/file"
	mockUploader.EXPECT().LocalUploader().Return(mockUploader)
	mockUploader.EXPECT().Stat(srcPath).Return(info, nil)
	mockUploader.EXPECT().Type().Return(schematype.Local)
	execDrain(ec)
	mockFileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{Path: srcPath}, nil)
	mockFileRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(&biz.File{}, nil)

	got, err := kr.CopyFileToPod(context.TODO(), &biz.CopyFileToPodInput{FileId: 1})
	require.NoError(t, err)
	assert.NotNil(t, got)
}

// TestK8sRepo_CopyFileToPod_UpdateError 覆盖 CopyFileToPod 最后写文件记录失败的分支
// copyToPod 成功后 fileRepo.Update 返回错误。
func TestK8sRepo_CopyFileToPod_UpdateError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFileRepo := NewMockFileRepo(m)
	kr, _, ec := copyFromPodK8sRepo(t, m, mockUploader, mockFileRepo)
	kr.archiver = NewDefaultArchiver()
	kr.maxUploadSize = 1 << 20

	info := uploader.NewMockFileInfo(m)
	info.EXPECT().Size().Return(uint64(8)).AnyTimes()
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "a.txt")
	require.NoError(t, os.WriteFile(srcPath, []byte("hi"), 0o644))
	mockUploader.EXPECT().LocalUploader().Return(uploader.NewMockUploader(m))
	mockUploader.EXPECT().Stat(srcPath).Return(info, nil)
	mockUploader.EXPECT().Type().Return(schematype.Local)
	execDrain(ec)
	mockFileRepo.EXPECT().GetByID(gomock.Any(), 7).Return(&biz.File{Path: srcPath}, nil)
	mockFileRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("update boom"))

	_, err := kr.CopyFileToPod(context.TODO(), &biz.CopyFileToPodInput{
		FileId: 7, Namespace: "ns", Pod: "p", Container: "c",
	})
	assert.EqualError(t, err, "update boom")
}

// 注：LogStream 成功/失败两条分支都是集成边界（真实 API server 流）——官方 fake 的
// GetLogs() 在 Invokes 阶段查 reactor 但返回值被
// `_, _ =` 丢弃，随后硬编码一个恒返回 200 "fake logs" 的 HTTP client 供 Stream() 用。
// 无论 reactor 注入错误还是内容都被吞掉：错误分支拿不到 err，
// 正常路径读循环拿到默认 "fake logs" 且 nil-deref panic，单测无法稳定驱动。

// newWorkloadIndexer 构造一个以命名空间索引的 shared indexer，供 workload lister 构造用。
func newWorkloadIndexer() cache.Indexer {
	return cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}

// NewDeploymentLister 构造带指定对象的 Deployment lister。
func NewDeploymentLister(deps ...*appsv1.Deployment) appsv1lister.DeploymentLister {
	idxer := newWorkloadIndexer()
	for _, d := range deps {
		_ = idxer.Add(d)
	}
	return appsv1lister.NewDeploymentLister(idxer)
}

// NewStatefulSetLister 构造带指定对象的 StatefulSet lister。
func NewStatefulSetLister(stsList ...*appsv1.StatefulSet) appsv1lister.StatefulSetLister {
	idxer := newWorkloadIndexer()
	for _, s := range stsList {
		_ = idxer.Add(s)
	}
	return appsv1lister.NewStatefulSetLister(idxer)
}

// NewDaemonSetLister 构造带指定对象的 DaemonSet lister。
func NewDaemonSetLister(dss ...*appsv1.DaemonSet) appsv1lister.DaemonSetLister {
	idxer := newWorkloadIndexer()
	for _, d := range dss {
		_ = idxer.Add(d)
	}
	return appsv1lister.NewDaemonSetLister(idxer)
}

// errNamespaceLister 是 List/Get 恒报错的命名空间级 workload lister 替身。
type errNamespaceLister[T any] struct{}

func (errNamespaceLister[T]) List(_ labels.Selector) ([]*T, error) {
	return nil, errors.New("list boom")
}
func (errNamespaceLister[T]) Get(string) (*T, error) { return nil, errors.New("get boom") }

// errDeploymentLister 是 Deployment lister 替身，Deployments 返回恒报错的命名空间级 lister。
type errDeploymentLister struct{}

func (errDeploymentLister) List(labels.Selector) ([]*appsv1.Deployment, error) {
	return nil, errors.New("boom")
}
func (errDeploymentLister) Deployments(string) appsv1lister.DeploymentNamespaceLister {
	return errNamespaceLister[appsv1.Deployment]{}
}

// errStatefulSetLister 是 StatefulSet lister 替身，StatefulSets 返回恒报错的命名空间级 lister。
type errStatefulSetLister struct{}

func (errStatefulSetLister) List(labels.Selector) ([]*appsv1.StatefulSet, error) {
	return nil, errors.New("boom")
}
func (errStatefulSetLister) StatefulSets(string) appsv1lister.StatefulSetNamespaceLister {
	return errNamespaceLister[appsv1.StatefulSet]{}
}
func (errStatefulSetLister) GetPodStatefulSets(*corev1.Pod) ([]*appsv1.StatefulSet, error) {
	return nil, errors.New("boom")
}

// errDaemonSetLister 是 DaemonSet lister 替身，DaemonSets 返回恒报错的命名空间级 lister。
type errDaemonSetLister struct{}

func (errDaemonSetLister) List(labels.Selector) ([]*appsv1.DaemonSet, error) {
	return nil, errors.New("boom")
}
func (errDaemonSetLister) DaemonSets(string) appsv1lister.DaemonSetNamespaceLister {
	return errNamespaceLister[appsv1.DaemonSet]{}
}
func (errDaemonSetLister) GetPodDaemonSets(*corev1.Pod) ([]*appsv1.DaemonSet, error) {
	return nil, errors.New("boom")
}
func (errDaemonSetLister) GetHistoryDaemonSets(*appsv1.ControllerRevision) ([]*appsv1.DaemonSet, error) {
	return nil, errors.New("boom")
}

// errReplicaSetLister 是 ReplicaSet lister 替身，ReplicaSets 返回恒报错的命名空间级 lister。
type errReplicaSetLister struct{}

func (errReplicaSetLister) List(labels.Selector) ([]*appsv1.ReplicaSet, error) {
	return nil, errors.New("boom")
}
func (errReplicaSetLister) ReplicaSets(string) appsv1lister.ReplicaSetNamespaceLister {
	return errNamespaceLister[appsv1.ReplicaSet]{}
}
func (errReplicaSetLister) GetPodReplicaSets(*corev1.Pod) ([]*appsv1.ReplicaSet, error) {
	return nil, errors.New("boom")
}

// newK8sRepoWithClient 构造包住指定 K8sClient 的 k8sRepo，供 lister 读取方法测试用。
func newK8sRepoWithClient(mockData *MockDataStore, c *K8sClient) *k8sRepo {
	mockData.EXPECT().K8s().Return(c).AnyTimes()
	return &k8sRepo{logger: mlog.NewForConfig(nil), data: mockData, cache: NewCacheImpl(&config.Config{}, nil, mlog.NewForConfig(nil))}
}

func TestK8sRepo_ListReplicaSets(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	rs := &appsv1.ReplicaSet{ObjectMeta: metav1.ObjectMeta{Name: "web-abc", Namespace: "ns"}}
	kr := newK8sRepoWithClient(NewMockDataStore(m), &K8sClient{ReplicaSetLister: NewRsLister(rs)})

	got, err := kr.ListReplicaSets("ns")
	assert.NoError(t, err)
	if assert.Len(t, got, 1) {
		assert.Equal(t, "web-abc", got[0].Name)
	}

	// informer List 失败 → errs.Wrap 上抛。
	krErr := newK8sRepoWithClient(NewMockDataStore(m), &K8sClient{ReplicaSetLister: errReplicaSetLister{}})
	_, err = krErr.ListReplicaSets("ns")
	assert.ErrorContains(t, err, "list replica sets")
}

func TestK8sRepo_GetDeployment(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "ns"}}
	kr := newK8sRepoWithClient(NewMockDataStore(m), &K8sClient{DeploymentLister: NewDeploymentLister(dep)})

	got, err := kr.GetDeployment("ns", "web")
	assert.NoError(t, err)
	assert.Equal(t, "web", got.Name)

	// informer Get 失败（not found）→ errs.Wrap 上抛。
	krErr := newK8sRepoWithClient(NewMockDataStore(m), &K8sClient{DeploymentLister: errDeploymentLister{}})
	_, err = krErr.GetDeployment("ns", "web")
	assert.ErrorContains(t, err, "get deployment")
}

func TestK8sRepo_GetStatefulSet(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	sts := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "ns"}}
	kr := newK8sRepoWithClient(NewMockDataStore(m), &K8sClient{StatefulSetLister: NewStatefulSetLister(sts)})

	got, err := kr.GetStatefulSet("ns", "db")
	assert.NoError(t, err)
	assert.Equal(t, "db", got.Name)

	krErr := newK8sRepoWithClient(NewMockDataStore(m), &K8sClient{StatefulSetLister: errStatefulSetLister{}})
	_, err = krErr.GetStatefulSet("ns", "db")
	assert.ErrorContains(t, err, "get statefulset")
}

func TestK8sRepo_GetDaemonSet(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ds := &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: "agent", Namespace: "ns"}}
	kr := newK8sRepoWithClient(NewMockDataStore(m), &K8sClient{DaemonSetLister: NewDaemonSetLister(ds)})

	got, err := kr.GetDaemonSet("ns", "agent")
	assert.NoError(t, err)
	assert.Equal(t, "agent", got.Name)

	krErr := newK8sRepoWithClient(NewMockDataStore(m), &K8sClient{DaemonSetLister: errDaemonSetLister{}})
	_, err = krErr.GetDaemonSet("ns", "agent")
	assert.ErrorContains(t, err, "get daemonset")
}

// TestK8sRepo_GetWorkloadsByManifest 覆盖从 manifest 解析三类工作负载：Deployment/STS/DS
// 各识别一个，Service/无法解码片段跳过。
func TestK8sRepo_GetWorkloadsByManifest(t *testing.T) {
	kr := &k8sRepo{logger: mlog.NewForConfig(nil)}
	deployments, statefulSets, daemonSets := kr.GetWorkloadsByManifest([]string{
		dedent.Dedent(`
			apiVersion: apps/v1
			kind: Deployment
			metadata:
			  name: web
			  namespace: ns
		`),
		dedent.Dedent(`
			apiVersion: apps/v1
			kind: StatefulSet
			metadata:
			  name: db
			  namespace: ns
		`),
		dedent.Dedent(`
			apiVersion: apps/v1
			kind: DaemonSet
			metadata:
			  name: agent
			  namespace: ns
		`),
		// Service 不属于滚动更新工作负载，跳过。
		dedent.Dedent(`
			apiVersion: v1
			kind: Service
			metadata:
			  name: svc
			  namespace: ns
		`),
		// 无法解码的片段（空对象）跳过。
		"",
	})
	if assert.Len(t, deployments, 1) {
		assert.Equal(t, "web", deployments[0].Name)
	}
	if assert.Len(t, statefulSets, 1) {
		assert.Equal(t, "db", statefulSets[0].Name)
	}
	if assert.Len(t, daemonSets, 1) {
		assert.Equal(t, "agent", daemonSets[0].Name)
	}
}

// TestK8sRepo_ClusterBoard 集群看板快照：一次拉取节点/节点指标/命名空间/Pod/Pod 指标，
// fake 客户端注入全量数据后断言各切片完整落位。
func TestK8sRepo_ClusterBoard(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	cpu := resource.MustParse("4")
	memory := resource.MustParse("8Gi")
	fc := fake.NewSimpleClientset(
		&corev1.NodeList{Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node01",
					Labels: map[string]string{
						"node-role.kubernetes.io/master": "",
						"kubernetes.io/hostname":         "host",
					},
				},
				Status: corev1.NodeStatus{
					Capacity: corev1.ResourceList{
						corev1.ResourceCPU:    cpu.DeepCopy(),
						corev1.ResourceMemory: memory.DeepCopy(),
					},
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
						{Type: corev1.NodeMemoryPressure, Status: corev1.ConditionFalse},
					},
				},
			},
			{ObjectMeta: metav1.ObjectMeta{Name: "node02"}, Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeMemoryPressure, Status: corev1.ConditionFalse},
			}}},
		}},
		&corev1.NamespaceList{Items: []corev1.Namespace{
			{ObjectMeta: metav1.ObjectMeta{Name: "ns-a"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "ns-b"}},
		}},
		&corev1.PodList{Items: []corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "p2", Namespace: "ns-b"}},
		}},
	)
	fcm := &fake2.Clientset{}
	fcm.AddReactor("list", "nodes", func(action testing2.Action) (bool, runtime.Object, error) {
		return true, &v1beta1.NodeMetricsList{Items: []v1beta1.NodeMetrics{
			{ObjectMeta: metav1.ObjectMeta{Name: "node01"}, Usage: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			}},
		}}, nil
	})
	fcm.AddReactor("list", "pods", func(action testing2.Action) (bool, runtime.Object, error) {
		return true, &v1beta1.PodMetricsList{Items: []v1beta1.PodMetrics{
			{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a"}, Containers: []v1beta1.ContainerMetrics{
				{Usage: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				}},
			}},
			{ObjectMeta: metav1.ObjectMeta{Name: "p2", Namespace: "ns-b"}, Containers: []v1beta1.ContainerMetrics{
				{Usage: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("250m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				}},
			}},
		}}, nil
	})
	mockData := NewMockDataStore(m)
	kr := &k8sRepo{logger: mlog.NewForConfig(nil), data: mockData, cache: NewCacheImpl(&config.Config{}, nil, mlog.NewForConfig(nil))}
	mockData.EXPECT().K8s().Return(&K8sClient{Client: fc, MetricsClient: fcm}).AnyTimes()

	got, err := kr.ClusterBoard(context.TODO(), false)
	assert.NoError(t, err)
	assert.Len(t, got.Nodes, 2)
	assert.Len(t, got.NodeMetrics, 1)
	assert.Len(t, got.Namespaces, 2)
	assert.Len(t, got.Pods, 2)
	assert.Len(t, got.PodMetrics, 2)
	assert.Equal(t, "node01", got.Nodes[0].Name)
	assert.Equal(t, int64(4000), got.Nodes[0].CpuCapacityMilli, "mapper 归约：4 核容量转毫核")
	assert.Equal(t, int64(8589934592), got.Nodes[0].MemCapacityBytes, "mapper 归约：8Gi 容量转字节")
	assert.Equal(t, "", got.Nodes[0].Labels["node-role.kubernetes.io/master"], "mapper 归约：保留角色标签")
	_, hasHostname := got.Nodes[0].Labels["kubernetes.io/hostname"]
	assert.False(t, hasHostname, "mapper 归约：丢弃非角色标签")
	assert.Equal(t, "True", got.Nodes[0].ReadyStatus, "mapper 归约：NodeReady 条件归约为 Status")
	assert.Equal(t, "", got.Nodes[1].ReadyStatus, "mapper 归约：仅非 Ready 条件时 ReadyStatus 为空")
	assert.Equal(t, "ns-a", got.Namespaces[0].Name)
	assert.Equal(t, "p1", got.PodMetrics[0].Name)
}

// TestK8sRepo_ClusterBoard_Cache 覆盖 30s 缓存语义：force=false 首次调用触发 List 并回填，
// 二次调用命中缓存不再触发 List；force=true 强制刷新重新触发 List（cron 预热路径）。
func TestK8sRepo_ClusterBoard_Cache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	var listCalls int32
	fc := fake.NewSimpleClientset()
	fc.PrependReactor("list", "nodes", func(action testing2.Action) (bool, runtime.Object, error) {
		listCalls++
		return true, &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "n1"}}}}, nil
	})
	fcm := &fake2.Clientset{}
	fcm.AddReactor("list", "nodes", func(action testing2.Action) (bool, runtime.Object, error) {
		return true, &v1beta1.NodeMetricsList{}, nil
	})
	mockData := NewMockDataStore(m)
	mockData.EXPECT().K8s().Return(&K8sClient{Client: fc, MetricsClient: fcm}).AnyTimes()
	kr := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
		cache:  NewCacheImpl(&config.Config{CacheDriver: "memory"}, nil, mlog.NewForConfig(nil)),
	}

	got, err := kr.ClusterBoard(context.TODO(), false)
	assert.NoError(t, err)
	assert.Len(t, got.Nodes, 1)
	assert.Equal(t, int32(1), listCalls, "首次调用触发 List 回填")

	got2, err := kr.ClusterBoard(context.TODO(), false)
	assert.NoError(t, err)
	assert.Equal(t, got.Nodes[0].Name, got2.Nodes[0].Name)
	assert.Equal(t, int32(1), listCalls, "force=false 二次调用命中缓存不触发 List")

	_, err = kr.ClusterBoard(context.TODO(), true)
	assert.NoError(t, err)
	assert.Equal(t, int32(2), listCalls, "force=true 强制刷新重新触发 List")
}

// TestK8sRepo_ClusterBoard_UnmarshalError 覆盖缓存值损坏（非合法 JSON）时反序列化
// 失败整体上抛，不产生半成品快照（防御缓存被外部污染的分支）。
func TestK8sRepo_ClusterBoard_UnmarshalError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cache := NewMockCache(m)
	cache.EXPECT().Remember(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte("not-json"), nil)
	kr := &k8sRepo{cache: cache}

	got, err := kr.ClusterBoard(context.TODO(), false)
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal cluster board cache")
}

// TestK8sRepo_CacheTTLs 断言两个快照缓存使用各自 TTL：ClusterBoard=30s、
// ResourceSnapshot=300s，锁死 Remember 的 seconds 参数——TTL 是本次分频的核心语义，
// 防止将来被误统一成同一刷新周期（cron 预热节律与之一致）。注意断言用字面契约值
// （30/300）而非常量引用：若常量被误改，测试仍能抓住错配。
func TestK8sRepo_CacheTTLs(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	cache := NewMockCache(m)
	cache.EXPECT().Remember(NewKey("cluster_board"), 30, gomock.Any(), gomock.Any()).Return([]byte("{}"), nil)
	cache.EXPECT().Remember(NewKey("resource_snapshot"), 300, gomock.Any(), gomock.Any()).Return([]byte("{}"), nil)

	kr := &k8sRepo{cache: cache}
	_, err := kr.ClusterBoard(context.TODO(), false)
	assert.NoError(t, err)
	_, err = kr.ResourceSnapshot(context.TODO(), false)
	assert.NoError(t, err)
}

// TestK8sRepo_ResourceSnapshot_UnmarshalError 同上：空间资源快照缓存损坏时反序列化失败上抛。
func TestK8sRepo_ResourceSnapshot_UnmarshalError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cache := NewMockCache(m)
	cache.EXPECT().Remember(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte("not-json"), nil)
	kr := &k8sRepo{cache: cache}

	got, err := kr.ResourceSnapshot(context.TODO(), false)
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal resource snapshot cache")
}

// TestK8sRepo_ClusterBoard_Errors 集群看板各资源 List 失败整体上抛：任一环节失败
// 返回 errs.Wrap 错误且快照为 nil，不产生半成品。
func TestK8sRepo_ClusterBoard_Errors(t *testing.T) {
	cases := []struct {
		name            string
		coreResource    string // fc 上注入失败 reactor 的资源；空串表示不注入
		metricsResource string // fcm 上注入失败 reactor 的资源；空串表示不注入
	}{
		{name: "nodes", coreResource: "nodes"},
		{name: "namespaces", coreResource: "namespaces"},
		{name: "pods", coreResource: "pods"},
		{name: "node metrics", metricsResource: "nodes"},
		{name: "pod metrics", metricsResource: "pods"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := gomock.NewController(t)
			defer m.Finish()
			fc := fake.NewSimpleClientset()
			if tc.coreResource != "" {
				fc.PrependReactor("list", tc.coreResource, func(action testing2.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("core list boom")
				})
			}
			fcm := &fake2.Clientset{}
			if tc.metricsResource != "" {
				fcm.PrependReactor("list", tc.metricsResource, func(action testing2.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("metrics list boom")
				})
			}
			mockData := NewMockDataStore(m)
			kr := &k8sRepo{logger: mlog.NewForConfig(nil), data: mockData, cache: NewCacheImpl(&config.Config{}, nil, mlog.NewForConfig(nil))}
			mockData.EXPECT().K8s().Return(&K8sClient{Client: fc, MetricsClient: fcm}).AnyTimes()

			got, err := kr.ClusterBoard(context.TODO(), false)
			assert.Nil(t, got)
			assert.Error(t, err)
		})
	}
}

// TestK8sRepo_ResourceSnapshot 空间资源快照：拉 Running Pod、ReplicaSet 与其指标
// 三段数据，Pod List 必须携带 status.phase=Running 字段选择器，切片完整落位。
func TestK8sRepo_ResourceSnapshot(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	fc := fake.NewSimpleClientset()
	fc.PrependReactor("list", "pods", func(action testing2.Action) (bool, runtime.Object, error) {
		restrictions := action.(testing2.ListAction).GetListRestrictions()
		assert.Equal(t, "status.phase=Running", restrictions.Fields.String(), "Pod List 必须过滤 Running")
		return true, &corev1.PodList{Items: []corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a", OwnerReferences: []metav1.OwnerReference{{Kind: "StatefulSet", Name: "db"}}}, Status: corev1.PodStatus{Phase: corev1.PodRunning}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")}}}}}},
		}}, nil
	})
	fc.PrependReactor("list", "replicasets", func(action testing2.Action) (bool, runtime.Object, error) {
		return true, &appsv1.ReplicaSetList{Items: []appsv1.ReplicaSet{
			{ObjectMeta: metav1.ObjectMeta{Name: "p1-abc", Namespace: "ns-a", UID: "rs-1", OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", Name: "web"}, {Kind: "Node", Name: "x"}}}},
		}}, nil
	})
	fcm := &fake2.Clientset{}
	fcm.AddReactor("list", "pods", func(action testing2.Action) (bool, runtime.Object, error) {
		return true, &v1beta1.PodMetricsList{Items: []v1beta1.PodMetrics{
			{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns-a"}, Containers: []v1beta1.ContainerMetrics{
				{Usage: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				}},
			}},
		}}, nil
	})
	mockData := NewMockDataStore(m)
	kr := &k8sRepo{logger: mlog.NewForConfig(nil), data: mockData, cache: NewCacheImpl(&config.Config{}, nil, mlog.NewForConfig(nil))}
	mockData.EXPECT().K8s().Return(&K8sClient{Client: fc, MetricsClient: fcm}).AnyTimes()

	got, err := kr.ResourceSnapshot(context.TODO(), false)
	assert.NoError(t, err)
	if assert.Len(t, got.Pods, 1) {
		assert.Equal(t, "p1", got.Pods[0].Name)
		assert.Equal(t, int64(500), got.Pods[0].CpuRequestMilli, "mapper 归约：容器 500m requests 聚合")
		if assert.Len(t, got.Pods[0].Owners, 1) {
			assert.Equal(t, "StatefulSet", got.Pods[0].Owners[0].Kind, "mapper 归约：保留 pod 属主")
		}
	}
	if assert.Len(t, got.ReplicaSets, 1) {
		assert.Equal(t, "rs-1", got.ReplicaSets[0].UID, "ReplicaSet 归约为 UID + Deployment 属主")
		if assert.Len(t, got.ReplicaSets[0].Owners, 1) {
			assert.Equal(t, "Deployment", got.ReplicaSets[0].Owners[0].Kind, "mapper 归约：RS 只保留 Deployment 属主")
		}
	}
	if assert.Len(t, got.PodMetrics, 1) {
		assert.Equal(t, "p1", got.PodMetrics[0].Name)
		assert.Equal(t, int64(500), got.PodMetrics[0].CpuMilli, "mapper 归约：容器用量聚合")
		assert.Equal(t, int64(268435456), got.PodMetrics[0].MemBytes, "mapper 归约：256Mi 用量转字节")
	}
}

// TestK8sRepo_ResourceSnapshot_Errors 空间资源快照各 List 失败整体上抛：
// Pod、ReplicaSet 与 Pod 指标任一环节失败均返回错误且快照为 nil，不产生半成品。
func TestK8sRepo_ResourceSnapshot_Errors(t *testing.T) {
	cases := []struct {
		name            string
		coreResource    string // fc 上注入失败 reactor 的资源；空串表示不注入
		metricsResource string // fcm 上注入失败 reactor 的资源；空串表示不注入
	}{
		{name: "pods", coreResource: "pods"},
		{name: "replica sets", coreResource: "replicasets"},
		{name: "pod metrics", metricsResource: "pods"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := gomock.NewController(t)
			defer m.Finish()
			fc := fake.NewSimpleClientset()
			if tc.coreResource != "" {
				fc.PrependReactor("list", tc.coreResource, func(action testing2.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("core list boom")
				})
			}
			fcm := &fake2.Clientset{}
			if tc.metricsResource != "" {
				fcm.PrependReactor("list", tc.metricsResource, func(action testing2.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("metrics list boom")
				})
			}
			mockData := NewMockDataStore(m)
			kr := &k8sRepo{logger: mlog.NewForConfig(nil), data: mockData, cache: NewCacheImpl(&config.Config{}, nil, mlog.NewForConfig(nil))}
			mockData.EXPECT().K8s().Return(&K8sClient{Client: fc, MetricsClient: fcm}).AnyTimes()

			got, err := kr.ResourceSnapshot(context.TODO(), false)
			assert.Nil(t, got)
			assert.Error(t, err)
		})
	}
}
