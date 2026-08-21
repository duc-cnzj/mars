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

// Test_namespaceRepo_FavoriteSort_QueryError 覆盖事务内查询两个关注空间失败的防御分支：
// 查询失败直接回滚事务返回错误，不进入移动/回填逻辑。
func Test_namespaceRepo_FavoriteSort_QueryError(t *testing.T) {
	client, mock := newNamespaceMockEntClient(t)
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}))
	ctx := context.TODO()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnError(errors.New("db boom"))
	mock.ExpectRollback()

	err := repo.FavoriteSort(ctx, "u@mars.com", 1, 2)
	require.Error(t, err)
}

// Test_namespaceRepo_FavoriteSort_UpdateError 覆盖事务内区间顺移写入失败的防御分支：
// 顺移/落位任一步写入失败即回滚整个事务，不留下半程移动状态。
func Test_namespaceRepo_FavoriteSort_UpdateError(t *testing.T) {
	client, mock := newNamespaceMockEntClient(t)
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}))
	ctx := context.TODO()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "namespace_id", "sort_order"}).
			AddRow(1, "u@mars.com", 1, 0).
			AddRow(2, "u@mars.com", 2, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `favorites`")).
		WillReturnError(errors.New("db boom"))
	mock.ExpectRollback()

	err := repo.FavoriteSort(ctx, "u@mars.com", 1, 2)
	require.Error(t, err)
}

// Test_namespaceRepo_FavoriteSort_BackwardUpdateError 覆盖后移分支（firstID 在 secondID 之后，
// 中间区间 +1）顺移写入失败的防御分支：失败即回滚整个事务。
func Test_namespaceRepo_FavoriteSort_BackwardUpdateError(t *testing.T) {
	client, mock := newNamespaceMockEntClient(t)
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}))
	ctx := context.TODO()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "namespace_id", "sort_order"}).
			AddRow(1, "u@mars.com", 1, 0).
			AddRow(2, "u@mars.com", 2, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `favorites`")).
		WillReturnError(errors.New("db boom"))
	mock.ExpectRollback()

	// firstID=2(order 1) 在 secondID=1(order 0) 之后 → 走 else 分支，区间 UPDATE 失败。
	err := repo.FavoriteSort(ctx, "u@mars.com", 2, 1)
	require.Error(t, err)
}

// Test_namespaceRepo_FavoriteSort_RenumberQueryError 覆盖两空间 sort_order 相同时触发的懒重排：
// 重排查询（按 email 取全部关注）失败的防御分支，DB 抖动必须返回错误而非静默原地移动。
func Test_namespaceRepo_FavoriteSort_RenumberQueryError(t *testing.T) {
	client, mock := newNamespaceMockEntClient(t)
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}))
	ctx := context.TODO()

	mock.ExpectBegin()
	// 两空间 sort_order 相同（0,0）→ 进入懒重排。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "namespace_id", "sort_order"}).
			AddRow(1, "u@mars.com", 1, 0).
			AddRow(2, "u@mars.com", 2, 0))
	// 重排查询失败 → 回滚。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnError(errors.New("db boom"))
	mock.ExpectRollback()

	err := repo.FavoriteSort(ctx, "u@mars.com", 1, 2)
	require.Error(t, err)
}

// Test_namespaceRepo_FavoriteSort_RenumberUpdateError 覆盖懒重排落位写入失败的防御分支：
// 重排 UPDATE 失败即回滚整个事务，不留半程重排状态。
func Test_namespaceRepo_FavoriteSort_RenumberUpdateError(t *testing.T) {
	client, mock := newNamespaceMockEntClient(t)
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}))
	ctx := context.TODO()

	mock.ExpectBegin()
	// 两空间 sort_order 相同 → 进入懒重排。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "namespace_id", "sort_order"}).
			AddRow(1, "u@mars.com", 1, 0).
			AddRow(2, "u@mars.com", 2, 0))
	// 重排查询：按 (sort_order, id) 稳定序返回两行，第二行 0!=1 需落位更新。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "namespace_id", "sort_order"}).
			AddRow(1, "u@mars.com", 1, 0).
			AddRow(2, "u@mars.com", 2, 0))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `favorites`")).
		WillReturnError(errors.New("db boom"))
	mock.ExpectRollback()

	err := repo.FavoriteSort(ctx, "u@mars.com", 1, 2)
	require.Error(t, err)
}

// Test_namespaceRepo_FavoriteSort_RenumberRereadError 覆盖懒重排成功后重读两空间落位失败的
// 防御分支：重读失败即回滚整个事务，不进入区间顺移。
func Test_namespaceRepo_FavoriteSort_RenumberRereadError(t *testing.T) {
	client, mock := newNamespaceMockEntClient(t)
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}))
	ctx := context.TODO()

	mock.ExpectBegin()
	// 两空间 sort_order 相同 → 进入懒重排。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "namespace_id", "sort_order"}).
			AddRow(1, "u@mars.com", 1, 0).
			AddRow(2, "u@mars.com", 2, 0))
	// 重排查询：已连续 0,1，无需 UPDATE。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "namespace_id", "sort_order"}).
			AddRow(1, "u@mars.com", 1, 0).
			AddRow(2, "u@mars.com", 2, 1))
	// 重读两空间落位失败 → 回滚。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `favorites`")).
		WillReturnError(errors.New("db boom"))
	mock.ExpectRollback()

	err := repo.FavoriteSort(ctx, "u@mars.com", 1, 2)
	require.Error(t, err)
}
