package services

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	errors2 "k8s.io/apimachinery/pkg/api/errors"

	"github.com/duc-cnzj/mars/api/v4/namespace"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewNamespaceSvc_Creation(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	helmer := repo.NewMockHelmerRepo(m)
	nsRepo := repo.NewMockNamespaceRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	logger := mlog.NewLogger(nil)
	eventRepo := repo.NewMockEventRepo(m)

	svc := NewNamespaceSvc(helmer, nsRepo, k8sRepo, logger, eventRepo)

	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*namespaceSvc).helmer)
	assert.NotNil(t, svc.(*namespaceSvc).nsRepo)
	assert.NotNil(t, svc.(*namespaceSvc).k8sRepo)
	assert.NotNil(t, svc.(*namespaceSvc).logger)
	assert.NotNil(t, svc.(*namespaceSvc).eventRepo)
}

func TestNamespaceSvc_All_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)

	nsRepo.EXPECT().All(gomock.Any(), &repo.AllNamespaceInput{
		Favorite: false,
		Email:    adminEmail,
	}).Return([]*repo.Namespace{
		{
			ID:   1,
			Name: "namespace1",
			Favorites: []*repo.Favorite{
				{Email: adminEmail},
			},
		},
		{
			ID:   2,
			Name: "namespace2",
		},
	}, nil)

	res, err := svc.All(newAdminUserCtx(), &namespace.AllRequest{
		Favorite: false,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 2, len(res.Items))
	assert.Equal(t, int32(1), res.Items[0].Id)
	assert.Equal(t, "namespace1", res.Items[0].Name)
	assert.Equal(t, int32(2), res.Items[1].Id)
	assert.Equal(t, "namespace2", res.Items[1].Name)
	assert.True(t, res.Items[0].Favorite)
}

func TestNamespaceSvc_All_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)

	nsRepo.EXPECT().All(gomock.Any(), &repo.AllNamespaceInput{
		Favorite: false,
		Email:    adminEmail,
	}).Return(nil, errors.New("error"))

	res, err := svc.All(newAdminUserCtx(), &namespace.AllRequest{
		Favorite: false,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Create_NamespaceTerminating(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewNamespaceSvc(repo.NewMockHelmerRepo(m), nsRepo, k8sRepo, mlog.NewLogger(nil), repo.NewMockEventRepo(m))

	nsRepo.EXPECT().GetMarsNamespace("test").Return("test")
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(nil, &ent.NotFoundError{})
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "test").Return(nil, &errors2.StatusError{
		ErrStatus: metav1.Status{
			Reason: metav1.StatusReasonAlreadyExists,
		},
	})
	k8sRepo.EXPECT().GetNamespace(gomock.Any(), "test").Return(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceTerminating,
		},
	}, nil)

	res, err := svc.Create(context.TODO(), &namespace.CreateRequest{
		Namespace:      "test",
		IgnoreIfExists: true,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
	t.Log(err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
	assert.Equal(t, "该名称空间正在删除中", status.Convert(err).Message())
}

func TestNamespaceSvc_Create_Exists(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewNamespaceSvc(repo.NewMockHelmerRepo(m), nsRepo, k8sRepo, mlog.NewLogger(nil), repo.NewMockEventRepo(m))

	nsRepo.EXPECT().GetMarsNamespace("test").Return("test")
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&repo.Namespace{}, nil)

	res, err := svc.Create(context.TODO(), &namespace.CreateRequest{
		Namespace:      "test",
		IgnoreIfExists: true,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.Exists)
}

func TestNamespaceSvc_Create_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		k8sRepo,
		mlog.NewLogger(nil),
		eventRepo,
	)
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, &ent.NotFoundError{})
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "namespace1").Return(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "namespace1",
		},
	}, nil)
	k8sRepo.EXPECT().CreateDockerSecrets(gomock.Any(), "namespace1").Return(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "docker-secret",
		},
	}, nil)
	nsRepo.EXPECT().Create(gomock.Any(), &repo.CreateNamespaceInput{
		Name:             "namespace1",
		ImagePullSecrets: []string{"docker-secret"},
	}).Return(&repo.Namespace{}, nil)
	nsRepo.EXPECT().Favorite(gomock.Any(), gomock.Any()).Return(nil)
	eventRepo.EXPECT().Dispatch(repo.EventNamespaceCreated, gomock.Any())
	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), "admin", gomock.Any(), gomock.Any())
	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{
		Namespace: "namespace1",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestNamespaceSvc_Create_AlreadyExists(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		k8sRepo,
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&repo.Namespace{}, nil)

	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{
		Namespace: "namespace1",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Create_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		k8sRepo,
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)

	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "namespace1").Return(nil, errors.New("error"))
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errors.New("error"))

	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{
		Namespace: "namespace1",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Delete_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	helmerRepo := repo.NewMockHelmerRepo(m)
	svc := NewNamespaceSvc(
		helmerRepo,
		nsRepo,
		k8sRepo,
		mlog.NewLogger(nil),
		eventRepo,
	)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Namespace{
		ID:               1,
		Name:             "namespace1",
		ImagePullSecrets: []string{"a", "b"},
		Projects: []*repo.Project{
			{
				ID:   1,
				Name: "projA",
			},
			{
				ID:   1,
				Name: "projB",
			},
		},
	}, nil)

	k8sRepo.EXPECT().DeleteSecret(gomock.Any(), "namespace1", "a").Return(nil)
	k8sRepo.EXPECT().DeleteSecret(gomock.Any(), "namespace1", "b").Return(nil)
	helmerRepo.EXPECT().Uninstall("projA", "namespace1", gomock.Any()).Return(nil)
	helmerRepo.EXPECT().Uninstall("projB", "namespace1", gomock.Any()).Return(nil)
	eventRepo.EXPECT().Dispatch(repo.EventNamespaceDeleted, repo.NamespaceDeletedData{ID: 1})
	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), "admin", gomock.Any(), gomock.Any())
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").Return(nil)
	k8sRepo.EXPECT().GetNamespace(gomock.Any(), "namespace1").Return(nil, errors.New("X"))
	nsRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)

	res, err := svc.Delete(newAdminUserCtx(), &namespace.DeleteRequest{
		Id: 1,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestNamespaceSvc_Delete_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))

	res, err := svc.Delete(newAdminUserCtx(), &namespace.DeleteRequest{
		Id: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Favorite_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)

	nsRepo.EXPECT().Favorite(gomock.Any(), &repo.FavoriteNamespaceInput{
		NamespaceID: 1,
		UserEmail:   adminEmail,
		Favorite:    true,
	}).Return(nil)

	res, err := svc.Favorite(newAdminUserCtx(), &namespace.FavoriteRequest{
		Id:       1,
		Favorite: true,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestNamespaceSvc_Favorite_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)

	nsRepo.EXPECT().Favorite(gomock.Any(), &repo.FavoriteNamespaceInput{
		NamespaceID: 1,
		UserEmail:   adminEmail,
		Favorite:    true,
	}).Return(errors.New("error"))

	res, err := svc.Favorite(newAdminUserCtx(), &namespace.FavoriteRequest{
		Id:       1,
		Favorite: true,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_IsExists_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)

	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&repo.Namespace{
		ID:   1,
		Name: "namespace1",
	}, nil)

	res, err := svc.IsExists(context.TODO(), &namespace.IsExistsRequest{
		Name: "namespace1",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.Exists)
	assert.Equal(t, int64(1), res.Id)
}

func TestNamespaceSvc_IsExists_NotFound(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)

	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, &ent.NotFoundError{})

	res, err := svc.IsExists(context.TODO(), &namespace.IsExistsRequest{
		Name: "namespace1",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.False(t, res.Exists)
}

func TestNamespaceSvc_IsExists_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errors.New("error"))

	res, err := svc.IsExists(context.TODO(), &namespace.IsExistsRequest{
		Name: "namespace1",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Show_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Namespace{
		ID:   1,
		Name: "namespace1",
	}, nil)

	res, err := svc.Show(context.TODO(), &namespace.ShowRequest{
		Id: 1,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(1), res.Item.Id)
	assert.Equal(t, "namespace1", res.Item.Name)
}

func TestNamespaceSvc_Show_NotFound(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, &ent.NotFoundError{})

	res, err := svc.Show(context.TODO(), &namespace.ShowRequest{
		Id: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Show_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		repo.NewMockEventRepo(m),
	)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))

	res, err := svc.Show(context.TODO(), &namespace.ShowRequest{
		Id: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}
