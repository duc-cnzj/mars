package socket

import (
	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

type messager struct {
	conn     *WsConn
	slugName string
	wsType   websocket_pb.Type
}

func NewMessageSender(conn *WsConn, slugName string, wsType websocket_pb.Type) contracts.DeployMsger {
	return &messager{conn: conn, slugName: slugName, wsType: wsType}
}

func (ms *messager) SendDeployedResult(result websocket_pb.ResultType, msg string, p *types.ProjectModel) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:    ms.slugName,
			Type:    ms.wsType,
			Result:  result,
			End:     true,
			Uid:     ms.conn.uid,
			Id:      ms.conn.id,
			Message: msg,
		},
	}
	ms.send(res)
}

func (ms *messager) SendEndError(err error) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:    ms.slugName,
			Type:    ms.wsType,
			Result:  ResultError,
			End:     true,
			Uid:     ms.conn.uid,
			Id:      ms.conn.id,
			Message: err.Error(),
		},
	}
	ms.send(res)
}

func (ms *messager) SendError(err error) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:    ms.slugName,
			Type:    ms.wsType,
			Result:  ResultError,
			End:     false,
			Uid:     ms.conn.uid,
			Id:      ms.conn.id,
			Message: err.Error(),
		},
	}
	ms.send(res)
}

func (ms *messager) SendProcessPercent(percent int64) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:    ms.slugName,
			Type:    WsProcessPercent,
			Result:  ResultSuccess,
			End:     false,
			Uid:     ms.conn.uid,
			Id:      ms.conn.id,
			Percent: percent,
		},
	}
	ms.send(res)
}

func (ms *messager) SendMsg(msg string) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:    ms.slugName,
			Type:    ms.wsType,
			Result:  ResultSuccess,
			End:     false,
			Uid:     ms.conn.uid,
			Id:      ms.conn.id,
			Message: msg,
		},
	}
	ms.send(res)
}

func (ms *messager) SendMsgWithContainerLog(msg string, containers []*types.Container) {
	res := &websocket_pb.WsWithContainerMessageResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:    ms.slugName,
			Type:    ms.wsType,
			Result:  ResultLogWithContainers,
			End:     false,
			Uid:     ms.conn.uid,
			Id:      ms.conn.id,
			Message: msg,
		},
		Containers: containers,
	}
	ms.send(res)
}

func (ms *messager) SendProtoMsg(msg contracts.WebsocketMessage) {
	ms.send(msg)
}

func (ms *messager) send(res contracts.WebsocketMessage) {
	ms.conn.pubSub.ToSelf(res)
}
