package version

import (
	"fmt"
	"runtime"
)

var (
	gitRepo        = ""
	gitBranch      = ""                     // `git rev-parse --abbrev-ref HEAD`
	buildDate      = "1970-01-01T00:00:00Z" // output from `date -u +'%Y-%m-%dT%H:%M:%SZ'`
	gitCommit      = ""                     // output from `git rev-parse --short HEAD`
	gitTag         = ""                     // output from `git describe --exact-match --tags HEAD` (if clean tree state)
	kubectlVersion = ""                     // determined from go.mod file `go list -m all | grep k8s.io/client-go | cut -d " " -f2`
	helmVersion    = ""                     // determined from go.mod file `go list -m all | grep helm.sh/helm/v3 | cut -d " " -f2`
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

// GetVersion returns the version information
func GetVersion() Version {
	var versionStr string = gitTag

	if versionStr == "" {
		versionStr = fmt.Sprintf("%s-%s", gitBranch, gitCommit)
	}

	return Version{
		GitRepo:        fmt.Sprintf("https://%s", gitRepo),
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
