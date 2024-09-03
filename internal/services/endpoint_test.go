package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v5/endpoint"
	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewEndpointSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewEndpointSvc(mlog.NewForConfig(nil), repo.NewMockEndpointRepo(m))
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*endpointSvc).epRepo)
}

func Test_endpointSvc_InNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	endpointRepo := repo.NewMockEndpointRepo(m)
	svc := NewEndpointSvc(mlog.NewForConfig(nil), endpointRepo)
	endpointRepo.EXPECT().InNamespace(gomock.Any(), 1).Return(nil, nil)
	namespace, err := svc.InNamespace(context.TODO(), &endpoint.InNamespaceRequest{
		NamespaceId: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, namespace)
}
func Test_endpointSvc_InNamespace_Fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	endpointRepo := repo.NewMockEndpointRepo(m)
	svc := NewEndpointSvc(mlog.NewForConfig(nil), endpointRepo)
	endpointRepo.EXPECT().InNamespace(gomock.Any(), 1).Return(nil, errors.New("x"))
	_, err := svc.InNamespace(context.TODO(), &endpoint.InNamespaceRequest{
		NamespaceId: 1,
	})
	assert.Error(t, err)
}

func TestEndpointSvc_InProject_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	epRepo := repo.NewMockEndpointRepo(m)
	svc := NewEndpointSvc(mlog.NewForConfig(nil), epRepo)

	epRepo.EXPECT().InProject(gomock.Any(), gomock.Any()).Return([]*types.ServiceEndpoint{}, nil)

	_, err := svc.InProject(context.Background(), &endpoint.InProjectRequest{
		ProjectId: 1,
	})
	assert.NoError(t, err)
}

func TestEndpointSvc_InProject_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	epRepo := repo.NewMockEndpointRepo(m)
	svc := NewEndpointSvc(mlog.NewForConfig(nil), epRepo)

	epRepo.EXPECT().InProject(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	_, err := svc.InProject(context.Background(), &endpoint.InProjectRequest{
		ProjectId: 1,
	})
	assert.Error(t, err)
}
