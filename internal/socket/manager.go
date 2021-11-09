package socket

import (
	"context"
	"errors"
	"regexp"
	"sync/atomic"

	"github.com/golang-jwt/jwt"

	"github.com/duc-cnzj/mars/internal/grpc/services"

	"github.com/duc-cnzj/mars/internal/enums"

	"github.com/duc-cnzj/mars/internal/plugins"

	"go.uber.org/config"
	"gopkg.in/yaml.v2"

	app "github.com/duc-cnzj/mars/internal/app/helper"
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
	WsSetUid         string = enums.WsSetUid
	WsReloadProjects string = enums.WsReloadProjects
	WsCancel         string = enums.WsCancel
	WsCreateProject  string = enums.WsCreateProject
	WsUpdateProject  string = enums.WsUpdateProject

	WsProcessPercent  string = enums.WsProcessPercent
	WsClusterInfoSync string = enums.WsClusterInfoSync

	WsInternalError string = enums.WsInternalError

	WsHandleExecShell    string = enums.WsHandleExecShell
	WsHandleExecShellMsg string = enums.WsHandleExecShellMsg
	WsHandleCloseShell   string = enums.WsHandleCloseShell
	WsAuthorize          string = enums.WsHandleAuthorize
)

var hostMatch = regexp.MustCompile(".*?=(.*?){{\\s*.Host\\d\\s*}}")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebsocketManager struct{}

func NewWebsocketManager() *WebsocketManager {
	wc := &WebsocketManager{}

	ticker := time.NewTicker(15 * time.Second)
	sub := plugins.GetWsSender().New("", "")
	go func() {
		for {
			select {
			case <-ticker.C:
				marshal, _ := json.Marshal(utils.ClusterInfo())

				sub.ToAll(&WsResponse{
					Type: WsClusterInfoSync,
					Data: string(marshal),
				})
			case <-app.App().Done():
				mlog.Info("[Websocket]: app shutdown and stop WsClusterInfoSync")
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

type WsResponse = plugins.WsResponse

const (
	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 5

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type WsConn struct {
	userMu sync.RWMutex

	user services.UserInfo
	ps   plugins.PubSub
	id   string
	uid  string
	c    *websocket.Conn
	cs   *CancelSignals

	terminalSessions *SessionMap
}

func (c *WsConn) SetUser(info services.UserInfo) {
	c.userMu.Lock()
	defer c.userMu.Unlock()
	c.user = info
}

func (c *WsConn) GetUser() services.UserInfo {
	c.userMu.RLock()
	defer c.userMu.RUnlock()
	return c.user
}

func (c *WsConn) GetShellChannel(sessionID string) (chan TerminalMessage, error) {
	if handler, ok := c.terminalSessions.Sessions[sessionID]; ok {
		return handler.shellCh, nil
	}

	return nil, fmt.Errorf("%v not found channel", sessionID)
}

func (c *WsConn) DeleteShellChannel(sessionID string) {
	ch := c.terminalSessions.Sessions[sessionID].shellCh
	for {
		select {
		case msg := <-ch:
			mlog.Warning("[Websocket] session: %v 未消费的数据 %v", sessionID, msg)
		default:
			close(ch)
			return
		}
	}
}

type CancelSignals struct {
	cs map[string]*ProcessControl
	sync.RWMutex
}

func (jobs *CancelSignals) Remove(id string) {
	delete(jobs.cs, id)
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

func (wc *WebsocketManager) Info(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	marshal, _ := json.Marshal(plugins.GetWsSender().New("", "").Info())
	writer.Write(marshal)
}

func (wc *WebsocketManager) Ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		mlog.Error(err)
		return
	}

	var uid string
	uid = r.URL.Query().Get("uid")
	if uid == "" {
		uid = uuid.New().String()
	}
	id := uuid.New().String()

	ps := plugins.GetWsSender().New(uid, id)
	var wsconn = &WsConn{
		ps:  ps,
		id:  id,
		uid: uid,
		c:   c,
		cs:  &CancelSignals{cs: map[string]*ProcessControl{}},
	}
	wsconn.terminalSessions = &SessionMap{Sessions: make(map[string]*MyPtyHandler), conn: wsconn}

	defer func() {
		mlog.Debug("[Websocket]: Ws exit ")
		wsconn.terminalSessions.CloseAll()
		ps.Close()
		c.Close()
		app.Metrics().DecWebsocketConn()
	}()

	app.Metrics().IncWebsocketConn()
	go write(wsconn)

	SendMsg(wsconn, "", WsSetUid, uid)

	ch := make(chan struct{}, 1)
	go func() {
		var err error
		defer func() {
			mlog.Debugf("[Websocket]: go read exit, err: %v", err)
		}()
		err = read(wsconn)
		ch <- struct{}{}
	}()

	select {
	case <-app.App().Done():
		return
	case <-ch:
		return
	}
}

func write(wsconn *WsConn) error {
	defer utils.HandlePanic()

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		mlog.Debugf("[Websocket]: go write exit")
		ticker.Stop()
		wsconn.c.Close()
	}()
	ch := wsconn.ps.Subscribe()
	for {
		select {
		case message, ok := <-ch:
			wsconn.c.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				return wsconn.c.WriteMessage(websocket.CloseMessage, []byte{})
			}

			w, err := wsconn.c.NextWriter(websocket.TextMessage)
			if err != nil {
				return err
			}
			w.Write([]byte(message))

			if err := w.Close(); err != nil {
				return err
			}
		case <-ticker.C:
			mlog.Debugf("[Websocket] tick ping/pong uid: %s, id: %s", wsconn.uid, wsconn.id)
			wsconn.c.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wsconn.c.WriteMessage(websocket.PingMessage, nil); err != nil {
				return err
			}
		}
	}
}

func read(wsconn *WsConn) error {
	wsconn.c.SetReadLimit(maxMessageSize)
	wsconn.c.SetReadDeadline(time.Now().Add(pongWait))
	wsconn.c.SetPongHandler(func(string) error {
		wsconn.c.SetReadDeadline(time.Now().Add(pongWait))
		mlog.Debugf("[Websocket] 收到心跳 id: %s, uid %s", wsconn.id, wsconn.uid)
		return nil
	})
	for {
		var wsRequest WsRequest
		_, message, err := wsconn.c.ReadMessage()
		if err != nil {
			mlog.Debugf("[Websocket] read error:", err, message)
			return err
		}
		if err := json.Unmarshal(message, &wsRequest); err != nil {
			SendEndError(wsconn, "", WsInternalError, err)
			continue
		}

		go serveWebsocket(wsconn, wsRequest)
	}
}

type Token struct {
	Token string `json:"token"`
}

func serveWebsocket(c *WsConn, wsRequest WsRequest) {
	defer utils.HandlePanic()
	mlog.Infof("[Websocket]: user: %v, type: %v, data: %v.", c.GetUser().Name, wsRequest.Type, wsRequest.Data)
	switch wsRequest.Type {
	case WsAuthorize:
		var input Token
		if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
			mlog.Error(wsRequest.Data, &input)
			SendEndError(c, "", wsRequest.Type, err)
			return
		}
		var token = strings.TrimSpace(strings.TrimLeft(input.Token, "Bearer"))
		parse, err := jwt.ParseWithClaims(token, &services.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
			return app.Config().Pubkey(), nil
		})
		if err == nil && parse.Valid {
			claims, _ := parse.Claims.(*services.JwtClaims)
			c.SetUser(claims.UserInfo)
		}
	case WsHandleCloseShell:
		var input TerminalMessage
		if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
			mlog.Error(wsRequest.Data, &input)
			SendEndError(c, "", wsRequest.Type, err)
			return
		}
		mlog.Debugf("[Websocket] %v 收到客户端主动断开的消息", input.SessionID)
		c.terminalSessions.Close(input.SessionID, 0, "")
	case WsHandleExecShellMsg:
		var input TerminalMessage
		if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
			mlog.Error(wsRequest.Data, &input)
			SendEndError(c, "", wsRequest.Type, err)
			return
		}
		go func() {
			defer utils.HandlePanic()
			if input.SessionID != "" {
				messages, err := c.GetShellChannel(input.SessionID)
				if err != nil {
					mlog.Error(err)
					return
				}
				mlog.Debugf("[Websocket] %v 收到 WsHandleExecShellMsg 消息: '%v' , op: %v chan size: %v", input.SessionID, input.Data, input.Op, len(messages))
				messages <- input
				mlog.Debugf("[Websocket] %v msg send: '%v'", input.SessionID, input.Data)
			}
		}()
	case WsHandleExecShell:
		var input WsHandleExecShellInput
		if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
			mlog.Error(wsRequest.Data, &input)
			SendEndError(c, "", wsRequest.Type, err)
			return
		}

		sessionID, err := HandleExecShell(input, c)
		if err != nil {
			mlog.Error(err)
			SendEndMsg(c, ResultError, "", WsHandleExecShell, err.Error())
			return
		}
		mlog.Debugf("[Websocket] 收到 初始化连接 WsHandleExecShell 消息, id: %v", sessionID)

		res := struct {
			WsHandleExecShellInput
			SessionID string `json:"session_id"`
		}{
			WsHandleExecShellInput: input,
			SessionID:              sessionID,
		}
		marshal, _ := json.Marshal(res)
		SendEndMsg(c, ResultSuccess, "", WsHandleExecShell, string(marshal))
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
			c.cs.Cancel(slugName)
		}
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
		if err := app.DB().Where("`id` = ?", input.ProjectId).First(&p).Error; err != nil {
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
		conn.ps.ToAll(&WsResponse{Type: WsReloadProjects})
	default:
		conn.ps.ToOthers(&WsResponse{Type: WsReloadProjects})
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
	if err := app.DB().Where("`id` = ?", pc.input.NamespaceId).First(&ns).Error; err != nil {
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
	if app.DB().Where("`name` = ? AND `namespace_id` = ?", pc.project.Name, pc.project.NamespaceId).First(&p).Error == gorm.ErrRecordNotFound {
		app.DB().Create(&pc.project)
		pc.new = true
	}

	return nil
}

func (pc *ProcessControl) PrepareConfigFiles() error {
	pc.SendMsg("生成配置文件...")
	pc.To(40)
	marsC := pc.marC
	loadArchive := pc.chart
	input := pc.input

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
	if input.Config != "" {
		configValues, err = marsC.GenerateConfigYamlByInput(input.Config)
		if err != nil {
			return err
		}
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
	pc.AddAfterInstalledFunc(func() { closer.Close() })

	pc.SendMsg("解析镜像tag")
	pc.To(45)
	t := template.New("tag_parse")
	parse, err := t.Parse(marsC.DockerTagFormat)
	if err != nil {
		return err
	}
	b := &bytes.Buffer{}
	commit, _, err := app.GitlabClient().Commits.GetCommit(pc.project.GitlabProjectId, pc.project.GitlabCommit)
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

	if app.Config().HasWildcardDomain() {
		sub := getPreOccupiedLen(marsC.IngressOverwriteValues)
		var host, secretName string = getDomain(pc.project.Name, pc.project.Namespace.Name, sub), fmt.Sprintf("%s-%s-tls", pc.project.Name, pc.project.Namespace.Name)
		var vars = map[string]string{}
		for i := 1; i <= 10; i++ {
			vars[fmt.Sprintf("Host%d", i)] = getDomainByIndex(pc.project.Name, pc.project.Namespace.Name, i, sub)
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

	pc.To(70)

	var valueOpts = &values.Options{
		ValueFiles: []string{mergedFile},
		Values:     append(append(commonValues, ingressConfig...), imagePullSecrets...),
	}

	indent, _ := json.MarshalIndent(append(append(commonValues, ingressConfig...), imagePullSecrets...), "", "\t")
	mlog.Debugf("values: %s", string(indent))

	pc.SendMsg(fmt.Sprintf("使用的镜像是: %s", fmt.Sprintf("%s:%s", marsC.DockerRepository, b.String())))

	for key, secret := range marsC.ImagePullSecrets {
		valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("imagePullSecrets[%d].name=%s", key, secret))
		pc.SendMsg(fmt.Sprintf("使用的imagepullsecrets是: %s", secret))
	}
	pc.valueOpts = valueOpts

	pc.log = func(format string, v ...interface{}) {
		if pc.Current() < 99 {
			pc.Add()
		}
		if pc.Current() >= 95 {
			format = "[如果长时间未部署成功，建议取消使用 debug 模式]: " + format
		}
		msg := fmt.Sprintf(format, v...)
		if pc.running.isSet() {
			pc.messageCh <- MessageItem{
				Msg:  msg,
				Type: "text",
			}
		}
	}

	return nil
}

func (pc *ProcessControl) CheckConfig() error {
	pc.SendMsg("校验项目配置传参...")
	pc.To(15)
	marsC, err := services.GetProjectMarsConfig(pc.input.GitlabProjectId, pc.input.GitlabBranch)
	if err != nil {
		return err
	}
	marsC.ImagePullSecrets = pc.project.Namespace.ImagePullSecretsArray()
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

func (pc *ProcessControl) CheckImage() error {
	image := strings.Split(pc.project.DockerImage, ":")
	if len(image) == 2 {
		if plugins.GetDockerPlugin().ImageNotExists(image[0], image[1]) {
			return errors.New(fmt.Sprintf("镜像 %s 不存在！", pc.project.DockerImage))
		}
	}

	return nil
}

func (pc *ProcessControl) Prune() {
	if pc.new {
		app.DB().Delete(&pc.project)
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

func (pc *ProcessControl) DoStop() {
	select {
	case <-pc.stopCtx.Done():
		pc.Prune()
		pc.SendEndMsg(ResultDeployCanceled, "收到停止信号")
	default:
		return
	}
}

func (pc *ProcessControl) CallAfterInstalledFuncs() {
	for _, f := range pc.customFuncAfterInstalled {
		f()
	}
}

func (pc *ProcessControl) AddAfterInstalledFunc(fn func()) {
	pc.customFuncAfterInstalled = append(pc.customFuncAfterInstalled, fn)
}

func (pc *ProcessControl) Run() {
	pc.running.setTrue()
	ch := pc.messageCh
	loadArchive := pc.chart
	valueOpts := pc.valueOpts
	go func() {
		defer utils.HandlePanic()
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
			if app.DB().Where("`name` = ? AND `namespace_id` = ?", pc.project.Name, pc.project.NamespaceId).First(&p).Error == nil {
				app.DB().Model(&models.Project{}).
					Select("Config", "GitlabProjectId", "GitlabCommit", "GitlabBranch", "DockerImage", "PodSelectors", "OverrideValues", "Atomic").
					Where("`id` = ?", p.ID).
					Updates(&pc.project)
			} else {
				app.DB().Create(&pc.project)
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
				app.DB().Delete(&pc.project)
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

func getPreOccupiedLen(values []string) int {
	var sub = 0
	if len(values) > 0 {
		for _, value := range values {
			submatch := hostMatch.FindAllStringSubmatch(value, -1)
			if len(submatch) == 1 && len(submatch[0]) >= 1 {
				sub = max(sub, len(submatch[0][1]))
			}
		}
	}
	return sub
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
	conn.ps.ToSelf(res)
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
	conn.ps.ToSelf(res)
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
	conn.ps.ToSelf(res)
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
	conn.ps.ToSelf(res)
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
	conn.ps.ToSelf(res)
}

func getDomain(project, namespace string, preOccupiedLen int) string {
	if !app.Config().HasWildcardDomain() {
		return ""
	}

	return plugins.GetDomainResolverPlugin().GetDomain(strings.TrimLeft(app.Config().WildcardDomain, "*."), project, namespace, preOccupiedLen)
}

func getDomainByIndex(project, namespace string, index, preOccupiedLen int) string {
	if !app.Config().HasWildcardDomain() {
		return ""
	}

	return plugins.GetDomainResolverPlugin().GetDomainByIndex(strings.TrimLeft(app.Config().WildcardDomain, "*."), project, namespace, index, preOccupiedLen)
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
