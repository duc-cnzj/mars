package plugins

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	cache2 "github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type gitPlugin struct {
	called bool
	GitServer
	err error
}

func (g *gitPlugin) Initialize(args map[string]any) error {
	g.called = true
	return g.err
}

func TestGetGitServer(t *testing.T) {
	p := &gitPlugin{}
	ma := &mockApp{
		p: map[string]contracts.PluginInterface{"git": p},
	}
	instance.SetInstance(ma)
	gitServerOnce = sync.Once{}
	GetGitServer()
	assert.Equal(t, 1, ma.callback)
	assert.True(t, p.called)
}

func TestGetGitServer2(t *testing.T) {
	p := &gitPlugin{
		err: errors.New("xxx"),
	}
	ma := &mockApp{
		p: map[string]contracts.PluginInterface{"git": p},
	}
	instance.SetInstance(ma)
	gitServerOnce = sync.Once{}
	assert.Panics(t, func() {
		GetGitServer()
	})
	assert.Equal(t, 0, ma.callback)
	assert.True(t, p.called)
}

func TestGetGitServerCached(t *testing.T) {
	p := &gitPlugin{}
	ma := &mockApp{
		cached: true,
		p:      map[string]contracts.PluginInterface{"git": p},
	}
	instance.SetInstance(ma)
	gitServerOnce = sync.Once{}
	assert.IsType(t, (*gitServerCache)(nil), GetGitServer())
	assert.Equal(t, 1, ma.callback)
	assert.True(t, p.called)
}

func Test_branch_GetName(t *testing.T) {
	b := branch{
		Name: "name",
	}
	assert.Equal(t, "name", b.GetName())
}

func Test_branch_GetWebURL(t *testing.T) {
	b := branch{
		WebUrl: "https://xxx",
	}
	assert.Equal(t, b.WebUrl, b.GetWebURL())
}

func Test_branch_IsDefault(t *testing.T) {
	b := branch{
		Default: true,
	}
	assert.Equal(t, b.Default, b.IsDefault())
}

func Test_commit_GetAuthorEmail(t *testing.T) {
	c := commit{
		AuthorEmail: "1025434218@qq.com",
	}
	assert.Equal(t, c.AuthorEmail, c.GetAuthorEmail())
}

func Test_commit_GetAuthorName(t *testing.T) {
	c := commit{
		AuthorName: "duc",
	}
	assert.Equal(t, c.AuthorName, c.GetAuthorName())
}

func Test_commit_GetCommittedDate(t *testing.T) {
	c := commit{
		CommittedDate: &time.Time{},
	}
	assert.Equal(t, c.CommittedDate, c.GetCommittedDate())
}

func Test_commit_GetCommitterEmail(t *testing.T) {
	c := commit{
		CommitterEmail: "1@q.c",
	}
	assert.Equal(t, c.CommitterEmail, c.GetCommitterEmail())
}

func Test_commit_GetCommitterName(t *testing.T) {
	c := commit{
		CommitterName: "name",
	}
	assert.Equal(t, c.CommitterName, c.GetCommitterName())
}

func Test_commit_GetCreatedAt(t *testing.T) {
	c := commit{
		CreatedAt: &time.Time{},
	}
	assert.Equal(t, c.CreatedAt, c.GetCreatedAt())
}

func Test_commit_GetID(t *testing.T) {
	c := commit{
		ID: "1",
	}
	assert.Equal(t, c.ID, c.GetID())
}

func Test_commit_GetMessage(t *testing.T) {
	c := commit{
		Message: "msg",
	}
	assert.Equal(t, c.Message, c.GetMessage())
}

func Test_commit_GetProjectID(t *testing.T) {
	c := commit{
		ProjectID: 1,
	}
	assert.Equal(t, c.ProjectID, c.GetProjectID())
}

func Test_commit_GetShortID(t *testing.T) {
	c := commit{
		ShortID: "1",
	}
	assert.Equal(t, c.ShortID, c.GetShortID())
}

func Test_commit_GetTitle(t *testing.T) {
	c := commit{
		Title: "t",
	}
	assert.Equal(t, c.Title, c.GetTitle())
}

func Test_commit_GetWebURL(t *testing.T) {
	c := commit{
		WebURL: "weburl",
	}
	assert.Equal(t, c.WebURL, c.GetWebURL())
}

func Test_project_GetAvatarURL(t *testing.T) {
	p := project{
		AvatarUrl: "a",
	}
	assert.Equal(t, p.AvatarUrl, p.GetAvatarURL())
}

func Test_project_GetDefaultBranch(t *testing.T) {
	p := project{
		DefaultBranch: "dev",
	}
	assert.Equal(t, p.DefaultBranch, p.GetDefaultBranch())
}

func Test_project_GetDescription(t *testing.T) {
	p := project{
		Description: "desc",
	}
	assert.Equal(t, p.Description, p.GetDescription())
}

func Test_project_GetID(t *testing.T) {
	p := project{
		ID: 1,
	}
	assert.Equal(t, p.ID, p.GetID())
}

func Test_project_GetName(t *testing.T) {
	p := project{
		Name: "name",
	}
	assert.Equal(t, p.Name, p.GetName())
}

func Test_project_GetPath(t *testing.T) {
	p := project{
		Path: "path",
	}
	assert.Equal(t, p.Path, p.GetPath())
}

func Test_project_GetWebURL(t *testing.T) {
	p := project{
		WebUrl: "weburl",
	}
	assert.Equal(t, p.WebUrl, p.GetWebURL())
}

func Test_gitServerCache_AllBranches(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	gits := mockGitServer(m, app)

	app.EXPECT().Cache().Return(c)
	c.EXPECT().Remember(cache2.NewKey("AllBranches-%v", 1), AllBranchesCacheSeconds, gomock.Any()).Return(nil, errors.New("xxx"))
	_, err := (&gitServerCache{s: gits}).AllBranches("1")
	assert.Equal(t, "xxx", err.Error())

	app.EXPECT().Cache().Return(&cache2.NoCache{})
	b1 := mock.NewMockBranchInterface(m)
	b1.EXPECT().GetWebURL().Return("weburl")
	b1.EXPECT().GetName().Return("name")
	b1.EXPECT().IsDefault().Return(false)
	b2 := mock.NewMockBranchInterface(m)
	b2.EXPECT().GetWebURL().Return("weburl2")
	b2.EXPECT().GetName().Return("name2")
	b2.EXPECT().IsDefault().Return(true)
	gits.EXPECT().AllBranches("1").Return([]contracts.BranchInterface{b1, b2}, nil)
	res, err := (&gitServerCache{s: gits}).AllBranches("1")
	assert.Nil(t, err)
	assert.Len(t, res, 2)
}

func Test_gitServerCache_AllProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	app.EXPECT().Cache().Return(c)
	gits := mockGitServer(m, app)

	c.EXPECT().Remember(cache2.NewKey("AllProjects"), AllProjectsCacheSeconds, gomock.Any()).Return(nil, errors.New("xxx"))
	_, err := (&gitServerCache{s: gits}).AllProjects()
	assert.Equal(t, "xxx", err.Error())

	app.EXPECT().Cache().Return(&cache2.NoCache{})
	p1 := mock.NewMockProjectInterface(m)
	p1.EXPECT()
	p1.EXPECT().GetID().Return(int64(1))
	p1.EXPECT().GetName().Return("name")
	p1.EXPECT().GetDefaultBranch().Return("dev")
	p1.EXPECT().GetPath().Return("path")
	p1.EXPECT().GetWebURL().Return("weburl")
	p1.EXPECT().GetAvatarURL().Return("avatar")
	p1.EXPECT().GetDescription().Return("desc")
	p2 := mock.NewMockProjectInterface(m)
	p2.EXPECT().GetID().Return(int64(2))
	p2.EXPECT().GetName().Return("name2")
	p2.EXPECT().GetDefaultBranch().Return("dev2")
	p2.EXPECT().GetPath().Return("path2")
	p2.EXPECT().GetWebURL().Return("weburl2")
	p2.EXPECT().GetAvatarURL().Return("avatar2")
	p2.EXPECT().GetDescription().Return("desc2")
	gits.EXPECT().AllProjects().Return([]contracts.ProjectInterface{p1, p2}, nil)
	res, err := (&gitServerCache{s: gits}).AllProjects()
	assert.Nil(t, err)
	assert.Len(t, res, 2)
}

func Test_gitServerCache_Destroy(t *testing.T) {
	assert.Nil(t, (&gitServerCache{}).Destroy())
}

type stateGitServer struct {
	calledMap map[string]bool
	GitServer
}

func (s *stateGitServer) GetCommit(pid string, sha string) (contracts.CommitInterface, error) {
	s.calledMap["GetCommit"] = true
	return nil, nil
}

func Test_gitServerCache_GetCommit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	gits := mockGitServer(m, app)

	app.EXPECT().Cache().Return(c)
	c.EXPECT().Remember(cache2.NewKey("GetCommit:%s-%s", "1", "xxx"), GetCommitCacheSeconds, gomock.Any()).Return(nil, errors.New("xxx"))
	_, err := (&gitServerCache{s: gits}).GetCommit("1", "xxx")
	assert.Equal(t, "xxx", err.Error())

	app.EXPECT().Cache().Return(&cache2.NoCache{})
	gits.EXPECT().GetCommit("1", "xxx").Return(nil, errors.New("xxx"))
	_, err = (&gitServerCache{s: gits}).GetCommit("1", "xxx")
	assert.Equal(t, "xxx", err.Error())

	app.EXPECT().Cache().Return(&cache2.NoCache{})
	commit := mock.NewMockCommitInterface(m)
	commit.EXPECT().GetID().Return("xxx")
	commit.EXPECT().GetShortID().Return("xxx")
	commit.EXPECT().GetTitle().Return("xxx")
	commit.EXPECT().GetCommittedDate().Return(nil)
	commit.EXPECT().GetAuthorName().Return("x")
	commit.EXPECT().GetAuthorEmail().Return("x")
	commit.EXPECT().GetCommitterName().Return("x")
	commit.EXPECT().GetCommitterEmail().Return("x")
	commit.EXPECT().GetCreatedAt().Return(nil)
	commit.EXPECT().GetMessage().Return("x")
	commit.EXPECT().GetProjectID().Return(int64(1))
	commit.EXPECT().GetWebURL().Return("x")
	gits.EXPECT().GetCommit("1", "xxx").Return(commit, nil)
	_, err = (&gitServerCache{s: gits}).GetCommit("1", "xxx")
	assert.Nil(t, err)
}

func (s *stateGitServer) GetCommitPipeline(pid string, sha string) (contracts.PipelineInterface, error) {
	s.calledMap["GetCommitPipeline"] = true
	return nil, nil
}
func Test_gitServerCache_GetCommitPipeline(t *testing.T) {
	s := &stateGitServer{calledMap: map[string]bool{}}
	(&gitServerCache{s: s}).GetCommitPipeline("", "")
	assert.True(t, s.calledMap["GetCommitPipeline"])
}

func (s *stateGitServer) GetDirectoryFilesWithBranch(pid string, branch string, path string, recursive bool) ([]string, error) {
	s.calledMap["GetDirectoryFilesWithBranch"] = true
	return nil, nil
}
func Test_gitServerCache_GetDirectoryFilesWithBranch(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	app.EXPECT().Cache().Return(c)
	gits := mockGitServer(m, app)
	c.EXPECT().Remember(cache2.NewKey("GetDirectoryFilesWithBranch-%s-%s-%s-%v", "", "", "", false), GetDirectoryFilesCacheSeconds, gomock.Any()).Times(1).Return(nil, errors.New("xxx"))

	_, err := (&gitServerCache{s: gits}).GetDirectoryFilesWithBranch("", "", "", false)
	assert.Equal(t, "xxx", err.Error())

	app.EXPECT().Cache().Return(&cache2.NoCache{})
	gits.EXPECT().GetDirectoryFilesWithBranch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"dev", "master"}, nil)
	res, err := (&gitServerCache{s: gits}).GetDirectoryFilesWithBranch("", "", "", false)
	assert.Nil(t, err)
	assert.Equal(t, []string{"dev", "master"}, res)

	app.EXPECT().Cache().Return(&cache2.NoCache{})
	gits.EXPECT().GetDirectoryFilesWithBranch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("xxx"))
	_, err = (&gitServerCache{s: gits}).GetDirectoryFilesWithBranch("", "", "", false)
	assert.Equal(t, "xxx", err.Error())
}

func Test_gitServerCache_GetDirectoryFilesWithSha(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	gits := mockGitServer(m, app)

	app.EXPECT().Cache().Return(c)
	c.EXPECT().Remember(cache2.NewKey("GetDirectoryFilesWithSha-%s-%s-%s-%v", "", "", "", false), GetDirectoryFilesCacheSeconds, gomock.Any()).Return(nil, errors.New("xxx"))
	_, err := (&gitServerCache{s: gits}).GetDirectoryFilesWithSha("", "", "", false)
	assert.Equal(t, "xxx", err.Error())

	app.EXPECT().Cache().Return(&cache2.NoCache{})
	gits.EXPECT().GetDirectoryFilesWithSha("", "", "", false).Return([]string{"aa.txt", "bb.txt"}, nil)
	res, _ := (&gitServerCache{s: gits}).GetDirectoryFilesWithSha("", "", "", false)
	assert.Len(t, res, 2)
}

func Test_gitServerCache_GetFileContentWithBranch(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	gits := mockGitServer(m, app)

	app.EXPECT().Cache().Return(c)
	c.EXPECT().Remember(cache2.NewKey("GetFileContentWithBranch-%s-%s-%s", "", "", ""), GetFileContentCacheSeconds, gomock.Any()).Return(nil, errors.New("xxx"))
	_, err := (&gitServerCache{s: gits}).GetFileContentWithBranch("", "", "")
	assert.Equal(t, "xxx", err.Error())

	app.EXPECT().Cache().Return(&cache2.NoCache{})
	gits.EXPECT().GetFileContentWithBranch("", "", "").Return("aaaa", nil)
	res, _ := (&gitServerCache{s: gits}).GetFileContentWithBranch("", "", "")
	assert.Equal(t, "aaaa", res)
}

func Test_gitServerCache_GetFileContentWithSha(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	gits := mockGitServer(m, app)

	app.EXPECT().Cache().Return(c)
	c.EXPECT().Remember(cache2.NewKey("GetFileContentWithSha-%s-%s-%s", "", "", ""), GetFileContentCacheSeconds, gomock.Any()).Return(nil, errors.New("xxx"))
	_, err := (&gitServerCache{s: gits}).GetFileContentWithSha("", "", "")
	assert.Equal(t, "xxx", err.Error())

	app.EXPECT().Cache().Return(&cache2.NoCache{})
	gits.EXPECT().GetFileContentWithSha("", "", "").Return("aaaa", nil)
	res, _ := (&gitServerCache{s: gits}).GetFileContentWithSha("", "", "")
	assert.Equal(t, "aaaa", res)
}

func (s *stateGitServer) GetProject(pid string) (contracts.ProjectInterface, error) {
	s.calledMap["GetProject"] = true
	return nil, nil
}

func Test_gitServerCache_GetProject(t *testing.T) {
	s := &stateGitServer{calledMap: map[string]bool{}}
	(&gitServerCache{s: s}).GetProject("1")
	assert.True(t, s.calledMap["GetProject"])
}

func Test_gitServerCache_Initialize(t *testing.T) {
	assert.Nil(t, (&gitServerCache{}).Initialize(nil))
}

func (s *stateGitServer) ListBranches(pid string, page, pageSize int) (contracts.ListBranchResponseInterface, error) {
	s.calledMap["ListBranches"] = true
	return nil, nil
}
func Test_gitServerCache_ListBranches(t *testing.T) {
	s := &stateGitServer{calledMap: map[string]bool{}}
	(&gitServerCache{s: s}).ListBranches("1", 1, 2)
	assert.True(t, s.calledMap["ListBranches"])
}
func (s *stateGitServer) ListCommits(pid string, branch string) ([]contracts.CommitInterface, error) {
	s.calledMap["ListCommits"] = true
	return nil, nil
}
func Test_gitServerCache_ListCommits(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	gits := mockGitServer(m, app)

	app.EXPECT().Cache().Return(c)
	c.EXPECT().Remember(cache2.NewKey("ListCommits:%s-%s", "1", "dev"), ListCommitsCacheSeconds, gomock.Any()).Return(nil, errors.New("xxx"))
	_, err := (&gitServerCache{s: gits}).ListCommits("1", "dev")
	assert.Equal(t, "xxx", err.Error())

	app.EXPECT().Cache().Return(&cache2.NoCache{})
	c1 := mock.NewMockCommitInterface(m)
	c1.EXPECT().GetID().Return("xxx")
	c1.EXPECT().GetShortID().Return("xxx")
	c1.EXPECT().GetTitle().Return("xxx")
	c1.EXPECT().GetCommittedDate().Return(nil)
	c1.EXPECT().GetAuthorName().Return("x")
	c1.EXPECT().GetAuthorEmail().Return("x")
	c1.EXPECT().GetCommitterName().Return("x")
	c1.EXPECT().GetCommitterEmail().Return("x")
	c1.EXPECT().GetCreatedAt().Return(nil)
	c1.EXPECT().GetMessage().Return("x")
	c1.EXPECT().GetProjectID().Return(int64(1))
	c1.EXPECT().GetWebURL().Return("x")
	c2 := mock.NewMockCommitInterface(m)
	c2.EXPECT().GetID().Return("xxx")
	c2.EXPECT().GetShortID().Return("xxx")
	c2.EXPECT().GetTitle().Return("xxx")
	c2.EXPECT().GetCommittedDate().Return(nil)
	c2.EXPECT().GetAuthorName().Return("x")
	c2.EXPECT().GetAuthorEmail().Return("x")
	c2.EXPECT().GetCommitterName().Return("x")
	c2.EXPECT().GetCommitterEmail().Return("x")
	c2.EXPECT().GetCreatedAt().Return(nil)
	c2.EXPECT().GetMessage().Return("x")
	c2.EXPECT().GetProjectID().Return(int64(1))
	c2.EXPECT().GetWebURL().Return("x")
	gits.EXPECT().ListCommits("1", "dev").Return([]contracts.CommitInterface{c1, c2}, nil)
	_, err = (&gitServerCache{s: gits}).ListCommits("1", "dev")
	assert.Nil(t, err)
}

func (s *stateGitServer) ListProjects(page, pageSize int) (contracts.ListProjectResponseInterface, error) {
	s.calledMap["ListProjects"] = true
	return nil, nil
}

func Test_gitServerCache_ListProjects(t *testing.T) {
	s := &stateGitServer{calledMap: map[string]bool{}}
	(&gitServerCache{s: s}).ListProjects(1, 14)
	assert.True(t, s.calledMap["ListProjects"])
}

func Test_gitServerCache_Name(t *testing.T) {
	assert.Equal(t, "", (&gitServerCache{}).Name())
}

func Test_newGitServerCache(t *testing.T) {
	assert.Implements(t, (*GitServer)(nil), newGitServerCache(nil))
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

func Test_gitServerCache_ReCacheAllProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	app.EXPECT().Cache().Return(c)
	server := mockGitServer(m, app)
	cache := newGitServerCache(server)
	server.EXPECT().AllProjects().Return(nil, errors.New("xx"))
	err := cache.ReCacheAllProjects()
	assert.Equal(t, "xx", err.Error())

	server.EXPECT().AllProjects().Return(nil, nil)
	c.EXPECT().SetWithTTL(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	err = cache.ReCacheAllProjects()
	assert.Nil(t, err)
}

func Test_gitServerCache_ReCacheAllBranches(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	app.EXPECT().Cache().Return(c)
	server := mockGitServer(m, app)
	cache := newGitServerCache(server)
	server.EXPECT().AllBranches("1").Return(nil, errors.New("xx"))
	err := cache.ReCacheAllBranches("1")
	assert.Equal(t, "xx", err.Error())

	server.EXPECT().AllBranches("1").Return(nil, nil)
	c.EXPECT().SetWithTTL(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	err = cache.ReCacheAllBranches("1")
	assert.Nil(t, err)
}
