package deploy

// apply.go 定义 gRPC/WebSocket 共享的部署编排用例函数 ApplyProject：
// 鉴权 → 仓库取回与名缺省 → git ensure → 版本反查 → 装配 Job → 传输层钩子
// → ctx watcher → InstallProject。原先 gRPC projectSvc.apply 与 ws installProject
// 双份编排各写一遍，行为漂移（WS 缺权限校验/ctx 取消/版本反查），现收敛到一处。
//
// 该函数必须放在 deploy 包而非 biz.ProjectBiz 方法：internal/deploy import
// internal/biz（JobInput.User *biz.UserInfo），biz 反向 import deploy 即成环，
// 部署编排必须调用 deploy.InstallProject/JobManager，故共享函数只能落 deploy。

import (
	"context"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/samber/lo"
)

// ApplyProjectDeps 是共享部署编排的依赖：AccessBiz 用于命名空间访问校验（WS 路径此前
// 完全无此校验，注入即补上 IDOR 漏洞；用户提取由 AccessBiz 内部走 MustGetUser——
// ApplyProject 入口已把 JobInput.User 物化进 ctx，gRPC/WS 两路径一致），RepoBiz/GitBiz 用于仓库
// 取回与 git ensure，ProjectBiz 用于版本反查，JobMgr 用于装配 Job。Logger 仅用于 ctx watcher
// 协程的 HandlePanic 兜底（本函数不打印业务错误日志，错误统一上抛由 services 层 logError 打印）。
type ApplyProjectDeps struct {
	AccessBiz  biz.AccessBiz
	RepoBiz    biz.RepoBiz
	GitBiz     biz.GitBiz
	ProjectBiz biz.ProjectBiz
	JobMgr     JobManager
	Logger     mlog.Logger
}

// ApplyProjectInput 是共享部署编排的输入：JobInput 为调用方已填好的部署参数
// （Type/NamespaceId/RepoID/GitBranch/GitCommit/Config/Atomic/ExtraValues/Version/
// User/DryRun/PubSub/Messager）；OnJob 为传输层钩子（WS 用它登记取消回调），
// 返回非 nil = 传输层已自行处理失败（如已发失败帧），跳过 watcher 与 InstallProject。
type ApplyProjectInput struct {
	JobInput *JobInput
	OnJob    func(job Job) error
}

// ApplyProject 执行部署用例编排（gRPC/WebSocket 共享），返回 (job, 流水线错误)。
// 顺序与现 projectSvc.apply 完全一致保证行为对齐；唯一的传输层差异（pubsub 的
// 创建/关闭、取消任务登记）由调用方在 JobInput/OnJob 中表达。
// 本函数不打印错误日志：所有失败路径（鉴权/仓库取回/git ensure/版本反查/InstallProject）
// 都直接上抛 err，由 services 层 logError（gRPC）或 wc.logger.Error（WS）统一打印一次。
func ApplyProject(ctx context.Context, deps ApplyProjectDeps, input *ApplyProjectInput) (Job, error) {
	// 匿名部署（User==nil）无法建档/审计（runner 会 nil-deref panic），一律拒绝。
	// 鉴权流程必带用户，此守卫对合法流恒不触发。
	if input.JobInput.User == nil {
		return nil, errs.WrapPermissionDenied(errs.ErrorPermissionDenied, "发起部署（需要登录用户）")
	}

	// 把部署身份物化进 ctx：WS 的 user 在 Conn 上不在 ctx，此处统一从 JobInput 取，
	// 使 gRPC/WS 两路径的 AccessBiz（内部走 MustGetUser）都能安全取值。
	ctx = biz.SetUser(ctx, input.JobInput.User)

	// 命名空间访问校验：私有命名空间对非成员/非 admin 隐藏部署入口。
	// 错误码与 services 侧 AccessBiz 语义一致（ErrorPermissionDenied）。
	if _, err := deps.AccessBiz.RequireNamespaceAccessByID(ctx, int(input.JobInput.NamespaceId)); err != nil {
		return nil, err
	}

	show, err := deps.RepoBiz.Get(ctx, int(input.JobInput.RepoID))
	if err != nil {
		return nil, err
	}
	// 未显式传项目名时以仓库名缺省（messager 的 slug 依赖它）。
	if input.JobInput.Name == "" {
		input.JobInput.Name = show.Name
	}
	// 部署帧 slug 的唯一权威来源：传输层构造 messager 不绑定 slug（构造期 name 可能为
	// 空，与前端 toSlug 关联的日志 key 不一致），此处名缺省解析后就地回填
	// GetSlugName(ns, 最终名)，保证后续所有帧（含 git ensure 消息）携带最终名。
	// Messager 为空（如测试中的空 messager）时跳过。
	if input.JobInput.Messager != nil {
		input.JobInput.Messager.SetSlug(GetSlugName(input.JobInput.NamespaceId, input.JobInput.Name))
	}

	if show.NeedGitRepo {
		var msgs []string
		input.JobInput.GitBranch, input.JobInput.GitCommit, msgs, err = deps.GitBiz.EnsureBranchAndCommit(ctx, show, input.JobInput.GitBranch, input.JobInput.GitCommit)
		if err != nil {
			return nil, err
		}
		for _, msg := range msgs {
			input.JobInput.Messager.SendMsg(msg)
		}
	}

	// 版本反查：仅当未显式传 ProjectID 且请求带 Version 时才按 name+ns 定位项目。
	// ProjectID==0 守卫防止 WS 显式传 ProjectID 被清零后误走反查的回归。
	if input.JobInput.ProjectID == 0 && lo.FromPtr(input.JobInput.Version) > 0 {
		proj, err := deps.ProjectBiz.FindByName(ctx, input.JobInput.Name, int(input.JobInput.NamespaceId))
		if err != nil {
			// 只有 NotFound 才说明是首次部署，可放行；真实 DB 故障必须上抛，
			// 否则 ProjectID=0 会让 runner 报"版本不匹配"，把故障伪装成预期。
			if !errs.IsNotFound(err) {
				return nil, err
			}
		} else {
			input.JobInput.ProjectID = int32(proj.ID)
		}
	}

	job := deps.JobMgr.NewJob(input.JobInput)

	// 传输层钩子：WS 用它登记取消任务，返回非 nil 说明传输层已发失败帧，
	// 跳过 watcher 与 InstallProject。
	if input.OnJob != nil {
		if err := input.OnJob(job); err != nil {
			return job, nil
		}
	}

	// ctx watcher：连接/请求取消即停任务。close(ch) 放 defer，InstallProject 一旦
	// panic 也会解除 select 阻塞，避免 watcher goroutine 泄漏到 ctx 结束才退出。
	ch := make(chan struct{})
	go func() {
		defer deps.Logger.HandlePanic("deploy.ApplyProject: stop watcher")
		select {
		case <-ctx.Done():
			job.Stop(ctx.Err())
		case <-ch:
		}
	}()
	defer close(ch)

	return job, InstallProject(ctx, job)
}
