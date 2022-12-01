// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.17.3
// source: token/token.proto

package token

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

// AccessTokenClient is the client API for AccessToken service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AccessTokenClient interface {
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
	Grant(ctx context.Context, in *GrantRequest, opts ...grpc.CallOption) (*GrantResponse, error)
	Lease(ctx context.Context, in *LeaseRequest, opts ...grpc.CallOption) (*LeaseResponse, error)
	Revoke(ctx context.Context, in *RevokeRequest, opts ...grpc.CallOption) (*RevokeResponse, error)
}

type accessTokenClient struct {
	cc grpc.ClientConnInterface
}

func NewAccessTokenClient(cc grpc.ClientConnInterface) AccessTokenClient {
	return &accessTokenClient{cc}
}

func (c *accessTokenClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, "/token.AccessToken/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accessTokenClient) Grant(ctx context.Context, in *GrantRequest, opts ...grpc.CallOption) (*GrantResponse, error) {
	out := new(GrantResponse)
	err := c.cc.Invoke(ctx, "/token.AccessToken/Grant", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accessTokenClient) Lease(ctx context.Context, in *LeaseRequest, opts ...grpc.CallOption) (*LeaseResponse, error) {
	out := new(LeaseResponse)
	err := c.cc.Invoke(ctx, "/token.AccessToken/Lease", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accessTokenClient) Revoke(ctx context.Context, in *RevokeRequest, opts ...grpc.CallOption) (*RevokeResponse, error) {
	out := new(RevokeResponse)
	err := c.cc.Invoke(ctx, "/token.AccessToken/Revoke", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AccessTokenServer is the server API for AccessToken service.
// All implementations must embed UnimplementedAccessTokenServer
// for forward compatibility
type AccessTokenServer interface {
	List(context.Context, *ListRequest) (*ListResponse, error)
	Grant(context.Context, *GrantRequest) (*GrantResponse, error)
	Lease(context.Context, *LeaseRequest) (*LeaseResponse, error)
	Revoke(context.Context, *RevokeRequest) (*RevokeResponse, error)
	mustEmbedUnimplementedAccessTokenServer()
}

// UnimplementedAccessTokenServer must be embedded to have forward compatible implementations.
type UnimplementedAccessTokenServer struct {
}

func (UnimplementedAccessTokenServer) List(context.Context, *ListRequest) (*ListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedAccessTokenServer) Grant(context.Context, *GrantRequest) (*GrantResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Grant not implemented")
}
func (UnimplementedAccessTokenServer) Lease(context.Context, *LeaseRequest) (*LeaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Lease not implemented")
}
func (UnimplementedAccessTokenServer) Revoke(context.Context, *RevokeRequest) (*RevokeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Revoke not implemented")
}
func (UnimplementedAccessTokenServer) mustEmbedUnimplementedAccessTokenServer() {}

// UnsafeAccessTokenServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AccessTokenServer will
// result in compilation errors.
type UnsafeAccessTokenServer interface {
	mustEmbedUnimplementedAccessTokenServer()
}

func RegisterAccessTokenServer(s grpc.ServiceRegistrar, srv AccessTokenServer) {
	s.RegisterService(&AccessToken_ServiceDesc, srv)
}

func _AccessToken_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccessTokenServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/token.AccessToken/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccessTokenServer).List(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccessToken_Grant_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GrantRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccessTokenServer).Grant(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/token.AccessToken/Grant",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccessTokenServer).Grant(ctx, req.(*GrantRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccessToken_Lease_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LeaseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccessTokenServer).Lease(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/token.AccessToken/Lease",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccessTokenServer).Lease(ctx, req.(*LeaseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccessToken_Revoke_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RevokeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccessTokenServer).Revoke(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/token.AccessToken/Revoke",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccessTokenServer).Revoke(ctx, req.(*RevokeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AccessToken_ServiceDesc is the grpc.ServiceDesc for AccessToken service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AccessToken_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "token.AccessToken",
	HandlerType: (*AccessTokenServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _AccessToken_List_Handler,
		},
		{
			MethodName: "Grant",
			Handler:    _AccessToken_Grant_Handler,
		},
		{
			MethodName: "Lease",
			Handler:    _AccessToken_Lease_Handler,
		},
		{
			MethodName: "Revoke",
			Handler:    _AccessToken_Revoke_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "token/token.proto",
}
