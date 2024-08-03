package gitlab

import (
	"errors"
	"strconv"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/xanzy/go-gitlab"
)

var _ application.GitServer = (*server)(nil)

var name = "gitlab"

func init() {
	dr := &server{}
	application.RegisterPlugin(dr.Name(), dr)
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

// GetStatus
// created, waiting_for_resource, preparing, pending, running, success, failed, canceled, skipped, manual, scheduled
func (p *pipeline) GetStatus() application.Status {
	switch p.p.Status {
	case "failed":
		return application.StatusFailed
	case "running":
		return application.StatusRunning
	case "success", "manual":
		return application.StatusSuccess
	default:
		return application.StatusUnknown
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
	logger mlog.Logger
}

func (g *server) GetCommitPipeline(pid string, branch string, sha string) (application.Pipeline, error) {
	var p *gitlab.PipelineInfo
	pipelines, _, err := g.client.Pipelines.ListProjectPipelines(pid, &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
		Ref: gitlab.String(branch),
		SHA: gitlab.String(sha),
	})
	if err != nil {
		return nil, err
	}
	// 只拿 push/web 的 pipeline
	for _, info := range pipelines {
		if info.Source == "push" || info.Source == "web" {
			p = info
			break
		}
	}

	if p == nil {
		return nil, errors.New("pipeline not found")
	}

	return &pipeline{p: p}, nil
}

func (g *server) Name() string {
	return name
}

func (g *server) Initialize(app application.App, args map[string]any) error {
	client, err := gitlab.NewClient(args["token"].(string), gitlab.WithBaseURL(args["baseurl"].(string)))
	if err != nil {
		return err
	}
	g.client = client
	g.logger = app.Logger()
	g.logger.Info("[Plugin]: " + g.Name() + " plugin Initialize...")
	return nil
}

func (g *server) Destroy() error {
	g.logger.Info("[Plugin]: " + g.Name() + " plugin Destroy...")
	return nil
}

func (g *server) GetProject(pid string) (application.Project, error) {
	p, _, err := g.client.Projects.GetProject(pid, &gitlab.GetProjectOptions{})

	return &project{p: p}, err
}

type listProjectResponse struct {
	items                    []application.Project
	page, pageSize, nextPage int
	hasMore                  bool
}

func (l *listProjectResponse) NextPage() int {
	return l.nextPage
}

func (l *listProjectResponse) HasMore() bool {
	return l.hasMore
}

func (l *listProjectResponse) GetItems() []application.Project {
	return l.items
}

func (l *listProjectResponse) Page() int {
	return l.page
}

func (l *listProjectResponse) PageSize() int {
	return l.pageSize
}

func (g *server) ListProjects(page, pageSize int) (application.ListProjectResponse, error) {
	res, r, err := g.client.Projects.ListProjects(&gitlab.ListProjectsOptions{
		MinAccessLevel: gitlab.AccessLevel(gitlab.DeveloperPermissions),
		ListOptions:    gitlab.ListOptions{PerPage: pageSize, Page: page},
	})
	if err != nil {
		return nil, err
	}
	nextPage := r.Header.Get("x-next-page")
	var projects = make([]application.Project, 0, len(res))
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

func (g *server) AllProjects() ([]application.Project, error) {
	var ps []application.Project
	page := 1
	pageSize := 100
	for page != -1 {
		projects, err := g.ListProjects(page, pageSize)
		if err != nil {
			return nil, err
		}
		if len(projects.GetItems()) < pageSize {
			page = -1
		} else {
			page++
		}
		// gitlab 分页貌似有 bug, 有时能返回分页信息有时不能
		//if projects.HasMore() {
		//	page = projects.NextPage()
		//} else {
		//	page = -1
		//}
		ps = append(ps, projects.GetItems()...)
	}

	return ps, nil
}

type listBranchResponse struct {
	items                    []application.Branch
	page, pageSize, nextPage int
	hasMore                  bool
}

func (l *listBranchResponse) NextPage() int {
	return l.nextPage
}

func (l *listBranchResponse) HasMore() bool {
	return l.hasMore
}

func (l *listBranchResponse) GetItems() []application.Branch {
	return l.items
}

func (l *listBranchResponse) Page() int {
	return l.page
}

func (l *listBranchResponse) PageSize() int {
	return l.pageSize
}

func (g *server) ListBranches(pid string, page, pageSize int) (application.ListBranchResponse, error) {
	var (
		branches []application.Branch
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

func (g *server) AllBranches(pid string) ([]application.Branch, error) {
	var branches []application.Branch
	page := 1
	pageSize := 100
	for page != -1 {
		gitlabBranches, err := g.ListBranches(pid, page, pageSize)
		if err != nil {
			return nil, err
		}
		if len(gitlabBranches.GetItems()) < pageSize {
			page = -1
		} else {
			page++
		}
		// gitlab 分页貌似有 bug, 有时能返回分页信息有时不能
		//if gitlabBranches.HasMore() {
		//	page = gitlabBranches.NextPage()
		//} else {
		//	page = -1
		//}
		branches = append(branches, gitlabBranches.GetItems()...)
	}

	return branches, nil
}

func (g *server) GetCommit(pid string, sha string) (application.Commit, error) {
	c, _, err := g.client.Commits.GetCommit(pid, sha)
	if err != nil {
		return nil, err
	}
	return &commit{c: c}, nil
}

func (g *server) ListCommits(pid string, branch string) ([]application.Commit, error) {
	commits, _, err := g.client.Commits.ListCommits(pid, &gitlab.ListCommitsOptions{RefName: gitlab.String(branch), ListOptions: gitlab.ListOptions{PerPage: 100}})

	res := make([]application.Commit, 0, len(commits))
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
		//mlog.Warning(err)
	}
	return string(raw), err
}

func (g *server) GetFileContentWithSha(pid string, sha string, filename string) (string, error) {
	return getRawFile(g.client, pid, sha, filename)
}

func (g *server) GetFileContentWithBranch(pid string, branch string, filename string) (string, error) {
	return getRawFile(g.client, pid, branch, filename)
}

func getDirectoryFiles(g *gitlab.Client, pid any, commit string, path string, recursive bool) ([]string, error) {
	var files []string

	opt := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
		Path:      gitlab.String(path),
		Recursive: gitlab.Bool(recursive),
	}
	if commit != "" {
		opt.Ref = gitlab.String(commit)
	}

	for opt.Page != -1 {
		tree, _, err := g.Repositories.ListTree(pid, opt)
		if err != nil {
			return nil, err
		}
		if len(tree) != opt.PerPage {
			opt.Page = -1
		} else {
			opt.Page++
		}
		for _, node := range tree {
			if node.Type == "blob" {
				files = append(files, node.Path)
			}
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
