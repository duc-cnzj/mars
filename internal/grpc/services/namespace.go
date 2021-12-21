package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/event"
	"github.com/duc-cnzj/mars/pkg/namespace"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Namespace struct {
	namespace.UnimplementedNamespaceServer
}

func (n *Namespace) Index(ctx context.Context, empty *emptypb.Empty) (*namespace.NamespaceList, error) {
	var namespaces []*models.Namespace
	app.DB().Preload("Projects").Find(&namespaces)
	var res = &namespace.NamespaceList{Data: make([]*namespace.NamespaceItem, 0, len(namespaces))}
	for _, ns := range namespaces {
		var projects = make([]*namespace.NamespaceItem_SimpleProjectItem, 0, len(ns.Projects))

		for _, project := range ns.Projects {
			status, err := utils.ReleaseStatus(project.Name, ns.Name)
			if err != nil {
				mlog.Error(err)
				status = utils.StatusUnknown
			}
			projects = append(projects, &namespace.NamespaceItem_SimpleProjectItem{
				Id:     int64(project.ID),
				Name:   project.Name,
				Status: status,
			})
		}

		res.Data = append(res.Data, &namespace.NamespaceItem{
			Id:        int64(ns.ID),
			Name:      ns.Name,
			CreatedAt: timestamppb.New(ns.CreatedAt),
			UpdatedAt: timestamppb.New(ns.UpdatedAt),
			Projects:  projects,
		})
	}

	return res, nil
}

func (n *Namespace) Store(ctx context.Context, request *namespace.NsStoreRequest) (*namespace.NsStoreResponse, error) {
	request.Namespace = utils.GetMarsNamespace(request.Namespace)

	if app.DB().Where("`name` = ?", request.Namespace).First(&models.Namespace{}).Error == nil {
		return nil, errors.New("名称空间已存在")
	}

	// 创建名称空间
	create, err := app.K8sClientSet().CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: request.Namespace}}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	var imagePullSecrets []string
	for _, secret := range app.Config().ImagePullSecrets {
		s, err := utils.CreateDockerSecret(request.Namespace, secret.Username, secret.Password, secret.Email, secret.Server)
		if err != nil {
			mlog.Error(err)
			continue
		}
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	mlog.Debug("成功创建namespace: ", create.Name)
	data := models.Namespace{Name: create.Name, ImagePullSecrets: strings.Join(imagePullSecrets, ",")}

	app.DB().Create(&data)

	app.Event().Dispatch(events.EventNamespaceCreated, events.NamespaceCreatedData{
		NsModel:  &data,
		NsK8sObj: create,
	})
	AuditLog(MustGetUser(ctx).Name, event.ActionType_Create, fmt.Sprintf("创建项目空间: %d: %s", data.ID, data.Name))

	return &namespace.NsStoreResponse{
		Data: &namespace.NamespaceResponse{
			Id:               int64(data.ID),
			Name:             data.Name,
			ImagePullSecrets: data.ImagePullSecretsArray(),
			CreatedAt:        timestamppb.New(data.CreatedAt),
			UpdatedAt:        timestamppb.New(data.UpdatedAt),
			DeletedAt:        timestamppb.New(data.DeletedAt.Time),
		},
	}, nil
}

func (n *Namespace) CpuAndMemory(ctx context.Context, id *namespace.NamespaceID) (*namespace.CpuAndMemoryResponse, error) {
	var ns models.Namespace
	if err := app.DB().Preload("Projects").Where("`id` = ?", id.NamespaceId).First(&ns).Error; err != nil {
		return nil, err
	}

	cpu, memory := utils.GetCpuAndMemoryInNamespace(ns.Name)

	return &namespace.CpuAndMemoryResponse{
		Cpu:    cpu,
		Memory: memory,
	}, nil
}

func (n *Namespace) ServiceEndpoints(ctx context.Context, id *namespace.ServiceEndpointsRequest) (*namespace.ServiceEndpointsResponse, error) {
	var ns models.Namespace
	if err := app.DB().Preload("Projects").Where("`id` = ?", id.NamespaceId).First(&ns).Error; err != nil {
		return nil, err
	}

	var res = []*namespace.ServiceEndpointsResponseItem{}
	nodePortMapping := utils.GetNodePortMappingByNamespace(ns.Name)
	ingMapping := utils.GetIngressMappingByNamespace(ns.Name)
	for projectName, hosts := range nodePortMapping {
		var items []string = hosts
		if v, ok := ingMapping[projectName]; ok {
			items = append(v, hosts...)
		}
		res = append(res, &namespace.ServiceEndpointsResponseItem{
			Name: projectName,
			Url:  items,
		})
	}
	for projectName, hosts := range ingMapping {
		for _, re := range res {
			if re.Name == projectName {
				continue
			}
		}
		res = append(res, &namespace.ServiceEndpointsResponseItem{
			Name: projectName,
			Url:  hosts,
		})
	}

	if id.ProjectName != "" {
		var data = []*namespace.ServiceEndpointsResponseItem{}
		for _, re := range res {
			if re.Name == id.ProjectName {
				data = []*namespace.ServiceEndpointsResponseItem{re}
			}
		}
		return &namespace.ServiceEndpointsResponse{Data: data}, nil
	}

	return &namespace.ServiceEndpointsResponse{Data: res}, nil
}

func (n *Namespace) Destroy(ctx context.Context, id *namespace.NamespaceID) (*emptypb.Empty, error) {
	var ns models.Namespace
	// 删除空间前，要先删除空间下的项目
	if app.DB().Preload("Projects").Where("`id` = ?", id.NamespaceId).First(&ns).Error == nil {
		wg := sync.WaitGroup{}
		wg.Add(len(ns.Projects))
		for _, project := range ns.Projects {
			go func(releaseName, namespace string) {
				defer wg.Done()
				mlog.Debugf("delete release %s namespace %s", releaseName, namespace)
				if err := utils.UninstallRelease(releaseName, namespace, mlog.Debugf); err != nil {
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

	AuditLog(MustGetUser(ctx).Name, event.ActionType_Delete, fmt.Sprintf("删除项目空间: id: %d %s", ns.ID, ns.Name))

	return &emptypb.Empty{}, nil
}
