package repo

import (
	"context"
	"fmt"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/mlog"

	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/stretchr/testify/assert"
)

func Test_cronRepo_allNamespaces(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	cr := &cronRepo{
		logger: mlog.NewLogger(nil),
		nsRepo: &namespaceRepo{
			logger: mlog.NewLogger(nil),
			data:   data.NewDataImpl(&data.NewDataParams{DB: db}),
		},
	}
	for i := 0; i < 10; i++ {
		db.Namespace.Create().SetName(fmt.Sprintf("test%d", i)).SaveX(context.TODO())
	}
	namespaces, err := cr.allNamespaces(3)
	assert.Nil(t, err)
	assert.Len(t, namespaces, 10)
	for i, namespace := range namespaces {
		assert.Equal(t, fmt.Sprintf("test%d", i), namespace.Name)
	}
}
