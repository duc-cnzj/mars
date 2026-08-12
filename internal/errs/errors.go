// Package errs 是 mars 应用领域错误词汇与协议码映射的单一事实来源。
//
// 它统一实现"底层错误 → gRPC/HTTP 状态码"的映射，取代原先由 internal/biz 承担
// 的协议映射职责（biz 只表达业务语义，不再触碰 grpc status/codes）：
//   - data 层在 repo 出口用 Wrap 构造器包裹不确定错误（查询/更新/外部 API 调用返回的错误
//     可能是"记录不存在"也可能是"DB 断开/网络抖动"），由 Wrap 按底层错误实际类型自动归类；
//     Wrap 自动识别 ent 错误（NotFound/Validation/Constraint）与 k8s apierrors 错误（NotFound），
//     避免 ORM/k8s 错误被误映射成 500 或把网络故障误判成 404；
//   - biz 层用语义构造器（WrapUnauthenticated/WrapInvalidArgument）包裹已确定语义的错误、
//     用消息构造器（NotFound/InvalidArgument/Unauthenticated/...）表达业务错误；
//   - services/transport 层透传错误，用 IsNotFound/ErrorPermissionDenied 判定协议语义；
//
// 各层均不触碰字面 HTTP/gRPC 码，协议码只在构造器内部决定。本包仅依赖 gRPC
// status/codes、pkg/errors、ent 与 k8s apimachinery（用于把 ORM/k8s 领域错误自动归类为对应协议码）。
package errs

import (
	"fmt"

	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// grpcStatusError 是"携带 gRPC 状态码的包裹错误"：GRPCStatus() 让 grpc-go 在传输层
// 直接拿到协议码；wrapped 保留 pkg/errors 的 wrap 链与堆栈供最上层日志打印。
type grpcStatusError struct {
	st      *status.Status
	wrapped error
}

// Error 返回 wrapped 链的错误信息（含上下文 msg）。
func (g *grpcStatusError) Error() string { return g.wrapped.Error() }

// GRPCStatus 返回固定的协议状态，传输层据此映射 http/grpc 码。
func (g *grpcStatusError) GRPCStatus() *status.Status { return g.st }

// Unwrap 暴露 wrap 链，供 errors.Is/As 与 status.FromError 穿透。
func (g *grpcStatusError) Unwrap() error { return g.wrapped }

// Format 透传 pkg/errors 的堆栈格式化（%+v 打起源栈），否则回退 Error()。
func (g *grpcStatusError) Format(s fmt.State, verb rune) {
	if formatter, ok := g.wrapped.(fmt.Formatter); ok {
		formatter.Format(s, verb)
		return
	}
	fmt.Fprint(s, g.wrapped.Error())
}

// wrapErr 是语义构造器（WrapNotFound/WrapInvalidArgument/WrapUnauthenticated）的公共实现：
// 协议码（codes.Code）由本包在构造器内决定，调用方只表达领域语义，不触碰字面 HTTP code。
// 若底层 err 已是 gRPC status 错误则保留其码（status.Convert 穿透），否则映射为入参 code；
// msg 通过 errors.Wrap 进入 wrap 链供最上层日志打印，客户端可见 message 仍是底层 err.Error()。
func wrapErr(code codes.Code, err error, msg string) error {
	if err == nil {
		return nil
	}
	st := status.Convert(err)
	if st.Code() == codes.Unknown || st.Code() == codes.OK {
		st = status.New(code, err.Error())
	}
	return &grpcStatusError{st: st, wrapped: errors.Wrap(err, msg)}
}

// WrapNotFound 构造"记录不存在"领域错误（gRPC NotFound / HTTP 404），是 Wrap 自动归类
// 的 NotFound 分支（ent.NotFound / k8s apierrors.NotFound 命中后经其包裹）；确定语义为
// "记录不存在"的场景也可直接使用，携带中文业务上下文（如 "query access token"）。
func WrapNotFound(err error, msg string) error {
	return wrapErr(codes.NotFound, err, msg)
}

// WrapInvalidArgument 构造"参数不合法"领域错误（gRPC InvalidArgument / HTTP 400）。
// 供业务校验类（名称已存在、非法路径等）在边界处包裹返回。
func WrapInvalidArgument(err error, msg string) error {
	return wrapErr(codes.InvalidArgument, err, msg)
}

// WrapUnauthenticated 构造"未认证"领域错误（gRPC Unauthenticated / HTTP 401）。
// 供认证/令牌校验类在边界处包裹返回。
func WrapUnauthenticated(err error, msg string) error {
	return wrapErr(codes.Unauthenticated, err, msg)
}

// Wrap 是 data 层边界统一使用的"自动归类"构造器：按底层错误实际类型推导协议码，
// 而非由调用方硬编码语义——这是处理不确定错误（查询/更新/外部 API 调用的返回错误
// 可能是"记录不存在"也可能是"DB 连接断开/网络抖动"）的正确姿势。
//
// 归类规则（按序，已带 status 码的错误仍保留原码）：
//   - ent.IsNotFound / k8s apierrors.IsNotFound → NotFound(404)：查询/取 k8s 资源未命中
//   - ent.IsValidationError → InvalidArgument(400)：ORM 校验失败
//   - ent.IsConstraintError → AlreadyExists(409)：唯一约束冲突
//   - 其余（DB 连接异常、k8s API 调用失败、上传存储故障等）→ Internal(500)
//
// 调用方只表达领域上下文（msg 经 pkg/errors.Wrap 进链供日志打印，客户端可见 message
// 仍是底层 err.Error()），协议码映射收口本包，data 层不再按操作意图硬编码 NotFound/500。
func Wrap(err error, msg string) error {
	switch {
	case ent.IsNotFound(err):
		return WrapNotFound(err, msg)
	case apierrors.IsNotFound(err):
		return WrapNotFound(err, msg)
	case ent.IsValidationError(err):
		return WrapInvalidArgument(err, msg)
	case ent.IsConstraintError(err):
		return wrapErr(codes.AlreadyExists, err, msg)
	default:
		return wrapErr(codes.Internal, err, msg)
	}
}

// NotFound 构造"记录不存在"状态错误（gRPC NotFound / HTTP 404），直接携带消息，
// 无底层错误可包裹。biz 层业务判断（容器/日志/指标不存在）用它表达领域语义。
func NotFound(msg string) error { return status.Error(codes.NotFound, msg) }

// InvalidArgument 构造"参数不合法"状态错误（gRPC InvalidArgument / HTTP 400），
// 直接携带消息，无底层错误可包裹。biz 层参数校验失败用它表达领域语义。
func InvalidArgument(msg string) error { return status.Error(codes.InvalidArgument, msg) }

// Unauthenticated 构造"未认证"状态错误（gRPC Unauthenticated / HTTP 401），
// 直接携带消息，无底层错误可包裹。biz 层登录/令牌校验失败用它表达领域语义。
func Unauthenticated(msg string) error { return status.Error(codes.Unauthenticated, msg) }

// PermissionDenied 构造"无权限"状态错误（gRPC PermissionDenied / HTTP 403），
// 直接携带消息，无底层错误可包裹。用于领域级权限拒绝。
func PermissionDenied(msg string) error { return status.Error(codes.PermissionDenied, msg) }

// IsNotFound 判断错误是否为"未找到"，作为 ent.IsNotFound 的上层等价物：
// 既能识别已映射为 NotFound 协议码的错误（data 边界 errs.Wrap 后的 grpcStatusError），
// 也能穿透识别尚未映射的裸 ent.NotFoundError，避免 transport/data 层直接依赖
// ent ORM 的错误类型（k8s apierrors.NotFound 仅在经 Wrap 映射为协议码后被识别）。
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return ent.IsNotFound(err) || status.Code(err) == codes.NotFound
}

// ErrorPermissionDenied 领域级权限拒绝 sentinel（gRPC PermissionDenied）。
// 访问谓词与传输层共用同一错误值，保证 errors.Is/assert.ErrorIs 可匹配。
var ErrorPermissionDenied = PermissionDenied("没有权限执行该操作")

// AlreadyExists 构造"资源已存在"领域错误（gRPC AlreadyExists / HTTP 409）。
// 供 biz 用例（命名空间已存在/Terminating）直接携带状态语义返回，transport
// 透传即可拿到 409——协议映射收口本包，传输层不再散落 status.Error 构造。
func AlreadyExists(msg string) error {
	return status.Error(codes.AlreadyExists, msg)
}
