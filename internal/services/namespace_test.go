package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNewNamespaceSvc_Creation(t *testing.T) {
	svc, _ := newNamespaceSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.nsRepoBiz)
	assert.NotNil(t, svc.nsBiz)
	assert.NotNil(t, svc.logger)
	assert.NotNil(t, svc.eventBiz)
}

func TestNamespaceSvc_Create_NamespaceTerminating(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo

	nsRepo.EXPECT().GetMarsNamespace("test").Return("test")
	nsRepo.EXPECT().FindByName(gomock.Any(), "test").Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "test").Return(nil, &k8sapierrors.StatusError{
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

	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{
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
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().GetMarsNamespace("test").Return("test")
	nsRepo.EXPECT().FindByName(gomock.Any(), "test").Return(&biz.Namespace{}, nil)

	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{
		Namespace:      "test",
		IgnoreIfExists: true,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.Exists)
}

// IgnoreIfExists 命中已存在的私有命名空间且当前用户无权访问时，必须 403 拒绝，
// 不能把私有空间完整对象（描述/成员/创建者）返回给无权限用户——与 IsExists
// "私有空间视同不存在"的隐藏语义对齐，闭合幂等放行的元数据泄露面。
func TestNamespaceSvc_Create_IgnoreIfExists_PrivateDenied(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().GetMarsNamespace("test").Return("test")
	nsRepo.EXPECT().FindByName(gomock.Any(), "test").Return(&biz.Namespace{
		ID:           1,
		Name:         "test",
		Private:      true,
		CreatorEmail: "someone-else@example.com",
	}, nil)

	res, err := svc.Create(newOtherUserCtx(), &namespace.CreateRequest{
		Namespace:      "test",
		IgnoreIfExists: true,
	})

	assert.Nil(t, res)
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestNamespaceSvc_Create_Success(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "namespace1").Return(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "namespace1",
		},
	}, nil)
	k8sRepo.EXPECT().CreateDockerSecret(gomock.Any(), "namespace1").Return(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "docker-secret",
		},
	}, nil)
	nsRepo.EXPECT().Create(gomock.Any(), &biz.CreateNamespaceInput{
		Name:             "namespace1",
		ImagePullSecrets: []string{"docker-secret"},
		CreatorEmail:     adminEmail,
	}).Return(&biz.Namespace{ID: 1}, nil)
	nsRepo.EXPECT().Favorite(gomock.Any(), &biz.FavoriteNamespaceInput{
		NamespaceID: 1,
		UserEmail:   adminEmail,
		Favorite:    true,
	}).Return(nil)
	eventRepo.EXPECT().Dispatch(biz.EventNamespaceCreated, gomock.Any())
	req := &namespace.CreateRequest{
		Namespace: "namespace1",
	}
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Create,
		"admin",
		gomock.Any(),
		req,
	)
	res, err := svc.Create(newAdminUserCtx(), req)

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

// 创建成功但自动关注失败：namespace 仍创建成功，错误只打日志不阻断。
func TestNamespaceSvc_Create_FavoriteError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "namespace1").Return(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "namespace1"},
	}, nil)
	k8sRepo.EXPECT().CreateDockerSecret(gomock.Any(), "namespace1").Return(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "docker-secret"},
	}, nil)
	nsRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&biz.Namespace{ID: 1}, nil)
	nsRepo.EXPECT().Favorite(gomock.Any(), gomock.Any()).Return(errors.New("favorite boom"))
	eventRepo.EXPECT().Dispatch(biz.EventNamespaceCreated, gomock.Any())
	req := &namespace.CreateRequest{Namespace: "namespace1"}
	eventRepo.EXPECT().AuditLogWithRequest(types.EventActionType_Create, "admin", gomock.Any(), req)

	res, err := svc.Create(newAdminUserCtx(), req)
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

// DB 记录创建失败时必须回滚本次刚创建的 k8s namespace，避免孤儿资源。
func TestNamespaceSvc_Create_RollbackOnDbError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "namespace1").Return(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "namespace1"},
	}, nil)
	k8sRepo.EXPECT().CreateDockerSecret(gomock.Any(), "namespace1").Return(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "docker-secret"},
	}, nil)
	nsRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db down"))
	// DB 失败后必须回滚刚创建的 namespace。
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").Return(nil)

	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{Namespace: "namespace1"})
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestNamespaceSvc_Create_RollbackDeleteError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "namespace1").Return(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "namespace1"},
	}, nil)
	k8sRepo.EXPECT().CreateDockerSecret(gomock.Any(), "namespace1").Return(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "docker-secret"},
	}, nil)
	nsRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db down"))
	// 回滚 DeleteNamespace 也失败时只打日志，不阻断返回原始 DB 错误。
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").Return(errors.New("delete boom"))

	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{Namespace: "namespace1"})
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestNamespaceSvc_Create_AlreadyExists(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{}, nil)

	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{
		Namespace: "namespace1",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Create_AlreadyExists_Adopt(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "namespace1").Return(nil, &k8sapierrors.StatusError{
		ErrStatus: metav1.Status{
			Reason: metav1.StatusReasonAlreadyExists,
		},
	})
	// 已存在的 k8s namespace，非 Terminating → 走收养路径
	k8sRepo.EXPECT().GetNamespace(gomock.Any(), "namespace1").Return(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "namespace1",
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}, nil)
	k8sRepo.EXPECT().CreateDockerSecret(gomock.Any(), "namespace1").Return(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "docker-secret",
		},
	}, nil)
	// 核心断言：收养路径必须用 found.Name，而不是失败 create 的空 Name
	nsRepo.EXPECT().Create(gomock.Any(), &biz.CreateNamespaceInput{
		Name:             "namespace1",
		ImagePullSecrets: []string{"docker-secret"},
		CreatorEmail:     adminEmail,
	}).Return(&biz.Namespace{ID: 1}, nil)
	nsRepo.EXPECT().Favorite(gomock.Any(), &biz.FavoriteNamespaceInput{
		NamespaceID: 1,
		UserEmail:   adminEmail,
		Favorite:    true,
	}).Return(nil)
	eventRepo.EXPECT().Dispatch(biz.EventNamespaceCreated, gomock.Any())
	req := &namespace.CreateRequest{Namespace: "namespace1"}
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Create,
		"admin",
		gomock.Any(),
		req,
	)
	res, err := svc.Create(newAdminUserCtx(), req)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.False(t, res.Exists)
}

func TestNamespaceSvc_Create_FindByNameError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	// 预检查返回非 NotFound 的真实错误（如 DB 故障）时必须直接上抛，
	// 不能误判为"不存在"后继续走 k8s 创建流程。
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errors.New("error"))

	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{
		Namespace: "namespace1",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Create_K8sCreateError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo

	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "namespace1").Return(nil, errors.New("error"))

	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{
		Namespace: "namespace1",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Delete_SecretError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{
		ID:               1,
		Name:             "namespace1",
		ImagePullSecrets: []string{"a"},
	}, nil)

	// DeleteSecret 失败只记录日志，不中断删除流程。
	k8sRepo.EXPECT().DeleteSecret(gomock.Any(), "namespace1", "a").Return(errors.New("secret error"))
	eventRepo.EXPECT().Dispatch(biz.EventNamespaceDeleted, biz.NamespaceDeletedData{ID: 1})
	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), "admin", gomock.Any(), gomock.Any())
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").Return(nil)
	k8sRepo.EXPECT().GetNamespace(gomock.Any(), "namespace1").Return(nil, k8sapierrors.NewNotFound(schema.GroupResource{Resource: "namespaces"}, "namespace1"))
	nsRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)

	res, err := svc.Delete(newAdminUserCtx(), &namespace.DeleteRequest{
		Id: 1,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestNamespaceSvc_Delete_Success(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo
	helmerRepo := mocks.helmerRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{
		ID:               1,
		Name:             "namespace1",
		ImagePullSecrets: []string{"a", "b"},
		Projects: []*biz.Project{
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
	eventRepo.EXPECT().Dispatch(biz.EventNamespaceDeleted, biz.NamespaceDeletedData{ID: 1})
	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), "admin", gomock.Any(), gomock.Any())
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").Return(nil)
	k8sRepo.EXPECT().GetNamespace(gomock.Any(), "namespace1").Return(nil, k8sapierrors.NewNotFound(schema.GroupResource{Resource: "namespaces"}, "namespace1"))
	nsRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)

	res, err := svc.Delete(newAdminUserCtx(), &namespace.DeleteRequest{
		Id: 1,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestNamespaceSvc_Delete_Error(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))

	res, err := svc.Delete(newAdminUserCtx(), &namespace.DeleteRequest{
		Id: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Delete_Error2(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{
		ID:           1,
		Name:         "namespace1",
		CreatorEmail: "someone-else@mars.com",
	}, nil)

	res, err := svc.Delete(newOtherUserCtx(), &namespace.DeleteRequest{
		Id: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Favorite_Success(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	eventRepo := mocks.eventRepo

	nsRepo.EXPECT().Favorite(gomock.Any(), &biz.FavoriteNamespaceInput{
		NamespaceID: 1,
		UserEmail:   adminEmail,
		Favorite:    true,
	}).Return(nil)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{
		Name: "namespace1",
	}, nil)

	req := &namespace.FavoriteRequest{
		Id:       1,
		Favorite: true,
	}
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Update,
		biz.MustGetUser(newAdminUserCtx()).Name,
		"用户关注项目空间: namespace1",
		req,
	)

	res, err := svc.Favorite(newAdminUserCtx(), req)

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestNamespaceSvc_Favorite_Error(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Favorite(gomock.Any(), &biz.FavoriteNamespaceInput{
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
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{
		ID:   1,
		Name: "namespace1",
	}, nil)

	res, err := svc.IsExists(newAdminUserCtx(), &namespace.IsExistsRequest{
		Name: "namespace1",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.Exists)
	assert.Equal(t, int64(1), res.Id)
}

func TestNamespaceSvc_IsExists_PrivateNamespaceHidden(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{
		ID:           1,
		Name:         "namespace1",
		Private:      true,
		CreatorEmail: "someone-else@example.com",
	}, nil)

	res, err := svc.IsExists(newOtherUserCtx(), &namespace.IsExistsRequest{
		Name: "namespace1",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.False(t, res.Exists)
}

func TestNamespaceSvc_IsExists_NotFound(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))

	res, err := svc.IsExists(context.TODO(), &namespace.IsExistsRequest{
		Name: "namespace1",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.False(t, res.Exists)
}

func TestNamespaceSvc_IsExists_Error(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errors.New("error"))

	res, err := svc.IsExists(context.TODO(), &namespace.IsExistsRequest{
		Name: "namespace1",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Show_Success(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{
		ID:   1,
		Name: "namespace1",
	}, nil)

	res, err := svc.Show(newAdminUserCtx(), &namespace.ShowRequest{
		Id: 1,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(1), res.Item.Id)
	assert.Equal(t, "namespace1", res.Item.Name)
}

func TestNamespaceSvc_Show_NotFound(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))

	res, err := svc.Show(newAdminUserCtx(), &namespace.ShowRequest{
		Id: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Show_Error(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))

	res, err := svc.Show(newAdminUserCtx(), &namespace.ShowRequest{
		Id: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Show_Error2(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{
		Private: true,
	}, nil)

	_, err := svc.Show(newOtherUserCtx(), &namespace.ShowRequest{
		Id: 1,
	})

	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func Test_namespaceSvc_UpdateDesc(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	eventRepo := mocks.eventRepo
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{
		ID:          1,
		Name:        "namespace1",
		Description: "old desc",
	}, nil)
	nsRepo.EXPECT().Update(gomock.Any(), &biz.UpdateNamespaceInput{
		ID:          1,
		Description: "new desc",
	}).Return(&biz.Namespace{
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
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))

	res, err := svc.UpdateDesc(newAdminUserCtx(), &namespace.UpdateDescRequest{
		Id:   1,
		Desc: "new desc",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func Test_namespaceSvc_UpdateDesc_fail2(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{
		ID:          1,
		Name:        "namespace1",
		Description: "old desc",
	}, nil)
	nsRepo.EXPECT().Update(gomock.Any(), &biz.UpdateNamespaceInput{
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
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().List(gomock.Any(), &biz.ListNamespaceInput{
		Favorite: false,
		Email:    "user@mars.com",
		Name:     lo.ToPtr("name"),
		PageSize: 15,
		Page:     1,
		IsAdmin:  false,
	}).Return([]*biz.Namespace{
		{
			ID:   1,
			Name: "namespace1",
			Favorites: []*biz.Favorite{
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
	// 反向断言：无收藏命名的命名空间必须保持 Favorite=false，防止聚合逻辑误标。
	assert.False(t, res.Items[1].Favorite)

	nsRepo.EXPECT().List(gomock.Any(), &biz.ListNamespaceInput{
		Favorite: false,
		Email:    "user@mars.com",
		Name:     lo.ToPtr("name"),
		PageSize: 15,
		Page:     1,
		IsAdmin:  false,
	}).Return(nil, nil, errors.New("x"))

	_, err = svc.List(newOtherUserCtx(), &namespace.ListRequest{
		Favorite: false,
		Name:     lo.ToPtr("name"),
	})
	assert.Error(t, err)
}

func Test_namespaceSvc_SyncMembers(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	eventRepo := mocks.eventRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{CreatorEmail: "other@email.com"}, nil)
	_, err := svc.SyncMembers(newOtherUserCtx(), &namespace.SyncMembersRequest{
		Id:     1,
		Emails: []string{"a"},
	})
	assert.Error(t, err)

	nsRepo.EXPECT().SyncMembers(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("x"))

	nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{CreatorEmail: "user@mars.com"}, nil)
	ns, err := svc.SyncMembers(newOtherUserCtx(), &namespace.SyncMembersRequest{
		Id:     1,
		Emails: []string{"a"},
	})
	assert.Equal(t, "x", err.Error())
	assert.Nil(t, ns)

	nsRepo.EXPECT().SyncMembers(gomock.Any(), gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil)

	nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{CreatorEmail: "user@mars.com"}, nil)
	eventRepo.EXPECT().AuditLogWithChange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	ns, err = svc.SyncMembers(newOtherUserCtx(), &namespace.SyncMembersRequest{
		Id:     1,
		Emails: []string{"a"},
	})
	assert.NotNil(t, ns)
	assert.Nil(t, err)
}

func Test_namespaceSvc_UpdatePrivate(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	eventRepo := mocks.eventRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{CreatorEmail: "other@email.com"}, nil)
	_, err := svc.UpdatePrivate(newOtherUserCtx(), &namespace.UpdatePrivateRequest{
		Id:      1,
		Private: true,
	})
	assert.Error(t, err)

	nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{CreatorEmail: "user@mars.com"}, nil)
	nsRepo.EXPECT().UpdatePrivate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("x"))
	ns, err := svc.UpdatePrivate(newOtherUserCtx(), &namespace.UpdatePrivateRequest{
		Id:      1,
		Private: true,
	})
	assert.Nil(t, ns)
	assert.Error(t, err)

	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{CreatorEmail: "user@mars.com"}, nil)
	nsRepo.EXPECT().UpdatePrivate(gomock.Any(), gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil)
	ns, err = svc.UpdatePrivate(newOtherUserCtx(), &namespace.UpdatePrivateRequest{
		Id:      1,
		Private: true,
	})
	assert.NotNil(t, ns)
	assert.Nil(t, err)
}

func Test_namespaceSvc_Transfer(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	eventRepo := mocks.eventRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{CreatorEmail: "other@email.com"}, nil)
	_, err := svc.Transfer(newOtherUserCtx(), &namespace.TransferRequest{
		Id:            1,
		NewAdminEmail: "a",
	})
	assert.Error(t, err)

	nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{CreatorEmail: "user@mars.com"}, nil)
	nsRepo.EXPECT().Transfer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("x"))
	ns, err := svc.Transfer(newOtherUserCtx(), &namespace.TransferRequest{
		Id:            1,
		NewAdminEmail: "a",
	})
	assert.Nil(t, ns)
	assert.Error(t, err)

	req := &namespace.TransferRequest{
		Id:            1,
		NewAdminEmail: "a",
	}
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Update,
		biz.MustGetUser(newOtherUserCtx()).Name,
		"转让项目空间给: a",
		req,
	)
	nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{CreatorEmail: "user@mars.com"}, nil)
	nsRepo.EXPECT().Transfer(gomock.Any(), int(req.Id), req.NewAdminEmail).Return(&biz.Namespace{}, nil)
	ns, err = svc.Transfer(newOtherUserCtx(), req)
	assert.NotNil(t, ns)
	assert.Nil(t, err)
}

func TestNamespaceSvc_Transfer_ShowError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))

	res, err := svc.Transfer(newAdminUserCtx(), &namespace.TransferRequest{Id: 1, NewAdminEmail: "a"})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Create_GetNamespaceError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo

	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))
	// CreateNamespace 报 AlreadyExists → 收养路径，但 GetNamespace 也失败
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "namespace1").Return(nil, &k8sapierrors.StatusError{
		ErrStatus: metav1.Status{Reason: metav1.StatusReasonAlreadyExists},
	})
	k8sRepo.EXPECT().GetNamespace(gomock.Any(), "namespace1").Return(nil, errors.New("x"))

	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{Namespace: "namespace1"})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Create_CreateDockerSecretError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo

	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "namespace1").Return(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "namespace1"},
	}, nil)
	// docker secret 创建失败 → 走 else Debug 分支，imagePullSecrets 为空
	k8sRepo.EXPECT().CreateDockerSecret(gomock.Any(), "namespace1").Return(nil, errors.New("secret boom"))
	nsRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&biz.Namespace{ID: 1}, nil)
	nsRepo.EXPECT().Favorite(gomock.Any(), &biz.FavoriteNamespaceInput{
		NamespaceID: 1,
		UserEmail:   adminEmail,
		Favorite:    true,
	}).Return(nil)
	eventRepo.EXPECT().Dispatch(biz.EventNamespaceCreated, gomock.Any())
	req := &namespace.CreateRequest{Namespace: "namespace1"}
	eventRepo.EXPECT().AuditLogWithRequest(types.EventActionType_Create, "admin", gomock.Any(), req)

	res, err := svc.Create(newAdminUserCtx(), req)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.False(t, res.Exists)
}

func TestNamespaceSvc_Create_RepoCreateError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo

	nsRepo.EXPECT().GetMarsNamespace("namespace1").Return("namespace1")
	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))
	k8sRepo.EXPECT().CreateNamespace(gomock.Any(), "namespace1").Return(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "namespace1"},
	}, nil)
	k8sRepo.EXPECT().CreateDockerSecret(gomock.Any(), "namespace1").Return(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "docker-secret"},
	}, nil)
	nsRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("x"))
	// DB 记录失败时必须回滚刚创建的 k8s namespace，避免留下孤儿资源。
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").Return(nil)

	res, err := svc.Create(newAdminUserCtx(), &namespace.CreateRequest{Namespace: "namespace1"})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Delete_UninstallError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo
	helmerRepo := mocks.helmerRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{
		ID:               1,
		Name:             "namespace1",
		ImagePullSecrets: []string{"a"},
		Projects:         []*biz.Project{{ID: 1, Name: "projA"}},
	}, nil)
	// Uninstall 失败：错误在 biz 并发卸载编排内被吞，继续删 secret/命名空间/DB，不阻塞整体流程
	helmerRepo.EXPECT().Uninstall("projA", "namespace1", gomock.Any()).Return(errors.New("helm boom"))
	k8sRepo.EXPECT().DeleteSecret(gomock.Any(), "namespace1", "a").Return(nil)
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").Return(nil)
	k8sRepo.EXPECT().GetNamespace(gomock.Any(), "namespace1").Return(nil, k8sapierrors.NewNotFound(schema.GroupResource{Resource: "namespaces"}, "namespace1"))
	nsRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)
	eventRepo.EXPECT().Dispatch(biz.EventNamespaceDeleted, biz.NamespaceDeletedData{ID: 1})
	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), "admin", gomock.Any(), gomock.Any())

	res, err := svc.Delete(newAdminUserCtx(), &namespace.DeleteRequest{Id: 1})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestNamespaceSvc_Delete_DeleteNamespaceError(t *testing.T) {
	// F18 同类回归：k8s DeleteNamespace 返回非 NotFound 的真实错误时必须 abort，
	// 不得继续删 DB 记录——否则留下孤儿 namespace，且轮询超时后会误发 NamespaceDeleted 事件。
	// 改坏实现（log-and-continue）时：nsRepo.Delete/Dispatch/AuditLog 被意外调用 → 测试 FAIL。
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1, Name: "namespace1"}, nil)
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").Return(errors.New("ns boom"))
	// 不设置 GetNamespace / nsRepo.Delete / Dispatch / AuditLog 期望：
	// 真实错误路径必须提前 return，任何后续调用都会 FAIL。

	res, err := svc.Delete(newAdminUserCtx(), &namespace.DeleteRequest{Id: 1})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

// injectFastNamespaceDeletePolling 用 100ms 超时 / 10ms 轮询覆盖 biz 导出默认（5s/500ms），
// 让 Delete 轮询相关测试覆盖 timer/ticker 分支而不支付真实墙钟，结束后自动恢复。
func injectFastNamespaceDeletePolling(t *testing.T) {
	t.Helper()
	biz.NamespaceDeleteTimeout = 100 * time.Millisecond
	biz.NamespacePollInterval = 10 * time.Millisecond
	t.Cleanup(func() {
		biz.NamespaceDeleteTimeout = 5 * time.Second
		biz.NamespacePollInterval = 500 * time.Millisecond
	})
}

func TestNamespaceSvc_Delete_DeleteNamespaceNotFoundIsClean(t *testing.T) {
	// F18 同类回归：k8s namespace 已不存在（NotFound）视为删除干净，正常继续删 DB 记录。
	// 改坏实现（把 NotFound 当真实错误 abort）时：Dispatch/AuditLog 期望落空 → 测试 FAIL。
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo

	injectFastNamespaceDeletePolling(t)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1, Name: "namespace1"}, nil)
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").
		Return(k8sapierrors.NewNotFound(schema.GroupResource{Resource: "namespaces"}, "namespace1"))
	k8sRepo.EXPECT().GetNamespace(gomock.Any(), "namespace1").
		Return(nil, k8sapierrors.NewNotFound(schema.GroupResource{Resource: "namespaces"}, "namespace1"))
	nsRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)
	eventRepo.EXPECT().Dispatch(biz.EventNamespaceDeleted, biz.NamespaceDeletedData{ID: 1})
	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), "admin", gomock.Any(), gomock.Any())

	res, err := svc.Delete(newAdminUserCtx(), &namespace.DeleteRequest{Id: 1})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestNamespaceSvc_Delete_RepoDeleteError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1, Name: "namespace1"}, nil)
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").Return(nil)
	// nsRepoBiz.Delete 失败 → 提前返回，轮询循环不会执行
	nsRepo.EXPECT().Delete(gomock.Any(), 1).Return(errors.New("x"))

	res, err := svc.Delete(newAdminUserCtx(), &namespace.DeleteRequest{Id: 1})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_Delete_Timer(t *testing.T) {
	// 轮询期间 GetNamespace 一直成功 → 只能等 5s 计时器触发 break loop（走 timer.C 分支）
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo

	injectFastNamespaceDeletePolling(t)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1, Name: "namespace1"}, nil)
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").Return(nil)
	k8sRepo.EXPECT().GetNamespace(gomock.Any(), "namespace1").Return(&corev1.Namespace{}, nil).AnyTimes()
	nsRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)
	eventRepo.EXPECT().Dispatch(biz.EventNamespaceDeleted, biz.NamespaceDeletedData{ID: 1})
	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), "admin", gomock.Any(), gomock.Any())

	start := time.Now()
	res, err := svc.Delete(newAdminUserCtx(), &namespace.DeleteRequest{Id: 1})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// 覆盖 timer.C break loop 分支：注入 100ms 超时后断言确实等待了轮询（而非瞬断），
	// 不再为这个分支支付真实 5s 墙钟。
	assert.GreaterOrEqual(t, time.Since(start), 80*time.Millisecond)
}

func TestNamespaceSvc_Delete_TransientErrorContinuesPolling(t *testing.T) {
	// 回归防护：轮询期间 GetNamespace 返回非 NotFound 瞬态错误（状态未知）时必须
	// 继续轮询，不能误判为"已删除"提前 break。改坏实现（任意错误都 break）时此测试 FAIL。
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{ID: 1, Name: "namespace1"}, nil)
	k8sRepo.EXPECT().DeleteNamespace(gomock.Any(), "namespace1").Return(nil)
	nsRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)
	eventRepo.EXPECT().Dispatch(biz.EventNamespaceDeleted, biz.NamespaceDeletedData{ID: 1})
	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), "admin", gomock.Any(), gomock.Any())

	injectFastNamespaceDeletePolling(t)
	calls := 0
	k8sRepo.EXPECT().GetNamespace(gomock.Any(), "namespace1").DoAndReturn(func(ctx context.Context, name string) (*corev1.Namespace, error) {
		calls++
		if calls == 1 {
			// 第一次轮询：瞬态错误，namespace 是否已删除未知
			return nil, errors.New("transient api error")
		}
		return nil, k8sapierrors.NewNotFound(schema.GroupResource{Resource: "namespaces"}, name)
	}).AnyTimes()

	res, err := svc.Delete(newAdminUserCtx(), &namespace.DeleteRequest{Id: 1})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	// 至少轮询 2 次：第一次瞬态错误后没有 break，而是继续轮询直到真正 NotFound
	assert.GreaterOrEqual(t, calls, 2)
}

func TestNamespaceSvc_Favorite_ShowError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Favorite(gomock.Any(), &biz.FavoriteNamespaceInput{
		NamespaceID: 1,
		UserEmail:   adminEmail,
		Favorite:    true,
	}).Return(nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))

	res, err := svc.Favorite(newAdminUserCtx(), &namespace.FavoriteRequest{Id: 1, Favorite: true})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_UpdatePrivate_ShowError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))

	res, err := svc.UpdatePrivate(newAdminUserCtx(), &namespace.UpdatePrivateRequest{Id: 1, Private: true})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNamespaceSvc_SyncMembers_ShowError(t *testing.T) {
	svc, mocks := newNamespaceSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))

	res, err := svc.SyncMembers(newAdminUserCtx(), &namespace.SyncMembersRequest{Id: 1, Emails: []string{"a"}})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

// namespaceSvcMocks 聚合 namespaceSvc 的全部下游 mock，由 newNamespaceSvcWithMocks 统一构造。
type namespaceSvcMocks struct {
	ctrl       *gomock.Controller
	helmerRepo *data.MockHelmerRepo
	nsRepo     *data.MockNamespaceRepo
	k8sRepo    *data.MockK8sRepo
	eventRepo  *data.MockEventRepo
}

func newNamespaceSvcWithMocks(t *testing.T) (*namespaceSvc, *namespaceSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &namespaceSvcMocks{
		ctrl:       ctrl,
		helmerRepo: data.NewMockHelmerRepo(ctrl),
		nsRepo:     data.NewMockNamespaceRepo(ctrl),
		k8sRepo:    data.NewMockK8sRepo(ctrl),
		eventRepo:  data.NewMockEventRepo(ctrl),
	}
	logger := mlog.NewForConfig(nil)
	s, ok := NewNamespaceSvc(NamespaceSvcDeps{
		NsRepoBiz: biz.NewNsRepoBiz(mocks.nsRepo),
		NsBiz: biz.NewNamespaceBiz(
			logger,
			mocks.nsRepo,
			mocks.k8sRepo,
			mocks.helmerRepo,
			mocks.eventRepo,
		),
		Logger:    logger,
		EventBiz:  biz.NewEventBiz(mocks.eventRepo),
		AccessBiz: biz.NewAccessBiz(biz.NewNsRepoBiz(mocks.nsRepo), nil),
	}).(*namespaceSvc)
	if !ok {
		panic("NewNamespaceSvc returned unexpected type")
	}
	return s, mocks
}
