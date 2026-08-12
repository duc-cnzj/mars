package biz

import (
	"context"
	"errors"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// DeployBiz 封装项目部署生命周期相关的用例编排。
type DeployBiz interface {
	// DeleteProject 删除项目：先卸载 helm release（非 NotFound 的卸载失败会中止删除，
	// 保留项目记录以便重试），成功后删除 DB 记录并派发项目删除事件。
	// 这是项目删除的领域不变式：卸载失败时不得删 DB，否则留下无法重试的孤儿 release。
	// id 为待删除的 DB 主键，proj 为调用方已校验权限的 Show 结果（用于卸载与事件派发）。
	DeleteProject(ctx context.Context, id int, proj *Project, log LogFn) error
}

type deployBiz struct {
	logger     mlog.Logger
	projRepo   ProjectRepo
	helmerRepo HelmerRepo
	eventRepo  EventRepo
}

// NewDeployBiz 构造 deploy biz。
func NewDeployBiz(logger mlog.Logger, projRepo ProjectRepo, helmerRepo HelmerRepo, eventRepo EventRepo) DeployBiz {
	return &deployBiz{
		logger:     logger.WithModule("biz/deploy"),
		projRepo:   projRepo,
		helmerRepo: helmerRepo,
		eventRepo:  eventRepo,
	}
}

// DeleteProject 校验 project 后删除项目：先卸载 helm release（已不存在的 release 不算失败），
// 成功后再删 DB 记录并派发删除事件。
func (d *deployBiz) DeleteProject(ctx context.Context, id int, proj *Project, log LogFn) error {
	if proj == nil {
		return errs.WrapInvalidArgument(errors.New("project 不能为空"), "delete project")
	}
	// 先卸载 release 再删 DB 记录：卸载失败（release 仍存在）时中止删除并保留
	// 项目记录，用户可重试。release 已不存在（手动清理/孤儿）不算失败。
	if err := d.helmerRepo.Uninstall(proj.Name, proj.Namespace.Name, log); err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		return err
	}
	if err := d.projRepo.Delete(ctx, id); err != nil {
		return err
	}
	d.eventRepo.Dispatch(EventProjectDeleted, &ProjectDeletedPayload{
		NamespaceID: proj.NamespaceID,
		ProjectID:   proj.ID,
	})
	return nil
}
