package services

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// stubServiceRegistrar 记录所有被注册的服务，让测试无需启动真实服务器
// 即可断言 grpc 装配关系。
type stubServiceRegistrar struct {
	services []string
}

func (s *stubServiceRegistrar) RegisterService(sd *grpc.ServiceDesc, _ any) {
	s.services = append(s.services, sd.ServiceName)
}

func Test_NewGrpcRegistry_Success(t *testing.T) {
	reg := NewGrpcRegistry(NewGrpcRegistryDeps{})

	assert.IsType(t, &application.GrpcRegistry{}, reg)
	assert.Len(t, reg.EndpointFuncs, 15)
	assert.NotNil(t, reg.RegistryFunc)
}

func Test_NewGrpcRegistry_RegistersAllServices(t *testing.T) {
	reg := NewGrpcRegistry(NewGrpcRegistryDeps{})

	registrar := &stubServiceRegistrar{}
	reg.RegistryFunc(registrar)

	want := []string{
		"repo.Repo",
		"container.Container",
		"cluster.Cluster",
		"endpoint.Endpoint",
		"event.Event",
		"file.File",
		"git.Git",
		"metrics.Metrics",
		"namespace.Namespace",
		"picture.Picture",
		"project.Project",
		"version.Version",
		"changelog.Changelog",
		"auth.Auth",
		"token.AccessToken",
	}
	assert.Equal(t, want, registrar.services)
}

// Test_AccessGetUserBinding 验证 AccessBiz 的用户提取契约：用户由内部
// MustGetUser 从 ctx 提取——admin 上下文过 admin 门禁、非 admin 拒绝。
// repo 传 nil 即可（RequireAdmin 不触达实体加载，符合 NewAccessBiz 的"repo 懒加载"约定）。
func Test_AccessGetUserBinding(t *testing.T) {
	ab := biz.NewAccessBiz(mlog.NewForConfig(nil), nil, nil)
	// 编译期断言返回类型满足 AccessBiz 接口。
	var _ biz.AccessBiz = ab

	// admin 上下文通过 admin 门禁（allowlist 未命中时仍放行）。
	ctx, err := ab.RequireAdmin(newAdminUserCtx(), "/api/v1/file/List")
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	// 非 admin 上下文且未命中 allowlist 时拒绝。
	_, err = ab.RequireAdmin(newOtherUserCtx(), "/api/v1/file/List")
	assert.Equal(t, biz.ErrorPermissionDenied, err)
}
