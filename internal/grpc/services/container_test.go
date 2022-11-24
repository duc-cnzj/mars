package services

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"sync"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/duc-cnzj/mars-client/v4/container"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	clientgoexec "k8s.io/client-go/util/exec"
)

func TestContainer_ContainerLog(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "duc",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	fk := fake.NewSimpleClientset(pod)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:    fk,
		PodLister: testutil.NewPodLister(pod),
	})

	_, err := new(Container).ContainerLog(context.TODO(), &container.LogRequest{
		Namespace: "naas",
		Pod:       "poaaa",
		Container: "app",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())

	log, err := new(Container).ContainerLog(context.TODO(), &container.LogRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
	})
	assert.Nil(t, err)
	assert.Equal(t, (&container.LogResponse{
		Namespace:     "duc",
		PodName:       "pod1",
		ContainerName: "app",
		Log:           "fake logs",
	}).String(), log.String())
}

func TestContainer_CopyToPod(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "po1",
			Namespace: "dev",
		},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: "Running",
		},
	}
	fk := fake.NewSimpleClientset(pod)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fk, PodLister: testutil.NewPodLister(pod)}).AnyTimes()
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.File{})
	file := &models.File{}
	db.Create(file)
	testutil.AssertAuditLogFired(m, app)
	pfc := mock.NewMockPodFileCopier(m)
	pfc.EXPECT().Copy(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&contracts.CopyFileToPodResult{
		TargetDir:     "/tmp",
		ContainerPath: "/tmp/aa.txt",
		FileName:      "aa.txt",
	}, nil)
	res, err := (&Container{
		PodFileCopier: pfc,
	}).CopyToPod(adminCtx(), &container.CopyToPodRequest{
		FileId:    int64(file.ID),
		Namespace: "dev",
		Pod:       "po1",
		Container: "app",
	})
	assert.Nil(t, err)
	assert.Equal(t, "/tmp", res.PodFilePath)
	assert.Equal(t, "aa.txt", res.FileName)

	_, err = (&Container{}).CopyToPod(adminCtx(), &container.CopyToPodRequest{
		FileId:    int64(file.ID),
		Namespace: "xxx",
		Pod:       "xxx",
		Container: "xxx",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())
	_, err = (&Container{}).CopyToPod(adminCtx(), &container.CopyToPodRequest{
		FileId:    1111111,
		Namespace: "dev",
		Pod:       "po1",
		Container: "app",
	})
	assert.Error(t, err)

	pfc.EXPECT().Copy(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("xxx"))
	_, err = (&Container{
		PodFileCopier: pfc,
	}).CopyToPod(adminCtx(), &container.CopyToPodRequest{
		FileId:    int64(file.ID),
		Namespace: "dev",
		Pod:       "po1",
		Container: "app",
	})
	fromError, _ = status.FromError(err)

	assert.Equal(t, "xxx", fromError.Message())
	assert.Equal(t, codes.Internal, fromError.Code())
}

type execServer struct {
	send func(res *container.ExecResponse) error
	ctx  context.Context
	container.Container_ExecServer
}

func (e *execServer) Context() context.Context {
	return e.ctx
}

func (e *execServer) Send(res *container.ExecResponse) error {
	return e.send(res)
}

func TestContainer_Exec(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "duc",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}

	fk := fake.NewSimpleClientset(pod)

	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:    fk,
		PodLister: testutil.NewPodLister(pod),
	})

	err := (&Container{}).Exec(&container.ExecRequest{
		Namespace: "ducxx",
		Pod:       "not_exists",
		Container: "app",
		Command:   []string{"sh", "-c", "ls"},
	}, nil)
	assert.Equal(t, "pod \"not_exists\" not found", err.Error())

	var mu sync.Mutex
	var result []*container.ExecResponse
	re := mock.NewMockRemoteExecutor(m)
	re.EXPECT().WithContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return(re)
	re.EXPECT().WithCommand(gomock.Any()).Return(re)
	re.EXPECT().WithMethod(gomock.Any()).Return(re)
	re.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("xxx"))
	err = (&Container{
		Executor: re,
	}).Exec(&container.ExecRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Command:   []string{"sh", "-c", "ls"},
	}, &execServer{
		send: func(res *container.ExecResponse) error {
			mu.Lock()
			defer mu.Unlock()
			result = append(result, res)
			return nil
		},
		ctx: context.TODO(),
	})
	assert.Nil(t, err)
	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, result, 1)
	assert.Equal(t, (&container.ExecResponse{
		Error: &container.ExecError{
			Code:    int64(1),
			Message: "xxx",
		},
	}).String(), result[0].String())
}

type fakeRemoteExecutor struct {
	execute func(clientSet kubernetes.Interface, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error
}

func (f *fakeRemoteExecutor) WithMethod(method string) contracts.RemoteExecutor {
	return f
}

func (f *fakeRemoteExecutor) WithContainer(namespace, pod, container string) contracts.RemoteExecutor {
	return f
}

func (f *fakeRemoteExecutor) WithCommand(cmd []string) contracts.RemoteExecutor {
	return f
}

func (f *fakeRemoteExecutor) Execute(clientSet kubernetes.Interface, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
	return f.execute(clientSet, config, stdin, stdout, stderr, tty, terminalSizeQueue)
}

func TestContainer_Exec_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "duc",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	fk := fake.NewSimpleClientset(pod)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:    fk,
		PodLister: testutil.NewPodLister(pod),
	})

	var mu sync.Mutex
	var result []*container.ExecResponse
	err := (&Container{
		Executor: &fakeRemoteExecutor{
			execute: func(clientSet kubernetes.Interface, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
				stdout.Write([]byte("aaa"))
				return nil
			},
		},
	}).Exec(&container.ExecRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
		Command:   []string{"sh", "-c", "ls"},
	}, &execServer{
		send: func(res *container.ExecResponse) error {
			mu.Lock()
			defer mu.Unlock()
			result = append(result, res)
			return nil
		},
		ctx: context.TODO(),
	})
	assert.Nil(t, err)
	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, result, 1)
	assert.Equal(t, (&container.ExecResponse{Message: "aaa"}).String(), result[0].String())
}

func TestContainer_Exec_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "duc",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	fk := fake.NewSimpleClientset(pod)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:    fk,
		PodLister: testutil.NewPodLister(pod),
	})

	var result []*container.ExecResponse
	err := (&Container{
		Executor: &fakeRemoteExecutor{execute: func(clientSet kubernetes.Interface, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
			return &clientgoexec.CodeExitError{
				Err:  errors.New("aaa"),
				Code: 100,
			}
		}},
	}).Exec(&container.ExecRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
		Command:   []string{"sh", "-c", "ls"},
	}, &execServer{
		send: func(res *container.ExecResponse) error {
			result = append(result, res)
			return nil
		},
		ctx: context.TODO(),
	})
	assert.Nil(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, (&container.ExecResponse{
		Error: &container.ExecError{
			Code:    int64(100),
			Message: "aaa",
		},
	}).String(), result[0].String())

	err = (&Container{
		Executor: &fakeRemoteExecutor{execute: func(clientSet kubernetes.Interface, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
			stdout.Write([]byte("aaa"))
			return nil
		}},
	}).Exec(&container.ExecRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
		Command:   []string{"sh", "-c", "ls"},
	}, &execServer{
		send: func(res *container.ExecResponse) error {
			return errors.New("xxx")
		},
		ctx: context.TODO(),
	})
	assert.Equal(t, "xxx", err.Error())
}

func TestContainer_Exec_ErrorWithClientCtxDone(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "duc",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	fk := fake.NewSimpleClientset(pod)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:    fk,
		PodLister: testutil.NewPodLister(pod),
	})

	var (
		mu     sync.Mutex
		result []*container.ExecResponse
	)
	cancel, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	err := (&Container{
		Executor: &fakeRemoteExecutor{
			execute: func(clientSet kubernetes.Interface, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
				time.Sleep(1 * time.Second)
				return nil
			},
		},
	}).Exec(&container.ExecRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
		Command:   []string{"sh", "-c", "ls"},
	}, &execServer{
		send: func(res *container.ExecResponse) error {
			mu.Lock()
			defer mu.Unlock()
			result = append(result, res)
			return nil
		},
		ctx: cancel,
	})
	assert.Equal(t, "context canceled", err.Error())
}

func TestContainer_IsPodExists(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "ns",
		},
	}
	app.EXPECT().K8sClient().Times(2).Return(&contracts.K8sClient{
		Client:    fake.NewSimpleClientset(pod1),
		PodLister: testutil.NewPodLister(pod1),
	})
	exist, _ := new(Container).IsPodExists(context.TODO(), &container.IsPodExistsRequest{
		Namespace: "nsxx",
		Pod:       "podxxx",
	})
	assert.Equal(t, false, exist.Exists)
	exist, _ = new(Container).IsPodExists(context.TODO(), &container.IsPodExistsRequest{
		Namespace: "ns",
		Pod:       "pod",
	})
	assert.Equal(t, true, exist.Exists)
}

func TestContainer_IsPodRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "ns",
		},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: "Running",
		},
	}
	app.EXPECT().K8sClient().Times(2).Return(&contracts.K8sClient{Client: fake.NewSimpleClientset(pod), PodLister: testutil.NewPodLister(pod)})
	running, _ := new(Container).IsPodRunning(context.TODO(), &container.IsPodRunningRequest{
		Namespace: "nsxx",
		Pod:       "podxxx",
	})
	assert.Equal(t, false, running.Running)
	running, _ = new(Container).IsPodRunning(context.TODO(), &container.IsPodRunningRequest{
		Namespace: "ns",
		Pod:       "pod",
	})
	assert.Equal(t, true, running.Running)
}

type streamLogServer struct {
	send func(res *container.LogResponse) error
	ctx  context.Context
	container.Container_StreamContainerLogServer
}

func (e *streamLogServer) Context() context.Context {
	return e.ctx
}

func (e *streamLogServer) Send(res *container.LogResponse) error {
	return e.send(res)
}

type errorStreamer struct{}

func (e *errorStreamer) Stream(ctx context.Context, namespace, pod, container string) (io.ReadCloser, error) {
	return nil, errors.New("xxx")
}

type streamReadClose struct {
	times  int
	closed bool
}

func (s *streamReadClose) Read(p []byte) (n int, err error) {
	time.Sleep(100 * time.Millisecond)
	if s.times >= 1 {
		return 0, errors.New("xxx")
	}

	s.times++
	bytes := []byte(`aaa
bbb
ccc
`)
	copy(p, bytes)
	return len(bytes), nil
}

func (s *streamReadClose) Close() error {
	s.closed = true
	return nil
}

type myStreamer struct {
	s io.ReadCloser
}

func (e *myStreamer) Stream(ctx context.Context, namespace, pod, container string) (io.ReadCloser, error) {
	e.s = &streamReadClose{}
	return e.s, nil
}

func TestContainer_StreamContainerLog(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Done().Return(nil).Times(1)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "duc",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	podSuccessed := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-succeeded",
			Namespace: "duc",
		},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	}
	fk := fake.NewSimpleClientset(pod, podSuccessed)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:    fk,
		PodLister: testutil.NewPodLister(pod, podSuccessed),
	})

	err := (&Container{Steamer: &DefaultStreamer{}}).StreamContainerLog(&container.LogRequest{
		Namespace: "naas",
		Pod:       "poaaa",
		Container: "app",
	}, nil)
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())

	err = (&Container{Steamer: &errorStreamer{}}).StreamContainerLog(&container.LogRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
	}, nil)
	assert.Equal(t, "xxx", err.Error())

	var mu sync.Mutex
	var result []*container.LogResponse
	err = (&Container{Steamer: &DefaultStreamer{}}).StreamContainerLog(&container.LogRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
	}, &streamLogServer{
		send: func(res *container.LogResponse) error {
			mu.Lock()
			defer mu.Unlock()
			result = append(result, res)
			return nil
		},
		ctx: context.TODO(),
	})
	assert.Nil(t, err)

	err = (&Container{Steamer: &DefaultStreamer{}}).StreamContainerLog(&container.LogRequest{
		Namespace: "duc",
		Pod:       "pod-succeeded",
		Container: "app",
	}, &streamLogServer{
		send: func(res *container.LogResponse) error {
			mu.Lock()
			defer mu.Unlock()
			result = append(result, res)
			return nil
		},
		ctx: context.TODO(),
	})
	assert.Nil(t, err)
}

func TestContainer_StreamContainerLog_AppDone(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "duc",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	fk := fake.NewSimpleClientset(pod)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:    fk,
		PodLister: testutil.NewPodLister(pod),
	})

	done := make(chan struct{})
	close(done)
	app.EXPECT().Done().Return(done).AnyTimes()

	ms := &myStreamer{}
	err := (&Container{Steamer: ms}).StreamContainerLog(&container.LogRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
	}, &streamLogServer{
		send: func(res *container.LogResponse) error {
			return nil
		},
		ctx: context.TODO(),
	})
	assert.Equal(t, "server shutdown", err.Error())
	assert.True(t, ms.s.(*streamReadClose).closed)
}

func TestContainer_StreamContainerLog_ServerSend(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "duc",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	fk := fake.NewSimpleClientset(pod)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:    fk,
		PodLister: testutil.NewPodLister(pod),
	})

	app.EXPECT().Done().Return(nil).AnyTimes()

	var mu sync.Mutex
	var result []*container.LogResponse
	ms := &myStreamer{}
	err := (&Container{Steamer: ms}).StreamContainerLog(&container.LogRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
	}, &streamLogServer{
		send: func(res *container.LogResponse) error {
			mu.Lock()
			defer mu.Unlock()
			result = append(result, res)
			return nil
		},
		ctx: context.TODO(),
	})
	assert.Nil(t, err)
	assert.Len(t, result, 3)
	assert.True(t, ms.s.(*streamReadClose).closed)

	ms2 := &myStreamer{}
	err = (&Container{Steamer: ms2}).StreamContainerLog(&container.LogRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
	}, &streamLogServer{
		send: func(res *container.LogResponse) error {
			return errors.New("xxx")
		},
		ctx: context.TODO(),
	})
	assert.Equal(t, "xxx", err.Error())
	assert.True(t, ms2.s.(*streamReadClose).closed)

	ms3 := &myStreamer{}
	c, cn := context.WithCancel(context.TODO())
	cn()
	err = (&Container{Steamer: ms3}).StreamContainerLog(&container.LogRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
	}, &streamLogServer{
		send: func(res *container.LogResponse) error {
			return nil
		},
		ctx: c,
	})
	assert.Nil(t, err)
	assert.True(t, ms3.s.(*streamReadClose).closed)
}

type streamCopyToPodServer struct {
	ctx context.Context
	res *container.StreamCopyToPodResponse
	container.Container_StreamCopyToPodServer
	sync.Mutex
	write      int
	totalWrite int
}

func (s *streamCopyToPodServer) SendAndClose(res *container.StreamCopyToPodResponse) error {
	s.res = res
	return nil
}

func (s *streamCopyToPodServer) Recv() (*container.StreamCopyToPodRequest, error) {
	s.Lock()
	defer s.Unlock()
	defer func() {
		s.write++
	}()
	if s.write >= s.totalWrite {
		return nil, io.EOF
	}
	return &container.StreamCopyToPodRequest{
		FileName:  "a.txt",
		Data:      []byte("aa"),
		Namespace: "dev",
		Pod:       "po1",
	}, nil
}

func (s *streamCopyToPodServer) Context() context.Context {
	return s.ctx
}

type mockFileInfo struct {
	size int64
}

func (m *mockFileInfo) Name() string {
	return ""
}

func (m *mockFileInfo) Size() int64 {
	return m.size
}

func (m *mockFileInfo) Mode() fs.FileMode {
	return fs.FileMode(0644)
}

func (m *mockFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (m *mockFileInfo) IsDir() bool {
	return false
}

func (m *mockFileInfo) Sys() any {
	return nil
}

func TestContainer_StreamCopyToPod(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "po1",
			Namespace: "dev",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "app"},
			},
		},
		Status: v1.PodStatus{
			Phase: "Running",
		},
	}
	fk := fake.NewSimpleClientset(pod)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fk, PodLister: testutil.NewPodLister(pod)}).AnyTimes()

	up := mock.NewMockUploader(m)
	up.EXPECT().Type().Return(contracts.Local).AnyTimes()
	app.EXPECT().Uploader().Return(up).AnyTimes()
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.File{})
	file := &models.File{}
	db.Create(file)
	testutil.AssertAuditLogFired(m, app)

	up.EXPECT().Disk(gomock.Any()).Return(up)
	up.EXPECT().AbsolutePath(gomock.Any()).Return("/tmp/aa.txt")
	up.EXPECT().MkDir(gomock.Any(), gomock.Any()).Times(1).Return(nil)

	ff := mock.NewMockFile(m)
	ff.EXPECT().Name().Return("name")
	ff.EXPECT().Stat().Return(&mockFileInfo{}, nil)
	ff.EXPECT().Close().MinTimes(1)
	ff.EXPECT().Write([]byte("aa")).Times(2)

	up.EXPECT().NewFile(gomock.Any()).Return(ff, nil)
	up.EXPECT().Delete(gomock.Any()).Times(0)

	s := &streamCopyToPodServer{ctx: adminCtx(), totalWrite: 2}
	pfc := mock.NewMockPodFileCopier(m)
	pfc.EXPECT().Copy(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&contracts.CopyFileToPodResult{
		TargetDir:     "/tmp",
		ContainerPath: "/tmp/aa.txt",
		FileName:      "aa.txt",
	}, nil)
	err := (&Container{
		PodFileCopier: pfc,
	}).StreamCopyToPod(s)
	assert.Nil(t, err)
	assert.Equal(t, (&container.StreamCopyToPodResponse{
		PodFilePath: "/tmp",
		Pod:         "po1",
		Namespace:   "dev",
		Container:   "app",
		Filename:    "aa.txt",
	}).String(), s.res.String())
}

func TestFindDefaultContainer(t *testing.T) {
	defaultContainer := FindDefaultContainer(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "xx",
			Namespace:   "ns",
			Annotations: map[string]string{},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "app",
				},
				{
					Name: "app2",
				},
			},
		},
	})
	assert.Equal(t, "app", defaultContainer)
	defaultContainer = FindDefaultContainer(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "xx",
			Namespace: "ns",
			Annotations: map[string]string{
				DefaultContainerAnnotationName: "app2",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "app",
				},
				{
					Name: "app2",
				},
			},
		},
	})
	assert.Equal(t, "app2", defaultContainer)
	defaultContainer = FindDefaultContainer(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "xx",
			Namespace: "ns",
			Annotations: map[string]string{
				DefaultContainerAnnotationName: "app3",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "app",
				},
				{
					Name: "app2",
				},
			},
		},
	})
	assert.Equal(t, "app", defaultContainer)
	defaultContainer = FindDefaultContainer(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "xx",
			Namespace:   "ns",
			Annotations: map[string]string{},
		},
		Spec: v1.PodSpec{},
	})
	assert.Equal(t, "", defaultContainer)
}

func Test_execWriter_IsClosed(t *testing.T) {
	w := newExecWriter()
	assert.False(t, w.IsClosed())
	w.Close()
	w.Close()
	assert.True(t, w.IsClosed())
	_, ok := <-w.ch
	assert.False(t, ok)
}

func Test_execWriter_Write(t *testing.T) {
	w := newExecWriter()
	_, err2 := w.Write([]byte("aaa"))
	assert.Nil(t, err2)
	data := <-w.ch
	assert.Equal(t, "aaa", data)
	w.Close()
	_, err := w.Write([]byte("bbb"))
	assert.Error(t, err)
}

func Test_newExecWriter(t *testing.T) {
	assert.NotNil(t, newExecWriter())
}
