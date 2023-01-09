package socket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duc-cnzj/mars-client/v4/mars"
	"github.com/duc-cnzj/mars-client/v4/types"
	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/event"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/plugins/domain_manager"
	"helm.sh/helm/v3/pkg/release"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
)

type fakeChartLoader struct {
	c *chart.Chart
}

func (f *fakeChartLoader) LoadArchive(in io.Reader) (*chart.Chart, error) {
	return f.c, nil
}

func (f *fakeChartLoader) LoadDir(dir string) (*chart.Chart, error) {
	return f.c, nil
}

type fakeOpener struct{}

func (f *fakeOpener) Open(name string) (io.ReadCloser, error) {
	return io.NopCloser(nil), nil
}

func (f *fakeOpener) Close() error {
	return nil
}

func TestChartFileLoader_Load(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := &emptyMsger{}
	h := mock.NewMockHelmer(m)
	job := &Jober{
		helmer: h,
		input: &JobInput{
			GitProjectId: 100,
		},
		messager:  em,
		percenter: &emptyPercenter{},
		config: &mars.Config{
			LocalChartPath: "9999|master|dir",
		},
	}
	l := &ChartFileLoader{
		chartLoader: &fakeChartLoader{
			c: &chart.Chart{
				Metadata: &chart.Metadata{
					Dependencies: []*chart.Dependency{
						{
							Repository: "file://xxxx",
						},
					},
				},
			},
		},
		fileOpener: &fakeOpener{},
	}
	gits := mock.NewMockGitServer(m)
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{GitServerPlugin: config.Plugin{Name: "gits"}}).AnyTimes()
	app.EXPECT().GetPluginByName("gits").Return(gits).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gits.EXPECT().Initialize(gomock.Any()).AnyTimes()

	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return(nil, errors.New("xxx"))

	err := l.Load(job)
	assert.Equal(t, "charts 文件不存在", err.Error())

	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return([]string{"file1", "file2"}, nil)

	gits.EXPECT().GetFileContentWithSha("9999", "master", "file1").Return("file1", nil).Times(1)
	gits.EXPECT().GetFileContentWithSha("9999", "master", "file2").Return("file2", nil).Times(1)
	up := mock.NewMockUploader(m)
	app.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().AbsolutePath(gomock.Any()).Return("")
	up.EXPECT().MkDir(gomock.Any(), false).Times(1)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Times(2)
	h.EXPECT().PackageChart(gomock.Any(), gomock.Any()).Times(1)

	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir/xxxx", true).Return([]string{}, nil)

	err = l.Load(job)
	assert.Len(t, job.destroyFuncs, 2)
	assert.Nil(t, err)
}

func TestChartFileLoader_Load2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := &emptyMsger{}
	h := mock.NewMockHelmer(m)
	job := &Jober{
		helmer: h,
		input: &JobInput{
			GitCommit:    "commit",
			GitProjectId: 100,
		},
		messager:  em,
		percenter: &emptyPercenter{},
		config: &mars.Config{
			LocalChartPath: "dir",
		},
	}
	l := &ChartFileLoader{
		chartLoader: &fakeChartLoader{
			c: &chart.Chart{
				Metadata: &chart.Metadata{},
			},
		},
		fileOpener: &fakeOpener{},
	}
	gits := mock.NewMockGitServer(m)
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{GitServerPlugin: config.Plugin{Name: "gits"}}).AnyTimes()
	app.EXPECT().GetPluginByName("gits").Return(gits).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gits.EXPECT().Initialize(gomock.Any()).AnyTimes()

	gits.EXPECT().GetDirectoryFilesWithSha("100", "commit", "dir", true).Return([]string{"file1", "file2"}, nil)

	up := mock.NewMockUploader(m)
	app.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().MkDir(gomock.Any(), false).Return(errors.New("mkdir err")).Times(1)

	err := l.Load(job)
	assert.Equal(t, "mkdir err", err.Error())
}

func TestDynamicLoader_Load(t *testing.T) {
	em := &emptyMsger{}
	job := &Jober{
		input: &JobInput{
			Config: "xxxx",
		},
		messager:  em,
		percenter: &emptyPercenter{},
		config: &mars.Config{
			ConfigField: "app->config",
			IsSimpleEnv: true,
		},
	}
	assert.Nil(t, (&DynamicLoader{}).Load(job))
	assert.Equal(t,
		`app:
  config: xxxx
`, job.dynamicConfigYaml)
	job2 := &Jober{
		input: &JobInput{
			Config: "name: duc\nage: 17",
		},
		messager:  em,
		percenter: &emptyPercenter{},
		config: &mars.Config{
			ConfigField: "app->config",
			IsSimpleEnv: false,
		},
	}
	assert.Nil(t, (&DynamicLoader{}).Load(job2))
	assert.Equal(t,
		`app:
  config:
    age: 17
    name: duc
`, job2.dynamicConfigYaml)

	em.msgs = []string{}
	job3 := &Jober{
		input:     &JobInput{},
		messager:  em,
		percenter: &emptyPercenter{},
	}
	assert.Nil(t, (&DynamicLoader{}).Load(job3))
	assert.Len(t, em.msgs, 2)
	assert.Equal(t, "[DynamicLoader]: 未发现用户自定义配置", em.msgs[1])
}

func TestExtraValuesLoader_Load(t *testing.T) {
	em := &emptyMsger{}
	job := &Jober{
		input: &JobInput{
			ExtraValues: []*types.ExtraValue{
				{
					Path:  "app->config",
					Value: "1",
				},
			},
		},
		messager:  em,
		percenter: &emptyPercenter{},
		config: &mars.Config{
			Elements: []*mars.Element{
				{
					Path:         "app->config",
					Type:         mars.ElementType_ElementTypeSelect,
					Default:      "1",
					Description:  "replica count",
					SelectValues: []string{"1", "2", "3"},
				},
				{
					Path:        "app->xxx",
					Type:        mars.ElementType_ElementTypeSwitch,
					Default:     "true",
					Description: "bool value",
				},
			},
		},
	}

	assert.Nil(t, (&ExtraValuesLoader{}).Load(job))
	sort.Strings(job.extraValues)
	assert.Equal(t,
		`app:
  config: "1"
`,
		job.extraValues[0])
	assert.Equal(t,
		`app:
  xxx: true
`,
		job.extraValues[1])

	err := (&ExtraValuesLoader{}).Load(&Jober{
		input: &JobInput{
			ExtraValues: []*types.ExtraValue{
				{
					Path:  "app->config",
					Value: "4",
				},
			},
		},
		messager:  em,
		percenter: &emptyPercenter{},
		config: &mars.Config{
			Elements: []*mars.Element{
				{
					Path:         "app->config",
					Type:         mars.ElementType_ElementTypeSelect,
					Default:      "1",
					Description:  "replica count",
					SelectValues: []string{"1", "2", "3"},
				},
			},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, "app->config 必须在 '1,2,3' 里面, 你传的是 4", err.Error())

	j := &Jober{
		input: &JobInput{
			ExtraValues: []*types.ExtraValue{
				{
					Path:  "app->config",
					Value: "4",
				},
				{
					Path:  "duc",
					Value: "xxx",
				},
			},
		},
		messager:  em,
		percenter: &emptyPercenter{},
		config: &mars.Config{
			Elements: []*mars.Element{
				{
					Path:    "duc",
					Type:    mars.ElementType_ElementTypeInput,
					Default: "input",
				},
			},
		},
	}
	err = (&ExtraValuesLoader{}).Load(j)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(strings.Join(em.msgs, " "), "不允许自定义字段 app->config"))
	assert.Equal(t, []string{"duc: xxx\n"}, j.extraValues)

	j2 := &Jober{
		input:     &JobInput{},
		messager:  em,
		percenter: &emptyPercenter{},
		config:    &mars.Config{},
	}
	em.msgs = []string{}
	assert.Nil(t, (&ExtraValuesLoader{}).Load(j2))
	assert.Len(t, em.msgs, 2)
	assert.Equal(t, "[ExtraValuesLoader]: 未发现项目额外的配置", em.msgs[1])
}

func TestExtraValuesLoader_deepSetItems(t *testing.T) {
	items := (&ExtraValuesLoader{}).deepSetItems(map[string]any{"a": "a"})
	assert.Equal(t, "a: a\n", items[0])
	items = (&ExtraValuesLoader{}).deepSetItems(map[string]any{"a->b": "ab"})
	assert.Equal(t,
		`a:
  b: ab
`, items[0])
}

func TestExtraValuesLoader_typeValue(t *testing.T) {
	var tests = []struct {
		ele    *mars.Element
		input  string
		result any
		err    string
	}{
		{
			ele: &mars.Element{
				Type: mars.ElementType_ElementTypeSwitch,
			},
			input:  "",
			result: false,
			err:    "",
		},
		{
			ele: &mars.Element{
				Type: mars.ElementType_ElementTypeSwitch,
			},
			input:  "true",
			result: true,
			err:    "",
		},
		{
			ele: &mars.Element{
				Path: "app->config",
				Type: mars.ElementType_ElementTypeSwitch,
			},
			input: "xxx",
			err:   "app->config 字段类型不正确，应该为 bool，你传入的是 xxx",
		},
		{
			ele: &mars.Element{
				Path: "app->config",
				Type: mars.ElementType_ElementTypeInputNumber,
			},
			input:  "",
			result: int64(0),
			err:    "",
		},
		{
			ele: &mars.Element{
				Path: "app->config",
				Type: mars.ElementType_ElementTypeInputNumber,
			},
			input:  "10",
			result: int64(10),
			err:    "",
		},
		{
			ele: &mars.Element{
				Path: "app->config",
				Type: mars.ElementType_ElementTypeInputNumber,
			},
			input:  "xxx",
			result: nil,
			err:    "app->config 字段类型不正确，应该为整数，你传入的是 xxx",
		},
		{
			ele: &mars.Element{
				Path: "app->config",
				Type: mars.ElementType_ElementTypeRadio,
				SelectValues: []string{
					"a", "b", "c",
				},
			},
			input:  "a",
			result: "a",
			err:    "",
		},
		{
			ele: &mars.Element{
				Path: "app->config",
				Type: mars.ElementType_ElementTypeRadio,
				SelectValues: []string{
					"a", "b", "c",
				},
			},
			input:  "d",
			result: "",
			err:    "app->config 必须在 'a,b,c' 里面, 你传的是 d",
		},
		{
			ele: &mars.Element{
				Path: "app->config",
				Type: mars.ElementType_ElementTypeSelect,
				SelectValues: []string{
					"a", "b", "c",
				},
			},
			input:  "b",
			result: "b",
			err:    "",
		},
		{
			ele: &mars.Element{
				Path: "app->config",
				Type: 10000,
			},
			input:  "xxx",
			result: "xxx",
			err:    "",
		},
	}
	for i, test := range tests {
		tt := test
		t.Run(fmt.Sprintf("test-%v", i), func(t *testing.T) {
			t.Parallel()
			value, err := (&ExtraValuesLoader{}).typeValue(tt.ele, tt.input)
			if err != nil {
				assert.Equal(t, err.Error(), tt.err)
			} else {
				assert.Equal(t, tt.result, value)
			}
		})
	}
}

func TestJober_AddDestroyFunc(t *testing.T) {
	j := &Jober{}
	called := 0
	j.AddDestroyFunc(func() {
		called++
	})
	j.AddDestroyFunc(func() {
		called++
	})
	j.CallDestroyFuncs()
	assert.Equal(t, 2, called)
}

func TestJober_CallDestroyFuncs(t *testing.T) {
	j := &Jober{}
	called := false
	j.AddDestroyFunc(func() {
		called = true
	})
	j.CallDestroyFuncs()
	assert.True(t, called)
}

func TestJober_Commit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := mock.NewMockCommitInterface(m)
	j := &Jober{
		commit: c,
	}
	assert.Same(t, c, j.Commit())
}

func TestJober_Done(t *testing.T) {
	ch := make(chan struct{})
	j := &Jober{
		done: ch,
	}
	fn := func() <-chan struct{} {
		return ch
	}
	assert.Equal(t, fn(), j.Done())
}

func TestJober_Finish(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	ps := mock.NewMockPubSub(m)
	j := &Jober{
		pubsub: ps,
		done:   make(chan struct{}),
		ns:     &models.Namespace{ID: 1},
	}

	ps.EXPECT().ToAll(reloadProjectsMessage(1)).Times(1)
	j.Finish()
	_, ok := <-j.done
	assert.False(t, ok)
}

func TestJober_GetStoppedErrorIfHas(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msg := mock.NewMockDeployMsger(m)
	msg.EXPECT().SendMsg(gomock.Any())
	context, f := utils.NewCustomErrorContext()
	j := &Jober{
		slugName: "aaa",
		messager: msg,
		stopCtx:  context,
		stopFn:   f,
	}
	assert.Nil(t, j.GetStoppedErrorIfHas())
	j.Stop(errors.New("xxx"))
	assert.Error(t, j.GetStoppedErrorIfHas())
}

func TestJober_HandleMessage_DoneClosed(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Done().Return(nil)
	ch := make(chan contracts.MessageItem, 10)
	done := make(chan struct{}, 1)
	msger := mock.NewMockDeployMsger(m)
	close(done)
	j := &Jober{
		done:     done,
		messager: msger,
		messageCh: &SafeWriteMessageCh{
			ch: ch,
		},
	}
	j.HandleMessage()
}

func TestJober_HandleMessage_AppDoneClosed(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	done := make(chan struct{}, 1)
	close(done)
	app := testutil.MockApp(m)
	app.EXPECT().Done().Return(done)
	ch := make(chan contracts.MessageItem, 10)
	msger := mock.NewMockDeployMsger(m)
	j := &Jober{
		done:     nil,
		messager: msger,
		messageCh: &SafeWriteMessageCh{
			ch: ch,
		},
	}
	j.HandleMessage()
}

func TestJober_HandleMessage_TextMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Done().Return(nil).AnyTimes()
	ch := make(chan contracts.MessageItem, 10)
	msger := mock.NewMockDeployMsger(m)
	j := &Jober{
		done:     nil,
		messager: msger,
		messageCh: &SafeWriteMessageCh{
			ch: ch,
		},
	}
	cs := []*types.Container{
		{
			Namespace: "ns",
			Pod:       "pod",
			Container: "c",
		},
	}
	go func() {
		ch <- contracts.MessageItem{Msg: "aa", Type: contracts.MessageText, Containers: cs}
		close(ch)
	}()
	msger.EXPECT().SendMsgWithContainerLog("aa", cs).Times(1)
	j.HandleMessage()
}

func TestJober_HandleMessage_ErrorMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Done().Return(nil).AnyTimes()
	ch := make(chan contracts.MessageItem, 10)
	msger := mock.NewMockDeployMsger(m)
	j := &Jober{
		project:  &models.Project{},
		stopCtx:  context.TODO(),
		dryRun:   true,
		done:     nil,
		messager: msger,
		messageCh: &SafeWriteMessageCh{
			ch: ch,
		},
	}
	go func() {
		ch <- contracts.MessageItem{Msg: "errors", Type: contracts.MessageError}
		close(ch)
	}()
	msger.EXPECT().SendDeployedResult(ResultDeployFailed, "errors", gomock.Any()).Times(1)
	j.HandleMessage()
}

func TestJober_HandleMessage_SuccessMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Done().Return(nil).AnyTimes()
	ch := make(chan contracts.MessageItem, 10)
	msger := mock.NewMockDeployMsger(m)
	j := &Jober{
		project:  &models.Project{},
		stopCtx:  context.TODO(),
		dryRun:   true,
		done:     nil,
		messager: msger,
		messageCh: &SafeWriteMessageCh{
			ch: ch,
		},
	}
	go func() {
		ch <- contracts.MessageItem{Msg: "ok", Type: contracts.MessageSuccess}
		close(ch)
	}()
	msger.EXPECT().SendDeployedResult(ResultDeployed, "ok", gomock.Any()).Times(1)
	j.HandleMessage()
}

func TestJober_HandleMessage_UserCanceled(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Done().Return(nil).AnyTimes()
	ch := make(chan contracts.MessageItem, 10)
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()
	msger := mock.NewMockDeployMsger(m)

	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	p := &models.Project{
		Name:         "aaa",
		GitProjectId: 100,
		GitBranch:    "dev",
		GitCommit:    "xxx",
		Namespace:    models.Namespace{Name: "dev-aaa"},
	}
	db.Create(&p)
	assert.Greater(t, p.ID, 0)
	j := &Jober{
		project:  p,
		stopCtx:  ctx,
		dryRun:   false,
		isNew:    true,
		done:     nil,
		messager: msger,
		messageCh: &SafeWriteMessageCh{
			ch: ch,
		},
	}
	go func() {
		ch <- contracts.MessageItem{Msg: "errors", Type: contracts.MessageError}
		close(ch)
	}()
	msger.EXPECT().SendDeployedResult(ResultDeployCanceled, gomock.Any(), gomock.Any()).Times(1)
	j.HandleMessage()
	newp := &models.Project{}
	db.Unscoped().Where("`id` = ?", p.ID).First(newp)
	assert.True(t, newp.DeletedAt.Valid)
}

func TestJober_ID(t *testing.T) {
	j := &Jober{
		slugName: "aaa",
	}
	assert.Equal(t, "aaa", j.ID())
}

func TestJober_IsDryRun(t *testing.T) {
	j := &Jober{
		dryRun: true,
	}
	assert.True(t, j.IsDryRun())
}

func TestJober_IsNew(t *testing.T) {
	j := &Jober{
		isNew: true,
	}
	assert.True(t, j.IsNew())
}

func TestJober_IsStopped(t *testing.T) {
	j := &Jober{
		stopCtx: context.TODO(),
	}
	assert.False(t, j.IsStopped())
	cancel, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	j2 := &Jober{
		stopCtx: cancel,
	}
	assert.True(t, j2.IsStopped())
}

type emptyLoader struct {
	err    error
	called bool
	sync.Mutex
}

func (e *emptyLoader) Load(jober *Jober) error {
	e.Lock()
	defer e.Unlock()
	e.called = true
	return e.err
}

func (e *emptyLoader) GetCalled() bool {
	e.Lock()
	defer e.Unlock()

	return e.called
}

func TestJober_LoadConfigs1(t *testing.T) {
	ctx, fn := context.WithCancel(context.TODO())
	fn()
	l := &emptyLoader{}
	assert.Equal(t, "context canceled", (&Jober{
		messager: &emptyMsger{},
		stopCtx:  ctx,
		loaders:  []Loader{l},
	}).LoadConfigs().Error())
	assert.False(t, l.GetCalled())
}

func TestJober_LoadConfigs(t *testing.T) {
	l := &emptyLoader{}
	assert.Nil(t, (&Jober{
		messager: &emptyMsger{},
		stopCtx:  context.TODO(),
		loaders:  []Loader{l},
	}).LoadConfigs())
	assert.True(t, l.GetCalled())

	l2 := &emptyLoader{}
	cancel, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	assert.Equal(t, "context canceled", (&Jober{
		messager: &emptyMsger{},
		stopCtx:  cancel,
		loaders:  []Loader{l2},
	}).LoadConfigs().Error())
	assert.False(t, l2.GetCalled())

	l3 := &emptyLoader{
		err: errors.New("xxx"),
	}
	assert.Equal(t, "xxx", (&Jober{
		messager: &emptyMsger{},
		stopCtx:  context.TODO(),
		loaders:  []Loader{l3},
	}).LoadConfigs().Error())
	assert.True(t, l3.GetCalled())
}

func TestJober_Logs(t *testing.T) {
	ri := newReleaseInstaller("a", "a", nil, nil, false, 10, false)
	j := &Jober{
		installer: ri,
	}
	ri.logs = &timeOrderedSetString{}
	assert.Equal(t, ri.logs.sortedItems(), j.Logs())
}

func TestJober_Manifests(t *testing.T) {
	j := &Jober{
		manifests: []string{"xxx"},
	}
	assert.Equal(t, []string{"xxx"}, j.Manifests())
}

func TestJober_Messager(t *testing.T) {
	m := NewMessageSender(nil, "", 0)
	j := &Jober{
		messager: m,
	}
	assert.Same(t, m, j.Messager())
}

func TestJober_Namespace(t *testing.T) {
	ns := &models.Namespace{Name: "aaa"}
	j := &Jober{
		ns: ns,
	}
	assert.Same(t, ns, j.Namespace())
}

func TestJober_Percenter(t *testing.T) {
	p := newProcessPercent(nil, nil)
	j := &Jober{
		percenter: p,
	}
	assert.Same(t, p, j.Percenter())
}

func TestJober_Project(t *testing.T) {
	p := &models.Project{}
	j := &Jober{
		project: p,
	}
	assert.Same(t, p, j.Project())
}

func TestJober_ProjectModel(t *testing.T) {
	j := &Jober{}
	assert.Nil(t, j.ProjectModel())
	j.project = &models.Project{Name: "aa"}
	assert.NotNil(t, j.ProjectModel())
}

func TestJober_Prune(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, cfn := testutil.SetGormDB(m, app)
	defer cfn()
	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	p := &models.Project{Name: "app", Namespace: models.Namespace{Name: "aaa"}}
	db.Create(p)
	j := &Jober{
		isNew:   true,
		dryRun:  false,
		project: &models.Project{ID: p.ID},
	}
	var c int64
	db.Model(&models.Project{}).Count(&c)
	assert.Equal(t, int64(1), c)
	j.Prune()
	db.Model(&models.Project{}).Count(&c)
	assert.Equal(t, int64(0), c)

	// 重置 version
	j2 := &Jober{
		isNew:       false,
		dryRun:      false,
		project:     &models.Project{ID: p.ID, Version: 101},
		prevProject: &models.Project{ID: p.ID, Version: 100},
	}
	j2.Prune()
	assert.Equal(t, 100, j2.project.Version)
}

func TestJober_PubSub(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	ps := mock.NewMockPubSub(m)
	j := &Jober{
		pubsub: ps,
	}
	assert.Same(t, ps, j.PubSub())
}

func TestJober_ReleaseInstaller(t *testing.T) {
	ri := &releaseInstaller{}
	j := &Jober{
		installer: ri,
	}
	assert.Same(t, ri, j.ReleaseInstaller())
}

func TestJober_Run_Fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	app := testutil.MockApp(m)
	app.EXPECT().Done().Return(nil).AnyTimes()

	db, closeFn := testutil.SetGormDB(m, app)
	defer closeFn()
	assert.Nil(t, db.AutoMigrate(&models.Project{}, &models.Namespace{}, &models.Event{}, &models.Changelog{}))

	msgCh := &SafeWriteMessageCh{
		ch: make(chan contracts.MessageItem, 100),
	}
	proj := &models.Project{
		Name:         "app",
		DeployStatus: uint8(types.Deploy_StatusDeploying),
		Namespace: models.Namespace{
			Name: "ns",
		},
	}
	assert.Nil(t, db.Create(proj).Error)

	stopCtx := context.TODO()
	rinstaller := mock.NewMockReleaseInstaller(m)

	commit := mock.NewMockCommitInterface(m)
	commit.EXPECT().GetTitle().Return("xxx").Times(1)
	rinstaller.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "xxx").Return(nil, errors.New("xxx"))
	job := &Jober{
		messager:  &emptyMsger{},
		project:   proj,
		isNew:     true,
		dryRun:    false,
		installer: rinstaller,
		stopCtx:   stopCtx,
		messageCh: msgCh,
		commit:    commit,
	}
	assert.Equal(t, "xxx", job.Run().Error())
	assert.Equal(t, "xxx", job.messager.(*emptyMsger).msgs[0])
	newProj := &models.Project{}
	assert.Nil(t, db.Unscoped().Where("`id` = ?", proj.ID).First(&newProj).Error)
	assert.True(t, newProj.DeletedAt.Valid)
}

func TestJober_Run_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	app := testutil.MockApp(m)
	app.EXPECT().Done().Return(nil).AnyTimes()

	db, closeFn := testutil.SetGormDB(m, app)
	defer closeFn()
	assert.Nil(t, db.AutoMigrate(&models.Project{}, &models.Namespace{}, &models.Event{}, &models.Changelog{}, &models.GitProject{}))

	msgCh := &SafeWriteMessageCh{
		ch: make(chan contracts.MessageItem, 100),
	}
	proj := &models.Project{
		GitProjectId:     100,
		Name:             "app",
		DeployStatus:     uint8(types.Deploy_StatusDeploying),
		ExtraValues:      `[{"path": "app->config", "value": "xxx"}]`,
		FinalExtraValues: `["xx", "yy"]`,
		Namespace: models.Namespace{
			Name: "ns",
		},
	}
	gp := &models.GitProject{
		DefaultBranch: "dev",
		Name:          "busyapp",
		GitProjectId:  100,
	}
	db.Create(gp)
	assert.Nil(t, db.Create(proj).Error)

	stopCtx := context.TODO()
	rinstaller := mock.NewMockReleaseInstaller(m)
	rinstaller.EXPECT().Chart().Return(&chart.Chart{Metadata: &chart.Metadata{Name: "busyapp"}, Values: map[string]any{}}).Times(1)
	manifest := `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: busybox
  name: busybox
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - command:
        - sh
        - -c
        - sleep 3600;
        image: busybox:latest
        imagePullPolicy: Always
        name: busybox
        resources:
          limits:
            cpu: 10m
            memory: 10Mi
          requests:
            cpu: 10m
            memory: 10Mi
`
	rinstaller.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "title").Return(&release.Release{
		Config:   map[string]any{},
		Manifest: manifest,
	}, nil)

	commit := mock.NewMockCommitInterface(m)
	commit.EXPECT().GetTitle().Return("title").Times(2)
	commit.EXPECT().GetWebURL().Return("url")
	commit.EXPECT().GetAuthorName().Return("duc")
	date := time.Now()
	commit.EXPECT().GetCommittedDate().Return(&date)

	gits := mock.NewMockGitServer(m)
	app.EXPECT().Config().Return(&config.Config{GitServerPlugin: config.Plugin{Name: "gits"}}).AnyTimes()
	app.EXPECT().GetPluginByName("gits").Return(gits)
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gits.EXPECT().Initialize(gomock.Any()).AnyTimes()

	commit2 := mock.NewMockCommitInterface(m)
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(commit2, nil)
	commit2.EXPECT().GetWebURL().Return("weburl2").Times(1)
	commit2.EXPECT().GetTitle().Return("title2").Times(1)

	d := event.NewDispatcher(app)
	d.Listen(events.EventProjectChanged, events.HandleProjectChanged)
	d.Listen(events.EventAuditLog, events.HandleAuditLog)
	app.EXPECT().EventDispatcher().Return(d).AnyTimes()

	job := &Jober{
		percenter: &emptyPercenter{},
		user: contracts.UserInfo{
			Name: "duc",
		},
		ns: &proj.Namespace,
		input: &JobInput{
			ExtraValues: []*types.ExtraValue{{
				Path:  "app->config",
				Value: "xxx",
			}},
		},
		prevProject: proj,
		config: &mars.Config{
			ConfigFileType: "go",
		},
		commit:      commit,
		messager:    &emptyMsger{},
		project:     proj,
		isNew:       false,
		dryRun:      false,
		installer:   rinstaller,
		stopCtx:     stopCtx,
		messageCh:   msgCh,
		extraValues: []string{"ex-aaa", "ex-bbb"},
		vars:        map[string]any{"var-a": "aaa", "var-b": "bbb"},
	}
	assert.Nil(t, job.Run())
	assert.Equal(t, "部署成功", job.messager.(*emptyMsger).msgs[0])

	assert.Equal(t, "{}\n", job.Project().OverrideValues)
	assert.Equal(t, manifest, job.Project().Manifest)
	assert.Equal(t, "app=busybox", job.Project().PodSelectors)
	assert.Equal(t, "busybox:latest", job.Project().DockerImage)
	assert.Equal(t, "title", job.Project().GitCommitTitle)
	assert.Equal(t, "url", job.Project().GitCommitWebUrl)
	assert.Equal(t, "duc", job.Project().GitCommitAuthor)
	assert.Equal(t, date.String(), job.Project().GitCommitDate.String())
	assert.Equal(t, "go", job.Project().ConfigType)
	assert.Equal(t, `[{"path":"app-\u003econfig","value":"xxx"}]`, job.Project().ExtraValues)
	assert.Equal(t, `["ex-aaa","ex-bbb"]`, job.Project().FinalExtraValues)
	assert.Equal(t, `{"var-a":"aaa","var-b":"bbb"}`, job.Project().EnvValues)

	var clog models.Changelog
	db.First(&clog)
	assert.Equal(t, int(1), int(clog.Version))
	assert.Equal(t, manifest, clog.Manifest)
	assert.Equal(t, gp.ID, clog.GitProjectID)

	var adlog models.Event
	db.First(&adlog)
	assert.Equal(t, "duc", adlog.Username)
	assert.Equal(t, types.EventActionType_Update, types.EventActionType(adlog.Action))
	assert.Equal(t,
		`config: ""
branch: ""
commit: ""
atomic: false
web_url: weburl2
title: title2
extra_values:
- path: app->config
  value: xxx
final_extra_values: |
  {}
env_values:
  var-a: aaa
  var-b: bbb
`, adlog.Old)
	assert.Equal(t,
		`config: ""
branch: ""
commit: ""
atomic: false
web_url: url
title: title
extra_values:
- path: app->config
  value: xxx
final_extra_values: |
  {}
env_values:
  var-a: aaa
  var-b: bbb
`, adlog.New)
}

func TestJober_Stop(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msg := mock.NewMockDeployMsger(m)
	msg.EXPECT().SendMsg(gomock.Any()).Times(1)
	called := 0
	j := &Jober{messager: msg, stopFn: func(err error) {
		called++
	}}
	wg := sync.WaitGroup{}
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()

			j.Stop(nil)
		}()
	}
	wg.Wait()
	assert.Equal(t, 1, called)
}

func TestJober_User(t *testing.T) {
	j := &Jober{}
	assert.IsType(t, contracts.UserInfo{}, j.User())
}

func TestJober_Validate(t *testing.T) {
	job := &Jober{
		wsType: 99999,
	}
	err := job.Validate()
	assert.Equal(t, "type error: 99999", err.Error())

	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	db.AutoMigrate(&models.Project{}, &models.Namespace{}, &models.GitProject{})

	job2 := &Jober{
		input: &JobInput{
			NamespaceId: 9999,
		},
		messager:  &emptyMsger{},
		percenter: &emptyPercenter{},
		wsType:    websocket_pb.Type_CreateProject,
	}
	assert.Equal(t, "[FAILED]: 校验名称空间: record not found", job2.Validate().Error())

	ns := &models.Namespace{Name: "ns", ImagePullSecrets: "aa,bb"}
	db.Create(ns)
	marsC := mars.Config{
		DisplayName:    "app",
		ConfigFileType: "go",
	}
	marshal, _ := json.Marshal(&marsC)
	gp := &models.GitProject{
		DefaultBranch: "dev",
		Name:          "git-app",
		GitProjectId:  100,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	}
	db.Create(gp)
	gits := mock.NewMockGitServer(m)
	app.EXPECT().Config().Return(&config.Config{GitServerPlugin: config.Plugin{Name: "gits"}}).AnyTimes()
	app.EXPECT().GetPluginByName("gits").Return(gits).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gits.EXPECT().Initialize(gomock.Any()).AnyTimes()
	commit := mock.NewMockCommitInterface(m)
	h := mock.NewMockHelmer(m)
	ps := mock.NewMockPubSub(m)
	var newJob3 = func() *Jober {
		return &Jober{
			pubsub: ps,
			helmer: h,
			input: &JobInput{
				NamespaceId:  int64(ns.ID),
				GitProjectId: 100,
				GitBranch:    "dev",
				GitCommit:    "commit",
				Config:       "xxx",
				Atomic:       true,
			},
			messager:  &emptyMsger{},
			percenter: &emptyPercenter{},
			wsType:    websocket_pb.Type_CreateProject,
			dryRun:    false,
		}
	}
	job3 := newJob3()
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(commit, nil).Times(1)
	ps.EXPECT().ToSelf(reloadProjectsMessage(ns.ID)).Times(1)
	// 正常创建
	assert.Nil(t, job3.Validate())
	assert.Equal(t, 1, len(job3.destroyFuncs))
	assert.Equal(t, "app", job3.input.Name)
	assert.Same(t, commit, job3.Commit())
	assert.Equal(t, []string{"aa", "bb"}, job3.imagePullSecrets)
	var p models.Project
	db.First(&p)
	assert.Equal(t, uint8(types.Deploy_StatusDeploying), p.DeployStatus)
	assert.Equal(t, int(1), p.Version)
	assert.Equal(t, 100, int(p.GitProjectId))
	assert.Equal(t, "dev", p.GitBranch)
	assert.Equal(t, "commit", p.GitCommit)
	assert.Equal(t, "xxx", p.Config)
	assert.Equal(t, true, p.Atomic)
	assert.Equal(t, "go", p.ConfigType)
	assert.Nil(t, job3.prevProject)

	// 创建后状态变成了 types.Deploy_StatusDeploying
	job3.input.Version = 1
	assert.Equal(t, "有别人也在操作这个项目，等等哦~", job3.Validate().Error())

	db.Model(&p).UpdateColumn("deploy_status", types.Deploy_StatusDeployed)
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(commit, nil).Times(1)
	job3 = newJob3()
	job3.input.Version = 2
	ps.EXPECT().ToSelf(reloadProjectsMessage(ns.ID)).Times(1)
	// 正常返回 commit，设置 prevProject
	assert.Nil(t, job3.Validate())
	assert.NotNil(t, job3.prevProject)

	marshal2, _ := json.Marshal(&mars.Config{
		DisplayName: "",
	})
	db.Model(&gp).UpdateColumn("global_config", string(marshal2))
	gitproj := mock.NewMockProjectInterface(m)
	gitproj.EXPECT().GetName().Return("app-git")
	gits.EXPECT().GetProject("100").Return(gitproj, nil)
	job3 = newJob3()
	job3.input.Name = ""
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(commit, nil).Times(1)
	ps.EXPECT().ToSelf(reloadProjectsMessage(ns.ID)).Times(1)
	// inputName 设置为空, 并且置空 displayName, app name 就会变成 git proj name
	// 此时应该为创建, version = 1
	assert.Nil(t, job3.Validate())
	assert.Equal(t, "app-git", job3.input.Name)
	var pp = models.Project{ID: job3.project.ID}
	db.First(&pp)
	assert.Equal(t, 1, pp.Version)

	h.EXPECT().ReleaseStatus("app-git", "ns").Return(types.Deploy_StatusUnknown).AnyTimes()
	job3.CallDestroyFuncs()
	assert.Equal(t, uint8(types.Deploy_StatusUnknown), job3.project.DeployStatus)

	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(nil, errors.New("aaa")).Times(1)
	job3.input.Version = int64(pp.Version)
	ps.EXPECT().ToSelf(reloadProjectsMessage(ns.ID)).Times(1)
	// GetCommit 返回 error 的情景
	assert.Equal(t, "aaa", job3.Validate().Error())
}

func TestJober_Validate_VersionMatch(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	db.AutoMigrate(&models.Project{}, &models.Namespace{}, &models.GitProject{})

	ns := &models.Namespace{Name: "ns", ImagePullSecrets: "aa,bb"}
	db.Create(ns)
	marsC := mars.Config{
		DisplayName:    "app",
		ConfigFileType: "go",
	}
	marshal, _ := json.Marshal(&marsC)
	gp := &models.GitProject{
		DefaultBranch: "dev",
		Name:          "git-app",
		GitProjectId:  100,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	}
	db.Create(gp)
	p := models.Project{
		NamespaceId:  ns.ID,
		GitProjectId: 100,
		Name:         "app",
		Version:      1,
		DeployStatus: uint8(types.Deploy_StatusDeployed),
	}
	assert.Nil(t, db.Create(&p).Error)
	gits := mock.NewMockGitServer(m)
	app.EXPECT().Config().Return(&config.Config{GitServerPlugin: config.Plugin{Name: "gits"}}).AnyTimes()
	app.EXPECT().GetPluginByName("gits").Return(gits).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gits.EXPECT().Initialize(gomock.Any()).AnyTimes()
	h := mock.NewMockHelmer(m)
	ps := mock.NewMockPubSub(m)
	job := &Jober{
		pubsub: ps,
		helmer: h,
		input: &JobInput{
			NamespaceId:  int64(ns.ID),
			GitProjectId: 100,
			Name:         p.Name,
			Version:      0,
		},
		messager:  &emptyMsger{},
		percenter: &emptyPercenter{},
		wsType:    websocket_pb.Type_UpdateProject,
	}
	assert.ErrorIs(t, ErrorVersionNotMatched, job.Validate())
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	ps.EXPECT().ToSelf(reloadProjectsMessage(ns.ID)).AnyTimes()

	wg := sync.WaitGroup{}
	var successedTimes int64
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			job2 := &Jober{
				pubsub: ps,
				helmer: h,
				input: &JobInput{
					NamespaceId:  int64(ns.ID),
					GitProjectId: 100,
					Name:         p.Name,
					Version:      1,
				},
				messager:  &emptyMsger{},
				percenter: &emptyPercenter{},
				wsType:    websocket_pb.Type_ApplyProject,
			}
			if job2.Validate() == nil {
				atomic.AddInt64(&successedTimes, 1)
			}
		}()
	}
	wg.Wait()
	var pp = models.Project{ID: p.ID}
	db.First(&pp)
	assert.Equal(t, int64(1), atomic.LoadInt64(&successedTimes))
	assert.Equal(t, int(2), pp.Version)
}

func Test_DisplayNameValidate(t *testing.T) {
	var tests = []struct {
		name     string
		wantsErr bool
	}{
		{
			name:     "-a",
			wantsErr: true,
		},
		{
			name:     "A-a_aA",
			wantsErr: false,
		},
		{
			name:     "a",
			wantsErr: false,
		},
		{
			name:     "a-",
			wantsErr: true,
		},
		{
			name:     "a-a",
			wantsErr: false,
		},
		{
			name:     "a-_a",
			wantsErr: false,
		},
		{
			name:     "a-_a_",
			wantsErr: true,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantsErr, (&mars.Config{DisplayName: tt.name}).Validate() != nil)
		})
	}
}

type dump struct {
	assertFn func(x any)
}

func (d *dump) Matches(x any) bool {
	d.assertFn(x)
	return true
}

func (d *dump) String() string {
	return ""
}

func TestMergeValuesLoader_Load(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().LocalUploader().Return(up).AnyTimes()
	finfo := mock.NewMockFileInfo(m)
	finfo.EXPECT().Path().Return("/app/config.yaml")
	up.EXPECT().Put(gomock.Any(), &dump{assertFn: func(x any) {
		all, _ := io.ReadAll(x.(io.Reader))
		result := map[string]any{}
		yaml.Unmarshal(all, &result)
		assert.Equal(t, 1, result["app"].(map[any]any)["one"])
		assert.Equal(t, "two", result["app"].(map[any]any)["two"])
		assert.Equal(t, 3, result["app"].(map[any]any)["three"])
		assert.Equal(t, 4, result["app"].(map[any]any)["four"])
		assert.Equal(t, []any{map[any]any{"name": "secret"}}, result["imagePullSecrets"])
	}}).Return(finfo, nil)
	vy := `
app:
  one: one
  two: 2
`
	dcy := `
app:
  one: 1
  two: two
`
	ev1 := `
app:
  three: 3
`
	ev2 := `
app:
  four: 4
`
	job := &Jober{
		dynamicConfigYaml: dcy,
		imagePullSecrets:  []string{"secret"},
		valuesYaml:        vy,
		extraValues:       []string{ev1, ev2},
		valuesOptions:     &values.Options{},
		input:             &JobInput{GitProjectId: 99, GitBranch: "dev"},
		messager:          &emptyMsger{},
		percenter:         &emptyPercenter{},
	}
	assert.Nil(t, (&MergeValuesLoader{}).Load(job))
	assert.Equal(t, "/app/config.yaml", job.valuesOptions.ValueFiles[0])

	job2 := &Jober{
		dynamicConfigYaml: "",
		imagePullSecrets:  nil,
		valuesYaml:        "",
		extraValues:       nil,
		valuesOptions:     &values.Options{},
		input:             &JobInput{GitProjectId: 99, GitBranch: "dev"},
		messager:          &emptyMsger{},
		percenter:         &emptyPercenter{},
	}
	assert.Nil(t, (&MergeValuesLoader{}).Load(job2))
}

func TestNewJober(t *testing.T) {
	assert.Implements(t, (*contracts.Job)(nil), NewJober(&JobInput{}, contracts.UserInfo{}, "", nil, nil, 10))
	jober := NewJober(&JobInput{}, contracts.UserInfo{}, "", nil, nil, 0, WithDryRun())
	assert.True(t, jober.IsDryRun())
}

type emptyMsger struct {
	sync.Mutex
	msgs   []string
	called int
	contracts.DeployMsger
}

func (e *emptyMsger) SendDeployedResult(t websocket_pb.ResultType, msg string, p *types.ProjectModel) {
	e.Lock()
	defer e.Unlock()
	if e.msgs == nil {
		e.msgs = []string{}
	}
	e.msgs = append(e.msgs, msg)
	e.called++
}

func (e *emptyMsger) SendMsg(s string) {
	e.Lock()
	defer e.Unlock()
	if e.msgs == nil {
		e.msgs = []string{}
	}
	e.msgs = append(e.msgs, s)
	e.called++
}

type emptyPercenter struct {
	called int
	contracts.Percentable
}

func (e *emptyPercenter) To(percent int64) {
	e.called++
}

func TestReleaseInstallerLoader_Load(t *testing.T) {
	ep := &emptyPercenter{}
	em := &emptyMsger{}
	job := &Jober{
		chart:          &chart.Chart{},
		valuesOptions:  &values.Options{},
		percenter:      ep,
		messager:       em,
		dryRun:         true,
		timeoutSeconds: 20,
		input: &JobInput{
			Atomic: true,
		},
		project: &models.Project{
			Name: "app",
		},
		ns: &models.Namespace{
			Name: "ns",
		},
	}
	(&ReleaseInstallerLoader{}).Load(job)
	assert.Equal(t, 1, ep.called)
	assert.Equal(t, 1, em.called)
	in := job.installer.(*releaseInstaller)
	assert.Equal(t, true, in.dryRun)
	assert.Equal(t, int64(20), in.timeoutSeconds)
	assert.Equal(t, "app", in.releaseName)
	assert.Equal(t, "ns", in.namespace)
	assert.Equal(t, true, in.wait)
	assert.Same(t, job.valuesOptions, in.valueOpts)
	assert.Same(t, job.chart, in.chart)
}

func TestSafeWriteMessageCh_Chan(t *testing.T) {
	ch := make(chan contracts.MessageItem, 10)
	sc := &SafeWriteMessageCh{
		ch: ch,
	}
	fn := func() <-chan contracts.MessageItem {
		return ch
	}
	assert.Equal(t, fn(), sc.Chan())
}

func TestSafeWriteMessageCh_Closed(t *testing.T) {
	ch := make(chan contracts.MessageItem, 10)
	sc := &SafeWriteMessageCh{
		ch: ch,
	}
	sc.Closed()
	_, ok := <-ch
	assert.False(t, ok)
}

func TestSafeWriteMessageCh_Send(t *testing.T) {
	ch := make(chan contracts.MessageItem, 10)
	sc := &SafeWriteMessageCh{
		ch: ch,
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sc.Send(contracts.MessageItem{})
		}()
	}
	wg.Wait()
	assert.Len(t, ch, 2)
	sc.Closed()
	sc.Send(contracts.MessageItem{})
	assert.Len(t, ch, 2)

}

func TestVariableLoader_Load(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{
		GitServerPlugin:     config.Plugin{Name: "gits"},
		DomainManagerPlugin: config.Plugin{Name: "domain"},
	}).AnyTimes()

	gitS := mock.NewMockGitServer(m)
	app.EXPECT().GetPluginByName("gits").Return(gitS).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gitS.EXPECT().Initialize(gomock.All()).AnyTimes()
	pipe := mock.NewMockPipelineInterface(m)
	pipe.EXPECT().GetID().Return(int64(9999))
	pipe.EXPECT().GetRef().Return("dev")
	gitS.EXPECT().GetCommitPipeline(gomock.Any(), gomock.Any()).Return(pipe, nil)

	app.EXPECT().GetPluginByName("domain").Return(&domain_manager.DefaultDomainManager{}).AnyTimes()

	em := &emptyMsger{}
	commit := mock.NewMockCommitInterface(m)
	commit.EXPECT().GetShortID().Return("short_id").AnyTimes()
	job := &Jober{
		commit: commit,
		config: &mars.Config{
			ValuesYaml: `
VarImagePullSecrets: <.ImagePullSecrets>
image: <.Pipeline>-<.Branch>
`,
		},
		project: &models.Project{
			Name:      "app",
			GitBranch: "dev",
		},
		ns:               &models.Namespace{Name: "ns"},
		imagePullSecrets: []string{"a", "b", "c"},
		messager:         em,
		percenter:        &emptyPercenter{},
	}
	assert.Nil(t, (&VariableLoader{}).Load(job))
	assert.Equal(t, `
VarImagePullSecrets: [{name: a},{name: b},{name: c},]
image: 9999-dev
`,
		job.valuesYaml)
	assert.Equal(t, "dev", job.vars[VarBranch])
	assert.Equal(t, "short_id", job.vars[VarCommit])
	assert.Equal(t, int64(9999), job.vars[VarPipeline])
	assert.Equal(t, "[{name: a},{name: b},{name: c},]", job.vars[VarImagePullSecrets])

	em2 := &emptyMsger{}
	job2 := &Jober{
		commit: commit,
		config: &mars.Config{
			ValuesYaml: "",
		},
		project: &models.Project{
			Name:      "app",
			GitBranch: "dev",
		},
		ns:               &models.Namespace{Name: "ns"},
		imagePullSecrets: []string{"a", "b", "c"},
		messager:         em2,
		percenter:        &emptyPercenter{},
	}
	assert.Nil(t, (&VariableLoader{}).Load(job2))
	assert.Equal(t, "[VariableLoader]: 未发现可用的 values.yaml", em2.msgs[1])
}

func TestVariableLoader_Load_ok(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{
		GitServerPlugin:     config.Plugin{Name: "gits"},
		DomainManagerPlugin: config.Plugin{Name: "domain"},
	}).AnyTimes()

	gitS := mock.NewMockGitServer(m)
	app.EXPECT().GetPluginByName("gits").Return(gitS).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gitS.EXPECT().Initialize(gomock.All()).AnyTimes()
	gitS.EXPECT().GetCommitPipeline(gomock.Any(), gomock.Any()).Return(nil, errors.New("xxx"))

	app.EXPECT().GetPluginByName("domain").Return(&domain_manager.DefaultDomainManager{}).AnyTimes()

	em := &emptyMsger{}
	commit := mock.NewMockCommitInterface(m)
	commit.EXPECT().GetShortID().Return("short_id").AnyTimes()
	job := &Jober{
		commit: commit,
		config: &mars.Config{
			ValuesYaml: `
VarImagePullSecrets: <.ImagePullSecrets>
`,
		},
		project: &models.Project{
			Name:      "app",
			GitBranch: "dev",
		},
		ns:               &models.Namespace{Name: "ns"},
		imagePullSecrets: []string{"a", "b", "c"},
		messager:         em,
		percenter:        &emptyPercenter{},
	}
	assert.Nil(t, (&VariableLoader{}).Load(job))
	assert.Equal(t, `
VarImagePullSecrets: [{name: a},{name: b},{name: c},]
`,
		job.valuesYaml)
	assert.Equal(t, "dev", job.vars[VarBranch])
	assert.Equal(t, "short_id", job.vars[VarCommit])
	assert.Equal(t, int64(0), job.vars[VarPipeline])
	assert.Equal(t, "[{name: a},{name: b},{name: c},]", job.vars[VarImagePullSecrets])
}

func TestVariableLoader_Load_fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{
		GitServerPlugin:     config.Plugin{Name: "gits"},
		DomainManagerPlugin: config.Plugin{Name: "domain"},
	}).AnyTimes()

	gitS := mock.NewMockGitServer(m)
	app.EXPECT().GetPluginByName("gits").Return(gitS).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gitS.EXPECT().Initialize(gomock.All()).AnyTimes()
	gitS.EXPECT().GetCommitPipeline(gomock.Any(), gomock.Any()).Return(nil, errors.New("xxx"))

	app.EXPECT().GetPluginByName("domain").Return(&domain_manager.DefaultDomainManager{}).AnyTimes()

	em := &emptyMsger{}
	commit := mock.NewMockCommitInterface(m)
	commit.EXPECT().GetShortID().Return("short_id").AnyTimes()
	job := &Jober{
		commit: commit,
		config: &mars.Config{
			ValuesYaml: `
VarImagePullSecrets: <.ImagePullSecrets>
image: <.Pipeline>-<.Branch>
`,
		},
		project: &models.Project{
			Name:      "app",
			GitBranch: "dev",
		},
		ns:               &models.Namespace{Name: "ns"},
		imagePullSecrets: []string{"a", "b", "c"},
		messager:         em,
		percenter:        &emptyPercenter{},
	}
	err := (&VariableLoader{}).Load(job)
	assert.Equal(t, "无法获取 Pipeline 信息", err.Error())
}

func TestWithDryRun(t *testing.T) {
	j := &Jober{dryRun: false}
	WithDryRun()(j)
	assert.True(t, j.dryRun)
}

func Test_defaultLoaders(t *testing.T) {
	assert.Equal(t, []Loader{
		&ChartFileLoader{
			chartLoader: &defaultChartLoader{},
			fileOpener:  &defaultFileOpener{},
		},
		&VariableLoader{},
		&DynamicLoader{},
		&ExtraValuesLoader{},
		&MergeValuesLoader{},
		&ReleaseInstallerLoader{},
	}, defaultLoaders())
}

func Test_mergeYamlString_MarshalYAML(t *testing.T) {
	s := mergeYamlString{`
user:
  name: duc
`, `
user:
  age: 18
`}
	marshalYAML, _ := s.MarshalYAML()
	m := map[string]any{}
	yaml.Unmarshal([]byte(marshalYAML.(string)), &m)
	assert.Equal(t, 18, m["user"].(map[any]any)["age"])
	assert.Equal(t, "duc", m["user"].(map[any]any)["name"])
	a, _ := mergeYamlString{}.MarshalYAML()
	assert.Equal(t, "", a)
}

func Test_sortableExtraItem(t *testing.T) {
	s := sortableExtraItem{
		{
			Path:  "a",
			Value: "",
		},
		{
			Path:  "c",
			Value: "",
		},
		{
			Path:  "b",
			Value: "",
		},
	}
	sort.Sort(s)
	assert.Equal(t, sortableExtraItem{
		{
			Path:  "a",
			Value: "",
		},
		{
			Path:  "b",
			Value: "",
		},
		{
			Path:  "c",
			Value: "",
		},
	}, s)
}

func Test_toUpdatesMap(t *testing.T) {
	p := &models.Project{}
	assert.Equal(t, map[string]any{
		"manifest":           p.Manifest,
		"config":             p.Config,
		"git_project_id":     p.GitProjectId,
		"git_commit":         p.GitCommit,
		"git_branch":         p.GitBranch,
		"docker_image":       p.DockerImage,
		"pod_selectors":      p.PodSelectors,
		"override_values":    p.OverrideValues,
		"atomic":             p.Atomic,
		"extra_values":       p.ExtraValues,
		"final_extra_values": p.FinalExtraValues,
		"env_values":         p.EnvValues,
		"deploy_status":      p.DeployStatus,
		"git_commit_title":   p.GitCommitTitle,
		"git_commit_web_url": p.GitCommitWebUrl,
		"git_commit_author":  p.GitCommitAuthor,
		"git_commit_date":    p.GitCommitDate,
		"config_type":        p.ConfigType,
	}, toUpdatesMap(p))
}

func Test_userConfig_PrettyYaml(t *testing.T) {
	res := userConfig{
		ExtraValues: []*types.ExtraValue{
			{
				Path:  "a",
				Value: "",
			},
			{
				Path:  "c",
				Value: "",
			},
			{
				Path:  "b",
				Value: "",
			},
		},
	}.PrettyYaml()
	assert.Equal(t, `config: ""
branch: ""
commit: ""
atomic: false
web_url: ""
title: ""
extra_values:
- path: a
  value: ""
- path: b
  value: ""
- path: c
  value: ""
final_extra_values: ""
env_values: {}
`, res)
}

func Test_vars_MustGetString(t *testing.T) {
	assert.Equal(t, "", vars{}.MustGetString("aa"))
	assert.Equal(t, "bb", vars{"aa": "bb"}.MustGetString("aa"))
}

func Test_defaultFileOpener_Open(t *testing.T) {
	_, err := (&defaultFileOpener{}).Open("not exist")
	assert.ErrorIs(t, err, os.ErrNotExist)
}
