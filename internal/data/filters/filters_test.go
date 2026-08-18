package filters

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	// 注册 sqlite3 驱动，避免依赖 internal/data（data 会反向 import filters）。
	_ "github.com/mattn/go-sqlite3"
	// 加载 ent mixin 默认值（与 data.NewSqliteDB 行为一致）。
	_ "github.com/duc-cnzj/mars/v6/internal/data/ent/runtime"
)

func openTestDB(t *testing.T) *ent.Client {
	t.Helper()
	db, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1&loc=Local")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Schema.Create(context.TODO()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestIfStrEQ(t *testing.T) {
	db := openTestDB(t)
	db.Namespace.Create().SetName("test").SetCreatorEmail("a").SaveX(context.TODO())
	db.Namespace.Create().SetName("test2").SetCreatorEmail("a").SaveX(context.TODO())
	x := db.Namespace.Query().Where(IfStrEQ("name")("test")).AllX(context.TODO())
	assert.Equal(t, 1, len(x))
	assert.Equal(t, "test", x[0].Name)
	x2 := db.Namespace.Query().Where(IfStrEQ("name")("")).AllX(context.TODO())
	assert.Equal(t, 2, len(x2))

	x3 := db.Namespace.Query().Where(IfNameLike("te")).AllX(context.TODO())
	assert.Equal(t, 2, len(x3))
	x4 := db.Namespace.Query().Where(IfNameLike("st2")).AllX(context.TODO())
	assert.Equal(t, 1, len(x4))
	assert.Equal(t, "test2", x4[0].Name)
}

func TestIfIntsIN(t *testing.T) {
	db := openTestDB(t)
	id1 := db.Namespace.Create().SetCreatorEmail("a").SetName("test").SaveX(context.TODO()).ID
	id2 := db.Namespace.Create().SetCreatorEmail("a").SetName("test2").SaveX(context.TODO()).ID
	db.Namespace.Create().SetCreatorEmail("a").SetName("test3").SaveX(context.TODO())
	x := db.Namespace.Query().Where(IfIntsIN[int]("id")([]int{id1, id2})).AllX(context.TODO())
	assert.Equal(t, 2, len(x))
	x2 := db.Namespace.Query().Where(IfIntsIN[int]("id")(nil)).AllX(context.TODO())
	assert.Equal(t, 3, len(x2))
	x2b := db.Namespace.Query().Where(IfIntsIN[int]("id")([]int{})).AllX(context.TODO())
	assert.Equal(t, 3, len(x2b))
	x3 := db.Namespace.Query().Where(IfIntsIN[int]("id")([]int{-999})).AllX(context.TODO())
	assert.Equal(t, 0, len(x3))
}

func TestIfOrderByDesc(t *testing.T) {
	db := openTestDB(t)
	db.Namespace.Create().SetCreatorEmail("a").SetName("test").SaveX(context.TODO())
	db.Namespace.Create().SetCreatorEmail("a").SetName("test2").SaveX(context.TODO())
	x := db.Namespace.Query().Where(IfOrderByDesc("id")(lo.ToPtr(false))).AllX(context.TODO())
	assert.Equal(t, 2, len(x))
	assert.Equal(t, "test", x[0].Name)
	x2 := db.Namespace.Query().Where(IfOrderByDesc("id")(lo.ToPtr(true))).AllX(context.TODO())
	assert.Equal(t, 2, len(x2))
	assert.Equal(t, "test2", x2[0].Name)
}

func TestIfBool(t *testing.T) {
	db := openTestDB(t)
	db.Repo.Create().SetName("test").SetEnabled(true).SaveX(context.TODO())
	db.Repo.Create().SetName("test2").SetEnabled(false).SaveX(context.TODO())
	x := db.Repo.Query().Where(IfBool("enabled")(lo.ToPtr(false))).AllX(context.TODO())
	assert.Equal(t, 1, len(x))
	assert.Equal(t, "test2", x[0].Name)
	x2 := db.Repo.Query().Where(IfBool("enabled")(nil)).AllX(context.TODO())
	assert.Equal(t, 2, len(x2))
	x3 := db.Repo.Query().Where(IfBool("enabled")(lo.ToPtr(true))).AllX(context.TODO())
	assert.Equal(t, 1, len(x3))
	assert.Equal(t, "test", x3[0].Name)
}
