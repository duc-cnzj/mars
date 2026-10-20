package data

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	repoq "github.com/duc-cnzj/mars/v6/internal/data/ent/repo"
	"github.com/duc-cnzj/mars/v6/internal/errs"
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

	_, err := repo.Create(context.TODO(), &biz.CreateRepoInput{Name: "plain", Enabled: true})
	assert.Nil(t, err)
	_, err = repo.Create(context.TODO(), &biz.CreateRepoInput{Name: "agent", Enabled: true, AIAgent: true})
	assert.Nil(t, err)

	// 不传 AIAgent 时不过滤，返回全部仓库。
	res, pag, err := repo.List(context.TODO(), &biz.ListRepoRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, pag)
	assert.Len(t, res, 2)

	// AIAgent=true 只返回 AI Agent 仓库。
	res, _, err = repo.List(context.TODO(), &biz.ListRepoRequest{AIAgent: lo.ToPtr(true)})
	assert.Nil(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "agent", res[0].Name)
	assert.True(t, res[0].AIAgent)

	// AIAgent=false 只返回非 AI Agent 仓库。
	res, _, err = repo.List(context.TODO(), &biz.ListRepoRequest{AIAgent: lo.ToPtr(false)})
	assert.Nil(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "plain", res[0].Name)
	assert.False(t, res[0].AIAgent)
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

// newRepoTestKit 构造 repo 测试的最小装配（SQLite + git mock），供 AI Agent 用例复用。
func newRepoTestKit(t *testing.T) (*repoImpl, *MockGitRepo) {
	t.Helper()
	m := gomock.NewController(t)
	entdb, _ := NewSqliteDB()
	t.Cleanup(func() {
		entdb.Close()
		m.Finish()
	})
	git := NewMockGitRepo(m)
	return NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), git).(*repoImpl), git
}

// TestRepoImpl_GetAIAgent 覆盖未配置（NotFound）与配置后查询的语义。
func TestRepoImpl_GetAIAgent(t *testing.T) {
	repo, _ := newRepoTestKit(t)
	ctx := context.TODO()

	_, err := repo.GetAIAgent(ctx)
	assert.True(t, errs.IsNotFound(err), "未配置 AI Agent 仓库时应返回 NotFound")

	created, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "app", AIAgent: true})
	require.NoError(t, err)
	got, err := repo.GetAIAgent(ctx)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.True(t, got.AIAgent)
}

// TestRepoImpl_Create_SwitchesAIAgentSlot 覆盖新建 AI 仓库时顶替旧槽位，且含软删行。
func TestRepoImpl_Create_SwitchesAIAgentSlot(t *testing.T) {
	repo, _ := newRepoTestKit(t)
	ctx := context.TODO()

	first, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "first", AIAgent: true})
	require.NoError(t, err)
	second, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "second", AIAgent: true})
	require.NoError(t, err)

	got, err := repo.GetAIAgent(ctx)
	require.NoError(t, err)
	assert.Equal(t, second.ID, got.ID, "新建 AI 仓库应顶替旧槽位")
	old, err := repo.Get(ctx, first.ID)
	require.NoError(t, err)
	assert.False(t, old.AIAgent, "旧 AI 仓库标记应被清除")

	// 软删行仍占用唯一索引：删除后新建 AI 仓库必须先把软删行槽位清掉才能成功。
	require.NoError(t, repo.Delete(ctx, second.ID))
	third, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "third", AIAgent: true})
	require.NoError(t, err, "软删行的 AI 槽位必须被清理，否则唯一索引冲突")
	got, err = repo.GetAIAgent(ctx)
	require.NoError(t, err)
	assert.Equal(t, third.ID, got.ID)
}

// TestRepoImpl_Update_SwitchesAIAgentSlot 覆盖更新时顶替槽位，以及取消 AI 标记后查询 NotFound。
func TestRepoImpl_Update_SwitchesAIAgentSlot(t *testing.T) {
	repo, _ := newRepoTestKit(t)
	ctx := context.TODO()

	first, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "first", AIAgent: true})
	require.NoError(t, err)
	second, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "second"})
	require.NoError(t, err)

	_, err = repo.Update(ctx, &biz.UpdateRepoInput{ID: int32(second.ID), Name: "second", AIAgent: lo.ToPtr(true)})
	require.NoError(t, err)
	got, err := repo.GetAIAgent(ctx)
	require.NoError(t, err)
	assert.Equal(t, second.ID, got.ID)
	old, err := repo.Get(ctx, first.ID)
	require.NoError(t, err)
	assert.False(t, old.AIAgent)

	_, err = repo.Update(ctx, &biz.UpdateRepoInput{ID: int32(second.ID), Name: "second", AIAgent: lo.ToPtr(false)})
	require.NoError(t, err)
	_, err = repo.GetAIAgent(ctx)
	assert.True(t, errs.IsNotFound(err), "取消 AI 标记后应查不到 AI Agent 仓库")
}

// TestRepoImpl_Update_KeepsAIAgentWhenUnspecified 旧客户端更新时未传 ai_agent（nil），
// 必须保持已有标记与唯一槽位——这是 optional 字段的核心兼容语义。
func TestRepoImpl_Update_KeepsAIAgentWhenUnspecified(t *testing.T) {
	repo, _ := newRepoTestKit(t)
	ctx := context.TODO()

	created, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "agent", AIAgent: true})
	require.NoError(t, err)

	got, err := repo.Update(ctx, &biz.UpdateRepoInput{ID: int32(created.ID), Name: "agent"})
	require.NoError(t, err)
	assert.True(t, got.AIAgent)
	agent, err := repo.GetAIAgent(ctx)
	require.NoError(t, err)
	assert.Equal(t, created.ID, agent.ID)
}

// TestRepoImpl_Clone_DoesNotCopyAIAgent 覆盖克隆不继承 AI 标记（克隆出的仓库只是普通仓库）。
func TestRepoImpl_Clone_DoesNotCopyAIAgent(t *testing.T) {
	repo, _ := newRepoTestKit(t)
	ctx := context.TODO()

	src, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "src", AIAgent: true})
	require.NoError(t, err)
	clone, err := repo.Clone(ctx, &biz.CloneRepoInput{ID: src.ID, Name: "clone"})
	require.NoError(t, err)
	assert.False(t, clone.AIAgent, "克隆不应继承 AI Agent 标记")

	got, err := repo.GetAIAgent(ctx)
	require.NoError(t, err)
	assert.Equal(t, src.ID, got.ID)
}

// TestRepoImpl_Import_AIAgentUniqueness 覆盖导入时的槽位顶替与取消。
func TestRepoImpl_Import_AIAgentUniqueness(t *testing.T) {
	repo, _ := newRepoTestKit(t)
	ctx := context.TODO()

	first, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "first", AIAgent: true})
	require.NoError(t, err)
	second, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "second"})
	require.NoError(t, err)

	created, updated, err := repo.Import(ctx, []*biz.ImportRepoItem{{Name: "second", AIAgent: true}})
	require.NoError(t, err)
	assert.Zero(t, created)
	assert.Equal(t, 1, updated)
	got, err := repo.GetAIAgent(ctx)
	require.NoError(t, err)
	assert.Equal(t, second.ID, got.ID)
	old, err := repo.Get(ctx, first.ID)
	require.NoError(t, err)
	assert.False(t, old.AIAgent)

	// 导入条目不带 AI 标记即视为取消（导出/导入 round-trip 语义）。
	_, _, err = repo.Import(ctx, []*biz.ImportRepoItem{{Name: "second"}})
	require.NoError(t, err)
	_, err = repo.GetAIAgent(ctx)
	assert.True(t, errs.IsNotFound(err))
}

// installAIAgentSlotClearTrigger 安装一条仅拦「清除 AI 槽位」的 SQLite 触发器：
// 让 clearOtherAIAgentSlots 确定性失败，覆盖各写入路径的事务内错误分支。
func installAIAgentSlotClearTrigger(t *testing.T) {
	t.Helper()
	// ent.Client 不暴露裸 SQL 执行口：另开一条到同一 shared-cache 内存库的连接装触发器。
	sqlDB, err := sql.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1&loc=Local")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	_, err = sqlDB.ExecContext(context.TODO(), `CREATE TRIGGER fail_ai_slot_clear
BEFORE UPDATE OF ai_agent_key ON repos
WHEN OLD.ai_agent_key IS NOT NULL AND NEW.ai_agent_key IS NULL
BEGIN SELECT RAISE(ABORT, 'slot clear boom'); END;`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = sqlDB.ExecContext(context.TODO(), `DROP TRIGGER IF EXISTS fail_ai_slot_clear`)
	})
}

// TestRepoImpl_AIAgentSlotClearErrors 覆盖 Create/Update/Import 三条写入路径中
// 「清理旧 AI 槽位失败 → 整体回滚并上抛错误」的分支。
func TestRepoImpl_AIAgentSlotClearErrors(t *testing.T) {
	ctx := context.TODO()

	t.Run("create", func(t *testing.T) {
		repo, _ := newRepoTestKit(t)
		_, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "first", AIAgent: true})
		require.NoError(t, err)
		installAIAgentSlotClearTrigger(t)

		_, err = repo.Create(ctx, &biz.CreateRepoInput{Name: "second", AIAgent: true})
		assert.ErrorContains(t, err, "slot clear boom")
	})

	t.Run("update", func(t *testing.T) {
		repo, _ := newRepoTestKit(t)
		_, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "first", AIAgent: true})
		require.NoError(t, err)
		second, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "second"})
		require.NoError(t, err)
		installAIAgentSlotClearTrigger(t)

		_, err = repo.Update(ctx, &biz.UpdateRepoInput{ID: int32(second.ID), Name: "second", AIAgent: lo.ToPtr(true)})
		assert.ErrorContains(t, err, "slot clear boom")
	})

	t.Run("import create", func(t *testing.T) {
		repo, _ := newRepoTestKit(t)
		_, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "first", AIAgent: true})
		require.NoError(t, err)
		installAIAgentSlotClearTrigger(t)

		_, _, err = repo.Import(ctx, []*biz.ImportRepoItem{{Name: "second", AIAgent: true}})
		assert.ErrorContains(t, err, "slot clear boom")
	})

	t.Run("import update", func(t *testing.T) {
		repo, _ := newRepoTestKit(t)
		_, err := repo.Create(ctx, &biz.CreateRepoInput{Name: "first", AIAgent: true})
		require.NoError(t, err)
		_, err = repo.Create(ctx, &biz.CreateRepoInput{Name: "second"})
		require.NoError(t, err)
		installAIAgentSlotClearTrigger(t)

		_, _, err = repo.Import(ctx, []*biz.ImportRepoItem{{Name: "second", AIAgent: true}})
		assert.ErrorContains(t, err, "slot clear boom")
	})
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

// 导入暂不支持定时休眠：新建条目即便带了 hibernate 配置，落库也应被抹掉（其余字段保留）。
func TestRepoImpl_Import_StripsHibernateOnCreate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	created, updated, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{
		{Name: "app", Enabled: true, MarsConfig: &mars.Config{
			ConfigField: "cfg",
			Hibernate:   &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"replicas"}, WindowNames: []string{"夜间"}},
		}},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, created)
	assert.Zero(t, updated)

	got, err := repo.GetByName(context.TODO(), "app")
	assert.Nil(t, err)
	require.NotNil(t, got.MarsConfig)
	assert.Nil(t, got.MarsConfig.Hibernate, "新建导入不应带入 hibernate")
	assert.Equal(t, "cfg", got.MarsConfig.ConfigField, "其余配置字段应保留")
}

// 导入暂不支持定时休眠：更新条目不带 hibernate 时，库内既有 hibernate 应原样保留（不被清空）。
func TestRepoImpl_Import_PreservesHibernateOnUpdate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	existing, err := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name: "app", Enabled: true,
		MarsConfig: &mars.Config{Hibernate: &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"replicas"}, WindowNames: []string{"夜间"}}},
	})
	assert.Nil(t, err)

	// 导入文件由导出产生、已不含 hibernate，更新后库内 hibernate 应保持不变。
	created, updated, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{
		{Name: "app", Enabled: false, MarsConfig: &mars.Config{ConfigField: "cfg"}},
	})
	assert.Nil(t, err)
	assert.Zero(t, created)
	assert.Equal(t, 1, updated)

	got, err := repo.Show(context.TODO(), existing.ID)
	assert.Nil(t, err)
	require.NotNil(t, got.MarsConfig)
	require.NotNil(t, got.MarsConfig.Hibernate, "更新导入应保留库内既有 hibernate")
	assert.Equal(t, []string{"夜间"}, got.MarsConfig.Hibernate.WindowNames)
	assert.Equal(t, "cfg", got.MarsConfig.ConfigField)
}

// 边界：导入条目不带 mars_config（文件里没有配置）覆盖已配 hibernate 的仓库——
// 其余配置按空处理，但库内既有 hibernate 仍应保留。
func TestRepoImpl_Import_PreservesHibernateWhenConfigNil(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	existing, err := repo.Create(context.TODO(), &biz.CreateRepoInput{
		Name: "app", Enabled: true,
		MarsConfig: &mars.Config{Hibernate: &mars.Hibernate{Enabled: true, WindowNames: []string{"夜间"}}},
	})
	assert.Nil(t, err)

	_, updated, err := repo.Import(context.TODO(), []*biz.ImportRepoItem{{Name: "app", Enabled: true}})
	assert.Nil(t, err)
	assert.Equal(t, 1, updated)

	got, err := repo.Show(context.TODO(), existing.ID)
	assert.Nil(t, err)
	require.NotNil(t, got.MarsConfig)
	require.NotNil(t, got.MarsConfig.Hibernate, "config 为空时仍应保留库内 hibernate")
	assert.Equal(t, []string{"夜间"}, got.MarsConfig.Hibernate.WindowNames)
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

// TestRepoImpl_BatchUpdateMarsConfig 覆盖批量覆盖 mars_config 的事务语义：全部成功时逐个写入；
// 含不存在 id 时错误归类为 404 且整体回滚（存在的仓库也不被覆盖）；事务内写失败（故障注入）时
// 同样整体回滚。该路径是「一键配置时间窗到多仓库」的落库口，只写配置列、不触发 git 解析。
func TestRepoImpl_BatchUpdateMarsConfig(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewRepo(NewDataImpl(&NewDataParams{DB: entdb}), NewMockGitRepo(m))

	mk := func(name string, cfg *mars.Config) int {
		b := entdb.Repo.Create().SetName(name)
		if cfg != nil {
			b = b.SetMarsConfig(cfg)
		}
		e, err := b.Save(context.TODO())
		require.NoError(t, err)
		return e.ID
	}
	cfgOf := func(id int) *mars.Config {
		e, err := entdb.Repo.Get(context.TODO(), id)
		require.NoError(t, err)
		return e.MarsConfig
	}

	a := mk("a", nil)
	b := mk("b", nil)

	// 全部成功：两个仓库的 mars_config 都被覆盖。
	err := repo.BatchUpdateMarsConfig(context.TODO(), map[int]*mars.Config{
		a: {Hibernate: &mars.Hibernate{WindowNames: []string{"夜间"}}},
		b: {Hibernate: &mars.Hibernate{WindowNames: []string{"午休"}}},
	})
	assert.Nil(t, err)
	assert.Equal(t, []string{"夜间"}, cfgOf(a).GetHibernate().GetWindowNames())
	assert.Equal(t, []string{"午休"}, cfgOf(b).GetHibernate().GetWindowNames())

	// 含不存在 id：错误归类为 404，且整体回滚——a 保持上一次的「夜间」未被本次覆盖。
	err = repo.BatchUpdateMarsConfig(context.TODO(), map[int]*mars.Config{
		a:     {Hibernate: &mars.Hibernate{WindowNames: []string{"午休"}}},
		99999: {Hibernate: &mars.Hibernate{WindowNames: []string{"幽灵"}}},
	})
	assert.NotNil(t, err)
	s, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, s.Code())
	assert.Equal(t, []string{"夜间"}, cfgOf(a).GetHibernate().GetWindowNames())

	// 事务内写失败（故障注入 exec#1）：整体回滚，配置未改动。
	client, fd := newFailDB(t, -1, 0)
	pre := client.Repo.Create()
	target, err := pre.SetName("tx-fail").Save(context.TODO())
	require.NoError(t, err)
	fd.Arm()
	err = NewRepo(NewDataImpl(&NewDataParams{DB: client}), NewMockGitRepo(m)).BatchUpdateMarsConfig(
		context.TODO(),
		map[int]*mars.Config{target.ID: {Hibernate: &mars.Hibernate{WindowNames: []string{"夜间-tx"}}}},
	)
	fd.Disarm()
	assert.NotNil(t, err)
	got, err := client.Repo.Get(context.TODO(), target.ID)
	require.NoError(t, err)
	assert.Nil(t, got.MarsConfig.GetHibernate())
}
