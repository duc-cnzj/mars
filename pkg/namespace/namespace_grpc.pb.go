// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package namespace

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// NamespaceClient is the client API for Namespace service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NamespaceClient interface {
	Index(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*NamespaceList, error)
	Store(ctx context.Context, in *NsStoreRequest, opts ...grpc.CallOption) (*NsStoreResponse, error)
	CpuAndMemory(ctx context.Context, in *NamespaceID, opts ...grpc.CallOption) (*CpuAndMemoryResponse, error)
	ServiceEndpoints(ctx context.Context, in *ServiceEndpointsRequest, opts ...grpc.CallOption) (*ServiceEndpointsResponse, error)
	Destroy(ctx context.Context, in *NamespaceID, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type namespaceClient struct {
	cc grpc.ClientConnInterface
}

func NewNamespaceClient(cc grpc.ClientConnInterface) NamespaceClient {
	return &namespaceClient{cc}
}

func (c *namespaceClient) Index(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*NamespaceList, error) {
	out := new(NamespaceList)
	err := c.cc.Invoke(ctx, "/Namespace/Index", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) Store(ctx context.Context, in *NsStoreRequest, opts ...grpc.CallOption) (*NsStoreResponse, error) {
	out := new(NsStoreResponse)
	err := c.cc.Invoke(ctx, "/Namespace/Store", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) CpuAndMemory(ctx context.Context, in *NamespaceID, opts ...grpc.CallOption) (*CpuAndMemoryResponse, error) {
	out := new(CpuAndMemoryResponse)
	err := c.cc.Invoke(ctx, "/Namespace/CpuAndMemory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) ServiceEndpoints(ctx context.Context, in *ServiceEndpointsRequest, opts ...grpc.CallOption) (*ServiceEndpointsResponse, error) {
	out := new(ServiceEndpointsResponse)
	err := c.cc.Invoke(ctx, "/Namespace/ServiceEndpoints", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) Destroy(ctx context.Context, in *NamespaceID, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/Namespace/Destroy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NamespaceServer is the server API for Namespace service.
// All implementations must embed UnimplementedNamespaceServer
// for forward compatibility
type NamespaceServer interface {
	Index(context.Context, *emptypb.Empty) (*NamespaceList, error)
	Store(context.Context, *NsStoreRequest) (*NsStoreResponse, error)
	CpuAndMemory(context.Context, *NamespaceID) (*CpuAndMemoryResponse, error)
	ServiceEndpoints(context.Context, *ServiceEndpointsRequest) (*ServiceEndpointsResponse, error)
	Destroy(context.Context, *NamespaceID) (*emptypb.Empty, error)
	mustEmbedUnimplementedNamespaceServer()
}

// UnimplementedNamespaceServer must be embedded to have forward compatible implementations.
type UnimplementedNamespaceServer struct {
}

func (UnimplementedNamespaceServer) Index(context.Context, *emptypb.Empty) (*NamespaceList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Index not implemented")
}
func (UnimplementedNamespaceServer) Store(context.Context, *NsStoreRequest) (*NsStoreResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Store not implemented")
}
func (UnimplementedNamespaceServer) CpuAndMemory(context.Context, *NamespaceID) (*CpuAndMemoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CpuAndMemory not implemented")
}
func (UnimplementedNamespaceServer) ServiceEndpoints(context.Context, *ServiceEndpointsRequest) (*ServiceEndpointsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ServiceEndpoints not implemented")
}
func (UnimplementedNamespaceServer) Destroy(context.Context, *NamespaceID) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Destroy not implemented")
}
func (UnimplementedNamespaceServer) mustEmbedUnimplementedNamespaceServer() {}

// UnsafeNamespaceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NamespaceServer will
// result in compilation errors.
type UnsafeNamespaceServer interface {
	mustEmbedUnimplementedNamespaceServer()
}

func RegisterNamespaceServer(s grpc.ServiceRegistrar, srv NamespaceServer) {
	s.RegisterService(&Namespace_ServiceDesc, srv)
}

func _Namespace_Index_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).Index(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Namespace/Index",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).Index(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_Store_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NsStoreRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).Store(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Namespace/Store",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).Store(ctx, req.(*NsStoreRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_CpuAndMemory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NamespaceID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).CpuAndMemory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Namespace/CpuAndMemory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).CpuAndMemory(ctx, req.(*NamespaceID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_ServiceEndpoints_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServiceEndpointsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).ServiceEndpoints(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Namespace/ServiceEndpoints",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).ServiceEndpoints(ctx, req.(*ServiceEndpointsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_Destroy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NamespaceID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).Destroy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Namespace/Destroy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).Destroy(ctx, req.(*NamespaceID))
	}
	return interceptor(ctx, in, info, handler)
}

// Namespace_ServiceDesc is the grpc.ServiceDesc for Namespace service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Namespace_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Namespace",
	HandlerType: (*NamespaceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Index",
			Handler:    _Namespace_Index_Handler,
		},
		{
			MethodName: "Store",
			Handler:    _Namespace_Store_Handler,
		},
		{
			MethodName: "CpuAndMemory",
			Handler:    _Namespace_CpuAndMemory_Handler,
		},
		{
			MethodName: "ServiceEndpoints",
			Handler:    _Namespace_ServiceEndpoints_Handler,
		},
		{
			MethodName: "Destroy",
			Handler:    _Namespace_Destroy_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "namespace/namespace.proto",
}