package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/DuC-cnZj/mars/pkg/mlog"
	"github.com/DuC-cnZj/mars/pkg/models"
	"github.com/DuC-cnZj/mars/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/gosimple/slug"
	"github.com/xanzy/go-gitlab"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"
)

const (
	ResultError        string = "error"
	ResultSuccess      string = "success"
	ResultDeployed     string = "deployed"
	ResultDeployFailed string = "deployed_failed"
)

const (
	WsCreateProject string = "create_project"
	WsUpdateProject string = "update_project"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebsocketController struct{}

func NewWebsocketController() *WebsocketController {
	return &WebsocketController{}
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
	c *websocket.Conn
}

func (*WebsocketController) Ws(ctx *gin.Context) {
	c, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		mlog.Error(err)
		return
	}
	defer c.Close()

	var wsconn = &WsConn{c: c}

	c.SetReadLimit(maxMessageSize)
	c.SetPongHandler(func(string) error { c.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	mlog.Debug("ws connected")

	for {
		var wsRequest WsRequest
		_, message, err := c.ReadMessage()
		if err != nil {
			mlog.Warning("read:", err, message)
			break
		}
		mlog.Infof("receive msg %s", message)
		if err := json.Unmarshal(message, &wsRequest); err != nil {
			SendEndError(wsconn, "", "", err)
			continue
		}

		mlog.Info("handle req", wsRequest)
		go serveWebsocket(wsconn, wsRequest)
	}
}

func serveWebsocket(c *WsConn, wsRequest WsRequest) {
	switch wsRequest.Type {
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

func installProject(input ProjectInput, wsType string, wsRequest WsRequest, conn *WsConn) {
	var slugName = utils.Md5(fmt.Sprintf("%d-%s", input.NamespaceId, input.Name))
	SendMsg(conn, slugName, wsRequest.Type, "收到请求，开始创建项目")

	var ns models.Namespace
	if err := utils.DB().Where("`id` = ?", input.NamespaceId).First(&ns).Error; err != nil {
		mlog.Error(err)
		SendEndError(conn, "", wsType, err)
		return
	}
	input.Name = slug.Make(input.Name)

	SendMsg(conn, slugName, wsType, "校验传参...")

	var project = models.Project{
		Name:            input.Name,
		GitlabProjectId: input.GitlabProjectId,
		GitlabBranch:    input.GitlabBranch,
		GitlabCommit:    input.GitlabCommit,
		Config:          input.Config,
		NamespaceId:     ns.ID,
	}

	var namespace models.Namespace

	utils.DB().Where("`id` = ?", ns.ID).First(&namespace)

	SendMsg(conn, slugName, wsType, "校验项目配置传参...")

	marsC, err := GetProjectMarsConfig(input.GitlabProjectId, input.GitlabBranch)
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}

	file, _, err := utils.GitlabClient().RepositoryFiles.GetFile(input.GitlabProjectId, marsC.LocalChartPath, &gitlab.GetFileOptions{Ref: gitlab.String(input.GitlabBranch)})
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}
	archive, _ := base64.StdEncoding.DecodeString(file.Content)

	SendMsg(conn, slugName, wsType, "加载 helm charts...")

	loadArchive, err := loader.LoadArchive(bytes.NewReader(archive))
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}

	SendMsg(conn, slugName, wsType, "生成配置文件...")
	filePath, deleteFn, err := marsC.GenerateConfigYamlFileByInput(input.Config)
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}
	defer deleteFn()

	SendMsg(conn, slugName, wsType, "解析镜像tag")
	t := template.New("tag_parse")
	parse, err := t.Parse(marsC.DockerTagFormat)
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}
	b := &bytes.Buffer{}
	commit, _, err := utils.GitlabClient().Commits.GetCommit(project.GitlabProjectId, project.GitlabCommit)
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}
	var pipelineID int

	if commit.LastPipeline != nil {
		pipelineID = commit.LastPipeline.ID
	}

	SendMsg(conn, slugName, wsType, fmt.Sprintf("镜像分支 %s 镜像commit %s 镜像 pipeline_id %d", project.GitlabBranch, project.GitlabCommit, pipelineID))

	if err := parse.Execute(b, struct {
		Branch   string
		Commit   string
		Pipeline int
	}{
		Branch:   project.GitlabBranch,
		Commit:   project.GitlabCommit,
		Pipeline: pipelineID,
	}); err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}

	var ingressConfig []string
	if utils.Config().HasWildcardDomain() {
		var host, secretName string = utils.Config().GetDomain(fmt.Sprintf("%s-%s", project.Name, namespace.Name)), fmt.Sprintf("%s-%s-tls", project.Name, namespace.Name)
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
				mlog.Info(string(f.Data))
				if strings.Contains(string(f.Data), "path: {{ . }}") {
					isOldVersion = true
					break
				}
			}
		}
		ingressConfig = []string{
			"ingress.enabled=true",
			"ingress.hosts[0].host=" + host,
			"ingress.tls[0].secretName=" + secretName,
			"ingress.tls[0].hosts[0]=" + host,
			"ingress.annotations.kubernetes\\.io\\/ingress\\.class=nginx",
			"ingress.annotations.cert\\-manager\\.io\\/cluster\\-issuer=" + utils.Config().ClusterIssuer,
		}
		if isOldVersion {
			ingressConfig = append(ingressConfig, "ingress.hosts[0].paths[0]=/")
		} else {
			ingressConfig = append(ingressConfig, "ingress.hosts[0].paths[0].path=/")
		}

		SendMsg(conn, slugName, wsType, fmt.Sprintf("已配置域名: %s", host))
	}

	var commonValues = []string{
		"image.pullPolicy=IfNotPresent",
		"image.repository=" + marsC.DockerRepository,
		"image.tag=" + b.String(),
	}

	var imagePullSecrets []string
	for k, s := range namespace.ImagePullSecretsArray() {
		imagePullSecrets = append(imagePullSecrets, fmt.Sprintf("imagePullSecrets[%d].name=%s", k, s))
	}

	var valueOpts = &values.Options{
		ValueFiles: []string{filePath},
		Values:     append(append(append(commonValues, ingressConfig...), marsC.DefaultValues...), imagePullSecrets...),
	}

	indent, _ := json.MarshalIndent(append(append(append(commonValues, ingressConfig...), marsC.DefaultValues...), imagePullSecrets...), "", "\t")
	mlog.Warningf("values: %s", string(indent))

	SendMsg(conn, slugName, wsType, fmt.Sprintf("使用的镜像是: %s", fmt.Sprintf("%s:%s", marsC.DockerRepository, b.String())))

	for key, secret := range namespace.ImagePullSecretsArray() {
		valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("imagePullSecrets[%d].name=%s", key, secret))
		SendMsg(conn, slugName, wsType, fmt.Sprintf("使用的imagepullsecrets是: %s", secret))
	}

	valueOpts.Values = append(valueOpts.Values, marsC.DefaultValues...)

	ch := make(chan MessageItem)
	fn := func(format string, v ...interface{}) {
		msg := fmt.Sprintf(format, v...)
		mlog.Debug(msg)
		ch <- MessageItem{
			Msg:  msg,
			Type: "text",
		}
	}

	SendMsg(conn, slugName, wsType, "准备部署...")

	go func() {
		if _, err := utils.UpgradeOrInstall(input.Name, ns.Name, loadArchive, valueOpts, fn); err != nil {
			mlog.Error(err)
			ch <- MessageItem{
				Msg:  err.Error(),
				Type: "error",
			}
			close(ch)
		} else {
			if utils.DB().Where("`name` = ? AND `namespace_id` = ?", input.Name, ns.ID).First(&project).Error == nil {
				utils.DB().Model(&models.Project{}).Where("`id` = ?", project.ID).Updates(map[string]interface{}{
					"config":            input.Config,
					"gitlab_project_id": input.GitlabProjectId,
					"gitlab_commit":     input.GitlabCommit,
					"gitlab_branch":     input.GitlabBranch,
				})
			} else {
				utils.DB().Create(&project)
			}
			ch <- MessageItem{
				Msg:  "部署成功",
				Type: "success",
			}
			close(ch)
		}
	}()

	for s := range ch {
		switch s.Type {
		case "text":
			SendMsg(conn, slugName, wsType, s.Msg)
		case "error":
			SendEndMsg(conn, ResultDeployFailed, slugName, wsType, s.Msg)
		case "success":
			SendEndMsg(conn, ResultDeployed, slugName, wsType, s.Msg)
		}
	}
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
	}
	conn.Lock()
	defer conn.Unlock()
	conn.c.WriteMessage(websocket.TextMessage, res.EncodeToBytes())
}
