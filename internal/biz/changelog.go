package biz

import (
	"context"
	"errors"

	"github.com/duc-cnzj/mars/v6/internal/errs"
)

// ChangelogBiz 收口项目变更记录业务：查询最近记录与创建记录。
type ChangelogBiz interface {
	// FindLastChangelogsByProjectID 查询项目最近一批变更记录。
	FindLastChangelogsByProjectID(ctx context.Context, input *FindLastChangelogsByProjectIDChangeLogInput) ([]*Changelog, error)
	// Create 创建一条变更记录。
	Create(ctx context.Context, input *CreateChangeLogInput) (*Changelog, error)
	// FindLastChangeByProjectID 查询项目最近一条变更记录。
	FindLastChangeByProjectID(ctx context.Context, projectID int) (*Changelog, error)
}

type changelogBiz struct {
	clRepo ChangelogRepo
}

// NewChangelogBiz 构造 changelog biz。
func NewChangelogBiz(clRepo ChangelogRepo) ChangelogBiz {
	return &changelogBiz{clRepo: clRepo}
}

// FindLastChangelogsByProjectID 查询项目的最近变更记录列表（透传 repo）。
func (c *changelogBiz) FindLastChangelogsByProjectID(ctx context.Context, input *FindLastChangelogsByProjectIDChangeLogInput) ([]*Changelog, error) {
	return c.clRepo.FindLastChangelogsByProjectID(ctx, input)
}

// Create 校验输入后创建变更记录。
func (c *changelogBiz) Create(ctx context.Context, input *CreateChangeLogInput) (*Changelog, error) {
	if input == nil || input.ProjectID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("changelog 不能为空或 project id 不能小于等于 0"), "create changelog")
	}
	return c.clRepo.Create(ctx, input)
}

// FindLastChangeByProjectID 查询项目的最近一条变更记录（透传 repo）。
func (c *changelogBiz) FindLastChangeByProjectID(ctx context.Context, projectID int) (*Changelog, error) {
	return c.clRepo.FindLastChangeByProjectID(ctx, projectID)
}

// ChangelogRepo 是项目变更记录仓库端口。
type ChangelogRepo interface {
	// FindLastChangelogsByProjectID 查询项目最近一批变更记录。
	FindLastChangelogsByProjectID(ctx context.Context, input *FindLastChangelogsByProjectIDChangeLogInput) ([]*Changelog, error)
	// Create 创建一条变更记录。
	Create(ctx context.Context, input *CreateChangeLogInput) (*Changelog, error)
	// FindLastChangeByProjectID 查询项目最近一条变更记录。
	FindLastChangeByProjectID(ctx context.Context, projectID int) (*Changelog, error)
	// CountByProjectIDs 按项目 ID 集合聚合各项目变更记录数（GROUP BY project_id）。
	CountByProjectIDs(ctx context.Context, ids ...int) (map[int]int, error)
}
