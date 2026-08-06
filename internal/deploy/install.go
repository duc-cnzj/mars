package deploy

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/closeable"
	"github.com/duc-cnzj/mars/v6/internal/util/hasher"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

// InstallProject 运行一个 Job 的完整部署流水线：
// GlobalLock → Validate → LoadConfigs → Run → Finish，最终返回流水线错误。
// 这是 gRPC 服务与 websocket 传输共同使用的部署入口。
func InstallProject(ctx context.Context, job Job) (err error) {
	return job.GlobalLock().Validate().LoadConfigs().Run(ctx).Finish().Error()
}

// GetSlugName 由 namespaceID + 服务名生成部署任务标识，供传输层区分同一服务的多次部署。
func GetSlugName[T int64 | int | int32](namespaceId T, name string) string {
	return hasher.Hash(fmt.Sprintf("%d-%s", namespaceId, name))
}

// MessageItem 是部署进度的一条消息，经 SafeWriteMessageChan 从安装侧投递到 jobRunner。
type MessageItem struct {
	Msg  string
	Type MessageType

	Containers []*websocket_pb.Container
}

// MessageType 是部署进度消息的类型（成功/失败/文本）。
type MessageType uint8

// 消息类型取值：1 成功、2 失败、3 普通文本进度。
const (
	_ MessageType = iota
	MessageSuccess
	MessageError
	MessageText
)

// ReleaseInstaller 执行 helm 安装/升级并返回 release 结果。
type ReleaseInstaller interface {
	Run(ctx context.Context, input *InstallInput) (*release.Release, error)
}

var _ ReleaseInstaller = (*releaseInstaller)(nil)

type releaseInstaller struct {
	logger         mlog.Logger
	helmer         biz.HelmerRepo
	timeoutSeconds int64
	timer          timer.Timer
}

// NewReleaseInstaller 构造部署执行器，构造时从 config 读取默认安装超时兜底。
func NewReleaseInstaller(
	logger mlog.Logger,
	helmer biz.HelmerRepo,
	config *config.Config,
	timer timer.Timer,
) ReleaseInstaller {
	return &releaseInstaller{
		timer:          timer,
		logger:         logger,
		helmer:         helmer,
		timeoutSeconds: int64(config.InstallTimeout.Seconds()),
	}
}

// InstallInput 是 releaseInstaller 的一次安装输入，由 jobRunner.Run 组装。
type InstallInput struct {
	IsNew        bool
	Wait         bool
	Chart        *chart.Chart
	ValueOptions *values.Options
	DryRun       bool
	ReleaseName  string
	Namespace    string
	Description  string

	// TimeoutSeconds 调用方按请求覆盖超时；0 表示用构造时读到的默认值。
	TimeoutSeconds int64

	messageChan SafeWriteMessageChan
	percenter   Percentable
}

// Run 执行一次 helm 安装/升级；失败时按是否新项目选择回滚或卸载清理。
func (r *releaseInstaller) Run(
	ctx context.Context,
	input *InstallInput,
) (*release.Release, error) {
	defer r.logger.Debug("releaseInstaller exit")

	var logger = newTimeOrderedSetString(r.timer)
	wrapLogFn := r.loggerWrap(input.messageChan, input.percenter, logger)
	var re *release.Release
	var err error

	// 超时优先级：调用方按请求覆盖 > 构造时从 config 读到的默认值。
	timeout := r.timeoutSeconds
	if input.TimeoutSeconds > 0 {
		timeout = input.TimeoutSeconds
	}

	re, err = r.helmer.UpgradeOrInstall(
		ctx,
		input.ReleaseName,
		input.Namespace,
		input.Chart,
		input.ValueOptions,
		wrapLogFn,
		input.Wait,
		timeout,
		input.DryRun,
		input.Description,
	)
	if err == nil {
		r.logger.Debug(re)
		return re, nil
	}
	wrapLogFn(nil, "部署出现问题: %s", err)
	if !input.DryRun {
		var rollbackErr error
		if !input.IsNew {
			// 失败了，需要手动回滚
			r.logger.Debug("rollback project")
			rollbackErr = r.helmer.Rollback(input.ReleaseName, input.Namespace, false, wrapLogFn.UnWrap(), input.DryRun)
		} else {
			r.logger.Debug("uninstall project")
			rollbackErr = r.helmer.Uninstall(input.ReleaseName, input.Namespace, wrapLogFn.UnWrap())
		}
		if rollbackErr != nil {
			wrapLogFn(nil, "回滚出现问题: %s", rollbackErr)
			r.logger.Debug(rollbackErr)
		}
	}
	return nil, err
}

// loggerWrap 包装 helm 进度回调：推进进度、去重日志并通过 messageChan 投递。
func (r *releaseInstaller) loggerWrap(messageChan SafeWriteMessageChan, percenter Percentable, logs *timeOrderedSetString) biz.WrapLogFn {
	return func(containers []*websocket_pb.Container, format string, v ...any) {
		if percenter.Current() < 99 {
			percenter.Add()
		}

		msg := fmt.Sprintf(format, v...)

		if !logs.has(msg) {
			logs.add(msg)
			messageChan.Send(MessageItem{
				Msg:        msg,
				Containers: containers,
				Type:       MessageText,
			})
		}
	}
}

type timeOrderedSetStringItem struct {
	t    time.Time
	data string
}

type orderedItemList []*timeOrderedSetStringItem

// Len 实现 sort.Interface，返回元素个数。
func (o orderedItemList) Len() int {
	return len(o)
}

// Less 实现 sort.Interface，按时间升序排列。
func (o orderedItemList) Less(i, j int) bool {
	return o[i].t.Before(o[j].t)
}

// Swap 实现 sort.Interface，交换两个元素。
func (o orderedItemList) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

// List 返回按插入时间排序后的字符串列表。
func (o orderedItemList) List() []string {
	res := make([]string, len(o))
	for i, item := range o {
		res[i] = item.data
	}
	return res
}

type timeOrderedSetString struct {
	mu    sync.RWMutex
	items map[string]time.Time
	timer timer.Timer
}

// newTimeOrderedSetString 构造按插入时间排序的去重字符串集合。
func newTimeOrderedSetString(timer timer.Timer) *timeOrderedSetString {
	return &timeOrderedSetString{
		items: make(map[string]time.Time),
		timer: timer,
	}
}

// add 记录字符串 s 的首次插入时间；重复插入直接忽略。
func (o *timeOrderedSetString) add(s string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if _, ok := o.items[s]; ok {
		return
	}
	o.items[s] = o.timer.Now()
}

// has 判断字符串 s 是否已记录。
func (o *timeOrderedSetString) has(s string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	_, ok := o.items[s]
	return ok
}

// sortedItems 返回全部已记录字符串，按插入时间升序。
func (o *timeOrderedSetString) sortedItems() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	oslist := make(orderedItemList, 0, len(o.items))
	for s, t := range o.items {
		oslist = append(oslist, &timeOrderedSetStringItem{
			t:    t,
			data: s,
		})
	}
	sort.Sort(oslist)
	return oslist.List()
}

// SafeWriteMessageChan 是部署进度消息的线程安全投递通道，
// releaseInstaller 写、jobRunner.HandleMessage 读。
type SafeWriteMessageChan interface {
	Close()
	Chan() <-chan MessageItem
	Send(m MessageItem)
}

var _ SafeWriteMessageChan = (*safeWriteMessageCh)(nil)

type safeWriteMessageCh struct {
	logger    mlog.Logger
	closeable closeable.Closeable

	chMu sync.Mutex
	ch   chan MessageItem
	once sync.Once
}

// newSafeWriteMessageCh 创建容量为 chSize 的线程安全消息通道，
// 供 jobRunner 向传输层投递部署进度消息；仅包内使用。
func newSafeWriteMessageCh(logger mlog.Logger, chSize int) SafeWriteMessageChan {
	return &safeWriteMessageCh{
		logger: logger,
		ch:     make(chan MessageItem, chSize),
	}
}

// Close 关闭消息通道。用 sync.Once 保证幂等，只真正 close 一次；
// 关闭后再 Send 的消息会被静默丢弃。
func (s *safeWriteMessageCh) Close() {
	s.once.Do(func() {
		s.logger.Debug("safeWriteMessageCh closed")
		s.closeable.Close()
		s.chMu.Lock()
		defer s.chMu.Unlock()
		close(s.ch)
	})
}

// Chan 返回只读消息队列，消费方从其上取 MessageItem。
func (s *safeWriteMessageCh) Chan() <-chan MessageItem {
	return s.ch
}

// Send 向通道投递一条消息；通道已关闭或缓冲已满时静默丢弃并打日志，
// 绝不在写侧阻塞部署流水线。
func (s *safeWriteMessageCh) Send(m MessageItem) {
	s.chMu.Lock()
	defer s.chMu.Unlock()

	if s.closeable.IsClosed() {
		s.logger.Debugf("[Websocket]: Drop %s type %v", m.Msg, m.Type)
		return
	}

	s.logger.Debugf("Send message to channel: %v", m.Msg)

	select {
	case s.ch <- m:
	default:
		s.logger.Warningf("Channel is full, dropping message: %s type %v", m.Msg, m.Type)
	}
}
