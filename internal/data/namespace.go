package data

import (
	"context"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/favorite"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/member"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/namespace"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/project"
	"github.com/duc-cnzj/mars/v6/internal/data/filters"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/samber/lo"
)

// toNamespace 把 ent.Namespace 转换为 biz.Namespace（nil 安全）。
// 创建者为内置超级管理员（biz.SuperAdminEmail）时，展示名替换为"超级管理员"。
func toNamespace(namespace *ent.Namespace) *biz.Namespace {
	if namespace == nil {
		return nil
	}
	cemail := namespace.CreatorEmail
	if cemail == biz.SuperAdminEmail {
		cemail = biz.SuperAdminName
	}

	return &biz.Namespace{
		ID:               namespace.ID,
		CreatedAt:        namespace.CreatedAt,
		UpdatedAt:        namespace.UpdatedAt,
		DeletedAt:        namespace.DeletedAt,
		Name:             namespace.Name,
		ImagePullSecrets: namespace.ImagePullSecrets,
		Description:      namespace.Description,
		Private:          namespace.Private,
		CreatorEmail:     cemail,
		Projects:         slice.Map(namespace.Edges.Projects, toProject),
		Favorites:        slice.Map(namespace.Edges.Favorites, toFavorite),
		Members:          slice.Map(namespace.Edges.Members, toMember),
	}
}

// toMember 把 ent.Member 转换为 biz.Member（nil 安全）。
func toMember(v *ent.Member) *biz.Member {
	if v == nil {
		return nil
	}
	return &biz.Member{
		ID:          v.ID,
		NamespaceID: v.NamespaceID,
		Email:       v.Email,
	}
}

// toFavorite 把 ent.Favorite 转换为 biz.Favorite（nil 安全）。
func toFavorite(v *ent.Favorite) *biz.Favorite {
	if v == nil {
		return nil
	}
	return &biz.Favorite{
		ID:          v.ID,
		NamespaceID: v.NamespaceID,
		Email:       v.Email,
	}
}

var _ biz.NamespaceRepo = (*namespaceRepo)(nil)

// namespaceRepo 是 biz.NamespaceRepo 的 data 实现：经 dataStore 访问 ent 客户端，
// 执行 namespace 及其成员/收藏/项目关系的读写。
type namespaceRepo struct {
	data     dataStore
	nsPrefix string
}

// Transfer 把 namespace 的创建者/归属转交给新邮箱：创建者已相同则跳过更新。
func (repo *namespaceRepo) Transfer(ctx context.Context, id int, email string) (out *biz.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/Transfer")
	defer func() { endSpan(span, err) }()
	ns, err := repo.data.DB().Namespace.Get(ctx, id)
	if err != nil {
		return nil, errs.Wrap(err, "transfer namespace")
	}
	if ns.CreatorEmail != email {
		ns, err = ns.Update().SetCreatorEmail(email).Save(ctx)
		if err != nil {
			return nil, errs.Wrap(err, "transfer namespace")
		}
	}
	return toNamespace(ns), nil
}

// SyncMembers 以 memberEmails 为最终名单同步 namespace 成员：事务内差量新增缺失成员、
// 删除已不在名单内的成员，随后返回最新 namespace。
func (repo *namespaceRepo) SyncMembers(ctx context.Context, namespaceID int, memberEmails []string) (out *biz.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/SyncMembers")
	defer func() { endSpan(span, err) }()
	if err := repo.data.WithTx(ctx, func(tx *ent.Tx) error {
		get, err := tx.Namespace.Query().WithMembers().Where(namespace.ID(namespaceID)).First(ctx)
		if err != nil {
			return err
		}
		del, add := lo.Difference(slice.Map(get.Edges.Members, func(v *ent.Member) string { return v.Email }), memberEmails)
		if len(add) > 0 {
			creates := make([]*ent.MemberCreate, 0, len(add))
			for _, addEmail := range add {
				creates = append(creates, tx.Member.Create().SetEmail(addEmail).SetNamespaceID(namespaceID))
			}
			if _, err := tx.Member.CreateBulk(creates...).Save(ctx); err != nil {
				return err
			}
		}
		// 按本 namespace 限定删除，避免误删其他 namespace 的同名成员行。
		if len(del) > 0 {
			if _, err := tx.Member.Delete().Where(member.NamespaceID(namespaceID), member.EmailIn(del...)).Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, errs.Wrap(err, "sync members")
	}
	return repo.Show(ctx, namespaceID)
}

// UpdatePrivate 切换 namespace 私有状态；转为公开（private=false）时清空全部成员。
func (repo *namespaceRepo) UpdatePrivate(ctx context.Context, namespaceID int, private bool) (out *biz.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/UpdatePrivate")
	defer func() { endSpan(span, err) }()
	if err := repo.data.WithTx(ctx, func(tx *ent.Tx) error {
		get, err := tx.Namespace.Get(ctx, namespaceID)
		if err != nil {
			return err
		}
		up := get.Update().
			SetPrivate(private)
		if !private {
			// 成员行已删光，无需再 ClearMembers（置 FK 为 NULL 影响 0 行）。
			if _, err = tx.Member.Delete().Where(member.NamespaceID(namespaceID)).Exec(ctx); err != nil {
				return err
			}
		}
		_, err = up.Save(ctx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, errs.Wrap(err, "update private")
	}
	return repo.Show(ctx, namespaceID)
}

// NewNamespaceRepo 构造 namespaceRepo：nsPrefix 来自配置，用于创建/查找时加前缀。
func NewNamespaceRepo(data dataStore) biz.NamespaceRepo {
	return &namespaceRepo{
		data:     data,
		nsPrefix: data.Config().NsPrefix,
	}
}

// List 按输入条件分页查询 namespace：支持名称模糊与收藏过滤；非管理员只可见
// 公开的、自己创建的、或自己是成员的私有 namespace。
func (repo *namespaceRepo) List(ctx context.Context, input *biz.ListNamespaceInput) (out []*biz.Namespace, pag *pagination.Pagination, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/List")
	defer func() { endSpan(span, err) }()
	query := repo.data.DB().Namespace.Query().
		Where(
			filters.IfNameLike(lo.FromPtr(input.Name)),
		)
	if !input.IsAdmin {
		query = query.Where(
			namespace.Or(
				namespace.And(
					namespace.HasMembersWith(member.Email(input.Email)),
					namespace.Private(true),
				),
				namespace.Private(false),
				namespace.CreatorEmail(input.Email),
			),
		)
	}

	if input.Favorite {
		query = query.Where(
			namespace.HasFavoritesWith(favorite.Email(input.Email)),
		)
	}

	all, err := query.Clone().
		Select(
			namespace.FieldID,
			namespace.FieldName,
			namespace.FieldDescription,
			namespace.FieldCreatedAt,
			namespace.FieldUpdatedAt,
			namespace.FieldCreatorEmail,
			namespace.FieldPrivate,
			namespace.FieldImagePullSecrets,
		).
		WithFavorites(func(query *ent.FavoriteQuery) {
			query.Where(favorite.Email(input.Email))
		}).
		WithMembers(func(query *ent.MemberQuery) {
			query.Select(member.FieldID, member.FieldEmail)
		}).
		WithProjects(
			func(query *ent.ProjectQuery) {
				query.Select(
					project.FieldID,
					project.FieldName,
					project.FieldDeployStatus,
					project.FieldNamespaceID,
					project.FieldCreatedAt,
					project.FieldUpdatedAt,
				)
			},
		).
		Offset(pagination.GetPageOffset(input.Page, input.PageSize)).
		Limit(int(input.PageSize)).
		All(ctx)
	if err != nil {
		return nil, nil, errs.Wrap(err, "list namespaces")
	}
	count, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, nil, errs.Wrap(err, "count namespaces")
	}
	return slice.Map(all, toNamespace), pagination.NewPagination(input.Page, input.PageSize, count), nil
}

// Create 创建 namespace：名称经 biz.GetNamespace 加 nsPrefix 前缀。
func (repo *namespaceRepo) Create(ctx context.Context, input *biz.CreateNamespaceInput) (out *biz.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/Create")
	defer func() { endSpan(span, err) }()
	save, err := repo.data.DB().Namespace.
		Create().
		SetName(biz.GetNamespace(input.Name, repo.nsPrefix)).
		SetImagePullSecrets(input.ImagePullSecrets).
		SetCreatorEmail(input.CreatorEmail).
		SetDescription(input.Description).
		Save(ctx)
	return toNamespace(save), errs.Wrap(err, "create namespace")
}

// Show 返回单个 namespace，附带项目（精简列）与成员列表。
func (repo *namespaceRepo) Show(ctx context.Context, id int) (out *biz.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/Show")
	defer func() { endSpan(span, err) }()
	first, err := repo.data.DB().Namespace.Query().
		WithProjects(func(query *ent.ProjectQuery) {
			query.Select(
				project.FieldID,
				project.FieldName,
				project.FieldNamespaceID,
			)
		}).
		WithMembers().
		Where(namespace.ID(id)).
		First(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "show namespace")
	}
	return toNamespace(first), nil
}

// Update 更新 namespace 的描述信息。
func (repo *namespaceRepo) Update(ctx context.Context, input *biz.UpdateNamespaceInput) (out *biz.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/Update")
	defer func() { endSpan(span, err) }()
	get, err := repo.data.DB().Namespace.Get(ctx, input.ID)
	if err != nil {
		return nil, errs.Wrap(err, "update namespace")
	}
	save, err := get.Update().SetDescription(input.Description).Save(ctx)
	return toNamespace(save), errs.Wrap(err, "update namespace")
}

// GetMarsNamespace 返回带 nsPrefix 前缀的完整 namespace 名称。
func (repo *namespaceRepo) GetMarsNamespace(name string) string {
	return biz.GetNamespace(name, repo.nsPrefix)
}

// FindByName 按名称（自动加 nsPrefix 前缀）精确查找 namespace。
func (repo *namespaceRepo) FindByName(ctx context.Context, name string) (out *biz.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/FindByName")
	defer func() { endSpan(span, err) }()
	first, err := repo.data.DB().Namespace.Query().Where(namespace.Name(biz.GetNamespace(name, repo.nsPrefix))).First(ctx)
	return toNamespace(first), errs.Wrap(err, "find namespace by name")
}

// ListAll 返回全部 namespace（含 ImagePullSecrets 列），cron 同步 imagePullSecrets
// 与 TLS 证书需全量遍历。
func (repo *namespaceRepo) ListAll(ctx context.Context) (out []*biz.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/ListAll")
	defer func() { endSpan(span, err) }()
	all, err := repo.data.DB().Namespace.Query().All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list all namespaces")
	}
	return slice.Map(all, toNamespace), nil
}

// UpdateImagePullSecrets 仅回写 namespace 的 imagePullSecrets 列表，
// cron 对账后把新增/清理后的 secret 名单持久化。
func (repo *namespaceRepo) UpdateImagePullSecrets(ctx context.Context, id int, secrets []string) (err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/UpdateImagePullSecrets")
	defer func() { endSpan(span, err) }()
	return errs.Wrap(repo.data.DB().Namespace.UpdateOneID(id).SetImagePullSecrets(secrets).Exec(ctx), "update image pull secrets")
}

// Delete 删除 namespace：若关联项目非空先删除项目，再在事务内删除 namespace。
func (repo *namespaceRepo) Delete(ctx context.Context, id int) (err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/Delete")
	defer func() { endSpan(span, err) }()
	first, err := repo.data.DB().Namespace.Query().WithProjects().Where(namespace.ID(id)).First(ctx)
	if err != nil {
		return errs.Wrap(err, "delete namespace")
	}
	return errs.Wrap(repo.data.WithTx(ctx, func(tx *ent.Tx) error {
		if len(first.Edges.Projects) > 0 {
			if _, err := tx.Project.
				Delete().
				Where(project.HasNamespaceWith(namespace.ID(id))).
				Exec(ctx); err != nil {
				return err
			}
		}
		return tx.Namespace.DeleteOneID(id).Exec(ctx)
	}), "delete namespace")
}

// Favorite 收藏/取消收藏 namespace：Favorite=true 幂等收藏，false 删除收藏。
func (repo *namespaceRepo) Favorite(ctx context.Context, input *biz.FavoriteNamespaceInput) (err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/Favorite")
	defer func() { endSpan(span, err) }()
	if !input.Favorite {
		_, err := repo.data.DB().Favorite.Delete().Where(favorite.NamespaceID(input.NamespaceID), favorite.Email(input.UserEmail)).Exec(ctx)
		return errs.Wrap(err, "favorite namespace")
	}

	exist, err := repo.data.DB().Favorite.Query().Where(favorite.NamespaceID(input.NamespaceID), favorite.Email(input.UserEmail)).Exist(ctx)
	if err != nil {
		return errs.Wrap(err, "favorite namespace")
	}
	if exist {
		return nil
	}
	return errs.Wrap(repo.data.DB().Favorite.Create().SetNamespaceID(input.NamespaceID).SetEmail(input.UserEmail).Exec(ctx), "favorite namespace")
}
