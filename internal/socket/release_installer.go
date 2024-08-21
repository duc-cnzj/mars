package socket

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/util/timer"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

type MessageItem struct {
	Msg  string
	Type MessageType

	Containers []*websocket_pb.Container
}

type MessageType uint8

const (
	_ MessageType = iota
	MessageSuccess
	MessageError
	MessageText
)

type ReleaseInstaller interface {
	Run(ctx context.Context, input *InstallInput) (*release.Release, error)
}

var _ ReleaseInstaller = (*releaseInstaller)(nil)

type releaseInstaller struct {
	logger         mlog.Logger
	helmer         repo.HelmerRepo
	timeoutSeconds int64
	timer          timer.Timer
}

func NewReleaseInstaller(
	logger mlog.Logger,
	helmer repo.HelmerRepo,
	data data.Data,
	timer timer.Timer,
) ReleaseInstaller {
	return &releaseInstaller{
		timer:          timer,
		logger:         logger,
		helmer:         helmer,
		timeoutSeconds: int64(data.Config().InstallTimeout.Seconds()),
	}
}

type InstallInput struct {
	IsNew        bool
	Wait         bool
	Chart        *chart.Chart
	ValueOptions *values.Options
	DryRun       bool
	ReleaseName  string
	Namespace    string
	Description  string

	messageChan SafeWriteMessageChan
	percenter   Percentable
}

func (r *releaseInstaller) Run(
	ctx context.Context,
	input *InstallInput,
) (*release.Release, error) {
	defer r.logger.Debug("releaseInstaller exit")

	var logger = newTimeOrderedSetString(r.timer)
	wrapLogFn := r.loggerWrap(input.messageChan, input.percenter, logger)
	re, err := r.helmer.UpgradeOrInstall(
		ctx,
		input.ReleaseName,
		input.Namespace,
		input.Chart,
		input.ValueOptions,
		wrapLogFn,
		input.Wait,
		r.timeoutSeconds,
		input.DryRun,
		input.Description,
	)
	if err == nil {
		r.logger.Debug(err)
		return re, nil
	}
	if !input.DryRun {
		if !input.IsNew {
			// 失败了，需要手动回滚
			r.logger.Debug("rollback project")
			if err = r.helmer.Rollback(input.ReleaseName, input.Namespace, false, wrapLogFn.UnWrap(), input.DryRun); err != nil {
				r.logger.Debug(err)
			}
		} else {
			r.logger.Debug("uninstall project")
			if err = r.helmer.Uninstall(input.ReleaseName, input.Namespace, wrapLogFn.UnWrap()); err != nil {
				r.logger.Debug(err)
			}
		}
	}
	return nil, err
}

func (r *releaseInstaller) loggerWrap(messageChan SafeWriteMessageChan, percenter Percentable, logs *timeOrderedSetString) repo.WrapLogFn {
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

func (o orderedItemList) Len() int {
	return len(o)
}

func (o orderedItemList) Less(i, j int) bool {
	return o[i].t.Before(o[j].t)
}

func (o orderedItemList) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

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

func newTimeOrderedSetString(timer timer.Timer) *timeOrderedSetString {
	return &timeOrderedSetString{
		items: make(map[string]time.Time),
		timer: timer,
	}
}

func (o *timeOrderedSetString) add(s string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if _, ok := o.items[s]; ok {
		return
	}
	o.items[s] = o.timer.Now()
}

func (o *timeOrderedSetString) has(s string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	_, ok := o.items[s]
	return ok
}

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
