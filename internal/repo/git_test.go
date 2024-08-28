package repo

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/config"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/cache"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGitRepo_AllProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewLogger(nil)
	mockCache := cache.NewMockCache(m)
	mockPluginManager := application.NewMockPluginManger(m)
	mockData := data.NewMockData(m)
	mockData.EXPECT().Config().Return(&config.Config{GitServerCached: false})

	repo := NewGitRepo(mockLogger, mockCache, mockPluginManager, mockData)

	git := application.NewMockGitServer(m)
	mockPluginManager.EXPECT().Git().Return(git)
	mockProject := application.NewMockProject(m)
	mockProject.EXPECT().GetID()
	mockProject.EXPECT().GetName()
	mockProject.EXPECT().GetDefaultBranch()
	mockProject.EXPECT().GetWebURL()
	mockProject.EXPECT().GetPath()
	mockProject.EXPECT().GetAvatarURL()
	mockProject.EXPECT().GetDescription()
	git.EXPECT().AllProjects().Return([]application.Project{
		mockProject,
	}, nil)
	projects, err := repo.AllProjects(context.Background(), true)

	assert.Nil(t, err)
	assert.NotNil(t, projects)
}

func TestGitRepo_AllProjects_Cache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewLogger(nil)
	mockPluginManager := application.NewMockPluginManger(m)
	mockData := data.NewMockData(m)
	mockData.EXPECT().Config().Return(&config.Config{GitServerCached: true})

	repo := NewGitRepo(mockLogger, &cache.NoCache{}, mockPluginManager, mockData)

	git := application.NewMockGitServer(m)
	mockPluginManager.EXPECT().Git().Return(git)
	mockProject := application.NewMockProject(m)
	mockProject.EXPECT().GetID()
	mockProject.EXPECT().GetName()
	mockProject.EXPECT().GetDefaultBranch()
	mockProject.EXPECT().GetWebURL()
	mockProject.EXPECT().GetPath()
	mockProject.EXPECT().GetAvatarURL()
	mockProject.EXPECT().GetDescription()
	git.EXPECT().AllProjects().Return([]application.Project{
		mockProject,
	}, nil)
	projects, err := repo.AllProjects(context.Background(), true)

	assert.Nil(t, err)
	assert.NotNil(t, projects)
}

func TestGitRepo_AllBranches(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewLogger(nil)
	mockCache := cache.NewMockCache(m)
	mockPluginManager := application.NewMockPluginManger(m)
	mockData := data.NewMockData(m)
	mockData.EXPECT().Config().Return(&config.Config{GitServerCached: false})

	git := application.NewMockGitServer(m)
	mockPluginManager.EXPECT().Git().Return(git)
	branch := application.NewMockBranch(m)
	branch.EXPECT().GetName()
	branch.EXPECT().IsDefault()
	branch.EXPECT().GetWebURL()
	git.EXPECT().AllBranches("1").Return([]application.Branch{branch}, nil)
	repo := NewGitRepo(mockLogger, mockCache, mockPluginManager, mockData)

	branches, err := repo.AllBranches(context.Background(), 1, true)

	assert.Nil(t, err)
	assert.NotNil(t, branches)
}

func TestGitRepo_AllBranches_Cache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewLogger(nil)
	mockPluginManager := application.NewMockPluginManger(m)
	mockData := data.NewMockData(m)
	mockData.EXPECT().Config().Return(&config.Config{GitServerCached: true})

	git := application.NewMockGitServer(m)
	mockPluginManager.EXPECT().Git().Return(git)
	branch := application.NewMockBranch(m)
	branch.EXPECT().GetName()
	branch.EXPECT().IsDefault()
	branch.EXPECT().GetWebURL()
	git.EXPECT().AllBranches("1").Return([]application.Branch{branch}, nil)
	repo := NewGitRepo(mockLogger, &cache.NoCache{}, mockPluginManager, mockData)

	branches, err := repo.AllBranches(context.Background(), 1, true)

	assert.Nil(t, err)
	assert.NotNil(t, branches)
}

func TestGitRepo_ListCommits(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewLogger(nil)
	mockCache := cache.NewMockCache(m)
	mockPluginManager := application.NewMockPluginManger(m)
	mockData := data.NewMockData(m)

	git := application.NewMockGitServer(m)
	mockPluginManager.EXPECT().Git().Return(git)
	git.EXPECT().ListCommits("1", "main").Return([]application.Commit{}, nil)
	repo := NewGitRepo(mockLogger, mockCache, mockPluginManager, mockData)

	commits, err := repo.ListCommits(context.Background(), 1, "main")

	assert.Nil(t, err)
	assert.NotNil(t, commits)
}

func TestGitRepo_GetProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewLogger(nil)
	mockCache := cache.NewMockCache(m)
	mockPluginManager := application.NewMockPluginManger(m)
	mockData := data.NewMockData(m)
	git := application.NewMockGitServer(m)
	mockPluginManager.EXPECT().Git().Return(git)
	mockProject := application.NewMockProject(m)
	git.EXPECT().GetProject("1").Return(mockProject, nil)
	mockProject.EXPECT().GetID()
	mockProject.EXPECT().GetName()
	mockProject.EXPECT().GetDefaultBranch()
	mockProject.EXPECT().GetWebURL()
	mockProject.EXPECT().GetPath()
	mockProject.EXPECT().GetAvatarURL()
	mockProject.EXPECT().GetDescription()
	repo := NewGitRepo(mockLogger, mockCache, mockPluginManager, mockData)

	project, err := repo.GetProject(context.Background(), 1)

	assert.Nil(t, err)
	assert.NotNil(t, project)
}

func TestGitRepo_GetFileContentWithBranch(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewLogger(nil)
	mockCache := cache.NewMockCache(m)
	mockPluginManager := application.NewMockPluginManger(m)
	mockData := data.NewMockData(m)

	git := application.NewMockGitServer(m)
	mockPluginManager.EXPECT().Git().Return(git)
	git.EXPECT().GetFileContentWithBranch("1", "main", "README.md").Return("aa", nil)

	repo := NewGitRepo(mockLogger, mockCache, mockPluginManager, mockData)

	content, err := repo.GetFileContentWithBranch(context.Background(), 1, "main", "README.md")

	assert.Nil(t, err)
	assert.Equal(t, "aa", content)
}

func TestGitRepo_GetCommit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewLogger(nil)
	mockCache := cache.NewMockCache(m)
	mockPluginManager := application.NewMockPluginManger(m)
	mockData := data.NewMockData(m)
	git := application.NewMockGitServer(m)
	mockPluginManager.EXPECT().Git().Return(git)
	mockCommit := application.NewMockCommit(m)
	mockCommit.EXPECT().GetID()
	mockCommit.EXPECT().GetShortID()
	mockCommit.EXPECT().GetAuthorName()
	mockCommit.EXPECT().GetAuthorEmail()
	mockCommit.EXPECT().GetCommitterName()
	mockCommit.EXPECT().GetCommitterEmail()
	mockCommit.EXPECT().GetMessage()
	mockCommit.EXPECT().GetTitle()
	mockCommit.EXPECT().GetWebURL()
	mockCommit.EXPECT().GetCreatedAt()
	mockCommit.EXPECT().GetCommittedDate()
	git.EXPECT().GetCommit("1", "abc123").Return(mockCommit, nil)

	repo := NewGitRepo(mockLogger, mockCache, mockPluginManager, mockData)

	commit, err := repo.GetCommit(context.Background(), 1, "abc123")

	assert.Nil(t, err)
	assert.NotNil(t, commit)
}

func TestGitRepo_GetCommitPipeline(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewLogger(nil)
	mockCache := cache.NewMockCache(m)
	mockPluginManager := application.NewMockPluginManger(m)
	mockData := data.NewMockData(m)

	git := application.NewMockGitServer(m)
	mockPluginManager.EXPECT().Git().Return(git)

	mockPipeline := application.NewMockPipeline(m)
	git.EXPECT().GetCommitPipeline("1", "main", "abc123").Return(mockPipeline, nil)
	mockPipeline.EXPECT().GetStatus()
	mockPipeline.EXPECT().GetWebURL()
	repo := NewGitRepo(mockLogger, mockCache, mockPluginManager, mockData)

	pipeline, err := repo.GetCommitPipeline(context.Background(), 1, "main", "abc123")

	assert.Nil(t, err)
	assert.NotNil(t, pipeline)
}

func TestGitRepo_GetByProjectID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewLogger(nil)
	mockCache := cache.NewMockCache(m)
	mockPluginManager := application.NewMockPluginManger(m)
	mockData := data.NewMockData(m)

	git := application.NewMockGitServer(m)
	mockPluginManager.EXPECT().Git().Return(git)
	mockProject := application.NewMockProject(m)
	git.EXPECT().GetProject("1").Return(mockProject, nil)
	mockProject.EXPECT().GetID()
	mockProject.EXPECT().GetName()
	mockProject.EXPECT().GetDefaultBranch()
	mockProject.EXPECT().GetWebURL()
	mockProject.EXPECT().GetPath()
	mockProject.EXPECT().GetAvatarURL()
	mockProject.EXPECT().GetDescription()
	repo := NewGitRepo(mockLogger, mockCache, mockPluginManager, mockData)

	project, err := repo.GetByProjectID(context.Background(), 1)

	assert.Nil(t, err)
	assert.NotNil(t, project)
}
