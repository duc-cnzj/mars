package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars-client/v4/metrics"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	fake2 "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

func TestMetricsSvc_CpuMemoryInNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	_, err := new(MetricsSvc).CpuMemoryInNamespace(context.TODO(), &metrics.CpuMemoryInNamespaceRequest{NamespaceId: 1})
	assert.Error(t, err)
	p := &models.Project{
		Name: "p",
		Namespace: models.Namespace{
			Name: "ns",
		},
	}
	db.Create(p)
	mk := &fake2.Clientset{}
	mk.AddReactor("list", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		res := &v1beta1.PodMetricsList{
			ListMeta: metav1.ListMeta{
				ResourceVersion: "2",
			},
			Items: []v1beta1.PodMetrics{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "ns",
					},
					Timestamp: metav1.Time{},
					Window:    metav1.Duration{},
					Containers: []v1beta1.ContainerMetrics{
						{
							Name: "container1",
							Usage: v1.ResourceList{
								v1.ResourceCPU:     *resource.NewMilliQuantity(4, resource.DecimalSI),
								v1.ResourceMemory:  *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
								v1.ResourceStorage: *resource.NewQuantity(6*(1000*1000), resource.DecimalSI),
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: "ns",
					},
					Timestamp: metav1.Time{},
					Window:    metav1.Duration{},
					Containers: []v1beta1.ContainerMetrics{
						{
							Name: "container1",
							Usage: v1.ResourceList{
								v1.ResourceCPU:     *resource.NewMilliQuantity(4, resource.DecimalSI),
								v1.ResourceMemory:  *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
								v1.ResourceStorage: *resource.NewQuantity(6*(1000*1000), resource.DecimalSI),
							},
						},
					},
				},
			},
		}
		return true, res, nil
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{MetricsClient: mk}).AnyTimes()
	namespace, err := new(MetricsSvc).CpuMemoryInNamespace(context.TODO(), &metrics.CpuMemoryInNamespaceRequest{NamespaceId: int64(p.Namespace.ID)})
	assert.Nil(t, err)
	assert.Equal(t, "8 m", namespace.Cpu)
	assert.Equal(t, "10 MB", namespace.Memory)
}

func TestMetricsSvc_CpuMemoryInProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	p := &models.Project{
		Name:         "p",
		PodSelectors: "app=test",
		Namespace: models.Namespace{
			Name: "ns",
		},
	}
	db.Create(p)
	mk := &fake2.Clientset{}
	mk.AddReactor("list", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		res := &v1beta1.PodMetricsList{
			ListMeta: metav1.ListMeta{
				ResourceVersion: "2",
			},
			Items: []v1beta1.PodMetrics{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "ns",
						Labels: map[string]string{
							"app": "test",
						},
					},
					Timestamp: metav1.Time{},
					Window:    metav1.Duration{},
					Containers: []v1beta1.ContainerMetrics{
						{
							Name: "container1",
							Usage: v1.ResourceList{
								v1.ResourceCPU:     *resource.NewMilliQuantity(4, resource.DecimalSI),
								v1.ResourceMemory:  *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
								v1.ResourceStorage: *resource.NewQuantity(6*(1000*1000), resource.DecimalSI),
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: "ns",
						Labels: map[string]string{
							"app": "xxx",
						},
					},
					Timestamp: metav1.Time{},
					Window:    metav1.Duration{},
					Containers: []v1beta1.ContainerMetrics{
						{
							Name: "container1",
							Usage: v1.ResourceList{
								v1.ResourceCPU:     *resource.NewMilliQuantity(4, resource.DecimalSI),
								v1.ResourceMemory:  *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
								v1.ResourceStorage: *resource.NewQuantity(6*(1000*1000), resource.DecimalSI),
							},
						},
					},
				},
			},
		}
		return true, res, nil
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{MetricsClient: mk}).AnyTimes()
	namespace, err := new(MetricsSvc).CpuMemoryInProject(context.TODO(), &metrics.CpuMemoryInProjectRequest{ProjectId: int64(p.ID)})
	assert.Nil(t, err)
	assert.Equal(t, "4 m", namespace.Cpu)
	assert.Equal(t, "5.0 MB", namespace.Memory)
}

type topServerMock struct {
	sendError bool
	ctx       context.Context
	result    []*metrics.TopPodResponse
	metrics.Metrics_StreamTopPodServer
}

func (t *topServerMock) Context() context.Context {
	return t.ctx
}

func (t *topServerMock) Send(response *metrics.TopPodResponse) error {
	t.result = append(t.result, response)
	if t.sendError {
		return errors.New("err")
	}
	return nil
}

func TestMetricsSvc_StreamTopPod(t *testing.T) {
	preTickDuration := tickDuration
	tickDuration = time.Millisecond * 500
	defer func() {
		tickDuration = preTickDuration
	}()
	preNow := now
	now = func() string {
		return "2022-01-01"
	}
	defer func() {
		now = preNow
	}()
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	app.EXPECT().Done().Return(nil).Times(3)
	instance.SetInstance(app)
	fk := fake.NewSimpleClientset(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "pod",
		},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	})
	mk := &fake2.Clientset{}
	mk.AddReactor("get", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1beta1.PodMetrics{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod",
				Namespace: "ns",
			},
			Timestamp: metav1.Time{},
			Window:    metav1.Duration{},
			Containers: []v1beta1.ContainerMetrics{
				{
					Name: "container1",
					Usage: v1.ResourceList{
						v1.ResourceCPU:     *resource.NewMilliQuantity(4, resource.DecimalSI),
						v1.ResourceMemory:  *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
						v1.ResourceStorage: *resource.NewQuantity(6*(1000*1000), resource.DecimalSI),
					},
				},
			},
		}, nil
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client:        fk,
		MetricsClient: mk,
	}).AnyTimes()
	ctx, cancel := context.WithCancel(context.TODO())
	done := make(chan struct{})
	tsm := &topServerMock{ctx: ctx}
	go func() {
		defer close(done)
		err := new(MetricsSvc).StreamTopPod(&metrics.TopPodRequest{
			Namespace: "ns",
			Pod:       "pod",
		}, tsm)
		t.Log(err)
		assert.Equal(t, "context canceled", err.Error())
	}()
	select {
	case <-time.After(1200 * time.Millisecond):
		cancel()
	}
	_, ok := <-done
	assert.False(t, ok)
	assert.Len(t, tsm.result, 3)
	for _, response := range tsm.result {
		assert.Equal(t, "2022-01-01", response.Time)
		assert.Equal(t, float64(4), response.Cpu)
		assert.Equal(t, float64(5000), response.Memory)
		assert.Equal(t, "4 m", response.HumanizeCpu)
		assert.Equal(t, "5.0 MB", response.HumanizeMemory)
	}
}

func TestMetricsSvc_StreamTopPod2(t *testing.T) {
	preTickDuration := tickDuration
	tickDuration = time.Millisecond * 500
	defer func() {
		tickDuration = preTickDuration
	}()
	preNow := now
	now = func() string {
		return "2022-01-01"
	}
	defer func() {
		now = preNow
	}()
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	ch := make(chan struct{})
	close(ch)
	app.EXPECT().Done().Return(ch).Times(1)
	instance.SetInstance(app)
	fk := fake.NewSimpleClientset(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "pod",
		},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	})
	mk := &fake2.Clientset{}
	mk.AddReactor("get", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1beta1.PodMetrics{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod",
				Namespace: "ns",
			},
			Timestamp: metav1.Time{},
			Window:    metav1.Duration{},
			Containers: []v1beta1.ContainerMetrics{
				{
					Name: "container1",
					Usage: v1.ResourceList{
						v1.ResourceCPU:     *resource.NewMilliQuantity(4, resource.DecimalSI),
						v1.ResourceMemory:  *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
						v1.ResourceStorage: *resource.NewQuantity(6*(1000*1000), resource.DecimalSI),
					},
				},
			},
		}, nil
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client:        fk,
		MetricsClient: mk,
	}).AnyTimes()
	ctx, cancel := context.WithCancel(context.TODO())
	done := make(chan struct{})
	tsm := &topServerMock{ctx: ctx}
	go func() {
		defer close(done)
		err := new(MetricsSvc).StreamTopPod(&metrics.TopPodRequest{
			Namespace: "ns",
			Pod:       "pod",
		}, tsm)
		assert.Nil(t, err)
	}()
	select {
	case <-time.After(1200 * time.Millisecond):
		cancel()
	}
	_, ok := <-done
	assert.False(t, ok)
	assert.Len(t, tsm.result, 1)
}

func TestMetricsSvc_StreamTopPod_Error(t *testing.T) {
	preTickDuration := tickDuration
	tickDuration = time.Millisecond * 500
	defer func() {
		tickDuration = preTickDuration
	}()
	preNow := now
	now = func() string {
		return "2022-01-01"
	}
	defer func() {
		now = preNow
	}()
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	app.EXPECT().Done().Return(nil).Times(0)
	instance.SetInstance(app)
	fk := fake.NewSimpleClientset()
	mk := &fake2.Clientset{}
	mk.AddReactor("get", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("xxx")
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client:        fk,
		MetricsClient: mk,
	}).AnyTimes()
	ctx, cancel := context.WithCancel(context.TODO())
	done := make(chan struct{})
	tsm := &topServerMock{ctx: ctx}
	go func() {
		defer close(done)
		err := new(MetricsSvc).StreamTopPod(&metrics.TopPodRequest{
			Namespace: "ns",
			Pod:       "pod",
		}, tsm)
		assert.Equal(t, "xxx", err.Error())
	}()
	select {
	case <-time.After(1200 * time.Millisecond):
		cancel()
	}
	_, ok := <-done
	assert.False(t, ok)
	assert.Len(t, tsm.result, 0)
}

func TestMetricsSvc_StreamTopPod_Error2(t *testing.T) {
	preTickDuration := tickDuration
	tickDuration = time.Millisecond * 500
	defer func() {
		tickDuration = preTickDuration
	}()
	preNow := now
	now = func() string {
		return "2022-01-01"
	}
	defer func() {
		now = preNow
	}()
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	app.EXPECT().Done().Return(nil).Times(0)
	instance.SetInstance(app)
	fk := fake.NewSimpleClientset()
	mk := &fake2.Clientset{}
	mk.AddReactor("get", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1beta1.PodMetrics{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod",
				Namespace: "ns",
			},
			Timestamp: metav1.Time{},
			Window:    metav1.Duration{},
			Containers: []v1beta1.ContainerMetrics{
				{
					Name: "container1",
					Usage: v1.ResourceList{
						v1.ResourceCPU:     *resource.NewMilliQuantity(4, resource.DecimalSI),
						v1.ResourceMemory:  *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
						v1.ResourceStorage: *resource.NewQuantity(6*(1000*1000), resource.DecimalSI),
					},
				},
			},
		}, nil
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client:        fk,
		MetricsClient: mk,
	}).AnyTimes()
	ctx, cancel := context.WithCancel(context.TODO())
	done := make(chan struct{})
	tsm := &topServerMock{ctx: ctx, sendError: true}
	go func() {
		defer close(done)
		err := new(MetricsSvc).StreamTopPod(&metrics.TopPodRequest{
			Namespace: "ns",
			Pod:       "pod",
		}, tsm)
		assert.Equal(t, "err", err.Error())
	}()
	select {
	case <-time.After(1200 * time.Millisecond):
		cancel()
	}
	_, ok := <-done
	assert.False(t, ok)
	assert.Len(t, tsm.result, 1)
}

func TestMetricsSvc_TopPod(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	fk := fake.NewSimpleClientset()
	mk := &fake2.Clientset{}
	mk.AddReactor("get", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("xx")
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client:        fk,
		MetricsClient: mk,
	}).AnyTimes()
	ms := new(MetricsSvc)
	_, err := ms.TopPod(context.TODO(), &metrics.TopPodRequest{
		Namespace: "ns",
		Pod:       "pod",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, `pods "pod" not found`, fromError.Message())
}

func TestMetricsSvc_TopPod2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	fk := fake.NewSimpleClientset(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "pod",
		},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	})
	mk := &fake2.Clientset{}
	mk.AddReactor("get", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("xx")
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client:        fk,
		MetricsClient: mk,
	}).AnyTimes()
	ms := new(MetricsSvc)
	_, err := ms.TopPod(context.TODO(), &metrics.TopPodRequest{
		Namespace: "ns",
		Pod:       "pod",
	})
	assert.Equal(t, "xx", err.Error())
}

func TestMetricsSvc_TopPod3(t *testing.T) {
	preNow := now
	now = func() string {
		return "2022-01-01"
	}
	defer func() {
		now = preNow
	}()
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	fk := fake.NewSimpleClientset(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "pod",
		},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	})
	mk := &fake2.Clientset{}
	mk.AddReactor("get", "pods", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1beta1.PodMetrics{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod",
				Namespace: "ns",
			},
			Timestamp: metav1.Time{},
			Window:    metav1.Duration{},
			Containers: []v1beta1.ContainerMetrics{
				{
					Name: "container1",
					Usage: v1.ResourceList{
						v1.ResourceCPU:     *resource.NewMilliQuantity(4, resource.DecimalSI),
						v1.ResourceMemory:  *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
						v1.ResourceStorage: *resource.NewQuantity(6*(1000*1000), resource.DecimalSI),
					},
				},
			},
		}, nil
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client:        fk,
		MetricsClient: mk,
	}).AnyTimes()
	ms := new(MetricsSvc)
	res, err := ms.TopPod(context.TODO(), &metrics.TopPodRequest{
		Namespace: "ns",
		Pod:       "pod",
	})
	assert.Nil(t, err)
	wants := &metrics.TopPodResponse{
		Cpu:            4,
		Memory:         5000,
		HumanizeCpu:    "4 m",
		HumanizeMemory: "5.0 MB",
		Time:           "2022-01-01",
		Length:         30,
	}
	assert.Equal(t, wants.Cpu, res.Cpu)
	assert.Equal(t, wants.Time, res.Time)
	assert.Equal(t, wants.HumanizeCpu, res.HumanizeCpu)
	assert.Equal(t, wants.Memory, res.Memory)
	assert.Equal(t, wants.HumanizeMemory, res.HumanizeMemory)
	assert.Equal(t, wants.Length, res.Length)
}

func TestMetricsSvc_metrics(t *testing.T) {
	response := new(MetricsSvc).metrics(&v1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "ns",
		},
		Timestamp: metav1.Time{},
		Window:    metav1.Duration{},
		Containers: []v1beta1.ContainerMetrics{
			{
				Name: "container1",
				Usage: v1.ResourceList{
					v1.ResourceCPU:     *resource.NewMilliQuantity(4000, resource.DecimalSI),
					v1.ResourceMemory:  *resource.NewQuantity(5*(1000*1000), resource.DecimalSI),
					v1.ResourceStorage: *resource.NewQuantity(6*(1000*1000), resource.DecimalSI),
				},
			},
		},
	})
	assert.Equal(t, "4.000", response.HumanizeCpu)
}
