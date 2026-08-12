package data

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
)

// fakeGitServer 是 data.GitServer 的替身：err 非 nil 时全部方法返回该错误，
// 否则返回固定数据，用于隔离 gitRepo 的缓存/透传与错误包装逻辑。
type fakeGitServer struct{ err error }

func (f fakeGitServer) AllProjects() ([]*biz.GitProject, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []*biz.GitProject{{ID: 1, Name: "p"}}, nil
}

func (f fakeGitServer) AllBranches(pid string) ([]*biz.Branch, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []*biz.Branch{{Name: "main"}}, nil
}

func (f fakeGitServer) GetProject(pid string) (*biz.GitProject, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &biz.GitProject{ID: 1, Name: "p"}, nil
}

func (f fakeGitServer) GetCommit(pid, sha string) (*biz.Commit, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &biz.Commit{ID: sha}, nil
}

func (f fakeGitServer) ListCommits(pid, branch string) ([]*biz.Commit, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []*biz.Commit{{ID: "abc"}}, nil
}

func (f fakeGitServer) GetCommitPipeline(pid, branch, sha string) (*biz.Pipeline, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &biz.Pipeline{SHA: sha, Status: biz.StatusSuccess}, nil
}

func (f fakeGitServer) GetFileContentWithBranch(pid, branch, filename string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return "content", nil
}

// newTestGitRepo 构造 gitRepo：注入返回 fakeGitServer 的惰性取数闭包（对齐生产
// 路径 provideGitServer 的解析方式），cached 决定走缓存还是透传路径。
func newTestGitRepo(cached bool, f fakeGitServer) *gitRepo {
	r := NewGitRepo(mlog.NewForConfig(nil), &noCache{}, func() GitServer { return f }, &config.Config{GitServerCached: cached})
	return r.(*gitRepo)
}

func TestGitRepo_AllProjects(t *testing.T) {
	for _, cached := range []bool{true, false} {
		g := newTestGitRepo(cached, fakeGitServer{})
		projects, err := g.AllProjects(context.TODO(), true)
		assert.NoError(t, err)
		assert.Len(t, projects, 1)
		assert.Equal(t, int64(1), projects[0].ID)

		g2 := newTestGitRepo(cached, fakeGitServer{err: errors.New("boom")})
		_, err = g2.AllProjects(context.TODO(), false)
		assert.Error(t, err)
	}
}

func TestGitRepo_AllBranches(t *testing.T) {
	for _, cached := range []bool{true, false} {
		g := newTestGitRepo(cached, fakeGitServer{})
		branches, err := g.AllBranches(context.TODO(), 100, true)
		assert.NoError(t, err)
		assert.Len(t, branches, 1)
		assert.Equal(t, "main", branches[0].Name)

		g2 := newTestGitRepo(cached, fakeGitServer{err: errors.New("boom")})
		_, err = g2.AllBranches(context.TODO(), 100, false)
		assert.Error(t, err)
	}
}

func TestGitRepo_GetChartValuesYaml(t *testing.T) {
	g := newTestGitRepo(false, fakeGitServer{})
	// 非远端路径（无 "|" 三段结构）直接返回空串。
	v, err := g.GetChartValuesYaml(context.TODO(), "local")
	assert.NoError(t, err)
	assert.Equal(t, "", v)
	// 远端路径（uid|branch|dir）解析出 values.yaml 走 GetFileContentWithBranch。
	v, err = g.GetChartValuesYaml(context.TODO(), "100|main|chart")
	assert.NoError(t, err)
	assert.Equal(t, "content", v)
}

func TestGitRepo_ListCommits(t *testing.T) {
	g := newTestGitRepo(false, fakeGitServer{})
	commits, err := g.ListCommits(context.TODO(), 100, "main")
	assert.NoError(t, err)
	assert.Len(t, commits, 1)

	g2 := newTestGitRepo(false, fakeGitServer{err: errors.New("boom")})
	_, err = g2.ListCommits(context.TODO(), 100, "main")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git list commits")
}

func TestGitRepo_GetProject(t *testing.T) {
	g := newTestGitRepo(false, fakeGitServer{})
	project, err := g.GetProject(context.TODO(), 100)
	assert.NoError(t, err)
	assert.Equal(t, "p", project.Name)

	g2 := newTestGitRepo(false, fakeGitServer{err: errors.New("boom")})
	_, err = g2.GetProject(context.TODO(), 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git get project")
}

func TestGitRepo_GetFileContentWithBranch(t *testing.T) {
	g := newTestGitRepo(false, fakeGitServer{})
	content, err := g.GetFileContentWithBranch(context.TODO(), 100, "main", "chart/values.yaml")
	assert.NoError(t, err)
	assert.Equal(t, "content", content)

	g2 := newTestGitRepo(false, fakeGitServer{err: errors.New("boom")})
	_, err = g2.GetFileContentWithBranch(context.TODO(), 100, "main", "chart/values.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git get file content")
}

func TestGitRepo_GetCommit(t *testing.T) {
	g := newTestGitRepo(false, fakeGitServer{})
	commit, err := g.GetCommit(context.TODO(), 100, "sha1")
	assert.NoError(t, err)
	assert.Equal(t, "sha1", commit.ID)

	g2 := newTestGitRepo(false, fakeGitServer{err: errors.New("boom")})
	_, err = g2.GetCommit(context.TODO(), 100, "sha1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git get commit")
	assert.Contains(t, err.Error(), "boom")
}

func TestGitRepo_GetCommitPipeline(t *testing.T) {
	g := newTestGitRepo(false, fakeGitServer{})
	pipeline, err := g.GetCommitPipeline(context.TODO(), 100, "main", "sha1")
	assert.NoError(t, err)
	assert.Equal(t, "sha1", pipeline.SHA)

	g2 := newTestGitRepo(false, fakeGitServer{err: errors.New("boom")})
	_, err = g2.GetCommitPipeline(context.TODO(), 100, "main", "sha1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git get commit pipeline")
}

func TestGitRepo_GetByProjectID(t *testing.T) {
	// 委托 GetProject：成功透传数据，错误带 "git get project" 上下文。
	g := newTestGitRepo(false, fakeGitServer{})
	project, err := g.GetByProjectID(context.TODO(), 100)
	assert.NoError(t, err)
	assert.Equal(t, "p", project.Name)

	g2 := newTestGitRepo(false, fakeGitServer{err: errors.New("boom")})
	_, err = g2.GetByProjectID(context.TODO(), 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git get project")
}
