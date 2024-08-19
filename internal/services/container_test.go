package services

import (
	"context"
	"errors"
	"io"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/auth"
	"k8s.io/client-go/tools/remotecommand"
	clientgoexec "k8s.io/client-go/util/exec"

	"github.com/duc-cnzj/mars/api/v4/container"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewContainerSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		repo.NewMockK8sRepo(m),
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*containerSvc).eventRepo)
	assert.NotNil(t, svc.(*containerSvc).k8sRepo)
	assert.NotNil(t, svc.(*containerSvc).fileRepo)
	assert.NotNil(t, svc.(*containerSvc).eventRepo)
	assert.NotNil(t, svc.(*containerSvc).logger)
}

func Test_containerSvc_IsPodRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().IsPodRunning(gomock.Any(), gomock.Any()).Return(false, "")
	running, err := svc.IsPodRunning(context.TODO(), &container.IsPodRunningRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.False(t, running.Running)
	assert.Nil(t, err)
}

func Test_containerSvc_IsPodExists(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(nil, nil)
	running, err := svc.IsPodExists(context.TODO(), &container.IsPodExistsRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.True(t, running.Exists)
	assert.Nil(t, err)
}

func Test_containerSvc_IsPodExists_Fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(nil, errors.New("X"))
	running, err := svc.IsPodExists(context.TODO(), &container.IsPodExistsRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.False(t, running.Exists)
	assert.Nil(t, err)
}

func TestContainerSvc_ContainerLog_PodNotFound(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(nil, nil)
	_, err := svc.ContainerLog(context.TODO(), &container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.NotNil(t, err)
}

func TestContainerSvc_ContainerLog_PodPending(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodPending}}, nil)
	_, err := svc.ContainerLog(context.TODO(), &container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.Error(t, err)
}

func TestContainerSvc_ContainerLog_PodRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning}}, nil)
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), gomock.Any(), gomock.Any()).Return("log", nil)
	_, err := svc.ContainerLog(context.TODO(), &container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.Nil(t, err)
}

func TestContainerSvc_ContainerLog_GetPodLogs_error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning}}, nil)
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("x"))
	_, err := svc.ContainerLog(context.TODO(), &container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.Equal(t, "x", err.Error())
}

func TestContainerSvc_ContainerLog_PodPending1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodPending}}, nil)
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
	resp, err := svc.ContainerLog(context.TODO(), &container.LogRequest{
		Namespace:  "a",
		Pod:        "b",
		ShowEvents: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, "aaa\nbbb", resp.Log)
}

type logStreamServer struct {
	ctx context.Context
	container.Container_StreamContainerLogServer
	res []string
}

func (l *logStreamServer) Send(response *container.LogResponse) error {
	l.res = append(l.res, response.Log)
	return nil
}

func (l *logStreamServer) Context() context.Context {
	if l.ctx != nil {
		return l.ctx
	}
	return context.TODO()
}

func TestContainerSvc_CopyToPod_PodNotRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		fileRepo,
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().IsPodRunning(gomock.Any(), gomock.Any()).Return(false, "")
	_, err := svc.CopyToPod(context.TODO(), &container.CopyToPodRequest{
		Namespace: "a",
		Pod:       "b",
	})
	assert.NotNil(t, err)
}

func TestContainerSvc_CopyToPod_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewContainerSvc(
		eventRepo,
		k8sRepo,
		fileRepo,
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().IsPodRunning(gomock.Any(), gomock.Any()).Return(true, "")
	eventRepo.EXPECT().FileAuditLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	k8sRepo.EXPECT().CopyFileToPod(gomock.Any(), &repo.CopyFileToPodRequest{
		FileId:    1,
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}).Return(&repo.File{}, nil)
	_, err := svc.CopyToPod(newAdminUserCtx(), &container.CopyToPodRequest{
		FileId:    1,
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	})
	assert.Nil(t, err)
}
func TestContainerSvc_CopyToPod_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewContainerSvc(
		eventRepo,
		k8sRepo,
		fileRepo,
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().IsPodRunning(gomock.Any(), gomock.Any()).Return(true, "")
	k8sRepo.EXPECT().CopyFileToPod(gomock.Any(), &repo.CopyFileToPodRequest{
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

func TestContainerSvc_StreamContainerLog_PodNotFound(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(nil, nil)
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, &logStreamServer{})
	assert.NotNil(t, err)
}

func TestContainerSvc_StreamContainerLog_PodPending(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodPending}}, nil)
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, &logStreamServer{})
	assert.Error(t, err)
}

func TestContainerSvc_StreamContainerLog_PodRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning}}, nil)
	ch := make(chan []byte, 2)
	ch <- []byte("log1")
	ch <- []byte("log2")
	close(ch)
	k8sRepo.EXPECT().LogStream(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(ch, nil)
	s := &logStreamServer{}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, s)
	assert.Nil(t, err)
	assert.Equal(t, []string{"log1", "log2"}, s.res)
}

func TestContainerSvc_StreamContainerLog_PodSucceeded(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodSucceeded}}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), gomock.Any(), gomock.Any()).Return("log", nil)
	s := &logStreamServer{}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, s)
	assert.Nil(t, err)
	assert.Equal(t, []string{"log"}, s.res)
}

func TestContainerSvc_StreamContainerLog_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodSucceeded}}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("x"))
	s := &logStreamServer{}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, s)
	assert.Equal(t, "x", err.Error())
}

func TestContainerSvc_StreamContainerLog_PodFailed(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodFailed}}, nil).AnyTimes()
	k8sRepo.EXPECT().GetPodLogs(gomock.Any(), gomock.Any(), gomock.Any()).Return("log", nil)
	s := &logStreamServer{}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, s)
	assert.Nil(t, err)
	assert.Equal(t, []string{"log"}, s.res)
}

func TestContainerSvc_StreamContainerLog_PodPending1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().GetPod(gomock.Any(), gomock.Any()).Return(&v1.Pod{Status: v1.PodStatus{Phase: v1.PodPending}}, nil).AnyTimes()
	s := &logStreamServer{}
	err := svc.StreamContainerLog(&container.LogRequest{
		Namespace: "a",
		Pod:       "b",
	}, s)

	assert.Equal(t, "未找到日志", status.Convert(err).Message())
}

type streamCopyToPodServer struct {
	container.Container_StreamCopyToPodServer
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
	return newAdminUserCtx()
}

func TestContainerSvc_StreamCopyToPod_PodNotRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		fileRepo,
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().IsPodRunning(gomock.Any(), gomock.Any()).Return(false, "")
	err := svc.StreamCopyToPod(&streamCopyToPodServer{recv: []*container.StreamCopyToPodRequest{
		{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
	}})
	assert.NotNil(t, err)
}

func TestContainerSvc_StreamCopyToPod_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewContainerSvc(
		eventRepo,
		k8sRepo,
		fileRepo,
		mlog.NewLogger(nil),
	)
	err := svc.StreamCopyToPod(&streamCopyToPodServer{err: errors.New("xx")})
	assert.Equal(t, "xx", err.Error())
}

func TestContainerSvc_StreamCopyToPod_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewContainerSvc(
		eventRepo,
		k8sRepo,
		fileRepo,
		mlog.NewLogger(nil),
	)
	eventRepo.EXPECT().FileAuditLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	k8sRepo.EXPECT().IsPodRunning(gomock.Any(), gomock.Any()).Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return("c", nil)
	fileRepo.EXPECT().StreamUploadFile(gomock.Any(), gomock.Any()).Return(&repo.File{
		ID:        1,
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}, nil)
	k8sRepo.EXPECT().CopyFileToPod(gomock.Any(), &repo.CopyFileToPodRequest{
		FileId:    1,
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}).Return(&repo.File{}, nil)
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

func TestSortEvents(t *testing.T) {
	event1 := &eventv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "1",
		},
	}
	event2 := &eventv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "2",
		},
	}
	event3 := &eventv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "3",
		},
	}

	events := sortEvents{event3, event1, event2}
	sort.Sort(events)

	assert.Equal(t, "1", events[0].ResourceVersion)
	assert.Equal(t, "2", events[1].ResourceVersion)
	assert.Equal(t, "3", events[2].ResourceVersion)
}

type execOnceServer struct {
	container.Container_ExecOnceServer
	res   []string
	Error *container.ExecError
}

func (l *execOnceServer) Context() context.Context {
	return newAdminUserCtx()
}

func (l *execOnceServer) Send(response *container.ExecResponse) error {
	l.res = append(l.res, string(response.Message))
	l.Error = response.Error
	return nil
}

func TestContainerSvc_ExecOnce_PodNotRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		repo.NewMockFileRepo(m),
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().IsPodRunning(gomock.Any(), gomock.Any()).Return(false, "")
	err := svc.ExecOnce(&container.ExecOnceRequest{
		Namespace: "a",
		Pod:       "b",
	}, &execOnceServer{})
	assert.NotNil(t, err)
}

func TestContainerSvc_ExecOnce_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewContainerSvc(
		eventRepo,
		k8sRepo,
		fileRepo,
		mlog.NewLogger(nil),
	)
	recorder := repo.NewMockRecorder(m)
	k8sRepo.EXPECT().IsPodRunning(gomock.Any(), gomock.Any()).Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return("c", nil)
	fileRepo.EXPECT().NewRecorder(gomock.Any(), gomock.Any(), gomock.Any()).Return(recorder)
	recorder.EXPECT().Write([]byte("mars@c:/# ls"))
	recorder.EXPECT().Write([]byte("\r\n"))
	recorder.EXPECT().Close()
	recorder.EXPECT().Container().Return(&repo.Container{}).AnyTimes()
	recorder.EXPECT().User().Return(&auth.UserInfo{Name: "mars"})
	recorder.EXPECT().File().Return(&repo.File{ID: 1})
	recorder.EXPECT().Duration().Return(time.Second)
	eventRepo.EXPECT().FileAuditLogWithDuration(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

	mac := &execOnceMatcher{
		tty: false,
		cmd: []string{"ls"},
	}
	k8sRepo.EXPECT().Execute(gomock.Any(), &repo.Container{
		Namespace: "a",
		Pod:       "b",
		Container: "c",
	}, mac).Return(clientgoexec.CodeExitError{
		Err:  errors.New("xx"),
		Code: 1,
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
	input *repo.ExecuteInput
	tty   bool
	cmd   []string
}

func (e *execOnceMatcher) Matches(x any) bool {
	input, ok := x.(*repo.ExecuteInput)
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

func TestContainerSvc_Exec_PodNotRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewContainerSvc(
		repo.NewMockEventRepo(m),
		k8sRepo,
		fileRepo,
		mlog.NewLogger(nil),
	)
	k8sRepo.EXPECT().IsPodRunning(gomock.Any(), gomock.Any()).Return(false, "Pod not running")
	err := svc.Exec(&execServerMock{})
	assert.NotNil(t, err)
}

func TestContainerSvc_Exec_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewContainerSvc(
		eventRepo,
		k8sRepo,
		fileRepo,
		mlog.NewLogger(nil),
	)
	eventRepo.EXPECT().FileAuditLogWithDuration(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	k8sRepo.EXPECT().IsPodRunning(gomock.Any(), gomock.Any()).Return(true, "")
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return("c", nil)
	fileRepo.EXPECT().NewRecorder(gomock.Any(), gomock.Any(), gomock.Any()).Return(&recorderMock{})
	k8sRepo.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	err := svc.Exec(&execServerMock{})
	assert.Nil(t, err)
}

type execServerMock struct {
	container.Container_ExecServer
}

func (e *execServerMock) Recv() (*container.ExecRequest, error) {
	return &container.ExecRequest{
		Namespace: "a",
		Pod:       "b",
		Command:   []string{"ls"},
	}, nil
}

func (e *execServerMock) Send(response *container.ExecResponse) error {
	return nil
}

func (e *execServerMock) Context() context.Context {
	return context.TODO()
}

type recorderMock struct {
	repo.Recorder
}

func (r *recorderMock) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (r *recorderMock) Close() error {
	return nil
}

func (r *recorderMock) Container() *repo.Container {
	return &repo.Container{}
}

func (r *recorderMock) User() *auth.UserInfo {
	return &auth.UserInfo{Name: "mars"}
}

func (r *recorderMock) File() *repo.File {
	return &repo.File{ID: 1}
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

func TestSizeQueue_Next_ContextDone(t *testing.T) {
	queue := &sizeQueue{
		ch:  make(chan *remotecommand.TerminalSize, 1),
		ctx: context.Background(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	queue.ctx = ctx
	cancel()

	assert.Nil(t, queue.Next())
}
func TestSizeQueue_Next_NotOk(t *testing.T) {
	queue := &sizeQueue{
		ch:  make(chan *remotecommand.TerminalSize, 1),
		ctx: context.Background(),
	}
	close(queue.ch)

	assert.Nil(t, queue.Next())
}

func TestSizeQueue_Next_SizeReceived(t *testing.T) {
	queue := &sizeQueue{
		ch:  make(chan *remotecommand.TerminalSize, 1),
		ctx: context.Background(),
	}

	expectedSize := &remotecommand.TerminalSize{Width: 10, Height: 20}
	queue.ch <- expectedSize

	assert.Equal(t, expectedSize, queue.Next())
}
