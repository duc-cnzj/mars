package helm

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/sirupsen/logrus"

	restclient "k8s.io/client-go/rest"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"

	"k8s.io/apimachinery/pkg/labels"

	v12 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPackageChart(t *testing.T) {}

func TestReleaseList_Add(t *testing.T) {
	rl := ReleaseList{}
	rl.Add(&release.Release{Name: "rl1", Namespace: "dev", Info: &release.Info{Status: "deployed"}})
	rl.Add(&release.Release{Name: "rl2", Namespace: "dev", Info: &release.Info{Status: "pending-upgrade"}})
	rl.Add(&release.Release{Name: "rl3", Namespace: "dev", Info: &release.Info{Status: "pending-rollback"}})
	rl.Add(&release.Release{Name: "rl4", Namespace: "dev", Info: &release.Info{Status: "pending-install"}})
	rl.Add(&release.Release{Name: "rl5", Namespace: "dev", Info: &release.Info{Status: "uninstalling"}})
	rl.Add(&release.Release{Name: "rl6", Namespace: "dev", Info: &release.Info{Status: "failed"}})
	rl.Add(&release.Release{Name: "rl7", Namespace: "dev", Info: &release.Info{Status: "superseded"}})
	rl.Add(&release.Release{Name: "rl8", Namespace: "dev", Info: &release.Info{Status: "unknown"}})
	assert.Len(t, rl, 8)
	_, ok := rl["dev-rl1"]
	assert.True(t, ok)
	assert.Equal(t, "deployed", rl["dev-rl1"].Status)
	assert.Equal(t, "pending", rl["dev-rl2"].Status)
	assert.Equal(t, "pending", rl["dev-rl3"].Status)
	assert.Equal(t, "pending", rl["dev-rl4"].Status)
	assert.Equal(t, "unknown", rl["dev-rl5"].Status)
	assert.Equal(t, "failed", rl["dev-rl6"].Status)
	assert.Equal(t, "unknown", rl["dev-rl7"].Status)
	assert.Equal(t, "unknown", rl["dev-rl8"].Status)
}

func TestReleaseList_GetStatus(t *testing.T) {
	rl := ReleaseList{}
	rl.Add(&release.Release{Name: "rl1", Namespace: "dev", Info: &release.Info{Status: "deployed"}})
	assert.Equal(t, "deployed", rl.GetStatus("dev", "rl1"))
	assert.Equal(t, "unknown", rl.GetStatus("dev", "xxx"))
}

func TestReleaseStatus(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{}).Times(1)
	status := releaseStatus("test", "ns")
	assert.Equal(t, types.Deploy_StatusUnknown, status)
}

func TestRollback(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{}).Times(1)
	err := rollback("test", "ns", false, nil, false)
	assert.Error(t, err)
}

func TestUninstallRelease(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{}).Times(1)
	err := uninstallRelease("test", "ns", func(format string, v ...any) {})
	assert.Error(t, err)
}

func TestUpgradeOrInstall(t *testing.T) {}

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
	assert.Error(t, err)
}

func Test_getActionConfigAndSettings(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{}).Times(1)
	settings := getActionConfigAndSettings("test", func(format string, v ...any) {})
	assert.NotNil(t, settings)
}

func Test_getActionConfigAndSettings1(t *testing.T) {
	getter.All(&cli.EnvSettings{PluginsDirectory: ""})
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{KubeConfig: "xxx"}).Times(2)
	settings := getActionConfigAndSettings("test", func(format string, v ...any) {})
	assert.NotNil(t, settings)
}

func Test_runInstall(t *testing.T) {}

func Test_send(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "po",
			Namespace: "ns",
			Labels: map[string]string{
				"app.kubernetes.io/instance": "app",
			},
		},
		Spec:   v1.PodSpec{},
		Status: v1.PodStatus{},
	}
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	idxer.Add(pod)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fake.NewSimpleClientset(pod), PodLister: v12.NewPodLister(idxer)})
	called := 0
	var str string
	var obj any = &eventv1.Event{
		Note: "aaa",
		Regarding: v1.ObjectReference{
			Name:            "po",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
	}
	event := obj.(*eventv1.Event)
	p := event.Regarding
	get, _ := app.K8sClient().PodLister.Pods(p.Namespace).Get(p.Name)

	for _, value := range get.Labels {
		if value == ("app") {
			func(cs []*types.Container, format string, v ...any) {
				called++
				str = format
			}(nil, event.Note)
			break
		}
	}
	assert.Equal(t, 1, called)
	assert.Equal(t, "aaa", str)
}

func Test_watchEvent(t *testing.T) {
	t.Parallel()
	ctx, cancelFn := context.WithCancel(context.TODO())
	ch := make(chan contracts.Obj[*eventv1.Event], 10)
	go func() {
		ch <- contracts.NewObj(nil, &eventv1.Event{
			Regarding: v1.ObjectReference{
				Namespace: "ns",
				Name:      "app",
			},
		}, contracts.Add)
		ch <- contracts.NewObj(nil, &eventv1.Event{
			Regarding: v1.ObjectReference{
				Namespace: "ns",
				Name:      "app1",
			},
		}, contracts.Update)
		ch <- contracts.NewObj(nil, &eventv1.Event{
			Regarding: v1.ObjectReference{
				Namespace: "ns",
				Name:      "app2",
			},
		}, contracts.Delete)
		time.Sleep(2 * time.Second)
		cancelFn()
	}()
	var called int64
	watchEvent(ctx, ch, "release", func(container []*types.Container, format string, v ...any) {
		atomic.AddInt64(&called, 1)
	}, testutil.NewPodLister(&v1.Pod{
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
	ch := make(chan contracts.Obj[*eventv1.Event], 10)
	go func() {
		close(ch)
		time.Sleep(2 * time.Second)
		cancelFn()
	}()
	var called int64
	watchEvent(ctx, ch, "release", func(container []*types.Container, format string, v ...any) {
		atomic.AddInt64(&called, 1)
	}, testutil.NewPodLister(&v1.Pod{
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
	ch := make(chan contracts.Obj[*eventv1.Event], 10)
	go func() {
		ch <- contracts.NewObj(nil, &eventv1.Event{
			Regarding: v1.ObjectReference{
				Namespace: "ns",
				Name:      "app1",
			},
		}, contracts.Add)
		ch <- contracts.NewObj(nil, &eventv1.Event{
			Regarding: v1.ObjectReference{
				Namespace: "ns",
				Name:      "app",
			},
		}, contracts.Add)
		time.Sleep(2 * time.Second)
		cancelFn()
	}()
	var called int64
	watchEvent(ctx, ch, "release", func(container []*types.Container, format string, v ...any) {
		atomic.AddInt64(&called, 1)
	}, testutil.NewPodLister(&v1.Pod{
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
	podCh := make(chan contracts.Obj[*v1.Pod], 10)
	ctx, cancelFn := context.WithCancel(context.TODO())
	go func() {
		podCh <- contracts.NewObj(nil, &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "app1",
				Labels: map[string]string{
					"name": "app",
				},
			},
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:         "aaa",
						Ready:        false,
						RestartCount: 6,
					},
				},
			},
		}, contracts.Add)
		podCh <- contracts.NewObj(nil, &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "app2",
				Labels: map[string]string{
					"name": "app",
				},
			},
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:         "bbb",
						Ready:        false,
						RestartCount: 5,
					},
				},
			},
		}, contracts.Delete)
		podCh <- contracts.NewObj[*v1.Pod](nil, nil, contracts.Update)
		podCh <- contracts.NewObj[*v1.Pod](nil, nil, contracts.Update)
		podCh <- contracts.NewObj[*v1.Pod](nil, nil, contracts.Update)
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
	watchPodStatus(ctx, podCh, selectorLists, func(container []*types.Container, format string, v ...any) {
		atomic.AddInt64(&called, 1)
	})
	assert.Equal(t, int64(2), atomic.LoadInt64(&called))

	podCh2 := make(chan contracts.Obj[*v1.Pod], 10)
	close(podCh2)
	assert.NotPanics(t, func() {
		watchPodStatus(context.TODO(), podCh2, nil, nil)
	})
}

func Test_watchPodStatus_Error1(t *testing.T) {
	t.Parallel()
	var called int64
	var cs = &ContainerGetterSetter{}
	podCh := make(chan contracts.Obj[*v1.Pod], 10)
	ctx, cancelFn := context.WithCancel(context.TODO())
	go func() {
		podCh <- contracts.NewObj[*v1.Pod](nil, &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "app-not-match",
				Labels: map[string]string{
					"name": "app-not-match",
				},
			},
		}, contracts.Add)
		podCh <- contracts.NewObj[*v1.Pod](nil, &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "app",
				Labels: map[string]string{
					"name": "app",
				},
			},
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:         "three",
						Ready:        false,
						RestartCount: 0,
					},
				},
			},
		}, contracts.Delete)
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
	watchPodStatus(ctx, podCh, selectorLists, func(container []*types.Container, format string, v ...any) {
		cs.Set(container)
		atomic.AddInt64(&called, 1)
	})
	assert.Len(t, cs.Get(), 0)
	assert.Equal(t, int64(0), atomic.LoadInt64(&called))
}

type ContainerGetterSetter struct {
	sync.Mutex
	cs []*types.Container
}

func (c *ContainerGetterSetter) Set(cs []*types.Container) {
	c.Lock()
	defer c.Unlock()
	c.cs = cs
}
func (c *ContainerGetterSetter) Get() []*types.Container {
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
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Debugf("ass").Times(1)
	n, err := (&logWriter{}).Write([]byte("ass"))
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
}

func Test_newDefaultRegistryClient(t *testing.T) {
	client, err := newDefaultRegistryClient(false, "")
	assert.Nil(t, err)
	assert.NotNil(t, client)
}
