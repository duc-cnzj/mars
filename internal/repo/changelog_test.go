package repo

import (
	"context"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestChangelogRepo_Create(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewChangelogRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		DB: db,
	}))

	ns := createNamespace(db)
	project := createProject(db, ns.ID)

	input := &CreateChangeLogInput{
		Version:          1,
		Username:         "testUser",
		Config:           "testConfig",
		GitBranch:        "testBranch",
		GitCommit:        "testCommit",
		DockerImage:      []string{"testImage"},
		EnvValues:        []*types.KeyValue{{Key: "testKey", Value: "testValue"}},
		ExtraValues:      []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue"}},
		FinalExtraValues: []string{"testFinalExtraValue"},
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
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewChangelogRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		DB: db,
	}))

	ns := createNamespace(db)
	project := createProject(db, ns.ID)

	for i := 0; i < 20; i++ {
		db.Changelog.Create().
			SetVersion(i).
			SetUsername("").
			SetProject(project).
			SaveX(context.TODO())
	}
	db.Changelog.Create().
		SetVersion(100).
		SetUsername("").
		SetConfigChanged(true).
		SetProject(project).
		SaveX(context.TODO())

	input := &FindLastChangelogsByProjectIDChangeLogInput{
		OnlyChanged:        false,
		ProjectID:          1,
		OrderByVersionDesc: lo.ToPtr(true),
		Limit:              10,
	}
	changelogs, err := repo.FindLastChangelogsByProjectID(context.TODO(), input)
	assert.Nil(t, err)
	assert.Len(t, changelogs, 10)
	assert.Equal(t, 100, changelogs[0].Version)

	input = &FindLastChangelogsByProjectIDChangeLogInput{
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
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewChangelogRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		DB: db,
	}))

	ns := createNamespace(db)
	project := createProject(db, ns.ID)

	db.Changelog.Create().
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
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewChangelogRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		DB: db,
	}))

	changelog, err := repo.FindLastChangeByProjectID(context.TODO(), -1)
	assert.NotNil(t, err)
	assert.Nil(t, changelog)
}

func TestChangelogRepo_FindLastChangeByProjectID_WithNoChangelog(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewChangelogRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		DB: db,
	}))

	ns := createNamespace(db)
	project := createProject(db, ns.ID)

	changelog, err := repo.FindLastChangeByProjectID(context.TODO(), project.ID)
	assert.NotNil(t, err)
	assert.Nil(t, changelog)
}

func TestNewChangelogRepo(t *testing.T) {
	repo := NewChangelogRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{}))
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
		FinalExtraValues: []string{"testFinalExtraValue"},
		GitCommitWebURL:  "testWebURL",
		GitCommitTitle:   "testTitle",
		GitCommitAuthor:  "testAuthor",
		GitCommitDate:    nil,
		ConfigChanged:    false,
		ProjectID:        1,
		Edges:            ent.ChangelogEdges{Project: &ent.Project{ID: 1}},
	}

	result := ToChangeLog(c)

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
	result := ToChangeLog(nil)
	assert.Nil(t, result)
}

func TestChangelogRepoCreate_WithValidInput(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewChangelogRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		DB: db,
	}))

	ns := createNamespace(db)
	project := createProject(db, ns.ID)

	input := &CreateChangeLogInput{
		Version:          1,
		Username:         "testUser",
		Config:           "testConfig",
		GitBranch:        "testBranch",
		GitCommit:        "testCommit",
		DockerImage:      []string{"testImage"},
		EnvValues:        []*types.KeyValue{{Key: "testKey", Value: "testValue"}},
		ExtraValues:      []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue"}},
		FinalExtraValues: []string{"testFinalExtraValue"},
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
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewChangelogRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		DB: db,
	}))

	input := &CreateChangeLogInput{
		Version:          1,
		Username:         "",
		Config:           "",
		GitBranch:        "",
		GitCommit:        "",
		DockerImage:      []string{},
		EnvValues:        []*types.KeyValue{},
		ExtraValues:      []*websocket_pb.ExtraValue{},
		FinalExtraValues: []string{},
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

func createProject(db *ent.Client, nsID int) *ent.Project {
	return db.Project.Create().
		SetGitBranch("").
		SetGitCommit("").
		SetConfig("").
		SetGitProjectID(1).
		SetCreator("").
		SetName("testProject").
		SetNamespaceID(nsID).
		SaveX(context.TODO())
}

func createNamespace(db *ent.Client) *ent.Namespace {
	return db.Namespace.Create().SetName("testns").SaveX(context.TODO())
}
