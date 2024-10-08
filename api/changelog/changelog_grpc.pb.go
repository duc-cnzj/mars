// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: changelog/changelog.proto

package changelog

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
	Changelog_FindLastChangelogsByProjectID_FullMethodName = "/changelog.Changelog/FindLastChangelogsByProjectID"
)

// ChangelogClient is the client API for Changelog service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChangelogClient interface {
	FindLastChangelogsByProjectID(ctx context.Context, in *FindLastChangelogsByProjectIDRequest, opts ...grpc.CallOption) (*FindLastChangelogsByProjectIDResponse, error)
}

type changelogClient struct {
	cc grpc.ClientConnInterface
}

func NewChangelogClient(cc grpc.ClientConnInterface) ChangelogClient {
	return &changelogClient{cc}
}

func (c *changelogClient) FindLastChangelogsByProjectID(ctx context.Context, in *FindLastChangelogsByProjectIDRequest, opts ...grpc.CallOption) (*FindLastChangelogsByProjectIDResponse, error) {
	out := new(FindLastChangelogsByProjectIDResponse)
	err := c.cc.Invoke(ctx, Changelog_FindLastChangelogsByProjectID_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChangelogServer is the server API for Changelog service.
// All implementations must embed UnimplementedChangelogServer
// for forward compatibility
type ChangelogServer interface {
	FindLastChangelogsByProjectID(context.Context, *FindLastChangelogsByProjectIDRequest) (*FindLastChangelogsByProjectIDResponse, error)
	mustEmbedUnimplementedChangelogServer()
}

// UnimplementedChangelogServer must be embedded to have forward compatible implementations.
type UnimplementedChangelogServer struct {
}

func (UnimplementedChangelogServer) FindLastChangelogsByProjectID(context.Context, *FindLastChangelogsByProjectIDRequest) (*FindLastChangelogsByProjectIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindLastChangelogsByProjectID not implemented")
}
func (UnimplementedChangelogServer) mustEmbedUnimplementedChangelogServer() {}

// UnsafeChangelogServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChangelogServer will
// result in compilation errors.
type UnsafeChangelogServer interface {
	mustEmbedUnimplementedChangelogServer()
}

func RegisterChangelogServer(s grpc.ServiceRegistrar, srv ChangelogServer) {
	s.RegisterService(&Changelog_ServiceDesc, srv)
}

func _Changelog_FindLastChangelogsByProjectID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FindLastChangelogsByProjectIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChangelogServer).FindLastChangelogsByProjectID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Changelog_FindLastChangelogsByProjectID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChangelogServer).FindLastChangelogsByProjectID(ctx, req.(*FindLastChangelogsByProjectIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Changelog_ServiceDesc is the grpc.ServiceDesc for Changelog service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Changelog_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "changelog.Changelog",
	HandlerType: (*ChangelogServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FindLastChangelogsByProjectID",
			Handler:    _Changelog_FindLastChangelogsByProjectID_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "changelog/changelog.proto",
}
