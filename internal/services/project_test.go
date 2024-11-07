package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v5/project"
	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/socket"
	"github.com/duc-cnzj/mars/v5/internal/util/pagination"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
		mlog.NewForConfig(nil),
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
		mlog.NewForConfig(nil),
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
		mlog.NewForConfig(nil),
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
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		k8sRepo,
		repo.NewMockEventRepo(m),
		mlog.NewForConfig(nil),
		repo.NewMockHelmerRepo(m),
		nsRepo,
	)
	nsRepo.EXPECT().CanAccess(gomock.Any(), 1, MustGetUser(newAdminUserCtx())).Return(true)
	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Project{
		Namespace:   &repo.Namespace{ID: 1},
		NamespaceID: 1,
	}, nil)
	res, err := svc.Show(newAdminUserCtx(), &project.ShowRequest{
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
		mlog.NewForConfig(nil),
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

func TestProjectSvc_Show_Failure2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	namespaceRepo := repo.NewMockNamespaceRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		repo.NewMockK8sRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewForConfig(nil),
		repo.NewMockHelmerRepo(m),
		namespaceRepo,
	)

	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Project{NamespaceID: 1}, nil)
	namespaceRepo.EXPECT().CanAccess(gomock.Any(), 1, gomock.Any()).Return(false)
	_, err := svc.Show(context.TODO(), &project.ShowRequest{
		Id: 1,
	})
	assert.ErrorIs(t, err, ErrorPermissionDenied)
}

func Test_projectSvc_Delete(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	helmerRepo := repo.NewMockHelmerRepo(m)
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		k8sRepo,
		eventRepo,
		mlog.NewForConfig(nil),
		helmerRepo,
		nsRepo,
	)
	nsRepo.EXPECT().CanAccess(gomock.Any(), 1, MustGetUser(newAdminUserCtx())).Return(true)
	helmerRepo.EXPECT().Uninstall("app", "ns", gomock.Any()).Return(errors.New("x"))
	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Project{
		ID:          2,
		Name:        "app",
		NamespaceID: 1,
		Namespace: &repo.Namespace{
			Name: "ns",
		},
	}, nil)
	req := &project.DeleteRequest{
		Id: 1,
	}
	eventRepo.EXPECT().AuditLogWithRequest(types.EventActionType_Delete, MustGetUser(newAdminUserCtx()).Name, gomock.Any(), req)
	eventRepo.EXPECT().Dispatch(repo.EventProjectDeleted, &repo.ProjectDeletedPayload{
		NamespaceID: 1,
		ProjectID:   2,
	})
	projectRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)

	response, err := svc.Delete(newAdminUserCtx(), req)
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
		mlog.NewForConfig(nil),
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

func Test_projectSvc_Delete_Fail2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	helmerRepo := repo.NewMockHelmerRepo(m)
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		k8sRepo,
		eventRepo,
		mlog.NewForConfig(nil),
		helmerRepo,
		nsRepo,
	)
	nsRepo.EXPECT().CanAccess(gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Project{}, nil)
	projectRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errors.New("x"))
	response, err := svc.Delete(newAdminUserCtx(), &project.DeleteRequest{
		Id: 1,
	})
	assert.NotNil(t, err)
	assert.Nil(t, response)
}

func Test_projectSvc_Delete_Fail3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := repo.NewMockProjectRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	helmerRepo := repo.NewMockHelmerRepo(m)
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		projectRepo,
		repo.NewMockGitRepo(m),
		k8sRepo,
		eventRepo,
		mlog.NewForConfig(nil),
		helmerRepo,
		nsRepo,
	)
	nsRepo.EXPECT().CanAccess(gomock.Any(), gomock.Any(), gomock.Any()).Return(false)
	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Project{}, nil)
	_, err := svc.Delete(newAdminUserCtx(), &project.DeleteRequest{
		Id: 1,
	})
	assert.ErrorIs(t, err, ErrorPermissionDenied)
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
		mlog.NewForConfig(nil),
		helmerRepo,
		repo.NewMockNamespaceRepo(m),
	)
	projectRepo.EXPECT().Version(gomock.Any(), 1).Return(100, nil)
	version, err := svc.Version(context.TODO(), &project.VersionRequest{Id: 1})
	assert.Nil(t, err)
	assert.NotNil(t, version)
	assert.Equal(t, int32(100), version.Version)

	projectRepo.EXPECT().Version(gomock.Any(), 1).Return(0, errors.New("x"))
	_, err = svc.Version(context.TODO(), &project.VersionRequest{Id: 1})
	assert.Error(t, err)
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
		mlog.NewForConfig(nil),
		helmerRepo,
		repo.NewMockNamespaceRepo(m),
	)
	projectRepo.EXPECT().GetAllActiveContainers(gomock.Any(), 1).Return([]*types.StateContainer{}, nil)
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
		mlog.NewForConfig(nil),
		helmerRepo,
		repo.NewMockNamespaceRepo(m),
	)
	projectRepo.EXPECT().GetAllActiveContainers(gomock.Any(), 1).Return(nil, errors.New("x"))
	containers, err := svc.AllContainers(context.TODO(), &project.AllContainersRequest{
		Id: 1,
	})
	assert.NotNil(t, err)
	assert.Nil(t, containers)
}

func TestProjectSvc_WebApply_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	projectRepo := repo.NewMockProjectRepo(m)
	repoRepo := repo.NewMockRepoRepo(m)
	jobManager := socket.NewMockJobManager(m)
	gitRepo := repo.NewMockGitRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	logger := mlog.NewForConfig(nil)
	helmerRepo := repo.NewMockHelmerRepo(m)
	nsRepo := repo.NewMockNamespaceRepo(m)
	plMgr := application.NewMockPluginManger(m)

	svc := NewProjectSvc(
		repoRepo,
		plMgr,
		jobManager,
		projectRepo,
		gitRepo,
		repo.NewMockK8sRepo(m),
		eventRepo,
		logger,
		helmerRepo,
		nsRepo,
	)

	nsRepo.EXPECT().CanAccess(gomock.Any(), 1, MustGetUser(newAdminUserCtx())).Return(true)
	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&repo.Repo{Name: "test", NeedGitRepo: true, GitProjectID: 100}, nil)
	gitRepo.EXPECT().ListCommits(gomock.Any(), 100, "dev").Return([]*repo.Commit{{ID: "commit-id"}}, nil)

	job := socket.NewMockJob(m)
	jobManager.EXPECT().NewJob(gomock.Any()).Return(job)

	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(nil)
	job.EXPECT().Manifests().Return([]string{"manifests"})
	job.EXPECT().IsNotDryRun().Return(true)
	job.EXPECT().Project().Return(&repo.Project{ID: 1})

	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Project{}, nil)

	_, err := svc.WebApply(newAdminUserCtx(), &project.WebApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		GitBranch:   "dev",
	})

	assert.Nil(t, err)
}

func TestProjectSvc_WebApply_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	projectRepo := repo.NewMockProjectRepo(m)
	repoRepo := repo.NewMockRepoRepo(m)
	jobManager := socket.NewMockJobManager(m)
	gitRepo := repo.NewMockGitRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	logger := mlog.NewForConfig(nil)
	helmerRepo := repo.NewMockHelmerRepo(m)
	nsRepo := repo.NewMockNamespaceRepo(m)
	plMgr := application.NewMockPluginManger(m)

	svc := NewProjectSvc(
		repoRepo,
		plMgr,
		jobManager,
		projectRepo,
		gitRepo,
		repo.NewMockK8sRepo(m),
		eventRepo,
		logger,
		helmerRepo,
		nsRepo,
	)

	nsRepo.EXPECT().CanAccess(gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
	repoRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	_, err := svc.WebApply(context.TODO(), &project.WebApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		Name:        "test",
	})

	assert.NotNil(t, err)
	assert.Equal(t, "error", err.Error())
}

func TestProjectSvc_Apply_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	projectRepo := repo.NewMockProjectRepo(m)
	repoRepo := repo.NewMockRepoRepo(m)
	jobManager := socket.NewMockJobManager(m)
	gitRepo := repo.NewMockGitRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	logger := mlog.NewForConfig(nil)
	helmerRepo := repo.NewMockHelmerRepo(m)
	nsRepo := repo.NewMockNamespaceRepo(m)
	plMgr := application.NewMockPluginManger(m)

	svc := NewProjectSvc(
		repoRepo,
		plMgr,
		jobManager,
		projectRepo,
		gitRepo,
		repo.NewMockK8sRepo(m),
		eventRepo,
		logger,
		helmerRepo,
		nsRepo,
	)

	nsRepo.EXPECT().CanAccess(gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
	repoRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&repo.Repo{Name: "test", NeedGitRepo: true}, nil)
	gitRepo.EXPECT().ListCommits(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*repo.Commit{{ID: "commit-id"}}, nil)

	job := socket.NewMockJob(m)
	jobManager.EXPECT().NewJob(gomock.Any()).Return(job)

	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(nil)

	mockServer := &mockProjectApplyServer{
		Req: &project.ApplyRequest{
			RepoId:      1,
			NamespaceId: 1,
			Name:        "test",
		},
	}

	applyRequest := &project.ApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		Name:        "test",
	}

	err := svc.Apply(applyRequest, mockServer)

	assert.Nil(t, err)
}

func TestProjectSvc_Apply_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	projectRepo := repo.NewMockProjectRepo(m)
	repoRepo := repo.NewMockRepoRepo(m)
	jobManager := socket.NewMockJobManager(m)
	gitRepo := repo.NewMockGitRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	logger := mlog.NewForConfig(nil)
	helmerRepo := repo.NewMockHelmerRepo(m)
	nsRepo := repo.NewMockNamespaceRepo(m)
	plMgr := application.NewMockPluginManger(m)

	svc := NewProjectSvc(
		repoRepo,
		plMgr,
		jobManager,
		projectRepo,
		gitRepo,
		repo.NewMockK8sRepo(m),
		eventRepo,
		logger,
		helmerRepo,
		nsRepo,
	)

	nsRepo.EXPECT().CanAccess(gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
	repoRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	mockServer := &mockProjectApplyServer{
		Req: &project.ApplyRequest{
			RepoId:      1,
			NamespaceId: 1,
			Name:        "test",
		},
	}

	applyRequest := &project.ApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		Name:        "test",
	}

	err := svc.Apply(applyRequest, mockServer)

	assert.NotNil(t, err)
	assert.Equal(t, "error", err.Error())
}

func TestProjectSvc_Apply_Failure2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	projectRepo := repo.NewMockProjectRepo(m)
	repoRepo := repo.NewMockRepoRepo(m)
	jobManager := socket.NewMockJobManager(m)
	gitRepo := repo.NewMockGitRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	logger := mlog.NewForConfig(nil)
	helmerRepo := repo.NewMockHelmerRepo(m)
	nsRepo := repo.NewMockNamespaceRepo(m)
	plMgr := application.NewMockPluginManger(m)

	svc := NewProjectSvc(
		repoRepo,
		plMgr,
		jobManager,
		projectRepo,
		gitRepo,
		repo.NewMockK8sRepo(m),
		eventRepo,
		logger,
		helmerRepo,
		nsRepo,
	)

	nsRepo.EXPECT().CanAccess(gomock.Any(), 1, gomock.Any()).Return(false)

	mockServer := &mockProjectApplyServer{
		Req: &project.ApplyRequest{
			RepoId:      1,
			NamespaceId: 1,
			Name:        "test",
		},
	}

	applyRequest := &project.ApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		Name:        "test",
	}

	err := svc.Apply(applyRequest, mockServer)

	assert.ErrorIs(t, err, ErrorPermissionDenied)
}

type mockProjectApplyServer struct {
	grpc.ServerStream
	Req *project.ApplyRequest
}

func (x *mockProjectApplyServer) Send(m *project.ApplyResponse) error {
	// Implement this method based on your testing needs.
	return nil
}

func (x *mockProjectApplyServer) SetHeader(md metadata.MD) error {
	// Implement this method based on your testing needs.
	return nil
}

func (x *mockProjectApplyServer) SendHeader(md metadata.MD) error {
	// Implement this method based on your testing needs.
	return nil
}

func (x *mockProjectApplyServer) SetTrailer(md metadata.MD) {
	// Implement this method based on your testing needs.
}

func (x *mockProjectApplyServer) Context() context.Context {
	// Implement this method based on your testing needs.
	return context.TODO()
}

func (x *mockProjectApplyServer) SendMsg(m any) error {
	// Implement this method based on your testing needs.
	return nil
}

func (x *mockProjectApplyServer) RecvMsg(m any) error {
	return nil
}

func TestEmptyMessager_Current(t *testing.T) {
	m := newEmptyMessager()
	current := m.Current()
	assert.Equal(t, int64(0), current)
}

func TestEmptyMessager_Add(t *testing.T) {
	m := newEmptyMessager()
	m.Add()
	current := m.Current()
	assert.Equal(t, int64(0), current)
}

func TestEmptyMessager_To(t *testing.T) {
	m := newEmptyMessager()
	m.To(50)
	current := m.Current()
	assert.Equal(t, int64(0), current)
}

func TestEmptyMessager_SendEndError(t *testing.T) {
	m := newEmptyMessager()
	m.SendEndError(errors.New("test error"))
	// As the emptyMessager does not handle errors, there's no assertion to be made here.
}

func TestEmptyMessager_SendMsg(t *testing.T) {
	m := newEmptyMessager()
	m.SendMsg("test message")
	// As the emptyMessager does not handle messages, there's no assertion to be made here.
}

func TestEmptyMessager_SendProtoMsg(t *testing.T) {
	m := newEmptyMessager()
	m.SendProtoMsg(nil)
	// As the emptyMessager does not handle messages, there's no assertion to be made here.
}

func TestEmptyMessager_SendProcessPercent(t *testing.T) {
	m := newEmptyMessager()
	m.SendProcessPercent(50)
	// As the emptyMessager does not handle process percent, there's no assertion to be made here.
}

func TestEmptyMessager_SendMsgWithContainerLog(t *testing.T) {
	m := newEmptyMessager()
	m.SendMsgWithContainerLog("test message", []*websocket.Container{})
	// As the emptyMessager does not handle container logs, there's no assertion to be made here.
}

func TestMessager_Current(t *testing.T) {
	m := NewMessager(true, "slug", websocket.Type_ApplyProject, nil)
	current := m.Current()
	assert.Equal(t, int64(0), current)
}

type mockApplyServer struct {
	project.Project_ApplyServer

	response *project.ApplyResponse
}

func (m *mockApplyServer) Send(response *project.ApplyResponse) error {
	m.response = response
	return nil
}

func TestMessager_Add(t *testing.T) {
	server := &mockApplyServer{}
	m := NewMessager(true, "slug", websocket.Type_ApplyProject, server)
	m.Add()
	current := m.Current()
	assert.Equal(t, websocket.Type_ProcessPercent, server.response.Metadata.Type)
	assert.Equal(t, int64(1), current)
}

func TestMessager_To(t *testing.T) {
	server := &mockApplyServer{}
	m := NewMessager(true, "slug", websocket.Type_ApplyProject, server)
	m.To(50)
	current := m.Current()
	assert.Equal(t, int32(50), server.response.Metadata.Percent)
	assert.Equal(t, int64(50), current)
}

func TestMessager_SendEndError(t *testing.T) {
	server := &mockApplyServer{}
	m := NewMessager(true, "slug", websocket.Type_ApplyProject, server)
	m.SendEndError(errors.New("test error"))
	assert.True(t, server.response.Metadata.End)
	assert.Equal(t, "test error", server.response.Metadata.Message)
}

func TestMessager_SendMsg(t *testing.T) {
	server := &mockApplyServer{}
	m := NewMessager(true, "slug", websocket.Type_ApplyProject, server)
	m.SendMsg("test message")
	assert.False(t, server.response.Metadata.End)
	assert.Equal(t, "test message", server.response.Metadata.Message)
}

type mockWsMessage struct {
	application.WebsocketMessage
}

func (m *mockWsMessage) GetMetadata() *websocket.Metadata {
	return &websocket.Metadata{
		Type: websocket.Type_ApplyProject,
	}
}

func TestMessager_SendProtoMsg(t *testing.T) {
	server := &mockApplyServer{}
	m := NewMessager(true, "slug", websocket.Type_ApplyProject, server)
	m.SendProtoMsg(&mockWsMessage{})
	assert.Equal(t, websocket.Type_ApplyProject, server.response.Metadata.Type)
}

func TestMessager_SendProcessPercent(t *testing.T) {
	server := &mockApplyServer{}
	m := NewMessager(true, "slug", websocket.Type_ApplyProject, server)
	m.SendProcessPercent(50)
	assert.Equal(t, websocket.Type_ProcessPercent, server.response.Metadata.Type)
	assert.Equal(t, int32(50), server.response.Metadata.Percent)
}

func TestMessager_SendMsgWithContainerLog(t *testing.T) {
	server := &mockApplyServer{}
	m := NewMessager(true, "slug", websocket.Type_ApplyProject, server)
	m.SendMsgWithContainerLog("test message", []*websocket.Container{})
	assert.False(t, server.response.Metadata.End)
	assert.Equal(t, "test message", server.response.Metadata.Message)
	assert.Equal(t, websocket.ResultType_LogWithContainers, server.response.Metadata.Result)
}

func TestMessager_SendDeployedResult(t *testing.T) {
	server := &mockApplyServer{}
	m := NewMessager(true, "slug", websocket.Type_ApplyProject, server)
	m.SendDeployedResult(websocket.ResultType_Success, "test message", &types.ProjectModel{})
	assert.True(t, server.response.Metadata.End)
	assert.Equal(t, "test message", server.response.Metadata.Message)
	assert.Equal(t, websocket.ResultType_Success, server.response.Metadata.Result)
}

func TestGetBranchAndCommitIfMissingReturnsDefaultBranchWhenBranchIsNotProvided(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	gitRepo := repo.NewMockGitRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		repo.NewMockProjectRepo(m),
		gitRepo,
		repo.NewMockK8sRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewForConfig(nil),
		repo.NewMockHelmerRepo(m),
		repo.NewMockNamespaceRepo(m),
	)

	gitRepo.EXPECT().ListCommits(gomock.Any(), 10, "default").Return([]*repo.Commit{{ID: "commit-id"}}, nil)

	branch, _, _ := svc.(*projectSvc).getBranchAndCommitIfMissing("", "", &repo.Repo{DefaultBranch: "default", GitProjectID: 10}, newEmptyMessager())
	assert.Equal(t, "default", branch)
}

func TestGetBranchAndCommitIfMissingReturnsLatestCommitWhenCommitIsNotProvided(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	gitRepo := repo.NewMockGitRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		repo.NewMockProjectRepo(m),
		gitRepo,
		repo.NewMockK8sRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewForConfig(nil),
		repo.NewMockHelmerRepo(m),
		repo.NewMockNamespaceRepo(m),
	)

	gitRepo.EXPECT().ListCommits(gomock.Any(), 10, "branch").Return([]*repo.Commit{{ID: "commit-id"}}, nil)

	_, commit, _ := svc.(*projectSvc).getBranchAndCommitIfMissing("branch", "", &repo.Repo{DefaultBranch: "default", GitProjectID: 10}, newEmptyMessager())
	assert.Equal(t, "commit-id", commit)
}

func TestGetBranchAndCommitIfMissingReturnsLatestCommitWhenCommitIsNotProvided2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	gitRepo := repo.NewMockGitRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		repo.NewMockProjectRepo(m),
		gitRepo,
		repo.NewMockK8sRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewForConfig(nil),
		repo.NewMockHelmerRepo(m),
		repo.NewMockNamespaceRepo(m),
	)

	gitRepo.EXPECT().ListCommits(gomock.Not(nil), 10, "branch").Return([]*repo.Commit{}, nil)

	_, _, err := svc.(*projectSvc).getBranchAndCommitIfMissing("branch", "", &repo.Repo{DefaultBranch: "default", GitProjectID: 10}, newEmptyMessager())
	assert.Equal(t, "没有可用的 commit", err.Error())
}

func TestGetBranchAndCommitIfMissingReturnsProvidedBranchAndCommit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	gitRepo := repo.NewMockGitRepo(m)
	svc := NewProjectSvc(
		repo.NewMockRepoRepo(m),
		application.NewMockPluginManger(m),
		socket.NewMockJobManager(m),
		repo.NewMockProjectRepo(m),
		gitRepo,
		repo.NewMockK8sRepo(m),
		repo.NewMockEventRepo(m),
		mlog.NewForConfig(nil),
		repo.NewMockHelmerRepo(m),
		repo.NewMockNamespaceRepo(m),
	)

	branch, commit, _ := svc.(*projectSvc).getBranchAndCommitIfMissing("branch", "commit", &repo.Repo{DefaultBranch: "default"}, newEmptyMessager())
	assert.Equal(t, "branch", branch)
	assert.Equal(t, "commit", commit)
}

func Test_projectSvc_MemoryCpuAndEndpoints(t *testing.T) {
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
		mlog.NewForConfig(nil),
		helmerRepo,
		repo.NewMockNamespaceRepo(m),
	)

	projModel := &repo.Project{
		ID:          2,
		Name:        "app",
		NamespaceID: 1,
		Namespace: &repo.Namespace{
			Name: "ns",
		},
		Version: 100,
	}
	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(projModel, nil)
	k8sRepo.EXPECT().GetAllPodMetrics(gomock.Any(), projModel)
	k8sRepo.EXPECT().GetCpuAndMemory(gomock.Any(), gomock.Any()).Return("1", "2Gi")
	projectRepo.EXPECT().GetProjectEndpointsInNamespace(gomock.Any(), projModel.Namespace.Name, projModel.ID).Return([]*types.ServiceEndpoint{
		{Name: "app"},
		{Name: "app1"},
		{Name: "app2"},
	}, nil)
	endpoints, err := svc.MemoryCpuAndEndpoints(context.TODO(), &project.MemoryCpuAndEndpointsRequest{Id: 1})
	assert.Nil(t, err)
	assert.Equal(t, "1", endpoints.Cpu)
	assert.Equal(t, "2Gi", endpoints.Memory)
	assert.Equal(t, 3, len(endpoints.Urls))
}

func Test_projectSvc_MemoryCpuAndEndpoints_Fail(t *testing.T) {
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
		mlog.NewForConfig(nil),
		helmerRepo,
		repo.NewMockNamespaceRepo(m),
	)

	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	_, err := svc.MemoryCpuAndEndpoints(context.TODO(), &project.MemoryCpuAndEndpointsRequest{Id: 1})
	assert.Error(t, err)
}

func Test_projectSvc_MemoryCpuAndEndpoints_fail2(t *testing.T) {
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
		mlog.NewForConfig(nil),
		helmerRepo,
		repo.NewMockNamespaceRepo(m),
	)

	projModel := &repo.Project{
		ID:          2,
		Name:        "app",
		NamespaceID: 1,
		Namespace: &repo.Namespace{
			Name: "ns",
		},
		Version: 100,
	}
	projectRepo.EXPECT().Show(gomock.Any(), 1).Return(projModel, nil)
	k8sRepo.EXPECT().GetAllPodMetrics(gomock.Any(), projModel)
	k8sRepo.EXPECT().GetCpuAndMemory(gomock.Any(), gomock.Any()).Return("1", "2Gi")
	projectRepo.EXPECT().GetProjectEndpointsInNamespace(gomock.Any(), projModel.Namespace.Name, projModel.ID).Return(nil, errors.New("x"))
	endpoints, err := svc.MemoryCpuAndEndpoints(context.TODO(), &project.MemoryCpuAndEndpointsRequest{Id: 1})
	assert.Error(t, err)
	assert.Nil(t, endpoints)
}
