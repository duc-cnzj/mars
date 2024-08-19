package filters

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestIfStrEQ(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	db.Namespace.Create().SetName("test").SaveX(context.TODO())
	db.Namespace.Create().SetName("test2").SaveX(context.TODO())
	x := db.Namespace.Query().Where(IfStrEQ("name")("test")).AllX(context.Background())
	assert.Equal(t, 1, len(x))
	assert.Equal(t, "test", x[0].Name)
	x2 := db.Namespace.Query().Where(IfStrEQ("name")("")).AllX(context.Background())
	assert.Equal(t, 2, len(x2))

	x3 := db.Namespace.Query().Where(IfNameLike("te")).AllX(context.Background())
	assert.Equal(t, 2, len(x3))
	x4 := db.Namespace.Query().Where(IfNameLike("st2")).AllX(context.Background())
	assert.Equal(t, 1, len(x4))
	assert.Equal(t, "test2", x4[0].Name)
}

func TestIfIntEQ(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	save1 := db.Namespace.Create().SetName("test").SaveX(context.TODO())
	db.Namespace.Create().SetName("test2").SaveX(context.TODO())
	x := db.Namespace.Query().Where(IfIntEQ[int]("id")(save1.ID)).AllX(context.Background())
	assert.Equal(t, 1, len(x))
	assert.Equal(t, "test", x[0].Name)
	x2 := db.Namespace.Query().Where(IfIntEQ[int]("name")(0)).AllX(context.Background())
	assert.Equal(t, 2, len(x2))
}

func TestIfOrderByDesc(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	db.Namespace.Create().SetName("test").SaveX(context.TODO())
	db.Namespace.Create().SetName("test2").SaveX(context.TODO())
	x := db.Namespace.Query().Where(IfOrderByDesc("id")(lo.ToPtr(false))).AllX(context.Background())
	assert.Equal(t, 2, len(x))
	assert.Equal(t, "test", x[0].Name)
	x2 := db.Namespace.Query().Where(IfOrderByDesc("id")(lo.ToPtr(true))).AllX(context.Background())
	assert.Equal(t, 2, len(x2))
	assert.Equal(t, "test2", x2[0].Name)
}

func TestIfBool(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	db.Repo.Create().SetName("test").SetEnabled(true).SaveX(context.TODO())
	db.Repo.Create().SetName("test2").SetEnabled(false).SaveX(context.TODO())
	x := db.Repo.Query().Where(IfBool("enabled")(lo.ToPtr(false))).AllX(context.Background())
	assert.Equal(t, 1, len(x))
	assert.Equal(t, "test2", x[0].Name)
	x2 := db.Repo.Query().Where(IfBool("enabled")(nil)).AllX(context.Background())
	assert.Equal(t, 2, len(x2))
	x3 := db.Repo.Query().Where(IfBool("enabled")(lo.ToPtr(true))).AllX(context.Background())
	assert.Equal(t, 1, len(x3))
	assert.Equal(t, "test", x3[0].Name)
}
