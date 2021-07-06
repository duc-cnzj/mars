package controllers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/pkg/event/events"
	"github.com/duc-cnzj/mars/pkg/mlog"
	"github.com/duc-cnzj/mars/pkg/models"
	"github.com/duc-cnzj/mars/pkg/response"
	"github.com/duc-cnzj/mars/pkg/scopes"
	"github.com/duc-cnzj/mars/pkg/utils"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceController 项目空间
type NamespaceController struct{}

func NewNamespaceController() *NamespaceController {
	return &NamespaceController{}
}

type SimpleProjectItem struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type NamespaceItem struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Projects []SimpleProjectItem `json:"projects"`
}

func (ns *NamespaceController) Index(ctx *gin.Context) {
	var namespaces []*models.Namespace
	utils.DB().Preload("Projects").Scopes(scopes.OrderByIdDesc()).Find(&namespaces)
	var res = make([]NamespaceItem, 0, len(namespaces))
	for _, namespace := range namespaces {
		var projects = make([]SimpleProjectItem, 0, len(namespace.Projects))

		for _, project := range namespace.Projects {
			status, err := utils.ReleaseStatus(project.Name, namespace.Name)
			if err != nil {
				mlog.Error(err)
				status = utils.StatusUnknown
			}
			projects = append(projects, SimpleProjectItem{
				ID:     project.ID,
				Name:   project.Name,
				Status: status,
			})
		}

		res = append(res, NamespaceItem{
			ID:        namespace.ID,
			Name:      namespace.Name,
			CreatedAt: namespace.CreatedAt,
			UpdatedAt: namespace.UpdatedAt,
			Projects:  projects,
		})
	}

	response.Success(ctx, http.StatusOK, res)
}

type NamespaceStoreInput struct {
	Namespace string `json:"namespace" binding:"required,namespace"`
}

func (ns *NamespaceController) Store(ctx *gin.Context) {
	var input NamespaceStoreInput
	if err := ctx.ShouldBind(&input); err != nil {
		response.Error(ctx, 422, err)
		return
	}
	input.Namespace = utils.GetMarsNamespace(input.Namespace)

	if utils.DB().Where("`name` = ?", input.Namespace).First(&models.Namespace{}).Error == nil {
		response.Error(ctx, 422, errors.New("namespace already exists"))
		return
	}

	// 创建名称空间
	create, err := utils.K8sClientSet().CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: input.Namespace}}, metav1.CreateOptions{})
	if err != nil {
		response.Error(ctx, 500, err)
		mlog.Error(err)
		return
	}

	var imagePullSecrets []string
	for _, secret := range utils.Config().ImagePullSecrets {
		s, err := utils.CreateDockerSecret(input.Namespace, secret.Username, secret.Password, secret.Email)
		if err != nil {
			mlog.Error(err)
			continue
		}
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	mlog.Debug("成功创建namespace: ", create.Name)
	data := models.Namespace{Name: create.Name, ImagePullSecrets: strings.Join(imagePullSecrets, ",")}

	utils.DB().Create(&data)

	utils.Event().Dispatch(events.EventNamespaceCreated, events.NamespaceCreatedData{
		NsModel:  &data,
		NsK8sObj: create,
	})

	response.Success(ctx, http.StatusCreated, data)
}

type NamespaceUri struct {
	NamespaceId int `uri:"namespace_id"`
}

func (ns *NamespaceController) Destroy(ctx *gin.Context) {
	var (
		input     NamespaceUri
		namespace models.Namespace
	)
	if err := ctx.ShouldBindUri(&input); err != nil {
		response.Error(ctx, 422, err)
		return
	}

	// 删除空间前，要先删除空间下的项目
	if utils.DB().Preload("Projects").Where("`id` = ?", input.NamespaceId).First(&namespace).Error == nil {
		wg := sync.WaitGroup{}
		wg.Add(len(namespace.Projects))
		for _, project := range namespace.Projects {
			go func(releaseName, namespace string) {
				defer wg.Done()
				mlog.Debugf("delete release %s namespace %s", releaseName, namespace)
				if err := utils.UninstallRelease(releaseName, namespace); err != nil {
					mlog.Error(err)
					return
				}
			}(project.Name, namespace.Name)
		}
		wg.Wait()
		for _, secret := range namespace.ImagePullSecretsArray() {
			mlog.Debugf("delete namespace %s secret %s", namespace.Name, secret)
			utils.K8sClientSet().CoreV1().Secrets(namespace.Name).Delete(context.Background(), secret, metav1.DeleteOptions{})
		}
		if err := utils.K8sClientSet().CoreV1().Namespaces().Delete(context.Background(), namespace.Name, metav1.DeleteOptions{}); err != nil {
			mlog.Error("删除 namespace 出现错误: ", err)
		}
		if len(namespace.Projects) > 0 {
			utils.DB().Delete(&namespace.Projects)
		}
		utils.DB().Delete(&namespace)
	}

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
LABEL:
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			if _, err := utils.K8sClientSet().CoreV1().Namespaces().Get(context.Background(), namespace.Name, metav1.GetOptions{}); err != nil {
				mlog.Error(err)
				break LABEL
			}
		case <-timer.C:
			break LABEL
		}
	}

	utils.Event().Dispatch(events.EventNamespaceDeleted, events.NamespaceDeletedData{NsModel: &namespace})

	response.Success(ctx, http.StatusNoContent, "")
}

func (*NamespaceController) CpuAndMemory(ctx *gin.Context) {
	var (
		input     NamespaceUri
		namespace models.Namespace
	)
	if err := ctx.ShouldBindUri(&input); err != nil {
		response.Error(ctx, 422, err)
		return
	}

	if err := utils.DB().Preload("Projects").Where("`id` = ?", input.NamespaceId).First(&namespace).Error; err != nil {
		response.Error(ctx, 500, err)

		return
	}

	cpu, memory := utils.GetCpuAndMemoryInNamespace(namespace.Name)

	response.Success(ctx, 200, gin.H{
		"cpu":    cpu,
		"memory": memory,
	})
}

type ServiceEndpointsQuery struct {
	ProjectName string `form:"project_name"`
}

func (*NamespaceController) ServiceEndpoints(ctx *gin.Context) {
	var (
		input     NamespaceUri
		query     ServiceEndpointsQuery
		namespace models.Namespace
	)
	if err := ctx.ShouldBindUri(&input); err != nil {
		response.Error(ctx, 422, err)
		return
	}

	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.Error(ctx, 400, err)
		return
	}

	if err := utils.DB().Preload("Projects").Where("`id` = ?", input.NamespaceId).First(&namespace).Error; err != nil {
		response.Error(ctx, 500, err)

		return
	}

	var res = map[string][]string{}
	nodePortMapping := utils.GetNodePortMappingByNamespace(namespace.Name)
	ingMapping := utils.GetIngressMappingByNamespace(namespace.Name)
	for projectName, hosts := range nodePortMapping {
		var items []string = hosts
		if v, ok := ingMapping[projectName]; ok {
			items = append(v, hosts...)
		}
		res[projectName] = items
	}
	for projectName, hosts := range ingMapping {
		if _, ok := res[projectName]; ok {
			continue
		}
		res[projectName] = append([]string{}, hosts...)
	}

	if query.ProjectName != "" {
		response.Success(ctx, 200, gin.H{query.ProjectName: res[query.ProjectName]})
		return
	}

	response.Success(ctx, 200, res)
}
