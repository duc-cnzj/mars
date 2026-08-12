package wssender

import (
	"encoding/json"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func testPodEventMsg() *websocket_pb.WsProjectPodEventResponse {
	return &websocket_pb.WsProjectPodEventResponse{
		Metadata: &websocket_pb.Metadata{
			Id:   "id-1",
			Uid:  "uid-1",
			Type: websocket_pb.Type_ProjectPodEvent,
			End:  true,
			To:   websocket_pb.To_ToSelf,
		},
		ProjectId: 7,
	}
}

// ---------------------------------------------------------------------------
// SendOrDrop
// ---------------------------------------------------------------------------

func TestSendOrDrop_sends_when_space_available(t *testing.T) {
	ch := make(chan []byte, 1)
	SendOrDrop(ch, []byte("hello"), mlog.NewForConfig(nil), "label")

	select {
	case got := <-ch:
		assert.Equal(t, "hello", string(got))
	default:
		t.Fatal("expected message to be delivered")
	}
}

func TestSendOrDrop_drops_when_channel_full(t *testing.T) {
	ch := make(chan []byte, 1)
	ch <- []byte("first")

	// Second send on a full channel must drop, not block or panic.
	assert.NotPanics(t, func() {
		SendOrDrop(ch, []byte("second"), mlog.NewForConfig(nil), "label")
	})

	// Still exactly one buffered message, unchanged.
	assert.Len(t, ch, 1)
	got := <-ch
	assert.Equal(t, "first", string(got))
}

// ---------------------------------------------------------------------------
// TransformToResponse
// ---------------------------------------------------------------------------

func TestTransformToResponse_marshals_proto(t *testing.T) {
	data := TransformToResponse(testPodEventMsg())
	var resp websocket_pb.WsProjectPodEventResponse
	require.NoError(t, proto.Unmarshal(data, &resp))
	assert.Equal(t, int32(7), resp.ProjectId)
}

func TestTransformToResponse_nil_returns_empty(t *testing.T) {
	assert.Empty(t, TransformToResponse(nil))
}

// ---------------------------------------------------------------------------
// Message / Marshal / DecodeMessage
// ---------------------------------------------------------------------------

func TestMessage_Marshal_and_DecodeMessage_roundtrip(t *testing.T) {
	msg := Message{
		Data: []byte("payload"),
		To:   websocket_pb.To_ToOthers,
		ID:   "conn-1",
	}
	encoded := msg.Marshal()
	require.NotEmpty(t, encoded)

	decoded, err := DecodeMessage(encoded)
	require.NoError(t, err)
	assert.Equal(t, []byte("payload"), decoded.Data)
	assert.Equal(t, websocket_pb.To_ToOthers, decoded.To)
	assert.Equal(t, "conn-1", decoded.ID)
}

func TestMessage_Marshal_is_valid_json(t *testing.T) {
	msg := Message{Data: []byte("x"), To: websocket_pb.To_ToSelf, ID: "i"}
	var m map[string]any
	require.NoError(t, json.Unmarshal(msg.Marshal(), &m))
	assert.Equal(t, "i", m["ID"])
}

func TestDecodeMessage_invalid_json_returns_error(t *testing.T) {
	_, err := DecodeMessage([]byte("{not json"))
	assert.Error(t, err)
}

func TestDecodeMessage_empty_json_returns_zero_value(t *testing.T) {
	msg, err := DecodeMessage([]byte("{}"))
	require.NoError(t, err)
	assert.Empty(t, msg.Data)
	assert.Empty(t, msg.ID)
	assert.Equal(t, websocket_pb.To_ToSelf, msg.To)
}

func TestDecodeMessage_nil_returns_error(t *testing.T) {
	_, err := DecodeMessage(nil)
	assert.Error(t, err)
}

func TestProtoToMessage_sets_fields(t *testing.T) {
	msg := ProtoToMessage(testPodEventMsg(), "conn-1", websocket_pb.To_ToAll)
	assert.Equal(t, "conn-1", msg.ID)
	assert.Equal(t, websocket_pb.To_ToAll, msg.To)
	assert.NotEmpty(t, msg.Data)

	var resp websocket_pb.WsProjectPodEventResponse
	require.NoError(t, proto.Unmarshal(msg.Data, &resp))
	assert.Equal(t, int32(7), resp.ProjectId)
}

// ---------------------------------------------------------------------------
// GetProjectPodEventRoom
// ---------------------------------------------------------------------------

func TestGetProjectPodEventRoom_int64(t *testing.T) {
	assert.Equal(t, "project-pod-events:42", GetProjectPodEventRoom(int64(42)))
}

func TestGetProjectPodEventRoom_int(t *testing.T) {
	assert.Equal(t, "project-pod-events:42", GetProjectPodEventRoom(42))
}

// ---------------------------------------------------------------------------
// BuildPodEventResponse
// ---------------------------------------------------------------------------

func TestBuildPodEventResponse(t *testing.T) {
	data := BuildPodEventResponse("id-9", "uid-9", 3)
	var resp websocket_pb.WsProjectPodEventResponse
	require.NoError(t, proto.Unmarshal(data, &resp))

	assert.Equal(t, int32(3), resp.ProjectId)
	assert.NotNil(t, resp.Metadata)
	assert.Equal(t, "id-9", resp.Metadata.Id)
	assert.Equal(t, "uid-9", resp.Metadata.Uid)
	assert.Equal(t, websocket_pb.Type_ProjectPodEvent, resp.Metadata.Type)
	assert.True(t, resp.Metadata.End)
	assert.Equal(t, websocket_pb.ResultType_Success, resp.Metadata.Result)
	assert.Equal(t, websocket_pb.To_ToSelf, resp.Metadata.To)
}

// ---------------------------------------------------------------------------
// MatchSelectorsAndSend
// ---------------------------------------------------------------------------

func TestMatchSelectorsAndSend_sends_to_matching_project(t *testing.T) {
	ch := make(chan []byte, 4)
	sel, err := labels.Parse("app=frontend")
	require.NoError(t, err)

	pidSelectors := map[int32][]labels.Selector{
		1: {sel},
	}
	MatchSelectorsAndSend(
		ch,
		labels.Set(map[string]string{"app": "frontend"}),
		pidSelectors,
		"id-1", "uid-1",
		mlog.NewForConfig(nil),
	)

	select {
	case data := <-ch:
		var resp websocket_pb.WsProjectPodEventResponse
		require.NoError(t, proto.Unmarshal(data, &resp))
		assert.Equal(t, int32(1), resp.ProjectId)
	case <-time.After(time.Second):
		t.Fatal("expected matching project to receive pod event")
	}
}

func TestMatchSelectorsAndSend_skips_non_matching(t *testing.T) {
	ch := make(chan []byte, 4)
	sel, err := labels.Parse("app=backend")
	require.NoError(t, err)

	MatchSelectorsAndSend(
		ch,
		labels.Set(map[string]string{"app": "frontend"}),
		map[int32][]labels.Selector{1: {sel}},
		"id-1", "uid-1",
		mlog.NewForConfig(nil),
	)

	select {
	case data := <-ch:
		t.Fatalf("unexpected message for non-matching selector: %v", data)
	case <-time.After(50 * time.Millisecond):
		// expected: no match, nothing sent
	}
}

func TestMatchSelectorsAndSend_multiple_projects_one_match(t *testing.T) {
	ch := make(chan []byte, 4)
	selA, err := labels.Parse("app=a")
	require.NoError(t, err)
	selB, err := labels.Parse("app=b")
	require.NoError(t, err)

	MatchSelectorsAndSend(
		ch,
		labels.Set(map[string]string{"app": "b"}),
		map[int32][]labels.Selector{1: {selA}, 2: {selB}},
		"id-1", "uid-1",
		mlog.NewForConfig(nil),
	)

	var got int32
	for len(ch) > 0 {
		data := <-ch
		var resp websocket_pb.WsProjectPodEventResponse
		require.NoError(t, proto.Unmarshal(data, &resp))
		got = resp.ProjectId
	}
	assert.Equal(t, int32(2), got, "only the matching project should receive the event")
}

func TestMatchSelectorsAndSend_first_matching_selector_wins(t *testing.T) {
	ch := make(chan []byte, 4)
	selA, err := labels.Parse("app=x")
	require.NoError(t, err)
	selB, err := labels.Parse("tier=web")
	require.NoError(t, err)

	// Two selectors on the same project; the pod matches both — must send exactly one message.
	MatchSelectorsAndSend(
		ch,
		labels.Set(map[string]string{"app": "x", "tier": "web"}),
		map[int32][]labels.Selector{5: {selA, selB}},
		"id-1", "uid-1",
		mlog.NewForConfig(nil),
	)

	assert.Len(t, ch, 1, "one project must receive exactly one event")
}

func TestProjectPodEventObj_json_roundtrip(t *testing.T) {
	pod := &corev1.Pod{}
	pod.Name = "pod-1"
	obj := ProjectPodEventObj{
		Channel:     "project-pod-events:9",
		NamespaceID: 9,
		Pod:         pod,
	}
	data, err := json.Marshal(&obj)
	require.NoError(t, err)

	var decoded ProjectPodEventObj
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "project-pod-events:9", decoded.Channel)
	assert.Equal(t, int64(9), decoded.NamespaceID)
	assert.Equal(t, "pod-1", decoded.Pod.Name)
}
