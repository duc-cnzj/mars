package git_server

import (
	"strconv"
	"time"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/xanzy/go-gitlab"
)

var gitlab_name = "gitlab"

func init() {
	dr := &GitlabServer{}
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

func (p *pipeline) GetStatus() string {
	return p.p.Status
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

func (c *commit) GetLastPipeline() plugins.PipelineInterface {
	return &pipeline{p: c.c.LastPipeline}
}

type GitlabServer struct {
	client *gitlab.Client
}

func (g *GitlabServer) Name() string {
	return gitlab_name
}

func (g *GitlabServer) Initialize(args map[string]interface{}) error {
	client, err := gitlab.NewClient(args["token"].(string), gitlab.WithBaseURL(args["baseurl"].(string)))
	if err != nil {
		return err
	}
	g.client = client

	mlog.Info("[Plugin]: " + g.Name() + " plugin Initialize...")
	return nil
}

func (g *GitlabServer) Destroy() error {
	mlog.Info("[Plugin]: " + g.Name() + " plugin Destroy...")
	return nil
}

func (g *GitlabServer) GetProject(pid string) (plugins.ProjectInterface, error) {
	p, _, err := g.client.Projects.GetProject(pid, &gitlab.GetProjectOptions{})

	return &project{p: p}, err
}

func (g *GitlabServer) AllProjects() ([]plugins.ProjectInterface, error) {
	res, _, err := g.client.Projects.ListProjects(&gitlab.ListProjectsOptions{
		MinAccessLevel: gitlab.AccessLevel(gitlab.DeveloperPermissions),
		ListOptions:    gitlab.ListOptions{PerPage: 100},
	})
	var projects = make([]plugins.ProjectInterface, 0, len(res))
	for _, re := range res {
		projects = append(projects, &project{p: re})
	}

	return projects, err
}

func (g *GitlabServer) AllBranches(pid string) ([]plugins.BranchInterface, error) {
	var branches []plugins.BranchInterface
	page := 1
	for page != -1 {
		gitlabBranches, r, e := g.client.Branches.ListBranches(pid, &gitlab.ListBranchesOptions{ListOptions: gitlab.ListOptions{PerPage: 100, Page: page}})
		if e != nil {
			return nil, e
		}
		nextPage := r.Header.Get("x-next-page")
		if nextPage == "" {
			page = -1
		} else {
			page, _ = strconv.Atoi(nextPage)
		}
		for _, gitlabBranch := range gitlabBranches {
			branches = append(branches, &branch{b: gitlabBranch})
		}
	}
	return branches, nil
}

func (g *GitlabServer) GetCommit(pid string, sha string) (plugins.CommitInterface, error) {
	c, _, err := g.client.Commits.GetCommit(pid, sha)
	return &commit{c: c}, err
}

func (g *GitlabServer) ListCommits(pid string, branch string) ([]plugins.CommitInterface, error) {
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

func (g *GitlabServer) GetFileContentWithSha(pid string, sha string, filename string) (string, error) {
	return getRawFile(g.client, pid, sha, filename)
}

func (g *GitlabServer) GetFileContentWithBranch(pid string, branch string, filename string) (string, error) {
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

func (g *GitlabServer) GetDirectoryFilesWithBranch(pid string, branch string, path string, recursive bool) ([]string, error) {
	return getDirectoryFiles(g.client, pid, branch, pid, recursive)
}

func (g *GitlabServer) GetDirectoryFilesWithSha(pid string, sha string, path string, recursive bool) ([]string, error) {
	return getDirectoryFiles(g.client, pid, sha, pid, recursive)
}
