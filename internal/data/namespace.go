package data

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"

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

// UpdateConfig 单事务原子更新 namespace 配置（描述/私有/成员/转让管理员）：
// 合并 UpdatePrivate/SyncMembers/Transfer/UpdateDesc 的既有业务规则。顺序为先写
// namespace 字段（描述/私有/转让），再以 final 名单差量同步成员——私有转公开时先
// 清空成员，随后若给定新名单则按名单重建。Emails 非 nil（含空）表示成员需全量同步。
func (repo *namespaceRepo) UpdateConfig(ctx context.Context, input *biz.UpdateConfigInput) (out *biz.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/UpdateConfig")
	defer func() { endSpan(span, err) }()
	if err := repo.data.WithTx(ctx, func(tx *ent.Tx) error {
		get, err := tx.Namespace.Query().Where(namespace.ID(input.ID)).First(ctx)
		if err != nil {
			return err
		}
		up := tx.Namespace.UpdateOneID(input.ID)
		if input.Description != nil {
			up = up.SetDescription(*input.Description)
		}
		if input.Private != nil {
			up = up.SetPrivate(*input.Private)
			// 转公开清空全部成员（对齐 UpdatePrivate 规则）；随后若给定新名单则按名单重建。
			if !*input.Private {
				if _, err = tx.Member.Delete().Where(member.NamespaceID(input.ID)).Exec(ctx); err != nil {
					return err
				}
			}
		}
		if input.NewAdminEmail != "" && input.NewAdminEmail != get.CreatorEmail {
			up = up.SetCreatorEmail(input.NewAdminEmail)
		}
		if _, err = up.Save(ctx); err != nil {
			return err
		}
		// 成员差量同步（对齐 SyncMembers：Emails 非 nil 即全量同步，含清空）。
		if input.Emails != nil {
			current, err := tx.Member.Query().Where(member.NamespaceID(input.ID)).Select(member.FieldEmail).Strings(ctx)
			if err != nil {
				return err
			}
			del, add := lo.Difference(current, input.Emails)
			if len(add) > 0 {
				creates := make([]*ent.MemberCreate, 0, len(add))
				for _, addEmail := range add {
					creates = append(creates, tx.Member.Create().SetEmail(addEmail).SetNamespaceID(input.ID))
				}
				if _, err := tx.Member.CreateBulk(creates...).Save(ctx); err != nil {
					return err
				}
			}
			if len(del) > 0 {
				if _, err := tx.Member.Delete().Where(member.NamespaceID(input.ID), member.EmailIn(del...)).Exec(ctx); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, errs.Wrap(err, "update config")
	}
	return repo.Show(ctx, input.ID)
}

// NewNamespaceRepo 构造 namespaceRepo：nsPrefix 来自配置，用于创建/查找时加前缀。
func NewNamespaceRepo(data dataStore) biz.NamespaceRepo {
	return &namespaceRepo{
		data:     data,
		nsPrefix: data.Config().NsPrefix,
	}
}

// adminNamespaceBaseQuery 构造管理员视角命名空间的过滤条件：名称模糊 + 管理后台搜索
// （匹配空间名/创建者邮箱）+ 只看私有。List 分页与 ListAdminPage 共用，保证同一份
// 过滤语义单一来源；边装配由 withAdminEdges 按需叠加（Count 走本 base 保持无边，避免
// 边 JOIN 放大计数）。
func (repo *namespaceRepo) adminNamespaceBaseQuery(ctx context.Context, input *biz.ListNamespaceInput) *ent.NamespaceQuery {
	return repo.data.DB().Namespace.Query().
		Where(
			filters.IfNameLike(lo.FromPtr(input.Name)),
			// 管理后台搜索：模糊匹配空间名或创建者邮箱，空串不过滤。
			filters.If(func(s string) bool {
				return s != ""
			}, func(t string) func(*sql.Selector) {
				return namespace.Or(namespace.NameContains(t), namespace.CreatorEmailContains(t))
			})(input.Search),
			// 管理后台私有过滤：只看私有空间。
			filters.If(func(b bool) bool {
				return b
			}, func(bool) func(*sql.Selector) {
				return namespace.Private(true)
			})(input.PrivateOnly),
		)
}

// withAdminEdges 给管理查询装配全量边（收藏/成员/项目），供列表展示下钻：成员列表
// （含邮箱）、项目（含 UpdatedAt 供活跃度聚合）、关注标记。email 为当前用户，关注列表
// 收敛为该用户行（admin 场景为空串，不匹配任何关注行，语义等价无边）。
func withAdminEdges(query *ent.NamespaceQuery, email string) *ent.NamespaceQuery {
	return query.
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
			query.Where(favorite.Email(email))
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
		)
}

// List 按输入条件分页查询 namespace：支持名称模糊与收藏过滤；非管理员只可见
// 公开的、自己创建的、或自己是成员的私有 namespace。
func (repo *namespaceRepo) List(ctx context.Context, input *biz.ListNamespaceInput) (out []*biz.Namespace, pag *pagination.Pagination, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/List")
	defer func() { endSpan(span, err) }()
	query := repo.adminNamespaceBaseQuery(ctx, input)
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
		// 关注列表按用户自定义排序：sort_order 升序（同值按 namespace id 兜底稳定序）。
		// LEFT JOIN favorites 取排序值；JOIN 带 email 条件收敛为该用户行（等价 INNER JOIN），
		// (email,namespace_id) 唯一索引保证每空间至多命中一行，分页/计数不被放大。
		query = query.Order(
			func(s *sql.Selector) {
				t := sql.Table(favorite.Table)
				s.LeftJoin(t).
					On(s.C(namespace.FieldID), t.C(favorite.FieldNamespaceID)).
					Where(sql.EQ(t.C(favorite.FieldEmail), input.Email)).
					OrderBy(t.C(favorite.FieldSortOrder), s.C(namespace.FieldID))
			},
		)
	}

	all, err := withAdminEdges(query.Clone(), input.Email).
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
			// 项目列与 List 路径保持一致：列表卡片要读 deploy_status 渲染部署状态，
			// 漏选会把 deployStatus 落成零值（StatusUnknown）——部署成功后
			// 前端 refreshNamespace 用 show 原地替换卡片，项目状态就"变没"了。
			query.Select(
				project.FieldID,
				project.FieldName,
				project.FieldDeployStatus,
				project.FieldNamespaceID,
				project.FieldCreatedAt,
				project.FieldUpdatedAt,
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

// namespaceLivenessPred 命名空间活跃度分类 SQL 谓词：分类键是「空间下项目 UpdatedAt 最大值」
// 的跨表聚合，SQL 侧以 ent 原生 EXISTS 谓词等价表达（MAX(updated_at) > X ⟺ EXISTS(项目
// updated_at > X)），免建 correlated subquery 新地基；边界由分类基准 now 推导（活跃=最大
// updated_at > now-31d；僵尸=<= now-90d；低活跃=两者之间），与 biz.classifyLiveness 阈值
// 数学等价。无项目（从未活跃 → 零值时间 → 僵尸）以 NOT EXISTS(任意项目) 表达。
// 非法 liveness 值返回恒假谓词，复现旧逻辑「无行命中非法分类」的空列表语义。
func namespaceLivenessPred(liveness string, now time.Time) func(*sql.Selector) {
	active, zombie := livenessBoundaries(now)
	// hasRecent = 是否存在项目 updated_at > boundary。
	hasRecent := func(boundary time.Time) func(*sql.Selector) {
		return namespace.HasProjectsWith(project.UpdatedAtGT(boundary))
	}
	switch liveness {
	case "active":
		return hasRecent(active)
	case "zombie":
		// 无项目 或 最近活跃已过僵尸边界。
		return namespace.Or(namespace.Not(namespace.HasProjects()), namespace.Not(hasRecent(zombie)))
	case "dormant":
		return namespace.And(hasRecent(zombie), namespace.Not(hasRecent(active)))
	default:
		return func(s *sql.Selector) { s.Where(sql.False()) }
	}
}

// ListAdminPage 分页列出管理员视角的命名空间（真 SQL 分页）：分类过滤/统计/分页全部下沉
// SQL，stats 基于 search 命中全量（无 edges 的 base 计数，避免 JOIN 放大），count 为分类
// 过滤后总数（无过滤 = total）。行级 lastActiveAt/活跃度仍由 biz 依已加载的项目边计算，
// 故分页行保留 withAdminEdges 全量边装配（成员/项目/关注，供前端下钻）。
// 排序按 id 升序（现状自然序），保证 LIMIT/OFFSET 翻页确定性。
func (repo *namespaceRepo) ListAdminPage(ctx context.Context, query *biz.AdminListPageQuery) (page *biz.AdminListPageResult, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/ListAdminPage")
	defer func() { endSpan(span, err) }()
	base := repo.adminNamespaceBaseQuery(ctx, &biz.ListNamespaceInput{
		Search:      query.Search,
		PrivateOnly: query.PrivateOnly,
	})
	total, err := base.Clone().Count(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "count admin namespaces total")
	}
	active, err := base.Clone().Where(namespaceLivenessPred("active", query.Now)).Count(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "count admin namespaces active")
	}
	dormant, err := base.Clone().Where(namespaceLivenessPred("dormant", query.Now)).Count(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "count admin namespaces dormant")
	}
	zombie, err := base.Clone().Where(namespaceLivenessPred("zombie", query.Now)).Count(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "count admin namespaces zombie")
	}
	count := total
	if query.Liveness != "" {
		filtered, err := base.Clone().Where(namespaceLivenessPred(query.Liveness, query.Now)).Count(ctx)
		if err != nil {
			return nil, errs.Wrap(err, "count admin namespaces filtered")
		}
		count = filtered
	}
	rows := base.Clone()
	if query.Liveness != "" {
		rows = rows.Where(namespaceLivenessPred(query.Liveness, query.Now))
	}
	// Where/Order 在 *NamespaceQuery 上叠加（Select 裁剪由 withAdminEdges 收尾），
	// 对齐 namespaceRepo.List 的既有写法。
	rows = rows.Order(ent.Asc(namespace.FieldID))
	all, err := withAdminEdges(rows, "").
		Offset(pagination.GetPageOffset(query.Page, query.PageSize)).
		Limit(int(query.PageSize)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list admin namespaces page")
	}
	return &biz.AdminListPageResult{
		Namespaces: slice.Map(all, toNamespace),
		Count:      count,
		Stats:      biz.AdminLivenessStats{Total: total, Active: active, Dormant: dormant, Zombie: zombie},
	}, nil
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
	// 新关注追加末尾：sort_order = 该用户现有最大 +1（无历史关注则 0，排最前）。
	// 先挡掉非 NotFound 的查询错误（早退，避免 if/else 嵌套），NotFound 视为无历史关注。
	last, err := repo.data.DB().Favorite.Query().
		Where(favorite.Email(input.UserEmail)).
		Order(ent.Desc(favorite.FieldSortOrder)).
		First(ctx)
	// 非 NotFound 的查询错误（DB 抖动等）：防御分支，SQLite 无法确定性触发。
	if err != nil && !ent.IsNotFound(err) {
		return errs.Wrap(err, "favorite namespace")
	}
	sortOrder := 0
	if last != nil {
		sortOrder = last.SortOrder + 1
	}
	return errs.Wrap(repo.data.DB().Favorite.Create().
		SetNamespaceID(input.NamespaceID).
		SetEmail(input.UserEmail).
		SetSortOrder(sortOrder).
		Exec(ctx), "favorite namespace")
}

// FavoriteSort 把 firstID 关注空间移动到 secondID 所在位置，中间元素 sort_order 整体顺移。
// 两个 id 必须都是该用户的关注空间；相同 id（firstID==secondID）视为无效请求直接拒绝，
// 避免半程移动静默破坏排序。事务内按 email+namespace_id 精确更新，越权写他人空间天然被隔离。
func (repo *namespaceRepo) FavoriteSort(ctx context.Context, email string, firstID, secondID int) (err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/FavoriteSort")
	defer func() { endSpan(span, err) }()
	return errs.Wrap(repo.data.WithTx(ctx, func(tx *ent.Tx) error {
		if firstID == secondID {
			return errs.WrapInvalidArgument(fmt.Errorf("两个空间 id 不能相同"), "favorite sort")
		}
		favs, err := tx.Favorite.Query().
			Where(favorite.Email(email), favorite.NamespaceIDIn(firstID, secondID)).
			All(ctx)
		// DB 查询失败（防御分支：SQLite 无法确定性触发）。
		if err != nil {
			return err
		}
		if len(favs) != 2 {
			return errs.WrapInvalidArgument(fmt.Errorf("关注列表必须同时包含这两个空间（实际命中 %d 个）", len(favs)), "favorite sort")
		}
		var orderA, orderB int
		for _, f := range favs {
			// 命中 firstID 即记录并继续，剩余必为 secondID（guard clause 替代 else）。
			if f.NamespaceID == firstID {
				orderA = f.SortOrder
				continue
			}
			orderB = f.SortOrder
		}
		// 两空间 sort_order 相同（历史迁移全 0 / 手工改库重复序）：区间顺移缺位置信息，
		// 先在事务内把该用户关注按 (sort_order, id) 稳定序重排为 0..N，再重读两空间落位。
		// 懒修复随首次拖拽触发，不随系统启动扫描全表。
		if orderA == orderB {
			if err := renumberFavoriteSortOrders(ctx, tx, email); err != nil {
				return err
			}
			fresh, err := tx.Favorite.Query().
				Where(favorite.Email(email), favorite.NamespaceIDIn(firstID, secondID)).
				All(ctx)
			if err != nil {
				return err
			}
			for _, f := range fresh {
				if f.NamespaceID == firstID {
					orderA = f.SortOrder
					continue
				}
				orderB = f.SortOrder
			}
		}
		// 区间顺移统一表达，避免 if/else 双分支重复 UPDATE：此时必满足 orderA != orderB，
		// 前移（A<B）区间 [A+1, B] 减 1，后移（A>B）区间 [B, A-1] 加 1，方向由 delta 表达。
		lower, upper, delta := orderA+1, orderB, -1
		if orderA > orderB {
			lower, upper, delta = orderB, orderA-1, 1
		}
		if err := tx.Favorite.
			Update().
			Where(favorite.Email(email), favorite.SortOrderGTE(lower), favorite.SortOrderLTE(upper)).
			AddSortOrder(delta).
			Exec(ctx); err != nil {
			// DB 写入失败（防御分支：SQLite 无法确定性触发）。
			return err
		}
		return tx.Favorite.
			Update().
			Where(favorite.Email(email), favorite.NamespaceID(firstID)).
			SetSortOrder(orderB).
			Exec(ctx)
	}), "favorite sort")
}

// renumberFavoriteSortOrders 把该用户全部关注按 (sort_order, id) 稳定序重排为 0..N。
// 仅在 FavoriteSort 发现两空间 sort_order 相同（历史迁移全 0 / 手工改库重复序，区间顺移
// 缺位置信息）时懒触发，事务内幂等执行；不随系统启动扫描全表。
func renumberFavoriteSortOrders(ctx context.Context, tx *ent.Tx, email string) error {
	userFavs, err := tx.Favorite.Query().
		Where(favorite.Email(email)).
		Order(ent.Asc(favorite.FieldSortOrder), ent.Asc(favorite.FieldID)).
		All(ctx)
	if err != nil {
		return err
	}
	for i, f := range userFavs {
		if f.SortOrder == i {
			continue
		}
		if err := tx.Favorite.UpdateOneID(f.ID).SetSortOrder(i).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
