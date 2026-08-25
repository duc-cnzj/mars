package data

import (
	"context"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	repoq "github.com/duc-cnzj/mars/v6/internal/data/ent/repo"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRepoImpl_All(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))
	res, err := repo.All(context.TODO(), &biz.AllRepoRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestRepoImpl_List(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))
	res, pag, err := repo.List(context.TODO(), &biz.ListRepoRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, pag)
}

func TestRepoImpl_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))
	_, err := repo.Show(context.TODO(), 1)
	s, _ := status.FromError(err)

	assert.Equal(t, "NotFound", s.Code().String())
}

func TestRepoImpl_Create(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), mockGitRepo)
	mockGitRepo.EXPECT().GetByProjectID(gomock.Any(), 100).Return(&biz.GitProject{}, nil)

	mockGitRepo.EXPECT().GetChartValuesYaml(gomock.Any(), "100|main|chart").Return(`config: config`, nil)

	res, err := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name:         "app",
		Enabled:      true,
		NeedGitRepo:  true,
		GitProjectID: lo.ToPtr(int32(100)),
		MarsConfig:   &mars.Config{ConfigField: "config", LocalChartPath: "100|main|chart"},
		Description:  "desc",
	})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.GitProjectID)
	assert.True(t, res.MarsConfig.IsSimpleEnv)
}

func TestRepoImpl_Update(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), mockGitRepo)

	create, err := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name: "app",
	})
	assert.Nil(t, err)

	mockGitRepo.EXPECT().GetByProjectID(gomock.Any(), 100).Return(&biz.GitProject{
		DefaultBranch: "dev",
		Name:          "a",
	}, nil)
	mockGitRepo.EXPECT().GetChartValuesYaml(gomock.Any(), "100|main|chart").Return(`config: config`, nil)

	res, err := repo.Update(context.TODO(), &biz.UpdateRepoInput{
		ID:           int32(create.ID),
		Name:         "abc",
		NeedGitRepo:  true,
		GitProjectID: lo.ToPtr(int32(100)),
		MarsConfig:   &mars.Config{ConfigField: "config", LocalChartPath: "100|main|chart"},
		Description:  "dex",
	})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.MarsConfig.IsSimpleEnv)
	assert.Equal(t, "abc", res.Name)
	assert.Equal(t, "dex", res.Description)
	assert.Equal(t, true, res.NeedGitRepo)
	assert.Equal(t, "config", res.MarsConfig.ConfigField)
	assert.Equal(t, "dev", res.DefaultBranch)
	assert.Equal(t, "a", res.GitProjectName)
	assert.Equal(t, int32(100), res.GitProjectID)
}

func TestRepoImpl_ToggleEnabled(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	create, err := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name: "app",
	})
	assert.Nil(t, err)
	assert.False(t, create.Enabled)

	res, err := repo.ToggleEnabled(context.TODO(), create.ID, true)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.Enabled)
}

func TestRepoImpl_ToggleEnabled_WithProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	create, err := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name:    "app",
		Enabled: true,
	})
	assert.Nil(t, err)
	assert.True(t, create.Enabled)

	ns := createNamespace(entdb)
	project := createProject(entdb, ns.ID)
	project.Update().SetRepoID(create.ID).SaveX(context.TODO())

	// Attempt to disable the repo, should succeed (validation in biz layer)
	_, err = repo.ToggleEnabled(context.TODO(), create.ID, false)
	assert.Nil(t, err)
}

func TestRepo_GetMarsConfig_WithExistingConfig(t *testing.T) {
	r := &biz.Repo{
		MarsConfig: &mars.Config{ConfigField: "existing_config"},
	}
	cfg := r.GetMarsConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "existing_config", cfg.ConfigField)
}

func TestRepo_GetMarsConfig_WithoutExistingConfig(t *testing.T) {
	r := &biz.Repo{}
	cfg := r.GetMarsConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "", cfg.ConfigField)
}

func TestRepoImpl_GetByName(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	// 空表按名称查询 → ent NotFound → errs.Wrap 归类为 404。
	_, err := repo.GetByName(context.TODO(), "nope")
	s, _ := status.FromError(err)
	assert.Equal(t, "NotFound", s.Code().String())

	created, err := repo.Create(context.TODO(), &biz.CreateRepoInput{Name: "app"})
	assert.Nil(t, err)
	got, err := repo.GetByName(context.TODO(), "app")
	assert.Nil(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "app", got.Name)
}

func TestRepoImpl_GetProjNameAndBranch_WithExistingProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(NewMockDataStore(m), mockGitRepo)
	mockGitRepo.EXPECT().GetByProjectID(gomock.Any(), gomock.Any()).Return(&biz.GitProject{
		DefaultBranch: "main",
		Name:          "projName",
	}, nil)
	projName, defaultBranch, err := repo.(*repoImpl).GetProjNameAndBranch(context.TODO(), 1)
	assert.Nil(t, err)
	assert.NotNil(t, projName)
	assert.NotNil(t, defaultBranch)
}

func TestRepoImpl_GetProjNameAndBranch_WithNonExistingProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(NewMockDataStore(m), mockGitRepo)
	mockGitRepo.EXPECT().GetByProjectID(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)

	projName, defaultBranch, err := repo.(*repoImpl).GetProjNameAndBranch(context.TODO(), 1)

	assert.NotNil(t, err)
	assert.Nil(t, projName)
	assert.Nil(t, defaultBranch)
}

func TestCloneRepoWithValidInput(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	entdb, _ := NewSqliteDB()
	defer entdb.Close()

	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), mockGitRepo)

	create, _ := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name: "app",
	})

	_, err := repo.Clone(context.TODO(), &biz.CloneRepoInput{
		ID:   create.ID,
		Name: "clone",
	})

	assert.Nil(t, err)
}

func TestRepoImpl_Delete_WithExistingRepo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	create, _ := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name: "app",
	})

	err := repo.Delete(context.TODO(), create.ID)
	assert.Nil(t, err)
}

func TestRepoImpl_Delete_WithNonExistingRepo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	err := repo.Delete(context.TODO(), 9999)
	assert.NotNil(t, err)
}

func TestRepoImpl_Delete_WithRepoHavingProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	create, _ := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name: "app",
	})

	ns := createNamespace(entdb)
	project := createProject(entdb, ns.ID)
	project.Update().SetRepoID(create.ID).SaveX(context.TODO())

	err := repo.Delete(context.TODO(), create.ID)
	assert.Nil(t, err)
}

func Test_repoImpl_Get(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	create, _ := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name: "app",
	})
	ns := createNamespace(entdb)
	project := createProject(entdb, ns.ID)
	project.Update().SetRepoID(create.ID).SaveX(context.TODO())

	get, err := repo.Get(context.TODO(), create.ID)
	assert.Nil(t, err)
	assert.Len(t, get.Projects, 0)
	s, err := repo.Show(context.TODO(), create.ID)
	assert.Nil(t, err)
	assert.Len(t, s.Projects, 1)
}

// TestRepoImpl_ErrorBranches 覆盖 repoImpl 的 DB 查询错误分支与 gitRepo 错误透传。
func TestRepoImpl_ErrorBranches(t *testing.T) {
	ctx := context.TODO()

	newClosed := func(t *testing.T) biz.RepoRepo {
		t.Helper()
		return NewRepo(NewDataImpl(&NewDataParams{DB: mustClosedDB(t), Cfg: &config.Config{}}), NewMockGitRepo(gomock.NewController(t)))
	}

	t.Run("All query error", func(t *testing.T) {
		_, err := newClosed(t).All(ctx, &biz.AllRepoRequest{})
		assert.Error(t, err)
	})

	t.Run("List query error", func(t *testing.T) {
		_, _, err := newClosed(t).List(ctx, &biz.ListRepoRequest{})
		assert.Error(t, err)
	})

	t.Run("Create Save error", func(t *testing.T) {
		// 名称唯一性校验已上收 biz 层，data Create 直落持久化，closed DB 在 Save 时报错。
		_, err := newClosed(t).Create(ctx, &biz.CreateRepoInput{Name: "x"})
		assert.Error(t, err)
	})

	t.Run("GetByName query error", func(t *testing.T) {
		_, err := newClosed(t).GetByName(ctx, "x")
		assert.Error(t, err)
	})

	t.Run("Create gitRepo error", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		entdb, err := NewSqliteDB()
		require.NoError(t, err)
		defer entdb.Close()
		git := NewMockGitRepo(m)
		git.EXPECT().GetByProjectID(gomock.Any(), 1).Return(nil, assert.AnError)
		repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), git)
		_, err = repo.Create(ctx, &biz.CreateRepoInput{Name: "fresh", NeedGitRepo: true, GitProjectID: lo.ToPtr(int32(1))})
		assert.Error(t, err)
	})

	t.Run("Update NotFound error", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		entdb, err := NewSqliteDB()
		require.NoError(t, err)
		defer entdb.Close()
		repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))
		// 校验已上收 biz 层，data Update 直落 UpdateOneID，目标行不存在 → ent NotFound。
		_, err = repo.Update(ctx, &biz.UpdateRepoInput{ID: 999999, Name: "nope"})
		assert.Error(t, err)
	})

	t.Run("Update gitRepo error", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		entdb, err := NewSqliteDB()
		require.NoError(t, err)
		defer entdb.Close()
		git := NewMockGitRepo(m)
		repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), git)
		created, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "base"})
		require.NoError(t, err)
		git.EXPECT().GetByProjectID(gomock.Any(), 1).Return(nil, assert.AnError)
		_, err = repo.Update(ctx, &biz.UpdateRepoInput{ID: int32(created.ID), Name: "base", NeedGitRepo: true, GitProjectID: lo.ToPtr(int32(1))})
		assert.Error(t, err)
	})

	t.Run("Update clear git fields branch", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		entdb, err := NewSqliteDB()
		require.NoError(t, err)
		defer entdb.Close()
		repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))
		created, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "clear"})
		require.NoError(t, err)
		// 同名同 ID 更新 + NeedGitRepo=false → 走到 ClearGitProjectID/ClearGitProjectName/ClearDefaultBranch 分支。
		res, err := repo.Update(ctx, &biz.UpdateRepoInput{ID: int32(created.ID), Name: "clear", NeedGitRepo: false})
		require.NoError(t, err)
		assert.False(t, res.NeedGitRepo)
	})

	t.Run("Clone Get error", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		repo := NewRepo(NewDataImpl(&NewDataParams{DB: mustClosedDB(t)}), NewMockGitRepo(m))
		_, err := repo.Clone(ctx, &biz.CloneRepoInput{ID: 1, Name: "x"})
		assert.Error(t, err)
	})

	t.Run("Clone Create error", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		entdb, err := NewSqliteDB()
		require.NoError(t, err)
		defer entdb.Close()
		git := NewMockGitRepo(m)
		// 源 repo 开启 NeedGitRepo → Clone 透传 → 内部 Create 走 GetProjNameAndBranch 报错。
		orig := entdb.Repo.Create().SetName("orig").SetNeedGitRepo(true).SaveX(ctx)
		git.EXPECT().GetByProjectID(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
		repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), git)
		_, err = repo.Clone(ctx, &biz.CloneRepoInput{ID: orig.ID, Name: "fresh"})
		assert.Error(t, err)
	})
}

// TestRepoImpl_IsSimpleEnv 覆盖 isSimpleEnv 的边界分支：
// 空配置早退 true 与本地 yaml 解析成功直接返回。
func TestRepoImpl_IsSimpleEnv(t *testing.T) {
	ctx := context.TODO()
	repo := &repoImpl{}

	t.Run("empty config returns true", func(t *testing.T) {
		assert.True(t, repo.isSimpleEnv(ctx, nil))
		assert.True(t, repo.isSimpleEnv(ctx, &mars.Config{}))
		assert.True(t, repo.isSimpleEnv(ctx, &mars.Config{ConfigField: "config"}))
		assert.True(t, repo.isSimpleEnv(ctx, &mars.Config{LocalChartPath: "1|main|chart"}))
	})

	t.Run("values yaml resolves to scalar", func(t *testing.T) {
		assert.True(t, repo.isSimpleEnv(ctx, &mars.Config{ConfigField: "config", LocalChartPath: "1|main|chart", ValuesYaml: "config: config"}))
		assert.False(t, repo.isSimpleEnv(ctx, &mars.Config{ConfigField: "config", LocalChartPath: "1|main|chart", ValuesYaml: "config:\n  a: b"}))
	})
}

func TestRepoImpl_Import_MixedCreateAndUpdate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	// 预置已存在仓库：导入同名走更新（保留 ID、覆盖 enabled/description），异名走创建。
	created, err := repo.Create(context.TODO(), &biz.CreateRepoInput{Name: "exists", Enabled: true, Description: "old"})
	assert.Nil(t, err)

	createdCount, updatedCount, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{
		{Name: "exists", Enabled: false, Description: "new"},
		{Name: "fresh", Enabled: true, Description: "fresh desc"},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, createdCount)
	assert.Equal(t, 1, updatedCount)

	got, err := repo.Show(context.TODO(), created.ID)
	assert.Nil(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.False(t, got.Enabled)
	assert.Equal(t, "new", got.Description)

	fresh, err := repo.GetByName(context.TODO(), "fresh")
	assert.Nil(t, err)
	assert.True(t, fresh.Enabled)
	assert.Equal(t, "fresh desc", fresh.Description)
}

func TestRepoImpl_Import_WithMarsConfig(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	git := NewMockGitRepo(m)
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), git)

	// 新建条目带 mars_config + 需要 git：git 解析项目名/分支，config 完整穿越事务并重算 IsSimpleEnv。
	git.EXPECT().GetByProjectID(gomock.Any(), 100).Return(&biz.GitProject{DefaultBranch: "main", Name: "proj"}, nil)
	git.EXPECT().GetChartValuesYaml(gomock.Any(), "100|main|chart").Return(`config: config`, nil)

	created, updated, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{
		{Name: "app", Enabled: true, NeedGitRepo: true, GitProjectID: lo.ToPtr(int32(100)),
			MarsConfig: &mars.Config{ConfigField: "config", LocalChartPath: "100|main|chart"}},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, created)
	assert.Zero(t, updated)

	got, err := repo.GetByName(context.TODO(), "app")
	assert.Nil(t, err)
	assert.Equal(t, "proj", got.GitProjectName)
	assert.Equal(t, "main", got.DefaultBranch)
	assert.Equal(t, int32(100), got.GitProjectID)
	require.NotNil(t, got.MarsConfig)
	assert.Equal(t, "config", got.MarsConfig.ConfigField)
	assert.True(t, got.MarsConfig.IsSimpleEnv)
}

func TestRepoImpl_Import_UnrelatedReposUntouched(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	// 预置一个不在导入列表里的仓库：导入只应触碰 items 内的名字，其余数据保持原状。
	other, err := repo.Create(context.TODO(), &biz.CreateRepoInput{Name: "other", Enabled: true, Description: "keep me"})
	assert.Nil(t, err)

	createdCount, updatedCount, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{
		{Name: "app", Enabled: true, Description: "fresh"},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, createdCount)
	assert.Zero(t, updatedCount)

	got, err := repo.Show(context.TODO(), other.ID)
	assert.Nil(t, err)
	assert.Equal(t, other.ID, got.ID)
	assert.True(t, got.Enabled)
	assert.Equal(t, "keep me", got.Description)

	_, err = repo.GetByName(context.TODO(), "app")
	assert.Nil(t, err)
}

func TestRepoImpl_Import_PreservesProjectsOnUpdate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	// 预置已存在仓库，并挂一个关联项目：导入覆盖仓库标量字段，不得冲掉 projects 边。
	created, err := repo.Create(context.TODO(), &biz.CreateRepoInput{Name: "app", Enabled: true})
	assert.Nil(t, err)
	ns := createNamespace(entdb)
	proj := entdb.Project.Create().
		SetName("p1").
		SetCreator("tester").
		SetRepoID(created.ID).
		SetNamespaceID(ns.ID).
		SaveX(context.TODO())

	createdCount, updatedCount, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{
		{Name: "app", Enabled: false, Description: "overwritten"},
	})
	assert.Nil(t, err)
	assert.Zero(t, createdCount)
	assert.Equal(t, 1, updatedCount)

	show, err := repo.Show(context.TODO(), created.ID)
	assert.Nil(t, err)
	require.Len(t, show.Projects, 1)
	assert.Equal(t, proj.ID, show.Projects[0].ID)
	assert.Equal(t, created.ID, show.Projects[0].RepoID)
}

func TestRepoImpl_Import_SoftDeletedNameReused(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	// 同名仓库已软删除：GetByName 被 soft-delete interceptor 过滤 → 判定不存在，创建新仓库。
	deleted, err := repo.Create(context.TODO(), &biz.CreateRepoInput{Name: "app", Enabled: true})
	assert.Nil(t, err)
	assert.Nil(t, repo.Delete(context.TODO(), deleted.ID))

	createdCount, updatedCount, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{
		{Name: "app", Enabled: false},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, createdCount)
	assert.Zero(t, updatedCount)

	got, err := repo.GetByName(context.TODO(), "app")
	assert.Nil(t, err)
	assert.NotEqual(t, deleted.ID, got.ID)
	assert.False(t, got.Enabled)
}

func TestRepoImpl_Import_RollbackOnWriteFailure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	// 第一条合法、第二条 name 违反 ent 字段校验 → 第二条 Save 失败，整体回滚，第一条不落库。
	created, updated, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{
		{Name: "good"},
		{Name: "bad name"},
	})
	assert.NotNil(t, err)
	assert.Zero(t, created)
	assert.Zero(t, updated)

	_, err = repo.GetByName(context.TODO(), "good")
	s, _ := status.FromError(err)
	assert.Equal(t, "NotFound", s.Code().String())
}

func TestRepoImpl_Import_RollbackOnUpdateFailure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	git := NewMockGitRepo(m)
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), git)

	// 预置一个无需 git 关联的已存在仓库（create 不触发 git 调用）。
	created, err := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name: "app", Enabled: true, NeedGitRepo: false,
	})
	assert.Nil(t, err)

	// 导入第二条同名覆盖并启用 git：git 解析出的默认分支超长（> default_branch MaxLen(255)），
	// update Save 触发字段校验失败 → 整体回滚。
	git.EXPECT().GetByProjectID(gomock.Any(), 100).Return(
		&biz.GitProject{DefaultBranch: strings.Repeat("x", 256), Name: "proj"}, nil)

	createdCount, updatedCount, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{
		{Name: "good"},
		{Name: "app", Enabled: true, NeedGitRepo: true, GitProjectID: lo.ToPtr(int32(100))},
	})
	assert.NotNil(t, err)
	assert.Zero(t, createdCount)
	assert.Zero(t, updatedCount)

	// 原子性：第一条 "good" 未落库，且 "app" 保持原状（未被部分覆盖）。
	_, err = repo.GetByName(context.TODO(), "good")
	s, _ := status.FromError(err)
	assert.Equal(t, "NotFound", s.Code().String())
	got, err := repo.GetByName(context.TODO(), "app")
	assert.Nil(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.False(t, got.NeedGitRepo)
}

func TestRepoImpl_Import_GitResolutionError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	git := NewMockGitRepo(m)
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), git)

	// 第二条需要 git 仓库，git 解析失败 → 事务前中止，第一条不落库（原子性）。
	git.EXPECT().GetByProjectID(gomock.Any(), 1).Return(nil, assert.AnError)

	created, updated, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{
		{Name: "good"},
		{Name: "git-one", NeedGitRepo: true, GitProjectID: lo.ToPtr(int32(1))},
	})
	assert.NotNil(t, err)
	assert.Zero(t, created)
	assert.Zero(t, updated)

	_, err = repo.GetByName(context.TODO(), "good")
	s, _ := status.FromError(err)
	assert.Equal(t, "NotFound", s.Code().String())
}

func TestRepoImpl_Import_GetByNameError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: mustClosedDB(t)}), NewMockGitRepo(m))

	_, _, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{{Name: "app"}})
	assert.NotNil(t, err)
}

func TestRepoImpl_Import_UpdateClearsGitFields(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	git := NewMockGitRepo(m)
	git.EXPECT().GetByProjectID(gomock.Any(), 100).Return(&biz.GitProject{DefaultBranch: "main", Name: "proj"}, nil)
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), git)

	// 预置带 git 关联的已存在仓库。
	created, err := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name:         "app",
		Enabled:      true,
		NeedGitRepo:  true,
		GitProjectID: lo.ToPtr(int32(100)),
	})
	assert.Nil(t, err)
	assert.Equal(t, int32(100), created.GitProjectID)

	// 导入覆盖：NeedGitRepo=false → 清空 git 字段，Enabled 覆盖为 false。
	_, _, err = repo.Import(context.TODO(), []*biz.ImportRepoItem{
		{Name: "app", Enabled: false, NeedGitRepo: false},
	})
	assert.Nil(t, err)

	got, err := repo.Show(context.TODO(), created.ID)
	assert.Nil(t, err)
	assert.False(t, got.Enabled)
	assert.False(t, got.NeedGitRepo)
	assert.Equal(t, int32(0), got.GitProjectID)
	assert.Empty(t, got.GitProjectName)
	assert.Empty(t, got.DefaultBranch)
}

func TestRepoImpl_Import_RejectsDuplicateNameInDB(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	// 预置库内同名重复行（绕过业务查重直接插入，模拟历史脏数据/并发竞争产物）。
	for i := 0; i < 2; i++ {
		_, err := entdb.Repo.Create().SetName("dup").SetEnabled(false).Save(context.TODO())
		assert.Nil(t, err)
	}

	// 导入同名条目：整体拒绝（InvalidArgument），不做任何部分变更。
	created, updated, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{{Name: "dup", Enabled: true, Description: "v2"}})
	assert.NotNil(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Zero(t, created)
	assert.Zero(t, updated)

	// 库内两行保持原样，未被部分覆盖。
	cnt, err := entdb.Repo.Query().Where(repoq.Name("dup")).Count(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 2, cnt)
}

// ---- PreviewImport ----

func TestRepoImpl_PreviewImport_MixedCreateAndUpdate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	// 预置已存在仓库：干跑应判定「exists→更新、fresh→新建」，且不落库。
	created, err := repo.Create(context.TODO(), &biz.CreateRepoInput{Name: "exists", Enabled: true, Description: "old"})
	assert.Nil(t, err)

	createdCount, updatedCount, err := repo.PreviewImport(context.TODO(), []*biz.ImportRepoItem{
		{Name: "exists", Enabled: false, Description: "new"},
		{Name: "fresh", Enabled: true, Description: "fresh desc"},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, createdCount)
	assert.Equal(t, 1, updatedCount)

	// 干跑零副作用：exists 保持原值，fresh 不存在。
	got, err := repo.Show(context.TODO(), created.ID)
	assert.Nil(t, err)
	assert.True(t, got.Enabled)
	assert.Equal(t, "old", got.Description)

	_, err = repo.GetByName(context.TODO(), "fresh")
	s, _ := status.FromError(err)
	assert.Equal(t, "NotFound", s.Code().String())
}

func TestRepoImpl_PreviewImport_WithGitResolution(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	git := NewMockGitRepo(m)
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), git)

	// 干跑同样走 planImport：需要 git 仓库时解析项目名/分支，但解析结果不落库。
	git.EXPECT().GetByProjectID(gomock.Any(), 100).Return(&biz.GitProject{DefaultBranch: "main", Name: "proj"}, nil)

	created, updated, err := repo.PreviewImport(context.TODO(), []*biz.ImportRepoItem{
		{Name: "app", Enabled: true, NeedGitRepo: true, GitProjectID: lo.ToPtr(int32(100))},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, created)
	assert.Zero(t, updated)

	_, err = repo.GetByName(context.TODO(), "app")
	s, _ := status.FromError(err)
	assert.Equal(t, "NotFound", s.Code().String())
}

func TestRepoImpl_PreviewImport_GitResolutionError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	git := NewMockGitRepo(m)
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), git)

	git.EXPECT().GetByProjectID(gomock.Any(), 1).Return(nil, assert.AnError)

	created, updated, err := repo.PreviewImport(context.TODO(), []*biz.ImportRepoItem{
		{Name: "git-one", NeedGitRepo: true, GitProjectID: lo.ToPtr(int32(1))},
	})
	assert.NotNil(t, err)
	assert.Zero(t, created)
	assert.Zero(t, updated)
}

func TestRepoImpl_PreviewImport_GetByNameError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: mustClosedDB(t)}), NewMockGitRepo(m))

	_, _, err := repo.PreviewImport(context.TODO(), []*biz.ImportRepoItem{{Name: "app"}})
	assert.NotNil(t, err)
}
