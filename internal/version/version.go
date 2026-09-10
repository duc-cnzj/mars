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

// 以下变量由 Makefile 的 -ldflags -X 在编译期覆盖注入（命令行见行尾注释），
// 未经注入时保持占位默认值。
var (
	gitBranch      string = unknown     // 由 `git rev-parse --abbrev-ref HEAD` 注入
	gitCommit      string = unknown     // 由 `git rev-parse --short HEAD` 注入
	gitTag         string = "dev"       // 由 `git describe --exact-match --tags HEAD` 注入（工作区干净时）
	kubectlVersion string = unknown     // 由 `go list -m -f "{{.Path}} {{.Version}}" all` 里的 k8s.io/client-go 版本注入
	helmVersion    string = unknown     // 由 `go list -m -f "{{.Path}} {{.Version}}" all` 里的 helm.sh/helm/v3 版本注入
	buildDate      string = defaultDate // 由 `date -u +'%Y-%m-%dT%H:%M:%SZ'` 注入
)

// Version 是 mars 编译期元数据：仓库、版本号、构建日期与各依赖版本。
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

// String 返回版本号字符串。
func (v Version) String() string {
	return v.Version
}

// HasBuildInfo 判断是否携带构建信息：BuildDate 非默认占位日期即视为已构建。
func (v Version) HasBuildInfo() bool {
	return v.BuildDate != defaultDate
}

// GetVersion 返回由编译期注入元数据组装成的 Version。
func GetVersion() Version {
	var versionStr = gitTag

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
