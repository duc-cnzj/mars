package repo

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/ent/project"

	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewEndpointRepo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo := NewEndpointRepo(mlog.NewForConfig(nil), data.NewMockData(m), NewMockProjectRepo(m))
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.(*endpointRepo).logger)
	assert.NotNil(t, repo.(*endpointRepo).data)
	assert.NotNil(t, repo.(*endpointRepo).projRepo)
}

func Test_endpointRepo_InNamespace_HappyPath(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	proj := NewMockProjectRepo(m)
	repo := NewEndpointRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		DB: db,
	}), proj)

	ns := createNamespace(db)
	createProject(db, ns.ID)
	p, _ := db.Project.Query().Select(
		project.FieldID,
		project.FieldName,
		project.FieldNamespaceID,
		project.FieldManifest,
	).First(context.TODO())

	proj.EXPECT().GetNodePortMappingByProjects(gomock.Any(), ns.Name, ToProject(p)).Return(EndpointMapping{"test": []*types.ServiceEndpoint{
		{
			Name:     "a1",
			Url:      "b1",
			PortName: "c1",
		},
	}})
	proj.EXPECT().GetIngressMappingByProjects(gomock.Any(), ns.Name, ToProject(p)).Return(EndpointMapping{"test": []*types.ServiceEndpoint{
		{
			Name:     "a2",
			Url:      "b2",
			PortName: "c2",
		},
	}})
	proj.EXPECT().GetLoadBalancerMappingByProjects(gomock.Any(), ns.Name, ToProject(p)).Return(EndpointMapping{"test": []*types.ServiceEndpoint{
		{
			Name:     "a3",
			Url:      "b3",
			PortName: "c4",
		},
	}})

	// Assuming namespace with ID 1 exists and has projects
	res, err := repo.InNamespace(context.TODO(), ns.ID)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Len(t, res, 3)
}

func Test_endpointRepo_InNamespace_NonExistentNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewEndpointRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		DB: db,
	}), NewMockProjectRepo(m))

	// Assuming namespace with ID 999 does not exist
	res, err := repo.InNamespace(context.TODO(), 999)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestInProject_HappyPath(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	proj := NewMockProjectRepo(m)
	repo := NewEndpointRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		DB: db,
	}), proj)

	ns := createNamespace(db)
	pr := createProject(db, ns.ID)

	proj.EXPECT().GetGatewayHTTPRouteMappingByProjects(gomock.Any(), gomock.Any(), gomock.Any()).Return(EndpointMapping{"test": []*types.ServiceEndpoint{
		{
			Name:     "ra1",
			Url:      "rb1",
			PortName: "rc1",
		},
	}})
	proj.EXPECT().GetNodePortMappingByProjects(gomock.Any(), gomock.Any(), gomock.Any()).Return(EndpointMapping{"test": []*types.ServiceEndpoint{
		{
			Name:     "a1",
			Url:      "b1",
			PortName: "c1",
		},
	}})
	proj.EXPECT().GetIngressMappingByProjects(gomock.Any(), gomock.Any(), gomock.Any()).Return(EndpointMapping{"test": []*types.ServiceEndpoint{
		{
			Name:     "a2",
			Url:      "b2",
			PortName: "c2",
		},
	}})
	proj.EXPECT().GetLoadBalancerMappingByProjects(gomock.Any(), gomock.Any(), gomock.Any()).Return(EndpointMapping{"test": []*types.ServiceEndpoint{
		{
			Name:     "a3",
			Url:      "b3",
			PortName: "c4",
		},
	}})

	// Assuming project with ID 1 exists
	res, err := repo.InProject(context.TODO(), pr.ID)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Len(t, res, 4)
}

func TestInProject_NonExistentProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewEndpointRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		DB: db,
	}), NewMockProjectRepo(m))

	// Assuming project with ID 999 does not exist
	res, err := repo.InProject(context.TODO(), 999)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}
