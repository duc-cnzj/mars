package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

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
	Slug    string `json:"slug"`
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Data    string `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	End     bool   `json:"end"`
}

func (r *WsResponse) EncodeToBytes() []byte {
	marshal, _ := json.Marshal(&r)
	return marshal
}

func (*WebsocketController) Ws(ctx *gin.Context) {
	c, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		mlog.Error(err)
		return
	}
	defer c.Close()

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
			r := &WsResponse{
				Type:    "error to Unmarshal",
				Success: false,
				Error:   err.Error(),
			}
			mlog.Error("json.Unmarshal", err.Error())
			c.WriteMessage(websocket.TextMessage, r.EncodeToBytes())
			continue
		}

		mlog.Info("handle req", wsRequest)
		switch wsRequest.Type {
		case "create_project":
			mlog.Infof("create_project type %s data %s", wsRequest.Type, wsRequest.Data)
			c.WriteMessage(1, []byte(fmt.Sprintf("create_project type %s data %s", wsRequest.Type, wsRequest.Data)))
			handleCreateProject(wsRequest.Type, wsRequest, c)
		}
	}
}

func handleCreateProject(wstype string, wsRequest WsRequest, conn *websocket.Conn) {
	var input ProjectStoreInput

	if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
		r := &WsResponse{
			Type:    wstype,
			Success: false,
			Error:   err.Error(),
		}
		mlog.Error(wsRequest.Data, &input)
		conn.WriteMessage(websocket.TextMessage, r.EncodeToBytes())
		return
	}

	var ns models.Namespace
	if err := utils.DB().Where("`id` = ?", input.NamespaceId).First(&ns).Error; err != nil {
		r := &WsResponse{
			Type:    wstype,
			Success: false,
			Error:   err.Error(),
		}
		conn.WriteMessage(websocket.TextMessage, r.EncodeToBytes())
		return
	}

	input.Name = slug.Make(input.Name)

	var project models.Project
	if utils.DB().Where("`name` = ? AND `namespace_id` = ?", input.Name, ns.ID).First(&project).Error == nil {
		utils.DB().Where("`id` = ?", project.ID).Updates(map[string]interface{}{
			"config":            input.Config,
			"gitlab_project_id": input.GitlabProjectId,
			"gitlab_commit":     input.GitlabCommit,
			"gitlab_branch":     input.GitlabBranch,
		})
	} else {
		project = models.Project{
			Name:            input.Name,
			GitlabProjectId: input.GitlabProjectId,
			GitlabBranch:    input.GitlabBranch,
			GitlabCommit:    input.GitlabCommit,
			Config:          input.Config,
			NamespaceId:     ns.ID,
		}
		utils.DB().Create(&project)
	}

	marsC, err := GetProjectMarsConfig(input.GitlabProjectId, input.GitlabBranch)
	if err != nil {
		r := &WsResponse{
			Type:    wstype,
			Success: false,
			Error:   err.Error(),
		}
		conn.WriteMessage(websocket.TextMessage, r.EncodeToBytes())
		return
	}
	file, _, err := utils.GitlabClient().RepositoryFiles.GetFile(input.GitlabProjectId, marsC.LocalChartPath, &gitlab.GetFileOptions{Ref: gitlab.String(input.GitlabBranch)})
	if err != nil {
		r := &WsResponse{
			Type:    wstype,
			Success: false,
			Error:   err.Error(),
		}
		conn.WriteMessage(websocket.TextMessage, r.EncodeToBytes())
		return
	}
	archive, _ := base64.StdEncoding.DecodeString(file.Content)

	loadArchive, err := loader.LoadArchive(bytes.NewReader(archive))
	if err != nil {
		r := &WsResponse{
			Type:    wstype,
			Success: false,
			Error:   err.Error(),
		}
		conn.WriteMessage(websocket.TextMessage, r.EncodeToBytes())
		return
	}

	filePath, deleteFn, err := marsC.GenerateConfigYamlFileByInput(input.Config)
	if err != nil {
		r := &WsResponse{
			Type:    wstype,
			Success: false,
			Error:   err.Error(),
		}
		conn.WriteMessage(websocket.TextMessage, r.EncodeToBytes())
		return
	}
	defer deleteFn()

	var valueOpts = &values.Options{
		ValueFiles: []string{filePath},
	}

	ch := make(chan MessageItem)
	fn := func(format string, v ...interface{}) {
		msg := fmt.Sprintf(format, v...)
		mlog.Debug(msg)
		ch <- MessageItem{
			Msg:  msg,
			Type: "text",
		}
	}
	go func() {
		if _, err := utils.UpgradeOrInstall(input.Name, ns.Name, loadArchive, valueOpts, fn); err != nil {
			mlog.Error(err)
			ch <- MessageItem{
				Msg:  err.Error(),
				Type: "error",
			}
			close(ch)
		} else {
			ch <- MessageItem{
				Msg:  "部署成功",
				Type: "success",
			}
			close(ch)
		}
	}()

	for s := range ch {
		r := &WsResponse{
			Type: wstype,
			Slug: slug.Make(fmt.Sprintf("%d-%s", input.NamespaceId, input.Name)),
		}
		r.Data = s.Msg
		if s.Type == "error" {
			r.Success = false
			r.End = true
		}
		if s.Type == "success" {
			r.Success = true
			r.End = true
		}
		conn.WriteMessage(websocket.TextMessage, r.EncodeToBytes())
	}
}

type MessageItem struct {
	Msg  string
	Type string
}
