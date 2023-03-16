package wssender

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/stretchr/testify/assert"
)

func TestDecodeMessage(t *testing.T) {
	m := Message{
		Data: []byte("111"),
		To:   1,
		ID:   "id",
	}

	s, _ := json.Marshal(&m)
	message, _ := DecodeMessage(s)
	assert.Equal(t, m, message)
}

func TestMessage_Marshal(t *testing.T) {
	m := Message{
		Data: []byte("111"),
		To:   1,
		ID:   "id",
	}

	s, _ := json.Marshal(&m)
	assert.Equal(t, s, m.Marshal())
}

func TestProtoToMessage(t *testing.T) {
	m := websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			Message: "aa",
		},
	}
	marshal, _ := proto.Marshal(&m)
	assert.Equal(t, Message{
		Data: marshal,
		To:   1,
		ID:   "idx",
	}, ProtoToMessage(&m, "idx"))
}

func TestTransformToResponse(t *testing.T) {
	m := &websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			Id:   "",
			Uid:  "uid",
			Slug: "slug",
			Type: 1,
			End:  true,
		},
	}
	marshal, _ := proto.Marshal(m)
	assert.Equal(t, marshal, TransformToResponse(m))
	assert.Equal(t, []byte{}, TransformToResponse(nil))
}
