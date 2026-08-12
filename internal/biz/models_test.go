package biz

import (
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/stretchr/testify/assert"
)

func TestEventKey_String(t *testing.T) {
	assert.Equal(t, "audit_log", AuditLogEvent.String())
	assert.Equal(t, "namespace_created", EventNamespaceCreated.String())
	assert.Equal(t, "namespace_deleted", EventNamespaceDeleted.String())
	assert.Equal(t, "project_changed", EventProjectChanged.String())
	assert.Equal(t, "project_deleted", EventProjectDeleted.String())
}

func TestNamespace_GetImagePullSecrets(t *testing.T) {
	ns := &Namespace{ImagePullSecrets: []string{"reg1", "reg2"}}
	got := ns.GetImagePullSecrets()
	assert.Len(t, got, 2)
	assert.Equal(t, []*types.ImagePullSecret{{Name: "reg1"}, {Name: "reg2"}}, got)

	// 空列表返回空切片而非 nil，避免前端展开 nil。
	empty := (&Namespace{}).GetImagePullSecrets()
	assert.NotNil(t, empty)
	assert.Empty(t, empty)
}

func TestRepo_GetMarsConfig(t *testing.T) {
	cfg := &mars.Config{ConfigFile: "app.yml", Branches: []string{"main"}}
	r := &Repo{MarsConfig: cfg}
	assert.Same(t, cfg, r.GetMarsConfig())

	// nil 配置回退为空配置（非 nil，防下游 nil-deref）。
	fallback := (&Repo{}).GetMarsConfig()
	assert.NotNil(t, fallback)
}

func TestWrapLogFn_UnWrap(t *testing.T) {
	var (
		gotContainer []*websocket_pb.Container
		gotFormat    string
		gotArgs      []any
	)
	fn := WrapLogFn(func(c []*websocket_pb.Container, format string, v ...any) {
		gotContainer = c
		gotFormat = format
		gotArgs = v
	})
	unwrapped := fn.UnWrap()
	unwrapped("hello %s", "world")

	// UnWrap 将容器列表置 nil，只转发格式串与参数。
	assert.Nil(t, gotContainer)
	assert.Equal(t, "hello %s", gotFormat)
	assert.Equal(t, []any{"world"}, gotArgs)
}
