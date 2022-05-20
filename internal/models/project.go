package models

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils/date"

	"gorm.io/gorm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type Project struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Name             string `json:"name" gorm:"size:100;not null;comment:项目名"`
	GitProjectId     int    `json:"git_project_id" gorm:"not null;type:integer;"`
	GitBranch        string `json:"git_branch" gorm:"not null;size:255;"`
	GitCommit        string `json:"git_commit" gorm:"not null;size:255;"`
	Config           string `json:"config"`
	OverrideValues   string `json:"override_values"`
	DockerImage      string `json:"docker_image" gorm:"not null;size:255;default:''"`
	PodSelectors     string `json:"pod_selectors" gorm:"type:text;nullable;"`
	NamespaceId      int    `json:"namespace_id"`
	Atomic           bool   `json:"atomic"`
	DeployStatus     uint8  `json:"deploy_status" gorm:"not null;default:0"`
	EnvValues        string `json:"env_values" gorm:"type:text;nullable;comment:可用的环境变量值"`
	ExtraValues      string `json:"extra_values" gorm:"type:text;nullable;comment:用户表单传入的额外值"`
	FinalExtraValues string `json:"final_extra_values" gorm:"type:text;nullable;comment:用户表单传入的额外值 + 系统默认的额外值"`

	ConfigType string `json:"config_type" gorm:"size:255;nullable;"`

	GitCommitWebUrl string     `json:"git_commit_web_url" gorm:"size:255;nullable;"`
	GitCommitTitle  string     `json:"git_commit_title" gorm:"size:255;nullable;"`
	GitCommitAuthor string     `json:"git_commit_author" gorm:"size:255;nullable;"`
	GitCommitDate   *time.Time `json:"git_commit_date"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	Namespace Namespace
}

func (project *Project) GetExtraValues() (res []*types.ExtraValue) {
	json.Unmarshal([]byte(project.ExtraValues), &res)
	return res
}

func (project *Project) GetEnvValues() (res map[string]any) {
	json.Unmarshal([]byte(project.EnvValues), &res)
	return res
}

func (project *Project) SetPodSelectors(selectors []string) {
	project.PodSelectors = strings.Join(selectors, "|")
}

// GetPodSelectors 不仅包括 deployment 的 pod 还包括其他的 stateful sets...
func (project *Project) GetPodSelectors() []string {
	return strings.Split(project.PodSelectors, "|")
}

func (project *Project) GetAllPods() []v1.Pod {
	var list []corev1.Pod
	var newList []corev1.Pod
	var split []string
	if len(project.PodSelectors) > 0 {
		split = strings.Split(project.PodSelectors, "|")
	}
	if len(split) == 0 {
		return nil
	}
	notEqualSelector := fields.OneTermNotEqualSelector("status.phase", string(v1.PodFailed))
	for _, labels := range split {
		l, _ := app.K8sClientSet().CoreV1().Pods(project.Namespace.Name).List(context.Background(), metav1.ListOptions{
			LabelSelector: labels,
			FieldSelector: notEqualSelector.String(),
		})

		list = append(list, l.Items...)
	}
	var m = make(map[string]*appsv1.ReplicaSet)
	for _, pod := range list {
		flag := true
		for _, reference := range pod.OwnerReferences {
			if reference.Kind == "ReplicaSet" {
				var (
					rs  *appsv1.ReplicaSet
					err error
				)
				if _, ok := m[string(reference.UID)]; !ok {
					rs, err = app.K8sClientSet().AppsV1().ReplicaSets(pod.Namespace).Get(context.TODO(), reference.Name, metav1.GetOptions{})
					if err != nil {
						mlog.Debug(err)
						continue
					}
					m[string(reference.UID)] = rs
				}
				if rs.Spec.Replicas != nil && *rs.Spec.Replicas == 0 {
					flag = false
					break
				}
			}
		}
		if flag {
			newList = append(newList, pod)
		}
	}

	return newList
}

func (project *Project) GetAllPodMetrics() []v1beta1.PodMetrics {
	app.DB().Preload("Namespace").First(&project)
	metricses := app.K8sMetrics().MetricsV1beta1().PodMetricses(project.Namespace.Name)
	var list []v1beta1.PodMetrics
	var split []string
	if len(project.PodSelectors) > 0 {
		split = strings.Split(project.PodSelectors, "|")
	}
	if len(split) == 0 {
		return nil
	}
	for _, labels := range split {
		l, _ := metricses.List(context.Background(), metav1.ListOptions{
			LabelSelector: labels,
		})

		list = append(list, l.Items...)
	}

	return list
}

func (project *Project) ProtoTransform() *types.ProjectModel {
	return &types.ProjectModel{
		Id:                int64(project.ID),
		Name:              project.Name,
		GitProjectId:      int64(project.GitProjectId),
		GitBranch:         project.GitBranch,
		GitCommit:         project.GitCommit,
		Config:            project.Config,
		OverrideValues:    project.OverrideValues,
		DockerImage:       project.DockerImage,
		PodSelectors:      project.PodSelectors,
		NamespaceId:       int64(project.NamespaceId),
		Atomic:            project.Atomic,
		EnvValues:         project.EnvValues,
		ExtraValues:       project.GetExtraValues(),
		FinalExtraValues:  project.FinalExtraValues,
		DeployStatus:      types.Deploy(project.DeployStatus),
		HumanizeCreatedAt: date.ToHumanizeDatetimeString(&project.CreatedAt),
		HumanizeUpdatedAt: date.ToHumanizeDatetimeString(&project.UpdatedAt),
		ConfigType:        project.ConfigType,
		GitCommitWebUrl:   project.GitCommitWebUrl,
		GitCommitTitle:    project.GitCommitTitle,
		GitCommitAuthor:   project.GitCommitAuthor,
		GitCommitDate:     date.ToHumanizeDatetimeString(project.GitCommitDate),
		Namespace:         project.Namespace.ProtoTransform(),
		CreatedAt:         date.ToRFC3339DatetimeString(&project.CreatedAt),
		UpdatedAt:         date.ToRFC3339DatetimeString(&project.UpdatedAt),
		DeletedAt:         date.ToRFC3339DatetimeString(&project.DeletedAt.Time),
	}
}
