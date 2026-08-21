package deploy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/pipeline"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

func TestNewJobManager(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	manager := NewJobManager(JobManagerDeps{
		Timer:            timer.NewReal(),
		Logger:           mlog.NewForConfig(nil),
		ReleaseInstaller: NewMockReleaseInstaller(m),
		RepoRepo:         data.NewMockRepoRepo(m),
		NsRepo:           data.NewMockNamespaceRepo(m),
		ProjRepo:         data.NewMockProjectRepo(m),
		Helmer:           data.NewMockHelmerRepo(m),
		Uploader:         uploader.NewMockUploader(m),
		Locker:           locker.NewMockLocker(m),
		K8sRepo:          data.NewMockK8sRepo(m),
		EventRepo:        data.NewMockEventRepo(m),
		PluginManager:    app.NewMockPluginManager(m),
	}).(*jobManager)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.timer)
	assert.NotNil(t, manager.logger)
	assert.NotNil(t, manager.releaseInstaller)
	assert.NotNil(t, manager.repoRepo)
	assert.NotNil(t, manager.nsRepo)
	assert.NotNil(t, manager.projRepo)
	assert.NotNil(t, manager.helmRepo)
	assert.NotNil(t, manager.uploader)
	assert.NotNil(t, manager.locker)
	assert.NotNil(t, manager.k8sRepo)
	assert.NotNil(t, manager.eventRepo)
	assert.NotNil(t, manager.pluginManager)
}

func TestNewJob(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	timer := timer.NewReal()
	logger := mlog.NewForConfig(nil)
	releaseInstaller := NewMockReleaseInstaller(m)
	repoRepo := data.NewMockRepoRepo(m)
	nsRepo := data.NewMockNamespaceRepo(m)
	projRepo := data.NewMockProjectRepo(m)
	helmer := data.NewMockHelmerRepo(m)
	uploader := uploader.NewMockUploader(m)
	locker := locker.NewMockLocker(m)
	k8sRepo := data.NewMockK8sRepo(m)
	eventRepo := data.NewMockEventRepo(m)
	pl := app.NewMockPluginManager(m)

	manager := NewJobManager(JobManagerDeps{
		Timer:            timer,
		Logger:           logger,
		ReleaseInstaller: releaseInstaller,
		RepoRepo:         repoRepo,
		NsRepo:           nsRepo,
		ProjRepo:         projRepo,
		Helmer:           helmer,
		Uploader:         uploader,
		Locker:           locker,
		K8sRepo:          k8sRepo,
		EventRepo:        eventRepo,
		PluginManager:    pl,
	})

	jobInput := &JobInput{
		Type:           websocket_pb.Type_CreateProject,
		NamespaceId:    1,
		Name:           "test",
		RepoID:         1,
		GitBranch:      "main",
		GitCommit:      "abc123",
		Config:         "test-config",
		Atomic:         new(bool),
		ExtraValues:    []*websocket_pb.ExtraValue{},
		Version:        new(int32),
		ProjectID:      1,
		TimeoutSeconds: 10,
		User:           &biz.UserInfo{},
		DryRun:         true,
		PubSub:         NewEmptyPubSub(),
		Messager:       nil,
	}

	job := manager.NewJob(jobInput)

	assert.NotNil(t, job)
	assert.Equal(t, jobInput, job.(*jobRunner).input)
	assert.NotNil(t, job.(*jobRunner).timer)
	assert.NotNil(t, job.(*jobRunner).logger)

	assert.NotNil(t, job.(*jobRunner).installer)
	assert.NotNil(t, job.(*jobRunner).repoRepo)
	assert.NotNil(t, job.(*jobRunner).nsRepo)
	assert.NotNil(t, job.(*jobRunner).projRepo)
	assert.NotNil(t, job.(*jobRunner).helmer)
	assert.NotNil(t, job.(*jobRunner).uploader)
	assert.NotNil(t, job.(*jobRunner).locker)
	assert.NotNil(t, job.(*jobRunner).k8sRepo)
	assert.NotNil(t, job.(*jobRunner).eventRepo)
	assert.NotNil(t, job.(*jobRunner).pluginMgr)
	assert.NotNil(t, job.(*jobRunner).PubSub())
	assert.Nil(t, job.(*jobRunner).Error())
	assert.NotNil(t, job.(*jobRunner).loaders)
	assert.True(t, job.(*jobRunner).dryRun)
	assert.NotNil(t, job.(*jobRunner).deployResult)
	assert.NotNil(t, job.(*jobRunner).valuesOptions)
	assert.NotNil(t, job.(*jobRunner).messageCh)
	assert.NotNil(t, job.(*jobRunner).stopCtx)
	assert.NotNil(t, job.(*jobRunner).stopFn)
	assert.Equal(t, int64(10), job.(*jobRunner).timeoutSeconds)

	assert.Equal(t, GetSlugName(jobInput.NamespaceId, jobInput.Name), job.ID())
	assert.Equal(t, false, job.IsNotDryRun())
	assert.True(t, job.(*jobRunner).typeValidated())

	job.(*jobRunner).manifests = []string{"test"}
	assert.Equal(t, []string{"test"}, job.Manifests())
	job.(*jobRunner).err = errors.New("x")
	assert.Equal(t, "x", job.Error().Error())
	job.(*jobRunner).SetError(nil)
	assert.Nil(t, job.Error())

	assert.False(t, job.(*jobRunner).HasError())

	ctx, cancelFunc := context.WithCancelCause(context.TODO())
	job.(*jobRunner).stopCtx = ctx
	job.(*jobRunner).stopFn = cancelFunc
	assert.False(t, job.(*jobRunner).IsStopped())
	assert.Nil(t, job.(*jobRunner).GetStoppedErrorIfHas())
	cancelFunc(ErrCancel)
	assert.True(t, job.(*jobRunner).IsStopped())
	assert.Equal(t, ErrCancel, job.(*jobRunner).GetStoppedErrorIfHas())
}

func TestJobInput_Slug(t *testing.T) {
	jp := &JobInput{
		NamespaceId: 1,
		Name:        "test",
		DryRun:      true,
	}
	assert.Equal(t, jp.Slug(), GetSlugName(jp.NamespaceId, jp.Name))
	assert.False(t, jp.IsNotDryRun())
}

func TestEmptyPubSubMethods(t *testing.T) {
	e := NewEmptyPubSub()

	assert.NoError(t, e.Join(1), "Join should not return an error")
	assert.NoError(t, e.Leave(1, 1), "Leave should not return an error")
	assert.NoError(t, e.Run(context.TODO()), "Run should not return an error")
	assert.NoError(t, e.Publish(1, nil), "Publish should not return an error")
	assert.Nil(t, e.Info(), "Info should return nil")
	assert.Equal(t, "", e.Uid(), "Uid should return an empty string")
	assert.Equal(t, "", e.ID(), "ID should return an empty string")
	assert.NoError(t, e.ToSelf(nil), "ToSelf should not return an error")
	assert.NoError(t, e.ToAll(nil), "ToAll should not return an error")
	assert.Nil(t, e.Subscribe(), "Subscribe should return nil")
	assert.NoError(t, e.Close(), "Close should not return an error")
}

func TestInternalCloser(t *testing.T) {
	// Test when the function returns nil
	closer := NewCloser(func() error {
		return nil
	})

	err := closer.Close()
	assert.NoError(t, err, "Close should not return an error when the function returns nil")

	// Test when the function returns an error
	expectedErr := errors.New("close error")
	closer = NewCloser(func() error {
		return expectedErr
	})

	err = closer.Close()
	assert.Error(t, err, "Close should return an error when the function returns an error")
	assert.Equal(t, expectedErr, err, "Close should return the expected error")
}

func TestMatchDockerImage(t *testing.T) {
	tests := []struct {
		name     string
		vars     pipelineVars
		manifest string
		expected []string
	}{
		{
			name: "Single match with pipeline variable",
			vars: pipelineVars{
				Pipeline: "pipeline1",
			},
			manifest: `image: "docker.io/pipeline1:latest"`,
			expected: []string{"docker.io/pipeline1:latest"},
		},
		{
			name: "Multiple matches with pipeline variable",
			vars: pipelineVars{
				Commit: "commit123",
			},
			manifest: `image: "docker.io/image1:latest"
					   image: "docker.io/image2:commit123"`,
			expected: []string{"docker.io/image2:commit123"},
		},
		{
			name: "No matches with pipeline variable",
			vars: pipelineVars{
				Branch: "main",
			},
			manifest: `image: "docker.io/image1:latest"
					   image: "docker.io/image2:latest"`,
			expected: []string{"docker.io/image1:latest", "docker.io/image2:latest"},
		},
		{
			name: "Duplicate images",
			vars: pipelineVars{
				Branch: "main",
			},
			manifest: `image: "docker.io/image1:latest"
					   image: "docker.io/image1:latest"`,
			expected: []string{"docker.io/image1:latest"},
		},
		{
			name: "No images in manifest",
			vars: pipelineVars{
				Pipeline: "pipeline1",
			},
			manifest: ``,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchDockerImage(tt.vars, tt.manifest)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVars_ToKeyValue(t *testing.T) {
	v := vars{
		"key1": "value1",
		"key2": "value2",
	}

	expected := []*types.KeyValue{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}

	result := v.ToKeyValue()
	assert.ElementsMatch(t, expected, result, "ToKeyValue should return the correct key-value pairs")
}

func TestVars_MustGetString(t *testing.T) {
	v := vars{
		"key1": "value1",
		"key2": "value2",
	}

	assert.Equal(t, "value1", v.MustGetString("key1"), "MustGetString should return the correct value for key1")
	assert.Equal(t, "value2", v.MustGetString("key2"), "MustGetString should return the correct value for key2")
	assert.Equal(t, "", v.MustGetString("key3"), "MustGetString should return an empty string for a non-existent key")
}

func TestVars_Add(t *testing.T) {
	v := vars{}
	v.Add("key1", "value1")
	v.Add("key2", "value2")

	assert.Equal(t, "value1", v["key1"], "Add should correctly add key1 with value1")
	assert.Equal(t, "value2", v["key2"], "Add should correctly add key2 with value2")
}

func TestDeployResult_IsSet(t *testing.T) {
	dr := &deployResult{}
	assert.False(t, dr.IsSet(), "IsSet should return false when set is not true")

	dr.Set(1, "test message", &types.ProjectModel{Name: "test"})
	assert.True(t, dr.IsSet(), "IsSet should return true when set is true")
}

func TestDeployResult_Msg(t *testing.T) {
	dr := &deployResult{}
	assert.Equal(t, "", dr.Msg(), "Msg should return an empty string when msg is not set")

	dr.Set(1, "test message", &types.ProjectModel{Name: "test"})
	assert.Equal(t, "test message", dr.Msg(), "Msg should return the correct message")
}

func TestDeployResult_Model(t *testing.T) {
	dr := &deployResult{}
	assert.Nil(t, dr.Model(), "Model should return nil when model is not set")

	model := &types.ProjectModel{Name: "test"}
	dr.Set(1, "test message", model)
	assert.Equal(t, model, dr.Model(), "Model should return the correct model")
}

func TestDeployResult_ResultType(t *testing.T) {
	dr := &deployResult{}
	assert.Equal(t, websocket_pb.ResultType(0), dr.ResultType(), "ResultType should return the default value when result is not set")

	dr.Set(1, "test message", &types.ProjectModel{Name: "test"})
	assert.Equal(t, websocket_pb.ResultType(1), dr.ResultType(), "ResultType should return the correct result type")
}

func TestDeployResult_Set(t *testing.T) {
	dr := &deployResult{}
	model := &types.ProjectModel{Name: "test"}

	dr.Set(1, "test message", model)
	assert.True(t, dr.IsSet(), "IsSet should return true after Set is called")
	assert.Equal(t, "test message", dr.Msg(), "Msg should return the correct message after Set is called")
	assert.Equal(t, model, dr.Model(), "Model should return the correct model after Set is called")
	assert.Equal(t, websocket_pb.ResultType(1), dr.ResultType(), "ResultType should return the correct result type after Set is called")
}

func TestToProjectEventYaml(t *testing.T) {
	tests := []struct {
		name     string
		project  *biz.Project
		expected biz.YamlPrettier
	}{
		{
			name:     "Nil project",
			project:  nil,
			expected: nil,
		},
		{
			name: "Non-nil project",
			project: &biz.Project{
				GitCommitTitle:  "Initial commit",
				GitBranch:       "main",
				GitCommit:       "abc123",
				Atomic:          true,
				GitCommitWebURL: "http://example.com",
				Config:          "some config",
				EnvValues: []*types.KeyValue{
					{Key: "key2", Value: "value2"},
					{Key: "key1", Value: "value1"},
				},
				ExtraValues: []*websocket_pb.ExtraValue{
					{Path: "path2", Value: "value2"},
					{Path: "path1", Value: "value1"},
				},
				FinalExtraValues: []*websocket_pb.ExtraValue{
					{Path: "path3", Value: "value3"},
					{Path: "path1", Value: "value1"},
				},
			},
			expected: biz.AnyYamlPrettier{
				"title":   "Initial commit",
				"branch":  "main",
				"commit":  "abc123",
				"atomic":  true,
				"web_url": "http://example.com",
				"config":  "some config",
				"env_values": []*types.KeyValue{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
				"extra_values": []*websocket_pb.ExtraValue{
					{Path: "path1", Value: "value1"},
					{Path: "path2", Value: "value2"},
				},
				"final_extra_values": []*websocket_pb.ExtraValue{
					{Path: "path1", Value: "value1"},
					{Path: "path3", Value: "value3"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.project.ToEventYaml()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleMessage(t *testing.T) {
	jr := &jobRunner{
		logger:       mlog.NewForConfig(nil),
		messageCh:    newSafeWriteMessageCh(mlog.NewForConfig(nil), 1),
		deployResult: &deployResult{},
	}
	ctx, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	jr.HandleMessage(ctx)
	// ctx 已取消且未发送任何消息：走 ctx.Done 分支直接返回，deployResult 必须保持未设置。
	assert.False(t, jr.deployResult.IsSet(), "ctx 取消且无消息时 deployResult 不应被设置")
}

func TestHandleMessage_2(t *testing.T) {
	ch := newSafeWriteMessageCh(mlog.NewForConfig(nil), 1)
	jr := &jobRunner{
		logger:       mlog.NewForConfig(nil),
		messageCh:    ch,
		deployResult: &deployResult{},
	}
	ch.Close()
	jr.HandleMessage(context.TODO())
	// channel 已关闭且从未发送消息：走 !ok 分支直接返回，deployResult 必须保持未设置。
	assert.False(t, jr.deployResult.IsSet(), "channel 关闭且无消息时 deployResult 不应被设置")
}

func TestHandleMessage_3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	ch := newSafeWriteMessageCh(mlog.NewForConfig(nil), 10)
	jr := &jobRunner{
		logger:       mlog.NewForConfig(nil),
		messageCh:    ch,
		messager:     msger,
		deployResult: &deployResult{},
	}
	ch.Send(MessageItem{
		Msg:  "a",
		Type: MessageText,
		Containers: []*websocket_pb.Container{
			{
				Namespace: "a",
				Pod:       "b",
				Container: "c",
			},
		},
	})
	ch.Send(MessageItem{
		Msg:  "success",
		Type: MessageSuccess,
	})
	ch.Close()
	msger.EXPECT().SendMsgWithContainerLog("a", gomock.Any())
	jr.HandleMessage(context.TODO())
	assert.True(t, jr.deployResult.IsSet())
	assert.Equal(t, "success", jr.deployResult.Msg())
	assert.Equal(t, ResultDeployed, jr.deployResult.ResultType())
}

func TestHandleMessage_4(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	ch := newSafeWriteMessageCh(mlog.NewForConfig(nil), 10)
	jr := &jobRunner{
		logger:       mlog.NewForConfig(nil),
		messageCh:    ch,
		messager:     msger,
		deployResult: &deployResult{},
		stopCtx:      context.TODO(),
	}
	ch.Send(MessageItem{
		Msg:  "err",
		Type: MessageError,
	})
	ch.Close()
	jr.HandleMessage(context.TODO())
	assert.True(t, jr.deployResult.IsSet())
	assert.Equal(t, "err", jr.deployResult.Msg())
	assert.Equal(t, ResultDeployFailed, jr.deployResult.ResultType())
}

func TestHandleMessage_5(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	ctx, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	ch := newSafeWriteMessageCh(mlog.NewForConfig(nil), 10)
	jr := &jobRunner{
		logger:       mlog.NewForConfig(nil),
		messageCh:    ch,
		messager:     msger,
		deployResult: &deployResult{},
		stopCtx:      ctx,
	}
	ch.Send(MessageItem{
		Msg:  "err",
		Type: MessageError,
	})
	ch.Close()
	jr.HandleMessage(context.TODO())
	assert.True(t, jr.deployResult.IsSet())
	assert.Equal(t, ResultDeployCanceled, jr.deployResult.ResultType())
}

type fakeChartLoader struct {
	loadDirErr error
	c          *chart.Chart
}

func (f *fakeChartLoader) LoadArchive(in io.Reader) (*chart.Chart, error) {
	return f.c, nil
}

func (f *fakeChartLoader) LoadDir(dir string) (*chart.Chart, error) {
	if f.loadDirErr != nil {
		return nil, f.loadDirErr
	}
	return f.c, nil
}

type fakeOpener struct{}

func (f *fakeOpener) Open(name string) (io.ReadCloser, error) {
	return io.NopCloser(nil), nil
}

func (f *fakeOpener) Close() error {
	return nil
}

func TestUserConfigLoader_Load(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().To(gomock.Any()).AnyTimes()
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ctx := &LoadContext{
		Input: &JobInput{
			Config: "xxxx",
		},
		Messager: em,
		Config: &mars.Config{
			ConfigField: "app->config",
			IsSimpleEnv: true,
		},
	}
	assert.Nil(t, (&UserConfigLoader{}).Load(ctx))
	assert.Equal(t,
		`app:
  config: xxxx
`, ctx.UserConfigYaml)
	ctx2 := &LoadContext{
		Input: &JobInput{
			Config: "name: duc\nage: 17",
		},
		Messager: em,
		Config: &mars.Config{
			ConfigField: "app->config",
			IsSimpleEnv: false,
		},
	}
	assert.Nil(t, (&UserConfigLoader{}).Load(ctx2))
	assert.Equal(t,
		`app:
  config:
    age: 17
    name: duc
`, ctx2.UserConfigYaml)

	ctx3 := &LoadContext{
		Input:    &JobInput{},
		Messager: em,
	}
	assert.Nil(t, (&UserConfigLoader{}).Load(ctx3))
}

func TestElementsLoader_Load(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().To(gomock.Any()).AnyTimes()
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ctx := &LoadContext{
		Input: &JobInput{
			ExtraValues: []*websocket_pb.ExtraValue{
				{
					Path:  "app->config",
					Value: "1",
				},
			},
		},
		Messager: em,
		Config: &mars.Config{
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

	assert.Nil(t, (&ElementsLoader{}).Load(ctx))
	sort.Strings(ctx.ElementValues)
	assert.Equal(t,
		`app:
  config: "1"
`,
		ctx.ElementValues[0])
	assert.Equal(t,
		`app:
  xxx: true
`,
		ctx.ElementValues[1])

	err := (&ElementsLoader{}).Load(&LoadContext{
		Input: &JobInput{
			ExtraValues: []*websocket_pb.ExtraValue{
				{
					Path:  "app->config",
					Value: "4",
				},
			},
		},
		Messager: em,
		Config: &mars.Config{
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

	ctx2 := &LoadContext{
		Input: &JobInput{
			ExtraValues: []*websocket_pb.ExtraValue{
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
		Messager: em,
		Config: &mars.Config{
			Elements: []*mars.Element{
				{
					Path:    "duc",
					Type:    mars.ElementType_ElementTypeInput,
					Default: "input",
				},
			},
		},
	}
	err = (&ElementsLoader{}).Load(ctx2)
	assert.Nil(t, err)
	assert.Equal(t, []string{"duc: xxx\n"}, ctx2.ElementValues)

	ctx3 := &LoadContext{
		Input:    &JobInput{},
		Messager: em,
		Config:   &mars.Config{},
	}
	assert.Nil(t, (&ElementsLoader{}).Load(ctx3))
}

func TestElementsLoader_deepSetItems(t *testing.T) {
	items := (&ElementsLoader{}).deepSetItems(map[string]any{"a": "a"})
	assert.Equal(t, "a: a\n", items[0])
	items = (&ElementsLoader{}).deepSetItems(map[string]any{"a->b": "ab"})
	assert.Equal(t,
		`a:
  b: ab
`, items[0])
}

func TestElementsLoader_typedValue(t *testing.T) {
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
			result: "",
			err:    "[ElementsLoader]: '2x' 非 number 类型, 无法转换",
		},
	}

	for i, test := range tests {
		tt := test
		t.Run(fmt.Sprintf("test-%v", i), func(t *testing.T) {
			t.Parallel()
			value, err := (&ElementsLoader{}).typedValue(tt.ele, tt.input)
			if err != nil {
				assert.Equal(t, err.Error(), tt.err)
			} else {
				assert.Equal(t, tt.result, value)
			}
		})
	}
}

func TestJober_GlobalLock(t *testing.T) {
	l := locker.NewMemoryLock(timer.NewReal(), [2]int{2, 100}, locker.NewMemStore(), mlog.NewForConfig(nil))
	job := &jobRunner{locker: l, input: &JobInput{NamespaceId: 1, Name: "app"}}
	assert.Nil(t, job.GlobalLock().Error())
	assert.Equal(t, "正在部署中，请稍后再试", (&jobRunner{locker: l, input: &JobInput{NamespaceId: 1, Name: "app"}}).GlobalLock().Error().Error())
	assert.Len(t, job.finallyCallback.Sort(), 1)
	called := 0
	pipeline.New[error]().Send(nil).Through(job.finallyCallback.Sort()...).Then(func(e error) {
		called++
		assert.Nil(t, e)
	})
	assert.Equal(t, 1, called)
	acquire := l.Acquire("id", 100)
	assert.True(t, acquire)

	m := gomock.NewController(t)
	defer m.Finish()
	ml := locker.NewMockLocker(m)
	ml.EXPECT().RenewalAcquire(GetSlugName(1, "app"), 30, 20).Times(0)
	assert.Equal(t, "xxx", (&jobRunner{err: errors.New("xxx"), locker: ml}).GlobalLock().Error().Error())
}

type emptyLoader struct {
	err    error
	called bool
	sync.Mutex
}

func (e *emptyLoader) Load(ctx *LoadContext) error {
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
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	ctx, fn := context.WithCancel(context.TODO())
	fn()
	l := &emptyLoader{}
	msger.EXPECT().SendMsg(gomock.Any())
	assert.Equal(t, "context canceled", (&jobRunner{
		stopCtx:  ctx,
		logger:   mlog.NewForConfig(nil),
		messager: msger,
		loaders:  []Loader{l},
	}).LoadConfigs().Error().Error())
	assert.False(t, l.GetCalled())
}

func TestJober_LoadConfigs(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	assert.Equal(t, "xxx", (&jobRunner{
		err:      errors.New("xxx"),
		loaders:  []Loader{},
		logger:   mlog.NewForConfig(nil),
		messager: msger,
	}).LoadConfigs().Error().Error())

	l := &emptyLoader{}
	assert.Nil(t, (&jobRunner{
		stopCtx:  context.TODO(),
		loaders:  []Loader{l},
		logger:   mlog.NewForConfig(nil),
		messager: msger,
	}).LoadConfigs().Error())
	assert.True(t, l.GetCalled())

	l2 := &emptyLoader{}
	cancel, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	assert.Equal(t, "context canceled", (&jobRunner{
		stopCtx:  cancel,
		loaders:  []Loader{l2},
		logger:   mlog.NewForConfig(nil),
		messager: msger,
	}).LoadConfigs().Error().Error())
	assert.False(t, l2.GetCalled())

	l3 := &emptyLoader{
		err: errors.New("xxx"),
	}
	assert.Equal(t, "xxx", (&jobRunner{
		stopCtx:  context.TODO(),
		loaders:  []Loader{l3},
		logger:   mlog.NewForConfig(nil),
		messager: msger,
	}).LoadConfigs().Error().Error())
	assert.True(t, l3.GetCalled())
}

// cleanupLoader 在加载期登记一个临时资源清理函数，用于验证 LoadConfigs 把
// ctx.cleanups 收口为 finally 回调（成功与失败两个分支都必须登记）。
type cleanupLoader struct {
	err     error
	cleanup func()
}

func (c *cleanupLoader) Load(ctx *LoadContext) error {
	ctx.AddCleanup(c.cleanup)
	return c.err
}

func TestJober_LoadConfigs_registerCleanup(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()

	// 链成功：清理函数被收口为 finally 回调，部署结束时执行一次
	var cleaned int
	j := (&jobRunner{
		stopCtx:  context.TODO(),
		loaders:  []Loader{&cleanupLoader{cleanup: func() { cleaned++ }}},
		logger:   mlog.NewForConfig(nil),
		messager: msger,
	}).LoadConfigs()
	assert.Nil(t, j.Error())
	assert.Len(t, j.(*jobRunner).finallyCallback.Sort(), 1)
	pipeline.New[error]().Send(nil).Through(j.(*jobRunner).finallyCallback.Sort()...).Then(func(error) {})
	assert.Equal(t, 1, cleaned)

	// 链失败但已登记清理：仍需注册回收，避免下载目录/临时文件泄漏
	cleaned = 0
	j = (&jobRunner{
		stopCtx:  context.TODO(),
		loaders:  []Loader{&cleanupLoader{err: errors.New("load err"), cleanup: func() { cleaned++ }}},
		logger:   mlog.NewForConfig(nil),
		messager: msger,
	}).LoadConfigs()
	assert.Equal(t, "load err", j.Error().Error())
	assert.Len(t, j.(*jobRunner).finallyCallback.Sort(), 1)
	pipeline.New[error]().Send(nil).Through(j.(*jobRunner).finallyCallback.Sort()...).Then(func(error) {})
	assert.Equal(t, 1, cleaned)
}

func TestJober_Stop(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msg := NewMockDeployMsger(m)
	msg.EXPECT().SendMsg(gomock.Any()).Times(3)
	var called int64 = 0
	j := &jobRunner{messager: msg, logger: mlog.NewForConfig(nil), stopFn: func(err error) {
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

func TestJober_OnError(t *testing.T) {
	job := &jobRunner{err: errors.New("xxx")}
	job.OnError(1, func(err error, sendResultToUser func()) {
		assert.Equal(t, "xxx", err.Error())
		sendResultToUser()
	})
	assert.Len(t, job.errorCallback.Sort(), 1)
	called := 0
	pipeline.New[error]().Send(job.Error()).Through(job.errorCallback.Sort()...).Then(func(err error) {
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
	pipeline.New[error]().Send(job.Error()).Through(job.finallyCallback.Sort()...).Then(func(err error) {
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
			pipeline.New[error]().Send(job.Error()).Through(job.finallyCallback.Sort()...).Then(func(err error) {
				assert.Equal(t, tt, err)
				called++
			})
			assert.Equal(t, 1, called)
		})
	}
}

func Test_jobRunner_Project(t *testing.T) {
	job := &jobRunner{project: &biz.Project{}}
	assert.NotNil(t, job.Project())
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
	up := uploader.NewMockUploader(m)
	finfo := uploader.NewMockFileInfo(m)
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

	up.EXPECT().LocalUploader().Return(up)
	msger := NewMockDeployMsger(m)
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ctx := &LoadContext{
		UserConfigYaml:   dcy,
		ImagePullSecrets: []string{"secret"},
		ElementValues:    []string{ev1, ev2},
		ValuesOptions:    &values.Options{},
		Input:            &JobInput{GitBranch: "dev"},
		Messager:         msger,
		SystemValuesYaml: vy,
		timer:            timer.NewReal(),
		uploader:         up,
	}
	assert.Nil(t, (&MergeValuesLoader{}).Load(ctx))
	assert.Equal(t, "/app/config.yaml", ctx.ValuesOptions.ValueFiles[0])

	ctx2 := &LoadContext{
		SystemValuesYaml: "",
		ImagePullSecrets: nil,
		UserConfigYaml:   "",
		ValuesOptions:    &values.Options{},
		Input:            &JobInput{GitBranch: "dev"},
		Messager:         msger,
	}
	assert.Nil(t, (&MergeValuesLoader{}).Load(ctx2))
}

func TestSystemVariableLoader_Load_ok1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	msger.EXPECT().To(gomock.Any()).Times(1)
	msger.EXPECT().SendMsg(gomock.Any()).Times(2)
	assert.Nil(t, (&SystemVariableLoader{}).Load(&LoadContext{
		Config:   &mars.Config{ValuesYaml: ""},
		Messager: msger,
	}))
}
func TestSystemVariableLoader_Load_ok(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	gitS := app.NewMockGitServer(m)
	gitS.EXPECT().GetCommitPipeline("10", "dev", "c").Return(&biz.Pipeline{Ref: "dev"}, nil)

	pl := app.NewMockPluginManager(m)
	pl.EXPECT().Git().Return(gitS)
	domain := app.NewMockDomainManager(m)
	pl.EXPECT().Domain().Return(domain).AnyTimes()
	domain.EXPECT().GetDomainByIndex(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	domain.EXPECT().GetCertSecretName("app", gomock.Any()).AnyTimes()
	domain.EXPECT().GetClusterIssuer().Return("cluster-issuer")
	em := NewMockDeployMsger(m)
	em.EXPECT().To(gomock.Any()).AnyTimes()
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ctx := &LoadContext{
		Commit:    &biz.Commit{ShortID: "short_id"},
		PluginMgr: pl,
		Config: &mars.Config{
			ValuesYaml: `
VarImagePullSecrets: <.ImagePullSecrets>
`,
		},
		Input: &JobInput{
			GitBranch: "dev",
			GitCommit: "c",
		},
		Logger: mlog.NewForConfig(nil),
		Project: &biz.Project{
			Name: "app",
		},
		Namespace:        &biz.Namespace{Name: "ns"},
		ImagePullSecrets: []string{"a", "b", "c"},
		Messager:         em,
		Repo:             &biz.Repo{NeedGitRepo: true, GitProjectID: 10},
	}
	assert.Nil(t, (&SystemVariableLoader{}).Load(ctx))
	assert.Equal(t, `
VarImagePullSecrets: [{name: a}, {name: b}, {name: c}, ]
`,
		ctx.SystemValuesYaml)
	assert.Equal(t, "dev", ctx.Vars[VarBranch])
	assert.Equal(t, "short_id", ctx.Vars[VarCommit])
	assert.Equal(t, "0", ctx.Vars[VarPipeline])
	assert.Equal(t, "c", ctx.Vars[VarLongCommit])
	assert.Equal(t, "ns", ctx.Vars[VarNamespace])
	assert.Equal(t, "[{name: a}, {name: b}, {name: c}, ]", ctx.Vars[VarImagePullSecrets])
}

func TestSystemVariableLoader_Load_fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	gitS := app.NewMockGitServer(m)
	gitS.EXPECT().GetCommitPipeline(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("x"))

	pl := app.NewMockPluginManager(m)
	pl.EXPECT().Git().Return(gitS)
	domain := app.NewMockDomainManager(m)
	pl.EXPECT().Domain().Return(domain).AnyTimes()
	domain.EXPECT().GetDomainByIndex(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	domain.EXPECT().GetCertSecretName(gomock.Any(), gomock.Any()).AnyTimes()
	em := NewMockDeployMsger(m)
	em.EXPECT().To(gomock.Any()).AnyTimes()
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ctx := &LoadContext{
		Commit:    &biz.Commit{ShortID: "short_id"},
		PluginMgr: pl,
		Input: &JobInput{
			GitBranch: "dev",
		},
		Config: &mars.Config{
			ValuesYaml: `
VarImagePullSecrets: <.ImagePullSecrets>
image: <.Pipeline>-<.Branch>
`,
		},
		Logger: mlog.NewForConfig(nil),
		Project: &biz.Project{
			Name: "app",
		},
		Namespace:        &biz.Namespace{Name: "ns"},
		ImagePullSecrets: []string{"a", "b", "c"},
		Messager:         em,
		Repo:             &biz.Repo{NeedGitRepo: true, GitProjectID: 1},
	}
	assert.Error(t, (&SystemVariableLoader{}).Load(ctx))
}

func TestChartFileLoader_Load2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	h := data.NewMockHelmerRepo(m)
	pl := app.NewMockPluginManager(m)
	gits := app.NewMockGitServer(m)
	up := uploader.NewMockUploader(m)
	pl.EXPECT().Git().Return(gits).AnyTimes()
	ctx := &LoadContext{
		uploader:  up,
		Helmer:    h,
		PluginMgr: pl,
		Input: &JobInput{
			GitCommit: "commit",
		},
		Messager: em,
		Config: &mars.Config{
			LocalChartPath: "100|main|dir",
		},
		Logger: mlog.NewForConfig(nil),
	}
	l := &ChartFileLoader{
		chartLoader: &fakeChartLoader{
			c: &chart.Chart{
				Metadata: &chart.Metadata{},
			},
		},
		fileOpener: &fakeOpener{},
	}

	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	em.EXPECT().To(gomock.Any())
	gits.EXPECT().GetDirectoryFilesWithBranch("100", "main", "dir", true).Return([]string{"file1", "file2"}, nil)
	up.EXPECT().LocalUploader().Return(up)
	up.EXPECT().MkDir(gomock.Any(), true).Return(errors.New("mkdir err")).Times(1)

	err := l.Load(ctx)
	assert.Equal(t, "mkdir err", err.Error())
}

func TestChartFileLoader_Load(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	em.EXPECT().To(gomock.Any()).AnyTimes()
	h := data.NewMockHelmerRepo(m)
	gits := app.NewMockGitServer(m)
	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return(nil, errors.New("xxx"))

	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return([]string{"file1", "file2"}, nil)
	gits.EXPECT().GetFileContentWithSha("9999", "master", "file1").Return("file1", nil).Times(1)
	gits.EXPECT().GetFileContentWithSha("9999", "master", "file2").Return("file2", nil).Times(1)
	up := uploader.NewMockUploader(m)
	up.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().AbsolutePath(gomock.Any()).Return("/dir")
	up.EXPECT().MkDir(gomock.Any(), true).Times(1)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Times(2)
	h.EXPECT().PackageChart(gomock.Any(), gomock.Any()).Times(1)
	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir/xxxx", true).Return([]string{}, nil)

	pl := app.NewMockPluginManager(m)
	pl.EXPECT().Git().Return(gits).AnyTimes()
	ctx := &LoadContext{
		uploader: up,
		Logger:   mlog.NewForConfig(nil),
		Helmer:   h,
		Input:    &JobInput{},
		Messager: em,
		Config: &mars.Config{
			LocalChartPath: "9999|master|dir",
		},
		PluginMgr: pl,
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

	err := l.Load(ctx)
	// git 故障如实上报（旧实现吞掉后误报"charts 文件不存在"）
	assert.Equal(t, "获取远程 charts 文件: xxx", err.Error())
	err = l.Load(ctx)
	assert.Len(t, ctx.cleanups, 2)
	assert.Nil(t, err)

	up.EXPECT().DeleteDir("/dir").Times(2)
	for _, c := range ctx.cleanups {
		c()
	}
}

func TestChartFileLoader_LoadWithChartMissing(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	em.EXPECT().To(gomock.Any()).AnyTimes()
	h := data.NewMockHelmerRepo(m)
	pl := app.NewMockPluginManager(m)
	gits := app.NewMockGitServer(m)
	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return([]string{"file1", "file2"}, nil)

	gits.EXPECT().GetFileContentWithSha("9999", "master", "file1").Return("file1", nil).Times(1)
	gits.EXPECT().GetFileContentWithSha("9999", "master", "file2").Return("file2", nil).Times(1)
	up := uploader.NewMockUploader(m)
	up.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().AbsolutePath(gomock.Any()).Return("/dir")
	up.EXPECT().MkDir(gomock.Any(), true).Times(1)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Times(2)
	pl.EXPECT().Git().Return(gits).AnyTimes()
	ctx := &LoadContext{
		Helmer:    h,
		PluginMgr: pl,
		uploader:  up,
		Logger:    mlog.NewForConfig(nil),
		Input:     &JobInput{},
		Messager:  em,
		Config: &mars.Config{
			LocalChartPath: "9999|master|dir",
		},
	}
	loadDirErr := errors.New("Chart.yaml file is missing")
	l := &ChartFileLoader{
		chartLoader: &fakeChartLoader{
			loadDirErr: loadDirErr,
		},
	}

	err := l.Load(ctx)
	assert.Equal(t, loadDirErr.Error(), err.Error())

	ctx.Config.LocalChartPath = "xxx"
	err = l.Load(ctx)
	assert.Equal(t, "LocalChartPath 格式不正确: xxx", err.Error())
}

func TestLoadContext_WriteConfigYamlToTmpFile(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockUploader.EXPECT().LocalUploader().Return(mockUploader).Times(2)
	ctx := &LoadContext{
		timer:    timer.NewReal(),
		uploader: mockUploader,
		Logger:   mlog.NewForConfig(nil),
	}
	info := uploader.NewMockFileInfo(m)
	mockUploader.EXPECT().Put(gomock.Not(nil), bytes.NewReader([]byte("xxx"))).Return(info, nil)
	info.EXPECT().Path().Return("/path")
	path, closer, err := ctx.WriteConfigYamlToTmpFile([]byte("xxx"))
	assert.Equal(t, "/path", path)
	assert.Nil(t, err)
	mockUploader.EXPECT().Delete("/path").Return(errors.New("x"))
	err = closer.Close()
	assert.Equal(t, "x", err.Error())
}

func Test_jobRunner_Validate_Fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	msger := NewMockDeployMsger(m)
	repoRepo := data.NewMockRepoRepo(m)
	projectRepo := data.NewMockProjectRepo(m)

	var job *jobRunner
	job = &jobRunner{
		input: &JobInput{Type: websocket_pb.Type_HandleAuthorize},
	}
	assert.Error(t, job.Validate().Error())

	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("xxx"))
	job = &jobRunner{
		messager: msger,
		nsRepo:   nsRepo,
		input:    &JobInput{Type: websocket_pb.Type_CreateProject, NamespaceId: 1},
	}
	assert.Error(t, job.Validate().Error())

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)
	repoRepo.EXPECT().Get(gomock.Any(), 12).Return(nil, errors.New("xxx"))
	job = &jobRunner{
		messager: msger,
		nsRepo:   nsRepo,
		repoRepo: repoRepo,
		input:    &JobInput{Type: websocket_pb.Type_CreateProject, NamespaceId: 1, RepoID: 12},
	}
	assert.Error(t, job.Validate().Error())

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1}, nil)
	repoRepo.EXPECT().Get(gomock.Any(), 12).Return(&biz.Repo{
		MarsConfig: &mars.Config{},
	}, nil)
	// 非 NotFound 错误（DB 抖动/网络）必须 fail-fast，绝不误走"新建项目"分支
	projectRepo.EXPECT().FindByName(gomock.Any(), "xx", 1).Return(nil, errors.New("xxx"))
	job = &jobRunner{
		logger:   mlog.NewForConfig(nil),
		messager: msger,
		nsRepo:   nsRepo,
		projRepo: projectRepo,
		repoRepo: repoRepo,
		user:     &biz.UserInfo{},
		input: &JobInput{
			Type:        websocket_pb.Type_CreateProject,
			NamespaceId: 1,
			Name:        "xx",
			RepoID:      12,
			DryRun:      false,
		},
	}
	assert.Error(t, job.Validate().Error())

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1}, nil)
	repoRepo.EXPECT().Get(gomock.Any(), 12).Return(&biz.Repo{
		MarsConfig: &mars.Config{},
	}, nil)
	projectRepo.EXPECT().FindByName(gomock.Any(), "xx", 1).Return(&biz.Project{}, nil)
	projectRepo.EXPECT().UpdateStatusByVersion(gomock.Any(), 1, types.Deploy_StatusDeploying, gomock.Any()).Return(nil, errors.New("xxx"))
	job = &jobRunner{
		logger:   mlog.NewForConfig(nil),
		messager: msger,
		nsRepo:   nsRepo,
		projRepo: projectRepo,
		repoRepo: repoRepo,
		user:     &biz.UserInfo{},
		input: &JobInput{
			Type:        websocket_pb.Type_CreateProject,
			NamespaceId: 1,
			Name:        "xx",
			RepoID:      12,
			DryRun:      false,
			ProjectID:   1,
		},
	}
	assert.Error(t, job.Validate().Error())
}

func Test_jobRunner_Validate_FindByName_NotFound_CreatesProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	msger := NewMockDeployMsger(m)
	repoRepo := data.NewMockRepoRepo(m)
	projectRepo := data.NewMockProjectRepo(m)
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()

	// 只有确切的 NotFound 才走"新建项目"分支；data 边界将 ent.NotFoundError 转成
	// gRPC NotFound，mock 里用 errs.WrapNotFound() 构造等价的 NotFound，断言错误
	// 源自 Create 而非 FindByName，与 Test_jobRunner_Validate_Fail 里的非 NotFound
	// fail-fast 用例形成判别对。
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1}, nil)
	repoRepo.EXPECT().Get(gomock.Any(), 12).Return(&biz.Repo{
		MarsConfig: &mars.Config{},
	}, nil)
	projectRepo.EXPECT().FindByName(gomock.Any(), "xx", 1).Return(nil, errs.WrapNotFound(errors.New("not found"), "find project by name"))
	projectRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("create failed"))

	job := &jobRunner{
		logger:   mlog.NewForConfig(nil),
		messager: msger,
		nsRepo:   nsRepo,
		projRepo: projectRepo,
		repoRepo: repoRepo,
		user:     &biz.UserInfo{},
		input: &JobInput{
			Type:        websocket_pb.Type_CreateProject,
			NamespaceId: 1,
			Name:        "xx",
			RepoID:      12,
			DryRun:      false,
		},
	}
	assert.Error(t, job.Validate().Error())
	// 证明错误来自 Create 而非 FindByName 的 NotFoundError
	assert.Contains(t, job.Error().Error(), "create failed")
}

func Test_jobRunner_Validate_FindByName_NotFound_CreatesProject_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	msger := NewMockDeployMsger(m)
	repoRepo := data.NewMockRepoRepo(m)
	projectRepo := data.NewMockProjectRepo(m)
	sub := app.NewMockPubSub(m)
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	sub.EXPECT().ToAll(gomock.Any())

	// create 成功路径：覆盖 createdID 赋值与 OnError 清理回调注册（失败用例覆盖不到）
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1}, nil)
	repoRepo.EXPECT().Get(gomock.Any(), 12).Return(&biz.Repo{
		MarsConfig:  &mars.Config{},
		NeedGitRepo: false,
	}, nil)
	projectRepo.EXPECT().FindByName(gomock.Any(), "xx", 1).Return(nil, errs.WrapNotFound(errors.New("not found"), "find project by name"))
	projectRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&biz.Project{ID: 7}, nil)

	job := &jobRunner{
		logger:   mlog.NewForConfig(nil),
		messager: msger,
		nsRepo:   nsRepo,
		projRepo: projectRepo,
		repoRepo: repoRepo,
		user:     &biz.UserInfo{},
		input: &JobInput{
			Type:        websocket_pb.Type_CreateProject,
			NamespaceId: 1,
			Name:        "xx",
			RepoID:      12,
			DryRun:      false,
			PubSub:      sub,
		},
	}
	assert.Nil(t, job.Validate().Error())
	assert.True(t, job.isNew)
	assert.Equal(t, 7, job.project.ID)
}

// dry-run 新建项目不落库，但必须合成占位 project，否则 Run 的 ReleaseName
// 与 loader 的 ctx.Project.Name 会对 nil 解引用 panic。
func Test_jobRunner_Validate_FindByName_NotFound_DryRun_SynthPlaceholderProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	msger := NewMockDeployMsger(m)
	repoRepo := data.NewMockRepoRepo(m)
	projectRepo := data.NewMockProjectRepo(m)
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1}, nil)
	repoRepo.EXPECT().Get(gomock.Any(), 12).Return(&biz.Repo{
		MarsConfig:  &mars.Config{},
		NeedGitRepo: false,
	}, nil)
	projectRepo.EXPECT().FindByName(gomock.Any(), "xx", 1).Return(nil, errs.WrapNotFound(errors.New("not found"), "find project by name"))
	// dry-run 不落库：不 mock Create，gomock 对意外调用即失败

	job := &jobRunner{
		logger:   mlog.NewForConfig(nil),
		messager: msger,
		nsRepo:   nsRepo,
		projRepo: projectRepo,
		repoRepo: repoRepo,
		user:     &biz.UserInfo{},
		input: &JobInput{
			Type:        websocket_pb.Type_CreateProject,
			NamespaceId: 1,
			Name:        "xx",
			RepoID:      12,
			DryRun:      true,
		},
		dryRun: true,
	}
	assert.Nil(t, job.Validate().Error())
	assert.True(t, job.isNew)
	require.NotNil(t, job.project)
	assert.Equal(t, "xx", job.project.Name)
	assert.Equal(t, "xx", job.Project().Name)
}

func Test_jobRunner_Validate_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	msger := NewMockDeployMsger(m)
	repoRepo := data.NewMockRepoRepo(m)
	projectRepo := data.NewMockProjectRepo(m)
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()

	var job *jobRunner

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1}, nil)
	repoRepo.EXPECT().Get(gomock.Any(), 12).Return(&biz.Repo{
		MarsConfig:  &mars.Config{},
		NeedGitRepo: false,
	}, nil)
	projectRepo.EXPECT().FindByName(gomock.Any(), "xx", 1).Return(&biz.Project{}, nil)
	projectRepo.EXPECT().UpdateStatusByVersion(gomock.Any(), 100, types.Deploy_StatusDeploying, 10).Return(&biz.Project{}, nil)
	sub := app.NewMockPubSub(m)
	sub.EXPECT().ToAll(gomock.Any())
	job = &jobRunner{
		logger:   mlog.NewForConfig(nil),
		messager: msger,
		nsRepo:   nsRepo,
		projRepo: projectRepo,
		repoRepo: repoRepo,
		user:     &biz.UserInfo{},
		input: &JobInput{
			Type:        websocket_pb.Type_CreateProject,
			NamespaceId: 1,
			Name:        "xx",
			RepoID:      12,
			DryRun:      false,
			PubSub:      sub,
			ProjectID:   100,
			Version:     lo.ToPtr(int32(10)),
		},
	}

	assert.Nil(t, job.Validate().Error())
	assert.NotNil(t, job.commit)
}

func TestJober_Finish_WhenError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	stopCtx, stopFn := context.WithCancelCause(context.TODO())
	stopFn(errors.New("stopped"))
	job := &jobRunner{
		deployResult: &deployResult{},
		logger:       mlog.NewForConfig(nil),
		err:          errors.New("xxx"),
		messager:     msger,
		stopCtx:      stopCtx,
		stopFn:       stopFn,
	}
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
	job2 := &jobRunner{
		deployResult: &deployResult{},
		logger:       mlog.NewForConfig(nil),
		err:          errors.New("xxx"),
		messager:     msger,
		stopCtx:      context.TODO(),
	}
	msger.EXPECT().SendDeployedResult(websocket_pb.ResultType_DeployedFailed, "xxx", nil).Times(1)
	job2.Finish()
}

func TestJober_Finish_WhenSuccess(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	job := &jobRunner{
		deployResult: &deployResult{},
		messager:     msger, logger: mlog.NewForConfig(nil),
	}
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
	job.deployResult.Set(websocket_pb.ResultType_Deployed, "ok", nil)
	// success
	assert.Nil(t, job.Finish().Error())
	assert.Equal(t, 1, finallyCalled)
	assert.Equal(t, 1, successCalled)
	assert.Equal(t, 0, errorCalled)
}

func Test_jobRunner_Run_Fail(t *testing.T) {
	assert.Error(t, (&jobRunner{
		err: errors.New("x"),
	}).Run(context.TODO()).Error())
}

func Test_jobRunner_Run_Fail_2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	installer := NewMockReleaseInstaller(m)
	msger := NewMockDeployMsger(m)
	k8sRepo := data.NewMockK8sRepo(m)
	projRepo := data.NewMockProjectRepo(m)
	eventRepo := data.NewMockEventRepo(m)
	messageChan := NewMockSafeWriteMessageChan(m)
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	installer.EXPECT().Run(gomock.Any(), gomock.Any()).Return(nil, errors.New("xx"))
	messageChan.EXPECT().Send(gomock.Any())
	messageChan.EXPECT().Close()
	messageChan.EXPECT().Chan().Return(make(chan MessageItem, 1))
	jb := &jobRunner{
		logger:       mlog.NewForConfig(nil),
		projRepo:     projRepo,
		k8sRepo:      k8sRepo,
		eventRepo:    eventRepo,
		messager:     msger,
		installer:    installer,
		ns:           &biz.Namespace{},
		project:      &biz.Project{},
		deployResult: &deployResult{},
		input:        &JobInput{},
		commit:       &biz.Commit{},
		messageCh:    messageChan,
	}
	assert.Error(t, jb.Run(context.TODO()).Error())
}

func Test_jobRunner_Run_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	installer := NewMockReleaseInstaller(m)
	msger := NewMockDeployMsger(m)
	k8sRepo := data.NewMockK8sRepo(m)
	projRepo := data.NewMockProjectRepo(m)
	eventRepo := data.NewMockEventRepo(m)
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ch := newSafeWriteMessageCh(mlog.NewForConfig(nil), 10)

	jb := &jobRunner{
		logger:    mlog.NewForConfig(nil),
		projRepo:  projRepo,
		k8sRepo:   k8sRepo,
		eventRepo: eventRepo,
		messager:  msger,
		installer: installer,
		config:    &mars.Config{},
		chart: &chart.Chart{
			Metadata: &chart.Metadata{},
		},
		ns:           &biz.Namespace{},
		project:      &biz.Project{},
		deployResult: &deployResult{},
		input:        &JobInput{},
		user:         &biz.UserInfo{Name: "duc"},
		commit:       &biz.Commit{},
		messageCh:    ch,
	}

	installer.EXPECT().Run(gomock.Any(), &InstallInput{
		IsNew:        jb.isNew,
		Wait:         lo.FromPtr(jb.input.Atomic),
		Chart:        jb.chart,
		ValueOptions: jb.valuesOptions,
		DryRun:       jb.dryRun,
		ReleaseName:  jb.project.Name,
		Namespace:    jb.ns.Name,
		Description:  jb.commit.Title,
		messageChan:  jb.messageCh,
		percenter:    jb.messager,
	}).Return(&release.Release{
		Config: map[string]any{},
	}, nil)
	projRepo.EXPECT().UpdateProject(gomock.Any(), gomock.Any()).Return(&biz.Project{}, nil)
	eventRepo.EXPECT().Dispatch(biz.EventProjectChanged, gomock.Any())
	eventRepo.EXPECT().AuditLogWithChange(
		types.EventActionType_Update, "duc", "",
		gomock.Any(), gomock.Any(),
		gomock.Any())
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	k8sRepo.EXPECT().SplitManifests(gomock.Any())
	k8sRepo.EXPECT().GetPodSelectorsByManifest(gomock.Any())

	assert.Nil(t, jb.Run(context.TODO()).Error())
}

// 新建路径下 Create 成功后注册的 OnError 清理回调（删除项目）与 OnFinally 状态回收
// 回调，只有 Finish 报错时才会真正执行；该用例通过 SetError + Finish 触发它们。
func Test_jobRunner_Validate_CreatePath_OnErrorOnFinally(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	msger := NewMockDeployMsger(m)
	repoRepo := data.NewMockRepoRepo(m)
	projectRepo := data.NewMockProjectRepo(m)
	sub := app.NewMockPubSub(m)
	helmer := data.NewMockHelmerRepo(m)
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1}, nil)
	repoRepo.EXPECT().Get(gomock.Any(), 12).Return(&biz.Repo{
		MarsConfig:  &mars.Config{},
		NeedGitRepo: false,
	}, nil)
	projectRepo.EXPECT().FindByName(gomock.Any(), "xx", 1).Return(nil, errs.WrapNotFound(errors.New("not found"), "find project by name"))
	projectRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&biz.Project{ID: 7}, nil)
	// OnError 清理回调：删除刚创建的项目
	projectRepo.EXPECT().Delete(gomock.Any(), 7).Return(nil)
	// OnFinally 回调：按 helm 实际状态回收部署状态
	helmer.EXPECT().ReleaseStatus(gomock.Any(), gomock.Any()).Return(types.Deploy_StatusDeployed)
	projectRepo.EXPECT().UpdateDeployStatus(gomock.Any(), 7, types.Deploy_StatusDeployed).Return(&biz.Project{ID: 7}, nil)
	// ToAll 在 Validate(376) 与 OnFinally 回调(380) 各触发一次
	sub.EXPECT().ToAll(gomock.Any()).AnyTimes()
	msger.EXPECT().SendDeployedResult(websocket_pb.ResultType_DeployedFailed, "boom", gomock.Any())

	job := &jobRunner{
		logger:       mlog.NewForConfig(nil),
		messager:     msger,
		nsRepo:       nsRepo,
		projRepo:     projectRepo,
		repoRepo:     repoRepo,
		helmer:       helmer,
		user:         &biz.UserInfo{Email: "duc@x.com"},
		deployResult: &deployResult{},
		stopCtx:      context.TODO(),
		input: &JobInput{
			Type:        websocket_pb.Type_CreateProject,
			NamespaceId: 1,
			Name:        "xx",
			RepoID:      12,
			DryRun:      false,
			PubSub:      sub,
		},
	}
	assert.Nil(t, job.Validate().Error())
	assert.True(t, job.isNew)
	assert.Equal(t, "boom", job.SetError(errors.New("boom")).Finish().Error().Error())
}

// 更新路径：UpdateStatusByVersion 成功后注册的 OnError 版本回滚回调，以及
// NeedGitRepo 时的 GetCommit 注入，都由 Finish 报错触发/覆盖。
func Test_jobRunner_Validate_UpdatePath_OnErrorOnFinally(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	msger := NewMockDeployMsger(m)
	repoRepo := data.NewMockRepoRepo(m)
	projectRepo := data.NewMockProjectRepo(m)
	sub := app.NewMockPubSub(m)
	helmer := data.NewMockHelmerRepo(m)
	gits := app.NewMockGitServer(m)
	pluginMgr := app.NewMockPluginManager(m)
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1}, nil)
	repoRepo.EXPECT().Get(gomock.Any(), 12).Return(&biz.Repo{
		MarsConfig:   &mars.Config{},
		NeedGitRepo:  true,
		GitProjectID: 88,
	}, nil)
	projectRepo.EXPECT().FindByName(gomock.Any(), "xx", 1).Return(&biz.Project{ID: 5, Version: 3}, nil)
	projectRepo.EXPECT().UpdateStatusByVersion(gomock.Any(), 100, types.Deploy_StatusDeploying, 10).Return(&biz.Project{ID: 5}, nil)
	pluginMgr.EXPECT().Git().Return(gits).AnyTimes()
	gits.EXPECT().GetCommit("88", "abc123").Return(&biz.Commit{ShortID: "abc123", Title: "t"}, nil)
	// OnError 版本回滚回调
	projectRepo.EXPECT().UpdateVersion(gomock.Any(), 5, 3).Return(&biz.Project{ID: 5}, nil)
	// OnFinally 状态回收回调
	helmer.EXPECT().ReleaseStatus(gomock.Any(), gomock.Any()).Return(types.Deploy_StatusDeploying)
	projectRepo.EXPECT().UpdateDeployStatus(gomock.Any(), 5, types.Deploy_StatusDeploying).Return(&biz.Project{ID: 5}, nil)
	// ToAll 在 Validate(376) 与 OnFinally 回调(380) 各触发一次
	sub.EXPECT().ToAll(gomock.Any()).AnyTimes()
	msger.EXPECT().SendDeployedResult(websocket_pb.ResultType_DeployedFailed, "boom", gomock.Any())

	job := &jobRunner{
		logger:       mlog.NewForConfig(nil),
		messager:     msger,
		nsRepo:       nsRepo,
		projRepo:     projectRepo,
		repoRepo:     repoRepo,
		helmer:       helmer,
		pluginMgr:    pluginMgr,
		user:         &biz.UserInfo{Email: "duc@x.com"},
		deployResult: &deployResult{},
		stopCtx:      context.TODO(),
		input: &JobInput{
			Type:        websocket_pb.Type_CreateProject,
			NamespaceId: 1,
			Name:        "xx",
			RepoID:      12,
			ProjectID:   100,
			Version:     lo.ToPtr(int32(10)),
			GitCommit:   "abc123",
			DryRun:      false,
			PubSub:      sub,
		},
	}
	assert.Nil(t, job.Validate().Error())
	assert.False(t, job.isNew)
	// 覆盖 GetCommit 注入
	assert.Equal(t, "abc123", job.commit.ShortID)
	assert.Equal(t, "boom", job.SetError(errors.New("boom")).Finish().Error().Error())
}

// Run 非 dryRun 成功安装后，UpdateProject 落库失败的错误分支。
func Test_jobRunner_Run_UpdateProjectError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	installer := NewMockReleaseInstaller(m)
	msger := NewMockDeployMsger(m)
	k8sRepo := data.NewMockK8sRepo(m)
	projRepo := data.NewMockProjectRepo(m)
	eventRepo := data.NewMockEventRepo(m)
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ch := newSafeWriteMessageCh(mlog.NewForConfig(nil), 10)

	jb := &jobRunner{
		logger:    mlog.NewForConfig(nil),
		projRepo:  projRepo,
		k8sRepo:   k8sRepo,
		eventRepo: eventRepo,
		messager:  msger,
		installer: installer,
		config:    &mars.Config{},
		chart: &chart.Chart{
			Metadata: &chart.Metadata{},
		},
		ns:           &biz.Namespace{},
		project:      &biz.Project{},
		deployResult: &deployResult{},
		input:        &JobInput{},
		user:         &biz.UserInfo{Name: "duc"},
		commit:       &biz.Commit{},
		messageCh:    ch,
	}

	installer.EXPECT().Run(gomock.Any(), gomock.Any()).Return(&release.Release{
		Config: map[string]any{},
	}, nil)
	k8sRepo.EXPECT().SplitManifests(gomock.Any())
	k8sRepo.EXPECT().GetPodSelectorsByManifest(gomock.Any())
	projRepo.EXPECT().UpdateProject(gomock.Any(), gomock.Any()).Return(nil, errors.New("update failed"))
	msger.EXPECT().To(gomock.Any()).AnyTimes()

	assert.Equal(t, "update failed", jb.Run(context.TODO()).Error().Error())
}

// Run dryRun 路径：跳过落库，act 记为 DryRun。
func Test_jobRunner_Run_DryRun(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	installer := NewMockReleaseInstaller(m)
	msger := NewMockDeployMsger(m)
	k8sRepo := data.NewMockK8sRepo(m)
	projRepo := data.NewMockProjectRepo(m)
	eventRepo := data.NewMockEventRepo(m)
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ch := newSafeWriteMessageCh(mlog.NewForConfig(nil), 10)

	jb := &jobRunner{
		logger:    mlog.NewForConfig(nil),
		projRepo:  projRepo,
		k8sRepo:   k8sRepo,
		eventRepo: eventRepo,
		messager:  msger,
		installer: installer,
		config:    &mars.Config{},
		chart: &chart.Chart{
			Metadata: &chart.Metadata{},
		},
		ns:           &biz.Namespace{},
		project:      &biz.Project{},
		deployResult: &deployResult{},
		input:        &JobInput{},
		user:         &biz.UserInfo{Name: "duc"},
		commit:       &biz.Commit{},
		messageCh:    ch,
		dryRun:       true,
	}

	installer.EXPECT().Run(gomock.Any(), gomock.Any()).Return(&release.Release{
		Config: map[string]any{},
	}, nil)
	k8sRepo.EXPECT().SplitManifests(gomock.Any())
	k8sRepo.EXPECT().GetPodSelectorsByManifest(gomock.Any())
	eventRepo.EXPECT().AuditLogWithChange(
		types.EventActionType_DryRun, "duc", "",
		gomock.Any(), gomock.Any(), gomock.Any())
	msger.EXPECT().To(gomock.Any()).AnyTimes()

	assert.Nil(t, jb.Run(context.TODO()).Error())
}
