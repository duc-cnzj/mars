package socket

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
	"time"

	"github.com/duc-cnzj/mars/api/v5/mars"
	"github.com/duc-cnzj/mars/api/v5/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/locker"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/uploader"
	"github.com/duc-cnzj/mars/v5/internal/util/pipeline"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

func TestNewJobManager(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	manager := NewJobManager(
		data.NewMockData(m),
		timer.NewReal(),
		mlog.NewForConfig(nil),
		NewMockReleaseInstaller(m),
		repo.NewMockRepoRepo(m),
		repo.NewMockNamespaceRepo(m),
		repo.NewMockProjectRepo(m),
		repo.NewMockHelmerRepo(m),
		uploader.NewMockUploader(m),
		locker.NewMockLocker(m),
		repo.NewMockK8sRepo(m),
		repo.NewMockEventRepo(m),
		application.NewMockPluginManger(m),
	).(*jobManager)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.data)
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
	assert.NotNil(t, manager.pluginManger)
}

func TestNewJob(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	data := data.NewMockData(m)
	timer := timer.NewReal()
	logger := mlog.NewForConfig(nil)
	releaseInstaller := NewMockReleaseInstaller(m)
	repoRepo := repo.NewMockRepoRepo(m)
	nsRepo := repo.NewMockNamespaceRepo(m)
	projRepo := repo.NewMockProjectRepo(m)
	helmer := repo.NewMockHelmerRepo(m)
	uploader := uploader.NewMockUploader(m)
	locker := locker.NewMockLocker(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	pl := application.NewMockPluginManger(m)
	data.EXPECT().Config().Return(&config.Config{InstallTimeout: 10 * time.Second})

	manager := NewJobManager(
		data,
		timer,
		logger,
		releaseInstaller,
		repoRepo,
		nsRepo,
		projRepo,
		helmer,
		uploader,
		locker,
		k8sRepo,
		eventRepo,
		pl,
	)

	jobInput := &JobInput{
		Type:        websocket_pb.Type_CreateProject,
		NamespaceId: 1,
		Name:        "test",
		RepoID:      1,
		GitBranch:   "main",
		GitCommit:   "abc123",
		Config:      "test-config",
		Atomic:      new(bool),
		ExtraValues: []*websocket_pb.ExtraValue{},
		Version:     new(int32),
		ProjectID:   1,
		User:        &auth.UserInfo{},
		DryRun:      true,
		PubSub:      NewEmptyPubSub(),
		Messager:    nil,
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
	cancelFunc(errCancel)
	assert.True(t, job.(*jobRunner).IsStopped())
	assert.Equal(t, errCancel, job.(*jobRunner).GetStoppedErrorIfHas())
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

func TestCommitMethods(t *testing.T) {
	now := time.Now()
	c := &commit{
		ID:             "12345",
		ShortID:        "123",
		Title:          "Initial commit",
		CommittedDate:  &now,
		AuthorName:     "Author",
		AuthorEmail:    "author@example.com",
		CommitterName:  "Committer",
		CommitterEmail: "committer@example.com",
		CreatedAt:      &now,
		Message:        "This is a commit message",
		ProjectID:      1,
		WebURL:         "http://example.com",
	}

	assert.Equal(t, "12345", c.GetID(), "ID should be '12345'")
	assert.Equal(t, "123", c.GetShortID(), "ShortID should be '123'")
	assert.Equal(t, "Initial commit", c.GetTitle(), "Title should be 'Initial commit'")
	assert.Equal(t, &now, c.GetCommittedDate(), "CommittedDate should be equal to now")
	assert.Equal(t, "Author", c.GetAuthorName(), "AuthorName should be 'Author'")
	assert.Equal(t, "author@example.com", c.GetAuthorEmail(), "AuthorEmail should be 'author@example.com'")
	assert.Equal(t, "Committer", c.GetCommitterName(), "CommitterName should be 'Committer'")
	assert.Equal(t, "committer@example.com", c.GetCommitterEmail(), "CommitterEmail should be 'committer@example.com'")
	assert.Equal(t, &now, c.GetCreatedAt(), "CreatedAt should be equal to now")
	assert.Equal(t, "This is a commit message", c.GetMessage(), "Message should be 'This is a commit message'")
	assert.Equal(t, int64(1), c.GetProjectID(), "ProjectID should be '1'")
	assert.Equal(t, "http://example.com", c.GetWebURL(), "WebURL should be 'http://example.com'")
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
	assert.NoError(t, e.ToOthers(nil), "ToOthers should not return an error")
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
		project  *repo.Project
		expected repo.YamlPrettier
	}{
		{
			name:     "Nil project",
			project:  nil,
			expected: nil,
		},
		{
			name: "Non-nil project",
			project: &repo.Project{
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
			expected: repo.AnyYamlPrettier{
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
			result := toProjectEventYaml(tt.project)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewEmptyCommit(t *testing.T) {
	assert.NotNil(t, NewEmptyCommit())
}

func TestHandleMessage(t *testing.T) {
	jr := &jobRunner{
		logger:    mlog.NewForConfig(nil),
		messageCh: NewSafeWriteMessageCh(mlog.NewForConfig(nil), 1),
	}
	ctx, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	jr.HandleMessage(ctx)
}

func TestHandleMessage_2(t *testing.T) {
	ch := NewSafeWriteMessageCh(mlog.NewForConfig(nil), 1)
	jr := &jobRunner{
		logger:    mlog.NewForConfig(nil),
		messageCh: ch,
	}
	ch.Close()
	jr.HandleMessage(context.TODO())
}

func TestHandleMessage_3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	ch := NewSafeWriteMessageCh(mlog.NewForConfig(nil), 10)
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
	ch := NewSafeWriteMessageCh(mlog.NewForConfig(nil), 10)
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
	ch := NewSafeWriteMessageCh(mlog.NewForConfig(nil), 10)
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
	job := &jobRunner{
		input: &JobInput{
			Config: "xxxx",
		},
		messager: em,
		config: &mars.Config{
			ConfigField: "app->config",
			IsSimpleEnv: true,
		},
	}
	assert.Nil(t, (&UserConfigLoader{}).Load(job))
	assert.Equal(t,
		`app:
  config: xxxx
`, job.userConfigYaml)
	job2 := &jobRunner{
		input: &JobInput{
			Config: "name: duc\nage: 17",
		},
		messager: em,
		config: &mars.Config{
			ConfigField: "app->config",
			IsSimpleEnv: false,
		},
	}
	assert.Nil(t, (&UserConfigLoader{}).Load(job2))
	assert.Equal(t,
		`app:
  config:
    age: 17
    name: duc
`, job2.userConfigYaml)

	job3 := &jobRunner{
		input:    &JobInput{},
		messager: em,
	}
	assert.Nil(t, (&UserConfigLoader{}).Load(job3))
}

func TestElementsLoader_Load(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().To(gomock.Any()).AnyTimes()
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	job := &jobRunner{
		input: &JobInput{
			ExtraValues: []*websocket_pb.ExtraValue{
				{
					Path:  "app->config",
					Value: "1",
				},
			},
		},
		messager: em,
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

	assert.Nil(t, (&ElementsLoader{}).Load(job))
	sort.Strings(job.elementValues)
	assert.Equal(t,
		`app:
  config: "1"
`,
		job.elementValues[0])
	assert.Equal(t,
		`app:
  xxx: true
`,
		job.elementValues[1])

	err := (&ElementsLoader{}).Load(&jobRunner{
		input: &JobInput{
			ExtraValues: []*websocket_pb.ExtraValue{
				{
					Path:  "app->config",
					Value: "4",
				},
			},
		},
		messager: em,
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
		messager: em,
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
	err = (&ElementsLoader{}).Load(j)
	assert.Nil(t, err)
	assert.Equal(t, []string{"duc: xxx\n"}, j.elementValues)

	j2 := &jobRunner{
		input:    &JobInput{},
		messager: em,
		config:   &mars.Config{},
	}
	assert.Nil(t, (&ElementsLoader{}).Load(j2))
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
	job := &jobRunner{project: &repo.Project{}}
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
	job := &jobRunner{
		userConfigYaml:   dcy,
		imagePullSecrets: []string{"secret"},
		timer:            timer.NewReal(),
		elementValues:    []string{ev1, ev2},
		valuesOptions:    &values.Options{},
		input:            &JobInput{GitBranch: "dev"},
		messager:         msger,
		systemValuesYaml: vy,
		uploader:         up,
	}
	assert.Nil(t, (&MergeValuesLoader{}).Load(job))
	assert.Equal(t, "/app/config.yaml", job.valuesOptions.ValueFiles[0])

	job2 := &jobRunner{
		systemValuesYaml: "",
		imagePullSecrets: nil,
		userConfigYaml:   "",
		valuesOptions:    &values.Options{},
		input:            &JobInput{GitBranch: "dev"},
		messager:         msger,
	}
	assert.Nil(t, (&MergeValuesLoader{}).Load(job2))
}

func TestSystemVariableLoader_Load_ok1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	msger.EXPECT().To(gomock.Any()).Times(1)
	msger.EXPECT().SendMsg(gomock.Any()).Times(2)
	assert.Nil(t, (&SystemVariableLoader{}).Load(&jobRunner{
		config:   &mars.Config{ValuesYaml: ""},
		messager: msger,
	}))
}
func TestSystemVariableLoader_Load_ok(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	gitS := application.NewMockGitServer(m)
	mockPipeline := application.NewMockPipeline(m)
	mockPipeline.EXPECT().GetID().Return(int64(0))
	mockPipeline.EXPECT().GetRef().Return("dev")
	gitS.EXPECT().GetCommitPipeline("10", "dev", "c").Return(mockPipeline, nil)

	projectRepo := repo.NewMockProjectRepo(m)
	projectRepo.EXPECT().GetPreOccupiedLenByValuesYaml(gomock.Any()).Return(1)
	pl := application.NewMockPluginManger(m)
	pl.EXPECT().Git().Return(gitS)
	domain := application.NewMockDomainManager(m)
	pl.EXPECT().Domain().Return(domain).AnyTimes()
	domain.EXPECT().GetDomainByIndex(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	domain.EXPECT().GetCertSecretName("app", gomock.Any()).AnyTimes()
	domain.EXPECT().GetClusterIssuer().Return("cluster-issuer")
	em := NewMockDeployMsger(m)
	em.EXPECT().To(gomock.Any()).AnyTimes()
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	commit := application.NewMockCommit(m)
	commit.EXPECT().GetShortID().Return("short_id").AnyTimes()
	job := &jobRunner{
		commit:    commit,
		pluginMgr: pl,
		projRepo:  projectRepo,
		config: &mars.Config{
			ValuesYaml: `
VarImagePullSecrets: <.ImagePullSecrets>
`,
		},
		input: &JobInput{
			GitBranch: "dev",
			GitCommit: "c",
		},
		logger: mlog.NewForConfig(nil),
		project: &repo.Project{
			Name: "app",
		},
		ns:               &repo.Namespace{Name: "ns"},
		imagePullSecrets: []string{"a", "b", "c"},
		messager:         em,
		repo:             &repo.Repo{NeedGitRepo: true, GitProjectID: 10},
	}
	assert.Nil(t, (&SystemVariableLoader{}).Load(job))
	assert.Equal(t, `
VarImagePullSecrets: [{name: a}, {name: b}, {name: c}, ]
`,
		job.systemValuesYaml)
	assert.Equal(t, "dev", job.vars[VarBranch])
	assert.Equal(t, "short_id", job.vars[VarCommit])
	assert.Equal(t, "0", job.vars[VarPipeline])
	assert.Equal(t, "[{name: a}, {name: b}, {name: c}, ]", job.vars[VarImagePullSecrets])
}

func TestSystemVariableLoader_Load_fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	gitS := application.NewMockGitServer(m)
	gitS.EXPECT().GetCommitPipeline(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("x"))

	projectRepo := repo.NewMockProjectRepo(m)
	projectRepo.EXPECT().GetPreOccupiedLenByValuesYaml(gomock.Any()).Return(1)
	pl := application.NewMockPluginManger(m)
	pl.EXPECT().Git().Return(gitS)
	domain := application.NewMockDomainManager(m)
	pl.EXPECT().Domain().Return(domain).AnyTimes()
	domain.EXPECT().GetDomainByIndex(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	domain.EXPECT().GetCertSecretName(gomock.Any(), gomock.Any()).AnyTimes()
	em := NewMockDeployMsger(m)
	em.EXPECT().To(gomock.Any()).AnyTimes()
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	commit := application.NewMockCommit(m)
	commit.EXPECT().GetShortID().Return("short_id").AnyTimes()
	job := &jobRunner{
		commit:    commit,
		pluginMgr: pl,
		projRepo:  projectRepo,
		input: &JobInput{
			GitBranch: "dev",
		},
		config: &mars.Config{
			ValuesYaml: `
VarImagePullSecrets: <.ImagePullSecrets>
image: <.Pipeline>-<.Branch>
`,
		},
		logger: mlog.NewForConfig(nil),
		project: &repo.Project{
			Name: "app",
		},
		ns:               &repo.Namespace{Name: "ns"},
		imagePullSecrets: []string{"a", "b", "c"},
		messager:         em,
		repo:             &repo.Repo{NeedGitRepo: true, GitProjectID: 1},
	}
	assert.Error(t, (&SystemVariableLoader{}).Load(job))
}

func TestChartFileLoader_Load2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	h := repo.NewMockHelmerRepo(m)
	pl := application.NewMockPluginManger(m)
	gits := application.NewMockGitServer(m)
	up := uploader.NewMockUploader(m)
	pl.EXPECT().Git().Return(gits).AnyTimes()
	job := &jobRunner{
		uploader:  up,
		helmer:    h,
		pluginMgr: pl,
		input: &JobInput{
			GitCommit: "commit",
		},
		messager: em,
		config: &mars.Config{
			LocalChartPath: "100|main|dir",
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

	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	em.EXPECT().To(gomock.Any())
	gits.EXPECT().GetDirectoryFilesWithBranch("100", "main", "dir", true).Return([]string{"file1", "file2"}, nil)
	up.EXPECT().LocalUploader().Return(up)
	up.EXPECT().MkDir(gomock.Any(), true).Return(errors.New("mkdir err")).Times(1)

	err := l.Load(job)
	assert.Equal(t, "mkdir err", err.Error())
}

func TestChartFileLoader_Load(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	em.EXPECT().To(gomock.Any()).AnyTimes()
	h := repo.NewMockHelmerRepo(m)
	gits := application.NewMockGitServer(m)
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

	pl := application.NewMockPluginManger(m)
	pl.EXPECT().Git().Return(gits).AnyTimes()
	job := &jobRunner{
		uploader: up,
		logger:   mlog.NewForConfig(nil),
		helmer:   h,
		input:    &JobInput{},
		messager: em,
		config: &mars.Config{
			LocalChartPath: "9999|master|dir",
		},
		pluginMgr: pl,
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

	err := l.Load(job)
	assert.Equal(t, "charts 文件不存在", err.Error())
	err = l.Load(job)
	assert.Len(t, job.finallyCallback.Sort(), 2)
	assert.Nil(t, err)

	up.EXPECT().DeleteDir("/dir").Times(2)
	called := 0

	pipeline.New[error]().Send(job.Error()).Through(job.finallyCallback.Sort()...).Then(func(err error) {
		called++
	})
	assert.Equal(t, 1, called)
}

func TestChartFileLoader_LoadWithChartMissing(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	em.EXPECT().To(gomock.Any()).AnyTimes()
	h := repo.NewMockHelmerRepo(m)
	pl := application.NewMockPluginManger(m)
	gits := application.NewMockGitServer(m)
	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return([]string{"file1", "file2"}, nil)

	gits.EXPECT().GetFileContentWithSha("9999", "master", "file1").Return("file1", nil).Times(1)
	gits.EXPECT().GetFileContentWithSha("9999", "master", "file2").Return("file2", nil).Times(1)
	up := uploader.NewMockUploader(m)
	up.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().AbsolutePath(gomock.Any()).Return("/dir")
	up.EXPECT().MkDir(gomock.Any(), true).Times(1)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Times(2)
	pl.EXPECT().Git().Return(gits).AnyTimes()
	job := &jobRunner{
		helmer:    h,
		pluginMgr: pl,
		uploader:  up,
		logger:    mlog.NewForConfig(nil),
		input:     &JobInput{},
		messager:  em,
		config: &mars.Config{
			LocalChartPath: "9999|master|dir",
		},
	}
	loadDirErr := errors.New("Chart.yaml file is missing")
	l := &ChartFileLoader{
		chartLoader: &fakeChartLoader{
			loadDirErr: loadDirErr,
		},
	}

	err := l.Load(job)
	assert.Equal(t, loadDirErr.Error(), err.Error())

	job.config.LocalChartPath = "xxx"
	err = l.Load(job)
	assert.Equal(t, "LocalChartPath 格式不正确", err.Error())
}

func Test_jobRunner_WriteConfigYamlToTmpFile(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockUploader.EXPECT().LocalUploader().Return(mockUploader).Times(2)
	job := &jobRunner{
		timer:    timer.NewReal(),
		uploader: mockUploader,
		logger:   mlog.NewForConfig(nil),
	}
	info := uploader.NewMockFileInfo(m)
	mockUploader.EXPECT().Put(gomock.Not(nil), bytes.NewReader([]byte("xxx"))).Return(info, nil)
	info.EXPECT().Path().Return("/path")
	path, closer, err := job.WriteConfigYamlToTmpFile([]byte("xxx"))
	assert.Equal(t, "/path", path)
	assert.Nil(t, err)
	mockUploader.EXPECT().Delete("/path").Return(errors.New("x"))
	err = closer.Close()
	assert.Equal(t, "x", err.Error())
}

func Test_jobRunner_Validate_Fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	msger := NewMockDeployMsger(m)
	repoRepo := repo.NewMockRepoRepo(m)
	projectRepo := repo.NewMockProjectRepo(m)

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

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Namespace{}, nil)
	repoRepo.EXPECT().Show(gomock.Any(), 12).Return(nil, errors.New("xxx"))
	job = &jobRunner{
		messager: msger,
		nsRepo:   nsRepo,
		repoRepo: repoRepo,
		input:    &JobInput{Type: websocket_pb.Type_CreateProject, NamespaceId: 1, RepoID: 12},
	}
	assert.Error(t, job.Validate().Error())

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Namespace{ID: 1}, nil)
	repoRepo.EXPECT().Show(gomock.Any(), 12).Return(&repo.Repo{
		MarsConfig: &mars.Config{},
	}, nil)
	projectRepo.EXPECT().FindByName(gomock.Any(), "xx", 1).Return(nil, errors.New("xxx"))
	projectRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("xxxa"))
	job = &jobRunner{
		logger:   mlog.NewForConfig(nil),
		messager: msger,
		nsRepo:   nsRepo,
		projRepo: projectRepo,
		repoRepo: repoRepo,
		user:     &auth.UserInfo{},
		input: &JobInput{
			Type:        websocket_pb.Type_CreateProject,
			NamespaceId: 1,
			Name:        "xx",
			RepoID:      12,
			DryRun:      false,
		},
	}
	assert.Error(t, job.Validate().Error())

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Namespace{ID: 1}, nil)
	repoRepo.EXPECT().Show(gomock.Any(), 12).Return(&repo.Repo{
		MarsConfig: &mars.Config{},
	}, nil)
	projectRepo.EXPECT().FindByName(gomock.Any(), "xx", 1).Return(&repo.Project{}, nil)
	projectRepo.EXPECT().UpdateStatusByVersion(gomock.Any(), 1, types.Deploy_StatusDeploying, gomock.Any()).Return(nil, errors.New("xxx"))
	job = &jobRunner{
		logger:   mlog.NewForConfig(nil),
		messager: msger,
		nsRepo:   nsRepo,
		projRepo: projectRepo,
		repoRepo: repoRepo,
		user:     &auth.UserInfo{},
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

func Test_jobRunner_Validate_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	msger := NewMockDeployMsger(m)
	repoRepo := repo.NewMockRepoRepo(m)
	projectRepo := repo.NewMockProjectRepo(m)
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()

	var job *jobRunner

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Namespace{ID: 1}, nil)
	repoRepo.EXPECT().Show(gomock.Any(), 12).Return(&repo.Repo{
		MarsConfig:  &mars.Config{},
		NeedGitRepo: false,
	}, nil)
	projectRepo.EXPECT().FindByName(gomock.Any(), "xx", 1).Return(&repo.Project{}, nil)
	projectRepo.EXPECT().UpdateStatusByVersion(gomock.Any(), 100, types.Deploy_StatusDeploying, 10).Return(&repo.Project{}, nil)
	sub := application.NewMockPubSub(m)
	sub.EXPECT().ToAll(gomock.Any())
	job = &jobRunner{
		logger:   mlog.NewForConfig(nil),
		messager: msger,
		nsRepo:   nsRepo,
		projRepo: projectRepo,
		repoRepo: repoRepo,
		user:     &auth.UserInfo{},
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
	k8sRepo := repo.NewMockK8sRepo(m)
	projRepo := repo.NewMockProjectRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
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
		ns:           &repo.Namespace{},
		project:      &repo.Project{},
		deployResult: &deployResult{},
		input:        &JobInput{},
		commit:       NewEmptyCommit(),
		messageCh:    messageChan,
	}
	assert.Error(t, jb.Run(context.TODO()).Error())
}

func Test_jobRunner_Run_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	installer := NewMockReleaseInstaller(m)
	msger := NewMockDeployMsger(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	projRepo := repo.NewMockProjectRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ch := NewSafeWriteMessageCh(mlog.NewForConfig(nil), 10)

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
		ns:           &repo.Namespace{},
		project:      &repo.Project{},
		deployResult: &deployResult{},
		input:        &JobInput{},
		user:         &auth.UserInfo{Name: "duc"},
		commit:       NewEmptyCommit(),
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
		Description:  jb.commit.GetTitle(),
		messageChan:  jb.messageCh,
		percenter:    jb.messager,
	}).Return(&release.Release{
		Config: map[string]any{},
	}, nil)
	projRepo.EXPECT().UpdateProject(gomock.Any(), gomock.Any()).Return(&repo.Project{}, nil)
	eventRepo.EXPECT().Dispatch(repo.EventProjectChanged, gomock.Any())
	eventRepo.EXPECT().AuditLogWithChange(
		types.EventActionType_Update, "duc",
		gomock.Any(), gomock.Any(),
		gomock.Any())
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	k8sRepo.EXPECT().SplitManifests(gomock.Any())
	k8sRepo.EXPECT().GetPodSelectorsByManifest(gomock.Any())

	assert.Nil(t, jb.Run(context.TODO()).Error())
}
