package services

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/endpoint"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_endpointSvc_InNamespace_ShowError(t *testing.T) {
	svc, mocks := newEndpointSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("boom"))
	_, err := svc.InNamespace(newAdminUserCtx(), &endpoint.InNamespaceRequest{NamespaceId: 1})
	assert.Error(t, err)
	assert.Equal(t, "boom", err.Error())
}

func TestEndpointSvc_InProject_ShowError(t *testing.T) {
	svc, mocks := newEndpointSvcWithMocks(t)
	projRepo := mocks.projRepo
	projRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("boom"))
	_, err := svc.InProject(newAdminUserCtx(), &endpoint.InProjectRequest{ProjectId: 1})
	assert.Error(t, err)
	assert.Equal(t, "boom", err.Error())
}

func TestEndpointSvc_InProject_NamespaceError(t *testing.T) {
	svc, mocks := newEndpointSvcWithMocks(t)
	projRepo := mocks.projRepo
	nsRepo := mocks.nsRepo
	projRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("boom"))
	_, err := svc.InProject(newAdminUserCtx(), &endpoint.InProjectRequest{ProjectId: 1})
	assert.Error(t, err)
	assert.Equal(t, "boom", err.Error())
}

func TestEndpointSvc_InProject_PermissionDenied(t *testing.T) {
	svc, mocks := newEndpointSvcWithMocks(t)
	projRepo := mocks.projRepo
	nsRepo := mocks.nsRepo
	projRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: true}, nil)
	_, err := svc.InProject(newOtherUserCtx(), &endpoint.InProjectRequest{ProjectId: 1})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestNewEndpointSvc(t *testing.T) {
	svc, _ := newEndpointSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.epBiz)
}

func Test_endpointSvc_InNamespace(t *testing.T) {
	svc, mocks := newEndpointSvcWithMocks(t)
	endpointBiz := mocks.epBiz
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)
	endpointBiz.EXPECT().InNamespace(gomock.Any(), 1).Return(nil, nil)
	namespace, err := svc.InNamespace(newAdminUserCtx(), &endpoint.InNamespaceRequest{
		NamespaceId: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, namespace)
}
func Test_endpointSvc_InNamespace_PermissionDenied(t *testing.T) {
	svc, mocks := newEndpointSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: true}, nil)
	_, err := svc.InNamespace(newOtherUserCtx(), &endpoint.InNamespaceRequest{
		NamespaceId: 1,
	})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func Test_endpointSvc_InNamespace_Fail(t *testing.T) {
	svc, mocks := newEndpointSvcWithMocks(t)
	endpointBiz := mocks.epBiz
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)
	endpointBiz.EXPECT().InNamespace(gomock.Any(), 1).Return(nil, errors.New("x"))
	_, err := svc.InNamespace(newAdminUserCtx(), &endpoint.InNamespaceRequest{
		NamespaceId: 1,
	})
	assert.Error(t, err)
}

func TestEndpointSvc_InProject_Success(t *testing.T) {
	svc, mocks := newEndpointSvcWithMocks(t)
	epBiz := mocks.epBiz
	projRepo := mocks.projRepo
	nsRepo := mocks.nsRepo

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)
	epBiz.EXPECT().InProject(gomock.Any(), 1).Return([]*types.ServiceEndpoint{}, nil)

	_, err := svc.InProject(newAdminUserCtx(), &endpoint.InProjectRequest{
		ProjectId: 1,
	})
	assert.NoError(t, err)
}

func TestEndpointSvc_InProject_Failure(t *testing.T) {
	svc, mocks := newEndpointSvcWithMocks(t)
	epBiz := mocks.epBiz
	projRepo := mocks.projRepo
	nsRepo := mocks.nsRepo

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)
	epBiz.EXPECT().InProject(gomock.Any(), 1).Return(nil, errors.New("error"))

	_, err := svc.InProject(newAdminUserCtx(), &endpoint.InProjectRequest{
		ProjectId: 1,
	})
	assert.Error(t, err)
}

// endpointSvcMocks 聚合 endpointSvc 的全部下游 mock，由 newEndpointSvcWithMocks 统一构造。
type endpointSvcMocks struct {
	ctrl     *gomock.Controller
	epBiz    *biz.MockEndpointBiz
	projRepo *data.MockProjectRepo
	nsRepo   *data.MockNamespaceRepo
}

func newEndpointSvcWithMocks(t *testing.T) (*endpointSvc, *endpointSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &endpointSvcMocks{
		ctrl:     ctrl,
		epBiz:    biz.NewMockEndpointBiz(ctrl),
		projRepo: data.NewMockProjectRepo(ctrl),
		nsRepo:   data.NewMockNamespaceRepo(ctrl),
	}
	logger := mlog.NewForConfig(nil)
	s, ok := NewEndpointSvc(EndpointSvcDeps{
		Logger:    logger,
		EpBiz:     mocks.epBiz,
		AccessBiz: biz.NewAccessBiz(biz.NewNamespaceBiz(logger, mocks.nsRepo, nil, nil, nil), biz.NewProjectBiz(logger, mocks.projRepo, nil)),
	}).(*endpointSvc)
	if !ok {
		panic("NewEndpointSvc returned unexpected type")
	}
	return s, mocks
}
