package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	v12 "k8s.io/client-go/listers/core/v1"

	"github.com/duc-cnzj/mars/internal/utils/recovery"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"net"
	"os"
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"

	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"

	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/api/events/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	kube.New(nil)
}

type internalCloser struct {
	closeFn func() error
}

func (i *internalCloser) Close() error {
	return i.closeFn()
}

func NewCloser(fn func() error) io.Closer {
	return &internalCloser{closeFn: fn}
}

func WriteConfigYamlToTmpFile(data []byte) (string, io.Closer, error) {
	var localUploader = app.LocalUploader()
	file := fmt.Sprintf("mars-%s-%s.yaml", time.Now().Format("2006-01-02"), RandomString(20))
	info, err := localUploader.Put(file, bytes.NewReader(data))
	if err != nil {
		return "", nil, err
	}
	path := info.Path()

	return path, NewCloser(func() error {
		mlog.Debug("delete file: " + path)
		if err := localUploader.Delete(path); err != nil {
			mlog.Error("WriteConfigYamlToTmpFile error: ", err)
			return err
		}

		return nil
	}), nil
}

// UpgradeOrInstall
// 不会自动回滚
func UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn contracts.WrapLogFn, wait bool, timeoutSeconds int64, dryRun bool, podSelectors []string) (*release.Release, error) {
	actionConfig, settings, err := getActionConfigAndSettings(namespace, fn.UnWrap())
	if err != nil {
		return nil, err
	}
	client := action.NewUpgrade(actionConfig)
	client.Install = true
	client.Atomic = false
	client.Wait = wait
	client.DryRun = dryRun
	client.DisableOpenAPIValidation = true
	var selectorList []labels.Selector
	for _, label := range podSelectors {
		selector, _ := metav1.ParseToLabelSelector(label)
		asSelector, _ := metav1.LabelSelectorAsSelector(selector)
		selectorList = append(selectorList, asSelector)
	}

	if wait && !dryRun {
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
			defer recovery.HandlePanic("UpgradeOrInstall pod-fan-out")
			watchPodStatus(fanOutCtx, podCh, selectorList, fn)
		}()
		go func() {
			defer recovery.HandlePanic("UpgradeOrInstall event-fan-out")
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

	client.Namespace = namespace
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
			mlog.Debug("start install release", valueOpts)
			return runInstall(ctx, releaseName, ch, instClient, valueOpts, settings)
		}
	}

	if client.Version == "" && client.Devel {
		mlog.Debug("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	vals, err := valueOpts.MergeValues(getter.All(settings))
	if err != nil {
		return nil, err
	}

	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			return nil, err
		}
	}

	if ch.Metadata.Deprecated {
		mlog.Warning("This chart is deprecated")
	}

	return client.RunWithContext(ctx, releaseName, ch, vals)
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
				contaniners    []*types.Container
			)
			for _, status := range p.Status.ContainerStatuses {
				if !status.Ready && status.RestartCount > 0 {
					containerNames = append(containerNames, status.Name)
				}
			}
			for _, name := range containerNames {
				//查看日志
				contaniners = append(contaniners, &types.Container{
					Namespace: p.Namespace,
					Pod:       p.Name,
					Container: name,
				})
			}
			if len(contaniners) > 0 {
				fn(contaniners, "容器多次异常重启")
			}
		}
	}
}

func watchEvent(ctx context.Context, evCh chan contracts.Obj[*v1.Event], releaseName string, fn contracts.WrapLogFn, lister v12.PodLister) {
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
			event := obj.(*v1.Event)
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

func UninstallRelease(releaseName, namespace string, log contracts.LogFn) error {
	actionConfig, _, err := getActionConfigAndSettings(namespace, log)
	if err != nil {
		return err
	}
	uninstall := action.NewUninstall(actionConfig)
	if _, err := uninstall.Run(releaseName); err != nil {
		mlog.Warning(err)
		return err
	}
	return nil
}

func runInstall(ctx context.Context, releaseName string, chartRequested *chart.Chart, client *action.Install, valueOpts *values.Options, settings *cli.EnvSettings) (*release.Release, error) {
	mlog.Debugf("Original chart version: %q", client.Version)
	if client.Version == "" && client.Devel {
		mlog.Debug("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	client.ReleaseName = releaseName

	p := getter.All(settings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	if err := checkIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		mlog.Warning("This chart is deprecated")
	}

	client.Namespace = settings.Namespace()
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
	tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

func ReleaseStatus(releaseName, namespace string) types.Deploy {
	actionConfig, _, err := getActionConfigAndSettings(namespace, mlog.Debugf)
	if err != nil {
		return types.Deploy_StatusUnknown
	}
	statusClient := action.NewStatus(actionConfig)
	run, err := statusClient.Run(releaseName)
	if err != nil {
		mlog.Warning(err)
		return types.Deploy_StatusUnknown
	}

	mlog.Debug(run.Info.Status)
	switch run.Info.Status {
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

func PackageChart(path string, destDir string) (string, error) {
	_, settings, err := getActionConfigAndSettings("", mlog.Debugf)
	if err != nil {
		return "", err
	}
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
		downloadManager := &downloader.Manager{
			Out:              ioutil.Discard,
			ChartPath:        path,
			Keyring:          newPackage.Keyring,
			Getters:          getter.All(settings),
			Debug:            settings.Debug,
			RepositoryConfig: settings.RepositoryConfig,
			RepositoryCache:  settings.RepositoryCache,
		}

		if err := downloadManager.Update(); err != nil {
			return "", err
		}
	}

	return newPackage.Run(path, nil)
}

func getActionConfigAndSettings(namespace string, log func(format string, v ...any)) (*action.Configuration, *cli.EnvSettings, error) {
	settings := cli.New()
	sflags := pflag.NewFlagSet("", pflag.ContinueOnError)
	settings.AddFlags(sflags)
	ssets := []string{"--namespace=" + namespace, fmt.Sprintf("--debug=%T", app.App().IsDebug())}

	actionConfig := new(action.Configuration)
	flags := genericclioptions.NewConfigFlags(true)
	set := pflag.NewFlagSet("", pflag.ContinueOnError)
	flags.AddFlags(set)
	sets := []string{"--namespace=" + namespace}
	if app.Config().KubeConfig != "" {
		sets = append(sets, "--kubeconfig="+app.Config().KubeConfig)
		ssets = append(ssets, "--kubeconfig="+app.Config().KubeConfig)
	} else {
		host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
		server := "https://" + net.JoinHostPort(host, port)
		token, _ := ioutil.ReadFile(tokenFile)
		sets = append(sets, "--server="+server, "--token="+string(token), "--certificate-authority="+rootCAFile)
		ssets = append(ssets, "--kube-apiserver="+server, "--kube-token="+string(token), "--kube-ca-file="+rootCAFile)
	}

	sflags.Parse(ssets)
	set.Parse(sets)

	if err := actionConfig.Init(flags, namespace, "", log); err != nil {
		return nil, nil, err
	}

	return actionConfig, settings, nil
}

func GetSlugName[T int64 | int](namespaceId T, name string) string {
	return Md5(fmt.Sprintf("%d-%s", namespaceId, name))
}

func Rollback(releaseName, namespace string, wait bool, log contracts.LogFn, dryRun bool) error {
	actionConfig, _, err := getActionConfigAndSettings(namespace, log)
	if err != nil {
		return err
	}
	client := action.NewRollback(actionConfig)
	client.Wait = wait
	client.DryRun = dryRun
	client.DisableHooks = true
	client.WaitForJobs = wait

	return client.Run(releaseName)
}
