package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/metrics"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestNewMetricsSvc(t *testing.T) {
	svc, _ := newMetricsSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.k8sBiz)
	assert.NotNil(t, svc.logger)
	assert.NotNil(t, svc.timer)
	assert.NotNil(t, svc.accessBiz)
}

func TestMetricsSvc_TopPod_Success(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{Name: "namespace1"}, nil)
	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(&v1beta1.PodMetrics{}, nil)
	k8sRepo.EXPECT().GetCpuAndMemoryQuantity(gomock.Any()).Return(&resource.Quantity{}, &resource.Quantity{})

	res, err := svc.TopPod(newAdminUserCtx(), &metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestMetricsSvc_TopPod_PodNotRunning(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{Name: "namespace1"}, nil)
	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(nil, errors.New("error"))
	k8sRepo.EXPECT().IsPodRunning("namespace1", "pod1").Return(false, "pod not running")

	res, err := svc.TopPod(newAdminUserCtx(), &metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMetricsSvc_TopPod_Error(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{Name: "namespace1"}, nil)
	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(nil, errors.New("error"))
	k8sRepo.EXPECT().IsPodRunning("namespace1", "pod1").Return(true, "")

	res, err := svc.TopPod(newAdminUserCtx(), &metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

// 回归防护：TopPod 接收原始 k8s namespace 名，私有命名空间的 pod 资源用量
// 不允许被非授权用户读取。去掉 checkNamespaceAccess 门禁本测试必须失败，
// 且是干净的 assert 失败（去除后流程继续，返回 200 响应而非 panic）。
func TestMetricsSvc_TopPod_AccessDenied(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{Private: true, CreatorEmail: "other@x.com"}, nil)
	// 去除门禁后 GetPodMetrics 成功 + metrics 组装成功 → 返回 200 响应，
	// assert.ErrorIs 干净失败（err=nil），不产生误导性 panic。
	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1beta1.PodMetrics{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetCpuAndMemoryQuantity(gomock.Any()).Return(&resource.Quantity{}, &resource.Quantity{}).AnyTimes()

	res, err := svc.TopPod(newOtherUserCtx(), &metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	})

	assert.Nil(t, res)
	assert.ErrorIs(t, err, biz.ErrorPermissionDenied)
}

func TestMetricsSvc_TopPod_FindByNameError(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(nil, errors.New("find error"))

	res, err := svc.TopPod(newAdminUserCtx(), &metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	})

	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestMetricsSvc_CpuMemoryInNamespace_Success(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	nsRepo := mocks.nsRepo
	k8sRepo := mocks.k8sRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "a"}, nil)
	k8sRepo.EXPECT().GetCpuAndMemoryInNamespace(gomock.Any(), "a").Return("cpu", "memory")

	res, err := svc.CpuMemoryInNamespace(newAdminUserCtx(), &metrics.CpuMemoryInNamespaceRequest{
		NamespaceId: 1,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "cpu", res.Cpu)
	assert.Equal(t, "memory", res.Memory)
}

func TestMetricsSvc_CpuMemoryInNamespace_NamespaceNotFound(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("namespace not found"))

	res, err := svc.CpuMemoryInNamespace(context.TODO(), &metrics.CpuMemoryInNamespaceRequest{
		NamespaceId: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMetricsSvc_CpuMemoryInNamespace_Error(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))

	res, err := svc.CpuMemoryInNamespace(context.TODO(), &metrics.CpuMemoryInNamespaceRequest{
		NamespaceId: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMetricsSvc_CpuMemoryInProject_Success(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	projRepo := mocks.projectRepo
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo

	p := &biz.Project{NamespaceID: 1}
	projRepo.EXPECT().Show(gomock.Any(), 1).Return(p, nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "a"}, nil)
	me := []v1beta1.PodMetrics{}
	k8sRepo.EXPECT().GetAllPodMetrics(gomock.Any(), p).Return(me)
	k8sRepo.EXPECT().GetCpuAndMemory(gomock.Any(), me).Return("cpu", "memory")

	res, err := svc.CpuMemoryInProject(newAdminUserCtx(), &metrics.CpuMemoryInProjectRequest{
		ProjectId: 1,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "cpu", res.Cpu)
	assert.Equal(t, "memory", res.Memory)
}

func TestMetricsSvc_CpuMemoryInProject_ProjectNotFound(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	projRepo := mocks.projectRepo

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("project not found"))

	res, err := svc.CpuMemoryInProject(context.TODO(), &metrics.CpuMemoryInProjectRequest{
		ProjectId: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMetricsSvc_CpuMemoryInProject_Error(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	projRepo := mocks.projectRepo

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))

	res, err := svc.CpuMemoryInProject(context.TODO(), &metrics.CpuMemoryInProjectRequest{
		ProjectId: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMetricsSvc_CpuMemoryInProject_NamespaceShowError(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	projRepo := mocks.projectRepo
	nsRepo := mocks.nsRepo

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("namespace error"))

	res, err := svc.CpuMemoryInProject(newAdminUserCtx(), &metrics.CpuMemoryInProjectRequest{
		ProjectId: 1,
	})

	assert.Nil(t, res)
	assert.Error(t, err)
}

// 回归防护：私有命名空间的项目资源用量不允许被非 admin / 非创建者 / 非成员读取。
// 去掉 CpuMemoryInProject 里的 CanAccess 检查，本测试必须失败。
func TestMetricsSvc_CpuMemoryInProject_AccessDenied(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	projRepo := mocks.projectRepo
	nsRepo := mocks.nsRepo

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: true, CreatorEmail: "other@x.com"}, nil)

	res, err := svc.CpuMemoryInProject(newOtherUserCtx(), &metrics.CpuMemoryInProjectRequest{
		ProjectId: 1,
	})

	assert.Nil(t, res)
	assert.ErrorIs(t, err, biz.ErrorPermissionDenied)
}

// 回归防护：私有命名空间资源用量不允许被非授权用户读取。
// 去掉 CpuMemoryInNamespace 里的 CanAccess 检查，本测试必须失败。
func TestMetricsSvc_CpuMemoryInNamespace_AccessDenied(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: true, CreatorEmail: "other@x.com"}, nil)

	res, err := svc.CpuMemoryInNamespace(newOtherUserCtx(), &metrics.CpuMemoryInNamespaceRequest{
		NamespaceId: 1,
	})

	assert.Nil(t, res)
	assert.ErrorIs(t, err, biz.ErrorPermissionDenied)
}

func TestMetricsSvc_Metrics_Success(t *testing.T) {
	svc, _ := newMetricsSvcWithMocks(t)

	res := svc.buildTopPodResponse(&biz.PodSample{
		Cpu:    &resource.Quantity{},
		Memory: &resource.Quantity{},
	})

	assert.NotNil(t, res)
	assert.Equal(t, float64(0), res.Cpu)
	assert.Equal(t, float64(0), res.Memory)
	assert.Equal(t, "0 m", res.HumanizeCpu)
	assert.Equal(t, "0 B", res.HumanizeMemory)
}

func TestMetricsSvc_Metrics_NonZeroValues(t *testing.T) {
	svc, _ := newMetricsSvcWithMocks(t)

	cpuQuantity := resource.NewMilliQuantity(1500, resource.DecimalSI)
	memoryQuantity := resource.NewQuantity(1024, resource.BinarySI)

	res := svc.buildTopPodResponse(&biz.PodSample{Cpu: cpuQuantity, Memory: memoryQuantity})

	assert.NotNil(t, res)
	assert.Equal(t, float64(1500), res.Cpu)
	assert.Equal(t, float64(2), res.Memory)
	assert.Equal(t, "1500 m", res.HumanizeCpu)
	assert.Equal(t, "1.0 kB", res.HumanizeMemory)
}

func TestMetricsSvc_Metrics_FractionalMemory(t *testing.T) {
	svc, _ := newMetricsSvcWithMocks(t)

	cpuQuantity := resource.NewMilliQuantity(100, resource.DecimalSI)
	// 分数内存值（如 1.5Gi）无法精确转为 int64 → memory.AsInt64() 返回 ok=false，
	// 覆盖 HumanizeMemory 的兜底日志分支；HumanizeCpu 仍是统一毫核格式。
	memoryQuantity := resource.MustParse("1.5Gi")

	res := svc.buildTopPodResponse(&biz.PodSample{Cpu: cpuQuantity, Memory: &memoryQuantity})

	assert.NotNil(t, res)
	assert.Equal(t, "100 m", res.HumanizeCpu)
	// AsInt64 失败 → asInt64 为 0，HumanizeMemory 回退为 "0 B"。
	assert.Equal(t, "0 B", res.HumanizeMemory)
}

func TestMetricsSvc_StreamTopPod_Success(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo

	// 注入 50ms 采样间隔：默认 5s 在测试超时窗口内永远不触发 ticker.C，周期发送路径从未被真实执行。
	// tickDuration 是包级 var（wire 无法为 wire.Struct 注入裸 time.Duration），覆盖后恢复。
	tickDuration = 50 * time.Millisecond
	t.Cleanup(func() { tickDuration = 5 * time.Second })

	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{Name: "namespace1"}, nil)
	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(&v1beta1.PodMetrics{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetCpuAndMemoryQuantity(gomock.Any()).Return(&resource.Quantity{}, &resource.Quantity{}).AnyTimes()

	server := NewMockMetrics_StreamTopPodServer(mocks.ctrl)
	timeout, cancelFunc := context.WithTimeout(newAdminUserCtx(), 500*time.Millisecond)
	defer cancelFunc()
	server.EXPECT().Context().Return(timeout).AnyTimes()
	// MinTimes(2)：初始 fn() 一次 + 至少一次 ticker.C 周期触发，锁定周期发送路径真实执行。
	server.EXPECT().Send(gomock.Any()).Return(nil).MinTimes(2)

	err := svc.StreamTopPod(&metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	}, server)

	assert.Nil(t, err)
}

func TestMetricsSvc_StreamTopPod_Error(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{Name: "namespace1"}, nil)
	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(nil, errors.New("x"))
	k8sRepo.EXPECT().IsPodRunning("namespace1", "pod1").Return(true, "")

	server := NewMockMetrics_StreamTopPodServer(mocks.ctrl)
	server.EXPECT().Context().Return(newAdminUserCtx()).AnyTimes()
	server.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()

	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(nil, errors.New("x"))
	k8sRepo.EXPECT().IsPodRunning("namespace1", "pod1").Return(false, "")

	err := svc.StreamTopPod(&metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	}, server)

	// pod 已不存在：返回 NotFound 状态（与 TopPod 一致），不再抛裸 metrics 错误。
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestMetricsSvc_StreamTopPod_SendError(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{Name: "namespace1"}, nil)
	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(&v1beta1.PodMetrics{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetCpuAndMemoryQuantity(gomock.Any()).Return(&resource.Quantity{}, &resource.Quantity{}).AnyTimes()

	server := NewMockMetrics_StreamTopPodServer(mocks.ctrl)
	server.EXPECT().Context().Return(newAdminUserCtx()).AnyTimes()
	server.EXPECT().Send(gomock.Any()).Return(errors.New("send error")).AnyTimes()

	err := svc.StreamTopPod(&metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	}, server)

	assert.NotNil(t, err)
}

func TestMetricsSvc_StreamTopPod_PodNotRunning(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{Name: "namespace1"}, nil)
	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), "namespace1", "pod1").Return(nil, errors.New("error")).AnyTimes()
	k8sRepo.EXPECT().IsPodRunning("namespace1", "pod1").Return(false, "pod not running").AnyTimes()

	server := NewMockMetrics_StreamTopPodServer(mocks.ctrl)
	timeout, cancelFunc := context.WithTimeout(newAdminUserCtx(), 3*time.Second)
	defer cancelFunc()
	server.EXPECT().Context().Return(timeout).AnyTimes()

	err := svc.StreamTopPod(&metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	}, server)

	// 回归防护：pod 不存在时 reason 必须透出（之前被 _ 丢弃），客户端拿到的是 NotFound + 原因。
	assert.Equal(t, codes.NotFound, status.Code(err))
	assert.Equal(t, "pod not running", status.Convert(err).Message())
}

// 回归防护：StreamTopPod 与 TopPod 一致，私有命名空间不允许未授权用户订阅 pod 资源用量。
// 去掉 checkNamespaceAccess 门禁本测试必须失败，且是干净的 assert 失败
// （去除后进入流循环直到 ctx 超时返回 nil，而非 panic）。
func TestMetricsSvc_StreamTopPod_AccessDenied(t *testing.T) {
	svc, mocks := newMetricsSvcWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().FindByName(gomock.Any(), "namespace1").Return(&biz.Namespace{Private: true, CreatorEmail: "other@x.com"}, nil)
	// 去除门禁后 GetPodMetrics 成功 + Send 成功 → 循环直到 ctx 3s 超时返回 nil，
	// assert.ErrorIs 干净失败（err=nil），不产生误导性 panic。
	k8sRepo.EXPECT().GetPodMetrics(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1beta1.PodMetrics{}, nil).AnyTimes()
	k8sRepo.EXPECT().GetCpuAndMemoryQuantity(gomock.Any()).Return(&resource.Quantity{}, &resource.Quantity{}).AnyTimes()

	server := NewMockMetrics_StreamTopPodServer(mocks.ctrl)
	timeout, cancelFunc := context.WithTimeout(newOtherUserCtx(), 3*time.Second)
	defer cancelFunc()
	server.EXPECT().Context().Return(timeout).AnyTimes()
	server.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()

	err := svc.StreamTopPod(&metrics.TopPodRequest{
		Namespace: "namespace1",
		Pod:       "pod1",
	}, server)

	assert.ErrorIs(t, err, biz.ErrorPermissionDenied)
}

// metricsSvcMocks 聚合 metricsSvc 的全部下游 mock，由 newMetricsSvcWithMocks 统一构造。
type metricsSvcMocks struct {
	ctrl        *gomock.Controller
	k8sRepo     *data.MockK8sRepo
	projectRepo *data.MockProjectRepo
	nsRepo      *data.MockNamespaceRepo
}

// newMetricsSvcWithMocks 构造被测 svc 并返回全部 mock。
// 各测试按需取用字段，未用到的 mock 不设 expectation 即可。
func newMetricsSvcWithMocks(t *testing.T) (*metricsSvc, *metricsSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &metricsSvcMocks{
		ctrl:        ctrl,
		k8sRepo:     data.NewMockK8sRepo(ctrl),
		projectRepo: data.NewMockProjectRepo(ctrl),
		nsRepo:      data.NewMockNamespaceRepo(ctrl),
	}
	logger := mlog.NewForConfig(nil)
	s, ok := NewMetricsSvc(MetricsSvcDeps{
		Timer:      timer.NewReal(),
		K8sBiz:     biz.NewK8sBiz(mocks.k8sRepo),
		MetricsBiz: biz.NewMetricsBiz(biz.NewK8sBiz(mocks.k8sRepo)),
		Logger:     logger,
		AccessBiz:  biz.NewAccessBiz(logger, biz.NewNsRepoBiz(mocks.nsRepo), biz.NewProjectBiz(logger, mocks.projectRepo, mocks.k8sRepo)),
	}).(*metricsSvc)
	if !ok {
		panic("NewMetricsSvc returned unexpected type")
	}
	return s, mocks
}
