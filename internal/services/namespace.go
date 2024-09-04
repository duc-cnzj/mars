package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v5/namespace"
	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/transformer"
	"github.com/duc-cnzj/mars/v5/internal/util/pagination"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
)

var _ namespace.NamespaceServer = (*namespaceSvc)(nil)

type namespaceSvc struct {
	namespace.UnimplementedNamespaceServer

	helmer    repo.HelmerRepo
	nsRepo    repo.NamespaceRepo
	k8sRepo   repo.K8sRepo
	logger    mlog.Logger
	eventRepo repo.EventRepo
}

func NewNamespaceSvc(helmer repo.HelmerRepo, nsRepo repo.NamespaceRepo, k8sRepo repo.K8sRepo, logger mlog.Logger, eventRepo repo.EventRepo) namespace.NamespaceServer {
	return &namespaceSvc{
		helmer:    helmer,
		nsRepo:    nsRepo,
		k8sRepo:   k8sRepo,
		logger:    logger.WithModule("services/namespace"),
		eventRepo: eventRepo,
	}
}

func (n *namespaceSvc) List(ctx context.Context, request *namespace.ListRequest) (*namespace.ListResponse, error) {
	user := MustGetUser(ctx)
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	namespaces, pag, err := n.nsRepo.List(ctx, &repo.ListNamespaceInput{
		Favorite: request.Favorite,
		Email:    user.Email,
		Page:     page,
		PageSize: size,
		Name:     request.Name,
		IsAdmin:  user.IsAdmin(),
	})
	if err != nil {
		return nil, err
	}
	var res = &namespace.ListResponse{
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

func (n *namespaceSvc) Create(ctx context.Context, request *namespace.CreateRequest) (*namespace.CreateResponse, error) {
	user := MustGetUser(ctx)
	nsName := n.nsRepo.GetMarsNamespace(request.Namespace)
	preCheckNs, err := n.nsRepo.FindByName(ctx, nsName)
	if err == nil {
		if request.IgnoreIfExists {
			return &namespace.CreateResponse{Item: transformer.FromNamespace(preCheckNs), Exists: true}, nil
		}
		return nil, status.Error(codes.AlreadyExists, "名称空间已存在")
	}

	create, err := n.k8sRepo.CreateNamespace(ctx, nsName)
	// 创建名称空间
	if err != nil {
		if !k8sapierrors.IsAlreadyExists(err) {
			return nil, err
		}
		found, err := n.k8sRepo.GetNamespace(ctx, nsName)
		if err != nil {
			return nil, err
		}
		if found.Status.Phase == v1.NamespaceTerminating {
			return nil, status.Error(codes.AlreadyExists, "该名称空间正在删除中")
		}
	}
	n.logger.Debug("成功创建namespace: ", create.Name)

	var imagePullSecrets []string
	secret, err := n.k8sRepo.CreateDockerSecret(ctx, create.Name)
	if err == nil {
		imagePullSecrets = append(imagePullSecrets, secret.Name)
	} else {
		n.logger.Error(err)
	}

	ns, err := n.nsRepo.Create(ctx, &repo.CreateNamespaceInput{
		Name:             create.Name,
		ImagePullSecrets: imagePullSecrets,
		Description:      request.Description,
		CreatorEmail:     user.Email,
	})
	if err != nil {
		return nil, err
	}
	n.nsRepo.Favorite(ctx, &repo.FavoriteNamespaceInput{
		NamespaceID: ns.ID,
		UserEmail:   user.Email,
		Favorite:    true,
	})

	n.eventRepo.Dispatch(repo.EventNamespaceCreated, repo.NamespaceCreatedData{
		NsModel:  ns,
		NsK8sObj: create,
	})

	n.eventRepo.AuditLogWithRequest(
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

func (n *namespaceSvc) Show(ctx context.Context, input *namespace.ShowRequest) (*namespace.ShowResponse, error) {
	ns, err := n.nsRepo.Show(ctx, int(input.Id))
	if err != nil {
		return nil, err
	}
	if access := n.nsRepo.CanAccess(ctx, int(input.Id), MustGetUser(ctx)); !access {
		return nil, repo.ToError(403, "没有权限")
	}
	return &namespace.ShowResponse{Item: transformer.FromNamespace(ns)}, nil
}

func (n *namespaceSvc) UpdateDesc(ctx context.Context, req *namespace.UpdateDescRequest) (*namespace.UpdateDescResponse, error) {
	old, err := n.nsRepo.Show(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	if access := n.nsRepo.CanAccess(ctx, int(req.Id), MustGetUser(ctx)); !access {
		return nil, repo.ToError(403, "没有权限")
	}

	ns, err := n.nsRepo.Update(ctx, &repo.UpdateNamespaceInput{
		ID:          int(req.Id),
		Description: req.Desc,
	})
	if err != nil {
		return nil, err
	}

	n.eventRepo.AuditLogWithChange(
		types.EventActionType_Update,
		MustGetUser(ctx).Name,
		fmt.Sprintf("更新项目空间描述: id: '%d' '%s'", ns.ID, ns.Name),
		&repo.AnyYamlPrettier{
			"namespace": old.Name,
			"desc":      old.Description,
		},
		&repo.AnyYamlPrettier{
			"namespace": ns.Name,
			"desc":      ns.Description,
		},
	)

	return &namespace.UpdateDescResponse{Item: transformer.FromNamespace(ns)}, nil
}

func (n *namespaceSvc) Delete(ctx context.Context, input *namespace.DeleteRequest) (*namespace.DeleteResponse, error) {
	user := auth.MustGetUser(ctx)
	if !user.IsAdmin() {
		return nil, ErrorPermissionDenied
	}

	ns, err := n.nsRepo.Show(ctx, int(input.Id))
	if err != nil {
		return nil, err
	}

	var deletedProjectNames []string
	if err == nil {
		wg := sync.WaitGroup{}
		wg.Add(len(ns.Projects))
		for _, project := range ns.Projects {
			deletedProjectNames = append(deletedProjectNames, project.Name)
			go func(releaseName, namespace string) {
				defer wg.Done()
				defer n.logger.HandlePanic("namespaceSvc.Delete")
				n.logger.Debugf("delete release %s namespace %s", releaseName, namespace)
				if err := n.helmer.Uninstall(releaseName, namespace, n.logger.Debugf); err != nil {
					n.logger.ErrorCtx(ctx, err)
					return
				}
			}(project.Name, ns.Name)
		}
		wg.Wait()
		for _, secret := range ns.ImagePullSecrets {
			n.logger.DebugCtxf(ctx, "delete ns %s secret %s", ns.Name, secret)
			n.k8sRepo.DeleteSecret(context.TODO(), ns.Name, secret)
		}
		if err := n.k8sRepo.DeleteNamespace(context.TODO(), ns.Name); err != nil {
			n.logger.ErrorCtx(ctx, "删除 namespace 出现错误: ", err)
		}
		if err := n.nsRepo.Delete(context.TODO(), ns.ID); err != nil {
			n.logger.ErrorCtx(ctx, "删除 namespace 出现错误: ", err)
			return nil, err
		}
	}

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			if _, err := n.k8sRepo.GetNamespace(context.TODO(), ns.Name); err != nil {
				n.logger.Error(err)
				break loop
			}
		case <-timer.C:
			break loop
		}
	}

	n.eventRepo.Dispatch(repo.EventNamespaceDeleted, repo.NamespaceDeletedData{ID: ns.ID})

	n.eventRepo.AuditLogWithRequest(
		types.EventActionType_Delete,
		MustGetUser(ctx).Name,
		fmt.Sprintf("删除项目空间: id: '%d' '%s', 删除的项目有: '%s'", ns.ID, ns.Name, strings.Join(deletedProjectNames, ", ")),
		input,
	)

	return &namespace.DeleteResponse{}, nil
}

func (n *namespaceSvc) IsExists(ctx context.Context, input *namespace.IsExistsRequest) (*namespace.IsExistsResponse, error) {
	ns, err := n.nsRepo.FindByName(ctx, n.nsRepo.GetMarsNamespace(input.Name))
	if err != nil {
		if ent.IsNotFound(err) {
			return &namespace.IsExistsResponse{Exists: false}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &namespace.IsExistsResponse{Exists: true, Id: int64(ns.ID)}, nil
}

func (n *namespaceSvc) Favorite(ctx context.Context, req *namespace.FavoriteRequest) (*namespace.FavoriteResponse, error) {
	user := MustGetUser(ctx)
	err := n.nsRepo.Favorite(ctx, &repo.FavoriteNamespaceInput{
		NamespaceID: int(req.Id),
		UserEmail:   user.Email,
		Favorite:    req.Favorite,
	})
	if err != nil {
		return nil, err
	}
	return &namespace.FavoriteResponse{}, nil
}

func (n *namespaceSvc) UpdatePrivate(ctx context.Context, req *namespace.UpdatePrivateRequest) (*namespace.UpdatePrivateResponse, error) {
	user := MustGetUser(ctx)
	isOwner, err := n.nsRepo.IsOwner(ctx, int(req.Id), user)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, repo.ToError(403, "没有权限")
	}

	private, err := n.nsRepo.UpdatePrivate(ctx, int(req.Id), req.Private)
	if err != nil {
		return nil, err
	}
	n.eventRepo.AuditLogWithRequest(
		types.EventActionType_Update,
		MustGetUser(ctx).Name,
		fmt.Sprintf("[更新空间访问权限] id: %v private: %t", req.Id, req.GetPrivate()),
		req,
	)
	return &namespace.UpdatePrivateResponse{Item: transformer.FromNamespace(private)}, nil
}

func (n *namespaceSvc) SyncMembers(ctx context.Context, req *namespace.SyncMembersRequest) (*namespace.SyncMembersResponse, error) {
	user := MustGetUser(ctx)
	isOwner, err := n.nsRepo.IsOwner(ctx, int(req.Id), user)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, repo.ToError(403, "没有权限")
	}
	show, err := n.nsRepo.Show(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	ns, err := n.nsRepo.SyncMembers(ctx, int(req.Id), lo.Uniq(req.Emails))
	if err != nil {
		return nil, err
	}

	n.eventRepo.AuditLogWithChange(
		types.EventActionType_Update,
		MustGetUser(ctx).Name,
		fmt.Sprintf("[同步空间成员] id: %v name: %v", show.ID, show.Name),
		&repo.AnyYamlPrettier{
			"members": show.Members,
		},
		&repo.AnyYamlPrettier{
			"members": ns.Members,
		},
	)

	return &namespace.SyncMembersResponse{Item: transformer.FromNamespace(ns)}, nil
}
