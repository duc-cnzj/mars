// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: container/container.proto

package container

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Container_CopyToPod_FullMethodName          = "/container.Container/CopyToPod"
	Container_Exec_FullMethodName               = "/container.Container/Exec"
	Container_ExecOnce_FullMethodName           = "/container.Container/ExecOnce"
	Container_StreamCopyToPod_FullMethodName    = "/container.Container/StreamCopyToPod"
	Container_IsPodRunning_FullMethodName       = "/container.Container/IsPodRunning"
	Container_IsPodExists_FullMethodName        = "/container.Container/IsPodExists"
	Container_ContainerLog_FullMethodName       = "/container.Container/ContainerLog"
	Container_StreamContainerLog_FullMethodName = "/container.Container/StreamContainerLog"
)

// ContainerClient is the client API for Container service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ContainerClient interface {
	CopyToPod(ctx context.Context, in *CopyToPodRequest, opts ...grpc.CallOption) (*CopyToPodResponse, error)
	// Exec grpc 执行 pod 命令，交互式, 使用方法见 examples/ 目录
	Exec(ctx context.Context, opts ...grpc.CallOption) (Container_ExecClient, error)
	// ExecOnce grpc 执行一次 pod 命令, 非 tty 模式, 适合一次性脚本, 使用方法见 examples/ 目录
	ExecOnce(ctx context.Context, in *ExecOnceRequest, opts ...grpc.CallOption) (Container_ExecOnceClient, error)
	// StreamCopyToPod grpc 上传文件到 pod, 使用方法见 examples/ 目录
	StreamCopyToPod(ctx context.Context, opts ...grpc.CallOption) (Container_StreamCopyToPodClient, error)
	IsPodRunning(ctx context.Context, in *IsPodRunningRequest, opts ...grpc.CallOption) (*IsPodRunningResponse, error)
	IsPodExists(ctx context.Context, in *IsPodExistsRequest, opts ...grpc.CallOption) (*IsPodExistsResponse, error)
	// ContainerLog 查看 pod 日志
	ContainerLog(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (*LogResponse, error)
	StreamContainerLog(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (Container_StreamContainerLogClient, error)
}

type containerClient struct {
	cc grpc.ClientConnInterface
}

func NewContainerClient(cc grpc.ClientConnInterface) ContainerClient {
	return &containerClient{cc}
}

func (c *containerClient) CopyToPod(ctx context.Context, in *CopyToPodRequest, opts ...grpc.CallOption) (*CopyToPodResponse, error) {
	out := new(CopyToPodResponse)
	err := c.cc.Invoke(ctx, Container_CopyToPod_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *containerClient) Exec(ctx context.Context, opts ...grpc.CallOption) (Container_ExecClient, error) {
	stream, err := c.cc.NewStream(ctx, &Container_ServiceDesc.Streams[0], Container_Exec_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &containerExecClient{stream}
	return x, nil
}

type Container_ExecClient interface {
	Send(*ExecRequest) error
	Recv() (*ExecResponse, error)
	grpc.ClientStream
}

type containerExecClient struct {
	grpc.ClientStream
}

func (x *containerExecClient) Send(m *ExecRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *containerExecClient) Recv() (*ExecResponse, error) {
	m := new(ExecResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *containerClient) ExecOnce(ctx context.Context, in *ExecOnceRequest, opts ...grpc.CallOption) (Container_ExecOnceClient, error) {
	stream, err := c.cc.NewStream(ctx, &Container_ServiceDesc.Streams[1], Container_ExecOnce_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &containerExecOnceClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Container_ExecOnceClient interface {
	Recv() (*ExecResponse, error)
	grpc.ClientStream
}

type containerExecOnceClient struct {
	grpc.ClientStream
}

func (x *containerExecOnceClient) Recv() (*ExecResponse, error) {
	m := new(ExecResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *containerClient) StreamCopyToPod(ctx context.Context, opts ...grpc.CallOption) (Container_StreamCopyToPodClient, error) {
	stream, err := c.cc.NewStream(ctx, &Container_ServiceDesc.Streams[2], Container_StreamCopyToPod_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &containerStreamCopyToPodClient{stream}
	return x, nil
}

type Container_StreamCopyToPodClient interface {
	Send(*StreamCopyToPodRequest) error
	CloseAndRecv() (*StreamCopyToPodResponse, error)
	grpc.ClientStream
}

type containerStreamCopyToPodClient struct {
	grpc.ClientStream
}

func (x *containerStreamCopyToPodClient) Send(m *StreamCopyToPodRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *containerStreamCopyToPodClient) CloseAndRecv() (*StreamCopyToPodResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(StreamCopyToPodResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *containerClient) IsPodRunning(ctx context.Context, in *IsPodRunningRequest, opts ...grpc.CallOption) (*IsPodRunningResponse, error) {
	out := new(IsPodRunningResponse)
	err := c.cc.Invoke(ctx, Container_IsPodRunning_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *containerClient) IsPodExists(ctx context.Context, in *IsPodExistsRequest, opts ...grpc.CallOption) (*IsPodExistsResponse, error) {
	out := new(IsPodExistsResponse)
	err := c.cc.Invoke(ctx, Container_IsPodExists_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *containerClient) ContainerLog(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (*LogResponse, error) {
	out := new(LogResponse)
	err := c.cc.Invoke(ctx, Container_ContainerLog_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *containerClient) StreamContainerLog(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (Container_StreamContainerLogClient, error) {
	stream, err := c.cc.NewStream(ctx, &Container_ServiceDesc.Streams[3], Container_StreamContainerLog_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &containerStreamContainerLogClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Container_StreamContainerLogClient interface {
	Recv() (*LogResponse, error)
	grpc.ClientStream
}

type containerStreamContainerLogClient struct {
	grpc.ClientStream
}

func (x *containerStreamContainerLogClient) Recv() (*LogResponse, error) {
	m := new(LogResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ContainerServer is the server API for Container service.
// All implementations must embed UnimplementedContainerServer
// for forward compatibility
type ContainerServer interface {
	CopyToPod(context.Context, *CopyToPodRequest) (*CopyToPodResponse, error)
	// Exec grpc 执行 pod 命令，交互式, 使用方法见 examples/ 目录
	Exec(Container_ExecServer) error
	// ExecOnce grpc 执行一次 pod 命令, 非 tty 模式, 适合一次性脚本, 使用方法见 examples/ 目录
	ExecOnce(*ExecOnceRequest, Container_ExecOnceServer) error
	// StreamCopyToPod grpc 上传文件到 pod, 使用方法见 examples/ 目录
	StreamCopyToPod(Container_StreamCopyToPodServer) error
	IsPodRunning(context.Context, *IsPodRunningRequest) (*IsPodRunningResponse, error)
	IsPodExists(context.Context, *IsPodExistsRequest) (*IsPodExistsResponse, error)
	// ContainerLog 查看 pod 日志
	ContainerLog(context.Context, *LogRequest) (*LogResponse, error)
	StreamContainerLog(*LogRequest, Container_StreamContainerLogServer) error
	mustEmbedUnimplementedContainerServer()
}

// UnimplementedContainerServer must be embedded to have forward compatible implementations.
type UnimplementedContainerServer struct {
}

func (UnimplementedContainerServer) CopyToPod(context.Context, *CopyToPodRequest) (*CopyToPodResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CopyToPod not implemented")
}
func (UnimplementedContainerServer) Exec(Container_ExecServer) error {
	return status.Errorf(codes.Unimplemented, "method Exec not implemented")
}
func (UnimplementedContainerServer) ExecOnce(*ExecOnceRequest, Container_ExecOnceServer) error {
	return status.Errorf(codes.Unimplemented, "method ExecOnce not implemented")
}
func (UnimplementedContainerServer) StreamCopyToPod(Container_StreamCopyToPodServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamCopyToPod not implemented")
}
func (UnimplementedContainerServer) IsPodRunning(context.Context, *IsPodRunningRequest) (*IsPodRunningResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsPodRunning not implemented")
}
func (UnimplementedContainerServer) IsPodExists(context.Context, *IsPodExistsRequest) (*IsPodExistsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsPodExists not implemented")
}
func (UnimplementedContainerServer) ContainerLog(context.Context, *LogRequest) (*LogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ContainerLog not implemented")
}
func (UnimplementedContainerServer) StreamContainerLog(*LogRequest, Container_StreamContainerLogServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamContainerLog not implemented")
}
func (UnimplementedContainerServer) mustEmbedUnimplementedContainerServer() {}

// UnsafeContainerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ContainerServer will
// result in compilation errors.
type UnsafeContainerServer interface {
	mustEmbedUnimplementedContainerServer()
}

func RegisterContainerServer(s grpc.ServiceRegistrar, srv ContainerServer) {
	s.RegisterService(&Container_ServiceDesc, srv)
}

func _Container_CopyToPod_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CopyToPodRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContainerServer).CopyToPod(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Container_CopyToPod_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContainerServer).CopyToPod(ctx, req.(*CopyToPodRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Container_Exec_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ContainerServer).Exec(&containerExecServer{stream})
}

type Container_ExecServer interface {
	Send(*ExecResponse) error
	Recv() (*ExecRequest, error)
	grpc.ServerStream
}

type containerExecServer struct {
	grpc.ServerStream
}

func (x *containerExecServer) Send(m *ExecResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *containerExecServer) Recv() (*ExecRequest, error) {
	m := new(ExecRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Container_ExecOnce_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ExecOnceRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ContainerServer).ExecOnce(m, &containerExecOnceServer{stream})
}

type Container_ExecOnceServer interface {
	Send(*ExecResponse) error
	grpc.ServerStream
}

type containerExecOnceServer struct {
	grpc.ServerStream
}

func (x *containerExecOnceServer) Send(m *ExecResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Container_StreamCopyToPod_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ContainerServer).StreamCopyToPod(&containerStreamCopyToPodServer{stream})
}

type Container_StreamCopyToPodServer interface {
	SendAndClose(*StreamCopyToPodResponse) error
	Recv() (*StreamCopyToPodRequest, error)
	grpc.ServerStream
}

type containerStreamCopyToPodServer struct {
	grpc.ServerStream
}

func (x *containerStreamCopyToPodServer) SendAndClose(m *StreamCopyToPodResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *containerStreamCopyToPodServer) Recv() (*StreamCopyToPodRequest, error) {
	m := new(StreamCopyToPodRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Container_IsPodRunning_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsPodRunningRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContainerServer).IsPodRunning(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Container_IsPodRunning_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContainerServer).IsPodRunning(ctx, req.(*IsPodRunningRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Container_IsPodExists_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsPodExistsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContainerServer).IsPodExists(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Container_IsPodExists_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContainerServer).IsPodExists(ctx, req.(*IsPodExistsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Container_ContainerLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContainerServer).ContainerLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Container_ContainerLog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContainerServer).ContainerLog(ctx, req.(*LogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Container_StreamContainerLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(LogRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ContainerServer).StreamContainerLog(m, &containerStreamContainerLogServer{stream})
}

type Container_StreamContainerLogServer interface {
	Send(*LogResponse) error
	grpc.ServerStream
}

type containerStreamContainerLogServer struct {
	grpc.ServerStream
}

func (x *containerStreamContainerLogServer) Send(m *LogResponse) error {
	return x.ServerStream.SendMsg(m)
}

// Container_ServiceDesc is the grpc.ServiceDesc for Container service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Container_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "container.Container",
	HandlerType: (*ContainerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CopyToPod",
			Handler:    _Container_CopyToPod_Handler,
		},
		{
			MethodName: "IsPodRunning",
			Handler:    _Container_IsPodRunning_Handler,
		},
		{
			MethodName: "IsPodExists",
			Handler:    _Container_IsPodExists_Handler,
		},
		{
			MethodName: "ContainerLog",
			Handler:    _Container_ContainerLog_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Exec",
			Handler:       _Container_Exec_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "ExecOnce",
			Handler:       _Container_ExecOnce_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "StreamCopyToPod",
			Handler:       _Container_StreamCopyToPod_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "StreamContainerLog",
			Handler:       _Container_StreamContainerLog_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "container/container.proto",
}
