package models

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/annotations"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"

	"gorm.io/gorm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	DockerImage      string `json:"docker_image" gorm:"not null;size:1024;default:''"`
	PodSelectors     string `json:"pod_selectors" gorm:"type:text;nullable;"`
	NamespaceId      int    `json:"namespace_id" gorm:"index:idx_namespace_id_deleted_at,priority:1;"`
	Atomic           bool   `json:"atomic"`
	DeployStatus     uint8  `json:"deploy_status" gorm:"index:idx_deploy_status;not null;default:0"`
	EnvValues        string `json:"env_values" gorm:"type:text;nullable;comment:可用的环境变量值"`
	ExtraValues      string `json:"extra_values" gorm:"type:longtext;nullable;comment:用户表单传入的额外值"`
	FinalExtraValues string `json:"final_extra_values" gorm:"type:longtext;nullable;comment:用户表单传入的额外值 + 系统默认的额外值"`
	Version          int    `json:"version" gorm:"type:int;not null;default:1;"`

	ConfigType string `json:"config_type" gorm:"size:255;nullable;"`
	Manifest   string `json:"manifest" gorm:"type:longtext;"`

	GitCommitWebUrl string     `json:"git_commit_web_url" gorm:"size:255;nullable;"`
	GitCommitTitle  string     `json:"git_commit_title" gorm:"size:255;nullable;"`
	GitCommitAuthor string     `json:"git_commit_author" gorm:"size:255;nullable;"`
	GitCommitDate   *time.Time `json:"git_commit_date"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index:idx_namespace_id_deleted_at,priority:2;"`

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

// GetPodSelectors 不仅包括 deployment 的 pod 还包括其他的: StatefulSet、DaemonSet...
func (project *Project) GetPodSelectors() []string {
	return strings.Split(project.PodSelectors, "|")
}

type StatePod struct {
	IsOld       bool
	Terminating bool
	Pending     bool
	OrderIndex  int
	Pod         *corev1.Pod
}

type SortStatePod []StatePod

func (s SortStatePod) Len() int {
	return len(s)
}

var allStatus = map[corev1.PodPhase]int{
	corev1.PodRunning:   1,
	corev1.PodSucceeded: 2,
	corev1.PodFailed:    3,
	corev1.PodPending:   4,
	corev1.PodUnknown:   5,
}

func (s SortStatePod) Less(i, j int) bool {
	if allStatus[s[i].Pod.Status.Phase] < allStatus[s[j].Pod.Status.Phase] {
		return true
	}

	if s[i].Pod.Status.Phase == s[j].Pod.Status.Phase {
		if !s[i].IsOld && s[j].IsOld {
			return true
		}

		if s[i].OrderIndex > s[j].OrderIndex && s[i].IsOld == s[j].IsOld {
			return true
		}

		if !s[i].Terminating && s[j].Terminating && s[i].IsOld == s[j].IsOld {
			return true
		}

		if s[i].OrderIndex == s[j].OrderIndex && s[i].IsOld == s[j].IsOld && s[i].Terminating == s[j].Terminating {
			return s[i].Pod.CreationTimestamp.Time.Before(s[j].Pod.CreationTimestamp.Time)
		}
	}

	return false
}

func (s SortStatePod) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

const RevisionAnnotation = "deployment.kubernetes.io/revision"

func (project *Project) GetAllPods() SortStatePod {
	var list = make(map[string]*corev1.Pod)
	var newList SortStatePod
	var split []string
	if len(project.PodSelectors) > 0 {
		split = strings.Split(project.PodSelectors, "|")
	}
	if len(split) == 0 {
		return nil
	}
	for _, ls := range split {
		selector, _ := metav1.ParseToLabelSelector(ls)
		asSelector, _ := metav1.LabelSelectorAsSelector(selector)

		l, _ := app.K8sClient().PodLister.Pods(project.Namespace.Name).List(asSelector)
		for _, pod := range l {
			if pod.Status.Phase != corev1.PodFailed {
				list[pod.Name] = pod
			}
		}
	}

	var m = make(map[string]*appsv1.ReplicaSet)
	var objectMap = make(map[string]runtime.Object)
	var oldReplicaMap = make(map[string]struct{})

	// TODO: 兼容 sts pod
	for _, pod := range list {
		for _, reference := range pod.OwnerReferences {
			if reference.Kind == "ReplicaSet" {
				var (
					rs  *appsv1.ReplicaSet
					err error
					ok  bool
				)
				if _, ok = m[string(reference.UID)]; !ok {
					rs, err = app.K8sClient().ReplicaSetLister.ReplicaSets(pod.Namespace).Get(reference.Name)
					if err != nil {
						mlog.Debug(err)
						continue
					}
					m[string(reference.UID)] = rs
					for _, re := range rs.OwnerReferences {
						if re.Kind == "Deployment" {
							uniqueKey := string(re.UID)
							if old, found := objectMap[uniqueKey]; found {
								accessor1, _ := meta.Accessor(old)
								accessor2, _ := meta.Accessor(rs)
								accessor1Revision := accessor1.GetAnnotations()[RevisionAnnotation]
								accessor2Revision := accessor2.GetAnnotations()[RevisionAnnotation]
								if accessor1Revision != "" && accessor2Revision != "" && accessor1Revision != accessor2Revision {
									if accessor1Revision < accessor2Revision {
										oldReplicaMap[string(accessor1.GetUID())] = struct{}{}
										objectMap[uniqueKey] = rs
									} else {
										oldReplicaMap[string(accessor2.GetUID())] = struct{}{}
									}
								}
							} else {
								objectMap[uniqueKey] = rs
							}
							break
						}
					}
				}
			}
		}
	}

	for _, pod := range list {
		var isOld bool
		for _, reference := range pod.OwnerReferences {
			if _, ok := oldReplicaMap[string(reference.UID)]; ok {
				isOld = true
				break
			}
		}

		idx := pod.Annotations[annotations.PodOrderIndex]

		newList = append(newList, StatePod{
			IsOld:       isOld,
			Terminating: pod.DeletionTimestamp != nil,
			Pending:     pod.Status.Phase == corev1.PodPending,
			OrderIndex:  mustInt(idx),
			Pod:         pod.DeepCopy(),
		})
	}
	sort.Sort(newList)

	return newList
}

func mustInt(num string) (res int) {
	var err error
	res, err = strconv.Atoi(num)
	if err != nil {
		res = 0
	}
	return
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
		Version:           int64(project.Version),
		Namespace:         project.Namespace.ProtoTransform(),
		CreatedAt:         date.ToRFC3339DatetimeString(&project.CreatedAt),
		UpdatedAt:         date.ToRFC3339DatetimeString(&project.UpdatedAt),
		DeletedAt:         date.ToRFC3339DatetimeString(&project.DeletedAt.Time),
	}
}
