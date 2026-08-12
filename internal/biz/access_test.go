package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// plainUser 构造普通登录用户（非 admin、非空间创建者），供各访问判定用例复用。
func plainUser() *UserInfo { return &UserInfo{Email: "user@example.com"} }

// accessBizMocks 聚合 AccessBiz 实现（accessBiz）的全部下游 mock。
type accessBizMocks struct {
	ctrl   *gomock.Controller
	nsRepo *MockNamespaceBiz
	proj   *MockProjectBiz
}

// newAccessBizFixture 构造 AccessBiz 并暴露其下游 mock。用户经 userCtx 注入
// 请求上下文——并包后 AccessBiz 内部走 MustGetUser 提取，不再注入 getUser 回调。
func newAccessBizFixture(t *testing.T) (AccessBiz, *accessBizMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	m := &accessBizMocks{ctrl: ctrl, nsRepo: NewMockNamespaceBiz(ctrl), proj: NewMockProjectBiz(ctrl)}
	ab := NewAccessBiz(m.nsRepo, m.proj)
	return ab, m
}

// userCtx 构造携带指定用户的请求上下文：user 经 SetUser 注入后由 MustGetUser 提取。
// 注意：ctx 无用户时 MustGetUser 会 panic（编程错误，见 context_test.go），
// 因此访问判定用例一律通过 userCtx 注入用户，不再测试 nil-user 拒绝路径。
func userCtx(u *UserInfo) context.Context { return SetUser(context.TODO(), u) }

// plainCtx / adminCtx / creatorCtx 分别是普通用户 / admin / 命名空间创建者的请求上下文。
func plainCtx() context.Context { return userCtx(plainUser()) }
func adminCtx() context.Context {
	return userCtx(&UserInfo{Email: "a@b.c", Roles: []string{MarsAdmin}})
}
func creatorCtx() context.Context { return userCtx(&UserInfo{Email: "owner@example.com"}) }

func TestNewAccessBiz(t *testing.T) {
	ab := NewAccessBiz(nil, nil)
	assert.NotNil(t, ab)
}

func TestAccessBiz_RequireNamespaceAccessByName(t *testing.T) {
	t.Run("public namespace accessible", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.nsRepo.EXPECT().FindByName(gomock.Any(), "ns").Return(&Namespace{Name: "ns", Private: false}, nil)

		ns, err := ab.RequireNamespaceAccessByName(plainCtx(), "ns")
		assert.NoError(t, err)
		assert.Equal(t, "ns", ns.Name)
	})

	t.Run("load error logged and propagated", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.nsRepo.EXPECT().FindByName(gomock.Any(), "ns").Return(nil, errors.New("boom"))

		_, err := ab.RequireNamespaceAccessByName(plainCtx(), "ns")
		assert.Error(t, err)
	})

	t.Run("private namespace denied to non-member", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.nsRepo.EXPECT().FindByName(gomock.Any(), "ns").Return(&Namespace{Private: true, CreatorEmail: "owner@example.com"}, nil)

		_, err := ab.RequireNamespaceAccessByName(plainCtx(), "ns")
		assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
	})
}

// TestAccessBiz_RequireNamespaceAccessByID 覆盖 CanAccessNamespace 谓词的分支：
// admin/创建者/成员/公开空间/非成员拒绝/nil ns。nil-user 不再是合法输入——ctx
// 无用户即编程错误，MustGetUser 直接 panic（见 context_test.go），实现无 nil-user 分支。
func TestAccessBiz_RequireNamespaceAccessByID(t *testing.T) {
	t.Run("public namespace accessible by non-member", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&Namespace{ID: 1, Private: false}, nil)

		ns, err := ab.RequireNamespaceAccessByID(plainCtx(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, ns.ID)
	})

	t.Run("load error logged and propagated", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("boom"))

		_, err := ab.RequireNamespaceAccessByID(plainCtx(), 1)
		assert.Error(t, err)
	})

	t.Run("admin accesses private namespace", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&Namespace{ID: 1, Private: true, CreatorEmail: "owner@example.com"}, nil)

		ns, err := ab.RequireNamespaceAccessByID(adminCtx(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, ns.ID)
	})

	t.Run("creator accesses private namespace", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&Namespace{ID: 1, Private: true, CreatorEmail: "owner@example.com"}, nil)

		ns, err := ab.RequireNamespaceAccessByID(creatorCtx(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, ns.ID)
	})

	t.Run("member accesses private namespace", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&Namespace{ID: 1, Private: true, CreatorEmail: "owner@example.com", Members: []*Member{{Email: "user@example.com"}}}, nil)

		ns, err := ab.RequireNamespaceAccessByID(plainCtx(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, ns.ID)
	})

	t.Run("private namespace denied to non-member", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&Namespace{ID: 1, Private: true, CreatorEmail: "owner@example.com"}, nil)

		_, err := ab.RequireNamespaceAccessByID(plainCtx(), 1)
		assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
	})

	t.Run("nil namespace denied", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, nil)

		_, err := ab.RequireNamespaceAccessByID(plainCtx(), 1)
		assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
	})
}

func TestAccessBiz_RequireProjectAccess(t *testing.T) {
	t.Run("accessible", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.proj.EXPECT().Show(gomock.Any(), 1).Return(&Project{ID: 1, NamespaceID: 5}, nil)
		m.nsRepo.EXPECT().Show(gomock.Any(), 5).Return(&Namespace{Private: false}, nil)

		proj, err := ab.RequireProjectAccess(plainCtx(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, proj.ID)
	})

	t.Run("project load error logged and propagated", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.proj.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("boom"))

		_, err := ab.RequireProjectAccess(plainCtx(), 1)
		assert.Error(t, err)
	})

	t.Run("namespace denied to non-member", func(t *testing.T) {
		ab, m := newAccessBizFixture(t)
		m.proj.EXPECT().Show(gomock.Any(), 1).Return(&Project{ID: 1, NamespaceID: 5}, nil)
		m.nsRepo.EXPECT().Show(gomock.Any(), 5).Return(&Namespace{Private: true, CreatorEmail: "owner@example.com"}, nil)

		_, err := ab.RequireProjectAccess(plainCtx(), 1)
		assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
	})
}

func TestAccessBiz_RequireNamespaceOwner(t *testing.T) {
	t.Run("owner passes", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		err := ab.RequireNamespaceOwner(creatorCtx(), &Namespace{Private: true, CreatorEmail: "owner@example.com"})
		assert.NoError(t, err)
	})

	t.Run("admin passes", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		err := ab.RequireNamespaceOwner(adminCtx(), &Namespace{Private: true, CreatorEmail: "owner@example.com"})
		assert.NoError(t, err)
	})

	t.Run("non-owner denied", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		err := ab.RequireNamespaceOwner(plainCtx(), &Namespace{Private: true, CreatorEmail: "owner@example.com"})
		assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
	})

	t.Run("nil namespace denied", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		err := ab.RequireNamespaceOwner(plainCtx(), nil)
		assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
	})
}

func TestAccessBiz_RequireAdmin(t *testing.T) {
	t.Run("allowlist hit bypasses user extraction", func(t *testing.T) {
		// ctx 不注入用户：allowlist 精确命中即放行，不触达 MustGetUser，故不 panic。
		ab, _ := newAccessBizFixture(t)

		ctx, err := ab.RequireAdmin(context.TODO(), "/file.File/MaxUploadSize", "/file.File/MaxUploadSize")
		assert.NoError(t, err)
		assert.NotNil(t, ctx)
	})

	t.Run("admin passes", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		_, err := ab.RequireAdmin(adminCtx(), "/file.File/List")
		assert.NoError(t, err)
	})

	t.Run("non-admin denied", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		_, err := ab.RequireAdmin(plainCtx(), "/file.File/List")
		assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
	})
}

// TestAccessBiz_CanAccessNamespace 直接覆盖 CanAccessNamespace 谓词的分支：nil ns
// 拒绝、admin/创建者/成员/公开空间放行、非成员拒绝。nil-user 不再是合法输入——
// ctx 无用户即编程错误，MustGetUser 直接 panic（见 context_test.go）。
func TestAccessBiz_CanAccessNamespace(t *testing.T) {
	t.Run("nil namespace denied", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		assert.False(t, ab.CanAccessNamespace(plainCtx(), nil))
	})

	t.Run("admin accesses private namespace", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		assert.True(t, ab.CanAccessNamespace(adminCtx(), &Namespace{Private: true, CreatorEmail: "owner@example.com"}))
	})

	t.Run("public namespace accessible by non-member", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		assert.True(t, ab.CanAccessNamespace(plainCtx(), &Namespace{Private: false}))
	})

	t.Run("creator accesses private namespace", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		assert.True(t, ab.CanAccessNamespace(creatorCtx(), &Namespace{Private: true, CreatorEmail: "owner@example.com"}))
	})

	t.Run("member accesses private namespace", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		assert.True(t, ab.CanAccessNamespace(plainCtx(), &Namespace{Private: true, CreatorEmail: "owner@example.com", Members: []*Member{{Email: "user@example.com"}}}))
	})

	t.Run("non-member denied private namespace", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		assert.False(t, ab.CanAccessNamespace(plainCtx(), &Namespace{Private: true, CreatorEmail: "owner@example.com"}))
	})
}

// TestAccessBiz_RequireFileAccess 覆盖文件访问判定的路径：所有者/admin 放行、
// 非所有者非 admin 拒绝。nil-user 不再是合法输入——ctx 无用户即编程错误，
// MustGetUser 直接 panic（见 context_test.go）。
func TestAccessBiz_RequireFileAccess(t *testing.T) {
	fil := &File{Username: "owner"}

	t.Run("owner passes", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		assert.NoError(t, ab.RequireFileAccess(userCtx(&UserInfo{Name: "owner"}), fil))
	})

	t.Run("admin passes", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		assert.NoError(t, ab.RequireFileAccess(adminCtx(), fil))
	})

	t.Run("non-owner non-admin denied", func(t *testing.T) {
		ab, _ := newAccessBizFixture(t)

		assert.ErrorIs(t, ab.RequireFileAccess(plainCtx(), fil), errs.ErrorPermissionDenied)
	})
}
