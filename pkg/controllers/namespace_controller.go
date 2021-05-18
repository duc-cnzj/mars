package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/DuC-cnZj/mars/pkg/event/events"
	"github.com/DuC-cnZj/mars/pkg/mlog"
	"github.com/DuC-cnZj/mars/pkg/models"
	"github.com/DuC-cnZj/mars/pkg/response"
	"github.com/DuC-cnZj/mars/pkg/scopes"
	"github.com/DuC-cnZj/mars/pkg/utils"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceController 项目空间
type NamespaceController struct{}

func NewNamespaceController() *NamespaceController {
	return &NamespaceController{}
}

type NamespaceItem struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
}

func (ns *NamespaceController) Index(ctx *gin.Context) {
	var namespaces []*models.Namespace
	utils.DB().Scopes(scopes.OrderByIdDesc()).Find(&namespaces)
	var res = make([]NamespaceItem, 0, len(namespaces))

	for _, namespace := range namespaces {
		cpu, memory := utils.GetCpuAndMemoryInNamespace(namespace.Name)
		res = append(res, NamespaceItem{
			ID:        namespace.ID,
			Name:      namespace.Name,
			CreatedAt: namespace.CreatedAt,
			UpdatedAt: namespace.UpdatedAt,
			Cpu:       cpu,
			Memory:    memory,
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

	// 创建名称空间
	create, err := utils.K8s().CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: input.Namespace}}, metav1.CreateOptions{})
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}
	mlog.Debug("成功创建namespace: ", create.Name)
	data := models.Namespace{Name: create.Name}
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

	if utils.DB().Where("`id` = ?", input.NamespaceId).First(&namespace).Error == nil {
		if err := utils.K8s().CoreV1().Namespaces().Delete(context.Background(), namespace.Name, metav1.DeleteOptions{}); err != nil {
			mlog.Error("删除 namespace 出现错误: ", err)
		}
		utils.DB().Delete(&namespace)
	}

	utils.Event().Dispatch(events.EventNamespaceDeleted, events.NamespaceDeletedData{NsModel: &namespace})

	response.Success(ctx, http.StatusNoContent, "")
}
