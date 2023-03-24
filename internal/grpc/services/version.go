package services

import (
	"context"

	"google.golang.org/grpc"

	"github.com/duc-cnzj/mars-client/v4/version"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	marsVersion "github.com/duc-cnzj/mars/v4/version"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		version.RegisterVersionServer(s, new(VersionSvc))
	})
	RegisterEndpoint(version.RegisterVersionHandlerFromEndpoint)
}

type VersionSvc struct {
	Guest

	version.UnimplementedVersionServer
}

func (*VersionSvc) Version(ctx context.Context, request *version.Request) (*version.Response, error) {
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
