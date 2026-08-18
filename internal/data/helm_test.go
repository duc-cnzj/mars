package data

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	restclient "k8s.io/client-go/rest"
)

func TestReleaseStatus(t *testing.T) {
	// 走导出方法覆盖包装层与 releaseStatus 实现：
	// 集群不可达时回退为 StatusUnknown。
	status := (&DefaultHelmer{
		logger: mlog.NewForConfig(nil),
	}).ReleaseStatus("a", "test")
	assert.Equal(t, types.Deploy_StatusUnknown, status)
}

func TestRollback(t *testing.T) {
	err := (&DefaultHelmer{}).Rollback("test", "ns", false, nil, false)
	assert.Error(t, err)
}

func TestUninstallRelease(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	err := (&DefaultHelmer{}).Uninstall("test", "ns", func(format string, v ...any) {})
	assert.Error(t, err)
}

func Test_checkIfInstallable(t *testing.T) {
	err := checkIfInstallable(&chart.Chart{
		Metadata: &chart.Metadata{
			Type: "",
		},
	})
	assert.Nil(t, err)
	err = checkIfInstallable(&chart.Chart{
		Metadata: &chart.Metadata{
			Type: "application",
		},
	})
	assert.Nil(t, err)
	err = checkIfInstallable(&chart.Chart{
		Metadata: &chart.Metadata{
			Type: "xxx",
		},
	})
	require.Error(t, err)
	// chart 类型非法是显式校验失败，映射为 InvalidArgument(400) 而非默认 500，
	// 上层 errs.Wrap 保留该状态码，客户端不会把"chart 不可安装"误判成服务器内部错误。
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestUpgradeOrInstall_ClusterError 覆盖 UpgradeOrInstall 两个失败路径：
// 集群不可达时（kubeconfig 无效）首轮 dry-run 探测失败 / 直连升级失败，错误均冒泡。
// happy path 需真实集群（helm 集群操作），属集成边界不在单测范围。
func TestUpgradeOrInstall_ClusterError(t *testing.T) {
	newHelmer := func() *DefaultHelmer {
		return &DefaultHelmer{
			logger:     mlog.NewForConfig(nil),
			KubeConfig: "/nonexistent/kubeconfig",
			data:       &dataImpl{},
		}
	}
	ch := &chart.Chart{Metadata: &chart.Metadata{}}
	fn := biz.WrapLogFn(func([]*websocket_pb.Container, string, ...any) {})

	t.Run("wait with dry-run probe fails", func(t *testing.T) {
		_, err := newHelmer().UpgradeOrInstall(context.TODO(), "rel", "ns", ch, &values.Options{}, fn, true, 0, false, "desc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "upgrade or install")
	})

	t.Run("direct non-wait upgrade fails with nil valueOpts", func(t *testing.T) {
		// nil valueOpts 触发 upgradeOrInstall 的缺省回落分支。
		_, err := newHelmer().UpgradeOrInstall(context.TODO(), "rel", "ns", ch, nil, fn, false, 0, false, "desc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "upgrade or install")
	})
}

func Test_getActionConfigAndSettings(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	// kubeconfig 为空走 in-cluster 分支，非空走 --kubeconfig 分支。
	for _, kc := range []string{"", "test"} {
		settings := getActionConfigAndSettings("test", kc, func(format string, v ...any) {})
		assert.NotNil(t, settings)
	}
}

func Test_watchEvent(t *testing.T) {
	t.Parallel()
	ctx, cancelFn := context.WithCancel(context.TODO())
	ch := make(chan Obj[*eventv1.Event], 10)
	go func() {
		ch <- newObj(nil, &eventv1.Event{
			Regarding: corev1.ObjectReference{
				Namespace: "ns",
				Name:      "app",
			},
		}, Add)
		ch <- newObj(nil, &eventv1.Event{
			Regarding: corev1.ObjectReference{
				Namespace: "ns",
				Name:      "app1",
			},
		}, Update)
		ch <- newObj(nil, &eventv1.Event{
			Regarding: corev1.ObjectReference{
				Namespace: "ns",
				Name:      "app2",
			},
		}, Delete)
		time.Sleep(2 * time.Second)
		cancelFn()
	}()
	var called int64
	(&DefaultHelmer{
		logger: mlog.NewForConfig(nil),
	}).watchEvent(ctx, ch, "release", func(container []*websocket_pb.Container, format string, v ...any) {
		atomic.AddInt64(&called, 1)
	}, NewPodLister(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "app",
			Labels: map[string]string{
				"xxx":          "xxx",
				"release-name": "release",
			},
		},
	}))

	assert.Equal(t, int64(1), atomic.LoadInt64(&called))
}

func Test_watchEvent_Error1(t *testing.T) {
	t.Parallel()
	ctx, cancelFn := context.WithCancel(context.TODO())
	ch := make(chan Obj[*eventv1.Event], 10)
	go func() {
		close(ch)
		time.Sleep(2 * time.Second)
		cancelFn()
	}()
	var called int64
	(&DefaultHelmer{
		logger: mlog.NewForConfig(nil),
	}).watchEvent(ctx, ch, "release", func(container []*websocket_pb.Container, format string, v ...any) {
		atomic.AddInt64(&called, 1)
	}, NewPodLister(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "app",
			Labels: map[string]string{
				"xxx":          "xxx",
				"release-name": "release",
			},
		},
	}))

	assert.Equal(t, int64(0), atomic.LoadInt64(&called))
}

func Test_watchEvent_Error2(t *testing.T) {
	t.Parallel()
	ctx, cancelFn := context.WithCancel(context.TODO())
	ch := make(chan Obj[*eventv1.Event], 10)
	go func() {
		ch <- newObj(nil, &eventv1.Event{
			Regarding: corev1.ObjectReference{
				Namespace: "ns",
				Name:      "app1",
			},
		}, Add)
		ch <- newObj(nil, &eventv1.Event{
			Regarding: corev1.ObjectReference{
				Namespace: "ns",
				Name:      "app",
			},
		}, Add)
		time.Sleep(2 * time.Second)
		cancelFn()
	}()
	var called int64
	(&DefaultHelmer{
		logger: mlog.NewForConfig(nil),
	}).watchEvent(ctx, ch, "release", func(container []*websocket_pb.Container, format string, v ...any) {
		atomic.AddInt64(&called, 1)
	}, NewPodLister(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "app",
			Labels: map[string]string{
				"xxx":          "xxx",
				"release-name": "release",
			},
		},
	}))

	assert.Equal(t, int64(1), atomic.LoadInt64(&called))
}

func Test_watchPodStatus(t *testing.T) {
	t.Parallel()
	var called int64
	podCh := make(chan Obj[*corev1.Pod], 10)
	ctx, cancelFn := context.WithCancel(context.TODO())
	go func() {
		podCh <- newObj(nil, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "app1",
				Labels: map[string]string{
					"name": "app",
				},
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:         "aaa",
						Ready:        false,
						RestartCount: 6,
					},
				},
			},
		}, Add)
		podCh <- newObj(nil, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "app2",
				Labels: map[string]string{
					"name": "app",
				},
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:         "bbb",
						Ready:        false,
						RestartCount: 5,
					},
				},
			},
		}, Delete)
		podCh <- newObj[*corev1.Pod](nil, nil, Update)
		podCh <- newObj[*corev1.Pod](nil, nil, Update)
		podCh <- newObj[*corev1.Pod](nil, nil, Update)
		time.Sleep(2 * time.Second)
		cancelFn()
	}()
	selectorLists := []labels.Selector{
		labels.SelectorFromSet(map[string]string{
			"name": "app",
		}),
		labels.SelectorFromSet(map[string]string{
			"release": "v1",
		}),
	}
	(&DefaultHelmer{
		logger: mlog.NewForConfig(nil),
	}).watchPodStatus(ctx, podCh, selectorLists, func(container []*websocket_pb.Container, format string, v ...any) {
		atomic.AddInt64(&called, 1)
	})
	assert.Equal(t, int64(2), atomic.LoadInt64(&called))

	podCh2 := make(chan Obj[*corev1.Pod], 10)
	close(podCh2)
	assert.NotPanics(t, func() {
		(&DefaultHelmer{
			logger: mlog.NewForConfig(nil),
		}).watchPodStatus(context.TODO(), podCh2, nil, nil)
	})
}

func Test_watchPodStatus_Error1(t *testing.T) {
	t.Parallel()
	var called int64
	var cs = &ContainerGetterSetter{}
	podCh := make(chan Obj[*corev1.Pod], 10)
	ctx, cancelFn := context.WithCancel(context.TODO())
	go func() {
		podCh <- newObj[*corev1.Pod](nil, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "app-not-match",
				Labels: map[string]string{
					"name": "app-not-match",
				},
			},
		}, Add)
		podCh <- newObj[*corev1.Pod](nil, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "app",
				Labels: map[string]string{
					"name": "app",
				},
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:         "three",
						Ready:        false,
						RestartCount: 0,
					},
				},
			},
		}, Delete)
		time.Sleep(2 * time.Second)
		cancelFn()
	}()
	selectorLists := []labels.Selector{
		labels.SelectorFromSet(map[string]string{
			"name": "app",
		}),
		labels.SelectorFromSet(map[string]string{
			"release": "v1",
		}),
	}
	(&DefaultHelmer{
		logger: mlog.NewForConfig(nil),
	}).watchPodStatus(ctx, podCh, selectorLists, func(container []*websocket_pb.Container, format string, v ...any) {
		cs.Set(container)
		atomic.AddInt64(&called, 1)
	})
	assert.Len(t, cs.Get(), 0)
	assert.Equal(t, int64(0), atomic.LoadInt64(&called))
}

type ContainerGetterSetter struct {
	sync.Mutex
	cs []*websocket_pb.Container
}

func (c *ContainerGetterSetter) Set(cs []*websocket_pb.Container) {
	c.Lock()
	defer c.Unlock()
	c.cs = cs
}
func (c *ContainerGetterSetter) Get() []*websocket_pb.Container {
	c.Lock()
	defer c.Unlock()
	return c.cs
}

func Test_formatStatus(t *testing.T) {
	var tests = []struct {
		input release.Status
		want  types.Deploy
	}{
		{
			input: release.StatusPendingUpgrade,
			want:  types.Deploy_StatusDeploying,
		},
		{
			input: release.StatusPendingInstall,
			want:  types.Deploy_StatusDeploying,
		},
		{
			input: release.StatusPendingRollback,
			want:  types.Deploy_StatusDeploying,
		},
		{
			input: release.StatusDeployed,
			want:  types.Deploy_StatusDeployed,
		},
		{
			input: release.StatusFailed,
			want:  types.Deploy_StatusFailed,
		},
		{
			input: "xxx",
			want:  types.Deploy_StatusUnknown,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run("", func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, formatStatus(tt.input))
		})
	}
}

func Test_fillInstall(t *testing.T) {
	i := &action.Install{}
	u := &action.Upgrade{
		Install:                  true,
		Devel:                    true,
		Namespace:                "xxx",
		SkipCRDs:                 true,
		Timeout:                  10,
		Wait:                     true,
		WaitForJobs:              true,
		DisableHooks:             true,
		DryRun:                   true,
		Force:                    true,
		Atomic:                   true,
		SubNotes:                 true,
		Description:              "desc",
		DisableOpenAPIValidation: true,
		DependencyUpdate:         true,
	}
	fillInstall(i, u)

	assert.Equal(t, i.CreateNamespace, true)
	assert.Equal(t, i.ChartPathOptions, u.ChartPathOptions)
	assert.Equal(t, i.DryRun, u.DryRun)
	assert.Equal(t, i.DisableHooks, u.DisableHooks)
	assert.Equal(t, i.SkipCRDs, u.SkipCRDs)
	assert.Equal(t, i.Timeout, u.Timeout)
	assert.Equal(t, i.Wait, u.Wait)
	assert.Equal(t, i.WaitForJobs, u.WaitForJobs)
	assert.Equal(t, i.Devel, u.Devel)
	assert.Equal(t, i.Namespace, u.Namespace)
	assert.Equal(t, i.Atomic, u.Atomic)
	assert.Equal(t, i.PostRenderer, u.PostRenderer)
	assert.Equal(t, i.DisableOpenAPIValidation, u.DisableOpenAPIValidation)
	assert.Equal(t, i.SubNotes, u.SubNotes)
	assert.Equal(t, i.Description, u.Description)
	assert.Equal(t, i.DependencyUpdate, u.DependencyUpdate)
}

func Test_wrapRestConfig(t *testing.T) {
	cfg := &restclient.Config{}
	wrapRestConfig(cfg)
	assert.Equal(t, float32(-1), cfg.QPS)
}

func Test_logWriter_Write(t *testing.T) {
	n, err := (&logWriter{}).Write([]byte("ass"))
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
}

func Test_newDefaultRegistryClient(t *testing.T) {
	client, err := newDefaultRegistryClient(false, "")
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// 错误分支：凭据文件存在但 JSON 格式错误，
	// dockerauth.NewClientWithDockerFallback 解析失败返回错误。
	bad := filepath.Join(t.TempDir(), "creds.json")
	require.NoError(t, os.WriteFile(bad, []byte("not-json{{{"), 0o600))
	_, err = newDefaultRegistryClient(false, bad)
	assert.Error(t, err)
}

func TestWrapLogFn_UnWrap(t *testing.T) {
	called := false
	biz.WrapLogFn(func(container []*websocket_pb.Container, format string, v ...any) {
		called = true
	})(nil, "", "")
	assert.True(t, called)
}

func TestNewDefaultHelmer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := NewMockDataStore(m)
	k8sRepo := NewMockK8sRepo(m)
	helmer := NewDefaultHelmer(k8sRepo, mockData, &config.Config{}, mlog.NewForConfig(nil)).(*DefaultHelmer)
	assert.NotNil(t, helmer.data)
	assert.NotNil(t, helmer.logger)
	assert.NotNil(t, helmer.k8sRepo)
}

// TestDefaultHelmer_PackageChart 覆盖 PackageChart 导出包装
// 与 packageChart 无依赖分支：打包本地 chart 生成 tgz。
func TestDefaultHelmer_PackageChart(t *testing.T) {
	chartDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(chartDir, "templates"), 0o755))
	chartYaml := "apiVersion: v2\nname: test-chart\nversion: 0.1.0\ntype: application\n"
	require.NoError(t, os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte(chartYaml), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(chartDir, "values.yaml"), []byte("replicaCount: 1\n"), 0o644))

	destDir := t.TempDir()
	got, err := (&DefaultHelmer{}).PackageChart(chartDir, destDir)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(destDir, "test-chart-0.1.0.tgz"), got)
	fi, err := os.Stat(got)
	require.NoError(t, err)
	assert.Greater(t, fi.Size(), int64(0))
}
