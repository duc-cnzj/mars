package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v4/git"
	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewGitSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockGitRepo(m),
		cache.NewMockCache(m),
	)

	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*gitSvc).eventRepo)
	assert.NotNil(t, svc.(*gitSvc).logger)
	assert.NotNil(t, svc.(*gitSvc).gitRepo)
	assert.NotNil(t, svc.(*gitSvc).cache)
	assert.NotNil(t, svc.(*gitSvc).repoRepo)
}

func Test_gitSvc_AllRepos(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)

	gitRepo.EXPECT().AllProjects(gomock.Any(), false).Return(nil, errors.New("error"))
	repos, err := svc.AllRepos(context.TODO(), nil)
	assert.Nil(t, repos)
	assert.NotNil(t, err)
}

func Test_gitSvc_AllRepos_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)

	gitRepo.EXPECT().AllProjects(gomock.Any(), false).Return([]*repo.GitProject{
		{ID: 1, Name: "a", Description: "aa"},
		nil,
	}, nil)
	repos, err := svc.AllRepos(context.TODO(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, repos)
	assert.Equal(t, 2, len(repos.Items))
	assert.Equal(t, int32(1), repos.Items[0].Id)
	assert.Equal(t, "a", repos.Items[0].Name)
	assert.Equal(t, "aa", repos.Items[0].Description)
}

func Test_gitSvc_ProjectOptions(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewGitSvc(
		repoRepo,
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockGitRepo(m),
		cache.NewMockCache(m),
	)

	repoRepo.EXPECT().All(gomock.Any(), &repo.AllRepoRequest{Enabled: lo.ToPtr(true)}).Return(nil, errors.New("error"))
	options, err := svc.ProjectOptions(context.TODO(), nil)
	assert.Nil(t, options)
	assert.NotNil(t, err)
}

func Test_gitSvc_ProjectOptions_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewGitSvc(
		repoRepo,
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockGitRepo(m),
		cache.NewMockCache(m),
	)

	repoRepo.EXPECT().All(gomock.Any(), &repo.AllRepoRequest{Enabled: lo.ToPtr(true)}).Return([]*repo.Repo{
		{
			ID:           1,
			Name:         "a",
			GitProjectID: 11,
			NeedGitRepo:  true,
			Description:  "desc",
		},
	}, nil)
	options, err := svc.ProjectOptions(context.TODO(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, options)
	assert.Equal(t, 1, len(options.Items))
	assert.Equal(t, "1", options.Items[0].Value)
	assert.Equal(t, "a", options.Items[0].Label)
	assert.Equal(t, OptionTypeProject, options.Items[0].Type)
	assert.Equal(t, false, options.Items[0].IsLeaf)
	assert.Equal(t, int32(11), options.Items[0].GitProjectId)
	assert.Equal(t, true, options.Items[0].NeedGitRepo)
	assert.Equal(t, "desc", options.Items[0].Description)
}

func Test_gitSvc_BranchOptions(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)
	gitRepo.EXPECT().AllBranches(gomock.Any(), 1, false).Return(nil, errors.New("error"))
	options, err := svc.BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: 1,
		RepoId:       1,
	})
	assert.Error(t, err)
	assert.Nil(t, options)
}

func Test_gitSvc_BranchOptions_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewGitSvc(
		repoRepo,
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)
	gitRepo.EXPECT().AllBranches(gomock.Any(), 1, false).Return([]*repo.Branch{
		{
			Name:      "br",
			IsDefault: true,
			WebURL:    "xxx",
		},
		{
			Name:      "ccc",
			IsDefault: true,
			WebURL:    "xxx",
		},
	}, nil)

	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Repo{
		MarsConfig: &mars.Config{Branches: []string{"ccc"}},
	}, nil)
	options, err := svc.BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: 1,
		RepoId:       1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, options)
	assert.Equal(t, 1, len(options.Items))
}

func Test_gitSvc_BranchOptions_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewGitSvc(
		repoRepo,
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)
	gitRepo.EXPECT().AllBranches(gomock.Any(), 1, false).Return([]*repo.Branch{
		{
			Name:      "br",
			IsDefault: true,
			WebURL:    "xxx",
		},
		{
			Name:      "ccc",
			IsDefault: true,
			WebURL:    "xxx",
		},
	}, nil)

	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))
	_, err := svc.BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: 1,
		RepoId:       1,
	})
	assert.Equal(t, "error", err.Error())
}

func Test_gitSvc_Commit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)
	gitRepo.EXPECT().GetCommit(gomock.Any(), 1, "commit").Return(nil, errors.New("error"))
	commit, err := svc.Commit(context.TODO(), &git.CommitRequest{
		GitProjectId: 1,
		Branch:       "branch",
		Commit:       "commit",
	})
	assert.Error(t, err)
	assert.Nil(t, commit)
}

func Test_gitSvc_Commit_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)
	gitRepo.EXPECT().GetCommit(gomock.Any(), 1, "commit").Return(&repo.Commit{}, nil)
	commit, err := svc.Commit(context.TODO(), &git.CommitRequest{
		GitProjectId: 1,
		Branch:       "branch",
		Commit:       "commit",
	})
	assert.Nil(t, err)
	assert.NotNil(t, commit)
}

func Test_gitSvc_CommitOptions(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)
	gitRepo.EXPECT().ListCommits(gomock.Any(), 1, "xxx").Return(nil, errors.New("error"))
	options, err := svc.CommitOptions(context.TODO(), &git.CommitOptionsRequest{
		GitProjectId: 1,
		Branch:       "xxx",
	})
	assert.Nil(t, options)
	assert.NotNil(t, err)
}

func Test_gitSvc_CommitOptions_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)
	gitRepo.EXPECT().ListCommits(gomock.Any(), 1, "xxx").Return([]*repo.Commit{
		{
			ID:         "x",
			ShortID:    "aaa",
			AuthorName: "aaaa",
		},
	}, nil)
	options, err := svc.CommitOptions(context.TODO(), &git.CommitOptionsRequest{
		GitProjectId: 1,
		Branch:       "xxx",
	})
	assert.NotNil(t, options)
	assert.Nil(t, err)
}

func Test_gitSvc_GetChartValuesYaml(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)

	resp, err := svc.GetChartValuesYaml(context.TODO(), &git.GetChartValuesYamlRequest{
		Input: "",
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func Test_gitSvc_GetChartValuesYaml_error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)

	gitRepo.EXPECT().GetFileContentWithBranch(gomock.Any(), 1, "branch", "path/values.yaml").Return("", errors.New("x"))
	resp, err := svc.GetChartValuesYaml(context.TODO(), &git.GetChartValuesYamlRequest{
		Input: "1|branch|path",
	})
	assert.Nil(t, resp)
	assert.Equal(t, "x", err.Error())
}

func Test_gitSvc_GetChartValuesYaml_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)

	gitRepo.EXPECT().GetFileContentWithBranch(gomock.Any(), 1, "branch", "path/values.yaml").Return("content", nil)
	resp, err := svc.GetChartValuesYaml(context.TODO(), &git.GetChartValuesYamlRequest{
		Input: "1|branch|path",
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "content", resp.Values)
}

func Test_gitSvc_PipelineInfo_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)

	gitRepo.EXPECT().GetCommitPipeline(gomock.Any(), 1, "main", "commit").Return(&repo.Pipeline{
		Status: "success",
		WebURL: "https://example.com",
	}, nil)

	res, err := svc.PipelineInfo(context.TODO(), &git.PipelineInfoRequest{
		GitProjectId: "1",
		Branch:       "main",
		Commit:       "commit",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "success", res.Status)
	assert.Equal(t, "https://example.com", res.WebUrl)
}

func Test_gitSvc_PipelineInfo_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := repo.NewMockGitRepo(m)
	svc := NewGitSvc(
		repo.NewMockRepoRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		gitRepo,
		cache.NewMockCache(m),
	)

	gitRepo.EXPECT().GetCommitPipeline(gomock.Any(), 1, "main", "commit").Return(nil, errors.New("error"))

	res, err := svc.PipelineInfo(context.TODO(), &git.PipelineInfoRequest{
		GitProjectId: "1",
		Branch:       "main",
		Commit:       "commit",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}
