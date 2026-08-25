package biz

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
)

// repoNamePattern 是 repo 名称的合法字符集，对齐 proto validate 与 ent schema。
// 导入请求复用 RepoModel 模型消息（无逐字段 validate 规则），故在 biz 层自行校验 name。
var repoNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

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
	// Import 批量导入仓库（按 name 幂等覆盖），返回新建/更新计数。
	Import(ctx context.Context, items []*ImportRepoItem) (created, updated int, err error)
	// PreviewImport 干跑导入：同 Import 的校验与计数判定，但不落库（导入前预览用）。
	PreviewImport(ctx context.Context, items []*ImportRepoItem) (created, updated int, err error)
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

// validateImportItems 整体校验导入条目：任一为空、name 非法或重复即返回参数错误，保证校验先行、零部分变更。
// 调用方（Import/PreviewImport）已对空切片短路，此处空切片自然返回 nil。
func validateImportItems(items []*ImportRepoItem) error {
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item == nil {
			return errs.WrapInvalidArgument(errors.New("导入数据不能为空"), "import repo")
		}
		if !repoNamePattern.MatchString(item.Name) {
			return errs.WrapInvalidArgument(fmt.Errorf("repo 名称 %q 不合法", item.Name), "import repo")
		}
		if _, dup := seen[item.Name]; dup {
			return errs.WrapInvalidArgument(fmt.Errorf("repo 名称 %q 在导入文件中重复", item.Name), "import repo")
		}
		seen[item.Name] = struct{}{}
	}
	return nil
}

// Import 批量导入仓库：先整体校验所有 name（任一非法或重复即返回、不做任何变更），
// 再委托 data 层在单个事务内按 name 幂等应用——已存在走 Update（保留原 ID 与关联项目），
// 不存在走 Create；任一条目落库失败整体回滚，不留下部分导入结果。
func (b *repoBiz) Import(ctx context.Context, items []*ImportRepoItem) (created, updated int, err error) {
	if len(items) == 0 {
		return 0, 0, nil
	}
	if err := validateImportItems(items); err != nil {
		return 0, 0, err
	}
	return b.repoRepo.Import(ctx, items)
}

// PreviewImport 干跑导入：与 Import 共用整体校验，委托 data 层只做计划不计落库（导入前预览用）。
func (b *repoBiz) PreviewImport(ctx context.Context, items []*ImportRepoItem) (created, updated int, err error) {
	if len(items) == 0 {
		return 0, 0, nil
	}
	if err := validateImportItems(items); err != nil {
		return 0, 0, err
	}
	return b.repoRepo.PreviewImport(ctx, items)
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
	// Import 批量导入仓库（按名称幂等，事务性整体应用）。
	Import(ctx context.Context, items []*ImportRepoItem) (created, updated int, err error)
	// PreviewImport 干跑导入：同 Import 的校验与计数判定，但不落库（导入前预览用）。
	PreviewImport(ctx context.Context, items []*ImportRepoItem) (created, updated int, err error)
}
