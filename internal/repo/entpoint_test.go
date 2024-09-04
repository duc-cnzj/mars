package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewEndpointRepo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo := NewEndpointRepo(mlog.NewForConfig(nil), data.NewMockData(m), NewMockProjectRepo(m), NewMockNamespaceRepo(m))
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.(*endpointRepo).logger)
	assert.NotNil(t, repo.(*endpointRepo).data)
	assert.NotNil(t, repo.(*endpointRepo).projRepo)
}

func Test_endpointRepo_InNamespace_HappyPath(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	proj := NewMockProjectRepo(m)
	mockNamespaceRepo := NewMockNamespaceRepo(m)
	repo := NewEndpointRepo(mlog.NewForConfig(nil), nil, proj, mockNamespaceRepo)

	mockNamespaceRepo.EXPECT().Show(gomock.Any(), 1).Return(&Namespace{
		Name: "ns",
		Projects: []*Project{
			{ID: 1},
		},
	}, nil)

	proj.EXPECT().GetProjectEndpointsInNamespace(gomock.Any(), "ns", 1).Return([]*types.ServiceEndpoint{
		{
			Name:     "a1",
			Url:      "b1",
			PortName: "c1",
		},
		{
			Name:     "a2",
			Url:      "b2",
			PortName: "c2",
		},
		{
			Name:     "a3",
			Url:      "b3",
			PortName: "c4",
		},
	}, nil)

	// Assuming namespace with ID 1 exists and has projects
	res, err := repo.InNamespace(context.TODO(), 1)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Len(t, res, 3)
}

func Test_endpointRepo_InNamespace_NonExistentNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	proj := NewMockProjectRepo(m)
	mockNamespaceRepo := NewMockNamespaceRepo(m)
	repo := NewEndpointRepo(mlog.NewForConfig(nil), nil, proj, mockNamespaceRepo)
	mockNamespaceRepo.EXPECT().Show(gomock.Any(), 999).Return(nil, errors.New("x"))

	// Assuming namespace with ID 999 does not exist
	res, err := repo.InNamespace(context.TODO(), 999)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestInProject_HappyPath(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	proj := NewMockProjectRepo(m)
	repo := NewEndpointRepo(mlog.NewForConfig(nil), nil, proj, NewMockNamespaceRepo(m))

	proj.EXPECT().Show(gomock.Any(), 1).Return(&Project{
		Namespace: &Namespace{Name: "ns"},
		ID:        1,
	}, nil)
	proj.EXPECT().GetProjectEndpointsInNamespace(gomock.Any(), "ns", 1).
		Return([]*types.ServiceEndpoint{
			{
				Name:     "ra1",
				Url:      "rb1",
				PortName: "rc1",
			},
			{
				Name:     "a1",
				Url:      "b1",
				PortName: "c1",
			},
			{
				Name:     "a2",
				Url:      "b2",
				PortName: "c2",
			},
			{
				Name:     "a3",
				Url:      "b3",
				PortName: "c4",
			},
		}, nil)

	// Assuming project with ID 1 exists
	res, err := repo.InProject(context.TODO(), 1)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Len(t, res, 4)
}

func TestInProject_NonExistentProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	proj := NewMockProjectRepo(m)
	repo := NewEndpointRepo(mlog.NewForConfig(nil), nil, proj, NewMockNamespaceRepo(m))

	proj.EXPECT().Show(gomock.Any(), 999).Return(nil, errors.New("x"))
	// Assuming project with ID 999 does not exist
	res, err := repo.InProject(context.TODO(), 999)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}
