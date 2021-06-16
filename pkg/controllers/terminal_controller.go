package controllers

import (
	"net/http"

	"github.com/DuC-cnZj/mars/pkg/controllers/terminal"
	"github.com/gin-gonic/gin"
)

type TerminalController struct{}

func NewTerminalController() *TerminalController {
	return &TerminalController{}
}

type HandleExecShellUri struct {
	Namespace string `uri:"namespace"`
	Pod       string `uri:"pod"`

	Container string `form:"container"`
}

func (*TerminalController) HandleExecShell(ctx *gin.Context) {
	terminal.HandleExecShell(ctx)
}

func (*TerminalController) HandleSocket(path string) http.Handler {
	return terminal.CreateAttachHandler(path)
}
