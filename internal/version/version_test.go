package version

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	assert.Equal(t, Version{
		GitRepo:        gitRepo,
		Version:        "dev",
		BuildDate:      buildDate,
		GitBranch:      gitBranch,
		GitCommit:      gitCommit,
		GitTag:         gitTag,
		GoVersion:      runtime.Version(),
		Compiler:       runtime.Compiler,
		Platform:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		KubectlVersion: kubectlVersion,
		HelmVersion:    helmVersion,
	}, GetVersion())

	gitBranch = "dev"
	gitCommit = "xx"
	gitTag = ""
	kubectlVersion = "v1.22.0"
	helmVersion = "v3.8.0"
	buildDate = "2022-01-02T00:00:00Z"
	assert.Equal(t, Version{
		GitRepo:        gitRepo,
		Version:        gitBranch + "-" + gitCommit,
		BuildDate:      buildDate,
		GitBranch:      gitBranch,
		GitCommit:      gitCommit,
		GitTag:         gitTag,
		GoVersion:      runtime.Version(),
		Compiler:       runtime.Compiler,
		Platform:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		KubectlVersion: kubectlVersion,
		HelmVersion:    helmVersion,
	}, GetVersion())
	gitTag = "v1"
	assert.Equal(t, Version{
		GitRepo:        gitRepo,
		Version:        gitTag,
		BuildDate:      buildDate,
		GitBranch:      gitBranch,
		GitCommit:      gitCommit,
		GitTag:         gitTag,
		GoVersion:      runtime.Version(),
		Compiler:       runtime.Compiler,
		Platform:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		KubectlVersion: kubectlVersion,
		HelmVersion:    helmVersion,
	}, GetVersion())
}

func TestVersion_HasBuildInfo(t *testing.T) {
	assert.True(t, (Version{
		BuildDate: "2022-01-01T00:00:00Z",
	}).HasBuildInfo())
	assert.False(t, (Version{
		BuildDate: defaultDate,
	}).HasBuildInfo())
}

func TestVersion_String(t *testing.T) {
	assert.Equal(t, "", Version{}.String())
	assert.Equal(t, "aa", Version{Version: "aa"}.String())
}
