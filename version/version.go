package version

import (
	"fmt"
	"runtime"
)

const (
	unknown     string = "<unknown>"
	defaultDate string = "1970-01-01T00:00:00Z"
	gitRepo     string = "https://github.com/duc-cnzj/mars"
)

var (
	gitBranch      string = unknown     // `git rev-parse --abbrev-ref HEAD`
	gitCommit      string = unknown     // output from `git rev-parse --short HEAD`
	gitTag         string = unknown     // output from `git describe --exact-match --tags HEAD` (if clean tree state)
	kubectlVersion string = unknown     // determined from go.mod file `go list -m all | grep k8s.io/client-go | cut -d " " -f2`
	helmVersion    string = unknown     // determined from go.mod file `go list -m all | grep helm.sh/helm/v3 | cut -d " " -f2`
	buildDate      string = defaultDate // output from `date -u +'%Y-%m-%dT%H:%M:%SZ'`
)

// Version contains Argo version information
type Version struct {
	GitRepo        string
	Version        string
	BuildDate      string
	GitCommit      string
	GitBranch      string
	GitTag         string
	GoVersion      string
	Compiler       string
	Platform       string
	KubectlVersion string
	HelmVersion    string
}

func (v Version) String() string {
	return v.Version
}

func (v Version) HasBuildInfo() bool {
	return v.BuildDate != defaultDate
}

// GetVersion returns the version information
func GetVersion() Version {
	var versionStr string = gitTag

	if versionStr == "" && gitBranch != "" && gitCommit != "" {
		versionStr = fmt.Sprintf("%s-%s", gitBranch, gitCommit)
	}

	return Version{
		GitRepo:        gitRepo,
		Version:        versionStr,
		BuildDate:      buildDate,
		GitBranch:      gitBranch,
		GitCommit:      gitCommit,
		GitTag:         gitTag,
		GoVersion:      runtime.Version(),
		Compiler:       runtime.Compiler,
		Platform:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		KubectlVersion: kubectlVersion,
		HelmVersion:    helmVersion,
	}
}
