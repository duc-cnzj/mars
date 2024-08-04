package socket

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/locker"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/util"
	mars2 "github.com/duc-cnzj/mars/v4/internal/util/mars"
	"github.com/duc-cnzj/mars/v4/internal/util/pipeline"
	mysort "github.com/duc-cnzj/mars/v4/internal/util/xsort"
	yaml2 "github.com/duc-cnzj/mars/v4/internal/util/yaml"
	"github.com/gosimple/slug"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

var _ JobManager = (*jobManager)(nil)

type JobManager interface {
	//	创建一个新的Job
	NewJob(input *JobInput) Job
}

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
	Manifests() []string

	OnError(p int, fn func(err error, sendResultToUser func())) Job
	OnSuccess(p int, fn func(err error, sendResultToUser func())) Job
	OnFinally(p int, fn func(err error, sendResultToUser func())) Job
}

type jobManager struct {
	logger mlog.Logger

	nsRepo   repo.NamespaceRepo
	projRepo repo.ProjectRepo
	k8sRepo  repo.K8sRepo

	gitServer    application.GitServer
	domainServer application.DomainManager
	wsServer     application.WsSender
	eventRepo    repo.EventRepo

	helmer   repo.HelmerRepo
	locker   locker.Locker
	toolRepo repo.ToolRepo
}

func NewJobManager(
	logger mlog.Logger,
	nsRepo repo.NamespaceRepo,
	projRepo repo.ProjectRepo,
	helmer repo.HelmerRepo,
	locker locker.Locker,
	k8sRepo repo.K8sRepo,
	eventRepo repo.EventRepo,
	toolRepo repo.ToolRepo,
	pl application.PluginManger,
) JobManager {
	return &jobManager{
		logger:       logger,
		nsRepo:       nsRepo,
		projRepo:     projRepo,
		k8sRepo:      k8sRepo,
		gitServer:    pl.Git(),
		domainServer: pl.Domain(),
		wsServer:     pl.Ws(),
		helmer:       helmer,
		locker:       locker,
		toolRepo:     toolRepo,
		eventRepo:    eventRepo,
	}
}

func (j *jobManager) NewJob(input *JobInput) Job {
	jb := &jobRunner{
		logger:          j.logger,
		nsRepo:          j.nsRepo,
		projRepo:        j.projRepo,
		gitServer:       j.gitServer,
		domainServer:    j.domainServer,
		helmer:          j.helmer,
		locker:          j.locker,
		k8sRepo:         j.k8sRepo,
		eventRepo:       j.eventRepo,
		toolRepo:        j.toolRepo,
		deployResult:    DeployResult{},
		loaders:         defaultLoaders(),
		input:           input,
		finallyCallback: mysort.PrioritySort[func(err error, next func())]{},
		errorCallback:   mysort.PrioritySort[func(err error, next func())]{},
		successCallback: mysort.PrioritySort[func(err error, next func())]{},
		vars:            vars{},
		valuesOptions:   &values.Options{},
		messageCh:       &safeWriteMessageCh{ch: make(chan contracts.MessageItem, 100)},
	}
	opts := []Option{
		WithDryRun(input.DryRun),
	}
	jb.stopCtx, jb.stopFn = NewCustomErrorContext()
	for _, opt := range opts {
		opt(jb)
	}

	return jb
}

type CustomErrorContext struct {
	sync.Mutex
	done     chan struct{}
	err      error
	canceled bool
}

func NewCustomErrorContext() (context.Context, func(error)) {
	ctx := &CustomErrorContext{done: make(chan struct{})}
	return ctx, func(err error) {
		ctx.Lock()
		defer ctx.Unlock()
		if ctx.canceled {
			return
		}
		ctx.err = err
		ctx.canceled = true
		close(ctx.done)
	}
}

func (m *CustomErrorContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (m *CustomErrorContext) Done() <-chan struct{} {
	return m.done
}

func (m *CustomErrorContext) Err() error {
	m.Lock()
	defer m.Unlock()
	return m.err
}

func (m *CustomErrorContext) Value(key any) any {
	return nil
}

type JobInput struct {
	Type         websocket_pb.Type
	NamespaceId  int32
	Name         string
	GitProjectId int32
	GitBranch    string
	GitCommit    string
	Config       string
	Atomic       bool
	ExtraValues  []*types.ExtraValue
	Version      *int32

	TimeoutSeconds int32
	User           *auth.UserInfo
	DryRun         bool

	PubSub application.PubSub
}

func (job *JobInput) Slug() string {
	return util.GetSlugName(job.NamespaceId, job.Name)
}

func (job *JobInput) IsNotDryRun() bool {
	return !job.DryRun
}

func (j *jobRunner) ID() string {
	return j.input.Slug()
}

func (j *jobRunner) IsNotDryRun() bool {
	return !j.IsDryRun()
}

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

func (j *jobRunner) Validate() Job {
	var err error
	if j.HasError() {
		return j
	}

	if !j.WsTypeValidated() {
		return j.SetError(errors.New("type error: " + j.input.Type.String()))
	}

	j.Messager().SendMsg("[start]: 收到请求，开始创建项目")
	j.Percenter().To(5)

	j.Messager().SendMsg("[Check]: 校验名称空间...")

	j.ns, err = j.nsRepo.Show(context.TODO(), int(j.input.NamespaceId))
	if err != nil {
		return j.SetError(fmt.Errorf("[FAILED]: 校验名称空间: %w", err))
	}

	j.Messager().SendMsg("[Loading]: 加载用户配置")
	j.Percenter().To(10)

	j.config, err = mars2.GetProjectConfig(j.input.GitProjectId, j.input.GitBranch)
	if err != nil {
		return j.SetError(err)
	}
	if j.input.Name == "" {
		j.input.Name = mars2.GetProjectName(j.input.GitProjectId, j.config)
	}

	createProjectInput := &repo.CreateProjectInput{
		Name:         slug.Make(j.input.Name),
		GitProjectID: int(j.input.GitProjectId),
		GitBranch:    j.input.GitBranch,
		GitCommit:    j.input.GitCommit,
		Config:       j.input.Config,
		Atomic:       j.input.Atomic,
		ConfigType:   j.config.ConfigFileType,
		NamespaceID:  j.ns.ID,
	}

	j.Messager().SendMsg("[Check]: 检查项目是否存在")

	found, err := j.projRepo.FindByName(context.TODO(), createProjectInput.Name, createProjectInput.NamespaceID)
	if err != nil {
		j.Messager().SendMsg("[Check]: 新建项目")
		createProjectInput.DeployStatus = types.Deploy_StatusDeploying
		j.isNew = true
		if j.IsNotDryRun() {
			j.project, _ = j.projRepo.Create(context.TODO(), createProjectInput)
			j.OnError(1, func(err error, sendResultToUser func()) {
				j.logger.Debug("清理项目")
				j.projRepo.Delete(context.TODO(), j.project.ID)
				sendResultToUser()
			})
		}
	} else {
		j.prevProject = found
		j.ns = found.Edges.Namespace
		if j.IsNotDryRun() {
			j.Messager().SendMsg("[Check]: 检查当前版本")
			j.project, err = j.projRepo.UpdateStatusByVersion(context.TODO(), j.project.ID, types.Deploy_StatusDeploying, j.project.Version+1)
			if err != nil {
				return j.SetError(ErrorVersionNotMatched)
			}
			j.OnError(1, func(err error, sendResultToUser func()) {
				j.project, _ = j.projRepo.UpdateVersion(context.TODO(), j.project.ID, j.prevProject.Version)
				sendResultToUser()
			})
		}
	}

	if j.IsNotDryRun() {
		j.PubSub().ToAll(reloadProjectsMessage(j.ns.ID))
		j.OnFinally(1, func(err error, sendResultToUser func()) {
			// 如果状态出现问题，只有拿到锁的才能更新状态
			j.project, _ = j.projRepo.UpdateDeployStatus(context.TODO(), j.project.ID, types.Deploy_StatusFailed)
			j.PubSub().ToAll(reloadProjectsMessage(j.Namespace().ID))
			sendResultToUser()
		})
	}
	j.imagePullSecrets = j.Namespace().ImagePullSecrets
	j.commit, err = j.gitServer.GetCommit(fmt.Sprintf("%d", j.project.GitProjectID), j.project.GitCommit)

	return j.SetError(err)
}

func (j *jobRunner) WsTypeValidated() bool {
	return j.input.Type == websocket_pb.Type_CreateProject || j.input.Type == websocket_pb.Type_UpdateProject || j.input.Type == websocket_pb.Type_ApplyProject
}

func (j *jobRunner) LoadConfigs() Job {
	if j.HasError() {
		return j
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	ch := make(chan error, 1)
	go func() {
		defer j.logger.HandlePanic("LoadConfigs")
		defer wg.Done()
		defer close(ch)
		err := func() error {
			j.Messager().SendMsg("[Check]: 加载项目文件")

			for _, defaultLoader := range j.loaders {
				if err := j.GetStoppedErrorIfHas(); err != nil {
					return err
				}
				if err := defaultLoader.Load(j); err != nil {
					return err
				}
			}

			return nil
		}()
		ch <- err
	}()

	var err error
	select {
	case err = <-ch:
	case <-j.stopCtx.Done():
		err = j.stopCtx.Err()
	}
	wg.Wait()

	return j.SetError(err)
}

func (j *jobRunner) Run(ctx context.Context) Job {
	if j.HasError() {
		return j
	}
	done := make(chan struct{}, 1)
	go func() {
		defer func() {
			done <- struct{}{}
			close(done)
		}()
		defer j.logger.HandlePanic("[Websocket]: jobRunner Run")
		j.HandleMessage(ctx)
	}()

	err := func() error {
		var (
			result *release.Release
			err    error
		)

		if result, err = j.ReleaseInstaller().Run(ctx, j.messageCh, j.Percenter(), j.IsNew(), j.Commit().GetTitle()); err != nil {
			j.logger.Errorf("[Websocket]: %v", err)
			j.messageCh.Send(contracts.MessageItem{
				Msg:  err.Error(),
				Type: contracts.MessageError,
			})
		} else {
			coalesceValues, _ := chartutil.CoalesceValues(j.ReleaseInstaller().Chart(), result.Config)
			marshal, _ := yaml2.PrettyMarshal(&coalesceValues)
			var updateProjectInput = &repo.UpdateProjectInput{
				ID:           j.project.ID,
				GitBranch:    j.input.GitBranch,
				GitCommit:    j.input.GitCommit,
				Config:       j.input.Config,
				Atomic:       j.input.Atomic,
				ConfigType:   j.config.GetConfigFileType(),
				PodSelectors: j.k8sRepo.GetPodSelectorsByManifest(j.manifests),
				Manifest:     util.SplitManifests(result.Manifest),
				DockerImage: matchDockerImage(pipelineVars{
					Pipeline: j.vars.MustGetString("Pipeline"),
					Commit:   j.vars.MustGetString("Commit"),
					Branch:   j.vars.MustGetString("Branch"),
				}, result.Manifest),
				GitCommitTitle:   j.Commit().GetTitle(),
				GitCommitWebURL:  j.Commit().GetWebURL(),
				GitCommitAuthor:  j.Commit().GetAuthorName(),
				GitCommitDate:    j.Commit().GetCommittedDate(),
				ExtraValues:      j.input.ExtraValues,
				FinalExtraValues: j.extraValues,
				EnvValues:        j.vars.ToKeyValue(),
				OverrideValues:   string(marshal),
			}

			var oldConf, newConf *userConfig

			if j.prevProject != nil && j.prevProject.ID > 0 {
				oldConf = newUserConfig(j.prevProject)
			}
			if j.IsNotDryRun() {
				j.project, _ = j.projRepo.UpdateProject(context.TODO(), updateProjectInput)
				newConf = newUserConfig(j.project)
				j.eventRepo.Dispatch(repo.EventProjectChanged, &repo.ProjectChangedData{
					Project:  j.project,
					Username: j.User().Name,
				})
			}

			var act types.EventActionType = types.EventActionType_Create
			if !j.IsNew() {
				act = types.EventActionType_Update
			}
			if j.IsDryRun() {
				act = types.EventActionType_DryRun
			}
			j.eventRepo.AuditLogWithChange(act, j.User().Name,
				fmt.Sprintf("%s 项目: %s/%s", act.String(), j.Namespace().Name, j.Project().Name),
				oldConf, newConf)
			j.Percenter().To(100)
			j.messageCh.Send(contracts.MessageItem{
				Msg:  "部署成功",
				Type: contracts.MessageSuccess,
			})
		}
		return err
	}()
	<-done
	return j.SetError(err)
}

func (j *jobRunner) Finish() Job {
	j.logger.Debug("finished")

	var callbacks []func(err error, next func())

	// Run error hooks
	if j.HasError() {
		func(err error) {
			j.SetDeployResult(websocket_pb.ResultType_DeployedFailed, err.Error(), j.ProjectModel())

			if e := j.GetStoppedErrorIfHas(); e != nil {
				j.SetDeployResult(websocket_pb.ResultType_DeployedCanceled, e.Error(), j.ProjectModel())
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

	pipeline.NewPipeline[error]().
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

func (j *jobRunner) Manifests() []string {
	return j.manifests
}

func (j *jobRunner) Stop(err error) {
	j.messager.SendMsg("收到取消信号, 开始停止部署~")
	j.logger.Debugf("stop deploy jobRunner, because '%v'", err)
	j.stopFn(err)
}

func (j *jobRunner) OnError(p int, fn func(err error, sendResultToUser func())) Job {
	j.errorCallback.Add(p, fn)
	return j
}

func (j *jobRunner) OnSuccess(p int, fn func(err error, sendResultToUser func())) Job {
	j.successCallback.Add(p, fn)
	return j
}

func (j *jobRunner) OnFinally(p int, fn func(err error, sendResultToUser func())) Job {
	j.finallyCallback.Add(p, fn)
	return j
}

func (j *jobRunner) Error() error {
	return j.err
}
