package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/duc-cnzj/mars-client/v4/git"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestGitSvc_All(t *testing.T) {
}

func TestGitSvc_BranchOptions(t *testing.T) {
}

func TestGitSvc_Commit(t *testing.T) {
}

func TestGitSvc_CommitOptions(t *testing.T) {
}

func TestGitSvc_DisableProject(t *testing.T) {
}

func TestGitSvc_EnableProject(t *testing.T) {
}

func TestGitSvc_MarsConfigFile(t *testing.T) {
}

func TestGitSvc_PipelineInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	pipe := mock.NewMockPipelineInterface(m)
	instance.SetInstance(app)
	gitS := mockGitServer(m, app)
	gitS.EXPECT().GetCommitPipeline("1", "xxx").Times(1).Return(pipe, nil)
	pipe.EXPECT().GetStatus().Times(1).Return("status")
	pipe.EXPECT().GetWebURL().Times(1).Return("weburl")
	info, _ := new(GitSvc).PipelineInfo(context.TODO(), &git.PipelineInfoRequest{
		GitProjectId: "1",
		Branch:       "dev",
		Commit:       "xxx",
	})
	assert.Equal(t, "status", info.Status)
	assert.Equal(t, "weburl", info.WebUrl)
}

func TestGitSvc_ProjectOptions(t *testing.T) {
}

func mockGitServer(m *gomock.Controller, app *mock.MockApplicationInterface) *mock.MockGitServer {
	gitS := mock.NewMockGitServer(m)
	app.EXPECT().Config().Return(&config.Config{
		GitServerPlugin: config.Plugin{
			Name: "test_git_server",
		},
	}).AnyTimes()
	app.EXPECT().GetPluginByName("test_git_server").Return(gitS).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gitS.EXPECT().Initialize(gomock.All()).AnyTimes()
	return gitS
}
