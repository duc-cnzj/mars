package services

import (
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewContainerSvc(t *testing.T) {
	svc, _ := newContainerSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.eventBiz)
	assert.NotNil(t, svc.k8sBiz)
	assert.NotNil(t, svc.fileBiz)
	assert.NotNil(t, svc.logger)
}

func Test_containerSvc_IsPodRunning(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(false, "")
	running, err := svc.IsPodRunning(newAdminUserCtx(), &container.IsPodRunningRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.False(t, running.Running)
	assert.Nil(t, err)
}

func Test_containerSvc_IsPodRunning_PermissionDenied(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Private: true}, nil).AnyTimes()
	_, err := svc.IsPodRunning(newOtherUserCtx(), &container.IsPodRunningRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func Test_containerSvc_IsPodExists(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(nil, nil)
	running, err := svc.IsPodExists(newAdminUserCtx(), &container.IsPodExistsRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.True(t, running.Exists)
	assert.Nil(t, err)
}

func Test_containerSvc_IsPodExists_Fail(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(nil, errors.New("X"))
	running, err := svc.IsPodExists(newAdminUserCtx(), &container.IsPodExistsRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.False(t, running.Exists)
	assert.Nil(t, err)
}

func Test_containerSvc_IsPodExists_NotFound(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	// pod 不存在（errs.NotFound）：探活接口静默返回 Exists:false，不记 Error 日志。
	k8sRepo.EXPECT().GetPod("a", "b").Return(nil, errs.NotFound("pod not found"))
	running, err := svc.IsPodExists(newAdminUserCtx(), &container.IsPodExistsRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.False(t, running.Exists)
	assert.Nil(t, err)
}

func TestContainerSvc_ContainerLog_PodNotFound(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(nil, nil)
	_, err := svc.ContainerLog(newAdminUserCtx(), &container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.NotNil(t, err)
}

func TestContainerSvc_ContainerLog_PodPending(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodPending}}, nil)
	_, err := svc.ContainerLog(newAdminUserCtx(), &container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.Error(t, err)
}

func TestContainerSvc_ContainerLog_PodRunning(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning}}, nil)
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), "a", "b", &v1.PodLogOptions{
		TailLines: &tailLines,
		Container: "c",
	}).Return("log", nil)
	_, err := svc.ContainerLog(newAdminUserCtx(), &container.LogRequest{
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	})
	assert.Nil(t, err)
}

// 终止（Succeeded/Failed）pod 的日志同样可能巨大，必须与 Running 一样只取尾部，
// 否则全量读进内存有 OOM 风险。锁定 TailLines 恒被设置的行为。
func TestContainerSvc_ContainerLog_PodSucceeded_TailLines(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodSucceeded}}, nil)
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), "a", "b", &v1.PodLogOptions{
		TailLines: &tailLines,
		Container: "c",
	}).Return("log", nil)
	_, err := svc.ContainerLog(newAdminUserCtx(), &container.LogRequest{
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	})
	assert.Nil(t, err)
}

func TestContainerSvc_ContainerLog_GetPodLogs_error(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning}}, nil)
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), "a", "b", gomock.Any()).Return("", errors.New("x"))
	_, err := svc.ContainerLog(newAdminUserCtx(), &container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.Equal(t, "x", err.Error())
}

func TestContainerSvc_ContainerLog_PodPending1(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodPending}}, nil)
	k8sRepo.EXPECT().ListEvents(gomock.Any()).Return([]*eventv1.Event{
		{
			Regarding: v1.ObjectReference{Kind: "Pod", Name: "b"},
			Note:      "aaa",
		},
		{
			Regarding: v1.ObjectReference{Kind: "Pod", Name: "b"},
			Note:      "bbb",
		},
	}, nil)
	resp, err := svc.ContainerLog(newAdminUserCtx(), &container.LogRequest{
		Namespace:  "a",
		Pod:        "b",
		ShowEvents: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, "aaa\nbbb", resp.Log)
}

func TestContainerSvc_ContainerLog_Pending_EventFilterIsolation(t *testing.T) {
	// 回归防护：Pending + ShowEvents 时，只收集「本 pod 且 Kind==Pod」的事件，
	// 其他 pod / 其他 Kind 的事件不得混入（隔离语义）。改坏过滤条件
	// （去掉 Kind 或 pod 名判断）时此测试 FAIL。
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodPending}}, nil)
	k8sRepo.EXPECT().ListEvents(gomock.Any()).Return([]*eventv1.Event{
		{Regarding: v1.ObjectReference{Kind: "Pod", Name: "b"}, Note: "match"},
		{Regarding: v1.ObjectReference{Kind: "Pod", Name: "other-pod"}, Note: "wrong-pod"},
		{Regarding: v1.ObjectReference{Kind: "Deployment", Name: "b"}, Note: "wrong-kind"},
	}, nil)
	resp, err := svc.ContainerLog(newAdminUserCtx(), &container.LogRequest{
		Namespace:  "a",
		Pod:        "b",
		ShowEvents: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, "match", resp.Log)
}

// tailLines 是 PodLogOptions.TailLines 的取址对象（gomock.Eq 按值 DeepEqual，与 biz 侧常量等价）。
var tailLines int64 = 1000

type logStreamServer struct {
	ctx context.Context
	container.Container_StreamContainerLogServer
	res     []string
	sendErr error
}

func (l *logStreamServer) Send(response *container.LogResponse) error {
	l.res = append(l.res, response.Log)
	return l.sendErr
}

func (l *logStreamServer) Context() context.Context {
	if l.ctx != nil {
		return l.ctx
	}
	return newAdminUserCtx()
}

func TestContainerSvc_CopyToPod_PodNotRunning(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(false, "")
	_, err := svc.CopyToPod(newAdminUserCtx(), &container.CopyToPodRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.NotNil(t, err)
}

func TestContainerSvc_CopyToPod_Success(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	eventRepo := mocks.eventRepo
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	eventRepo.EXPECT().FileAuditLog(
		types.EventActionType_Upload,
		biz.MustGetUser(newAdminUserCtx()).Name,
		gomock.Any(),
		11,
	)
	k8sRepo.EXPECT().CopyFileToPod(gomock.Any(), &biz.CopyFileToPodInput{
		FileId:    1,
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}).Return(&biz.File{ID: 11}, nil)
	_, err := svc.CopyToPod(newAdminUserCtx(), &container.CopyToPodRequest{
		FileId:    1,
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	})
	assert.Nil(t, err)
}
func TestContainerSvc_CopyToPod_Error(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	k8sRepo.EXPECT().CopyFileToPod(gomock.Any(), &biz.CopyFileToPodInput{
		FileId:    1,
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}).Return(nil, errors.New("xx"))
	_, err := svc.CopyToPod(newAdminUserCtx(), &container.CopyToPodRequest{
		FileId:    1,
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	})
	assert.Equal(t, "xx", err.Error())
}

// 回归防护：CopyToPod 空 container 必须走 ResolveContainer 找默认容器，
// 与 StreamCopyToPod 的"空则找默认"语义对齐。去掉 ResolveContainer 时本测试
// 会因 FindDefaultContainer 未被调用而失败。
func TestContainerSvc_CopyToPod_DefaultContainer(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	eventRepo := mocks.eventRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("c", nil)
	eventRepo.EXPECT().FileAuditLog(
		types.EventActionType_Upload,
		biz.MustGetUser(newAdminUserCtx()).Name,
		gomock.Any(),
		11,
	)
	k8sRepo.EXPECT().CopyFileToPod(gomock.Any(), &biz.CopyFileToPodInput{
		FileId:    1,
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}).Return(&biz.File{ID: 11}, nil)
	_, err := svc.CopyToPod(newAdminUserCtx(), &container.CopyToPodRequest{
		FileId:    1,
		Namespace: "a",
		Pod:       "b",
	})
	assert.Nil(t, err)
}

// CopyToPod 在"空 container + 默认容器解析失败"时 fail-fast：不触达 CopyFileToPod。
func TestContainerSvc_CopyToPod_ResolveContainerError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("", errors.New("no default"))
	_, err := svc.CopyToPod(newAdminUserCtx(), &container.CopyToPodRequest{
		FileId:    1,
		Namespace: "a",
		Pod:       "b",
	})
	assert.Equal(t, "no default", err.Error())
}

func TestContainerSvc_StreamContainerLog_PodNotFound(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(nil, nil)
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, &logStreamServer{})
	assert.NotNil(t, err)
}

func TestContainerSvc_StreamContainerLog_PodPending(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodPending}}, nil)
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, &logStreamServer{})
	assert.Error(t, err)
}

func TestContainerSvc_StreamContainerLog_PodRunning(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning}}, nil)
	ch := make(chan []byte, 2)
	ch <- []byte("log1")
	ch <- []byte("log2")
	close(ch)
	k8sRepo.EXPECT().LogStream(gomock.Any(), "a", "b", "c").Return(ch, nil)
	s := &logStreamServer{}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}, s)
	assert.Nil(t, err)
	assert.Equal(t, []string{"log1", "log2"}, s.res)
}

func TestContainerSvc_StreamContainerLog_PodSucceeded(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodSucceeded}}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), "a", "b", gomock.Any()).Return("log", nil)
	s := &logStreamServer{}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, s)
	assert.Nil(t, err)
	assert.Equal(t, []string{"log"}, s.res)
}

// scannerText 回调里 Send 失败（客户端断连）时只打日志，scannerText 仍返回 nil。
func TestContainerSvc_StreamContainerLog_SendErrorInScanner(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodSucceeded}}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), "a", "b", gomock.Any()).Return("log", nil)
	s := &logStreamServer{sendErr: errors.New("send boom")}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, s)
	assert.Nil(t, err)
	assert.Equal(t, []string{"log"}, s.res)
}

func TestContainerSvc_StreamContainerLog_Error(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodSucceeded}}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), "a", "b", gomock.Any()).Return("", errors.New("x"))
	s := &logStreamServer{}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, s)
	assert.Equal(t, "x", err.Error())
}

func TestContainerSvc_StreamContainerLog_PodFailed(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodFailed}}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), "a", "b", gomock.Any()).Return("log", nil)
	s := &logStreamServer{}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, s)
	assert.Nil(t, err)
	assert.Equal(t, []string{"log"}, s.res)
}

func TestContainerSvc_StreamContainerLog_PodPending1(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodPending}}, nil).AnyTimes()
	s := &logStreamServer{}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, s)

	assert.Equal(t, "未找到日志", status.Convert(err).Message())
}

func TestContainerSvc_StreamContainerLog_PodPending_ShowEvents(t *testing.T) {
	// 回归防护：Pending + ShowEvents=true 必须走 ContainerLog 事件路径（container.go
	// line 300 的 PodPending 分组），把 ListEvents 过滤后的 Note 逐行流式下发；而非落到
	// LogStream 实时流（实时流对 Pending pod 无日志可读，行为完全不同）。变异去掉
	// line 300 的 PodPending 分组后，Pending pod 会进入 LogStream → gomock "未期望调用"
	// FAIL；同时 wrong-pod 事件不得混入流中。
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodPending}}, nil).AnyTimes()
	k8sRepo.EXPECT().ListEvents("a").Return([]*eventv1.Event{
		{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "2"}, Regarding: v1.ObjectReference{Kind: "Pod", Name: "b"}, Note: "match"},
		{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"}, Regarding: v1.ObjectReference{Kind: "Pod", Name: "other-pod"}, Note: "wrong-pod"},
	}, nil)
	s := &logStreamServer{}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace:  "a",
		Pod:        "b",
		ShowEvents: true,
	}, s)
	assert.Nil(t, err)
	assert.Equal(t, []string{"match"}, s.res)
}

type streamCopyToPodServer struct {
	container.Container_StreamCopyToPodServer
	ctx  context.Context
	err  error
	idx  int
	recv []*container.StreamCopyToPodRequest
}

func (l *streamCopyToPodServer) Send(response *container.StreamCopyToPodResponse) error {
	return nil
}

func (l *streamCopyToPodServer) SendAndClose(response *container.StreamCopyToPodResponse) error {
	return nil
}

func (l *streamCopyToPodServer) Recv() (*container.StreamCopyToPodRequest, error) {
	if l.err != nil {
		return nil, l.err
	}
	if l.idx < len(l.recv) {
		l.idx++
		return l.recv[l.idx-1], nil
	}
	return nil, io.EOF
}

func (l *streamCopyToPodServer) Context() context.Context {
	if l.ctx != nil {
		return l.ctx
	}
	return newAdminUserCtx()
}

func TestContainerSvc_StreamCopyToPod_PodNotRunning(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(false, "")
	err := svc.StreamCopyToPod(&streamCopyToPodServer{recv: []*container.StreamCopyToPodRequest{
		{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
	}})
	// pod 未运行 → NotFound，与 CopyToPod/metrics 的语义对齐。
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestContainerSvc_StreamCopyToPod_Error(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	err := svc.StreamCopyToPod(&streamCopyToPodServer{err: errors.New("xx")})
	assert.Equal(t, "xx", err.Error())
}

func TestContainerSvc_StreamCopyToPod_Success(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo
	eventRepo.EXPECT().FileAuditLog(
		types.EventActionType_Upload,
		biz.MustGetUser(newAdminUserCtx()).Name,
		gomock.Any(),
		1,
	)
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("c", nil)
	fileRepo.EXPECT().StreamUploadFile(gomock.Any(), gomock.Cond(func(x any) bool {
		v, ok := x.(*biz.StreamUploadFileRequest)
		if !ok {
			return false
		}
		return v.Namespace == "a" &&
			v.Pod == "b" &&
			v.Container == "c" &&
			v.Username == biz.MustGetUser(newAdminUserCtx()).Name &&
			v.FileName == "a.txt"
	})).Return(&biz.File{
		ID:        1,
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}, nil)
	k8sRepo.EXPECT().CopyFileToPod(gomock.Any(), &biz.CopyFileToPodInput{
		FileId:    1,
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}).Return(&biz.File{}, nil)
	err := svc.StreamCopyToPod(&streamCopyToPodServer{recv: []*container.StreamCopyToPodRequest{
		{
			Namespace: "a",
			Pod:       "b",
			Container: "",
			FileName:  "a.txt",
			Data:      []byte("data"),
		},
	}})
	assert.Nil(t, err)
}

type execOnceServer struct {
	container.Container_ExecOnceServer
	ctx   context.Context
	res   []string
	Error *container.ExecError
}

func (l *execOnceServer) Context() context.Context {
	if l.ctx != nil {
		return l.ctx
	}
	return newAdminUserCtx()
}

func (l *execOnceServer) Send(response *container.ExecResponse) error {
	l.res = append(l.res, string(response.Message))
	l.Error = response.Error
	return nil
}

func TestContainerSvc_ExecOnce_PodNotRunning(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(false, "")
	err := svc.ExecOnce(&container.ExecOnceRequest{
		Namespace: "a",
		Pod:       "b",
	}, &execOnceServer{})
	// pod 未运行 → NotFound，与 CopyToPod/metrics 的语义对齐。
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestContainerSvc_ExecOnce_Success(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	eventRepo := mocks.eventRepo
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("c", nil)
	eventRepo.EXPECT().AuditLogWithChange(
		types.EventActionType_Exec,
		"admin",
		gomock.Any(),
		nil,
		gomock.Cond(func(x any) bool {
			v, ok := x.(biz.AnyYamlPrettier)
			if !ok {
				return false
			}
			ns, ok := v["namespace"].(string)
			if !ok {
				return false
			}
			pod, ok := v["pod"].(string)
			if !ok {
				return false
			}
			containerName, ok := v["container"].(string)
			if !ok {
				return false
			}
			command, ok := v["command"].([]string)
			if !ok {
				return false
			}
			return ns == "a" &&
				pod == "b" &&
				containerName == "c" &&
				slices.Equal(command, []string{"ls"}) &&
				v["error"] == "xx" &&
				v["result"] == ""
		}),
	)

	mac := &execOnceMatcher{
		tty: false,
		cmd: []string{"ls"},
	}
	k8sRepo.EXPECT().Execute(gomock.Any(), &biz.Container{
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}, mac).Return(&biz.ExecExitError{
		Code:    1,
		Message: "xx",
	})
	ser := &execOnceServer{}
	err := svc.ExecOnce(&container.ExecOnceRequest{
		Namespace: "a",
		Pod:       "b",
		Command:   []string{"ls"},
	}, ser)
	assert.Error(t, err)
	assert.Equal(t, int64(1), ser.Error.Code)
	assert.Equal(t, "xx", ser.Error.Message)
}

type execOnceMatcher struct {
	input *biz.ExecuteInput
	tty   bool
	cmd   []string
}

func (e *execOnceMatcher) Matches(x any) bool {
	input, ok := x.(*biz.ExecuteInput)
	if !ok {
		return false
	}
	e.input = input
	if e.tty != input.TTY {
		return false
	}
	if !slices.Equal(e.cmd, input.Cmd) {
		return false
	}
	return true
}

func (e *execOnceMatcher) String() string {
	return ""
}

// --- 分支补充:权限/错误路径 ---

func Test_containerSvc_CheckNamespaceAccess_FindByNameError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(nil, errors.New("boom"))

	_, err := svc.IsPodRunning(newAdminUserCtx(), &container.IsPodRunningRequest{Namespace: "a", Pod: "b"})
	assert.Equal(t, "boom", err.Error())
}

func TestContainerSvc_ContainerLog_PermissionDenied(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Private: true}, nil).AnyTimes()

	_, err := svc.ContainerLog(newOtherUserCtx(), &container.LogRequest{Namespace: "a", Pod: "b"})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestContainerSvc_ContainerLog_GetPodError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()

	k8sRepo.EXPECT().GetPod("a", "b").Return(nil, errors.New("x"))
	_, err := svc.ContainerLog(newAdminUserCtx(), &container.LogRequest{Namespace: "a", Pod: "b"})
	assert.Error(t, err)
}

func TestContainerSvc_ContainerLog_Pending_ListEventsError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()

	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodPending}}, nil)
	k8sRepo.EXPECT().ListEvents(gomock.Any()).Return(nil, errors.New("boom"))
	resp, err := svc.ContainerLog(newAdminUserCtx(), &container.LogRequest{Namespace: "a", Pod: "b", ShowEvents: true})
	assert.Nil(t, err)
	assert.Equal(t, "", resp.Log)
}

func TestContainerSvc_CopyToPod_PermissionDenied(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Private: true}, nil).AnyTimes()

	_, err := svc.CopyToPod(newOtherUserCtx(), &container.CopyToPodRequest{Namespace: "a", Pod: "b"})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestContainerSvc_StreamCopyToPod_PermissionDenied(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Private: true}, nil).AnyTimes()

	err := svc.StreamCopyToPod(&streamCopyToPodServer{ctx: newOtherUserCtx(), recv: []*container.StreamCopyToPodRequest{
		{Namespace: "a", Pod: "b", Container: "c"},
	}})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestContainerSvc_StreamCopyToPod_FindDefaultContainerError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()

	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("", errors.New("boom"))
	err := svc.StreamCopyToPod(&streamCopyToPodServer{recv: []*container.StreamCopyToPodRequest{
		{Namespace: "a", Pod: "b", Container: "", FileName: "a.txt"},
	}})
	assert.Equal(t, "boom", err.Error())
}

func TestContainerSvc_StreamCopyToPod_StreamUploadError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	fileRepo := mocks.fileRepo

	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	fileRepo.EXPECT().StreamUploadFile(gomock.Any(), gomock.Any()).Return(nil, errors.New("upload boom"))
	err := svc.StreamCopyToPod(&streamCopyToPodServer{recv: []*container.StreamCopyToPodRequest{
		{Namespace: "a", Pod: "b", Container: "c", FileName: "a.txt", Data: []byte("d")},
	}})
	assert.Equal(t, "upload boom", err.Error())
}

func TestContainerSvc_StreamCopyToPod_CopyError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	fileRepo := mocks.fileRepo

	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	fileRepo.EXPECT().StreamUploadFile(gomock.Any(), gomock.Any()).Return(&biz.File{ID: 1, Namespace: "a", Pod: "b", Container: "c"}, nil)
	k8sRepo.EXPECT().CopyFileToPod(gomock.Any(), gomock.Any()).Return(nil, errors.New("copy boom"))
	err := svc.StreamCopyToPod(&streamCopyToPodServer{recv: []*container.StreamCopyToPodRequest{
		{Namespace: "a", Pod: "b", Container: "c", FileName: "a.txt", Data: []byte("d")},
	}})
	assert.Equal(t, "copy boom", err.Error())
}

func TestContainerSvc_StreamContainerLog_PermissionDenied(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Private: true}, nil).AnyTimes()

	err := svc.StreamContainerLog(&container.LogRequest{Namespace: "a", Pod: "b"}, &logStreamServer{ctx: newOtherUserCtx()})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestContainerSvc_StreamContainerLog_GetPodError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()

	k8sRepo.EXPECT().GetPod("a", "b").Return(nil, errors.New("x"))
	err := svc.StreamContainerLog(&container.LogRequest{Namespace: "a", Pod: "b"}, &logStreamServer{})
	assert.Error(t, err)
}

func TestContainerSvc_StreamContainerLog_LogStreamError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()

	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning}}, nil)
	k8sRepo.EXPECT().LogStream(gomock.Any(), "a", "b", "c").Return(nil, errors.New("stream boom"))
	err := svc.StreamContainerLog(&container.LogRequest{Namespace: "a", Pod: "b", Container: "c"}, &logStreamServer{})
	assert.Equal(t, "stream boom", err.Error())
}

type logStreamServerErr struct {
	container.Container_StreamContainerLogServer
}

func (l *logStreamServerErr) Send(response *container.LogResponse) error {
	return errors.New("send boom")
}

func (l *logStreamServerErr) Context() context.Context {
	return newAdminUserCtx()
}

func TestContainerSvc_StreamContainerLog_SendError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()

	k8sRepo.EXPECT().GetPod("a", "b").Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning}}, nil)
	ch := make(chan []byte, 1)
	ch <- []byte("log1")
	close(ch)
	k8sRepo.EXPECT().LogStream(gomock.Any(), "a", "b", "c").Return(ch, nil)
	err := svc.StreamContainerLog(&container.LogRequest{Namespace: "a", Pod: "b", Container: "c"}, &logStreamServerErr{})
	assert.Equal(t, "send boom", err.Error())
}

func TestContainerSvc_ExecOnce_PermissionDenied(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Private: true}, nil).AnyTimes()

	err := svc.ExecOnce(&container.ExecOnceRequest{Namespace: "a", Pod: "b"}, &execOnceServer{ctx: newOtherUserCtx()})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestContainerSvc_ExecOnce_FindDefaultContainerError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()

	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("", errors.New("boom"))
	err := svc.ExecOnce(&container.ExecOnceRequest{Namespace: "a", Pod: "b"}, &execOnceServer{})
	assert.Equal(t, "boom", err.Error())
}

func TestContainerSvc_Exec_RecvError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()

	err := svc.Exec(&execServerRecvErr{})
	assert.Equal(t, "recv boom", err.Error())
}

func TestContainerSvc_Exec_PermissionDenied(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Private: true}, nil).AnyTimes()

	err := svc.Exec(&execServerMock{})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestContainerSvc_Exec_FindDefaultContainerError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()

	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("", errors.New("boom"))
	err := svc.Exec(&execServerMock{})
	assert.Equal(t, "boom", err.Error())
}

func Test_containerSvc_IsPodExists_PermissionDenied(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Private: true}, nil).AnyTimes()

	_, err := svc.IsPodExists(newOtherUserCtx(), &container.IsPodExistsRequest{Namespace: "a", Pod: "b"})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestContainerSvc_StreamCopyToPod_MultipleMessages(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().FileAuditLog(types.EventActionType_Upload, "admin", gomock.Any(), 1)
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	fileRepo.EXPECT().StreamUploadFile(gomock.Any(), gomock.Any()).Return(&biz.File{ID: 1, Namespace: "a", Pod: "b", Container: "c"}, nil)
	k8sRepo.EXPECT().CopyFileToPod(gomock.Any(), gomock.Any()).Return(&biz.File{}, nil)
	err := svc.StreamCopyToPod(&streamCopyToPodServer{recv: []*container.StreamCopyToPodRequest{
		{Namespace: "a", Pod: "b", Container: "c", FileName: "a.txt", Data: []byte("d1")},
		{Namespace: "a", Pod: "b", Container: "c", FileName: "a.txt", Data: []byte("d2")},
	}})
	assert.Nil(t, err)
}

type streamCopyToPodServerRecvErr struct {
	container.Container_StreamCopyToPodServer
	count int
}

func (l *streamCopyToPodServerRecvErr) Recv() (*container.StreamCopyToPodRequest, error) {
	l.count++
	if l.count == 1 {
		return &container.StreamCopyToPodRequest{Namespace: "a", Pod: "b", Container: "c", FileName: "a.txt", Data: []byte("d")}, nil
	}
	return nil, errors.New("recv boom")
}

func (l *streamCopyToPodServerRecvErr) Send(*container.StreamCopyToPodResponse) error {
	return nil
}

func (l *streamCopyToPodServerRecvErr) SendAndClose(*container.StreamCopyToPodResponse) error {
	return nil
}

func (l *streamCopyToPodServerRecvErr) Context() context.Context {
	return newAdminUserCtx()
}

func TestContainerSvc_StreamCopyToPod_RecvError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().FileAuditLog(types.EventActionType_Upload, "admin", gomock.Any(), 1)
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	fileRepo.EXPECT().StreamUploadFile(gomock.Any(), gomock.Any()).Return(&biz.File{ID: 1, Namespace: "a", Pod: "b", Container: "c"}, nil)
	k8sRepo.EXPECT().CopyFileToPod(gomock.Any(), gomock.Any()).Return(&biz.File{}, nil)
	err := svc.StreamCopyToPod(&streamCopyToPodServerRecvErr{})
	assert.Nil(t, err)
}

func TestContainerSvc_StreamCopyToPod_CtxCancelled(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	ctx, cancel := context.WithCancel(newAdminUserCtx())
	cancel()
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().FileAuditLog(types.EventActionType_Upload, "admin", gomock.Any(), 1)
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	fileRepo.EXPECT().StreamUploadFile(gomock.Any(), gomock.Any()).Return(&biz.File{ID: 1, Namespace: "a", Pod: "b", Container: "c"}, nil)
	k8sRepo.EXPECT().CopyFileToPod(gomock.Any(), gomock.Any()).Return(&biz.File{}, nil)
	var recvs []*container.StreamCopyToPodRequest
	for i := 0; i < 110; i++ {
		recvs = append(recvs, &container.StreamCopyToPodRequest{Namespace: "a", Pod: "b", Container: "c", FileName: "a.txt"})
	}
	err := svc.StreamCopyToPod(&streamCopyToPodServer{ctx: ctx, recv: recvs})
	assert.Nil(t, err)
}

type execServerAll struct {
	container.Container_ExecServer
	count int
	err   *container.ExecError
}

func (e *execServerAll) Recv() (*container.ExecRequest, error) {
	e.count++
	if e.count == 1 {
		return &container.ExecRequest{Namespace: "a", Pod: "b", Container: "c", Command: []string{"ls"}, Message: []byte("hello")}, nil
	}
	return nil, errors.New("recv done")
}

func (e *execServerAll) Send(response *container.ExecResponse) error {
	if response.Error != nil {
		e.err = response.Error
	}
	return errors.New("send boom")
}

func (e *execServerAll) Context() context.Context {
	return newAdminUserCtx()
}

func TestContainerSvc_Exec_SendError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().FileAuditLogWithDuration(types.EventActionType_Exec, "mars", gomock.Any(), 1, time.Second)
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	reco := &recorderMock{}
	fileRepo.EXPECT().NewRecorder(gomock.Any(), gomock.Any()).Return(reco)
	k8sRepo.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Cond(func(x any) bool {
		v, ok := x.(*biz.ExecuteInput)
		if !ok {
			return false
		}
		// 消费 stdin，让 recv goroutine 的 writer.Write 立即返回
		rbuf := make([]byte, 64)
		_, _ = v.Stdin.Read(rbuf)
		// 写入 stdout，让 send goroutine 读到数据并触发 Send 错误
		_, _ = v.Stdout.Write([]byte("A"))
		return true
	})).Return(&biz.ExecExitError{Code: 3, Message: "xx"})

	err := svc.Exec(&execServerAll{})
	assert.Error(t, err)
}

// recv goroutine 的首个 message 写入失败（reader 已被关闭）时，
// 只能记录日志，不能 panic / 中断整个 Exec。
func TestContainerSvc_Exec_FirstWriteError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().FileAuditLogWithDuration(types.EventActionType_Exec, "mars", gomock.Any(), 1, time.Second)
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	reco := &recorderMock{}
	fileRepo.EXPECT().NewRecorder(gomock.Any(), gomock.Any()).Return(reco)
	k8sRepo.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Return(&biz.ExecExitError{Code: 4, Message: "xx"})

	err := svc.Exec(&execServerAll{})
	assert.Error(t, err)
}

// 并发 Send 检测 mock：追踪任何时刻正在执行 Send 的 goroutine 数。
// grpc-go stream.go 明确 "not safe to call SendMsg on the same stream in
// different goroutines"，Exec/ExecOnce 的 send loop 与主流程 exit error Send
// 必须串行化，任何时刻最多 1 个 Send 在执行。
type concurrentSendExecServer struct {
	container.Container_ExecServer
	count     int
	recvBlock chan struct{}
	inSend    atomic.Int32
	max       atomic.Int32
	release   chan struct{}
	bothIn    chan struct{}
	signaled  atomic.Bool
}

func (e *concurrentSendExecServer) Recv() (*container.ExecRequest, error) {
	e.count++
	if e.count == 1 {
		return &container.ExecRequest{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
			Command:   []string{"ls"},
		}, nil
	}
	// 阻塞住 recv goroutine，避免它提前触发 closeAll 关闭 pipe。
	<-e.recvBlock
	return nil, nil
}

func (e *concurrentSendExecServer) Send(resp *container.ExecResponse) error {
	n := e.inSend.Add(1)
	if n > e.max.Load() {
		e.max.Store(n)
	}
	if n == 2 && e.signaled.CompareAndSwap(false, true) {
		close(e.bothIn)
	}
	<-e.release
	e.inSend.Add(-1)
	return nil
}

func (e *concurrentSendExecServer) Context() context.Context {
	return newAdminUserCtx()
}

func TestContainerSvc_Exec_ConcurrentSend(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().FileAuditLogWithDuration(types.EventActionType_Exec, "mars", gomock.Any(), 1, time.Second)
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	fileRepo.EXPECT().NewRecorder(gomock.Any(), gomock.Any()).Return(&recorderMock{})
	k8sRepo.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Cond(func(x any) bool {
		v, ok := x.(*biz.ExecuteInput)
		if !ok {
			return false
		}
		// 写入 stdout，让 send loop goroutine 读到数据并进入 Send。
		_, _ = v.Stdout.Write([]byte("x"))
		return true
	})).Return(&biz.ExecExitError{Code: 4, Message: "xx"})

	server := &concurrentSendExecServer{
		recvBlock: make(chan struct{}),
		release:   make(chan struct{}),
		bothIn:    make(chan struct{}),
	}
	done := make(chan error, 1)
	go func() { done <- svc.Exec(server) }()

	// 无锁时 send loop 与 exit error Send 会并发进入（bothIn 立即触发）；
	// 有锁时 exit error Send 被互斥锁挡住，只能等超时后放行 —— 两种情况都能确定性地结束。
	select {
	case <-server.bothIn:
	case <-time.After(300 * time.Millisecond):
	}
	close(server.release)

	assert.Error(t, <-done)
	// gRPC SendMsg 不保证并发安全：任何时刻最多 1 个 Send 在执行。
	assert.Equal(t, int32(1), server.max.Load())
}

type concurrentSendExecOnceServer struct {
	container.Container_ExecOnceServer
	inSend   atomic.Int32
	max      atomic.Int32
	release  chan struct{}
	bothIn   chan struct{}
	signaled atomic.Bool
}

func (e *concurrentSendExecOnceServer) Send(resp *container.ExecResponse) error {
	n := e.inSend.Add(1)
	if n > e.max.Load() {
		e.max.Store(n)
	}
	if n == 2 && e.signaled.CompareAndSwap(false, true) {
		close(e.bothIn)
	}
	<-e.release
	e.inSend.Add(-1)
	return nil
}

func (e *concurrentSendExecOnceServer) Context() context.Context {
	return newAdminUserCtx()
}

func TestContainerSvc_ExecOnce_ConcurrentSend(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().AuditLogWithChange(types.EventActionType_Exec, "admin", gomock.Any(), nil, gomock.Any())
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("c", nil)
	k8sRepo.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Cond(func(x any) bool {
		v, ok := x.(*biz.ExecuteInput)
		if !ok {
			return false
		}
		_, _ = v.Stdout.Write([]byte("x"))
		return true
	})).Return(&biz.ExecExitError{Code: 4, Message: "xx"})

	server := &concurrentSendExecOnceServer{
		release: make(chan struct{}),
		bothIn:  make(chan struct{}),
	}
	done := make(chan error, 1)
	go func() { done <- svc.ExecOnce(&container.ExecOnceRequest{Namespace: "a", Pod: "b"}, server) }()

	select {
	case <-server.bothIn:
	case <-time.After(300 * time.Millisecond):
	}
	close(server.release)

	assert.Error(t, <-done)
	assert.Equal(t, int32(1), server.max.Load())
}

type execOnceServerErr struct {
	container.Container_ExecOnceServer
	ctx context.Context
}

func (l *execOnceServerErr) Context() context.Context {
	if l.ctx != nil {
		return l.ctx
	}
	return newAdminUserCtx()
}

func (l *execOnceServerErr) Send(response *container.ExecResponse) error {
	return errors.New("send boom")
}

func TestContainerSvc_ExecOnce_SendError(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().AuditLogWithChange(types.EventActionType_Exec, "admin", gomock.Any(), nil, gomock.Any())
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("c", nil)
	k8sRepo.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Cond(func(x any) bool {
		v, ok := x.(*biz.ExecuteInput)
		if !ok {
			return false
		}
		// 写入 stdout 让 send goroutine 读到数据并触发 Send 错误
		_, _ = v.Stdout.Write([]byte("A"))
		return true
	})).Return(&biz.ExecExitError{Code: 1, Message: "xx"})

	err := svc.ExecOnce(&container.ExecOnceRequest{Namespace: "a", Pod: "b"}, &execOnceServerErr{})
	assert.Error(t, err)
}

func TestContainerSvc_Exec_PodNotRunning(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(false, "Pod not running")
	err := svc.Exec(&execServerMock{})
	// pod 未运行 → NotFound，与 CopyToPod/metrics 的语义对齐。
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestContainerSvc_Exec_Success(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	nsRepo.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(&biz.Namespace{}, nil).AnyTimes()
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo
	eventRepo.EXPECT().FileAuditLogWithDuration(
		types.EventActionType_Exec,
		"mars",
		gomock.Any(),
		1,
		time.Second,
	)
	k8sRepo.EXPECT().IsPodRunning("a", "b").Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("c", nil)
	reco := &recorderMock{}
	fileRepo.EXPECT().NewRecorder(gomock.Any(), &biz.Container{
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}).Return(reco)
	k8sRepo.EXPECT().Execute(gomock.Any(), &biz.Container{
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}, gomock.Cond(func(x any) bool {
		v, ok := x.(*biz.ExecuteInput)
		if !ok {
			return false
		}
		return v.TTY &&
			slices.Equal(v.Cmd, []string{"ls"}) &&
			v.Stdin != nil &&
			v.Stdout != nil &&
			v.Stderr != nil &&
			v.TerminalSizeQueue != nil
	})).Return(&biz.ExecExitError{
		Code:    2,
		Message: "xx",
	})
	mock := &execServerMock{}
	err := svc.Exec(mock)
	assert.Equal(t, int64(2), mock.err.Code)
	assert.NotNil(t, err)
	assert.Equal(t, uint16(10), reco.w)
	assert.Equal(t, uint16(20), reco.h)
}

type execServerRecvErr struct {
	container.Container_ExecServer
}

func (e *execServerRecvErr) Recv() (*container.ExecRequest, error) {
	return nil, errors.New("recv boom")
}

func (e *execServerRecvErr) Context() context.Context {
	return newAdminUserCtx()
}

type execServerMock struct {
	container.Container_ExecServer
	err *container.ExecError
}

func (e *execServerMock) Recv() (*container.ExecRequest, error) {
	return &container.ExecRequest{
		Namespace: "a",
		Pod:       "b",
		Command:   []string{"ls"},
		SizeQueue: &container.TerminalSize{
			Width:  10,
			Height: 20,
		},
	}, nil
}

func (e *execServerMock) Send(response *container.ExecResponse) error {
	if response.Error != nil {
		e.err = response.Error
	}
	return nil
}

func (e *execServerMock) Context() context.Context {
	return newOtherUserCtx()
}

type recorderMock struct {
	biz.Recorder
	w uint16
	h uint16
}

func (r *recorderMock) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (r *recorderMock) Resize(width, height uint16) {
	r.h = height
	r.w = width
}

func (r *recorderMock) Close() error {
	return nil
}

func (r *recorderMock) Container() *biz.Container {
	return &biz.Container{}
}

func (r *recorderMock) User() *biz.UserInfo {
	return &biz.UserInfo{Name: "mars"}
}

func (r *recorderMock) File() *biz.File {
	return &biz.File{ID: 1}
}

func (r *recorderMock) Duration() time.Duration {
	return time.Second
}

func TestScannerText_SingleLine(t *testing.T) {
	var result string
	err := scannerText("single line", func(s string) {
		result = s
	})
	assert.Nil(t, err)
	assert.Equal(t, "single line", result)
}

func TestScannerText_MultipleLines(t *testing.T) {
	var result []string
	err := scannerText("line1\nline2\nline3", func(s string) {
		result = append(result, s)
	})
	assert.Nil(t, err)
	assert.Equal(t, []string{"line1", "line2", "line3"}, result)
}

func TestScannerText_EmptyString(t *testing.T) {
	var result []string
	err := scannerText("", func(s string) {
		result = append(result, s)
	})
	assert.Nil(t, err)
	assert.Nil(t, result)
}

// 单行超过默认 64KB token 上限时必须仍能完整流式返回（容器日志常见超长行）。
func TestScannerText_LongLine(t *testing.T) {
	long := strings.Repeat("a", 100*1024) // 100KB > bufio.MaxScanTokenSize(64KB)
	var result string
	err := scannerText(long, func(s string) {
		result = s
	})
	assert.Nil(t, err)
	assert.Equal(t, long, result)
}

func Test_toValidUTF8String(t *testing.T) {
	// 测试有效的 UTF-8 字符串
	validUTF8 := []byte("hello, 世界")
	assert.Equal(t, "hello, 世界", toValidUTF8String(validUTF8))

	// 测试空字节
	assert.Equal(t, "", toValidUTF8String([]byte{}))

	// 测试包含无效 UTF-8 字符的字节
	// 0xff 是无效的 UTF-8 字节
	invalidUTF8 := []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0xff, 0xfe}
	result := toValidUTF8String(invalidUTF8)
	assert.Equal(t, "hello", result) // 无效字符被移除

	// 测试混合有效和无效 UTF-8
	mixed := []byte("valid\x80\x81string")
	result = toValidUTF8String(mixed)
	assert.Equal(t, "validstring", result)

	// 测试纯 ASCII
	ascii := []byte("pure ascii text")
	assert.Equal(t, "pure ascii text", toValidUTF8String(ascii))
}

func Test_containerSvc_ForceDeletePod_Success(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	mocks.nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{}, nil).AnyTimes()
	mocks.k8sRepo.EXPECT().DeletePod(gomock.Any(), "a", "b", gomock.Any()).Return(nil)
	mocks.eventRepo.EXPECT().AuditLog(types.EventActionType_ForceDeletePod, "admin", gomock.Any()).Times(1)
	resp, err := svc.ForceDeletePod(newAdminUserCtx(), &container.ForceDeletePodRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.Nil(t, err)
	assert.True(t, resp.Deleted)
	assert.Equal(t, "a", resp.Namespace)
	assert.Equal(t, "b", resp.Pod)
	assert.Contains(t, resp.Message, "已强制删除")
}

func Test_containerSvc_ForceDeletePod_Error(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	mocks.nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{}, nil).AnyTimes()
	mocks.k8sRepo.EXPECT().DeletePod(gomock.Any(), "a", "b", gomock.Any()).Return(errors.New("k8s delete failed"))
	mocks.eventRepo.EXPECT().AuditLog(types.EventActionType_ForceDeletePod, "admin", gomock.Any()).Times(1)
	_, err := svc.ForceDeletePod(newAdminUserCtx(), &container.ForceDeletePodRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "k8s delete failed")
}

func Test_containerSvc_ForceDeletePod_PermissionDenied(t *testing.T) {
	svc, mocks := newContainerSvcWithMocks(t)
	mocks.nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Private: true}, nil).AnyTimes()
	_, err := svc.ForceDeletePod(newOtherUserCtx(), &container.ForceDeletePodRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

// containerSvcMocks 聚合 containerSvc 的全部下游 mock，由 newContainerSvcWithMocks 统一构造。
type containerSvcMocks struct {
	ctrl      *gomock.Controller
	k8sRepo   *data.MockK8sRepo
	fileRepo  *data.MockFileRepo
	nsRepo    *data.MockNamespaceRepo
	eventRepo *data.MockEventRepo
}

func newContainerSvcWithMocks(t *testing.T) (*containerSvc, *containerSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &containerSvcMocks{
		ctrl:      ctrl,
		k8sRepo:   data.NewMockK8sRepo(ctrl),
		fileRepo:  data.NewMockFileRepo(ctrl),
		nsRepo:    data.NewMockNamespaceRepo(ctrl),
		eventRepo: data.NewMockEventRepo(ctrl),
	}
	s, ok := NewContainerSvc(ContainerSvcDeps{
		ContainerBiz: biz.NewContainerBiz(mlog.NewForConfig(nil), biz.NewK8sBiz(mocks.k8sRepo), biz.NewFileBiz(mocks.fileRepo), biz.NewEventBiz(mocks.eventRepo), timer.NewReal()),
		EventBiz:     biz.NewEventBiz(mocks.eventRepo),
		K8sBiz:       biz.NewK8sBiz(mocks.k8sRepo),
		FileBiz:      biz.NewFileBiz(mocks.fileRepo),
		AccessBiz:    biz.NewAccessBiz(biz.NewNamespaceBiz(mlog.NewForConfig(nil), mocks.nsRepo, nil, nil, nil), nil),
		Logger:       mlog.NewForConfig(nil),
	}).(*containerSvc)
	if !ok {
		panic("NewContainerSvc returned unexpected type")
	}
	return s, mocks
}
