package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmetatypes "k8s.io/apimachinery/pkg/types"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// fakeProjectRepoForProjectBiz 记录各写操作是否被调用，输入校验测试中 repo 不被调用（调用即 panic）。
// 透传查询方法返回罐头数据并记录调用，供纯透传测试断言。
type fakeProjectRepoForProjectBiz struct {
	ProjectRepo
	createCalled, deleteCalled, updateProjectCalled  bool
	findByIDsCalled, listCalled, listLivenessCalled  bool
	listAllCalled                                    bool
	showCalled, versionCalled, findByNameCalled      bool
	updateDeployStatusCalled, updateVersionCalled    bool
	findByVersionCalled, updateStatusByVersionCalled bool
	findByIDsErr, showErr, listLivenessErr           error
	listLivenessPageResult                           *LivenessPageResult
	listLivenessPageQuery                            *LivenessPageQuery
}

func (f *fakeProjectRepoForProjectBiz) Create(ctx context.Context, project *CreateProjectInput) (*Project, error) {
	f.createCalled = true
	return &Project{ID: 1, Name: project.Name}, nil
}

func (f *fakeProjectRepoForProjectBiz) Delete(ctx context.Context, id int) error {
	f.deleteCalled = true
	return nil
}

func (f *fakeProjectRepoForProjectBiz) UpdateProject(ctx context.Context, input *UpdateProjectInput) (*Project, error) {
	f.updateProjectCalled = true
	return &Project{ID: input.ID}, nil
}

func (f *fakeProjectRepoForProjectBiz) FindProjectsByIDs(ctx context.Context, ids ...int) ([]*Project, error) {
	f.findByIDsCalled = true
	if f.findByIDsErr != nil {
		return nil, f.findByIDsErr
	}
	return []*Project{{Name: "proj1", Namespace: &Namespace{Name: "ns"}, Manifest: []string{svcManifest, ingressManifest, httpRouteManifest}}}, nil
}

func (f *fakeProjectRepoForProjectBiz) List(ctx context.Context, input *ListProjectInput) ([]*Project, *pagination.Pagination, error) {
	f.listCalled = true
	return []*Project{{ID: 1}}, nil, nil
}

func (f *fakeProjectRepoForProjectBiz) ListAll(ctx context.Context) ([]*Project, error) {
	f.listAllCalled = true
	return []*Project{{Name: "proj1", Namespace: &Namespace{Name: "ns"}}}, nil
}

func (f *fakeProjectRepoForProjectBiz) ListLivenessPage(ctx context.Context, query *LivenessPageQuery) (*LivenessPageResult, error) {
	f.listLivenessCalled = true
	f.listLivenessPageQuery = query
	if f.listLivenessErr != nil {
		return nil, f.listLivenessErr
	}
	return f.listLivenessPageResult, nil
}

func (f *fakeProjectRepoForProjectBiz) Show(ctx context.Context, id int) (*Project, error) {
	f.showCalled = true
	if f.showErr != nil {
		return nil, f.showErr
	}
	return &Project{ID: id, Namespace: &Namespace{Name: "ns"}, PodSelectors: []string{"app=a"}}, nil
}

func (f *fakeProjectRepoForProjectBiz) Version(ctx context.Context, id int) (int, error) {
	f.versionCalled = true
	return 3, nil
}

func (f *fakeProjectRepoForProjectBiz) FindByName(ctx context.Context, name string, nsID int) (*Project, error) {
	f.findByNameCalled = true
	return &Project{ID: 1, Name: name}, nil
}

func (f *fakeProjectRepoForProjectBiz) UpdateDeployStatus(ctx context.Context, id int, status types.Deploy) (*Project, error) {
	f.updateDeployStatusCalled = true
	return &Project{ID: id, DeployStatus: status}, nil
}

func (f *fakeProjectRepoForProjectBiz) UpdateVersion(ctx context.Context, id int, version int) (*Project, error) {
	f.updateVersionCalled = true
	return &Project{ID: id, Version: version}, nil
}

func (f *fakeProjectRepoForProjectBiz) FindByVersion(ctx context.Context, id, version int) (*Project, error) {
	f.findByVersionCalled = true
	return &Project{ID: id, Version: version}, nil
}

func (f *fakeProjectRepoForProjectBiz) UpdateStatusByVersion(ctx context.Context, id int, status types.Deploy, version int) (*Project, error) {
	f.updateStatusByVersionCalled = true
	return &Project{ID: id, DeployStatus: status, Version: version}, nil
}

// fakeClRepoForLiveness 是活跃度测试专用的 changelog 假实现：CountByProjectIDs 返回
// 罐头计数或注入错误，供 projectBiz.Liveness 聚合测试。
type fakeClRepoForLiveness struct {
	ChangelogRepo
	counts map[int]int
	err    error
}

func (f *fakeClRepoForLiveness) CountByProjectIDs(_ context.Context, _ ...int) (map[int]int, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.counts, nil
}

func newProjectBizForTest(repo ProjectRepo) ProjectBiz {
	return NewProjectBiz(mlog.NewForConfig(nil), repo, nil, &fakeClRepoForLiveness{})
}

// newProjectBizWithK8s 组装注入 k8sRepo 的 project biz，供容器/端点派生测试使用。
func newProjectBizWithK8s(repo ProjectRepo, k8sRepo K8sRepo) ProjectBiz {
	return NewProjectBiz(mlog.NewForConfig(nil), repo, k8sRepo, &fakeClRepoForLiveness{})
}

// newProjectBizWithClRepo 组装注入自定义 changelog 假实现的 project biz，供活跃度测试使用。
func newProjectBizWithClRepo(repo ProjectRepo, clRepo ChangelogRepo) ProjectBiz {
	return NewProjectBiz(mlog.NewForConfig(nil), repo, nil, clRepo)
}

func TestProjectBiz_Create_NilInput(t *testing.T) {
	b := newProjectBizForTest(&fakeProjectRepoForProjectBiz{})
	got, err := b.Create(context.TODO(), nil)
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "project 不能为空", status.Convert(err).Message())
}

func TestProjectBiz_Create_EmptyName(t *testing.T) {
	b := newProjectBizForTest(&fakeProjectRepoForProjectBiz{})
	got, err := b.Create(context.TODO(), &CreateProjectInput{Name: "", NamespaceID: 1, RepoID: 1})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "project 名称不能为空", status.Convert(err).Message())
}

func TestProjectBiz_Create_InvalidNamespaceID(t *testing.T) {
	b := newProjectBizForTest(&fakeProjectRepoForProjectBiz{})
	got, err := b.Create(context.TODO(), &CreateProjectInput{Name: "app", NamespaceID: 0, RepoID: 1})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "project 所属 namespace 不能为空", status.Convert(err).Message())
}

func TestProjectBiz_Create_InvalidRepoID(t *testing.T) {
	b := newProjectBizForTest(&fakeProjectRepoForProjectBiz{})
	got, err := b.Create(context.TODO(), &CreateProjectInput{Name: "app", NamespaceID: 1, RepoID: 0})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "project 所属 repo 不能为空", status.Convert(err).Message())
}

func TestProjectBiz_Create_Valid(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	got, err := b.Create(context.TODO(), &CreateProjectInput{Name: "app", NamespaceID: 1, RepoID: 1})
	assert.NoError(t, err)
	assert.True(t, f.createCalled)
	assert.Equal(t, "app", got.Name)
}

func TestProjectBiz_Delete_InvalidID(t *testing.T) {
	b := newProjectBizForTest(&fakeProjectRepoForProjectBiz{})
	err := b.Delete(context.TODO(), 0)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "project id 不能小于等于 0", status.Convert(err).Message())
}

func TestProjectBiz_Delete_Valid(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	assert.NoError(t, b.Delete(context.TODO(), 1))
	assert.True(t, f.deleteCalled)
}

func TestProjectBiz_UpdateProject_InvalidID(t *testing.T) {
	b := newProjectBizForTest(&fakeProjectRepoForProjectBiz{})
	got, err := b.UpdateProject(context.TODO(), &UpdateProjectInput{ID: 0})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "project id 不能小于等于 0", status.Convert(err).Message())
}

func TestProjectBiz_UpdateProject_Valid(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	got, err := b.UpdateProject(context.TODO(), &UpdateProjectInput{ID: 1})
	assert.NoError(t, err)
	assert.True(t, f.updateProjectCalled)
	assert.Equal(t, 1, got.ID)
}

// TestProject_ToEventYaml 直接测模型方法 ToEventYaml：nil receiver 返回 nil，非 nil 时
// 三个值序列（EnvValues/ExtraValues/FinalExtraValues）按 key/path 排序后封进 AnyYamlPrettier。
// 在 biz 包内直测而非跨包，保证计入 biz 自身 coverprofile（跨包测试不贡献）。
func TestProject_ToEventYaml(t *testing.T) {
	tests := []struct {
		name     string
		project  *Project
		expected YamlPrettier
	}{
		{
			name:     "Nil receiver",
			project:  nil,
			expected: nil,
		},
		{
			name: "Non-nil project",
			project: &Project{
				GitCommitTitle:  "Initial commit",
				GitBranch:       "main",
				GitCommit:       "abc123",
				Atomic:          true,
				GitCommitWebURL: "http://example.com",
				Config:          "some config",
				EnvValues: []*types.KeyValue{
					{Key: "key2", Value: "value2"},
					{Key: "key1", Value: "value1"},
				},
				ExtraValues: []*websocket_pb.ExtraValue{
					{Path: "path2", Value: "value2"},
					{Path: "path1", Value: "value1"},
				},
				FinalExtraValues: []*websocket_pb.ExtraValue{
					{Path: "path3", Value: "value3"},
					{Path: "path1", Value: "value1"},
				},
			},
			expected: AnyYamlPrettier{
				"title":   "Initial commit",
				"branch":  "main",
				"commit":  "abc123",
				"atomic":  true,
				"web_url": "http://example.com",
				"config":  "some config",
				"env_values": []*types.KeyValue{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
				"extra_values": []*websocket_pb.ExtraValue{
					{Path: "path1", Value: "value1"},
					{Path: "path2", Value: "value2"},
				},
				"final_extra_values": []*websocket_pb.ExtraValue{
					{Path: "path1", Value: "value1"},
					{Path: "path3", Value: "value3"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.project.ToEventYaml()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ---- 容器/端点派生 ----

func TestProjectBiz_GetAllActiveContainers(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	k := &fakeEndpointK8sRepo{pods: []*corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "ns"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status: corev1.PodStatus{
			Phase:             corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{{Name: "web", Ready: true}},
		},
	}}}
	b := newProjectBizWithK8s(f, k)
	got, err := b.GetAllActiveContainers(context.TODO(), 1)
	assert.NoError(t, err)
	assert.True(t, f.showCalled)
	if assert.Len(t, got, 1) {
		assert.Equal(t, "pod", got[0].Pod)
		assert.Equal(t, "web", got[0].Container)
		assert.True(t, got[0].Ready)
	}
}

func TestProjectBiz_GetProjectEndpointsInNamespace(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	k := &fakeEndpointK8sRepo{
		gatewayInstalled: true,
		externalIP:       "10.0.0.1",
		ingresses: []*networkingv1.Ingress{{
			ObjectMeta: metav1.ObjectMeta{Name: "web-ing", Namespace: "ns"},
			Spec:       networkingv1.IngressSpec{Rules: []networkingv1.IngressRule{{Host: "a.example.com"}}},
		}},
		services: []*corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "web-svc", Namespace: "ns"},
			Spec:       corev1.ServiceSpec{Type: corev1.ServiceTypeNodePort, Ports: []corev1.ServicePort{{Name: "web", Port: 80, NodePort: 30080}}},
		}},
		httpRoutes: []*gatewayv1.HTTPRoute{{
			ObjectMeta: metav1.ObjectMeta{Name: "web-route", Namespace: "ns"},
			Spec:       gatewayv1.HTTPRouteSpec{Hostnames: []gatewayv1.Hostname{"x.example.com"}},
		}},
	}
	b := newProjectBizWithK8s(f, k)
	got, err := b.GetProjectEndpointsInNamespace(context.TODO(), "ns", 1)
	assert.NoError(t, err)
	assert.True(t, f.findByIDsCalled)
	// ing + nodePort + httpRoute 三类来源聚合（LB 无对象不产出）。
	urls := make([]string, 0, len(got))
	for _, e := range got {
		urls = append(urls, e.Url)
	}
	assert.ElementsMatch(t, []string{"http://a.example.com", "http://10.0.0.1:30080", "https://x.example.com"}, urls)
}

func TestProjectBiz_GetAllActiveContainers_ShowErr(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{showErr: errors.New("show down")}
	b := newProjectBizWithK8s(f, &fakeEndpointK8sRepo{})
	got, err := b.GetAllActiveContainers(context.TODO(), 1)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "show down")
}

func TestProjectBiz_ResourceTree(t *testing.T) {
	uid := kmetatypes.UID("dep-uid")
	f := &fakeProjectRepoForProjectBiz{}
	k := &fakeTreeK8sRepo{
		deployments:  map[string]*appsv1.Deployment{"app": readyDeployment("app", uid)},
		manifestDeps: []*appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "app"}}},
	}
	b := newProjectBizWithK8s(f, k)
	got, err := b.ResourceTree(context.TODO(), 1)
	assert.NoError(t, err)
	assert.True(t, f.showCalled)
	if assert.NotNil(t, got) {
		assert.Equal(t, "application-1", got.Nodes[0].ID)
		if assert.Len(t, got.Nodes, 2) {
			assert.Equal(t, "deployment-app", got.Nodes[1].ID)
			assert.Equal(t, "healthy", got.Nodes[1].Status)
		}
		assert.True(t, hasTreeEdge(got, "owner", "application-1", "deployment-app"))
	}
}

func TestProjectBiz_ResourceTree_ShowErr(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{showErr: errors.New("show down")}
	b := newProjectBizWithK8s(f, &fakeTreeK8sRepo{})
	got, err := b.ResourceTree(context.TODO(), 1)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "show down")
}

func TestProjectBiz_GetProjectEndpointsInNamespace_FindByIDsErr(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{findByIDsErr: errors.New("ids down")}
	b := newProjectBizWithK8s(f, &fakeEndpointK8sRepo{})
	got, err := b.GetProjectEndpointsInNamespace(context.TODO(), "ns", 1)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "ids down")
}

// 以下四个测试覆盖 GetProjectEndpointsInNamespace 内各 Build* 的 List 失败分支：
// 任一来源失败即整体上抛。NodePort 与 LB 共用 ListServices，故用 servicesFailOnCall
// 指定在第几次调用返回错误，以分别命中 LoadBalancer 与 NodePort 的错误分支。

func TestProjectBiz_GetProjectEndpointsInNamespace_IngressErr(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	k := &fakeEndpointK8sRepo{listIngressesErr: errors.New("ing down")}
	b := newProjectBizWithK8s(f, k)
	got, err := b.GetProjectEndpointsInNamespace(context.TODO(), "ns", 1)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "ing down")
}

func TestProjectBiz_GetProjectEndpointsInNamespace_LBErr(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	k := &fakeEndpointK8sRepo{listServicesErr: errors.New("svc down")}
	b := newProjectBizWithK8s(f, k)
	got, err := b.GetProjectEndpointsInNamespace(context.TODO(), "ns", 1)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "svc down")
}

func TestProjectBiz_GetProjectEndpointsInNamespace_NodePortErr(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	// 第一次 ListServices（LB 调用）成功，第二次（NodePort 调用）失败。
	k := &fakeEndpointK8sRepo{listServicesErr: errors.New("svc down"), servicesFailOnCall: 2}
	b := newProjectBizWithK8s(f, k)
	got, err := b.GetProjectEndpointsInNamespace(context.TODO(), "ns", 1)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "svc down")
}

func TestProjectBiz_GetProjectEndpointsInNamespace_HTTPRouteErr(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	k := &fakeEndpointK8sRepo{gatewayInstalled: true, listHTTPRoutesErr: errors.New("route down")}
	b := newProjectBizWithK8s(f, k)
	got, err := b.GetProjectEndpointsInNamespace(context.TODO(), "ns", 1)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "route down")
}

// ---- 纯透传查询 ----

func TestProjectBiz_ListAll(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	got, err := b.ListAll(context.TODO())
	assert.NoError(t, err)
	assert.True(t, f.listAllCalled)
	assert.Len(t, got, 1)
	assert.Equal(t, "proj1", got[0].Name)
}

func TestProjectBiz_List(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	got, pag, err := b.List(context.TODO(), &ListProjectInput{})
	assert.NoError(t, err)
	assert.True(t, f.listCalled)
	assert.Len(t, got, 1)
	assert.Nil(t, pag)
}

func TestProjectBiz_Show(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	got, err := b.Show(context.TODO(), 5)
	assert.NoError(t, err)
	assert.True(t, f.showCalled)
	assert.Equal(t, 5, got.ID)
}

func TestProjectBiz_Version(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	ver, err := b.Version(context.TODO(), 1)
	assert.NoError(t, err)
	assert.True(t, f.versionCalled)
	assert.Equal(t, 3, ver)
}

func TestProjectBiz_FindByName(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	got, err := b.FindByName(context.TODO(), "app", 2)
	assert.NoError(t, err)
	assert.True(t, f.findByNameCalled)
	assert.Equal(t, "app", got.Name)
}

func TestProjectBiz_UpdateDeployStatus(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	got, err := b.UpdateDeployStatus(context.TODO(), 1, types.Deploy_StatusDeployed)
	assert.NoError(t, err)
	assert.True(t, f.updateDeployStatusCalled)
	assert.Equal(t, types.Deploy_StatusDeployed, got.DeployStatus)
}

func TestProjectBiz_UpdateVersion(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	got, err := b.UpdateVersion(context.TODO(), 1, 9)
	assert.NoError(t, err)
	assert.True(t, f.updateVersionCalled)
	assert.Equal(t, 9, got.Version)
}

func TestProjectBiz_FindByVersion(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	got, err := b.FindByVersion(context.TODO(), 1, 4)
	assert.NoError(t, err)
	assert.True(t, f.findByVersionCalled)
	assert.Equal(t, 4, got.Version)
}

func TestProjectBiz_UpdateStatusByVersion(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{}
	b := newProjectBizForTest(f)
	got, err := b.UpdateStatusByVersion(context.TODO(), 1, types.Deploy_StatusFailed, 4)
	assert.NoError(t, err)
	assert.True(t, f.updateStatusByVersionCalled)
	assert.Equal(t, types.Deploy_StatusFailed, got.DeployStatus)
	assert.Equal(t, 4, got.Version)
}

// ---- 项目活跃度 ----

// Test_ClassifyLiveness 表驱动锁定分类边界：≤30 天活跃、≥90 天僵尸、中间休眠。
func Test_ClassifyLiveness(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		updatedAt time.Time
		want      LivenessKind
	}{
		{"刚更新", now.Add(-time.Hour), LivenessActive},
		{"活跃边界 30 天", now.Add(-30 * 24 * time.Hour), LivenessActive},
		{"休眠 60 天", now.Add(-60 * 24 * time.Hour), LivenessDormant},
		{"僵尸边界 90 天", now.Add(-90 * 24 * time.Hour), LivenessZombie},
		{"僵尸 120 天", now.Add(-120 * 24 * time.Hour), LivenessZombie},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ClassifyLiveness(tt.updatedAt, now))
		})
	}
}

// TestProjectBiz_Liveness 薄传递：分类/统计/过滤/分页由 repo 下沉 SQL，biz 透传入参并
// 装配部署次数与结果。
func TestProjectBiz_Liveness(t *testing.T) {
	now := time.Now()
	f := &fakeProjectRepoForProjectBiz{
		listLivenessPageResult: &LivenessPageResult{
			Projects: []*Project{
				{ID: 1, UpdatedAt: now.Add(-10 * 24 * time.Hour)},
				{ID: 2, UpdatedAt: now.Add(-60 * 24 * time.Hour)},
				{ID: 3, UpdatedAt: now.Add(-120 * 24 * time.Hour)},
			},
			Count: 3,
			Stats: LivenessStats{Total: 3, Active: 1, Dormant: 1, Zombie: 1},
		},
	}
	cl := &fakeClRepoForLiveness{counts: map[int]int{1: 5, 2: 3, 3: 0}}
	b := newProjectBizWithClRepo(f, cl)
	res, err := b.Liveness(context.TODO(), &LivenessInput{Page: 1, PageSize: 10, Search: "app"})
	assert.NoError(t, err)
	assert.True(t, f.listLivenessCalled)
	if assert.NotNil(t, f.listLivenessPageQuery) {
		assert.Equal(t, "app", f.listLivenessPageQuery.Search)
		assert.Equal(t, int32(1), f.listLivenessPageQuery.Page)
		assert.Equal(t, int32(10), f.listLivenessPageQuery.PageSize)
		// Now 为捕获的分类基准时间，注入 repo 与行级标记共用同一基准。
		assert.False(t, f.listLivenessPageQuery.Now.IsZero())
	}
	if assert.NotNil(t, res) {
		assert.Equal(t, LivenessStats{Total: 3, Active: 1, Dormant: 1, Zombie: 1}, res.Stats)
		assert.Equal(t, int32(3), res.Count)
		if assert.Len(t, res.Items, 3) {
			assert.Equal(t, 5, res.Items[0].DeployCount)
			assert.Equal(t, 3, res.Items[1].DeployCount)
			assert.Equal(t, 0, res.Items[2].DeployCount)
		}
	}
}

// TestProjectBiz_Liveness_FilterByKind 分类过滤透传：Liveness 参数透传 repo，统计/条目
// 直接反映 repo 结果（SQL 侧过滤）。
func TestProjectBiz_Liveness_FilterByKind(t *testing.T) {
	now := time.Now()
	f := &fakeProjectRepoForProjectBiz{
		listLivenessPageResult: &LivenessPageResult{
			Projects: []*Project{{ID: 3, UpdatedAt: now.Add(-120 * 24 * time.Hour)}},
			Count:    1,
			Stats:    LivenessStats{Total: 3, Active: 1, Dormant: 1, Zombie: 1},
		},
	}
	b := newProjectBizForTest(f)
	res, err := b.Liveness(context.TODO(), &LivenessInput{Page: 1, PageSize: 10, Liveness: "zombie"})
	assert.NoError(t, err)
	if assert.NotNil(t, f.listLivenessPageQuery) {
		assert.Equal(t, "zombie", f.listLivenessPageQuery.Liveness)
	}
	if assert.NotNil(t, res) {
		assert.Equal(t, LivenessStats{Total: 3, Active: 1, Dormant: 1, Zombie: 1}, res.Stats)
		assert.Equal(t, int32(1), res.Count)
		assert.Len(t, res.Items, 1)
	}
}

// TestProjectBiz_Liveness_Pagination 分页参数透传：Page/PageSize 交给 repo，结果直接反映
// repo 返回的页内条目。
func TestProjectBiz_Liveness_Pagination(t *testing.T) {
	now := time.Now()
	f := &fakeProjectRepoForProjectBiz{
		listLivenessPageResult: &LivenessPageResult{
			Projects: []*Project{{ID: 3, UpdatedAt: now.Add(-3 * 24 * time.Hour)}, {ID: 4, UpdatedAt: now.Add(-4 * 24 * time.Hour)}},
			Count:    5,
		},
	}
	b := newProjectBizForTest(f)
	res, err := b.Liveness(context.TODO(), &LivenessInput{Page: 2, PageSize: 2})
	assert.NoError(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, int32(5), res.Count)
		assert.Equal(t, int32(2), res.Page)
		assert.Equal(t, int32(2), res.PageSize)
		if assert.Len(t, res.Items, 2) {
			assert.Equal(t, 3, res.Items[0].Project.ID)
			assert.Equal(t, 4, res.Items[1].Project.ID)
		}
	}
}

// TestProjectBiz_Liveness_PaginationOutOfRange 越界页：repo 返回空条目但计数保留全量。
func TestProjectBiz_Liveness_PaginationOutOfRange(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{
		listLivenessPageResult: &LivenessPageResult{Projects: nil, Count: 5},
	}
	b := newProjectBizForTest(f)
	res, err := b.Liveness(context.TODO(), &LivenessInput{Page: 99, PageSize: 10})
	assert.NoError(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, int32(5), res.Count)
		assert.Empty(t, res.Items)
	}
}

// TestProjectBiz_Liveness_ListLivenessPageErr 分页查询失败整体上抛。
func TestProjectBiz_Liveness_ListLivenessPageErr(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{listLivenessErr: errors.New("list down")}
	b := newProjectBizForTest(f)
	res, err := b.Liveness(context.TODO(), &LivenessInput{})
	assert.Nil(t, res)
	assert.ErrorContains(t, err, "list down")
}

// TestProjectBiz_Liveness_CountErr 部署次数聚合失败整体上抛。
func TestProjectBiz_Liveness_CountErr(t *testing.T) {
	f := &fakeProjectRepoForProjectBiz{listLivenessPageResult: &LivenessPageResult{Projects: []*Project{{ID: 1, UpdatedAt: time.Now()}}}}
	cl := &fakeClRepoForLiveness{err: errors.New("count down")}
	b := newProjectBizWithClRepo(f, cl)
	res, err := b.Liveness(context.TODO(), &LivenessInput{})
	assert.Nil(t, res)
	assert.ErrorContains(t, err, "count down")
}
