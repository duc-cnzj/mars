package socket

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/types"
	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/duc-cnzj/mars/internal/utils"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestChartFileLoader_Load(t *testing.T) {
}

func TestDynamicLoader_Load(t *testing.T) {
}

func TestExtraValuesLoader_Load(t *testing.T) {
}

func TestExtraValuesLoader_deepSetItems(t *testing.T) {
}

func TestExtraValuesLoader_typeValue(t *testing.T) {
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
}

func TestJober_LoadConfigs(t *testing.T) {
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
}

func TestMergeValuesLoader_Load(t *testing.T) {
}

func TestNewJober(t *testing.T) {
	assert.Implements(t, (*contracts.Job)(nil), NewJober(&websocket_pb.CreateProjectInput{}, contracts.UserInfo{}, "", nil, nil, 10))
	jober := NewJober(&websocket_pb.CreateProjectInput{}, contracts.UserInfo{}, "", nil, nil, 0, WithDryRun())
	assert.True(t, jober.IsDryRun())
}

func TestReleaseInstallerLoader_Load(t *testing.T) {

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
