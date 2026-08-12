package data

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/namespace"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	appsv1lister "k8s.io/client-go/listers/apps/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	networkingv1lister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubectl/pkg/util/deployment"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	httproutev1 "sigs.k8s.io/gateway-api/pkg/client/listers/apis/v1"
)

func createRepo(db *ent.Client) *ent.Repo {
	return db.Repo.Create().SetName(rand.String(10)).SaveX(context.TODO())
}
func TestProjectRepoCreate(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)
	repo := createRepo(db)
	ns := createNamespace(db)
	input := &biz.CreateProjectInput{
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
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	p := createProject(db, createNamespace(db).ID)

	input := &biz.UpdateProjectInput{
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
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	err := r.Delete(ctx, 1)
	assert.Error(t, err)

	project := createProject(db, createNamespace(db).ID)
	err = r.Delete(ctx, project.ID)
	assert.Nil(t, err)
}

func TestProjectRepoFindByName(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	p := createProject(db, createNamespace(db).ID)

	project, err := r.FindByName(ctx, p.Name, p.NamespaceID)
	assert.NoError(t, err)
	assert.Equal(t, p.Name, project.Name)
	assert.Equal(t, 1, project.NamespaceID)
}

func TestProjectRepoUpdateDeployStatus(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	p := createProject(db, createNamespace(db).ID)

	project, err := r.UpdateDeployStatus(ctx, p.ID, types.Deploy_StatusDeploying)
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusDeploying, project.DeployStatus)
}

func TestProjectRepoUpdateVersion(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	p := createProject(db, createNamespace(db).ID)

	project, err := r.UpdateVersion(ctx, p.ID, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, project.Version)
}

func TestProjectRepoFindProjectsByIDs(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	// 空 ids 直接返回 nil，不触达 DB。
	empty, err := r.FindProjectsByIDs(ctx)
	assert.Nil(t, err)
	assert.Nil(t, empty)

	p1 := createProject(db, createNamespace(db).ID)
	p2 := createProject(db, createNamespace(db).ID)
	projs, err := r.FindProjectsByIDs(ctx, p1.ID, p2.ID, 99999)
	assert.NoError(t, err)
	assert.Len(t, projs, 2)
	assert.Equal(t, p1.Name, projs[0].Name)
	assert.Equal(t, p2.Name, projs[1].Name)
}

func TestProjectRepoList(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	// Create some projects
	for i := 0; i < 5; i++ {
		createProject(db, createNamespace(db).ID)
	}

	// Test list with pagination
	input := &biz.ListProjectInput{
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

	// Test list without pagination (admin 跳过访问过滤分支)
	input = &biz.ListProjectInput{
		Page:     1,
		PageSize: 10,
		IsAdmin:  true,
	}
	projects, pagination, err = r.List(ctx, input)
	assert.NoError(t, err)
	assert.Len(t, projects, 5)
	assert.Equal(t, int32(1), pagination.Page)
	assert.Equal(t, int32(10), pagination.PageSize)
	assert.Equal(t, int32(5), pagination.Count)
}

// 回归防护：project.List 非 admin 必须按命名空间访问谓词过滤——
// 私有命名空间（非创建者/非成员）下的项目不得出现在列表中。
// 去掉 data 层 List 里的访问过滤，本测试必须失败。
func TestProjectRepoList_AccessFilter(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	// 公开命名空间 + 私有命名空间（创建者 owner@x.com）+ 私有命名空间带成员 member@x.com
	pub := db.Namespace.Create().SetCreatorEmail("pub-owner@x.com").SetName("pub").SaveX(ctx)
	pri := db.Namespace.Create().SetCreatorEmail("owner@x.com").SetName("pri").SetPrivate(true).SaveX(ctx)
	priMember := db.Namespace.Create().SetCreatorEmail("owner2@x.com").SetName("priMember").SetPrivate(true).SaveX(ctx)
	db.Member.Create().SetEmail("member@x.com").SetNamespaceID(priMember.ID).SaveX(ctx)

	for _, ns := range []int{pub.ID, pri.ID, priMember.ID} {
		db.Project.Create().SetName("p-" + strconv.Itoa(ns)).SetNamespaceID(ns).SetGitProjectID(1).SetCreator("").SaveX(ctx)
	}

	// 陌生人：只能看到公开命名空间的项目（pub 那 1 个）
	projects, pagination, err := r.List(ctx, &biz.ListProjectInput{Page: 1, PageSize: 10, Email: "stranger@x.com"})
	assert.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, int32(1), pagination.Count)
	assert.Equal(t, "p-"+strconv.Itoa(pub.ID), projects[0].Name)

	// 创建者：能看到自己的私有命名空间（pri）内的项目
	projects, _, err = r.List(ctx, &biz.ListProjectInput{Page: 1, PageSize: 10, Email: "owner@x.com"})
	assert.NoError(t, err)
	assert.Len(t, projects, 2) // pub + pri

	// 成员：能看到作为成员的私有命名空间（priMember）内的项目
	projects, _, err = r.List(ctx, &biz.ListProjectInput{Page: 1, PageSize: 10, Email: "member@x.com"})
	assert.NoError(t, err)
	assert.Len(t, projects, 2) // pub + priMember

	// admin：看到全部
	projects, _, err = r.List(ctx, &biz.ListProjectInput{Page: 1, PageSize: 10, IsAdmin: true})
	assert.NoError(t, err)
	assert.Len(t, projects, 3)
}

func TestProjectRepoList_Empty(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	// Test list when no projects exist
	input := &biz.ListProjectInput{
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
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	repo := createRepo(db)
	project := createProject(db, createNamespace(db).ID)
	project.Update().SetRepo(repo).SaveX(context.TODO())

	show, err := r.Show(context.TODO(), project.ID)
	assert.Nil(t, err)
	assert.NotNil(t, show)
	assert.NotNil(t, show.Namespace)
	assert.NotNil(t, show.Repo)
}

func Test_projectRepo_FindByVersion(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
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

	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
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
				biz.IgnoreContainerNames: "x",
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

	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}, K8sClient: &K8sClient{
		Client:           fk,
		PodLister:        NewPodLister(pod1, pod2, pod3, pod4, pod5, podWithErrorRsName),
		ReplicaSetLister: NewRsLister(rs, rs2, rs3, rs4),
	}})
	r := NewProjectRepo(logger, data)
	// 容器拓扑推导已迁入 biz.ProjectBiz，data 侧用真实 k8sRepo 薄包装跑全链路。
	b := biz.NewProjectBiz(mlog.NewForConfig(nil), r, &k8sRepo{data: data, logger: logger})
	project.Update().SetPodSelectors(nil).Save(context.TODO())
	_, err := b.GetAllActiveContainers(context.TODO(), project.ID)

	assert.Nil(t, err)

	project.Update().SetPodSelectors([]string{"a=a", "b=b"}).Save(context.TODO())
	pods, _ := b.GetAllActiveContainers(context.TODO(), project.ID)
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
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	project := createProject(db, createNamespace(db).ID)
	project.Update().SetDeployStatus(types.Deploy_StatusDeployed).Save(context.TODO())

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

	data := NewMockDataStore(m)
	k8s := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   data,
	}
	data.EXPECT().K8s().Return(&K8sClient{
		ServiceLister: lister,
	})
	db, err := NewSqliteDB()
	assert.Nil(t, err)
	defer db.Close()
	namespace := createNamespace(db)
	namespace.Update().SetName("duc").Save(context.TODO())
	project := createProject(db, namespace.ID)
	project = project.Update().SetName("svc1").
		SetManifest(encodeToYaml(&svc1)).
		SaveX(context.TODO())
	mapping, err := biz.BuildLoadBalancerMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k8s, "duc", toProject(project))
	assert.Nil(t, err)
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
	db, _ := NewSqliteDB()
	defer db.Close()
	save, _ := db.Namespace.Create().SetCreatorEmail("a").SetName("duc").Save(context.TODO())
	p1, _ := db.Project.Create().SetName("app1").
		SetManifest(encodeToYaml(&ing1)).
		SetNamespaceID(save.ID).
		SetCreator("").
		Save(context.TODO())
	p2, _ := db.Project.Create().SetName("app2").
		SetManifest(encodeToYaml(&ing2)).
		SetNamespaceID(save.ID).
		SetCreator("").
		Save(context.TODO())
	p3, _ := db.Project.Create().SetName("xxx").
		SetNamespaceID(save.ID).
		SetCreator("").
		Save(context.TODO())
	p4 := db.Project.Create().SetName("yyy").
		SetManifest(encodeToYaml(&ing4)).
		SetNamespaceID(save.ID).
		SetCreator("").
		SaveX(context.TODO())
	data := NewMockDataStore(m)
	k8s := &k8sRepo{
		logger: mlog.NewForConfig(nil),
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
	data.EXPECT().K8s().Return(&K8sClient{
		IngressLister: NewIngressLister(&ing1, &ing2, &ing3, &ing4),
		Client:        fk,
	})

	only, _ := db.Namespace.Query().Where(namespace.ID(save.ID)).WithProjects().Only(context.TODO())
	mapping, err := biz.BuildIngressMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k8s, save.Name, slice.Map(only.Edges.Projects, toProject)...)
	assert.Nil(t, err)

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
	data := NewMockDataStore(m)
	k8s := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   data,
	}
	data.EXPECT().K8s().Return(&K8sClient{
		ServiceLister: lister,
	})
	// NodePort 端点 URL 的节点 IP 现在经 k8sRepo.ExternalIp() 读取，收敛到配置。
	data.EXPECT().Config().Return(&config.Config{ExternalIp: "127.0.0.1"})
	db, _ := NewSqliteDB()
	defer db.Close()
	ns, _ := db.Namespace.Create().SetCreatorEmail("a").SetName("duc").Save(context.TODO())
	p1 := db.Project.Create().SetName("svc1").
		SetManifest(encodeToYaml(&svc1)).
		SetNamespaceID(ns.ID).
		SetCreator("").
		SaveX(context.TODO())
	_ = p1
	only, _ := db.Namespace.Query().Where(namespace.ID(ns.ID)).WithProjects().Only(context.TODO())
	mapping, err := biz.BuildNodePortMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k8s, ns.Name, slice.Map(only.Edges.Projects, toProject)...)
	assert.Nil(t, err)
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
	data := NewMockDataStore(m)
	k8s := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   data,
	}
	data.EXPECT().K8s().Return(&K8sClient{
		HTTPRouteLister:     lister,
		GatewayApiInstalled: true,
	}).AnyTimes()
	db, _ := NewSqliteDB()
	defer db.Close()
	ns, _ := db.Namespace.Create().SetCreatorEmail("a").SetName("duc").Save(context.TODO())
	p1 := db.Project.Create().SetName("r1").
		SetManifest(encodeToYaml(&route1)).
		SetNamespaceID(ns.ID).
		SetCreator("").
		SaveX(context.TODO())
	_ = p1
	only, _ := db.Namespace.Query().Where(namespace.ID(ns.ID)).WithProjects().Only(context.TODO())
	mapping, err := biz.BuildGatewayHTTPRouteMappingByProjects(context.TODO(), mlog.NewForConfig(nil), k8s, ns.Name, slice.Map(only.Edges.Projects, toProject)...)
	assert.Nil(t, err)
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

func Test_projectRepo_GetProjectEndpointsInNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := NewSqliteDB()
	defer db.Close()
	mockData := NewMockDataStore(m)
	projRepo := &projectRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	k8s := &k8sRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	mockData.EXPECT().K8s().Return(&K8sClient{}).AnyTimes()
	mockData.EXPECT().DB().Return(db)
	b := biz.NewProjectBiz(mlog.NewForConfig(nil), projRepo, k8s)
	_, err := b.GetProjectEndpointsInNamespace(context.TODO(), "duc", 1)
	assert.Nil(t, err)
}

func Test_projectRepo_Version(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := NewSqliteDB()
	defer db.Close()
	mockData := NewMockDataStore(m)
	repo := &projectRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	mockData.EXPECT().DB().Return(db).AnyTimes()
	_, err := repo.Version(context.TODO(), 1)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, s.Code())

	e := createNamespace(db)
	project := createProject(db, e.ID)
	project.Update().SetVersion(100).SaveX(context.TODO())
	res, _ := repo.Version(context.TODO(), project.ID)
	assert.Equal(t, 100, res)
}

// TestProjectRepo_ListByDeployStatus 覆盖按部署状态过滤并携带 namespace 的端口，
// cron FixDeployStatus 依赖该端口修复失败/未知项目。
func TestProjectRepo_ListByDeployStatus(t *testing.T) {
	ctx := context.TODO()
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(mlog.NewForConfig(nil), data)
	repo := createRepo(db)
	ns := createNamespace(db)

	_, err := r.Create(ctx, &biz.CreateProjectInput{
		Name:         "app",
		GitProjectID: 1,
		GitBranch:    "main",
		GitCommit:    "abc",
		Config:       "c",
		ConfigType:   "yaml",
		NamespaceID:  ns.ID,
		DeployStatus: types.Deploy_StatusFailed,
		RepoID:       repo.ID,
		Creator:      "u",
	})
	assert.NoError(t, err)

	list, err := r.ListByDeployStatus(ctx, types.Deploy_StatusFailed)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "app", list[0].Name)
	assert.Equal(t, ns.Name, list[0].Namespace.Name)

	// 不在集合内的状态不命中。
	empty, err := r.ListByDeployStatus(ctx, types.Deploy_StatusUnknown)
	assert.NoError(t, err)
	assert.Len(t, empty, 0)
}

// TestProjectRepo_ErrorBranches 用 closed DB 触发 projectRepo 各查询错误分支。
func TestProjectRepo_ErrorBranches(t *testing.T) {
	closed := NewDataImpl(&NewDataParams{DB: mustClosedDB(t), Cfg: &config.Config{}})
	repo := NewProjectRepo(mlog.NewForConfig(nil), closed).(*projectRepo)
	ctx := context.TODO()

	t.Run("UpdateProject query error", func(t *testing.T) {
		_, err := repo.UpdateProject(ctx, &biz.UpdateProjectInput{ID: 1})
		assert.Error(t, err)
	})

	t.Run("Show query error", func(t *testing.T) {
		_, err := repo.Show(ctx, 1)
		assert.Error(t, err)
	})

	t.Run("ListByDeployStatus query error", func(t *testing.T) {
		_, err := repo.ListByDeployStatus(ctx, types.Deploy_StatusDeployed)
		assert.Error(t, err)
	})

	t.Run("FindProjectsByIDs query error", func(t *testing.T) {
		_, err := repo.FindProjectsByIDs(ctx, 1)
		assert.Error(t, err)
	})
}
