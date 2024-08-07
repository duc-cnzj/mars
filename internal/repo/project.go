package repo

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/annotation"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/project"
	"github.com/duc-cnzj/mars/v4/internal/filters"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/pagination"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type Project struct {
	ID               int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	Name             string
	GitProjectID     int
	GitBranch        string
	GitCommit        string
	Config           string
	OverrideValues   string
	DockerImage      []string
	PodSelectors     []string
	Atomic           bool
	DeployStatus     types.Deploy
	EnvValues        []*types.KeyValue
	ExtraValues      []*websocket_pb.ExtraValue
	FinalExtraValues []string
	Version          int
	ConfigType       string
	Manifest         []string
	GitCommitWebURL  string
	GitCommitTitle   string
	GitCommitAuthor  string
	GitCommitDate    *time.Time
	NamespaceID      int

	Namespace *Namespace
}

func ToProject(project *ent.Project) *Project {
	if project == nil {

		return nil
	}
	return &Project{
		ID:               project.ID,
		CreatedAt:        project.CreatedAt,
		UpdatedAt:        project.UpdatedAt,
		DeletedAt:        project.DeletedAt,
		Name:             project.Name,
		GitProjectID:     project.GitProjectID,
		GitBranch:        project.GitBranch,
		GitCommit:        project.GitCommit,
		Config:           project.Config,
		OverrideValues:   project.OverrideValues,
		DockerImage:      project.DockerImage,
		PodSelectors:     project.PodSelectors,
		Atomic:           project.Atomic,
		DeployStatus:     project.DeployStatus,
		EnvValues:        project.EnvValues,
		ExtraValues:      project.ExtraValues,
		FinalExtraValues: project.FinalExtraValues,
		Version:          project.Version,
		ConfigType:       project.ConfigType,
		Manifest:         project.Manifest,
		GitCommitWebURL:  project.GitCommitWebURL,
		GitCommitTitle:   project.GitCommitTitle,
		GitCommitAuthor:  project.GitCommitAuthor,
		GitCommitDate:    project.GitCommitDate,
		NamespaceID:      project.NamespaceID,
		Namespace:        ToNamespace(project.Edges.Namespace),
	}
}

type ProjectRepo interface {
	GetAllPods(project *Project) SortStatePod
	GetAllPodMetrics(project *Project) []v1beta1.PodMetrics
	GetNodePortMappingByProjects(namespace string, projects ...*Project) EndpointMapping
	GetLoadBalancerMappingByProjects(namespace string, projects ...*Project) EndpointMapping
	GetIngressMappingByProjects(namespace string, projects ...*Project) EndpointMapping

	List(ctx context.Context, input *ListProjectInput) ([]*Project, *pagination.Pagination, error)
	Create(ctx context.Context, project *CreateProjectInput) (*Project, error)
	Show(ctx context.Context, id int) (*Project, error)
	Delete(ctx context.Context, id int) error
	FindByName(ctx context.Context, name string, nsID int) (*Project, error)
	UpdateDeployStatus(ctx context.Context, id int, status types.Deploy) (*Project, error)
	UpdateVersion(ctx context.Context, id int, version int) (*Project, error)
	FindByVersion(ctx context.Context, id, version int) (*Project, error)
	UpdateStatusByVersion(ctx context.Context, id int, status types.Deploy, version int) (*Project, error)

	UpdateProject(ctx context.Context, input *UpdateProjectInput) (*Project, error)
}

var _ ProjectRepo = (*projectRepo)(nil)

type projectRepo struct {
	logger mlog.Logger

	externalIp string
	data       data.Data
}

func NewProjectRepo(logger mlog.Logger, data data.Data) ProjectRepo {
	return &projectRepo{
		logger:     logger,
		externalIp: data.Config().ExternalIp,
		data:       data,
	}
}

type ListProjectInput struct {
	Page          int32
	PageSize      int32
	OrderByIDDesc *bool
}

func (repo *projectRepo) List(ctx context.Context, input *ListProjectInput) ([]*Project, *pagination.Pagination, error) {
	query := repo.data.DB().Project.Query().
		WithNamespace().
		Where(filters.IfOrderByDesc("id")(input.OrderByIDDesc))
	all := query.Clone().
		Offset(pagination.GetPageOffset(input.Page, input.PageSize)).
		Limit(int(input.PageSize)).
		AllX(ctx)
	count := query.Clone().CountX(ctx)
	return serialize.Serialize(all, ToProject), pagination.NewPagination(input.Page, input.PageSize, count), nil
}

type CreateProjectInput struct {
	Name         string
	GitProjectID int
	GitBranch    string
	GitCommit    string
	Config       string
	Atomic       *bool
	ConfigType   string
	NamespaceID  int
	Manifest     []string
	PodSelectors []string
	DeployStatus types.Deploy
}

func (repo *projectRepo) Create(ctx context.Context, input *CreateProjectInput) (*Project, error) {
	save, err := repo.data.DB().Project.Create().
		SetName(input.Name).
		SetGitProjectID(input.GitProjectID).
		SetGitBranch(input.GitBranch).
		SetGitCommit(input.GitCommit).
		SetConfig(input.Config).
		SetNillableAtomic(input.Atomic).
		SetConfigType(input.ConfigType).
		SetNamespaceID(input.NamespaceID).
		SetManifest(input.Manifest).
		SetPodSelectors(input.PodSelectors).
		Save(ctx)
	return ToProject(save), err
}

type UpdateProjectInput struct {
	ID         int
	GitBranch  string
	GitCommit  string
	Config     string
	Atomic     *bool
	ConfigType string

	PodSelectors     []string
	Manifest         []string
	DockerImage      []string
	GitCommitTitle   string
	GitCommitWebURL  string
	GitCommitAuthor  string
	GitCommitDate    *time.Time
	ExtraValues      []*websocket_pb.ExtraValue
	FinalExtraValues []string
	EnvValues        []*types.KeyValue
	OverrideValues   string
}

func (repo *projectRepo) UpdateProject(ctx context.Context, input *UpdateProjectInput) (*Project, error) {
	first, err := repo.data.DB().Project.Query().Where(project.ID(input.ID)).First(ctx)
	if err != nil {
		return nil, err
	}
	save, err := first.Update().
		SetGitBranch(input.GitBranch).
		SetGitCommit(input.GitCommit).
		SetConfig(input.Config).
		SetNillableAtomic(input.Atomic).
		SetConfigType(input.ConfigType).
		SetPodSelectors(input.PodSelectors).
		SetManifest(input.Manifest).
		SetDockerImage(input.DockerImage).
		SetGitCommitTitle(input.GitCommitTitle).
		SetGitCommitWebURL(input.GitCommitWebURL).
		SetGitCommitAuthor(input.GitCommitAuthor).
		SetNillableGitCommitDate(input.GitCommitDate).
		SetExtraValues(input.ExtraValues).
		SetFinalExtraValues(input.FinalExtraValues).
		SetEnvValues(input.EnvValues).
		SetOverrideValues(input.OverrideValues).
		Save(ctx)
	return ToProject(save), err
}

func (repo *projectRepo) Show(ctx context.Context, id int) (*Project, error) {
	first, err := repo.data.DB().Project.Query().WithNamespace().Where(project.ID(id)).First(ctx)
	return ToProject(first), err
}

func (repo *projectRepo) Delete(ctx context.Context, id int) error {
	return repo.data.DB().Project.DeleteOneID(id).Exec(ctx)
}

func (repo *projectRepo) UpdateStatusByVersion(ctx context.Context, id int, status types.Deploy, version int) (*Project, error) {
	if _, err := repo.FindByVersion(ctx, id, version); err != nil {
		return nil, err
	}
	save, err := repo.data.DB().Project.UpdateOneID(id).SetDeployStatus(status).SetVersion(version).Save(ctx)
	return ToProject(save), err
}

func (repo *projectRepo) FindByVersion(ctx context.Context, id, version int) (*Project, error) {
	first, err := repo.data.DB().Project.Query().Where(project.ID(id), project.Version(version)).First(ctx)
	return ToProject(first), err
}

func (repo *projectRepo) UpdateVersion(ctx context.Context, id int, version int) (*Project, error) {
	save, err := repo.data.DB().Project.UpdateOneID(id).SetVersion(version).Save(ctx)
	return ToProject(save), err
}

func (repo *projectRepo) UpdateDeployStatus(ctx context.Context, id int, status types.Deploy) (*Project, error) {
	save, err := repo.data.DB().Project.UpdateOneID(id).SetDeployStatus(status).Save(ctx)
	return ToProject(save), err
}

func (repo *projectRepo) FindByName(ctx context.Context, name string, nsID int) (*Project, error) {
	first, err := repo.data.DB().Project.Query().Where(project.Name(name), project.NamespaceID(nsID)).First(ctx)
	return ToProject(first), err
}

func (repo *projectRepo) IsPodRunning(namespace, podName string) (running bool, notRunningReason string) {
	podInfo, err := repo.data.K8sClient().PodLister.Pods(namespace).Get(podName)
	if err != nil {
		return false, err.Error()
	}

	if podInfo.Status.Phase == v1.PodRunning {
		return true, ""
	}

	if podInfo.Status.Phase == v1.PodFailed && podInfo.Status.Reason == "Evicted" {
		return false, fmt.Sprintf("po %s already evicted in namespace %s!", podName, namespace)
	}

	for _, status := range podInfo.Status.ContainerStatuses {
		return false, fmt.Sprintf("%s %s", status.State.Waiting.Reason, status.State.Waiting.Message)
	}

	return false, "pod not running."
}

func (repo *projectRepo) GetNodePortMappingByProjects(namespace string, projects ...*Project) EndpointMapping {
	var (
		projectMap = make(projectObjectMap)
		k8sCli     = repo.data.K8sClient()
	)
	for _, project := range projects {
		projectMap[project.Name] = FilterRuntimeObjectFromManifests[*v1.Service](repo.logger, project.Manifest)
	}

	list, _ := k8sCli.ServiceLister.Services(namespace).List(labels.Everything())
	var m = map[string][]*types.ServiceEndpoint{}

	for _, item := range list {
		if projectName, ok := projectMap.GetProject(item); ok && item.Spec.Type == v1.ServiceTypeNodePort {
			for _, port := range item.Spec.Ports {
				data := m[projectName]

				switch {
				case isHttpPortName(port.Name):
					m[projectName] = append(data, &types.ServiceEndpoint{
						Name:     projectName,
						PortName: port.Name,
						Url:      fmt.Sprintf("http://%s:%d", repo.externalIp, port.NodePort),
					})
				default:
					m[projectName] = append(data, &types.ServiceEndpoint{
						Name:     projectName,
						PortName: port.Name,
						Url:      fmt.Sprintf("%s:%d", repo.externalIp, port.NodePort),
					})
				}
			}
		}
	}

	return m
}

func (repo *projectRepo) GetAllPods(project *Project) SortStatePod {
	var (
		list      = make(map[string]*corev1.Pod)
		newList   SortStatePod
		split     []string = project.PodSelectors
		k8sClient          = repo.data.K8sClient()
	)
	if len(split) == 0 {
		return nil
	}
	for _, ls := range split {
		selector, _ := metav1.ParseToLabelSelector(ls)
		asSelector, _ := metav1.LabelSelectorAsSelector(selector)

		l, _ := k8sClient.PodLister.Pods(project.Namespace.Name).List(asSelector)
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
					rs, err = k8sClient.ReplicaSetLister.ReplicaSets(pod.Namespace).Get(reference.Name)
					if err != nil {
						repo.logger.Debug(err)
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

		idx := pod.Annotations[annotation.PodOrderIndex]

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

func (repo *projectRepo) GetAllPodMetrics(project *Project) []v1beta1.PodMetrics {
	var (
		metrics = repo.data.K8sClient().MetricsClient
	)
	//db.Preload("Namespace").First(&project)
	metricses := metrics.MetricsV1beta1().PodMetricses(project.Namespace.Name)
	var list []v1beta1.PodMetrics
	var split []string = project.PodSelectors
	if len(split) == 0 {
		return nil
	}
	for _, labels := range split {
		l, _ := metricses.List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels,
		})

		list = append(list, l.Items...)
	}

	return list
}

func (repo *projectRepo) GetLoadBalancerMappingByProjects(namespace string, projects ...*Project) EndpointMapping {
	var projectMap = make(projectObjectMap)
	for _, project := range projects {
		projectMap[project.Name] = FilterRuntimeObjectFromManifests[*v1.Service](repo.logger, project.Manifest)
	}
	var k8sCli = repo.data.K8sClient()
	list, _ := k8sCli.ServiceLister.Services(namespace).List(labels.Everything())
	var m = EndpointMapping{}

	for _, item := range list {
		if projectName, ok := projectMap.GetProject(item); ok && item.Spec.Type == v1.ServiceTypeLoadBalancer && len(item.Status.LoadBalancer.Ingress) > 0 {
			lbIP := item.Status.LoadBalancer.Ingress[0].IP
			for _, port := range item.Spec.Ports {
				data := m[projectName]

				switch {
				case isHttpPortName(port.Name):
					var url string = fmt.Sprintf("http://%s:%d", lbIP, port.Port)
					if port.Port == 80 {
						url = fmt.Sprintf("http://%s", lbIP)
					}
					if port.Port == 443 {
						url = fmt.Sprintf("https://%s", lbIP)
					}
					m[projectName] = append(data, &types.ServiceEndpoint{
						Name:     projectName,
						PortName: port.Name,
						Url:      url,
					})
				default:
					m[projectName] = append(data, &types.ServiceEndpoint{
						Name:     projectName,
						PortName: port.Name,
						Url:      fmt.Sprintf("%s:%d", lbIP, port.Port),
					})
				}
			}
		}
	}
	m.Sort()

	return m
}

func (repo *projectRepo) GetIngressMappingByProjects(namespace string, projects ...*Project) EndpointMapping {
	var projectMap = make(projectObjectMap)
	for _, project := range projects {
		projectMap[project.Name] = FilterRuntimeObjectFromManifests[*networkingv1.Ingress](repo.logger, project.Manifest)
	}

	var m = EndpointMapping{}
	var k8sCli = repo.data.K8sClient()
	list, _ := k8sCli.IngressLister.Ingresses(namespace).List(labels.Everything())
	type Host = string
	var allHosts = make(map[Host]struct {
		projectName string
		tls         bool
	})
	for _, item := range list {
		for _, rules := range item.Spec.Rules {
			if projectName, ok := projectMap.GetProject(item); ok {
				allHosts[rules.Host] = struct {
					projectName string
					tls         bool
				}{projectName: projectName, tls: false}
			}
		}
		for _, tls := range item.Spec.TLS {
			if projectName, ok := projectMap.GetProject(item); ok {
				for _, host := range tls.Hosts {
					allHosts[host] = struct {
						projectName string
						tls         bool
					}{projectName: projectName, tls: true}
				}
			}
		}
	}
	for host, data := range allHosts {
		urlScheme := "http"
		if data.tls {
			urlScheme = "https"
		}
		m[data.projectName] = append(m[data.projectName], &types.ServiceEndpoint{
			Name: data.projectName,
			Url:  fmt.Sprintf("%s://%s", urlScheme, host),
		})
	}
	m.Sort()

	return m
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

func mustInt(num string) (res int) {
	var err error
	res, err = strconv.Atoi(num)
	if err != nil {
		res = 0
	}
	return
}

func FilterRuntimeObjectFromManifests[T runtime.Object](logger mlog.Logger, manifests []string) RuntimeObjectList {
	var m = make(RuntimeObjectList, 0)
	info, _ := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
	for _, f := range manifests {
		obj, _, err := info.Serializer.Decode([]byte(f), nil, nil)
		if err != nil {
			logger.Warning(err.Error())
			continue
		}
		switch obj.(type) {
		case T:
			m = append(m, obj)
		}
	}

	return m
}

func isHttpPortName(name string) bool {
	switch {
	case strings.Contains(name, "web"):
		fallthrough
	case strings.Contains(name, "ui"):
		fallthrough
	case strings.Contains(name, "api"):
		fallthrough
	case strings.Contains(name, "http"):
		return true
	default:
		return false
	}
}

type RuntimeObjectList []runtime.Object

type projectObjectMap map[string]RuntimeObjectList

func (l RuntimeObjectList) Has(in runtime.Object) bool {
	inAccessor, _ := meta.Accessor(in)
	for _, set := range l {
		accessor, _ := meta.Accessor(set)
		if reflect.TypeOf(set) == reflect.TypeOf(in) && accessor.GetName() == inAccessor.GetName() {
			return true
		}
	}

	return false
}

func (m projectObjectMap) GetProject(svc runtime.Object) (string, bool) {
	for projectName, set := range m {
		if set.Has(svc) {
			return projectName, true
		}
	}
	return "", false
}

type sortEndpoint []*types.ServiceEndpoint

func (s sortEndpoint) Len() int {
	return len(s)
}

func (s sortEndpoint) Less(i, j int) bool {
	return strings.HasPrefix(s[i].Url, "https") && !strings.HasPrefix(s[j].Url, "https")
}

func (s sortEndpoint) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type EndpointMapping map[string][]*types.ServiceEndpoint

func (e EndpointMapping) Sort() {
	for _, endpoints := range e {
		sort.Sort(sortEndpoint(endpoints))
	}
}

func (e EndpointMapping) Get(projName string) []*types.ServiceEndpoint {
	return e[projName]
}

func (e EndpointMapping) AllEndpoints() []*types.ServiceEndpoint {
	var res = make([]*types.ServiceEndpoint, 0)
	for _, endpoints := range e {
		res = append(res, endpoints...)
	}
	return res
}
