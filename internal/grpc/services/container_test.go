package services

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"sync"
	"testing"
	"time"

	eventsv1 "k8s.io/api/events/v1"

	"github.com/duc-cnzj/mars/v4/internal/utils/timer"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"k8s.io/client-go/kubernetes"

	"github.com/duc-cnzj/mars/api/v4/container"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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

	_, err := new(containerSvc).ContainerLog(context.TODO(), &container.LogRequest{
		Namespace: "naas",
		Pod:       "poaaa",
		Container: "app",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())

	log, err := new(containerSvc).ContainerLog(context.TODO(), &container.LogRequest{
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

func TestContainer_ContainerLogWhenPodPending(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "duc",
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
	}
	ev1 := &eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "ev1",
			Namespace:       "duc",
			ResourceVersion: "3",
		},
		Regarding: v1.ObjectReference{
			Kind: "Pod",
			Name: "pod1",
		},
		Note: "event note 1",
	}
	ev2 := &eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "ev2",
			Namespace:       "duc",
			ResourceVersion: "2",
		},
		Regarding: v1.ObjectReference{
			Kind: "Pod",
			Name: "pod1",
		},
		Note: "event note 2",
	}
	ev3 := &eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "ev3",
			Namespace:       "duc",
			ResourceVersion: "1",
		},
		Regarding: v1.ObjectReference{
			Kind: "Pod",
			Name: "pod1",
		},
		Note: "event note 3",
	}
	fk := fake.NewSimpleClientset(pod)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:      fk,
		PodLister:   testutil.NewPodLister(pod),
		EventLister: testutil.NewEventLister(ev1, ev2, ev3),
	})

	log, err := new(containerSvc).ContainerLog(context.TODO(), &container.LogRequest{
		Namespace:  "duc",
		Pod:        "pod1",
		Container:  "app",
		ShowEvents: false,
	})
	assert.Error(t, err)
	assert.Nil(t, log)

	log, err = new(containerSvc).ContainerLog(context.TODO(), &container.LogRequest{
		Namespace:  "duc",
		Pod:        "pod1",
		Container:  "app",
		ShowEvents: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, (&container.LogResponse{
		Namespace:     "duc",
		PodName:       "pod1",
		ContainerName: "app",
		Log:           "event note 3\nevent note 2\nevent note 1",
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
	res, err := (&containerSvc{
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

	_, err = (&containerSvc{}).CopyToPod(adminCtx(), &container.CopyToPodRequest{
		FileId:    int64(file.ID),
		Namespace: "xxx",
		Pod:       "xxx",
		Container: "xxx",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())
	_, err = (&containerSvc{}).CopyToPod(adminCtx(), &container.CopyToPodRequest{
		FileId:    1111111,
		Namespace: "dev",
		Pod:       "po1",
		Container: "app",
	})
	assert.Error(t, err)

	pfc.EXPECT().Copy(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("xxx"))
	_, err = (&containerSvc{
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

	err := (&containerSvc{}).Exec(&container.ExecRequest{
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
	re.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("xxx"))
	r := mock.NewMockRecorderInterface(m)
	r.EXPECT().Write("mars@app:/# sh -c ls").Times(1)
	r.EXPECT().Write("\r\n").Times(1)
	r.EXPECT().Close().Times(1)
	err = (&containerSvc{
		Executor: re,
		NewRecorderFunc: func(action types.EventActionType, user contracts.UserInfo, timer timer.Timer, container contracts.Container) contracts.RecorderInterface {
			assert.Equal(t, "duc", user.Name)
			return r
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
		ctx: auth.SetUser(context.TODO(), &contracts.UserInfo{Name: "duc"}),
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

func (f *fakeRemoteExecutor) Execute(ctx context.Context, clientSet kubernetes.Interface, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
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
	r := mock.NewMockRecorderInterface(m)
	r.EXPECT().Write("mars@app:/# sh -c ls").Times(1)
	r.EXPECT().Write("\r\n").Times(1)
	r.EXPECT().Write("aaa").Times(1)
	r.EXPECT().Close().Times(1)
	var mu sync.Mutex
	var result []*container.ExecResponse
	err := (&containerSvc{
		NewRecorderFunc: func(actionType types.EventActionType, info contracts.UserInfo, timer timer.Timer, c contracts.Container) contracts.RecorderInterface {
			assert.Equal(t, "duc", info.Name)
			return r
		},
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
		ctx: auth.SetUser(context.TODO(), &contracts.UserInfo{Name: "duc"}),
	})
	assert.Nil(t, err)
	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, result, 1)
	assert.Equal(t, (&container.ExecResponse{Message: "aaa"}).String(), result[0].String())
}

func TestContainer_Exec_SuccessWithDefaultContainer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "duc",
			Annotations: map[string]string{
				defaultContainerAnnotationName: "app-default",
			},
		},
		Spec: v1.PodSpec{Containers: []v1.Container{
			{Name: "app"},
			{Name: "app-default"},
		}},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	fk := fake.NewSimpleClientset(pod)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client:    fk,
		PodLister: testutil.NewPodLister(pod),
	})
	r := mock.NewMockRecorderInterface(m)
	r.EXPECT().Write("mars@app-default:/# sh -c ls").Times(1)
	r.EXPECT().Write("\r\n").Times(1)
	r.EXPECT().Write("aaa").Times(1)
	r.EXPECT().Close().Times(1)
	var mu sync.Mutex
	var result []*container.ExecResponse
	err := (&containerSvc{
		NewRecorderFunc: func(actionType types.EventActionType, info contracts.UserInfo, timer timer.Timer, c contracts.Container) contracts.RecorderInterface {
			assert.Equal(t, "app-default", c.Container)
			assert.Equal(t, "duc", info.Name)
			return r
		},
		Executor: &fakeRemoteExecutor{
			execute: func(clientSet kubernetes.Interface, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
				stdout.Write([]byte("aaa"))
				return nil
			},
		},
	}).Exec(&container.ExecRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "",
		Command:   []string{"sh", "-c", "ls"},
	}, &execServer{
		send: func(res *container.ExecResponse) error {
			mu.Lock()
			defer mu.Unlock()
			result = append(result, res)
			return nil
		},
		ctx: auth.SetUser(context.TODO(), &contracts.UserInfo{Name: "duc"}),
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
	r := mock.NewMockRecorderInterface(m)
	r.EXPECT().Write("mars@app:/# sh -c ls").Times(1)
	r.EXPECT().Write("\r\n").Times(1)
	r.EXPECT().Close().Times(1)
	var result []*container.ExecResponse
	err := (&containerSvc{
		NewRecorderFunc: func(actionType types.EventActionType, info contracts.UserInfo, timer timer.Timer, c contracts.Container) contracts.RecorderInterface {
			return r
		},
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
		ctx: auth.SetUser(context.TODO(), &contracts.UserInfo{Name: "duc"}),
	})
	assert.Nil(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, (&container.ExecResponse{
		Error: &container.ExecError{
			Code:    int64(100),
			Message: "aaa",
		},
	}).String(), result[0].String())

	r2 := mock.NewMockRecorderInterface(m)
	r2.EXPECT().Write("mars@app:/# sh -c ls").Times(1)
	r2.EXPECT().Write("\r\n").Times(1)
	r2.EXPECT().Write("aaa").Times(1)
	r2.EXPECT().Close().Times(1)
	err = (&containerSvc{
		NewRecorderFunc: func(actionType types.EventActionType, info contracts.UserInfo, timer timer.Timer, c contracts.Container) contracts.RecorderInterface {
			return r2
		},
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
		ctx: auth.SetUser(context.TODO(), &contracts.UserInfo{Name: "duc"}),
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
	r := mock.NewMockRecorderInterface(m)
	r.EXPECT().Write(gomock.Any()).AnyTimes()
	r.EXPECT().Close().Times(1)
	err := (&containerSvc{
		NewRecorderFunc: func(actionType types.EventActionType, info contracts.UserInfo, timer timer.Timer, c contracts.Container) contracts.RecorderInterface {
			return r
		},
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
		ctx: auth.SetUser(cancel, &contracts.UserInfo{Name: "duc"}),
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
	exist, _ := new(containerSvc).IsPodExists(context.TODO(), &container.IsPodExistsRequest{
		Namespace: "nsxx",
		Pod:       "podxxx",
	})
	assert.Equal(t, false, exist.Exists)
	exist, _ = new(containerSvc).IsPodExists(context.TODO(), &container.IsPodExistsRequest{
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
	running, _ := new(containerSvc).IsPodRunning(context.TODO(), &container.IsPodRunningRequest{
		Namespace: "nsxx",
		Pod:       "podxxx",
	})
	assert.Equal(t, false, running.Running)
	running, _ = new(containerSvc).IsPodRunning(context.TODO(), &container.IsPodRunningRequest{
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
	err := (&containerSvc{Steamer: &defaultStreamer{}}).StreamContainerLog(&container.LogRequest{
		Namespace: "naas",
		Pod:       "poaaa",
		Container: "app",
	}, nil)
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())

	err = (&containerSvc{Steamer: &errorStreamer{}}).StreamContainerLog(&container.LogRequest{
		Namespace: "duc",
		Pod:       "pod1",
		Container: "app",
	}, nil)
	assert.Equal(t, "xxx", err.Error())

	var mu sync.Mutex
	var result []*container.LogResponse
	err = (&containerSvc{Steamer: &defaultStreamer{}}).StreamContainerLog(&container.LogRequest{
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
	result = nil
	err = (&containerSvc{Steamer: &defaultStreamer{}}).StreamContainerLog(&container.LogRequest{
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
	assert.Equal(t, "fake logs", result[0].Log)
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
	err := (&containerSvc{Steamer: ms}).StreamContainerLog(&container.LogRequest{
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
	err := (&containerSvc{Steamer: ms}).StreamContainerLog(&container.LogRequest{
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
	err = (&containerSvc{Steamer: ms2}).StreamContainerLog(&container.LogRequest{
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
	err = (&containerSvc{Steamer: ms3}).StreamContainerLog(&container.LogRequest{
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
	err := (&containerSvc{
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
				defaultContainerAnnotationName: "app2",
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
				defaultContainerAnnotationName: "app3",
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
	m := gomock.NewController(t)
	defer m.Finish()
	r := mock.NewMockRecorderInterface(m)
	w := newExecWriter(r)
	assert.False(t, w.IsClosed())
	r.EXPECT().Close().Times(1)
	w.Close()
	w.Close()
	assert.True(t, w.IsClosed())
	_, ok := <-w.ch
	assert.False(t, ok)
}

func Test_execWriter_Write(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	r := mock.NewMockRecorderInterface(m)

	w := newExecWriter(r)
	r.EXPECT().Write("aaa").Times(1)
	_, err2 := w.Write([]byte("aaa"))
	assert.Nil(t, err2)
	data := <-w.ch
	assert.Equal(t, "aaa", data)
	r.EXPECT().Close().Times(1)
	w.Close()
	_, err := w.Write([]byte("bbb"))
	assert.Error(t, err)
}

func Test_newExecWriter(t *testing.T) {
	assert.NotNil(t, newExecWriter(nil))
}

func Test_scannerText(t *testing.T) {
	var tests = []struct {
		input string
		want  []string
	}{
		{
			input: "a",
			want:  []string{"a"},
		},
		{
			input: "a\nb\nc",
			want:  []string{"a", "b", "c"},
		},
		{
			input: "a\nb\n\nc",
			want:  []string{"a", "b", "", "c"},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			var res []string
			err := scannerText(tt.input, func(s string) {
				res = append(res, s)
			})
			assert.Nil(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}
