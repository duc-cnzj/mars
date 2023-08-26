// Package github
// github api 好烂啊, 接了半天这不行那不行, 真的垃圾
package github

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
	"github.com/google/go-github/v47/github"

	"golang.org/x/oauth2"
)

var _ plugins.GitServer = (*server)(nil)

var name = "github"

func init() {
	dr := &server{}
	plugins.RegisterPlugin(dr.Name(), dr)
}

type project struct {
	p *github.Repository
}

func (p *project) GetID() int64 {
	return p.p.GetID()
}

func (p *project) GetName() string {
	return p.p.GetName()
}

func (p *project) GetDefaultBranch() string {
	return p.p.GetDefaultBranch()
}

func (p *project) GetPath() string {
	return p.p.GetFullName()
}

func (p *project) GetWebURL() string {
	return p.p.GetHTMLURL()
}

func (p *project) GetAvatarURL() string {
	return ""
}

func (p *project) GetDescription() string {
	return p.p.GetDescription()
}

type server struct {
	client   *github.Client
	user     *github.User
	username string
}

func (g *server) Name() string {
	return name
}

func (g *server) Initialize(args map[string]any) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: args["token"].(string)},
	)
	g.client = github.NewClient(oauth2.NewClient(context.TODO(), ts))

	user, _, err := g.client.Users.Get(context.TODO(), "")
	if err != nil {
		return err
	}
	g.user = user
	g.username = user.GetLogin()
	mlog.Info("[Plugin]: " + g.Name() + " plugin Initialize...")
	return nil
}

func (g *server) Destroy() error {
	mlog.Info("[Plugin]: " + g.Name() + " plugin Destroy...")
	return nil
}

func (g *server) GetProject(pid string) (contracts.ProjectInterface, error) {
	p, _, err := g.client.Repositories.GetByID(context.TODO(), toInt64(pid))
	if err != nil {
		return nil, err
	}

	return &project{p: p}, nil
}

type listProjectResponse struct {
	items                    []contracts.ProjectInterface
	page, pageSize, nextPage int
	hasMore                  bool
}

func (l *listProjectResponse) NextPage() int {
	return l.nextPage
}

func (l *listProjectResponse) HasMore() bool {
	return l.hasMore
}

func (l *listProjectResponse) GetItems() []contracts.ProjectInterface {
	return l.items
}

func (l *listProjectResponse) Page() int {
	return l.page
}

func (l *listProjectResponse) PageSize() int {
	return l.pageSize
}

func (g *server) ListProjects(page, pageSize int) (contracts.ListProjectResponseInterface, error) {
	list, _, err := g.client.Repositories.List(context.TODO(), g.username, &github.RepositoryListOptions{
		Sort:        "updated",
		ListOptions: github.ListOptions{Page: page, PerPage: pageSize},
	})
	if err != nil {
		return nil, err
	}
	ps := make([]contracts.ProjectInterface, 0, len(list))
	for _, repository := range list {
		ps = append(ps, &project{p: repository})
	}
	var (
		nextPage int = page + 1
		hasMore  bool
	)
	if pageSize != len(list) {
		hasMore = false
		nextPage = -1
	}

	return &listProjectResponse{
		items:    ps,
		page:     page,
		pageSize: pageSize,
		nextPage: nextPage,
		hasMore:  hasMore,
	}, nil
}

func (g *server) AllProjects() ([]contracts.ProjectInterface, error) {
	var ps []contracts.ProjectInterface
	page := 1
	for page != -1 {
		projects, err := g.ListProjects(page, 100)
		if err != nil {
			return nil, err
		}
		if projects.HasMore() {
			page = projects.NextPage()
		} else {
			page = -1
		}
		ps = append(ps, projects.GetItems()...)
	}

	return ps, nil
}

type listBranchResponse struct {
	items                    []contracts.BranchInterface
	page, pageSize, nextPage int
	hasMore                  bool
}

func (l *listBranchResponse) NextPage() int {
	return l.nextPage
}

func (l *listBranchResponse) HasMore() bool {
	return l.hasMore
}

func (l *listBranchResponse) GetItems() []contracts.BranchInterface {
	return l.items
}

func (l *listBranchResponse) Page() int {
	return l.page
}

func (l *listBranchResponse) PageSize() int {
	return l.pageSize
}

type branch struct {
	b    *github.Branch
	repo *github.Repository
}

func (b *branch) GetName() string {
	return b.b.GetName()
}

func (b *branch) IsDefault() bool {
	return b.b.GetName() == b.repo.GetDefaultBranch()
}

func (b *branch) GetWebURL() string {
	return ""
}

func (g *server) ListBranches(pid string, page, pageSize int) (contracts.ListBranchResponseInterface, error) {
	p, _, _ := g.client.Repositories.GetByID(context.TODO(), toInt64(pid))

	branches, _, err := g.client.Repositories.ListBranches(context.TODO(), g.username, p.GetName(), &github.BranchListOptions{
		ListOptions: github.ListOptions{Page: page, PerPage: pageSize},
	})
	if err != nil {
		return nil, err
	}
	bs := make([]contracts.BranchInterface, 0, len(branches))
	for _, b := range branches {
		bs = append(bs, &branch{
			b:    b,
			repo: p,
		})
	}
	var (
		nextPage int = page + 1
		hasMore  bool
	)
	if pageSize != len(branches) {
		hasMore = false
		nextPage = -1
	}

	return &listBranchResponse{
		items:    bs,
		page:     page,
		pageSize: pageSize,
		nextPage: nextPage,
		hasMore:  hasMore,
	}, nil
}

func (g *server) AllBranches(pid string) ([]contracts.BranchInterface, error) {
	var branches []contracts.BranchInterface
	page := 1
	for page != -1 {
		githubBranches, err := g.ListBranches(pid, page, 100)
		if err != nil {
			return nil, err
		}
		if githubBranches.HasMore() {
			page = githubBranches.NextPage()
		} else {
			page = -1
		}
		branches = append(branches, githubBranches.GetItems()...)
	}

	return branches, nil
}

type commit struct {
	c      *github.RepositoryCommit
	p      *github.Repository
	status *github.RepoStatus
}

func (c *commit) GetID() string {
	return c.c.GetSHA()
}

func (c *commit) GetShortID() string {
	return c.c.GetSHA()[0:7]
}

func (c *commit) GetCommittedDate() *time.Time {
	t := c.c.GetCommit().GetCommitter().GetDate()
	return &t
}

func (c *commit) GetTitle() string {
	return c.c.GetCommit().GetMessage()
}

func (c *commit) GetAuthorName() string {
	return c.c.GetCommit().GetAuthor().GetName()
}

func (c *commit) GetAuthorEmail() string {
	return c.c.GetCommit().GetAuthor().GetEmail()
}

func (c *commit) GetCommitterName() string {
	return c.c.GetCommit().GetCommitter().GetName()
}

func (c *commit) GetCommitterEmail() string {
	return c.c.GetCommit().GetCommitter().GetEmail()
}

func (c *commit) GetCreatedAt() *time.Time {
	return c.GetCommittedDate()
}

func (c *commit) GetMessage() string {
	return c.c.GetCommit().GetMessage()
}

func (c *commit) GetProjectID() int64 {
	return c.p.GetID()
}

func (c *commit) GetWebURL() string {
	return c.c.GetHTMLURL()
}

func (g *server) GetCommitPipeline(pid string, branch string, sha string) (contracts.PipelineInterface, error) {
	return nil, errors.New("github unimplemented this func")
}

func (g *server) GetCommit(pid string, sha string) (contracts.CommitInterface, error) {
	p, _, _ := g.client.Repositories.GetByID(context.TODO(), toInt64(pid))
	c, _, err := g.client.Repositories.GetCommit(context.TODO(), g.username, p.GetName(), sha, &github.ListOptions{})
	if err != nil {
		return nil, err
	}
	return &commit{c: c, p: p, status: nil}, nil
}

func (g *server) ListCommits(pid string, branch string) ([]contracts.CommitInterface, error) {
	p, _, _ := g.client.Repositories.GetByID(context.TODO(), toInt64(pid))
	cs, _, err := g.client.Repositories.ListCommits(context.TODO(), g.username, p.GetName(), &github.CommitsListOptions{
		SHA:         branch,
		ListOptions: github.ListOptions{Page: 1, PerPage: 100},
	})
	if err != nil {
		return nil, err
	}
	res := make([]contracts.CommitInterface, 0, len(cs))
	for _, c := range cs {
		res = append(res, &commit{c: c, status: nil})
	}
	return res, nil
}

func (g *server) GetFileContentWithBranch(pid string, branch string, filename string) (string, error) {
	return g.getRaw(pid, branch, filename)
}

func (g *server) getRaw(pid string, branch string, filename string) (string, error) {
	p, _, err := g.client.Repositories.GetByID(context.TODO(), toInt64(pid))
	if err != nil {
		return "", err
	}
	contents, _, _, err := g.client.Repositories.GetContents(context.TODO(), g.username, p.GetName(), filename, &github.RepositoryContentGetOptions{
		Ref: branch,
	})
	if err != nil || contents == nil {
		return "", err
	}

	return contents.GetContent()
}

func (g *server) GetFileContentWithSha(pid string, sha string, filename string) (string, error) {
	return g.getRaw(pid, sha, filename)
}

func (g *server) GetDirectoryFilesWithBranch(pid string, branch string, path string, recursive bool) ([]string, error) {
	p, _, _ := g.client.Repositories.GetByID(context.TODO(), toInt64(pid))

	return g.getDirectoryFilesWithBranch(pid, branch, p.GetName(), path, recursive)
}

func (g *server) getDirectoryFilesWithBranch(pid string, branch string, repo string, path string, recursive bool) ([]string, error) {
	_, directoryContent, _, _ := g.client.Repositories.GetContents(context.TODO(), g.username, repo, path, &github.RepositoryContentGetOptions{
		Ref: branch,
	})
	var res []string
	for _, content := range directoryContent {
		if content.GetType() == "file" {
			res = append(res, content.GetPath())
		}
		if content.GetType() == "dir" && recursive {
			files, _ := g.getDirectoryFilesWithBranch(pid, branch, repo, content.GetPath(), recursive)
			res = append(res, files...)
		}
	}

	return res, nil
}

func (g *server) GetDirectoryFilesWithSha(pid string, sha string, path string, recursive bool) ([]string, error) {
	return g.GetDirectoryFilesWithBranch(pid, sha, path, recursive)
}

func toInt64(s string) int64 {
	atoi, _ := strconv.Atoi(s)
	return int64(atoi)
}
