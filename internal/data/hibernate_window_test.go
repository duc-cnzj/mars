package data

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
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
