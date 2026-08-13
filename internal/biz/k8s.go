package biz

import (
	"context"
	"errors"
	"regexp"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// K8sBiz 封装 k8s 基础设施操作（secret/namespace/pod/exec/文件拷贝/指标等）。
type K8sBiz interface {
	// SplitManifests 按文档分隔符把 manifest 切分为多段。
	SplitManifests(manifest string) []string
	// AddTlsSecret 校验命名后创建 TLS 证书 secret。
	AddTlsSecret(ns string, name string, key string, crt string) (*corev1.Secret, error)
	// GetPodMetrics 查询单个 pod 的资源指标。
	GetPodMetrics(ctx context.Context, namespace, podName string) (*v1beta1.PodMetrics, error)
	// CreateDockerSecret 创建 docker registry 登录 secret。
	CreateDockerSecret(ctx context.Context, namespace string) (*corev1.Secret, error)
	// GetNamespace 查询命名空间。
	GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error)
	// CreateNamespace 校验名称后创建命名空间。
	CreateNamespace(ctx context.Context, name string) (*corev1.Namespace, error)
	// LogStream 流式返回 pod 容器日志（channel 逐行投递，EOF 后关闭）。
	LogStream(ctx context.Context, namespace, pod, container string) (chan []byte, error)
	// GetPodLogs 一次性返回 pod 容器日志文本。
	GetPodLogs(ctx context.Context, namespace, podName string, options *corev1.PodLogOptions) (string, error)
	// FindDefaultContainer 推断 pod 的默认主容器名。
	FindDefaultContainer(ctx context.Context, namespace string, pod string) (string, error)
	// GetPod 按名称查询 pod。
	GetPod(namespace, podName string) (*corev1.Pod, error)
	// ListEvents 列出命名空间下的事件。
	ListEvents(namespace string) ([]*eventv1.Event, error)
	// IsPodRunning 判断 pod 是否处于运行态，非运行态同时返回原因。
	IsPodRunning(namespace, podName string) (running bool, notRunningReason string)
	// GetPodSelectorsByManifest 从 manifest 推导 pod 的 label selectors。
	GetPodSelectorsByManifest(manifests []string) []string
	// GetCpuAndMemoryInNamespace 汇总命名空间内全部 pod 的 CPU/内存用量。
	GetCpuAndMemoryInNamespace(ctx context.Context, namespace string) (string, string)
	// GetCpuAndMemory 汇总一组 pod 指标列表的 CPU/内存用量。
	GetCpuAndMemory(ctx context.Context, list []v1beta1.PodMetrics) (string, string)
	// GetCpuAndMemoryQuantity 从单个 pod 指标提取 CPU/内存量。
	GetCpuAndMemoryQuantity(pod v1beta1.PodMetrics) (cpu *resource.Quantity, memory *resource.Quantity)
	// ClusterInfo 返回集群健康与资源汇总信息。
	ClusterInfo() *ClusterInfo
	// Execute 在容器内执行命令，stdin/stdout/stderr 接线由输入提供。
	Execute(ctx context.Context, c *Container, input *ExecuteInput) error
	// DeleteSecret 校验命名后删除 secret。
	DeleteSecret(ctx context.Context, namespace, secret string) error
	// DeleteNamespace 校验名称后删除命名空间。
	DeleteNamespace(ctx context.Context, name string) error
	// ForceDeletePod 强制删除 pod：以指定宽限期与后台传播策略删除，
	// gracePeriodSeconds=0 即不等优雅终止直接移除，等价 kubectl delete pod --force。
	// 用于卡死/无法正常终止的 pod。
	ForceDeletePod(ctx context.Context, namespace, pod string, gracePeriodSeconds int64) error
	// GetAllPodMetrics 按项目 selectors 返回全部 pod 的指标列表。
	GetAllPodMetrics(ctx context.Context, proj *Project) []v1beta1.PodMetrics
	// CopyFileToPod 把文件拷贝进 pod 指定容器，返回落库的 File 记录。
	CopyFileToPod(ctx context.Context, input *CopyFileToPodInput) (*File, error)
	// CopyFromPod 从 pod 拷贝文件出来，返回落库的 File 记录。
	CopyFromPod(ctx context.Context, input *CopyFromPodInput) (*File, error)
}

type k8sBiz struct {
	k8sRepo K8sRepo
}

// NewK8sBiz 构造 k8s biz。
func NewK8sBiz(k8sRepo K8sRepo) K8sBiz {
	return &k8sBiz{k8sRepo: k8sRepo}
}

// SplitManifests 按文档分隔符把 manifest 切分为多段（透传 repo）。
func (k *k8sBiz) SplitManifests(manifest string) []string { return k.k8sRepo.SplitManifests(manifest) }

// AddTlsSecret 校验命名后创建 TLS secret。
func (k *k8sBiz) AddTlsSecret(ns string, name string, key string, crt string) (*corev1.Secret, error) {
	if ns == "" || name == "" {
		return nil, errs.WrapInvalidArgument(errors.New("namespace 或 secret 名称不能为空"), "add tls secret")
	}
	return k.k8sRepo.AddTlsSecret(ns, name, key, crt)
}

// GetPodMetrics 查询单个 pod 的指标（透传 repo）。
func (k *k8sBiz) GetPodMetrics(ctx context.Context, namespace, podName string) (*v1beta1.PodMetrics, error) {
	return k.k8sRepo.GetPodMetrics(ctx, namespace, podName)
}

// CreateDockerSecret 创建 docker registry secret（透传 repo）。
func (k *k8sBiz) CreateDockerSecret(ctx context.Context, namespace string) (*corev1.Secret, error) {
	return k.k8sRepo.CreateDockerSecret(ctx, namespace)
}

// GetNamespace 查询 namespace（透传 repo）。
func (k *k8sBiz) GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	return k.k8sRepo.GetNamespace(ctx, name)
}

// CreateNamespace 校验名称后创建 namespace。
func (k *k8sBiz) CreateNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	if name == "" {
		return nil, errs.WrapInvalidArgument(errors.New("namespace 名称不能为空"), "create namespace")
	}
	return k.k8sRepo.CreateNamespace(ctx, name)
}

// LogStream 返回 pod 容器的实时日志流（透传 repo）。
func (k *k8sBiz) LogStream(ctx context.Context, namespace, pod, container string) (chan []byte, error) {
	return k.k8sRepo.LogStream(ctx, namespace, pod, container)
}

// GetPodLogs 返回 pod 容器的一次性日志（透传 repo）。
func (k *k8sBiz) GetPodLogs(ctx context.Context, namespace, podName string, options *corev1.PodLogOptions) (string, error) {
	return k.k8sRepo.GetPodLogs(ctx, namespace, podName, options)
}

// FindDefaultContainer 查询 pod 的默认容器名（透传 repo）。
func (k *k8sBiz) FindDefaultContainer(ctx context.Context, namespace string, pod string) (string, error) {
	return k.k8sRepo.FindDefaultContainer(ctx, namespace, pod)
}

// GetPod 按名称查询 pod（透传 repo）。
func (k *k8sBiz) GetPod(namespace, podName string) (*corev1.Pod, error) {
	return k.k8sRepo.GetPod(namespace, podName)
}

// ListEvents 列出 namespace 下的 k8s 事件（透传 repo）。
func (k *k8sBiz) ListEvents(namespace string) ([]*eventv1.Event, error) {
	return k.k8sRepo.ListEvents(namespace)
}

// IsPodRunning 判断 pod 是否处于 Running，未运行同时返回原因（透传 repo）。
func (k *k8sBiz) IsPodRunning(namespace, podName string) (running bool, notRunningReason string) {
	return k.k8sRepo.IsPodRunning(namespace, podName)
}

// GetPodSelectorsByManifest 从 manifests 推导 pod 选择器列表（透传 repo）。
func (k *k8sBiz) GetPodSelectorsByManifest(manifests []string) []string {
	return k.k8sRepo.GetPodSelectorsByManifest(manifests)
}

// GetCpuAndMemoryInNamespace 聚合 namespace 下全部 pod 的 CPU/内存用量（透传 repo）。
func (k *k8sBiz) GetCpuAndMemoryInNamespace(ctx context.Context, namespace string) (string, string) {
	return k.k8sRepo.GetCpuAndMemoryInNamespace(ctx, namespace)
}

// GetCpuAndMemory 聚合给定 pod 指标列表的 CPU/内存用量（透传 repo）。
func (k *k8sBiz) GetCpuAndMemory(ctx context.Context, list []v1beta1.PodMetrics) (string, string) {
	return k.k8sRepo.GetCpuAndMemory(ctx, list)
}

// GetCpuAndMemoryQuantity 提取单个 pod 指标的 CPU/内存用量（透传 repo）。
func (k *k8sBiz) GetCpuAndMemoryQuantity(pod v1beta1.PodMetrics) (cpu *resource.Quantity, memory *resource.Quantity) {
	return k.k8sRepo.GetCpuAndMemoryQuantity(pod)
}

// ClusterInfo 返回集群信息（透传 repo）。
func (k *k8sBiz) ClusterInfo() *ClusterInfo { return k.k8sRepo.ClusterInfo() }

// Execute 执行容器内命令，stdin/stdout 接线由调用方经 input 提供（透传 repo）。
func (k *k8sBiz) Execute(ctx context.Context, c *Container, input *ExecuteInput) error {
	return k.k8sRepo.Execute(ctx, c, input)
}

// DeleteSecret 校验命名后删除 secret。
func (k *k8sBiz) DeleteSecret(ctx context.Context, namespace, secret string) error {
	if namespace == "" || secret == "" {
		return errs.WrapInvalidArgument(errors.New("namespace 或 secret 名称不能为空"), "delete secret")
	}
	return k.k8sRepo.DeleteSecret(ctx, namespace, secret)
}

// DeleteNamespace 校验名称后删除 namespace。
func (k *k8sBiz) DeleteNamespace(ctx context.Context, name string) error {
	if name == "" {
		return errs.WrapInvalidArgument(errors.New("namespace 名称不能为空"), "delete namespace")
	}
	return k.k8sRepo.DeleteNamespace(ctx, name)
}

// ForceDeletePod 强制删除 pod：空参或负宽限期返回 InvalidArgument，否则以
// 指定宽限期 + PropagationPolicy=Background 透传 repo；gracePeriodSeconds=0 立即移除。
func (k *k8sBiz) ForceDeletePod(ctx context.Context, namespace, pod string, gracePeriodSeconds int64) error {
	if namespace == "" || pod == "" {
		return errs.WrapInvalidArgument(errors.New("namespace 或 pod 名称不能为空"), "force delete pod")
	}
	if gracePeriodSeconds < 0 {
		return errs.WrapInvalidArgument(errors.New("grace period seconds 不能为负数"), "force delete pod")
	}
	background := metav1.DeletePropagationBackground
	return k.k8sRepo.DeletePod(ctx, namespace, pod, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
		PropagationPolicy:  &background,
	})
}

// GetAllPodMetrics 按项目 PodSelectors 返回全部 pod 指标（透传 repo）。
func (k *k8sBiz) GetAllPodMetrics(ctx context.Context, proj *Project) []v1beta1.PodMetrics {
	return k.k8sRepo.GetAllPodMetrics(ctx, proj)
}

// CopyFileToPod 把文件拷贝进 pod 并返回落库的 File 记录（透传 repo）。
func (k *k8sBiz) CopyFileToPod(ctx context.Context, input *CopyFileToPodInput) (*File, error) {
	return k.k8sRepo.CopyFileToPod(ctx, input)
}

// CopyFromPod 从 pod 拷贝文件出来并返回落库的 File 记录（透传 repo）。
func (k *k8sBiz) CopyFromPod(ctx context.Context, input *CopyFromPodInput) (*File, error) {
	return k.k8sRepo.CopyFromPod(ctx, input)
}

// ProjectCpuMemory 计算项目下全部 pod 的 CPU/内存聚合用量（毫核 + 字节字符串）。
// metrics.CpuMemoryInProject 与 project.MemoryCpuAndEndpoints 两个传输服务共用同一
// 聚合规则（先取项目全部 pod metrics，再聚合成量），下沉到 biz 消除两处重复，防止
// 聚合口径（pod 选择/单位换算）在 transport 层各自演化而漂移。
func ProjectCpuMemory(ctx context.Context, k8sBiz K8sBiz, proj *Project) (string, string) {
	return k8sBiz.GetCpuAndMemory(ctx, k8sBiz.GetAllPodMetrics(ctx, proj))
}

var hostMatch = regexp.MustCompile(`\s+([\w-_]*)<\s*.Host\d+\s*>`)

// GetPreOccupiedLenByValuesYaml 计算 values.yaml 中 host 占位符（<.HostN>）前缀的最大长度，
// 供 DomainManager 生成域名时计算 pre-occupied 长度使用。纯函数，无外部依赖。
func GetPreOccupiedLenByValuesYaml(values string) int {
	var sub = 0
	if len(values) > 0 {
		submatch := hostMatch.FindAllStringSubmatch(values, -1)
		for _, i := range submatch {
			if len(i) == 2 {
				sub = max(sub, len(i[1]))
			}
		}
	}
	return sub
}

// K8sRepo 是 k8s 集群操作端口，聚合资源读写、指标、日志与容器执行能力。
type K8sRepo interface {
	// SplitManifests 把多对象 manifest 拆分为独立片段列表。
	SplitManifests(manifest string) []string
	// AddTlsSecret 创建 TLS 证书 secret。
	AddTlsSecret(ns string, name string, key string, crt string) (*corev1.Secret, error)
	// GetPodMetrics 查询单个 pod 的指标。
	GetPodMetrics(ctx context.Context, namespace, podName string) (*v1beta1.PodMetrics, error)
	// CreateDockerSecret 创建 docker registry 登录 secret。
	CreateDockerSecret(ctx context.Context, namespace string) (*corev1.Secret, error)
	// GetNamespace 查询命名空间。
	GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error)
	// CreateNamespace 创建命名空间。
	CreateNamespace(ctx context.Context, name string) (*corev1.Namespace, error)
	// LogStream 流式返回 pod 容器日志（channel 逐行投递，EOF 后关闭）。
	LogStream(ctx context.Context, namespace, pod, container string) (chan []byte, error)
	// GetPodLogs 一次性返回 pod 容器日志文本。
	GetPodLogs(ctx context.Context, namespace, podName string, options *corev1.PodLogOptions) (string, error)
	// FindDefaultContainer 推断 pod 的默认主容器名。
	FindDefaultContainer(ctx context.Context, namespace string, pod string) (string, error)
	// GetPod 查询单个 pod。
	GetPod(namespace, podName string) (*corev1.Pod, error)
	// ListEvents 列出命名空间下的事件。
	ListEvents(namespace string) ([]*eventv1.Event, error)
	// IsPodRunning 判断 pod 是否处于运行态，非运行态返回原因。
	IsPodRunning(namespace, podName string) (running bool, notRunningReason string)
	// GetPodSelectorsByManifest 从 manifest 推导 pod 的 label selectors。
	GetPodSelectorsByManifest(manifests []string) []string
	// GetCpuAndMemoryInNamespace 汇总命名空间内全部 pod 的 CPU/内存用量。
	GetCpuAndMemoryInNamespace(ctx context.Context, namespace string) (string, string)
	// GetCpuAndMemory 汇总一组 pod 指标列表的 CPU/内存用量。
	GetCpuAndMemory(ctx context.Context, list []v1beta1.PodMetrics) (string, string)
	// GetCpuAndMemoryQuantity 从单个 pod 指标提取 CPU/内存量。
	GetCpuAndMemoryQuantity(pod v1beta1.PodMetrics) (cpu *resource.Quantity, memory *resource.Quantity)
	// ClusterInfo 返回集群健康与资源汇总信息。
	ClusterInfo() *ClusterInfo
	// Execute 在容器内执行命令，stdin/stdout/stderr 接线由输入提供。
	Execute(ctx context.Context, c *Container, input *ExecuteInput) error
	// GetSecret 读取命名空间下指定名称的 k8s secret（domainmanager 插件校验 TLS 证书用）。
	GetSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error)
	// UpdateSecret 更新指定命名空间下的 secret 内容（cron TLS 证书同步用）。
	UpdateSecret(ctx context.Context, namespace, name string, secret *corev1.Secret) (*corev1.Secret, error)
	// CreateDockerSecrets 为命名空间下指定的 docker registry 集合创建
	// DockerConfigJson 类型 secret。servers 是 config.ImagePullSecrets 的子集，
	// 实现侧自行按 server 过滤凭据，避免把 config 类型泄漏进业务层。
	CreateDockerSecrets(ctx context.Context, namespace string, servers []string) (*corev1.Secret, error)
	// SubscribePodEvents 订阅 informer 的 Pod 生命周期事件，返回事件通道与取消订阅函数。
	// 取消订阅会关闭事件通道，消费方 range 循环随之退出；data 层内部以 Obj 泛型 fanout
	// 实现，转换收敛在适配层，业务侧只见领域 PodEvent 类型。
	SubscribePodEvents(listener string) (<-chan PodEvent, func())
	// DeleteSecret 删除命名空间下的 secret。
	DeleteSecret(ctx context.Context, namespace, secret string) error
	// DeleteNamespace 删除命名空间。
	DeleteNamespace(ctx context.Context, name string) error
	// DeletePod 删除 pod，删除策略由 opts 决定（强制删除传 GracePeriodSeconds=0）。
	DeletePod(ctx context.Context, namespace, pod string, opts metav1.DeleteOptions) error
	// GetAllPodMetrics 返回项目全部 pod 的指标列表。
	GetAllPodMetrics(ctx context.Context, proj *Project) []v1beta1.PodMetrics
	// CopyFileToPod 把文件拷贝进 pod 指定容器。
	CopyFileToPod(ctx context.Context, input *CopyFileToPodInput) (*File, error)
	// CopyFromPod 从 pod 拷贝文件出来。
	CopyFromPod(ctx context.Context, input *CopyFromPodInput) (*File, error)
	// 以下为容器拓扑推导与 endpoint 编排所需的低层读取原语。
	// ListPodsBySelectors 按一组 label selector（如 "app=a,env=prod"）列出命名空间内的
	// 运行中 Pod，selector 解析与 informer 过滤属于基础设施细节，收敛在 data 侧。
	ListPodsBySelectors(namespace string, selectors []string) ([]*corev1.Pod, error)
	// GetReplicaSet 查询指定 ReplicaSet。
	GetReplicaSet(namespace, name string) (*appsv1.ReplicaSet, error)
	// ListServices 列出命名空间下全部 Service。
	ListServices(namespace string) ([]*corev1.Service, error)
	// ListIngresses 列出命名空间下全部 Ingress。
	ListIngresses(namespace string) ([]*networkingv1.Ingress, error)
	// ListHTTPRoutes 列出命名空间下全部 Gateway API HTTPRoute。
	ListHTTPRoutes(namespace string) ([]*gatewayv1.HTTPRoute, error)
	// GatewayApiInstalled 返回集群是否安装 Gateway API。
	GatewayApiInstalled() bool
	// ExternalIp 返回集群对外访问 IP。
	ExternalIp() string
}
