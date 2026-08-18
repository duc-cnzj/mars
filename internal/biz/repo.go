package biz

import (
	"context"
	"errors"
	"fmt"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
)

// RepoBiz 收口仓库领域业务：增删改查、克隆与启用开关。
type RepoBiz interface {
	// All 返回全部仓库（可过滤）。
	All(ctx context.Context, in *AllRepoRequest) ([]*Repo, error)
	// List 分页列出仓库。
	List(ctx context.Context, in *ListRepoRequest) ([]*Repo, *pagination.Pagination, error)
	// Create 校验输入后创建仓库。
	Create(ctx context.Context, in *CreateRepoInput) (*Repo, error)
	// Get 按 id 查询仓库。
	Get(ctx context.Context, id int) (*Repo, error)
	// GetByGitProjectID 按 git 项目 ID 查询已启用的单个仓库；NotFound 由 errs 归类。
	GetByGitProjectID(ctx context.Context, projectID int32) (*Repo, error)
	// Show 按 id 查询仓库（携带关联项目）。
	Show(ctx context.Context, id int) (*Repo, error)
	// Update 校验输入后更新仓库（同名冲突校验见实现）。
	Update(ctx context.Context, in *UpdateRepoInput) (*Repo, error)
	// Delete 校验 id 后删除仓库。
	Delete(ctx context.Context, id int) error
	// Clone 克隆仓库为新仓库。
	Clone(ctx context.Context, input *CloneRepoInput) (*Repo, error)
	// ToggleEnabled 启用/禁用仓库。
	ToggleEnabled(ctx context.Context, id int, enabled bool) (*Repo, error)
}

type repoBiz struct {
	repoRepo RepoRepo
}

// NewRepoBiz 构造 repo biz。
func NewRepoBiz(repoRepo RepoRepo) RepoBiz {
	return &repoBiz{repoRepo: repoRepo}
}

// All 返回全部仓库（透传 repo）。
func (b *repoBiz) All(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
	return b.repoRepo.All(ctx, in)
}

// List 分页列出仓库（透传 repo）。
func (b *repoBiz) List(ctx context.Context, in *ListRepoRequest) ([]*Repo, *pagination.Pagination, error) {
	return b.repoRepo.List(ctx, in)
}

// ensureRepoNameFree 校验仓库名称未被占用（Create/Clone 前置）。
// GetByName 返回 NotFound 视为名称空闲；查询失败（DB 抖动等）原样透传；
// 名称已存在则按 op 构造参数错误。
func (b *repoBiz) ensureRepoNameFree(ctx context.Context, name, op string) error {
	_, err := b.repoRepo.GetByName(ctx, name)
	switch {
	case err == nil:
		return errs.WrapInvalidArgument(errors.New("repo 名称已经存在"), op)
	case errs.IsNotFound(err):
		return nil
	default:
		return err
	}
}

// Create 创建仓库，先校验名称唯一（名称唯一是业务规则，不依赖 DB 唯一约束）。
func (b *repoBiz) Create(ctx context.Context, in *CreateRepoInput) (*Repo, error) {
	if in == nil {
		return nil, errs.WrapInvalidArgument(errors.New("repo 不能为空"), "create repo")
	}
	if in.Name == "" {
		return nil, errs.WrapInvalidArgument(errors.New("repo 名称不能为空"), "create repo")
	}
	if err := b.ensureRepoNameFree(ctx, in.Name, "create repo"); err != nil {
		return nil, err
	}
	return b.repoRepo.Create(ctx, in)
}

// Get 按 id 查询仓库（透传 repo）。
func (b *repoBiz) Get(ctx context.Context, id int) (*Repo, error) {
	return b.repoRepo.Get(ctx, id)
}

// GetByGitProjectID 按 git 项目 ID 查询仓库（透传 repo）。
func (b *repoBiz) GetByGitProjectID(ctx context.Context, projectID int32) (*Repo, error) {
	return b.repoRepo.GetByGitProjectID(ctx, projectID)
}

// Show 按 id 查询仓库（透传 repo）。
func (b *repoBiz) Show(ctx context.Context, id int) (*Repo, error) {
	return b.repoRepo.Show(ctx, id)
}

// Update 更新仓库，先校验名称唯一（排除自身）与"有项目不可改名"两条业务规则。
func (b *repoBiz) Update(ctx context.Context, in *UpdateRepoInput) (*Repo, error) {
	if in.ID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("repo id 不能小于等于 0"), "update repo")
	}
	found, err := b.repoRepo.GetByName(ctx, in.Name)
	if err != nil && !errs.IsNotFound(err) {
		return nil, err
	}
	// 同名仓库存在且不是自身时，名称冲突（guard clause 后主路径不缩进）。
	if err == nil && found.ID != int(in.ID) {
		return nil, errs.WrapInvalidArgument(errors.New("repo 名称已经存在"), "update repo")
	}
	show, err := b.repoRepo.Show(ctx, int(in.ID))
	if err != nil {
		return nil, err
	}
	if show.Name != in.Name && len(show.Projects) > 0 {
		return nil, errs.WrapInvalidArgument(fmt.Errorf("repo 下面还有 %d 个项目，不能修改名称", len(show.Projects)), "update repo")
	}
	return b.repoRepo.Update(ctx, in)
}

// Delete 删除仓库，先验证是否有已关联的项目。
func (b *repoBiz) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return errs.WrapInvalidArgument(errors.New("repo id 不能小于等于 0"), "delete repo")
	}
	show, err := b.repoRepo.Show(ctx, id)
	if err != nil {
		return err
	}
	if len(show.Projects) > 0 {
		return errs.WrapInvalidArgument(fmt.Errorf("repo 下面还有 %d 个项目，不能删除", len(show.Projects)), "delete repo")
	}
	return b.repoRepo.Delete(ctx, id)
}

// Clone 克隆仓库，先校验目标名称唯一。
func (b *repoBiz) Clone(ctx context.Context, input *CloneRepoInput) (*Repo, error) {
	if input.ID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("repo id 不能小于等于 0"), "clone repo")
	}
	if input.Name == "" {
		return nil, errs.WrapInvalidArgument(errors.New("repo 名称不能为空"), "clone repo")
	}
	if err := b.ensureRepoNameFree(ctx, input.Name, "clone repo"); err != nil {
		return nil, err
	}
	return b.repoRepo.Clone(ctx, input)
}

// ToggleEnabled 启用/禁用仓库，禁用时需验证无关联项目。
func (b *repoBiz) ToggleEnabled(ctx context.Context, id int, enabled bool) (*Repo, error) {
	if id <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("repo id 不能小于等于 0"), "toggle enabled")
	}
	get, err := b.repoRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !enabled && get.Enabled {
		show, err := b.repoRepo.Show(ctx, id)
		if err != nil {
			return nil, err
		}
		if len(show.Projects) > 0 {
			return nil, errs.WrapInvalidArgument(fmt.Errorf("repo 下面还有 %d 个项目，不能禁用", len(show.Projects)), "toggle enabled")
		}
	}
	return b.repoRepo.ToggleEnabled(ctx, id, enabled)
}

// RepoRepo 是仓库（存储库）仓库端口。
type RepoRepo interface {
	// All 返回全部仓库。
	All(ctx context.Context, in *AllRepoRequest) ([]*Repo, error)
	// List 分页列出仓库。
	List(ctx context.Context, in *ListRepoRequest) ([]*Repo, *pagination.Pagination, error)
	// Create 创建仓库。
	Create(ctx context.Context, in *CreateRepoInput) (*Repo, error)
	// Get 按 id 查询仓库。
	Get(ctx context.Context, id int) (*Repo, error)
	// GetByName 按名称查询仓库。
	GetByName(ctx context.Context, name string) (*Repo, error)
	// GetByGitProjectID 按 git 项目 ID 查询已启用的单个仓库；NotFound 由 errs 归类。
	GetByGitProjectID(ctx context.Context, projectID int32) (*Repo, error)
	// Show 按 id 查询仓库（携带关联项目）。
	Show(ctx context.Context, id int) (*Repo, error)
	// Update 更新仓库。
	Update(ctx context.Context, in *UpdateRepoInput) (*Repo, error)
	// Delete 删除仓库。
	Delete(ctx context.Context, id int) error
	// Clone 克隆仓库为新仓库。
	Clone(ctx context.Context, input *CloneRepoInput) (*Repo, error)
	// ToggleEnabled 启用/禁用仓库。
	ToggleEnabled(ctx context.Context, id int, enabled bool) (*Repo, error)
}
