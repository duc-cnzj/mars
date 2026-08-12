package biz

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fakeK8sBizForLog 只覆写 Log/LogStream 用到的四个读取原语，其余方法 panic 兜底。
type fakeK8sBizForLog struct {
	K8sBiz
	getPod     func(namespace, podName string) (*v1.Pod, error)
	listEvents func(namespace string) ([]*eventv1.Event, error)
	getPodLogs func(ctx context.Context, namespace, podName string, options *v1.PodLogOptions) (string, error)
	logStream  func(ctx context.Context, namespace, podName, container string) (chan []byte, error)
}

func (f *fakeK8sBizForLog) GetPod(namespace, podName string) (*v1.Pod, error) {
	return f.getPod(namespace, podName)
}

func (f *fakeK8sBizForLog) ListEvents(namespace string) ([]*eventv1.Event, error) {
	return f.listEvents(namespace)
}

func (f *fakeK8sBizForLog) GetPodLogs(ctx context.Context, namespace, podName string, options *v1.PodLogOptions) (string, error) {
	return f.getPodLogs(ctx, namespace, podName, options)
}

func (f *fakeK8sBizForLog) LogStream(ctx context.Context, namespace, podName, container string) (chan []byte, error) {
	return f.logStream(ctx, namespace, podName, container)
}

func podOfPhase(phase v1.PodPhase) *v1.Pod {
	return &v1.Pod{Status: v1.PodStatus{Phase: phase}}
}

func TestContainerBiz_Log_PodNotFound(t *testing.T) {
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return nil, nil },
	}, nil, nil)
	_, err := cb.Log(context.TODO(), &LogInput{Namespace: "a", Pod: "b"})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestContainerBiz_Log_GetPodError_Propagate(t *testing.T) {
	// GetPod 返回错误时原样上抛（错误由最上层 services 统一打印），不吞错继续走后续逻辑。
	var hit bool
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) {
			return nil, errors.New("get pod boom")
		},
		getPodLogs: func(ctx context.Context, ns, pod string, opts *v1.PodLogOptions) (string, error) {
			hit = true
			return "log", nil
		},
	}, nil, nil)
	res, err := cb.Log(context.TODO(), &LogInput{Namespace: "a", Pod: "b", Container: "c"})
	assert.ErrorContains(t, err, "get pod boom")
	assert.Nil(t, res)
	assert.False(t, hit)
}

func TestContainerBiz_Log_Pending_NoShowEvents_NotFound(t *testing.T) {
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return podOfPhase(v1.PodPending), nil },
	}, nil, nil)
	_, err := cb.Log(context.TODO(), &LogInput{Namespace: "a", Pod: "b"})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestContainerBiz_Log_Pending_ShowEvents_AggregateAndSort(t *testing.T) {
	// 只聚合「本 pod 且 Kind==Pod」的事件，且按 ResourceVersion 数值升序（"10" 在 "9" 之后）。
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return podOfPhase(v1.PodPending), nil },
		listEvents: func(namespace string) ([]*eventv1.Event, error) {
			return []*eventv1.Event{
				{Regarding: v1.ObjectReference{Kind: "Pod", Name: "b"}, Note: "e9", ObjectMeta: metav1.ObjectMeta{ResourceVersion: "9"}},
				{Regarding: v1.ObjectReference{Kind: "Pod", Name: "b"}, Note: "e10", ObjectMeta: metav1.ObjectMeta{ResourceVersion: "10"}},
				{Regarding: v1.ObjectReference{Kind: "Pod", Name: "other"}, Note: "wrong-pod"},
				{Regarding: v1.ObjectReference{Kind: "Deployment", Name: "b"}, Note: "wrong-kind"},
			}, nil
		},
	}, nil, nil)
	res, err := cb.Log(context.TODO(), &LogInput{Namespace: "a", Pod: "b", ShowEvents: true})
	assert.NoError(t, err)
	assert.Equal(t, "e9\ne10", res.Content)
}

func TestContainerBiz_Log_Pending_ListEventsError_EmptyContent(t *testing.T) {
	// ListEvents 失败只打 Debug 日志，事件区置空不报错。
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return podOfPhase(v1.PodPending), nil },
		listEvents: func(namespace string) ([]*eventv1.Event, error) {
			return nil, errors.New("list events boom")
		},
	}, nil, nil)
	res, err := cb.Log(context.TODO(), &LogInput{Namespace: "a", Pod: "b", ShowEvents: true})
	assert.NoError(t, err)
	assert.Equal(t, "", res.Content)
}

func TestContainerBiz_Log_Pending_ShowEvents_NonNumericVersionFallback(t *testing.T) {
	// ResourceVersion 非数值时回退字符串比较（保持调用方顺序的稳定性），不 panic。
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return podOfPhase(v1.PodPending), nil },
		listEvents: func(namespace string) ([]*eventv1.Event, error) {
			return []*eventv1.Event{
				{Regarding: v1.ObjectReference{Kind: "Pod", Name: "b"}, Note: "first", ObjectMeta: metav1.ObjectMeta{ResourceVersion: "abc"}},
				{Regarding: v1.ObjectReference{Kind: "Pod", Name: "b"}, Note: "second", ObjectMeta: metav1.ObjectMeta{ResourceVersion: "abc"}},
			}, nil
		},
	}, nil, nil)
	res, err := cb.Log(context.TODO(), &LogInput{Namespace: "a", Pod: "b", ShowEvents: true})
	assert.NoError(t, err)
	assert.Equal(t, "first\nsecond", res.Content)
}

func TestContainerBiz_Log_GetPodLogsError(t *testing.T) {
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return podOfPhase(v1.PodSucceeded), nil },
		getPodLogs: func(ctx context.Context, ns, pod string, opts *v1.PodLogOptions) (string, error) {
			return "", errors.New("x")
		},
	}, nil, nil)
	_, err := cb.Log(context.TODO(), &LogInput{Namespace: "a", Pod: "b"})
	assert.Equal(t, "x", err.Error())
}

func TestContainerBiz_LogStream_Running_Live(t *testing.T) {
	ch := make(chan []byte, 1)
	ch <- []byte("live")
	close(ch)
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return podOfPhase(v1.PodRunning), nil },
		logStream: func(ctx context.Context, ns, pod, container string) (chan []byte, error) {
			assert.Equal(t, "c", container)
			return ch, nil
		},
	}, nil, nil)
	res, err := cb.LogStream(context.TODO(), &LogInput{Namespace: "a", Pod: "b", Container: "c"})
	assert.NoError(t, err)
	assert.Equal(t, LogSourceLive, res.Source)
	assert.Equal(t, []byte("live"), <-res.Stream)
}

func TestContainerBiz_LogStream_Running_LogStreamError(t *testing.T) {
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return podOfPhase(v1.PodRunning), nil },
		logStream: func(ctx context.Context, ns, pod, container string) (chan []byte, error) {
			return nil, errors.New("stream boom")
		},
	}, nil, nil)
	_, err := cb.LogStream(context.TODO(), &LogInput{Namespace: "a", Pod: "b"})
	assert.Equal(t, "stream boom", err.Error())
}

func TestContainerBiz_LogStream_Succeeded_Tail(t *testing.T) {
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return podOfPhase(v1.PodSucceeded), nil },
		getPodLogs: func(ctx context.Context, ns, pod string, opts *v1.PodLogOptions) (string, error) {
			return "log", nil
		},
	}, nil, nil)
	res, err := cb.LogStream(context.TODO(), &LogInput{Namespace: "a", Pod: "b"})
	assert.NoError(t, err)
	assert.Equal(t, LogSourceContent, res.Source)
	assert.Equal(t, "log", res.Content)
}

func TestContainerBiz_LogStream_Pending_ShowEvents(t *testing.T) {
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return podOfPhase(v1.PodPending), nil },
		listEvents: func(namespace string) ([]*eventv1.Event, error) {
			return []*eventv1.Event{
				{Regarding: v1.ObjectReference{Kind: "Pod", Name: "b"}, Note: "ev"},
			}, nil
		},
	}, nil, nil)
	res, err := cb.LogStream(context.TODO(), &LogInput{Namespace: "a", Pod: "b", ShowEvents: true})
	assert.NoError(t, err)
	assert.Equal(t, LogSourceContent, res.Source)
	assert.Equal(t, "ev", res.Content)
}

func TestContainerBiz_LogStream_LogError(t *testing.T) {
	// 非 Running pod 且 Log 内部出错（如 GetPodLogs 失败）时，错误原样透传。
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return podOfPhase(v1.PodSucceeded), nil },
		getPodLogs: func(ctx context.Context, ns, pod string, opts *v1.PodLogOptions) (string, error) {
			return "", errors.New("tail boom")
		},
	}, nil, nil)
	_, err := cb.LogStream(context.TODO(), &LogInput{Namespace: "a", Pod: "b"})
	assert.Equal(t, "tail boom", err.Error())
}

func TestContainerBiz_LogStream_PodNotFound(t *testing.T) {
	cb := newTestContainerBiz(&fakeK8sBizForLog{
		getPod: func(namespace, pod string) (*v1.Pod, error) { return nil, nil },
	}, nil, nil)
	_, err := cb.LogStream(context.TODO(), &LogInput{Namespace: "a", Pod: "b"})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// ---- sortEvents（随日志规则从 transport 迁入 biz）----

func TestSortEvents(t *testing.T) {
	event1 := &eventv1.Event{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"}}
	event2 := &eventv1.Event{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "2"}}
	event3 := &eventv1.Event{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "3"}}

	events := sortEvents{event3, event1, event2}
	sort.Sort(events)

	assert.Equal(t, "1", events[0].ResourceVersion)
	assert.Equal(t, "2", events[1].ResourceVersion)
	assert.Equal(t, "3", events[2].ResourceVersion)
}

// ResourceVersion 是数字字符串，必须按数值比较，否则 "10" 会排在 "9" 前面。
func TestSortEvents_NumericOrder(t *testing.T) {
	mk := func(rv string) *eventv1.Event {
		return &eventv1.Event{ObjectMeta: metav1.ObjectMeta{ResourceVersion: rv}}
	}
	events := sortEvents{mk("9"), mk("10"), mk("2")}
	sort.Sort(events)
	assert.Equal(t, "2", events[0].ResourceVersion)
	assert.Equal(t, "9", events[1].ResourceVersion)
	assert.Equal(t, "10", events[2].ResourceVersion)
}

// 非数字 ResourceVersion 时回退到字符串比较，不 panic。
func TestSortEvents_NonNumericFallback(t *testing.T) {
	mk := func(rv string) *eventv1.Event {
		return &eventv1.Event{ObjectMeta: metav1.ObjectMeta{ResourceVersion: rv}}
	}
	events := sortEvents{mk("b"), mk("a")}
	sort.Sort(events)
	assert.Equal(t, "a", events[0].ResourceVersion)
	assert.Equal(t, "b", events[1].ResourceVersion)
}

// 超过 int64 上限（2^63-1）的大版本号：k8s ResourceVersion 语义上是 uint64，
// 若用 ParseInt 会解析失败回退字符串比较，导致位数不同的两个大版本号排错顺序。
// 9.99e18（19 位）必须排在 1e19（20 位）前面。
func TestSortEvents_HugeNumericOrder(t *testing.T) {
	mk := func(rv string) *eventv1.Event {
		return &eventv1.Event{ObjectMeta: metav1.ObjectMeta{ResourceVersion: rv}}
	}
	events := sortEvents{mk("10000000000000000000"), mk("9999999999999999999")}
	sort.Sort(events)
	assert.Equal(t, "9999999999999999999", events[0].ResourceVersion)
	assert.Equal(t, "10000000000000000000", events[1].ResourceVersion)
}

// FuzzSortEventsLess 属性：自反性（Less(i,i)==false）与反对称性
// （Less(a,b) 与 Less(b,a) 不得同时为真），防止排序错乱回归。
func FuzzSortEventsLess(f *testing.F) {
	f.Add("9", "10")
	f.Add("", "")
	f.Add("999999999999999999999", "1")
	f.Add("abc", "def")
	f.Fuzz(func(t *testing.T, rvA, rvB string) {
		s := sortEvents{
			{ObjectMeta: metav1.ObjectMeta{ResourceVersion: rvA}},
			{ObjectMeta: metav1.ObjectMeta{ResourceVersion: rvB}},
		}
		if s.Less(0, 0) {
			t.Fatalf("Less(i,i) 必须为 false, rv=%q", rvA)
		}
		if s.Less(1, 1) {
			t.Fatalf("Less(i,i) 必须为 false, rv=%q", rvB)
		}
		if s.Less(0, 1) && s.Less(1, 0) {
			t.Fatalf("Less 反对称性被破坏: rv=%q vs rv=%q", rvA, rvB)
		}
	})
}
