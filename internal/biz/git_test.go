package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeGitRepo 覆盖 GitBiz 透传测试用到的 GitRepo 方法，其余由嵌入接口兜底。
type fakeGitRepo struct {
	GitRepo
	listCommits        func(ctx context.Context, projectID int, branch string) ([]*Commit, error)
	allProjects        func(ctx context.Context, forceFresh bool) ([]*GitProject, error)
	allBranches        func(ctx context.Context, projectID int, forceFresh bool) ([]*Branch, error)
	getCommit          func(ctx context.Context, projectID int, sha string) (*Commit, error)
	getCommitPipeline  func(ctx context.Context, projectID int, branch, sha string) (*Pipeline, error)
	getByProjectID     func(ctx context.Context, id int) (*GitProject, error)
	getFileContent     func(ctx context.Context, projectID int, branch, path string) (string, error)
	getProject         func(ctx context.Context, id int) (*GitProject, error)
	getChartValuesYaml func(ctx context.Context, localChartPath string) (string, error)
}

func (f *fakeGitRepo) ListCommits(ctx context.Context, projectID int, branch string) ([]*Commit, error) {
	return f.listCommits(ctx, projectID, branch)
}

func (f *fakeGitRepo) AllProjects(ctx context.Context, forceFresh bool) ([]*GitProject, error) {
	return f.allProjects(ctx, forceFresh)
}

func (f *fakeGitRepo) AllBranches(ctx context.Context, projectID int, forceFresh bool) ([]*Branch, error) {
	return f.allBranches(ctx, projectID, forceFresh)
}

func (f *fakeGitRepo) GetCommit(ctx context.Context, projectID int, sha string) (*Commit, error) {
	return f.getCommit(ctx, projectID, sha)
}

func (f *fakeGitRepo) GetCommitPipeline(ctx context.Context, projectID int, branch, sha string) (*Pipeline, error) {
	return f.getCommitPipeline(ctx, projectID, branch, sha)
}

func (f *fakeGitRepo) GetByProjectID(ctx context.Context, id int) (*GitProject, error) {
	return f.getByProjectID(ctx, id)
}

func (f *fakeGitRepo) GetFileContentWithBranch(ctx context.Context, projectID int, branch, path string) (string, error) {
	return f.getFileContent(ctx, projectID, branch, path)
}

func (f *fakeGitRepo) GetProject(ctx context.Context, id int) (*GitProject, error) {
	return f.getProject(ctx, id)
}

func (f *fakeGitRepo) GetChartValuesYaml(ctx context.Context, localChartPath string) (string, error) {
	return f.getChartValuesYaml(ctx, localChartPath)
}

func TestGitBiz_EnsureBranchAndCommit_NoBranchNoCommit(t *testing.T) {
	g := NewGitBiz(&fakeGitRepo{
		listCommits: func(ctx context.Context, projectID int, branch string) ([]*Commit, error) {
			assert.Equal(t, "main", branch)
			assert.Equal(t, 1, projectID)
			return []*Commit{{ID: "abc", Title: "fix", WebURL: "http://x/abc"}}, nil
		},
	})
	branch, commit, msgs, err := g.EnsureBranchAndCommit(context.TODO(), &Repo{DefaultBranch: "main", GitProjectID: 1}, "", "")
	assert.Nil(t, err)
	assert.Equal(t, "main", branch)
	assert.Equal(t, "abc", commit)
	assert.Len(t, msgs, 2)
}

func TestGitBiz_EnsureBranchAndCommit_BranchProvided(t *testing.T) {
	g := NewGitBiz(&fakeGitRepo{
		listCommits: func(ctx context.Context, projectID int, branch string) ([]*Commit, error) {
			assert.Equal(t, "dev", branch)
			return []*Commit{{ID: "xyz", Title: "feat", WebURL: "http://x/xyz"}}, nil
		},
	})
	branch, commit, msgs, err := g.EnsureBranchAndCommit(context.TODO(), &Repo{DefaultBranch: "main", GitProjectID: 1}, "dev", "")
	assert.Nil(t, err)
	assert.Equal(t, "dev", branch)
	assert.Equal(t, "xyz", commit)
	assert.Len(t, msgs, 1)
}

func TestGitBiz_EnsureBranchAndCommit_BothProvided(t *testing.T) {
	g := NewGitBiz(&fakeGitRepo{})
	branch, commit, msgs, err := g.EnsureBranchAndCommit(context.TODO(), &Repo{DefaultBranch: "main"}, "dev", "abc123")
	assert.Nil(t, err)
	assert.Equal(t, "dev", branch)
	assert.Equal(t, "abc123", commit)
	assert.Empty(t, msgs)
}

func TestGitBiz_EnsureBranchAndCommit_ListCommitsError(t *testing.T) {
	g := NewGitBiz(&fakeGitRepo{
		listCommits: func(ctx context.Context, projectID int, branch string) ([]*Commit, error) {
			return nil, errors.New("git down")
		},
	})
	branch, commit, msgs, err := g.EnsureBranchAndCommit(context.TODO(), &Repo{DefaultBranch: "main", GitProjectID: 1}, "", "")
	assert.Equal(t, "", branch)
	assert.Equal(t, "", commit)
	assert.Nil(t, msgs)
	assert.NotNil(t, err)
}

func TestGitBiz_EnsureBranchAndCommit_NoCommits(t *testing.T) {
	g := NewGitBiz(&fakeGitRepo{
		listCommits: func(ctx context.Context, projectID int, branch string) ([]*Commit, error) {
			return nil, nil
		},
	})
	branch, commit, msgs, err := g.EnsureBranchAndCommit(context.TODO(), &Repo{DefaultBranch: "main", GitProjectID: 1}, "", "")
	assert.Equal(t, "", branch)
	assert.Equal(t, "", commit)
	assert.Nil(t, msgs)
	assert.ErrorContains(t, err, "没有可用的 commit")
}

func TestGitBiz_EnsureBranchAndCommit_NilShow(t *testing.T) {
	g := NewGitBiz(&fakeGitRepo{})
	branch, commit, msgs, err := g.EnsureBranchAndCommit(context.TODO(), nil, "", "")
	assert.Empty(t, branch)
	assert.Empty(t, commit)
	assert.Empty(t, msgs)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo 不能为空", status.Convert(err).Message())
}

// ---- 纯透传查询 ----

func TestGitBiz_AllProjects_Passthrough(t *testing.T) {
	var called bool
	g := NewGitBiz(&fakeGitRepo{
		allProjects: func(ctx context.Context, forceFresh bool) ([]*GitProject, error) {
			called = true
			assert.True(t, forceFresh)
			return []*GitProject{{ID: 1, Name: "p"}}, nil
		},
	})
	got, err := g.AllProjects(context.TODO(), true)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Len(t, got, 1)
}

func TestGitBiz_AllBranches_Passthrough(t *testing.T) {
	var called bool
	g := NewGitBiz(&fakeGitRepo{
		allBranches: func(ctx context.Context, projectID int, forceFresh bool) ([]*Branch, error) {
			called = true
			assert.Equal(t, 1, projectID)
			return []*Branch{{Name: "main"}}, nil
		},
	})
	got, err := g.AllBranches(context.TODO(), 1, false)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Len(t, got, 1)
}

func TestGitBiz_GetCommit_Passthrough(t *testing.T) {
	var called bool
	g := NewGitBiz(&fakeGitRepo{
		getCommit: func(ctx context.Context, projectID int, sha string) (*Commit, error) {
			called = true
			assert.Equal(t, "sha1", sha)
			return &Commit{ID: "sha1"}, nil
		},
	})
	got, err := g.GetCommit(context.TODO(), 1, "sha1")
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "sha1", got.ID)
}

func TestGitBiz_GetCommitPipeline_Passthrough(t *testing.T) {
	var called bool
	g := NewGitBiz(&fakeGitRepo{
		getCommitPipeline: func(ctx context.Context, projectID int, branch, sha string) (*Pipeline, error) {
			called = true
			assert.Equal(t, "main", branch)
			assert.Equal(t, "sha1", sha)
			return &Pipeline{ID: 9}, nil
		},
	})
	got, err := g.GetCommitPipeline(context.TODO(), 1, "main", "sha1")
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, int64(9), got.ID)
}

func TestGitBiz_GetByProjectID_Passthrough(t *testing.T) {
	var called bool
	g := NewGitBiz(&fakeGitRepo{
		getByProjectID: func(ctx context.Context, id int) (*GitProject, error) {
			called = true
			assert.Equal(t, 3, id)
			return &GitProject{ID: 3}, nil
		},
	})
	got, err := g.GetByProjectID(context.TODO(), 3)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, int64(3), got.ID)
}

func TestGitBiz_GetFileContentWithBranch_Passthrough(t *testing.T) {
	var called bool
	g := NewGitBiz(&fakeGitRepo{
		getFileContent: func(ctx context.Context, projectID int, branch, path string) (string, error) {
			called = true
			assert.Equal(t, "deploy.yaml", path)
			return "content", nil
		},
	})
	got, err := g.GetFileContentWithBranch(context.TODO(), 1, "main", "deploy.yaml")
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "content", got)
}

func TestGitBiz_GetProject_Passthrough(t *testing.T) {
	var called bool
	g := NewGitBiz(&fakeGitRepo{
		getProject: func(ctx context.Context, id int) (*GitProject, error) {
			called = true
			return &GitProject{ID: 4, Name: "repo"}, nil
		},
	})
	got, err := g.GetProject(context.TODO(), 4)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "repo", got.Name)
}

func TestGitBiz_ListCommits_Passthrough(t *testing.T) {
	var called bool
	g := NewGitBiz(&fakeGitRepo{
		listCommits: func(ctx context.Context, projectID int, branch string) ([]*Commit, error) {
			called = true
			assert.Equal(t, 7, projectID)
			assert.Equal(t, "dev", branch)
			return []*Commit{{ID: "c1"}}, nil
		},
	})
	got, err := g.ListCommits(context.TODO(), 7, "dev")
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Len(t, got, 1)
}

func TestGitBiz_GetChartValuesYaml_Passthrough(t *testing.T) {
	var called bool
	g := NewGitBiz(&fakeGitRepo{
		getChartValuesYaml: func(ctx context.Context, localChartPath string) (string, error) {
			called = true
			assert.Equal(t, "charts/app", localChartPath)
			return "values: 1", nil
		},
	})
	got, err := g.GetChartValuesYaml(context.TODO(), "charts/app")
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "values: 1", got)
}
