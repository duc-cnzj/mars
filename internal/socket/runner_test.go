package socket

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/locker"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"helm.sh/helm/v3/pkg/chart"
)

func TestNewJobManager(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	manager := NewJobManager(
		data.NewMockData(m),
		timer.NewRealTimer(),
		mlog.NewLogger(nil),
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
	timer := timer.NewRealTimer()
	logger := mlog.NewLogger(nil)
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

	ctx, cancelFunc := context.WithCancelCause(context.Background())
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
			expected: &repo.AnyYamlPrettier{
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
		logger:    mlog.NewLogger(nil),
		messageCh: NewSafeWriteMessageCh(mlog.NewLogger(nil), 1),
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	jr.HandleMessage(ctx)
}

func TestHandleMessage_2(t *testing.T) {
	ch := NewSafeWriteMessageCh(mlog.NewLogger(nil), 1)
	jr := &jobRunner{
		logger:    mlog.NewLogger(nil),
		messageCh: ch,
	}
	ch.Close()
	jr.HandleMessage(context.TODO())
}

func TestHandleMessage_3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	ch := NewSafeWriteMessageCh(mlog.NewLogger(nil), 10)
	jr := &jobRunner{
		logger:       mlog.NewLogger(nil),
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
	ch := NewSafeWriteMessageCh(mlog.NewLogger(nil), 10)
	jr := &jobRunner{
		logger:       mlog.NewLogger(nil),
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
	ch := NewSafeWriteMessageCh(mlog.NewLogger(nil), 10)
	jr := &jobRunner{
		logger:       mlog.NewLogger(nil),
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
	up.EXPECT().MkDir(gomock.Any(), false).Return(errors.New("mkdir err")).Times(1)

	err := l.Load(job)
	assert.Equal(t, "mkdir err", err.Error())
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
