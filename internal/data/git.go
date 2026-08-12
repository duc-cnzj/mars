package data

import (
	"context"
	"encoding/json"
	"fmt"
	gopath "path"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/spf13/cast"
)

// GitServer 是 gitRepo 消费的 git 端口：从 app.GitServer 收窄而来，
// 只声明 gitRepo 实际用到的方法（GetFileContentWithSha/GetDirectoryFilesWith*
// 等由上层 git 服务直连插件，不走 gitRepo）。
type GitServer interface {
	// AllProjects 返回全部 git 项目。
	AllProjects() ([]*biz.GitProject, error)
	// AllBranches 返回指定项目的全部分支。
	AllBranches(pid string) ([]*biz.Branch, error)
	// GetProject 返回单个 git 项目。
	GetProject(pid string) (*biz.GitProject, error)
	// GetCommit 返回单个提交。
	GetCommit(pid string, sha string) (*biz.Commit, error)
	// ListCommits 返回某分支的提交列表。
	ListCommits(pid string, branch string) ([]*biz.Commit, error)
	// GetCommitPipeline 返回某提交的流水线。
	GetCommitPipeline(pid string, branch string, sha string) (*biz.Pipeline, error)
	// GetFileContentWithBranch 按分支返回文件内容。
	GetFileContentWithBranch(pid string, branch, filename string) (string, error)
}

var _ biz.GitRepo = (*gitRepo)(nil)

// gitRepo 是 git 仓库的 biz.GitRepo 实现：经惰性取数闭包实时解析已加载的 git
// 插件（插件在 bootstrap 阶段才 Initialize，wire 期拿不到实例），GitServerCached
// 开关决定走缓存还是透传，透传时统一 errs.Wrap 携带上下文。
type gitRepo struct {
	logger    mlog.Logger
	gitServer func() GitServer
	cfg       *config.Config
	cache     Cache
}

// NewGitRepo 构造 git 仓库的 repo 实现。注入 *config.Config 而非完整 dataStore：
// gitRepo 只读 GitServerCached 判断缓存开关（ISP），不摸 DB/k8s 客户端。
// gitServer 是惰性取数闭包：插件在 bootstrap 阶段才完成 Initialize，wire 期
// pm.Git() 恒为 nil，首次调用方法时才实时解析（替代原 GitServerHolder 快照）。
func NewGitRepo(logger mlog.Logger, c Cache, gitServer func() GitServer, cfg *config.Config) biz.GitRepo {
	return &gitRepo{
		logger:    logger.WithModule("repo/git"),
		gitServer: gitServer,
		cache:     c,
		cfg:       cfg,
	}
}

// AllProjects 返回全部 git 项目：未开启缓存时透传插件调用并包装错误，开启时
// 经缓存 Remember 合并重复读，forceFresh 强制跳过缓存。
func (g *gitRepo) AllProjects(ctx context.Context, forceFresh bool) (projects []*biz.GitProject, err error) {
	_, span := tracer.Start(ctx, "gitRepo/AllProjects")
	defer func() { endSpan(span, err) }()
	fn := func() ([]*biz.GitProject, error) {
		return g.gitServer().AllProjects()
	}
	if !g.cfg.GitServerCached {
		projects, err := fn()
		return projects, errs.Wrap(err, "git all projects")
	}
	remember, err := g.cache.Remember(NewKey("all_projects"), 600, func() ([]byte, error) {
		projects, err := fn()
		if err != nil {
			return nil, err
		}
		return json.Marshal(projects)
	}, forceFresh)
	if err == nil {
		err = json.Unmarshal(remember, &projects)
	}
	return projects, errs.Wrap(err, "git all projects")
}

// GetChartValuesYaml 解析"远端 chart 路径"（uid|branch|dir），非远端路径返回空串。
func (g *gitRepo) GetChartValuesYaml(ctx context.Context, localChartPath string) (content string, err error) {
	ctx, span := tracer.Start(ctx, "gitRepo/GetChartValuesYaml")
	defer func() { endSpan(span, err) }()
	if !biz.IsRemoteLocalChartPath(localChartPath) {
		return "", nil
	}
	split := strings.Split(localChartPath, "|")
	pid := split[0]
	branch := split[1]
	filename := gopath.Join(split[2], "values.yaml")
	return g.GetFileContentWithBranch(ctx, cast.ToInt(pid), branch, filename)
}

// AllBranches 返回项目的全部分支：与 AllProjects 同款缓存/透传路径。
func (g *gitRepo) AllBranches(ctx context.Context, projectID int, forceFresh bool) (branches []*biz.Branch, err error) {
	_, span := tracer.Start(ctx, "gitRepo/AllBranches")
	defer func() { endSpan(span, err) }()
	fn := func() ([]*biz.Branch, error) {
		return g.gitServer().AllBranches(fmt.Sprintf("%d", projectID))
	}
	if !g.cfg.GitServerCached {
		branches, err := fn()
		return branches, errs.Wrap(err, "git all branches")
	}
	remember, err := g.cache.Remember(NewKey("all_branches_%d", projectID), 600, func() ([]byte, error) {
		branches, err := fn()
		if err != nil {
			return nil, err
		}
		return json.Marshal(branches)
	}, forceFresh)
	if err == nil {
		err = json.Unmarshal(remember, &branches)
	}
	return branches, errs.Wrap(err, "git all branches")
}

// ListCommits 返回项目某分支的提交列表。
func (g *gitRepo) ListCommits(ctx context.Context, projectID int, branch string) (commits []*biz.Commit, err error) {
	_, span := tracer.Start(ctx, "gitRepo/ListCommits")
	defer func() { endSpan(span, err) }()
	commits, err = g.gitServer().ListCommits(fmt.Sprintf("%d", projectID), branch)
	if err != nil {
		return nil, errs.Wrap(err, "git list commits")
	}
	return commits, nil
}

// GetProject 返回单个 git 项目。
func (g *gitRepo) GetProject(ctx context.Context, id int) (get *biz.GitProject, err error) {
	_, span := tracer.Start(ctx, "gitRepo/GetProject")
	defer func() { endSpan(span, err) }()
	get, err = g.gitServer().GetProject(fmt.Sprintf("%d", id))
	if err != nil {
		return nil, errs.Wrap(err, "git get project")
	}
	return get, nil
}

// GetFileContentWithBranch 按分支返回文件内容。
func (g *gitRepo) GetFileContentWithBranch(ctx context.Context, projectID int, branch, path string) (content string, err error) {
	_, span := tracer.Start(ctx, "gitRepo/GetFileContentWithBranch")
	defer func() { endSpan(span, err) }()
	withBranch, err := g.gitServer().GetFileContentWithBranch(fmt.Sprintf("%d", projectID), branch, path)
	if err != nil {
		return "", errs.Wrap(err, "git get file content")
	}
	return withBranch, nil
}

// GetCommit 返回单个提交。
func (g *gitRepo) GetCommit(ctx context.Context, projectID int, sha string) (commit *biz.Commit, err error) {
	_, span := tracer.Start(ctx, "gitRepo/GetCommit")
	defer func() { endSpan(span, err) }()
	commit, err = g.gitServer().GetCommit(fmt.Sprintf("%d", projectID), sha)
	return commit, errs.Wrap(err, "git get commit")
}

// GetCommitPipeline 返回某个提交的流水线。
func (g *gitRepo) GetCommitPipeline(ctx context.Context, projectID int, branch, sha string) (pipeline *biz.Pipeline, err error) {
	_, span := tracer.Start(ctx, "gitRepo/GetCommitPipeline")
	defer func() { endSpan(span, err) }()
	pipeline, err = g.gitServer().GetCommitPipeline(fmt.Sprintf("%d", projectID), branch, sha)
	if err != nil {
		return nil, errs.Wrap(err, "git get commit pipeline")
	}
	return pipeline, nil
}

// GetByProjectID 别名 GetProject（biz.GitRepo 接口必需）。
func (g *gitRepo) GetByProjectID(ctx context.Context, id int) (get *biz.GitProject, err error) {
	ctx, span := tracer.Start(ctx, "gitRepo/GetByProjectID")
	defer func() { endSpan(span, err) }()
	return g.GetProject(ctx, id)
}
