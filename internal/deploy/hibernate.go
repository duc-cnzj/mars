// hibernate.go 提供 biz.HibernateDeployer 的部署实现：把「定时休眠/唤醒」算好的副本字段
// 目标值经共享编排 ApplyProject 重部署落库。调度与状态推进在 cronjob，本文件只负责
// 「给一组 extraValues，重部署项目」这一件事，保持 deploy → biz 单向（biz 不能反向 import）。
package deploy

import (
	"context"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/samber/lo"
)

// hibernateUser 是定时休眠重部署的系统身份：ApplyProject 拒绝匿名部署（runner 建档/审计
// 需要 User 非空），休眠是系统行为，用一个固定超管身份做审计归属——updated_by 会显示它。
var hibernateUser = &biz.UserInfo{
	ID:    "system:hibernate",
	Email: schematype.SuperAdminEmail,
	Name:  "系统（定时休眠）",
	Roles: []string{schematype.MarsAdmin},
}

var _ biz.HibernateDeployer = (*hibernateDeployer)(nil)

// hibernateDeployer 以共享编排 ApplyProject 重部署项目，是 biz.HibernateDeployer 的实现。
type hibernateDeployer struct {
	deps ApplyProjectDeps
}

// NewHibernateDeployer 构造休眠重部署器，deps 与 gRPC/WS 部署入口同源（鉴权/仓库/git/版本/JobMgr）。
func NewHibernateDeployer(deps ApplyProjectDeps) biz.HibernateDeployer {
	return &hibernateDeployer{deps: deps}
}

// Redeploy 以给定额外配置项重部署项目：ExtraValues 由调用方（对账）算好——副本字段的目标值
// 已就位，其余自定义字段沿用项目当前值（ElementsLoader 从元素默认值起算，只传副本字段会把
// 其它字段重置回默认，故调用方须基于项目全量 FinalExtraValues 构造）。
// 复用项目当前 version 走乐观锁：期间有用户并发部署则版本不匹配报错，留给下一轮对账重试。
func (d *hibernateDeployer) Redeploy(ctx context.Context, proj *biz.Project, extraValues []*websocket_pb.ExtraValue) error {
	input := &JobInput{
		Type:        websocket_pb.Type_UpdateProject,
		NamespaceId: int32(proj.NamespaceID),
		Name:        proj.Name,
		RepoID:      int32(proj.RepoID),
		GitBranch:   proj.GitBranch,
		GitCommit:   proj.GitCommit,
		Config:      proj.Config,
		Atomic:      lo.ToPtr(proj.Atomic),
		ExtraValues: extraValues,
		Version:     lo.ToPtr(int32(proj.Version)),
		ProjectID:   int32(proj.ID),
		User:        hibernateUser,
		DryRun:      false,
		PubSub:      NewEmptyPubSub(),
		Messager:    NoopMsger{},
	}
	_, err := ApplyProject(ctx, d.deps, &ApplyProjectInput{JobInput: input})
	return err
}
