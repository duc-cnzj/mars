package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v4/version"
	marsVersion "github.com/duc-cnzj/mars/v4/version"
)

var _ version.VersionServer = (*versionSvc)(nil)

type versionSvc struct {
	guest

	version.UnimplementedVersionServer
}

func NewVersionSvc() version.VersionServer {
	return &versionSvc{}
}

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
