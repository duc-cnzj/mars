package deploy

// apply_test.go 覆盖 ApplyProject 共享部署编排的全部分支：
// 匿名拒绝、权限校验、仓库取回、名缺省、git ensure、版本反查守卫、
// OnJob 短路、ctx watcher、InstallProject 成功/失败。
// 断言级：对 JobInput 的原地变更（名缺省/版本反查/ensure 赋值）逐一验证。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// applyTestKit 组装 ApplyProject 的最小依赖替身集合，各用例按需覆写返回值。
func applyTestKit(t *testing.T, ctrl *gomock.Controller) (
	nsRepo *biz.MockNamespaceBiz, repo *biz.MockRepoBiz, git *biz.MockGitBiz,
	proj *biz.MockProjectBiz, jobMgr *MockJobManager, msger *MockDeployMsger) {
	return biz.NewMockNamespaceBiz(ctrl), biz.NewMockRepoBiz(ctrl), biz.NewMockGitBiz(ctrl),
		biz.NewMockProjectBiz(ctrl), NewMockJobManager(ctrl), NewMockDeployMsger(ctrl)
}

// applyDepsOf 把替身包装成 ApplyProjectDeps：AccessBiz 用户提取内部走 MustGetUser
// （与生产语义一致——ApplyProject 入口已把 JobInput.User 物化进 ctx，从 ctx 取值）。
func applyDepsOf(nsRepo *biz.MockNamespaceBiz, repo *biz.MockRepoBiz, git *biz.MockGitBiz,
	proj *biz.MockProjectBiz, jobMgr *MockJobManager, msger *MockDeployMsger) ApplyProjectDeps {
	return ApplyProjectDeps{
		AccessBiz:  biz.NewAccessBiz(nsRepo, nil),
		RepoBiz:    repo,
		GitBiz:     git,
		ProjectBiz: proj,
		JobMgr:     jobMgr,
		Logger:     mlog.NewForConfig(nil),
	}
}

// pubNs 返回一个公开命名空间（Private=false），任何非 admin 用户都可访问。
func pubNs() *biz.Namespace { return &biz.Namespace{ID: 1, Name: "ns-1", Private: false} }

// privNs 返回一个私有命名空间：creator/members 均非 tester@mars.local，非 admin 无法访问。
func privNs() *biz.Namespace {
	return &biz.Namespace{
		ID:           1,
		Name:         "ns-priv",
		Private:      true,
		CreatorEmail: "owner@mars.local",
	}
}

// baseJobInput 返回一份默认 JobInput：已带用户与公开空间 ID，各用例按需覆写。
func baseJobInput() *JobInput {
	return &JobInput{
		NamespaceId: 1,
		Name:        "app",
		RepoID:      2,
		User:        &biz.UserInfo{Email: "tester@mars.local", Name: "tester"},
	}
}

// expectInstallChain 预置 InstallProject 的链式调用全部命中并返回流水线错误 err。
func expectInstallChain(job *MockJob, err error) {
	gomock.InOrder(
		job.EXPECT().GlobalLock().Return(job),
		job.EXPECT().Validate().Return(job),
		job.EXPECT().LoadConfigs().Return(job),
		job.EXPECT().Run(gomock.Any()).Return(job),
		job.EXPECT().Finish().Return(job),
		job.EXPECT().Error().Return(err),
	)
}

func TestApplyProject_AnonymousRejected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	_, repo, _, _, jobMgr, _ := applyTestKit(t, ctrl)

	job, err := ApplyProject(context.Background(), applyDepsOf(biz.NewMockNamespaceBiz(ctrl), repo, biz.NewMockGitBiz(ctrl), biz.NewMockProjectBiz(ctrl), jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{
		// User 为 nil：匿名部署直接拒绝，不触碰任何下游依赖。
		JobInput: &JobInput{NamespaceId: 1, RepoID: 2, User: nil},
	})
	assert.Nil(t, job)
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestApplyProject_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, _, _, jobMgr, _ := applyTestKit(t, ctrl)

	// 私有命名空间 + 非成员非 admin 用户 → RequireNamespaceAccessByID 拒绝。
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(privNs(), nil)

	job, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, biz.NewMockGitBiz(ctrl), biz.NewMockProjectBiz(ctrl), jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{JobInput: baseJobInput()})
	assert.Nil(t, job)
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

func TestApplyProject_RepoGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, _, _, jobMgr, _ := applyTestKit(t, ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(nil, errors.New("repo boom"))

	job, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, biz.NewMockGitBiz(ctrl), biz.NewMockProjectBiz(ctrl), jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{JobInput: baseJobInput()})
	assert.Nil(t, job)
	assert.Error(t, err)
}

func TestApplyProject_NameDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, _ := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	// 未传 Name：取仓库名缺省（messager 的 slug 依赖它）。
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "repo-app", NeedGitRepo: false}, nil)
	jobMgr.EXPECT().NewJob(gomock.Any()).DoAndReturn(func(input *JobInput) Job {
		assert.Equal(t, "repo-app", input.Name)
		return job
	})
	expectInstallChain(job, nil)

	input := baseJobInput()
	input.Name = ""
	got, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, git, proj, jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{JobInput: input})
	assert.NoError(t, err)
	assert.Same(t, job, got)
	// 名缺省是 JobInput 原地变更，调用方读回可见。
	assert.Equal(t, "repo-app", input.Name)
}

// TestApplyProject_NameDefaultSyncsSlug 回归：创建项目场景前端不发 name，传输层构造
// messager 不绑定 slug；名缺省解析后 ApplyProject 必须把 messager 的 slug 回填为最终名
// （GetSlugName），与前端 toSlug 关联的日志 key 对齐——否则创建部署所有帧落错 key
// （前端日志区空）。
func TestApplyProject_NameDefaultSyncsSlug(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, _ := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "repo-app", NeedGitRepo: false}, nil)
	jobMgr.EXPECT().NewJob(gomock.Any()).Return(job)
	expectInstallChain(job, nil)

	msger := &slugRecorder{MockDeployMsger: NewMockDeployMsger(ctrl)}
	input := baseJobInput()
	input.Name = ""
	input.Messager = msger
	_, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, git, proj, jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{JobInput: input})
	assert.NoError(t, err)
	// name 就地解析为仓库名 + messager slug 以最终名重算，与前端 toSlug(namespaceId, repo.Name) 对齐。
	assert.Equal(t, "repo-app", input.Name)
	assert.Equal(t, GetSlugName(1, "repo-app"), msger.slug)
}

// slugRecorder 实现 DeployMsger 契约的 SetSlug，记录最新值；其余方法委托 mock。
type slugRecorder struct {
	*MockDeployMsger
	slug string
}

// SetSlug 记录调用参数。
func (s *slugRecorder) SetSlug(slug string) {
	s.slug = slug
}

func TestApplyProject_GitEnsureAndMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, msger := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "app", NeedGitRepo: true}, nil)
	git.EXPECT().EnsureBranchAndCommit(gomock.Any(), gomock.Any(), "", "").Return("main", "abc123", []string{"已解析分支: main"}, nil)
	msger.EXPECT().SetSlug(GetSlugName(1, "app"))
	msger.EXPECT().SendMsg("已解析分支: main")
	jobMgr.EXPECT().NewJob(gomock.Any()).DoAndReturn(func(input *JobInput) Job {
		assert.Equal(t, "main", input.GitBranch)
		assert.Equal(t, "abc123", input.GitCommit)
		return job
	})
	expectInstallChain(job, nil)

	input := baseJobInput()
	input.Messager = msger
	got, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, git, proj, jobMgr, msger), &ApplyProjectInput{JobInput: input})
	assert.NoError(t, err)
	assert.Same(t, job, got)
}

func TestApplyProject_GitEnsureError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, _, jobMgr, _ := applyTestKit(t, ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "app", NeedGitRepo: true}, nil)
	git.EXPECT().EnsureBranchAndCommit(gomock.Any(), gomock.Any(), "", "").Return("", "", nil, errors.New("git boom"))

	job, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, git, biz.NewMockProjectBiz(ctrl), jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{JobInput: baseJobInput()})
	assert.Nil(t, job)
	assert.Error(t, err)
}

func TestApplyProject_VersionReverseLookup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, _ := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "app", NeedGitRepo: false}, nil)
	// ProjectID==0 且 Version>0：按 name+ns 反查项目并回填 ProjectID。
	proj.EXPECT().FindByName(gomock.Any(), "app", 1).Return(&biz.Project{ID: 7}, nil)
	jobMgr.EXPECT().NewJob(gomock.Any()).DoAndReturn(func(input *JobInput) Job {
		assert.Equal(t, int32(7), input.ProjectID)
		return job
	})
	expectInstallChain(job, nil)

	input := baseJobInput()
	v := int32(5)
	input.Version = &v
	got, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, git, proj, jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{JobInput: input})
	assert.NoError(t, err)
	assert.Same(t, job, got)
	assert.Equal(t, int32(7), input.ProjectID)
}

func TestApplyProject_VersionReverseLookupNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, _ := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "app", NeedGitRepo: false}, nil)
	// FindByName 返回 biz NotFound：视为首次部署，ProjectID 保持 0，不阻断。
	proj.EXPECT().FindByName(gomock.Any(), "app", 1).Return(nil, errs.WrapNotFound(errors.New("not found"), "not found"))
	jobMgr.EXPECT().NewJob(gomock.Any()).DoAndReturn(func(input *JobInput) Job {
		assert.Equal(t, int32(0), input.ProjectID)
		return job
	})
	expectInstallChain(job, nil)

	input := baseJobInput()
	v := int32(5)
	input.Version = &v
	got, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, git, proj, jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{JobInput: input})
	assert.NoError(t, err)
	assert.Same(t, job, got)
}

func TestApplyProject_VersionReverseLookupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, _ := applyTestKit(t, ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "app", NeedGitRepo: false}, nil)
	// 非 NotFound 的 DB 故障必须上抛，不能放行成 ProjectID=0 的"版本不匹配"伪装。
	proj.EXPECT().FindByName(gomock.Any(), "app", 1).Return(nil, errors.New("db down"))

	input := baseJobInput()
	v := int32(5)
	input.Version = &v
	job, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, git, proj, jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{JobInput: input})
	assert.Nil(t, job)
	assert.Error(t, err)
}

func TestApplyProject_VersionNoLookupWhenProjectIDSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, _ := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "app", NeedGitRepo: false}, nil)
	jobMgr.EXPECT().NewJob(gomock.Any()).DoAndReturn(func(input *JobInput) Job {
		assert.Equal(t, int32(7), input.ProjectID)
		return job
	})
	expectInstallChain(job, nil)

	// ProjectID 已显式传入：即使 Version>0 也不反查（防止 WS 显式 ProjectID 被清零的回归）。
	input := baseJobInput()
	input.ProjectID = 7
	v := int32(5)
	input.Version = &v
	got, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, git, proj, jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{JobInput: input})
	assert.NoError(t, err)
	assert.Same(t, job, got)
	assert.Equal(t, int32(7), input.ProjectID)
}

func TestApplyProject_OnJobShortCircuit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, msger := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "app", NeedGitRepo: false}, nil)
	jobMgr.EXPECT().NewJob(gomock.Any()).Return(job)

	// OnJob 返回非 nil：传输层已发失败帧，跳过 watcher 与 InstallProject（无链式调用预期）。
	got, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, git, proj, jobMgr, msger), &ApplyProjectInput{
		JobInput: baseJobInput(),
		OnJob: func(j Job) error {
			assert.Same(t, job, j)
			return errors.New("已发失败帧")
		},
	})
	assert.NoError(t, err)
	assert.Same(t, job, got)
}

func TestApplyProject_InstallSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, _ := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "app", NeedGitRepo: false}, nil)
	jobMgr.EXPECT().NewJob(gomock.Any()).Return(job)
	expectInstallChain(job, nil)

	got, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, git, proj, jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{JobInput: baseJobInput()})
	assert.NoError(t, err)
	assert.Same(t, job, got)
}

func TestApplyProject_InstallError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, _ := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "app", NeedGitRepo: false}, nil)
	jobMgr.EXPECT().NewJob(gomock.Any()).Return(job)
	installErr := errors.New("install boom")
	expectInstallChain(job, installErr)

	got, err := ApplyProject(context.Background(), applyDepsOf(nsRepo, repo, git, proj, jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{JobInput: baseJobInput()})
	assert.ErrorIs(t, err, installErr)
	// 流水线失败仍返回 job（调用方读 Manifests/Project 兜底）。
	assert.Same(t, job, got)
}

func TestApplyProject_CtxCancelStopsJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, _ := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "app", NeedGitRepo: false}, nil)
	jobMgr.EXPECT().NewJob(gomock.Any()).Return(job)

	// watcher 在 OnJob 里 cancel ctx → goroutine 选中 ctx.Done 调 job.Stop。
	// Run 阻塞到 Stop 已观测，消除 select 在 ctx.Done/close(ch) 同时就绪时的随机性。
	stopCh := make(chan struct{})
	job.EXPECT().Stop(gomock.Any()).DoAndReturn(func(error) { close(stopCh) })
	gomock.InOrder(
		job.EXPECT().GlobalLock().Return(job),
		job.EXPECT().Validate().Return(job),
		job.EXPECT().LoadConfigs().Return(job),
		job.EXPECT().Run(gomock.Any()).DoAndReturn(func(ctx context.Context) Job {
			select {
			case <-stopCh:
			case <-time.After(2 * time.Second):
			}
			return job
		}),
		job.EXPECT().Finish().Return(job),
		job.EXPECT().Error().Return(nil),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	got, err := ApplyProject(ctx, applyDepsOf(nsRepo, repo, git, proj, jobMgr, NewMockDeployMsger(ctrl)), &ApplyProjectInput{
		JobInput: baseJobInput(),
		OnJob: func(j Job) error {
			cancel()
			return nil
		},
	})
	assert.NoError(t, err)
	assert.Same(t, job, got)
}
