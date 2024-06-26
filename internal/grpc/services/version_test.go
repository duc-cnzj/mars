package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/api/v4/version"
	marsVersion "github.com/duc-cnzj/mars/v4/version"
	"github.com/stretchr/testify/assert"
)

func TestVersionSvc_AuthFuncOverride(t *testing.T) {
	v := new(versionSvc)
	_, err := v.AuthFuncOverride(context.TODO(), "")
	assert.Nil(t, err)
}

func TestVersionSvc_Version(t *testing.T) {
	v := new(versionSvc)
	response, _ := v.Version(context.TODO(), &version.Request{})
	vv := marsVersion.GetVersion()
	assert.Equal(t, &version.Response{
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
	}, response)
}
