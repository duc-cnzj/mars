package socket

import (
	"sync"
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

type Percentable interface {
	Current() int64
	Add()
	To(percent int64)
}

var _ Percentable = (*processPercent)(nil)

type processPercent struct {
	msger DeployMsger

	s           Sleeper
	percentLock sync.RWMutex
	percent     int64
}

func NewProcessPercent(sender DeployMsger, s Sleeper) Percentable {
	return &processPercent{
		s:     s,
		msger: sender,
	}
}

func (pp *processPercent) Current() int64 {
	pp.percentLock.RLock()
	defer pp.percentLock.RUnlock()

	return pp.percent
}

func (pp *processPercent) Add() {
	pp.percentLock.Lock()
	defer pp.percentLock.Unlock()

	if pp.percent < 100 {
		pp.percent++
		pp.msger.SendProcessPercent(pp.percent)
	}
}

func (pp *processPercent) To(percent int64) {
	pp.percentLock.Lock()
	defer pp.percentLock.Unlock()

	sleepTime := 100 * time.Millisecond
	var step int64 = 2
	for pp.percent+step <= percent {
		pp.s.Sleep(sleepTime)
		pp.percent += step
		if sleepTime > 50*time.Millisecond {
			sleepTime = sleepTime / 2
		}
		pp.msger.SendProcessPercent(pp.percent)
	}
	if pp.percent != percent {
		pp.msger.SendProcessPercent(pp.percent)
		pp.percent = percent
	}
}
