package plugins

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
)

var gitServerOnce = sync.Once{}

type Status = string

const (
	StatusUnknown Status = "unknown"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
	StatusRunning Status = "running"
)

type ProjectInterface interface {
	GetID() int64
	GetName() string
	GetDefaultBranch() string
	GetPath() string
	GetWebURL() string
	GetAvatarURL() string
	GetDescription() string
}

type BranchInterface interface {
	GetName() string
	IsDefault() bool
	GetWebURL() string
}

type PipelineInterface interface {
	GetID() int64
	GetProjectID() int64
	GetStatus() Status
	GetRef() string
	GetSHA() string
	GetWebURL() string
	GetUpdatedAt() *time.Time
	GetCreatedAt() *time.Time
}

type CommitInterface interface {
	GetID() string
	GetShortID() string
	GetTitle() string
	GetCommittedDate() *time.Time
	GetAuthorName() string
	GetAuthorEmail() string
	GetCommitterName() string
	GetCommitterEmail() string
	GetCreatedAt() *time.Time
	GetMessage() string
	GetProjectID() int64
	GetWebURL() string
}

type paginate interface {
	Page() int
	PageSize() int
	HasMore() bool
	NextPage() int
}

type ListProjectResponseInterface interface {
	paginate
	GetItems() []ProjectInterface
}

type ListBranchResponseInterface interface {
	paginate
	GetItems() []BranchInterface
}

type GitServer interface {
	GetProject(pid string) (ProjectInterface, error)
	ListProjects(page, pageSize int) (ListProjectResponseInterface, error)
	AllProjects() ([]ProjectInterface, error)

	ListBranches(pid string, page, pageSize int) (ListBranchResponseInterface, error)
	AllBranches(pid string) ([]BranchInterface, error)

	GetCommit(pid string, sha string) (CommitInterface, error)
	GetCommitPipeline(pid string, sha string) (PipelineInterface, error)
	ListCommits(pid string, branch string) ([]CommitInterface, error)

	GetFileContentWithBranch(pid string, branch string, filename string) (string, error)
	GetFileContentWithSha(pid string, sha string, filename string) (string, error)

	GetDirectoryFilesWithBranch(pid string, branch string, path string, recursive bool) ([]string, error)
	GetDirectoryFilesWithSha(pid string, sha string, path string, recursive bool) ([]string, error)
}

func GetGitServer() GitServer {
	pcfg := app.Config().GitServerPlugin
	p := app.App().GetPluginByName(pcfg.Name)
	gitServerOnce.Do(func() {
		if err := p.Initialize(pcfg.GetArgs()); err != nil {
			panic(err)
		}
		app.App().RegisterAfterShutdownFunc(func(app contracts.ApplicationInterface) {
			p.Destroy()
		})
	})

	if app.Config().GitServerCached {
		return newGitServerCache(p.(GitServer))
	}

	return p.(GitServer)
}

var (
	ListCommitsCacheSeconds       int = 10
	AllBranchesCacheSeconds       int = 60 * 2
	AllProjectsCacheSeconds       int = 60 * 5
	GetFileContentCacheSeconds    int = 0
	GetDirectoryFilesCacheSeconds int = 0
)

// gitServerCache
// 用来缓存一些耗时比较久的请求
type gitServerCache struct {
	s GitServer
}

func newGitServerCache(s GitServer) *gitServerCache {
	return &gitServerCache{s: s}
}

func (g *gitServerCache) GetProject(pid string) (ProjectInterface, error) {
	return g.s.GetProject(pid)
}

func (g *gitServerCache) ListProjects(page, pageSize int) (ListProjectResponseInterface, error) {
	return g.s.ListProjects(page, pageSize)
}

func (g *gitServerCache) AllProjects() ([]ProjectInterface, error) {
	remember, err := app.Cache().Remember("AllProjects", AllProjectsCacheSeconds, func() ([]byte, error) {
		projects, err := g.s.AllProjects()
		if err != nil {
			return nil, err
		}
		var all = make([]ProjectInterface, 0, len(projects))
		for _, projectInterface := range projects {
			all = append(all, &project{
				ID:            projectInterface.GetID(),
				Name:          projectInterface.GetName(),
				DefaultBranch: projectInterface.GetDefaultBranch(),
				Path:          projectInterface.GetPath(),
				WebUrl:        projectInterface.GetWebURL(),
				AvatarUrl:     projectInterface.GetAvatarURL(),
				Description:   projectInterface.GetDescription(),
			})
		}
		marshal, _ := json.Marshal(all)
		return marshal, nil
	})
	if err != nil {
		return nil, err
	}
	var res []*project
	json.Unmarshal(remember, &res)
	var all = make([]ProjectInterface, 0, len(res))
	for _, re := range res {
		all = append(all, re)
	}
	return all, nil
}

func (g *gitServerCache) ListBranches(pid string, page, pageSize int) (ListBranchResponseInterface, error) {
	return g.s.ListBranches(pid, page, pageSize)
}

func (g *gitServerCache) AllBranches(pid string) ([]BranchInterface, error) {
	remember, err := app.Cache().Remember(fmt.Sprintf("AllBranches-%v", pid), AllBranchesCacheSeconds, func() ([]byte, error) {
		b, err := g.s.AllBranches(pid)
		if err != nil {
			return nil, err
		}
		var all = make([]BranchInterface, 0, len(b))
		for _, branchInterface := range b {
			all = append(all, &branch{
				Name:    branchInterface.GetName(),
				Default: branchInterface.IsDefault(),
				WebUrl:  branchInterface.GetWebURL(),
			})
		}

		marshal, _ := json.Marshal(all)
		return marshal, nil
	})
	if err != nil {
		return nil, err
	}
	var res []*branch
	json.Unmarshal(remember, &res)
	// Why? 为什么我不能直接返回 res，奇怪的 go 语法
	var all = make([]BranchInterface, 0, len(res))
	for _, b := range res {
		all = append(all, b)
	}
	return all, nil
}

func (g *gitServerCache) GetCommit(pid string, sha string) (CommitInterface, error) {
	return g.s.GetCommit(pid, sha)
}

func (g *gitServerCache) GetCommitPipeline(pid string, sha string) (PipelineInterface, error) {
	return g.s.GetCommitPipeline(pid, sha)
}

func (g *gitServerCache) ListCommits(pid string, branch string) ([]CommitInterface, error) {
	remember, err := app.Cache().Remember(fmt.Sprintf("ListCommits:%s-%s", pid, branch), ListCommitsCacheSeconds, func() ([]byte, error) {
		commits, err := g.s.ListCommits(pid, branch)
		if err != nil {
			return nil, err
		}
		var result = make([]CommitInterface, 0, len(commits))
		for _, commitInterface := range commits {
			result = append(result, &commit{
				ID:             commitInterface.GetID(),
				ShortID:        commitInterface.GetShortID(),
				Title:          commitInterface.GetTitle(),
				CommittedDate:  commitInterface.GetCommittedDate(),
				AuthorName:     commitInterface.GetAuthorName(),
				AuthorEmail:    commitInterface.GetAuthorEmail(),
				CommitterName:  commitInterface.GetCommitterName(),
				CommitterEmail: commitInterface.GetCommitterEmail(),
				CreatedAt:      commitInterface.GetCreatedAt(),
				Message:        commitInterface.GetMessage(),
				ProjectID:      commitInterface.GetProjectID(),
				WebURL:         commitInterface.GetWebURL(),
			})
		}
		marshal, _ := json.Marshal(result)
		return marshal, nil
	})
	if err != nil {
		return nil, err
	}
	var res []*commit
	json.Unmarshal(remember, &res)
	var lists = make([]CommitInterface, 0, len(res))

	for _, re := range res {
		lists = append(lists, re)
	}
	return lists, nil
}

func (g *gitServerCache) GetFileContentWithBranch(pid string, branch string, filename string) (string, error) {
	remember, err := app.Cache().Remember(fmt.Sprintf("GetFileContentWithBranch-%s-%s-%s", pid, branch, filename), GetFileContentCacheSeconds, func() ([]byte, error) {
		content, err := g.s.GetFileContentWithBranch(pid, branch, filename)
		if err != nil {
			return nil, err
		}
		return []byte(content), nil
	})
	if err != nil {
		return "", err
	}
	return string(remember), nil
}

func (g *gitServerCache) GetFileContentWithSha(pid string, sha string, filename string) (string, error) {
	remember, err := app.Cache().Remember(fmt.Sprintf("GetFileContentWithSha-%s-%s-%s", pid, sha, filename), GetFileContentCacheSeconds, func() ([]byte, error) {
		content, err := g.s.GetFileContentWithSha(pid, sha, filename)
		if err != nil {
			return nil, err
		}
		return []byte(content), nil
	})
	if err != nil {
		return "", err
	}
	return string(remember), nil
}

func (g *gitServerCache) GetDirectoryFilesWithBranch(pid string, branch string, path string, recursive bool) ([]string, error) {
	remember, err := app.Cache().Remember(fmt.Sprintf("GetDirectoryFilesWithBranch-%s-%s-%s-%v", pid, branch, path, recursive), GetDirectoryFilesCacheSeconds, func() ([]byte, error) {
		withBranch, err := g.s.GetDirectoryFilesWithBranch(pid, branch, path, recursive)
		if err != nil {
			return nil, err
		}
		marshal, _ := json.Marshal(withBranch)
		return marshal, nil
	})
	if err != nil {
		return nil, err
	}
	var res []string
	json.Unmarshal(remember, &res)
	return res, nil
}

func (g *gitServerCache) GetDirectoryFilesWithSha(pid string, sha string, path string, recursive bool) ([]string, error) {
	remember, err := app.Cache().Remember(fmt.Sprintf("GetDirectoryFilesWithSha-%s-%s-%s-%v", pid, sha, path, recursive), GetDirectoryFilesCacheSeconds, func() ([]byte, error) {
		withBranch, err := g.s.GetDirectoryFilesWithSha(pid, sha, path, recursive)
		if err != nil {
			return nil, err
		}
		marshal, _ := json.Marshal(withBranch)
		return marshal, nil
	})
	if err != nil {
		return nil, err
	}
	var res []string
	json.Unmarshal(remember, &res)
	return res, nil
}

type project struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	Path          string `json:"path"`
	WebUrl        string `json:"web_url"`
	AvatarUrl     string `json:"avatar_url"`
	Description   string `json:"description"`
}

func (p *project) GetID() int64 {
	return p.ID
}

func (p *project) GetName() string {
	return p.Name
}

func (p *project) GetDefaultBranch() string {
	return p.DefaultBranch
}

func (p *project) GetPath() string {
	return p.Path
}

func (p *project) GetWebURL() string {
	return p.WebUrl
}

func (p *project) GetAvatarURL() string {
	return p.AvatarUrl
}

func (p *project) GetDescription() string {
	return p.Description
}

type branch struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
	WebUrl  string `json:"web_url"`
}

func (b *branch) GetName() string {
	return b.Name
}

func (b *branch) IsDefault() bool {
	return b.Default
}

func (b *branch) GetWebURL() string {
	return b.WebUrl
}

type commit struct {
	ID             string     `json:"id"`
	ShortID        string     `json:"short_id"`
	Title          string     `json:"title"`
	CommittedDate  *time.Time `json:"committed_date"`
	AuthorName     string     `json:"author_name"`
	AuthorEmail    string     `json:"author_email"`
	CommitterName  string     `json:"committer_name"`
	CommitterEmail string     `json:"committer_email"`
	CreatedAt      *time.Time `json:"created_at"`
	Message        string     `json:"message"`
	ProjectID      int64      `json:"project_id"`
	WebURL         string     `json:"web_url"`
}

func (c *commit) GetID() string {
	return c.ID
}

func (c *commit) GetShortID() string {
	return c.ShortID
}

func (c *commit) GetTitle() string {
	return c.Title
}

func (c *commit) GetCommittedDate() *time.Time {
	return c.CommittedDate
}

func (c *commit) GetAuthorName() string {
	return c.AuthorName
}

func (c *commit) GetAuthorEmail() string {
	return c.AuthorEmail
}

func (c *commit) GetCommitterName() string {
	return c.CommitterName
}

func (c *commit) GetCommitterEmail() string {
	return c.CommitterEmail
}

func (c *commit) GetCreatedAt() *time.Time {
	return c.CreatedAt
}

func (c *commit) GetMessage() string {
	return c.Message
}

func (c *commit) GetProjectID() int64 {
	return c.ProjectID
}

func (c *commit) GetWebURL() string {
	return c.WebURL
}
