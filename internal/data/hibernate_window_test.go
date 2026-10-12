package data

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	entrepo "github.com/duc-cnzj/mars/v6/internal/data/ent/repo"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHibernateWindowRepo_CRUD 覆盖全局休眠时间窗的增删改查：List 名称升序、
// GetByName 命中/未命中、Create、Update（含不存在）、Delete（幂等）。
func TestHibernateWindowRepo_CRUD(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	defer entdb.Close()
	repo := NewHibernateWindowRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: entdb}))

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
	repo := NewHibernateWindowRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: entdb}))

	night, err := repo.Create(context.TODO(), &biz.HibernateWindow{Name: "夜间", Start: "23:00", End: "06:00"})
	assert.NoError(t, err)

	mkRepo := func(name, window string, withHibernate bool) {
		cfg := &mars.Config{}
		if withHibernate {
			cfg.Hibernate = &mars.Hibernate{WindowName: window}
		}
		_, err := entdb.Repo.Create().SetName(name).SetMarsConfig(cfg).Save(context.TODO())
		assert.NoError(t, err)
	}
	mkRepo("a", "夜间", true) // 命中
	mkRepo("e", "夜间", true) // 命中
	mkRepo("b", "午休", true) // 非命中（别的窗）
	mkRepo("d", "", false)  // hibernate 为空：非命中

	windowNameOf := func(name string) string {
		e, err := entdb.Repo.Query().Where(entrepo.Name(name)).Only(context.TODO())
		assert.NoError(t, err)
		return e.MarsConfig.GetHibernate().GetWindowName()
	}

	// 级联改名：命中 2 个引用仓库，计数正确，命中项改写为新名，非命中项不动。
	updated, migrated, err := repo.UpdateCascade(context.TODO(), &biz.HibernateWindow{
		ID: night.ID, Name: "深夜", Start: "22:00", End: "07:00",
	}, "夜间")
	assert.NoError(t, err)
	assert.Equal(t, 2, migrated)
	assert.Equal(t, night.ID, updated.ID)
	assert.Equal(t, "深夜", updated.Name)
	assert.Equal(t, "22:00", updated.Start)
	assert.Equal(t, "深夜", windowNameOf("a"))
	assert.Equal(t, "深夜", windowNameOf("e"))
	assert.Equal(t, "午休", windowNameOf("b"))
	assert.Equal(t, "", windowNameOf("d"))
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
	// 回滚证据：窗口仍叫「凌晨」，引用仓库仍指向「凌晨」。
	if kept, err := repo.GetByID(context.TODO(), night.ID); assert.NoError(t, err) {
		assert.Equal(t, "凌晨", kept.Name)
	}
	assert.Equal(t, "深夜", windowNameOf("a"))

	// 窗口不存在：上抛（交由 errs.Wrap 归为 404）。
	_, _, err = repo.UpdateCascade(context.TODO(), &biz.HibernateWindow{
		ID: 99999, Name: "幽灵", Start: "01:00", End: "02:00",
	}, "凌晨")
	assert.Error(t, err)
}

// TestHibernateWindowRepo_UpdateCascade_SoftDeletedExcluded 验证级联改名不吃进软删仓库：
// 引用旧名的仓库中，软删的不参与迁移（迁移数只算活跃引用），即 SoftDeleteMixin 拦截器在
// 事务内同样生效——避免给已删仓库写回 window_name（依赖 UpdateCascade 注释所声称的不变量）。
func TestHibernateWindowRepo_UpdateCascade_SoftDeletedExcluded(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	defer entdb.Close()
	repo := NewHibernateWindowRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: entdb}))

	night, err := repo.Create(context.TODO(), &biz.HibernateWindow{Name: "夜间-sd", Start: "23:00", End: "06:00"})
	require.NoError(t, err)

	mk := func(name string) int {
		e, err := entdb.Repo.Create().SetName(name).
			SetMarsConfig(&mars.Config{Hibernate: &mars.Hibernate{WindowName: "夜间-sd"}}).
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
	assert.Equal(t, "深夜-sd", e.MarsConfig.GetHibernate().GetWindowName())
}

// TestHibernateWindowRepo_UpdateCascade_TxErrors 用故障注入覆盖事务内两条错误分支：
// 引用仓库查询失败（qAfter）与引用仓库更新失败（eAfter）——两者都必须整体回滚
// （窗口未改名、引用未迁移），证明半迁移不会发生。名称加后缀以避免与共享内存库的其他用例撞名。
func TestHibernateWindowRepo_UpdateCascade_TxErrors(t *testing.T) {
	cases := []struct {
		name           string
		qAfter, eAfter int32
	}{
		{"引用查询失败", 1, -1},
		{"引用更新失败", -1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, fd := newFailDB(t, tc.qAfter, tc.eAfter)
			repo := NewHibernateWindowRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: client}))

			night, err := repo.Create(context.TODO(), &biz.HibernateWindow{Name: "夜间-txerr", Start: "23:00", End: "06:00"})
			require.NoError(t, err)
			_, err = client.Repo.Create().SetName("a-txerr").
				SetMarsConfig(&mars.Config{Hibernate: &mars.Hibernate{WindowName: "夜间-txerr"}}).
				Save(context.TODO())
			require.NoError(t, err)

			fd.Arm()
			_, _, err = repo.UpdateCascade(context.TODO(), &biz.HibernateWindow{
				ID: night.ID, Name: "深夜-txerr", Start: "22:00", End: "07:00",
			}, "夜间-txerr")
			fd.Disarm()
			assert.Error(t, err)

			// 回滚证据：窗口仍叫「夜间-txerr」，引用仓库仍指向「夜间-txerr」。
			kept, err := repo.GetByID(context.TODO(), night.ID)
			require.NoError(t, err)
			assert.Equal(t, "夜间-txerr", kept.Name)
			e, err := client.Repo.Query().Where(entrepo.Name("a-txerr")).Only(context.TODO())
			require.NoError(t, err)
			assert.Equal(t, "夜间-txerr", e.MarsConfig.GetHibernate().GetWindowName())
		})
	}
}

// TestHibernateWindowRepo_GetByID 覆盖命中、未命中（返回 nil,nil）与 DB 故障上抛（不误判为未命中）。
func TestHibernateWindowRepo_GetByID(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	repo := NewHibernateWindowRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: entdb}))

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

// TestHibernateWindowRepo_ReposReferencing 覆盖按 window_name 过滤与升序：命中多条、
// 非命中、无 mars_config、hibernate 为空均不被误纳；无引用返回空；DB 故障上抛。
func TestHibernateWindowRepo_ReposReferencing(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	repo := NewHibernateWindowRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: entdb}))

	mk := func(name string, cfg *mars.Config) {
		b := entdb.Repo.Create().SetName(name)
		if cfg != nil {
			b = b.SetMarsConfig(cfg)
		}
		_, err := b.Save(context.TODO())
		assert.NoError(t, err)
	}
	mk("b", &mars.Config{Hibernate: &mars.Hibernate{WindowName: "午休"}}) // 非命中
	mk("a", &mars.Config{Hibernate: &mars.Hibernate{WindowName: "夜间"}}) // 命中
	mk("c", nil)                                                        // 无 mars_config
	mk("d", &mars.Config{Hibernate: nil})                               // hibernate 为空
	mk("e", &mars.Config{Hibernate: &mars.Hibernate{WindowName: "夜间"}}) // 命中

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

// TestHibernateWindowRepo_Errors 覆盖各方法在 DB 故障时的错误上抛（关闭连接触发）。
func TestHibernateWindowRepo_Errors(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	repo := NewHibernateWindowRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: entdb}))
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
