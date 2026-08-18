package errs

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	_ "github.com/duc-cnzj/mars/v6/internal/data/ent/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	_ "github.com/mattn/go-sqlite3"
)

// TestWrapConstructors_NilErr 覆盖 wrapErr 的 nil 提前返回分支：底层 err 为 nil 时
// 四个构造器（含通用 Wrap）均返回 nil，不构造无意义的错误。
func TestWrapConstructors_NilErr(t *testing.T) {
	assert.Nil(t, WrapNotFound(nil, "query access token"))
	assert.Nil(t, WrapInvalidArgument(nil, "create repo"))
	assert.Nil(t, WrapUnauthenticated(nil, "verify token"))
	assert.Nil(t, WrapPermissionDenied(nil, "permission denied"))
	assert.Nil(t, Wrap(nil, "revoke access token"))
}

// TestWrapConstructors 覆盖三个语义构造器的协议码映射与 grpcStatusError 的全部方法
// （Error/Unwrap/Format/GRPCStatus）：客户端可见 message 为底层 err.Error()，
// 上下文 msg 进入 wrap 链供日志打印，%v/%+v 格式化不 panic。
func TestWrapConstructors(t *testing.T) {
	t.Run("WrapNotFound", func(t *testing.T) {
		err := WrapNotFound(assert.AnError, "query access token")
		assert.Equal(t, codes.NotFound, status.Code(err))
		assert.ErrorIs(t, err, assert.AnError)
		assert.Contains(t, err.Error(), "query access token")
		assert.NotPanics(t, func() { _ = fmt.Sprintf("%v", err) })
		assert.NotPanics(t, func() { _ = fmt.Sprintf("%+v", err) })
	})

	t.Run("WrapInvalidArgument", func(t *testing.T) {
		underlying := errors.New("repo 名称已经存在")
		err := WrapInvalidArgument(underlying, "create repo")
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.ErrorIs(t, err, underlying)
	})

	t.Run("WrapUnauthenticated", func(t *testing.T) {
		underlying := errors.New("token 验证失败")
		err := WrapUnauthenticated(underlying, "verify token")
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
		assert.ErrorIs(t, err, underlying)
	})

	t.Run("WrapPermissionDenied", func(t *testing.T) {
		underlying := errors.New("gitlab 无权限")
		err := WrapPermissionDenied(underlying, "permission denied")
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
		assert.ErrorIs(t, err, underlying)
	})
}

// TestMessageConstructors 覆盖五个消息构造器（NotFound/InvalidArgument/Unauthenticated/
// PermissionDenied/AlreadyExists）：直接携带消息返回对应协议码，无底层错误可包裹。
func TestMessageConstructors(t *testing.T) {
	testCases := []struct {
		name string
		got  error
		code codes.Code
		msg  string
	}{
		{"NotFound", NotFound("记录不存在"), codes.NotFound, "记录不存在"},
		{"InvalidArgument", InvalidArgument("参数非法"), codes.InvalidArgument, "参数非法"},
		{"Unauthenticated", Unauthenticated("未认证"), codes.Unauthenticated, "未认证"},
		{"PermissionDenied", PermissionDenied("无权限"), codes.PermissionDenied, "无权限"},
		{"AlreadyExists", AlreadyExists("资源已存在"), codes.AlreadyExists, "资源已存在"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.code, status.Code(tc.got))
			assert.Equal(t, tc.msg, status.Convert(tc.got).Message())
		})
	}
}

// TestErrorPermissionDenied_Sentinel 验证权限拒绝 sentinel 的协议码与消息，
// 保证 services 层 errors.Is(err, ErrorPermissionDenied) 可匹配同一实例。
func TestErrorPermissionDenied_Sentinel(t *testing.T) {
	assert.Equal(t, codes.PermissionDenied, status.Code(ErrorPermissionDenied))
	assert.Equal(t, "没有权限执行该操作", status.Convert(ErrorPermissionDenied).Message())
	assert.True(t, errors.Is(ErrorPermissionDenied, ErrorPermissionDenied))
}

// TestWrap_GenericInternal 覆盖通用 Wrap 构造器：默认映射 codes.Internal（HTTP 500），
// 兜底堆栈捕获；已带 status 码的底层错误保留原码（status.Convert 穿透）。
func TestWrap_GenericInternal(t *testing.T) {
	err := Wrap(assert.AnError, "revoke access token")
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.ErrorIs(t, err, assert.AnError)
	assert.Contains(t, err.Error(), "revoke access token")
	assert.NotPanics(t, func() { _ = fmt.Sprintf("%+v", err) })

	kept := Wrap(status.Error(codes.NotFound, "missing"), "keep code")
	assert.Equal(t, codes.NotFound, status.Code(kept))
}

// TestWrap_EntAware 覆盖通用 Wrap 对 ent 领域错误的自动归类：ent.NotFound 不应被
// 误映射成 Internal(500) 而是 NotFound(404)、Validation 错误→InvalidArgument(400)、
// Constraint 错误→AlreadyExists(409)；纯底层错误仍落 Internal。同时验证 errors.As
// 能穿透 wrap 链还原 ent 错误类型（k8s apierrors/ent 消费方依赖该穿透）。
func TestWrap_EntAware(t *testing.T) {
	t.Run("NotFoundError maps to NotFound", func(t *testing.T) {
		nf := &ent.NotFoundError{}
		err := Wrap(nf, "get repo")
		assert.Equal(t, codes.NotFound, status.Code(err))
		assert.True(t, IsNotFound(err))
		var got *ent.NotFoundError
		assert.True(t, errors.As(err, &got))
		assert.Same(t, nf, got)
	})

	t.Run("ValidationError maps to InvalidArgument", func(t *testing.T) {
		// ValidationError 的 err 字段未导出，包外无法构造合法实例（零值 Error() 会
		// nil-deref）。用真实 ent 客户端触发：DBCache.key 有 NotEmpty 验证器，
		// SetKey("") 在写库前必返 ValidationError。
		client, err := ent.Open("sqlite3", "file:errs-test?mode=memory&cache=shared&_fk=1")
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, client.Close()) })
		_, verr := client.DBCache.Create().SetKey("").Save(context.Background())
		require.True(t, ent.IsValidationError(verr))

		err = Wrap(verr, "create repo")
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("ConstraintError maps to AlreadyExists", func(t *testing.T) {
		err := Wrap(&ent.ConstraintError{}, "create repo")
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	t.Run("plain error stays Internal", func(t *testing.T) {
		err := Wrap(errors.New("connection refused"), "open db")
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("preexisting status code preserved", func(t *testing.T) {
		err := Wrap(status.Error(codes.Unauthenticated, "unauth"), "wrap")
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

// TestWrap_K8sAware 覆盖通用 Wrap 对 k8s apierrors 错误的自动归类：apierrors.NotFound
// （k8s 资源未命中）→NotFound(404)，其余 k8s 错误（内部错误/网络层）仍落 Internal(500)——
// 保证 data 边界取 k8s 资源时用统一 Wrap 不会把"网络故障"误判成 404，也不会把
// "资源不存在"误映射成 500。errors.As 可穿透 wrap 链还原 *apierrors.StatusError。
func TestWrap_K8sAware(t *testing.T) {
	t.Run("apierrors NotFound maps to NotFound", func(t *testing.T) {
		nf := apierrors.NewNotFound(schema.GroupResource{Resource: "secrets"}, "my-secret")
		err := Wrap(nf, "get secret")
		assert.Equal(t, codes.NotFound, status.Code(err))
		assert.True(t, IsNotFound(err))
		var got *apierrors.StatusError
		assert.True(t, errors.As(err, &got))
		assert.Same(t, nf, got)
	})

	t.Run("other apierrors stays Internal", func(t *testing.T) {
		internal := apierrors.NewInternalError(errors.New("boom"))
		err := Wrap(internal, "get namespace")
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.False(t, IsNotFound(err))
	})
}

// TestIsNotFound 覆盖 IsNotFound 的 nil/非 NotFound/NotFound 分支：
// WrapNotFound 对原始错误按 NotFound 包装后应判为 NotFound；
// 对已是 status 错误的结果保留原 code，不应误判；
// 裸 ent.NotFoundError 可被穿透识别，裸 ent 非 NotFound 与裸 k8s apierrors.NotFound
// 不误判（k8s 仅在经 Wrap 映射为协议码后被识别，见 Wrap 的 k8s 归类分支）。
func TestIsNotFound(t *testing.T) {
	assert.False(t, IsNotFound(nil))
	assert.False(t, IsNotFound(status.Error(codes.Internal, "boom")))
	assert.True(t, IsNotFound(status.Error(codes.NotFound, "missing")))
	assert.True(t, IsNotFound(WrapNotFound(assert.AnError, "query project")))
	assert.False(t, IsNotFound(WrapNotFound(status.Error(codes.Internal, "underlying"), "query project")))
	// 裸 ent 错误穿透识别：作为 ent.IsNotFound 的上层等价物
	assert.True(t, IsNotFound(&ent.NotFoundError{}))
	// 裸 ent 非 NotFound 错误不误判（errors.As 不匹配 NotFoundError 接口；
	// 不用零值 ValidationError——其 Error() 会 nil-deref）
	assert.False(t, IsNotFound(&ent.ConstraintError{}))
	// 裸 k8s apierrors.NotFound 仅在 Wrap 映射为协议码后识别（此处为未映射状态）
	assert.False(t, IsNotFound(apierrors.NewNotFound(schema.GroupResource{Resource: "secrets"}, "my-secret")))
}

// TestGrpcStatusError_FormatFallback 直接构造 wrapped 为非 fmt.Formatter 的错误，
// 覆盖 Format 的 fmt.Fprint 兜底分支（pkg/errors.Wrap 的 withStack 恒实现 Formatter，
// 正常路径走代理，兜底仅防手工构造）。
func TestGrpcStatusError_FormatFallback(t *testing.T) {
	g := &grpcStatusError{
		st:      status.New(codes.NotFound, "not found"),
		wrapped: errors.New("plain error"),
	}
	assert.Equal(t, "plain error", fmt.Sprintf("%v", g))
}
