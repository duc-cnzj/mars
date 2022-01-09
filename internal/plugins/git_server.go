package plugins

import (
	"sync"
	"time"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
)

var gitServerOnce = sync.Once{}

type Status = string

const (
	StatusUnknown Status = "unknown"
	StatusSuccess        = "success"
	StatusFailed         = "failed"
	StatusRunning        = "running"
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

	return p.(GitServer)
}
