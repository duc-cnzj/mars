package socket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/mars"
	"github.com/duc-cnzj/mars-client/v4/types"
	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/plugins/domain_manager"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
)

func TestChartFileLoader_Load(t *testing.T) {
}

func TestDynamicLoader_Load(t *testing.T) {
	em := &emptyMsger{}
	job := &Jober{
		input: &websocket_pb.CreateProjectInput{
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
		input: &websocket_pb.CreateProjectInput{
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
}

func TestExtraValuesLoader_Load(t *testing.T) {
	em := &emptyMsger{}
	job := &Jober{
		input: &websocket_pb.CreateProjectInput{
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
		input: &websocket_pb.CreateProjectInput{
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
	j := &Jober{
		done: make(chan struct{}),
	}

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
			closed: false,
			ch:     ch,
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
			closed: false,
			ch:     ch,
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
			closed: false,
			ch:     ch,
		},
	}
	go func() {
		ch <- contracts.MessageItem{Msg: "aa", Type: contracts.MessageText}
		close(ch)
	}()
	msger.EXPECT().SendMsg("aa").Times(1)
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
			closed: false,
			ch:     ch,
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
			closed: false,
			ch:     ch,
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
			closed: false,
			ch:     ch,
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
}

func (e *emptyLoader) Load(jober *Jober) error {
	e.called = true
	return e.err
}

func TestJober_LoadConfigs(t *testing.T) {
	l := &emptyLoader{}
	assert.Nil(t, (&Jober{
		messager: &emptyMsger{},
		stopCtx:  context.TODO(),
		loaders:  []Loader{l},
	}).LoadConfigs())
	assert.True(t, l.called)

	l2 := &emptyLoader{}
	cancel, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	assert.Equal(t, "context canceled", (&Jober{
		messager: &emptyMsger{},
		stopCtx:  cancel,
		loaders:  []Loader{l2},
	}).LoadConfigs().Error())
	assert.False(t, l2.called)

	l3 := &emptyLoader{
		err: errors.New("xxx"),
	}
	assert.Equal(t, "xxx", (&Jober{
		messager: &emptyMsger{},
		stopCtx:  context.TODO(),
		loaders:  []Loader{l3},
	}).LoadConfigs().Error())
	assert.True(t, l3.called)
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

func TestJober_Run(t *testing.T) {
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
}

func TestMarsLoader_Load(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	db, closeFn := testutil.SetGormDB(ctrl, app)
	defer closeFn()
	db.AutoMigrate(&models.GitProject{})
	mc := mars.Config{
		ConfigFile:       "cf",
		ConfigFileValues: "vv",
		ConfigField:      "f",
		IsSimpleEnv:      true,
		ConfigFileType:   "php",
		LocalChartPath:   "./charts",
		Branches:         []string{"dev", "master"},
		ValuesYaml:       "xxx",
		Elements:         nil,
	}
	marshal, _ := json.Marshal(&mc)
	db.Create(&models.GitProject{
		GitProjectId:  99,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})
	job := &Jober{
		input:     &websocket_pb.CreateProjectInput{GitProjectId: 99, GitBranch: "dev"},
		messager:  &emptyMsger{},
		percenter: &emptyPercenter{},
	}
	assert.Nil(t, (&MarsLoader{}).Load(job))
	assert.Equal(t, mc.String(), job.config.String())
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
	app.EXPECT().Uploader().Return(up).AnyTimes()
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
		input:             &websocket_pb.CreateProjectInput{GitProjectId: 99, GitBranch: "dev"},
		messager:          &emptyMsger{},
		percenter:         &emptyPercenter{},
	}
	assert.Nil(t, (&MergeValuesLoader{}).Load(job))
	assert.Equal(t, "/app/config.yaml", job.valuesOptions.ValueFiles[0])
}

func TestNewJober(t *testing.T) {
	assert.Implements(t, (*contracts.Job)(nil), NewJober(&websocket_pb.CreateProjectInput{}, contracts.UserInfo{}, "", nil, nil, 10))
	jober := NewJober(&websocket_pb.CreateProjectInput{}, contracts.UserInfo{}, "", nil, nil, 0, WithDryRun())
	assert.True(t, jober.IsDryRun())
}

type emptyMsger struct {
	msgs   []string
	called int
	contracts.DeployMsger
}

func (e *emptyMsger) SendMsg(s string) {
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
		input: &websocket_pb.CreateProjectInput{
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
		closed: false,
		ch:     ch,
	}
	fn := func() <-chan contracts.MessageItem {
		return ch
	}
	assert.Equal(t, fn(), sc.Chan())
}

func TestSafeWriteMessageCh_Closed(t *testing.T) {
	ch := make(chan contracts.MessageItem, 10)
	sc := &SafeWriteMessageCh{
		closed: false,
		ch:     ch,
	}
	sc.Closed()
	_, ok := <-ch
	assert.False(t, ok)
}

func TestSafeWriteMessageCh_Send(t *testing.T) {
	ch := make(chan contracts.MessageItem, 10)
	sc := &SafeWriteMessageCh{
		closed: false,
		ch:     ch,
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
		&MarsLoader{},
		&ChartFileLoader{},
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
