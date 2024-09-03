package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v5/mars"
	reposerver "github.com/duc-cnzj/mars/api/v5/repo"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/util/pagination"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewRepoSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewRepoSvc(mlog.NewForConfig(nil), repo.NewMockEventRepo(m), repo.NewMockGitRepo(m), repo.NewMockRepoRepo(m))
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*repoSvc).eventRepo)
	assert.NotNil(t, svc.(*repoSvc).gitRepo)
	assert.NotNil(t, svc.(*repoSvc).repoRepo)
	assert.NotNil(t, svc.(*repoSvc).logger)
}

func Test_repoSvc_Clone_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		eventRepo,
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	repoRepo.EXPECT().Show(gomock.Any(), int(1)).Return(&repo.Repo{}, nil)
	repoRepo.EXPECT().Clone(gomock.Any(), &repo.CloneRepoInput{
		ID:   1,
		Name: "clone",
	}).Return(&repo.Repo{
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
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().Clone(gomock.Any(), &repo.CloneRepoInput{
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

func TestRepoSvc_Create_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		eventRepo,
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().Create(gomock.Any(), &repo.CreateRepoInput{
		Name:         "newRepo",
		Enabled:      true,
		NeedGitRepo:  true,
		GitProjectID: lo.ToPtr(int32(1)),
		MarsConfig:   &mars.Config{},
		Description:  "description",
	}).Return(&repo.Repo{
		ID:   1,
		Name: "newRepo",
	}, nil)

	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

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
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().Create(gomock.Any(), &repo.CreateRepoInput{
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
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		eventRepo,
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)
	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

	res, err := svc.Delete(newAdminUserCtx(), &reposerver.DeleteRequest{
		Id: 1,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestRepoSvc_Delete_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().Delete(gomock.Any(), 1).Return(errors.New("error"))

	res, err := svc.Delete(newAdminUserCtx(), &reposerver.DeleteRequest{
		Id: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_List_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().List(gomock.Any(), &repo.ListRepoRequest{
		Page:          1,
		PageSize:      10,
		Enabled:       lo.ToPtr(true),
		OrderByIDDesc: lo.ToPtr(true),
		Name:          "test",
	}).Return([]*repo.Repo{
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
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().List(gomock.Any(), &repo.ListRepoRequest{
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
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Repo{
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
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))

	res, err := svc.Show(context.TODO(), &reposerver.ShowRequest{
		Id: 1,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_ToggleEnabled_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		eventRepo,
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().ToggleEnabled(gomock.Any(), 1, true).Return(&repo.Repo{
		ID:      1,
		Name:    "toggle",
		Enabled: true,
	}, nil)

	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

	res, err := svc.ToggleEnabled(newAdminUserCtx(), &reposerver.ToggleEnabledRequest{
		Id:      1,
		Enabled: true,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(1), res.Item.Id)
	assert.Equal(t, "toggle", res.Item.Name)
	assert.Equal(t, true, res.Item.Enabled)
}

func TestRepoSvc_ToggleEnabled_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().ToggleEnabled(gomock.Any(), 1, true).Return(nil, errors.New("error"))

	res, err := svc.ToggleEnabled(newAdminUserCtx(), &reposerver.ToggleEnabledRequest{
		Id:      1,
		Enabled: true,
	})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepoSvc_Update_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		eventRepo,
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Repo{
		ID:   1,
		Name: "update",
	}, nil)

	repoRepo.EXPECT().Update(gomock.Any(), &repo.UpdateRepoInput{
		ID:           1,
		Name:         "updated",
		NeedGitRepo:  true,
		GitProjectID: lo.ToPtr(int32(1)),
		MarsConfig:   &mars.Config{},
		Description:  "updated description",
	}).Return(&repo.Repo{
		ID:   1,
		Name: "updated",
	}, nil)

	eventRepo.EXPECT().AuditLogWithChange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

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
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("error"))

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
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := repo.NewMockRepoRepo(m)
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repoRepo,
	)

	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Repo{}, nil)
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
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repo.NewMockRepoRepo(m),
	).(*repoSvc)

	ctx := newAdminUserCtx()
	_, err := svc.Authorize(ctx, "List")

	assert.Nil(t, err)
}

func TestRepoSvc_Authorize_AdminUser2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repo.NewMockRepoRepo(m),
	).(*repoSvc)

	ctx := newAdminUserCtx()
	_, err := svc.Authorize(ctx, "XX")

	assert.Nil(t, err)
}

func TestRepoSvc_Authorize_ListMethod(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repo.NewMockRepoRepo(m),
	).(*repoSvc)

	ctx := newOtherUserCtx()
	_, err := svc.Authorize(ctx, "List")

	assert.Nil(t, err)
}

func TestRepoSvc_Authorize_NonListMethod(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewRepoSvc(
		mlog.NewForConfig(nil),
		repo.NewMockEventRepo(m),
		repo.NewMockGitRepo(m),
		repo.NewMockRepoRepo(m),
	).(*repoSvc)

	ctx := newOtherUserCtx()
	_, err := svc.Authorize(ctx, "NonList")

	assert.NotNil(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}
