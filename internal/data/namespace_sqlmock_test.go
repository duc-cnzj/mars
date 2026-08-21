package data

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/stretchr/testify/require"
)

// newNamespaceMockEntClient 构造一个由 go-sqlmock 驱动的 ent client，
// 用于覆盖 namespaceRepo 中 DB 操作失败的分支，无需真实 MySQL。
func newNamespaceMockEntClient(t *testing.T) (*ent.Client, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))
	t.Cleanup(func() { _ = db.Close() })
	return client, mock
}

// Test_namespaceRepo_Favorite_MaxSortOrderQueryError 覆盖 Favorite 追加末尾时
// 查询该用户最大 sort_order 失败（非 NotFound）的防御分支：DB 抖动必须返回错误而非
// 被误判为"无历史关注"继续写入。
func Test_namespaceRepo_Favorite_MaxSortOrderQueryError(t *testing.T) {
	client, mock := newNamespaceMockEntClient(t)
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}))
	ctx := context.TODO()

	// 已有关注检查（ent Exist → SELECT id ... LIMIT 1）：空行视为无历史关注，继续走到 max sort_order 查询。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`.`id` FROM `favorites`")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	// max sort_order 查询返回 DB 错误（非 NotFound）→ 触发防御分支。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`.`id`, `favorites`.`email`, `favorites`.`namespace_id`, `favorites`.`sort_order` FROM `favorites`")).
		WillReturnError(errors.New("db boom"))

	err := repo.Favorite(ctx, &biz.FavoriteNamespaceInput{NamespaceID: 1, UserEmail: "u@mars.com", Favorite: true})
	require.Error(t, err)
}

// Test_namespaceRepo_FavoriteSort_QueryError 覆盖事务内查询该用户关注列表失败的防御分支：
// 查询失败直接回滚事务返回错误，不进入校验/回填逻辑。
func Test_namespaceRepo_FavoriteSort_QueryError(t *testing.T) {
	client, mock := newNamespaceMockEntClient(t)
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}))
	ctx := context.TODO()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnError(errors.New("db boom"))
	mock.ExpectRollback()

	err := repo.FavoriteSort(ctx, "u@mars.com", []int{1, 2})
	require.Error(t, err)
}

// Test_namespaceRepo_FavoriteSort_UpdateError 覆盖事务内回填 sort_order 写入失败的防御分支：
// 逐行更新中任一行写入失败即回滚整个事务，不留下半程重排状态。
func Test_namespaceRepo_FavoriteSort_UpdateError(t *testing.T) {
	client, mock := newNamespaceMockEntClient(t)
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}))
	ctx := context.TODO()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnRows(sqlmock.NewRows([]string{"namespace_id"}).AddRow(1).AddRow(2).AddRow(3))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `favorites`")).
		WillReturnError(errors.New("db boom"))
	mock.ExpectRollback()

	err := repo.FavoriteSort(ctx, "u@mars.com", []int{1, 2, 3})
	require.Error(t, err)
}
