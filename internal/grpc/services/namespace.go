package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/client/event"
	modelspb "github.com/duc-cnzj/mars/client/model"
	"github.com/duc-cnzj/mars/client/namespace"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var sf utils.SingleflightGroup

type Namespace struct {
	namespace.UnimplementedNamespaceServer
}

func (n *Namespace) All(ctx context.Context, request *namespace.NamespaceAllRequest) (*namespace.NamespaceAllResponse, error) {
	var namespaces []*models.Namespace
	app.DB().Preload("Projects").Find(&namespaces)
	var res = &namespace.NamespaceAllResponse{Data: make([]*namespace.NamespaceItem, 0, len(namespaces))}
	releaseStatus := utils.ReleaseList{}
	if len(namespaces) > 0 {
		do, _, _ := sf.Do("ListRelease", func() (interface{}, error) {
			mlog.Debug("ListRelease.....")
			return utils.ListRelease()
		})
		releaseStatus = do.(utils.ReleaseList)
	}
	for _, ns := range namespaces {
		var projects = make([]*namespace.NamespaceSimpleProject, 0, len(ns.Projects))

		for _, project := range ns.Projects {
			projects = append(projects, &namespace.NamespaceSimpleProject{
				Id:     int64(project.ID),
				Name:   project.Name,
				Status: releaseStatus.GetStatus(ns.Name, project.Name),
			})
		}

		res.Data = append(res.Data, &namespace.NamespaceItem{
			Id:        int64(ns.ID),
			Name:      ns.Name,
			CreatedAt: utils.ToRFC3339DatetimeString(&ns.CreatedAt),
			UpdatedAt: utils.ToRFC3339DatetimeString(&ns.UpdatedAt),
			Projects:  projects,
		})
	}

	return res, nil
}

func (n *Namespace) Create(ctx context.Context, request *namespace.NamespaceCreateRequest) (*namespace.NamespaceCreateResponse, error) {
	request.Namespace = utils.GetMarsNamespace(request.Namespace)

	if app.DB().Where("`name` = ?", request.Namespace).First(&models.Namespace{}).Error == nil {
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

	return &namespace.NamespaceCreateResponse{
		Id:               int64(data.ID),
		Name:             data.Name,
		ImagePullSecrets: data.ImagePullSecretsArray(),
		CreatedAt:        utils.ToRFC3339DatetimeString(&data.CreatedAt),
		UpdatedAt:        utils.ToRFC3339DatetimeString(&data.UpdatedAt),
	}, nil
}

func (n *Namespace) CpuMemory(ctx context.Context, id *namespace.NamespaceCpuMemoryRequest) (*namespace.NamespaceCpuMemoryResponse, error) {
	var ns models.Namespace
	if err := app.DB().Preload("Projects").Where("`id` = ?", id.NamespaceId).First(&ns).Error; err != nil {
		return nil, err
	}

	cpu, memory := utils.GetCpuAndMemoryInNamespace(ns.Name)

	return &namespace.NamespaceCpuMemoryResponse{
		Cpu:    cpu,
		Memory: memory,
	}, nil
}

func (n *Namespace) ServiceEndpoints(ctx context.Context, id *namespace.NamespaceServiceEndpointsRequest) (*namespace.NamespaceServiceEndpointsResponse, error) {
	var ns models.Namespace
	if err := app.DB().Preload("Projects").Where("`id` = ?", id.NamespaceId).First(&ns).Error; err != nil {
		return nil, err
	}

	var res = []*namespace.NamespaceServiceEndpointsResponseItem{}
	nodePortMapping := utils.GetNodePortMappingByNamespace(ns.Name)
	ingMapping := utils.GetIngressMappingByNamespace(ns.Name)
	for projectName, hosts := range nodePortMapping {
		var items []string = hosts
		if v, ok := ingMapping[projectName]; ok {
			items = append(v, hosts...)
		}
		res = append(res, &namespace.NamespaceServiceEndpointsResponseItem{
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
		res = append(res, &namespace.NamespaceServiceEndpointsResponseItem{
			Name: projectName,
			Url:  hosts,
		})
	}

	if id.ProjectName != "" {
		var data = []*namespace.NamespaceServiceEndpointsResponseItem{}
		for _, re := range res {
			if re.Name == id.ProjectName {
				data = []*namespace.NamespaceServiceEndpointsResponseItem{re}
			}
		}
		return &namespace.NamespaceServiceEndpointsResponse{Data: data}, nil
	}

	return &namespace.NamespaceServiceEndpointsResponse{Data: res}, nil
}

func (n *Namespace) Delete(ctx context.Context, id *namespace.NamespaceDeleteRequest) (*namespace.NamespaceDeleteResponse, error) {
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

	return &namespace.NamespaceDeleteResponse{}, nil
}

func (n *Namespace) IsExists(ctx context.Context, input *namespace.NamespaceIsExistsRequest) (*namespace.NamespaceIsExistsResponse, error) {
	var ns models.Namespace

	err := app.DB().Select("ID", "Name").Where("`name` = ?", utils.GetMarsNamespace(input.Name)).First(&ns).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &namespace.NamespaceIsExistsResponse{Exists: false}, nil
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &namespace.NamespaceIsExistsResponse{Exists: true, Id: int64(ns.ID)}, nil
}

func (n *Namespace) Show(ctx context.Context, id *namespace.NamespaceShowRequest) (*namespace.NamespaceShowResponse, error) {
	var ns models.Namespace

	err := app.DB().Preload("Projects").Where("`id` = ?", id.NamespaceId).First(&ns).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}
	ps := make([]*modelspb.ProjectModel, 0, len(ns.Projects))
	for _, project := range ns.Projects {
		ps = append(ps, &modelspb.ProjectModel{
			Id:              int64(project.ID),
			Name:            project.Name,
			GitlabProjectId: int64(project.GitlabProjectId),
			GitlabBranch:    project.GitlabBranch,
			GitlabCommit:    project.GitlabCommit,
			Config:          project.Config,
			OverrideValues:  project.OverrideValues,
			DockerImage:     project.DockerImage,
			PodSelectors:    project.PodSelectors,
			NamespaceId:     int64(project.NamespaceId),
			Atomic:          project.Atomic,
			CreatedAt:       utils.ToRFC3339DatetimeString(&project.CreatedAt),
			UpdatedAt:       utils.ToRFC3339DatetimeString(&project.UpdatedAt),
		})
	}

	return &namespace.NamespaceShowResponse{
		Id:               int64(ns.ID),
		Name:             ns.Name,
		ImagePullSecrets: ns.ImagePullSecretsArray(),
		CreatedAt:        utils.ToRFC3339DatetimeString(&ns.CreatedAt),
		UpdatedAt:        utils.ToRFC3339DatetimeString(&ns.UpdatedAt),
		Projects:         ps,
	}, nil
}
