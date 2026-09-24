package data

import (
	"context"
	"errors"
	"regexp"
	"strings"
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

// sqlRecorder 是记录型 QueryMatcher：放行所有语句并留存实际 SQL，用于断言查询形状
// （列/过滤条件），不依赖 ent 生成语句的逐字文本。
type sqlRecorder struct{ sqls []string }

func (r *sqlRecorder) Match(_, actualSQL string) error {
	r.sqls = append(r.sqls, actualSQL)
	return nil
}

// find 返回第一条包含子串的已记录语句，找不到返回空串。
func (r *sqlRecorder) find(sub string) string {
	for _, q := range r.sqls {
		if strings.Contains(q, sub) {
			return q
		}
	}
	return ""
}

// Test_namespaceRepo_FindByName_LoadsMembers 断言 FindByName 预加载 members 边，且该边
// 查询过滤软删。
//
// 为什么必须预加载：按名字寻址的访问门卫（biz.RequireNamespaceAccessByName）直接拿本方法的
// 返回值判权限，而私有空间的成员判定读的正是 ns.Members（biz.CanAccessNamespace）。漏加载会让
// Members 恒空，私有空间的普通成员被误判成无权访问（403），而其按 ID 寻址的孪生入口
// （RequireNamespaceAccessByID → Show，同包已 WithMembers）却能通过——同名/同 id 两个入口行为
// 不一致，且与接口注释承诺的「私有空间仅 admin/创建者/成员」放行规则相悖。
//
// 为什么还要断言软删过滤：成员表带 SoftDeleteMixin，授权类查询靠软删过滤排除已移除成员；
// 嵌套边查询是否吃到该过滤不能靠假设，必须实证（漏过滤 = 已移除成员仍被认定有权限）。
func Test_namespaceRepo_FindByName_LoadsMembers(t *testing.T) {
	rec := &sqlRecorder{}
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(rec))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.MySQL, db)))
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}))

	// 命名空间主查询命中一个私有空间行。
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "name",
		"image_pull_secrets", "private", "creator_email", "description",
	}).AddRow(7, nil, nil, nil, "devops-demo", nil, true, "owner@mars.com", nil))
	// members 边查询返回一个成员。
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "email", "namespace_id",
	}).AddRow(1, nil, nil, nil, "member@mars.com", 7))

	ns, err := repo.FindByName(context.TODO(), "demo")
	require.NoError(t, err)
	require.Len(t, ns.Members, 1, "FindByName 必须带上成员边，否则私有空间成员会被误判无权")

	membersSQL := rec.find("FROM `members`")
	require.NotEmpty(t, membersSQL, "FindByName 未发出 members 边查询（WithMembers 缺失）")
	require.Contains(t, membersSQL, "`members`.`deleted_at` IS NULL",
		"成员边查询必须过滤软删——已移除的成员不得继续被判定为有访问权")
}
