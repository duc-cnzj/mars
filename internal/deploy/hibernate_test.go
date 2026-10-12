package deploy

import (
	"context"
	"errors"
	"testing"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// hibernateProj 返回一份已部署项目的休眠重部署入参。
func hibernateProj() *biz.Project {
	return &biz.Project{
		ID:          9,
		Name:        "app",
		NamespaceID: 1,
		RepoID:      2,
		GitBranch:   "main",
		GitCommit:   "abc",
		Config:      "c",
		Atomic:      true,
		Version:     3,
		FinalExtraValues: []*websocket_pb.ExtraValue{
			{Path: "app->replicas", Value: "3"},
		},
	}
}

// TestHibernateDeployer_Redeploy 覆盖系统身份重部署：以 UpdateProject 走共享编排，
// 项目身份/版本/额外配置项原样透传到 JobInput，返回安装流水线错误。
func TestHibernateDeployer_Redeploy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, _ := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "repo-app", NeedGitRepo: false}, nil)
	jobMgr.EXPECT().NewJob(gomock.Any()).DoAndReturn(func(input *JobInput) Job {
		assert.Equal(t, websocket_pb.Type_UpdateProject, input.Type)
		assert.Equal(t, int32(9), input.ProjectID)
		assert.Equal(t, int32(2), input.RepoID)
		assert.Equal(t, "app", input.Name)
		assert.Equal(t, int32(3), *input.Version)
		assert.Same(t, hibernateUser, input.User)
		assert.Len(t, input.ExtraValues, 1)
		return job
	})
	expectInstallChain(job, nil)

	d := NewHibernateDeployer(applyDepsOf(nsRepo, repo, git, proj, jobMgr, NewMockDeployMsger(ctrl)))
	err := d.Redeploy(context.Background(), hibernateProj(), []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "0"}})
	assert.NoError(t, err)
}

// TestHibernateDeployer_Redeploy_PipelineError 覆盖流水线失败时错误上抛。
func TestHibernateDeployer_Redeploy_PipelineError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsRepo, repo, git, proj, jobMgr, _ := applyTestKit(t, ctrl)
	job := NewMockJob(ctrl)

	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(pubNs(), nil)
	repo.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, Name: "repo-app", NeedGitRepo: false}, nil)
	jobMgr.EXPECT().NewJob(gomock.Any()).Return(job)
	expectInstallChain(job, errors.New("install boom"))

	d := NewHibernateDeployer(applyDepsOf(nsRepo, repo, git, proj, jobMgr, NewMockDeployMsger(ctrl)))
	assert.Error(t, d.Redeploy(context.Background(), hibernateProj(), nil))
}

// TestNoopMsger 覆盖空进度端口的全部方法（丢弃输入、不 panic）。
func TestNoopMsger(t *testing.T) {
	var m DeployMsger = NoopMsger{}
	assert.Zero(t, m.Current())
	m.Add()
	m.To(50)
	m.SetSlug("slug")
	m.SendProcessPercent(10)
	m.SendDeployedResult(websocket_pb.ResultType_Success, "ok", nil)
	m.SendEndError(errors.New("x"))
	m.SendMsg("msg")
	m.SendProtoMsg(nil)
	m.SendMsgWithContainerLog("log", nil)
}
