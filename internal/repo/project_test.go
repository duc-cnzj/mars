package repo

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/annotation"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/ent/namespace"
	"github.com/duc-cnzj/mars/v5/internal/util/rand"
	"github.com/duc-cnzj/mars/v5/internal/util/serialize"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	networkingv1lister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubectl/pkg/util/deployment"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	httproutev1 "sigs.k8s.io/gateway-api/pkg/client/listers/apis/v1"

	"github.com/duc-cnzj/mars/api/v5/types"
	data2 "github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"

	"github.com/stretchr/testify/assert"
	appsv1lister "k8s.io/client-go/listers/apps/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
)

func createRepo(db *ent.Client) *ent.Repo {
	return db.Repo.Create().SetName(rand.String(10)).SaveX(context.Background())
}
func TestProjectRepoCreate(t *testing.T) {
	ctx := context.Background()
	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)
	repo := createRepo(db)
	ns := createNamespace(db)
	input := &CreateProjectInput{
		Name:         "TestProject",
		GitProjectID: 1,
		GitBranch:    "master",
		GitCommit:    "abc123",
		Config:       "testConfig",
		Atomic:       nil,
		ConfigType:   "testConfigType",
		NamespaceID:  ns.ID,
		PodSelectors: []string{"testSelector"},
		DeployStatus: types.Deploy_StatusDeployed,
		RepoID:       repo.ID,
		Creator:      "testCreator",
	}

	project, err := r.Create(ctx, input)
	assert.NoError(t, err)
	assert.Equal(t, input.Name, project.Name)
	assert.Equal(t, input.GitProjectID, project.GitProjectID)
	assert.Equal(t, input.GitBranch, project.GitBranch)
	assert.Equal(t, input.GitCommit, project.GitCommit)
	assert.Equal(t, input.Config, project.Config)
	assert.Equal(t, input.ConfigType, project.ConfigType)
	assert.Equal(t, input.NamespaceID, project.NamespaceID)
	assert.Equal(t, input.PodSelectors, project.PodSelectors)
	assert.Equal(t, input.DeployStatus, project.DeployStatus)
	assert.Equal(t, input.RepoID, project.RepoID)
}

func TestProjectRepoUpdateProject(t *testing.T) {
	ctx := context.Background()
	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	p := createProject(db, createNamespace(db).ID)

	input := &UpdateProjectInput{
		ID:           p.ID,
		GitBranch:    "updatedBranch",
		GitCommit:    "updatedCommit",
		Config:       "updatedConfig",
		Atomic:       nil,
		ConfigType:   "updatedConfigType",
		PodSelectors: []string{"updatedSelector"},
		DockerImage:  []string{"updatedImage"},
		Manifest:     []string{"updatedManifest"},
	}

	project, err := r.UpdateProject(ctx, input)
	assert.NoError(t, err)
	assert.Equal(t, input.GitBranch, project.GitBranch)
	assert.Equal(t, input.GitCommit, project.GitCommit)
	assert.Equal(t, input.Config, project.Config)
	assert.Equal(t, input.ConfigType, project.ConfigType)
	assert.Equal(t, input.PodSelectors, project.PodSelectors)
	assert.Equal(t, input.DockerImage, project.DockerImage)
	assert.Equal(t, input.Manifest, project.Manifest)
}

func TestProjectRepoDelete(t *testing.T) {
	ctx := context.Background()
	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	err := r.Delete(ctx, 1)
	assert.Error(t, err)

	project := createProject(db, createNamespace(db).ID)
	err = r.Delete(ctx, project.ID)
	assert.Nil(t, err)
}

func TestProjectRepoFindByName(t *testing.T) {
	ctx := context.Background()
	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	p := createProject(db, createNamespace(db).ID)

	project, err := r.FindByName(ctx, p.Name, p.NamespaceID)
	assert.NoError(t, err)
	assert.Equal(t, p.Name, project.Name)
	assert.Equal(t, 1, project.NamespaceID)
}

func TestProjectRepoUpdateDeployStatus(t *testing.T) {
	ctx := context.Background()
	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	p := createProject(db, createNamespace(db).ID)

	project, err := r.UpdateDeployStatus(ctx, p.ID, types.Deploy_StatusDeploying)
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusDeploying, project.DeployStatus)
}

func TestProjectRepoUpdateVersion(t *testing.T) {
	ctx := context.Background()
	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	p := createProject(db, createNamespace(db).ID)

	project, err := r.UpdateVersion(ctx, p.ID, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, project.Version)
}

func TestIsContainerReady(t *testing.T) {
	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "testContainer",
					Ready: true,
				},
			},
		},
	}

	assert.True(t, isContainerReady(pod, "testContainer"))

	pod.Status.ContainerStatuses[0].Ready = false
	assert.False(t, isContainerReady(pod, "testContainer"))

	assert.False(t, isContainerReady(pod, "nonExistingContainer"))
}

func TestGetPreOccupiedLenByValuesYaml(t *testing.T) {
	repo := &projectRepo{}

	t.Run("returns zero when values is empty", func(t *testing.T) {
		values := ""
		got := repo.GetPreOccupiedLenByValuesYaml(values)
		assert.Equal(t, 0, got)
	})

	t.Run("returns correct length when values contains host", func(t *testing.T) {
		values := "  testHost< .Host1 >"
		got := repo.GetPreOccupiedLenByValuesYaml(values)
		assert.Equal(t, len("testHost"), got)
	})

	t.Run("returns max length when values contains multiple hosts", func(t *testing.T) {
		values := "  testHost< .Host1 >  longerTestHost< .Host2 >"
		got := repo.GetPreOccupiedLenByValuesYaml(values)
		assert.Equal(t, len("longerTestHost"), got)
	})

	t.Run("ignores non-host values", func(t *testing.T) {
		values := "  testHost< .Host1 >  nonHostValue"
		got := repo.GetPreOccupiedLenByValuesYaml(values)
		assert.Equal(t, len("testHost"), got)
	})
}

func TestIsHttpPortName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Should return true when input contains 'web'",
			input:    "web",
			expected: true,
		},
		{
			name:     "Should return true when input contains 'ui'",
			input:    "ui",
			expected: true,
		},
		{
			name:     "Should return true when input contains 'api'",
			input:    "api",
			expected: true,
		},
		{
			name:     "Should return true when input contains 'http'",
			input:    "http",
			expected: true,
		},
		{
			name:     "Should return false when input does not contain 'web', 'ui', 'api', or 'http'",
			input:    "test",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHttpPortName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortStatePod_Len(t *testing.T) {
	pods := SortStatePod{
		{Pod: &corev1.Pod{}},
		{Pod: &corev1.Pod{}},
		{Pod: &corev1.Pod{}},
	}
	assert.Equal(t, 3, pods.Len())
}

func TestSortStatePod_Swap(t *testing.T) {
	pods := SortStatePod{
		{Pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "Pod1",
			},
		}},
		{Pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "Pod2",
			},
		}},
	}
	pods.Swap(0, 1)
	assert.Equal(t, "Pod2", pods[0].Pod.Name)
	assert.Equal(t, "Pod1", pods[1].Pod.Name)
}

func TestSortStatePod_Less(t *testing.T) {
	pods := SortStatePod{
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}},
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}, Status: corev1.PodStatus{Phase: corev1.PodPending}}},
	}
	assert.True(t, pods.Less(0, 1))
	pods = SortStatePod{
		{OrderIndex: 2, Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}},
		{OrderIndex: 1, Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}, Status: corev1.PodStatus{Phase: corev1.PodPending}}},
	}
	assert.True(t, pods.Less(0, 1))

	pods = SortStatePod{
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: true},
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: false},
	}
	assert.True(t, pods.Less(1, 0))

	pods = SortStatePod{
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: true, Terminating: true},
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: true},
	}
	assert.True(t, pods.Less(1, 0))

	pods = SortStatePod{
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1", CreationTimestamp: metav1.Time{Time: time.Now()}}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: true},
		{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2", CreationTimestamp: metav1.Time{Time: time.Now().Add(-1 * time.Hour)}}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}, IsOld: true},
	}
	assert.True(t, pods.Less(1, 0))
}

func TestSortEndpoint_Len(t *testing.T) {
	endpoints := sortEndpoint{
		{Name: "Endpoint1"},
		{Name: "Endpoint2"},
		{Name: "Endpoint3"},
	}
	assert.Equal(t, 3, endpoints.Len())
}

func TestSortEndpoint_Swap(t *testing.T) {
	endpoints := sortEndpoint{
		{Name: "Endpoint1"},
		{Name: "Endpoint2"},
	}
	endpoints.Swap(0, 1)
	assert.Equal(t, "Endpoint2", endpoints[0].Name)
	assert.Equal(t, "Endpoint1", endpoints[1].Name)
}

func TestSortEndpoint_Less(t *testing.T) {
	endpoints := sortEndpoint{
		{Name: "Endpoint1", Url: "http://example.com"},
		{Name: "Endpoint2", Url: "https://example.com"},
	}
	assert.False(t, endpoints.Less(0, 1))
}

func TestRuntimeObjectList_Has(t *testing.T) {
	list := RuntimeObjectList{
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}},
	}

	t.Run("returns true when object is in list", func(t *testing.T) {
		obj := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}}
		assert.True(t, list.Has(obj))
	})

	t.Run("returns false when object is not in list", func(t *testing.T) {
		obj := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod3"}}
		assert.False(t, list.Has(obj))
	})
}

func TestProjectObjectMap_GetProject(t *testing.T) {
	mapObj := projectObjectMap{
		"Project1": RuntimeObjectList{
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}},
		},
		"Project2": RuntimeObjectList{
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod2"}},
		},
	}

	t.Run("returns project name and true when object is in map", func(t *testing.T) {
		obj := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod1"}}
		projectName, ok := mapObj.GetProject(obj)
		assert.True(t, ok)
		assert.Equal(t, "Project1", projectName)
	})

	t.Run("returns empty string and false when object is not in map", func(t *testing.T) {
		obj := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "Pod3"}}
		projectName, ok := mapObj.GetProject(obj)
		assert.False(t, ok)
		assert.Equal(t, "", projectName)
	})
}

func TestEndpointMapping_Get(t *testing.T) {
	mapping := EndpointMapping{
		"Project1": []*types.ServiceEndpoint{
			{Name: "Endpoint1", Url: "http://example.com"},
		},
		"Project2": []*types.ServiceEndpoint{
			{Name: "Endpoint2", Url: "https://example.com"},
		},
	}

	t.Run("returns endpoints for given project name", func(t *testing.T) {
		endpoints := mapping.Get("Project1")
		assert.Equal(t, 1, len(endpoints))
		assert.Equal(t, "Endpoint1", endpoints[0].Name)
	})

	t.Run("returns empty slice for non-existing project name", func(t *testing.T) {
		endpoints := mapping.Get("Project3")
		assert.Equal(t, 0, len(endpoints))
	})
}

func TestEndpointMapping_AllEndpoints(t *testing.T) {
	mapping := EndpointMapping{
		"Project1": []*types.ServiceEndpoint{
			{Name: "Endpoint1", Url: "http://example.com"},
		},
		"Project2": []*types.ServiceEndpoint{
			{Name: "Endpoint2", Url: "https://example.com"},
		},
	}

	t.Run("returns all endpoints", func(t *testing.T) {
		endpoints := mapping.AllEndpoints()
		assert.Equal(t, 2, len(endpoints))
		assert.Equal(t, "Endpoint2", endpoints[0].Name)
		assert.Equal(t, "Endpoint1", endpoints[1].Name)
	})
}

func TestEndpointMapping_Sort(t *testing.T) {
	mapping := EndpointMapping{
		"Project1": []*types.ServiceEndpoint{
			{Name: "Endpoint1", Url: "http://example.com"},
			{Name: "Endpoint2", Url: "https://example.com"},
		},
		"Project2": []*types.ServiceEndpoint{
			{Name: "Endpoint3", Url: "http://example.com"},
			{Name: "Endpoint4", Url: "https://example.com"},
		},
	}

	mapping.Sort()

	t.Run("Endpoints should be sorted by Url", func(t *testing.T) {
		for _, endpoints := range mapping {
			for i := 0; i < len(endpoints)-1; i++ {
				if strings.HasPrefix(endpoints[i].Url, "http") && strings.HasPrefix(endpoints[i+1].Url, "https") {
					t.Errorf("Endpoints are not sorted correctly")
				}
			}
		}
	})
}

func TestFilterK8sTypeFromManifest(t *testing.T) {
	data := []string{`apiVersion: v1
kind: Service
metadata:
 name: devops-misc-consul-server
 namespace: devops-aa
 labels:
   app: consul
   chart: consul-helm
   heritage: Helm
   release: devops-misc
   component: server
 annotations:
   service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
spec:
 publishNotReadyAddresses: true
 ports:
 - name: http
   port: 8500
   targetPort: 8500
`, `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: busybox
  name: busybox-56c8cc5468-fd59w
  namespace: default
spec:
  containers:
  - command:
    - sh
    - -c
    - sleep 3600;
    image: busybox:latest
    name: busybox
    resources:
      limits:
        cpu: 10m
        memory: 10Mi
      requests:
        cpu: 10m
        memory: 10Mi
`}
	res := FilterRuntimeObjectFromManifests[*corev1.Service](mlog.NewLogger(nil), data)
	assert.Len(t, res, 1)
	res1 := FilterRuntimeObjectFromManifests[*corev1.Pod](mlog.NewLogger(nil), data)
	assert.Len(t, res1, 1)
	res2 := FilterRuntimeObjectFromManifests[*corev1.Namespace](mlog.NewLogger(nil), data)
	assert.Len(t, res2, 0)
}

func TestProjectRepoList(t *testing.T) {
	ctx := context.Background()
	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	// Create some projects
	for i := 0; i < 5; i++ {
		createProject(db, createNamespace(db).ID)
	}

	// Test list with pagination
	input := &ListProjectInput{
		Page:          1,
		PageSize:      2,
		OrderByIDDesc: lo.ToPtr(true),
	}
	projects, pagination, err := r.List(ctx, input)
	assert.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.True(t, projects[1].ID < projects[0].ID)
	assert.Equal(t, int32(1), pagination.Page)
	assert.Equal(t, int32(2), pagination.PageSize)
	assert.Equal(t, int32(5), pagination.Count)

	// Test list without pagination
	input = &ListProjectInput{
		Page:     1,
		PageSize: 10,
	}
	projects, pagination, err = r.List(ctx, input)
	assert.NoError(t, err)
	assert.Len(t, projects, 5)
	assert.Equal(t, int32(1), pagination.Page)
	assert.Equal(t, int32(10), pagination.PageSize)
	assert.Equal(t, int32(5), pagination.Count)
}

func TestProjectRepoList_Empty(t *testing.T) {
	ctx := context.Background()
	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	// Test list when no projects exist
	input := &ListProjectInput{
		Page:     1,
		PageSize: 2,
	}
	projects, pagination, err := r.List(ctx, input)
	assert.NoError(t, err)
	assert.Empty(t, projects)
	assert.Equal(t, int32(1), pagination.Page)
	assert.Equal(t, int32(2), pagination.PageSize)
	assert.Equal(t, int32(0), pagination.Count)
}

func Test_projectRepo_Show(t *testing.T) {
	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	repo := createRepo(db)
	project := createProject(db, createNamespace(db).ID)
	project.Update().SetRepo(repo).SaveX(context.Background())

	show, err := r.Show(context.TODO(), project.ID)
	assert.Nil(t, err)
	assert.NotNil(t, show)
	assert.NotNil(t, show.Namespace)
	assert.NotNil(t, show.Repo)
}

func Test_projectRepo_FindByVersion(t *testing.T) {
	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	project := createProject(db, createNamespace(db).ID)
	version, err := r.FindByVersion(context.TODO(), project.ID, 1)
	assert.Nil(t, err)
	assert.NotNil(t, version)
	_, err = r.FindByVersion(context.TODO(), project.ID, 2)
	assert.Error(t, err)
}

func Test_projectRepo_GetAllPods(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()

	rs := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				deployment.RevisionAnnotation: "1",
			},
			UID:       "aaaa",
			Namespace: "test",
			Name:      "rs",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "Deployment",
					UID:  "deploy-1",
				},
			},
		},
	}
	rs2 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				deployment.RevisionAnnotation: "5",
			},
			UID:       "bbbb",
			Namespace: "test",
			Name:      "rs2",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "Deployment",
					UID:  "deploy-1",
				},
			},
		},
	}
	rs3 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				deployment.RevisionAnnotation: "2",
			},
			UID:       "cccc",
			Namespace: "test",
			Name:      "rs3",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "Deployment",
					UID:  "deploy-1",
				},
			},
		},
	}
	rs4 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				deployment.RevisionAnnotation: "4",
			},
			UID:       "dddd",
			Namespace: "test",
			Name:      "rs4",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "Deployment",
					UID:  "deploy-1",
				},
			},
		},
	}
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "test",
			Labels: map[string]string{
				"a": "a",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "rs",
					UID:        "aaaa",
				},
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app"},
			},
		},
	}
	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod4",
			Namespace: "test",
			Labels: map[string]string{
				"b": "b",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "rs3",
					UID:        "cccc",
				},
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app"},
			},
		},
	}
	pod3 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod3",
			Namespace: "test",
			Labels: map[string]string{
				"c": "c",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app"},
			},
		},
	}
	pod4 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "test",
			Labels: map[string]string{
				"b": "b",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "rs2",
					UID:        "bbbb",
				},
			},
			Annotations: map[string]string{
				annotation.IgnoreContainerNames: "x",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app"},
				{Name: "x"},
			},
		},
	}
	pod5 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod5",
			Namespace: "test",
			Labels: map[string]string{
				"b": "b",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "rs4",
					UID:        "dddd",
				},
			},
		}, Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "cpp"},
			},
		},
	}

	podWithErrorRsName := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-x-error",
			Namespace: "test",
			Labels: map[string]string{
				"b": "b",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "rs-x",
					UID:        "uid-not-exist",
				},
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app"},
			},
		},
	}

	namespace := createNamespace(db)
	project := createProject(db, namespace.ID)
	namespace.Update().SetName("test").Save(context.TODO())
	fk := fake.NewSimpleClientset(rs, rs2, rs3, rs4, pod1, pod2, pod3, pod4, pod5, podWithErrorRsName)

	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}, K8sClient: &data2.K8sClient{
		Client:           fk,
		PodLister:        NewPodLister(pod1, pod2, pod3, pod4, pod5, podWithErrorRsName),
		ReplicaSetLister: NewRsLister(rs, rs2, rs3, rs4),
	}})
	r := NewProjectRepo(logger, data)
	project.Update().SetPodSelectors(nil).Save(context.TODO())
	_, err := r.GetAllActiveContainers(context.TODO(), project.ID)

	assert.Error(t, err)

	project.Update().SetPodSelectors([]string{"a=a", "b=b"}).Save(context.TODO())
	pods, _ := r.GetAllActiveContainers(context.TODO(), project.ID)
	assert.Len(t, pods, 5)
	var oldCount int
	for _, po := range pods {
		if po.IsOld {
			oldCount++
			continue
		}
		assert.True(t, po.Pod == "pod2" || po.Pod == "pod-x-error" || po.Pod == "pod5")
	}
	assert.Equal(t, 3, oldCount)
}

func Test_projectRepo_UpdateStatusByVersion(t *testing.T) {
	logger := mlog.NewLogger(nil)
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	data := data2.NewDataImpl(&data2.NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	project := createProject(db, createNamespace(db).ID)
	project.Update().SetDeployStatus(types.Deploy_StatusDeployed).Save(context.Background())

	version, err := r.UpdateStatusByVersion(context.TODO(), project.ID, types.Deploy_StatusFailed, 2)
	assert.Error(t, err)
	assert.Nil(t, version)
	version, err = r.UpdateStatusByVersion(context.TODO(), project.ID, types.Deploy_StatusFailed, 1)
	assert.Nil(t, err)
	assert.NotNil(t, version)
}

func encodeToYaml(objs ...runtime.Object) []string {
	var results []string
	for _, obj := range objs {
		bf := bytes.Buffer{}
		info, _ := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
		info.Serializer.Encode(obj, &bf)
		results = append(results, bf.String())
	}
	return results
}

func Test_projectRepo_GetLoadBalancerMappingByProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc1 := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "duc",
			Name:      "svc1",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30000,
				},
				{
					Name:     "https",
					Protocol: "tcp",
					Port:     443,
					NodePort: 30001,
				},
				{
					Name:     "xxxx",
					Protocol: "tcp",
					Port:     8080,
					NodePort: 30005,
				},
				{
					Name:     "httpx",
					Protocol: "tcp",
					Port:     8080,
					NodePort: 30006,
				},
			},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{
						IP: "111.111.111.111",
					},
				},
			},
		},
	}
	lister := NewServiceLister(&svc1)

	data := data2.NewMockData(m)
	repo := &projectRepo{
		logger: mlog.NewLogger(nil),
		data:   data,
	}
	data.EXPECT().K8sClient().Return(&data2.K8sClient{
		ServiceLister: lister,
	})
	db, err := data2.NewSqliteDB()
	assert.Nil(t, err)
	defer db.Close()
	namespace := createNamespace(db)
	namespace.Update().SetName("duc").Save(context.TODO())
	project := createProject(db, namespace.ID)
	project = project.Update().SetName("svc1").
		SetManifest(encodeToYaml(&svc1)).
		SaveX(context.TODO())
	mapping := repo.GetLoadBalancerMappingByProjects(context.TODO(), "duc", ToProject(project))
	for _, endpoints := range mapping {
		for _, endpoint := range endpoints {
			if endpoint.Name == "http" {
				assert.Equal(t, "http://111.111.111.111", endpoint.Url)
			}
			if endpoint.Name == "https" {
				assert.Equal(t, "https://111.111.111.111", endpoint.Url)
			}
			if endpoint.Name == "xxxx" {
				assert.Equal(t, "111.111.111.111:8080", endpoint.Url)
			}
			if endpoint.Name == "httpx" {
				assert.Equal(t, "http://111.111.111.111:8080", endpoint.Url)
			}
		}
	}
}

func Test_projectRepo_GetIngressMappingByNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ing1 := networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "duc",
			Name:      "ing1",
		},
		Spec: networkingv1.IngressSpec{
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{"app1.com", "app1.io"},
					SecretName: "sec1",
				},
				{
					Hosts:      []string{"app1.org"},
					SecretName: "sec2",
				},
			},
		},
	}
	ing2 := networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "duc",
			Name:      "ing2",
		},
		Spec: networkingv1.IngressSpec{
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{"app2.org"},
					SecretName: "sec2",
				},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: "http.com",
				},
				{
					Host: "app2.org",
				},
			},
		},
	}
	ing3 := networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "duc",
			Name:      "xxx",
		},
		Spec: networkingv1.IngressSpec{
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{"xxx.org"},
					SecretName: "sec2",
				},
			},
		},
	}
	ing4 := networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "duc",
			Name:      "yyy",
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "yyy.com",
				},
				{
					Host: "zzz.com",
				},
			},
		},
	}
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	save, _ := db.Namespace.Create().SetName("duc").Save(context.Background())
	p1, _ := db.Project.Create().SetName("app1").
		SetManifest(encodeToYaml(&ing1)).
		SetNamespaceID(save.ID).
		SetCreator("").
		Save(context.Background())
	p2, _ := db.Project.Create().SetName("app2").
		SetManifest(encodeToYaml(&ing2)).
		SetNamespaceID(save.ID).
		SetCreator("").
		Save(context.Background())
	p3, _ := db.Project.Create().SetName("xxx").
		SetNamespaceID(save.ID).
		SetCreator("").
		Save(context.Background())
	p4 := db.Project.Create().SetName("yyy").
		SetManifest(encodeToYaml(&ing4)).
		SetNamespaceID(save.ID).
		SetCreator("").
		SaveX(context.Background())
	data := data2.NewMockData(m)
	repo := &projectRepo{
		logger: mlog.NewLogger(nil),
		data:   data,
	}
	fk := fake.NewSimpleClientset(
		&networkingv1.IngressList{
			Items: []networkingv1.Ingress{
				ing1,
				ing2,
				ing3,
				ing4,
			},
		},
	)
	var (
		_ = p1
		_ = p2
		_ = p3
		_ = p4
	)
	data.EXPECT().K8sClient().Return(&data2.K8sClient{
		IngressLister: NewIngressLister(&ing1, &ing2, &ing3, &ing4),
		Client:        fk,
	})

	only, _ := db.Namespace.Query().Where(namespace.ID(save.ID)).WithProjects().Only(context.TODO())
	mapping := repo.GetIngressMappingByProjects(context.TODO(), save.Name, serialize.Serialize(only.Edges.Projects, ToProject)...)

	assert.Len(t, mapping["app1"], 3)
	assert.Len(t, mapping["app2"], 2)
	assert.Equal(t, "https://app2.org", mapping["app2"][0].Url)
	assert.Equal(t, "http://http.com", mapping["app2"][1].Url)
	assert.Len(t, mapping["xxx"], 0)
	assert.Len(t, mapping["yyy"], 2)
	for _, endpoint := range mapping["yyy"] {
		assert.True(t, strings.HasPrefix(endpoint.Url, "http://"))
	}
}

func Test_projectRepo_GetNodePortMappingByProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc1 := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "duc",
			Name:      "svc1",
		},
		Spec: corev1.ServiceSpec{
			Type: "NodePort",
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30000,
				},
				{
					Name:     "ui",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30001,
				},
				{
					Name:     "web",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30002,
				},
				{
					Name:     "api",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30003,
				},
				{
					Name:     "grpc",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30004,
				},
				{
					Name:     "xxxx",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30005,
				},
			},
		},
	}
	lister := NewServiceLister(&svc1)
	data := data2.NewMockData(m)
	repo := &projectRepo{
		logger:     mlog.NewLogger(nil),
		data:       data,
		externalIp: "127.0.0.1",
	}
	data.EXPECT().K8sClient().Return(&data2.K8sClient{
		ServiceLister: lister,
	})
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	ns, _ := db.Namespace.Create().SetName("duc").Save(context.Background())
	p1 := db.Project.Create().SetName("svc1").
		SetManifest(encodeToYaml(&svc1)).
		SetNamespaceID(ns.ID).
		SetCreator("").
		SaveX(context.Background())
	_ = p1
	only, _ := db.Namespace.Query().Where(namespace.ID(ns.ID)).WithProjects().Only(context.Background())
	mapping := repo.GetNodePortMappingByProjects(context.TODO(), ns.Name, serialize.Serialize(only.Edges.Projects, ToProject)...)
	httpCount := 0
	total := 0
	for _, endpoints := range mapping {
		for _, endpoint := range endpoints {
			total++
			if strings.HasPrefix(endpoint.Url, "http") {
				httpCount++
			}
		}
	}
	assert.Equal(t, 4, httpCount)
	assert.Equal(t, 6, total)
}

func Test_projectRepo_GetGatewayHTTPRouteMappingByProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	route1 := gatewayv1.HTTPRoute{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HTTPRoute",
			APIVersion: "gateway.networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "duc",
			Name:      "r1",
		},
		Spec: gatewayv1.HTTPRouteSpec{
			Hostnames: []gatewayv1.Hostname{"app1.com", "app1.io"},
		},
	}
	lister := NewHTTPRouteLister(&route1)
	data := data2.NewMockData(m)
	repo := &projectRepo{
		logger: mlog.NewLogger(nil),
		data:   data,
	}
	data.EXPECT().K8sClient().Return(&data2.K8sClient{
		HTTPRouteLister:     lister,
		GatewayApiInstalled: true,
	})
	db, _ := data2.NewSqliteDB()
	defer db.Close()
	ns, _ := db.Namespace.Create().SetName("duc").Save(context.Background())
	p1 := db.Project.Create().SetName("r1").
		SetManifest(encodeToYaml(&route1)).
		SetNamespaceID(ns.ID).
		SetCreator("").
		SaveX(context.Background())
	_ = p1
	only, _ := db.Namespace.Query().Where(namespace.ID(ns.ID)).WithProjects().Only(context.Background())
	mapping := repo.GetGatewayHTTPRouteMappingByProjects(context.TODO(), ns.Name, serialize.Serialize(only.Edges.Projects, ToProject)...)
	assert.Len(t, mapping["r1"], 2)
}

func NewServiceLister(svcs ...*corev1.Service) corev1lister.ServiceLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range svcs {
		idxer.Add(po)
	}
	return corev1lister.NewServiceLister(idxer)
}

func NewHTTPRouteLister(svcs ...*gatewayv1.HTTPRoute) httproutev1.HTTPRouteLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range svcs {
		idxer.Add(po)
	}
	return httproutev1.NewHTTPRouteLister(idxer)
}

func NewIngressLister(svcs ...*networkingv1.Ingress) networkingv1lister.IngressLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range svcs {
		idxer.Add(po)
	}
	return networkingv1lister.NewIngressLister(idxer)
}

func NewPodLister(pods ...*corev1.Pod) corev1lister.PodLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range pods {
		idxer.Add(po)
	}
	return corev1lister.NewPodLister(idxer)
}

func NewRsLister(rs ...*appsv1.ReplicaSet) appsv1lister.ReplicaSetLister {
	idxer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, po := range rs {
		idxer.Add(po)
	}
	return appsv1lister.NewReplicaSetLister(idxer)
}
