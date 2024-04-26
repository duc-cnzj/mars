package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/duc-cnzj/mars/api/v4/git"
	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGitSvc_All(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	gits := mockGitServer(m, app)
	// p1 id: 1 enabled: false
	// p2 id: 2 enabled: true
	// p3 id: 3 enabled: false
	// p4 id: 4 enabled: true
	// ->
	// p2, p4, p1, p3
	p1 := mock.NewMockProjectInterface(m)
	p1.EXPECT().GetID().Return(int64(1)).Times(2)
	p1.EXPECT().GetName().Return("name1")
	p1.EXPECT().GetPath().Return("path1")
	p1.EXPECT().GetWebURL().Return("weburl1")
	p1.EXPECT().GetAvatarURL().Return("avatar1")
	p1.EXPECT().GetDescription().Return("desc1")
	p2 := mock.NewMockProjectInterface(m)
	p2.EXPECT().GetID().Return(int64(2)).Times(2)
	p2.EXPECT().GetName().Return("name2")
	p2.EXPECT().GetPath().Return("path2")
	p2.EXPECT().GetWebURL().Return("weburl2")
	p2.EXPECT().GetAvatarURL().Return("avatar2")
	p2.EXPECT().GetDescription().Return("desc2")
	p3 := mock.NewMockProjectInterface(m)
	p3.EXPECT().GetID().Return(int64(3)).Times(2)
	p3.EXPECT().GetName().Return("name3")
	p3.EXPECT().GetPath().Return("path3")
	p3.EXPECT().GetWebURL().Return("weburl3")
	p3.EXPECT().GetAvatarURL().Return("avatar3")
	p3.EXPECT().GetDescription().Return("desc3")
	p4 := mock.NewMockProjectInterface(m)
	p4.EXPECT().GetID().Return(int64(4)).Times(2)
	p4.EXPECT().GetName().Return("name4")
	p4.EXPECT().GetPath().Return("path4")
	p4.EXPECT().GetWebURL().Return("weburl4")
	p4.EXPECT().GetAvatarURL().Return("avatar4")
	p4.EXPECT().GetDescription().Return("desc4")
	gits.EXPECT().AllProjects().Return([]contracts.ProjectInterface{p3, p1, p2, p4}, nil).Times(1)

	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	db.AutoMigrate(&models.GitProject{})
	marshal, _ := json.Marshal(&mars.Config{DisplayName: "app"})
	db.Create(&models.GitProject{
		GitProjectId:  2,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})
	db.Create(&models.GitProject{
		GitProjectId:  4,
		Enabled:       true,
		GlobalEnabled: true,
	})

	all, err := new(gitSvc).All(context.TODO(), &git.AllRequest{})
	assert.Nil(t, err)
	assert.Equal(t, int64(2), all.Items[0].Id)
	assert.Equal(t, true, all.Items[0].Enabled)
	assert.Equal(t, true, all.Items[0].GlobalEnabled)
	assert.Equal(t, "app", all.Items[0].DisplayName)
	assert.Equal(t, int64(4), all.Items[1].Id)
	assert.Equal(t, true, all.Items[1].Enabled)
	assert.Equal(t, true, all.Items[1].GlobalEnabled)
	assert.Equal(t, "", all.Items[1].DisplayName)
	assert.Equal(t, int64(1), all.Items[2].Id)
	assert.Equal(t, false, all.Items[2].Enabled)
	assert.Equal(t, false, all.Items[2].GlobalEnabled)
	assert.Equal(t, "", all.Items[2].DisplayName)
	assert.Equal(t, int64(3), all.Items[3].Id)
	assert.Equal(t, false, all.Items[3].Enabled)
	assert.Equal(t, false, all.Items[3].GlobalEnabled)
	assert.Equal(t, "", all.Items[3].DisplayName)

	gits.EXPECT().AllProjects().Return(nil, errors.New("xxx")).Times(1)
	_, err = new(gitSvc).All(context.TODO(), &git.AllRequest{})
	assert.Equal(t, "xxx", err.Error())
}

func TestGitSvc_BranchOptions(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	gits := mockGitServer(m, app)
	b1 := mock.NewMockBranchInterface(m)
	b1.EXPECT().GetName().Return("b1").MinTimes(2)
	b2 := mock.NewMockBranchInterface(m)
	b2.EXPECT().GetName().Return("b2").MinTimes(2)
	gits.EXPECT().AllBranches("1").Return([]contracts.BranchInterface{b1, b2}, nil).Times(1)
	options, err := new(gitSvc).BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: "1",
		All:          true,
	})
	assert.Nil(t, err)
	assert.Len(t, options.Items, 2)

	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	marsC := mars.Config{
		LocalChartPath: "",
		Branches:       []string{"b1"},
	}
	marshal, _ := json.Marshal(&marsC)
	db.AutoMigrate(&models.GitProject{})
	db.Create(&models.GitProject{
		DefaultBranch: "dev",
		Name:          "app",
		GitProjectId:  10,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})
	b1.EXPECT().IsDefault().Return(true).Times(1)
	b2.EXPECT().IsDefault().Return(false).Times(1)
	gits.EXPECT().AllBranches("10").Return(nil, errors.New("xxx")).Times(1)
	_, err = new(gitSvc).BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: "10",
		All:          true,
	})
	assert.Equal(t, "xxx", err.Error())
	assert.Error(t, err)
	gits.EXPECT().AllBranches("10").Return([]contracts.BranchInterface{b1, b2}, nil)
	options, _ = new(gitSvc).BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: "10",
		All:          false,
	})
	assert.Equal(t, "b1", options.Items[0].Branch)

	gits.EXPECT().AllBranches("100").Return([]contracts.BranchInterface{b1}, nil).Times(1)
	b1.EXPECT().IsDefault().Return(false)
	gits.EXPECT().GetFileContentWithBranch("100", "", ".mars.yaml").Return("", errors.New("xx")).Times(1)
	res, err := new(gitSvc).BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: "100",
		All:          false,
	})
	assert.Nil(t, err)
	assert.Len(t, res.Items, 0)
}

func TestGitSvc_Commit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	gits := mockGitServer(m, app)
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(nil, errors.New("")).Times(1)
	_, err := new(gitSvc).Commit(context.TODO(), &git.CommitRequest{})
	assert.Error(t, err)
	cm := mock.NewMockCommitInterface(m)
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(cm, nil).Times(1)
	cm.EXPECT().GetID().Return("aaa").Times(1)
	cm.EXPECT().GetShortID().Return("xxx").Times(1)
	cm.EXPECT().GetCommittedDate().Return(nil).Times(1)
	cm.EXPECT().GetTitle().Return("t").Times(2)
	cm.EXPECT().GetAuthorName().Return("duc").Times(1)
	cm.EXPECT().GetAuthorEmail().Return("1@q.c").Times(1)
	cm.EXPECT().GetCommitterName().Return("cname").Times(1)
	cm.EXPECT().GetCommitterEmail().Return("cemail").Times(1)
	cm.EXPECT().GetWebURL().Return("weburl").Times(1)
	cm.EXPECT().GetMessage().Return("msg").Times(1)
	cm.EXPECT().GetCommittedDate().Return(nil).Times(1)
	cm.EXPECT().GetCreatedAt().Return(nil).Times(1)
	res, _ := new(gitSvc).Commit(context.TODO(), &git.CommitRequest{
		GitProjectId: "11",
		Branch:       "dev",
		Commit:       "xxx",
	})
	assert.NotNil(t, res)
}

func TestGitSvc_CommitOptions(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	gits := mockGitServer(m, app)
	m1 := mock.NewMockCommitInterface(m)
	m1.EXPECT().GetID().Return("xx")
	m1.EXPECT().GetTitle().Return("tt")
	m1.EXPECT().GetCommittedDate().Return(nil)
	m2 := mock.NewMockCommitInterface(m)
	m2.EXPECT().GetID().Return("yyy")
	m2.EXPECT().GetTitle().Return("zzz")
	m2.EXPECT().GetCommittedDate().Return(nil)
	gits.EXPECT().ListCommits(gomock.Any(), gomock.Any()).Return([]contracts.CommitInterface{m1, m2}, nil).Times(1)
	options, err := new(gitSvc).CommitOptions(context.TODO(), &git.CommitOptionsRequest{
		GitProjectId: "",
		Branch:       "",
	})
	assert.Nil(t, err)
	assert.Len(t, options.Items, 2)
	gits.EXPECT().ListCommits(gomock.Any(), gomock.Any()).Return(nil, errors.New("")).Times(1)
	_, err = new(gitSvc).CommitOptions(context.TODO(), &git.CommitOptionsRequest{
		GitProjectId: "",
		Branch:       "",
	})
	assert.Error(t, err)
}

func TestGitSvc_DisableProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	db.AutoMigrate(&models.GitProject{})

	gits := mockGitServer(m, app)
	p := mock.NewMockProjectInterface(m)
	gits.EXPECT().GetProject("123").Return(p, nil)

	p.EXPECT().GetName().Return("n").Times(2)
	p.EXPECT().GetDefaultBranch().Return("dex")

	testutil.AssertAuditLogFired(m, app)

	_, err := new(gitSvc).DisableProject(adminCtx(), &git.DisableProjectRequest{
		GitProjectId: "123",
	})
	assert.Nil(t, err)
	_, err = new(gitSvc).DisableProject(auth.SetUser(context.TODO(), &contracts.UserInfo{}), &git.DisableProjectRequest{
		GitProjectId: "123",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, fromError.Code())
}

func TestGitSvc_DisableProject2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	db.AutoMigrate(&models.GitProject{})
	pmodel := &models.GitProject{
		GitProjectId: 123,
		Enabled:      true,
	}
	db.Create(pmodel)
	gits := mockGitServer(m, app)
	p := mock.NewMockProjectInterface(m)
	gits.EXPECT().GetProject("123").Return(p, nil)

	p.EXPECT().GetName().Return("n").Times(2)
	p.EXPECT().GetDefaultBranch().Return("dex")

	testutil.AssertAuditLogFired(m, app)

	_, err := new(gitSvc).DisableProject(adminCtx(), &git.DisableProjectRequest{
		GitProjectId: "123",
	})
	assert.Nil(t, err)
	db.First(pmodel)
	assert.False(t, pmodel.Enabled)
}

func TestGitSvc_EnableProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	db.AutoMigrate(&models.GitProject{})

	gits := mockGitServer(m, app)
	p := mock.NewMockProjectInterface(m)
	gits.EXPECT().GetProject("123").Return(p, nil)

	p.EXPECT().GetName().Return("n").Times(2)
	p.EXPECT().GetDefaultBranch().Return("dex")

	testutil.AssertAuditLogFiredWithMsg(m, app, "关闭项目: n")

	_, err := new(gitSvc).DisableProject(adminCtx(), &git.DisableProjectRequest{
		GitProjectId: "123",
	})
	assert.Nil(t, err)
	_, err = new(gitSvc).EnableProject(auth.SetUser(context.TODO(), &contracts.UserInfo{}), &git.EnableProjectRequest{
		GitProjectId: "123",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, fromError.Code())
}

func TestGitSvc_EnableProject2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	db.AutoMigrate(&models.GitProject{})
	pmodel := &models.GitProject{
		GitProjectId: 123,
		Enabled:      false,
	}
	db.Create(pmodel)
	gits := mockGitServer(m, app)
	p := mock.NewMockProjectInterface(m)
	gits.EXPECT().GetProject("123").Return(p, nil)

	p.EXPECT().GetName().Return("n").Times(2)
	p.EXPECT().GetDefaultBranch().Return("dex")

	testutil.AssertAuditLogFired(m, app)

	_, err := new(gitSvc).EnableProject(adminCtx(), &git.EnableProjectRequest{
		GitProjectId: "123",
	})
	assert.Nil(t, err)
	db.First(pmodel)
	assert.True(t, pmodel.Enabled)
}

func TestGitSvc_EnableProject_NotExistsInDB(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	db.AutoMigrate(&models.GitProject{})
	gits := mockGitServer(m, app)
	p := mock.NewMockProjectInterface(m)
	gits.EXPECT().GetProject("123").Return(p, nil)

	p.EXPECT().GetName().Return("n").Times(2)
	p.EXPECT().GetDefaultBranch().Return("dex")

	testutil.AssertAuditLogFired(m, app)

	_, err := new(gitSvc).EnableProject(adminCtx(), &git.EnableProjectRequest{
		GitProjectId: "123",
	})
	assert.Nil(t, err)
	pp := &models.GitProject{}
	db.First(pp)
	assert.Equal(t, "n", pp.Name)
	assert.Equal(t, "dex", pp.DefaultBranch)
	assert.Equal(t, 123, pp.GitProjectId)
	assert.Equal(t, true, pp.Enabled)
}

func TestGitSvc_MarsConfigFile(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	db.AutoMigrate(&models.GitProject{})

	marsC := mars.Config{
		LocalChartPath: "",
		Branches:       []string{"b1"},
	}
	marshal, _ := json.Marshal(&marsC)
	db.Create(&models.GitProject{
		DefaultBranch: "dev",
		Name:          "app",
		GitProjectId:  10,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})
	marsC2 := mars.Config{
		LocalChartPath: "",
		Branches:       []string{"b1"},
		ConfigFile:     "cfg.yaml",
	}

	marshal2, _ := json.Marshal(&marsC2)
	db.Create(&models.GitProject{
		DefaultBranch: "dev2",
		Name:          "app2",
		GitProjectId:  11,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal2),
	})
	marsC3 := mars.Config{
		LocalChartPath: "",
		Branches:       []string{"b1"},
		ConfigFile:     "1|master|cfg.yaml",
	}
	marshal3, _ := json.Marshal(&marsC3)
	db.Create(&models.GitProject{
		DefaultBranch: "dev3",
		Name:          "app3",
		GitProjectId:  12,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal3),
	})
	file, err := new(gitSvc).MarsConfigFile(context.TODO(), &git.MarsConfigFileRequest{
		GitProjectId: "10",
		Branch:       "dev",
	})
	assert.Nil(t, err)
	assert.Equal(t, "yaml", file.Type)
	assert.Equal(t, "", file.Data)
	gits := mockGitServer(m, app)
	gits.EXPECT().GetFileContentWithBranch("11", "dev", "cfg.yaml").Return("", errors.New("aaa")).Times(1)
	file, _ = new(gitSvc).MarsConfigFile(context.TODO(), &git.MarsConfigFileRequest{
		GitProjectId: "11",
		Branch:       "dev",
	})
	assert.Equal(t, "", file.Data)
	gits.EXPECT().GetFileContentWithBranch("11", "dev", "cfg.yaml").Return("aaa", nil).Times(1)
	file, _ = new(gitSvc).MarsConfigFile(context.TODO(), &git.MarsConfigFileRequest{
		GitProjectId: "11",
		Branch:       "dev",
	})
	assert.Equal(t, "aaa", file.Data)
	gits.EXPECT().GetFileContentWithBranch("1", "master", "cfg.yaml").Return("aaa", nil).Times(1)

	new(gitSvc).MarsConfigFile(context.TODO(), &git.MarsConfigFileRequest{
		GitProjectId: "12",
		Branch:       "dev",
	})

	gits.EXPECT().GetFileContentWithBranch("9999", "dev", ".mars.yaml").Return("", errors.New("aaa")).Times(1)
	_, err = new(gitSvc).MarsConfigFile(context.TODO(), &git.MarsConfigFileRequest{
		GitProjectId: "9999",
		Branch:       "dev",
	})
	assert.Equal(t, "aaa", err.Error())
}

func TestGitSvc_PipelineInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	pipe := mock.NewMockPipelineInterface(m)
	instance.SetInstance(app)
	gitS := mockGitServer(m, app)
	gitS.EXPECT().GetCommitPipeline("1", "dev", "xxx").Times(1).Return(pipe, nil)
	pipe.EXPECT().GetStatus().Times(1).Return("status")
	pipe.EXPECT().GetWebURL().Times(1).Return("weburl")
	info, _ := new(gitSvc).PipelineInfo(context.TODO(), &git.PipelineInfoRequest{
		GitProjectId: "1",
		Branch:       "dev",
		Commit:       "xxx",
	})
	assert.Equal(t, "status", info.Status)
	assert.Equal(t, "weburl", info.WebUrl)
	gitS.EXPECT().GetCommitPipeline("1", "dev", "xxx").Times(1).Return(nil, errors.New("xxx"))
	_, err := new(gitSvc).PipelineInfo(context.TODO(), &git.PipelineInfoRequest{
		GitProjectId: "1",
		Branch:       "dev",
		Commit:       "xxx",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())
	assert.Equal(t, "xxx", fromError.Message())
}

func TestGitSvc_ProjectOptions(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	app.EXPECT().Cache().Return(c)
	c.EXPECT().Remember(cache.NewKey("ProjectOptions"), 30, gomock.Any()).Return(nil, errors.New("xxx"))
	_, err := new(gitSvc).ProjectOptions(context.TODO(), &git.ProjectOptionsRequest{})
	assert.Equal(t, "xxx", err.Error())

	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.GitProject{})
	marshal, _ := json.Marshal(&mars.Config{DisplayName: "app"})
	marshal2, _ := json.Marshal(&mars.Config{DisplayName: "b"})

	p1 := &models.GitProject{
		DefaultBranch: "dev",
		Name:          "a",
		GitProjectId:  1,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	}
	p2 := &models.GitProject{
		DefaultBranch: "dev1",
		Name:          "b",
		GitProjectId:  2,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal2),
	}
	p3 := &models.GitProject{
		DefaultBranch: "dev2",
		Name:          "c",
		GitProjectId:  3,
		Enabled:       true,
		GlobalEnabled: false,
		GlobalConfig:  "",
	}
	server := mockGitServer(m, app)
	server.EXPECT().GetFileContentWithBranch("3", "dev2", ".mars.yaml").Return("", errors.New("xxx"))
	db.CreateInBatches([]*models.GitProject{p1, p2, p3}, 2)
	app.EXPECT().Cache().Return(&cache.NoCache{}).Times(1)
	res, err := new(gitSvc).ProjectOptions(context.TODO(), &git.ProjectOptionsRequest{})
	assert.Nil(t, err)
	assert.Len(t, res.Items, 2)
	assert.Equal(t, "a(app)", res.Items[0].Label)
	assert.Equal(t, "", res.Items[0].Branch)
	assert.Equal(t, OptionTypeProject, res.Items[0].Type)
	assert.Equal(t, fmt.Sprintf("%d", p1.ID), res.Items[0].Value)
	assert.Equal(t, false, res.Items[0].IsLeaf)
	assert.Equal(t, "1", res.Items[0].GitProjectId)
	assert.Equal(t, "app", res.Items[0].DisplayName)

	assert.Equal(t, "", res.Items[1].Branch)
	assert.Equal(t, "b", res.Items[1].Label)
	assert.Equal(t, OptionTypeProject, res.Items[1].Type)
	assert.Equal(t, fmt.Sprintf("%d", p2.ID), res.Items[1].Value)
	assert.Equal(t, false, res.Items[1].IsLeaf)
	assert.Equal(t, "2", res.Items[1].GitProjectId)
	assert.Equal(t, "b", res.Items[1].DisplayName)
}

func mockGitServer(m *gomock.Controller, app *mock.MockApplicationInterface) *mock.MockGitServer {
	return testutil.MockGitServer(m, app)
}

func TestSortableOption(t *testing.T) {
	var options = []*git.Option{
		{
			Value:        "2",
			Label:        "2",
			Type:         "2",
			IsLeaf:       false,
			GitProjectId: "2",
			Branch:       "dev2",
		},
		{
			Value:        "1",
			Label:        "1",
			Type:         "1",
			IsLeaf:       false,
			GitProjectId: "1",
			Branch:       "dev1",
		},
	}
	sort.Sort(sortableOption(options))
	assert.Equal(t, []*git.Option{
		{
			Value:        "1",
			Label:        "1",
			Type:         "1",
			IsLeaf:       false,
			GitProjectId: "1",
			Branch:       "dev1",
		},
		{
			Value:        "2",
			Label:        "2",
			Type:         "2",
			IsLeaf:       false,
			GitProjectId: "2",
			Branch:       "dev2",
		},
	}, options)
}
