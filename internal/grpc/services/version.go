package services

import (
	"context"

	"github.com/duc-cnzj/mars/client/version"
	"github.com/duc-cnzj/mars/internal/mlog"
	marsVersion "github.com/duc-cnzj/mars/version"
)

type VersionService struct {
	version.UnsafeVersionServer
}

func (*VersionService) Version(ctx context.Context, request *version.VersionRequest) (*version.VersionResponse, error) {
	vv := marsVersion.GetVersion()

	return &version.VersionResponse{
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

func (*VersionService) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}
