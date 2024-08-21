package socket

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMessageSenderFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	ms := NewMessageSender(conn, "slug", websocket_pb.Type_SetUid)

	assert.NotNil(t, ms)
	assert.Equal(t, conn, ms.(*messager).conn)
	assert.Equal(t, "slug", ms.(*messager).slugName)
	assert.Equal(t, websocket_pb.Type_SetUid, ms.(*messager).wsType)
	assert.NotNil(t, ms.(*messager).percent)
}

func TestSendDeployedResultFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := application.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)
	sub.EXPECT().ToSelf(gomock.Any())

	ms := NewMessageSender(conn, "slug", websocket_pb.Type_HandleAuthorize)
	ms.SendDeployedResult(websocket_pb.ResultType_Deployed, "message", &types.ProjectModel{})
}

func TestSendEndErrorFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := application.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	sub.EXPECT().ToSelf(gomock.Any())
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)
	ms := NewMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.SendEndError(errors.New("error"))
}

func TestSendProcessPercentFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := application.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	sub.EXPECT().ToSelf(gomock.Any())
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)

	ms := NewMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.SendProcessPercent(50)
}

func TestSendMsgFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := application.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	sub.EXPECT().ToSelf(gomock.Any())
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)

	ms := NewMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.SendMsg("message")
}

func TestSendMsgWithContainerLogFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := application.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	sub.EXPECT().ToSelf(gomock.Any())
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)

	ms := NewMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.SendMsgWithContainerLog("message", []*websocket_pb.Container{})
}

func TestSendProtoMsgFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := application.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	sub.EXPECT().ToSelf(gomock.Any())

	ms := NewMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.SendProtoMsg(&WsResponse{})
}

func TestProcessPercentAddFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := application.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	sub.EXPECT().ToSelf(gomock.Any())
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)

	ms := NewMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.Add()
}

func TestProcessPercentToFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := application.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).AnyTimes()
	sub.EXPECT().ToSelf(gomock.Any()).AnyTimes()
	conn.EXPECT().UID().Return("uid").AnyTimes()
	conn.EXPECT().ID().Return("id").AnyTimes()

	ms := NewMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.To(3)
}

func Test_messager_Current(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := application.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).AnyTimes()
	sub.EXPECT().ToSelf(gomock.Any()).AnyTimes()
	conn.EXPECT().UID().Return("uid").AnyTimes()
	conn.EXPECT().ID().Return("id").AnyTimes()

	ms := NewMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	assert.Equal(t, int64(0), ms.Current())
	ms.Add()
	assert.Equal(t, int64(1), ms.Current())
}
