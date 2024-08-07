package socket

import (
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/application"
)

type DeployMsger interface {
	Percentable

	SendProcessPercent(int64)
	SendDeployedResult(t websocket_pb.ResultType, msg string, p *types.ProjectModel)
	SendEndError(error)
	SendMsg(string)
	SendProtoMsg(application.WebsocketMessage)
	SendMsgWithContainerLog(msg string, containers []*websocket_pb.Container)
}

var _ DeployMsger = (*messager)(nil)

type messager struct {
	conn     Conn
	slugName string
	wsType   websocket_pb.Type

	percent Percentable
}

func NewMessageSender(
	conn Conn,
	slugName string,
	wsType websocket_pb.Type,
) DeployMsger {
	m := &messager{
		conn:     conn,
		slugName: slugName,
		wsType:   wsType,
	}
	m.percent = NewProcessPercent(m, NewRealSleeper())
	return m
}

func (ms *messager) SendDeployedResult(result websocket_pb.ResultType, msg string, p *types.ProjectModel) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:    ms.slugName,
			Type:    ms.wsType,
			Result:  result,
			End:     true,
			Uid:     ms.conn.UID(),
			Id:      ms.conn.ID(),
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
			Uid:     ms.conn.UID(),
			Id:      ms.conn.ID(),
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
			Uid:     ms.conn.UID(),
			Id:      ms.conn.ID(),
			Percent: int32(percent),
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
			Uid:     ms.conn.UID(),
			Id:      ms.conn.ID(),
			Message: msg,
		},
	}
	ms.send(res)
}

func (ms *messager) SendMsgWithContainerLog(msg string, containers []*websocket_pb.Container) {
	res := &websocket_pb.WsWithContainerMessageResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:    ms.slugName,
			Type:    ms.wsType,
			Result:  ResultLogWithContainers,
			End:     false,
			Uid:     ms.conn.UID(),
			Id:      ms.conn.ID(),
			Message: msg,
		},
		Containers: containers,
	}
	ms.send(res)
}

func (ms *messager) SendProtoMsg(msg application.WebsocketMessage) {
	ms.send(msg)
}

func (ms *messager) send(res application.WebsocketMessage) {
	ms.conn.PubSub().ToSelf(res)
}

func (ms *messager) Current() int64 {
	return ms.percent.Current()
}

func (ms *messager) Add() {
	ms.percent.Add()
}

func (ms *messager) To(percent int64) {
	ms.percent.To(percent)
}

type Sleeper interface {
	Sleep(time.Duration)
}

type realSleeper struct{}

func NewRealSleeper() Sleeper {
	return &realSleeper{}
}

func (r *realSleeper) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

var _ Percentable = (*processPercent)(nil)
