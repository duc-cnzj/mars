package repo

import (
	"context"
	"fmt"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type GitRepo interface {
	All(ctx context.Context) (projects []application.Project, err error)
	AllProjectBranches(ctx context.Context, projectID int) (branches []application.Branch, err error)
	ListCommits(ctx context.Context, projectID int, branch string) ([]application.Commit, error)
	GetCommit(ctx context.Context, projectID int, sha string) (application.Commit, error)
	GetCommitPipeline(ctx context.Context, projectID int, branch, sha string) (application.Pipeline, error)
	GetByProjectID(ctx context.Context, id int) (project application.Project, err error)
	GetFileContentWithBranch(ctx context.Context, projectID int, branch, path string) (string, error)
	GetProject(ctx context.Context, id int) (project application.Project, err error)
}

var _ GitRepo = (*gitRepo)(nil)

type gitRepo struct {
	logger mlog.Logger
	pl     application.PluginManger
	db     *ent.Client
}

func NewGitRepo(logger mlog.Logger, pl application.PluginManger, data *data.Data) GitRepo {
	return &gitRepo{
		logger: logger,
		pl:     pl,
		db:     data.DB,
	}
}

func (g *gitRepo) All(ctx context.Context) (projects []application.Project, err error) {
	return g.pl.Git().AllProjects()
}

func (g *gitRepo) AllProjectBranches(ctx context.Context, projectID int) (branches []application.Branch, err error) {
	return g.pl.Git().AllBranches(fmt.Sprintf("%d", projectID))
}

func (g *gitRepo) ListCommits(ctx context.Context, projectID int, branch string) ([]application.Commit, error) {
	return g.pl.Git().ListCommits(fmt.Sprintf("%d", projectID), branch)
}

func (g *gitRepo) GetProject(ctx context.Context, id int) (project application.Project, err error) {
	return g.pl.Git().GetProject(fmt.Sprintf("%d", id))
}

func (g *gitRepo) GetFileContentWithBranch(ctx context.Context, projectID int, branch, path string) (string, error) {
	return g.pl.Git().GetFileContentWithBranch(fmt.Sprintf("%d", projectID), branch, path)
}

func (g *gitRepo) GetCommit(ctx context.Context, projectID int, sha string) (application.Commit, error) {
	return g.pl.Git().GetCommit(fmt.Sprintf("%d", projectID), sha)
}

func (g *gitRepo) GetCommitPipeline(ctx context.Context, projectID int, branch, sha string) (application.Pipeline, error) {
	return g.pl.Git().GetCommitPipeline(fmt.Sprintf("%d", projectID), branch, sha)
}

func (g *gitRepo) GetByProjectID(ctx context.Context, id int) (project application.Project, err error) {
	return g.pl.Git().GetProject(fmt.Sprintf("%d", id))
}
