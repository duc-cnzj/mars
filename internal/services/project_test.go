package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v4/project"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/socket"
	"github.com/duc-cnzj/mars/v4/internal/util/pagination"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestNewProjectSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		repo.NewMockProjectRepo(m),
		repo.NewMockGitRepo(m),
		repo.NewMockK8sRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockHelmerRepo(m),
		repo.NewMockNamespaceRepo(m),
	)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*projectSvc).repoRepo)
	assert.NotNil(t, svc.(*projectSvc).plMgr)
	assert.NotNil(t, svc.(*projectSvc).jobManager)
	assert.NotNil(t, svc.(*projectSvc).projRepo)
	assert.NotNil(t, svc.(*projectSvc).gitRepo)
	assert.NotNil(t, svc.(*projectSvc).k8sRepo)
	assert.NotNil(t, svc.(*projectSvc).eventRepo)
	assert.NotNil(t, svc.(*projectSvc).logger)
	assert.NotNil(t, svc.(*projectSvc).helmer)
	assert.NotNil(t, svc.(*projectSvc).nsRepo)
}

func Test_projectSvc_List(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		repo.NewMockK8sRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockHelmerRepo(m),
		repo.NewMockNamespaceRepo(m),
	)

	projectRepo.EXPECT().List(gomock.Any(), &repo.ListProjectInput{
		Page:          1,
		PageSize:      11,
		OrderByIDDesc: lo.ToPtr(true),
	}).Return(nil, nil, errors.New("x"))
	list, err := svc.List(context.TODO(), &project.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(11)),
	})
	assert.Error(t, err)
	assert.Nil(t, list)
}

func Test_projectSvc_List_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		repo.NewMockK8sRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockHelmerRepo(m),
		repo.NewMockNamespaceRepo(m),
	)

	projectRepo.EXPECT().List(gomock.Any(), &repo.ListProjectInput{
		Page:          1,
		PageSize:      11,
		OrderByIDDesc: lo.ToPtr(true),
	}).Return([]*repo.Project{}, &pagination.Pagination{}, nil)
	list, err := svc.List(context.TODO(), &project.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(11)),
	})
	assert.Nil(t, err)
	assert.NotNil(t, list)
}

func TestProjectSvc_Show_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		k8sRepo,
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockHelmerRepo(m),
		repo.NewMockNamespaceRepo(m),
	)

	k8sRepo.EXPECT().GetAllPodMetrics(gomock.Any()).Return([]v1beta1.PodMetrics{})
	k8sRepo.EXPECT().GetCpuAndMemory(gomock.Any()).Return("", "")
	projectRepo.EXPECT().GetNodePortMappingByProjects(gomock.Any(), gomock.Any()).Return(repo.EndpointMapping{})
	projectRepo.EXPECT().GetIngressMappingByProjects(gomock.Any(), gomock.Any()).Return(repo.EndpointMapping{})
	projectRepo.EXPECT().GetLoadBalancerMappingByProjects(gomock.Any(), gomock.Any()).Return(repo.EndpointMapping{})

	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Project{
		Namespace: &repo.Namespace{},
	}, nil)
	res, err := svc.Show(context.TODO(), &project.ShowRequest{
		Id: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestProjectSvc_Show_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		repo.NewMockK8sRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockHelmerRepo(m),
		repo.NewMockNamespaceRepo(m),
	)

	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	res, err := svc.Show(context.TODO(), &project.ShowRequest{
		Id: 1,
	})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func Test_projectSvc_Delete(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	helmerRepo := repo.NewMockHelmerRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		k8sRepo,
		eventRepo,
		mlog.NewLogger(nil),
		helmerRepo,
		repo.NewMockNamespaceRepo(m),
	)
	helmerRepo.EXPECT().Uninstall("app", "ns", gomock.Any()).Return(errors.New("x"))
	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Project{
		ID:          2,
		Name:        "app",
		NamespaceID: 1,
		Namespace: &repo.Namespace{
			Name: "ns",
		},
	}, nil)
	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	eventRepo.EXPECT().Dispatch(repo.EventProjectDeleted, &repo.ProjectDeletedPayload{
		NamespaceID: 1,
		ProjectID:   2,
	})
	projectRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)
	response, err := svc.Delete(newAdminUserCtx(), &project.DeleteRequest{
		Id: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, response)
}
func Test_projectSvc_Delete_Fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	helmerRepo := repo.NewMockHelmerRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		k8sRepo,
		eventRepo,
		mlog.NewLogger(nil),
		helmerRepo,
		repo.NewMockNamespaceRepo(m),
	)
	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	response, err := svc.Delete(newAdminUserCtx(), &project.DeleteRequest{
		Id: 1,
	})
	assert.NotNil(t, err)
	assert.Nil(t, response)
}

func Test_projectSvc_Version(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	helmerRepo := repo.NewMockHelmerRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		k8sRepo,
		eventRepo,
		mlog.NewLogger(nil),
		helmerRepo,
		repo.NewMockNamespaceRepo(m),
	)
	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Project{
		ID:          2,
		Name:        "app",
		NamespaceID: 1,
		Namespace: &repo.Namespace{
			Name: "ns",
		},
		Version: 100,
	}, nil)
	version, err := svc.Version(context.Background(), &project.VersionRequest{Id: 1})
	assert.Nil(t, err)
	assert.NotNil(t, version)
	assert.Equal(t, int32(100), version.Version)
}

func Test_projectSvc_AllContainers(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	helmerRepo := repo.NewMockHelmerRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		k8sRepo,
		eventRepo,
		mlog.NewLogger(nil),
		helmerRepo,
		repo.NewMockNamespaceRepo(m),
	)
	projectRepo.EXPECT().GetAllPods(gomock.Any(), 1).Return([]*types.StateContainer{}, nil)
	containers, err := svc.AllContainers(context.TODO(), &project.AllContainersRequest{
		Id: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, containers)
}
func Test_projectSvc_AllContainers_Fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	helmerRepo := repo.NewMockHelmerRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		k8sRepo,
		eventRepo,
		mlog.NewLogger(nil),
		helmerRepo,
		repo.NewMockNamespaceRepo(m),
	)
	projectRepo.EXPECT().GetAllPods(gomock.Any(), 1).Return(nil, errors.New("x"))
	containers, err := svc.AllContainers(context.TODO(), &project.AllContainersRequest{
		Id: 1,
	})
	assert.NotNil(t, err)
	assert.Nil(t, containers)
}

func Test_projectSvc_WebApply(t *testing.T) {}

func Test_projectSvc_Apply(t *testing.T) {}

func TestNewMessager(t *testing.T) {}

func Test_emptyMessager_Add(t *testing.T) {}

func Test_emptyMessager_Current(t *testing.T) {}

func Test_emptyMessager_SendDeployedResult(t *testing.T) {}

func Test_emptyMessager_SendEndError(t *testing.T) {}

func Test_emptyMessager_SendError(t *testing.T) {}

func Test_emptyMessager_SendMsg(t *testing.T) {}

func Test_emptyMessager_SendMsgWithContainerLog(t *testing.T) {}

func Test_emptyMessager_SendProcessPercent(t *testing.T) {}

func Test_emptyMessager_SendProtoMsg(t *testing.T) {}

func Test_emptyMessager_To(t *testing.T) {}

func Test_isContainerReady(t *testing.T) {}

func Test_messager_Add(t *testing.T) {}

func Test_messager_Current(t *testing.T) {}

func Test_messager_SendDeployedResult(t *testing.T) {}

func Test_messager_SendEndError(t *testing.T) {}

func Test_messager_SendError(t *testing.T) {}

func Test_messager_SendMsg(t *testing.T) {}

func Test_messager_SendMsgWithContainerLog(t *testing.T) {}

func Test_messager_SendProcessPercent(t *testing.T) {}

func Test_messager_SendProtoMsg(t *testing.T) {}

func Test_messager_To(t *testing.T) {}

func Test_messager_send(t *testing.T) {}

func Test_newEmptyMessager(t *testing.T) {}

func Test_projectSvc_getBranchAndCommitIfMissing(t *testing.T) {}
