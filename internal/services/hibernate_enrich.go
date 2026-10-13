package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/project"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/samber/lo"
)

// enrichProjectHibernate 给一批项目模型补充项目级休眠运行态（开关 + 状态）。
//
// 项目列表/详情与命名空间列表/详情（其中的嵌套项目）统一走这个出口——工作台的
// 项目行直接从命名空间列表取数，若只补项目侧接口，行上的休眠图标在点开弹层前
// 永远显示为「未开启」（列表项零值），用户必须先点一下才能看到真实的开关态。
//
// 未配置休眠的项目保持零值（Enabled=false、State=""）。补充失败只记日志不阻断主响应——
// 休眠信息是附加展示，其查询故障不应让列表/详情整体 500。
// 传入的模型可跨多个命名空间：id 汇总后一次批量查询，避免逐命名空间 N+1。
func enrichProjectHibernate(ctx context.Context, logger mlog.Logger, hb biz.HibernateBiz, models ...*types.ProjectModel) {
	ids := lo.FilterMap(models, func(m *types.ProjectModel, _ int) (int, bool) {
		return int(m.Id), m.Id > 0
	})
	if len(ids) == 0 {
		return
	}
	hmap, err := hb.ListByProjectIDs(ctx, ids...)
	if err != nil {
		logger.WarningCtxf(ctx, "补充项目休眠状态失败: %v", err)
		return
	}
	for _, m := range models {
		h, ok := hmap[int(m.Id)]
		if !ok || h == nil {
			continue
		}
		m.HibernateEnabled = h.Enabled
		m.HibernateState = string(h.State)
		m.HibernateWindowNames = h.WindowNames
	}
}

// enrichLivenessHibernate 给项目治理清单条目补充项目级休眠运行态（开关 + 状态 + 选定时间窗），
// 使后台治理列表能标注并过滤「已开启定时休眠」的项目。与列表/详情同款附加展示语义：id 汇总后
// 一次批量查询，补充失败只记日志不阻断主响应（休眠信息查询故障不应让治理清单整体 500）。
func enrichLivenessHibernate(ctx context.Context, logger mlog.Logger, hb biz.HibernateBiz, items []*project.LivenessItem) {
	ids := lo.FilterMap(items, func(m *project.LivenessItem, _ int) (int, bool) {
		return int(m.Id), m.Id > 0
	})
	if len(ids) == 0 {
		return
	}
	hmap, err := hb.ListByProjectIDs(ctx, ids...)
	if err != nil {
		logger.WarningCtxf(ctx, "补充项目休眠状态失败: %v", err)
		return
	}
	for _, m := range items {
		h, ok := hmap[int(m.Id)]
		if !ok || h == nil {
			continue
		}
		m.HibernateEnabled = h.Enabled
		m.HibernateState = string(h.State)
		m.HibernateWindowNames = h.WindowNames
	}
}

// repoHibernateWindowNames 取仓库声明的可用休眠时间窗名列表（项目选定子集的合法范围）。
// 仓库为空、或其 mars 配置未启用休眠段时返回 nil——治理列表据此判定「无可选窗」，
// 不向用户提供时间窗编辑入口（越界选择后端会 400）。
func repoHibernateWindowNames(repo *biz.Repo) []string {
	if repo == nil {
		return nil
	}
	cfg := repo.GetMarsConfig()
	if cfg.Hibernate == nil {
		return nil
	}
	return cfg.Hibernate.WindowNames
}
