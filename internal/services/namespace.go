package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v4/namespace"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
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
	return &namespaceSvc{helmer: helmer, nsRepo: nsRepo, k8sRepo: k8sRepo, eventRepo: eventRepo, logger: logger.WithModule("services/namespace")}
}

func (n *namespaceSvc) All(ctx context.Context, request *namespace.AllRequest) (*namespace.AllResponse, error) {
	email := MustGetUser(ctx).Email
	namespaces, err := n.nsRepo.All(ctx, &repo.AllNamespaceInput{
		Favorite: request.Favorite,
		Email:    email,
	})
	if err != nil {
		return nil, err
	}
	var res = &namespace.AllResponse{Items: make([]*types.NamespaceModel, 0, len(namespaces))}
	for _, ns := range namespaces {
		fav := false
		for _, f := range ns.Favorites {
			if f.Email == email {
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
	secret, err := n.k8sRepo.CreateDockerSecrets(ctx, create.Name)
	if err == nil {
		imagePullSecrets = append(imagePullSecrets, secret.Name)
	} else {
		n.logger.Error(err)
	}

	ns, err := n.nsRepo.Create(ctx, &repo.CreateNamespaceInput{
		Name:             create.Name,
		ImagePullSecrets: imagePullSecrets,
	})
	if err != nil {
		return nil, err
	}
	n.nsRepo.Favorite(ctx, &repo.FavoriteNamespaceInput{
		NamespaceID: ns.ID,
		UserEmail:   MustGetUser(ctx).Email,
		Favorite:    true,
	})

	n.eventRepo.Dispatch(repo.EventNamespaceCreated, repo.NamespaceCreatedData{
		NsModel:  ns,
		NsK8sObj: create,
	})

	n.eventRepo.AuditLog(types.EventActionType_Create, MustGetUser(ctx).Name, fmt.Sprintf("创建项目空间: %d: %s", ns.ID, ns.Name))

	return &namespace.CreateResponse{
		Item:   transformer.FromNamespace(ns),
		Exists: false,
	}, nil
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
			n.k8sRepo.DeleteSecret(context.Background(), ns.Name, secret)
		}
		if err := n.k8sRepo.DeleteNamespace(context.Background(), ns.Name); err != nil {
			n.logger.ErrorCtx(ctx, "删除 namespace 出现错误: ", err)
		}
		n.nsRepo.Delete(context.TODO(), ns.ID)
	}

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
loop:
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			if _, err := n.k8sRepo.GetNamespace(context.Background(), ns.Name); err != nil {
				n.logger.Error(err)
				break loop
			}
		case <-timer.C:
			break loop
		}
	}

	n.eventRepo.Dispatch(repo.EventNamespaceDeleted, repo.NamespaceDeletedData{NsModel: ns})

	n.eventRepo.AuditLog(types.EventActionType_Delete, MustGetUser(ctx).Name, fmt.Sprintf("删除项目空间: id: '%d' '%s', 删除的项目有: '%s'", ns.ID, ns.Name, strings.Join(deletedProjectNames, ", ")))

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

func (n *namespaceSvc) Show(ctx context.Context, input *namespace.ShowRequest) (*namespace.ShowResponse, error) {
	ns, err := n.nsRepo.Show(ctx, int(input.Id))
	if err != nil {
		return nil, err
	}
	return &namespace.ShowResponse{Item: transformer.FromNamespace(ns)}, nil
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
