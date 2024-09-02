package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v5/namespace"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/util/pagination"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
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

func TestNamespaceSvc_Delete_Error2(t *testing.T) {
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

	res, err := svc.Delete(newOtherUserCtx(), &namespace.DeleteRequest{
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

func Test_namespaceSvc_UpdateDesc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		eventRepo,
	)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Namespace{
		ID:          1,
		Name:        "namespace1",
		Description: "old desc",
	}, nil)
	nsRepo.EXPECT().Update(gomock.Any(), &repo.UpdateNamespaceInput{
		ID:          1,
		Description: "new desc",
	}).Return(&repo.Namespace{
		ID:          1,
		Name:        "namespace1",
		Description: "new desc",
	}, nil)
	eventRepo.EXPECT().AuditLogWithChange(gomock.Any(), "admin", gomock.Any(), gomock.Any(), gomock.Any())

	res, err := svc.UpdateDesc(newAdminUserCtx(), &namespace.UpdateDescRequest{
		Id:   1,
		Desc: "new desc",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func Test_namespaceSvc_UpdateDesc_fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		eventRepo,
	)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))

	res, err := svc.UpdateDesc(newAdminUserCtx(), &namespace.UpdateDescRequest{
		Id:   1,
		Desc: "new desc",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func Test_namespaceSvc_UpdateDesc_fail2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		eventRepo,
	)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Namespace{
		ID:          1,
		Name:        "namespace1",
		Description: "old desc",
	}, nil)
	nsRepo.EXPECT().Update(gomock.Any(), &repo.UpdateNamespaceInput{
		ID:          1,
		Description: "new desc",
	}).Return(nil, errors.New("x"))

	res, err := svc.UpdateDesc(newAdminUserCtx(), &namespace.UpdateDescRequest{
		Id:   1,
		Desc: "new desc",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func Test_namespaceSvc_List(t *testing.T) {
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

	nsRepo.EXPECT().List(gomock.Any(), &repo.ListNamespaceInput{
		Favorite: false,
		Email:    "user@mars.com",
		Name:     lo.ToPtr("name"),
		PageSize: 15,
		Page:     1,
		IsAdmin:  false,
	}).Return([]*repo.Namespace{
		{
			ID:   1,
			Name: "namespace1",
			Favorites: []*repo.Favorite{
				{Email: "user@mars.com"},
			},
		},
		{
			ID:   2,
			Name: "namespace2",
		},
	}, &pagination.Pagination{}, nil)

	res, err := svc.List(newOtherUserCtx(), &namespace.ListRequest{
		Favorite: false,
		Name:     lo.ToPtr("name"),
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

func Test_namespaceSvc_SyncMembers(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		eventRepo,
	)

	nsRepo.EXPECT().IsOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	_, err := svc.SyncMembers(newOtherUserCtx(), &namespace.SyncMembersRequest{
		Id:     1,
		Emails: []string{"a"},
	})
	assert.Error(t, err)

	nsRepo.EXPECT().IsOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	nsRepo.EXPECT().SyncMembers(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("x"))

	nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&repo.Namespace{}, nil)
	ns, err := svc.SyncMembers(newOtherUserCtx(), &namespace.SyncMembersRequest{
		Id:     1,
		Emails: []string{"a"},
	})
	assert.Equal(t, "x", err.Error())
	assert.Nil(t, ns)

	nsRepo.EXPECT().IsOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	nsRepo.EXPECT().SyncMembers(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repo.Namespace{}, nil)

	nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&repo.Namespace{}, nil)
	eventRepo.EXPECT().AuditLogWithChange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	ns, err = svc.SyncMembers(newOtherUserCtx(), &namespace.SyncMembersRequest{
		Id:     1,
		Emails: []string{"a"},
	})
	assert.NotNil(t, ns)
	assert.Nil(t, err)
}

func Test_namespaceSvc_UpdatePrivate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewNamespaceSvc(
		repo.NewMockHelmerRepo(m),
		nsRepo,
		repo.NewMockK8sRepo(m),
		mlog.NewLogger(nil),
		eventRepo,
	)

	nsRepo.EXPECT().IsOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	_, err := svc.UpdatePrivate(newOtherUserCtx(), &namespace.UpdatePrivateRequest{
		Id:      1,
		Private: true,
	})
	assert.Error(t, err)

	nsRepo.EXPECT().IsOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	nsRepo.EXPECT().UpdatePrivate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("x"))
	ns, err := svc.UpdatePrivate(newOtherUserCtx(), &namespace.UpdatePrivateRequest{
		Id:      1,
		Private: true,
	})
	assert.Nil(t, ns)
	assert.Error(t, err)

	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	nsRepo.EXPECT().IsOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	nsRepo.EXPECT().UpdatePrivate(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repo.Namespace{}, nil)
	ns, err = svc.UpdatePrivate(newOtherUserCtx(), &namespace.UpdatePrivateRequest{
		Id:      1,
		Private: true,
	})
	assert.NotNil(t, ns)
	assert.Nil(t, err)
}
