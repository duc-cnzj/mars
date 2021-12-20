package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"helm.sh/helm/v3/pkg/kube"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/spf13/pflag"

	"helm.sh/helm/v3/pkg/chart/loader"

	"helm.sh/helm/v3/pkg/downloader"

	"github.com/duc-cnzj/mars/internal/mlog"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	kube.New(nil)
}

type DeleteFunc func()

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
	file := fmt.Sprintf("mars-%s-%s.yaml", time.Now().Format("2006-01-02"), RandomString(20))
	info, err := app.Uploader().Put(file, bytes.NewReader(data))
	if err != nil {
		return "", nil, err
	}

	return info.Name(), NewCloser(func() error {
		mlog.Debug("delete file: " + info.Name())
		if err := app.Uploader().Delete(info.Name()); err != nil {
			mlog.Error("WriteConfigYamlToTmpFile error: ", err)
			return err
		}

		return nil
	}), nil
}

// UpgradeOrInstall TODO
func UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn func(format string, v ...interface{}), atomic bool) (*release.Release, error) {
	actionConfig, settings, err := getActionConfigAndSettings(namespace, fn)
	if err != nil {
		return nil, err
	}
	client := action.NewUpgrade(actionConfig)
	client.Install = true
	client.DisableOpenAPIValidation = true

	if atomic {
		client.Atomic = true
		client.Wait = true
		if app.Config().InstallTimeout != 0 {
			client.Timeout = app.Config().InstallTimeout
		} else {
			client.Timeout = 5 * 60 * time.Second
		}
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

func UninstallRelease(releaseName, namespace string, log action.DebugLog) error {
	actionConfig, _, err := getActionConfigAndSettings(namespace, log)
	if err != nil {
		return err
	}
	uninstall := action.NewUninstall(actionConfig)
	if _, err := uninstall.Run(releaseName); err != nil {
		mlog.Error(err)
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

func ReleaseStatus(releaseName, namespace string) (string, error) {
	actionConfig, _, err := getActionConfigAndSettings(namespace, mlog.Debugf)
	if err != nil {
		return "", err
	}
	statusClient := action.NewStatus(actionConfig)
	run, err := statusClient.Run(releaseName)
	if err != nil {
		mlog.Error(err)
		return "", err
	}

	mlog.Debug(run.Info.Status)
	switch run.Info.Status {
	case release.StatusPendingUpgrade, release.StatusPendingRollback, release.StatusPendingInstall:
		return StatusPending, nil
	case release.StatusDeployed:
		return StatusDeployed, nil
	case release.StatusFailed:
		return StatusFailed, nil
	default:
		return StatusUnknown, nil
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

func getActionConfigAndSettings(namespace string, log func(format string, v ...interface{})) (*action.Configuration, *cli.EnvSettings, error) {
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
