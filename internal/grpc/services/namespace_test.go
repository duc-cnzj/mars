package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/namespace"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/action"
	v12 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNamespaceSvc_All(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	db.Create(&models.Namespace{
		ID:               0,
		Name:             "ns1",
		ImagePullSecrets: "",
		Projects: []models.Project{
			{
				Name:         "deploy1",
				DeployStatus: 1,
			},
			{
				Name:         "deploy2",
				DeployStatus: 2,
			},
		},
	})
	db.Create(&models.Namespace{
		Name:             "ns2",
		ImagePullSecrets: "",
		Projects: []models.Project{
			{
				Name:         "deploy3",
				DeployStatus: 1,
			},
			{
				Name:         "deploy4",
				DeployStatus: 2,
			},
		},
	})
	all, _ := new(NamespaceSvc).All(context.TODO(), &namespace.AllRequest{})
	assert.Equal(t, "ns1", all.Items[0].Name)
	assert.Equal(t, "ns2", all.Items[1].Name)
	assert.Len(t, all.Items[0].Projects, 2)
}

func SetGormDB(m *gomock.Controller, app *mock.MockApplicationInterface) (*gorm.DB, func()) {
	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	return db, func() {
		s.Close()
	}
}

func TestNamespaceSvc_Create(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	db, closeFn := SetGormDB(m, app)
	defer closeFn()
	db.AutoMigrate(&models.Namespace{})
	app.EXPECT().Config().Return(&config.Config{NsPrefix: "dev-", ImagePullSecrets: []config.DockerAuth{
		{
			Username: "duc",
			Password: "pwd",
			Email:    "email",
			Server:   "server",
		},
	}}).AnyTimes()

	db.Create(&models.Namespace{
		Name: "dev-aaa",
	})
	_, err := new(NamespaceSvc).Create(adminCtx(), &namespace.CreateRequest{
		Namespace:      "aaa",
		IgnoreIfExists: false,
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.AlreadyExists, fromError.Code())
	res, err := new(NamespaceSvc).Create(adminCtx(), &namespace.CreateRequest{
		Namespace:      "aaa",
		IgnoreIfExists: true,
	})
	assert.Nil(t, err)
	assert.True(t, res.Exists)
	assert.Equal(t, "dev-aaa", res.Namespace.Name)
	clientset := fake.NewSimpleClientset(&v12.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "dev-terminating-ns",
		},
		Spec: v12.NamespaceSpec{},
		Status: v12.NamespaceStatus{
			Phase: v12.NamespaceTerminating,
		},
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: clientset}).AnyTimes()

	d := assertAuditLogFired(m, app)
	d.EXPECT().Dispatch(events.EventNamespaceCreated, gomock.Any())
	res, err = new(NamespaceSvc).Create(adminCtx(), &namespace.CreateRequest{
		Namespace: "bbb",
	})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	list, _ := clientset.CoreV1().Secrets("dev-bbb").List(context.TODO(), v1.ListOptions{})
	assert.Len(t, list.Items, 1)
	assert.Equal(t, "mars-", list.Items[0].GenerateName)

	res, err = new(NamespaceSvc).Create(adminCtx(), &namespace.CreateRequest{
		Namespace: "terminating-ns",
	})
	s, _ := status.FromError(err)
	assert.Equal(t, "该名称空间正在删除中", s.Message())
	assert.Equal(t, codes.AlreadyExists, s.Code())
}

func TestNamespaceSvc_Delete(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	db, closeFn := SetGormDB(m, app)
	defer closeFn()
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	ns := &models.Namespace{
		Name:             "dev-aaa",
		ImagePullSecrets: "mars-xxx",
		Projects: []models.Project{
			{
				Name: "duc",
			},
			{
				Name: "abc",
			},
		},
	}
	db.Create(ns)
	d := assertAuditLogFired(m, app)
	clientset := fake.NewSimpleClientset(
		&v12.Secret{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "dev-aaa",
				Name:      "mars-xxx",
			},
		},
		&v12.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: "dev-aaa",
			},
		},
	)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: clientset}).AnyTimes()

	d.EXPECT().Dispatch(events.EventNamespaceDeleted, gomock.Any()).Times(1)
	_, err := (&NamespaceSvc{UninstallReleaseFunc: func(releaseName, namespace string, log action.DebugLog) error {
		return nil
	}}).Delete(adminCtx(), &namespace.DeleteRequest{
		NamespaceId: int64(ns.ID),
	})
	assert.Nil(t, err)
	procjCount := int64(0)
	nsCount := int64(0)
	db.Model(&models.Project{}).Count(&procjCount)
	db.Model(&models.Namespace{}).Count(&nsCount)
	assert.Equal(t, int64(0), procjCount)
	assert.Equal(t, int64(0), nsCount)
	_, err = clientset.CoreV1().Namespaces().Get(context.TODO(), "dev-aaa", v1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err))
}

func TestNamespaceSvc_IsExists(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	db, closeFn := SetGormDB(m, app)
	defer closeFn()
	app.EXPECT().Config().Return(&config.Config{NsPrefix: "dev-"}).AnyTimes()
	_, err := new(NamespaceSvc).IsExists(context.TODO(), &namespace.IsExistsRequest{
		Name: "dev-not-exists",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, fromError.Code())
	db.AutoMigrate(&models.Namespace{})

	db.Create(&models.Namespace{
		Name: "dev-aaa",
	})
	exists, _ := new(NamespaceSvc).IsExists(context.TODO(), &namespace.IsExistsRequest{
		Name: "aaa",
	})
	assert.True(t, exists.Exists)
	exists, _ = new(NamespaceSvc).IsExists(context.TODO(), &namespace.IsExistsRequest{
		Name: "dev-aaa",
	})
	assert.True(t, exists.Exists)
	exists, err = new(NamespaceSvc).IsExists(context.TODO(), &namespace.IsExistsRequest{
		Name: "dev-not-exists",
	})
	assert.False(t, exists.Exists)
	assert.Nil(t, err)
}

func TestNamespaceSvc_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	db, closeFn := SetGormDB(m, app)
	defer closeFn()
	_, err := new(NamespaceSvc).Show(context.TODO(), &namespace.ShowRequest{
		NamespaceId: 678,
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, fromError.Code())
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	app.EXPECT().Config().Return(&config.Config{NsPrefix: "dev-"}).AnyTimes()

	ns := &models.Namespace{
		Name: "dev-aaa",
		Projects: []models.Project{
			{
				Name: "duc",
			},
		},
	}
	db.Create(ns)
	show, _ := new(NamespaceSvc).Show(context.TODO(), &namespace.ShowRequest{
		NamespaceId: int64(ns.ID),
	})
	assert.Equal(t, "dev-aaa", show.Namespace.Name)
	assert.Equal(t, "duc", show.Namespace.Projects[0].Name)
	_, err = new(NamespaceSvc).Show(context.TODO(), &namespace.ShowRequest{
		NamespaceId: 678,
	})
	fromError, _ = status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())
}
