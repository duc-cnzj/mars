package data

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
