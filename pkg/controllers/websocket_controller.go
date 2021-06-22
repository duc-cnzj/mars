package controllers

import (
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

	"github.com/duc-cnzj/mars/pkg/mlog"
	"github.com/duc-cnzj/mars/pkg/models"
	"github.com/duc-cnzj/mars/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/gosimple/slug"
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

	WsProcessPercent string = "process_percent"
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
	var pp = NewProcessPercent(conn, slugName, 0)
	pp.To(5)

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
	pp.To(15)

	marsC, err := GetProjectMarsConfig(input.GitlabProjectId, input.GitlabBranch)
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}

	// 下载 helm charts
	SendMsg(conn, slugName, wsType, fmt.Sprintf("下载 helm charts path: %s ...", marsC.LocalChartPath))
	files := utils.GetDirectoryFiles(input.GitlabProjectId, input.GitlabCommit, marsC.LocalChartPath)
	tmpChartsDir, deleteDirFn := utils.DownloadFiles(input.GitlabProjectId, input.GitlabCommit, files)
	defer deleteDirFn()
	chartDir := filepath.Join(tmpChartsDir, marsC.LocalChartPath)
	chart, err := utils.PackageChart(chartDir, chartDir)
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}
	archive, err := os.Open(chart)
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}
	defer archive.Close()

	pp.To(30)

	SendMsg(conn, slugName, wsType, "加载 helm charts...")

	loadArchive, err := loader.LoadArchive(archive)
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}

	SendMsg(conn, slugName, wsType, "生成配置文件...")
	pp.To(40)

	filePath, deleteFn, err := marsC.GenerateConfigYamlFileByInput(input.Config)
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}
	defer deleteFn()

	SendMsg(conn, slugName, wsType, "解析镜像tag")
	pp.To(45)
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

	pp.To(60)

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

	pp.To(65)

	tag := b.String()
	var commonValues = []string{
		"image.pullPolicy=IfNotPresent",
		"image.repository=" + marsC.DockerRepository,
		"image.tag=" + tag,
	}

	project.DockerImage = fmt.Sprintf("%s:%s", marsC.DockerRepository, tag)

	var imagePullSecrets []string
	for k, s := range namespace.ImagePullSecretsArray() {
		imagePullSecrets = append(imagePullSecrets, fmt.Sprintf("imagePullSecrets[%d].name=%s", k, s))
	}

	// default_values 也需要一个 file
	file, deleteDefaultValuesFileFn, err := marsC.GenerateDefaultValuesYamlFile()
	if err != nil {
		SendEndError(conn, slugName, wsType, err)
		return
	}

	pp.To(70)
	defer deleteDefaultValuesFileFn()
	var valueOpts = &values.Options{
		ValueFiles: []string{filePath, file},
		Values:     append(append(commonValues, ingressConfig...), imagePullSecrets...),
	}

	indent, _ := json.MarshalIndent(append(append(commonValues, ingressConfig...), imagePullSecrets...), "", "\t")
	mlog.Warningf("values: %s", string(indent))

	SendMsg(conn, slugName, wsType, fmt.Sprintf("使用的镜像是: %s", fmt.Sprintf("%s:%s", marsC.DockerRepository, b.String())))

	for key, secret := range namespace.ImagePullSecretsArray() {
		valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("imagePullSecrets[%d].name=%s", key, secret))
		SendMsg(conn, slugName, wsType, fmt.Sprintf("使用的imagepullsecrets是: %s", secret))
	}

	ch := make(chan MessageItem)
	fn := func(format string, v ...interface{}) {
		pp.AddOne()
		msg := fmt.Sprintf(format, v...)
		mlog.Debug(msg)
		ch <- MessageItem{
			Msg:  msg,
			Type: "text",
		}
	}

	SendMsg(conn, slugName, wsType, "准备部署...")

	go func() {
		if result, err := utils.UpgradeOrInstall(input.Name, ns.Name, loadArchive, valueOpts, fn); err != nil {
			mlog.Error(err)
			ch <- MessageItem{
				Msg:  err.Error(),
				Type: "error",
			}
			close(ch)
		} else {
			project.SetPodSelectors(getPodSelectorsInDeploymentAndStatefulSetByManifest(result.Manifest))
			var p models.Project
			if utils.DB().Where("`name` = ? AND `namespace_id` = ?", input.Name, ns.ID).First(&p).Error == nil {
				utils.DB().Model(&models.Project{}).
					Select("Config", "GitlabProjectId", "GitlabCommit", "GitlabBranch", "DockerImage", "PodSelectors").
					Where("`id` = ?", p.ID).
					Updates(&project)
			} else {
				utils.DB().Create(&project)
			}
			pp.To(100)
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

// getPodSelectorsInDeploymentAndStatefulSetByManifest TODO: 比较 hack
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

func SendProcessPercent(conn *WsConn, slug, percent string) {
	res := &WsResponse{
		Slug:   slug,
		Type:   WsProcessPercent,
		Result: ResultSuccess,
		End:    false,
		Data:   percent,
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

type ProcessPercent struct {
	sync.Mutex
	percent int64
	slug    string
	conn    *WsConn
}

func NewProcessPercent(conn *WsConn, slug string, percent int64) *ProcessPercent {
	return &ProcessPercent{
		percent: percent,
		slug:    slug,
		conn:    conn,
	}
}

func (pp *ProcessPercent) AddOne() {
	pp.Lock()
	defer pp.Unlock()

	if pp.percent < 100 {
		pp.percent++
		SendProcessPercent(pp.conn, pp.slug, fmt.Sprintf("%d", pp.percent))
	}
}

func (pp *ProcessPercent) To(to int64) {
	pp.Lock()
	defer pp.Unlock()

	for pp.percent < to {
		time.Sleep(100 * time.Millisecond)
		pp.percent++
		SendProcessPercent(pp.conn, pp.slug, fmt.Sprintf("%d", pp.percent))
	}
}
