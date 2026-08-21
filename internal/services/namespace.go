package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/samber/lo"
)

var _ namespace.NamespaceServer = (*namespaceSvc)(nil)

// namespaceSvc 是 namespace.NamespaceServer 的 gRPC 实现：管理命名空间全生命周期
// （增删改查/成员同步/收藏/私有化/转移），经 access 校验访问权限，由 NewNamespaceSvc 构造。
type namespaceSvc struct {
	namespace.UnimplementedNamespaceServer

	nsBiz     biz.NamespaceBiz
	logger    mlog.Logger
	eventBiz  biz.EventBiz
	accessBiz biz.AccessBiz
}

// NamespaceSvcDeps 收口 NewNamespaceSvc 的构造依赖，由 wire 按字段注入。
// namespace 域单一业务接口 NamespaceBiz 承载 CRUD 门面 + Create/Delete 编排，
// 传输层只保留鉴权、协议映射与审计日志（EventBiz）。
type NamespaceSvcDeps struct {
	NsBiz     biz.NamespaceBiz
	Logger    mlog.Logger
	EventBiz  biz.EventBiz
	AccessBiz biz.AccessBiz
}

// NewNamespaceSvc 收口命名空间服务的构造依赖，由 wire 按字段注入。
func NewNamespaceSvc(deps NamespaceSvcDeps) namespace.NamespaceServer {
	logger := deps.Logger.WithModule("services/namespace")
	return &namespaceSvc{
		nsBiz:     deps.NsBiz,
		logger:    logger,
		eventBiz:  deps.EventBiz,
		accessBiz: deps.AccessBiz,
	}
}

// showNsAndCheckOwner 加载命名空间并校验当前用户为 owner：Transfer/Delete/
// UpdatePrivate/SyncMembers 四个 owner 变更入口的公共前置，加载失败或非 owner
// 直接返回错误，消除四处重复的"Show + Owner 校验"样板。
func (n *namespaceSvc) showNsAndCheckOwner(ctx context.Context, id int) (*biz.Namespace, error) {
	show, err := n.nsBiz.Show(ctx, id)
	if err != nil {
		return nil, logError(ctx, n.logger, err)
	}
	if err := n.accessBiz.RequireNamespaceOwner(ctx, show); err != nil {
		return nil, logError(ctx, n.logger, err)
	}
	return show, nil
}

// Transfer 把命名空间的所有权转让给指定邮箱的用户，仅 owner 可操作，落更新审计日志。
func (n *namespaceSvc) Transfer(ctx context.Context, request *namespace.TransferRequest) (*namespace.TransferResponse, error) {
	user := biz.MustGetUser(ctx)
	// 返回值 show 仅用于 owner 校验，审计消息不引用，用 _ 忽略。
	if _, err := n.showNsAndCheckOwner(ctx, int(request.Id)); err != nil {
		return nil, err
	}
	transfer, err := n.nsBiz.Transfer(ctx, int(request.Id), request.NewAdminEmail)
	if err != nil {
		return nil, logError(ctx, n.logger, err)
	}
	n.eventBiz.AuditLogWithRequest(
		types.EventActionType_Update,
		user.Name,
		"转让项目空间给: "+request.NewAdminEmail,
		request,
	)
	return &namespace.TransferResponse{Item: transformer.FromNamespace(transfer)}, nil
}

// List 分页列出当前用户可见的命名空间（admin 看全部，普通用户看加入的），并标记是否已关注。
func (n *namespaceSvc) List(ctx context.Context, request *namespace.ListRequest) (*namespace.ListResponse, error) {
	user := biz.MustGetUser(ctx)
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	namespaces, pag, err := n.nsBiz.List(ctx, &biz.ListNamespaceInput{
		Favorite: request.Favorite,
		Email:    user.Email,
		Page:     page,
		PageSize: size,
		Name:     request.Name,
		IsAdmin:  user.IsAdmin(),
	})
	if err != nil {
		return nil, logError(ctx, n.logger, err)
	}
	res := &namespace.ListResponse{
		Items:    make([]*types.NamespaceModel, 0, len(namespaces)),
		Count:    pag.Count,
		Page:     pag.Page,
		PageSize: pag.PageSize,
	}
	for _, ns := range namespaces {
		fav := false
		for _, f := range ns.Favorites {
			if f.Email == user.Email {
				fav = true
				break
			}
		}
		v := transformer.FromNamespace(ns)
		v.Favorite = fav
		res.Items = append(res.Items, v)
	}

	return res, nil
}

// Create 创建命名空间：已存在时按 IgnoreIfExists 策略放行或返回 AlreadyExists；放行前
// 校验当前用户对已存在空间的访问权限（无权访问私有空间直接 403），落创建审计日志。
func (n *namespaceSvc) Create(ctx context.Context, request *namespace.CreateRequest) (*namespace.CreateResponse, error) {
	user := biz.MustGetUser(ctx)
	ns, exists, err := n.nsBiz.Create(ctx, request.Namespace, request.Description, user.Email)
	if err != nil {
		// Terminating 已是 biz 携带 AlreadyExists status 的领域错误（ErrNamespaceTerminating），
		// 其余错误（k8s/DB 故障）原样上抛——协议映射收口 biz，错误日志统一由本层 logError 打印。
		return nil, logError(ctx, n.logger, err)
	}
	if exists {
		// 命名空间已存在：按 IgnoreIfExists 策略放行（携带已存在的空间）或拒绝。
		if request.IgnoreIfExists {
			// 幂等放行不得把无权限用户无权看到的私有空间对象吐出去：ns 来自全局
			// FindByName（不感知权限），若当前用户无权访问直接 403，与 IsExists
			// "私有空间视同不存在"的隐藏语义对齐，闭合元数据泄露面。
			if !n.accessBiz.CanAccessNamespace(ctx, ns) {
				return nil, logError(ctx, n.logger, errs.ErrorPermissionDenied)
			}
			return &namespace.CreateResponse{Item: transformer.FromNamespace(ns), Exists: true}, nil
		}
		// exists 且非 IgnoreIfExists：拒绝并返回 AlreadyExists——状态码由 biz 工厂提供，
		// transport 不再散落 status.Error 构造（协议映射收口 biz）。
		return nil, errs.AlreadyExists(fmt.Sprintf("名称空间 %s 已存在", request.Namespace))
	}

	n.eventBiz.AuditLogWithRequest(
		types.EventActionType_Create,
		user.Name,
		fmt.Sprintf("创建项目空间: %d: %s", ns.ID, ns.Name),
		request,
	)

	return &namespace.CreateResponse{
		Item:   transformer.FromNamespace(ns),
		Exists: false,
	}, nil
}

// Show 返回命名空间详情，响应前做命名空间级访问控制。
func (n *namespaceSvc) Show(ctx context.Context, input *namespace.ShowRequest) (*namespace.ShowResponse, error) {
	ns, nerr := n.accessBiz.RequireNamespaceAccessByID(ctx, int(input.Id))
	if nerr != nil {
		return nil, logError(ctx, n.logger, nerr)
	}
	return &namespace.ShowResponse{Item: transformer.FromNamespace(ns)}, nil
}

// UpdateDesc 更新命名空间描述，落变更审计日志（含前后 diff）。
func (n *namespaceSvc) UpdateDesc(ctx context.Context, req *namespace.UpdateDescRequest) (*namespace.UpdateDescResponse, error) {
	old, err := n.nsBiz.Show(ctx, int(req.Id))
	if err != nil {
		return nil, logError(ctx, n.logger, err)
	}

	ns, err := n.nsBiz.Update(ctx, &biz.UpdateNamespaceInput{
		ID:          int(req.Id),
		Description: req.Desc,
	})
	if err != nil {
		return nil, logError(ctx, n.logger, err)
	}

	n.eventBiz.AuditLogWithChange(
		types.EventActionType_Update,
		biz.MustGetUser(ctx).Name,
		fmt.Sprintf("更新项目空间描述: id: '%d' '%s'", ns.ID, ns.Name),
		biz.AnyYamlPrettier{
			"namespace": old.Name,
			"desc":      old.Description,
		},
		biz.AnyYamlPrettier{
			"namespace": ns.Name,
			"desc":      ns.Description,
		},
	)

	return &namespace.UpdateDescResponse{Item: transformer.FromNamespace(ns)}, nil
}

// Delete 删除命名空间（含其下项目），仅 owner 可操作，落删除审计日志。
func (n *namespaceSvc) Delete(ctx context.Context, input *namespace.DeleteRequest) (*namespace.DeleteResponse, error) {
	user := biz.MustGetUser(ctx)
	ns, err := n.showNsAndCheckOwner(ctx, int(input.Id))
	if err != nil {
		return nil, err
	}

	deletedProjectNames, err := n.nsBiz.Delete(ctx, ns)
	if err != nil {
		// 删除编排（并发卸载/删 secret/删 k8s/删 DB/轮询确认）已下沉 biz，错误原样上抛——
		// 错误日志统一由本层 logError 打印。
		return nil, logError(ctx, n.logger, err)
	}

	n.eventBiz.AuditLogWithRequest(
		types.EventActionType_Delete,
		user.Name,
		fmt.Sprintf("删除项目空间: id: '%d' '%s', 删除的项目有: '%s'", ns.ID, ns.Name, strings.Join(deletedProjectNames, ", ")),
		input,
	)

	return &namespace.DeleteResponse{}, nil
}

// IsExists 查询命名空间是否存在：对无权限用户隐藏存在性，私有空间视同不存在。
func (n *namespaceSvc) IsExists(ctx context.Context, input *namespace.IsExistsRequest) (*namespace.IsExistsResponse, error) {
	ns, err := n.nsBiz.FindByName(ctx, n.nsBiz.GetMarsNamespace(input.Name))
	if err != nil {
		if errs.IsNotFound(err) {
			return &namespace.IsExistsResponse{Exists: false}, nil
		}
		// 与其他 service 一致：非 NotFound 直接返回原始错误，不额外包装 codes.Internal。
		return nil, logError(ctx, n.logger, err)
	}

	// 私有命名空间对无权限用户隐藏存在性：不可访问视同不存在（Exists: false）。
	// 若返回 403 反而暴露"存在但无权限"，IsExists 就成了探测私有命名空间的
	// IDOR 侧信道——与 List 按 user 过滤的语义对齐。
	if !n.accessBiz.CanAccessNamespace(ctx, ns) {
		return &namespace.IsExistsResponse{Exists: false}, nil
	}

	return &namespace.IsExistsResponse{Exists: true, Id: int64(ns.ID)}, nil
}

// Favorite 关注/取消关注命名空间，落更新审计日志。
func (n *namespaceSvc) Favorite(ctx context.Context, req *namespace.FavoriteRequest) (*namespace.FavoriteResponse, error) {
	user := biz.MustGetUser(ctx)
	err := n.nsBiz.Favorite(ctx, &biz.FavoriteNamespaceInput{
		NamespaceID: int(req.Id),
		UserEmail:   user.Email,
		Favorite:    req.Favorite,
	})
	if err != nil {
		return nil, logError(ctx, n.logger, err)
	}
	str := "取消关注"
	if req.Favorite {
		str = "关注"
	}
	ns, err := n.nsBiz.Show(ctx, int(req.Id))
	if err != nil {
		return nil, logError(ctx, n.logger, err)
	}
	n.eventBiz.AuditLogWithRequest(
		types.EventActionType_Update,
		user.Name,
		fmt.Sprintf("用户%s项目空间: %s", str, ns.Name),
		req,
	)
	return &namespace.FavoriteResponse{}, nil
}

// FavoriteSort 整体重排当前用户的关注列表，落更新审计日志。
func (n *namespaceSvc) FavoriteSort(ctx context.Context, req *namespace.FavoriteSortRequest) (*namespace.FavoriteSortResponse, error) {
	user := biz.MustGetUser(ctx)
	err := n.nsBiz.FavoriteSort(ctx, &biz.FavoriteSortNamespaceInput{
		UserEmail:           user.Email,
		OrderedNamespaceIDs: lo.Map(req.NamespaceIds, func(id int32, _ int) int { return int(id) }),
	})
	if err != nil {
		return nil, logError(ctx, n.logger, err)
	}
	n.eventBiz.AuditLogWithRequest(
		types.EventActionType_Update,
		user.Name,
		"用户重排关注列表",
		req,
	)
	return &namespace.FavoriteSortResponse{}, nil
}

// UpdatePrivate 切换命名空间私密属性，仅 owner 可操作，落更新审计日志。
func (n *namespaceSvc) UpdatePrivate(ctx context.Context, req *namespace.UpdatePrivateRequest) (*namespace.UpdatePrivateResponse, error) {
	user := biz.MustGetUser(ctx)
	// 返回值 show 仅用于 owner 校验，审计消息不引用，用 _ 忽略。
	if _, err := n.showNsAndCheckOwner(ctx, int(req.Id)); err != nil {
		return nil, err
	}

	private, err := n.nsBiz.UpdatePrivate(ctx, int(req.Id), req.Private)
	if err != nil {
		return nil, logError(ctx, n.logger, err)
	}
	n.eventBiz.AuditLogWithRequest(
		types.EventActionType_Update,
		user.Name,
		fmt.Sprintf("[更新空间访问权限] id: %v private: %t, name: %v", req.Id, req.GetPrivate(), private.Name),
		req,
	)
	return &namespace.UpdatePrivateResponse{Item: transformer.FromNamespace(private)}, nil
}

// SyncMembers 同步命名空间成员（按邮箱去重），仅 owner 可操作。
func (n *namespaceSvc) SyncMembers(ctx context.Context, req *namespace.SyncMembersRequest) (*namespace.SyncMembersResponse, error) {
	user := biz.MustGetUser(ctx)
	show, err := n.showNsAndCheckOwner(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	ns, err := n.nsBiz.SyncMembers(ctx, int(req.Id), lo.Uniq(req.Emails))
	if err != nil {
		return nil, logError(ctx, n.logger, err)
	}

	n.eventBiz.AuditLogWithChange(
		types.EventActionType_Update,
		user.Name,
		fmt.Sprintf("[同步空间成员] id: %v name: %v", show.ID, show.Name),
		biz.AnyYamlPrettier{
			"members": show.Members,
		},
		biz.AnyYamlPrettier{
			"members": ns.Members,
		},
	)

	return &namespace.SyncMembersResponse{Item: transformer.FromNamespace(ns)}, nil
}

// UpdateConfig 一次性原子更新命名空间配置（描述/私有/成员/转让管理员），
// 仅空间管理员（创建者）/超级管理员（admin）可操作，落变更审计日志（含前后 diff）。
func (n *namespaceSvc) UpdateConfig(ctx context.Context, req *namespace.UpdateConfigRequest) (*namespace.UpdateConfigResponse, error) {
	user := biz.MustGetUser(ctx)
	// show 用于 owner 校验（非空间管理员/超级管理员直接 403）与审计前值。
	show, err := n.showNsAndCheckOwner(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	ns, err := n.nsBiz.UpdateConfig(ctx, &biz.UpdateConfigInput{
		ID:            int(req.Id),
		Description:   req.Desc,
		Private:       req.Private,
		Emails:        req.Emails,
		NewAdminEmail: req.NewAdminEmail,
	})
	if err != nil {
		return nil, logError(ctx, n.logger, err)
	}

	n.eventBiz.AuditLogWithChange(
		types.EventActionType_Update,
		user.Name,
		fmt.Sprintf("[批量更新空间配置] id: %v name: %v", show.ID, show.Name),
		biz.AnyYamlPrettier{
			"namespace": show.Name,
			"desc":      show.Description,
			"private":   show.Private,
			"members":   show.Members,
			"admin":     show.CreatorEmail,
		},
		biz.AnyYamlPrettier{
			"namespace": ns.Name,
			"desc":      ns.Description,
			"private":   ns.Private,
			"members":   ns.Members,
			"admin":     ns.CreatorEmail,
		},
	)

	return &namespace.UpdateConfigResponse{Item: transformer.FromNamespace(ns)}, nil
}
