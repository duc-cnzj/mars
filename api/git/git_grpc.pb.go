// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: git/git.proto

package git

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
	Git_AllRepos_FullMethodName           = "/git.Git/AllRepos"
	Git_ProjectOptions_FullMethodName     = "/git.Git/ProjectOptions"
	Git_BranchOptions_FullMethodName      = "/git.Git/BranchOptions"
	Git_CommitOptions_FullMethodName      = "/git.Git/CommitOptions"
	Git_Commit_FullMethodName             = "/git.Git/Commit"
	Git_PipelineInfo_FullMethodName       = "/git.Git/PipelineInfo"
	Git_GetChartValuesYaml_FullMethodName = "/git.Git/GetChartValuesYaml"
)

// GitClient is the client API for Git service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GitClient interface {
	AllRepos(ctx context.Context, in *AllReposRequest, opts ...grpc.CallOption) (*AllReposResponse, error)
	ProjectOptions(ctx context.Context, in *ProjectOptionsRequest, opts ...grpc.CallOption) (*ProjectOptionsResponse, error)
	BranchOptions(ctx context.Context, in *BranchOptionsRequest, opts ...grpc.CallOption) (*BranchOptionsResponse, error)
	CommitOptions(ctx context.Context, in *CommitOptionsRequest, opts ...grpc.CallOption) (*CommitOptionsResponse, error)
	Commit(ctx context.Context, in *CommitRequest, opts ...grpc.CallOption) (*CommitResponse, error)
	PipelineInfo(ctx context.Context, in *PipelineInfoRequest, opts ...grpc.CallOption) (*PipelineInfoResponse, error)
	GetChartValuesYaml(ctx context.Context, in *GetChartValuesYamlRequest, opts ...grpc.CallOption) (*GetChartValuesYamlResponse, error)
}

type gitClient struct {
	cc grpc.ClientConnInterface
}

func NewGitClient(cc grpc.ClientConnInterface) GitClient {
	return &gitClient{cc}
}

func (c *gitClient) AllRepos(ctx context.Context, in *AllReposRequest, opts ...grpc.CallOption) (*AllReposResponse, error) {
	out := new(AllReposResponse)
	err := c.cc.Invoke(ctx, Git_AllRepos_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitClient) ProjectOptions(ctx context.Context, in *ProjectOptionsRequest, opts ...grpc.CallOption) (*ProjectOptionsResponse, error) {
	out := new(ProjectOptionsResponse)
	err := c.cc.Invoke(ctx, Git_ProjectOptions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitClient) BranchOptions(ctx context.Context, in *BranchOptionsRequest, opts ...grpc.CallOption) (*BranchOptionsResponse, error) {
	out := new(BranchOptionsResponse)
	err := c.cc.Invoke(ctx, Git_BranchOptions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitClient) CommitOptions(ctx context.Context, in *CommitOptionsRequest, opts ...grpc.CallOption) (*CommitOptionsResponse, error) {
	out := new(CommitOptionsResponse)
	err := c.cc.Invoke(ctx, Git_CommitOptions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitClient) Commit(ctx context.Context, in *CommitRequest, opts ...grpc.CallOption) (*CommitResponse, error) {
	out := new(CommitResponse)
	err := c.cc.Invoke(ctx, Git_Commit_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitClient) PipelineInfo(ctx context.Context, in *PipelineInfoRequest, opts ...grpc.CallOption) (*PipelineInfoResponse, error) {
	out := new(PipelineInfoResponse)
	err := c.cc.Invoke(ctx, Git_PipelineInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitClient) GetChartValuesYaml(ctx context.Context, in *GetChartValuesYamlRequest, opts ...grpc.CallOption) (*GetChartValuesYamlResponse, error) {
	out := new(GetChartValuesYamlResponse)
	err := c.cc.Invoke(ctx, Git_GetChartValuesYaml_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GitServer is the server API for Git service.
// All implementations must embed UnimplementedGitServer
// for forward compatibility
type GitServer interface {
	AllRepos(context.Context, *AllReposRequest) (*AllReposResponse, error)
	ProjectOptions(context.Context, *ProjectOptionsRequest) (*ProjectOptionsResponse, error)
	BranchOptions(context.Context, *BranchOptionsRequest) (*BranchOptionsResponse, error)
	CommitOptions(context.Context, *CommitOptionsRequest) (*CommitOptionsResponse, error)
	Commit(context.Context, *CommitRequest) (*CommitResponse, error)
	PipelineInfo(context.Context, *PipelineInfoRequest) (*PipelineInfoResponse, error)
	GetChartValuesYaml(context.Context, *GetChartValuesYamlRequest) (*GetChartValuesYamlResponse, error)
	mustEmbedUnimplementedGitServer()
}

// UnimplementedGitServer must be embedded to have forward compatible implementations.
type UnimplementedGitServer struct {
}

func (UnimplementedGitServer) AllRepos(context.Context, *AllReposRequest) (*AllReposResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AllRepos not implemented")
}
func (UnimplementedGitServer) ProjectOptions(context.Context, *ProjectOptionsRequest) (*ProjectOptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProjectOptions not implemented")
}
func (UnimplementedGitServer) BranchOptions(context.Context, *BranchOptionsRequest) (*BranchOptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BranchOptions not implemented")
}
func (UnimplementedGitServer) CommitOptions(context.Context, *CommitOptionsRequest) (*CommitOptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommitOptions not implemented")
}
func (UnimplementedGitServer) Commit(context.Context, *CommitRequest) (*CommitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Commit not implemented")
}
func (UnimplementedGitServer) PipelineInfo(context.Context, *PipelineInfoRequest) (*PipelineInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PipelineInfo not implemented")
}
func (UnimplementedGitServer) GetChartValuesYaml(context.Context, *GetChartValuesYamlRequest) (*GetChartValuesYamlResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChartValuesYaml not implemented")
}
func (UnimplementedGitServer) mustEmbedUnimplementedGitServer() {}

// UnsafeGitServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GitServer will
// result in compilation errors.
type UnsafeGitServer interface {
	mustEmbedUnimplementedGitServer()
}

func RegisterGitServer(s grpc.ServiceRegistrar, srv GitServer) {
	s.RegisterService(&Git_ServiceDesc, srv)
}

func _Git_AllRepos_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AllReposRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitServer).AllRepos(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Git_AllRepos_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitServer).AllRepos(ctx, req.(*AllReposRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Git_ProjectOptions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProjectOptionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitServer).ProjectOptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Git_ProjectOptions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitServer).ProjectOptions(ctx, req.(*ProjectOptionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Git_BranchOptions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BranchOptionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitServer).BranchOptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Git_BranchOptions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitServer).BranchOptions(ctx, req.(*BranchOptionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Git_CommitOptions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitOptionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitServer).CommitOptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Git_CommitOptions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitServer).CommitOptions(ctx, req.(*CommitOptionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Git_Commit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitServer).Commit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Git_Commit_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitServer).Commit(ctx, req.(*CommitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Git_PipelineInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PipelineInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitServer).PipelineInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Git_PipelineInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitServer).PipelineInfo(ctx, req.(*PipelineInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Git_GetChartValuesYaml_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetChartValuesYamlRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitServer).GetChartValuesYaml(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Git_GetChartValuesYaml_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitServer).GetChartValuesYaml(ctx, req.(*GetChartValuesYamlRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Git_ServiceDesc is the grpc.ServiceDesc for Git service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Git_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "git.Git",
	HandlerType: (*GitServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AllRepos",
			Handler:    _Git_AllRepos_Handler,
		},
		{
			MethodName: "ProjectOptions",
			Handler:    _Git_ProjectOptions_Handler,
		},
		{
			MethodName: "BranchOptions",
			Handler:    _Git_BranchOptions_Handler,
		},
		{
			MethodName: "CommitOptions",
			Handler:    _Git_CommitOptions_Handler,
		},
		{
			MethodName: "Commit",
			Handler:    _Git_Commit_Handler,
		},
		{
			MethodName: "PipelineInfo",
			Handler:    _Git_PipelineInfo_Handler,
		},
		{
			MethodName: "GetChartValuesYaml",
			Handler:    _Git_GetChartValuesYaml_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "git/git.proto",
}