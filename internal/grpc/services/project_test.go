package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"helm.sh/helm/v3/pkg/action"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	fake2 "k8s.io/metrics/pkg/client/clientset/versioned/fake"

	"github.com/duc-cnzj/mars-client/v4/mars"
	"github.com/duc-cnzj/mars-client/v4/project"
	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/socket"
	"github.com/duc-cnzj/mars/internal/testutil"
)

func TestProjectSvc_AllContainers(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	p := &models.Project{
		Namespace:    models.Namespace{Name: "test"},
		PodSelectors: "c=c",
	}
	db.Create(p)
	fk := fake.NewSimpleClientset(
		&v1.PodList{
			Items: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod3",
						Namespace: "test",
						Labels: map[string]string{
							"c": "c",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "c1",
							},
						},
					},
					// FIXME: kubeclient 不能做 fieldSelector 过滤
					//Status: v1.PodStatus{
					//	Phase: v1.PodFailed,
					//},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: "test",
						Labels: map[string]string{
							"b": "b",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "c3",
							},
							{
								Name: "c4",
							},
						},
					},
				},
			},
		})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fk}).AnyTimes()
	containers, err := new(ProjectSvc).AllContainers(context.TODO(), &project.AllContainersRequest{ProjectId: int64(p.ID)})
	assert.Nil(t, err)
	assert.Len(t, containers.Items, 1)
	assert.Equal(t, "pod3", containers.Items[0].Pod)
	assert.Equal(t, "c1", containers.Items[0].Container)
}

func TestProjectSvc_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	app.EXPECT().IsDebug().Return(false).AnyTimes()
	_, err := new(ProjectSvc).Show(adminCtx(), &project.ShowRequest{
		ProjectId: 990,
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, fromError.Code())
	db.AutoMigrate(&models.Project{}, &models.Namespace{}, &models.GitProject{})
	_, err = new(ProjectSvc).Show(adminCtx(), &project.ShowRequest{
		ProjectId: 990,
	})
	fromError, _ = status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())
	ing1 := v12.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "yyy",
			Labels: map[string]string{
				"app.kubernetes.io/instance": "yyy",
			}},
		Spec: v12.IngressSpec{
			Rules: []v12.IngressRule{
				{
					Host: "yyy.com",
				},
			},
		},
	}
	svc1 := v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "svc1",
			Labels: map[string]string{
				"app.kubernetes.io/instance": "yyy",
			},
		},
		Spec: v1.ServiceSpec{
			Type: "NodePort",
			Ports: []v1.ServicePort{
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
			},
		},
	}
	p := &models.Project{Namespace: models.Namespace{Name: "test"}, GitProjectId: 100, Name: "yyy", PodSelectors: "a=a", Manifest: strings.Join(encodeToYaml(&ing1, &svc1), "---")}
	db.Create(p)
	mcfg := mars.Config{
		Elements: []*mars.Element{
			{
				Path:         "conf->env",
				Type:         0,
				Default:      "xx",
				Description:  "xx",
				SelectValues: []string{"1", "2", "3"},
			},
		},
	}
	clone := proto.Clone(&mcfg)
	marshal, _ := json.Marshal(clone)
	db.Create(&models.GitProject{
		DefaultBranch: "dev",
		Name:          "gitcfg",
		GitProjectId:  100,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})

	fm := &fake2.Clientset{}
	ex := []v1beta1.PodMetrics{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "test", ResourceVersion: "10", Labels: map[string]string{"a": "a"}},
			Window:     metav1.Duration{Duration: time.Minute},
			Containers: []v1beta1.ContainerMetrics{
				{
					Name: "container1-2",
					Usage: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(4, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
					},
				},
			},
		},
	}
	fm.AddReactor("list", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		res := &v1beta1.PodMetricsList{
			ListMeta: metav1.ListMeta{
				ResourceVersion: "2",
			},
			Items: ex,
		}
		return true, res, nil
	})
	fk := fake.NewSimpleClientset(&v1.ServiceList{
		Items: []v1.Service{
			svc1,
		},
	},
		&v12.IngressList{
			Items: []v12.Ingress{
				ing1,
			},
		},
	)

	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		MetricsClient: fm,
		Client:        fk,
	}).AnyTimes()
	app.EXPECT().Config().Return(&config.Config{ExternalIp: "127.0.0.1"})

	show, err := new(ProjectSvc).Show(adminCtx(), &project.ShowRequest{
		ProjectId: int64(p.ID),
	})
	assert.Nil(t, err)
	assert.Equal(t, "4 m", show.Cpu)
	assert.Equal(t, "5.0 MB", show.Memory)
	assert.Len(t, show.Urls, 3)
	assert.Equal(t, p.ProtoTransform().String(), show.Project.String())
	assert.Equal(t, "conf->env", show.Elements[0].Path)
	assert.Equal(t, "xx", show.Elements[0].Description)
	assert.Equal(t, "xx", show.Elements[0].Default)
	assert.Equal(t, []string{"1", "2", "3"}, show.Elements[0].SelectValues)
}

func TestProjectSvc_Delete(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	app.EXPECT().IsDebug().Return(false).AnyTimes()
	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	p := &models.Project{Namespace: models.Namespace{Name: "test"}}
	db.Create(p)
	d := assertAuditLogFired(m, app)
	assert.False(t, p.DeletedAt.Valid)
	d.EXPECT().Dispatch(events.EventProjectDeleted, gomock.Any()).Times(1)
	_, err := (&ProjectSvc{UninstallReleaseFunc: func(releaseName, namespace string, log action.DebugLog) error {
		return nil
	}}).Delete(adminCtx(), &project.DeleteRequest{
		ProjectId: int64(p.ID),
	})
	assert.Nil(t, err)
	db.Unscoped().First(&p)
	assert.True(t, p.DeletedAt.Valid)
}

func TestProjectSvc_Apply(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	job := mock.NewMockJob(m)
	msg := mock.NewMockDeployMsger(m)
	job.EXPECT().Messager().Return(msg).AnyTimes()
	job.EXPECT().Validate().Return(nil).Times(1)
	job.EXPECT().LoadConfigs().Return(nil).Times(1)
	job.EXPECT().Run().Return(nil).Times(1)
	pubsub := mock.NewMockPubSub(m)
	job.EXPECT().PubSub().Return(pubsub).Times(1)
	job.EXPECT().CallDestroyFuncs().Times(1)
	job.EXPECT().Finish().Times(1)
	pubsub.EXPECT().ToOthers(gomock.Any()).Times(1)
	app := testutil.MockApp(m)
	ws := mockWsServer(m, app)
	ps := mock.NewMockPubSub(m)
	ws.EXPECT().New("", "").Return(ps)
	req := &project.ApplyRequest{
		NamespaceId:   1,
		Name:          "aaa",
		GitProjectId:  100,
		GitBranch:     "dev",
		GitCommit:     "xxx",
		Config:        "cfg",
		WebsocketSync: true,
	}
	job.EXPECT().Stop(gomock.Any()).Times(0)
	ma := &mockApplyServer{}
	(&ProjectSvc{NewJobFunc: func(input *websocket.CreateProjectInput, user contracts.UserInfo, slugName string, messager contracts.DeployMsger, pubsub contracts.PubSub, timeoutSeconds int64, opts ...socket.Option) contracts.Job {
		return job
	}}).Apply(req, ma)
}

type mockJob struct {
	sync.Mutex
	t     *testing.T
	msger contracts.DeployMsger
	contracts.Job
	e error
}

func (m *mockJob) Validate() error {
	time.Sleep(1 * time.Second)
	return errors.New("timeout")
}

func (m *mockJob) GetStoppedErrorIfHas() error {
	m.Lock()
	defer m.Unlock()
	return m.e
}

func (m *mockJob) IsDryRun() bool {
	return true
}

func (m *mockJob) Finish() {
}

func (m *mockJob) Messager() contracts.DeployMsger {
	return m.msger
}

func (m *mockJob) Stop(e error) {
	m.Lock()
	defer m.Unlock()
	m.e = e
	assert.Equal(m.t, "context canceled", e.Error())
}

func (m *mockJob) CallDestroyFuncs() {
}

func (m *mockJob) ProjectModel() *types.ProjectModel {
	return nil
}

func TestProjectSvc_Apply_WithClientStop(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := mock.NewMockDeployMsger(m)
	msger.EXPECT().SendDeployedResult(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	msger.EXPECT().Stop(gomock.Any()).Times(1)
	req := &project.ApplyRequest{
		NamespaceId:  1,
		Name:         "aaa",
		GitProjectId: 100,
		GitBranch:    "dev",
		GitCommit:    "xxx",
		Config:       "cfg",
	}

	ctx, cancel := context.WithCancel(adminCtx())
	cancel()
	ma := &mockApplyServer{ctx: ctx}

	err := (&ProjectSvc{NewJobFunc: func(input *websocket.CreateProjectInput, user contracts.UserInfo, slugName string, messager contracts.DeployMsger, pubsub contracts.PubSub, timeoutSeconds int64, opts ...socket.Option) contracts.Job {
		return &mockJob{msger: msger, t: t}
	}}).Apply(req, ma)
	assert.Equal(t, "context canceled", err.Error())
}

func TestProjectSvc_ApplyDryRun(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	job := mock.NewMockJob(m)
	msg := mock.NewMockDeployMsger(m)
	req := &project.ApplyRequest{
		NamespaceId:  1,
		Name:         "aaa",
		GitProjectId: 100,
		GitBranch:    "dev",
		Config:       "cfg",
	}

	app := testutil.MockApp(m)
	gits := mockGitServer(m, app)
	gits.EXPECT().ListCommits(gomock.Any(), gomock.Any()).Return(nil, nil)
	_, err := (&ProjectSvc{NewJobFunc: func(input *websocket.CreateProjectInput, user contracts.UserInfo, slugName string, messager contracts.DeployMsger, pubsub contracts.PubSub, timeoutSeconds int64, opts ...socket.Option) contracts.Job {
		return job
	}}).ApplyDryRun(adminCtx(), req)
	assert.Error(t, err)
	req.GitCommit = "xxx"

	job.EXPECT().Messager().Return(msg).AnyTimes()
	job.EXPECT().Validate().Return(nil).Times(1)
	job.EXPECT().LoadConfigs().Return(nil).Times(1)
	job.EXPECT().Run().Return(nil).Times(1)
	pubsub := mock.NewMockPubSub(m)
	job.EXPECT().PubSub().Return(pubsub).Times(1)
	job.EXPECT().CallDestroyFuncs().Times(1)
	job.EXPECT().Finish().Times(1)
	pubsub.EXPECT().ToOthers(gomock.Any()).Times(1)
	job.EXPECT().Manifests().Return([]string{"Manifests"}).Times(1)

	run, err := (&ProjectSvc{NewJobFunc: func(input *websocket.CreateProjectInput, user contracts.UserInfo, slugName string, messager contracts.DeployMsger, pubsub contracts.PubSub, timeoutSeconds int64, opts ...socket.Option) contracts.Job {
		return job
	}}).ApplyDryRun(adminCtx(), req)
	assert.Nil(t, err)
	assert.Equal(t, []string{"Manifests"}, run.Results)
}

func TestProjectSvc_ApplyDryRun_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := mock.NewMockDeployMsger(m)
	msger.EXPECT().SendDeployedResult(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	msger.EXPECT().Stop(gomock.Any()).Times(1)
	req := &project.ApplyRequest{
		NamespaceId:  1,
		Name:         "aaa",
		GitProjectId: 100,
		GitBranch:    "dev",
		GitCommit:    "xxx",
		Config:       "cfg",
	}

	ctx, cancel := context.WithCancel(adminCtx())
	cancel()

	run, err := (&ProjectSvc{NewJobFunc: func(input *websocket.CreateProjectInput, user contracts.UserInfo, slugName string, messager contracts.DeployMsger, pubsub contracts.PubSub, timeoutSeconds int64, opts ...socket.Option) contracts.Job {
		return &mockJob{msger: msger, t: t}
	}}).ApplyDryRun(ctx, req)
	assert.Nil(t, run)
	assert.Equal(t, "context canceled", err.Error())
}

func TestProjectSvc_List(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	_, err := new(ProjectSvc).List(context.TODO(), &project.ListRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.Error(t, err)
	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	p := &models.Project{
		Name:             "duc",
		GitBranch:        "dev",
		GitCommit:        "commit",
		Config:           "cfg",
		OverrideValues:   "xx",
		DockerImage:      "xx:v1",
		PodSelectors:     "a=b",
		Atomic:           true,
		DeployStatus:     1,
		EnvValues:        "x",
		ExtraValues:      "xa",
		FinalExtraValues: "xaa",
		ConfigType:       "yaml",
		GitCommitWebUrl:  "url",
		GitCommitTitle:   "title",
		GitCommitAuthor:  "author",
		Namespace: models.Namespace{
			Name: "ns",
		},
	}
	db.Create(p)
	list, _ := new(ProjectSvc).List(context.TODO(), &project.ListRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.Equal(t, p.ProtoTransform(), list.Items[0])
}

func TestProjectSvc_completeInput(t *testing.T) {
	req := &project.ApplyRequest{
		Name:      "xxx",
		GitCommit: "xxx",
	}
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	gits := mockGitServer(m, app)
	msger := mock.NewMockDeployMsger(m)
	msger.EXPECT().SendMsg(gomock.Any()).Times(0)
	new(ProjectSvc).completeInput(req, msger)
	req.GitCommit = ""
	req.Name = ""
	gits.EXPECT().ListCommits(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	err := new(ProjectSvc).completeInput(req, msger)
	assert.Equal(t, "没有可用的 commit", err.Error())
	commit := mock.NewMockCommitInterface(m)
	gits.EXPECT().ListCommits(gomock.Any(), gomock.Any()).Return([]contracts.CommitInterface{commit}, nil).Times(1)
	commit.EXPECT().GetID().Return("1")
	commit.EXPECT().GetTitle().Return("").Times(1)
	commit.EXPECT().GetWebURL().Return("").Times(1)
	msger.EXPECT().SendMsg(gomock.Any()).Times(1)
	project := mock.NewMockProjectInterface(m)
	project.EXPECT().GetName().Return("pro").Times(1)
	gits.EXPECT().GetProject(gomock.Any()).Return(project, nil).Times(1)
	err = new(ProjectSvc).completeInput(req, msger)
	assert.Nil(t, err)
	assert.Equal(t, "pro", req.Name)
	assert.Equal(t, "1", req.GitCommit)
}

func Test_messager_SendDeployedResult(t *testing.T) {
	m := &messager{server: &mockApplyServer{}, sendPercent: false}
	m.SendDeployedResult(websocket.ResultType_Deployed, "", nil)
	assert.Equal(t, 1, m.server.(*mockApplyServer).send)
	assert.Equal(t, websocket.ResultType_Deployed, m.server.(*mockApplyServer).ar.Metadata.Result)
	assert.Equal(t, true, m.server.(*mockApplyServer).ar.Metadata.End)
}

func Test_messager_SendEndError(t *testing.T) {
	m := &messager{server: &mockApplyServer{}, sendPercent: false}
	m.SendEndError(errors.New("a"))
	assert.Equal(t, 1, m.server.(*mockApplyServer).send)
	assert.Equal(t, websocket.ResultType_Error, m.server.(*mockApplyServer).ar.Metadata.Result)
	assert.Equal(t, true, m.server.(*mockApplyServer).ar.Metadata.End)
}

func Test_messager_SendError(t *testing.T) {
	m := &messager{server: &mockApplyServer{}, sendPercent: false}
	m.SendError(errors.New("a"))
	assert.Equal(t, 1, m.server.(*mockApplyServer).send)
	assert.Equal(t, websocket.ResultType_Error, m.server.(*mockApplyServer).ar.Metadata.Result)
	assert.Equal(t, false, m.server.(*mockApplyServer).ar.Metadata.End)
}

func Test_messager_SendMsg(t *testing.T) {
	m := &messager{server: &mockApplyServer{}, sendPercent: false}
	m.SendMsg("")
	assert.Equal(t, 1, m.server.(*mockApplyServer).send)
	assert.Equal(t, websocket.ResultType_Success, m.server.(*mockApplyServer).ar.Metadata.Result)
	assert.Equal(t, false, m.server.(*mockApplyServer).ar.Metadata.End)
}

func Test_messager_SendProcessPercent(t *testing.T) {
	m := &messager{server: &mockApplyServer{}, sendPercent: false}
	m.SendProcessPercent("10")
	assert.Equal(t, 0, m.server.(*mockApplyServer).send)
	m.sendPercent = true
	m.SendProcessPercent("10")
	assert.Equal(t, 1, m.server.(*mockApplyServer).send)
	assert.Equal(t, websocket.Type_ProcessPercent, m.server.(*mockApplyServer).ar.Metadata.Type)
	assert.Equal(t, websocket.ResultType_Success, m.server.(*mockApplyServer).ar.Metadata.Result)
	assert.Equal(t, false, m.server.(*mockApplyServer).ar.Metadata.End)
}

func Test_messager_SendProtoMsg(t *testing.T) {
	m := &messager{server: &mockApplyServer{}}
	m.SendProtoMsg(&websocket.WsHandleClusterResponse{})
	assert.Equal(t, 1, m.server.(*mockApplyServer).send)
}

func Test_messager_Stop(t *testing.T) {
	m := &messager{}
	m.Stop(nil)
	assert.True(t, m.IsStopped())
}

func Test_messager_send(t *testing.T) {
	m := &messager{server: &mockApplyServer{}}
	m.send(nil)
	assert.Equal(t, 1, m.server.(*mockApplyServer).send)
	m.Stop(nil)
	assert.Equal(t, 1, m.server.(*mockApplyServer).send)
}

func Test_newEmptyMessager(t *testing.T) {
	assert.Implements(t, (*contracts.DeployMsger)(nil), newEmptyMessager())
}

type mockApplyServer struct {
	send int
	ctx  context.Context
	ar   *project.ApplyResponse
	project.Project_ApplyServer
}

func (m *mockApplyServer) Send(ar *project.ApplyResponse) error {
	m.send++
	m.ar = ar
	return nil
}

func (m *mockApplyServer) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return adminCtx()
}

func TestEmptyMessager(t *testing.T) {
	em := &emptyMessager{}
	em.SendEndError(nil)
	em.SendError(nil)
	em.SendDeployedResult(0, "", nil)
	em.SendProcessPercent("10")
	em.SendMsg("")
	em.SendProtoMsg(nil)
	em.Stop(nil)
	assert.True(t, true)
}
