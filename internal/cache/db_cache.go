package cache

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/dbcache"
)

type dbStore struct {
	db *ent.Client
}

func NewDBStore(db *ent.Client) Store {
	return &dbStore{db: db}
}

func (d *dbStore) Get(key string) (value []byte, err error) {
	first, err := d.db.DBCache.Query().Where(dbcache.Key(key), dbcache.ExpiredAtGTE(time.Now())).First(context.TODO())
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(first.Value)
}

func (d *dbStore) Set(key string, value []byte, seconds int) (err error) {
	toString := base64.StdEncoding.EncodeToString(value)

	return d.db.DBCache.Create().
		SetKey(key).
		SetValue(toString).
		SetExpiredAt(time.Now().Add(time.Duration(seconds) * time.Second)).
		OnConflict().
		UpdateNewValues().
		Exec(context.TODO())
}

func (d *dbStore) Delete(key string) error {
	_, err := d.db.DBCache.Delete().Where(dbcache.Key(key)).Exec(context.TODO())
	return err
}
