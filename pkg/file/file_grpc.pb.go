// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.17.3
// source: file/file.proto

package file

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

// FileClient is the client API for File service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FileClient interface {
	//  文件列表
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
	//  删除文件
	Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	//  DeleteUndocumentedFiles 删除未被记录的文件，model 表中没有，但是文件目录中有
	DeleteUndocumentedFiles(ctx context.Context, in *DeleteUndocumentedFilesRequest, opts ...grpc.CallOption) (*DeleteUndocumentedFilesResponse, error)
	// DiskInfo 查看上传文件目录大小
	DiskInfo(ctx context.Context, in *DiskInfoRequest, opts ...grpc.CallOption) (*DiskInfoResponse, error)
}

type fileClient struct {
	cc grpc.ClientConnInterface
}

func NewFileClient(cc grpc.ClientConnInterface) FileClient {
	return &fileClient{cc}
}

func (c *fileClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, "/file.File/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileClient) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, "/file.File/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileClient) DeleteUndocumentedFiles(ctx context.Context, in *DeleteUndocumentedFilesRequest, opts ...grpc.CallOption) (*DeleteUndocumentedFilesResponse, error) {
	out := new(DeleteUndocumentedFilesResponse)
	err := c.cc.Invoke(ctx, "/file.File/DeleteUndocumentedFiles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileClient) DiskInfo(ctx context.Context, in *DiskInfoRequest, opts ...grpc.CallOption) (*DiskInfoResponse, error) {
	out := new(DiskInfoResponse)
	err := c.cc.Invoke(ctx, "/file.File/DiskInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FileServer is the server API for File service.
// All implementations must embed UnimplementedFileServer
// for forward compatibility
type FileServer interface {
	//  文件列表
	List(context.Context, *ListRequest) (*ListResponse, error)
	//  删除文件
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	//  DeleteUndocumentedFiles 删除未被记录的文件，model 表中没有，但是文件目录中有
	DeleteUndocumentedFiles(context.Context, *DeleteUndocumentedFilesRequest) (*DeleteUndocumentedFilesResponse, error)
	// DiskInfo 查看上传文件目录大小
	DiskInfo(context.Context, *DiskInfoRequest) (*DiskInfoResponse, error)
	mustEmbedUnimplementedFileServer()
}

// UnimplementedFileServer must be embedded to have forward compatible implementations.
type UnimplementedFileServer struct {
}

func (UnimplementedFileServer) List(context.Context, *ListRequest) (*ListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedFileServer) Delete(context.Context, *DeleteRequest) (*DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedFileServer) DeleteUndocumentedFiles(context.Context, *DeleteUndocumentedFilesRequest) (*DeleteUndocumentedFilesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUndocumentedFiles not implemented")
}
func (UnimplementedFileServer) DiskInfo(context.Context, *DiskInfoRequest) (*DiskInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DiskInfo not implemented")
}
func (UnimplementedFileServer) mustEmbedUnimplementedFileServer() {}

// UnsafeFileServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FileServer will
// result in compilation errors.
type UnsafeFileServer interface {
	mustEmbedUnimplementedFileServer()
}

func RegisterFileServer(s grpc.ServiceRegistrar, srv FileServer) {
	s.RegisterService(&File_ServiceDesc, srv)
}

func _File_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/file.File/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServer).List(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _File_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/file.File/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServer).Delete(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _File_DeleteUndocumentedFiles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUndocumentedFilesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServer).DeleteUndocumentedFiles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/file.File/DeleteUndocumentedFiles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServer).DeleteUndocumentedFiles(ctx, req.(*DeleteUndocumentedFilesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _File_DiskInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DiskInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServer).DiskInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/file.File/DiskInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServer).DiskInfo(ctx, req.(*DiskInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// File_ServiceDesc is the grpc.ServiceDesc for File service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var File_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "file.File",
	HandlerType: (*FileServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _File_List_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _File_Delete_Handler,
		},
		{
			MethodName: "DeleteUndocumentedFiles",
			Handler:    _File_DeleteUndocumentedFiles_Handler,
		},
		{
			MethodName: "DiskInfo",
			Handler:    _File_DiskInfo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "file/file.proto",
}
