package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/types"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	reposerver "github.com/duc-cnzj/mars/api/v6/proto/repo"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
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
		gomock.Any(),
		gomock.Not(nil),
	)
	repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{}, nil)
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
		AccessBiz: biz.NewAccessBiz(logger, nil, nil),
	}).(*repoSvc)
	if !ok {
		panic("NewRepoSvc returned unexpected type")
	}
	return s, mocks
}
