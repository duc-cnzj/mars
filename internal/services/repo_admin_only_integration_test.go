package services

// 本文件是 admin_only 的跨层集成测试：不 mock repo 数据链路，
// 用 sqlite + 真实 data/biz/services + 真实审计落库，验证单元测试无法覆盖的接缝：
//  1. 管理端列表筛选、普通用户强制只见公开仓库（含传 adminOnly=true 的越权尝试）；
//  2. proto3 optional 的线上语义：JSON 不传 adminOnly = 保留库内现值，显式 false 才重置；
//  3. 导出 → protojson（EmitUnpopulated，等价 grpc-gateway）→ 导入另一库，admin_only 无损；
//  4. git.ProjectOptions 普通用户看不到管理员仓库，管理员能看到且带标记。

import (
	"context"
	"fmt"
	"testing"

	gitpb "github.com/duc-cnzj/mars/api/v6/proto/git"
	reposerver "github.com/duc-cnzj/mars/api/v6/proto/repo"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/event"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/encoding/protojson"
)

// repoStack 是一套真实的 repo 依赖链（sqlite data + biz + services），不含 repo 层 mock。
type repoStack struct {
	svc     reposerver.RepoServer
	repoBiz biz.RepoBiz
}

// newRepoStack 组装一套独立库的真实链路；事件审计走真实 data 落库（验证审计不破坏写路径）。
func newRepoStack(t *testing.T) *repoStack {
	t.Helper()
	ctrl := gomock.NewController(t)
	entdb, err := data.NewSqliteDB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = entdb.Close() })

	logger := mlog.NewForConfig(nil)
	dataImpl := data.NewDataImpl(&data.NewDataParams{DB: entdb, Logger: logger})
	// NeedGitRepo=false 的仓库不触达 git；替身仅满足构造签名。
	repoRepo := data.NewRepo(dataImpl, data.NewMockGitRepo(ctrl))
	repoBiz := biz.NewRepoBiz(data.NewHibernateRepo(dataImpl), repoRepo)
	eventBiz := biz.NewEventBiz(data.NewEventRepo(dataImpl, event.NewDispatcher(logger)))
	return &repoStack{
		svc:     NewRepoSvc(RepoSvcDeps{Logger: logger, EventBiz: eventBiz, RepoBiz: repoBiz}),
		repoBiz: repoBiz,
	}
}

func itAdminCtx() context.Context {
	return biz.SetUser(context.TODO(), &biz.UserInfo{Name: "admin", Email: "admin@x.com", Roles: []string{biz.MarsAdmin}})
}

func itUserCtx() context.Context {
	return biz.SetUser(context.TODO(), &biz.UserInfo{Name: "user", Email: "user@x.com"})
}

// TestIntegration_AdminOnly_EndToEnd 覆盖从创建、列表过滤、optional JSON 语义、
// 导出导入 round-trip 到 ProjectOptions 的全链路。
func TestIntegration_AdminOnly_EndToEnd(t *testing.T) {
	stack := newRepoStack(t)
	admin, user := itAdminCtx(), itUserCtx()

	secret, err := stack.svc.Create(admin, &reposerver.CreateRequest{Name: "secret", AdminOnly: true})
	require.NoError(t, err)
	require.True(t, secret.Item.AdminOnly)

	public, err := stack.svc.Create(admin, &reposerver.CreateRequest{Name: "public"})
	require.NoError(t, err)
	require.False(t, public.Item.AdminOnly)

	// 1) 管理端筛选
	tru := true
	adminList, err := stack.svc.List(admin, &reposerver.ListRequest{AdminOnly: &tru})
	require.NoError(t, err)
	require.Equal(t, int32(1), adminList.Count)
	require.Equal(t, "secret", adminList.Items[0].Name)
	require.True(t, adminList.Items[0].AdminOnly)

	// 2) 普通用户：默认列表 + 显式传 adminOnly=true 的越权尝试，都必须只回公开仓库
	userList, err := stack.svc.List(user, &reposerver.ListRequest{})
	require.NoError(t, err)
	require.Equal(t, int32(1), userList.Count)
	require.Equal(t, "public", userList.Items[0].Name)

	userProbe, err := stack.svc.List(user, &reposerver.ListRequest{AdminOnly: &tru})
	require.NoError(t, err)
	require.Equal(t, int32(1), userProbe.Count, "普通用户传 adminOnly=true 不得反查出管理员仓库")
	require.Equal(t, "public", userProbe.Items[0].Name)

	// 3) optional 线上语义：JSON 不传 adminOnly → 保留现值（protojson 解码等价 grpc-gateway）
	var omitReq reposerver.UpdateRequest
	require.NoError(t, protojson.Unmarshal(
		[]byte(fmt.Sprintf(`{"id":%d,"name":"secret"}`, secret.Item.Id)), &omitReq))
	require.Nil(t, omitReq.AdminOnly, "JSON 不传 adminOnly 时字段必须缺省")
	updated, err := stack.svc.Update(admin, &omitReq)
	require.NoError(t, err)
	require.True(t, updated.Item.AdminOnly, "不传 adminOnly 的更新必须保留库内管理员可见状态")

	// 显式传 false → 重置为公开
	var explicitFalse reposerver.UpdateRequest
	require.NoError(t, protojson.Unmarshal(
		[]byte(fmt.Sprintf(`{"id":%d,"name":"secret","adminOnly":false}`, secret.Item.Id)), &explicitFalse))
	require.NotNil(t, explicitFalse.AdminOnly)
	updated, err = stack.svc.Update(admin, &explicitFalse)
	require.NoError(t, err)
	require.False(t, updated.Item.AdminOnly)

	// 恢复为管理员可见，供导出导入与下拉断言
	var explicitTrue reposerver.UpdateRequest
	require.NoError(t, protojson.Unmarshal(
		[]byte(fmt.Sprintf(`{"id":%d,"name":"secret","adminOnly":true}`, secret.Item.Id)), &explicitTrue))
	_, err = stack.svc.Update(admin, &explicitTrue)
	require.NoError(t, err)

	// 4) 导出 → protojson(EmitUnpopulated，等价 gateway 输出) → 导入另一套真实库
	exported, err := stack.svc.Export(admin, &reposerver.ExportRequest{})
	require.NoError(t, err)
	raw, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(exported)
	require.NoError(t, err)
	var decoded reposerver.ExportResponse
	require.NoError(t, protojson.Unmarshal(raw, &decoded))
	require.Len(t, decoded.Items, 2)
	var exportedSecret *types.RepoModel
	for _, item := range decoded.Items {
		if item.Name == "secret" {
			exportedSecret = item
		}
	}
	require.NotNil(t, exportedSecret)
	require.True(t, exportedSecret.AdminOnly, "导出 JSON 必须保留 admin_only=true")

	other := newRepoStack(t)
	_, err = other.svc.Import(admin, &reposerver.ImportRequest{Items: decoded.Items})
	require.NoError(t, err)
	imported, err := other.svc.List(admin, &reposerver.ListRequest{AdminOnly: &tru})
	require.NoError(t, err)
	require.Equal(t, int32(1), imported.Count, "导入后管理员仓库仍应是管理员可见")
	require.Equal(t, "secret", imported.Items[0].Name)

	// 5) git.ProjectOptions：普通用户不可见，管理员可见且带标记
	gitSvc := NewGitSvc(GitSvcDeps{Logger: mlog.NewForConfig(nil), RepoBiz: stack.repoBiz})
	adminOpts, err := gitSvc.ProjectOptions(admin, &gitpb.ProjectOptionsRequest{})
	require.NoError(t, err)
	require.Len(t, adminOpts.Items, 2)
	var secretOpt *gitpb.Option
	for _, o := range adminOpts.Items {
		if o.Label == "secret" {
			secretOpt = o
		}
	}
	require.NotNil(t, secretOpt)
	require.True(t, secretOpt.AdminOnly)

	userOpts, err := gitSvc.ProjectOptions(user, &gitpb.ProjectOptionsRequest{})
	require.NoError(t, err)
	require.Len(t, userOpts.Items, 1, "普通用户的仓库下拉只能有公开仓库")
	require.Equal(t, "public", userOpts.Items[0].Label)
}
