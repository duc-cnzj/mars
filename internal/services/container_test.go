package services

import (
	"context"
	"errors"
	"io"
	"sort"
	"testing"

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
	ctx context.Context
	container.Container_StreamCopyToPodServer
	res  []string
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

func Test_containerSvc_Exec(t *testing.T) {}

func Test_scannerText(t *testing.T) {}

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

func TestNewMultiWriterCloser(t *testing.T) {
	closer1 := &wcloser{}
	closer2 := &wcloser{}
	mwc := NewMultiWriterCloser(closer1, closer2)

	mwc.Write([]byte("hello"))

	err := mwc.Close()
	assert.Nil(t, err)

	assert.True(t, closer1.closed)
	assert.True(t, closer2.closed)

	assert.Equal(t, "hello", string(closer1.w))
	assert.Equal(t, "hello", string(closer2.w))
}

type wcloser struct {
	closed bool
	w      []byte
}

func (w *wcloser) Write(p []byte) (n int, err error) {
	w.w = p
	return 0, nil
}

func (w *wcloser) Close() error {
	w.closed = true
	return nil
}
