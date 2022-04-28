package socket

import (
	"sync"

	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
)

type Msger interface {
	SendEndError(error)
	SendError(error)
	SendMsg(string)
	SendProtoMsg(plugins.WebsocketMessage)
}

type ProcessPercentMsger interface {
	SendProcessPercent(string)
}

type DeployMsger interface {
	Msger
	ProcessPercentMsger

	Stop(error)
	SendDeployedResult(websocket_pb.ResultType, string, *models.Project)
}

type messager struct {
	mu        sync.RWMutex
	isStopped bool
	stoperr   error

	conn     *WsConn
	slugName string
	wsType   websocket_pb.Type
}

func NewMessageSender(conn *WsConn, slugName string, wsType websocket_pb.Type) DeployMsger {
	return &messager{conn: conn, slugName: slugName, wsType: wsType}
}

func (ms *messager) Stop(err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.stoperr = err
	ms.isStopped = true
}

func (ms *messager) IsStopped() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.isStopped
}

func (ms *messager) SendDeployedResult(result websocket_pb.ResultType, msg string, project *models.Project) {
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

func (ms *messager) SendProcessPercent(percent string) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:    ms.slugName,
			Type:    WsProcessPercent,
			Result:  ResultSuccess,
			End:     false,
			Uid:     ms.conn.uid,
			Id:      ms.conn.id,
			Message: percent,
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

func (ms *messager) SendProtoMsg(msg plugins.WebsocketMessage) {
	ms.send(msg)
}

func (ms *messager) send(res plugins.WebsocketMessage) {
	if ms.IsStopped() {
		return
	}
	ms.conn.pubSub.ToSelf(res)
}
