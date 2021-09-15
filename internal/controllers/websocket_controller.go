package controllers

import (
	"context"
	"errors"
	"sync/atomic"

	"go.uber.org/config"
	"gopkg.in/yaml.v2"

	"github.com/duc-cnzj/mars/internal/response"

	"helm.sh/helm/v3/pkg/action"

	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/google/uuid"

	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/internal/mars"
	"helm.sh/helm/v3/pkg/chart"

	"k8s.io/client-go/kubernetes/scheme"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	v1 "k8s.io/api/apps/v1"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/gosimple/slug"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"
)

const (
	ResultError          string = "error"
	ResultSuccess        string = "success"
	ResultDeployed       string = "deployed"
	ResultDeployFailed   string = "deployed_failed"
	ResultDeployCanceled string = "deployed_canceled"
)

const (
	WsSetUid         string = "set_uid"
	WsReloadProjects string = "reload_projects"
	WsCancel         string = "cancel_project"
	WsCreateProject  string = "create_project"
	WsUpdateProject  string = "update_project"

	WsProcessPercent  string = "process_percent"
	WsClusterInfoSync string = "cluster_info_sync"
)

type AllWsConnections struct {
	sync.RWMutex
	conns map[string]map[string]*WsConn
}

func (awc *AllWsConnections) Add(uid string, conn *WsConn) {
	awc.Lock()
	defer awc.Unlock()
	if m, ok := awc.conns[uid]; ok {
		m[conn.id] = conn
	} else {
		awc.conns[uid] = map[string]*WsConn{conn.id: conn}
	}
}

func (awc *AllWsConnections) Delete(uid string, id string) {
	awc.Lock()
	defer awc.Unlock()
	if m, ok := awc.conns[uid]; ok {
		delete(m, id)
		if len(m) == 0 {
			delete(awc.conns, uid)
		}
	}
}

func (awc *AllWsConnections) SendExceptUid(uid, wsType, data string) {
	awc.RLock()
	defer awc.RUnlock()
	for u, uidconns := range awc.conns {
		if u != uid {
			for _, conn := range uidconns {
				SendMsg(conn, "", wsType, data)
			}
		}
	}
}

func (awc *AllWsConnections) SendExceptId(id, wsType, data string) {
	awc.RLock()
	defer awc.RUnlock()
	for _, uidconns := range awc.conns {
		for idx, conn := range uidconns {
			if idx != id {
				SendMsg(conn, "", wsType, data)
			}
		}
	}
}

func (awc *AllWsConnections) SendToAll(wsType, data string) {
	awc.Lock()
	defer awc.Unlock()
	for _, uidconns := range awc.conns {
		for _, conn := range uidconns {
			SendMsg(conn, "", wsType, data)
		}
	}
}

func (awc *AllWsConnections) Count() int {
	awc.RLock()
	defer awc.RUnlock()
	return len(awc.conns)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebsocketController struct {
	*AllWsConnections
}

func NewWebsocketController() *WebsocketController {
	wc := &WebsocketController{
		AllWsConnections: &AllWsConnections{conns: map[string]map[string]*WsConn{}},
	}

	ticker := time.NewTicker(15 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				marshal, _ := json.Marshal(utils.ClusterInfo())
				wc.SendToAll(WsClusterInfoSync, string(marshal))
			case <-utils.App().Done():
				mlog.Warning("app shutdown and stop WsClusterInfoSync")
				ticker.Stop()
				return
			}
		}
	}()

	return wc
}

type WsRequest struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type WsResponse struct {
	// 有可能同一个用户同时部署两个环境, 必须要有 slug 区分
	Slug   string `json:"slug"`
	Type   string `json:"type"`
	Result string `json:"result"`
	Data   string `json:"data"`
	End    bool   `json:"end"`
	Uid    string `json:"uid"`
	ID     string `json:"id"`
}

func (r *WsResponse) EncodeToBytes() []byte {
	marshal, _ := json.Marshal(&r)
	return marshal
}

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 5
)

type WsConn struct {
	sync.Mutex

	id  string
	uid string
	c   *websocket.Conn
	cs  *CancelSignals
}

type CancelSignals struct {
	cs map[string]*ProcessControl
	sync.RWMutex
}

func (jobs *CancelSignals) Remove(id string) {
	delete(jobs.cs, id)
}

func (jobs *CancelSignals) CancelAll() {
	mlog.Warning("cancel all!")
	for _, control := range jobs.cs {
		control.SendStopSignal()
	}
}

func (jobs *CancelSignals) Has(id string) bool {
	jobs.RLock()
	defer jobs.RUnlock()

	_, ok := jobs.cs[id]

	return ok
}
func (jobs *CancelSignals) Cancel(id string) {
	jobs.Lock()
	defer jobs.Unlock()
	if pc, ok := jobs.cs[id]; ok {
		pc.SendMsg("收到取消信号，开始停止部署！！！")
		pc.SendStopSignal()
		jobs.Remove(id)
	}
}

func (jobs *CancelSignals) Add(id string, pc *ProcessControl) {
	jobs.Lock()
	defer jobs.Unlock()
	jobs.cs[id] = pc
}

func (wc *WebsocketController) Info(ctx *gin.Context) {
	detail := map[string]interface{}{}
	wc.RLock()
	for s, m := range wc.conns {
		detail[s] = len(m)
	}
	wc.RUnlock()

	response.Success(ctx, 200, gin.H{
		"count":  len(wc.AllWsConnections.conns),
		"detail": detail,
	})
}

func (wc *WebsocketController) Ws(ctx *gin.Context) {
	c, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		mlog.Error(err)
		return
	}
	defer c.Close()
	var uid string
	uid = ctx.Query("uid")
	if uid == "" {
		uid = uuid.New().String()
	}
	id := uuid.New().String()
	var wsconn = &WsConn{id: id, uid: uid, c: c, cs: &CancelSignals{cs: map[string]*ProcessControl{}}}
	wc.Add(uid, wsconn)
	SendMsg(wsconn, "", WsSetUid, uid)

	c.SetReadLimit(maxMessageSize)
	// 未设置 c.SetReadDeadline()，所以不需要 ping/pong 续命
	c.SetPongHandler(func(string) error { mlog.Infof("收到心跳 id: %s, uid %s", id, uid); return nil })

	mlog.Debug("ws connected")

	for {
		var wsRequest WsRequest
		_, message, err := c.ReadMessage()
		if err != nil {
			wsconn.cs.CancelAll()
			wc.Delete(uid, id)
			mlog.Debug("read:", err, message)
			break
		}
		mlog.Debugf("receive msg %s", message)
		if err := json.Unmarshal(message, &wsRequest); err != nil {
			SendEndError(wsconn, "", "", err)
			continue
		}

		mlog.Debug("handle req", wsRequest)
		go wc.serveWebsocket(wsconn, wsRequest)
	}
}

func (wc *WebsocketController) serveWebsocket(c *WsConn, wsRequest WsRequest) {
	switch wsRequest.Type {
	case WsCancel:
		var input CancelInput
		if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
			mlog.Error(wsRequest.Data, &input)
			SendEndError(c, "", wsRequest.Type, err)
			return
		}

		// cancel
		var slugName = utils.Md5(fmt.Sprintf("%d-%s", input.NamespaceId, input.Name))
		if c.cs.Has(slugName) {
			input.Name = slug.Make(input.Name)
			var ns models.Namespace
			utils.DB().Where("`id` = ?", input.NamespaceId).First(&ns)
			c.cs.Cancel(slugName)
		}
	case WsCreateProject:
		var input ProjectInput
		if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
			mlog.Error(wsRequest.Data, &input)
			SendEndError(c, "", wsRequest.Type, err)
			return
		}

		wc.installProject(input, wsRequest.Type, wsRequest, c)
	case WsUpdateProject:
		var input UpdateProject
		if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
			mlog.Error(wsRequest.Data, &input)
			SendEndError(c, "", wsRequest.Type, err)
			return
		}
		var p models.Project
		if err := utils.DB().Where("`id` = ?", input.ProjectId).First(&p).Error; err != nil {
			mlog.Error(wsRequest.Data, &input)
			SendEndError(c, "", wsRequest.Type, err)
			return
		}

		wc.installProject(ProjectInput{
			NamespaceId:     p.NamespaceId,
			Name:            p.Name,
			GitlabProjectId: p.GitlabProjectId,
			GitlabBranch:    input.GitlabBranch,
			GitlabCommit:    input.GitlabCommit,
			Config:          input.Config,
			Atomic:          input.Atomic,
		}, wsRequest.Type, wsRequest, c)
	}
}

type UpdateProject struct {
	ProjectId int `json:"project_id"`

	GitlabBranch string `json:"gitlab_branch"`
	GitlabCommit string `json:"gitlab_commit"`
	Config       string `json:"config"`
	Atomic       bool   `json:"atomic"`
}

type ProjectInput struct {
	NamespaceId int `uri:"namespace_id" json:"namespace_id"`

	Name            string `json:"name"`
	GitlabProjectId int    `json:"gitlab_project_id"`
	GitlabBranch    string `json:"gitlab_branch"`
	GitlabCommit    string `json:"gitlab_commit"`
	Config          string `json:"config"`
	Atomic          bool   `json:"atomic"`
}

type CancelInput struct {
	NamespaceId int    `uri:"namespace_id" json:"namespace_id"`
	Name        string `json:"name"`
}

func (wc *WebsocketController) installProject(input ProjectInput, wsType string, wsRequest WsRequest, conn *WsConn) {
	// step 1: 初始化
	pc := &ProcessControl{
		input:     input,
		wsType:    wsType,
		wsRequest: wsRequest,
		conn:      conn,
		st:        time.Now(),
	}
	if err := pc.SetUp(); err != nil {
		pc.SendEndError(err)
		return
	}

	defer func() {
		pc.stopFunc(nil)
		pc.conn.cs.Remove(pc.slugName)
		pc.CallAfterInstalledFuncs()
		mlog.Warningf("done!!!!")
	}()

	var checkList = []func() error{
		pc.CheckConfig,
		pc.PrepareConfigFiles,
		pc.CheckImage,
	}

	for _, fn := range checkList {
		if err := fn(); err != nil {
			pc.Prune()
			pc.SendEndError(err)
			return
		}

		if pc.NeedStop() {
			pc.DoStop()
			return
		}
	}

	pc.SendMsg("准备部署...")

	// step 4: 开始部署
	pc.Run()

	// 返回结果
	pc.Wait()
	select {
	case <-pc.stopCtx.Done():
		wc.SendToAll(WsReloadProjects, "")
	default:
		wc.SendExceptId(conn.id, WsReloadProjects, "")
	}
}

// getPodSelectorsInDeploymentAndStatefulSetByManifest FIXME: 比较 hack
// 参考 https://github.com/kubernetes/client-go/issues/193#issuecomment-363240636
func getPodSelectorsInDeploymentAndStatefulSetByManifest(manifest string) []string {
	var selectors []string
	split := strings.Split(manifest, "---")
	for _, f := range split {
		obj, _, _ := scheme.Codecs.UniversalDeserializer().Decode([]byte(f), nil, nil)
		switch a := obj.(type) {
		case *v1.Deployment:
			mlog.Debug("############### getPodSelectorsInDeploymentAndStatefulSetByManifest(manifest string) ###############")
			var labels []string
			for k, v := range a.Spec.Selector.MatchLabels {
				labels = append(labels, fmt.Sprintf("%s=%s", k, v))
			}
			selectors = append(selectors, strings.Join(labels, ","))
		case *v1.StatefulSet:
			mlog.Debug("############### getPodSelectorsInDeploymentAndStatefulSetByManifest(manifest string) ###############")
			var labels []string
			for k, v := range a.Spec.Selector.MatchLabels {
				labels = append(labels, fmt.Sprintf("%s=%s", k, v))
			}
			selectors = append(selectors, strings.Join(labels, ","))
		}
	}

	return selectors
}

type MessageItem struct {
	Msg  string
	Type string
}

func SendEndError(conn *WsConn, slug, wsType string, err error) {
	res := &WsResponse{
		Slug:   slug,
		Type:   wsType,
		Result: ResultError,
		Data:   err.Error(),
		End:    true,
		Uid:    conn.uid,
		ID:     conn.id,
	}
	conn.Lock()
	defer conn.Unlock()
	conn.c.WriteMessage(websocket.TextMessage, res.EncodeToBytes())
}

func SendError(conn *WsConn, slug, wsType string, err error) {
	res := &WsResponse{
		Slug:   slug,
		Type:   wsType,
		Result: ResultError,
		Data:   err.Error(),
		End:    false,
		Uid:    conn.uid,
		ID:     conn.id,
	}
	conn.Lock()
	defer conn.Unlock()
	conn.c.WriteMessage(websocket.TextMessage, res.EncodeToBytes())
}

func SendProcessPercent(conn *WsConn, slug, percent string) {
	res := &WsResponse{
		Slug:   slug,
		Type:   WsProcessPercent,
		Result: ResultSuccess,
		End:    false,
		Data:   percent,
		Uid:    conn.uid,
		ID:     conn.id,
	}
	conn.Lock()
	defer conn.Unlock()
	conn.c.WriteMessage(websocket.TextMessage, res.EncodeToBytes())
}

func SendMsg(conn *WsConn, slug, wsType string, msg string) {
	res := &WsResponse{
		Slug:   slug,
		Type:   wsType,
		Result: ResultSuccess,
		End:    false,
		Data:   msg,
		Uid:    conn.uid,
		ID:     conn.id,
	}
	conn.Lock()
	defer conn.Unlock()
	conn.c.WriteMessage(websocket.TextMessage, res.EncodeToBytes())
}

func SendEndMsg(conn *WsConn, result, slug, wsType string, msg string) {
	res := &WsResponse{
		Slug:   slug,
		Type:   wsType,
		Result: result,
		End:    true,
		Data:   msg,
		Uid:    conn.uid,
		ID:     conn.id,
	}
	conn.Lock()
	defer conn.Unlock()
	conn.c.WriteMessage(websocket.TextMessage, res.EncodeToBytes())
}

type ProcessPercent struct {
	sync.RWMutex
	percent  int64
	slugName string
	conn     *WsConn
}

func NewProcessPercent(conn *WsConn, slug string, percent int64) *ProcessPercent {
	return &ProcessPercent{
		percent:  percent,
		slugName: slug,
		conn:     conn,
	}
}

func (pp *ProcessPercent) Current() int64 {
	pp.RLock()
	defer pp.RUnlock()

	return pp.percent
}

func (pp *ProcessPercent) Add() {
	pp.Lock()
	defer pp.Unlock()

	if pp.percent < 100 {
		pp.percent++
		SendProcessPercent(pp.conn, pp.slugName, fmt.Sprintf("%d", pp.percent))
	}
}

func (pp *ProcessPercent) To(percent int64) {
	pp.Lock()
	defer pp.Unlock()

	sleepTime := 100 * time.Millisecond
	for pp.percent < percent {
		time.Sleep(sleepTime)
		pp.percent++
		if sleepTime > 50*time.Millisecond {
			sleepTime = sleepTime / 2
		}
		SendProcessPercent(pp.conn, pp.slugName, fmt.Sprintf("%d", pp.percent))
	}
}

type atomicBool int32

func (b *atomicBool) isSet() bool { return atomic.LoadInt32((*int32)(b)) != 0 }
func (b *atomicBool) setTrue()    { atomic.StoreInt32((*int32)(b), 1) }
func (b *atomicBool) setFalse()   { atomic.StoreInt32((*int32)(b), 0) }

type ProcessControl struct {
	*ProcessPercent
	*MessageSender

	running atomicBool
	st      time.Time

	customFuncAfterInstalled []func()

	stopCtx  context.Context
	stopFunc func(error)

	input     ProjectInput
	wsType    string
	wsRequest WsRequest

	new       bool
	marC      *mars.Config
	project   models.Project
	conn      *WsConn
	slugName  string
	chart     *chart.Chart
	valueOpts *values.Options
	log       func(format string, v ...interface{})
	messageCh chan MessageItem
}

func (pc *ProcessControl) CallAfterInstalledFuncs() {
	for _, f := range pc.customFuncAfterInstalled {
		f()
	}
}

func (pc *ProcessControl) AddAfterInstalledFunc(fn func()) {
	pc.customFuncAfterInstalled = append(pc.customFuncAfterInstalled, fn)
}

func (pc *ProcessControl) Prune() {
	if pc.new {
		utils.DB().Delete(&pc.project)
	}
}

func (pc *ProcessControl) DoStop() {
	select {
	case <-pc.stopCtx.Done():
		pc.Prune()
		pc.SendEndMsg(ResultDeployCanceled, "收到停止信号")
	default:
		return
	}
}

func (pc *ProcessControl) SendStopSignal() {
	if pc.stopFunc != nil {
		pc.stopFunc(errors.New("receive canceled signal"))
	}
}

func (pc *ProcessControl) NeedStop() bool {
	if pc.stopCtx == nil {
		return false
	}
	select {
	case <-pc.stopCtx.Done():
		return true
	default:
		return false
	}
}

func (pc *ProcessControl) SetUp() error {
	pc.slugName = utils.Md5(fmt.Sprintf("%d-%s", pc.input.NamespaceId, pc.input.Name))
	pc.stopCtx, pc.stopFunc = utils.NewCustomErrorContext()
	pc.conn.cs.Add(pc.slugName, pc)
	pc.messageCh = make(chan MessageItem)
	pc.ProcessPercent = NewProcessPercent(pc.conn, pc.slugName, 0)
	pc.MessageSender = NewMessageSender(pc.conn, pc.slugName, pc.wsType)

	pc.SendMsg("收到请求，开始创建项目")
	pc.To(5)
	pc.SendMsg("校验传参...")

	var ns models.Namespace
	if err := utils.DB().Where("`id` = ?", pc.input.NamespaceId).First(&ns).Error; err != nil {
		return err
	}

	pc.project = models.Project{
		Name:            slug.Make(pc.input.Name),
		GitlabProjectId: pc.input.GitlabProjectId,
		GitlabBranch:    pc.input.GitlabBranch,
		GitlabCommit:    pc.input.GitlabCommit,
		Config:          pc.input.Config,
		NamespaceId:     ns.ID,
		Namespace:       ns,
		Atomic:          pc.input.Atomic,
	}

	var p models.Project
	if utils.DB().Where("`name` = ? AND `namespace_id` = ?", pc.project.Name, pc.project.NamespaceId).First(&p).Error == gorm.ErrRecordNotFound {
		utils.DB().Create(&pc.project)
		pc.new = true
	}

	return nil
}

func (pc *ProcessControl) CheckConfig() error {
	pc.SendMsg("校验项目配置传参...")
	pc.To(15)
	marsC, err := GetProjectMarsConfig(pc.input.GitlabProjectId, pc.input.GitlabBranch)
	if err != nil {
		return err
	}
	pc.marC = marsC

	// 下载 helm charts
	pc.SendMsg(fmt.Sprintf("下载 helm charts path: %s ...", marsC.LocalChartPath))
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
					pc.AddAfterInstalledFunc(depDeleteFn)
					pc.SendMsg(fmt.Sprintf("下载本地依赖 %s", dependency.Name))
				}
			}
		}
		pc.SendMsg(fmt.Sprintf("识别为远程仓库 uid %v branch %s path %s", pid, branch, path))
	} else {
		dir = marsC.LocalChartPath
		files = utils.GetDirectoryFiles(pc.input.GitlabProjectId, pc.input.GitlabCommit, marsC.LocalChartPath)
		tmpChartsDir, deleteDirFn = utils.DownloadFiles(pc.input.GitlabProjectId, pc.input.GitlabCommit, files)
	}
	pc.AddAfterInstalledFunc(deleteDirFn)
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

	pc.To(30)

	pc.SendMsg("加载 helm charts...")

	loadArchive, err := loader.LoadArchive(archive)
	if err != nil {
		return err
	}
	pc.chart = loadArchive

	return nil
}

func (pc *ProcessControl) PrepareConfigFiles() error {
	pc.SendMsg("生成配置文件...")
	pc.To(40)
	marsC := pc.marC
	loadArchive := pc.chart
	input := pc.input

	// default_values 也需要一个 file
	defaultValues, err := marsC.GenerateDefaultValuesYaml()
	if err != nil {
		return err
	}

	// 传入自定义配置必须在默认配置之后，不然会被上面的 default_values 覆盖，导致不管你怎么更新配置文件都无法正正的更新到容器
	configValues, err := marsC.GenerateConfigYamlByInput(input.Config)
	if err != nil {
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
	mergedFile, deleteFn, err := utils.WriteConfigYamlToTmpFile(bf.Bytes())
	if err != nil {
		return err
	}
	pc.AddAfterInstalledFunc(deleteFn)

	pc.SendMsg("解析镜像tag")
	pc.To(45)
	t := template.New("tag_parse")
	parse, err := t.Parse(marsC.DockerTagFormat)
	if err != nil {
		return err
	}
	b := &bytes.Buffer{}
	commit, _, err := utils.GitlabClient().Commits.GetCommit(pc.project.GitlabProjectId, pc.project.GitlabCommit)
	if err != nil {
		return err
	}
	var pipelineID int

	if commit.LastPipeline != nil {
		pipelineID = commit.LastPipeline.ID
	}

	pc.SendMsg(fmt.Sprintf("镜像分支 %s 镜像commit %s 镜像 pipeline_id %d", pc.project.GitlabBranch, pc.project.GitlabCommit, pipelineID))

	if err := parse.Execute(b, struct {
		Branch   string
		Commit   string
		Pipeline int
	}{
		Branch:   pc.project.GitlabBranch,
		Commit:   pc.project.GitlabCommit,
		Pipeline: pipelineID,
	}); err != nil {
		return err
	}

	pc.To(60)

	var ingressConfig []string
	if utils.Config().HasWildcardDomain() {
		var host, secretName string = utils.Config().GetDomain(fmt.Sprintf("%s-%s", pc.project.Name, pc.project.Namespace.Name)), fmt.Sprintf("%s-%s-tls", pc.project.Name, pc.project.Namespace.Name)
		var vars = map[string]string{}
		for i := 1; i <= 10; i++ {
			vars[fmt.Sprintf("Host%d", i)] = utils.Config().GetDomain(fmt.Sprintf("%s-%s-%d", pc.project.Name, pc.project.Namespace.Name, i))
			vars[fmt.Sprintf("TlsSecret%d", i)] = fmt.Sprintf("%s-%s-%d-tls", pc.project.Name, pc.project.Namespace.Name, i)
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
		if utils.Config().ClusterIssuer != "" {
			ingressConfig = append(ingressConfig, "ingress.annotations.cert\\-manager\\.io\\/cluster\\-issuer="+utils.Config().ClusterIssuer)
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

		pc.SendMsg(fmt.Sprintf("已配置域名: %s", host))
	}

	pc.To(65)

	tag := b.String()
	var commonValues = []string{
		"image.pullPolicy=IfNotPresent",
		"image.repository=" + marsC.DockerRepository,
		"image.tag=" + tag,
	}

	pc.project.DockerImage = fmt.Sprintf("%s:%s", marsC.DockerRepository, tag)

	var imagePullSecrets []string
	for k, s := range pc.project.Namespace.ImagePullSecretsArray() {
		imagePullSecrets = append(imagePullSecrets, fmt.Sprintf("imagePullSecrets[%d].name=%s", k, s))
		marsC.ImagePullSecrets = append(marsC.ImagePullSecrets, s)
	}

	pc.To(70)

	var valueOpts = &values.Options{
		ValueFiles: []string{mergedFile},
		Values:     append(append(commonValues, ingressConfig...), imagePullSecrets...),
	}

	indent, _ := json.MarshalIndent(append(append(commonValues, ingressConfig...), imagePullSecrets...), "", "\t")
	mlog.Debugf("values: %s", string(indent))

	pc.SendMsg(fmt.Sprintf("使用的镜像是: %s", fmt.Sprintf("%s:%s", marsC.DockerRepository, b.String())))

	for key, secret := range pc.project.Namespace.ImagePullSecretsArray() {
		valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("imagePullSecrets[%d].name=%s", key, secret))
		pc.SendMsg(fmt.Sprintf("使用的imagepullsecrets是: %s", secret))
	}
	pc.valueOpts = valueOpts

	pc.log = func(format string, v ...interface{}) {
		if pc.Current() < 99 {
			pc.Add()
		}
		msg := fmt.Sprintf(format, v...)
		mlog.Debug(msg)
		if pc.running.isSet() {
			pc.messageCh <- MessageItem{
				Msg:  msg,
				Type: "text",
			}
		}
	}

	return nil
}

func (pc *ProcessControl) Run() {
	pc.running.setTrue()
	ch := pc.messageCh
	loadArchive := pc.chart
	valueOpts := pc.valueOpts
	go func() {
		defer func() {
			pc.running.setFalse()
			close(ch)
		}()

		if result, err := utils.UpgradeOrInstall(pc.stopCtx, pc.project.Name, pc.project.Namespace.Name, loadArchive, valueOpts, pc.log, pc.input.Atomic); err != nil {
			mlog.Error(err)
			ch <- MessageItem{
				Msg:  err.Error(),
				Type: "error",
			}
		} else {
			coalesceValues, _ := chartutil.CoalesceValues(loadArchive, result.Config)
			pc.project.OverrideValues, _ = coalesceValues.YAML()
			pc.project.SetPodSelectors(getPodSelectorsInDeploymentAndStatefulSetByManifest(result.Manifest))
			var p models.Project
			if utils.DB().Where("`name` = ? AND `namespace_id` = ?", pc.project.Name, pc.project.NamespaceId).First(&p).Error == nil {
				utils.DB().Model(&models.Project{}).
					Select("Config", "GitlabProjectId", "GitlabCommit", "GitlabBranch", "DockerImage", "PodSelectors", "OverrideValues", "Atomic").
					Where("`id` = ?", p.ID).
					Updates(&pc.project)
			} else {
				utils.DB().Create(&pc.project)
			}
			pc.To(100)
			ch <- MessageItem{
				Msg:  "部署成功",
				Type: "success",
			}
		}
	}()
}

func (pc *ProcessControl) Wait() {
	for s := range pc.messageCh {
		switch s.Type {
		case "text":
			pc.SendMsg(s.Msg)
		case "error":
			if pc.new {
				utils.DB().Delete(&pc.project)
			}
			select {
			case <-pc.stopCtx.Done():
				pc.SendEndMsg(ResultDeployCanceled, s.Msg)
			default:
				pc.SendEndMsg(ResultDeployFailed, s.Msg)
			}
		case "success":
			pc.SendEndMsg(ResultDeployed, s.Msg)
		}
	}
}

func (pc *ProcessControl) CheckImage() error {
	image := strings.Split(pc.project.DockerImage, ":")
	if len(image) == 2 {
		if utils.ImageNotExists(image[0], image[1]) {
			return errors.New(fmt.Sprintf("镜像 %s 不存在！", pc.project.DockerImage))
		}
	}

	return nil
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
	SendEndError(ms.conn, ms.slugName, ms.wsType, err)
}

func (ms *MessageSender) SendError(err error) {
	SendError(ms.conn, ms.slugName, ms.wsType, err)
}

func (ms *MessageSender) SendProcessPercent(percent string) {
	SendProcessPercent(ms.conn, ms.slugName, percent)
}

func (ms *MessageSender) SendMsg(msg string) {
	SendMsg(ms.conn, ms.slugName, ms.wsType, msg)
}

func (ms *MessageSender) SendEndMsg(result, msg string) {
	SendEndMsg(ms.conn, result, ms.slugName, ms.wsType, msg)
}
