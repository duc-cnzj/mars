package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGitConfigSvc_Authorize(t *testing.T) {
	e := new(GitConfigSvc)
	ctx := context.TODO()
	ctx = auth.SetUser(ctx, &contracts.UserInfo{})
	_, err := e.Authorize(ctx, "")
	assert.ErrorIs(t, err, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error()))
	ctx = auth.SetUser(ctx, &contracts.UserInfo{
		Roles: []string{"admin"},
	})
	_, err = e.Authorize(ctx, "")
	assert.Nil(t, err)
}

func TestGitConfigSvc_GetDefaultChartValues(t *testing.T) {
}

func TestGitConfigSvc_GlobalConfig(t *testing.T) {
}

func TestGitConfigSvc_Show(t *testing.T) {
}

func TestGitConfigSvc_ToggleGlobalStatus(t *testing.T) {
}

func TestGitConfigSvc_Update(t *testing.T) {
}

func Test_getDefaultBranch(t *testing.T) {
}
