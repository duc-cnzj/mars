package socket

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gosimple/slug"
	"go.uber.org/config"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"

	"github.com/duc-cnzj/mars/client/event"
	"github.com/duc-cnzj/mars/client/mars"
	websocket_pb "github.com/duc-cnzj/mars/client/websocket"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
)

const (
	ResultError          = websocket_pb.ResultType_Error
	ResultSuccess        = websocket_pb.ResultType_Success
	ResultDeployed       = websocket_pb.ResultType_Deployed
	ResultDeployFailed   = websocket_pb.ResultType_DeployedFailed
	ResultDeployCanceled = websocket_pb.ResultType_DeployedCanceled

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

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 5
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	hostMatch = regexp.MustCompile(`.*?=(.*?){{\s*.Host\d\s*}}`)
	tagRegex  = regexp.MustCompile(`{{\s*(\.Branch|\.Commit|\.Pipeline)\s*}}`)

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type WsResponse = websocket_pb.WsMetadataResponse

type CancelSignaler interface {
	Remove(id string)
	Has(id string) bool
	Cancel(id string)
	Add(id string, fn func(error)) error
	CancelAll()
}

type CancelSignals struct {
	cs map[string]func(error)
	sync.RWMutex
}

func (cs *CancelSignals) Remove(id string) {
	cs.Lock()
	defer cs.Unlock()
	delete(cs.cs, id)
}

func (cs *CancelSignals) Has(id string) bool {
	cs.RLock()
	defer cs.RUnlock()

	_, ok := cs.cs[id]

	return ok
}

func (cs *CancelSignals) Cancel(id string) {
	cs.Lock()
	defer cs.Unlock()
	if fn, ok := cs.cs[id]; ok {
		fn(errors.New("收到取消信号，开始停止部署！！！"))
	}
}

func (cs *CancelSignals) Add(id string, fn func(error)) error {
	cs.Lock()
	defer cs.Unlock()
	if _, ok := cs.cs[id]; ok {
		return errors.New("项目已经存在")
	}
	cs.cs[id] = fn
	return nil
}

func (cs *CancelSignals) CancelAll() {
	cs.Lock()
	defer cs.Unlock()
	for _, f := range cs.cs {
		f(errors.New("收到取消信号，开始停止部署！！！"))
	}
}

type MessageType uint8

const (
	_ MessageType = iota
	MessageSuccess
	MessageError
	MessageText
)

type MessageItem struct {
	Msg  string
	Type MessageType
}

type ReleaseInstaller interface {
	Chart() *chart.Chart
	Run(stopCtx context.Context, messageCh chan MessageItem) (*release.Release, error)
}

type releaseInstaller struct {
	chart       *chart.Chart
	releaseName string
	namespace   string
	atomic      bool
	valueOpts   *values.Options
	logger      func(format string, v ...interface{})
}

func newReleaseInstaller(releaseName, namespace string, chart *chart.Chart, valueOpts *values.Options, logger func(format string, v ...interface{}), atomic bool) *releaseInstaller {
	return &releaseInstaller{chart: chart, valueOpts: valueOpts, logger: logger, releaseName: releaseName, atomic: atomic, namespace: namespace}
}

func (r *releaseInstaller) Chart() *chart.Chart {
	return r.chart
}

func (r *releaseInstaller) Run(stopCtx context.Context, messageCh chan MessageItem) (*release.Release, error) {
	defer utils.HandlePanic("releaseInstaller: Run")

	return utils.UpgradeOrInstall(stopCtx, r.releaseName, r.namespace, r.chart, r.valueOpts, r.logger, r.atomic)
}

type ProjectManager interface {
	Get() *models.Project
	IsNew() bool
	Delete() error
	Save() error
}

type Msger interface {
	SendEndError(error)
	SendError(error)
	SendMsg(string)
	SendProtoMsg(plugins.WebsocketMessage)
}

type DeployMsger interface {
	Msger
	ProcessPercentMsger
	SendDeployedResult(websocket_pb.ResultType, string, *models.Project)
}

type ProcessPercentMsger interface {
	SendProcessPercent(string)
}

type MessageSender struct {
	conn     *WsConn
	slugName string
	wsType   websocket_pb.Type
}

func (ms *MessageSender) SendDeployedResult(result websocket_pb.ResultType, msg string, project *models.Project) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:   ms.slugName,
			Type:   ms.wsType,
			Result: result,
			End:    true,
			Uid:    ms.conn.uid,
			Id:     ms.conn.id,
			Data:   msg,
		},
	}
	ms.send(res)
}

func NewMessageSender(conn *WsConn, slugName string, wsType websocket_pb.Type) DeployMsger {
	return &MessageSender{conn: conn, slugName: slugName, wsType: wsType}
}

func (ms *MessageSender) SendEndError(err error) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:   ms.slugName,
			Type:   ms.wsType,
			Result: ResultError,
			End:    true,
			Uid:    ms.conn.uid,
			Id:     ms.conn.id,
			Data:   err.Error(),
		},
	}
	ms.send(res)
}

func (ms *MessageSender) SendError(err error) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:   ms.slugName,
			Type:   ms.wsType,
			Result: ResultError,
			End:    false,
			Uid:    ms.conn.uid,
			Id:     ms.conn.id,
			Data:   err.Error(),
		},
	}
	ms.send(res)
}

func (ms *MessageSender) SendProcessPercent(percent string) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:   ms.slugName,
			Type:   WsProcessPercent,
			Result: ResultSuccess,
			End:    false,
			Uid:    ms.conn.uid,
			Id:     ms.conn.id,
			Data:   percent,
		},
	}
	ms.send(res)
}

func (ms *MessageSender) SendMsg(msg string) {
	res := &WsResponse{
		Metadata: &websocket_pb.Metadata{
			Slug:   ms.slugName,
			Type:   ms.wsType,
			Result: ResultSuccess,
			End:    false,
			Uid:    ms.conn.uid,
			Id:     ms.conn.id,
			Data:   msg,
		},
	}
	ms.send(res)
}

func (ms *MessageSender) SendProtoMsg(msg plugins.WebsocketMessage) {
	ms.send(msg)
}

func (ms *MessageSender) send(res plugins.WebsocketMessage) {
	ms.conn.pubSub.ToSelf(res)
}

type Percentable interface {
	Current() int64
	Add()
	To(percent int64)
}

type ProcessPercent struct {
	ProcessPercentMsger

	percentLock sync.RWMutex
	percent     int64
}

func NewProcessPercent(sender ProcessPercentMsger) Percentable {
	return &ProcessPercent{
		percent:             0,
		ProcessPercentMsger: sender,
	}
}

func (pp *ProcessPercent) Current() int64 {
	pp.percentLock.RLock()
	defer pp.percentLock.RUnlock()

	return pp.percent
}

func (pp *ProcessPercent) Add() {
	pp.percentLock.Lock()
	defer pp.percentLock.Unlock()

	if pp.percent < 100 {
		pp.percent++
		pp.SendProcessPercent(fmt.Sprintf("%d", pp.percent))
	}
}

func (pp *ProcessPercent) To(percent int64) {
	pp.percentLock.Lock()
	defer pp.percentLock.Unlock()

	sleepTime := 100 * time.Millisecond
	for pp.percent < percent {
		time.Sleep(sleepTime)
		pp.percent++
		if sleepTime > 50*time.Millisecond {
			sleepTime = sleepTime / 2
		}
		pp.SendProcessPercent(fmt.Sprintf("%d", pp.percent))
	}
}

type Runnable interface {
	IsRunning() bool
	SetRunning(running bool)
	Run() error
}

type Stoppable interface {
	GetStoppedErrorIfHas() error
	IsStopped() bool
	Stop(error)
}

type Job interface {
	Runnable
	Stoppable

	User() contracts.UserInfo
	IsNew() bool
	Project() *models.Project
	Commit() plugins.CommitInterface

	ID() string
	Validate() error
	LoadConfigs() error
	HandleMessage()
	Prune()
	AddDestroyFunc(fn func())
	CallDestroyFuncs()

	ReleaseInstaller() ReleaseInstaller
	Messager() DeployMsger
	PubSub() plugins.PubSub
	Percenter() Percentable
}

type running struct {
	running bool
	sync.Mutex
}

func (run *running) SetRunning(r bool) {
	run.Lock()
	defer run.Unlock()
	run.running = r
}

func (run *running) IsRunning() bool {
	run.Lock()
	defer run.Unlock()
	return run.running
}

type vars map[string]interface{}

func (v vars) MustGetString(key string) string {
	if v != nil {
		if value, ok := v[key]; ok {
			return fmt.Sprintf("%v", value)
		}
	}

	return ""
}

type Jober struct {
	running

	input    *websocket_pb.ProjectInput
	wsType   websocket_pb.Type
	slugName string

	destroyFuncLock sync.RWMutex
	destroyFuncs    []func()

	imagePullSecrets  []string
	vars              vars
	dynamicConfigYaml string
	valuesYaml        string
	chart             *chart.Chart
	valuesOptions     *values.Options
	installer         ReleaseInstaller
	commit            plugins.CommitInterface

	messageCh chan MessageItem
	stopCtx   context.Context
	stopFn    func(error)

	isNew   bool
	config  *mars.MarsConfig
	project *models.Project

	percenter Percentable
	messager  DeployMsger

	pubsub plugins.PubSub

	user contracts.UserInfo
}

func NewJober(
	input *websocket_pb.ProjectInput,
	user contracts.UserInfo,
	slugName string,
	messager DeployMsger,
	pubsub plugins.PubSub,
) Job {
	return &Jober{
		user:          user,
		pubsub:        pubsub,
		messager:      messager,
		vars:          vars{},
		valuesOptions: &values.Options{},
		slugName:      slugName,
		input:         input,
		wsType:        input.Type,
	}
}

func (j *Jober) ID() string {
	return j.slugName
}

func (j *Jober) IsNew() bool {
	return j.isNew
}

func (j *Jober) Commit() plugins.CommitInterface {
	return j.commit
}

func (j *Jober) User() contracts.UserInfo {
	return j.user
}

func (j *Jober) Project() *models.Project {
	return j.project
}

func (j *Jober) Stop(err error) {
	j.stopFn(err)
}

func (j *Jober) IsStopped() bool {
	select {
	case <-j.stopCtx.Done():
		return true
	default:
	}

	return false
}

func (j *Jober) Prune() {
	if j.IsNew() {
		mlog.Debug("清理项目")
		app.DB().Delete(&j.project)
	}
}

func (j *Jober) CallDestroyFuncs() {
	j.destroyFuncLock.RLock()
	defer j.destroyFuncLock.RUnlock()
	for _, destroyFunc := range j.destroyFuncs {
		destroyFunc()
	}
}

func (j *Jober) HandleMessage() {
	defer mlog.Debug("HandleMessage exit")
	for {
		select {
		case <-app.App().Done():
			return
		case s, ok := <-j.messageCh:
			if !ok {
				return
			}
			switch s.Type {
			case MessageText:
				j.Messager().SendMsg(s.Msg)
			case MessageError:
				if j.IsNew() {
					app.DB().Delete(&j.project)
				}
				select {
				case <-j.stopCtx.Done():
					j.Messager().SendDeployedResult(ResultDeployCanceled, j.stopCtx.Err().Error(), j.Project())
				default:
					j.Messager().SendDeployedResult(ResultDeployFailed, s.Msg, j.Project())
				}
			case MessageSuccess:
				j.Messager().SendDeployedResult(ResultDeployed, s.Msg, j.Project())
			}
		}
	}
}

func (j *Jober) AddDestroyFunc(fn func()) {
	j.destroyFuncLock.Lock()
	defer j.destroyFuncLock.Unlock()
	j.destroyFuncs = append(j.destroyFuncs, fn)
}

type userConfig struct {
	Config string `yaml:"config"`
	Branch string `yaml:"branch"`
	Commit string `yaml:"commit"`
	Atomic bool   `yaml:"atomic"`
	WebUrl string `yaml:"web_url"`
}

func (u userConfig) PrettyYaml() string {
	bf := bytes.Buffer{}
	yaml.NewEncoder(&bf).Encode(&u)
	return bf.String()
}

func (j *Jober) Run() error {
	j.SetRunning(true)

	go j.HandleMessage()

	return func() error {
		defer func() {
			j.SetRunning(false)
			close(j.messageCh)
		}()
		var (
			result *release.Release
			err    error
		)

		if result, err = j.ReleaseInstaller().Run(j.stopCtx, j.messageCh); err != nil {
			j.messageCh <- MessageItem{
				Msg:  err.Error(),
				Type: MessageError,
			}
		} else {
			coalesceValues, _ := chartutil.CoalesceValues(j.ReleaseInstaller().Chart(), result.Config)
			j.project.OverrideValues, _ = coalesceValues.YAML()
			j.project.SetPodSelectors(getPodSelectorsInDeploymentAndStatefulSetByManifest(result.Manifest))
			j.project.DockerImage = matchDockerImage(pipelineVars{
				Pipeline: j.vars.MustGetString("Pipeline"),
				Commit:   j.vars.MustGetString("Commit"),
				Branch:   j.vars.MustGetString("Branch"),
			}, result.Manifest)

			var (
				p                models.Project
				oldConf, newConf userConfig
			)
			if app.DB().
				Select("ID", "GitlabProjectId", "Name", "NamespaceId", "Config", "GitlabBranch", "GitlabCommit", "Atomic").
				Where("`name` = ? AND `namespace_id` = ?", j.project.Name, j.project.NamespaceId).
				First(&p).Error == nil {
				j.project.ID = p.ID
				app.DB().Model(j.project).
					Select("Config", "GitlabProjectId", "GitlabCommit", "GitlabBranch", "DockerImage", "PodSelectors", "OverrideValues", "Atomic").
					Updates(&j.project)
				oldConf = userConfig{
					Config: p.Config,
					Branch: p.GitlabBranch,
					Commit: p.GitlabCommit,
					Atomic: p.Atomic,
				}
				commit, err := plugins.GetGitServer().GetCommit(fmt.Sprintf("%d", p.GitlabProjectId), p.GitlabCommit)
				if err == nil {
					oldConf.WebUrl = commit.GetWebURL()
				}
			} else {
				app.DB().Create(&j.project)
			}
			newConf = userConfig{
				Config: j.project.Config,
				Branch: j.project.GitlabBranch,
				Commit: j.project.GitlabCommit,
				Atomic: j.project.Atomic,
				WebUrl: j.Commit().GetWebURL(),
			}
			app.Event().Dispatch(events.EventProjectChanged, &events.ProjectChangedData{
				Project:  j.project,
				Manifest: result.Manifest,
				Config:   j.input.Config,
				Username: j.User().Name,
			})
			var act event.ActionType = event.ActionType_Create
			if !j.IsNew() {
				act = event.ActionType_Update
			}
			AuditLogWithChange(j.User().Name, act,
				fmt.Sprintf("%s 项目: %s/%s", act.String(), j.Project().Namespace.Name, j.Project().Name),
				oldConf, newConf)
			j.Percenter().To(100)
			j.messageCh <- MessageItem{
				Msg:  "部署成功",
				Type: MessageSuccess,
			}
		}
		return err
	}()
}

func (j *Jober) GetStoppedErrorIfHas() error {
	if j.IsStopped() {
		return j.stopCtx.Err()
	}
	return nil
}

func (j *Jober) ReleaseInstaller() ReleaseInstaller {
	return j.installer
}

func (j *Jober) Messager() DeployMsger {
	return j.messager
}

func (j *Jober) PubSub() plugins.PubSub {
	return j.pubsub
}

func (j *Jober) Percenter() Percentable {
	return j.percenter
}

func (j *Jober) Validate() error {
	if !(j.wsType == websocket_pb.Type_CreateProject || j.wsType == websocket_pb.Type_UpdateProject || j.wsType == websocket_pb.Type_ApplyProject) {
		return errors.New("type error: " + j.wsType.String())
	}

	j.percenter = NewProcessPercent(j.messager)
	j.stopCtx, j.stopFn = utils.NewCustomErrorContext()
	j.messageCh = make(chan MessageItem)

	j.Messager().SendMsg("[Start]: 收到请求，开始创建项目")
	j.Percenter().To(5)

	j.Messager().SendMsg("[Check]: 校验名称空间...")

	var ns models.Namespace
	if err := app.DB().Where("`id` = ?", j.input.NamespaceId).First(&ns).Error; err != nil {
		j.Messager().SendMsg(fmt.Sprintf("[FAILED]: 校验名称空间: %s", err.Error()))
		return err
	}

	j.project = &models.Project{
		Name:            slug.Make(j.input.Name),
		GitlabProjectId: int(j.input.GitlabProjectId),
		GitlabBranch:    j.input.GitlabBranch,
		GitlabCommit:    j.input.GitlabCommit,
		Config:          j.input.Config,
		NamespaceId:     ns.ID,
		Namespace:       ns,
		Atomic:          j.input.Atomic,
	}

	j.Messager().SendMsg("[Check]: 检查项目是否存在")

	var p models.Project
	if app.DB().Where("`name` = ? AND `namespace_id` = ?", j.project.Name, j.project.NamespaceId).First(&p).Error == gorm.ErrRecordNotFound {
		app.DB().Create(&j.project)
		j.Messager().SendMsg("[Check]: 新建项目")
		j.isNew = true
	}
	j.imagePullSecrets = j.project.Namespace.ImagePullSecretsArray()

	status, _ := utils.ReleaseStatus(j.project.Name, j.project.Namespace.Name)
	if status == utils.StatusPending {
		return errors.New("有别人也在操作这个项目，等等哦~")
	}

	return nil
}

func defaultLoaders() []Loader {
	return []Loader{
		&MarsLoader{},
		&ChartFileLoader{},
		&VariableLoader{},
		&DynamicLoader{},
		&MergeValuesLoader{},
		&ReleaseInstallerLoader{},
	}
}

func (j *Jober) LoadConfigs() error {
	ch := make(chan error)
	go func() {
		ch <- func() error {
			j.Messager().SendMsg("[Check]: 加载项目文件")

			for _, defaultLoader := range defaultLoaders() {
				if err := j.GetStoppedErrorIfHas(); err != nil {
					return err
				}
				if err := defaultLoader.Load(j); err != nil {
					return err
				}
			}

			return nil
		}()
	}()

	select {
	case err := <-ch:
		return err
	case <-j.stopCtx.Done():
		return j.stopCtx.Err()
	}
}

type Loader interface {
	Load(*Jober) error
}

type MarsLoader struct{}

func (m *MarsLoader) Load(j *Jober) error {
	const loaderName = "[MarsLoader]: "

	j.Messager().SendMsg(loaderName + "加载用户配置")
	j.Percenter().To(10)

	marsC, err := utils.GetProjectMarsConfig(j.input.GitlabProjectId, j.input.GitlabBranch)
	if err != nil {
		j.Messager().SendMsg(fmt.Sprintf(loaderName+"加载 mars config 失败: %s", err.Error()))
		return err
	}
	j.config = marsC

	return nil
}

type ChartFileLoader struct{}

func (c *ChartFileLoader) Load(j *Jober) error {
	const loaderName = "[ChartFileLoader]: "
	j.Messager().SendMsg(loaderName + "加载 helm chart 文件")
	j.Percenter().To(20)

	// 下载 helm charts
	split := strings.Split(j.config.LocalChartPath, "|")
	var (
		files        []string
		tmpChartsDir string
		deleteDirFn  func()
		dir          string
	)
	// 如果是这个格式意味着是远程项目, 'uid|branch|path'
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"下载 helm charts path: %s", j.config.LocalChartPath))

	if utils.IsRemoteChart(j.config) {
		pid := split[0]
		branch := split[1]
		path := split[2]
		files, _ = plugins.GetGitServer().GetDirectoryFilesWithBranch(pid, branch, path, true)
		if len(files) < 1 {
			return errors.New("charts 文件不存在")
		}
		mlog.Warning(files)
		var err error
		tmpChartsDir, deleteDirFn, err = utils.DownloadFiles(pid, branch, files)
		if err != nil {
			return err
		}

		dir = path

		loadDir, _ := loader.LoadDir(filepath.Join(tmpChartsDir, dir))
		if loadDir.Metadata.Dependencies != nil && action.CheckDependencies(loadDir, loadDir.Metadata.Dependencies) != nil {
			for _, dependency := range loadDir.Metadata.Dependencies {
				if strings.HasPrefix(dependency.Repository, "file://") {
					depFiles, _ := plugins.GetGitServer().GetDirectoryFilesWithBranch(pid, branch, filepath.Join(path, strings.TrimPrefix(dependency.Repository, "file://")), true)
					_, depDeleteFn, err := utils.DownloadFilesToDir(pid, branch, depFiles, tmpChartsDir)
					if err != nil {
						return err
					}
					j.AddDestroyFunc(depDeleteFn)
					j.Messager().SendMsg(fmt.Sprintf("下载本地依赖 %s", dependency.Name))
				}
			}
		}
		j.Messager().SendMsg(fmt.Sprintf(loaderName+"识别为远程仓库 uid %v branch %s path %s", pid, branch, path))
	} else {
		var err error
		dir = j.config.LocalChartPath
		files, _ = plugins.GetGitServer().GetDirectoryFilesWithSha(fmt.Sprintf("%d", j.input.GitlabProjectId), j.input.GitlabCommit, j.config.LocalChartPath, true)
		tmpChartsDir, deleteDirFn, err = utils.DownloadFiles(j.input.GitlabProjectId, j.input.GitlabCommit, files)
		if err != nil {
			return err
		}
	}

	j.AddDestroyFunc(deleteDirFn)

	chartDir := filepath.Join(tmpChartsDir, dir)

	j.Percenter().To(30)
	j.Messager().SendMsg(loaderName + "打包 helm charts")
	chart, err := utils.PackageChart(chartDir, chartDir)
	if err != nil {
		return err
	}
	archive, err := os.Open(chart)
	if err != nil {
		return err
	}
	defer archive.Close()

	if j.chart, err = loader.LoadArchive(archive); err != nil {
		return err
	}

	return nil
}

type DynamicLoader struct {
	values map[string]interface{}
}

func (d *DynamicLoader) Load(j *Jober) error {
	const loaderName = "[DynamicLoader]: "

	j.Percenter().To(40)
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "检查到用户传入的配置"))

	if j.input.Config == "" {
		j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "未发现用户自定义配置"))
		return nil
	}

	dynamicConfigYaml, err := utils.ParseInputConfig(j.config, j.input.Config)
	if err != nil {
		return err
	}
	j.dynamicConfigYaml = dynamicConfigYaml

	return nil
}

const (
	leftDelim  = "<"
	rightDelim = ">"

	VarImagePullSecrets = "ImagePullSecrets"
	VarBranch           = "Branch"
	VarCommit           = "Commit"
	VarPipeline         = "Pipeline"
	VarClusterIssuer    = "ClusterIssuer"
	VarHost             = "Host"
	VarTlsSecret        = "TlsSecret"
)

type VariableLoader struct {
	values vars
}

func (v *VariableLoader) Load(j *Jober) error {
	const loaderName = "[VariableLoader]: "
	j.Percenter().To(50)
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "注入内置环境变量"))

	if j.config.ValuesYaml == "" {
		j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "未发现可用的 values.yaml"))
		return nil
	}

	if v.values == nil {
		v.values = vars{}
	}

	//ImagePullSecrets
	parse, e := template.New("ImagePullSecrets").Parse(fmt.Sprintf("[{{- range .%s }}{name: {{ . }}}, {{- end }}]", VarImagePullSecrets))
	if e != nil {
		return e
	}

	renderResult := &bytes.Buffer{}
	if err := parse.Execute(renderResult, struct {
		ImagePullSecrets []string
	}{
		ImagePullSecrets: j.imagePullSecrets,
	}); err != nil {
		return err
	}

	v.values[VarImagePullSecrets] = renderResult.String()

	//Host1...Host10
	sub := getPreOccupiedLenByValuesYaml(j.config.ValuesYaml)
	for i := 1; i <= 10; i++ {
		v.values[fmt.Sprintf("%s%d", VarHost, i)] = getDomainByIndex(j.project.Name, j.project.Namespace.Name, i, sub)
		v.values[fmt.Sprintf("%s%d", VarTlsSecret, i)] = fmt.Sprintf("mars-tls-%s", utils.Md5(fmt.Sprintf("%s-%d", j.project.Name, i)))
	}

	//{{.Branch}}{{.Commit}}{{.Pipeline}}
	commit, err := plugins.GetGitServer().GetCommit(fmt.Sprintf("%d", j.project.GitlabProjectId), j.project.GitlabCommit)
	if err != nil {
		return err
	}
	j.commit = commit
	var (
		pipelineID     int64
		pipelineBranch string
		pipelineCommit string = commit.GetShortID()
	)

	// 如果存在需要传变量的，则必须有流水线信息
	if commit.GetLastPipeline() != nil {
		pipelineID = commit.GetLastPipeline().GetID()
		pipelineBranch = commit.GetLastPipeline().GetRef()

		j.Messager().SendMsg(fmt.Sprintf(loaderName+"镜像分支 %s 镜像commit %s 镜像 pipeline_id %d", pipelineBranch, pipelineCommit, pipelineID))
	} else {
		if tagRegex.MatchString(j.config.ValuesYaml) {
			return errors.New("无法获取 Pipeline 信息")
		}
	}

	v.values[VarBranch] = pipelineBranch
	v.values[VarCommit] = pipelineCommit
	v.values[VarPipeline] = pipelineID

	// ingress
	v.values[VarClusterIssuer] = app.Config().ClusterIssuer

	tpl, err := template.New("values_yaml").Delims(leftDelim, rightDelim).Parse(j.config.ValuesYaml)
	if err != nil {
		return err
	}
	bf := bytes.Buffer{}
	tpl.Execute(&bf, v.values)
	j.valuesYaml = bf.String()
	j.vars = v.values

	return nil
}

type MergeValuesLoader struct{}

// Load
// imagePullSecrets 会自动注入到 imagePullSecrets 中
func (m *MergeValuesLoader) Load(j *Jober) error {
	const loaderName = "[MergeValuesLoader]: "
	j.Percenter().To(60)
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "合并配置文件到 values.yaml"))

	// 自动注入 imagePullSecrets
	var imagePullSecrets = make([]map[string]interface{}, len(j.imagePullSecrets))
	for i, s := range j.imagePullSecrets {
		imagePullSecrets[i] = map[string]interface{}{"name": s}
	}
	yamlImagePullSecrets, _ := yaml.Marshal(map[string]interface{}{
		"imagePullSecrets": imagePullSecrets,
	})

	var opts []config.YAMLOption
	if j.valuesYaml != "" {
		opts = append(opts, config.Source(strings.NewReader(j.valuesYaml)))
	}
	if j.dynamicConfigYaml != "" {
		opts = append(opts, config.Source(strings.NewReader(j.dynamicConfigYaml)))
	}
	if len(yamlImagePullSecrets) != 0 {
		opts = append(opts, config.Source(bytes.NewReader(yamlImagePullSecrets)))
	}

	if len(opts) < 1 {
		return nil
	}

	// 5. 用用户传入的yaml配置去合并 `default_values`
	provider, err := config.NewYAML(opts...)
	if err != nil {
		mlog.Error(loaderName, err, j.valuesYaml, j.dynamicConfigYaml)

		return err
	}
	var mergedDefaultAndConfigYamlValues map[string]interface{}
	if err := provider.Get("").Populate(&mergedDefaultAndConfigYamlValues); err != nil {
		mlog.Error(loaderName, mergedDefaultAndConfigYamlValues, err)
		return err
	}

	bf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(bf)
	if err := encoder.Encode(&mergedDefaultAndConfigYamlValues); err != nil {
		return err
	}
	fileData := bf.String()
	mergedFile, closer, err := utils.WriteConfigYamlToTmpFile([]byte(fileData))
	if err != nil {
		return err
	}
	j.AddDestroyFunc(func() { closer.Close() })
	j.valuesOptions.ValueFiles = append(j.valuesOptions.ValueFiles, mergedFile)

	return nil
}

type ReleaseInstallerLoader struct{}

func (r *ReleaseInstallerLoader) Load(j *Jober) error {
	const loaderName = "ReleaseInstallerLoader"
	j.Messager().SendMsg(loaderName + "worker 已就绪, 准备安装")
	j.Percenter().To(70)
	j.installer = newReleaseInstaller(j.project.Name, j.project.Namespace.Name, j.chart, j.valuesOptions, func(format string, v ...interface{}) {
		if j.Percenter().Current() < 99 {
			j.Percenter().Add()
		}
		if j.Percenter().Current() >= 95 {
			format = "[如果长时间未部署成功，建议取消使用 debug 模式]: " + format
		}
		msg := fmt.Sprintf(format, v...)
		if j.IsRunning() {
			select {
			case j.messageCh <- MessageItem{
				Msg:  msg,
				Type: MessageText,
			}:
			default:
			}
		}
	}, j.input.Atomic)
	return nil
}
