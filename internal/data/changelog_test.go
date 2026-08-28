package data

import (
	"context"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestChangelogRepo_Create(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{
		DB: entdb,
	}))

	ns := createNamespace(entdb)
	project := createProject(entdb, ns.ID)

	input := &biz.CreateChangeLogInput{
		Version:          1,
		Username:         "testUser",
		Config:           "testConfig",
		GitBranch:        "testBranch",
		GitCommit:        "testCommit",
		DockerImage:      []string{"testImage"},
		EnvValues:        []*types.KeyValue{{Key: "testKey", Value: "testValue"}},
		ExtraValues:      []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue"}},
		FinalExtraValues: []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue1"}},
		GitCommitWebURL:  "testWebURL",
		GitCommitTitle:   "testTitle",
		GitCommitAuthor:  "testAuthor",
		GitCommitDate:    nil,
		ConfigChanged:    false,
		ProjectID:        project.ID,
	}
	changelog, err := repo.Create(context.TODO(), input)
	assert.Nil(t, err)
	assert.Equal(t, input.Version, changelog.Version)
	assert.Equal(t, input.Username, changelog.Username)
	assert.Equal(t, input.Config, changelog.Config)
	assert.Equal(t, input.GitBranch, changelog.GitBranch)
	assert.Equal(t, input.GitCommit, changelog.GitCommit)
	assert.Equal(t, input.DockerImage, changelog.DockerImage)
	assert.Equal(t, input.EnvValues, changelog.EnvValues)
	assert.Equal(t, input.ExtraValues, changelog.ExtraValues)
	assert.Equal(t, input.FinalExtraValues, changelog.FinalExtraValues)
	assert.Equal(t, input.GitCommitWebURL, changelog.GitCommitWebURL)
	assert.Equal(t, input.GitCommitTitle, changelog.GitCommitTitle)
	assert.Equal(t, input.GitCommitAuthor, changelog.GitCommitAuthor)
	assert.Equal(t, input.GitCommitDate, changelog.GitCommitDate)
	assert.Equal(t, input.ConfigChanged, changelog.ConfigChanged)
	assert.Equal(t, input.ProjectID, changelog.ProjectID)
}

func TestChangelogRepo_FindLastChangelogsByProjectID(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{
		DB: entdb,
	}))

	ns := createNamespace(entdb)
	project := createProject(entdb, ns.ID)

	for i := 0; i < 20; i++ {
		entdb.Changelog.Create().
			SetVersion(i).
			SetUsername("").
			SetProject(project).
			SaveX(context.TODO())
	}
	entdb.Changelog.Create().
		SetVersion(100).
		SetUsername("").
		SetConfigChanged(true).
		SetProject(project).
		SaveX(context.TODO())

	input := &biz.FindLastChangelogsByProjectIDChangeLogInput{
		OnlyChanged:        false,
		ProjectID:          1,
		OrderByVersionDesc: lo.ToPtr(true),
		Limit:              10,
	}
	changelogs, err := repo.FindLastChangelogsByProjectID(context.TODO(), input)
	assert.Nil(t, err)
	assert.Len(t, changelogs, 10)
	assert.Equal(t, 100, changelogs[0].Version)

	input = &biz.FindLastChangelogsByProjectIDChangeLogInput{
		OnlyChanged:        true,
		ProjectID:          1,
		OrderByVersionDesc: lo.ToPtr(true),
		Limit:              10,
	}
	changelogs, err = repo.FindLastChangelogsByProjectID(context.TODO(), input)
	assert.Nil(t, err)
	assert.Len(t, changelogs, 1)
	assert.Equal(t, 100, changelogs[0].Version)
}

func TestChangelogRepo_FindLastChangeByProjectID_WithValidProjectID(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{
		DB: entdb,
	}))

	ns := createNamespace(entdb)
	project := createProject(entdb, ns.ID)

	entdb.Changelog.Create().
		SetVersion(1).
		SetUsername("testUser").
		SetProject(project).
		SaveX(context.TODO())

	changelog, err := repo.FindLastChangeByProjectID(context.TODO(), project.ID)
	assert.Nil(t, err)
	assert.NotNil(t, changelog)
	assert.Equal(t, 1, changelog.Version)
	assert.Equal(t, "testUser", changelog.Username)
	assert.Equal(t, project.ID, changelog.ProjectID)
}

func TestChangelogRepo_FindLastChangeByProjectID_WithInvalidProjectID(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{
		DB: entdb,
	}))

	changelog, err := repo.FindLastChangeByProjectID(context.TODO(), -1)
	assert.NotNil(t, err)
	assert.Nil(t, changelog)
}

func TestChangelogRepo_FindLastChangeByProjectID_WithNoChangelog(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{
		DB: entdb,
	}))

	ns := createNamespace(entdb)
	project := createProject(entdb, ns.ID)

	changelog, err := repo.FindLastChangeByProjectID(context.TODO(), project.ID)
	assert.NotNil(t, err)
	assert.Nil(t, changelog)
}

func TestNewChangelogRepo(t *testing.T) {
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{}))
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.(*changelogRepo).logger)
	assert.NotNil(t, repo.(*changelogRepo).data)
}

func TestToChangeLog_WithValidChangelog(t *testing.T) {
	c := &ent.Changelog{
		ID:               1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DeletedAt:        nil,
		Version:          1,
		Username:         "testUser",
		Config:           "testConfig",
		GitBranch:        "testBranch",
		GitCommit:        "testCommit",
		DockerImage:      []string{"testImage"},
		EnvValues:        []*types.KeyValue{{Key: "testKey", Value: "testValue"}},
		ExtraValues:      []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue"}},
		FinalExtraValues: []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue1"}},
		GitCommitWebURL:  "testWebURL",
		GitCommitTitle:   "testTitle",
		GitCommitAuthor:  "testAuthor",
		GitCommitDate:    nil,
		ConfigChanged:    false,
		ProjectID:        1,
		Edges:            ent.ChangelogEdges{Project: &ent.Project{ID: 1}},
	}

	result := toChangeLog(c)

	assert.NotNil(t, result)
	assert.Equal(t, c.ID, result.ID)
	assert.Equal(t, c.Version, result.Version)
	assert.Equal(t, c.Username, result.Username)
	assert.Equal(t, c.Config, result.Config)
	assert.Equal(t, c.GitBranch, result.GitBranch)
	assert.Equal(t, c.GitCommit, result.GitCommit)
	assert.Equal(t, c.DockerImage, result.DockerImage)
	assert.Equal(t, c.EnvValues, result.EnvValues)
	assert.Equal(t, c.ExtraValues, result.ExtraValues)
	assert.Equal(t, c.FinalExtraValues, result.FinalExtraValues)
	assert.Equal(t, c.GitCommitWebURL, result.GitCommitWebURL)
	assert.Equal(t, c.GitCommitTitle, result.GitCommitTitle)
	assert.Equal(t, c.GitCommitAuthor, result.GitCommitAuthor)
	assert.Equal(t, c.GitCommitDate, result.GitCommitDate)
	assert.Equal(t, c.ConfigChanged, result.ConfigChanged)
	assert.Equal(t, c.ProjectID, result.ProjectID)
}

func TestToChangeLog_WithNilChangelog(t *testing.T) {
	result := toChangeLog(nil)
	assert.Nil(t, result)
}

func TestChangelogRepoCreate_WithValidInput(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{
		DB: entdb,
	}))

	ns := createNamespace(entdb)
	project := createProject(entdb, ns.ID)

	input := &biz.CreateChangeLogInput{
		Version:          1,
		Username:         "testUser",
		Config:           "testConfig",
		GitBranch:        "testBranch",
		GitCommit:        "testCommit",
		DockerImage:      []string{"testImage"},
		EnvValues:        []*types.KeyValue{{Key: "testKey", Value: "testValue"}},
		ExtraValues:      []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue"}},
		FinalExtraValues: []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue1"}},
		GitCommitWebURL:  "testWebURL",
		GitCommitTitle:   "testTitle",
		GitCommitAuthor:  "testAuthor",
		GitCommitDate:    nil,
		ConfigChanged:    false,
		ProjectID:        project.ID,
	}
	changelog, err := repo.Create(context.TODO(), input)
	assert.Nil(t, err)
	assert.NotNil(t, changelog)
}

func TestChangelogRepoCreate_WithInvalidInput(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{
		DB: entdb,
	}))

	input := &biz.CreateChangeLogInput{
		Version:          1,
		Username:         "",
		Config:           "",
		GitBranch:        "",
		GitCommit:        "",
		DockerImage:      []string{},
		EnvValues:        []*types.KeyValue{},
		ExtraValues:      []*websocket_pb.ExtraValue{},
		FinalExtraValues: []*websocket_pb.ExtraValue{},
		GitCommitWebURL:  "",
		GitCommitTitle:   "",
		GitCommitAuthor:  "",
		GitCommitDate:    nil,
		ConfigChanged:    false,
		ProjectID:        0,
	}
	changelog, err := repo.Create(context.TODO(), input)
	assert.NotNil(t, err)
	assert.Nil(t, changelog)
}

func createProject(entdb *ent.Client, nsID int) *ent.Project {
	return entdb.Project.Create().
		SetGitBranch("").
		SetGitCommit("").
		SetConfig("").
		SetGitProjectID(1).
		SetCreator("").
		SetName("testProject").
		SetNamespaceID(nsID).
		SaveX(context.TODO())
}

func createNamespace(entdb *ent.Client) *ent.Namespace {
	return entdb.Namespace.Create().SetName(rand.String(10)).SetCreatorEmail(rand.String(20) + "@q.c").SaveX(context.TODO())
}

// TestChangelogRepo_CountByProjectIDs 按项目聚合变更记录数：两项目各落多条，计数分组正确。
func TestChangelogRepo_CountByProjectIDs(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: entdb}))

	ns := createNamespace(entdb)
	p1 := createProject(entdb, ns.ID)
	p2 := createProject(entdb, ns.ID)
	for i := 0; i < 2; i++ {
		entdb.Changelog.Create().SetVersion(i).SetUsername("").SetProject(p1).SaveX(context.TODO())
	}
	entdb.Changelog.Create().SetVersion(0).SetUsername("").SetProject(p2).SaveX(context.TODO())

	counts, err := repo.CountByProjectIDs(context.TODO(), p1.ID, p2.ID)
	assert.NoError(t, err)
	assert.Equal(t, map[int]int{p1.ID: 2, p2.ID: 1}, counts)
}

// TestChangelogRepo_CountByProjectIDs_Empty 空 ID 集合返回空 map 不报错。
func TestChangelogRepo_CountByProjectIDs_Empty(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: entdb}))

	counts, err := repo.CountByProjectIDs(context.TODO())
	assert.NoError(t, err)
	assert.Empty(t, counts)
}

// TestChangelogRepo_CountByProjectIDs_Error GROUP BY 聚合失败整体上抛（关闭的 DB）。
func TestChangelogRepo_CountByProjectIDs_Error(t *testing.T) {
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: mustClosedDB(t)}))
	counts, err := repo.CountByProjectIDs(context.TODO(), 1, 2)
	assert.Nil(t, counts)
	assert.Error(t, err)
}
