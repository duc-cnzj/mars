package data

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	v12 "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
)

func init() {
	kube.New(nil)
}

var _ biz.HelmerRepo = (*DefaultHelmer)(nil)

// DefaultHelmer 是 biz.HelmerRepo 的实现：封装 helm 部署/回滚/卸载/打包/状态查询，
// 依赖 k8sRepo 与 dataStore 取集群上下文。
type DefaultHelmer struct {
	Debug      bool
	DockerAuth config.DockerAuths
	KubeConfig string
	k8sRepo    biz.K8sRepo
	logger     mlog.Logger
	data       dataStore
}

// NewDefaultHelmer 构造 helm 仓库实现，注入 k8s repo、dataStore、配置与日志。
func NewDefaultHelmer(k8sRepo biz.K8sRepo, data dataStore, cfg *config.Config, logger mlog.Logger) biz.HelmerRepo {
	return &DefaultHelmer{
		k8sRepo:    k8sRepo,
		Debug:      cfg.Debug,
		logger:     logger.WithModule("repo/helmer"),
		DockerAuth: cfg.ImagePullSecrets,
		KubeConfig: cfg.KubeConfig,
		data:       data,
	}
}

// UpgradeOrInstall 安装或升级 chart：wait 场景先干跑一次取 Pod 选择器，
// 再正式执行并附带 Pod/事件监听日志。
func (d *DefaultHelmer) UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn biz.WrapLogFn, wait bool, timeoutSeconds int64, dryRun bool, desc string) (out *release.Release, err error) {
	ctx, span := tracer.Start(ctx, "DefaultHelmer/UpgradeOrInstall")
	defer func() { endSpan(span, err) }()
	var podSelectors []string
	if wait && !dryRun {
		re, err := d.upgradeOrInstall(ctx, releaseName, namespace, ch, valueOpts, fn, false, timeoutSeconds, true, nil, desc, d.KubeConfig, d.data.K8s())
		if err != nil {
			return nil, errs.Wrap(err, "upgrade or install")
		}

		podSelectors = d.k8sRepo.GetPodSelectorsByManifest(d.k8sRepo.SplitManifests(re.Manifest))
	}

	rel, err := d.upgradeOrInstall(ctx, releaseName, namespace, ch, valueOpts, fn, wait, timeoutSeconds, dryRun, podSelectors, desc, d.KubeConfig, d.data.K8s())
	return rel, errs.Wrap(err, "upgrade or install")
}

// Rollback 回滚 release 到上一版本。
func (d *DefaultHelmer) Rollback(releaseName, namespace string, wait bool, log biz.LogFn, dryRun bool) error {
	return errs.Wrap(rollback(releaseName, namespace, wait, log, dryRun, d.KubeConfig), "rollback release")
}

// Uninstall 卸载指定 release。
func (d *DefaultHelmer) Uninstall(releaseName, namespace string, log biz.LogFn) error {
	return errs.Wrap(uninstallRelease(releaseName, namespace, log, d.KubeConfig), "uninstall release")
}

// PackageChart 把本地 chart 目录打包为 tgz 到 destDir。
func (d *DefaultHelmer) PackageChart(path string, destDir string) (string, error) {
	p, err := packageChart(path, destDir, d.DockerAuth, d.Debug)
	return p, errs.Wrap(err, "package chart")
}

// ReleaseStatus 查询 release 的部署状态（unknown/pending/deployed/failed）。
func (d *DefaultHelmer) ReleaseStatus(releaseName, namespace string) types.Deploy {
	return d.releaseStatus(releaseName, namespace, d.KubeConfig)
}

// upgradeOrInstall
// 不会自动回滚
func (d *DefaultHelmer) upgradeOrInstall(
	ctx context.Context,
	releaseName, namespace string,
	ch *chart.Chart,
	valueOpts *values.Options,
	fn biz.WrapLogFn,
	wait bool,
	timeoutSeconds int64,
	dryRun bool,
	podSelectors []string,
	desc string,
	kubeconfig string,
	k8sClient *K8sClient,
) (*release.Release, error) {
	actionConfig := newActionConfig(namespace, kubeconfig, fn.UnWrap())
	client := action.NewUpgrade(actionConfig)
	client.Install = true
	client.Atomic = false
	client.Wait = wait
	client.Description = desc
	client.DryRun = dryRun
	client.DependencyUpdate = true
	client.DisableOpenAPIValidation = true
	client.MaxHistory = 5

	if wait && !dryRun {
		var selectorList []labels.Selector
		for _, label := range podSelectors {
			selector, _ := metav1.ParseToLabelSelector(label)
			asSelector, _ := metav1.LabelSelectorAsSelector(selector)
			selectorList = append(selectorList, asSelector)
		}
		fanOutCtx, cancelFn := context.WithCancel(ctx)
		key := fmt.Sprintf("%s-%s", namespace, releaseName)
		podCh := make(chan Obj[*corev1.Pod], 200)
		evCh := make(chan Obj[*eventsv1.Event], 200)
		defer func() {
			cancelFn()
			k8sClient.podFanOut.RemoveListener(key)
			k8sClient.eventFanOut.RemoveListener(key)
		}()
		k8sClient.podFanOut.AddListener(key, podCh)
		k8sClient.eventFanOut.AddListener(key, evCh)
		go func() {
			defer d.logger.HandlePanic("upgradeOrInstall pod-fan-out")
			d.watchPodStatus(fanOutCtx, podCh, selectorList, fn)
		}()
		go func() {
			defer d.logger.HandlePanic("upgradeOrInstall event-fan-out")
			d.watchEvent(fanOutCtx, evCh, releaseName, fn, k8sClient.PodLister)
		}()
	}

	if timeoutSeconds != 0 {
		client.Timeout = time.Duration(timeoutSeconds) * time.Second
	} else {
		client.Timeout = 5 * 60 * time.Second
	}

	if valueOpts == nil {
		valueOpts = &values.Options{}
	}

	client.Namespace = namespace
	if client.Install {
		// If a release does not exist, install it.
		histClient := action.NewHistory(actionConfig)
		histClient.Max = 1
		if _, err := histClient.Run(releaseName); err == driver.ErrReleaseNotFound {
			instClient := action.NewInstall(actionConfig)
			fillInstall(instClient, client)
			d.logger.Debug("start install release", valueOpts)
			return d.runInstall(ctx, releaseName, ch, instClient, valueOpts)
		}
	}

	vals, err := valueOpts.MergeValues(getter.All(&cli.EnvSettings{PluginsDirectory: ""}))
	if err != nil {
		return nil, err
	}

	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			return nil, err
		}
	}

	return client.RunWithContext(ctx, releaseName, ch, vals)
}

// fillInstall 把 Upgrade 客户端的公共参数拷贝到 Install 客户端（首次安装复用升级配置）。
func fillInstall(instClient *action.Install, client *action.Upgrade) {
	instClient.CreateNamespace = true
	instClient.ChartPathOptions = client.ChartPathOptions
	instClient.DryRun = client.DryRun
	instClient.DisableHooks = client.DisableHooks
	instClient.SkipCRDs = client.SkipCRDs
	instClient.Timeout = client.Timeout
	instClient.Wait = client.Wait
	instClient.WaitForJobs = client.WaitForJobs
	instClient.Devel = client.Devel
	instClient.Namespace = client.Namespace
	instClient.Atomic = client.Atomic
	instClient.PostRenderer = client.PostRenderer
	instClient.DisableOpenAPIValidation = client.DisableOpenAPIValidation
	instClient.SubNotes = client.SubNotes
	instClient.Description = client.Description
	instClient.DependencyUpdate = client.DependencyUpdate
}

// watchPodStatus 消费 Pod 扇出事件：对命中选择器且多次异常重启的容器回调日志函数。
func (d *DefaultHelmer) watchPodStatus(ctx context.Context, podCh chan Obj[*corev1.Pod], selectorList []labels.Selector, fn biz.WrapLogFn) {
	for {
		select {
		case <-ctx.Done():
			d.logger.Debug("ctx.Done pod")
			return
		case obj, ok := <-podCh:
			if !ok {
				return
			}
			if obj.Type() == Update {
				continue
			}
			p := obj.Current()
			var matched bool
			for _, selector := range selectorList {
				if selector.Matches(labels.Set(p.Labels)) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}

			var (
				containerNames []string
				containers     []*websocket_pb.Container
			)
			for _, status := range p.Status.ContainerStatuses {
				if !status.Ready && status.RestartCount > 0 {
					containerNames = append(containerNames, status.Name)
				}
			}
			for _, name := range containerNames {
				//查看日志
				containers = append(containers, &websocket_pb.Container{
					Namespace: p.Namespace,
					Pod:       p.Name,
					Container: name,
				})
			}
			if len(containers) > 0 {
				fn(containers, "容器多次异常重启")
			}
		}
	}
}

// watchEvent 消费 Event 扇出事件：对命中 release 标签的 Pod 事件回调日志函数。
func (d *DefaultHelmer) watchEvent(ctx context.Context, evCh chan Obj[*eventsv1.Event], releaseName string, fn biz.WrapLogFn, lister v12.PodLister) {
	for {
		select {
		case <-ctx.Done():
			d.logger.Debug("ctx.Done event")
			return
		case evobj, ok := <-evCh:
			if !ok {
				return
			}
			if evobj.Type() != Add {
				continue
			}

			var obj any = evobj.Current()
			event := obj.(*eventsv1.Event)
			p := event.Regarding
			get, err := lister.Pods(p.Namespace).Get(p.Name)
			if err != nil {
				d.logger.Warningf("can't get pod ns: '%s', name: '%s'", p.Namespace, p.Name)
				continue
			}

			for _, value := range get.Labels {
				if value == releaseName {
					fn(nil, event.Note)
					break
				}
			}
		}
	}
}

// uninstallRelease 经 helm action 卸载指定 release。
func uninstallRelease(releaseName, namespace string, log biz.LogFn, kubeconfig string) error {
	actionConfig := newActionConfig(namespace, kubeconfig, log)
	uninstall := action.NewUninstall(actionConfig)
	_, err := uninstall.Run(releaseName)
	return err
}

// runInstall 执行首次安装：合并 values 并校验 chart 可安装性后运行。
func (d *DefaultHelmer) runInstall(ctx context.Context, releaseName string, chartRequested *chart.Chart, client *action.Install, valueOpts *values.Options) (*release.Release, error) {
	d.logger.Debugf("Original chart version: %q", client.Version)
	if client.Version == "" && client.Devel {
		d.logger.Debug("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	client.ReleaseName = releaseName

	vals, err := valueOpts.MergeValues(getter.All(&cli.EnvSettings{PluginsDirectory: ""}))
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	if err := checkIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	return client.RunWithContext(ctx, chartRequested, vals)
}

// checkIfInstallable 校验 chart 类型可安装（空或 application 才允许）。
// chart 类型非法是显式构造的校验失败，用语义构造器映射为 InvalidArgument(400)，
// 上层 errs.Wrap 保留该状态码，避免客户端把"chart 不可安装"误判成服务器内部错误。
func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errs.WrapInvalidArgument(fmt.Errorf("%s charts are not installable", ch.Metadata.Type), "check chart installable")
}

// Status* 是 release 状态字符串常量（与 helm release.Info.Status 对齐）。
const (
	StatusUnknown  string = "unknown"
	StatusPending  string = "pending"
	StatusDeployed string = "deployed"
	StatusFailed   string = "failed"
)

// tokenFile/rootCAFile 是集群内 service account 的 token 与 CA 证书挂载路径，
// 供无 kubeconfig 时走集群内认证连接。
const (
	tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token" // #nosec G101
	rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

// releaseStatus 查询 release 状态并映射为 types.Deploy 枚举。
func (d *DefaultHelmer) releaseStatus(releaseName, namespace, kubeconfig string) types.Deploy {
	actionConfig := newActionConfig(namespace, kubeconfig, d.logger.Debugf)
	statusClient := action.NewStatus(actionConfig)
	run, err := statusClient.Run(releaseName)
	if err != nil {
		d.logger.Warning(err)
		return types.Deploy_StatusUnknown
	}

	d.logger.Debug(run.Info.Status)
	return formatStatus(run.Info.Status)
}

// formatStatus 把 helm release 状态映射为 types.Deploy 枚举。
func formatStatus(input release.Status) types.Deploy {
	switch input {
	case release.StatusPendingUpgrade, release.StatusPendingInstall, release.StatusPendingRollback:
		return types.Deploy_StatusDeploying
	case release.StatusDeployed:
		return types.Deploy_StatusDeployed
	case release.StatusFailed:
		return types.Deploy_StatusFailed
	default:
		return types.Deploy_StatusUnknown
	}
}

// dockerCfgOnce 保证 docker 配置写盘只执行一次（打包依赖拉取用）。
var dockerCfgOnce sync.Once

// dockerCfgOncePath 是临时 docker 配置文件落盘路径。
const dockerCfgOncePath = "/tmp/mars-docker-config.json"

// packageChart 打包本地 chart 目录；有依赖时先经 registry 拉取更新依赖。
func packageChart(path string, destDir string, imagePullSecrets config.DockerAuths, debug bool) (string, error) {
	newPackage := action.NewPackage()
	if destDir != "" {
		newPackage.Destination = destDir
	}

	chartLocal, err := loader.LoadDir(path)
	if err != nil {
		return "", err
	}
	if chartLocal.Metadata.Dependencies != nil && action.CheckDependencies(chartLocal, chartLocal.Metadata.Dependencies) != nil {
		// 更新依赖 dependency, 防止没有依赖文件打包失败
		dockerCfgOnce.Do(func() {
			os.WriteFile(dockerCfgOncePath, imagePullSecrets.FormatDockerCfg(), 0600)
		})
		client, err := newDefaultRegistryClient(debug, dockerCfgOncePath)
		if err != nil {
			return "", err
		}
		downloadManager := &downloader.Manager{
			Out:            io.Discard,
			ChartPath:      path,
			Debug:          debug,
			Keyring:        newPackage.Keyring,
			Getters:        getter.All(&cli.EnvSettings{PluginsDirectory: ""}),
			RegistryClient: client,
		}

		if err := downloadManager.Update(); err != nil {
			return "", err
		}
	}

	return newPackage.Run(path, nil)
}

// logWriter 是丢弃日志的 io.Writer（registry 客户端调试输出丢弃）。
type logWriter struct{}

// Write 吞掉所有日志字节并返回成功。
func (l *logWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// newDefaultRegistryClient 构造带调试/缓存/凭据文件的 helm registry 客户端。
func newDefaultRegistryClient(debug bool, cfg string) (*registry.Client, error) {
	opts := []registry.ClientOption{
		registry.ClientOptDebug(debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(&logWriter{}),
		registry.ClientOptCredentialsFile(cfg),
	}

	registryClient, err := registry.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

// newActionConfig 是 helm action 配置构建函数，默认委托 getActionConfigAndSettings
// 连接真实集群；测试可替换为内存存储 + 假 kube 客户端的配置，以覆盖集群操作的成功分支。
var newActionConfig = getActionConfigAndSettings

// getActionConfigAndSettings 构造 helm action 配置：走 kubeconfig 或集群内
// service account token（KUBERNETES_SERVICE_HOST）两种连接方式。
func getActionConfigAndSettings(namespace string, kubeconfig string, log func(format string, v ...any)) *action.Configuration {
	actionConfig := new(action.Configuration)
	flags := genericclioptions.NewConfigFlags(true)
	flags = flags.WithDiscoveryQPS(-1)
	flags = flags.WithWrapConfigFn(wrapRestConfig)
	set := pflag.NewFlagSet("", pflag.ContinueOnError)
	flags.AddFlags(set)
	sets := []string{"--namespace=" + namespace}
	if kubeconfig != "" {
		sets = append(sets, "--kubeconfig="+kubeconfig)
	} else {
		host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
		server := "https://" + net.JoinHostPort(host, port)
		token, _ := os.ReadFile(tokenFile)
		sets = append(sets, "--server="+server, "--token="+string(token), "--certificate-authority="+rootCAFile)
	}

	set.Parse(sets)
	actionConfig.Init(flags, namespace, "", log)

	return actionConfig
}

// wrapRestConfig 关闭 rest.Config 的 QPS 限流（-1 表示不限）。
func wrapRestConfig(config *restclient.Config) *restclient.Config {
	config.QPS = -1
	return config
}

// rollback 执行 helm rollback：等待与 dry-run 由参数控制。
func rollback(releaseName, namespace string, wait bool, log biz.LogFn, dryRun bool, kubeconfig string) error {
	actionConfig := newActionConfig(namespace, kubeconfig, log)
	client := action.NewRollback(actionConfig)
	client.Wait = wait
	client.DryRun = dryRun
	client.DisableHooks = true
	client.WaitForJobs = wait

	return client.Run(releaseName)
}
