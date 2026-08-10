package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/version"
	marsVersion "github.com/duc-cnzj/mars/v6/internal/version"
)

var _ version.VersionServer = (*versionSvc)(nil)

// versionSvc 是 version.VersionServer 的 gRPC 实现：返回构建版本信息，无状态，由 NewVersionSvc 构造。
type versionSvc struct {
	version.UnimplementedVersionServer
}

// NewVersionSvc 构造版本信息服务，无外部依赖，为免登录公开接口（白名单见 middlewares.PublicMethods）。
func NewVersionSvc() version.VersionServer {
	return &versionSvc{}
}

// Version 返回编译期注入的版本信息（版本号、构建时间、git 分支/提交等），
// 供前端展示与部署排查使用。
func (*versionSvc) Version(ctx context.Context, request *version.Request) (*version.Response, error) {
	vv := marsVersion.GetVersion()

	return &version.Response{
		Version:        vv.Version,
		BuildDate:      vv.BuildDate,
		GitBranch:      vv.GitBranch,
		GitCommit:      vv.GitCommit,
		GitTag:         vv.GitTag,
		GoVersion:      vv.GoVersion,
		Compiler:       vv.Compiler,
		Platform:       vv.Platform,
		KubectlVersion: vv.KubectlVersion,
		HelmVersion:    vv.HelmVersion,
		GitRepo:        vv.GitRepo,
	}, nil
}
