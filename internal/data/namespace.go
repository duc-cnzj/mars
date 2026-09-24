package data

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/favorite"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/member"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/namespace"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/project"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
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
// 过滤语义单一来源；边装配由 withAdminEdges 按需叠加，而计数另在只带过滤条件的 query 上取
// ——COUNT 会原样带上行查询挂的 Order/Select 修饰符（namespaceRepo.List 的 favorites 排序就是
// LEFT JOIN，见该函数内注），与行查询分开取才能保证计数只反映过滤条件。
func (repo *namespaceRepo) adminNamespaceBaseQuery(input *biz.ListNamespaceInput) *ent.NamespaceQuery {
	return repo.data.DB().Namespace.Query().
		Where(
			// 名称模糊：匹配空间名，**或**该空间下任一项目名——首页搜索框输入项目名
			// 同样能定位到它所属的空间（用户往往记得项目名而记不住空间名）。
			// 空间名一侧保持原有的 Contains 语义不变，仅 OR 叠加项目名条件。
			// ⚠️ 子查询必须显式带 project.DeletedAtIsNil()：SoftDeleteMixin 靠 ent
			// Interceptor 注入 deleted_at IS NULL，而 Interceptor 只遍历生成的查询图，
			// **不会**进入 Has*With 产生的裸 sql.Selector 子查询；漏掉即让已软删项目
			// 仍能被搜到（回归见 TestNamespaceRepo_List_NameMatchesProject_SoftDeletedProject）。
			filters.If(func(s string) bool {
				return s != ""
			}, func(t string) func(*sql.Selector) {
				return namespace.Or(
					namespace.NameContains(t),
					namespace.HasProjectsWith(project.DeletedAtIsNil(), project.NameContainsFold(t)),
				)
			})(lo.FromPtr(input.Name)),
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

// adminNamespaceColumns 返回管理端命名空间列表共用的列裁剪清单（在册列表与已删除列表的
// 唯一事实来源，模型加字段只改这里）：只取列表要展示的列，避免拉大列。恢复页额外需要的
// namespace.FieldDeletedAt **不**入本表——在册查询选它无意义——由 withAdminDeletedEdges
// 自行 append。
func adminNamespaceColumns() []string {
	return []string{
		namespace.FieldID,
		namespace.FieldName,
		namespace.FieldDescription,
		namespace.FieldCreatedAt,
		namespace.FieldUpdatedAt,
		namespace.FieldCreatorEmail,
		namespace.FieldPrivate,
		namespace.FieldImagePullSecrets,
	}
}

// withAdminEdges 给管理查询装配全量边（收藏/成员/项目），供列表展示下钻：成员列表
// （含邮箱）、项目（含 UpdatedAt 供活跃度聚合）、关注标记。email 为当前用户，关注列表
// 收敛为该用户行（admin 场景为空串，不匹配任何关注行，语义等价无边）。
func withAdminEdges(query *ent.NamespaceQuery, email string) *ent.NamespaceQuery {
	return query.
		Select(adminNamespaceColumns()...).
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

// withAdminDeletedEdges 给「已删除空间」查询装配边与列裁剪，供 ListAdminDeletedPage 使用。
//
// 与 withAdminEdges 的两处差异都是刻意的：
//  1. 列裁剪在 adminNamespaceColumns 的基础上多带 namespace.FieldDeletedAt——恢复页要展示
//     「何时被删」并按它倒序，而 toNamespace 只有在列被 SELECT 出来时才能读到 DeletedAt
//     （否则静默为 nil，UI 显示不出删除时间）。
//  2. 项目边额外用 project.DeletedWithNamespace(true) 收敛到「随空间一起被删」的那一批，
//     即 RestoreDeleted 实际会恢复的集合——前端展示的项目数因此就是恢复后的项目数，不多不少
//     （用户早先单独删除的项目 flag=false，不会被恢复，也不该出现在这个计数里）。
//     不装配 favorites/members 边：空间已删除，收藏与成员列表无展示价值，省掉两条无用查询。
//
// 不复用 withAdminEdges 的原因：ent 的 WithProjects 是**整体替换**（内部 `_q.withProjects =
// query`），二次调用会丢掉前一次的列裁剪；而 Select 又返回另一种类型（*NamespaceSelect），
// 链式叠加会引入难以察觉的类型/顺序耦合。列清单则共用 adminNamespaceColumns 单一来源，
// 不存在两份手工清单漂移的问题。
func withAdminDeletedEdges(query *ent.NamespaceQuery) *ent.NamespaceQuery {
	// append 的目标是本次调用现分配的切片（adminNamespaceColumns 每次返回新切片），
	// 不会被别处共享，故就地追加安全。
	cols := append(adminNamespaceColumns(), namespace.FieldDeletedAt)
	return query.
		Select(cols...).
		WithProjects(func(query *ent.ProjectQuery) {
			query.
				Where(project.DeletedWithNamespace(true)).
				Select(
					project.FieldID,
					project.FieldName,
					project.FieldDeployStatus,
					project.FieldNamespaceID,
					project.FieldCreatedAt,
					project.FieldUpdatedAt,
				)
		})
}

// List 按输入条件分页查询 namespace：支持名称模糊与收藏过滤；非管理员只可见
// 公开的、自己创建的、或自己是成员的私有 namespace。
func (repo *namespaceRepo) List(ctx context.Context, input *biz.ListNamespaceInput) (out []*biz.Namespace, pag *pagination.Pagination, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/List")
	defer func() { endSpan(span, err) }()
	query := repo.adminNamespaceBaseQuery(input)
	if !input.IsAdmin {
		query = query.Where(
			namespace.Or(
				namespace.And(
					// ⚠️ member.DeletedAtIsNil() 必须显式带：成员被移出空间走 Member.Delete()
					// → SoftDeleteMixin 钩子转软删，而 HasMembersWith 的裸 sql.Selector 子查询
					// 不被 ent Interceptor 覆盖；漏掉即让「被移出者仍能看见私有空间」（越权）。
					namespace.HasMembersWith(member.DeletedAtIsNil(), member.Email(input.Email)),
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

// FindByName 按名称（自动加 nsPrefix 前缀）精确查找 namespace，并预加载成员列表。
//
// 必须 WithMembers：按名字寻址的访问门卫（RequireNamespaceAccessByName）用它的返回值
// 做权限判定，而私有空间的成员判定读的正是 ns.Members（biz/access.go CanAccessNamespace）。
// 漏加载会让 Members 恒空——私有空间的普通成员会被误判成无权访问（403），而其按 ID
// 寻址的孪生入口（RequireNamespaceAccessByID → Show，本文件已 WithMembers）却能通过。
// 预加载只补边、不过滤行，Create 的全局预查（不感知权限）语义不受影响。
func (repo *namespaceRepo) FindByName(ctx context.Context, name string) (out *biz.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/FindByName")
	defer func() { endSpan(span, err) }()
	first, err := repo.data.DB().Namespace.Query().
		WithMembers().
		Where(namespace.Name(biz.GetNamespace(name, repo.nsPrefix))).
		First(ctx)
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
	// hasRecent = 是否存在「存活且」updated_at > boundary 的项目。
	// ⚠️ project.DeletedAtIsNil() 必须显式带：SoftDeleteMixin 靠 ent Interceptor 注入
	// deleted_at IS NULL，而 Interceptor 不覆盖 Has*With 的裸 sql.Selector 子查询；漏掉即把
	// 「项目已被软删」的空间误判为活跃（回归见
	// TestNamespaceRepo_Liveness_SoftDeletedProjectsNotActive）。
	hasRecent := func(boundary time.Time) func(*sql.Selector) {
		return namespace.HasProjectsWith(project.DeletedAtIsNil(), project.UpdatedAtGT(boundary))
	}
	switch liveness {
	case "active":
		return hasRecent(active)
	case "zombie":
		// 无（存活）项目 或 最近活跃已过僵尸边界。同样补软删过滤，与 hasRecent 语义对齐。
		return namespace.Or(
			namespace.Not(namespace.HasProjectsWith(project.DeletedAtIsNil())),
			namespace.Not(hasRecent(zombie)),
		)
	case "dormant":
		return namespace.And(hasRecent(zombie), namespace.Not(hasRecent(active)))
	default:
		return func(s *sql.Selector) { s.Where(sql.False()) }
	}
}

// ListAdminPage 分页列出管理员视角的命名空间（真 SQL 分页）：分类过滤/统计/分页全部下沉
// SQL，stats 基于 search 命中全量（不带边的 base 计数：COUNT 会带上行查询的 Order/Select
// 修饰符，分开取才不把行侧排序 JOIN 带进计数），count 为分类
// 过滤后总数（无过滤 = total）。行级 lastActiveAt/活跃度仍由 biz 依已加载的项目边计算，
// 故分页行保留 withAdminEdges 全量边装配（成员/项目/关注，供前端下钻）。
// 排序按 id 升序（现状自然序），保证 LIMIT/OFFSET 翻页确定性。
func (repo *namespaceRepo) ListAdminPage(ctx context.Context, query *biz.AdminListPageQuery) (page *biz.AdminListPageResult, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/ListAdminPage")
	defer func() { endSpan(span, err) }()
	base := repo.adminNamespaceBaseQuery(&biz.ListNamespaceInput{
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

// ListAdminDeletedPage 分页列出**已软删**的命名空间（仅超管恢复流程使用）：Restore 的配套
// 「选谁恢复」视图，供恢复页按行选择后调 Restore。
//
// 必须绕过 SoftDeleteMixin：拦截器会给查询补 deleted_at IS NULL，与「只取软删行」自相矛盾，
// 结果恒为空。故整段查询都在 mixin.SkipSoftDelete(ctx) 下执行——注意该 ctx 也必须传给
// Count 与边加载，否则计数与行集口径不一致（计数 0 而列表非空）。
//
// SkipSoftDelete 同时会抑制 eager-load 边的软删过滤（拦截器不进子查询），故项目边天然会
// 连带捞出软删项目；withAdminDeletedEdges 用 DeletedWithNamespace(true) 把它收敛回
// 「随空间级联删除」的精确集合，恰好等于 RestoreDeleted 的恢复范围。
//
// 搜索复用 adminNamespaceBaseQuery 的 search 语义（空间名 OR 创建者邮箱），保证与
// ListAdminPage 同一份定义；Name/PrivateOnly 留空即不生效（filters.If 对空值跳过）。
// 排序按 deleted_at 倒序（最近删除的排最前），id 倒序作第二决胜键——deleted_at 是秒级
// datetime，同秒删除的多行仅按它排序翻页顺序不确定，接上 id 才能保证 LIMIT/OFFSET 稳定，
// 与 FindDeletedByName 的双键定序同理。
func (repo *namespaceRepo) ListAdminDeletedPage(ctx context.Context, query *biz.NamespaceDeletedListPageQuery) (page *biz.NamespaceDeletedListPageResult, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/ListAdminDeletedPage")
	defer func() { endSpan(span, err) }()
	ctx = mixin.SkipSoftDelete(ctx)
	base := repo.adminNamespaceBaseQuery(&biz.ListNamespaceInput{Search: query.Search}).
		Where(namespace.DeletedAtNotNil())
	// 计数走不带边的 base（与 ListAdminPage 同策略）：COUNT 会原样带上行查询挂的 Order/
	// Select 修饰符，故计数与行查询分开取；ent 的边 eager load 是独立往返，本就不参与 COUNT。
	count, err := base.Clone().Count(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "count deleted namespaces")
	}
	all, err := withAdminDeletedEdges(base.Clone()).
		Order(ent.Desc(namespace.FieldDeletedAt)).
		Order(ent.Desc(namespace.FieldID)).
		Offset(pagination.GetPageOffset(query.Page, query.PageSize)).
		Limit(int(query.PageSize)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list deleted namespaces page")
	}
	return &biz.NamespaceDeletedListPageResult{
		Namespaces: slice.Map(all, toNamespace),
		Count:      count,
	}, nil
}

// UpdateImagePullSecrets 仅回写 namespace 的 imagePullSecrets 列表，
// cron 对账后把新增/清理后的 secret 名单持久化。
//
// ⚠️ 本方法刻意**不**加 namespace.DeletedAtIsNil() 谓词（与 projectRepo.UpdateVersion /
// UpdateDeployStatus 的处理不同）。原因是它有一个必须写「已软删行」的生产调用方：
// namespaceBiz.Restore 在清 deleted_at **之前**先调本方法回写重建的 docker secret 名单
// （biz/namespace.go 的「先补 DB 骨架再清软删标记」顺序是刻意的，为的是失败可重试而非留下
// 不可恢复的孤儿空间）。加上谓词会让该调用命中 0 行、经 ensureExists 报 404，直接把「恢复被
// 误删空间」这条链路打断。要让本方法也能加谓词，须先让 biz 侧以 mixin.SkipSoftDelete 的 ctx
// 调用（或调整调用顺序），那超出本次改动文件域——回归护栏见
// TestNamespaceRepo_UpdateImagePullSecrets_SoftDeletedRowForRestore。
func (repo *namespaceRepo) UpdateImagePullSecrets(ctx context.Context, id int, secrets []string) (err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/UpdateImagePullSecrets")
	defer func() { endSpan(span, err) }()
	return errs.Wrap(repo.data.DB().Namespace.UpdateOneID(id).SetImagePullSecrets(secrets).Exec(ctx), "update image pull secrets")
}

// Delete 删除 namespace：若关联项目非空先删除项目，再在事务内删除 namespace。
//
// 级联删除的项目在同一事务内置 deleted_with_namespace=true（显式批次标识），RestoreDeleted
// 据此精确区分「随空间一起删的项目」与「用户早先单独删的项目」。
//
// 不用 deleted_at 相等去近似批次归属：该列是 MySQL `datetime`（= datetime(0)，秒级），
// 亚秒部分落库时被截断/进位，若同一秒内先单独删了某项目、又删了它所属空间，两者落库后
// 时间戳完全相同，"同秒"被误读成"同批"——恢复空间时把用户早已主动删除的项目一并复活成
// deploy_status 未知的幽灵记录。批次归属是一条确定的事实，就该用确定的标识表达。
//
// 空间与级联项目仍写入同一个 deleted_at：此时它只承担"这批是同一时刻下线的"这层可读语义，
// 不再参与任何判定。
func (repo *namespaceRepo) Delete(ctx context.Context, id int) (err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/Delete")
	defer func() { endSpan(span, err) }()
	deletedAt := time.Now()
	return errs.Wrap(repo.data.WithTx(ctx, func(tx *ent.Tx) error {
		// 无条件标记级联项目，不再用「先查存活项目数、非空才 UPDATE」的事务外守卫：该守卫与
		// 本 UPDATE 的 WHERE 语义本就等价（无存活项目时 UPDATE 自然是 0 行），是纯冗余优化；
		// 但它读的是**事务外快照**，快照读完到事务开启之间新建的项目会被漏标，产出「空间已
		// 软删、其下仍有存活且未打标项目」的脏数据（TOCTOU）。去掉后判定回归事务内的 UPDATE
		// 本身，顺带省掉 WithProjects 的全列边加载（只为算一个 bool 就把 projects 全部列、
		// 含大文本字段一并拉进内存）。
		//
		// 项目按 namespace_id 外键列直连过滤，不用 HasNamespaceWith：后者产出裸
		// sql.Selector 子查询，不被软删拦截器覆盖（已知的软删陷阱），外键判定等价且更直接。
		if err := tx.Project.
			Update().
			Where(project.NamespaceID(id), project.DeletedAtIsNil()).
			SetDeletedAt(deletedAt).
			SetDeletedWithNamespace(true).
			Exec(ctx); err != nil {
			return err
		}
		// 空间不存在或已被软删：本条 UPDATE 命中 0 行，UpdateOne 经 sqlgraph 的 ensureExists
		// 复查（谓词被带进 EXISTS 子查询）返回 ent NotFound，errs.Wrap 归类为 404——与原先
		// 事务外预查询 First 的 NotFound 语义一致。整个事务随即回滚，不会残留上面那条项目标记。
		return tx.Namespace.
			UpdateOneID(id).
			Where(namespace.DeletedAtIsNil()).
			SetDeletedAt(deletedAt).
			Exec(ctx)
	}), "delete namespace")
}

// FindDeletedByName 按名称查询**已软删**的 namespace，供超管恢复被误删的空间使用。
// 名称按 ns_prefix 幂等补全（对齐 FindByName）：调用方传界面展示名，带不带前缀均可命中。
// 同名空间可能被反复"删除→新建→再删除"，故按 deleted_at 倒序取最近一次删除的那行。
//
// id 是必须的第二决胜键：deleted_at 是 MySQL datetime（秒级），"删→同名重建→再删"落在同一秒
// 时两行时间戳相同，仅按它排序取哪行**不确定**——取错（通常命中小 id 的旧行）会让旧记录转在册，
// 随后新记录被同名在册检查永久挡死。id 越大者创建越晚，同秒内也就是后删的那条，故同样倒序。
// 「字段 + 主键」双键定序与本包 renumberFavoriteSortOrders 的 (sort_order, id) 稳定序同理；
// 那边写在单个 Order(f1, f2) 调用内，此处分两次 Order——两者语义等价（ent 的 Order 是
// `_q.order = append(_q.order, o...)` 累加而非覆盖），分开写只是为让主键决胜单独成句。
// 必须绕过软删拦截器：拦截器会给查询补 deleted_at IS NULL，与「只取软删行」自相矛盾。
func (repo *namespaceRepo) FindDeletedByName(ctx context.Context, name string) (out *biz.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/FindDeletedByName")
	defer func() { endSpan(span, err) }()
	first, err := repo.data.DB().Namespace.Query().
		Where(
			namespace.Name(biz.GetNamespace(name, repo.nsPrefix)),
			namespace.DeletedAtNotNil(),
		).
		Order(ent.Desc(namespace.FieldDeletedAt)).
		Order(ent.Desc(namespace.FieldID)).
		First(mixin.SkipSoftDelete(ctx))
	return toNamespace(first), errs.Wrap(err, "find deleted namespace by name")
}

// RestoreDeleted 恢复软删的 namespace 及其**同批次**级联删除的项目（清空 deleted_at）。
//
// 只还原 deleted_with_namespace=true 的项目：该标记由 Delete 与软删在同一事务写入，是
// "随空间一起被删"的确定事实，与时间戳精度无关；用户早先单独删除的项目标记为 false，不在
// 还原范围内——否则恢复空间会把那些早已被主动删除的项目静默复活成无部署资源的幽灵记录。
// 还原时一并把标记置回 false，维持不变式「标记为 true ⟺ 项目正处于随空间级联软删的状态」。
// （本列上线前已软删的存量项目保持列默认值 false，不在同批还原范围内——见迁移文件
// 20260921012959_add_project_deleted_with_namespace.sql。）
//
// 被还原项目的 deploy_status 一律重置为 StatusUnknown：删除空间时 helm release 已被物理
// 卸载，残留的"已部署"是对不存在资源的错误描述。StatusUnknown 与 ReleaseStatus 对不存在
// release 的返回值一致，也对应"骨架已恢复、请重新部署"的语义。
//
// 返回值是本次一并恢复的项目名（按 id 升序），供上位层落审计日志——与 Delete 返回被删项目名
// 对称：恢复范围只存在于本函数 UPDATE 的 WHERE 里，事务一结束就与"本就在册的项目"不可区分，
// 调用方无法在事后自行补算（空间下的存活项目还含竞态孤儿，见 cronjob.FixDeployStatus 的守卫）。
func (repo *namespaceRepo) RestoreDeleted(ctx context.Context, id int) (restored []string, err error) {
	ctx, span := tracer.Start(ctx, "namespaceRepo/RestoreDeleted")
	defer func() { endSpan(span, err) }()
	// 目标行本身是软删行，而本函数后续无论 SELECT 还是清 deleted_at 的 UPDATE 都与软删拦截器
	// 补的 deleted_at IS NULL 相矛盾（SELECT 取不到会误报 NotFound，UPDATE 命中 0 行且静默
	// "成功"），故整段统一绕过拦截器。
	ctx = mixin.SkipSoftDelete(ctx)
	ns, err := repo.data.DB().Namespace.Query().Where(namespace.ID(id)).Only(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "restore namespace")
	}
	if ns.DeletedAt == nil {
		// 未软删的空间无需恢复：显式报错优于静默成功，避免调用方把「什么都没做」当完成。
		return nil, errs.WrapInvalidArgument(fmt.Errorf("空间 %d 未被删除，无需恢复", id), "restore namespace")
	}
	if err = errs.Wrap(repo.data.WithTx(ctx, func(tx *ent.Tx) error {
		// 名单必须在清标记**之前**捞：UPDATE 只回影响行数，而这批行的 deleted_with_namespace
		// 一旦被清零就与"本就在册的项目"不可区分，事后再查无从分辨。读取仅为留痕，恢复范围仍
		// 由下面 UPDATE 的 WHERE 单方裁定——两者同事务同谓词，不存在读到的集合与写掉的集合不一致
		// 的窗口。显式 Order 是因为无序 SELECT 的行序在 SQL 里不确定，审计日志顺序得稳定。
		names, qerr := tx.Project.Query().
			Where(project.NamespaceID(id), project.DeletedWithNamespace(true)).
			Order(ent.Asc(project.FieldID)).
			Select(project.FieldName).
			Strings(ctx)
		if qerr != nil {
			return qerr
		}
		if err := tx.Project.
			Update().
			Where(project.NamespaceID(id), project.DeletedWithNamespace(true)).
			SetDeployStatus(types.Deploy_StatusUnknown).
			SetDeletedWithNamespace(false).
			ClearDeletedAt().
			Exec(ctx); err != nil {
			return err
		}
		restored = names
		return tx.Namespace.UpdateOneID(id).ClearDeletedAt().Exec(ctx)
	}), "restore namespace"); err != nil {
		// 事务已回滚：恢复名单随之一并作废，不向调用方交"半份结果 + 一个错误"的歧义。
		return nil, err
	}
	return restored, nil
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
