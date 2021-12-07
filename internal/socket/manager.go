package socket

import (
	"bytes"
	"context"
	"encoding/json"
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

	"google.golang.org/protobuf/proto"

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

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/grpc/services"
	"github.com/duc-cnzj/mars/internal/mars"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
	websocket_pb "github.com/duc-cnzj/mars/pkg/websocket"
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
	hostMatch = regexp.MustCompile(".*?=(.*?){{\\s*.Host\\d\\s*}}")
	tagRegex  = regexp.MustCompile("{{\\s*(\\.Branch|\\.Commit|\\.Pipeline)\\s*}}")

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type WsResponse = websocket_pb.WsResponseMetadata

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

type Messageable interface {
	SendEndError(error)
	SendError(error)
	SendProcessPercent(string)
	SendMsg(string)
	SendProtoMsg( proto.Message)
	SendEndMsg(websocket_pb.ResultType, string)
}

type MessageSender struct {
	conn     *WsConn
	slugName string
	wsType   websocket_pb.Type
}

func NewMessageSender(conn *WsConn, slugName string, wsType websocket_pb.Type) *MessageSender {
	return &MessageSender{conn: conn, slugName: slugName, wsType: wsType}
}

func (ms *MessageSender) SendEndError(err error) {
	res := &WsResponse{
		Metadata: &websocket_pb.ResponseMetadata{
			Slug:   ms.slugName,
			Type:   ms.wsType,
			Result: ResultError,
			Data:   err.Error(),
			End:    true,
			Uid:    ms.conn.uid,
			Id:     ms.conn.id,
		},
	}
	ms.send(res)
}

func (ms *MessageSender) SendError(err error) {
	res := &WsResponse{
		Metadata: &websocket_pb.ResponseMetadata{
			Slug:   ms.slugName,
			Type:   ms.wsType,
			Result: ResultError,
			Data:   err.Error(),
			End:    false,
			Uid:    ms.conn.uid,
			Id:     ms.conn.id,
		},
	}
	ms.send(res)
}

func (ms *MessageSender) SendProcessPercent(percent string) {
	res := &WsResponse{
		Metadata: &websocket_pb.ResponseMetadata{
			Slug:   ms.slugName,
			Type:   WsProcessPercent,
			Result: ResultSuccess,
			End:    false,
			Data:   percent,
			Uid:    ms.conn.uid,
			Id:     ms.conn.id,
		},
	}
	ms.send(res)
}

func (ms *MessageSender) SendMsg(msg string) {
	res := &WsResponse{
		Metadata: &websocket_pb.ResponseMetadata{
			Slug:   ms.slugName,
			Type:   ms.wsType,
			Result: ResultSuccess,
			End:    false,
			Data:   msg,
			Uid:    ms.conn.uid,
			Id:     ms.conn.id,
		},
	}
	ms.send(res)
}

func (ms *MessageSender) SendProtoMsg(msg proto.Message) {
	ms.send(msg)
}

func (ms *MessageSender) SendEndMsg(result websocket_pb.ResultType, msg string) {
	res := &WsResponse{
		Metadata: &websocket_pb.ResponseMetadata{
			Slug:   ms.slugName,
			Type:   ms.wsType,
			Result: result,
			End:    true,
			Data:   msg,
			Uid:    ms.conn.uid,
			Id:     ms.conn.id,
		},
	}
	ms.send(res)
}

func (ms *MessageSender) send(res proto.Message) {
	ms.conn.pubSub.ToSelf(res)
}

type Percentable interface {
	Current() int64
	Add()
	To(percent int64)
}

type ProcessPercent struct {
	Messageable

	percentLock sync.RWMutex
	percent     int64
}

func NewProcessPercent(sender Messageable) Percentable {
	return &ProcessPercent{
		percent:     0,
		Messageable: sender,
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

	ID() string
	Validate() error
	LoadConfigs() error
	HandleMessage()
	Prune()
	AddDestroyFunc(fn func())
	CallDestroyFuncs()

	ReleaseInstaller() ReleaseInstaller
	Messager() Messageable
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

type Jober struct {
	running

	id       string
	input    *websocket_pb.ProjectInput
	wsType   websocket_pb.Type
	conn     *WsConn
	slugName string

	destroyFuncLock sync.RWMutex
	destroyFuncs    []func()

	imageTag       string
	chart          *chart.Chart
	valuesOptions  *values.Options
	ingressOptions []string
	installer      ReleaseInstaller

	messageCh chan MessageItem
	stopCtx   context.Context
	stopFn    func(error)

	isNew     bool
	config    *mars.Config
	messager  Messageable
	project   *models.Project
	percenter Percentable
}

func NewJober(input *websocket_pb.ProjectInput, wsType websocket_pb.Type, conn *WsConn) Job {
	return &Jober{
		slugName: utils.Md5(fmt.Sprintf("%d-%s", input.NamespaceId, input.Name)),
		input:    input,
		wsType:   wsType,
		conn:     conn,
	}
}

func (j *Jober) ID() string {
	return j.slugName
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
	if j.isNew {
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
				if j.isNew {
					app.DB().Delete(&j.project)
				}
				select {
				case <-j.stopCtx.Done():
					j.Messager().SendEndMsg(ResultDeployCanceled, j.stopCtx.Err().Error())
				default:
					j.Messager().SendEndMsg(ResultDeployFailed, s.Msg)
				}
			case MessageSuccess:
				j.Messager().SendEndMsg(ResultDeployed, s.Msg)
			}
		}
	}
}

func (j *Jober) AddDestroyFunc(fn func()) {
	j.destroyFuncLock.Lock()
	defer j.destroyFuncLock.Unlock()
	j.destroyFuncs = append(j.destroyFuncs, fn)
}

func (j *Jober) Run() error {
	j.SetRunning(true)
	defer j.SetRunning(false)

	go j.HandleMessage()

	return func() error {
		defer close(j.messageCh)
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
			var p models.Project
			if app.DB().Where("`name` = ? AND `namespace_id` = ?", j.project.Name, j.project.NamespaceId).First(&p).Error == nil {
				app.DB().Model(&models.Project{}).
					Select("Config", "GitlabProjectId", "GitlabCommit", "GitlabBranch", "DockerImage", "PodSelectors", "OverrideValues", "Atomic").
					Where("`id` = ?", p.ID).
					Updates(&j.project)
			} else {
				app.DB().Create(&j.project)
			}
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

func (j *Jober) Messager() Messageable {
	return j.messager
}

func (j *Jober) PubSub() plugins.PubSub {
	return j.conn.pubSub
}

func (j *Jober) Percenter() Percentable {
	return j.percenter
}

func (j *Jober) Validate() error {
	j.messager = NewMessageSender(j.conn, j.slugName, j.wsType)
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

	return nil
}

var defaultLoaders = []Loader{
	MarsLoader{},
	ChartFileLoader{},
	TagLoader{},
	IngressLoader{},
	ValuesFileLoader{},
	ReleaseInstallerLoader{},
}

func (j *Jober) LoadConfigs() error {
	ch := make(chan error)
	go func() {
		ch <- func() error {
			j.Messager().SendMsg("[Check]: 加载项目文件")

			for _, defaultLoader := range defaultLoaders {
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

func (m MarsLoader) Load(j *Jober) error {
	const loaderName = "[MarsLoader]: "

	j.Messager().SendMsg(loaderName + "加载用户配置")
	j.Percenter().To(10)

	marsC, err := services.GetProjectMarsConfig(j.input.GitlabProjectId, j.input.GitlabBranch)
	if err != nil {
		j.Messager().SendMsg(fmt.Sprintf(loaderName+"加载 mars config 失败: %s", err.Error()))
		return err
	}
	marsC.ImagePullSecrets = j.project.Namespace.ImagePullSecretsArray()
	j.config = marsC

	return nil
}

type ChartFileLoader struct{}

func (c ChartFileLoader) Load(j *Jober) error {
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

	if j.config.IsRemoteChart() {
		pid := split[0]
		branch := split[1]
		path := split[2]
		files = utils.GetDirectoryFiles(pid, branch, path)
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
					depFiles := utils.GetDirectoryFiles(pid, branch, filepath.Join(path, strings.TrimPrefix(dependency.Repository, "file://")))
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
		files = utils.GetDirectoryFiles(j.input.GitlabProjectId, j.input.GitlabCommit, j.config.LocalChartPath)
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

type TagLoader struct{}

func (t TagLoader) Load(j *Jober) error {
	const loaderName = "[TagLoader]: "

	j.Messager().SendMsg(loaderName + "解析镜像tag")
	j.Percenter().To(40)
	tagParse := template.New("tag_parse")
	parse, err := tagParse.Parse(j.config.DockerTagFormat)
	if err != nil {
		return err
	}
	b := &bytes.Buffer{}
	commit, _, err := app.GitlabClient().Commits.GetCommit(j.project.GitlabProjectId, j.project.GitlabCommit)
	if err != nil {
		return err
	}
	var (
		pipelineID     int
		pipelineBranch string
		pipelineCommit string = commit.ShortID
	)

	// 如果存在需要传变量的，则必须有流水线信息
	if commit.LastPipeline != nil {
		pipelineID = commit.LastPipeline.ID
		pipelineBranch = commit.LastPipeline.Ref
	} else {
		if tagRegex.MatchString(j.config.DockerTagFormat) {
			return errors.New("无法获取 Pipeline 信息")
		}
	}
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"镜像分支 %s 镜像commit %s 镜像 pipeline_id %d", pipelineBranch, pipelineCommit, pipelineID))

	if err := parse.Execute(b, struct {
		Branch   string
		Commit   string
		Pipeline int
	}{
		Branch:   pipelineBranch,
		Commit:   pipelineCommit,
		Pipeline: pipelineID,
	}); err != nil {
		return err
	}
	j.imageTag = b.String()
	j.project.DockerImage = fmt.Sprintf("%s:%s", j.config.DockerRepository, j.imageTag)
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"使用的镜像是: %s", fmt.Sprintf("%s:%s", j.config.DockerRepository, j.imageTag)))

	image := strings.Split(j.project.DockerImage, ":")
	if len(image) == 2 {
		if plugins.GetDockerPlugin().ImageNotExists(image[0], image[1]) {
			return errors.New(fmt.Sprintf("镜像 %s 不存在！", j.project.DockerImage))
		}
	}

	return nil
}

type IngressLoader struct{}

func (IngressLoader) Load(j *Jober) error {
	const loaderName = "[IngressLoader]: "

	j.Messager().SendMsg(loaderName + "解析域名配置")
	j.Percenter().To(50)

	var ingressConfig []string

	if app.Config().HasWildcardDomain() {
		sub := getPreOccupiedLen(j.config.IngressOverwriteValues)
		var host, secretName string = getDomain(j.project.Name, j.project.Namespace.Name, sub), fmt.Sprintf("%s-%s-tls", j.project.Name, j.project.Namespace.Name)
		var vars = map[string]string{}
		for i := 1; i <= 10; i++ {
			vars[fmt.Sprintf("Host%d", i)] = getDomainByIndex(j.project.Name, j.project.Namespace.Name, i, sub)
			vars[fmt.Sprintf("TlsSecret%d", i)] = fmt.Sprintf("%s-%s-%d-tls", j.project.Name, j.project.Namespace.Name, i)
		}
		//TODO: 不同k8s版本 ingress 定义不一样, helm 生成的 template 不一样。
		//旧版长这样
		//ingress:
		// enabled: true
		// annotations: {}
		//   # kubernetes.io/ingress.class: nginx
		//   # kubernetes.io/tls-acme: "true"
		// hosts:
		//   - host: chart-example.local
		//     paths: []
		//新版长这样
		//ingress:
		//enabled: false
		//annotations: {}
		//	# kubernetes.io/ingress.class: nginx
		//	# kubernetes.io/tls-acme: "true"
		//hosts:
		//	- host: chart-example.local
		//paths:
		//	- path: /
		//   backend:
		//     serviceName: chart-example.local
		//     servicePort: 80
		var isOldVersion bool
		for _, f := range j.chart.Templates {
			if strings.Contains(f.Name, "ingress") {
				if strings.Contains(string(f.Data), "path: {{ . }}") {
					isOldVersion = true
					break
				}
			}
		}
		ingressConfig = []string{
			"ingress.enabled=true",
			"ingress.annotations.kubernetes\\.io\\/ingress\\.class=nginx",
		}
		if app.Config().ClusterIssuer != "" {
			ingressConfig = append(ingressConfig, "ingress.annotations.cert\\-manager\\.io\\/cluster\\-issuer="+app.Config().ClusterIssuer)
		}

		if len(j.config.IngressOverwriteValues) > 0 {
			var overwrites []string
			for _, value := range j.config.IngressOverwriteValues {
				bb := &bytes.Buffer{}
				ingressT := template.New("")
				t2, _ := ingressT.Parse(value)
				if err := t2.Execute(bb, vars); err != nil {
					return err
				}
				overwrites = append(overwrites, bb.String())
			}
			mlog.Warning(overwrites)
			ingressConfig = append(ingressConfig, overwrites...)
		} else {
			ingressConfig = append(ingressConfig, []string{
				"ingress.hosts[0].host=" + host,
				"ingress.tls[0].secretName=" + secretName,
				"ingress.tls[0].hosts[0]=" + host,
			}...)

			if isOldVersion {
				ingressConfig = append(ingressConfig, "ingress.hosts[0].paths[0]=/")
			} else {
				ingressConfig = append(ingressConfig, "ingress.hosts[0].paths[0].path=/", "ingress.hosts[0].paths[0].pathType=Prefix")
			}
		}

		j.Messager().SendMsg(fmt.Sprintf(loaderName+"已配置域名: %s", host))
	}
	j.ingressOptions = ingressConfig
	return nil
}

type ValuesFileLoader struct{}

func (v ValuesFileLoader) Load(j *Jober) error {
	const loaderName = "[ValuesFileLoader]: "
	j.Messager().SendMsg(loaderName + "写入配置文件")
	j.Percenter().To(60)

	var inputValuesMap = make(map[string]interface{})
	var err error
	// 1. 把用户提交的配置变成 yaml 文件
	if j.input.Config != "" {
		inputValuesMap, err = j.config.ParseInputConfigToMap(j.input.Config)
		if err != nil {
			return err
		}
	}

	// 2. 设置自定义的镜像和仓库
	inputValuesMap["image"] = map[string]interface{}{
		"tag":        j.imageTag,
		"repository": j.config.DockerRepository,
	}
	var imagePullSecrets = make([]map[string]interface{}, len(j.config.ImagePullSecrets))
	for i, s := range j.config.ImagePullSecrets {
		imagePullSecrets[i] = map[string]interface{}{"name": s}
	}
	inputValuesMap["imagePullSecrets"] = imagePullSecrets

	// 3. 融合镜像和自定义配置文件
	inputValues, err := yaml.Marshal(inputValuesMap)
	if err != nil {
		return err
	}

	// 4. default_values 也需要一个 file
	defaultValues, err := j.config.GenerateDefaultValuesYaml()
	if err != nil {
		return err
	}
	base := strings.NewReader(defaultValues)
	override := bytes.NewReader(inputValues)

	// 5. 用用户传入的yaml配置去合并 `default_values`
	provider, err := config.NewYAML(config.Source(base), config.Source(override))
	if err != nil {
		return err
	}
	var mergedDefaultAndConfigYamlValues map[string]interface{}
	if err := provider.Get("").Populate(&mergedDefaultAndConfigYamlValues); err != nil {
		return err
	}

	indent, _ := json.MarshalIndent(mergedDefaultAndConfigYamlValues, "", "\t")
	mlog.Debugf("用户自定义的 values.yaml: \n %v", string(indent))

	bf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(bf)
	if err := encoder.Encode(&mergedDefaultAndConfigYamlValues); err != nil {
		return err
	}
	mergedFile, closer, err := utils.WriteConfigYamlToTmpFile(bf.Bytes())
	if err != nil {
		return err
	}
	j.AddDestroyFunc(func() { closer.Close() })

	j.valuesOptions = &values.Options{
		ValueFiles: []string{mergedFile},
		Values:     j.ingressOptions,
	}

	return nil
}

type ReleaseInstallerLoader struct{}

func (r ReleaseInstallerLoader) Load(j *Jober) error {
	const loaderName = "ReleaseInstallerLoader"
	j.Messager().SendMsg(loaderName + "worker 安装成功, 准备安装")
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
			j.messageCh <- MessageItem{
				Msg:  msg,
				Type: MessageText,
			}
		}
	}, j.input.Atomic)
	return nil
}
