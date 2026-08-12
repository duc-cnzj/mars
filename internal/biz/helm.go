package biz

import (
	"context"
	"errors"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

// HelmerBiz 收口 helm release 生命周期业务：安装/升级、回滚、卸载、状态查询与打包。
type HelmerBiz interface {
	// UpgradeOrInstall 安装或升级 release。
	UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn WrapLogFn, wait bool, timeoutSeconds int64, dryRun bool, desc string) (*release.Release, error)
	// Rollback 回滚 release 到上一个版本。
	Rollback(releaseName, namespace string, wait bool, log LogFn, dryRun bool) error
	// Uninstall 卸载 release。
	Uninstall(releaseName, namespace string, log LogFn) error
	// ReleaseStatus 查询 release 的部署状态。
	ReleaseStatus(releaseName, namespace string) types.Deploy
	// PackageChart 把本地 chart 目录打包成 tgz。
	PackageChart(path string, destDir string) (string, error)
}

type helmerBiz struct {
	helmer HelmerRepo
}

// NewHelmerBiz 构造 helm biz。
func NewHelmerBiz(helmer HelmerRepo) HelmerBiz {
	return &helmerBiz{helmer: helmer}
}

// UpgradeOrInstall 校验 releaseName/namespace 后安装或升级 helm release。
func (h *helmerBiz) UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn WrapLogFn, wait bool, timeoutSeconds int64, dryRun bool, desc string) (*release.Release, error) {
	if releaseName == "" || namespace == "" {
		return nil, errs.WrapInvalidArgument(errors.New("releaseName 或 namespace 不能为空"), "upgrade or install")
	}
	return h.helmer.UpgradeOrInstall(ctx, releaseName, namespace, ch, valueOpts, fn, wait, timeoutSeconds, dryRun, desc)
}

// Rollback 校验 releaseName/namespace 后回滚 helm release。
func (h *helmerBiz) Rollback(releaseName, namespace string, wait bool, log LogFn, dryRun bool) error {
	if releaseName == "" || namespace == "" {
		return errs.WrapInvalidArgument(errors.New("releaseName 或 namespace 不能为空"), "rollback")
	}
	return h.helmer.Rollback(releaseName, namespace, wait, log, dryRun)
}

// Uninstall 校验 releaseName/namespace 后卸载 helm release。
func (h *helmerBiz) Uninstall(releaseName, namespace string, log LogFn) error {
	if releaseName == "" || namespace == "" {
		return errs.WrapInvalidArgument(errors.New("releaseName 或 namespace 不能为空"), "uninstall")
	}
	return h.helmer.Uninstall(releaseName, namespace, log)
}

// ReleaseStatus 查询 helm release 的部署状态（透传 repo）。
func (h *helmerBiz) ReleaseStatus(releaseName, namespace string) types.Deploy {
	return h.helmer.ReleaseStatus(releaseName, namespace)
}

// PackageChart 打包本地 chart 到目标目录（透传 repo）。
func (h *helmerBiz) PackageChart(path string, destDir string) (string, error) {
	return h.helmer.PackageChart(path, destDir)
}

// HelmerRepo 是 helm release 生命周期操作端口。
type HelmerRepo interface {
	// UpgradeOrInstall 安装或升级 release。
	UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn WrapLogFn, wait bool, timeoutSeconds int64, dryRun bool, desc string) (*release.Release, error)
	// Rollback 回滚 release 到上一个版本。
	Rollback(releaseName, namespace string, wait bool, log LogFn, dryRun bool) error
	// Uninstall 卸载 release。
	Uninstall(releaseName, namespace string, log LogFn) error
	// ReleaseStatus 查询 release 的部署状态。
	ReleaseStatus(releaseName, namespace string) types.Deploy
	// PackageChart 把本地 chart 目录打包成 tgz。
	PackageChart(path string, destDir string) (string, error)
}
