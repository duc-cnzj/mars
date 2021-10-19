// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package gitlab

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

// GitlabClient is the client API for Gitlab service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GitlabClient interface {
	EnableProject(ctx context.Context, in *EnableProjectRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	DisableProject(ctx context.Context, in *DisableProjectRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ProjectList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ProjectListResponse, error)
	Projects(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ProjectsResponse, error)
	Branches(ctx context.Context, in *BranchesRequest, opts ...grpc.CallOption) (*BranchesResponse, error)
	Commits(ctx context.Context, in *CommitsRequest, opts ...grpc.CallOption) (*CommitsResponse, error)
	Commit(ctx context.Context, in *CommitRequest, opts ...grpc.CallOption) (*CommitResponse, error)
	PipelineInfo(ctx context.Context, in *PipelineInfoRequest, opts ...grpc.CallOption) (*PipelineInfoResponse, error)
	ConfigFile(ctx context.Context, in *ConfigFileRequest, opts ...grpc.CallOption) (*ConfigFileResponse, error)
}

type gitlabClient struct {
	cc grpc.ClientConnInterface
}

func NewGitlabClient(cc grpc.ClientConnInterface) GitlabClient {
	return &gitlabClient{cc}
}

func (c *gitlabClient) EnableProject(ctx context.Context, in *EnableProjectRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/Gitlab/EnableProject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitlabClient) DisableProject(ctx context.Context, in *DisableProjectRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/Gitlab/DisableProject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitlabClient) ProjectList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ProjectListResponse, error) {
	out := new(ProjectListResponse)
	err := c.cc.Invoke(ctx, "/Gitlab/ProjectList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitlabClient) Projects(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ProjectsResponse, error) {
	out := new(ProjectsResponse)
	err := c.cc.Invoke(ctx, "/Gitlab/Projects", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitlabClient) Branches(ctx context.Context, in *BranchesRequest, opts ...grpc.CallOption) (*BranchesResponse, error) {
	out := new(BranchesResponse)
	err := c.cc.Invoke(ctx, "/Gitlab/Branches", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitlabClient) Commits(ctx context.Context, in *CommitsRequest, opts ...grpc.CallOption) (*CommitsResponse, error) {
	out := new(CommitsResponse)
	err := c.cc.Invoke(ctx, "/Gitlab/Commits", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitlabClient) Commit(ctx context.Context, in *CommitRequest, opts ...grpc.CallOption) (*CommitResponse, error) {
	out := new(CommitResponse)
	err := c.cc.Invoke(ctx, "/Gitlab/Commit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitlabClient) PipelineInfo(ctx context.Context, in *PipelineInfoRequest, opts ...grpc.CallOption) (*PipelineInfoResponse, error) {
	out := new(PipelineInfoResponse)
	err := c.cc.Invoke(ctx, "/Gitlab/PipelineInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitlabClient) ConfigFile(ctx context.Context, in *ConfigFileRequest, opts ...grpc.CallOption) (*ConfigFileResponse, error) {
	out := new(ConfigFileResponse)
	err := c.cc.Invoke(ctx, "/Gitlab/ConfigFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GitlabServer is the server API for Gitlab service.
// All implementations must embed UnimplementedGitlabServer
// for forward compatibility
type GitlabServer interface {
	EnableProject(context.Context, *EnableProjectRequest) (*emptypb.Empty, error)
	DisableProject(context.Context, *DisableProjectRequest) (*emptypb.Empty, error)
	ProjectList(context.Context, *emptypb.Empty) (*ProjectListResponse, error)
	Projects(context.Context, *emptypb.Empty) (*ProjectsResponse, error)
	Branches(context.Context, *BranchesRequest) (*BranchesResponse, error)
	Commits(context.Context, *CommitsRequest) (*CommitsResponse, error)
	Commit(context.Context, *CommitRequest) (*CommitResponse, error)
	PipelineInfo(context.Context, *PipelineInfoRequest) (*PipelineInfoResponse, error)
	ConfigFile(context.Context, *ConfigFileRequest) (*ConfigFileResponse, error)
	mustEmbedUnimplementedGitlabServer()
}

// UnimplementedGitlabServer must be embedded to have forward compatible implementations.
type UnimplementedGitlabServer struct {
}

func (UnimplementedGitlabServer) EnableProject(context.Context, *EnableProjectRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EnableProject not implemented")
}
func (UnimplementedGitlabServer) DisableProject(context.Context, *DisableProjectRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DisableProject not implemented")
}
func (UnimplementedGitlabServer) ProjectList(context.Context, *emptypb.Empty) (*ProjectListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProjectList not implemented")
}
func (UnimplementedGitlabServer) Projects(context.Context, *emptypb.Empty) (*ProjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Projects not implemented")
}
func (UnimplementedGitlabServer) Branches(context.Context, *BranchesRequest) (*BranchesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Branches not implemented")
}
func (UnimplementedGitlabServer) Commits(context.Context, *CommitsRequest) (*CommitsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Commits not implemented")
}
func (UnimplementedGitlabServer) Commit(context.Context, *CommitRequest) (*CommitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Commit not implemented")
}
func (UnimplementedGitlabServer) PipelineInfo(context.Context, *PipelineInfoRequest) (*PipelineInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PipelineInfo not implemented")
}
func (UnimplementedGitlabServer) ConfigFile(context.Context, *ConfigFileRequest) (*ConfigFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfigFile not implemented")
}
func (UnimplementedGitlabServer) mustEmbedUnimplementedGitlabServer() {}

// UnsafeGitlabServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GitlabServer will
// result in compilation errors.
type UnsafeGitlabServer interface {
	mustEmbedUnimplementedGitlabServer()
}

func RegisterGitlabServer(s grpc.ServiceRegistrar, srv GitlabServer) {
	s.RegisterService(&Gitlab_ServiceDesc, srv)
}

func _Gitlab_EnableProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EnableProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitlabServer).EnableProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Gitlab/EnableProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitlabServer).EnableProject(ctx, req.(*EnableProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gitlab_DisableProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DisableProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitlabServer).DisableProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Gitlab/DisableProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitlabServer).DisableProject(ctx, req.(*DisableProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gitlab_ProjectList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitlabServer).ProjectList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Gitlab/ProjectList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitlabServer).ProjectList(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gitlab_Projects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitlabServer).Projects(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Gitlab/Projects",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitlabServer).Projects(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gitlab_Branches_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BranchesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitlabServer).Branches(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Gitlab/Branches",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitlabServer).Branches(ctx, req.(*BranchesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gitlab_Commits_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitlabServer).Commits(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Gitlab/Commits",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitlabServer).Commits(ctx, req.(*CommitsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gitlab_Commit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitlabServer).Commit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Gitlab/Commit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitlabServer).Commit(ctx, req.(*CommitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gitlab_PipelineInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PipelineInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitlabServer).PipelineInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Gitlab/PipelineInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitlabServer).PipelineInfo(ctx, req.(*PipelineInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gitlab_ConfigFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfigFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitlabServer).ConfigFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Gitlab/ConfigFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitlabServer).ConfigFile(ctx, req.(*ConfigFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Gitlab_ServiceDesc is the grpc.ServiceDesc for Gitlab service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Gitlab_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Gitlab",
	HandlerType: (*GitlabServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "EnableProject",
			Handler:    _Gitlab_EnableProject_Handler,
		},
		{
			MethodName: "DisableProject",
			Handler:    _Gitlab_DisableProject_Handler,
		},
		{
			MethodName: "ProjectList",
			Handler:    _Gitlab_ProjectList_Handler,
		},
		{
			MethodName: "Projects",
			Handler:    _Gitlab_Projects_Handler,
		},
		{
			MethodName: "Branches",
			Handler:    _Gitlab_Branches_Handler,
		},
		{
			MethodName: "Commits",
			Handler:    _Gitlab_Commits_Handler,
		},
		{
			MethodName: "Commit",
			Handler:    _Gitlab_Commit_Handler,
		},
		{
			MethodName: "PipelineInfo",
			Handler:    _Gitlab_PipelineInfo_Handler,
		},
		{
			MethodName: "ConfigFile",
			Handler:    _Gitlab_ConfigFile_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "gitlab/gitlab.proto",
}