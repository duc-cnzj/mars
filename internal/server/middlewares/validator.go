package middlewares

import (
	"context"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	"google.golang.org/grpc"
)

// Validator 是可选的消息校验能力：请求（Unary 的 req 或 Stream 的 RecvMsg 消息）
// 实现它时，拦截器会在进业务逻辑前调用 Validate，失败即拦截。
type Validator interface {
	Validate() error
}

// ValidatorUnaryServerInterceptor 是 Unary 校验拦截器：req 实现 Validator 时先校验，
// 失败返回 InvalidArgument，通过才透传 handler。
func ValidatorUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if validator, ok := req.(Validator); ok {
			if err := validator.Validate(); err != nil {
				return nil, errs.InvalidArgument(err.Error())
			}
		}

		return handler(ctx, req)
	}
}

// ValidatorStreamServerInterceptor 是 Stream 校验拦截器：用 recvWrapper 包住流，
// 让 RecvMsg 时对收到的每条消息做校验。
func ValidatorStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &recvWrapper{stream}
		return handler(srv, wrapper)
	}
}

type recvWrapper struct {
	grpc.ServerStream
}

// RecvMsg 覆盖 grpc.ServerStream.RecvMsg：先取原始消息，若实现 Validator 则校验，
// 失败返回错误，通过才交付给 handler。
func (s *recvWrapper) RecvMsg(m any) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	if validator, ok := m.(Validator); ok {
		if err := validator.Validate(); err != nil {
			// 生成的 *.pb.validate.go 返回的是 MultiError（非 status 错误），原样上抛会
			// 让 stream 校验失败落成 Unknown，与 unary 路径的 InvalidArgument 不一致——
			// 两条路径统一经 errs 构造器映射，保证同一条校验规则的错误码相同。
			return errs.InvalidArgument(err.Error())
		}
	}

	return nil
}
