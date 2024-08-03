package socket

import (
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type pipelineVars struct {
	Pipeline string
	Commit   string
	Branch   string
}

var matchTag = regexp.MustCompile(`image:\s+(\S+)`)

func matchDockerImage(v pipelineVars, manifest string) []string {
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
		return candidateImages
	}

	return all
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

func newTimeOrderedSetString(nowFunc func() time.Time) *timeOrderedSetString {
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
