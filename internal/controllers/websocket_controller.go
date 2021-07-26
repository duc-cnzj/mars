package controllers

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"

	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/internal/mars"
	"helm.sh/helm/v3/pkg/chart"

	"gopkg.in/yaml.v2"
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

	WsProcessPercent string = "process_percent"
)

type AllWsConnections struct {
	sync.RWMutex
	conns map[string]map[string]*WsConn
}

func (awc *AllWsConnections) Add(uid string, conn *WsConn) {
	awc.Lock()
	defer awc.Unlock()
	if awc.conns == nil {
		awc.conns = map[string]map[string]*WsConn{}
	}
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

func (awc *AllWsConnections) SendExcept(uid, wsType, data string) {
	awc.Lock()
	defer awc.Unlock()
	for u, uidconns := range awc.conns {
		if u != uid {
			for _, conn := range uidconns {
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

var awc *AllWsConnections

type WebsocketController struct {
	*AllWsConnections
}

func NewWebsocketController() *WebsocketController {
	awc = &AllWsConnections{conns: nil}
	if utils.App().IsDebug() {
		go func() {
			t := time.NewTicker(1 * time.Second)
			for {
				select {
				case <-t.C:
					mlog.Warningf("################################# 一共有 %d 个客户端 #################################", len(awc.conns))
					for s, m := range awc.conns {
						mlog.Warningf("当前uid：%s 该uid下面的连接数: %d", s, len(m))
					}
				}
			}
		}()
	}
	return &WebsocketController{
		AllWsConnections: awc,
	}
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
	sync.Mutex
}

func (jobs *CancelSignals) Remove(id string) {
	if _, ok := jobs.cs[id]; ok {
		delete(jobs.cs, id)
	}
}

func (jobs *CancelSignals) CancelAll() {
	mlog.Warning("cancel all!")
	for _, control := range jobs.cs {
		control.SendStopSignal()
	}
}

func (jobs *CancelSignals) Cancel(id string) {
	jobs.Lock()
	defer jobs.Unlock()
	if pc, ok := jobs.cs[id]; ok {
		if pc.running {
			pc.SendError(errors.New("已经开始部署了，无法停止！！！！！！！"))
			return
		}
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
	c.SetPongHandler(func(string) error { c.SetReadDeadline(time.Now().Add(pongWait)); return nil })

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
		go serveWebsocket(wsconn, wsRequest)
	}
}

func serveWebsocket(c *WsConn, wsRequest WsRequest) {
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
		input.Name = slug.Make(input.Name)
		var ns models.Namespace
		utils.DB().Where("`id` = ?", input.NamespaceId).First(&ns)
		c.cs.Cancel(slugName)
	case WsCreateProject:
		var input ProjectInput
		if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
			mlog.Error(wsRequest.Data, &input)
			SendEndError(c, "", wsRequest.Type, err)
			return
		}

		installProject(input, wsRequest.Type, wsRequest, c)
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

		installProject(ProjectInput{
			NamespaceId:     p.NamespaceId,
			Name:            p.Name,
			GitlabProjectId: p.GitlabProjectId,
			GitlabBranch:    input.GitlabBranch,
			GitlabCommit:    input.GitlabCommit,
			Config:          input.Config,
		}, wsRequest.Type, wsRequest, c)
	}
}

type UpdateProject struct {
	ProjectId int `json:"project_id"`

	GitlabBranch string `json:"gitlab_branch"`
	GitlabCommit string `json:"gitlab_commit"`
	Config       string `json:"config"`
}

type ProjectInput struct {
	NamespaceId int `uri:"namespace_id" json:"namespace_id"`

	Name            string `json:"name"`
	GitlabProjectId int    `json:"gitlab_project_id"`
	GitlabBranch    string `json:"gitlab_branch"`
	GitlabCommit    string `json:"gitlab_commit"`
	Config          string `json:"config"`
}

type CancelInput struct {
	NamespaceId int    `uri:"namespace_id" json:"namespace_id"`
	Name        string `json:"name"`
}

func installProject(input ProjectInput, wsType string, wsRequest WsRequest, conn *WsConn) {
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
		pc.stopFunc()
		pc.conn.cs.Remove(pc.slugName)
		pc.CallAfterInstalledFuncs()
		mlog.Warningf("done!!!!")
	}()

	var checkList = []func() error{
		pc.CheckConfig,
		pc.PrepareConfigFiles,
	}

	for _, fn := range checkList {
		if err := fn(); err != nil {
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
		awc.SendToAll(WsReloadProjects, "")
	default:
		awc.SendExcept(conn.uid, WsReloadProjects, "")
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

type ProcessControl struct {
	*ProcessPercent
	*MessageSender

	running bool
	st      time.Time

	customFuncAfterInstalled []func()

	stopCtx  context.Context
	stopFunc func()

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

func (pc *ProcessControl) DoStop() {
	select {
	case <-pc.stopCtx.Done():
		if pc.new {
			utils.DB().Delete(&pc.project)
		}
		pc.SendEndMsg(ResultDeployCanceled, "收到停止信号")
	default:
		return
	}
}

func (pc *ProcessControl) SendStopSignal() {
	if pc.stopFunc != nil {
		pc.stopFunc()
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
	pc.stopCtx, pc.stopFunc = context.WithCancel(context.Background())
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
	intPid := func(pid string) bool {
		if _, err := strconv.ParseInt(pid, 10, 64); err == nil {
			return true
		}
		return false
	}
	var (
		files        []string
		tmpChartsDir string
		deleteDirFn  func()
		dir          string
	)
	// uid|branch|path
	if len(split) == 3 && intPid(split[0]) {
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

	filePath, deleteFn, err := marsC.GenerateConfigYamlFileByInput(input.Config)
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
			"ingress.annotations.cert\\-manager\\.io\\/cluster\\-issuer=" + utils.Config().ClusterIssuer,
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
	}

	// default_values 也需要一个 file
	file, deleteDefaultValuesFileFn, err := marsC.GenerateDefaultValuesYamlFile()
	if err != nil {
		return err
	}

	pc.To(70)
	pc.AddAfterInstalledFunc(deleteDefaultValuesFileFn)
	var vf []string

	if filePath != "" {
		vf = append(vf, filePath)
	}

	if file != "" {
		vf = append(vf, file)
	}

	var valueOpts = &values.Options{
		ValueFiles: vf,
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
		pc.messageCh <- MessageItem{
			Msg:  msg,
			Type: "text",
		}
	}

	return nil
}

func (pc *ProcessControl) Run() {
	pc.running = true
	ch := pc.messageCh
	loadArchive := pc.chart
	valueOpts := pc.valueOpts

	go func() {
		if result, err := utils.UpgradeOrInstall(pc.project.Name, pc.project.Namespace.Name, loadArchive, valueOpts, pc.log); err != nil {
			mlog.Error(err)
			ch <- MessageItem{
				Msg:  err.Error(),
				Type: "error",
			}
			close(ch)
		} else {
			bf := &bytes.Buffer{}
			yaml.NewEncoder(bf).Encode(&result.Config)
			pc.project.OverrideValues = bf.String()
			pc.project.SetPodSelectors(getPodSelectorsInDeploymentAndStatefulSetByManifest(result.Manifest))
			var p models.Project
			if utils.DB().Where("`name` = ? AND `namespace_id` = ?", pc.project.Name, pc.project.NamespaceId).First(&p).Error == nil {
				utils.DB().Model(&models.Project{}).
					Select("Config", "GitlabProjectId", "GitlabCommit", "GitlabBranch", "DockerImage", "PodSelectors", "OverrideValues").
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
			close(ch)
		}
	}()
}

func (pc *ProcessControl) Wait() {
	for s := range pc.messageCh {
		switch s.Type {
		case "text":
			pc.SendMsg(s.Msg)
		case "error":
			pc.SendEndMsg(ResultDeployFailed, s.Msg)
		case "success":
			pc.SendEndMsg(ResultDeployed, s.Msg)
		}
	}
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
