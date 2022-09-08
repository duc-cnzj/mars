package bootstrappers

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"

	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestK8sClientBootstrapper_Bootstrap(t *testing.T) {}

func TestK8sClientBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&K8sClientBootstrapper{}).Tags())
}

func Test_filterEvent(t *testing.T) {
	var tests = []struct {
		nsPrefix string

		reason    string
		kind      string
		namespace string
		name      string
		wants     bool
	}{
		{
			nsPrefix:  "ns",
			reason:    "",
			kind:      "Pod",
			namespace: "ns-xxx",
			name:      "app",
			wants:     true,
		},
		{
			nsPrefix:  "ns",
			reason:    "Unhealthy",
			kind:      "Pod",
			namespace: "ns-xxx",
			name:      "app-unhealthy",
			wants:     false,
		},
		{
			reason:    "",
			nsPrefix:  "ns",
			kind:      "Deployment",
			namespace: "xx-xxx",
			name:      "app",
			wants:     false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.namespace, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, filterEvent(tt.nsPrefix)(&eventsv1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tt.namespace,
				},
				Reason: tt.reason,
				Regarding: v1.ObjectReference{
					Kind:      tt.kind,
					Namespace: tt.namespace,
					Name:      tt.name,
				},
			}))
		})
	}
}

func Test_filterPod(t *testing.T) {
	var tests = []struct {
		nsPrefix string

		namespace string
		name      string
		wants     bool
	}{
		{
			nsPrefix:  "ns",
			namespace: "ns-xxx",
			name:      "app",
			wants:     true,
		},
		{
			nsPrefix:  "ns",
			namespace: "xx-xxx",
			name:      "app",
			wants:     false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.namespace, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, filterPod(tt.nsPrefix)(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tt.namespace,
					Name:      tt.name,
				},
			}))
		})
	}

}

func Test_fanOut_AddListener(t *testing.T) {
	ch := make(chan contracts.Obj[*v1.Pod], 10)
	ch2 := make(chan contracts.Obj[*v1.Pod], 10)
	fo := &fanOut[*v1.Pod]{
		name:      "test-pod",
		ch:        ch,
		listeners: make(map[string]chan<- contracts.Obj[*v1.Pod]),
	}
	fo.AddListener("aa", ch2)
	fo.AddListener("aa", ch2)
	fo.AddListener("bb", ch2)
	assert.Len(t, fo.listeners, 2)
}

func Test_fanOut_RemoveListener(t *testing.T) {
	ch := make(chan contracts.Obj[*v1.Pod], 10)
	ch2 := make(chan contracts.Obj[*v1.Pod], 10)
	fo := &fanOut[*v1.Pod]{
		name:      "test-pod",
		ch:        ch,
		listeners: make(map[string]chan<- contracts.Obj[*v1.Pod]),
	}
	fo.AddListener("aa", ch2)
	fo.AddListener("aa", ch2)
	fo.AddListener("bb", ch2)
	fo.RemoveListener("aa")
	fo.RemoveListener("aa")
	fo.RemoveListener("bb")
	fo.RemoveListener("bb")
	assert.Len(t, fo.listeners, 0)
}

func Test_fanOut_Distribute(t *testing.T) {
	ch := make(chan contracts.Obj[*v1.Pod], 10)
	ch2 := make(chan contracts.Obj[*v1.Pod], 10)
	fo := &fanOut[*v1.Pod]{
		name:      "test-pod",
		ch:        ch,
		listeners: make(map[string]chan<- contracts.Obj[*v1.Pod]),
	}
	fo.AddListener("aa", ch2)
	ctx := make(chan struct{}, 1)
	go func() {
		pod := &v1.Pod{}
		ch <- contracts.NewObj(nil, pod, contracts.Add)
		p := <-ch2
		assert.Len(t, ch2, 0)
		assert.Len(t, ch, 0)
		assert.Same(t, pod, p.Current())
		time.Sleep(2 * time.Second)
		close(ctx)
	}()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		fo.Distribute(ctx)
	}()
	go func() {
		defer wg.Done()
		fo.Distribute(ctx)
	}()
	wg.Wait()
}

func TestStartable_start(t *testing.T) {
	var s Startable
	var num int64
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if s.start() {
				atomic.AddInt64(&num, 1)
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, int64(1), atomic.LoadInt64(&num))
}
