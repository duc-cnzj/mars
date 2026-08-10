package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/git"
	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewGitSvc(t *testing.T) {
	svc, _ := newGitSvcWithMocks(t)

	assert.NotNil(t, svc)
	assert.NotNil(t, svc.logger)
	assert.NotNil(t, svc.gitBiz)
	assert.NotNil(t, svc.repoBiz)
}

func Test_gitSvc_AllRepos(t *testing.T) {
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo

	gitRepo.EXPECT().AllProjects(gomock.Any(), false).Return(nil, errors.New("error"))
	repos, err := svc.AllRepos(context.TODO(), nil)
	assert.Nil(t, repos)
	assert.NotNil(t, err)
}

func Test_gitSvc_AllRepos_Success(t *testing.T) {
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo

	gitRepo.EXPECT().AllProjects(gomock.Any(), false).Return([]*biz.GitProject{
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
	svc, mocks := newGitSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{Enabled: lo.ToPtr(true)}).Return(nil, errors.New("error"))
	options, err := svc.ProjectOptions(context.TODO(), nil)
	assert.Nil(t, options)
	assert.NotNil(t, err)
}

func Test_gitSvc_ProjectOptions_Success(t *testing.T) {
	svc, mocks := newGitSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{Enabled: lo.ToPtr(true)}).Return([]*biz.Repo{
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
	assert.Equal(t, optionTypeProject, options.Items[0].Type)
	assert.Equal(t, false, options.Items[0].IsLeaf)
	assert.Equal(t, int32(11), options.Items[0].GitProjectId)
	assert.Equal(t, true, options.Items[0].NeedGitRepo)
	assert.Equal(t, "desc", options.Items[0].Description)
}

func Test_gitSvc_BranchOptions(t *testing.T) {
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo
	gitRepo.EXPECT().AllBranches(gomock.Any(), 1, false).Return(nil, errors.New("error"))
	options, err := svc.BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: 1,
		RepoId:       1,
	})
	assert.Error(t, err)
	assert.Nil(t, options)
}

func Test_gitSvc_BranchOptions_Success(t *testing.T) {
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo
	repoRepo := mocks.repoRepo
	gitRepo.EXPECT().AllBranches(gomock.Any(), 1, false).Return([]*biz.Branch{
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

	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{
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
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo
	repoRepo := mocks.repoRepo
	gitRepo.EXPECT().AllBranches(gomock.Any(), 1, false).Return([]*biz.Branch{
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

	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(nil, errors.New("error"))
	_, err := svc.BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: 1,
		RepoId:       1,
	})
	assert.Equal(t, "error", err.Error())
}

func Test_gitSvc_BranchOptions_NoRepo(t *testing.T) {
	// 回归防护：RepoId=0（表单未选仓库）时跳过分支白名单过滤，直接返回全部分支。
	// 改坏实现（>0 写成 >=0 或去掉守卫）会误调用 repoRepo.Get(0)，
	// 下方未设置该期望，gomock 遇到意外调用必然 FAIL。
	svc, mocks := newGitSvcWithMocks(t)
	mocks.gitRepo.EXPECT().AllBranches(gomock.Any(), 1, false).Return([]*biz.Branch{
		{Name: "br", IsDefault: true, WebURL: "xxx"},
		{Name: "ccc", IsDefault: true, WebURL: "xxx"},
	}, nil)

	options, err := svc.BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: 1,
		RepoId:       0,
	})
	assert.Nil(t, err)
	if assert.NotNil(t, options) {
		assert.Equal(t, 2, len(options.Items), "未选仓库时应返回全部分支，不做白名单过滤")
	}
}

func Test_gitSvc_Commit(t *testing.T) {
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo
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
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo
	gitRepo.EXPECT().GetCommit(gomock.Any(), 1, "commit").Return(&biz.Commit{
		ID:             "abc123",
		ShortID:        "abc",
		Title:          "fix bug",
		AuthorName:     "duc",
		AuthorEmail:    "duc@example.com",
		CommitterName:  "duc",
		CommitterEmail: "duc@example.com",
		WebURL:         "http://git/abc123",
		Message:        "fix the bug",
		CommittedDate:  &time.Time{},
		CreatedAt:      &time.Time{},
	}, nil)
	commit, err := svc.Commit(context.TODO(), &git.CommitRequest{
		GitProjectId: 1,
		Branch:       "branch",
		Commit:       "commit",
	})
	assert.Nil(t, err)
	if assert.NotNil(t, commit) {
		assert.Equal(t, "abc123", commit.Id)
		assert.Equal(t, "abc", commit.ShortId)
		assert.Equal(t, int32(1), commit.GitProjectId)
		assert.Equal(t, "fix bug", commit.Title)
		assert.Equal(t, "branch", commit.Branch)
		assert.Equal(t, "duc", commit.AuthorName)
		assert.Equal(t, "duc@example.com", commit.AuthorEmail)
		assert.Equal(t, "duc", commit.CommitterName)
		assert.Equal(t, "duc@example.com", commit.CommitterEmail)
		assert.Equal(t, "http://git/abc123", commit.WebUrl)
		assert.Equal(t, "fix the bug", commit.Message)
	}
}

func Test_gitSvc_CommitOptions(t *testing.T) {
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo
	gitRepo.EXPECT().ListCommits(gomock.Any(), 1, "xxx").Return(nil, errors.New("error"))
	options, err := svc.CommitOptions(context.TODO(), &git.CommitOptionsRequest{
		GitProjectId: 1,
		Branch:       "xxx",
	})
	assert.Nil(t, options)
	assert.NotNil(t, err)
}

func Test_gitSvc_CommitOptions_Success(t *testing.T) {
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo
	gitRepo.EXPECT().ListCommits(gomock.Any(), 1, "xxx").Return([]*biz.Commit{
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
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo
	gitRepo.EXPECT().GetChartValuesYaml(gomock.Any(), "chart-values").Return("image: nginx:latest", nil)
	resp, err := svc.GetChartValuesYaml(context.TODO(), &git.GetChartValuesYamlRequest{
		Input: "chart-values",
	})
	assert.Nil(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, "image: nginx:latest", resp.Values)
	}
}

func Test_gitSvc_GetChartValuesYaml_error(t *testing.T) {
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo
	gitRepo.EXPECT().GetChartValuesYaml(gomock.Any(), "").Return("", errors.New("x"))
	resp, err := svc.GetChartValuesYaml(context.TODO(), &git.GetChartValuesYamlRequest{
		Input: "",
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_gitSvc_PipelineInfo_Success(t *testing.T) {
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo

	gitRepo.EXPECT().GetCommitPipeline(gomock.Any(), 1, "main", "commit").Return(&biz.Pipeline{
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
	svc, mocks := newGitSvcWithMocks(t)
	gitRepo := mocks.gitRepo

	gitRepo.EXPECT().GetCommitPipeline(gomock.Any(), 1, "main", "commit").Return(nil, errors.New("error"))

	res, err := svc.PipelineInfo(context.TODO(), &git.PipelineInfoRequest{
		GitProjectId: "1",
		Branch:       "main",
		Commit:       "commit",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

type gitSvcMocks struct {
	ctrl     *gomock.Controller
	gitRepo  *data.MockGitRepo
	repoRepo *data.MockRepoRepo
}

func newGitSvcWithMocks(t *testing.T) (*gitSvc, *gitSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &gitSvcMocks{
		ctrl:     ctrl,
		gitRepo:  data.NewMockGitRepo(ctrl),
		repoRepo: data.NewMockRepoRepo(ctrl),
	}
	s, ok := NewGitSvc(GitSvcDeps{
		RepoBiz: biz.NewRepoBiz(mocks.repoRepo),
		Logger:  mlog.NewForConfig(nil),
		GitBiz:  biz.NewGitBiz(mocks.gitRepo),
	}).(*gitSvc)
	if !ok {
		panic("NewGitSvc returned unexpected type")
	}
	return s, mocks
}
