package services

import (
	"context"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/pkg/version"
	marsVersion "github.com/duc-cnzj/mars/version"
	"google.golang.org/protobuf/types/known/emptypb"
)

type VersionService struct {
	version.UnsafeVersionServer
}

func (*VersionService) Get(ctx context.Context, empty *emptypb.Empty) (*version.VersionResponse, error) {
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
