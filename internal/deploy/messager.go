package deploy

import (
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/application"
)

// DeployMsger 是部署流水线向用户汇报进度的端口。实现方由传输层提供：
// websocket 侧用 messageSender，gRPC 侧用 services 里的 messager/emptyMessager。
type DeployMsger interface {
	Percentable

	SendProcessPercent(int64)
	SendDeployedResult(t websocket_pb.ResultType, msg string, p *types.ProjectModel)
	SendEndError(error)
	SendMsg(string)
	SendProtoMsg(application.WebsocketMessage)
	SendMsgWithContainerLog(msg string, containers []*websocket_pb.Container)
}

// Sleeper 是休眠抽象，进度上报的节流间隔可注入（测试用假休眠器）。
type Sleeper interface {
	Sleep(time.Duration)
}

type realSleeper struct{}

// NewRealSleeper 返回真实休眠实现，测试可注入 mock Sleeper 以加速。
func NewRealSleeper() Sleeper {
	return &realSleeper{}
}

// Sleep 真实休眠指定时长。
func (r *realSleeper) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

var _ Percentable = (*processPercent)(nil)

// Percentable 是进度追踪器端口：Current 查询、Add 步进、To 定位到目标百分比。
type Percentable interface {
	Current() int64
	Add()
	To(percent int64)
}

// processPercent 是通用的进度追踪器：Add 每次 +1 并上报，To(percent) 平滑推进到目标百分比。
// 它只依赖 DeployMsger 端口，不绑定任何传输实现，供所有 DeployMsger 适配器复用。
type processPercent struct {
	msger DeployMsger

	s           Sleeper
	percentLock sync.RWMutex
	percent     int64
}

// NewProcessPercent 构造进度追踪器，进度变化通过 sender 上报给用户。
func NewProcessPercent(sender DeployMsger, s Sleeper) Percentable {
	return &processPercent{
		s:     s,
		msger: sender,
	}
}

// Current 返回当前进度百分比。
func (pp *processPercent) Current() int64 {
	pp.percentLock.RLock()
	defer pp.percentLock.RUnlock()

	return pp.percent
}

// Add 进度 +1 并上报，封顶 100。
func (pp *processPercent) Add() {
	pp.percentLock.Lock()
	defer pp.percentLock.Unlock()

	if pp.percent < 100 {
		pp.percent++
		pp.msger.SendProcessPercent(pp.percent)
	}
}

// To 平滑推进进度到 percent：先按递减步长逼近，最后精确对齐目标值。
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
