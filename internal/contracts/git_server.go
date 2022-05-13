package contracts

//go:generate mockgen -destination ../mock/mock_git_server_pipeline.go -package mock github.com/duc-cnzj/mars/internal/contracts PipelineInterface
//go:generate mockgen -destination ../mock/mock_git_server_commit.go -package mock github.com/duc-cnzj/mars/internal/contracts CommitInterface
//go:generate mockgen -destination ../mock/mock_git_server_project.go -package mock github.com/duc-cnzj/mars/internal/contracts ProjectInterface

import "time"

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
