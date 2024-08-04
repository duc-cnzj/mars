package cache

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent/dbcache"
)

type dbStore struct {
	data data.Data
}

func NewDBStore(data data.Data) Store {
	return &dbStore{data: data}
}

func (d *dbStore) Get(key string) (value []byte, err error) {
	first, err := d.data.DB().DBCache.Query().Where(dbcache.Key(key), dbcache.ExpiredAtGTE(time.Now())).First(context.TODO())
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(first.Value)
}

func (d *dbStore) Set(key string, value []byte, seconds int) (err error) {
	toString := base64.StdEncoding.EncodeToString(value)

	return d.data.DB().DBCache.Create().
		SetKey(key).
		SetValue(toString).
		SetExpiredAt(time.Now().Add(time.Duration(seconds) * time.Second)).
		OnConflict().
		UpdateNewValues().
		Exec(context.TODO())
}

func (d *dbStore) Delete(key string) error {
	_, err := d.data.DB().DBCache.Delete().Where(dbcache.Key(key)).Exec(context.TODO())
	return err
}
