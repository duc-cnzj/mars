package data

import (
	"context"
	"strings"
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
	"github.com/stretchr/testify/require"
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

// TestChangelogRepo_Create_LongGitCommitTitle 回归：变更记录的 git commit title 超过
// 255 字节（含多字节 UTF-8）仍可完整落库。project 放宽为 longtext 后，从 DB 回读的
// 超长 title 会透传到 changelog，changelog 的 git_commit_title 必须同步放宽，
// 否则变更记录会静默写入失败、审计丢数据。
func TestChangelogRepo_Create_LongGitCommitTitle(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{
		DB: entdb,
	}))

	project := createProject(entdb, createNamespace(entdb).ID)

	longTitle := strings.Repeat("长", 100) // 300 字节 > 255
	require.Greater(t, len(longTitle), 255)

	changelog, err := repo.Create(context.TODO(), &biz.CreateChangeLogInput{
		Version:        1,
		Username:       "testUser",
		GitCommitTitle: longTitle,
		ProjectID:      project.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, longTitle, changelog.GitCommitTitle)
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

// TestChangelogRepo_SelectCreatedAtBetween 真库验证部署趋势取数：只回窗口内 changelog 的
// created_at（Scan 到 []time.Time 单列），软删行被拦截器排除，越过边界的行不返回。
func TestChangelogRepo_SelectCreatedAtBetween(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: entdb}))

	// 窗口锚点：留足前后余量，确保下述 created_at 全部落在 [lo, hi) 内。
	lo := time.Now().Add(-time.Hour)
	r1 := entdb.Changelog.Create().SetVersion(1).SetUsername("u1").SaveX(context.TODO())
	r2 := entdb.Changelog.Create().SetVersion(2).SetUsername("u2").SaveX(context.TODO())
	hi := time.Now().Add(time.Hour)

	created, err := repo.SelectCreatedAtBetween(context.TODO(), lo, hi)
	assert.NoError(t, err)
	if assert.Len(t, created, 2) {
		// ent 客户端时间戳落库再读回，忽略 monotonic/location 差异后应等于原值。
		assert.True(t, created[0].Equal(r1.CreatedAt) || created[0].Equal(r2.CreatedAt))
		assert.True(t, created[1].Equal(r1.CreatedAt) || created[1].Equal(r2.CreatedAt))
	}

	// 软删 r2（SoftDeleteMixin 把 Delete 转 OpUpdate+SetDeletedAt）：再查同窗口只剩 r1。
	delErr := entdb.Changelog.DeleteOneID(r2.ID).Exec(context.TODO())
	require.NoError(t, delErr)
	created, err = repo.SelectCreatedAtBetween(context.TODO(), lo, hi)
	assert.NoError(t, err)
	if assert.Len(t, created, 1) {
		assert.True(t, created[0].Equal(r1.CreatedAt))
	}
}

// TestChangelogRepo_SelectCreatedAtBetween_Empty 空窗口返回空切片不报错。
func TestChangelogRepo_SelectCreatedAtBetween_Empty(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: entdb}))

	created, err := repo.SelectCreatedAtBetween(context.TODO(), time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC))
	assert.NoError(t, err)
	assert.Empty(t, created)
}

// TestChangelogRepo_SelectCreatedAtBetween_Error 查询失败整体上抛（关闭的 DB）。
func TestChangelogRepo_SelectCreatedAtBetween_Error(t *testing.T) {
	repo := NewChangelogRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: mustClosedDB(t)}))
	created, err := repo.SelectCreatedAtBetween(context.TODO(), time.Now().Add(-time.Hour), time.Now())
	assert.Nil(t, created)
	assert.Error(t, err)
}
