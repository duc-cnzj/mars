package data

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/data/k8sutil"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/dustin/go-humanize"
	"github.com/mholt/archiver/v3"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel/attribute"
	"helm.sh/helm/v3/pkg/releaseutil"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	clientgoexec "k8s.io/client-go/util/exec"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// defaultContainerAnnotationName 是标注 Pod 默认容器的 kubectl annotation 键。
const defaultContainerAnnotationName = "kubectl.kubernetes.io/default-container"

// Status* 是集群健康度定级枚举：bad/not good/health。
const (
	StatusBad     biz.ClusterStatus = "bad"
	StatusNotGood biz.ClusterStatus = "not good"
	StatusHealth  biz.ClusterStatus = "health"
)

var _ biz.K8sRepo = (*k8sRepo)(nil)

// k8sRepo 是 biz.K8sRepo 的持久化实现：聚合 k8s 客户端、上传器、归档器与远程
// 执行器，承载 Pod 日志/复制、Secret/Namespace 管理、集群信息统计等全部 k8s 能力。
type k8sRepo struct {
	logger        mlog.Logger
	uploader      uploader.Uploader
	maxUploadSize uint64
	archiver      Archiver
	executor      ExecutorManager
	data          dataStore
	fileRepo      biz.FileRepo
	timer         timer.Timer
}

// NewK8sRepo 构造 k8s repo：从 data 读取上传大小上限，组合归档器与远程执行器。
func NewK8sRepo(
	logger mlog.Logger,
	timer timer.Timer,
	data dataStore,
	fileRepo biz.FileRepo,
	uploader uploader.Uploader,
	archiver Archiver,
	remoteExecutor ExecutorManager,
) biz.K8sRepo {
	return &k8sRepo{
		fileRepo:      fileRepo,
		timer:         timer,
		data:          data,
		logger:        logger.WithModule("repo/k8s"),
		uploader:      uploader,
		maxUploadSize: data.Config().MaxUploadSize(),
		archiver:      archiver,
		executor:      remoteExecutor,
	}
}

// CopyFromPod 把 Pod 内文件复制到上传存储并落一条文件记录：先经 exec 校验目标
// 必须是文件且位于工作目录内，再 tar 打包经 SPDY 流拉回，最后转存并登记。
func (repo *k8sRepo) CopyFromPod(ctx context.Context, input *biz.CopyFromPodInput) (f *biz.File, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/CopyFromPod")
	defer func() { endSpan(span, err) }()

	var file uploader.File

	lsbf := &bytes.Buffer{}

	_ = repo.Execute(ctx, &biz.Container{
		Namespace: input.Namespace,
		Pod:       input.Pod,
		Container: input.Container,
	}, &biz.ExecuteInput{
		Stdout: lsbf,
		TTY:    false,
		Cmd:    []string{"sh", "-c", fmt.Sprintf("test -f %s && echo 1 || echo 0", input.FilePath)},
	})
	isFile := strings.Trim(lsbf.String(), "\n") == "1"
	if !isFile {
		return nil, errs.WrapInvalidArgument(errors.New("下载内容必须是文件"), "k8s download file")
	}

	pwdbf := &bytes.Buffer{}
	err = repo.Execute(ctx, &biz.Container{
		Namespace: input.Namespace,
		Pod:       input.Pod,
		Container: input.Container,
	}, &biz.ExecuteInput{
		Stdout: pwdbf,
		TTY:    false,
		Cmd:    []string{"sh", "-c", "pwd"},
	})
	if err != nil {
		return nil, err
	}
	base := strings.Trim(pwdbf.String(), "\n")

	if !strings.HasPrefix(input.FilePath, base) {
		return nil, errs.WrapInvalidArgument(errors.New("invalid file path"), "k8s download file")
	}

	// base 为容器内 pwd 绝对路径，FilePath 已在上方校验以 base 为前缀，二者同为绝对路径，
	// filepath.Rel 的绝对/相对混用错误分支恒不可达，直接取相对路径。
	input.FilePath, _ = filepath.Rel(base, input.FilePath)
	repo.logger.Debugf("[CopyFromPod]: rel: %q base: %q", input.FilePath, base)

	bf := &bytes.Buffer{}
	fileCopy := repo.executor.NewFileCopy(5, bf)

	remotePath := input.FilePath

	filename := fmt.Sprintf("%s-%s.tar", input.Pod, rand.String(10))

	up := repo.uploader.Disk("podfile")

	file, err = up.NewFile(
		fmt.Sprintf("%s/%s/%s/%s",
			input.UserName,
			repo.timer.Now().Format("2006-01-02"),
			fmt.Sprintf("%s-%s", repo.timer.Now().Format("15-04-05"), rand.String(20)),
			filename),
	)
	if err != nil {
		return nil, errs.Wrap(err, "k8s download file")
	}
	defer func() {
		file.Close()
		if err != nil {
			up.Delete(file.Name())
		}
	}()

	if err = fileCopy.CopyFromPod(ctx, k8sutil.CopyFileSpec{
		PodName:       input.Pod,
		PodNamespace:  input.Namespace,
		ContainerName: input.Container,
		File:          k8sutil.NewRemotePath(remotePath),
	}, file); err != nil {
		return nil, errs.Wrap(err, fmt.Sprintf("copy from pod error, output: %v", bf.String()))
	}

	var stat os.FileInfo
	stat, err = file.Stat()
	if err != nil {
		return nil, errs.Wrap(err, "k8s download file")
	}
	return repo.fileRepo.Create(ctx, &biz.CreateFileInput{
		Path:       file.Name(),
		Username:   input.UserName,
		Size:       uint64(stat.Size()),
		UploadType: up.Type(),
		Namespace:  input.Namespace,
		Pod:        input.Pod,
		Container:  input.Container,
	})
}

// SplitManifests 用 helm 的 SplitManifests 切分多资源 manifest 并按字符串排序。
// 不走朴素 strings.Split("---")：有些 secret 的值本身含 ---，朴素切分会解析异常。
func (repo *k8sRepo) SplitManifests(manifest string) []string {
	mapManifests := releaseutil.SplitManifests(manifest)
	var manifests = make([]string, 0, len(mapManifests))
	for _, s := range mapManifests {
		manifests = append(manifests, s)
	}
	sort.Strings(manifests)

	return manifests
}

// AddTlsSecret 在命名空间创建 tls 类型的 secret，写入 key/crt 与 created-by 注解。
func (repo *k8sRepo) AddTlsSecret(ns string, name string, key string, crt string) (*corev1.Secret, error) {
	secret, err := repo.data.K8s().Client.CoreV1().Secrets(ns).Create(context.TODO(), &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Annotations: map[string]string{
				"created-by": "mars",
			},
		},
		StringData: map[string]string{
			"tls.key": key,
			"tls.crt": crt,
		},
		Type: corev1.SecretTypeTLS,
	}, metav1.CreateOptions{})
	return secret, errs.Wrap(err, "add tls secret")
}

// GetPodMetrics 返回单个 Pod 的指标（CPU/内存用量）。
func (repo *k8sRepo) GetPodMetrics(ctx context.Context, namespace, podName string) (metrics *v1beta1.PodMetrics, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/GetPodMetrics")
	defer func() { endSpan(span, err) }()
	metrics, err = repo.data.K8s().MetricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	return metrics, errs.Wrap(err, "get pod metrics")
}

// GetAllPodMetrics 汇总项目全部 Pod selector 命中的指标。单个 selector 解析/查询
// 失败只记录日志并跳过（返回签名无 error，无法冒泡），不影响其余 selector 结果。
func (repo *k8sRepo) GetAllPodMetrics(ctx context.Context, proj *biz.Project) []v1beta1.PodMetrics {
	ctx, span := tracer.Start(ctx, "k8sRepo/GetAllPodMetrics")
	defer span.End()
	metricses := repo.data.K8s().MetricsClient.MetricsV1beta1().PodMetricses(proj.Namespace.Name)
	var list []v1beta1.PodMetrics
	if len(proj.PodSelectors) == 0 {
		return nil
	}
	for _, podlabels := range proj.PodSelectors {
		labelsMap, err := labels.ConvertSelectorToLabelsMap(podlabels)
		if err != nil {
			// selector 非法时跳过该组，避免 labelsMap 为 nil 退化成列出全命名空间 pod。
			repo.logger.Debugf("[GetAllPodMetrics]: parse selector %q: %v", podlabels, err)
			continue
		}
		ret, err := repo.data.K8s().PodLister.Pods(proj.Namespace.Name).List(labelsMap.AsSelector())
		if err != nil {
			repo.logger.Debugf("[GetAllPodMetrics]: list pods %q: %v", podlabels, err)
			continue
		}
		if len(ret) == 0 {
			continue
		}
		l, err := metricses.List(ctx, metav1.ListOptions{
			LabelSelector: podlabels,
		})
		if err != nil {
			repo.logger.Debugf("[GetAllPodMetrics]: list metrics %q: %v", podlabels, err)
			continue
		}
		repo.logger.DebugCtx(ctx, "[GetAllPodMetrics]: ", podlabels, " ", len(l.Items))

		list = append(list, l.Items...)
	}

	return list
}

// DeleteNamespace 删除命名空间。
func (repo *k8sRepo) DeleteNamespace(ctx context.Context, name string) (err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/DeleteNamespace")
	defer func() { endSpan(span, err) }()
	return errs.Wrap(repo.data.K8s().Client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{}), "delete namespace")
}

// DeleteSecret 删除命名空间下的 secret。
func (repo *k8sRepo) DeleteSecret(ctx context.Context, namespace, secret string) (err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/DeleteSecret")
	defer func() { endSpan(span, err) }()
	return errs.Wrap(repo.data.K8s().Client.CoreV1().Secrets(namespace).Delete(ctx, secret, metav1.DeleteOptions{}), "delete secret")
}

// DeletePod 删除命名空间下的 pod，删除策略由 opts 决定（如强制删除传
// GracePeriodSeconds=0）。仅做 k8s API 级删除，不参与业务状态机。
func (repo *k8sRepo) DeletePod(ctx context.Context, namespace, pod string, opts metav1.DeleteOptions) (err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/DeletePod")
	defer func() { endSpan(span, err) }()
	return errs.Wrap(repo.data.K8s().Client.CoreV1().Pods(namespace).Delete(ctx, pod, opts), "delete pod")
}

// GetSecret 按命名空间与名称读取 secret。
func (repo *k8sRepo) GetSecret(ctx context.Context, namespace, name string) (secret *corev1.Secret, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/GetSecret")
	defer func() { endSpan(span, err) }()
	secret, err = repo.data.K8s().Client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	return secret, errs.Wrap(err, "get secret")
}

// CreateDockerSecret 为 namespace 创建包含全部已配置 registry 凭据的 docker secret。
func (repo *k8sRepo) CreateDockerSecret(ctx context.Context, namespace string) (secret *corev1.Secret, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/CreateDockerSecret")
	defer func() { endSpan(span, err) }()
	var servers []string
	for _, auth := range repo.data.Config().ImagePullSecrets {
		servers = append(servers, auth.Server)
	}
	return repo.CreateDockerSecrets(ctx, namespace, servers)
}

// CreateDockerSecrets 为 namespace 创建 DockerConfigJson 类型 secret，凭据取
// config.ImagePullSecrets 中 servers 命中的子集。servers 为空时生成空 auths 的 secret，
// 与旧 CreateDockerSecret 在零配置下的行为保持一致。
func (repo *k8sRepo) CreateDockerSecrets(ctx context.Context, namespace string, servers []string) (secret *corev1.Secret, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/CreateDockerSecrets")
	defer func() { endSpan(span, err) }()
	var entries = make(map[string]biz.DockerConfigEntry)
	for _, auth := range repo.data.Config().ImagePullSecrets {
		if lo.Contains(servers, auth.Server) {
			entries[auth.Server] = biz.DockerConfigEntry{
				Username: auth.Username,
				Password: auth.Password,
				Email:    auth.Email,
				Auth:     base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password)),
			}
		}
	}

	dockerCfgJSON := biz.DockerConfigJSON{
		Auths: entries,
	}

	marshal, _ := json.Marshal(dockerCfgJSON)

	secret, err = repo.data.K8s().Client.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "mars-" + strings.ToLower(rand.String(10)),
		},
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: marshal,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}, metav1.CreateOptions{})
	return secret, errs.Wrap(err, "create docker secrets")
}

// UpdateSecret 更新指定命名空间下的 secret 内容，返回更新后的 secret。
func (repo *k8sRepo) UpdateSecret(ctx context.Context, namespace, name string, secret *corev1.Secret) (updated *corev1.Secret, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/UpdateSecret")
	defer func() { endSpan(span, err) }()
	updated, err = repo.data.K8s().Client.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	return updated, errs.Wrap(err, "update secret")
}

// SubscribePodEvents 订阅 informer 的 Pod 事件并转换为领域 PodEvent 类型。
// 取消订阅函数会移除 fanout 监听并关闭事件通道，消费方 range 循环随之退出；
// 转换 goroutine 在通道关闭后退出，不留泄漏。
func (repo *k8sRepo) SubscribePodEvents(listener string) (<-chan biz.PodEvent, func()) {
	raw := make(chan Obj[*corev1.Pod], 500)
	out := make(chan biz.PodEvent, 500)
	repo.data.K8s().podFanOut.AddListener(listener, raw)
	go func() {
		defer close(out)
		for obj := range raw {
			out <- biz.PodEvent{
				Type:    biz.PodEventType(obj.Type()),
				Old:     obj.Old(),
				Current: obj.Current(),
			}
		}
	}()
	return out, func() {
		repo.data.K8s().podFanOut.RemoveListener(listener)
	}
}

// GetNamespace 读取命名空间。
func (repo *k8sRepo) GetNamespace(ctx context.Context, name string) (ns *corev1.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/GetNamespace")
	defer func() { endSpan(span, err) }()
	ns, err = repo.data.K8s().Client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	return ns, errs.Wrap(err, "get namespace")
}

// CreateNamespace 创建命名空间。
func (repo *k8sRepo) CreateNamespace(ctx context.Context, name string) (ns *corev1.Namespace, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/CreateNamespace")
	defer func() { endSpan(span, err) }()
	ns, err = repo.data.K8s().Client.CoreV1().
		Namespaces().
		Create(ctx,
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			},
			metav1.CreateOptions{},
		)
	return ns, errs.Wrap(err, "create namespace")
}

// Execute 在 Pod 容器内执行命令：经 executor 构造 SPDY exec 请求并透传输入输出流。
func (repo *k8sRepo) Execute(ctx context.Context, c *biz.Container, input *biz.ExecuteInput) (err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/Execute")
	defer func() { endSpan(span, err) }()
	return errs.Wrap(repo.executor.New().
		WithContainer(c.Namespace, c.Pod, c.Container).
		WithMethod("POST").
		WithCommand(input.Cmd).
		Execute(ctx, input), "k8s exec")
}

// GetPodLogs 一次性拉取 Pod 日志（按 options 过滤）并以字符串返回。
func (repo *k8sRepo) GetPodLogs(ctx context.Context, namespace, podName string, options *corev1.PodLogOptions) (out string, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/GetPodLogs")
	defer func() { endSpan(span, err) }()
	logs := repo.data.K8s().Client.CoreV1().Pods(namespace).GetLogs(podName, options)
	do := logs.Do(ctx)
	raw, err := do.Raw()
	return string(raw), errs.Wrap(err, "get pod logs")
}

// ListEvents 列出命名空间下全部事件。
func (repo *k8sRepo) ListEvents(namespace string) ([]*eventv1.Event, error) {
	events, err := repo.data.K8s().EventLister.Events(namespace).List(labels.Everything())
	return events, errs.Wrap(err, "list events")
}

// FindDefaultContainer 返回 Pod 的默认容器：优先命中 kubectl.kubernetes.io/
// default-container 注解且真实存在的容器，否则取第一个容器；无容器返回 NotFound。
func (repo *k8sRepo) FindDefaultContainer(ctx context.Context, namespace string, pod string) (container string, err error) {
	_, span := tracer.Start(ctx, "k8sRepo/FindDefaultContainer")
	defer func() { endSpan(span, err) }()
	corev1pod, err := repo.GetPod(namespace, pod)
	if err != nil {
		return "", err
	}
	if name := corev1pod.Annotations[defaultContainerAnnotationName]; len(name) > 0 {
		for _, co := range corev1pod.Spec.Containers {
			if name == co.Name {
				return name, nil
			}
		}
	}

	for _, co := range corev1pod.Spec.Containers {
		return co.Name, nil
	}

	return "", errs.NotFound("未找到容器")
}

// GetPod 经 informer lister 读取 Pod（不实时访问 apiserver）。
func (repo *k8sRepo) GetPod(namespace, podName string) (*corev1.Pod, error) {
	pod, err := repo.data.K8s().PodLister.Pods(namespace).Get(podName)
	return pod, errs.Wrap(err, "get pod")
}

// IsPodRunning 判断 Pod 是否 Running，非 Running 时返回人类可读的原因
// （Evicted/容器 Waiting 状态等），供上游决定是否提示或拦截操作。
func (repo *k8sRepo) IsPodRunning(namespace, podName string) (running bool, notRunningReason string) {
	podInfo, err := repo.data.K8s().PodLister.Pods(namespace).Get(podName)
	if err != nil {
		return false, err.Error()
	}

	if podInfo.Status.Phase == corev1.PodRunning {
		return true, ""
	}

	if podInfo.Status.Phase == corev1.PodFailed && podInfo.Status.Reason == "Evicted" {
		return false, fmt.Sprintf("po %s already evicted in namespace %s!", podName, namespace)
	}

	for _, status := range podInfo.Status.ContainerStatuses {
		return false, fmt.Sprintf("%s %s", status.State.Waiting.Reason, status.State.Waiting.Message)
	}

	return false, "pod not running."
}

// GetPodSelectorsByManifest 从一组资源 manifest 中提取 Pod label selector：
// Deployment/StatefulSet/DaemonSet 取 spec.selector，Job/CronJob 取模板 labels。
// 无法解码的资源跳过。参考 client-go#193 的 selector 序列化行为。
func (repo *k8sRepo) GetPodSelectorsByManifest(manifests []string) []string {
	var selectors []string
	info, _ := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
	for _, f := range manifests {
		obj, _, _ := info.Serializer.Decode([]byte(f), nil, nil)
		switch a := obj.(type) {
		case *appsv1.Deployment:
			selector, _ := metav1.LabelSelectorAsSelector(a.Spec.Selector)
			selectors = append(selectors, selector.String())
		case *appsv1.StatefulSet:
			selector, _ := metav1.LabelSelectorAsSelector(a.Spec.Selector)
			selectors = append(selectors, selector.String())
		case *appsv1.DaemonSet:
			selector, _ := metav1.LabelSelectorAsSelector(a.Spec.Selector)
			selectors = append(selectors, selector.String())
		case *batchv1.Job:
			jobPodLabels := a.Spec.Template.Labels
			if jobPodLabels != nil {
				selectors = append(selectors, labels.SelectorFromSet(jobPodLabels).String())
			}
		case *batchv1beta1.CronJob:
			jobPodLabels := a.Spec.JobTemplate.Spec.Template.Labels
			if jobPodLabels != nil {
				selectors = append(selectors, labels.SelectorFromSet(jobPodLabels).String())
			}
		case *batchv1.CronJob:
			jobPodLabels := a.Spec.JobTemplate.Spec.Template.Labels
			if jobPodLabels != nil {
				selectors = append(selectors, labels.SelectorFromSet(jobPodLabels).String())
			}
		default:
			repo.logger.Debugf("GetPodSelectorsByManifest Default: %#v", a)
		}
	}

	return selectors
}

// GetCpuAndMemoryInNamespace 汇总命名空间内全部 Pod 的 CPU/内存用量并格式化。
func (repo *k8sRepo) GetCpuAndMemoryInNamespace(ctx context.Context, namespace string) (string, string) {
	ctx, span := tracer.Start(ctx, "k8sRepo/GetCpuAndMemoryInNamespace")
	defer span.End()
	metricses := repo.data.K8s().MetricsClient.MetricsV1beta1().PodMetricses(namespace)
	list, _ := metricses.List(ctx, metav1.ListOptions{})
	return repo.GetCpuAndMemory(ctx, list.Items)
}

// GetCpuAndMemory 汇总一组 PodMetrics 的 CPU/内存用量，格式化为 "N m" / 人类可读字节。
func (repo *k8sRepo) GetCpuAndMemory(ctx context.Context, list []v1beta1.PodMetrics) (string, string) {
	_, span := tracer.Start(ctx, "k8sRepo/GetCpuAndMemory")
	defer span.End()
	var cpu, memory *resource.Quantity
	for _, item := range list {
		itemCPU, itemMemory := repo.GetCpuAndMemoryQuantity(item)
		if itemCPU == nil || itemMemory == nil {
			continue
		}
		if cpu == nil {
			cpu = itemCPU
		} else {
			cpu.Add(*itemCPU)
		}
		if memory == nil {
			memory = itemMemory
		} else {
			memory.Add(*itemMemory)
		}
	}

	var cpuStr, memoryStr = "0 m", "0 MB"

	if cpu != nil {
		cpuStr = fmt.Sprintf("%d m", cpu.MilliValue())
	}
	if memory != nil {
		asInt64, _ := memory.AsInt64()
		memoryStr = humanize.Bytes(uint64(asInt64))
	}

	return cpuStr, memoryStr
}

// GetCpuAndMemoryQuantity 累加单个 Pod 内全部容器的 CPU/内存用量；
// 无容器时返回双 nil。
func (repo *k8sRepo) GetCpuAndMemoryQuantity(pod v1beta1.PodMetrics) (cpu *resource.Quantity, memory *resource.Quantity) {
	for _, container := range pod.Containers {
		if cpu == nil {
			cpu = container.Usage.Cpu()
		} else {
			cpu.Add(*container.Usage.Cpu())
		}

		if memory == nil {
			memory = container.Usage.Memory()
		} else {
			memory.Add(*container.Usage.Memory())
		}
	}

	return cpu, memory
}

// copyToPodResult 记录 copyToPod 的执行结果：目标目录、双端输出与容器内落盘路径。
type copyToPodResult struct {
	TargetDir     string
	ErrOut        string
	StdOut        string
	ContainerPath string
	FileName      string
}

// copyToPod 把上传文件打包为 tar.gz 并经 SPDY 流管道解压到容器目标目录：
// 非 Local 类型上传先拉回本地再打包，包体经 io.Pipe 边拷贝边喂给 tar 解压命令。
func (repo *k8sRepo) copyToPod(ctx context.Context, namespace, pod, container, fpath, targetContainerDir string) (result *copyToPodResult, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/copyToPod")
	defer func() { endSpan(span, err) }()
	span.SetAttributes(
		attribute.Key("namespace").String(namespace),
		attribute.Key("pod").String(pod),
		attribute.Key("container").String(container),
		attribute.Key("file").String(fpath),
	)

	var (
		errbf, outbf      = bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
		reader, outStream = io.Pipe()
		uploader          = repo.uploader
		localUploader     = repo.uploader.LocalUploader()
	)
	if targetContainerDir == "" {
		targetContainerDir = "/tmp"
	}
	st, err := uploader.Stat(fpath)
	if err != nil {
		return nil, err
	}
	if st.Size() > repo.maxUploadSize {
		// 上传超限是显式构造的校验失败，用语义构造器映射为 InvalidArgument(400)，
		// 上层 errs.Wrap 保留该状态码，避免客户端把"文件太大"误判成服务器内部错误。
		return nil, errs.WrapInvalidArgument(fmt.Errorf("最大不得超过 %s, 你上传的文件大小是 %s", humanize.Bytes(repo.maxUploadSize), humanize.Bytes(uint64(st.Size()))), "copy file to pod")
	}

	baseName := filepath.Base(fpath)
	path := filepath.Join(filepath.Dir(fpath), baseName+".tar.gz")
	repo.logger.Debugf("[CopyFileToPod]: %v", path)
	var localPath = fpath
	// 如果是非 local 类型的，需要远程下载到 local 进行打包，再上传到容器
	if uploader.Type() != schematype.Local {
		read, err := uploader.Read(fpath)
		if err != nil {
			return nil, err
		}
		defer read.Close()
		if localUploader.Exists(localPath) {
			localUploader.Delete(localPath)
		}
		put, err := localUploader.Put(localPath, read)
		if err != nil {
			return nil, err
		}
		localPath = put.Path()
		defer localUploader.Delete(localPath)
	}
	if err := repo.archiver.Archive([]string{localPath}, path); err != nil {
		return nil, err
	}
	defer repo.archiver.Remove(path)
	src, err := repo.archiver.Open(path)
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	go func(reader *io.PipeReader, outStream *io.PipeWriter, src io.ReadCloser) {
		defer func() {
			reader.Close()
			outStream.Close()
			src.Close()
			wg.Done()
		}()
		defer repo.logger.HandlePanic("CopyFileToPod")

		if _, err := io.Copy(outStream, src); err != nil {
			repo.logger.Errorf("CopyFileToPod: io.Copy to container failed: %v", err)
		}
	}(reader, outStream, src)

	err = repo.executor.
		New().
		WithCommand([]string{"tar", "-zmxf", "-", "-C", targetContainerDir}).
		WithMethod("POST").
		WithContainer(namespace, pod, container).
		Execute(
			ctx,
			&biz.ExecuteInput{
				Stdin:  reader,
				Stdout: outbf,
				Stderr: errbf,
				TTY:    false,
			},
		)

	return &copyToPodResult{
		TargetDir:     targetContainerDir,
		ErrOut:        errbf.String(),
		StdOut:        outbf.String(),
		ContainerPath: filepath.Join(targetContainerDir, baseName),
		FileName:      baseName,
	}, err
}

// CopyFileToPod 把已登记的 file 复制进 Pod 容器，并回写容器内路径到该文件记录。
func (repo *k8sRepo) CopyFileToPod(ctx context.Context, input *biz.CopyFileToPodInput) (f *biz.File, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/CopyFileToPod")
	defer func() { endSpan(span, err) }()
	file, err := repo.fileRepo.GetByID(ctx, int(input.FileId))
	if err != nil {
		return nil, err
	}

	result, err := repo.copyToPod(ctx, input.Namespace, input.Pod, input.Container, file.Path, "")
	if err != nil {
		return nil, errs.Wrap(err, "copy file to pod")
	}

	return repo.fileRepo.Update(ctx, &biz.UpdateFileRequest{
		ID:            int(input.FileId),
		ContainerPath: result.ContainerPath,
		Namespace:     input.Namespace,
		Pod:           input.Pod,
		Container:     input.Container,
	})
}

// ClusterInfo 统计集群资源：节点调度能力/实际用量/请求量，换算 CPU/内存余量
// 与使用率，并按请求率阈值给出健康状态。节点或指标 List 失败只记日志降级返回。
func (repo *k8sRepo) ClusterInfo() *biz.ClusterInfo {
	selector := labels.Everything()
	var nodes []corev1.Node

	// 获取已经使用的 cpu, memory
	nodeList, err := repo.data.K8s().Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		// List 失败时 nodeList 为 nil，直接访问 .Items 会 panic，
		// 记录错误后按空集群返回，避免集群信息接口直接崩溃。
		repo.logger.Error(err)
		return &biz.ClusterInfo{}
	}
	nodes = append(nodes, nodeList.Items...)
	allocatable := make(map[string]corev1.ResourceList)

	var (
		totalCpu    = &resource.Quantity{}
		totalMemory = &resource.Quantity{}
	)

	var (
		workerNodes  []corev1.Node
		notWorkNodes []corev1.Node
	)

	for _, node := range nodes {
		notWork := false
		for _, taint := range node.Spec.Taints {
			if taint.Effect == corev1.TaintEffectNoExecute || taint.Effect == corev1.TaintEffectNoSchedule {
				notWork = true
				break
			}
		}
		if !notWork {
			workerNodes = append(workerNodes, node)
		} else {
			notWorkNodes = append(notWorkNodes, node)
		}
	}

	for _, n := range workerNodes {
		allocatable[n.Name] = n.Status.Allocatable
		totalCpu.Add(n.Status.Allocatable.Cpu().DeepCopy())
		totalMemory.Add(n.Status.Allocatable.Memory().DeepCopy())
	}

	requestCpu, requestMemory := repo.getNodeRequestCpuAndMemory(notWorkNodes)
	var (
		usedCpu    = &resource.Quantity{}
		usedMemory = &resource.Quantity{}
	)

	list, err := repo.data.K8s().MetricsClient.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		// metrics List 失败时 list 为 nil，range list.Items 会 panic，
		// 记录错误后用空列表继续（节点信息仍有效，仅用量统计为空）。
		repo.logger.Error(err)
		list = &v1beta1.NodeMetricsList{}
	}

	IsStatisticalNode := func(workerNodes []corev1.Node, name string) bool {
		for _, node := range workerNodes {
			if node.Name == name {
				return true
			}
		}
		return false
	}

	for _, item := range list.Items {
		if !IsStatisticalNode(workerNodes, item.Name) {
			continue
		}
		usedCpu.Add(item.Usage.Cpu().DeepCopy())
		usedMemory.Add(item.Usage.Memory().DeepCopy())
	}

	freeMemory := totalMemory.DeepCopy()
	freeMemory.Sub(*usedMemory)
	freeCpu := totalCpu.DeepCopy()
	freeCpu.Sub(*usedCpu)

	freeRequestMemory := totalMemory.DeepCopy()
	freeRequestMemory.Sub(*requestMemory)
	freeRequestCpu := totalCpu.DeepCopy()
	freeRequestCpu.Sub(*requestCpu)

	rateMemory := float64(usedMemory.Value()) / float64(totalMemory.Value()) * 100
	rateCpu := float64(usedCpu.Value()) / float64(totalCpu.Value()) * 100
	rateRequestMemory := float64(requestMemory.Value()) / float64(totalMemory.Value()) * 100
	rateRequestCpu := float64(requestCpu.Value()) / float64(totalCpu.Value()) * 100

	return &biz.ClusterInfo{
		Status:            repo.getStatus(rateRequestMemory, rateRequestCpu),
		FreeRequestMemory: humanize.IBytes(uint64(freeRequestMemory.Value())),
		FreeRequestCpu:    fmt.Sprintf("%.2f core", float64(freeRequestCpu.MilliValue())/1000),
		FreeMemory:        humanize.IBytes(uint64(freeMemory.Value())),
		FreeCpu:           fmt.Sprintf("%.2f core", float64(freeCpu.MilliValue())/1000),
		TotalMemory:       humanize.IBytes(uint64(totalMemory.Value())),
		TotalCpu:          fmt.Sprintf("%.2f core", float64(totalCpu.MilliValue())/1000),
		UsageMemoryRate:   fmt.Sprintf("%.1f%%", rateMemory),
		UsageCpuRate:      fmt.Sprintf("%.1f%%", rateCpu),
		RequestCpuRate:    fmt.Sprintf("%.1f%%", rateRequestCpu),
		RequestMemoryRate: fmt.Sprintf("%.1f%%", rateRequestMemory),
	}
}

// getStatus 按请求率阈值定级：任一超 80 为 bad，超 60 为 not good，否则 health。
func (repo *k8sRepo) getStatus(rateRequestMemory float64, rateRequestCpu float64) biz.ClusterStatus {
	var status = StatusHealth
	if rateRequestMemory > 60 || rateRequestCpu > 60 {
		status = StatusNotGood
	}
	if rateRequestMemory > 80 || rateRequestCpu > 80 {
		status = StatusBad
	}
	return status
}

// getNodeRequestCpuAndMemory 统计全部 Running Pod 的 CPU/内存 request，排除落在
// 不可调度节点（污点节点）上的 Pod，得到真正占用调度资源的那部分。
func (repo *k8sRepo) getNodeRequestCpuAndMemory(noExecuteNodes []corev1.Node) (*resource.Quantity, *resource.Quantity) {
	var (
		requestCpu    = &resource.Quantity{}
		requestMemory = &resource.Quantity{}
	)

	var nodeSelector = []string{
		"status.phase==" + string(corev1.PodRunning),
	}
	for _, node := range noExecuteNodes {
		nodeSelector = append(nodeSelector, "spec.nodeName!="+node.Name)
	}
	fieldSelector, _ := fields.ParseSelector(strings.Join(nodeSelector, ","))
	nodeNonTerminatedPodsList, _ := repo.data.K8s().
		Client.
		CoreV1().
		Pods("").
		List(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector.String()})
	for _, item := range nodeNonTerminatedPodsList.Items {
		for _, container := range item.Spec.Containers {
			requestCpu.Add(container.Resources.Requests.Cpu().DeepCopy())
			requestMemory.Add(container.Resources.Requests.Memory().DeepCopy())
		}
	}

	return requestCpu, requestMemory
}

// LogStream 建立跟随日志流：Follow 模式流式读取并按行投递到缓冲 channel，
// ctx 取消时关闭流与 channel。读满缓冲时丢行并打 Debug 日志，避免阻塞读循环。
func (repo *k8sRepo) LogStream(
	ctx context.Context,
	namespace,
	pod,
	container string,
) (out chan []byte, err error) {
	ctx, span := tracer.Start(ctx, "k8sRepo/LogStream")
	defer func() { endSpan(span, err) }()
	tailLines := 1000
	logs := repo.data.K8s().Client.
		CoreV1().
		Pods(namespace).
		GetLogs(pod, &corev1.PodLogOptions{
			Follow:    true,
			Container: container,
			TailLines: lo.ToPtr(int64(tailLines)),
		})

	stream, err := logs.Stream(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "stream container log")
	}

	bf := bufio.NewReader(stream)
	ctx, cancelFunc := context.WithCancel(ctx)
	ch := make(chan []byte, tailLines)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			close(ch)
			cancelFunc()
			wg.Done()
		}()
		defer repo.logger.HandlePanic("StreamContainerLog")

		for {
			bytes, err := bf.ReadBytes('\n')
			if err != nil {
				repo.logger.DebugCtxf(ctx, "[LogStream]: %v", err)
				return
			}
			select {
			case ch <- bytes:
			default:
				repo.logger.DebugCtx(ctx, "[LogStream]:  drop line!")
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		stream.Close()
		repo.logger.DebugCtx(ctx, "[LogStream]:  ctx done!")
	}()
	go func() {
		wg.Wait()
		repo.logger.DebugCtx(ctx, "[LogStream]:  exit!")
	}()

	return ch, nil
}

// ListPodsBySelectors 按一组 label selector 列出命名空间内匹配的 Pod。
// selector 解析与 informer 过滤属于基础设施细节，收敛在 data 侧；单个 selector
// 解析/查询失败只记录日志并跳过，不影响其余 selector 的结果。
//
// 注意：LabelSelectorAsSelector 的返回值忽略。入参 selector 一律来自
// metav1.ParseToLabelSelector，其词法/语义校验保证 operator/key/value 全部合法，
// 因此 AsSelector 分支必然成功（已用探针验证多组输入），无错误分支可走。
func (repo *k8sRepo) ListPodsBySelectors(namespace string, selectors []string) ([]*corev1.Pod, error) {
	var res []*corev1.Pod
	seen := map[string]struct{}{}
	for _, ls := range selectors {
		selector, err := metav1.ParseToLabelSelector(ls)
		if err != nil {
			repo.logger.Debugf("[ListPodsBySelectors]: parse selector %q: %v", ls, err)
			continue
		}
		asSelector, _ := metav1.LabelSelectorAsSelector(selector)
		l, err := repo.data.K8s().PodLister.Pods(namespace).List(asSelector)
		if err != nil {
			repo.logger.Debugf("[ListPodsBySelectors]: list pods %q: %v", ls, err)
			continue
		}
		for _, pod := range l {
			if _, ok := seen[pod.Name]; ok {
				continue
			}
			seen[pod.Name] = struct{}{}
			res = append(res, pod)
		}
	}
	return res, nil
}

// GetReplicaSet 经 informer lister 读取 ReplicaSet。
func (repo *k8sRepo) GetReplicaSet(namespace, name string) (*appsv1.ReplicaSet, error) {
	rs, err := repo.data.K8s().ReplicaSetLister.ReplicaSets(namespace).Get(name)
	if err != nil {
		repo.logger.Debugf("[GetReplicaSet]: %v", err)
	}
	return rs, errs.Wrap(err, "get replica set")
}

// ListServices 列出命名空间下全部 Service。
func (repo *k8sRepo) ListServices(namespace string) ([]*corev1.Service, error) {
	services, err := repo.data.K8s().ServiceLister.Services(namespace).List(labels.Everything())
	return services, errs.Wrap(err, "list services")
}

// ListIngresses 列出命名空间下全部 Ingress。
func (repo *k8sRepo) ListIngresses(namespace string) ([]*networkingv1.Ingress, error) {
	ingresses, err := repo.data.K8s().IngressLister.Ingresses(namespace).List(labels.Everything())
	return ingresses, errs.Wrap(err, "list ingresses")
}

// ListHTTPRoutes 列出命名空间下全部 Gateway API HTTPRoute。
func (repo *k8sRepo) ListHTTPRoutes(namespace string) ([]*gatewayv1.HTTPRoute, error) {
	routes, err := repo.data.K8s().HTTPRouteLister.HTTPRoutes(namespace).List(labels.Everything())
	return routes, errs.Wrap(err, "list http routes")
}

// GatewayApiInstalled 返回集群是否安装了 Gateway API（bootstrap 探测结果）。
func (repo *k8sRepo) GatewayApiInstalled() bool {
	return repo.data.K8s().GatewayApiInstalled
}

// ExternalIp 返回配置的对外 IP。
func (repo *k8sRepo) ExternalIp() string {
	return repo.data.Config().ExternalIp
}

// Archiver 抽象文件归档/解压操作：归档多个源路径、打开归档文件、删除归档。
// 具体实现委托第三方 archiver 库与 os 包，接口化便于测试替换与 k8sutil 复用。
type Archiver interface {
	// Archive 将多个源文件/目录归档到目标路径。
	Archive(sources []string, destination string) error
	// Open 打开归档文件用于读取。
	Open(path string) (io.ReadCloser, error)
	// Remove 删除归档文件。
	Remove(path string) error
}

// defaultArchiver 是 Archiver 的默认实现，直接透传 archiver.Archive 与 os 操作。
type defaultArchiver struct{}

// NewDefaultArchiver 构造默认 Archiver 实现。
func NewDefaultArchiver() Archiver {
	return &defaultArchiver{}
}

// Archive 将多个源文件/目录归档到目标路径（委托 archiver.Archive）。
func (m *defaultArchiver) Archive(sources []string, destination string) error {
	return archiver.Archive(sources, destination)
}

// Open 打开归档文件用于读取。
func (m *defaultArchiver) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// Remove 删除归档文件。
func (m *defaultArchiver) Remove(path string) error {
	return os.Remove(path)
}

// ExecutorManager 提供 Pod 远程执行相关对象的工厂：裸 Executor 与文件拷贝器。
type ExecutorManager interface {
	// New 构造裸 Executor（命令执行构建器）。
	New() Executor
	// NewFileCopy 构造从 Pod 复制文件的拷贝器（kubectl tar 流）。
	NewFileCopy(maxTries int, errOut io.Writer) k8sutil.FileCopy
}

// Executor 描述在指定 Pod 容器内执行命令的可配置构建器：
// 链式设置方法/容器/命令后调用 Execute 发起 SPDY exec 流。
type Executor interface {
	// WithMethod 设置执行方法（exec/attach 等）。
	WithMethod(method string) Executor
	// WithContainer 设置目标命名空间/Pod/容器。
	WithContainer(namespace, pod, container string) Executor
	// WithCommand 设置要执行的命令。
	WithCommand(cmd []string) Executor
	// Execute 发起 SPDY exec 流执行命令。
	Execute(context.Context, *biz.ExecuteInput) error
}

// defaultRemoteExecutor 是 ExecutorManager 的默认实现，从 dataStore 提取
// K8s client 与 rest config 构造执行器。
type defaultRemoteExecutor struct {
	data   dataStore
	logger mlog.Logger
}

// NewExecutorManager 构造默认 ExecutorManager 实现。
func NewExecutorManager(data dataStore, logger mlog.Logger) ExecutorManager {
	return &defaultRemoteExecutor{
		data:   data,
		logger: logger.WithModule("ExecutorManager"),
	}
}

// NewFileCopy 基于集群 rest config 构造文件拷贝器（k8sutil.FileCopy）。
// 拷贝器需要 v1 core API 路径与协商编解码器，故在构造时改写 config 的
// APIPath/GroupVersion/NegotiatedSerializer 并复用集群 client。
// 改写必须在 DeepCopy 副本上进行：原 config 与 New/Execute 等执行器共享，
// 原地改写会造成并发数据竞争并永久污染后续 executor 的请求路径。
func (e *defaultRemoteExecutor) NewFileCopy(
	maxTries int,
	errOut io.Writer,
) k8sutil.FileCopy {
	// 用 rest.CopyConfig 而非 .DeepCopy()：Config 匿名内嵌 TLSClientConfig，
	// 后者的 DeepCopy 会被方法提升返回 *TLSClientConfig，只拷贝 TLS 部分，
	// 达不到隔离共享 config 的目的。
	restCfg := restclient.CopyConfig(e.data.K8s().RestConfig)
	restCfg.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	restCfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	restCfg.APIPath = "/api"

	return k8sutil.NewCopyOptions(
		e.logger,
		restCfg,
		e.data.K8s().Client,
		maxTries,
		errOut,
	)
}

// New 构造裸 Executor，绑定集群 client 与 rest config，目标容器在后续链式调用中设置。
func (e *defaultRemoteExecutor) New() Executor {
	return &executor{
		clientSet: e.data.K8s().Client,
		config:    e.data.K8s().RestConfig,
	}
}

// executor 是 Executor 的实现：保存目标 namespace/pod/container、请求方法、
// 命令与集群连接信息，Execute 时拼装 SPDY exec 请求。
type executor struct {
	namespace, pod, container string
	method                    string
	cmd                       []string

	clientSet kubernetes.Interface
	config    *restclient.Config
}

// WithMethod 设置 SPDY 请求方法（如 POST）。
func (e *executor) WithMethod(method string) Executor {
	e.method = method
	return e
}

// WithContainer 设置目标 Pod 与容器。
func (e *executor) WithContainer(namespace, pod, container string) Executor {
	e.namespace = namespace
	e.pod = pod
	e.container = container
	return e
}

// newSPDYExecutor 是 remotecommand.NewSPDYExecutor 的构造缝：测试可注入 fake
// executor，以覆盖 Execute 成功路径的 StreamWithContext 返回 nil（真实握手属
// 集成边界，单测不连集群）。
var newSPDYExecutor = remotecommand.NewSPDYExecutor

// WithCommand 设置在容器内执行的命令。
func (e *executor) WithCommand(cmd []string) Executor {
	e.cmd = cmd
	return e
}

// Execute 向目标 Pod 发起 exec 请求并流式交换标准输入/输出/错误流。
// 错误直接透传 SPDY 握手与流处理失败，不做额外包装；容器退出码错误在出口
// 经 translateExecError 翻译为领域错误，隔离 client-go 类型。
func (e *executor) Execute(ctx context.Context, input *biz.ExecuteInput) error {
	var (
		terminalSizeQueue = toRemotecommandTerminalSizeQueue(input.TerminalSizeQueue)
		stdin             = input.Stdin
		stdout            = input.Stdout
		stderr            = input.Stderr
		tty               = input.TTY
	)
	peo := e.newOption(stdin, stdout, stderr, tty)

	req := e.clientSet.CoreV1().
		RESTClient().
		Post().
		Namespace(e.namespace).
		Resource("pods").
		SubResource("exec").
		Name(e.pod)

	exec, err := newSPDYExecutor(e.config, e.method, req.VersionedParams(peo, scheme.ParameterCodec).URL())
	if err != nil {
		return err
	}

	if err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               tty,
		TerminalSizeQueue: terminalSizeQueue,
	}); err != nil {
		return translateExecError(err)
	}
	return nil
}

// newOption 依据传入流与 TTY 标志构造 PodExecOptions，作为 exec 请求的版本化参数。
// 流是否为 nil 决定对应布尔开关（Stdin/Stdout/Stderr）与真实流注入。
func (e *executor) newOption(stdin io.Reader, stdout io.Writer, stderr io.Writer, tty bool) *corev1.PodExecOptions {
	return &corev1.PodExecOptions{
		Stdin:     stdin != nil,
		Stdout:    stdout != nil,
		Stderr:    stderr != nil,
		TTY:       tty,
		Container: e.container,
		Command:   e.cmd,
	}
}

// terminalSizeQueueAdapter 把 biz 领域 TerminalSizeQueue 适配为 client-go 的
// remotecommand.TerminalSizeQueue，使 SPDY executor 无需感知领域尺寸类型。
type terminalSizeQueueAdapter struct {
	queue biz.TerminalSizeQueue
}

// Next 转换并透传下一次终端尺寸；领域队列返回 nil（流结束）时返回 nil。
func (a *terminalSizeQueueAdapter) Next() *remotecommand.TerminalSize {
	s := a.queue.Next()
	if s == nil {
		return nil
	}
	return &remotecommand.TerminalSize{Width: s.Width, Height: s.Height}
}

// toRemotecommandTerminalSizeQueue 把领域尺寸队列适配为 remotecommand 队列；nil 输入返回 nil。
func toRemotecommandTerminalSizeQueue(q biz.TerminalSizeQueue) remotecommand.TerminalSizeQueue {
	if q == nil {
		return nil
	}
	return &terminalSizeQueueAdapter{queue: q}
}

// translateExecError 把 client-go 的容器退出码错误（CodeExitError）翻译为 biz 领域错误
// ExecExitError，使 biz 层不依赖 client-go 错误类型；其余错误原样透传。
func translateExecError(err error) error {
	if err == nil {
		return nil
	}
	var exitErr clientgoexec.ExitError
	if errors.As(err, &exitErr) {
		return &biz.ExecExitError{Code: exitErr.ExitStatus(), Message: exitErr.Error()}
	}
	return err
}
