package wssender

import (
	"encoding/json"

	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	"google.golang.org/protobuf/proto"
)

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

func ProtoToMessage(m proto.Message, to websocket_pb.To, id string) Message {
	return Message{
		Data: TransformToResponse(m),
		To:   to,
		ID:   id,
	}
}
