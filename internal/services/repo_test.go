package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	reposerver "github.com/duc-cnzj/mars/api/v6/proto/repo"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewRepoSvc(t *testing.T) {
	svc, _ := newRepoSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.eventBiz)
	assert.NotNil(t, svc.repoBiz)
	assert.NotNil(t, svc.logger)
}

func Test_repoSvc_Clone_Success(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Create,
		biz.MustGetUser(newAdminUserCtx()).Name,
		biz.MustGetUser(newAdminUserCtx()).Email,
		gomock.Any(),
		gomock.Not(nil),
	)
	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{}, nil)
	// biz.Clone 前置名称唯一性校验：GetByName NotFound 视为名称空闲。
	repoRepo.EXPECT().GetByName(gomock.Any(), "clone").Return(nil, repoNotFound())
	repoRepo.EXPECT().Clone(gomock.Any(), &biz.CloneRepoInput{
		ID:   1,
		Name: "clone",
	}).Return(&biz.Repo{
		ID:   2,
		Name: "clone",
	}, nil)

	res, err := svc.Clone(newAdminUserCtx(), &reposerver.CloneRequest{
		Id:   1,
		Name: "clone",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(2), res.Item.Id)
	assert.Equal(t, "clone", res.Item.Name)
}

func Test_repoSvc_Clone_Error(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	// 修复后 Get 先于 Clone；Get 成功后才可能走到 Clone 失败分支
	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{}, nil)
	repoRepo.EXPECT().GetByName(gomock.Any(), "clone").Return(nil, repoNotFound())
	repoRepo.EXPECT().Clone(gomock.Any(), &biz.CloneRepoInput{
		ID:   1,
		Name: "clone",
	}).Return(nil, errors.New("error"))

	res, err := svc.Clone(newAdminUserCtx(), &reposerver.CloneRequest{
		Id:   1,
		Name: "clone",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func Test_repoSvc_Clone_GetError(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	// 回归防护：Get（查询源仓库）失败必须 fail-fast，Clone 副作用不得发生。
	// 否则会出现"克隆已成功但 Get 失败返回错误"→ 客户端重试产生重复克隆。
	// 改坏实现（先 Clone 后 Get / Get 失败仍继续）时此测试 FAIL（未期望 Clone 调用）。
	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(nil, errors.New("x"))
	// 不设置 Clone 期望：若实现先执行克隆，gomock 会以"未期望调用"中止

	res, err := svc.Clone(newAdminUserCtx(), &reposerver.CloneRequest{
		Id:   1,
		Name: "clone",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Create_Success(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo
	eventRepo := mocks.eventRepo

	// biz.Create 前置名称唯一性校验：GetByName NotFound 视为名称空闲。
	repoRepo.EXPECT().GetByName(gomock.Any(), "newRepo").Return(nil, repoNotFound())
	repoRepo.EXPECT().Create(gomock.Any(), &biz.CreateRepoInput{
		Name:         "newRepo",
		Enabled:      true,
		NeedGitRepo:  true,
		GitProjectID: lo.ToPtr(int32(1)),
		MarsConfig:   &mars.Config{},
		Description:  "description",
	}).Return(&biz.Repo{
		ID:   1,
		Name: "newRepo",
	}, nil)

	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Create,
		biz.MustGetUser(newAdminUserCtx()).Name,
		biz.MustGetUser(newAdminUserCtx()).Email,
		gomock.Any(),
		gomock.Not(nil),
	)

	res, err := svc.Create(newAdminUserCtx(), &reposerver.CreateRequest{
		Name:         "newRepo",
		NeedGitRepo:  true,
		GitProjectId: lo.ToPtr(int32(1)),
		MarsConfig:   &mars.Config{},
		Description:  "description",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(1), res.Item.Id)
	assert.Equal(t, "newRepo", res.Item.Name)
}

func TestRepoSvc_Create_Error(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().GetByName(gomock.Any(), "newRepo").Return(nil, repoNotFound())
	repoRepo.EXPECT().Create(gomock.Any(), &biz.CreateRepoInput{
		Name:         "newRepo",
		Enabled:      true,
		NeedGitRepo:  true,
		GitProjectID: lo.ToPtr(int32(1)),
		MarsConfig:   &mars.Config{},
		Description:  "description",
	}).Return(nil, errors.New("error"))

	res, err := svc.Create(newAdminUserCtx(), &reposerver.CreateRequest{
		Name:         "newRepo",
		NeedGitRepo:  true,
		GitProjectId: lo.ToPtr(int32(1)),
		MarsConfig:   &mars.Config{},
		Description:  "description",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Delete_Success(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo
	eventRepo := mocks.eventRepo

	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Repo{ID: 1}, nil)
	repoRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)
	req := &reposerver.DeleteRequest{
		Id: 1,
	}
	eventRepo.EXPECT().AuditLogWithRequest(types.EventActionType_Delete,
		biz.MustGetUser(newAdminUserCtx()).Name,
		biz.MustGetUser(newAdminUserCtx()).Email,
		gomock.Any(),
		req,
	)

	res, err := svc.Delete(newAdminUserCtx(), req)

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestRepoSvc_Delete_Error(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Repo{ID: 1}, nil)
	repoRepo.EXPECT().Delete(gomock.Any(), 1).Return(errors.New("error"))

	res, err := svc.Delete(newAdminUserCtx(), &reposerver.DeleteRequest{
		Id: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_List_Success(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().List(gomock.Any(), &biz.ListRepoRequest{
		Page:          1,
		PageSize:      10,
		Enabled:       lo.ToPtr(true),
		OrderByIDDesc: lo.ToPtr(true),
		Name:          "test",
	}).Return([]*biz.Repo{
		{
			ID:   1,
			Name: "test",
		},
	}, &pagination.Pagination{
		Page:     1,
		PageSize: 10,
		Count:    1,
	}, nil)

	res, err := svc.List(context.TODO(), &reposerver.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(10)),
		Enabled:  lo.ToPtr(true),
		Name:     "test",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(1), res.Page)
	assert.Equal(t, int32(10), res.PageSize)
	assert.Equal(t, int32(1), res.Count)
	assert.Equal(t, int32(1), res.Items[0].Id)
	assert.Equal(t, "test", res.Items[0].Name)
}

func TestRepoSvc_List_Error(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().List(gomock.Any(), &biz.ListRepoRequest{
		Page:          1,
		PageSize:      10,
		Enabled:       lo.ToPtr(true),
		OrderByIDDesc: lo.ToPtr(true),
		Name:          "test",
	}).Return(nil, nil, errors.New("error"))

	res, err := svc.List(context.TODO(), &reposerver.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(10)),
		Enabled:  lo.ToPtr(true),
		Name:     "test",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Show_Success(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Repo{
		ID:   1,
		Name: "show",
	}, nil)

	res, err := svc.Show(context.TODO(), &reposerver.ShowRequest{
		Id: 1,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(1), res.Item.Id)
	assert.Equal(t, "show", res.Item.Name)
}

func TestRepoSvc_Show_Error(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))

	res, err := svc.Show(context.TODO(), &reposerver.ShowRequest{
		Id: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_ToggleEnabled_Success(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo
	eventRepo := mocks.eventRepo

	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{ID: 1, Enabled: false}, nil)
	repoRepo.EXPECT().ToggleEnabled(gomock.Any(), 1, true).Return(&biz.Repo{
		ID:      1,
		Name:    "toggle",
		Enabled: true,
	}, nil)

	req := &reposerver.ToggleEnabledRequest{
		Id:      1,
		Enabled: true,
	}
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Update,
		biz.MustGetUser(newAdminUserCtx()).Name,
		biz.MustGetUser(newAdminUserCtx()).Email,
		gomock.Any(),
		req,
	)
	res, err := svc.ToggleEnabled(newAdminUserCtx(), req)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(1), res.Item.Id)
	assert.Equal(t, "toggle", res.Item.Name)
	assert.Equal(t, true, res.Item.Enabled)
}

func TestRepoSvc_ToggleEnabled_Error(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{ID: 1, Enabled: false}, nil)
	repoRepo.EXPECT().ToggleEnabled(gomock.Any(), 1, true).Return(nil, errors.New("error"))

	res, err := svc.ToggleEnabled(newAdminUserCtx(), &reposerver.ToggleEnabledRequest{
		Id:      1,
		Enabled: true,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Update_Success(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo
	eventRepo := mocks.eventRepo

	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{
		ID:   1,
		Name: "update",
	}, nil)

	// biz.Update 前置：GetByName NotFound（名称空闲）+ Show 校验有项目不可改名。
	repoRepo.EXPECT().GetByName(gomock.Any(), "updated").Return(nil, repoNotFound())
	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Repo{ID: 1, Name: "update"}, nil)

	repoRepo.EXPECT().Update(gomock.Any(), &biz.UpdateRepoInput{
		ID:           1,
		Name:         "updated",
		NeedGitRepo:  true,
		GitProjectID: lo.ToPtr(int32(1)),
		MarsConfig:   &mars.Config{},
		Description:  "updated description",
	}).Return(&biz.Repo{
		ID:   1,
		Name: "updated",
	}, nil)

	eventRepo.EXPECT().AuditLogWithChange(types.EventActionType_Update,
		biz.MustGetUser(newAdminUserCtx()).Name,
		biz.MustGetUser(newAdminUserCtx()).Email,
		gomock.Any(),
		gomock.Not(nil),
		gomock.Not(nil),
	)

	res, err := svc.Update(newAdminUserCtx(), &reposerver.UpdateRequest{
		Id:           1,
		Name:         "updated",
		NeedGitRepo:  true,
		GitProjectId: lo.ToPtr(int32(1)),
		MarsConfig:   &mars.Config{},
		Description:  "updated description",
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(1), res.Item.Id)
	assert.Equal(t, "updated", res.Item.Name)
}

func TestRepoSvc_Update_Error(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(nil, errors.New("error"))

	res, err := svc.Update(newAdminUserCtx(), &reposerver.UpdateRequest{
		Id:           1,
		Name:         "updated",
		NeedGitRepo:  true,
		GitProjectId: lo.ToPtr(int32(1)),
		MarsConfig:   &mars.Config{},
		Description:  "updated description",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Update_Error2(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{}, nil)
	// biz.Update 前置：GetByName NotFound + Show（无项目）后走到 Update 失败。
	repoRepo.EXPECT().GetByName(gomock.Any(), "updated").Return(nil, repoNotFound())
	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Repo{}, nil)
	repoRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
	res, err := svc.Update(newAdminUserCtx(), &reposerver.UpdateRequest{
		Id:           1,
		Name:         "updated",
		NeedGitRepo:  true,
		GitProjectId: lo.ToPtr(int32(1)),
		MarsConfig:   &mars.Config{},
		Description:  "updated description",
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Authorize_AdminUser(t *testing.T) {
	svc, _ := newRepoSvcWithMocks(t)

	ctx := newAdminUserCtx()
	_, err := svc.Authorize(ctx, "List")

	assert.Nil(t, err)
}

func TestRepoSvc_Authorize_AdminUser2(t *testing.T) {
	svc, _ := newRepoSvcWithMocks(t)

	ctx := newAdminUserCtx()
	_, err := svc.Authorize(ctx, "XX")

	assert.Nil(t, err)
}

func TestRepoSvc_Authorize_ListMethod(t *testing.T) {
	svc, _ := newRepoSvcWithMocks(t)

	ctx := newOtherUserCtx()
	_, err := svc.Authorize(ctx, "/repo.Repo/List")
	assert.Nil(t, err)
	_, err = svc.Authorize(ctx, "/repo.Repo/Show")
	assert.Nil(t, err)
}

func TestRepoSvc_Authorize_NonListMethod(t *testing.T) {
	svc, _ := newRepoSvcWithMocks(t)

	ctx := newOtherUserCtx()
	_, err := svc.Authorize(ctx, "NonList")

	assert.NotNil(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestRepoSvc_Export_Success(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{}).Return([]*biz.Repo{
		{ID: 1, Name: "a"},
		{ID: 2, Name: "b"},
	}, nil)

	res, err := svc.Export(context.TODO(), &reposerver.ExportRequest{})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	require.Len(t, res.Items, 2)
	assert.Equal(t, "a", res.Items[0].Name)
	assert.Equal(t, "b", res.Items[1].Name)
}

func TestRepoSvc_Export_Error(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{}).Return(nil, errors.New("boom"))

	res, err := svc.Export(context.TODO(), &reposerver.ExportRequest{})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Import_Success(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo
	eventRepo := mocks.eventRepo

	// 导入前/后各采集一次全量快照：before 用于审计 old 字段，after 用于审计 new 字段。
	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{}).Return([]*biz.Repo{
		{ID: 1, Name: "existing", Enabled: true, Description: "d1"},
	}, nil).Times(2)

	// 幂等覆盖在 data 层事务内完成，services 只委托一次。
	repoRepo.EXPECT().Import(gomock.Any(), []*biz.ImportRepoItem{
		{Name: "existing", Enabled: true, Description: "d1"},
		{Name: "new", Enabled: true, Description: "d2"},
	}).Return(1, 1, nil)

	req := &reposerver.ImportRequest{Items: []*types.RepoModel{
		{Name: "existing", Enabled: true, Description: "d1"},
		{Name: "new", Enabled: true, Description: "d2"},
	}}
	eventRepo.EXPECT().AuditLogWithChange(
		types.EventActionType_Update,
		biz.MustGetUser(newAdminUserCtx()).Name,
		biz.MustGetUser(newAdminUserCtx()).Email,
		"导入 repo: 共 2 个（新建 1，覆盖 1）",
		gomock.Not(nil),
		gomock.Not(nil),
	)

	res, err := svc.Import(newAdminUserCtx(), req)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(2), res.Total)
	assert.Equal(t, int32(1), res.Created)
	assert.Equal(t, int32(1), res.Updated)
}

func TestRepoSvc_Import_Empty(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo
	eventRepo := mocks.eventRepo

	// 空导入仍走快照流程（无数据变更），审计记录计数为 0。
	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{}).Return(nil, nil).Times(2)

	req := &reposerver.ImportRequest{}
	eventRepo.EXPECT().AuditLogWithChange(
		types.EventActionType_Update,
		biz.MustGetUser(newAdminUserCtx()).Name,
		biz.MustGetUser(newAdminUserCtx()).Email,
		"导入 repo: 共 0 个（新建 0，覆盖 0）",
		gomock.Not(nil),
		gomock.Not(nil),
	)

	res, err := svc.Import(newAdminUserCtx(), req)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(0), res.Total)
	assert.Equal(t, int32(0), res.Created)
	assert.Equal(t, int32(0), res.Updated)
}

func TestRepoSvc_Import_Error(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	// 导入前快照成功后导入失败：数据未变（事务回滚），返回错误、无审计。
	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{}).Return(nil, nil)
	repoRepo.EXPECT().Import(gomock.Any(), []*biz.ImportRepoItem{
		{Name: "new"},
	}).Return(0, 0, errors.New("db down"))

	res, err := svc.Import(newAdminUserCtx(), &reposerver.ImportRequest{Items: []*types.RepoModel{
		{Name: "new"},
	}})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Import_DryRun(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	// dry_run：委托 PreviewImport 干跑计数 + 采集全量旧值，按 name 匹配出将被覆盖条目。
	repoRepo.EXPECT().PreviewImport(gomock.Any(), []*biz.ImportRepoItem{
		{Name: "existing", Enabled: true},
		{Name: "new", Enabled: true},
	}).Return(1, 1, nil)
	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{}).Return([]*biz.Repo{
		{ID: 1, Name: "existing", Enabled: true},
	}, nil)

	res, err := svc.Import(newAdminUserCtx(), &reposerver.ImportRequest{
		Items: []*types.RepoModel{
			{Name: "existing", Enabled: true},
			{Name: "new", Enabled: true},
		},
		DryRun: true,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(2), res.Total)
	assert.Equal(t, int32(1), res.Created)
	assert.Equal(t, int32(1), res.Updated)
	// 仅「existing」在库中，updated_old 只含它的旧值；「new」是新建不出现在 updated_old。
	require.Len(t, res.UpdatedOld, 1)
	assert.Equal(t, "existing", res.UpdatedOld[0].Name)
	assert.Equal(t, true, res.UpdatedOld[0].Enabled)
}

func TestRepoSvc_Import_DryRun_Error(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	// 至少一条 item 才会委托到 repo（biz.PreviewImport 对空切片短路）。
	repoRepo.EXPECT().PreviewImport(gomock.Any(), gomock.Any()).Return(0, 0, errors.New("boom"))

	res, err := svc.Import(newAdminUserCtx(), &reposerver.ImportRequest{
		Items:  []*types.RepoModel{{Name: "new"}},
		DryRun: true,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Import_DryRun_SnapshotError(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	// dry_run 预览计数成功后采集旧值快照失败：整体返回错误（拿不到 diff 依据就不放行预览）。
	repoRepo.EXPECT().PreviewImport(gomock.Any(), gomock.Any()).Return(0, 1, nil)
	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{}).Return(nil, errors.New("boom"))

	res, err := svc.Import(newAdminUserCtx(), &reposerver.ImportRequest{
		Items:  []*types.RepoModel{{Name: "existing"}},
		DryRun: true,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Import_SnapshotBeforeError(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	// 导入前快照采集失败：fail-fast，不进入导入、无审计。
	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{}).Return(nil, errors.New("boom"))

	res, err := svc.Import(newAdminUserCtx(), &reposerver.ImportRequest{Items: []*types.RepoModel{{Name: "new"}}})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Import_SnapshotAfterError(t *testing.T) {
	svc, mocks := newRepoSvcWithMocks(t)
	repoRepo := mocks.repoRepo

	// 导入成功后再采集新快照失败：数据已变更但无法产出审计，返回错误。
	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{}).Return(nil, nil)                    // before
	repoRepo.EXPECT().Import(gomock.Any(), []*biz.ImportRepoItem{{Name: "new"}}).Return(1, 0, nil) // import
	repoRepo.EXPECT().All(gomock.Any(), &biz.AllRepoRequest{}).Return(nil, errors.New("boom"))     // after

	res, err := svc.Import(newAdminUserCtx(), &reposerver.ImportRequest{Items: []*types.RepoModel{{Name: "new"}}})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func Test_ExportImport_RoundTrip(t *testing.T) {
	// round-trip 闭环：biz.Repo → FromRepo（导出）→ toImportRepoItem（导入抽取）
	// 必须保住可落库字段（name/enabled/need_git_repo/git_project_id/mars_config/description），
	// 服务器生成字段（id/时间戳/git 项目名）有意丢弃由 data 层重新推导。
	repo := &biz.Repo{
		ID:             42,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Name:           "app",
		DefaultBranch:  "main",
		GitProjectName: "proj",
		GitProjectID:   100,
		Enabled:        true,
		NeedGitRepo:    true,
		MarsConfig:     &mars.Config{ConfigField: "c", LocalChartPath: "100|main|chart"},
		Description:    "desc",
	}

	item := toImportRepoItem(transformer.FromRepo(repo))
	require.NotNil(t, item)
	assert.Equal(t, "app", item.Name)
	assert.True(t, item.Enabled)
	assert.True(t, item.NeedGitRepo)
	assert.Equal(t, lo.ToPtr(int32(100)), item.GitProjectID)
	assert.Equal(t, &mars.Config{ConfigField: "c", LocalChartPath: "100|main|chart"}, item.MarsConfig)
	assert.Equal(t, "desc", item.Description)

	// 非 git 仓库导出后回导：不携带 git id（git 字段由 data 层按 NeedGitRepo 清空）。
	noGit := toImportRepoItem(transformer.FromRepo(&biz.Repo{Name: "plain", Enabled: false, NeedGitRepo: false}))
	require.NotNil(t, noGit)
	assert.Equal(t, "plain", noGit.Name)
	assert.False(t, noGit.Enabled)
	assert.False(t, noGit.NeedGitRepo)
	assert.Nil(t, noGit.GitProjectID)
}

func Test_toImportRepoItem(t *testing.T) {
	assert.Nil(t, toImportRepoItem(nil))

	item := toImportRepoItem(&types.RepoModel{
		Name:         "app",
		Enabled:      true,
		NeedGitRepo:  true,
		GitProjectId: 100,
		MarsConfig:   &mars.Config{ConfigField: "c"},
		Description:  "desc",
	})
	require.NotNil(t, item)
	assert.Equal(t, "app", item.Name)
	assert.True(t, item.Enabled)
	assert.True(t, item.NeedGitRepo)
	assert.Equal(t, lo.ToPtr(int32(100)), item.GitProjectID)
	assert.Equal(t, &mars.Config{ConfigField: "c"}, item.MarsConfig)
	assert.Equal(t, "desc", item.Description)

	// 需要 git 仓库但 id 无效（0）→ 不携带 GitProjectID，避免落库无效 0 值。
	noID := toImportRepoItem(&types.RepoModel{NeedGitRepo: true})
	require.NotNil(t, noID)
	assert.Nil(t, noID.GitProjectID)

	// 不需要 git 仓库 → 即使带了 id 也不携带。
	noGit := toImportRepoItem(&types.RepoModel{NeedGitRepo: false, GitProjectId: 100})
	require.NotNil(t, noGit)
	assert.Nil(t, noGit.GitProjectID)
}

func TestRepoSvc_Authorize_ExportImport_Admin(t *testing.T) {
	svc, _ := newRepoSvcWithMocks(t)

	_, err := svc.Authorize(newAdminUserCtx(), reposerver.Repo_Export_FullMethodName)
	assert.Nil(t, err)
	_, err = svc.Authorize(newAdminUserCtx(), reposerver.Repo_Import_FullMethodName)
	assert.Nil(t, err)
}

func TestRepoSvc_Authorize_Export_NonAdmin(t *testing.T) {
	svc, _ := newRepoSvcWithMocks(t)

	_, err := svc.Authorize(newOtherUserCtx(), reposerver.Repo_Export_FullMethodName)

	assert.NotNil(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestRepoSvc_Authorize_Import_NonAdmin(t *testing.T) {
	svc, _ := newRepoSvcWithMocks(t)

	_, err := svc.Authorize(newOtherUserCtx(), reposerver.Repo_Import_FullMethodName)

	assert.NotNil(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

// Test_repoAuditYaml 锁定审计 YAML 字段白名单：只保留业务字段
// （name/enabled/need_git_repo/git_project_id/description/mars_config），
// 剔除服务器生成的 id/时间戳/git 项目名——否则导入/更新审计 diff 会被
// updated_at 刷新等业务无关噪声刷屏（用户明确不关注 created_at/updated_at）。
func Test_repoAuditYaml(t *testing.T) {
	out := repoAuditYaml([]*types.RepoModel{
		{
			Id:             1,
			CreatedAt:      "2020-01-02T03:04:05Z",
			UpdatedAt:      "2026-08-25T00:00:00Z",
			Name:           "name_4",
			Enabled:        true,
			NeedGitRepo:    false,
			GitProjectId:   0,
			GitProjectName: "git-name",
			Description:    "desc",
			MarsConfig:     &mars.Config{ConfigField: "c"},
		},
	})

	// 业务字段全保留（断言 mars_config 嵌套内容可序列化）。
	assert.Contains(t, out, "name: name_4")
	assert.Contains(t, out, "enabled: true")
	assert.Contains(t, out, "need_git_repo: false")
	assert.Contains(t, out, "git_project_id: 0")
	assert.Contains(t, out, "description: desc")
	assert.Contains(t, out, "config_field: c")

	// 服务器生成字段全部剔除。
	assert.NotContains(t, out, "created_at:")
	assert.NotContains(t, out, "updated_at:")
	assert.NotContains(t, out, "git_project_name:")
	// 用 "\n  id:" 而非 "id:"——避免误伤合法的 git_project_id: 子串。
	assert.NotContains(t, out, "\n  id:")

	// nil/空输入不 panic，输出空列表。
	assert.Equal(t, "[]\n", repoAuditYaml(nil))
}

// TestRepoGatewayRoute_ExportNotShadowedByShow 路由回归：GET /api/repos/export 必须命中
// Export 而非被 Show 的 /api/repos/{id} 通配吞掉。grpc-gateway 按注册逆序试匹配，
// Export 排在 Show 之后（proto service 块顺序），晚注册先命中；此测试锁定该顺序。
func TestRepoGatewayRoute_ExportNotShadowedByShow(t *testing.T) {
	mux := runtime.NewServeMux()
	server := &routeRecorderRepoServer{}
	require.NoError(t, reposerver.RegisterRepoHandlerServer(context.TODO(), mux, server))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/repos/export", nil))

	// Export 命中 → 200 + 响应含 exported；若被 Show 吞掉，id="export" 解析失败返回 400。
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "exported")
	assert.Equal(t, 1, server.exportCalls)
	assert.Zero(t, server.showCalls)
}

// routeRecorderRepoServer 记录 Export/Show 被调用次数，用于锁定 gateway 路由命中。
type routeRecorderRepoServer struct {
	reposerver.UnimplementedRepoServer
	exportCalls int
	showCalls   int
}

func (s *routeRecorderRepoServer) Export(_ context.Context, _ *reposerver.ExportRequest) (*reposerver.ExportResponse, error) {
	s.exportCalls++
	return &reposerver.ExportResponse{Items: []*types.RepoModel{{Name: "exported"}}}, nil
}

func (s *routeRecorderRepoServer) Show(_ context.Context, _ *reposerver.ShowRequest) (*reposerver.ShowResponse, error) {
	s.showCalls++
	return &reposerver.ShowResponse{Item: &types.RepoModel{Name: "show"}}, nil
}

// repoNotFound 返回 biz 层 errs.IsNotFound 判定的 NotFound 错误，
// 供 repoBiz 前置 GetByName 预查"名称空闲"分支使用。
func repoNotFound() error {
	return errs.WrapNotFound(errors.New("not found"), "not found")
}

type repoSvcMocks struct {
	ctrl      *gomock.Controller
	eventRepo *data.MockEventRepo
	repoRepo  *data.MockRepoRepo
}

func newRepoSvcWithMocks(t *testing.T) (*repoSvc, *repoSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &repoSvcMocks{
		ctrl:      ctrl,
		eventRepo: data.NewMockEventRepo(ctrl),
		repoRepo:  data.NewMockRepoRepo(ctrl),
	}
	logger := mlog.NewForConfig(nil)
	s, ok := NewRepoSvc(RepoSvcDeps{
		Logger:    logger,
		EventBiz:  biz.NewEventBiz(mocks.eventRepo),
		RepoBiz:   biz.NewRepoBiz(mocks.repoRepo),
		AccessBiz: biz.NewAccessBiz(nil, nil),
	}).(*repoSvc)
	if !ok {
		panic("NewRepoSvc returned unexpected type")
	}
	return s, mocks
}
