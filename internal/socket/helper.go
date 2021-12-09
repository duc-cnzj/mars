package socket

import (
	"fmt"
	"strings"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// getPodSelectorsInDeploymentAndStatefulSetByManifest FIXME: 比较 hack
// 参考 https://github.com/kubernetes/client-go/issues/193#issuecomment-363240636
func getPodSelectorsInDeploymentAndStatefulSetByManifest(manifest string) []string {
	var selectors []string
	split := strings.Split(manifest, "---")
	for _, f := range split {
		obj, _, _ := scheme.Codecs.UniversalDeserializer().Decode([]byte(f), nil, nil)
		switch a := obj.(type) {
		case *v1.Deployment:
			mlog.Debug("############### getPodSelectorsInDeploymentAndStatefulSetByManifest(manifest string) ###############")
			var labels []string
			for k, v := range a.Spec.Selector.MatchLabels {
				labels = append(labels, fmt.Sprintf("%s=%s", k, v))
			}
			selectors = append(selectors, strings.Join(labels, ","))
		case *v1.StatefulSet:
			mlog.Debug("############### getPodSelectorsInDeploymentAndStatefulSetByManifest(manifest string) ###############")
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
