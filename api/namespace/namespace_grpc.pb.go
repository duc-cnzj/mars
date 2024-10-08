// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: namespace/namespace.proto

package namespace

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
	Namespace_List_FullMethodName          = "/namespace.Namespace/List"
	Namespace_UpdatePrivate_FullMethodName = "/namespace.Namespace/UpdatePrivate"
	Namespace_SyncMembers_FullMethodName   = "/namespace.Namespace/SyncMembers"
	Namespace_Create_FullMethodName        = "/namespace.Namespace/Create"
	Namespace_Show_FullMethodName          = "/namespace.Namespace/Show"
	Namespace_UpdateDesc_FullMethodName    = "/namespace.Namespace/UpdateDesc"
	Namespace_Delete_FullMethodName        = "/namespace.Namespace/Delete"
	Namespace_IsExists_FullMethodName      = "/namespace.Namespace/IsExists"
	Namespace_Favorite_FullMethodName      = "/namespace.Namespace/Favorite"
	Namespace_Transfer_FullMethodName      = "/namespace.Namespace/Transfer"
)

// NamespaceClient is the client API for Namespace service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NamespaceClient interface {
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
	UpdatePrivate(ctx context.Context, in *UpdatePrivateRequest, opts ...grpc.CallOption) (*UpdatePrivateResponse, error)
	SyncMembers(ctx context.Context, in *SyncMembersRequest, opts ...grpc.CallOption) (*SyncMembersResponse, error)
	Create(ctx context.Context, in *CreateRequest, opts ...grpc.CallOption) (*CreateResponse, error)
	Show(ctx context.Context, in *ShowRequest, opts ...grpc.CallOption) (*ShowResponse, error)
	UpdateDesc(ctx context.Context, in *UpdateDescRequest, opts ...grpc.CallOption) (*UpdateDescResponse, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	IsExists(ctx context.Context, in *IsExistsRequest, opts ...grpc.CallOption) (*IsExistsResponse, error)
	Favorite(ctx context.Context, in *FavoriteRequest, opts ...grpc.CallOption) (*FavoriteResponse, error)
	Transfer(ctx context.Context, in *TransferRequest, opts ...grpc.CallOption) (*TransferResponse, error)
}

type namespaceClient struct {
	cc grpc.ClientConnInterface
}

func NewNamespaceClient(cc grpc.ClientConnInterface) NamespaceClient {
	return &namespaceClient{cc}
}

func (c *namespaceClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, Namespace_List_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) UpdatePrivate(ctx context.Context, in *UpdatePrivateRequest, opts ...grpc.CallOption) (*UpdatePrivateResponse, error) {
	out := new(UpdatePrivateResponse)
	err := c.cc.Invoke(ctx, Namespace_UpdatePrivate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) SyncMembers(ctx context.Context, in *SyncMembersRequest, opts ...grpc.CallOption) (*SyncMembersResponse, error) {
	out := new(SyncMembersResponse)
	err := c.cc.Invoke(ctx, Namespace_SyncMembers_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) Create(ctx context.Context, in *CreateRequest, opts ...grpc.CallOption) (*CreateResponse, error) {
	out := new(CreateResponse)
	err := c.cc.Invoke(ctx, Namespace_Create_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) Show(ctx context.Context, in *ShowRequest, opts ...grpc.CallOption) (*ShowResponse, error) {
	out := new(ShowResponse)
	err := c.cc.Invoke(ctx, Namespace_Show_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) UpdateDesc(ctx context.Context, in *UpdateDescRequest, opts ...grpc.CallOption) (*UpdateDescResponse, error) {
	out := new(UpdateDescResponse)
	err := c.cc.Invoke(ctx, Namespace_UpdateDesc_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, Namespace_Delete_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) IsExists(ctx context.Context, in *IsExistsRequest, opts ...grpc.CallOption) (*IsExistsResponse, error) {
	out := new(IsExistsResponse)
	err := c.cc.Invoke(ctx, Namespace_IsExists_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) Favorite(ctx context.Context, in *FavoriteRequest, opts ...grpc.CallOption) (*FavoriteResponse, error) {
	out := new(FavoriteResponse)
	err := c.cc.Invoke(ctx, Namespace_Favorite_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *namespaceClient) Transfer(ctx context.Context, in *TransferRequest, opts ...grpc.CallOption) (*TransferResponse, error) {
	out := new(TransferResponse)
	err := c.cc.Invoke(ctx, Namespace_Transfer_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NamespaceServer is the server API for Namespace service.
// All implementations must embed UnimplementedNamespaceServer
// for forward compatibility
type NamespaceServer interface {
	List(context.Context, *ListRequest) (*ListResponse, error)
	UpdatePrivate(context.Context, *UpdatePrivateRequest) (*UpdatePrivateResponse, error)
	SyncMembers(context.Context, *SyncMembersRequest) (*SyncMembersResponse, error)
	Create(context.Context, *CreateRequest) (*CreateResponse, error)
	Show(context.Context, *ShowRequest) (*ShowResponse, error)
	UpdateDesc(context.Context, *UpdateDescRequest) (*UpdateDescResponse, error)
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	IsExists(context.Context, *IsExistsRequest) (*IsExistsResponse, error)
	Favorite(context.Context, *FavoriteRequest) (*FavoriteResponse, error)
	Transfer(context.Context, *TransferRequest) (*TransferResponse, error)
	mustEmbedUnimplementedNamespaceServer()
}

// UnimplementedNamespaceServer must be embedded to have forward compatible implementations.
type UnimplementedNamespaceServer struct {
}

func (UnimplementedNamespaceServer) List(context.Context, *ListRequest) (*ListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedNamespaceServer) UpdatePrivate(context.Context, *UpdatePrivateRequest) (*UpdatePrivateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePrivate not implemented")
}
func (UnimplementedNamespaceServer) SyncMembers(context.Context, *SyncMembersRequest) (*SyncMembersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SyncMembers not implemented")
}
func (UnimplementedNamespaceServer) Create(context.Context, *CreateRequest) (*CreateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedNamespaceServer) Show(context.Context, *ShowRequest) (*ShowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Show not implemented")
}
func (UnimplementedNamespaceServer) UpdateDesc(context.Context, *UpdateDescRequest) (*UpdateDescResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDesc not implemented")
}
func (UnimplementedNamespaceServer) Delete(context.Context, *DeleteRequest) (*DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedNamespaceServer) IsExists(context.Context, *IsExistsRequest) (*IsExistsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsExists not implemented")
}
func (UnimplementedNamespaceServer) Favorite(context.Context, *FavoriteRequest) (*FavoriteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Favorite not implemented")
}
func (UnimplementedNamespaceServer) Transfer(context.Context, *TransferRequest) (*TransferResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Transfer not implemented")
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

func _Namespace_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Namespace_List_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).List(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_UpdatePrivate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdatePrivateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).UpdatePrivate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Namespace_UpdatePrivate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).UpdatePrivate(ctx, req.(*UpdatePrivateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_SyncMembers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SyncMembersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).SyncMembers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Namespace_SyncMembers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).SyncMembers(ctx, req.(*SyncMembersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Namespace_Create_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).Create(ctx, req.(*CreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_Show_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShowRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).Show(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Namespace_Show_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).Show(ctx, req.(*ShowRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_UpdateDesc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateDescRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).UpdateDesc(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Namespace_UpdateDesc_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).UpdateDesc(ctx, req.(*UpdateDescRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Namespace_Delete_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).Delete(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_IsExists_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsExistsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).IsExists(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Namespace_IsExists_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).IsExists(ctx, req.(*IsExistsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_Favorite_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FavoriteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).Favorite(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Namespace_Favorite_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).Favorite(ctx, req.(*FavoriteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Namespace_Transfer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransferRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceServer).Transfer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Namespace_Transfer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceServer).Transfer(ctx, req.(*TransferRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Namespace_ServiceDesc is the grpc.ServiceDesc for Namespace service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Namespace_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "namespace.Namespace",
	HandlerType: (*NamespaceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _Namespace_List_Handler,
		},
		{
			MethodName: "UpdatePrivate",
			Handler:    _Namespace_UpdatePrivate_Handler,
		},
		{
			MethodName: "SyncMembers",
			Handler:    _Namespace_SyncMembers_Handler,
		},
		{
			MethodName: "Create",
			Handler:    _Namespace_Create_Handler,
		},
		{
			MethodName: "Show",
			Handler:    _Namespace_Show_Handler,
		},
		{
			MethodName: "UpdateDesc",
			Handler:    _Namespace_UpdateDesc_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _Namespace_Delete_Handler,
		},
		{
			MethodName: "IsExists",
			Handler:    _Namespace_IsExists_Handler,
		},
		{
			MethodName: "Favorite",
			Handler:    _Namespace_Favorite_Handler,
		},
		{
			MethodName: "Transfer",
			Handler:    _Namespace_Transfer_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "namespace/namespace.proto",
}
