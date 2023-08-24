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
	"github.com/duc-cnzj/mars/v4/internal/cachelock"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/event"
	"github.com/duc-cnzj/mars/v4/internal/event/events"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/duc-cnzj/mars/v4/internal/utils"
	"github.com/duc-cnzj/mars/v4/internal/utils/pipeline"
	"github.com/duc-cnzj/mars/v4/plugins/domainmanager"
	"helm.sh/helm/v3/pkg/release"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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

func TestChartFileLoader_Load2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := &emptyMsger{}
	h := mock.NewMockHelmer(m)
	job := &jobRunner{
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
	app := testutil.MockApp(m)
	gits := testutil.MockGitServer(m, app)

	gits.EXPECT().GetDirectoryFilesWithSha("100", "commit", "dir", true).Return([]string{"file1", "file2"}, nil)

	up := mock.NewMockUploader(m)
	app.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().MkDir(gomock.Any(), false).Return(errors.New("mkdir err")).Times(1)

	err := l.Load(job)
	assert.Equal(t, "mkdir err", err.Error())
}

func TestDynamicLoader_Load(t *testing.T) {
	em := &emptyMsger{}
	job := &jobRunner{
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
	job2 := &jobRunner{
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
	job3 := &jobRunner{
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
	job := &jobRunner{
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

	err := (&ExtraValuesLoader{}).Load(&jobRunner{
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

	j := &jobRunner{
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

	j2 := &jobRunner{
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

func TestExtraValuesLoader_typedValue(t *testing.T) {
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
		{
			ele: &mars.Element{
				Path: "app->config",
				Type: mars.ElementType_ElementTypeNumberSelect,
				SelectValues: []string{
					"1", "2", "3",
				},
			},
			input:  "1",
			result: 1,
			err:    "",
		},
		{
			ele: &mars.Element{
				Path: "app->config",
				Type: mars.ElementType_ElementTypeNumberRadio,
				SelectValues: []string{
					"1", "2", "3",
				},
			},
			input:  "2",
			result: 2,
			err:    "",
		},
		{
			// 如果输入本身不是 num 则原样返回
			ele: &mars.Element{
				Path: "app->config",
				Type: mars.ElementType_ElementTypeNumberRadio,
				SelectValues: []string{
					"1x", "2x", "3x",
				},
			},
			input:  "2x",
			result: "2x",
			err:    "",
		},
	}

	for i, test := range tests {
		tt := test
		t.Run(fmt.Sprintf("test-%v", i), func(t *testing.T) {
			t.Parallel()
			value, err := (&ExtraValuesLoader{}).typedValue(tt.ele, tt.input)
			if err != nil {
				assert.Equal(t, err.Error(), tt.err)
			} else {
				assert.Equal(t, tt.result, value)
			}
		})
	}
}

func TestJober_Commit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := mock.NewMockCommitInterface(m)
	j := &jobRunner{
		commit: c,
	}
	assert.Same(t, c, j.Commit())
}

func TestJober_GetStoppedErrorIfHas(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msg := mock.NewMockDeployMsger(m)
	msg.EXPECT().SendMsg(gomock.Any())
	context, f := utils.NewCustomErrorContext()
	j := &jobRunner{
		slugName: "aaa",
		messager: msg,
		stopCtx:  context,
		stopFn:   f,
	}
	assert.Nil(t, j.GetStoppedErrorIfHas())
	j.Stop(errors.New("xxx"))
	assert.Error(t, j.GetStoppedErrorIfHas())
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
	j := &jobRunner{
		messager: msger,
		messageCh: &safeWriteMessageCh{
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
	j := &jobRunner{
		messager: msger,
		messageCh: &safeWriteMessageCh{
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
	j := &jobRunner{
		project:  &models.Project{},
		stopCtx:  context.TODO(),
		dryRun:   true,
		messager: msger,
		messageCh: &safeWriteMessageCh{
			ch: ch,
		},
	}
	go func() {
		ch <- contracts.MessageItem{Msg: "errors", Type: contracts.MessageError}
		close(ch)
	}()

	j.HandleMessage()
	assert.True(t, j.deployResult.IsSet())
	assert.Equal(t, "errors", j.deployResult.Msg())
	assert.Equal(t, ResultDeployFailed, j.deployResult.ResultType())
	assert.NotNil(t, j.deployResult.Model())
}

func TestJober_HandleMessage_SuccessMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Done().Return(nil).AnyTimes()
	ch := make(chan contracts.MessageItem, 10)
	msger := mock.NewMockDeployMsger(m)
	j := &jobRunner{
		project:  &models.Project{},
		stopCtx:  context.TODO(),
		dryRun:   true,
		messager: msger,
		messageCh: &safeWriteMessageCh{
			ch: ch,
		},
	}
	go func() {
		ch <- contracts.MessageItem{Msg: "ok", Type: contracts.MessageSuccess}
		close(ch)
	}()
	j.HandleMessage()
	assert.True(t, j.deployResult.IsSet())
	assert.Equal(t, "ok", j.deployResult.Msg())
	assert.Equal(t, ResultDeployed, j.deployResult.ResultType())
	assert.NotNil(t, j.deployResult.Model())
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
	j := &jobRunner{
		project:  p,
		stopCtx:  ctx,
		dryRun:   false,
		isNew:    true,
		messager: msger,
		messageCh: &safeWriteMessageCh{
			ch: ch,
		},
	}
	go func() {
		ch <- contracts.MessageItem{Msg: "errors", Type: contracts.MessageError}
		close(ch)
	}()
	j.HandleMessage()
	assert.True(t, j.deployResult.IsSet())
	assert.NotEmpty(t, j.deployResult.Msg())
	assert.Equal(t, ResultDeployCanceled, j.deployResult.ResultType())
	assert.NotNil(t, j.deployResult.Model())
}

func TestJober_ID(t *testing.T) {
	j := &jobRunner{
		slugName: "aaa",
	}
	assert.Equal(t, "aaa", j.ID())
}

func TestJober_IsDryRun(t *testing.T) {
	j := &jobRunner{
		dryRun: true,
	}
	assert.True(t, j.IsDryRun())
}

func TestJober_IsNew(t *testing.T) {
	j := &jobRunner{
		isNew: true,
	}
	assert.True(t, j.IsNew())
}

func TestJober_IsStopped(t *testing.T) {
	j := &jobRunner{
		stopCtx: context.TODO(),
	}
	assert.False(t, j.IsStopped())
	cancel, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	j2 := &jobRunner{
		stopCtx: cancel,
	}
	assert.True(t, j2.IsStopped())
}

type emptyLoader struct {
	err    error
	called bool
	sync.Mutex
}

func (e *emptyLoader) Load(jober *jobRunner) error {
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
	assert.Equal(t, "context canceled", (&jobRunner{
		messager: &emptyMsger{},
		stopCtx:  ctx,
		loaders:  []Loader{l},
	}).LoadConfigs().Error().Error())
	assert.False(t, l.GetCalled())
}

func TestJober_LoadConfigs(t *testing.T) {
	assert.Equal(t, "xxx", (&jobRunner{
		err:     errors.New("xxx"),
		loaders: []Loader{},
	}).LoadConfigs().Error().Error())

	l := &emptyLoader{}
	assert.Nil(t, (&jobRunner{
		messager: &emptyMsger{},
		stopCtx:  context.TODO(),
		loaders:  []Loader{l},
	}).LoadConfigs().Error())
	assert.True(t, l.GetCalled())

	l2 := &emptyLoader{}
	cancel, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	assert.Equal(t, "context canceled", (&jobRunner{
		messager: &emptyMsger{},
		stopCtx:  cancel,
		loaders:  []Loader{l2},
	}).LoadConfigs().Error().Error())
	assert.False(t, l2.GetCalled())

	l3 := &emptyLoader{
		err: errors.New("xxx"),
	}
	assert.Equal(t, "xxx", (&jobRunner{
		messager: &emptyMsger{},
		stopCtx:  context.TODO(),
		loaders:  []Loader{l3},
	}).LoadConfigs().Error().Error())
	assert.True(t, l3.GetCalled())
}

func TestJober_Manifests(t *testing.T) {
	j := &jobRunner{
		manifests: []string{"xxx"},
	}
	assert.Equal(t, []string{"xxx"}, j.Manifests())
}

func TestJober_Messager(t *testing.T) {
	m := NewMessageSender(nil, "", 0)
	j := &jobRunner{
		messager: m,
	}
	assert.Same(t, m, j.Messager())
}

func TestJober_Namespace(t *testing.T) {
	ns := &models.Namespace{Name: "aaa"}
	j := &jobRunner{
		ns: ns,
	}
	assert.Same(t, ns, j.Namespace())
}

func TestJober_Percenter(t *testing.T) {
	p := newProcessPercent(nil, nil)
	j := &jobRunner{
		percenter: p,
	}
	assert.Same(t, p, j.Percenter())
}

func TestJober_Project(t *testing.T) {
	p := &models.Project{}
	j := &jobRunner{
		project: p,
	}
	assert.Same(t, p, j.Project())
}

func TestJober_ProjectModel(t *testing.T) {
	j := &jobRunner{}
	assert.Nil(t, j.ProjectModel())
	j.project = &models.Project{Name: "aa"}
	assert.NotNil(t, j.ProjectModel())
}

func TestJober_PubSub(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	ps := mock.NewMockPubSub(m)
	j := &jobRunner{
		pubsub: ps,
	}
	assert.Same(t, ps, j.PubSub())
}

func TestJober_ReleaseInstaller(t *testing.T) {
	ri := &releaseInstaller{}
	j := &jobRunner{
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

	msgCh := &safeWriteMessageCh{
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
	job := &jobRunner{
		project:   proj,
		isNew:     true,
		dryRun:    false,
		installer: rinstaller,
		stopCtx:   stopCtx,
		messageCh: msgCh,
		commit:    commit,
	}
	assert.Equal(t, "xxx", job.Run().Error().Error())
	assert.Equal(t, "xxx", job.deployResult.Msg())
}

func TestJober_Run_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	app := testutil.MockApp(m)
	app.EXPECT().Done().Return(nil).AnyTimes()

	db, closeFn := testutil.SetGormDB(m, app)
	defer closeFn()
	assert.Nil(t, db.AutoMigrate(&models.Project{}, &models.Namespace{}, &models.Event{}, &models.Changelog{}, &models.GitProject{}))

	msgCh := &safeWriteMessageCh{
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

	gits := testutil.MockGitServer(m, app)

	commit2 := mock.NewMockCommitInterface(m)
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(commit2, nil)
	commit2.EXPECT().GetWebURL().Return("weburl2").Times(1)
	commit2.EXPECT().GetTitle().Return("title2").Times(1)

	d := event.NewDispatcher(app)
	d.Listen(events.EventProjectChanged, events.HandleProjectChanged)
	d.Listen(events.EventAuditLog, events.HandleAuditLog)
	app.EXPECT().EventDispatcher().Return(d).AnyTimes()

	job := &jobRunner{
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
	assert.Nil(t, job.Run().Error())
	assert.Equal(t, "部署成功", job.deployResult.Msg())

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
	msg.EXPECT().SendMsg(gomock.Any()).Times(3)
	var called int64 = 0
	j := &jobRunner{messager: msg, stopFn: func(err error) {
		atomic.AddInt64(&called, 1)
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
	assert.Equal(t, int64(3), atomic.LoadInt64(&called))
}

func TestJober_User(t *testing.T) {
	j := &jobRunner{}
	assert.IsType(t, contracts.UserInfo{}, j.User())
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
	assert.Nil(t, db.Create(gp).Error)
	p := models.Project{
		NamespaceId:  ns.ID,
		GitProjectId: 100,
		Name:         "app",
		Version:      1,
		DeployStatus: uint8(types.Deploy_StatusDeployed),
	}
	assert.Nil(t, db.Create(&p).Error)
	gits := testutil.MockGitServer(m, app)
	gits.EXPECT().GetFileContentWithBranch(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	h := mock.NewMockHelmer(m)
	ps := mock.NewMockPubSub(m)
	job := &jobRunner{
		pubsub: ps,
		helmer: h,
		input: &JobInput{
			NamespaceId:  int64(ns.ID),
			GitProjectId: 100,
			Name:         p.Name,
			Version:      0,
			Type:         websocket_pb.Type_UpdateProject,
		},
		messager:  &emptyMsger{},
		percenter: &emptyPercenter{},
	}
	assert.ErrorIs(t, ErrorVersionNotMatched, job.Validate().Error())
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	ps.EXPECT().ToAll(reloadProjectsMessage(ns.ID)).AnyTimes()
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
		assert.Equal(t, 1, result["app"].(map[string]any)["one"])
		assert.Equal(t, "two", result["app"].(map[string]any)["two"])
		assert.Equal(t, 3, result["app"].(map[string]any)["three"])
		assert.Equal(t, 4, result["app"].(map[string]any)["four"])
		assert.Equal(t, []any{map[string]any{"name": "secret"}}, result["imagePullSecrets"])
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
	job := &jobRunner{
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

	job2 := &jobRunner{
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
	m := gomock.NewController(t)
	defer m.Finish()
	msg := mock.NewMockDeployMsger(m)
	ps := mock.NewMockPubSub(m)
	l := mock.NewMockLocker(m)
	testutil.MockApp(m).EXPECT().CacheLock().Return(l).Times(2)
	input := &JobInput{}
	user := contracts.UserInfo{Name: "duc"}
	job := NewJober(input, user, "x", msg, ps, 10, WithDryRun()).(*jobRunner)
	assert.False(t, job.IsNotDryRun())

	assert.Equal(t, &DefaultHelmer{}, job.helmer)
	assert.Equal(t, defaultLoaders(), job.loaders)
	assert.Equal(t, user, job.user)
	assert.Same(t, ps, job.pubsub)
	assert.Same(t, msg, job.messager)
	assert.Equal(t, vars{}, job.vars)
	assert.Equal(t, &values.Options{}, job.valuesOptions)
	assert.Equal(t, "x", job.slugName)
	assert.Equal(t, input, job.input)
	assert.Equal(t, int64(10), job.timeoutSeconds)
	assert.Same(t, l, job.locker)
	assert.NotNil(t, job.messageCh)
	assert.Equal(t, 100, cap(job.messageCh.(*safeWriteMessageCh).ch))
	assert.Equal(t, newProcessPercent(msg, &realSleeper{}), job.percenter)

	assert.Implements(t, (*contracts.Job)(nil), NewJober(&JobInput{}, contracts.UserInfo{}, "", nil, nil, 10))

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
	job := &jobRunner{
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
	sc := &safeWriteMessageCh{
		ch: ch,
	}
	fn := func() <-chan contracts.MessageItem {
		return ch
	}
	assert.Equal(t, fn(), sc.Chan())
}

func TestSafeWriteMessageCh_Closed(t *testing.T) {
	ch := make(chan contracts.MessageItem, 10)
	sc := &safeWriteMessageCh{
		ch: ch,
	}
	sc.Close()
	_, ok := <-ch
	assert.False(t, ok)
}

func TestSafeWriteMessageCh_Send(t *testing.T) {
	ch := make(chan contracts.MessageItem, 10)
	sc := &safeWriteMessageCh{
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
	sc.Close()
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

	app.EXPECT().GetPluginByName("domain").Return(domainmanager.NewDefaultDomainManager()).AnyTimes()

	em := &emptyMsger{}
	commit := mock.NewMockCommitInterface(m)
	commit.EXPECT().GetShortID().Return("short_id").AnyTimes()
	job := &jobRunner{
		commit: commit,
		config: &mars.Config{
			ValuesYaml: `
if-master-test: < if ne .Branch "master" >not-master/is- <- .Branch >< else >< .Branch >< end >
if-dev-test: < if eq .Branch "dev" >dev-< .Branch >   <- end >
VarImagePullSecrets: <.ImagePullSecrets>
image: <.Pipeline>-<.Branch>
VarImagePullSecretsNoName: <.ImagePullSecretsNoName>
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
if-master-test: not-master/is-dev
if-dev-test: dev-dev
VarImagePullSecrets: [{name: a}, {name: b}, {name: c}, ]
image: 9999-dev
VarImagePullSecretsNoName: [a, b, c, ]
`,
		job.valuesYaml)
	assert.Equal(t, "dev", job.vars[VarBranch])
	assert.Equal(t, "short_id", job.vars[VarCommit])
	assert.Equal(t, "9999", job.vars[VarPipeline])
	assert.Equal(t, "[{name: a}, {name: b}, {name: c}, ]", job.vars[VarImagePullSecrets])
	assert.Equal(t, "[a, b, c, ]", job.vars[VarImagePullSecretsNoName])

	em2 := &emptyMsger{}
	job2 := &jobRunner{
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

	app.EXPECT().GetPluginByName("domain").Return(domainmanager.NewDefaultDomainManager()).AnyTimes()

	em := &emptyMsger{}
	commit := mock.NewMockCommitInterface(m)
	commit.EXPECT().GetShortID().Return("short_id").AnyTimes()
	job := &jobRunner{
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
VarImagePullSecrets: [{name: a}, {name: b}, {name: c}, ]
`,
		job.valuesYaml)
	assert.Equal(t, "dev", job.vars[VarBranch])
	assert.Equal(t, "short_id", job.vars[VarCommit])
	assert.Equal(t, "0", job.vars[VarPipeline])
	assert.Equal(t, "[{name: a}, {name: b}, {name: c}, ]", job.vars[VarImagePullSecrets])
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

	app.EXPECT().GetPluginByName("domain").Return(domainmanager.NewDefaultDomainManager()).AnyTimes()

	em := &emptyMsger{}
	commit := mock.NewMockCommitInterface(m)
	commit.EXPECT().GetShortID().Return("short_id").AnyTimes()
	job := &jobRunner{
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
	j := &jobRunner{dryRun: false}
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
	assert.Equal(t, 18, m["user"].(map[string]any)["age"])
	assert.Equal(t, "duc", m["user"].(map[string]any)["name"])
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
		"git_commit_title":   p.GitCommitTitle,
		"git_commit_web_url": p.GitCommitWebUrl,
		"git_commit_author":  p.GitCommitAuthor,
		"git_commit_date":    p.GitCommitDate,
		"config_type":        p.ConfigType,
	}, toUpdatesMap(p))
}

func Test_userConfig_PrettyYaml(t *testing.T) {
	res := userConfig{
		Config: `name: duc
age: 18  
realAge: 28
`,
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
	assert.Equal(t, `config: |
  name: duc
  age: 18  
  realAge: 28
branch: ""
commit: ""
atomic: false
web_url: ""
title: ""
extra_values:
  - path: a
  - path: b
  - path: c
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

func TestJober_GlobalLock(t *testing.T) {
	l := cachelock.NewMemoryLock([2]int{2, 100}, cachelock.NewMemStore())
	job := &jobRunner{locker: l, slugName: "id"}
	assert.Nil(t, job.GlobalLock().Error())
	assert.Equal(t, "正在部署中，请稍后再试", (&jobRunner{locker: l, slugName: "id"}).GlobalLock().Error().Error())
	assert.Len(t, job.finallyCallback.Sort(), 1)
	called := 0
	pipeline.NewPipeline[error]().Send(nil).Through(job.finallyCallback.Sort()...).Then(func(e error) {
		called++
		assert.Nil(t, e)
	})
	assert.Equal(t, 1, called)
	acquire := l.Acquire("id", 100)
	assert.True(t, acquire)

	m := gomock.NewController(t)
	defer m.Finish()
	ml := mock.NewMockLocker(m)
	ml.EXPECT().RenewalAcquire(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	assert.Equal(t, "xxx", (&jobRunner{err: errors.New("xxx"), locker: ml}).GlobalLock().Error().Error())
}

func TestJober_Validate_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := mock.NewMockDeployMsger(m)
	msger.EXPECT().SendMsg(gomock.Any()).Times(0)
	assert.Equal(t, "xxx", (&jobRunner{err: errors.New("xxx"), messager: msger}).Validate().Error().Error())

	assert.Equal(t, errors.New("type error: "+websocket_pb.Type_TypeUnknown.String()), (&jobRunner{messager: msger, input: &JobInput{Type: websocket_pb.Type_TypeUnknown}}).Validate().Error())
}

func TestJober_Validate_NamespaceNotExists(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.Namespace{})
	err := (&jobRunner{input: &JobInput{Type: websocket_pb.Type_ApplyProject, NamespaceId: 100}, messager: &emptyMsger{}, percenter: &emptyPercenter{}}).Validate().Error()
	assert.Equal(t, "[FAILED]: 校验名称空间: record not found", err.Error())
}

func TestJober_Validate_GetProjectMarsConfigError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.Namespace{})
	ns := &models.Namespace{Name: "ns"}
	db.Create(ns)
	gits := testutil.MockGitServer(m, app)
	gits.EXPECT().GetFileContentWithBranch("100", "dev", ".mars.yaml").Return("", errors.New("xxx"))
	err := (&jobRunner{input: &JobInput{Type: websocket_pb.Type_ApplyProject, NamespaceId: int64(ns.ID), GitProjectId: 100, GitBranch: "dev"}, messager: &emptyMsger{}, percenter: &emptyPercenter{}}).Validate().Error()
	assert.Equal(t, "xxx", err.Error())
}

func TestJober_Validate_Create(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	ns := &models.Namespace{Name: "ns"}
	db.Create(ns)
	gits := testutil.MockGitServer(m, app)
	gits.EXPECT().GetFileContentWithBranch("100", "dev", ".mars.yaml").Return("{}", nil)
	gitp := mock.NewMockProjectInterface(m)
	gits.EXPECT().GetProject("100").Return(gitp, nil)
	gitp.EXPECT().GetName().Return("git-app").Times(1)
	pubsub := mock.NewMockPubSub(m)
	pubsub.EXPECT().ToAll(reloadProjectsMessage(ns.ID)).Times(1)
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(nil, errors.New("commit err"))
	job := &jobRunner{pubsub: pubsub, input: &JobInput{Type: websocket_pb.Type_ApplyProject, NamespaceId: int64(ns.ID), GitProjectId: 100, GitBranch: "dev"}, messager: &emptyMsger{}, percenter: &emptyPercenter{}}
	err := job.
		Validate().
		Error()
	assert.Equal(t, "git-app", job.input.Name)
	assert.Equal(t, "commit err", err.Error())

	assert.Len(t, job.errorCallback.Sort(), 1)
	assert.Len(t, job.finallyCallback.Sort(), 1)

	called := 0
	hm := mock.NewMockHelmer(m)
	job.helmer = hm
	hm.EXPECT().ReleaseStatus("git-app", "ns").Return(types.Deploy_StatusFailed)
	pubsub.EXPECT().ToAll(reloadProjectsMessage(ns.ID)).Times(1)
	pipeline.NewPipeline[error]().Send(nil).Through(job.finallyCallback.Sort()...).Then(func(err2 error) {
		called++
	})
	assert.Equal(t, 1, called)
	p1 := &models.Project{ID: job.project.ID}
	db.First(&p1)
	assert.Equal(t, p1.DeployStatus, uint8(types.Deploy_StatusFailed))

	pipeline.NewPipeline[error]().Send(nil).Through(job.errorCallback.Sort()...).Then(func(err2 error) {
		called++
	})
	assert.Equal(t, 2, called)
	p := &models.Project{ID: job.project.ID}
	db.Unscoped().First(&p)
	assert.True(t, p.DeletedAt.Valid)
}

func TestJober_Validate_Update(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	ns := &models.Namespace{Name: "ns"}
	db.Create(ns)
	gits := testutil.MockGitServer(m, app)
	gits.EXPECT().GetFileContentWithBranch("100", "dev", ".mars.yaml").Return("{}", nil)
	pubsub := mock.NewMockPubSub(m)
	pubsub.EXPECT().ToAll(reloadProjectsMessage(ns.ID)).Times(1)
	commit := mock.NewMockCommitInterface(m)
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(commit, nil)
	p := &models.Project{Name: "app", NamespaceId: ns.ID}
	db.Create(p)
	assert.Equal(t, 1, p.Version)
	job := &jobRunner{pubsub: pubsub, input: &JobInput{Version: int64(p.Version), Name: "app", Type: websocket_pb.Type_ApplyProject, NamespaceId: int64(ns.ID), GitProjectId: 100, GitBranch: "dev"}, messager: &emptyMsger{}, percenter: &emptyPercenter{}}
	err := job.
		Validate().
		Error()
	job.prevProject.Version = 1
	var newp = models.Project{ID: p.ID}
	db.First(&newp)

	assert.Equal(t, p.ProtoTransform(), job.prevProject.ProtoTransform())
	assert.Equal(t, uint8(0), job.prevProject.DeployStatus)
	assert.Equal(t, int(2), newp.Version)
	assert.Equal(t, "app", job.input.Name)
	assert.Nil(t, err)
	assert.Same(t, commit, job.commit)

	assert.Len(t, job.errorCallback.Sort(), 1)
	called := false
	pipeline.NewPipeline[error]().Send(nil).Through(job.errorCallback.Sort()...).Then(func(err error) {
		called = true
	})
	assert.True(t, called)
	pp := &models.Project{ID: job.project.ID}
	db.First(&pp)
	assert.Equal(t, 1, pp.Version)
}

func TestJober_Validate_ErrVersionNotMatched(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	ns := &models.Namespace{Name: "ns"}
	db.Create(ns)
	gits := testutil.MockGitServer(m, app)
	gits.EXPECT().GetFileContentWithBranch("100", "dev", ".mars.yaml").Return("{}", nil)
	p := &models.Project{Name: "app", NamespaceId: ns.ID}
	db.Create(p)
	assert.Equal(t, 1, p.Version)
	job := &jobRunner{input: &JobInput{Version: int64(p.Version) + 1, Name: "app", Type: websocket_pb.Type_ApplyProject, NamespaceId: int64(ns.ID), GitProjectId: 100, GitBranch: "dev"}, messager: &emptyMsger{}, percenter: &emptyPercenter{}}
	err := job.
		Validate().
		Error()
	assert.Equal(t, ErrorVersionNotMatched, err)
}

func TestJober_WsTypeValidated(t *testing.T) {
	var tests = []struct {
		t      websocket_pb.Type
		result bool
	}{
		{t: websocket_pb.Type_TypeUnknown, result: false},
		{t: websocket_pb.Type_SetUid, result: false},
		{t: websocket_pb.Type_ReloadProjects, result: false},
		{t: websocket_pb.Type_CancelProject, result: false},
		{t: websocket_pb.Type_CreateProject, result: true},
		{t: websocket_pb.Type_UpdateProject, result: true},
		{t: websocket_pb.Type_ProcessPercent, result: false},
		{t: websocket_pb.Type_ClusterInfoSync, result: false},
		{t: websocket_pb.Type_InternalError, result: false},
		{t: websocket_pb.Type_ApplyProject, result: true},
		{t: websocket_pb.Type_ProjectPodEvent, result: false},
		{t: websocket_pb.Type_HandleExecShell, result: false},
		{t: websocket_pb.Type_HandleExecShellMsg, result: false},
		{t: websocket_pb.Type_HandleCloseShell, result: false},
		{t: websocket_pb.Type_HandleAuthorize, result: false},
	}
	for _, test := range tests {
		tt := test
		t.Run(fmt.Sprintf("WsTypeValidated: %v %t", tt.t.String(), tt.result), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.result, (&jobRunner{input: &JobInput{Type: tt.t}}).WsTypeValidated())
		})
	}
}

func TestJober_Run(t *testing.T) {
	assert.Equal(t, "xxx", (&jobRunner{err: errors.New("xxx")}).Run().Error().Error())
}

func TestJober_Finish_WhenError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := mock.NewMockDeployMsger(m)
	stopCtx, stopFn := utils.NewCustomErrorContext()
	stopFn(errors.New("stopped"))
	job := &jobRunner{err: errors.New("xxx"), messager: msger, stopCtx: stopCtx, stopFn: stopFn}
	successCalled := 0
	job.OnSuccess(1, func(err error, sendResultToUser func()) {
		sendResultToUser()
		successCalled++
	})
	errorCalled := 0
	job.OnError(1, func(err error, sendResultToUser func()) {
		sendResultToUser()
		errorCalled++
	})
	finallyCalled := 0
	job.OnFinally(1, func(err error, sendResultToUser func()) {
		sendResultToUser()
		finallyCalled++
	})
	msger.EXPECT().SendDeployedResult(websocket_pb.ResultType_DeployedCanceled, "stopped", nil).Times(1)
	// canceled
	assert.Equal(t, "xxx", job.Finish().Error().Error())
	assert.Equal(t, 1, finallyCalled)
	assert.Equal(t, 0, successCalled)
	assert.Equal(t, 1, errorCalled)

	// failed
	job2 := &jobRunner{err: errors.New("xxx"), messager: msger, stopCtx: context.TODO()}
	msger.EXPECT().SendDeployedResult(websocket_pb.ResultType_DeployedFailed, "xxx", nil).Times(1)
	job2.Finish()
}
func TestJober_Finish_WhenSuccess(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := mock.NewMockDeployMsger(m)
	job := &jobRunner{messager: msger}
	successCalled := 0
	job.OnSuccess(1, func(err error, sendResultToUser func()) {
		sendResultToUser()
		successCalled++
	})
	errorCalled := 0
	job.OnError(1, func(err error, sendResultToUser func()) {
		sendResultToUser()
		errorCalled++
	})
	finallyCalled := 0
	job.OnFinally(1, func(err error, sendResultToUser func()) {
		sendResultToUser()
		finallyCalled++
	})
	msger.EXPECT().SendDeployedResult(websocket_pb.ResultType_Deployed, "ok", nil).Times(1)
	job.SetDeployResult(websocket_pb.ResultType_Deployed, "ok", nil)
	// success
	assert.Nil(t, job.Finish().Error())
	assert.Equal(t, 1, finallyCalled)
	assert.Equal(t, 1, successCalled)
	assert.Equal(t, 0, errorCalled)
}

func TestJober_OnError(t *testing.T) {
	job := &jobRunner{err: errors.New("xxx")}
	job.OnError(1, func(err error, sendResultToUser func()) {
		assert.Equal(t, "xxx", err.Error())
		sendResultToUser()
	})
	assert.Len(t, job.errorCallback.Sort(), 1)
	called := 0
	pipeline.NewPipeline[error]().Send(job.Error()).Through(job.errorCallback.Sort()...).Then(func(err error) {
		assert.Equal(t, "xxx", err.Error())
		called++
	})
	assert.Equal(t, 1, called)
}

func TestJober_OnSuccess(t *testing.T) {
	job := &jobRunner{}
	job.OnSuccess(1, func(err error, sendResultToUser func()) {
		assert.Nil(t, err)
		sendResultToUser()
	})
	assert.Len(t, job.successCallback.Sort(), 1)
	called := 0
	pipeline.NewPipeline[error]().Send(job.Error()).Through(job.finallyCallback.Sort()...).Then(func(err error) {
		assert.Nil(t, err)
		called++
	})
	assert.Equal(t, 1, called)
}

func TestJober_OnFinally(t *testing.T) {
	var tests = []error{
		errors.New("xxx"),
		nil,
	}
	for _, test := range tests {
		tt := test
		t.Run("", func(t *testing.T) {
			t.Parallel()
			job := &jobRunner{err: tt}
			job.OnFinally(1, func(err error, sendResultToUser func()) {
				assert.Equal(t, tt, err)
				sendResultToUser()
			})
			assert.Len(t, job.finallyCallback.Sort(), 1)
			called := 0
			pipeline.NewPipeline[error]().Send(job.Error()).Through(job.finallyCallback.Sort()...).Then(func(err error) {
				assert.Equal(t, tt, err)
				called++
			})
			assert.Equal(t, 1, called)
		})
	}
}
func TestChartFileLoader_Load(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := &emptyMsger{}
	h := mock.NewMockHelmer(m)
	job := &jobRunner{
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
	app := testutil.MockApp(m)
	gits := testutil.MockGitServer(m, app)

	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return(nil, errors.New("xxx"))

	err := l.Load(job)
	assert.Equal(t, "charts 文件不存在", err.Error())

	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return([]string{"file1", "file2"}, nil)

	gits.EXPECT().GetFileContentWithSha("9999", "master", "file1").Return("file1", nil).Times(1)
	gits.EXPECT().GetFileContentWithSha("9999", "master", "file2").Return("file2", nil).Times(1)
	up := mock.NewMockUploader(m)
	app.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().AbsolutePath(gomock.Any()).Return("/dir")
	up.EXPECT().MkDir(gomock.Any(), false).Times(1)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Times(2)
	h.EXPECT().PackageChart(gomock.Any(), gomock.Any()).Times(1)

	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir/xxxx", true).Return([]string{}, nil)

	err = l.Load(job)
	assert.Len(t, job.finallyCallback.Sort(), 2)
	assert.Nil(t, err)

	up.EXPECT().DeleteDir("/dir").Times(2)
	called := 0

	pipeline.NewPipeline[error]().Send(job.Error()).Through(job.finallyCallback.Sort()...).Then(func(err error) {
		called++
	})
	assert.Equal(t, 1, called)
}
