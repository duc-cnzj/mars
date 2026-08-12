package transport

import (
	"context"
	"reflect"

	"google.golang.org/protobuf/proto"
)

// Stream 是 server-streaming 方法的类型安全接收器，语义与 gRPC SDK 对齐：
// Recv 阻塞等待下一条消息，io.EOF 表示流正常结束，其余错误即流级错误。
// T 是指针类型（如 *container.LogResponse），Recv 返回每次新实例化的 T。
type Stream[T proto.Message] interface {
	Recv() (T, error)
	Close() error
}

// RawStream 是非泛型底层流，由 Conn.OpenStream 返回，供泛型工厂包装。
type RawStream interface {
	// Recv 把下一条消息反序列化进 out；io.EOF 表示流结束。
	Recv(proto.Message) error
	Close() error
}

// OpenStream 通过 conn 打开一条 server-streaming 流并包装成类型安全的 Stream[T]。
// 生成器对带 google.api.http 注解的 server-streaming 方法调用本工厂。
func OpenStream[T proto.Message](conn Conn, ctx context.Context, method, path string, req proto.Message) (Stream[T], error) {
	raw, err := conn.OpenStream(ctx, method, path, req)
	if err != nil {
		return nil, err
	}
	return &typedStream[T]{raw: raw}, nil
}

type typedStream[T proto.Message] struct {
	raw RawStream
}

// Recv 阻塞读取下一条消息并反序列化为类型 T；io.EOF 表示流正常结束，其余错误即流级错误。
func (s *typedStream[T]) Recv() (T, error) {
	var zero T
	msg := reflect.New(reflect.TypeOf(zero).Elem()).Interface().(proto.Message)
	if err := s.raw.Recv(msg); err != nil {
		return zero, err
	}
	return msg.(T), nil
}

// Close 关闭底层原始流，释放连接资源。
func (s *typedStream[T]) Close() error { return s.raw.Close() }
