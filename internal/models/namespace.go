package models

import (
	"strings"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"

	"gorm.io/gorm"
)

type Namespace struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Name             string `json:"name" gorm:"size:100;not null;comment:项目空间名"`
	ImagePullSecrets string `json:"image_pull_secrets" gorm:"size:255;not null;default:'';comment:项目空间拉取镜像的secrets，数组"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index;"`

	Projects []Project
}

func (ns *Namespace) ImagePullSecretsArray() []string {
	if ns.ImagePullSecrets == "" {
		return []string{}
	}
	return strings.Split(ns.ImagePullSecrets, ",")
}

func (ns *Namespace) ProtoTransform() *types.NamespaceModel {
	secrets := ns.GetImagePullSecrets()
	var projects []*types.ProjectModel
	for _, project := range ns.Projects {
		projects = append(projects, project.ProtoTransform())
	}
	return &types.NamespaceModel{
		Id:               int64(ns.ID),
		Name:             ns.Name,
		ImagePullSecrets: secrets,
		Projects:         projects,
		CreatedAt:        date.ToRFC3339DatetimeString(&ns.CreatedAt),
		UpdatedAt:        date.ToRFC3339DatetimeString(&ns.UpdatedAt),
		DeletedAt:        date.ToRFC3339DatetimeString(&ns.DeletedAt.Time),
	}
}

func (ns *Namespace) GetImagePullSecrets() []*types.ImagePullSecret {
	var secrets = make([]*types.ImagePullSecret, 0)
	for _, s := range ns.ImagePullSecretsArray() {
		secrets = append(secrets, &types.ImagePullSecret{Name: s})
	}
	return secrets
}
