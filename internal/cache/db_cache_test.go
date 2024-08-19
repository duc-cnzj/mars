package cache

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent/dbcache"
	"github.com/stretchr/testify/assert"
)

func Test_dbStore_Get(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()

	s := &dbStore{
		data: data.NewDataImpl(&data.NewDataParams{DB: db}),
	}
	_, err := s.Get("test")
	assert.Error(t, err)

	db.DBCache.Create().SetKey("test").SetValue("test").SetExpiredAt(time.Now().Add(10 * time.Second)).Exec(context.TODO())

	v, err := s.Get("test")
	assert.Nil(t, err)
	decodeString, _ := base64.StdEncoding.DecodeString("test")
	assert.Equal(t, decodeString, v)
}

func Test_dbStore_Set(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()

	s := &dbStore{
		data: data.NewDataImpl(&data.NewDataParams{DB: db}),
	}
	err := s.Set("test", []byte("test"), 10)
	assert.Nil(t, err)
	only1, _ := db.DBCache.Query().Where(dbcache.Key("test")).Only(context.TODO())
	assert.Equal(t, "test", B64toStr(only1.Value))
	err = s.Set("test", []byte("testxxx"), 10)
	assert.Nil(t, err)
	only2, _ := db.DBCache.Query().Where(dbcache.Key("test")).Only(context.TODO())
	assert.Equal(t, "testxxx", B64toStr(only2.Value))
	assert.NotEqual(t, only1.ExpiredAt, only2.ExpiredAt)
}

func B64toStr(v string) string {
	decodeString, _ := base64.StdEncoding.DecodeString(v)
	return string(decodeString)
}

func Test_dbStore_Delete(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()

	s := &dbStore{
		data: data.NewDataImpl(&data.NewDataParams{DB: db}),
	}
	err := s.Delete("test")
	assert.Nil(t, err)
	db.DBCache.Create().SetKey("test").SetValue("test").SetExpiredAt(time.Now().Add(10 * time.Second)).Exec(context.TODO())
	err = s.Delete("test")
	assert.Nil(t, err)
	_, err = db.DBCache.Query().Where(dbcache.Key("test")).Only(context.TODO())
	assert.Error(t, err)
}
