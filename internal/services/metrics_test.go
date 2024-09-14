package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v5/metrics"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestNewMetricsSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewMetricsSvc(timer.NewReal(), repo.NewMockK8sRepo(m), mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), repo.NewMockNamespaceRepo(m))
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*metricsSvc).k8sRepo)
	assert.NotNil(t, svc.(*metricsSvc).logger)
	assert.NotNil(t, svc.(*metricsSvc).timer)
	assert.NotNil(t, svc.(*metricsSvc).projRepo)
	assert.NotNil(t, svc.(*metricsSvc).nsRepo)
}

func TestMetricsSvc_TopPod_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), k8sRepo, mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), repo.NewMockNamespaceRepo(m))

	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(&v1beta1.PodMetrics{}, nil)
	k8sRepo.EXPECT().GetCpuAndMemoryQuantity(gomock.Any()).Return(&resource.Quantity{}, &resource.Quantity{})

	res, err := svc.TopPod(context.TODO(), &metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestMetricsSvc_TopPod_PodNotRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), k8sRepo, mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), repo.NewMockNamespaceRepo(m))

	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(nil, errors.New("error"))
	k8sRepo.EXPECT().IsPodRunning("namespace1", "pod1").Return(false, "pod not running")

	res, err := svc.TopPod(context.TODO(), &metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMetricsSvc_TopPod_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), k8sRepo, mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), repo.NewMockNamespaceRepo(m))

	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(nil, errors.New("error"))
	k8sRepo.EXPECT().IsPodRunning("namespace1", "pod1").Return(true, "")

	res, err := svc.TopPod(context.TODO(), &metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMetricsSvc_CpuMemoryInNamespace_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), k8sRepo, mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), nsRepo)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Namespace{Name: "a"}, nil)
	k8sRepo.EXPECT().GetCpuAndMemoryInNamespace(gomock.Any(), "a").Return("cpu", "memory")

	res, err := svc.CpuMemoryInNamespace(context.TODO(), &metrics.CpuMemoryInNamespaceRequest{
		NamespaceId: 1,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "cpu", res.Cpu)
	assert.Equal(t, "memory", res.Memory)
}

func TestMetricsSvc_CpuMemoryInNamespace_NamespaceNotFound(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), repo.NewMockK8sRepo(m), mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), nsRepo)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("namespace not found"))

	res, err := svc.CpuMemoryInNamespace(context.TODO(), &metrics.CpuMemoryInNamespaceRequest{
		NamespaceId: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMetricsSvc_CpuMemoryInNamespace_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := repo.NewMockNamespaceRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), repo.NewMockK8sRepo(m), mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), nsRepo)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))

	res, err := svc.CpuMemoryInNamespace(context.TODO(), &metrics.CpuMemoryInNamespaceRequest{
		NamespaceId: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMetricsSvc_CpuMemoryInProject_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projRepo := repo.NewMockProjectRepo(m)
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), k8sRepo, mlog.NewForConfig(nil), projRepo, repo.NewMockNamespaceRepo(m))

	p := &repo.Project{}
	projRepo.EXPECT().Show(gomock.Any(), 1).Return(p, nil)
	me := []v1beta1.PodMetrics{}
	k8sRepo.EXPECT().GetAllPodMetrics(gomock.Any(), p).Return(me)
	k8sRepo.EXPECT().GetCpuAndMemory(gomock.Any(), me).Return("cpu", "memory")

	res, err := svc.CpuMemoryInProject(context.TODO(), &metrics.CpuMemoryInProjectRequest{
		ProjectId: 1,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "cpu", res.Cpu)
	assert.Equal(t, "memory", res.Memory)
}

func TestMetricsSvc_CpuMemoryInProject_ProjectNotFound(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projRepo := repo.NewMockProjectRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), repo.NewMockK8sRepo(m), mlog.NewForConfig(nil), projRepo, repo.NewMockNamespaceRepo(m))

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("project not found"))

	res, err := svc.CpuMemoryInProject(context.TODO(), &metrics.CpuMemoryInProjectRequest{
		ProjectId: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMetricsSvc_CpuMemoryInProject_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projRepo := repo.NewMockProjectRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), repo.NewMockK8sRepo(m), mlog.NewForConfig(nil), projRepo, repo.NewMockNamespaceRepo(m))

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))

	res, err := svc.CpuMemoryInProject(context.TODO(), &metrics.CpuMemoryInProjectRequest{
		ProjectId: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMetricsSvc_Metrics_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), k8sRepo, mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), repo.NewMockNamespaceRepo(m)).(*metricsSvc)

	k8sRepo.EXPECT().GetCpuAndMemoryQuantity(gomock.Any()).Return(&resource.Quantity{}, &resource.Quantity{})

	res := svc.metrics(&v1beta1.PodMetrics{})

	assert.NotNil(t, res)
	assert.Equal(t, float64(0), res.Cpu)
	assert.Equal(t, float64(0), res.Memory)
	assert.Equal(t, "0 m", res.HumanizeCpu)
	assert.Equal(t, "0 B", res.HumanizeMemory)
}

func TestMetricsSvc_Metrics_NonZeroValues(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), k8sRepo, mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), repo.NewMockNamespaceRepo(m)).(*metricsSvc)

	cpuQuantity := resource.NewMilliQuantity(1500, resource.DecimalSI)
	memoryQuantity := resource.NewQuantity(1024, resource.BinarySI)
	k8sRepo.EXPECT().GetCpuAndMemoryQuantity(gomock.Any()).Return(cpuQuantity, memoryQuantity)

	res := svc.metrics(&v1beta1.PodMetrics{})

	assert.NotNil(t, res)
	assert.Equal(t, float64(1500), res.Cpu)
	assert.Equal(t, float64(2), res.Memory)
	assert.Equal(t, "1.500", res.HumanizeCpu)
	assert.Equal(t, "1.0 kB", res.HumanizeMemory)
}

func TestMetricsSvc_StreamTopPod_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), k8sRepo, mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), repo.NewMockNamespaceRepo(m))

	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(&v1beta1.PodMetrics{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetCpuAndMemoryQuantity(gomock.Any()).Return(&resource.Quantity{}, &resource.Quantity{}).AnyTimes()

	server := NewMockMetrics_StreamTopPodServer(m)
	timeout, cancelFunc := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancelFunc()
	server.EXPECT().Context().Return(timeout).AnyTimes()
	server.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()

	err := svc.StreamTopPod(&metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	}, server)

	assert.Nil(t, err)
}

func TestMetricsSvc_StreamTopPod_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), k8sRepo, mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), repo.NewMockNamespaceRepo(m))

	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(nil, errors.New("x"))
	k8sRepo.EXPECT().IsPodRunning("namespace1", "pod1").Return(true, "")

	server := NewMockMetrics_StreamTopPodServer(m)
	server.EXPECT().Context().Return(context.TODO()).AnyTimes()
	server.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()

	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(nil, errors.New("x"))
	k8sRepo.EXPECT().IsPodRunning("namespace1", "pod1").Return(false, "")

	err := svc.StreamTopPod(&metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	}, server)

	assert.Equal(t, "x", err.Error())
}

func TestMetricsSvc_StreamTopPod_SendError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), k8sRepo, mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), repo.NewMockNamespaceRepo(m))

	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(&v1beta1.PodMetrics{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetCpuAndMemoryQuantity(gomock.Any()).Return(&resource.Quantity{}, &resource.Quantity{}).AnyTimes()

	server := NewMockMetrics_StreamTopPodServer(m)
	server.EXPECT().Context().Return(context.TODO()).AnyTimes()
	server.EXPECT().Send(gomock.Any()).Return(errors.New("send error")).AnyTimes()

	err := svc.StreamTopPod(&metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	}, server)

	assert.NotNil(t, err)
}

func TestMetricsSvc_StreamTopPod_PodNotRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := repo.NewMockK8sRepo(m)
	svc := NewMetricsSvc(timer.NewReal(), k8sRepo, mlog.NewForConfig(nil), repo.NewMockProjectRepo(m), repo.NewMockNamespaceRepo(m))

	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(nil, errors.New("error")).AnyTimes()
	k8sRepo.EXPECT().IsPodRunning("namespace1", "pod1").Return(false, "pod not running").AnyTimes()

	server := NewMockMetrics_StreamTopPodServer(m)
	timeout, cancelFunc := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancelFunc()
	server.EXPECT().Context().Return(timeout).AnyTimes()

	err := svc.StreamTopPod(&metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	}, server)

	assert.Equal(t, "error", err.Error())
}
