package services

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/changelog"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// changelogSvcMocks 聚合 changelogSvc 的全部下游 mock，由 newChangelogSvcWithMocks 统一构造。
type changelogSvcMocks struct {
	ctrl     *gomock.Controller
	clRepo   *data.MockChangelogRepo
	projRepo *data.MockProjectRepo
	nsRepo   *data.MockNamespaceRepo
}

func newChangelogSvcWithMocks(t *testing.T) (*changelogSvc, *changelogSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &changelogSvcMocks{
		ctrl:     ctrl,
		clRepo:   data.NewMockChangelogRepo(ctrl),
		projRepo: data.NewMockProjectRepo(ctrl),
		nsRepo:   data.NewMockNamespaceRepo(ctrl),
	}
	logger := mlog.NewForConfig(nil)
	s, ok := NewChangelogSvc(ChangelogSvcDeps{
		ClBiz:     biz.NewChangelogBiz(mocks.clRepo),
		Logger:    logger,
		AccessBiz: biz.NewAccessBiz(biz.NewNamespaceBiz(logger, mocks.nsRepo, nil, nil, nil), biz.NewProjectBiz(logger, mocks.projRepo, nil, nil)),
	}).(*changelogSvc)
	if !ok {
		panic("NewChangelogSvc returned unexpected type")
	}
	return s, mocks
}

func TestNewChangelogSvc(t *testing.T) {
	svc, _ := newChangelogSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.clBiz)
	assert.NotNil(t, svc.accessBiz)
}

func Test_changelogSvc_FindLastChangelogsByProjectID_RepoError(t *testing.T) {
	svc, mocks := newChangelogSvcWithMocks(t)
	clRepo, projRepo, nsRepo := mocks.clRepo, mocks.projRepo, mocks.nsRepo

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "a"}, nil)
	clRepo.EXPECT().FindLastChangelogsByProjectID(gomock.Any(), &biz.FindLastChangelogsByProjectIDChangeLogInput{
		OnlyChanged:        true,
		ProjectID:          1,
		OrderByVersionDesc: lo.ToPtr(true),
		Limit:              5,
	}).Return(nil, errors.New("x"))

	_, err := svc.FindLastChangelogsByProjectID(newAdminUserCtx(), &changelog.FindLastChangelogsByProjectIDRequest{
		ProjectId:   1,
		OnlyChanged: true,
	})
	assert.Error(t, err)
}

func Test_changelogSvc_FindLastChangelogsByProjectID_ProjectShowError(t *testing.T) {
	svc, mocks := newChangelogSvcWithMocks(t)
	projRepo := mocks.projRepo

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("project error"))

	_, err := svc.FindLastChangelogsByProjectID(newAdminUserCtx(), &changelog.FindLastChangelogsByProjectIDRequest{
		ProjectId:   1,
		OnlyChanged: true,
	})
	assert.Error(t, err)
}

func Test_changelogSvc_FindLastChangelogsByProjectID_NamespaceShowError(t *testing.T) {
	svc, mocks := newChangelogSvcWithMocks(t)
	projRepo, nsRepo := mocks.projRepo, mocks.nsRepo

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("namespace error"))

	_, err := svc.FindLastChangelogsByProjectID(newAdminUserCtx(), &changelog.FindLastChangelogsByProjectIDRequest{
		ProjectId:   1,
		OnlyChanged: true,
	})
	assert.Error(t, err)
}

// 回归防护：私有命名空间项目的 changelog（含部署配置 + 环境变量）不允许被
// 非 admin / 非创建者 / 非成员读取。去掉 FindLastChangelogsByProjectID 里的
// CanAccess 检查，本测试必须失败。
func Test_changelogSvc_FindLastChangelogsByProjectID_AccessDenied(t *testing.T) {
	svc, mocks := newChangelogSvcWithMocks(t)
	projRepo, nsRepo := mocks.projRepo, mocks.nsRepo

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: true, CreatorEmail: "other@x.com"}, nil)

	resp, err := svc.FindLastChangelogsByProjectID(newOtherUserCtx(), &changelog.FindLastChangelogsByProjectIDRequest{
		ProjectId:   1,
		OnlyChanged: true,
	})
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func Test_changelogSvc_FindLastChangelogsByProjectID_Success(t *testing.T) {
	svc, mocks := newChangelogSvcWithMocks(t)
	clRepo, projRepo, nsRepo := mocks.clRepo, mocks.projRepo, mocks.nsRepo

	projRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "a"}, nil)
	clRepo.EXPECT().FindLastChangelogsByProjectID(gomock.Any(), &biz.FindLastChangelogsByProjectIDChangeLogInput{
		OnlyChanged:        true,
		ProjectID:          1,
		OrderByVersionDesc: lo.ToPtr(true),
		Limit:              5,
	}).Return([]*biz.Changelog{}, nil)

	resp, err := svc.FindLastChangelogsByProjectID(newAdminUserCtx(), &changelog.FindLastChangelogsByProjectIDRequest{
		ProjectId:   1,
		OnlyChanged: true,
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}
