package socket

import (
	"fmt"
	"regexp"
	"strings"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/plugins"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

type pipelineVars struct {
	Pipeline string
	Commit   string
	Branch   string
}

var matchTag = regexp.MustCompile(`image:\s+(\S+)`)

func matchDockerImage(v pipelineVars, manifest string) string {
	var (
		candidateImages []string
		all             []string
		existsMap       = make(map[string]struct{})
	)
	submatch := matchTag.FindAllStringSubmatch(manifest, -1)
	for _, matches := range submatch {
		if len(matches) == 2 {
			image := strings.Trim(matches[1], "\"")

			if _, ok := existsMap[image]; ok {
				continue
			}
			existsMap[image] = struct{}{}
			all = append(all, image)
			if imageUsedPipelineVars(v, image) {
				candidateImages = append(candidateImages, image)
			}
		}
	}
	// 如果找到至少一个镜像就直接返回，如果未找到，则返回所有匹配到的镜像
	if len(candidateImages) > 0 {
		return strings.Join(candidateImages, " ")
	}

	return strings.Join(all, " ")
}

// imageUsedPipelineVars 使用的流水线变量的镜像，都把他当成是我们的目标镜像
func imageUsedPipelineVars(v pipelineVars, s string) bool {
	var pipelineVarsSlice []string
	if v.Pipeline != "" {
		pipelineVarsSlice = append(pipelineVarsSlice, v.Pipeline)
	}
	if v.Commit != "" {
		pipelineVarsSlice = append(pipelineVarsSlice, v.Commit)
	}
	if v.Branch != "" {
		pipelineVarsSlice = append(pipelineVarsSlice, v.Branch)
	}
	for _, pvar := range pipelineVarsSlice {
		if strings.Contains(s, pvar) {
			return true
		}
	}

	return false
}

// getPodSelectorsInDeploymentAndStatefulSetByManifest
// 参考 https://github.com/kubernetes/client-go/issues/193#issuecomment-363240636
// 参考源码
func getPodSelectorsInDeploymentAndStatefulSetByManifest(manifest string) []string {
	var selectors []string
	split := strings.Split(manifest, "---")
	info, _ := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
	for _, f := range split {
		obj, _, _ := info.Serializer.Decode([]byte(f), nil, nil)
		switch a := obj.(type) {
		case *v1.Deployment:
			var labels []string
			for k, v := range a.Spec.Selector.MatchLabels {
				labels = append(labels, fmt.Sprintf("%s=%s", k, v))
			}
			selectors = append(selectors, strings.Join(labels, ","))
		case *v1.StatefulSet:
			var labels []string
			for k, v := range a.Spec.Selector.MatchLabels {
				labels = append(labels, fmt.Sprintf("%s=%s", k, v))
			}
			selectors = append(selectors, strings.Join(labels, ","))
		}
	}

	return selectors
}

func getPreOccupiedLenByValuesYaml(values string) int {
	var sub = 0
	if len(values) > 0 {
		submatch := hostMatch.FindAllStringSubmatch(values, -1)
		if len(submatch) == 1 && len(submatch[0]) >= 1 {
			sub = max(sub, len(submatch[0][1]))
		}
	}
	return sub
}

func getDomainByIndex(project, namespace string, index, preOccupiedLen int) string {
	if !app.Config().HasWildcardDomain() {
		return ""
	}

	return plugins.GetDomainResolverPlugin().GetDomainByIndex(strings.TrimLeft(app.Config().WildcardDomain, "*."), project, namespace, index, preOccupiedLen)
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

var AuditLogWithChange = events.AuditLog
