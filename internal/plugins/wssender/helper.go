package wssender

import (
	"encoding/json"
	"fmt"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/mlog"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// SendOrDrop 非阻塞投递消息到通道；通道已满则丢弃该消息，避免头对头阻塞。
func SendOrDrop(ch chan<- []byte, data []byte, log mlog.Logger, label string) {
	select {
	case ch <- data:
	default:
		log.Debugf("[WS] drop msg to %s: channel full", label)
	}
}

// TransformToResponse 将 websocket 响应序列化为 proto 字节；nil 消息返回空切片。
func TransformToResponse(message proto.Message) []byte {
	if message == nil {
		return []byte{}
	}
	marshal, _ := proto.Marshal(message)
	return marshal
}

// Message 是 PubSub 传输的通用消息载体：Data 为序列化后的负载，To 决定投递目标，ID 标记消息归属连接。
type Message struct {
	Data []byte
	To   websocket_pb.To
	ID   string
}

// Marshal 将 Message 编码为 JSON。
func (m Message) Marshal() []byte {
	marshal, _ := json.Marshal(&m)
	return marshal
}

// DecodeMessage 将 JSON 解码回 Message。
func DecodeMessage(data []byte) (msg Message, err error) {
	err = json.Unmarshal(data, &msg)
	return
}

// ProtoToMessage 把 websocket proto 消息转成带目标与归属的 Message。
func ProtoToMessage(m app.WebsocketMessage, id string, to websocket_pb.To) Message {
	return Message{
		Data: TransformToResponse(m),
		To:   to,
		ID:   id,
	}
}

// ---------------------------------------------------------------------------
// 共享 pod 事件类型与工具（redis / nsq / memory 三种后端共用）
// ---------------------------------------------------------------------------

// ProjectPodEventObj 是经 PubSub 发布的 pod 事件 JSON 负载。
type ProjectPodEventObj struct {
	Channel     string  `json:"channel"`
	NamespaceID int64   `json:"namespace_id"`
	Pod         *v1.Pod `json:"pod"`
}

// GetProjectPodEventRoom 返回指定命名空间的 pod 事件 PubSub 频道名。
// NSQ 会在该基础名后追加 "#ephemeral" 后缀。
func GetProjectPodEventRoom[T int64 | int](nsID T) string {
	return fmt.Sprintf("project-pod-events:%d", nsID)
}

// BuildPodEventResponse 构造 protobuf 序列化的 WsProjectPodEventResponse。
func BuildPodEventResponse(id, uid string, pid int32) []byte {
	return TransformToResponse(&websocket_pb.WsProjectPodEventResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     id,
			Uid:    uid,
			Type:   websocket_pb.Type_ProjectPodEvent,
			End:    true,
			Result: websocket_pb.ResultType_Success,
			To:     websocket_pb.To_ToSelf,
		},
		ProjectId: pid,
	})
}

// MatchSelectorsAndSend 遍历 pidSelectors，用 podLabels 匹配每个项目的选择器，
// 命中即向 ch 投递一条 pod 事件。
func MatchSelectorsAndSend(
	ch chan<- []byte,
	podLabels labels.Set,
	pidSelectors map[int32][]labels.Selector,
	id, uid string,
	log mlog.Logger,
) {
	for pid, selectors := range pidSelectors {
		for _, selector := range selectors {
			if selector.Matches(podLabels) {
				SendOrDrop(ch, BuildPodEventResponse(id, uid, pid), log, fmt.Sprintf("pod-event:pid=%d", pid))
				break
			}
		}
	}
}
