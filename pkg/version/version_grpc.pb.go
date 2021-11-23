// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package version

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

// VersionClient is the client API for Version service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VersionClient interface {
	Get(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*VersionResponse, error)
}

type versionClient struct {
	cc grpc.ClientConnInterface
}

func NewVersionClient(cc grpc.ClientConnInterface) VersionClient {
	return &versionClient{cc}
}

func (c *versionClient) Get(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*VersionResponse, error) {
	out := new(VersionResponse)
	err := c.cc.Invoke(ctx, "/Version/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VersionServer is the server API for Version service.
// All implementations must embed UnimplementedVersionServer
// for forward compatibility
type VersionServer interface {
	Get(context.Context, *emptypb.Empty) (*VersionResponse, error)
	mustEmbedUnimplementedVersionServer()
}

// UnimplementedVersionServer must be embedded to have forward compatible implementations.
type UnimplementedVersionServer struct {
}

func (UnimplementedVersionServer) Get(context.Context, *emptypb.Empty) (*VersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedVersionServer) mustEmbedUnimplementedVersionServer() {}

// UnsafeVersionServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VersionServer will
// result in compilation errors.
type UnsafeVersionServer interface {
	mustEmbedUnimplementedVersionServer()
}

func RegisterVersionServer(s grpc.ServiceRegistrar, srv VersionServer) {
	s.RegisterService(&Version_ServiceDesc, srv)
}

func _Version_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VersionServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Version/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VersionServer).Get(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Version_ServiceDesc is the grpc.ServiceDesc for Version service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Version_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Version",
	HandlerType: (*VersionServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _Version_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "version/version.proto",
}
