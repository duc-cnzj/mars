package data

import (
	"context"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/projecthibernate"
	entrepo "github.com/duc-cnzj/mars/v6/internal/data/ent/repo"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHibernateWindowRepo_CRUD 覆盖全局休眠时间窗的增删改查：List 名称升序、
// GetByName 命中/未命中、Create、Update（含不存在）、Delete（幂等）。
func TestHibernateWindowRepo_CRUD(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	defer entdb.Close()
	repo := NewHibernateWindowRepo(NewDataImpl(&NewDataParams{DB: entdb}))

	// 空表。
	all, err := repo.List(context.TODO())
	assert.NoError(t, err)
	assert.Empty(t, all)

	// GetByName 未命中：返回 (nil, nil)。
	miss, err := repo.GetByName(context.TODO(), "夜间")
	assert.NoError(t, err)
	assert.Nil(t, miss)

	// 新建两条，验证按名称升序。
	night, err := repo.Create(context.TODO(), &biz.HibernateWindow{Name: "夜间", Start: "23:00", End: "06:00", Description: "过夜"})
	assert.NoError(t, err)
	assert.NotZero(t, night.ID)
	assert.Equal(t, "夜间", night.Name)
	assert.Equal(t, "23:00", night.Start)
	assert.Equal(t, "06:00", night.End)
	assert.Equal(t, "过夜", night.Description)
	assert.False(t, night.CreatedAt.IsZero())

	if _, err := repo.Create(context.TODO(), &biz.HibernateWindow{Name: "午休", Start: "12:00", End: "13:00"}); err != nil {
		t.Fatalf("create 午休: %v", err)
	}

	all, err = repo.List(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, all, 2)
	// 名称升序：中文按 unicode 码位，午(U+5348) < 夜(U+591C)。
	assert.Equal(t, "午休", all[0].Name)
	assert.Equal(t, "夜间", all[1].Name)

	// GetByName 命中。
	hit, err := repo.GetByName(context.TODO(), "夜间")
	assert.NoError(t, err)
	assert.NotNil(t, hit)
	assert.Equal(t, night.ID, hit.ID)

	// 重名：唯一索引冲突归类为 409。
	_, err = repo.Create(context.TODO(), &biz.HibernateWindow{Name: "夜间", Start: "01:00", End: "02:00"})
	assert.Error(t, err)

	// 更新字段。
	night.Start = "22:30"
	night.Description = "改过"
	if _, err := repo.Update(context.TODO(), night); err != nil {
		t.Fatalf("update: %v", err)
	}
	updated, err := repo.GetByName(context.TODO(), "夜间")
	assert.NoError(t, err)
	assert.Equal(t, "22:30", updated.Start)
	assert.Equal(t, "改过", updated.Description)

	// 更新不存在的记录：上抛错误。
	if _, err := repo.Update(context.TODO(), &biz.HibernateWindow{ID: 99999, Name: "幽灵", Start: "01:00", End: "02:00"}); err == nil {
		t.Fatal("更新不存在记录应报错")
	}

	// 删除后查不到；再次删除幂等不报错。
	assert.NoError(t, repo.Delete(context.TODO(), night.ID))
	gone, err := repo.GetByName(context.TODO(), "夜间")
	assert.NoError(t, err)
	assert.Nil(t, gone)
	assert.NoError(t, repo.Delete(context.TODO(), night.ID))
}

// TestToHibernateWindow_Nil 覆盖 nil 安全转换。
func TestToHibernateWindow_Nil(t *testing.T) {
	assert.Nil(t, toHibernateWindow(nil))
}

// TestHibernateWindowRepo_UpdateCascade 覆盖级联改名事务：
// 改名 + 重写引用旧名的仓库（计数正确、命中项改写、非命中项不动）；
// 无引用时迁移数为 0；目标名重复时整体回滚（窗口未改名、引用未改）；窗口不存在上抛。
func TestHibernateWindowRepo_UpdateCascade(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	defer entdb.Close()
	repo := NewHibernateWindowRepo(NewDataImpl(&NewDataParams{DB: entdb}))

	night, err := repo.Create(context.TODO(), &biz.HibernateWindow{Name: "夜间", Start: "23:00", End: "06:00"})
	assert.NoError(t, err)

	mkRepo := func(name string, windows []string, withHibernate bool) {
		cfg := &mars.Config{}
		if withHibernate {
			cfg.Hibernate = &mars.Hibernate{WindowNames: windows}
		}
		_, err := entdb.Repo.Create().SetName(name).SetMarsConfig(cfg).Save(context.TODO())
		assert.NoError(t, err)
	}
	mkRepo("a", []string{"夜间", "午休"}, true) // 命中（多窗里含旧名，另窗应保留）
	mkRepo("e", []string{"夜间"}, true)       // 命中
	mkRepo("b", []string{"午休"}, true)       // 非命中（别的窗）
	mkRepo("d", nil, false)                 // hibernate 为空：非命中

	windowNamesOf := func(name string) []string {
		e, err := entdb.Repo.Query().Where(entrepo.Name(name)).Only(context.TODO())
		assert.NoError(t, err)
		return e.MarsConfig.GetHibernate().GetWindowNames()
	}

	// 项目级选择也按名引用，级联改名须一并重写。p1 含旧名（应迁移并保留另窗），
	// p2 只是别窗（不动）。项目需真实存在（测试库 _fk=1 强制外键）。
	ns := createNamespace(entdb)
	p1 := createProject(entdb, ns.ID)
	p2 := createProject(entdb, ns.ID)
	_, err = entdb.ProjectHibernate.Create().SetProjectID(p1.ID).SetWindowNames([]string{"夜间", "午休"}).Save(context.TODO())
	assert.NoError(t, err)
	_, err = entdb.ProjectHibernate.Create().SetProjectID(p2.ID).SetWindowNames([]string{"午休"}).Save(context.TODO())
	assert.NoError(t, err)
	projectWindowsOf := func(id int) []string {
		e, err := entdb.ProjectHibernate.Query().Where(projecthibernate.ProjectID(id)).Only(context.TODO())
		assert.NoError(t, err)
		return e.WindowNames
	}

	// 级联改名：命中 2 个引用仓库 + 1 个项目，计数正确，命中项改写为新名（其余窗保留），
	// 非命中项不动。
	updated, migrated, err := repo.UpdateCascade(context.TODO(), &biz.HibernateWindow{
		ID: night.ID, Name: "深夜", Start: "22:00", End: "07:00",
	}, "夜间")
	assert.NoError(t, err)
	assert.Equal(t, 3, migrated)
	assert.Equal(t, night.ID, updated.ID)
	assert.Equal(t, "深夜", updated.Name)
	assert.Equal(t, "22:00", updated.Start)
	assert.Equal(t, []string{"深夜", "午休"}, windowNamesOf("a"))
	assert.Equal(t, []string{"深夜"}, windowNamesOf("e"))
	assert.Equal(t, []string{"午休"}, windowNamesOf("b"))
	assert.Empty(t, windowNamesOf("d"))
	assert.Equal(t, []string{"深夜", "午休"}, projectWindowsOf(p1.ID))
	assert.Equal(t, []string{"午休"}, projectWindowsOf(p2.ID))
	// 旧名不再命中，新名命中。
	if miss, err := repo.GetByName(context.TODO(), "夜间"); assert.NoError(t, err) {
		assert.Nil(t, miss)
	}

	// 无引用：迁移数为 0，仍改名。
	_, migrated, err = repo.UpdateCascade(context.TODO(), &biz.HibernateWindow{
		ID: night.ID, Name: "凌晨", Start: "01:00", End: "05:00",
	}, "不存在的旧名")
	assert.NoError(t, err)
	assert.Zero(t, migrated)

	// 目标名重复：撞唯一索引，整体回滚——窗口保持旧名、引用不被改写。
	if _, err := repo.Create(context.TODO(), &biz.HibernateWindow{Name: "午休2", Start: "08:00", End: "09:00"}); err != nil {
		t.Fatalf("create 午休2: %v", err)
	}
	_, _, err = repo.UpdateCascade(context.TODO(), &biz.HibernateWindow{
		ID: night.ID, Name: "午休2", Start: "01:00", End: "05:00",
	}, "凌晨")
	assert.Error(t, err)
	// 回滚证据：窗口仍叫「凌晨」，仓库与项目引用均未被改写。
	if kept, err := repo.GetByID(context.TODO(), night.ID); assert.NoError(t, err) {
		assert.Equal(t, "凌晨", kept.Name)
	}
	assert.Equal(t, []string{"深夜", "午休"}, windowNamesOf("a"))
	assert.Equal(t, []string{"深夜", "午休"}, projectWindowsOf(p1.ID))

	// 窗口不存在：上抛（交由 errs.Wrap 归为 404）。
	_, _, err = repo.UpdateCascade(context.TODO(), &biz.HibernateWindow{
		ID: 99999, Name: "幽灵", Start: "01:00", End: "02:00",
	}, "凌晨")
	assert.Error(t, err)
}

// TestHibernateWindowRepo_UpdateCascade_SoftDeletedExcluded 验证级联改名不吃进软删仓库：
// 引用旧名的仓库中，软删的不参与迁移（迁移数只算活跃引用），即 SoftDeleteMixin 拦截器在
// 事务内同样生效——避免给已删仓库写回 window_names（依赖 UpdateCascade 注释所声称的不变量）。
func TestHibernateWindowRepo_UpdateCascade_SoftDeletedExcluded(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	defer entdb.Close()
	repo := NewHibernateWindowRepo(NewDataImpl(&NewDataParams{DB: entdb}))

	night, err := repo.Create(context.TODO(), &biz.HibernateWindow{Name: "夜间-sd", Start: "23:00", End: "06:00"})
	require.NoError(t, err)

	mk := func(name string) int {
		e, err := entdb.Repo.Create().SetName(name).
			SetMarsConfig(&mars.Config{Hibernate: &mars.Hibernate{WindowNames: []string{"夜间-sd"}}}).
			Save(context.TODO())
		require.NoError(t, err)
		return e.ID
	}
	mk("active-sd")
	deadID := mk("dead-sd")

	// 软删其中一个引用仓库：级联改名应只迁移活跃引用（软删的由拦截器滤除）。
	require.NoError(t, entdb.Repo.DeleteOneID(deadID).Exec(context.TODO()))

	_, migrated, err := repo.UpdateCascade(context.TODO(), &biz.HibernateWindow{
		ID: night.ID, Name: "深夜-sd", Start: "22:00", End: "07:00",
	}, "夜间-sd")
	require.NoError(t, err)
	assert.Equal(t, 1, migrated)

	// 活跃引用已改写为新名。
	e, err := entdb.Repo.Query().Where(entrepo.Name("active-sd")).Only(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, []string{"深夜-sd"}, e.MarsConfig.GetHibernate().GetWindowNames())
}

// TestHibernateWindowRepo_UpdateCascade_TxErrors 用故障注入覆盖事务内的四条错误分支：
// 仓库引用查询/更新失败、项目引用查询/更新失败——四者都必须整体回滚（窗口未改名、仓库与项目
// 引用均未迁移），证明半迁移不会发生。故障注入的操作序号由 failDriver 的 Query/Exec 计数决定
// （窗口更新=exec#1，窗口回读=query#1，仓库查询=query#2，仓库更新=exec#2，项目查询=query#4，
// 项目更新=exec#3）。名称加后缀避免与共享内存库的其他用例撞名。
func TestHibernateWindowRepo_UpdateCascade_TxErrors(t *testing.T) {
	cases := []struct {
		name           string
		qAfter, eAfter int32
	}{
		{"仓库引用查询失败", 1, -1},
		{"仓库引用更新失败", -1, 1},
		{"项目引用查询失败", 3, -1},
		{"项目引用更新失败", -1, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, fd := newFailDB(t, tc.qAfter, tc.eAfter)
			repo := NewHibernateWindowRepo(NewDataImpl(&NewDataParams{DB: client}))

			night, err := repo.Create(context.TODO(), &biz.HibernateWindow{Name: "夜间-txerr", Start: "23:00", End: "06:00"})
			require.NoError(t, err)
			_, err = client.Repo.Create().SetName("a-txerr").
				SetMarsConfig(&mars.Config{Hibernate: &mars.Hibernate{WindowNames: []string{"夜间-txerr"}}}).
				Save(context.TODO())
			require.NoError(t, err)
			// 项目级引用（真实项目以满足测试库外键约束）。
			ns := createNamespace(client)
			p := createProject(client, ns.ID)
			_, err = client.ProjectHibernate.Create().SetProjectID(p.ID).SetWindowNames([]string{"夜间-txerr"}).Save(context.TODO())
			require.NoError(t, err)

			fd.Arm()
			_, _, err = repo.UpdateCascade(context.TODO(), &biz.HibernateWindow{
				ID: night.ID, Name: "深夜-txerr", Start: "22:00", End: "07:00",
			}, "夜间-txerr")
			fd.Disarm()
			assert.Error(t, err)

			// 回滚证据：窗口仍叫「夜间-txerr」，仓库与项目引用均仍指向「夜间-txerr」。
			kept, err := repo.GetByID(context.TODO(), night.ID)
			require.NoError(t, err)
			assert.Equal(t, "夜间-txerr", kept.Name)
			e, err := client.Repo.Query().Where(entrepo.Name("a-txerr")).Only(context.TODO())
			require.NoError(t, err)
			assert.Equal(t, []string{"夜间-txerr"}, e.MarsConfig.GetHibernate().GetWindowNames())
			ph, err := client.ProjectHibernate.Query().Where(projecthibernate.ProjectID(p.ID)).Only(context.TODO())
			require.NoError(t, err)
			assert.Equal(t, []string{"夜间-txerr"}, ph.WindowNames)
		})
	}
}

// TestHibernateWindowRepo_GetByID 覆盖命中、未命中（返回 nil,nil）与 DB 故障上抛（不误判为未命中）。
func TestHibernateWindowRepo_GetByID(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	repo := NewHibernateWindowRepo(NewDataImpl(&NewDataParams{DB: entdb}))

	w, err := repo.Create(context.TODO(), &biz.HibernateWindow{Name: "夜间", Start: "23:00", End: "06:00"})
	assert.NoError(t, err)

	got, err := repo.GetByID(context.TODO(), w.ID)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, w.ID, got.ID)
	assert.Equal(t, "夜间", got.Name)

	// 未命中：返回 (nil, nil)。
	miss, err := repo.GetByID(context.TODO(), 99999)
	assert.NoError(t, err)
	assert.Nil(t, miss)

	// DB 故障（非 NotFound）上抛，而非误判为未命中。
	entdb.Close()
	_, err = repo.GetByID(context.TODO(), w.ID)
	assert.Error(t, err)
}

// TestHibernateWindowRepo_ReposReferencing 覆盖按 window_names 过滤与升序：命中多条（含多窗里
// 命中其一）、非命中、无 mars_config、hibernate 为空均不被误纳；无引用返回空；DB 故障上抛。
func TestHibernateWindowRepo_ReposReferencing(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	repo := NewHibernateWindowRepo(NewDataImpl(&NewDataParams{DB: entdb}))

	mk := func(name string, cfg *mars.Config) {
		b := entdb.Repo.Create().SetName(name)
		if cfg != nil {
			b = b.SetMarsConfig(cfg)
		}
		_, err := b.Save(context.TODO())
		assert.NoError(t, err)
	}
	mk("b", &mars.Config{Hibernate: &mars.Hibernate{WindowNames: []string{"午休"}}})       // 非命中
	mk("a", &mars.Config{Hibernate: &mars.Hibernate{WindowNames: []string{"午休", "夜间"}}}) // 命中（多窗含目标）
	mk("c", nil)                                                                         // 无 mars_config
	mk("d", &mars.Config{Hibernate: nil})                                                // hibernate 为空
	mk("e", &mars.Config{Hibernate: &mars.Hibernate{WindowNames: []string{"夜间"}}})       // 命中

	refs, err := repo.ReposReferencing(context.TODO(), "夜间")
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "e"}, refs)

	// 无任何仓库引用：返回空。
	refs, err = repo.ReposReferencing(context.TODO(), "不存在")
	assert.NoError(t, err)
	assert.Empty(t, refs)

	// DB 故障上抛。
	entdb.Close()
	_, err = repo.ReposReferencing(context.TODO(), "夜间")
	assert.Error(t, err)
}

// TestHibernateWindowRepo_ProjectsReferencing 覆盖按 project_hibernates.window_names 过滤与
// ID 升序：命中（含多窗里命中其一）、非命中、空选择均正确处理；**禁用行（Enabled=false）即便命中
// 窗名也不计入**（幽灵引用锁修复：禁用项目不参与对账，其窗名只是保留的选择草稿）；无引用返回空；
// DB 故障上抛。
func TestHibernateWindowRepo_ProjectsReferencing(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	repo := NewHibernateWindowRepo(NewDataImpl(&NewDataParams{DB: entdb}))

	ns := createNamespace(entdb)
	mk := func(enabled bool, windows []string) int {
		p := createProject(entdb, ns.ID)
		_, err := entdb.ProjectHibernate.Create().SetProjectID(p.ID).SetEnabled(enabled).SetWindowNames(windows).Save(context.TODO())
		require.NoError(t, err)
		return p.ID
	}
	mk(true, []string{"午休"})               // 已开、非命中
	hitA := mk(true, []string{"午休", "夜间"}) // 已开、命中（多窗含目标）
	mk(true, nil)                          // 已开、空选择：不按名引用任何窗
	hitB := mk(true, []string{"夜间"})       // 已开、命中
	mk(false, []string{"夜间"})              // 禁用、命中窗名：开关未开不参与对账 → 不算引用

	refs, err := repo.ProjectsReferencing(context.TODO(), "夜间")
	assert.NoError(t, err)
	assert.Equal(t, []int{hitA, hitB}, refs)

	// 无任何项目引用：返回空。
	refs, err = repo.ProjectsReferencing(context.TODO(), "不存在")
	assert.NoError(t, err)
	assert.Empty(t, refs)

	// DB 故障上抛。
	entdb.Close()
	_, err = repo.ProjectsReferencing(context.TODO(), "夜间")
	assert.Error(t, err)
}

// TestHibernateWindowRepo_ProjectReferencingRepoIDs 覆盖「引用该窗的项目 ID → 所属仓库 ID」映射：
// 命中项目映射到各自仓库、项目无仓库时 repo_id 为 0、软删项目的悬留休眠行不计入、
// **禁用行（Enabled=false）即便命中窗名也不计入**（同款幽灵引用锁修复）、无引用返回空映射、DB 故障上抛。
func TestHibernateWindowRepo_ProjectReferencingRepoIDs(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	defer entdb.Close()
	repo := NewHibernateWindowRepo(NewDataImpl(&NewDataParams{DB: entdb}))

	ns := createNamespace(entdb)
	r1 := entdb.Repo.Create().SetName("r1").SaveX(context.TODO())
	r2 := entdb.Repo.Create().SetName("r2").SaveX(context.TODO())

	// mk 造「带休眠行」的项目；repoID 为 0 时不设 repo_id（存 NULL，读回 0）。
	mk := func(repoID int, enabled bool, windows []string) int {
		b := entdb.Project.Create().
			SetGitBranch("").SetGitCommit("").SetConfig("").SetGitProjectID(1).
			SetCreator("").SetName("p").SetNamespaceID(ns.ID)
		if repoID != 0 {
			b = b.SetRepoID(repoID)
		}
		p := b.SaveX(context.TODO())
		_, e := entdb.ProjectHibernate.Create().SetProjectID(p.ID).SetEnabled(enabled).SetWindowNames(windows).Save(context.TODO())
		require.NoError(t, e)
		return p.ID
	}
	hitA := mk(r1.ID, true, []string{"午休", "夜间"}) // 命中 → r1
	hitB := mk(r2.ID, true, []string{"夜间"})       // 命中 → r2
	mk(r1.ID, true, []string{"午休"})               // 非命中（不含目标窗）
	noRepo := mk(0, true, []string{"夜间"})         // 命中但无仓库 → repo_id 0
	mk(r1.ID, false, []string{"夜间"})              // 禁用、命中窗名：开关未开不参与对账 → 不算引用

	// 软删项目：其悬留的休眠行不应计入（Project 查询经拦截器过滤软删行）。
	ghost := mk(r1.ID, true, []string{"夜间"})
	entdb.Project.UpdateOneID(ghost).SetDeletedAt(time.Now()).SaveX(mixin.SkipSoftDelete(context.TODO()))

	got, err := repo.ProjectReferencingRepoIDs(context.TODO(), "夜间")
	assert.NoError(t, err)
	assert.Equal(t, map[int]int{hitA: r1.ID, hitB: r2.ID, noRepo: 0}, got)

	// 无任何项目引用：返回空映射。
	empty, err := repo.ProjectReferencingRepoIDs(context.TODO(), "不存在")
	assert.NoError(t, err)
	assert.Empty(t, empty)

	// DB 故障上抛。
	entdb.Close()
	_, err = repo.ProjectReferencingRepoIDs(context.TODO(), "夜间")
	assert.Error(t, err)
}

// TestHibernateWindowRepo_ProjectReferencingRepoIDs_SecondQueryFails 覆盖「按 ID 批量取仓库归属」
// 那一步的 DB 故障上抛：首个 ProjectHibernate 查询成功、其后 Project 查询被注入失败。
func TestHibernateWindowRepo_ProjectReferencingRepoIDs_SecondQueryFails(t *testing.T) {
	client, fd := newFailDB(t, 1, -1)
	repo := NewHibernateWindowRepo(NewDataImpl(&NewDataParams{DB: client}))

	ns := createNamespace(client)
	p := createProject(client, ns.ID)
	_, err := client.ProjectHibernate.Create().SetProjectID(p.ID).SetEnabled(true).SetWindowNames([]string{"夜间-repofail"}).Save(context.TODO())
	require.NoError(t, err)

	fd.Arm()
	_, err = repo.ProjectReferencingRepoIDs(context.TODO(), "夜间-repofail")
	fd.Disarm()
	assert.Error(t, err)
}

// TestHibernateWindowRepo_WindowRepoRefs 覆盖单趟扫描的「窗名 → 引用仓库名」分组映射：
// 多窗仓库同时计入各所属组、每组按仓库名升序；无 mars_config 与 hibernate 为空的仓库均不计入；
// 无任何引用返回空映射；DB 故障上抛。
func TestHibernateWindowRepo_WindowRepoRefs(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	defer entdb.Close()
	repo := NewHibernateWindowRepo(NewDataImpl(&NewDataParams{DB: entdb}))

	mk := func(name string, cfg *mars.Config) {
		b := entdb.Repo.Create().SetName(name)
		if cfg != nil {
			b = b.SetMarsConfig(cfg)
		}
		_, err := b.Save(context.TODO())
		assert.NoError(t, err)
	}
	mk("b", &mars.Config{Hibernate: &mars.Hibernate{WindowNames: []string{"午休"}}})
	mk("a", &mars.Config{Hibernate: &mars.Hibernate{WindowNames: []string{"午休", "夜间"}}})
	mk("c", nil)                          // 无 mars_config：不计入
	mk("d", &mars.Config{Hibernate: nil}) // hibernate 为空：不计入
	mk("e", &mars.Config{Hibernate: &mars.Hibernate{WindowNames: []string{"夜间"}}})
	mk("c2", &mars.Config{Hibernate: &mars.Hibernate{WindowNames: []string{"夜间", "午休"}}})

	refs, err := repo.WindowRepoRefs(context.TODO())
	assert.NoError(t, err)
	// 按窗名分组，每组仓库名升序；多窗仓库同时出现在两个组里。
	assert.Equal(t, map[string][]string{
		"午休": {"a", "b", "c2"},
		"夜间": {"a", "c2", "e"},
	}, refs)

	// 单名薄封装 ReposReferencing 与映射查表等价。
	refsNight, err := repo.ReposReferencing(context.TODO(), "夜间")
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "c2", "e"}, refsNight)

	// 无任何引用：返回空映射。
	empty, err := repo.ReposReferencing(context.TODO(), "不存在")
	assert.NoError(t, err)
	assert.Empty(t, empty)

	// DB 故障上抛。
	entdb.Close()
	_, err = repo.WindowRepoRefs(context.TODO())
	assert.Error(t, err)
}

// TestHibernateWindowRepo_Errors 覆盖各方法在 DB 故障时的错误上抛（关闭连接触发）。
func TestHibernateWindowRepo_Errors(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	repo := NewHibernateWindowRepo(NewDataImpl(&NewDataParams{DB: entdb}))
	entdb.Close()

	_, err = repo.List(context.TODO())
	assert.Error(t, err)
	// GetByName 查询失败（非 NotFound）时上抛，而非误判为未命中。
	_, err = repo.GetByName(context.TODO(), "夜间")
	assert.Error(t, err)
	_, err = repo.Create(context.TODO(), &biz.HibernateWindow{Name: "夜间", Start: "23:00", End: "06:00"})
	assert.Error(t, err)
	_, err = repo.Update(context.TODO(), &biz.HibernateWindow{ID: 1, Name: "夜间", Start: "23:00", End: "06:00"})
	assert.Error(t, err)
	assert.Error(t, repo.Delete(context.TODO(), 1))
}
