package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars-client/v4/namespace"
	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/socket"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/internal/utils/recovery"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		namespace.RegisterNamespaceServer(s, &NamespaceSvc{
			helmer: &socket.DefaultHelmer{},
		})
	})
	RegisterEndpoint(namespace.RegisterNamespaceHandlerFromEndpoint)
}

type NamespaceSvc struct {
	helmer contracts.Helmer
	namespace.UnimplementedNamespaceServer
}

func (n *NamespaceSvc) All(ctx context.Context, request *namespace.AllRequest) (*namespace.AllResponse, error) {
	var namespaces []*models.Namespace
	app.DB().Preload("Projects", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID", "Name", "DeployStatus", "NamespaceId")
	}).
		Select("id", "name", "created_at").
		Find(&namespaces)
	var res = &namespace.AllResponse{Items: make([]*types.NamespaceModel, 0, len(namespaces))}
	for _, ns := range namespaces {
		res.Items = append(res.Items, ns.ProtoTransform())
	}

	return res, nil
}

func (n *NamespaceSvc) Create(ctx context.Context, request *namespace.CreateRequest) (*namespace.CreateResponse, error) {
	request.Namespace = utils.GetMarsNamespace(request.Namespace)

	preCheckNs := &models.Namespace{}
	if app.DB().Where("`name` = ?", request.Namespace).First(&preCheckNs).Error == nil {
		if request.IgnoreIfExists {
			return &namespace.CreateResponse{Namespace: preCheckNs.ProtoTransform(), Exists: true}, nil
		}
		return nil, status.Error(codes.AlreadyExists, "名称空间已存在")
	}

	// 创建名称空间
	create, err := app.K8sClientSet().CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: request.Namespace}}, metav1.CreateOptions{})
	if err != nil {
		if !k8sapierrors.IsAlreadyExists(err) {
			return nil, err
		}
		create, err = app.K8sClientSet().CoreV1().Namespaces().Get(context.TODO(), request.Namespace, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if create.Status.Phase == v1.NamespaceTerminating {
			return nil, status.Error(codes.AlreadyExists, "该名称空间正在删除中")
		}
	}
	mlog.Debug("成功创建namespace: ", create.Name)

	var imagePullSecrets []string
	secret, err := utils.CreateDockerSecrets(app.K8sClientSet(), request.Namespace, app.Config().ImagePullSecrets)
	if err == nil {
		imagePullSecrets = append(imagePullSecrets, secret.Name)
	}
	data := models.Namespace{Name: create.Name, ImagePullSecrets: strings.Join(imagePullSecrets, ",")}

	app.DB().Create(&data)

	app.Event().Dispatch(events.EventNamespaceCreated, events.NamespaceCreatedData{
		NsModel:  &data,
		NsK8sObj: create,
	})
	AuditLog(MustGetUser(ctx).Name, types.EventActionType_Create, fmt.Sprintf("创建项目空间: %d: %s", data.ID, data.Name))

	return &namespace.CreateResponse{
		Namespace: data.ProtoTransform(),
		Exists:    false,
	}, nil
}

func (n *NamespaceSvc) Delete(ctx context.Context, id *namespace.DeleteRequest) (*namespace.DeleteResponse, error) {
	var ns models.Namespace
	var deletedProjectNames []string
	// 删除空间前，要先删除空间下的项目
	if app.DB().Preload("Projects").Where("`id` = ?", id.NamespaceId).First(&ns).Error == nil {
		wg := sync.WaitGroup{}
		wg.Add(len(ns.Projects))
		for _, project := range ns.Projects {
			deletedProjectNames = append(deletedProjectNames, project.Name)
			go func(releaseName, namespace string) {
				defer wg.Done()
				defer recovery.HandlePanic("NamespaceSvc.Delete")
				mlog.Debugf("delete release %s namespace %s", releaseName, namespace)
				if err := n.helmer.Uninstall(releaseName, namespace, mlog.Debugf); err != nil {
					mlog.Error(err)
					return
				}
			}(project.Name, ns.Name)
		}
		wg.Wait()
		for _, secret := range ns.ImagePullSecretsArray() {
			mlog.Debugf("delete ns %s secret %s", ns.Name, secret)
			app.K8sClientSet().CoreV1().Secrets(ns.Name).Delete(context.Background(), secret, metav1.DeleteOptions{})
		}
		if err := app.K8sClientSet().CoreV1().Namespaces().Delete(context.Background(), ns.Name, metav1.DeleteOptions{}); err != nil {
			mlog.Error("删除 namespace 出现错误: ", err)
		}
		if len(ns.Projects) > 0 {
			app.DB().Delete(&ns.Projects)
		}
		app.DB().Delete(&ns)
	}

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
loop:
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			if _, err := app.K8sClientSet().CoreV1().Namespaces().Get(context.Background(), ns.Name, metav1.GetOptions{}); err != nil {
				mlog.Error(err)
				break loop
			}
		case <-timer.C:
			break loop
		}
	}

	app.Event().Dispatch(events.EventNamespaceDeleted, events.NamespaceDeletedData{NsModel: &ns})

	AuditLog(MustGetUser(ctx).Name, types.EventActionType_Delete, fmt.Sprintf("删除项目空间: id: '%d' '%s', 删除的项目有: '%s'", ns.ID, ns.Name, strings.Join(deletedProjectNames, ", ")))

	return &namespace.DeleteResponse{}, nil
}

func (n *NamespaceSvc) IsExists(ctx context.Context, input *namespace.IsExistsRequest) (*namespace.IsExistsResponse, error) {
	var ns models.Namespace

	err := app.DB().Select("ID", "Name").Where("`name` = ?", utils.GetMarsNamespace(input.Name)).First(&ns).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &namespace.IsExistsResponse{Exists: false}, nil
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &namespace.IsExistsResponse{Exists: true, Id: int64(ns.ID)}, nil
}

func (n *NamespaceSvc) Show(ctx context.Context, id *namespace.ShowRequest) (*namespace.ShowResponse, error) {
	var ns models.Namespace

	err := app.DB().Preload("Projects").Where("`id` = ?", id.NamespaceId).First(&ns).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &namespace.ShowResponse{Namespace: ns.ProtoTransform()}, nil
}
