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

	"go.uber.org/config"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/enums"
	"github.com/duc-cnzj/mars/internal/grpc/services"
	"github.com/duc-cnzj/mars/internal/mars"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/gorilla/websocket"
	"github.com/gosimple/slug"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

const (
	ResultError          string = "error"
	ResultSuccess        string = "success"
	ResultDeployed       string = "deployed"
	ResultDeployFailed   string = "deployed_failed"
	ResultDeployCanceled string = "deployed_canceled"

	WsSetUid             string = enums.WsSetUid
	WsReloadProjects     string = enums.WsReloadProjects
	WsCancel             string = enums.WsCancel
	WsCreateProject      string = enums.WsCreateProject
	WsUpdateProject      string = enums.WsUpdateProject
	WsProcessPercent     string = enums.WsProcessPercent
	WsClusterInfoSync    string = enums.WsClusterInfoSync
	WsInternalError      string = enums.WsInternalError
	WsHandleExecShell    string = enums.WsHandleExecShell
	WsHandleExecShellMsg string = enums.WsHandleExecShellMsg
	WsHandleCloseShell   string = enums.WsHandleCloseShell
	WsAuthorize          string = enums.WsHandleAuthorize

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

type WsRequest struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type WsResponse = plugins.WsResponse

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

type MessageItem struct {
	Msg  string
	Type string
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
	var (
		err    error
		result *release.Release
	)
	done := make(chan error, 1)
	go func() {
		defer func() {
			close(messageCh)
			close(done)
		}()
		defer utils.HandlePanic("ProcessControl: Run")

		if result, err = utils.UpgradeOrInstall(stopCtx, r.releaseName, r.namespace, r.chart, r.valueOpts, r.logger, r.atomic); err != nil {
			mlog.Error(err)
			messageCh <- MessageItem{
				Msg:  err.Error(),
				Type: "error",
			}
			done <- err
		} else {
			messageCh <- MessageItem{
				Msg:  "部署成功",
				Type: "success",
			}
			done <- nil
		}
	}()

	return result, <-done
}

type ProjectManager interface {
	Get() *models.Project
	IsNew() bool
	Delete() error
	Save() error
}

type ConfigManager interface {
	Get() *mars.Config
}

type Messageable interface {
	SendEndError(err error)
	SendError(err error)
	SendProcessPercent(percent string)
	SendMsg(msg string)
	SendEndMsg(result, msg string)
}

type MessageSender struct {
	conn     *WsConn
	slugName string
	wsType   string
}

func NewMessageSender(conn *WsConn, slugName string, wsType string) *MessageSender {
	return &MessageSender{conn: conn, slugName: slugName, wsType: wsType}
}

func (ms *MessageSender) SendEndError(err error) {
	res := &WsResponse{
		Slug:   ms.slugName,
		Type:   ms.wsType,
		Result: ResultError,
		Data:   err.Error(),
		End:    true,
		Uid:    ms.conn.uid,
		ID:     ms.conn.id,
	}
	ms.send(res)
}

func (ms *MessageSender) SendError(err error) {
	res := &WsResponse{
		Slug:   ms.slugName,
		Type:   ms.wsType,
		Result: ResultError,
		Data:   err.Error(),
		End:    false,
		Uid:    ms.conn.uid,
		ID:     ms.conn.id,
	}
	ms.send(res)
}

func (ms *MessageSender) SendProcessPercent(percent string) {
	res := &WsResponse{
		Slug:   ms.slugName,
		Type:   WsProcessPercent,
		Result: ResultSuccess,
		End:    false,
		Data:   percent,
		Uid:    ms.conn.uid,
		ID:     ms.conn.id,
	}
	ms.send(res)
}

func (ms *MessageSender) SendMsg(msg string) {
	res := &WsResponse{
		Slug:   ms.slugName,
		Type:   ms.wsType,
		Result: ResultSuccess,
		End:    false,
		Data:   msg,
		Uid:    ms.conn.uid,
		ID:     ms.conn.id,
	}
	ms.send(res)
}

func (ms *MessageSender) SendEndMsg(result, msg string) {
	res := &WsResponse{
		Slug:   ms.slugName,
		Type:   ms.wsType,
		Result: result,
		End:    true,
		Data:   msg,
		Uid:    ms.conn.uid,
		ID:     ms.conn.id,
	}
	ms.send(res)
}

func (ms *MessageSender) send(res *WsResponse) {
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

	id        string
	input     ProjectInput
	wsType    string
	wsRequest WsRequest
	conn      *WsConn
	slugName  string

	destroyFuncLock sync.RWMutex
	destroyFuncs    []func()

	installer ReleaseInstaller

	messageCh chan MessageItem
	stopCtx   context.Context
	stopFn    func(error)

	isNew     bool
	config    *mars.Config
	messager  Messageable
	project   *models.Project
	percenter Percentable
}

func NewJober(input ProjectInput, wsType string, wsRequest WsRequest, conn *WsConn) Job {
	return &Jober{
		slugName:  utils.Md5(fmt.Sprintf("%d-%s", input.NamespaceId, input.Name)),
		input:     input,
		wsType:    wsType,
		wsRequest: wsRequest,
		conn:      conn,
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
			case "text":
				j.Messager().SendMsg(s.Msg)
			case "error":
				if j.isNew {
					app.DB().Delete(&j.project)
				}
				select {
				case <-j.stopCtx.Done():
					j.Messager().SendEndMsg(ResultDeployCanceled, j.stopCtx.Err().Error())
				default:
					j.Messager().SendEndMsg(ResultDeployFailed, s.Msg)
				}
			case "success":
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
	result, err := j.ReleaseInstaller().Run(j.stopCtx, j.messageCh)
	if err == nil {
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
	}

	return err
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

	j.Messager().SendMsg("收到请求，开始创建项目")
	j.Percenter().To(5)
	j.Messager().SendMsg("校验名称空间...")

	var ns models.Namespace
	if err := app.DB().Where("`id` = ?", j.input.NamespaceId).First(&ns).Error; err != nil {
		return err
	}

	j.project = &models.Project{
		Name:            slug.Make(j.input.Name),
		GitlabProjectId: j.input.GitlabProjectId,
		GitlabBranch:    j.input.GitlabBranch,
		GitlabCommit:    j.input.GitlabCommit,
		Config:          j.input.Config,
		NamespaceId:     ns.ID,
		Namespace:       ns,
		Atomic:          j.input.Atomic,
	}

	j.Messager().SendMsg("检查项目是否存在")

	var p models.Project
	if app.DB().Where("`name` = ? AND `namespace_id` = ?", j.project.Name, j.project.NamespaceId).First(&p).Error == gorm.ErrRecordNotFound {
		app.DB().Create(&j.project)
		j.Messager().SendMsg("项目不存在新建项目")
		j.isNew = true
	}

	return nil
}

func (j *Jober) LoadConfigs() error {
	ch := make(chan error)
	go func() {
		ch <- func() error {
			j.Messager().SendMsg("加载项目文件...")
			j.Percenter().To(15)
			marsC, err := services.GetProjectMarsConfig(j.input.GitlabProjectId, j.input.GitlabBranch)
			if err != nil {
				return err
			}
			marsC.ImagePullSecrets = j.project.Namespace.ImagePullSecretsArray()
			j.config = marsC

			// 下载 helm charts
			j.Messager().SendMsg(fmt.Sprintf("下载 helm charts path: %s ...", marsC.LocalChartPath))
			split := strings.Split(marsC.LocalChartPath, "|")
			var (
				files        []string
				tmpChartsDir string
				deleteDirFn  func()
				dir          string
			)
			// 如果是这个格式意味着是远程项目, 'uid|branch|path'
			if marsC.IsRemoteChart() {
				pid := split[0]
				branch := split[1]
				path := split[2]
				files = utils.GetDirectoryFiles(pid, branch, path)
				if len(files) < 1 {
					return errors.New("charts 文件不存在")
				}
				mlog.Warning(files)
				tmpChartsDir, deleteDirFn = utils.DownloadFiles(pid, branch, files)
				dir = path

				loadDir, _ := loader.LoadDir(filepath.Join(tmpChartsDir, dir))
				if loadDir.Metadata.Dependencies != nil && action.CheckDependencies(loadDir, loadDir.Metadata.Dependencies) != nil {
					for _, dependency := range loadDir.Metadata.Dependencies {
						if strings.HasPrefix(dependency.Repository, "file://") {
							depFiles := utils.GetDirectoryFiles(pid, branch, filepath.Join(path, strings.TrimPrefix(dependency.Repository, "file://")))
							_, depDeleteFn := utils.DownloadFilesToDir(pid, branch, depFiles, tmpChartsDir)
							j.AddDestroyFunc(depDeleteFn)
							j.Messager().SendMsg(fmt.Sprintf("下载本地依赖 %s", dependency.Name))
						}
					}
				}
				j.Messager().SendMsg(fmt.Sprintf("识别为远程仓库 uid %v branch %s path %s", pid, branch, path))
			} else {
				dir = marsC.LocalChartPath
				files = utils.GetDirectoryFiles(j.input.GitlabProjectId, j.input.GitlabCommit, marsC.LocalChartPath)
				tmpChartsDir, deleteDirFn = utils.DownloadFiles(j.input.GitlabProjectId, j.input.GitlabCommit, files)
			}

			j.AddDestroyFunc(deleteDirFn)

			if err := j.GetStoppedErrorIfHas(); err != nil {
				return err
			}

			chartDir := filepath.Join(tmpChartsDir, dir)

			chart, err := utils.PackageChart(chartDir, chartDir)
			if err != nil {
				return err
			}
			archive, err := os.Open(chart)
			if err != nil {
				return err
			}
			defer archive.Close()

			j.Percenter().To(30)
			j.Messager().SendMsg("加载 helm charts...")

			loadArchive, err := loader.LoadArchive(archive)
			if err != nil {
				return err
			}

			if err := j.GetStoppedErrorIfHas(); err != nil {
				return err
			}

			// ##############
			j.Messager().SendMsg("生成配置文件...")
			j.Percenter().To(40)

			var imagePullSecrets []string
			for k, s := range marsC.ImagePullSecrets {
				imagePullSecrets = append(imagePullSecrets, fmt.Sprintf("imagePullSecrets[%d].name=%s", k, s))
			}

			// default_values 也需要一个 file
			defaultValues, err := marsC.GenerateDefaultValuesYaml()
			if err != nil {
				return err
			}

			// 传入自定义配置必须在默认配置之后，不然会被上面的 default_values 覆盖，导致不管你怎么更新配置文件都无法正正的更新到容器
			var configValues string
			if j.input.Config != "" {
				configValues, err = marsC.GenerateConfigYamlByInput(j.input.Config)
				if err != nil {
					return err
				}
			}
			if err := j.GetStoppedErrorIfHas(); err != nil {
				return err
			}
			base := strings.NewReader(defaultValues)
			override := strings.NewReader(configValues)

			provider, err := config.NewYAML(config.Source(base), config.Source(override))
			if err != nil {
				return err
			}
			var mergedDefaultAndConfigYamlValues map[string]interface{}
			if err := provider.Get("").Populate(&mergedDefaultAndConfigYamlValues); err != nil {
				return err
			}

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
			if err := j.GetStoppedErrorIfHas(); err != nil {
				return err
			}
			j.Messager().SendMsg("解析镜像tag")
			j.Percenter().To(45)
			t := template.New("tag_parse")
			parse, err := t.Parse(marsC.DockerTagFormat)
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
				if tagRegex.MatchString(marsC.DockerTagFormat) {
					return errors.New("无法获取 Pipeline 信息")
				}
			}
			if err := j.GetStoppedErrorIfHas(); err != nil {
				return err
			}
			j.Messager().SendMsg(fmt.Sprintf("镜像分支 %s 镜像commit %s 镜像 pipeline_id %d", pipelineBranch, pipelineCommit, pipelineID))

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
			tag := b.String()

			j.Percenter().To(60)
			if err := j.GetStoppedErrorIfHas(); err != nil {
				return err
			}
			var ingressConfig []string

			if app.Config().HasWildcardDomain() {
				sub := getPreOccupiedLen(marsC.IngressOverwriteValues)
				var host, secretName string = getDomain(j.project.Name, j.project.Namespace.Name, sub), fmt.Sprintf("%s-%s-tls", j.project.Name, j.project.Namespace.Name)
				var vars = map[string]string{}
				for i := 1; i <= 10; i++ {
					vars[fmt.Sprintf("Host%d", i)] = getDomainByIndex(j.project.Name, j.project.Namespace.Name, i, sub)
					vars[fmt.Sprintf("TlsSecret%d", i)] = fmt.Sprintf("%s-%s-%d-tls", j.project.Name, j.project.Namespace.Name, i)
				}
				// TODO: 不同k8s版本 ingress 定义不一样, helm 生成的 template 不一样。
				// 旧版长这样
				// ingress:
				//  enabled: true
				//  annotations: {}
				//    # kubernetes.io/ingress.class: nginx
				//    # kubernetes.io/tls-acme: "true"
				//  hosts:
				//    - host: chart-example.local
				//      paths: []
				// 新版长这样
				// ingress:
				// enabled: false
				// annotations: {}
				// 	# kubernetes.io/ingress.class: nginx
				// 	# kubernetes.io/tls-acme: "true"
				// hosts:
				// 	- host: chart-example.local
				// paths:
				// 	- path: /
				//    backend:
				//      serviceName: chart-example.local
				//      servicePort: 80
				var isOldVersion bool
				for _, f := range loadArchive.Templates {
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

				if len(marsC.IngressOverwriteValues) > 0 {
					var overwrites []string
					for _, value := range marsC.IngressOverwriteValues {
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

				j.Messager().SendMsg(fmt.Sprintf("已配置域名: %s", host))
			}

			j.Percenter().To(65)
			if err := j.GetStoppedErrorIfHas(); err != nil {
				return err
			}
			var commonValues = []string{
				"image.pullPolicy=IfNotPresent",
				"image.repository=" + marsC.DockerRepository,
				"image.tag=" + tag,
			}

			j.project.DockerImage = fmt.Sprintf("%s:%s", marsC.DockerRepository, tag)

			j.Percenter().To(70)

			var valueOpts = &values.Options{
				ValueFiles: []string{mergedFile},
				Values:     append(append(commonValues, ingressConfig...), imagePullSecrets...),
			}

			indent, _ := json.MarshalIndent(append(append(commonValues, ingressConfig...), imagePullSecrets...), "", "\t")
			mlog.Debugf("values: %s", string(indent))

			j.Messager().SendMsg(fmt.Sprintf("使用的镜像是: %s", fmt.Sprintf("%s:%s", marsC.DockerRepository, b.String())))

			for key, secret := range marsC.ImagePullSecrets {
				valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("imagePullSecrets[%d].name=%s", key, secret))
				j.Messager().SendMsg(fmt.Sprintf("使用的imagepullsecrets是: %s", secret))
			}

			j.installer = newReleaseInstaller(j.project.Name, j.project.Namespace.Name, loadArchive, valueOpts, func(format string, v ...interface{}) {
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
						Type: "text",
					}
				}
			}, j.input.Atomic)

			image := strings.Split(j.project.DockerImage, ":")
			if len(image) == 2 {
				if plugins.GetDockerPlugin().ImageNotExists(image[0], image[1]) {
					return errors.New(fmt.Sprintf("镜像 %s 不存在！", j.project.DockerImage))
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
