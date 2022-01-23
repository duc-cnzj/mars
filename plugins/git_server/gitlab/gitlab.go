package gitlab

import (
	"errors"
	"strconv"
	"time"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/xanzy/go-gitlab"
)

var _ plugins.GitServer = (*server)(nil)

var name = "gitlab"

func init() {
	dr := &server{}
	plugins.RegisterPlugin(dr.Name(), dr)
}

type project struct {
	p *gitlab.Project
}

func (p *project) GetID() int64 {
	return int64(p.p.ID)
}

func (p *project) GetName() string {
	return p.p.Name
}

func (p *project) GetDefaultBranch() string {
	return p.p.DefaultBranch
}

func (p *project) GetPath() string {
	return p.p.Path
}

func (p *project) GetWebURL() string {
	return p.p.WebURL
}

func (p *project) GetAvatarURL() string {
	return p.p.AvatarURL
}

func (p *project) GetDescription() string {
	return p.p.Description
}

type branch struct {
	b *gitlab.Branch
}

func (b *branch) GetName() string {
	return b.b.Name
}

func (b *branch) IsDefault() bool {
	return b.b.Default
}

func (b *branch) GetWebURL() string {
	return b.b.WebURL
}

type commit struct {
	c *gitlab.Commit
}

func (c *commit) GetID() string {
	return c.c.ID
}

func (c *commit) GetShortID() string {
	return c.c.ShortID
}

func (c *commit) GetCommittedDate() *time.Time {
	return c.c.CommittedDate
}

func (c *commit) GetTitle() string {
	return c.c.Title
}

func (c *commit) GetAuthorName() string {
	return c.c.AuthorName
}

func (c *commit) GetAuthorEmail() string {
	return c.c.AuthorEmail
}

func (c *commit) GetCommitterName() string {
	return c.c.CommitterName
}

func (c *commit) GetCommitterEmail() string {
	return c.c.CommitterEmail
}

func (c *commit) GetCreatedAt() *time.Time {
	return c.c.CreatedAt
}

func (c *commit) GetMessage() string {
	return c.c.Message
}

func (c *commit) GetProjectID() int64 {
	return int64(c.c.ProjectID)
}

func (c *commit) GetWebURL() string {
	return c.c.WebURL
}

type pipeline struct {
	p *gitlab.PipelineInfo
}

func (p *pipeline) GetID() int64 {
	return int64(p.p.ID)
}

func (p *pipeline) GetProjectID() int64 {
	return int64(p.p.ProjectID)
}

func (p *pipeline) GetStatus() plugins.Status {
	switch p.p.Status {
	case "failed":
		return plugins.StatusFailed
	case "running":
		return plugins.StatusRunning
	case "success":
		return plugins.StatusSuccess
	default:
		return plugins.StatusUnknown
	}
}

func (p *pipeline) GetRef() string {
	return p.p.Ref
}

func (p *pipeline) GetSHA() string {
	return p.p.SHA
}

func (p *pipeline) GetWebURL() string {
	return p.p.WebURL
}

func (p *pipeline) GetUpdatedAt() *time.Time {
	return p.p.UpdatedAt
}

func (p *pipeline) GetCreatedAt() *time.Time {
	return p.p.CreatedAt
}

type server struct {
	client *gitlab.Client
}

func (g *server) GetCommitPipeline(pid string, sha string) (plugins.PipelineInterface, error) {
	c, _, err := g.client.Commits.GetCommit(pid, sha)
	if err != nil {
		return nil, err
	}

	if c.LastPipeline == nil {
		return nil, errors.New("pipeline not found")
	}

	return &pipeline{p: c.LastPipeline}, nil
}

func (g *server) Name() string {
	return name
}

func (g *server) Initialize(args map[string]interface{}) error {
	client, err := gitlab.NewClient(args["token"].(string), gitlab.WithBaseURL(args["baseurl"].(string)))
	if err != nil {
		return err
	}
	g.client = client

	mlog.Info("[Plugin]: " + g.Name() + " plugin Initialize...")
	return nil
}

func (g *server) Destroy() error {
	mlog.Info("[Plugin]: " + g.Name() + " plugin Destroy...")
	return nil
}

func (g *server) GetProject(pid string) (plugins.ProjectInterface, error) {
	p, _, err := g.client.Projects.GetProject(pid, &gitlab.GetProjectOptions{})

	return &project{p: p}, err
}

type listProjectResponse struct {
	items                    []plugins.ProjectInterface
	page, pageSize, nextPage int
	hasMore                  bool
}

func (l *listProjectResponse) NextPage() int {
	return l.nextPage
}

func (l *listProjectResponse) HasMore() bool {
	return l.hasMore
}

func (l *listProjectResponse) GetItems() []plugins.ProjectInterface {
	return l.items
}

func (l *listProjectResponse) Page() int {
	return l.page
}

func (l *listProjectResponse) PageSize() int {
	return l.pageSize
}

func (g *server) ListProjects(page, pageSize int) (plugins.ListProjectResponseInterface, error) {
	res, r, err := g.client.Projects.ListProjects(&gitlab.ListProjectsOptions{
		MinAccessLevel: gitlab.AccessLevel(gitlab.DeveloperPermissions),
		ListOptions:    gitlab.ListOptions{PerPage: pageSize, Page: page},
	})
	if err != nil {
		return nil, err
	}
	nextPage := r.Header.Get("x-next-page")
	var projects = make([]plugins.ProjectInterface, 0, len(res))
	for _, re := range res {
		projects = append(projects, &project{p: re})
	}

	var next int
	if nextPage != "" {
		next, _ = strconv.Atoi(nextPage)
	}

	return &listProjectResponse{
		items:    projects,
		page:     page,
		pageSize: pageSize,
		nextPage: next,
		hasMore:  nextPage != "",
	}, err
}

func (g *server) AllProjects() ([]plugins.ProjectInterface, error) {
	var ps []plugins.ProjectInterface
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
		for _, p := range projects.GetItems() {
			ps = append(ps, p)
		}
	}

	return ps, nil
}

type listBranchResponse struct {
	items                    []plugins.BranchInterface
	page, pageSize, nextPage int
	hasMore                  bool
}

func (l *listBranchResponse) NextPage() int {
	return l.nextPage
}

func (l *listBranchResponse) HasMore() bool {
	return l.hasMore
}

func (l *listBranchResponse) GetItems() []plugins.BranchInterface {
	return l.items
}

func (l *listBranchResponse) Page() int {
	return l.page
}

func (l *listBranchResponse) PageSize() int {
	return l.pageSize
}

func (g *server) ListBranches(pid string, page, pageSize int) (plugins.ListBranchResponseInterface, error) {
	var (
		branches []plugins.BranchInterface
		next     int
	)

	gitlabBranches, r, e := g.client.Branches.ListBranches(pid, &gitlab.ListBranchesOptions{ListOptions: gitlab.ListOptions{PerPage: pageSize, Page: page}})
	if e != nil {
		return nil, e
	}
	nextPage := r.Header.Get("x-next-page")
	for _, gitlabBranch := range gitlabBranches {
		branches = append(branches, &branch{b: gitlabBranch})
	}
	if nextPage != "" {
		next, _ = strconv.Atoi(nextPage)
	}
	return &listBranchResponse{
		items:    branches,
		page:     page,
		pageSize: pageSize,
		nextPage: next,
		hasMore:  nextPage != "",
	}, nil
}

func (g *server) AllBranches(pid string) ([]plugins.BranchInterface, error) {
	var branches []plugins.BranchInterface
	page := 1
	for page != -1 {
		gitlabBranches, err := g.ListBranches(pid, page, 100)
		if err != nil {
			return nil, err
		}
		if gitlabBranches.HasMore() {
			page = gitlabBranches.NextPage()
		} else {
			page = -1
		}
		for _, gitlabBranch := range gitlabBranches.GetItems() {
			branches = append(branches, gitlabBranch)
		}
	}

	return branches, nil
}

func (g *server) GetCommit(pid string, sha string) (plugins.CommitInterface, error) {
	c, _, err := g.client.Commits.GetCommit(pid, sha)
	if err != nil {
		return nil, err
	}
	return &commit{c: c}, nil
}

func (g *server) ListCommits(pid string, branch string) ([]plugins.CommitInterface, error) {
	commits, _, err := g.client.Commits.ListCommits(pid, &gitlab.ListCommitsOptions{RefName: gitlab.String(branch), ListOptions: gitlab.ListOptions{PerPage: 100}})

	res := make([]plugins.CommitInterface, 0, len(commits))
	for _, c := range commits {
		res = append(res, &commit{c: c})
	}

	return res, err
}

func getRawFile(client *gitlab.Client, pid string, shaOrBranch string, filename string) (string, error) {
	opt := gitlab.GetRawFileOptions{}
	if shaOrBranch != "" {
		opt.Ref = gitlab.String(shaOrBranch)
	}
	raw, _, err := client.RepositoryFiles.GetRawFile(pid, filename, &opt)
	if err != nil {
		mlog.Error(err)
	}
	return string(raw), err
}

func (g *server) GetFileContentWithSha(pid string, sha string, filename string) (string, error) {
	return getRawFile(g.client, pid, sha, filename)
}

func (g *server) GetFileContentWithBranch(pid string, branch string, filename string) (string, error) {
	return getRawFile(g.client, pid, branch, filename)
}

func getDirectoryFiles(g *gitlab.Client, pid interface{}, commit string, path string, recursive bool) ([]string, error) {
	var files []string

	// TODO: 坑, GitlabClient().Repositories.ListTree 带分页！！凸(艹皿艹 )
	opt := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
		Path:      gitlab.String(path),
		Recursive: gitlab.Bool(recursive),
	}
	if commit != "" {
		opt.Ref = gitlab.String(commit)
	}

	tree, _, _ := g.Repositories.ListTree(pid, opt)

	for _, node := range tree {
		if node.Type == "blob" {
			files = append(files, node.Path)
		}
	}

	return files, nil
}

func (g *server) GetDirectoryFilesWithBranch(pid string, branch string, path string, recursive bool) ([]string, error) {
	return getDirectoryFiles(g.client, pid, branch, path, recursive)
}

func (g *server) GetDirectoryFilesWithSha(pid string, sha string, path string, recursive bool) ([]string, error) {
	return getDirectoryFiles(g.client, pid, sha, path, recursive)
}
