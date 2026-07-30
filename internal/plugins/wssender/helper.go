package wssender

import (
	"encoding/json"
	"fmt"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/mlog"

	websocket_pb "github.com/duc-cnzj/mars/api/v5/websocket"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"google.golang.org/protobuf/proto"
)

// SendOrDrop tries non-blocking channel send; drops on full channel to avoid head-of-line blocking.
func SendOrDrop(ch chan<- []byte, data []byte, log mlog.Logger, label string) {
	select {
	case ch <- data:
	default:
		log.Debugf("[WS] drop msg to %s: channel full", label)
	}
}

func TransformToResponse(message proto.Message) []byte {
	if message == nil {
		return []byte{}
	}
	marshal, _ := proto.Marshal(message)
	return marshal
}

type Message struct {
	Data []byte
	To   websocket_pb.To
	ID   string
}

func (m Message) Marshal() []byte {
	marshal, _ := json.Marshal(&m)
	return marshal
}

func DecodeMessage(data []byte) (msg Message, err error) {
	err = json.Unmarshal(data, &msg)
	return
}

func ProtoToMessage(m application.WebsocketMessage, id string, to websocket_pb.To) Message {
	return Message{
		Data: TransformToResponse(m),
		To:   to,
		ID:   id,
	}
}

// ---------------------------------------------------------------------------
// Shared pod event types and helpers (used by redis, nsq, memory backends)
// ---------------------------------------------------------------------------

// ProjectPodEventObj is the JSON payload for pod events published via PubSub.
type ProjectPodEventObj struct {
	Channel     string  `json:"channel"`
	NamespaceID int64   `json:"namespace_id"`
	Pod         *v1.Pod `json:"pod"`
}

// GetProjectPodEventRoom returns the PubSub channel name for a given namespace.
// NSQ appends "#ephemeral" suffix to this base name.
func GetProjectPodEventRoom[T int64 | int](nsID T) string {
	return fmt.Sprintf("project-pod-events:%d", nsID)
}

// BuildPodEventResponse builds a protobuf-marshalled WsProjectPodEventResponse.
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

// MatchSelectorsAndSend iterates pidSelectors, matches against podLabels,
// and sends a pod event to ch for every matching project.
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
