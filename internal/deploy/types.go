package deploy

import (
	"context"
	"io"
	"regexp"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/application"
)

// deployResult 是部署最终结果的线程安全容器，由 HandleMessage/Finish 写入，
// Finish 结束时通过 messager 一次性上报给用户。
type deployResult struct {
	sync.RWMutex
	result websocket_pb.ResultType
	msg    string
	model  *types.ProjectModel
	set    bool
}

// IsSet 判断部署结果是否已写入。
func (d *deployResult) IsSet() bool {
	d.RLock()
	defer d.RUnlock()
	return d.set
}

// Msg 返回部署结果消息文本。
func (d *deployResult) Msg() string {
	d.RLock()
	defer d.RUnlock()
	return d.msg
}

// Model 返回部署结果关联的项目模型。
func (d *deployResult) Model() *types.ProjectModel {
	d.RLock()
	defer d.RUnlock()
	return d.model
}

// ResultType 返回部署结果状态类型。
func (d *deployResult) ResultType() websocket_pb.ResultType {
	d.RLock()
	defer d.RUnlock()
	return d.result
}

// Set 一次性写入部署结果（状态/消息/模型），并标记已写入。
func (d *deployResult) Set(t websocket_pb.ResultType, msg string, model *types.ProjectModel) {
	d.Lock()
	defer d.Unlock()
	d.result = t
	d.msg = msg
	d.model = model
	d.set = true
}

// vars 是系统内置变量表，SystemVariableLoader 注入后供 values.yaml 渲染与持久化。
type vars map[string]string

// ToKeyValue 把变量表转成 KeyValue 列表，供项目记录持久化。
func (v vars) ToKeyValue() (res []*types.KeyValue) {
	for k, va := range v {
		res = append(res, &types.KeyValue{
			Key:   k,
			Value: va,
		})
	}
	return
}

// MustGetString 读取变量值，缺失时返回空串。
func (v vars) MustGetString(key string) string {
	if value, ok := v[key]; ok {
		return value
	}

	return ""
}

// Add 写入一个变量。
func (v vars) Add(key, value string) {
	v[key] = value
}

// pipelineVars 是流水线注入的变量子集，用于从 manifest 中挑选目标镜像。
type pipelineVars struct {
	Pipeline string
	Commit   string
	Branch   string
}

var matchTag = regexp.MustCompile(`image:\s+(\S+)`)

// matchDockerImage 从 manifest 提取镜像列表，优先返回使用了流水线变量的目标镜像。
func matchDockerImage(v pipelineVars, manifest string) []string {
	var (
		candidateImages = make([]string, 0)
		all             = make([]string, 0)
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

type internalCloser func() error

// Close 执行包装的清理函数，满足 io.Closer 契约。
func (fn internalCloser) Close() error {
	return fn()
}

// NewCloser 把任意清理函数包装为 io.Closer。
func NewCloser(fn func() error) io.Closer {
	return internalCloser(fn)
}

type emptyPubSub struct{}

// NewEmptyPubSub 返回无操作的 PubSub 实现，供未绑定消息总线的场景占位。
func NewEmptyPubSub() application.PubSub {
	return &emptyPubSub{}
}

// Join 空实现。
func (e *emptyPubSub) Join(projectID int64) error {
	return nil
}

// Leave 空实现。
func (e *emptyPubSub) Leave(nsID int64, projectID int64) error {
	return nil
}

// Run 空实现。
func (e *emptyPubSub) Run(ctx context.Context) error {
	return nil
}

// Publish 空实现。
func (e *emptyPubSub) Publish(nsID int64, pod *corev1.Pod) error {
	return nil
}

// Info 空实现，恒返回 nil。
func (e *emptyPubSub) Info() any {
	return nil
}

// Uid 空实现，恒返回空串。
func (e *emptyPubSub) Uid() string {
	return ""
}

// ID 空实现，恒返回空串。
func (e *emptyPubSub) ID() string {
	return ""
}

// ToSelf 空实现。
func (e *emptyPubSub) ToSelf(message application.WebsocketMessage) error {
	return nil
}

// ToAll 空实现。
func (e *emptyPubSub) ToAll(message application.WebsocketMessage) error {
	return nil
}

// Subscribe 空实现，恒返回 nil channel。
func (e *emptyPubSub) Subscribe() <-chan []byte {
	return nil
}

// Close 空实现。
func (e *emptyPubSub) Close() error {
	return nil
}
