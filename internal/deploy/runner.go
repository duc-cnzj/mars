package deploy

import (
	"context"
	"errors"
	"fmt"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/pipeline"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	yaml2 "github.com/duc-cnzj/mars/v6/internal/util/yaml"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

// ErrorVersionNotMatched 是部署版本冲突错误：多为并发部署或项目已存在导致。
var ErrorVersionNotMatched = errors.New("[部署冲突]: 1. 可能是多个人同时部署导致 2. 项目已经存在")

// ErrCancel 是部署被取消的信号错误，jobRunner.Stop 会记录它；
// ws 传输层的 TaskManager 停止任务时也用它作为取消原因。
var ErrCancel = errors.New("取消本次部署，自动回滚到上一个版本！")

// Result* 是部署结果状态（websocket_pb.ResultType 的语义别名），
// 由 deployResult 记录并由 DeployMsger 上报，故保留在部署侧。
// Ws* 协议类型标签与连接调优常量属于 ws 传输层，见 internal/services/websocket/conn.go。
const (
	ResultError             = websocket_pb.ResultType_Error
	ResultSuccess           = websocket_pb.ResultType_Success
	ResultDeployed          = websocket_pb.ResultType_Deployed
	ResultDeployFailed      = websocket_pb.ResultType_DeployedFailed
	ResultDeployCanceled    = websocket_pb.ResultType_DeployedCanceled
	ResultLogWithContainers = websocket_pb.ResultType_LogWithContainers

	// WsReloadProjects 由部署完成后的 reload 广播使用（runner.go），故保留在部署侧。
	WsReloadProjects = websocket_pb.Type_ReloadProjects
)

var _ JobManager = (*jobManager)(nil)

// JobManager 创建部署 Job，是部署入口的抽象，供 websocket/gRPC 传输层复用。
type JobManager interface {
	//	创建一个新的Job
	NewJob(input *JobInput) Job
}

// Job 是一次部署流水线的完整接口：链式调用 GlobalLock→Validate→LoadConfigs→Run→Finish。
type Job interface {
	Stop(error)
	IsNotDryRun() bool

	ID() string
	GlobalLock() Job
	Validate() Job
	LoadConfigs() Job
	Run(ctx context.Context) Job
	Finish() Job
	Error() error
	Project() *biz.Project
	Manifests() []string

	OnError(p int, fn func(err error, sendResultToUser func())) Job
	OnSuccess(p int, fn func(err error, sendResultToUser func())) Job
	OnFinally(p int, fn func(err error, sendResultToUser func())) Job
}

// JobManagerDeps 收口 NewJobManager 的构造依赖，由 wire 按字段注入。
// 与 14 个 gRPC service 的 XxxSvcDeps / websocket 的 WebsocketManagerDeps 对齐，
// 消灭 13 个位置参数的 god 构造器。
type JobManagerDeps struct {
	Timer            timer.Timer
	Logger           mlog.Logger
	ReleaseInstaller ReleaseInstaller
	RepoRepo         biz.RepoRepo
	NsRepo           biz.NamespaceRepo
	ProjRepo         biz.ProjectRepo
	Helmer           biz.HelmerRepo
	Uploader         uploader.Uploader
	Locker           locker.Locker
	K8sRepo          biz.K8sRepo
	EventRepo        biz.EventRepo
	PluginManager    app.PluginManager
}

type jobManager struct {
	logger mlog.Logger

	timer            timer.Timer
	releaseInstaller ReleaseInstaller
	nsRepo           biz.NamespaceRepo
	projRepo         biz.ProjectRepo
	eventRepo        biz.EventRepo
	k8sRepo          biz.K8sRepo
	helmRepo         biz.HelmerRepo
	repoRepo         biz.RepoRepo

	locker        locker.Locker
	uploader      uploader.Uploader
	pluginManager app.PluginManager
}

// NewJobManager 构造部署任务管理器，依赖由 wire 从 JobManagerDeps 注入。
func NewJobManager(deps JobManagerDeps) JobManager {
	return &jobManager{
		timer:            deps.Timer,
		releaseInstaller: deps.ReleaseInstaller,
		uploader:         deps.Uploader,
		repoRepo:         deps.RepoRepo,
		logger:           deps.Logger,
		nsRepo:           deps.NsRepo,
		projRepo:         deps.ProjRepo,
		k8sRepo:          deps.K8sRepo,
		pluginManager:    deps.PluginManager,
		helmRepo:         deps.Helmer,
		locker:           deps.Locker,
		eventRepo:        deps.EventRepo,
	}
}

// NewJob 基于输入创建一次部署任务（jobRunner）并组装其全部执行依赖。
func (j *jobManager) NewJob(input *JobInput) Job {
	jb := &jobRunner{
		installer:       j.releaseInstaller,
		logger:          j.logger.WithModule("socket/job"),
		nsRepo:          j.nsRepo,
		projRepo:        j.projRepo,
		repoRepo:        j.repoRepo,
		pluginMgr:       j.pluginManager,
		helmer:          j.helmRepo,
		locker:          j.locker,
		k8sRepo:         j.k8sRepo,
		eventRepo:       j.eventRepo,
		timer:           j.timer,
		uploader:        j.uploader,
		loaders:         defaultLoaders(),
		dryRun:          input.DryRun,
		input:           input,
		finallyCallback: priority[func(err error, next func())]{},
		errorCallback:   priority[func(err error, next func())]{},
		successCallback: priority[func(err error, next func())]{},
		deployResult:    &deployResult{},
		valuesOptions:   &values.Options{},
		messageCh:       newSafeWriteMessageCh(j.logger, 100),
		messager:        input.Messager,
		user:            input.User,
		// 透传调用方按请求传入的超时；为 0 时由 releaseInstaller 用构造时
		// 读到的 config.InstallTimeout 兜底，默认值只此一处，避免双份来源。
		timeoutSeconds: int64(input.TimeoutSeconds),
	}
	jb.stopCtx, jb.stopFn = context.WithCancelCause(context.TODO())

	return jb
}

// JobInput 是一次部署的输入参数：来源、目标项目、镜像/配置与取消信号等。
type JobInput struct {
	Type        websocket_pb.Type
	NamespaceId int32
	Name        string
	RepoID      int32
	GitBranch   string
	GitCommit   string
	Config      string
	Atomic      *bool
	ExtraValues []*websocket_pb.ExtraValue
	Version     *int32
	ProjectID   int32

	TimeoutSeconds int32
	User           *biz.UserInfo
	DryRun         bool

	PubSub   app.PubSub  `json:"-"`
	Messager DeployMsger `json:"-"`
}

// Slug 返回该输入的部署任务标识（namespace+name 哈希），用于全局锁区分。
func (job *JobInput) Slug() string {
	return GetSlugName(job.NamespaceId, job.Name)
}

// IsNotDryRun 判断是否为真实部署（非 dry-run）。
func (job *JobInput) IsNotDryRun() bool {
	return !job.DryRun
}

type jobRunner struct {
	// 这些属性在 new runner 的时候就已经初始化了
	logger          mlog.Logger
	nsRepo          biz.NamespaceRepo
	projRepo        biz.ProjectRepo
	repoRepo        biz.RepoRepo
	helmer          biz.HelmerRepo
	locker          locker.Locker
	k8sRepo         biz.K8sRepo
	eventRepo       biz.EventRepo
	messager        DeployMsger
	timeoutSeconds  int64
	uploader        uploader.Uploader
	pluginMgr       app.PluginManager
	installer       ReleaseInstaller
	messageCh       SafeWriteMessageChan
	deployResult    *deployResult
	loaders         []Loader
	input           *JobInput
	finallyCallback priority[func(err error, next func())]
	errorCallback   priority[func(err error, next func())]
	successCallback priority[func(err error, next func())]
	stopCtx         context.Context
	stopFn          func(error)
	dryRun          bool
	user            *biz.UserInfo
	timer           timer.Timer

	// 这些属性在执行的时候才会初始化
	// Validate 阶段被初始化
	isNew            bool
	ns               *biz.Namespace
	repo             *biz.Repo
	config           *mars.Config
	project          *biz.Project
	imagePullSecrets []string
	commit           *biz.Commit
	oldConf          biz.YamlPrettier

	// LoadConfigs 阶段被初始化。
	// 加载链内部的中间值（用户配置/元素片段/系统渲染值）只活在 LoadContext 里，
	// 这里只保留 Run 阶段仍要消费的产物，链结束后由 LoadConfigs 收口写回。
	chart            *chart.Chart
	finalExtraValues []*websocket_pb.ExtraValue
	vars             vars
	valuesOptions    *values.Options

	err error

	// 部署成功后的 manifest
	manifests []string
}

// ID 返回部署任务的全局锁标识。
func (j *jobRunner) ID() string {
	return j.input.Slug()
}

// IsNotDryRun 判断是否为真实部署（非 dry-run）。
func (j *jobRunner) IsNotDryRun() bool {
	return !j.dryRun
}

// GlobalLock 抢占该任务的全局锁，未抢到则报"正在部署中"并终止流水线。
func (j *jobRunner) GlobalLock() Job {
	if j.HasError() {
		return j
	}
	releaseFn, acquired := j.locker.RenewalAcquire(j.ID(), 30, 20)
	if !acquired {
		return j.SetError(errors.New("正在部署中，请稍后再试"))
	}

	return j.OnFinally(-1, func(err error, sendResultToUser func()) {
		sendResultToUser()
		releaseFn()
	})
}

// Validate 校验部署类型、命名空间、仓库与项目存在性；
// 不存在则新建项目，存在则校验版本并置为部署中状态。
func (j *jobRunner) Validate() Job {
	var err error
	if !j.typeValidated() {
		return j.SetError(errors.New("type error: " + j.input.Type.String()))
	}

	j.messager.SendMsg("[start]: 收到请求，开始创建项目")
	j.messager.To(5)

	j.messager.SendMsg("[Check]: 校验名称空间...")

	j.ns, err = j.nsRepo.Show(context.TODO(), int(j.input.NamespaceId))
	if err != nil {
		return j.SetError(fmt.Errorf("[FAILED]: 校验名称空间: %w", err))
	}

	j.messager.SendMsg("[Loading]: 加载用户配置")
	j.messager.To(10)

	j.repo, err = j.repoRepo.Get(context.TODO(), int(j.input.RepoID))
	if err != nil {
		return j.SetError(err)
	}
	j.config = j.repo.MarsConfig

	j.messager.SendMsg("[Check]: 检查项目是否存在")

	found, err := j.projRepo.FindByName(context.TODO(), j.input.Name, j.ns.ID)
	if err != nil {
		// 只有确切的"记录不存在"才走新建分支。DB 抖动/网络错误等非 NotFound
		// 一律直接失败：否则会把"查询失败"误判为"项目不存在"，
		// 客户端重试时会对同一项目反复创建，产生重复项目行。
		if !errs.IsNotFound(err) {
			return j.SetError(err)
		}
		createProjectInput := &biz.CreateProjectInput{
			Name:         j.input.Name,
			GitProjectID: int(j.repo.GitProjectID),
			GitBranch:    j.input.GitBranch,
			GitCommit:    j.input.GitCommit,
			Config:       j.input.Config,
			Atomic:       j.input.Atomic,
			ConfigType:   j.config.ConfigFileType,
			NamespaceID:  j.ns.ID,
			RepoID:       j.repo.ID,
			Creator:      j.user.Email,
		}
		j.messager.SendMsg("[Check]: 新建项目")
		createProjectInput.DeployStatus = types.Deploy_StatusDeploying
		j.isNew = true
		if j.IsNotDryRun() {
			j.project, err = j.projRepo.Create(context.TODO(), createProjectInput)
			if err != nil {
				return j.SetError(err)
			}
			createdID := j.project.ID
			j.OnError(1, func(err error, sendResultToUser func()) {
				j.logger.Debug("清理项目")
				j.projRepo.Delete(context.TODO(), createdID)
				sendResultToUser()
			})
		} else {
			// dry-run 不落库，但 Run 的 ReleaseName 与 loader 的 host 变量渲染
			// （ctx.Project.Name）仍依赖 project 的身份字段——合成占位项目，
			// 避免后续链路对 nil project 解引用 panic。
			j.project = &biz.Project{
				Name:      createProjectInput.Name,
				GitBranch: createProjectInput.GitBranch,
				GitCommit: createProjectInput.GitCommit,
			}
			// 清掉 FindByName 的 NotFound：dry-run 新建已由占位项目承接，
			// 否则末尾 SetError 会把"预期内的不存在"误报成部署失败。
			err = nil
		}
	} else {
		j.project = found
		version := j.project.Version
		if j.IsNotDryRun() {
			j.messager.SendMsg(fmt.Sprintf("[Check]: 检查当前版本, version: %v", lo.FromPtr(j.input.Version)))
			j.project, err = j.projRepo.UpdateStatusByVersion(context.TODO(), int(j.input.ProjectID), types.Deploy_StatusDeploying, int(lo.FromPtr(j.input.Version)))
			if err != nil {
				return j.SetError(fmt.Errorf("%w: %w", ErrorVersionNotMatched, err))
			}
			j.OnError(1, func(err error, sendResultToUser func()) {
				j.project, _ = j.projRepo.UpdateVersion(context.TODO(), j.project.ID, version)
				sendResultToUser()
			})
		}
	}

	if j.IsNotDryRun() {
		reloadMessage := &websocket_pb.WsReloadProjectsResponse{
			Metadata:    &websocket_pb.Metadata{Type: WsReloadProjects},
			NamespaceId: int32(j.ns.ID),
		}
		j.PubSub().ToAll(reloadMessage)
		j.OnFinally(1, func(err error, sendResultToUser func()) {
			// 如果状态出现问题，只有拿到锁的才能更新状态
			j.project, _ = j.projRepo.UpdateDeployStatus(context.TODO(), j.project.ID, j.helmer.ReleaseStatus(j.Project().Name, j.ns.Name))
			j.PubSub().ToAll(reloadMessage)
			sendResultToUser()
		})
	}

	j.imagePullSecrets = j.ns.ImagePullSecrets
	j.commit = &biz.Commit{}
	if j.repo.NeedGitRepo {
		j.commit, err = j.pluginMgr.Git().GetCommit(fmt.Sprintf("%d", j.repo.GitProjectID), j.input.GitCommit)
	}
	if !j.isNew {
		j.oldConf = j.project.ToEventYaml()
	}

	return j.SetError(err)
}

// typeValidated 判断输入部署类型是否属于 Create/Update/Apply 三类。
func (j *jobRunner) typeValidated() bool {
	return j.input.Type == websocket_pb.Type_CreateProject ||
		j.input.Type == websocket_pb.Type_UpdateProject ||
		j.input.Type == websocket_pb.Type_ApplyProject
}

// newLoadContext 组装加载链的显式上下文：填入 Validate 阶段准备好的只读输入与依赖，
// 中间产物留待加载链写入。加载期产生的临时资源由 Loader 通过 AddCleanup 登记，
// 链结束后统一回收；下载/写临时文件所需的基础设施（uploader/timer）也注入其中。
func (j *jobRunner) newLoadContext() *LoadContext {
	return &LoadContext{
		Config:           j.config,
		Input:            j.input,
		Project:          j.project,
		Namespace:        j.ns,
		Repo:             j.repo,
		Commit:           j.commit,
		ImagePullSecrets: j.imagePullSecrets,
		Messager:         j.messager,
		PluginMgr:        j.pluginMgr,
		Helmer:           j.helmer,
		Logger:           j.logger,
		ValuesOptions:    &values.Options{},
		uploader:         j.uploader,
		timer:            j.timer,
	}
}

// LoadConfigs 按默认加载链依次加载 chart/配置，中途检测到停止信号即中断。
// 链无论成败，都会把加载期登记的临时资源清理注册为 finally 回调；
// 链成功后把中间产物收口回 jobRunner 供 Run 阶段消费。
func (j *jobRunner) LoadConfigs() Job {
	if j.HasError() {
		return j
	}
	ctx := j.newLoadContext()
	eg, _ := errgroup.WithContext(j.stopCtx)
	eg.Go(func() error {
		defer j.logger.HandlePanic("LoadConfigs")
		err := func() error {
			j.messager.SendMsg("[Check]: 加载项目文件")

			for _, defaultLoader := range j.loaders {
				if err := j.GetStoppedErrorIfHas(); err != nil {
					return err
				}
				if err := defaultLoader.Load(ctx); err != nil {
					return err
				}
			}

			return nil
		}()
		// 链失败时临时资源尚未回写，仍需注册清理；失败后 Run 不会执行，
		// 但 Finish 的 finally 回调保证下载目录/临时 values 文件被回收。
		for _, cleanup := range ctx.cleanups {
			j.OnFinally(1, func(err error, sendResultToUser func()) {
				sendResultToUser()
				cleanup()
			})
		}
		return err
	})

	if err := eg.Wait(); err != nil {
		return j.SetError(err)
	}
	// 收口中间产物：只有链成功才回写，失败时 Run 已短路不会消费。
	j.chart = ctx.Chart
	j.finalExtraValues = ctx.FinalExtraValues
	j.vars = ctx.Vars
	j.valuesOptions = ctx.ValuesOptions

	return j
}

// Run 执行安装并持久化项目/事件数据：一个 goroutine 消费消息通道，
// 另一个执行 releaseInstaller 安装并汇总更新项目记录。
func (j *jobRunner) Run(ctx context.Context) Job {
	if j.HasError() {
		return j
	}
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer j.logger.HandlePanic("[Websocket]: jobRunner Run")
		defer j.messageCh.Close()
		j.HandleMessage(ctx)
		return nil
	})

	eg.Go(func() error {
		defer j.logger.HandlePanic("[Websocket]: jobRunner Run")
		var (
			result *release.Release
			err    error
		)

		j.messager.SendMsg("worker 已就绪, 准备安装")
		if result, err = j.installer.Run(ctx, &InstallInput{
			IsNew:          j.isNew,
			Wait:           lo.FromPtr(j.input.Atomic),
			Chart:          j.chart,
			ValueOptions:   j.valuesOptions,
			DryRun:         j.dryRun,
			ReleaseName:    j.project.Name,
			Namespace:      j.ns.Name,
			Description:    j.commit.Title,
			TimeoutSeconds: j.timeoutSeconds,
			messageChan:    j.messageCh,
			percenter:      j.messager,
		}); err != nil {
			// 失败帧照常下发给客户端（进度流以 MessageError 收尾），错误本身交由上层
			// services logError 统一打印，避免同一错误在 deploy 层双留痕。
			j.messageCh.Send(MessageItem{
				Msg:  err.Error(),
				Type: MessageError,
			})
			return err
		}

		coalesceValues, _ := chartutil.CoalesceValues(j.chart, result.Config)
		marshal, _ := yaml2.PrettyMarshal(&coalesceValues)
		manifests := j.k8sRepo.SplitManifests(result.Manifest)
		j.manifests = manifests
		var updateProjectInput = &biz.UpdateProjectInput{
			ID:           j.project.ID,
			GitBranch:    j.input.GitBranch,
			GitCommit:    j.input.GitCommit,
			Config:       j.input.Config,
			Atomic:       j.input.Atomic,
			ConfigType:   j.config.GetConfigFileType(),
			PodSelectors: j.k8sRepo.GetPodSelectorsByManifest(manifests),
			DockerImage: matchDockerImage(pipelineVars{
				Pipeline: j.vars.MustGetString("Pipeline"),
				Commit:   j.vars.MustGetString("Commit"),
				Branch:   j.vars.MustGetString("Branch"),
			}, result.Manifest),
			GitCommitTitle:   j.commit.Title,
			GitCommitWebURL:  j.commit.WebURL,
			GitCommitAuthor:  j.commit.AuthorName,
			GitCommitDate:    j.commit.CommittedDate,
			ExtraValues:      j.input.ExtraValues,
			FinalExtraValues: j.finalExtraValues,
			EnvValues:        j.vars.ToKeyValue(),
			OverrideValues:   string(marshal),
			Manifest:         j.manifests,
		}

		var (
			oldConf = j.oldConf
			newConf biz.YamlPrettier
		)

		if j.IsNotDryRun() {
			j.project, err = j.projRepo.UpdateProject(context.TODO(), updateProjectInput)
			if err != nil {
				return err
			}

			newConf = j.project.ToEventYaml()
			j.eventRepo.Dispatch(biz.EventProjectChanged, &biz.ProjectChangedData{
				ID:       j.project.ID,
				Username: j.user.Name,
			})
		}

		var act = types.EventActionType_Create
		if !j.isNew {
			act = types.EventActionType_Update
		}
		if j.dryRun {
			act = types.EventActionType_DryRun
			prettyMarshal, _ := yaml2.PrettyMarshal(j.input)
			newConf = &biz.StringYamlPrettier{Str: string(prettyMarshal)}
		}
		j.eventRepo.AuditLogWithChange(
			act,
			j.user.Name,
			fmt.Sprintf("%s 项目: %s/%s", act.String(), j.ns.Name, j.Project().Name),
			oldConf, newConf)
		j.messager.To(100)
		j.messageCh.Send(MessageItem{
			Msg:  "部署成功",
			Type: MessageSuccess,
		})
		return nil
	})

	return j.SetError(eg.Wait())
}

// Finish 汇总部署结果，按 error/success/finally 顺序触发回调并向用户上报最终状态。
func (j *jobRunner) Finish() Job {
	j.logger.Debug("finished")

	var callbacks []func(err error, next func())

	// Run error hooks
	if j.HasError() {
		func(err error) {
			pmodel := transformer.FromProject(j.project)
			j.deployResult.Set(websocket_pb.ResultType_DeployedFailed, err.Error(), pmodel)

			if e := j.GetStoppedErrorIfHas(); e != nil {
				j.deployResult.Set(websocket_pb.ResultType_DeployedCanceled, e.Error(), pmodel)
				err = e
			}
		}(j.Error())
		callbacks = append(callbacks, j.errorCallback.Sort()...)
	}

	// Run success hooks
	if !j.HasError() {
		callbacks = append(callbacks, j.successCallback.Sort()...)
	}

	// run finally hooks
	callbacks = append(callbacks, j.finallyCallback.Sort()...)

	pipeline.New[error]().
		Send(j.Error()).
		Through(callbacks...).
		Then(func(error) {
			if j.deployResult.IsSet() {
				j.messager.SendDeployedResult(j.deployResult.ResultType(), j.deployResult.Msg(), j.deployResult.Model())
			}
			j.logger.Debug("SendDeployedResult")
		})

	return j
}

// Manifests 返回部署成功后生成的 manifest 列表。
func (j *jobRunner) Manifests() []string {
	return j.manifests
}

// Stop 向用户广播取消信号并以 err 作为取消原因终止部署流水线。
func (j *jobRunner) Stop(err error) {
	j.messager.SendMsg("收到取消信号, 开始停止部署~")
	j.logger.Debugf("stop deploy jobRunner, because '%v'", err)
	j.stopFn(err)
}

// OnError 注册失败回调，p 越大优先级越高（降序，与 priority.Sort 一致）。
func (j *jobRunner) OnError(p int, fn func(err error, sendResultToUser func())) Job {
	j.errorCallback.Add(p, fn)
	return j
}

// OnSuccess 注册成功回调，p 越大优先级越高（降序，与 priority.Sort 一致）。
func (j *jobRunner) OnSuccess(p int, fn func(err error, sendResultToUser func())) Job {
	j.successCallback.Add(p, fn)
	return j
}

// OnFinally 注册无论成败都会执行的回调，p 越大优先级越高（降序，与 priority.Sort 一致）。
func (j *jobRunner) OnFinally(p int, fn func(err error, sendResultToUser func())) Job {
	j.finallyCallback.Add(p, fn)
	return j
}

// Error 返回流水线当前错误（可为 nil）。
func (j *jobRunner) Error() error {
	return j.err
}

// SetError 记录流水线错误并返回自身，支持链式短路。
func (j *jobRunner) SetError(err error) *jobRunner {
	j.err = err
	return j
}

// HasError 判断流水线是否已处于错误状态。
func (j *jobRunner) HasError() bool {
	return j.err != nil
}

// Project 返回当前部署的项目实体。
func (j *jobRunner) Project() *biz.Project {
	return j.project
}

// PubSub 返回输入绑定的消息总线（广播 reload 等事件）。
func (j *jobRunner) PubSub() app.PubSub {
	return j.input.PubSub
}

// IsStopped 判断部署是否已被取消（stopCtx 已关闭）。
func (j *jobRunner) IsStopped() bool {
	select {
	case <-j.stopCtx.Done():
		return true
	default:
	}

	return false
}

// GetStoppedErrorIfHas 返回取消原因；未取消返回 nil，供加载链中断检测。
func (j *jobRunner) GetStoppedErrorIfHas() error {
	if j.IsStopped() {
		return context.Cause(j.stopCtx)
	}
	return nil
}

// HandleMessage 消费消息通道并把部署进度/结果透传给 messager；
// 收到 Error/Success 或 ctx 取消后退出。
func (j *jobRunner) HandleMessage(ctx context.Context) {
	defer j.logger.Debug("HandleMessage exit")
	ch := j.messageCh.Chan()
	for {
		select {
		case <-ctx.Done():
			return
		case s, ok := <-ch:
			if !ok {
				return
			}
			switch s.Type {
			case MessageText:
				j.messager.SendMsgWithContainerLog(s.Msg, s.Containers)
			case MessageError:
				select {
				case <-j.stopCtx.Done():
					j.deployResult.Set(ResultDeployCanceled, context.Cause(j.stopCtx).Error(), transformer.FromProject(j.project))
				default:
					j.deployResult.Set(ResultDeployFailed, s.Msg, transformer.FromProject(j.project))
				}
				return
			case MessageSuccess:
				j.deployResult.Set(ResultDeployed, s.Msg, transformer.FromProject(j.project))
				return
			}
		}
	}
}
