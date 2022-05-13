package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/internal/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars-client/v4/mars"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"

	"github.com/duc-cnzj/mars-client/v4/git"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestGitSvc_All(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	gits := mockGitServer(m, app)
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
	gits.EXPECT().AllProjects().Return([]contracts.ProjectInterface{p1, p2}, nil)

	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	db.AutoMigrate(&models.GitProject{})

	all, err := new(GitSvc).All(context.TODO(), &git.AllRequest{})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), all.Items[0].Id)
	assert.Equal(t, int64(2), all.Items[1].Id)
}

func TestGitSvc_BranchOptions(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	gits := mockGitServer(m, app)
	b1 := mock.NewMockBranchInterface(m)
	b1.EXPECT().GetName().Return("b1").MinTimes(2)
	b2 := mock.NewMockBranchInterface(m)
	b2.EXPECT().GetName().Return("b2").MinTimes(2)
	gits.EXPECT().AllBranches("1").Return([]contracts.BranchInterface{b1, b2}, nil)
	options, err := new(GitSvc).BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: "1",
		All:          true,
	})
	assert.Nil(t, err)
	assert.Len(t, options.Items, 2)

	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
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
	gits.EXPECT().AllBranches("10").Return([]contracts.BranchInterface{b1, b2}, nil)
	options, _ = new(GitSvc).BranchOptions(context.TODO(), &git.BranchOptionsRequest{
		GitProjectId: "10",
		All:          false,
	})
	assert.Equal(t, "b1", options.Items[0].Branch)
}

func TestGitSvc_Commit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	gits := mockGitServer(m, app)
	gits.EXPECT().GetCommit(gomock.Any(), gomock.Any()).Return(nil, errors.New("")).Times(1)
	_, err := new(GitSvc).Commit(context.TODO(), &git.CommitRequest{})
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
	res, _ := new(GitSvc).Commit(context.TODO(), &git.CommitRequest{
		GitProjectId: "11",
		Branch:       "dev",
		Commit:       "xxx",
	})
	assert.NotNil(t, res)
}

func TestGitSvc_CommitOptions(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
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
	options, err := new(GitSvc).CommitOptions(context.TODO(), &git.CommitOptionsRequest{
		GitProjectId: "",
		Branch:       "",
	})
	assert.Nil(t, err)
	assert.Len(t, options.Items, 2)
	gits.EXPECT().ListCommits(gomock.Any(), gomock.Any()).Return(nil, errors.New("")).Times(1)
	_, err = new(GitSvc).CommitOptions(context.TODO(), &git.CommitOptionsRequest{
		GitProjectId: "",
		Branch:       "",
	})
	assert.Error(t, err)
}

func TestGitSvc_DisableProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)

	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	db.AutoMigrate(&models.GitProject{})

	gits := mockGitServer(m, app)
	p := mock.NewMockProjectInterface(m)
	gits.EXPECT().GetProject("123").Return(p, nil)

	p.EXPECT().GetName().Return("n").Times(2)
	p.EXPECT().GetDefaultBranch().Return("dex")

	assertAuditLogFired(m, app)

	_, err := new(GitSvc).DisableProject(adminCtx(), &git.DisableProjectRequest{
		GitProjectId: "123",
	})
	assert.Nil(t, err)
	_, err = new(GitSvc).DisableProject(auth.SetUser(context.TODO(), &contracts.UserInfo{}), &git.DisableProjectRequest{
		GitProjectId: "123",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, fromError.Code())
}

func TestGitSvc_DisableProject2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)

	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
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

	assertAuditLogFired(m, app)

	_, err := new(GitSvc).DisableProject(adminCtx(), &git.DisableProjectRequest{
		GitProjectId: "123",
	})
	assert.Nil(t, err)
	db.First(pmodel)
	assert.False(t, pmodel.Enabled)
}

func TestGitSvc_EnableProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)

	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	db.AutoMigrate(&models.GitProject{})

	gits := mockGitServer(m, app)
	p := mock.NewMockProjectInterface(m)
	gits.EXPECT().GetProject("123").Return(p, nil)

	p.EXPECT().GetName().Return("n").Times(2)
	p.EXPECT().GetDefaultBranch().Return("dex")

	assertAuditLogFired(m, app)

	_, err := new(GitSvc).DisableProject(adminCtx(), &git.DisableProjectRequest{
		GitProjectId: "123",
	})
	assert.Nil(t, err)
	_, err = new(GitSvc).EnableProject(auth.SetUser(context.TODO(), &contracts.UserInfo{}), &git.EnableProjectRequest{
		GitProjectId: "123",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, fromError.Code())
}
func TestGitSvc_EnableProject2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)

	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
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

	assertAuditLogFired(m, app)

	_, err := new(GitSvc).EnableProject(adminCtx(), &git.EnableProjectRequest{
		GitProjectId: "123",
	})
	assert.Nil(t, err)
	db.First(pmodel)
	assert.True(t, pmodel.Enabled)
}

func TestGitSvc_MarsConfigFile(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	marsC := mars.Config{
		LocalChartPath: "",
		Branches:       []string{"b1"},
	}
	marshal, _ := json.Marshal(&marsC)
	marsC2 := mars.Config{
		LocalChartPath: "",
		Branches:       []string{"b1"},
		ConfigFile:     "cfg.yaml",
	}
	marshal2, _ := json.Marshal(&marsC2)
	db.AutoMigrate(&models.GitProject{})
	db.Create(&models.GitProject{
		DefaultBranch: "dev",
		Name:          "app",
		GitProjectId:  10,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})
	db.Create(&models.GitProject{
		DefaultBranch: "dev2",
		Name:          "app2",
		GitProjectId:  11,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal2),
	})
	file, err := new(GitSvc).MarsConfigFile(context.TODO(), &git.MarsConfigFileRequest{
		GitProjectId: "10",
		Branch:       "dev",
	})
	assert.Nil(t, err)
	assert.Equal(t, "yaml", file.Type)
	assert.Equal(t, "", file.Data)
	gits := mockGitServer(m, app)
	gits.EXPECT().GetFileContentWithBranch("11", "dev", "cfg.yaml").Return("", errors.New("aaa")).Times(1)
	file, _ = new(GitSvc).MarsConfigFile(context.TODO(), &git.MarsConfigFileRequest{
		GitProjectId: "11",
		Branch:       "dev",
	})
	assert.Equal(t, "", file.Data)
	gits.EXPECT().GetFileContentWithBranch("11", "dev", "cfg.yaml").Return("aaa", nil).Times(1)
	file, _ = new(GitSvc).MarsConfigFile(context.TODO(), &git.MarsConfigFileRequest{
		GitProjectId: "11",
		Branch:       "dev",
	})
	assert.Equal(t, "aaa", file.Data)
}

func TestGitSvc_PipelineInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	pipe := mock.NewMockPipelineInterface(m)
	instance.SetInstance(app)
	gitS := mockGitServer(m, app)
	gitS.EXPECT().GetCommitPipeline("1", "xxx").Times(1).Return(pipe, nil)
	pipe.EXPECT().GetStatus().Times(1).Return("status")
	pipe.EXPECT().GetWebURL().Times(1).Return("weburl")
	info, _ := new(GitSvc).PipelineInfo(context.TODO(), &git.PipelineInfoRequest{
		GitProjectId: "1",
		Branch:       "dev",
		Commit:       "xxx",
	})
	assert.Equal(t, "status", info.Status)
	assert.Equal(t, "weburl", info.WebUrl)
}

func TestGitSvc_ProjectOptions(t *testing.T) {
}

func mockGitServer(m *gomock.Controller, app *mock.MockApplicationInterface) *mock.MockGitServer {
	gitS := mock.NewMockGitServer(m)
	app.EXPECT().Config().Return(&config.Config{
		GitServerPlugin: config.Plugin{
			Name: "test_git_server",
		},
	}).AnyTimes()
	app.EXPECT().GetPluginByName("test_git_server").Return(gitS).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gitS.EXPECT().Initialize(gomock.All()).AnyTimes()
	return gitS
}

func mockWsServer(m *gomock.Controller, app *mock.MockApplicationInterface) *mock.MockWsSender {
	ws := mock.NewMockWsSender(m)
	app.EXPECT().Config().Return(&config.Config{
		WsSenderPlugin: config.Plugin{
			Name: "test_ws",
		},
	}).AnyTimes()
	app.EXPECT().GetPluginByName("test_ws").Return(ws).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	ws.EXPECT().Initialize(gomock.All()).AnyTimes()
	return ws
}
