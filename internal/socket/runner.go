package socket

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/locker"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util"
	"github.com/duc-cnzj/mars/v4/internal/util/pipeline"
	"github.com/duc-cnzj/mars/v4/internal/util/rand"
	"github.com/duc-cnzj/mars/v4/internal/util/timer"
	mysort "github.com/duc-cnzj/mars/v4/internal/util/xsort"
	yaml2 "github.com/duc-cnzj/mars/v4/internal/util/yaml"
	"github.com/samber/lo"
	"go.uber.org/config"
	"golang.org/x/sync/errgroup"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

var ErrorVersionNotMatched = errors.New("[部署冲突]: 1. 可能是多个人同时部署导致 2. 项目已经存在")

const (
	ResultError             = websocket_pb.ResultType_Error
	ResultSuccess           = websocket_pb.ResultType_Success
	ResultDeployed          = websocket_pb.ResultType_Deployed
	ResultDeployFailed      = websocket_pb.ResultType_DeployedFailed
	ResultDeployCanceled    = websocket_pb.ResultType_DeployedCanceled
	ResultLogWithContainers = websocket_pb.ResultType_LogWithContainers

	WsSetUid             = websocket_pb.Type_SetUid
	WsReloadProjects     = websocket_pb.Type_ReloadProjects
	WsCancel             = websocket_pb.Type_CancelProject
	WsCreateProject      = websocket_pb.Type_CreateProject
	WsUpdateProject      = websocket_pb.Type_UpdateProject
	WsProcessPercent     = websocket_pb.Type_ProcessPercent
	WsClusterInfoSync    = websocket_pb.Type_ClusterInfoSync
	WsInternalError      = websocket_pb.Type_InternalError
	WsHandleExecShell    = websocket_pb.Type_HandleExecShell
	WsHandleExecShellMsg = websocket_pb.Type_HandleExecShellMsg
	WsHandleCloseShell   = websocket_pb.Type_HandleCloseShell
	WsAuthorize          = websocket_pb.Type_HandleAuthorize
	ProjectPodEvent      = websocket_pb.Type_ProjectPodEvent

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 1024 * 20 // 20MB
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 8) / 10
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
	Project() *repo.Project
	Manifests() []string

	OnError(p int, fn func(err error, sendResultToUser func())) Job
	OnSuccess(p int, fn func(err error, sendResultToUser func())) Job
	OnFinally(p int, fn func(err error, sendResultToUser func())) Job
}

type jobManager struct {
	logger mlog.Logger
	data   data.Data

	timer            timer.Timer
	releaseInstaller ReleaseInstaller
	nsRepo           repo.NamespaceRepo
	projRepo         repo.ProjectRepo
	eventRepo        repo.EventRepo
	k8sRepo          repo.K8sRepo
	helmRepo         repo.HelmerRepo
	toolRepo         repo.ToolRepo
	repoRepo         repo.RepoImp

	locker       locker.Locker
	uploader     uploader.Uploader
	pluginManger application.PluginManger
}

func NewJobManager(
	data data.Data,
	timer timer.Timer,
	logger mlog.Logger,
	releaseInstaller ReleaseInstaller,
	repoRepo repo.RepoImp,
	nsRepo repo.NamespaceRepo,
	projRepo repo.ProjectRepo,
	helmer repo.HelmerRepo,
	uploader uploader.Uploader,
	locker locker.Locker,
	k8sRepo repo.K8sRepo,
	eventRepo repo.EventRepo,
	toolRepo repo.ToolRepo,
	pl application.PluginManger,
) JobManager {
	return &jobManager{
		timer:            timer,
		releaseInstaller: releaseInstaller,
		uploader:         uploader,
		repoRepo:         repoRepo,
		data:             data,
		logger:           logger,
		nsRepo:           nsRepo,
		projRepo:         projRepo,
		k8sRepo:          k8sRepo,
		pluginManger:     pl,
		helmRepo:         helmer,
		locker:           locker,
		toolRepo:         toolRepo,
		eventRepo:        eventRepo,
	}
}

func (j *jobManager) NewJob(input *JobInput) Job {
	var timeoutSeconds int64 = int64(input.TimeoutSeconds)
	if timeoutSeconds == 0 {
		timeoutSeconds = int64(j.data.Config().InstallTimeout.Seconds())
	}
	jb := &jobRunner{
		installer:       j.releaseInstaller,
		logger:          j.logger,
		nsRepo:          j.nsRepo,
		projRepo:        j.projRepo,
		repoRepo:        j.repoRepo,
		pluginMgr:       j.pluginManger,
		helmer:          j.helmRepo,
		locker:          j.locker,
		k8sRepo:         j.k8sRepo,
		eventRepo:       j.eventRepo,
		toolRepo:        j.toolRepo,
		timer:           j.timer,
		uploader:        j.uploader,
		loaders:         defaultLoaders(),
		dryRun:          input.DryRun,
		input:           input,
		finallyCallback: mysort.PrioritySort[func(err error, next func())]{},
		errorCallback:   mysort.PrioritySort[func(err error, next func())]{},
		successCallback: mysort.PrioritySort[func(err error, next func())]{},
		deployResult:    &DeployResult{},
		valuesOptions:   &values.Options{},
		messageCh:       NewSafeWriteMessageCh(j.logger, 100),
		messager:        input.Messager,
		user:            input.User,
		timeoutSeconds:  timeoutSeconds,
	}
	jb.stopCtx, jb.stopFn = context.WithCancelCause(context.TODO())

	return jb
}

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
	User           *auth.UserInfo
	DryRun         bool

	PubSub   application.PubSub `json:"-"`
	Messager DeployMsger        `json:"-"`
}

func (job *JobInput) Slug() string {
	return util.GetSlugName(job.NamespaceId, job.Name)
}

func (job *JobInput) IsNotDryRun() bool {
	return !job.DryRun
}

type jobRunner struct {
	// 这些属性在 new runner 的时候就已经初始化了
	logger          mlog.Logger
	nsRepo          repo.NamespaceRepo
	projRepo        repo.ProjectRepo
	repoRepo        repo.RepoImp
	helmer          repo.HelmerRepo
	locker          locker.Locker
	k8sRepo         repo.K8sRepo
	eventRepo       repo.EventRepo
	messager        DeployMsger
	timeoutSeconds  int64
	toolRepo        repo.ToolRepo
	uploader        uploader.Uploader
	pluginMgr       application.PluginManger
	installer       ReleaseInstaller
	messageCh       SafeWriteMessageChan
	deployResult    *DeployResult
	loaders         []Loader
	input           *JobInput
	finallyCallback mysort.PrioritySort[func(err error, next func())]
	errorCallback   mysort.PrioritySort[func(err error, next func())]
	successCallback mysort.PrioritySort[func(err error, next func())]
	stopCtx         context.Context
	stopFn          func(error)
	dryRun          bool
	user            *auth.UserInfo
	timer           timer.Timer

	// 这些属性在执行的时候才会初始化
	// Validate 阶段被初始化
	isNew            bool
	ns               *repo.Namespace
	repo             *repo.Repo
	config           *mars.Config
	project          *repo.Project
	imagePullSecrets []string
	commit           application.Commit
	oldConf          repo.YamlPrettier

	// LoadConfigs 阶段被初始化

	// 1. ChartFileLoader 时加载
	chart *chart.Chart

	// 2. UserConfigLoader 时加载
	userConfigYaml string

	// 3. ElementsLoader(自定义配置) 时加载
	elementValues []string

	// 4. SystemVariableLoader 时加载
	// chart 的 替换完所有 <.Var> 之后的 values.yaml 内容
	systemValuesYaml string
	// systemValuesYaml 注入的变量
	vars vars

	// 5. MergeValuesLoader 时加载
	// 把 values.yaml + elementValues + 自定义配置 合并后的结果
	valuesOptions *values.Options

	err error

	// 部署成功后的 manifest
	manifests []string
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

	if !j.typeValidated() {
		return j.SetError(errors.New("type error: " + j.input.Type.String()))
	}

	j.Messager().SendMsg("[start]: 收到请求，开始创建项目")
	j.Messager().To(5)

	j.Messager().SendMsg("[Check]: 校验名称空间...")

	j.ns, err = j.nsRepo.Show(context.TODO(), int(j.input.NamespaceId))
	if err != nil {
		return j.SetError(fmt.Errorf("[FAILED]: 校验名称空间: %w", err))
	}

	j.Messager().SendMsg("[Loading]: 加载用户配置")
	j.Messager().To(10)

	j.repo, err = j.repoRepo.Show(context.TODO(), int(j.input.RepoID))
	if err != nil {
		return j.SetError(err)
	}
	j.config = j.repo.MarsConfig

	createProjectInput := &repo.CreateProjectInput{
		Name:         j.input.Name,
		GitProjectID: int(j.repo.GitProjectID),
		GitBranch:    j.input.GitBranch,
		GitCommit:    j.input.GitCommit,
		Config:       j.input.Config,
		Atomic:       j.input.Atomic,
		ConfigType:   j.config.ConfigFileType,
		NamespaceID:  j.ns.ID,
		RepoID:       j.repo.ID,
	}

	j.Messager().SendMsg("[Check]: 检查项目是否存在")

	found, err := j.projRepo.FindByName(context.TODO(), createProjectInput.Name, createProjectInput.NamespaceID)
	if err != nil {
		j.Messager().SendMsg("[Check]: 新建项目")
		createProjectInput.DeployStatus = types.Deploy_StatusDeploying
		j.isNew = true
		if j.IsNotDryRun() {
			j.project, err = j.projRepo.Create(context.TODO(), createProjectInput)
			if err != nil {
				j.logger.Warning(err)
				return j.SetError(err)
			}
			createdID := j.project.ID
			j.OnError(1, func(err error, sendResultToUser func()) {
				j.logger.Debug("清理项目")
				j.projRepo.Delete(context.TODO(), createdID)
				sendResultToUser()
			})
		}
	} else {
		j.project = found
		version := j.project.Version
		if j.IsNotDryRun() {
			j.Messager().SendMsg(fmt.Sprintf("[Check]: 检查当前版本, version: %v", lo.FromPtr(j.input.Version)))
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
			NamespaceId: int32(j.Namespace().ID),
		}
		j.PubSub().ToAll(reloadMessage)
		j.OnFinally(1, func(err error, sendResultToUser func()) {
			// 如果状态出现问题，只有拿到锁的才能更新状态
			j.project, _ = j.projRepo.UpdateDeployStatus(context.TODO(), j.project.ID, j.helmer.ReleaseStatus(j.Project().Name, j.Namespace().Name))
			j.PubSub().ToAll(reloadMessage)
			sendResultToUser()
		})
	}

	j.imagePullSecrets = j.Namespace().ImagePullSecrets
	if j.repo.NeedGitRepo {
		j.commit, err = j.pluginMgr.Git().GetCommit(fmt.Sprintf("%d", j.project.GitProjectID), j.project.GitCommit)
	} else {
		j.commit = application.NewEmptyCommit()
	}
	j.oldConf = toProjectEventYaml(j.project)

	return j.SetError(err)
}

func (j *jobRunner) typeValidated() bool {
	return j.input.Type == websocket_pb.Type_CreateProject ||
		j.input.Type == websocket_pb.Type_UpdateProject ||
		j.input.Type == websocket_pb.Type_ApplyProject
}

func (j *jobRunner) LoadConfigs() Job {
	if j.HasError() {
		return j
	}
	eg, _ := errgroup.WithContext(j.stopCtx)
	eg.Go(func() error {
		defer j.logger.HandlePanic("LoadConfigs")
		return func() error {
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
	})

	return j.SetError(eg.Wait())
}

func (j *jobRunner) Run(ctx context.Context) Job {
	if j.HasError() {
		return j
	}
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer j.logger.HandlePanic("[Websocket]: jobRunner Run")
		j.HandleMessage(ctx)
		return nil
	})

	eg.Go(func() error {
		var (
			result *release.Release
			err    error
		)

		j.Messager().SendMsg("worker 已就绪, 准备安装")
		if result, err = j.installer.Run(ctx, &InstallInput{
			IsNew:        j.IsNew(),
			Wait:         lo.FromPtr(j.input.Atomic),
			Chart:        j.chart,
			ValueOptions: j.valuesOptions,
			DryRun:       j.IsDryRun(),
			ReleaseName:  j.project.Name,
			Namespace:    j.Namespace().Name,
			Description:  j.Commit().GetTitle(),
			messageChan:  NewSafeWriteMessageCh(j.logger, 100),
			percenter:    j.messager,
		}); err != nil {
			j.logger.Errorf("[Websocket]: %v", err)
			j.messageCh.Send(MessageItem{
				Msg:  err.Error(),
				Type: MessageError,
			})
			return err
		}

		coalesceValues, _ := chartutil.CoalesceValues(j.chart, result.Config)
		marshal, _ := yaml2.PrettyMarshal(&coalesceValues)
		manifests := util.SplitManifests(result.Manifest)
		j.manifests = manifests
		var updateProjectInput = &repo.UpdateProjectInput{
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
			GitCommitTitle:   j.Commit().GetTitle(),
			GitCommitWebURL:  j.Commit().GetWebURL(),
			GitCommitAuthor:  j.Commit().GetAuthorName(),
			GitCommitDate:    j.Commit().GetCommittedDate(),
			ExtraValues:      j.input.ExtraValues,
			FinalExtraValues: j.elementValues,
			EnvValues:        j.vars.ToKeyValue(),
			OverrideValues:   string(marshal),
			Manifest:         j.manifests,
		}

		var (
			oldConf repo.YamlPrettier = j.oldConf
			newConf repo.YamlPrettier
		)

		if j.IsNotDryRun() {
			j.project, err = j.projRepo.UpdateProject(context.TODO(), updateProjectInput)
			if err != nil {
				j.logger.Warning(err)
				return err
			}

			newConf = toProjectEventYaml(j.project)
			j.eventRepo.Dispatch(repo.EventProjectChanged, &repo.ProjectChangedData{
				ID:       j.project.ID,
				Username: j.User().Name,
			})
		}

		var act types.EventActionType = types.EventActionType_Create
		if !j.IsNew() {
			act = types.EventActionType_Update
		}
		if j.IsDryRun() {
			act = types.EventActionType_DryRun
			prettyMarshal, _ := yaml2.PrettyMarshal(j.input)
			newConf = &repo.StringYamlPrettier{Str: string(prettyMarshal)}
		}
		j.eventRepo.AuditLogWithChange(act, j.User().Name,
			fmt.Sprintf("%s 项目: %s/%s", act.String(), j.Namespace().Name, j.Project().Name),
			oldConf, newConf)
		j.Messager().To(100)
		j.messageCh.Send(MessageItem{
			Msg:  "部署成功",
			Type: MessageSuccess,
		})
		return nil
	})

	return j.SetError(eg.Wait())
}

func (j *jobRunner) Finish() Job {
	j.logger.Debug("finished")

	var callbacks []func(err error, next func())

	// Run error hooks
	if j.HasError() {
		func(err error) {
			j.deployResult.Set(websocket_pb.ResultType_DeployedFailed, err.Error(), j.ProjectModel())

			if e := j.GetStoppedErrorIfHas(); e != nil {
				j.deployResult.Set(websocket_pb.ResultType_DeployedCanceled, e.Error(), j.ProjectModel())
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

func (j *jobRunner) SetError(err error) *jobRunner {
	j.err = err
	return j
}

func (j *jobRunner) HasError() bool {
	return j.err != nil
}

func (j *jobRunner) IsNew() bool {
	return j.isNew
}

func (j *jobRunner) IsDryRun() bool {
	return j.dryRun
}

func (j *jobRunner) Commit() application.Commit {
	return j.commit
}

func (j *jobRunner) User() *auth.UserInfo {
	return j.user
}

func (j *jobRunner) ProjectModel() *types.ProjectModel {
	if j.project == nil {
		return nil
	}
	return transformer.FromProject(j.project)
}

func (j *jobRunner) Project() *repo.Project {
	return j.project
}

func (j *jobRunner) Namespace() *repo.Namespace {
	return j.ns
}

func (j *jobRunner) PubSub() application.PubSub {
	return j.input.PubSub
}

func (j *jobRunner) IsStopped() bool {
	select {
	case <-j.stopCtx.Done():
		return true
	default:
	}

	return false
}
func (j *jobRunner) GetStoppedErrorIfHas() error {
	if j.IsStopped() {
		return context.Cause(j.stopCtx)
	}
	return nil
}

func (j *jobRunner) WriteConfigYamlToTmpFile(data []byte) (string, io.Closer, error) {
	file := fmt.Sprintf("mars-%s-%s.yaml", j.timer.Now().Format("2006-01-02"), rand.String(20))
	info, err := j.uploader.LocalUploader().Put(file, bytes.NewReader(data))
	if err != nil {
		return "", nil, err
	}
	path := info.Path()

	return path, util.NewCloser(func() error {
		j.logger.Debug("delete file: " + path)
		if err := j.uploader.LocalUploader().Delete(path); err != nil {
			j.logger.Error("WriteConfigYamlToTmpFile error: ", err)
			return err
		}

		return nil
	}), nil
}

func (j *jobRunner) DownloadFiles(pid any, commit string, files []string) (string, func(), error) {
	id := fmt.Sprintf("%v", pid)
	dir := fmt.Sprintf("mars_tmp_%s", rand.String(10))
	if err := j.uploader.LocalUploader().MkDir(dir, false); err != nil {
		return "", nil, err
	}

	return j.DownloadFilesToDir(id, commit, files, j.uploader.LocalUploader().AbsolutePath(dir))
}

func (j *jobRunner) DownloadFilesToDir(pid any, commit string, files []string, dir string) (string, func(), error) {
	wg := &sync.WaitGroup{}
	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			defer j.logger.HandlePanic("DownloadFilesToDir")
			raw, err := j.pluginMgr.Git().GetFileContentWithSha(fmt.Sprintf("%v", pid), commit, file)
			if err != nil {
				j.logger.Error(err)
			}
			localPath := filepath.Join(dir, file)
			if _, err := j.uploader.LocalUploader().Put(localPath, strings.NewReader(raw)); err != nil {
				j.logger.Errorf("[DownloadFilesToDir]: err '%s'", err.Error())
			}
		}(file)
	}
	wg.Wait()

	return dir, func() {
		err := j.uploader.LocalUploader().DeleteDir(dir)
		if err != nil {
			j.logger.Warning(err)
			return
		}
		j.logger.Debug("remove " + dir)
	}, nil
}

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
				j.Messager().SendMsgWithContainerLog(s.Msg, s.Containers)
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

func (j *jobRunner) Messager() DeployMsger {
	return j.messager
}

func toProjectEventYaml(p *repo.Project) repo.YamlPrettier {
	if p == nil {
		return nil
	}

	var finalExtraValues string
	var opts []config.YAMLOption
	for _, item := range p.FinalExtraValues {
		opts = append(opts, config.Source(strings.NewReader(item)))
	}
	if len(opts) != 0 {
		provider, _ := config.NewYAML(opts...)
		var merged map[string]any
		provider.Get("").Populate(&merged)

		out, _ := yaml2.PrettyMarshal(&merged)
		finalExtraValues = string(out)
	}

	out, _ := yaml2.PrettyMarshal(map[string]any{
		"title":   p.GitCommitTitle,
		"branch":  p.GitBranch,
		"commit":  p.GitCommit,
		"atomic":  p.Atomic,
		"web_url": p.GitCommitWebURL,
		"config":  p.Config,
		"env_values": lo.IsSortedByKey(
			p.EnvValues,
			func(item *types.KeyValue) string {
				return item.Key
			}),
		"extra_values": lo.IsSortedByKey(
			p.ExtraValues,
			func(item *websocket_pb.ExtraValue) string {
				return item.Path
			}),
		"final_extra_values": finalExtraValues,
	})

	return &repo.StringYamlPrettier{Str: string(out)}
}

type DeployResult struct {
	sync.RWMutex
	result websocket_pb.ResultType
	msg    string
	model  *types.ProjectModel
	set    bool
}

func (d *DeployResult) IsSet() bool {
	d.RLock()
	defer d.RUnlock()
	return d.set
}

func (d *DeployResult) Msg() string {
	d.RLock()
	defer d.RUnlock()
	return d.msg
}

func (d *DeployResult) Model() *types.ProjectModel {
	d.RLock()
	defer d.RUnlock()
	return d.model
}

func (d *DeployResult) ResultType() websocket_pb.ResultType {
	d.RLock()
	defer d.RUnlock()
	return d.result
}

func (d *DeployResult) Set(t websocket_pb.ResultType, msg string, model *types.ProjectModel) {
	d.Lock()
	defer d.Unlock()
	d.result = t
	d.msg = msg
	d.model = model
	d.set = true
}

type vars map[string]string

func (v vars) ToKeyValue() (res []*types.KeyValue) {
	for k, va := range v {
		res = append(res, &types.KeyValue{
			Key:   k,
			Value: va,
		})
	}
	return
}

func (v vars) MustGetString(key string) string {
	if value, ok := v[key]; ok {
		return value
	}

	return ""
}

func (v vars) Add(key, value string) {
	v[key] = value
}

type pipelineVars struct {
	Pipeline string
	Commit   string
	Branch   string
}

var matchTag = regexp.MustCompile(`image:\s+(\S+)`)

func matchDockerImage(v pipelineVars, manifest string) []string {
	var (
		candidateImages []string
		all             []string
		existsMap       = make(map[string]struct{})
	)
	submatch := matchTag.FindAllStringSubmatch(manifest, -1)
	for _, matches := range submatch {
		if len(matches) == 2 {
			image := strings.Trim(matches[1], "\"")

			if _, ok := existsMap[image]; ok {
				continue
			}
			existsMap[image] = struct{}{}
			all = append(all, image)
			if imageUsedPipelineVars(v, image) {
				candidateImages = append(candidateImages, image)
			}
		}
	}
	// 如果找到至少一个镜像就直接返回，如果未找到，则返回所有匹配到的镜像
	if len(candidateImages) > 0 {
		return candidateImages
	}

	return all
}

// imageUsedPipelineVars 使用的流水线变量的镜像，都把他当成是我们的目标镜像
func imageUsedPipelineVars(v pipelineVars, s string) bool {
	var pipelineVarsSlice []string
	if v.Pipeline != "" {
		pipelineVarsSlice = append(pipelineVarsSlice, v.Pipeline)
	}
	if v.Commit != "" {
		pipelineVarsSlice = append(pipelineVarsSlice, v.Commit)
	}
	if v.Branch != "" {
		pipelineVarsSlice = append(pipelineVarsSlice, v.Branch)
	}
	for _, pvar := range pipelineVarsSlice {
		if strings.Contains(s, pvar) {
			return true
		}
	}

	return false
}
