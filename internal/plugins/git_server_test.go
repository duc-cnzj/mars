package plugins

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
)

type gitPlugin struct {
	called bool
	GitServer
}

func (g *gitPlugin) Initialize(args map[string]any) error {
	g.called = true
	return nil
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
		AuthorEmail: "admin@mars.com",
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

type cache struct {
	key     string
	seconds int
	called  bool
}

func (c *cache) Remember(key string, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	c.key = key
	c.seconds = seconds
	c.called = true
	return nil, nil
}

func Test_gitServerCache_AllBranches(t *testing.T) {
	c := &cache{}
	ma := &mockApp{cache: c}
	instance.SetInstance(ma)
	(&gitServerCache{}).AllBranches("1")
	assert.True(t, c.called)
	assert.Equal(t, AllBranchesCacheSeconds, c.seconds)
	assert.Equal(t, fmt.Sprintf("AllBranches-%v", 1), c.key)
}

func Test_gitServerCache_AllProjects(t *testing.T) {
	c := &cache{}
	ma := &mockApp{cache: c}
	instance.SetInstance(ma)
	(&gitServerCache{}).AllProjects()
	assert.True(t, c.called)
	assert.Equal(t, AllProjectsCacheSeconds, c.seconds)
	assert.Equal(t, "AllProjects", c.key)
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
	s := &stateGitServer{calledMap: map[string]bool{}}
	c := &cache{}
	ma := &mockApp{cache: c}
	instance.SetInstance(ma)
	(&gitServerCache{s: s}).GetCommit("", "")
	assert.True(t, c.called)
	assert.Equal(t, GetCommitCacheSeconds, c.seconds)
	assert.Equal(t, fmt.Sprintf("GetCommit:%s-%s", "", ""), c.key)
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
	s := &stateGitServer{calledMap: map[string]bool{}}
	c := &cache{}
	ma := &mockApp{cache: c}
	instance.SetInstance(ma)
	(&gitServerCache{s: s}).GetDirectoryFilesWithBranch("", "", "", false)
	assert.True(t, c.called)
	assert.Equal(t, GetDirectoryFilesCacheSeconds, c.seconds)
	assert.Equal(t, fmt.Sprintf("GetDirectoryFilesWithBranch-%s-%s-%s-%v", "", "", "", false), c.key)
}

func Test_gitServerCache_GetDirectoryFilesWithSha(t *testing.T) {
	c := &cache{}
	ma := &mockApp{cache: c}
	instance.SetInstance(ma)
	(&gitServerCache{}).GetDirectoryFilesWithSha("", "", "", false)
	assert.True(t, c.called)
	assert.Equal(t, GetDirectoryFilesCacheSeconds, c.seconds)
	assert.Equal(t, fmt.Sprintf("GetDirectoryFilesWithSha-%s-%s-%s-%v", "", "", "", false), c.key)
}

func Test_gitServerCache_GetFileContentWithBranch(t *testing.T) {
	c := &cache{}
	ma := &mockApp{cache: c}
	instance.SetInstance(ma)
	(&gitServerCache{}).GetFileContentWithBranch("", "", "")
	assert.True(t, c.called)
	assert.Equal(t, GetFileContentCacheSeconds, c.seconds)
	assert.Equal(t, fmt.Sprintf("GetFileContentWithBranch-%s-%s-%s", "", "", ""), c.key)
}

func Test_gitServerCache_GetFileContentWithSha(t *testing.T) {
	c := &cache{}
	ma := &mockApp{cache: c}
	instance.SetInstance(ma)
	(&gitServerCache{}).GetFileContentWithSha("", "", "")
	assert.True(t, c.called)
	assert.Equal(t, GetFileContentCacheSeconds, c.seconds)
	assert.Equal(t, fmt.Sprintf("GetFileContentWithSha-%s-%s-%s", "", "", ""), c.key)
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

func (s *stateGitServer) ListBranches(pid string, page, pageSize int) (ListBranchResponseInterface, error) {
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
	c := &cache{}
	ma := &mockApp{cache: c}
	instance.SetInstance(ma)
	(&gitServerCache{}).ListCommits("", "")
	assert.True(t, c.called)
	assert.Equal(t, ListCommitsCacheSeconds, c.seconds)
	assert.Equal(t, fmt.Sprintf("ListCommits:%s-%s", "", ""), c.key)
}

func (s *stateGitServer) ListProjects(page, pageSize int) (ListProjectResponseInterface, error) {
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
