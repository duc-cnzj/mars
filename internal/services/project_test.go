package services

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/project"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"helm.sh/helm/v3/pkg/storage/driver"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewProjectSvc(t *testing.T) {
	svc, _ := newProjectSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.repoBiz)
	assert.NotNil(t, svc.plMgr)
	assert.NotNil(t, svc.jobManager)
	assert.NotNil(t, svc.projBiz)
	assert.NotNil(t, svc.gitBiz)
	assert.NotNil(t, svc.k8sBiz)
	assert.NotNil(t, svc.eventBiz)
	assert.NotNil(t, svc.logger)
}

func Test_projectSvc_List(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().List(gomock.Any(), &biz.ListProjectInput{
		Page:          1,
		PageSize:      11,
		OrderByIDDesc: lo.ToPtr(true),
		Email:         adminEmail,
		IsAdmin:       true,
	}).Return(nil, nil, errors.New("x"))
	list, err := svc.List(newAdminUserCtx(), &project.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(11)),
	})
	assert.Error(t, err)
	assert.Nil(t, list)
}

func Test_projectSvc_List_Success(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().List(gomock.Any(), &biz.ListProjectInput{
		Page:          1,
		PageSize:      11,
		OrderByIDDesc: lo.ToPtr(true),
		Email:         adminEmail,
		IsAdmin:       true,
	}).Return([]*biz.Project{{ID: 1, Name: "p1"}}, &pagination.Pagination{Page: 3, PageSize: 11, Count: 1}, nil)
	list, err := svc.List(newAdminUserCtx(), &project.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(11)),
	})
	assert.Nil(t, err)
	if assert.NotNil(t, list) {
		assert.Equal(t, int32(3), list.Page)
		assert.Equal(t, int32(11), list.PageSize)
		assert.Equal(t, int32(1), list.Count)
		if assert.Len(t, list.Items, 1) {
			assert.Equal(t, int32(1), list.Items[0].Id)
			assert.Equal(t, "p1", list.Items[0].Name)
		}
	}
}

func Test_projectSvc_List_NonAdmin(t *testing.T) {
	// 回归防护：非 admin 用户透传自己的 Email 与 IsAdmin=false 到 data 层，
	// data 层才会按命名空间访问谓词过滤可见项目。改坏实现（IsAdmin 恒 true /
	// Email 写死为他人）会让非 admin 看到全部项目，此测试必须 FAIL。
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().List(gomock.Any(), &biz.ListProjectInput{
		Page:          1,
		PageSize:      11,
		OrderByIDDesc: lo.ToPtr(true),
		Email:         "user@mars.com",
		IsAdmin:       false,
	}).Return([]*biz.Project{{ID: 7, Name: "p7"}}, &pagination.Pagination{Page: 1, PageSize: 11, Count: 1}, nil)
	list, err := svc.List(newOtherUserCtx(), &project.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(11)),
	})
	assert.Nil(t, err)
	if assert.NotNil(t, list) {
		assert.Equal(t, int32(1), list.Count)
		if assert.Len(t, list.Items, 1) {
			assert.Equal(t, int32(7), list.Items[0].Id)
			assert.Equal(t, "p7", list.Items[0].Name)
		}
	}
}

func TestProjectSvc_Show_Success(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: false}, nil)
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{
		Namespace:   &biz.Namespace{ID: 1},
		NamespaceID: 1,
	}, nil)
	res, err := svc.Show(newAdminUserCtx(), &project.ShowRequest{
		Id: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestProjectSvc_Show_Failure(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	res, err := svc.Show(context.TODO(), &project.ShowRequest{
		Id: 1,
	})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestProjectSvc_Show_Failure2(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: true}, nil)
	_, err := svc.Show(newOtherUserCtx(), &project.ShowRequest{
		Id: 1,
	})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func Test_projectSvc_Delete(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: false}, nil)
	// 卸载成功 → 继续删 DB 记录
	mocks.helmerRepo.EXPECT().Uninstall("app", "ns", gomock.Any()).Return(nil)
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{
		ID:          2,
		Name:        "app",
		NamespaceID: 1,
		Namespace: &biz.Namespace{
			Name: "ns",
		},
	}, nil)
	req := &project.DeleteRequest{
		Id: 1,
	}
	mocks.eventRepo.EXPECT().AuditLogWithRequest(types.EventActionType_Delete, biz.MustGetUser(newAdminUserCtx()).Name, gomock.Any(), req)
	mocks.eventRepo.EXPECT().Dispatch(biz.EventProjectDeleted, &biz.ProjectDeletedPayload{
		NamespaceID: 1,
		ProjectID:   2,
	})
	mocks.projectRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)

	response, err := svc.Delete(newAdminUserCtx(), req)
	assert.Nil(t, err)
	assert.NotNil(t, response)
}

func Test_projectSvc_Delete_UninstallNotFound(t *testing.T) {
	// release 已不存在（手动清理/孤儿）不算失败，继续删除 DB 记录，
	// 避免把已经没 release 的项目锁死无法删除。
	svc, mocks := newProjectSvcWithMocks(t)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: false}, nil)
	mocks.helmerRepo.EXPECT().Uninstall("app", "ns", gomock.Any()).Return(driver.ErrReleaseNotFound)
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{
		ID:          2,
		Name:        "app",
		NamespaceID: 1,
		Namespace:   &biz.Namespace{Name: "ns"},
	}, nil)
	mocks.eventRepo.EXPECT().AuditLogWithRequest(types.EventActionType_Delete, biz.MustGetUser(newAdminUserCtx()).Name, gomock.Any(), gomock.Any())
	mocks.eventRepo.EXPECT().Dispatch(biz.EventProjectDeleted, &biz.ProjectDeletedPayload{NamespaceID: 1, ProjectID: 2})
	mocks.projectRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)

	res, err := svc.Delete(newAdminUserCtx(), &project.DeleteRequest{Id: 1})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func Test_projectSvc_Delete_UninstallError(t *testing.T) {
	// 回归防护：卸载 release 失败（非 not-found）必须中止删除、保留 DB 记录，
	// 否则会留下无记录、无法重试的孤儿 release。改坏实现（先删 DB 后卸载/忽略
	// 卸载错误继续删）时此测试 FAIL。
	svc, mocks := newProjectSvcWithMocks(t)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: false}, nil)
	mocks.helmerRepo.EXPECT().Uninstall("app", "ns", gomock.Any()).Return(errors.New("uninstall boom"))
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{
		ID:          2,
		Name:        "app",
		NamespaceID: 1,
		Namespace:   &biz.Namespace{Name: "ns"},
	}, nil)
	// 卸载失败后必须中止：以下三个副作用都不该发生（gomock 遇到未期望调用会 FAIL）
	// projectRepo.Delete / Dispatch / AuditLogWithRequest 均未设置期望

	res, err := svc.Delete(newAdminUserCtx(), &project.DeleteRequest{Id: 1})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func Test_projectSvc_Delete_Fail(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	response, err := svc.Delete(newAdminUserCtx(), &project.DeleteRequest{
		Id: 1,
	})
	assert.NotNil(t, err)
	assert.Nil(t, response)
}

func Test_projectSvc_Delete_Fail2(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{Private: false}, nil)
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{Namespace: &biz.Namespace{Name: "ns"}}, nil)
	// 新顺序下先卸载再删 DB：卸载成功才轮到 projectRepo.Delete 失败
	mocks.helmerRepo.EXPECT().Uninstall(gomock.Any(), "ns", gomock.Any()).Return(nil)
	mocks.projectRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errors.New("x"))
	response, err := svc.Delete(newAdminUserCtx(), &project.DeleteRequest{
		Id: 1,
	})
	assert.NotNil(t, err)
	assert.Nil(t, response)
}

func Test_projectSvc_Delete_Fail3(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{Private: true}, nil)
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{Name: "app", Namespace: &biz.Namespace{Name: "ns"}}, nil)
	// 若访问控制被删除，请求会继续卸载→删 DB→发事件→审计，返回成功（非 403），
	// 让 assert.ErrorIs 干净失败而非 nil 解引用 panic。下游 mock 用 AnyTimes 保证走通。
	mocks.helmerRepo.EXPECT().Uninstall(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mocks.projectRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mocks.eventRepo.EXPECT().Dispatch(gomock.Any(), gomock.Any()).AnyTimes()
	mocks.eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	_, err := svc.Delete(newOtherUserCtx(), &project.DeleteRequest{
		Id: 1,
	})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func Test_projectSvc_Version(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "a"}, nil)
	mocks.projectRepo.EXPECT().Version(gomock.Any(), 1).Return(100, nil)
	version, err := svc.Version(newAdminUserCtx(), &project.VersionRequest{Id: 1})
	assert.Nil(t, err)
	assert.NotNil(t, version)
	assert.Equal(t, int32(100), version.Version)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "a"}, nil)
	mocks.projectRepo.EXPECT().Version(gomock.Any(), 1).Return(0, errors.New("x"))
	_, err = svc.Version(newAdminUserCtx(), &project.VersionRequest{Id: 1})
	assert.Error(t, err)
}

// 回归防护：私有命名空间的项目版本不允许被非 admin / 非创建者 / 非成员读取。
// 去掉 Version 里的 CanAccess 检查，本测试必须失败。
func Test_projectSvc_Version_AccessDenied(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: true, CreatorEmail: "other@x.com"}, nil)
	// 若访问控制被删除，请求会继续走到 Version → 返回 100，而非 errs.ErrorPermissionDenied
	mocks.projectRepo.EXPECT().Version(gomock.Any(), 1).Return(100, nil).AnyTimes()

	resp, err := svc.Version(newOtherUserCtx(), &project.VersionRequest{Id: 1})
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

// 回归防护：Version 的 Show/nsRepo 门禁分支（项目或命名空间不存在/DB 故障）。
func Test_projectSvc_Version_ShowError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("project error"))

	resp, err := svc.Version(newAdminUserCtx(), &project.VersionRequest{Id: 1})
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func Test_projectSvc_Version_NamespaceError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("namespace error"))

	resp, err := svc.Version(newAdminUserCtx(), &project.VersionRequest{Id: 1})
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func Test_projectSvc_AllContainers(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	// Show 被调用两次：一次是 svc 层 access-check，一次是 biz.GetAllActiveContainers 内部取项目。
	// 项目无 PodSelectors 时 buildStateContainers 提前返回，不触发 k8sRepo 调用。
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil).Times(2)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)
	containers, err := svc.AllContainers(newAdminUserCtx(), &project.AllContainersRequest{
		Id: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, containers)
}

func Test_projectSvc_AllContainers_Fail(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	// 第一次 Show（access-check）成功，第二次 Show（biz 内部）失败 → 整体报错。
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil).Times(1)
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)
	containers, err := svc.AllContainers(newAdminUserCtx(), &project.AllContainersRequest{
		Id: 1,
	})
	assert.NotNil(t, err)
	assert.Nil(t, containers)
}

func TestProjectSvc_WebApply_Success(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{Name: "test", NeedGitRepo: true, GitProjectID: 100}, nil)
	mocks.gitRepo.EXPECT().ListCommits(gomock.Any(), 100, "dev").Return([]*biz.Commit{{ID: "commit-id"}}, nil)

	job := deploy.NewMockJob(mocks.ctrl)
	mocks.jobManager.EXPECT().NewJob(gomock.Any()).Return(job)

	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(nil)
	job.EXPECT().Manifests().Return([]string{"manifests"})
	job.EXPECT().IsNotDryRun().Return(true)
	job.EXPECT().Project().Return(&biz.Project{ID: 1})

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{}, nil)

	_, err := svc.WebApply(newAdminUserCtx(), &project.WebApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		GitBranch:   "dev",
	})

	assert.Nil(t, err)
}

func TestProjectSvc_WebApply_DryRun(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{Name: "test", NeedGitRepo: true, GitProjectID: 100}, nil)
	mocks.gitRepo.EXPECT().ListCommits(gomock.Any(), 100, "dev").Return([]*biz.Commit{{ID: "commit-id"}}, nil)

	job := deploy.NewMockJob(mocks.ctrl)
	var gotInput *deploy.JobInput
	mocks.jobManager.EXPECT().NewJob(gomock.Any()).DoAndReturn(func(input *deploy.JobInput) deploy.Job {
		gotInput = input
		return job
	})

	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(nil)
	job.EXPECT().Manifests().Return([]string{"manifests"})
	// dry-run 不创建项目：IsNotDryRun 为 false，不走 projectRepo.Show。
	job.EXPECT().IsNotDryRun().Return(false)

	resp, err := svc.WebApply(newAdminUserCtx(), &project.WebApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		GitBranch:   "dev",
		DryRun:      true,
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetDryRun())
	assert.True(t, gotInput.DryRun)
	assert.Equal(t, []string{"manifests"}, resp.GetYamlFiles())
	assert.Nil(t, resp.GetProject())
}

func TestProjectSvc_WebApply_Failure(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	_, err := svc.WebApply(newAdminUserCtx(), &project.WebApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		Name:        "test",
	})

	assert.NotNil(t, err)
	assert.Equal(t, "error", err.Error())
}

func TestProjectSvc_Apply_Success(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&biz.Repo{Name: "test", NeedGitRepo: true}, nil)
	mocks.gitRepo.EXPECT().ListCommits(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*biz.Commit{{ID: "commit-id"}}, nil)

	job := deploy.NewMockJob(mocks.ctrl)
	mocks.jobManager.EXPECT().NewJob(gomock.Any()).Return(job)

	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(nil)

	mockServer := &mockProjectApplyServer{
		Req: &project.ApplyRequest{
			RepoId:      1,
			NamespaceId: 1,
			Name:        "test",
		},
	}

	applyRequest := &project.ApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		Name:        "test",
	}

	err := svc.Apply(applyRequest, mockServer)

	assert.Nil(t, err)
}

func TestProjectSvc_Apply_Failure(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	mockServer := &mockProjectApplyServer{
		Req: &project.ApplyRequest{
			RepoId:      1,
			NamespaceId: 1,
			Name:        "test",
		},
	}

	applyRequest := &project.ApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		Name:        "test",
	}

	err := svc.Apply(applyRequest, mockServer)

	assert.NotNil(t, err)
	assert.Equal(t, "error", err.Error())
}

func TestProjectSvc_Apply_Failure2(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: true}, nil)

	mockServer := &mockProjectApplyServer{
		Ctx: newOtherUserCtx(),
		Req: &project.ApplyRequest{
			RepoId:      1,
			NamespaceId: 1,
			Name:        "test",
		},
	}

	applyRequest := &project.ApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		Name:        "test",
	}

	err := svc.Apply(applyRequest, mockServer)

	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestProjectSvc_Apply_InstallError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&biz.Repo{Name: "test", NeedGitRepo: false}, nil)

	job := deploy.NewMockJob(mocks.ctrl)
	mocks.jobManager.EXPECT().NewJob(gomock.Any()).Return(job)

	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(errors.New("install boom"))

	mockServer := &mockProjectApplyServer{
		Req: &project.ApplyRequest{
			RepoId:      1,
			NamespaceId: 1,
			Name:        "test",
		},
	}

	applyRequest := &project.ApplyRequest{
		RepoId:      1,
		NamespaceId: 1,
		Name:        "test",
	}

	err := svc.Apply(applyRequest, mockServer)

	assert.NotNil(t, err)
	assert.Equal(t, "install boom", err.Error())
}

type mockProjectApplyServer struct {
	grpc.ServerStream
	Req *project.ApplyRequest
	Ctx context.Context
}

func (x *mockProjectApplyServer) Send(m *project.ApplyResponse) error {
	return nil
}

func (x *mockProjectApplyServer) SetHeader(md metadata.MD) error {
	return nil
}

func (x *mockProjectApplyServer) SendHeader(md metadata.MD) error {
	return nil
}

func (x *mockProjectApplyServer) SetTrailer(md metadata.MD) {
}

func (x *mockProjectApplyServer) Context() context.Context {
	if x.Ctx != nil {
		return x.Ctx
	}
	// 默认注入 admin 用户，适配 MustGetUser panic 语义；权限测试通过 Ctx 覆盖。
	return newAdminUserCtx()
}

func (x *mockProjectApplyServer) SendMsg(m any) error {
	return nil
}

func (x *mockProjectApplyServer) RecvMsg(m any) error {
	return nil
}

func TestMessager_Current(t *testing.T) {
	m := newMessager(true, websocket.Type_ApplyProject, nil)
	current := m.Current()
	assert.Equal(t, int64(0), current)
}

// TestMessager_SetSlug 覆盖 slug 就地回填：创建部署名缺省解析后由 ApplyProject 调用，
// 保证出站帧携带最终名（前端 toSlug 关联的日志 key）。
func TestMessager_SetSlug(t *testing.T) {
	m := newMessager(true, websocket.Type_ApplyProject, nil)
	m.(*messager).SetSlug("new-slug")
	assert.Equal(t, "new-slug", m.(*messager).slugName)
}

type mockApplyServer struct {
	project.Project_ApplyServer

	response *project.ApplyResponse
}

func (m *mockApplyServer) Send(response *project.ApplyResponse) error {
	m.response = response
	return nil
}

func TestMessager_Add(t *testing.T) {
	server := &mockApplyServer{}
	m := newMessager(true, websocket.Type_ApplyProject, server)
	m.Add()
	current := m.Current()
	assert.Equal(t, websocket.Type_ProcessPercent, server.response.Metadata.Type)
	assert.Equal(t, int64(1), current)
}

func TestMessager_To(t *testing.T) {
	server := &mockApplyServer{}
	m := newMessager(true, websocket.Type_ApplyProject, server)
	m.To(50)
	current := m.Current()
	assert.Equal(t, int32(50), server.response.Metadata.Percent)
	assert.Equal(t, int64(50), current)
}

func TestMessager_SendEndError(t *testing.T) {
	server := &mockApplyServer{}
	m := newMessager(true, websocket.Type_ApplyProject, server)
	m.SendEndError(errors.New("test error"))
	assert.True(t, server.response.Metadata.End)
	assert.Equal(t, "test error", server.response.Metadata.Message)
}

func TestMessager_SendMsg(t *testing.T) {
	server := &mockApplyServer{}
	m := newMessager(true, websocket.Type_ApplyProject, server)
	m.SendMsg("test message")
	assert.False(t, server.response.Metadata.End)
	assert.Equal(t, "test message", server.response.Metadata.Message)
}

type mockWsMessage struct {
	app.WebsocketMessage
}

func (m *mockWsMessage) GetMetadata() *websocket.Metadata {
	return &websocket.Metadata{
		Type: websocket.Type_ApplyProject,
	}
}

func TestMessager_SendProtoMsg(t *testing.T) {
	server := &mockApplyServer{}
	m := newMessager(true, websocket.Type_ApplyProject, server)
	m.SendProtoMsg(&mockWsMessage{})
	assert.Equal(t, websocket.Type_ApplyProject, server.response.Metadata.Type)
}

func TestMessager_SendProcessPercent(t *testing.T) {
	server := &mockApplyServer{}
	m := newMessager(true, websocket.Type_ApplyProject, server)
	m.SendProcessPercent(50)
	assert.Equal(t, websocket.Type_ProcessPercent, server.response.Metadata.Type)
	assert.Equal(t, int32(50), server.response.Metadata.Percent)
}

func TestMessager_SendMsgWithContainerLog(t *testing.T) {
	server := &mockApplyServer{}
	m := newMessager(true, websocket.Type_ApplyProject, server)
	m.SendMsgWithContainerLog("test message", []*websocket.Container{})
	assert.False(t, server.response.Metadata.End)
	assert.Equal(t, "test message", server.response.Metadata.Message)
	assert.Equal(t, websocket.ResultType_LogWithContainers, server.response.Metadata.Result)
}

func TestMessager_SendDeployedResult(t *testing.T) {
	server := &mockApplyServer{}
	m := newMessager(true, websocket.Type_ApplyProject, server)
	m.SendDeployedResult(websocket.ResultType_Success, "test message", &types.ProjectModel{})
	assert.True(t, server.response.Metadata.End)
	assert.Equal(t, "test message", server.response.Metadata.Message)
	assert.Equal(t, websocket.ResultType_Success, server.response.Metadata.Result)
}

func Test_projectSvc_MemoryCpuAndEndpoints(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	projModel := &biz.Project{
		ID:          2,
		Name:        "app",
		NamespaceID: 1,
		Namespace: &biz.Namespace{
			Name: "ns",
		},
		Version: 100,
	}
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(projModel, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)
	mocks.k8sRepo.EXPECT().GetAllPodMetrics(gomock.Any(), projModel)
	mocks.k8sRepo.EXPECT().GetCpuAndMemory(gomock.Any(), gomock.Any()).Return("1", "2Gi")
	// endpoint 编排已迁入 biz：由 FindProjectsByIDs 取项目 + 空 listers 产出 0 个端点。
	mocks.projectRepo.EXPECT().FindProjectsByIDs(gomock.Any(), projModel.ID).Return([]*biz.Project{projModel}, nil)
	mocks.k8sRepo.EXPECT().GatewayApiInstalled().Return(false)
	mocks.k8sRepo.EXPECT().ListServices(gomock.Any()).Return(nil, nil).AnyTimes()
	mocks.k8sRepo.EXPECT().ListIngresses(gomock.Any()).Return(nil, nil)
	mocks.k8sRepo.EXPECT().ExternalIp().Return("127.0.0.1")
	endpoints, err := svc.MemoryCpuAndEndpoints(newAdminUserCtx(), &project.MemoryCpuAndEndpointsRequest{Id: 1})
	assert.Nil(t, err)
	assert.Equal(t, "1", endpoints.Cpu)
	assert.Equal(t, "2Gi", endpoints.Memory)
	assert.Equal(t, 0, len(endpoints.Urls))
}

func Test_projectSvc_MemoryCpuAndEndpoints_Fail(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	_, err := svc.MemoryCpuAndEndpoints(context.TODO(), &project.MemoryCpuAndEndpointsRequest{Id: 1})
	assert.Error(t, err)
}

func Test_projectSvc_MemoryCpuAndEndpoints_fail2(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	projModel := &biz.Project{
		ID:          2,
		Name:        "app",
		NamespaceID: 1,
		Namespace: &biz.Namespace{
			Name: "ns",
		},
		Version: 100,
	}
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(projModel, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)
	mocks.k8sRepo.EXPECT().GetAllPodMetrics(gomock.Any(), projModel)
	mocks.k8sRepo.EXPECT().GetCpuAndMemory(gomock.Any(), gomock.Any()).Return("1", "2Gi")
	// endpoint 编排失败路径：FindProjectsByIDs 报错 → svc 整体返回错误，listers 不会被调用。
	mocks.projectRepo.EXPECT().FindProjectsByIDs(gomock.Any(), projModel.ID).Return(nil, errors.New("x"))
	endpoints, err := svc.MemoryCpuAndEndpoints(newAdminUserCtx(), &project.MemoryCpuAndEndpointsRequest{Id: 1})
	assert.Error(t, err)
	assert.Nil(t, endpoints)
}

// newProjectSvcWithMocks 构建一个挂满 fresh mock 的 projectSvc，用于 apply 路径的测试。
// projectSvcMocks 聚合 newProjectSvcWithMocks 创建的全部 mock，测试按字段取用。
// mocks.ctrl 仅供临时创建额外 mock（如 deploy.NewMockJob(mocks.ctrl)）。
type projectSvcMocks struct {
	ctrl        *gomock.Controller
	projectRepo *data.MockProjectRepo
	repoRepo    *data.MockRepoRepo
	gitRepo     *data.MockGitRepo
	k8sRepo     *data.MockK8sRepo
	eventRepo   *data.MockEventRepo
	helmerRepo  *data.MockHelmerRepo
	nsRepo      *data.MockNamespaceRepo
	plMgr       *app.MockPluginManager
	jobManager  *deploy.MockJobManager
}

// newProjectSvcWithMocks 构造带全套 mock 的 projectSvc，消除各测试重复的
// NewProjectSvc(ProjectSvcDeps{...}) 样板。controller 由 testing.T 托管，无需手动 Finish。
func newProjectSvcWithMocks(t *testing.T) (*projectSvc, *projectSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &projectSvcMocks{
		ctrl:        ctrl,
		projectRepo: data.NewMockProjectRepo(ctrl),
		repoRepo:    data.NewMockRepoRepo(ctrl),
		gitRepo:     data.NewMockGitRepo(ctrl),
		k8sRepo:     data.NewMockK8sRepo(ctrl),
		eventRepo:   data.NewMockEventRepo(ctrl),
		helmerRepo:  data.NewMockHelmerRepo(ctrl),
		nsRepo:      data.NewMockNamespaceRepo(ctrl),
		plMgr:       app.NewMockPluginManager(ctrl),
		jobManager:  deploy.NewMockJobManager(ctrl),
	}
	logger := mlog.NewForConfig(nil)
	s, ok := NewProjectSvc(ProjectSvcDeps{
		RepoBiz:    biz.NewRepoBiz(mocks.repoRepo),
		PluginMgr:  mocks.plMgr,
		JobManager: mocks.jobManager,
		ProjBiz:    biz.NewProjectBiz(logger, mocks.projectRepo, mocks.k8sRepo),
		GitBiz:     biz.NewGitBiz(mocks.gitRepo),
		K8sBiz:     biz.NewK8sBiz(mocks.k8sRepo),
		EventBiz:   biz.NewEventBiz(mocks.eventRepo),
		Logger:     logger,
		DeployBiz:  biz.NewDeployBiz(logger, mocks.projectRepo, mocks.helmerRepo, mocks.eventRepo),
		AccessBiz:  biz.NewAccessBiz(biz.NewNamespaceBiz(logger, mocks.nsRepo, nil, nil, nil), biz.NewProjectBiz(logger, mocks.projectRepo, mocks.k8sRepo)),
	}).(*projectSvc)
	if !ok {
		panic("NewProjectSvc returned unexpected type")
	}
	return s, mocks
}

// mockInstallProjectChain 把 InstallProject 的整条 job 管线串起来。
func mockInstallProjectChain(job *deploy.MockJob) {
	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(nil)
}

// ctxAwareApplyServer 允许 Apply 测试注入自定义 context。
type ctxAwareApplyServer struct {
	mockProjectApplyServer
	ctx context.Context
}

func (x *ctxAwareApplyServer) Context() context.Context { return x.ctx }

func TestProjectSvc_WebApply_ShowError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{Name: "test"}, nil)
	job := deploy.NewMockJob(mocks.ctrl)
	mocks.jobManager.EXPECT().NewJob(gomock.Any()).Return(job)
	mockInstallProjectChain(job)
	// 非 dry-run：apply 成功后再 Show 项目详情，但 Show 失败
	job.EXPECT().IsNotDryRun().Return(true)
	job.EXPECT().Project().Return(&biz.Project{ID: 1})
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))

	_, err := svc.WebApply(newAdminUserCtx(), &project.WebApplyRequest{NamespaceId: 1, RepoId: 1})
	assert.Error(t, err)
}

func Test_projectSvc_apply_NsShowError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))

	_, err := svc.WebApply(newAdminUserCtx(), &project.WebApplyRequest{NamespaceId: 1, RepoId: 1})
	assert.Error(t, err)
}

func TestProjectSvc_Apply_WebsocketSync(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&biz.Repo{Name: "test"}, nil)

	wsSender := app.NewMockWsSender(mocks.ctrl)
	pubsub := app.NewMockPubSub(mocks.ctrl)
	mocks.plMgr.EXPECT().Ws().Return(wsSender)
	wsSender.EXPECT().New("", "").Return(pubsub)
	pubsub.EXPECT().Close()

	job := deploy.NewMockJob(mocks.ctrl)
	mocks.jobManager.EXPECT().NewJob(gomock.Any()).Return(job)
	mockInstallProjectChain(job)

	err := svc.Apply(&project.ApplyRequest{
		RepoId: 1, NamespaceId: 1, Name: "test", WebsocketSync: true,
	}, &mockProjectApplyServer{})
	assert.Nil(t, err)
}

func TestProjectSvc_Apply_NoCommits(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{Name: "test", NeedGitRepo: true, GitProjectID: 10, DefaultBranch: "dev"}, nil)
	// 分支/commit 都缺省，ListCommits 返回空 → apply 提前返回错误
	mocks.gitRepo.EXPECT().ListCommits(gomock.Any(), 10, "dev").Return([]*biz.Commit{}, nil)

	_, err := svc.WebApply(newAdminUserCtx(), &project.WebApplyRequest{NamespaceId: 1, RepoId: 1})
	assert.ErrorContains(t, err, "没有可用的 commit")
}

func TestProjectSvc_Apply_VersionFindSuccess(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{Name: "test"}, nil)
	// Version > 0 → 查已存在的项目，命中则用其 ID
	mocks.projectRepo.EXPECT().FindByName(gomock.Any(), "test", 1).Return(&biz.Project{ID: 5}, nil)
	job := deploy.NewMockJob(mocks.ctrl)
	mocks.jobManager.EXPECT().NewJob(gomock.Any()).Return(job)
	mockInstallProjectChain(job)
	job.EXPECT().IsNotDryRun().Return(false)
	job.EXPECT().Manifests().Return([]string{})

	_, err := svc.WebApply(newAdminUserCtx(), &project.WebApplyRequest{NamespaceId: 1, RepoId: 1, Version: lo.ToPtr(int32(1))})
	assert.Nil(t, err)
}

func TestProjectSvc_Apply_VersionFindError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{Name: "test"}, nil)
	// Version > 0 但 FindByName 找不到（真 NotFound）→ 放行，projectID 保持 0；
	// data 边界把 ent 的 NotFound 转成 gRPC NotFound，mock 用 errs.WrapNotFound() 模拟。
	mocks.projectRepo.EXPECT().FindByName(gomock.Any(), "test", 1).Return(nil, errs.WrapNotFound(errors.New("not found"), "find project by name"))
	job := deploy.NewMockJob(mocks.ctrl)
	mocks.jobManager.EXPECT().NewJob(gomock.Any()).Return(job)
	mockInstallProjectChain(job)
	job.EXPECT().IsNotDryRun().Return(false)
	job.EXPECT().Manifests().Return([]string{})

	_, err := svc.WebApply(newAdminUserCtx(), &project.WebApplyRequest{NamespaceId: 1, RepoId: 1, Version: lo.ToPtr(int32(1))})
	assert.Nil(t, err)
}

func TestProjectSvc_Apply_VersionFindDBError(t *testing.T) {
	// 回归防护：Version > 0 时 FindByName 返回非 NotFound 的真实 DB 故障必须
	// 上抛，不能当成"首次部署"吞掉——否则 ProjectID=0 会让 runner 报
	// "版本不匹配"，把故障伪装成预期。改坏实现（所有错误都放行）时此测试 FAIL。
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{Name: "test"}, nil)
	// DB 故障（非 NotFound）→ 必须上抛
	mocks.projectRepo.EXPECT().FindByName(gomock.Any(), "test", 1).Return(nil, errors.New("db down"))
	// 上抛后 NewJob/InstallProject 都不该发生（gomock 未期望调用会 FAIL）

	_, err := svc.WebApply(newAdminUserCtx(), &project.WebApplyRequest{NamespaceId: 1, RepoId: 1, Version: lo.ToPtr(int32(1))})
	assert.Equal(t, "db down", err.Error())
}

func TestProjectSvc_Apply_CtxCancelled(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&biz.Repo{Name: "test"}, nil)

	job := deploy.NewMockJob(mocks.ctrl)
	mocks.jobManager.EXPECT().NewJob(gomock.Any()).Return(job)
	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	runEntered := make(chan struct{})
	release := make(chan struct{})
	job.EXPECT().Run(gomock.Any()).DoAndReturn(func(ctx context.Context) deploy.Job {
		close(runEntered)
		<-release
		return job
	})
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(nil)
	stopCalled := make(chan struct{})
	// watcher goroutine: ctx 取消 → job.Stop
	job.EXPECT().Stop(gomock.Any()).Do(func(error) { close(stopCalled) })

	ctx, cancel := context.WithCancel(newAdminUserCtx())
	defer cancel()
	server := &ctxAwareApplyServer{ctx: ctx}

	done := make(chan error, 1)
	go func() {
		done <- svc.Apply(&project.ApplyRequest{
			RepoId: 1, NamespaceId: 1, Name: "test",
		}, server)
	}()

	<-runEntered
	cancel()
	// 等 watcher 调用了 job.Stop 再放行 Run，保证时序确定
	<-stopCalled
	close(release)

	assert.NoError(t, <-done)
}

func TestProjectSvc_apply_InstallProjectPanic_UnblocksWatcher(t *testing.T) {
	// 回归防护：InstallProject 一旦 panic，apply 的 defer close(ch) 必须解除 watcher
	// goroutine 的 select 阻塞并使其退出，否则 watcher 泄漏到 ctx 结束才回收。
	// 改坏实现（把 close(ch) 移出 defer / 移到 InstallProject 之后）时：
	// panic 后 watcher 仍阻塞在 select，goroutine 数保持 基线+1 不回落 → 此测试 FAIL。
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.nsRepo.EXPECT().Show(gomock.Any(), gomock.Any()).Return(&biz.Namespace{Private: false}, nil)
	mocks.repoRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&biz.Repo{Name: "test"}, nil)

	job := deploy.NewMockJob(mocks.ctrl)
	mocks.jobManager.EXPECT().NewJob(gomock.Any()).Return(job)
	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Do(func() { panic("install boom") })

	base := runtime.NumGoroutine()
	ctx := newAdminUserCtx()
	assert.Panics(t, func() {
		_, _ = svc.apply(ctx, biz.MustGetUser(ctx), newEmptyMessager(), &project.ApplyRequest{
			RepoId: 1, NamespaceId: 1, Name: "test",
		}, false)
	})

	// 修复后 watcher 因 defer close(ch) 退出，goroutine 数回落 ≤ 基线；
	// 未修复时 watcher 一直阻塞在 select（ctx.Done() 为 nil 永不触发）→ 恒为 基线+1。
	// 注意：不能再用 assert.Eventually —— 它内部会起 goroutine 跑 condition，
	// 使 runtime.NumGoroutine() 恒比基线多 1，条件永假。必须在主 goroutine 手动轮询。
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && runtime.NumGoroutine() > base {
		time.Sleep(10 * time.Millisecond)
	}
	assert.LessOrEqual(t, runtime.NumGoroutine(), base)
}

func TestProjectSvc_Show_NsShowError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	_, err := svc.Show(context.TODO(), &project.ShowRequest{Id: 1})
	assert.Error(t, err)
}

func Test_projectSvc_MemoryCpuAndEndpoints_NsShowError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	_, err := svc.MemoryCpuAndEndpoints(context.TODO(), &project.MemoryCpuAndEndpointsRequest{Id: 1})
	assert.Error(t, err)
}

func Test_projectSvc_MemoryCpuAndEndpoints_PermissionDenied(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: true}, nil)
	_, err := svc.MemoryCpuAndEndpoints(newOtherUserCtx(), &project.MemoryCpuAndEndpointsRequest{Id: 1})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func Test_projectSvc_Delete_NsShowError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	_, err := svc.Delete(newAdminUserCtx(), &project.DeleteRequest{Id: 1})
	assert.Error(t, err)
}

func Test_projectSvc_AllContainers_ShowError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	_, err := svc.AllContainers(context.TODO(), &project.AllContainersRequest{Id: 1})
	assert.Error(t, err)
}

func Test_projectSvc_AllContainers_NsShowError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))
	_, err := svc.AllContainers(context.TODO(), &project.AllContainersRequest{Id: 1})
	assert.Error(t, err)
}

func Test_projectSvc_AllContainers_PermissionDenied(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)

	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: true}, nil)
	_, err := svc.AllContainers(newOtherUserCtx(), &project.AllContainersRequest{Id: 1})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

// Test_projectSvc_CheckApplyStatus 覆盖 CheckApplyStatus 成功路径：无工作负载 → UNKNOWN，
// 容器/失败明细空，且响应前做项目级访问控制。
func Test_projectSvc_CheckApplyStatus(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	// Show 被调用两次：一次 svc 层 access-check，一次 biz.CheckApplyStatus 内部取项目。
	// 项目无 PodSelectors 时 buildStateContainers 提前返回；GetWorkloadsByManifest 返回空 → UNKNOWN。
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil).Times(2)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)
	mocks.k8sRepo.EXPECT().GetWorkloadsByManifest(gomock.Any()).Return(nil, nil, nil)

	resp, err := svc.CheckApplyStatus(newAdminUserCtx(), &project.CheckApplyStatusRequest{Id: 1})
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusUnknown, resp.GetStatus())
	assert.Contains(t, resp.GetReason(), "未发现")
	assert.Empty(t, resp.GetContainers())
	assert.Empty(t, resp.GetFailures())
}

func Test_projectSvc_CheckApplyStatus_BizError(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	// access-check 通过，biz 内部 Show 失败 → 整体报错。
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil).Times(1)
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("boom"))
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)

	resp, err := svc.CheckApplyStatus(newAdminUserCtx(), &project.CheckApplyStatusRequest{Id: 1})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "boom")
}

// Test_projectSvc_CheckApplyStatus_FailedWithFailures 覆盖 FAILED 判定 + failures 映射链路：
// Deployment 最新版本 pod CrashLoopBackOff → svc 层把领域失败诊断映射成 proto 明细并附日志。
func Test_projectSvc_CheckApplyStatus_FailedWithFailures(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{
		NamespaceID: 1,
		Namespace:   &biz.Namespace{Name: "ns"},
		Manifest:    []string{"deploy"},
	}, nil).Times(2)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "ns", Generation: 2, UID: "dep-uid"},
		Spec:       appsv1.DeploymentSpec{Replicas: lo.ToPtr(int32(1))},
		Status:     appsv1.DeploymentStatus{ObservedGeneration: 2, UpdatedReplicas: 1, AvailableReplicas: 0},
	}
	rs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rs-new", Namespace: "ns", UID: "rs-uid",
			Annotations:     map[string]string{"deployment.kubernetes.io/revision": "2"},
			OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", UID: "dep-uid"}},
		},
	}
	failPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-fail", Namespace: "ns", OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", UID: "rs-uid"}}},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "web"}}},
		Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
			Name:  "web",
			State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}},
		}}},
	}

	mocks.k8sRepo.EXPECT().GetWorkloadsByManifest(gomock.Any()).Return([]*appsv1.Deployment{dep}, nil, nil)
	mocks.k8sRepo.EXPECT().ListReplicaSets("ns").Return([]*appsv1.ReplicaSet{rs}, nil)
	mocks.k8sRepo.EXPECT().GetDeployment("ns", "web").Return(dep, nil)
	mocks.k8sRepo.EXPECT().ListPodsBySelectors(gomock.Any(), gomock.Any()).Return([]*corev1.Pod{failPod}, nil)
	mocks.k8sRepo.EXPECT().GetPodLogs(gomock.Any(), "ns", "pod-fail", gomock.Any()).Return("crash trace", nil)

	resp, err := svc.CheckApplyStatus(newAdminUserCtx(), &project.CheckApplyStatusRequest{Id: 1})
	assert.NoError(t, err)
	assert.Equal(t, types.Deploy_StatusFailed, resp.GetStatus())
	assert.Contains(t, resp.GetReason(), "CrashLoopBackOff")
	if assert.Len(t, resp.GetFailures(), 1) {
		assert.Equal(t, "Deployment", resp.GetFailures()[0].GetKind())
		assert.Equal(t, "web", resp.GetFailures()[0].GetWorkload())
		assert.Equal(t, "pod-fail", resp.GetFailures()[0].GetPod())
		assert.Equal(t, "CrashLoopBackOff", resp.GetFailures()[0].GetReason())
		assert.Equal(t, "crash trace", resp.GetFailures()[0].GetLogs())
	}
}

func Test_projectSvc_CheckApplyStatus_PermissionDenied(t *testing.T) {
	svc, mocks := newProjectSvcWithMocks(t)
	mocks.projectRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{NamespaceID: 1}, nil)
	mocks.nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Private: true}, nil)
	_, err := svc.CheckApplyStatus(newOtherUserCtx(), &project.CheckApplyStatusRequest{Id: 1})
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

// TestEmptyMessager 覆盖 WebApply 场景下空 messager 的全部 no-op 方法：
// 空对象模式要求所有方法可被安全调用且不产生任何副作用，WebApply 走部署编排时
// deploy 侧无条件调用这些方法（不判 nil），此处逐一调用断言不 panic、无状态残留。
func TestEmptyMessager(t *testing.T) {
	e := newEmptyMessager()

	// 唯一有返回值的方法：Current 恒为 0（进度未启动）。
	assert.Equal(t, int64(0), e.Current())

	// 其余方法均为空实现：调用不 panic、结果可忽略，逐一遍历防接口扩充时漏测。
	e.Add()
	e.To(100)
	e.SendEndError(errors.New("boom"))
	e.SendMsg("msg")
	e.SendProtoMsg(nil)
	e.SendProcessPercent(50)
	e.SendMsgWithContainerLog("log", nil)
	e.SendDeployedResult(websocket.ResultType_Success, "ok", nil)

	// 全部调用后状态不变（无内部计数器/缓冲可泄漏）。
	assert.Equal(t, int64(0), e.Current())
}
