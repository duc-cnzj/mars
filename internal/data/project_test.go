package data

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	entgo "entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/namespace"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/project"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		UpdatedBy:    "testUpdatedBy",
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
	assert.Equal(t, input.UpdatedBy, project.UpdatedBy)
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
		UpdatedBy:    "updatedByUser",
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
	assert.Equal(t, input.UpdatedBy, project.UpdatedBy)
}

// TestProjectRepoUpdateProject_LongGitCommitTitle 回归：git commit title 超过 255 字节
// （含多字节 UTF-8）仍可完整落库。历史上 git_commit_title MaxLen(255) 按字节校验，
// 超长 commit 标题会炸整个部署；放宽为 longtext 后应原样保存、不截断。
func TestProjectRepoUpdateProject_LongGitCommitTitle(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	p := createProject(db, createNamespace(db).ID)

	// 100 个中文字符 = 300 字节 > 255，但只有 100 个 rune：精准命中"按字节校验"的老坑。
	longTitle := strings.Repeat("长", 100)
	require.Greater(t, len(longTitle), 255)

	project, err := r.UpdateProject(ctx, &biz.UpdateProjectInput{
		ID:             p.ID,
		GitCommitTitle: longTitle,
	})
	require.NoError(t, err)
	assert.Equal(t, longTitle, project.GitCommitTitle)
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

// 回归防护：成员被移出私有命名空间后，其项目不得再出现在该成员的列表中。
// 移出走 Member.Delete() → SoftDeleteMixin 钩子转软删，而 HasMembersWith 的裸
// sql.Selector 子查询不被 ent Interceptor 覆盖；谓词里漏写 member.DeletedAtIsNil()
// 时，被移出者仍能看到私有空间内容（越权，曾实测复现）。
func TestProjectRepoList_AccessFilter_RemovedMemberSoftDeleted(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	data := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	r := NewProjectRepo(logger, data)

	pub := db.Namespace.Create().SetCreatorEmail("pub-owner@x.com").SetName("pub").SaveX(ctx)
	pri := db.Namespace.Create().SetCreatorEmail("owner@x.com").SetName("pri").SetPrivate(true).SaveX(ctx)
	mem := db.Member.Create().SetEmail("member@x.com").SetNamespaceID(pri.ID).SaveX(ctx)
	db.Project.Create().SetName("p-pub").SetNamespaceID(pub.ID).SetGitProjectID(1).SetCreator("").SaveX(ctx)
	db.Project.Create().SetName("p-pri").SetNamespaceID(pri.ID).SetGitProjectID(1).SetCreator("").SaveX(ctx)

	// 在册：成员可见公开 + 其私有空间，共 2 个。
	projects, _, err := r.List(ctx, &biz.ListProjectInput{Page: 1, PageSize: 10, Email: "member@x.com"})
	assert.NoError(t, err)
	assert.Len(t, projects, 2, "在册成员应能看到所属私有空间的项目")

	// 移出（生产路径：Member.Delete() → 钩子转软删）。
	require.NoError(t, db.Member.DeleteOneID(mem.ID).Exec(ctx))
	_, gerr := db.Member.Get(ctx, mem.ID)
	require.Error(t, gerr, "钩子须把成员删除转成软删，否则本测试不判别")

	// 移出后：只剩公开空间的项目。
	projects, pagination, err := r.List(ctx, &biz.ListProjectInput{Page: 1, PageSize: 10, Email: "member@x.com"})
	assert.NoError(t, err)
	require.Len(t, projects, 1, "被移出成员不得再看到私有空间的项目")
	assert.Equal(t, "p-pub", projects[0].Name)
	assert.Equal(t, int32(1), pagination.Count, "count 也须排除私有空间，否则分页数泄漏")
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
	b := biz.NewProjectBiz(mlog.NewForConfig(nil), r, &k8sRepo{data: data, logger: logger}, nil)
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
	b := biz.NewProjectBiz(mlog.NewForConfig(nil), projRepo, k8s, nil)
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

	t.Run("ListLivenessPage query error", func(t *testing.T) {
		_, err := repo.ListLivenessPage(ctx, &biz.LivenessPageQuery{Page: 1, PageSize: 10})
		assert.Error(t, err)
	})

	t.Run("ListAllProjectBriefs query error", func(t *testing.T) {
		_, err := repo.ListAllProjectBriefs(ctx)
		assert.Error(t, err)
	})
}

// TestProjectRepo_ListLivenessPage 真 SQL 分页：全量 + 搜索 + 边加载 + stats 全量口径。
// 覆盖：无搜索返回全量且边（命名空间/仓库）已加载、按项目名/命名空间名模糊搜索（不分大小写）、
// 无命中返回空。分页/分类/排序由 kind 专项测试覆盖。
func TestProjectRepo_ListLivenessPage(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	r := NewProjectRepo(logger, NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))

	ns := db.Namespace.Create().SetName("team-apps").SetCreatorEmail("a@q.c").SaveX(ctx)
	legacyNS := db.Namespace.Create().SetName("legacy").SetCreatorEmail("b@q.c").SaveX(ctx)
	repo1 := createRepo(db)
	repo2 := createRepo(db)
	// 项目名命中 "web"；另一项目名不命中但命名空间名命中 "legacy"。
	db.Project.Create().SetName("web-api").SetGitBranch("main").SetGitCommit("c1").SetCreator("u").SetGitProjectID(1).SetNamespaceID(ns.ID).SetRepoID(repo1.ID).SaveX(ctx)
	db.Project.Create().SetName("old-svc").SetGitBranch("dev").SetGitCommit("c2").SetCreator("u").SetGitProjectID(2).SetNamespaceID(legacyNS.ID).SetRepoID(repo2.ID).SaveX(ctx)

	t.Run("无搜索返回全量且边已加载", func(t *testing.T) {
		page, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Page: 1, PageSize: 10, Now: time.Now()})
		assert.NoError(t, err)
		assert.Equal(t, 2, page.Count)
		// 刚创建的项目皆活跃；stats 基于搜索命中全量。
		assert.Equal(t, biz.LivenessStats{Total: 2, Active: 2}, page.Stats)
		if assert.Len(t, page.Projects, 2) {
			for _, p := range page.Projects {
				assert.NotNil(t, p.Namespace, "命名空间边应加载")
				assert.NotNil(t, p.Repo, "仓库边应加载")
			}
		}
	})

	t.Run("按项目名模糊搜索", func(t *testing.T) {
		page, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Search: "web-api", Page: 1, PageSize: 10, Now: time.Now()})
		assert.NoError(t, err)
		assert.Equal(t, 1, page.Count)
		assert.Equal(t, biz.LivenessStats{Total: 1, Active: 1}, page.Stats)
		if assert.Len(t, page.Projects, 1) {
			assert.Equal(t, "web-api", page.Projects[0].Name)
		}
	})

	t.Run("按命名空间名搜索不分大小写", func(t *testing.T) {
		page, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Search: "LEGACY", Page: 1, PageSize: 10, Now: time.Now()})
		assert.NoError(t, err)
		assert.Equal(t, 1, page.Count)
		if assert.Len(t, page.Projects, 1) {
			assert.Equal(t, "old-svc", page.Projects[0].Name)
			assert.Equal(t, "legacy", page.Projects[0].Namespace.Name)
		}
	})

	t.Run("无命中返回空", func(t *testing.T) {
		page, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Search: "zzz", Page: 1, PageSize: 10, Now: time.Now()})
		assert.NoError(t, err)
		assert.Equal(t, 0, page.Count)
		assert.Empty(t, page.Projects)
		assert.Equal(t, biz.LivenessStats{}, page.Stats)
	})

	t.Run("越界页返回空但计数保留", func(t *testing.T) {
		page, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Page: 99, PageSize: 10, Now: time.Now()})
		assert.NoError(t, err)
		assert.Equal(t, 2, page.Count)
		assert.Empty(t, page.Projects)
	})
}

// TestProjectRepo_ListLivenessPage_KindFilter 分类过滤 + 排序决胜键：kind 谓词命中正确行，
// 排序为 updated_at {desc|asc} + id 决胜键（同秒不漂移），stats 不随分类过滤裁剪。
func TestProjectRepo_ListLivenessPage_KindFilter(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	r := NewProjectRepo(logger, NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	ns := db.Namespace.Create().SetName("kind").SetCreatorEmail("a@q.c").SaveX(ctx)
	repo := createRepo(db)
	now := time.Now()
	// 三个分类各一条：活跃/低活跃/僵尸。
	db.Project.Create().SetName("p-active").SetUpdatedAt(now.Add(-10 * 24 * time.Hour)).SetGitBranch("main").SetGitCommit("c1").SetCreator("u").SetGitProjectID(1).SetNamespaceID(ns.ID).SetRepoID(repo.ID).SaveX(ctx)
	db.Project.Create().SetName("p-dormant").SetUpdatedAt(now.Add(-60 * 24 * time.Hour)).SetGitBranch("main").SetGitCommit("c2").SetCreator("u").SetGitProjectID(2).SetNamespaceID(ns.ID).SetRepoID(repo.ID).SaveX(ctx)
	db.Project.Create().SetName("p-zombie").SetUpdatedAt(now.Add(-120 * 24 * time.Hour)).SetGitBranch("main").SetGitCommit("c3").SetCreator("u").SetGitProjectID(3).SetNamespaceID(ns.ID).SetRepoID(repo.ID).SaveX(ctx)

	t.Run("活跃分类命中一行且 stats 为全量", func(t *testing.T) {
		page, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Liveness: "active", Page: 1, PageSize: 10, Now: now})
		assert.NoError(t, err)
		assert.Equal(t, 1, page.Count)
		assert.Equal(t, biz.LivenessStats{Total: 3, Active: 1, Dormant: 1, Zombie: 1}, page.Stats)
		if assert.Len(t, page.Projects, 1) {
			assert.Equal(t, "p-active", page.Projects[0].Name)
		}
	})

	t.Run("低活跃分类命中一行", func(t *testing.T) {
		page, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Liveness: "dormant", Page: 1, PageSize: 10, Now: now})
		assert.NoError(t, err)
		assert.Equal(t, 1, page.Count)
		if assert.Len(t, page.Projects, 1) {
			assert.Equal(t, "p-dormant", page.Projects[0].Name)
		}
	})

	t.Run("僵尸分类命中一行", func(t *testing.T) {
		page, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Liveness: "zombie", Page: 1, PageSize: 10, Now: now})
		assert.NoError(t, err)
		assert.Equal(t, 1, page.Count)
		if assert.Len(t, page.Projects, 1) {
			assert.Equal(t, "p-zombie", page.Projects[0].Name)
		}
	})

	t.Run("非法分类恒空", func(t *testing.T) {
		page, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Liveness: "bogus", Page: 1, PageSize: 10, Now: now})
		assert.NoError(t, err)
		assert.Equal(t, 0, page.Count)
		assert.Empty(t, page.Projects)
	})

	t.Run("排序 updated_at + id 决胜键", func(t *testing.T) {
		// 同 updated_at 的多条：id 决胜键保证 desc/asc 翻页确定性。
		base := now.Add(-5 * 24 * time.Hour)
		for i := 1; i <= 3; i++ {
			db.Project.Create().SetName("p-tie-" + strconv.Itoa(i)).SetUpdatedAt(base).SetGitBranch("main").SetGitCommit("c").SetCreator("u").SetGitProjectID(10 + i).SetNamespaceID(ns.ID).SetRepoID(repo.ID).SaveX(ctx)
		}
		desc, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Sort: "desc", Page: 1, PageSize: 10, Now: now})
		assert.NoError(t, err)
		asc, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Sort: "asc", Page: 1, PageSize: 10, Now: now})
		assert.NoError(t, err)
		// desc: updated_at 大的在前，同秒按 id 大者在前（p-tie-3/2/1 在最前，其次 10d/60d/120d）。
		if assert.Len(t, desc.Projects, 6) {
			assert.Equal(t, "p-tie-3", desc.Projects[0].Name)
			assert.Equal(t, "p-tie-2", desc.Projects[1].Name)
			assert.Equal(t, "p-tie-1", desc.Projects[2].Name)
			assert.Equal(t, "p-active", desc.Projects[3].Name)
			assert.Equal(t, "p-dormant", desc.Projects[4].Name)
			assert.Equal(t, "p-zombie", desc.Projects[5].Name)
		}
		// asc: updated_at 小的在前（最久未更新 p-zombie 打头），同秒 p-tie 按 id 小者在前收尾。
		if assert.Len(t, asc.Projects, 6) {
			assert.Equal(t, "p-zombie", asc.Projects[0].Name)
			assert.Equal(t, "p-dormant", asc.Projects[1].Name)
			assert.Equal(t, "p-active", asc.Projects[2].Name)
			assert.Equal(t, "p-tie-1", asc.Projects[3].Name)
			assert.Equal(t, "p-tie-2", asc.Projects[4].Name)
			assert.Equal(t, "p-tie-3", asc.Projects[5].Name)
		}
	})
}

// TestProjectRepo_ListLivenessPage_BoundaryParity 边界奇偶性守护：SQL kind 计数 == Go
// classifyLiveness 逐行分类。种入跨边界项目（now-31d±1s、now-90d±1s、从未更新），断言
// stats 四分类计数与各 kind 过滤命中行 == Go 逐行分类结果，防边界公式漂移。
func TestProjectRepo_ListLivenessPage_BoundaryParity(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	r := NewProjectRepo(logger, NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	ns := db.Namespace.Create().SetName("parity").SetCreatorEmail("a@q.c").SaveX(ctx)
	repo := createRepo(db)
	now := time.Now()
	updates := []time.Time{
		now.Add(-1 * 24 * time.Hour),            // 活跃
		now.Add(-31*24*time.Hour + time.Second), // 活跃（31d 临界后一秒）
		now.Add(-31 * 24 * time.Hour),           // 低活跃（恰 31d：int 向下取整 31 → 非活跃）
		now.Add(-60 * 24 * time.Hour),           // 低活跃
		now.Add(-90 * 24 * time.Hour),           // 僵尸（恰 90d：int 取整 90 → 僵尸）
		now.Add(-120 * 24 * time.Hour),          // 僵尸
	}
	var wantCounts = map[string]int{"active": 0, "dormant": 0, "zombie": 0}
	wantProjects := map[string]string{} // 名 → kind
	for i, ts := range updates {
		name := "p" + strconv.Itoa(i)
		db.Project.Create().SetName(name).SetUpdatedAt(ts).SetGitBranch("main").SetGitCommit("c").SetCreator("u").SetGitProjectID(i).SetNamespaceID(ns.ID).SetRepoID(repo.ID).SaveX(ctx)
		kind := biz.ClassifyLiveness(ts, now) // Go 逐行分类
		wantCounts[string(kind)]++
		wantProjects[name] = string(kind)
	}

	page, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Page: 1, PageSize: 20, Now: now})
	assert.NoError(t, err)
	assert.Equal(t, biz.LivenessStats{Total: len(updates), Active: wantCounts["active"], Dormant: wantCounts["dormant"], Zombie: wantCounts["zombie"]}, page.Stats)

	// 每个 kind 的过滤命中行 == Go 逐行分类结果。
	for _, kind := range []string{"active", "dormant", "zombie"} {
		filtered, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Liveness: kind, Page: 1, PageSize: 20, Now: now})
		assert.NoError(t, err)
		assert.Len(t, filtered.Projects, wantCounts[kind], "kind=%s SQL 命中行数应等于 Go 分类计数", kind)
		for _, p := range filtered.Projects {
			assert.Equal(t, kind, wantProjects[p.Name], "kind=%s 命中行分类应一致", kind)
		}
	}
}

// TestProjectRepo_ListLivenessPage_QueryErrors 注入第 N 次查询失败，逐一覆盖 ListLivenessPage
// 各 COUNT/list 错误分支（total 已由 TestProjectRepo_ErrorBranches 用 closed DB 覆盖）：
// 无过滤查询序 total→active→dormant→zombie→listAll；有 liveness 过滤时第 5 次为 filtered，
// 第 6 次为 listAll。故 n=2/3/4 覆盖三分类 COUNT 错误，n=5+liveness 覆盖 filtered COUNT 错误，
// n=5 无过滤覆盖 listAll 错误。
func TestProjectRepo_ListLivenessPage_QueryErrors(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	cases := []struct {
		name     string
		failAt   int
		liveness string
	}{
		{name: "active count error", failAt: 2},
		{name: "dormant count error", failAt: 3},
		{name: "zombie count error", failAt: 4},
		{name: "filtered count error", failAt: 5, liveness: "active"},
		{name: "list page error", failAt: 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, _ := NewSqliteDB()
			defer db.Close()
			var n int
			db.Intercept(entgo.InterceptFunc(func(next entgo.Querier) entgo.Querier {
				return entgo.QuerierFunc(func(ctx context.Context, q entgo.Query) (entgo.Value, error) {
					n++
					if n == tc.failAt {
						return nil, errors.New("inject project query error")
					}
					return next.Query(ctx, q)
				})
			}))
			r := NewProjectRepo(logger, NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
			_, err := r.ListLivenessPage(ctx, &biz.LivenessPageQuery{Liveness: tc.liveness, Page: 1, PageSize: 10, Now: time.Now()})
			require.Error(t, err)
			assert.ErrorContains(t, err, "inject project query error")
		})
	}
}

// TestProjectRepo_ListAllProjectBriefs 全量项目精简投影（仅 Name/PodSelectors + 命名空间边）：
// 不分页返回全部并按 id 倒序，供空间资源聚合做 pod→项目归属映射。
func TestProjectRepo_ListAllProjectBriefs(t *testing.T) {
	ctx := context.TODO()
	logger := mlog.NewForConfig(nil)
	db, _ := NewSqliteDB()
	defer db.Close()
	r := NewProjectRepo(logger, NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))

	ns := db.Namespace.Create().SetName("team-apps").SetCreatorEmail("a@q.c").SaveX(ctx)
	repo1 := createRepo(db)
	repo2 := createRepo(db)
	// 落库时写入 config/override_values/manifest 大列真实值，反向断言投影应裁剪它们为零值。
	db.Project.Create().SetName("app-1").SetGitBranch("main").SetGitCommit("c1").SetCreator("u").SetGitProjectID(1).SetNamespaceID(ns.ID).SetRepoID(repo1.ID).SetPodSelectors([]string{"app=1"}).SetConfig("big-config-1").SetOverrideValues("big-override-1").SetManifest([]string{"manifest-1"}).SaveX(ctx)
	db.Project.Create().SetName("app-2").SetGitBranch("main").SetGitCommit("c2").SetCreator("u").SetGitProjectID(2).SetNamespaceID(ns.ID).SetRepoID(repo2.ID).SetPodSelectors([]string{"app=2"}).SetConfig("big-config-2").SetOverrideValues("big-override-2").SetManifest([]string{"manifest-2"}).SaveX(ctx)

	all, err := r.ListAllProjectBriefs(ctx)
	assert.NoError(t, err)
	if assert.Len(t, all, 2) {
		// 按 id 倒序：后创建的 app-2 在前。
		assert.Equal(t, "app-2", all[0].Name)
		assert.Equal(t, "app-1", all[1].Name)
		for _, p := range all {
			assert.Equal(t, "team-apps", p.Namespace.Name, "命名空间边应加载")
			assert.NotEmpty(t, p.PodSelectors, "Pod selectors 应加载")
			// 反向断言：未投影的 longtext/JSON 大列应保持零值——这正是精简投影的卖点，
			// 若未来回归全列拉取（如误加 Select(FieldAll)），此处应失败。
			assert.Empty(t, p.Config, "未投影的 config 大列应保持空")
			assert.Empty(t, p.OverrideValues, "未投影的 override_values 大列应保持空")
			assert.Nil(t, p.Manifest, "未投影的 manifest 大列应保持 nil")
			assert.Nil(t, p.Repo, "未投影的 repo 边应保持 nil")
		}
	}
}

// TestProjectRepo_FindAndRestoreDeleted 项目级恢复：按(空间, 名字)定位软删行、恢复后
// deploy_status 归零为未知、未删除的项目调用恢复显式报错。
func TestProjectRepo_FindAndRestoreDeleted(t *testing.T) {
	ctx := context.TODO()
	db, _ := NewSqliteDB()
	defer db.Close()
	repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	ns := createNamespace(db)
	p := createNamedProject(db, ns.ID, "victim")
	db.Project.UpdateOneID(p.ID).SetDeployStatus(types.Deploy_StatusDeployed).SaveX(ctx)

	// 未软删：软删查询不命中，恢复显式报错（避免把"什么都没做"当成功）。
	_, err := repo.FindDeletedByName(ctx, "victim", ns.ID)
	require.Error(t, err)
	assert.True(t, errs.IsNotFound(err))
	err = repo.RestoreDeleted(ctx, p.ID)
	require.Error(t, err)
	assert.ErrorContains(t, err, "未被删除")

	require.NoError(t, repo.Delete(ctx, p.ID))

	got, err := repo.FindDeletedByName(ctx, "victim", ns.ID)
	require.NoError(t, err)
	assert.Equal(t, p.ID, got.ID)

	require.NoError(t, repo.RestoreDeleted(ctx, p.ID))
	live := db.Project.Query().Where(project.ID(p.ID)).OnlyX(ctx)
	assert.Nil(t, live.DeletedAt)
	assert.Equal(t, types.Deploy_StatusUnknown, live.DeployStatus,
		"恢复后部署状态必须归零为未知：release 已被卸载，残留'已部署'是错误描述")
}

// TestProjectRepo_FindDeletedByName_Latest 同名项目被反复"删除→重建→再删除"时，
// 取最近一次删除的那行（否则会恢复出一份过期配置）。
func TestProjectRepo_FindDeletedByName_Latest(t *testing.T) {
	ctx := context.TODO()
	db, _ := NewSqliteDB()
	defer db.Close()
	repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	ns := createNamespace(db)

	older := createNamedProject(db, ns.ID, "dup")
	require.NoError(t, repo.Delete(ctx, older.ID))
	newer := createNamedProject(db, ns.ID, "dup")
	require.NoError(t, repo.Delete(ctx, newer.ID))

	got, err := repo.FindDeletedByName(ctx, "dup", ns.ID)
	require.NoError(t, err)
	assert.Equal(t, newer.ID, got.ID)

	// 其他空间的同名项目不得被跨空间命中。
	other := createNamespace(db)
	_, err = repo.FindDeletedByName(ctx, "dup", other.ID)
	assert.True(t, errs.IsNotFound(err))
}

// TestProjectRepo_FindDeletedByName_SameSecondTie 同名项目在同一秒内被删两次时，必须取 id
// 更大（= 后删）的那条。
//
// deleted_at 列是 MySQL `datetime`（秒级），"删→同名重建→再删"完全可能落在同一秒，两行时间戳
// 逐字相等。此时仅按 deleted_at 排序取哪行**不确定**：一旦取到旧行，恢复出来的是过期配置，而
// 真正该恢复的新行会被 biz 层"同名在册"检查永久挡死。本用例显式把两行 deleted_at 覆写成同一
// 值来复现该碰撞（不依赖测试库的时间精度）。
func TestProjectRepo_FindDeletedByName_SameSecondTie(t *testing.T) {
	ctx := context.TODO()
	db, _ := NewSqliteDB()
	defer db.Close()
	repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	ns := createNamespace(db)

	older := createNamedProject(db, ns.ID, "dup")
	require.NoError(t, repo.Delete(ctx, older.ID))
	newer := createNamedProject(db, ns.ID, "dup")
	require.NoError(t, repo.Delete(ctx, newer.ID))

	// 复现"同秒"：把后删那行的删除时间强行对齐到与前一行逐字相等。
	skip := mixin.SkipSoftDelete(ctx)
	olderDeletedAt := *db.Project.Query().Where(project.ID(older.ID)).OnlyX(skip).DeletedAt
	db.Project.UpdateOneID(newer.ID).SetDeletedAt(olderDeletedAt).SaveX(skip)
	require.Equal(t, olderDeletedAt,
		*db.Project.Query().Where(project.ID(newer.ID)).OnlyX(skip).DeletedAt,
		"前置：两行必须处于同一 deleted_at（碰撞已构造）")

	got, err := repo.FindDeletedByName(ctx, "dup", ns.ID)
	require.NoError(t, err)
	assert.Equal(t, newer.ID, got.ID, "deleted_at 相同时必须由主键决胜，取 id 更大的那行")
}

// softDeleteNamespace 直接给空间落软删标记（绕过 Delete 的集群侧编排，只造数据状态）。
func softDeleteNamespace(entdb *ent.Client, id int, deletedAt time.Time) {
	entdb.Namespace.UpdateOneID(id).SetDeletedAt(deletedAt).SaveX(mixin.SkipSoftDelete(context.TODO()))
}

// softDeleteProject 直接给项目落软删状态，可精确控制「是否随空间级联删除」的批次标识与删除
// 时间——恢复页的三个排除条件都依赖这两列，必须能独立构造而不是靠 Delete 编排的副作用。
func softDeleteProject(entdb *ent.Client, id int, deletedAt time.Time, withNamespace bool) {
	entdb.Project.UpdateOneID(id).
		SetDeletedAt(deletedAt).
		SetDeletedWithNamespace(withNamespace).
		SaveX(mixin.SkipSoftDelete(context.TODO()))
}

// TestProjectRepo_ListAdminDeletedPage_Filters 列「可恢复项目」的三条核心口径：
//  1. 只软删行可见（在册项目不出现）——拦截器被绕过不等于不过滤。
//  2. 所属空间已软删的行不出现——既覆盖随空间级联删除的那批（标记 true，由恢复空间连带还原），
//     也覆盖「先单独删项目、后删其空间」的行（标记 false，namespaceRepo.Delete 只标记存活项目）。
//     两类都只由「空间存活」这一条谓词排除，故用两个夹具分别锁住，避免将来有人把谓词换成
//     `deleted_with_namespace=false` 后只剩一个夹具仍然通过。
//  3. 行内 Namespace 边真的被加载（恢复请求要按空间名定位）+ DeletedAt 真的被 SELECT 出来
//     （列裁剪漏 FieldDeletedAt 会静默为 nil，前端就显示不出删除时间）。
func TestProjectRepo_ListAdminDeletedPage_Filters(t *testing.T) {
	ctx := context.TODO()
	db, _ := NewSqliteDB()
	defer db.Close()
	repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	base := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)

	liveNS := createNamespace(db)
	deadNS := createNamespace(db)
	softDeleteNamespace(db, deadNS.ID, base)

	// 唯一一条「可恢复」的项目：单独删除，所属空间健在。
	standalone := createNamedProject(db, liveNS.ID, "standalone")
	softDeleteProject(db, standalone.ID, base.Add(time.Minute), false)

	// 随空间级联删除：批次标识为 true，其空间已软删。
	cascade := createNamedProject(db, deadNS.ID, "cascade")
	softDeleteProject(db, cascade.ID, base.Add(2*time.Minute), true)

	// 先单独删项目、后删空间：批次标识仍是列默认的 false，但空间已软删。
	legacy := createNamedProject(db, deadNS.ID, "legacy")
	softDeleteProject(db, legacy.ID, base.Add(3*time.Minute), false)

	// 在册项目：被条件 1 排除。
	live := createNamedProject(db, liveNS.ID, "live")

	page, err := repo.ListAdminDeletedPage(ctx, &biz.ProjectDeletedListPageQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count, "只有单独删除且空间健在的项目可恢复")
	require.Len(t, page.Projects, 1)

	got := page.Projects[0]
	assert.Equal(t, standalone.ID, got.ID)
	assert.NotContains(t, []int{got.ID}, cascade.ID, "随空间级联删除的项目不得出现（须走空间恢复）")
	assert.NotContains(t, []int{got.ID}, legacy.ID, "所属空间已软删的行不得出现")
	assert.NotContains(t, []int{got.ID}, live.ID, "在册项目不得出现")

	require.NotNil(t, got.DeletedAt, "deleted_at 必须被列裁剪 SELECT 出来，否则前端无法展示删除时间")
	assert.WithinDuration(t, base.Add(time.Minute), *got.DeletedAt, time.Second)

	require.NotNil(t, got.Namespace, "必须带出命名空间边：恢复请求按「空间名 + 项目名」定位")
	assert.Equal(t, liveNS.ID, got.Namespace.ID)
}

// TestProjectRepo_ListAdminDeletedPage_OrderAndPagination 排序与分页：按删除时间倒序（「最近
// 删除」是页面语义），同秒时由主键倒序决胜；Count 是搜索命中总数（未分页），翻页不重不漏。
func TestProjectRepo_ListAdminDeletedPage_OrderAndPagination(t *testing.T) {
	ctx := context.TODO()
	db, _ := NewSqliteDB()
	defer db.Close()
	repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	ns := createNamespace(db)
	base := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)

	// 两行同秒（构造碰撞）+ 两行错开：验证双键定序既不因同秒抖动、也不因主键键抢了时间排序。
	sameA := createNamedProject(db, ns.ID, "same-a")
	softDeleteProject(db, sameA.ID, base, false)
	sameB := createNamedProject(db, ns.ID, "same-b")
	softDeleteProject(db, sameB.ID, base, false)
	older := createNamedProject(db, ns.ID, "older")
	softDeleteProject(db, older.ID, base.Add(-time.Hour), false)
	newer := createNamedProject(db, ns.ID, "newer")
	softDeleteProject(db, newer.ID, base.Add(time.Hour), false)

	first, err := repo.ListAdminDeletedPage(ctx, &biz.ProjectDeletedListPageQuery{Page: 1, PageSize: 3})
	require.NoError(t, err)
	assert.Equal(t, 4, first.Count, "Count 是未分页的命中总数，分页器靠它算总页数")
	ids := lo.Map(first.Projects, func(p *biz.Project, _ int) int { return p.ID })
	assert.Equal(t, []int{newer.ID, sameB.ID, sameA.ID}, ids,
		"按删除时间倒序；同秒的两行由主键倒序决胜（id 大者=后创建者在前）")

	second, err := repo.ListAdminDeletedPage(ctx, &biz.ProjectDeletedListPageQuery{Page: 2, PageSize: 3})
	require.NoError(t, err)
	assert.Equal(t, 4, second.Count)
	rest := lo.Map(second.Projects, func(p *biz.Project, _ int) int { return p.ID })
	assert.Equal(t, []int{older.ID}, rest, "第二页只剩最后一行，且不与第一页重复")

	// 越界页返回空行集但 Count 仍为总数（前端据此判断是否还有下一页）。
	third, err := repo.ListAdminDeletedPage(ctx, &biz.ProjectDeletedListPageQuery{Page: 9, PageSize: 3})
	require.NoError(t, err)
	assert.Empty(t, third.Projects)
	assert.Equal(t, 4, third.Count)
}

// TestProjectRepo_ListAdminDeletedPage_Search 关键词同时匹配项目名与所属空间名（与项目治理
// 列表的搜索口径一致），且搜索结果同样只含可恢复行。
//
// 两个关键词刻意选成互不为子串（"alpha" vs "team-one"），否则一条查询会同时命中两条路径，
// 断言就分不清是项目名匹配还是空间名匹配生效了。
func TestProjectRepo_ListAdminDeletedPage_Search(t *testing.T) {
	ctx := context.TODO()
	db, _ := NewSqliteDB()
	defer db.Close()
	repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	base := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)

	nsOne := db.Namespace.Create().SetName("team-one").SetCreatorEmail("one@q.c").SaveX(ctx)
	nsTwo := db.Namespace.Create().SetName("team-two").SetCreatorEmail("two@q.c").SaveX(ctx)

	byName := createNamedProject(db, nsTwo.ID, "alpha-worker")
	softDeleteProject(db, byName.ID, base, false)
	byNS := createNamedProject(db, nsOne.ID, "worker")
	softDeleteProject(db, byNS.ID, base.Add(-time.Minute), false)
	other := createNamedProject(db, nsTwo.ID, "unrelated")
	softDeleteProject(db, other.ID, base.Add(-2*time.Minute), false)

	ids := func(page *biz.ProjectDeletedListPageResult) []int {
		return lo.Map(page.Projects, func(p *biz.Project, _ int) int { return p.ID })
	}

	// 命中项目名（"alpha-worker"），空间名 team-two 不含 alpha。
	hit, err := repo.ListAdminDeletedPage(ctx, &biz.ProjectDeletedListPageQuery{Search: "alpha", Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, []int{byName.ID}, ids(hit))

	// 命中空间名（team-one）——项目名 "worker" 本身不含该词，只能靠命名空间边 JOIN 命中。
	bySpace, err := repo.ListAdminDeletedPage(ctx, &biz.ProjectDeletedListPageQuery{Search: "team-one", Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, []int{byNS.ID}, ids(bySpace))

	// 无命中：Count 与行集都为空，而不是把所有行都放回来。
	none, err := repo.ListAdminDeletedPage(ctx, &biz.ProjectDeletedListPageQuery{Search: "no-such-thing", Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Empty(t, none.Projects)
	assert.Equal(t, 0, none.Count)
}

// TestProjectRepo_ListAdminDeletedPage_QueryErrors 覆盖两条真实数据库错误分支：计数失败与取页失败
// 都必须带着上下文上抛，不能返回"空页"——那会让超管把 DB 故障误读成"没有可恢复项目"。
// SQL 序列：Q1 COUNT(*) → Q2 SELECT projects(带空间边)。
//
//	qAfter=0 → 第 1 条查询（COUNT）失败，命中 "count deleted projects"；
//	qAfter=1 → COUNT 成功后第 2 条查询（SELECT projects）失败，命中 "list deleted projects page"。
func TestProjectRepo_ListAdminDeletedPage_QueryErrors(t *testing.T) {
	ctx := context.TODO()
	base := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)

	// setup 建「存活空间 + 其下已单独软删项目」：空间须存活否则被 HasNamespaceWith 过滤，
	// 两条查询就都处理空集，测不出真实错误路径。
	setup := func(t *testing.T, db *ent.Client) {
		t.Helper()
		ns := createNamespace(db)
		p := createNamedProject(db, ns.ID, "ldp-err")
		softDeleteProject(db, p.ID, base, false)
	}

	t.Run("Count query error", func(t *testing.T) {
		db, fd := newFailDB(t, 0, -1)
		repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
		setup(t, db)
		fd.Arm()
		_, err := repo.ListAdminDeletedPage(ctx, &biz.ProjectDeletedListPageQuery{Page: 1, PageSize: 10})
		assert.ErrorContains(t, err, "count deleted projects")
	})

	t.Run("All query error", func(t *testing.T) {
		db, fd := newFailDB(t, 1, -1)
		repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
		setup(t, db)
		fd.Arm()
		_, err := repo.ListAdminDeletedPage(ctx, &biz.ProjectDeletedListPageQuery{Page: 1, PageSize: 10})
		assert.ErrorContains(t, err, "list deleted projects page")
	})
}

// TestProjectRepo_RestoreDeleted_MissingProject 恢复一个不存在的 id：加载软删项目即失败，
// 错误按 NotFound 表达（而非 500），调用方才好回「项目不存在」。
func TestProjectRepo_RestoreDeleted_MissingProject(t *testing.T) {
	db, _ := NewSqliteDB()
	defer db.Close()
	repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))

	err := repo.RestoreDeleted(context.TODO(), 999999)
	require.Error(t, err)
	assert.True(t, errs.IsNotFound(err), "不存在的项目应表达为 NotFound，实际: %v", err)
	assert.ErrorContains(t, err, "restore project")
}

// TestProjectRepo_RestoreDeleted_NamespaceQueryError 加载所属空间失败（真实查询错误）必须上抛，
// 不能让恢复继续——否则会在"空间状态未知"的前提下清掉软删标记。
// SQL 序列：Q1 SELECT projects(Only) → Q2 SELECT namespaces(Only，失败注入点)。
func TestProjectRepo_RestoreDeleted_NamespaceQueryError(t *testing.T) {
	ctx := context.TODO()
	db, fd := newFailDB(t, 1, -1)
	repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	ns := createNamespace(db)
	p := createNamedProject(db, ns.ID, "ns-lookup-fail")
	require.NoError(t, repo.Delete(ctx, p.ID))
	fd.Arm()

	err := repo.RestoreDeleted(ctx, p.ID)
	require.ErrorContains(t, err, "restore project")

	// 空间查询失败时不得清软删标记（解除注入后再回读，否则断言自身会被注入失败）。
	fd.Disarm()
	skip := mixin.SkipSoftDelete(ctx)
	assert.NotNil(t, db.Project.Query().Where(project.ID(p.ID)).OnlyX(skip).DeletedAt)
}

// TestProjectRepo_RestoreDeleted_NamespaceDeleted 所属空间仍处软删态时不得恢复该项目：
// 复活到已删空间下等于制造孤儿数据，应引导调用方先恢复空间（连带恢复整批项目）。
//
// 一并锁死协议码：该错误在事务闭包内构造、又被外层 errs.Wrap（默认 Internal）再包一次，靠
// wrapErr 对"已带确定码的错误原样保留"才维持 400。若哪天这段被判为"500 也无妨"而改错，
// 前端就再也收不到可引导用户的 400，故这里断言码而非只断言文案。
func TestProjectRepo_RestoreDeleted_NamespaceDeleted(t *testing.T) {
	ctx := context.TODO()
	db, _ := NewSqliteDB()
	defer db.Close()
	nsRepo := NewNamespaceRepo(NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	ns := createNamespace(db)
	p := createNamedProject(db, ns.ID, "orphan-candidate")

	require.NoError(t, nsRepo.Delete(ctx, ns.ID))

	err := repo.RestoreDeleted(ctx, p.ID)
	require.Error(t, err)
	assert.ErrorContains(t, err, "请先恢复空间")
	assert.Equal(t, codes.InvalidArgument, status.Code(err),
		"外层 errs.Wrap 默认映射 Internal，只有保留内层已带的 400 才不降级成 500")

	skip := mixin.SkipSoftDelete(ctx)
	assert.NotNil(t, db.Project.Query().Where(project.ID(p.ID)).OnlyX(skip).DeletedAt,
		"被拒的恢复不得留下半个状态")
}

// txTraceDriver 包装 sqlite 驱动，记录每条语句**是否在事务内**执行，供断言
// 「校验与写入同事务」这类原子性契约使用（无法用结果观测的并发窗口，只能验事务边界）。
type txTraceDriver struct {
	dialect.Driver
	inTx bool
	ops  []txTraceOp
}

type txTraceOp struct {
	inTx bool
	kind string // "query" / "exec"
	sql  string
}

func (d *txTraceDriver) record(inTx bool, kind, query string) {
	d.ops = append(d.ops, txTraceOp{inTx: inTx, kind: kind, sql: query})
}

func (d *txTraceDriver) Exec(ctx context.Context, query string, args, v any) error {
	d.record(d.inTx, "exec", query)
	return d.Driver.Exec(ctx, query, args, v)
}

func (d *txTraceDriver) Query(ctx context.Context, query string, args, v any) error {
	d.record(d.inTx, "query", query)
	return d.Driver.Query(ctx, query, args, v)
}

func (d *txTraceDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := d.Driver.Tx(ctx)
	if err != nil {
		return nil, err
	}
	d.inTx = true
	return &txTraceTx{Tx: tx, parent: d}, nil
}

type txTraceTx struct {
	dialect.Tx
	parent *txTraceDriver
}

func (t *txTraceTx) Query(ctx context.Context, query string, args, v any) error {
	t.parent.record(true, "query", query)
	return t.Tx.Query(ctx, query, args, v)
}

func (t *txTraceTx) Exec(ctx context.Context, query string, args, v any) error {
	t.parent.record(true, "exec", query)
	return t.Tx.Exec(ctx, query, args, v)
}

func (t *txTraceTx) Commit() error {
	err := t.Tx.Commit()
	t.parent.inTx = false
	return err
}

func (t *txTraceTx) Rollback() error {
	err := t.Tx.Rollback()
	t.parent.inTx = false
	return err
}

// newTxTraceDB 基于 sqlite + 事务追踪驱动构造 ent client。
func newTxTraceDB(t *testing.T) (*ent.Client, *txTraceDriver) {
	t.Helper()
	sqlDrv, err := sql.Open("sqlite3", "file:txtrace?mode=memory&cache=shared&_fk=1&loc=Local")
	require.NoError(t, err)
	d := &txTraceDriver{Driver: sqlDrv}
	client := ent.NewClient(ent.Driver(d))
	require.NoError(t, client.Schema.Create(context.TODO()))
	t.Cleanup(func() { client.Close() })
	return client, d
}

// TestProjectRepo_RestoreDeleted_NamespaceCheckAndWriteInSameTx 原子性契约：空间存活校验
// （SELECT namespaces）与项目写入（UPDATE projects）必须落在**同一个事务**内——否则校验通过
// 后、写入前空间被并发软删，照样写出一条「项目存活、所属空间软删」的孤儿。
//
// 并发窗口本身无法用结果确定性复现，故用事务追踪驱动直接断言这两条语句都在事务内执行；
// 回退成「三条独立语句」的旧实现会把它们记在事务外，本用例立刻转红。
func TestProjectRepo_RestoreDeleted_NamespaceCheckAndWriteInSameTx(t *testing.T) {
	ctx := context.TODO()
	db, tracer := newTxTraceDB(t)
	repo := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	ns := createNamespace(db)
	p := createNamedProject(db, ns.ID, "atomic")
	require.NoError(t, repo.Delete(ctx, p.ID))

	tracer.ops = nil
	require.NoError(t, repo.RestoreDeleted(ctx, p.ID))

	var nsSelectInTx, projUpdateInTx bool
	for _, op := range tracer.ops {
		if op.inTx && op.kind == "query" && strings.Contains(op.sql, "namespaces") {
			nsSelectInTx = true
		}
		if op.inTx && op.kind == "exec" && strings.Contains(op.sql, "projects") {
			projUpdateInTx = true
		}
	}
	assert.True(t, nsSelectInTx, "空间存活校验必须发生在事务内")
	assert.True(t, projUpdateInTx, "项目写入必须发生在事务内")
}

// TestProjectRepo_UpdateVersion_SoftDeletedRejected 已软删项目不得被 UPDATE 静默改写：
// UPDATE 路径没有自动软删过滤（拦截器只管 SELECT、钩子只挂删除），显式谓词是承重代码。
// 去掉 project.DeletedAtIsNil() 谓词本用例即转红——旧实现会把版本号写进软删行并报成功。
func TestProjectRepo_UpdateVersion_SoftDeletedRejected(t *testing.T) {
	ctx := context.TODO()
	db, _ := NewSqliteDB()
	defer db.Close()
	r := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	ns := createNamespace(db)
	p := createNamedProject(db, ns.ID, "victim")
	require.NoError(t, r.Delete(ctx, p.ID))

	got, err := r.UpdateVersion(ctx, p.ID, 99)
	require.Error(t, err)
	assert.True(t, errs.IsNotFound(err), "写已软删项目应表达为 NotFound，实际: %v", err)
	assert.Nil(t, got)

	skip := mixin.SkipSoftDelete(ctx)
	assert.NotEqual(t, 99, db.Project.Query().Where(project.ID(p.ID)).OnlyX(skip).Version,
		"软删行不得被静默改写")
}

// TestProjectRepo_UpdateDeployStatus_SoftDeletedRejected 同 UpdateVersion：UPDATE 路径
// 缺显式软删谓词会把「更新已软删项目」变成静默成功；去谓词即转红。
func TestProjectRepo_UpdateDeployStatus_SoftDeletedRejected(t *testing.T) {
	ctx := context.TODO()
	db, _ := NewSqliteDB()
	defer db.Close()
	r := NewProjectRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}}))
	ns := createNamespace(db)
	p := createNamedProject(db, ns.ID, "victim")
	require.NoError(t, r.Delete(ctx, p.ID))

	got, err := r.UpdateDeployStatus(ctx, p.ID, types.Deploy_StatusDeployed)
	require.Error(t, err)
	assert.True(t, errs.IsNotFound(err), "写已软删项目应表达为 NotFound，实际: %v", err)
	assert.Nil(t, got)

	skip := mixin.SkipSoftDelete(ctx)
	assert.NotEqual(t, types.Deploy_StatusDeployed,
		db.Project.Query().Where(project.ID(p.ID)).OnlyX(skip).DeployStatus, "软删行不得被静默改写")
}
