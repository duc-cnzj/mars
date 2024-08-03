package data

import (
	"context"

	"github.com/duc-cnzj/mars/v4/internal/ent"

	_ "github.com/duc-cnzj/mars/v4/internal/ent/runtime"
	_ "github.com/mattn/go-sqlite3"
)

func NewSqliteDB() (*ent.Client, error) {
	client, _ := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1&loc=Local")
	_ = client.Schema.Create(context.TODO())

	return client, nil
}
