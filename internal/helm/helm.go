package helm

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/utils"
	"github.com/duc-cnzj/mars/v4/internal/utils/recovery"
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

type DefaultHelmer struct{}

func (d *DefaultHelmer) UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn contracts.WrapLogFn, wait bool, timeoutSeconds int64, dryRun bool, desc string) (*release.Release, error) {
	var podSelectors []string
	if wait && !dryRun {
		re, err := upgradeOrInstall(context.TODO(), releaseName, namespace, ch, valueOpts, func(container []*types.Container, format string, v ...any) {}, false, timeoutSeconds, true, nil, desc)
		if err != nil {
			return nil, err
		}

		podSelectors = utils.GetPodSelectorsByManifest(utils.SplitManifests(re.Manifest))
	}

	return upgradeOrInstall(ctx, releaseName, namespace, ch, valueOpts, fn, wait, timeoutSeconds, dryRun, podSelectors, desc)
}

func (d *DefaultHelmer) Rollback(releaseName, namespace string, wait bool, log contracts.LogFn, dryRun bool) error {
	return rollback(releaseName, namespace, wait, log, dryRun)
}

func (d *DefaultHelmer) PackageChart(path string, destDir string) (string, error) {
	return packageChart(path, destDir)
}

func (d *DefaultHelmer) Uninstall(releaseName, namespace string, log contracts.LogFn) error {
	return uninstallRelease(releaseName, namespace, log)
}

func (d *DefaultHelmer) ReleaseStatus(releaseName, namespace string) types.Deploy {
	return releaseStatus(releaseName, namespace)
}

// upgradeOrInstall
// 不会自动回滚
func upgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn contracts.WrapLogFn, wait bool, timeoutSeconds int64, dryRun bool, podSelectors []string, desc string) (*release.Release, error) {
	actionConfig := getActionConfigAndSettings(namespace, fn.UnWrap())
	client := action.NewUpgrade(actionConfig)
	client.Install = true
	client.Atomic = false
	client.Wait = wait
	client.Description = desc
	client.DryRun = dryRun
	client.DependencyUpdate = true
	client.DisableOpenAPIValidation = true

	if wait && !dryRun {
		var selectorList []labels.Selector
		for _, label := range podSelectors {
			selector, _ := metav1.ParseToLabelSelector(label)
			asSelector, _ := metav1.LabelSelectorAsSelector(selector)
			selectorList = append(selectorList, asSelector)
		}
		fanOutCtx, cancelFn := context.WithCancel(context.TODO())
		key := fmt.Sprintf("%s-%s", namespace, releaseName)
		k8sClient := app.K8sClient()
		podCh := make(chan contracts.Obj[*corev1.Pod], 100)
		evCh := make(chan contracts.Obj[*eventsv1.Event], 100)
		defer func() {
			cancelFn()
			k8sClient.PodFanOut.RemoveListener(key)
			k8sClient.EventFanOut.RemoveListener(key)
			close(podCh)
			close(evCh)
		}()
		k8sClient.PodFanOut.AddListener(key, podCh)
		k8sClient.EventFanOut.AddListener(key, evCh)
		go func() {
			defer recovery.HandlePanic("upgradeOrInstall pod-fan-out")
			watchPodStatus(fanOutCtx, podCh, selectorList, fn)
		}()
		go func() {
			defer recovery.HandlePanic("upgradeOrInstall event-fan-out")
			watchEvent(fanOutCtx, evCh, releaseName, fn, k8sClient.PodLister)
		}()
	}

	if timeoutSeconds != 0 {
		client.Timeout = time.Duration(timeoutSeconds) * time.Second
	} else if app.Config().InstallTimeout != 0 {
		client.Timeout = app.Config().InstallTimeout
	} else {
		client.Timeout = 5 * 60 * time.Second
	}

	if valueOpts == nil {
		valueOpts = &values.Options{}
	}

	client.Namespace = namespace
	if client.Version == "" && client.Devel {
		mlog.Debug("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}
	if client.Install {
		// If a release does not exist, install it.
		histClient := action.NewHistory(actionConfig)
		histClient.Max = 1
		if _, err := histClient.Run(releaseName); err == driver.ErrReleaseNotFound {
			instClient := action.NewInstall(actionConfig)
			fillInstall(instClient, client)
			mlog.Debug("start install release", valueOpts)
			return runInstall(ctx, releaseName, ch, instClient, valueOpts)
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

func watchPodStatus(ctx context.Context, podCh chan contracts.Obj[*corev1.Pod], selectorList []labels.Selector, fn contracts.WrapLogFn) {
	for {
		select {
		case <-ctx.Done():
			mlog.Debug("ctx.Done pod")
			return
		case obj, ok := <-podCh:
			if !ok {
				return
			}
			if obj.Type() == contracts.Update {
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
				containers     []*types.Container
			)
			for _, status := range p.Status.ContainerStatuses {
				if !status.Ready && status.RestartCount > 0 {
					containerNames = append(containerNames, status.Name)
				}
			}
			for _, name := range containerNames {
				//查看日志
				containers = append(containers, &types.Container{
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

func watchEvent(ctx context.Context, evCh chan contracts.Obj[*eventsv1.Event], releaseName string, fn contracts.WrapLogFn, lister v12.PodLister) {
	for {
		select {
		case <-ctx.Done():
			mlog.Debug("ctx.Done event")
			return
		case evobj, ok := <-evCh:
			if !ok {
				return
			}
			if evobj.Type() != contracts.Add {
				continue
			}

			var obj any = evobj.Current()
			event := obj.(*eventsv1.Event)
			p := event.Regarding
			get, err := lister.Pods(p.Namespace).Get(p.Name)
			if err != nil {
				mlog.Warningf("can't get pod ns: '%s', name: '%s'", p.Namespace, p.Name)
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

func uninstallRelease(releaseName, namespace string, log contracts.LogFn) error {
	actionConfig := getActionConfigAndSettings(namespace, log)
	uninstall := action.NewUninstall(actionConfig)
	_, err := uninstall.Run(releaseName)
	return err
}

func runInstall(ctx context.Context, releaseName string, chartRequested *chart.Chart, client *action.Install, valueOpts *values.Options) (*release.Release, error) {
	mlog.Debugf("Original chart version: %q", client.Version)
	if client.Version == "" && client.Devel {
		mlog.Debug("setting version to >0.0.0-0")
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

func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return fmt.Errorf("%s charts are not installable", ch.Metadata.Type)
}

const (
	StatusUnknown  string = "unknown"
	StatusPending  string = "pending"
	StatusDeployed string = "deployed"
	StatusFailed   string = "failed"
)

var (
	tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token" // #nosec G101
	rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

func releaseStatus(releaseName, namespace string) types.Deploy {
	actionConfig := getActionConfigAndSettings(namespace, mlog.Debugf)
	statusClient := action.NewStatus(actionConfig)
	run, err := statusClient.Run(releaseName)
	if err != nil {
		mlog.Warning(err)
		return types.Deploy_StatusUnknown
	}

	mlog.Debug(run.Info.Status)
	return formatStatus(run.Info.Status)
}

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

type ListReleaseItem struct {
	Release *release.Release
	Status  string
}

type ReleaseList map[string]ListReleaseItem

func (l ReleaseList) GetStatus(namespace, name string) string {
	if item, ok := l[fmt.Sprintf("%s-%s", namespace, name)]; ok {
		return item.Status
	}
	return StatusUnknown
}

func (l ReleaseList) Add(r *release.Release) {
	fn := func(run *release.Release) string {
		switch run.Info.Status {
		case release.StatusPendingUpgrade, release.StatusPendingRollback, release.StatusPendingInstall:
			return StatusPending
		case release.StatusDeployed:
			return StatusDeployed
		case release.StatusFailed:
			return StatusFailed
		default:
			return StatusUnknown
		}
	}
	l[fmt.Sprintf("%s-%s", r.Namespace, r.Name)] = ListReleaseItem{
		Release: r,
		Status:  fn(r),
	}
}

var dockerCfgOnce sync.Once

const dockerCfgOncePath = "/tmp/mars-docker-config.json"

func packageChart(path string, destDir string) (string, error) {
	newPackage := action.NewPackage()
	debug := app.Config().Debug
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
			os.WriteFile(dockerCfgOncePath, app.Config().ImagePullSecrets.FormatDockerCfg(), 0600)
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

type logWriter struct{}

func (l *logWriter) Write(p []byte) (n int, err error) {
	mlog.Debugf(string(p))
	return len(p), nil
}

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

func getActionConfigAndSettings(namespace string, log func(format string, v ...any)) *action.Configuration {
	actionConfig := new(action.Configuration)
	flags := genericclioptions.NewConfigFlags(true)
	flags = flags.WithDiscoveryQPS(-1)
	flags = flags.WithWrapConfigFn(wrapRestConfig)
	set := pflag.NewFlagSet("", pflag.ContinueOnError)
	flags.AddFlags(set)
	sets := []string{"--namespace=" + namespace}
	if app.Config().KubeConfig != "" {
		sets = append(sets, "--kubeconfig="+app.Config().KubeConfig)
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

func wrapRestConfig(config *restclient.Config) *restclient.Config {
	config.QPS = -1
	return config
}

func rollback(releaseName, namespace string, wait bool, log contracts.LogFn, dryRun bool) error {
	actionConfig := getActionConfigAndSettings(namespace, log)
	client := action.NewRollback(actionConfig)
	client.Wait = wait
	client.DryRun = dryRun
	client.DisableHooks = true
	client.WaitForJobs = wait

	return client.Run(releaseName)
}
