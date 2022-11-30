package socket

import (
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"

	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

// getPodSelectorsByManifest
// 参考 https://github.com/kubernetes/client-go/issues/193#issuecomment-363240636
// 参考源码
func getPodSelectorsByManifest(manifests []string) []string {
	var selectors []string
	info, _ := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
	for _, f := range manifests {
		obj, _, _ := info.Serializer.Decode([]byte(f), nil, nil)
		switch a := obj.(type) {
		case *v1.Deployment:
			selector, _ := metav1.LabelSelectorAsSelector(a.Spec.Selector)
			selectors = append(selectors, selector.String())
		case *v1.StatefulSet:
			selector, _ := metav1.LabelSelectorAsSelector(a.Spec.Selector)
			selectors = append(selectors, selector.String())
		case *v1.DaemonSet:
			selector, _ := metav1.LabelSelectorAsSelector(a.Spec.Selector)
			selectors = append(selectors, selector.String())
		case *batchv1.Job:
			jobPodLabels := a.Spec.Template.Labels
			if jobPodLabels != nil {
				selectors = append(selectors, labels.SelectorFromSet(jobPodLabels).String())
			}
		case *v1beta1.CronJob:
			jobPodLabels := a.Spec.JobTemplate.Spec.Template.Labels
			if jobPodLabels != nil {
				selectors = append(selectors, labels.SelectorFromSet(jobPodLabels).String())
			}
		case *batchv1.CronJob:
			jobPodLabels := a.Spec.JobTemplate.Spec.Template.Labels
			if jobPodLabels != nil {
				selectors = append(selectors, labels.SelectorFromSet(jobPodLabels).String())
			}
		default:
			mlog.Debugf("未知: %#v", a)
		}
	}

	return selectors
}

var AuditLogWithChange = events.AuditLog

type timeOrderedSetStringItem struct {
	t    time.Time
	data string
}

type orderedItemList []*timeOrderedSetStringItem

func (o orderedItemList) Len() int {
	return len(o)
}

func (o orderedItemList) Less(i, j int) bool {
	return o[i].t.Before(o[j].t)
}

func (o orderedItemList) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o orderedItemList) List() (res []string) {
	for _, item := range o {
		res = append(res, item.data)
	}
	return res
}

type timeOrderedSetString struct {
	mu      sync.RWMutex
	items   map[string]time.Time
	nowFunc func() time.Time
}

func NewTimeOrderedSetString(nowFunc func() time.Time) *timeOrderedSetString {
	return &timeOrderedSetString{items: make(map[string]time.Time), nowFunc: nowFunc}
}

func (o *timeOrderedSetString) add(s string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if _, ok := o.items[s]; ok {
		return
	}
	o.items[s] = o.nowFunc()
}

func (o *timeOrderedSetString) has(s string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	_, ok := o.items[s]
	return ok
}

func (o *timeOrderedSetString) sortedItems() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	oslist := orderedItemList{}
	for s, t := range o.items {
		oslist = append(oslist, &timeOrderedSetStringItem{
			t:    t,
			data: s,
		})
	}
	sort.Sort(oslist)
	return oslist.List()
}
