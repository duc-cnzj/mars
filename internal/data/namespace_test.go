package data

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	entgo "entgo.io/ent"

	"github.com/duc-cnzj/mars/api/v6/proto/types"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/favorite"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/namespace"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/project"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNamespaceRepo_List_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ctx := context.TODO()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{},
		DB:  entdb,
	}))

	// seed data
	ns1 := entdb.Namespace.Create().SetCreatorEmail("a").SetName("test").SaveX(ctx)
	entdb.Favorite.Create().SetNamespaceID(ns1.ID).SetEmail("test@example.com").Save(context.TODO())
	entdb.Namespace.Create().SetCreatorEmail("a").SetName("tes2").SaveX(ctx)
	entdb.Namespace.Create().SetCreatorEmail("a").SetName("tes3").SaveX(ctx)
	pri1 := entdb.Namespace.Create().SetCreatorEmail("a").SetName("pri1").SetPrivate(true).SaveX(ctx)
	entdb.Member.Create().SetEmail("user@mars.com").SetNamespaceID(pri1.ID).SaveX(ctx)
	entdb.Namespace.Create().SetCreatorEmail("a").SetName("pri2").SetPrivate(true).SaveX(ctx)

	input := &biz.ListNamespaceInput{
		Favorite: true,
		Email:    "test@example.com",
		Page:     1,
		PageSize: 10,
		Name:     nil,
		IsAdmin:  true,
	}

	res, pag, err := repo.List(ctx, input)
	assert.NotNil(t, res)
	assert.NotNil(t, pag)
	assert.Nil(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, int32(1), pag.Count)
	assert.Equal(t, int32(1), pag.Page)
	assert.Equal(t, int32(10), pag.PageSize)
	assert.Equal(t, ns1.ID, res[0].ID)

	input = &biz.ListNamespaceInput{
		Favorite: false,
		Page:     1,
		PageSize: 10,
		Name:     lo.ToPtr("es3"),
	}

	res, pag, _ = repo.List(ctx, input)
	assert.Len(t, res, 1)
	assert.Equal(t, int32(1), pag.Count)

	input = &biz.ListNamespaceInput{
		Page:     1,
		PageSize: 10,
		IsAdmin:  false,
		Email:    "",
	}

	res, pag, _ = repo.List(ctx, input)
	assert.Len(t, res, 3)
	assert.Equal(t, int32(3), pag.Count)

	input = &biz.ListNamespaceInput{
		Page:     1,
		PageSize: 10,
		IsAdmin:  false,
		Email:    "user@mars.com",
	}

	res, pag, _ = repo.List(ctx, input)
	assert.Len(t, res, 4)
	assert.Equal(t, int32(4), pag.Count)

	input = &biz.ListNamespaceInput{
		Page:     1,
		PageSize: 10,
		IsAdmin:  true,
	}

	res, pag, _ = repo.List(ctx, input)
	assert.Len(t, res, 5)
	assert.Equal(t, int32(5), pag.Count)

	input = &biz.ListNamespaceInput{
		Email:    "a",
		Page:     1,
		PageSize: 10,
		IsAdmin:  false,
	}

	res, pag, _ = repo.List(ctx, input)
	assert.Len(t, res, 5)
	assert.Equal(t, int32(5), pag.Count)

	// 管理员后台搜索/私有过滤：Search 模糊匹配空间名或创建者邮箱，PrivateOnly 只看私有。
	t.Run("admin search by name", func(t *testing.T) {
		res, _, err := repo.List(ctx, &biz.ListNamespaceInput{Page: 1, PageSize: 10, IsAdmin: true, Search: "es3"})
		assert.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, "tes3", res[0].Name)
	})

	t.Run("admin search by creator email", func(t *testing.T) {
		res, pag, err := repo.List(ctx, &biz.ListNamespaceInput{Page: 1, PageSize: 10, IsAdmin: true, Search: "a"})
		assert.NoError(t, err)
		assert.Equal(t, int32(5), pag.Count)
		assert.Len(t, res, 5)
	})

	t.Run("admin private only", func(t *testing.T) {
		res, pag, err := repo.List(ctx, &biz.ListNamespaceInput{Page: 1, PageSize: 10, IsAdmin: true, PrivateOnly: true})
		assert.NoError(t, err)
		assert.Equal(t, int32(2), pag.Count)
		require.Len(t, res, 2)
		for _, ns := range res {
			assert.True(t, ns.Private)
		}
	})
}

func Test_namespaceRepo_Create(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: entdb,
	}))

	create, err := repo.Create(context.TODO(), &biz.CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc-aaa", create.Name)
	assert.Equal(t, "aa", create.CreatorEmail)
}

func Test_namespaceRepo_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc",
		},
		DB: entdb,
	}))

	create, err := repo.Create(context.TODO(), &biz.CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
	})
	createProject(entdb, create.ID)
	assert.Nil(t, err)
	show, err := repo.Show(context.TODO(), create.ID)
	assert.Nil(t, err)
	assert.Len(t, show.Projects, 1)
}

// Test_namespaceRepo_Show_ProjectDeployFields 回归 Show 的项目列需与 List 路径一致：
// 必须携带 deploy_status/created_at/updated_at，否则项目 DeployStatus 会落成零值
// （StatusUnknown），部署成功后前端 refreshNamespace 用 show 原地替换卡片，
// 项目部署状态就"变没"了。
func Test_namespaceRepo_Show_ProjectDeployFields(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc",
		},
		DB: entdb,
	}))

	create, err := repo.Create(context.TODO(), &biz.CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
	})
	require.NoError(t, err)

	// 项目写入非默认部署状态（StatusDeployed），Show 需原样带回该状态而非零值。
	entdb.Project.Create().
		SetGitBranch("").
		SetGitCommit("").
		SetConfig("").
		SetGitProjectID(1).
		SetCreator("").
		SetName("testProject").
		SetNamespaceID(create.ID).
		SetDeployStatus(types.Deploy_StatusDeployed).
		SaveX(context.TODO())

	show, err := repo.Show(context.TODO(), create.ID)
	require.NoError(t, err)
	require.Len(t, show.Projects, 1)

	got := show.Projects[0]
	assert.Equal(t, types.Deploy_StatusDeployed, got.DeployStatus)
	assert.False(t, got.CreatedAt.IsZero(), "Show 的项目应返回非零 CreatedAt")
	assert.False(t, got.UpdatedAt.IsZero(), "Show 的项目应返回非零 UpdatedAt")
}

func Test_namespaceRepo_Update(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc",
		},
		DB: entdb,
	}))

	create, err := repo.Create(context.TODO(), &biz.CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
	})
	assert.Nil(t, err)
	update, err := repo.Update(context.TODO(), &biz.UpdateNamespaceInput{
		ID:          create.ID,
		Description: "aaaaaa",
	})
	assert.Nil(t, err)
	assert.Equal(t, "aaaaaa", update.Description)
}

func Test_namespaceRepo_GetMarsNamespace(t *testing.T) {
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
	}))
	namespace := repo.GetMarsNamespace("a")
	assert.Equal(t, "abc-a", namespace)
	marsNamespace := repo.GetMarsNamespace("abc-a")
	assert.Equal(t, "abc-a", marsNamespace)
}

func Test_namespaceRepo_Delete(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc",
		},
		DB: entdb,
	}))

	create, err := repo.Create(context.TODO(), &biz.CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
	})
	assert.Nil(t, err)

	create2, err := repo.Create(context.TODO(), &biz.CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
	})
	assert.Nil(t, err)

	createProject(entdb, create.ID)
	createProject(entdb, create.ID)
	createProject(entdb, create.ID)
	createProject(entdb, create.ID)
	p2 := createProject(entdb, create2.ID)

	err = repo.Delete(context.TODO(), create.ID)
	assert.Nil(t, err)
	softDelete := mixin.SkipSoftDelete(context.TODO())
	x := entdb.Project.Query().Where(project.NamespaceID(create.ID)).AllX(softDelete)

	for _, p := range x {
		assert.NotZero(t, p.DeletedAt)
	}
	first, _ := entdb.Project.Query().Where(project.ID(p2.ID)).First(context.TODO())
	assert.Zero(t, first.DeletedAt)

	n, err := entdb.Namespace.Query().Where(namespace.ID(create.ID)).First(softDelete)
	assert.Nil(t, err)
	assert.NotZero(t, n.DeletedAt)
}

func Test_namespaceRepo_Favorite_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{},
		DB:  entdb,
	}))

	ns := entdb.Namespace.Create().SetCreatorEmail("a").SetName("test").SaveX(context.TODO())
	input := &biz.FavoriteNamespaceInput{
		NamespaceID: ns.ID,
		UserEmail:   "test@example.com",
		Favorite:    true,
	}

	err := repo.Favorite(context.TODO(), input)
	assert.Nil(t, err)

	fav := entdb.Favorite.Query().Where(favorite.NamespaceID(ns.ID), favorite.Email("test@example.com")).OnlyX(context.TODO())
	assert.NotNil(t, fav)
}

func Test_namespaceRepo_Favorite_AlreadyExists(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{},
		DB:  entdb,
	}))

	ns := entdb.Namespace.Create().SetCreatorEmail("a").SetName("test").SaveX(context.TODO())
	entdb.Favorite.Create().SetNamespaceID(ns.ID).SetEmail("test@example.com").Save(context.TODO())
	input := &biz.FavoriteNamespaceInput{
		NamespaceID: ns.ID,
		UserEmail:   "test@example.com",
		Favorite:    true,
	}

	err := repo.Favorite(context.TODO(), input)
	assert.Nil(t, err)
}

func Test_namespaceRepo_Favorite_Unfavorite(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{},
		DB:  entdb,
	}))

	ns := entdb.Namespace.Create().SetCreatorEmail("a").SetName("test").SaveX(context.TODO())
	entdb.Favorite.Create().SetNamespaceID(ns.ID).SetEmail("test@example.com").Save(context.TODO())
	input := &biz.FavoriteNamespaceInput{
		NamespaceID: ns.ID,
		UserEmail:   "test@example.com",
		Favorite:    false,
	}

	err := repo.Favorite(context.TODO(), input)
	assert.Nil(t, err)

	_, err = entdb.Favorite.Query().Where(favorite.NamespaceID(ns.ID), favorite.Email("test@example.com")).Only(context.TODO())
	assert.Error(t, err)
}

func Test_namespaceRepo_Favorite_SortOrderAppend(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{},
		DB:  entdb,
	}))
	ctx := context.TODO()
	ns1 := entdb.Namespace.Create().SetCreatorEmail("a").SetName("ns-1").SaveX(ctx)
	ns2 := entdb.Namespace.Create().SetCreatorEmail("a").SetName("ns-2").SaveX(ctx)

	require.NoError(t, repo.Favorite(ctx, &biz.FavoriteNamespaceInput{NamespaceID: ns1.ID, UserEmail: "u@mars.com", Favorite: true}))
	fav1 := entdb.Favorite.Query().Where(favorite.NamespaceID(ns1.ID), favorite.Email("u@mars.com")).OnlyX(ctx)
	require.Equal(t, 0, fav1.SortOrder, "首个关注排最前")

	require.NoError(t, repo.Favorite(ctx, &biz.FavoriteNamespaceInput{NamespaceID: ns2.ID, UserEmail: "u@mars.com", Favorite: true}))
	fav2 := entdb.Favorite.Query().Where(favorite.NamespaceID(ns2.ID), favorite.Email("u@mars.com")).OnlyX(ctx)
	require.Equal(t, 1, fav2.SortOrder, "新关注追加末尾")
}

func Test_namespaceRepo_List_FavoriteOrdered(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{},
		DB:  entdb,
	}))
	ctx := context.TODO()
	var ns []*ent.Namespace
	for i := 0; i < 4; i++ {
		ns = append(ns, entdb.Namespace.Create().SetCreatorEmail("a").SetName(fmt.Sprintf("ns-%d", i)).SaveX(ctx))
	}
	email := "u@mars.com"
	// 打乱 sort_order：期望按 0,1,2,3 升序返回（ns3, ns1, ns4, ns2）
	entdb.Favorite.Create().SetNamespaceID(ns[2].ID).SetEmail(email).SetSortOrder(0).SaveX(ctx)
	entdb.Favorite.Create().SetNamespaceID(ns[0].ID).SetEmail(email).SetSortOrder(1).SaveX(ctx)
	entdb.Favorite.Create().SetNamespaceID(ns[3].ID).SetEmail(email).SetSortOrder(2).SaveX(ctx)
	entdb.Favorite.Create().SetNamespaceID(ns[1].ID).SetEmail(email).SetSortOrder(3).SaveX(ctx)

	res, _, err := repo.List(ctx, &biz.ListNamespaceInput{Favorite: true, Email: email, Page: 1, PageSize: 10, IsAdmin: true})
	require.NoError(t, err)
	require.Len(t, res, 4)
	want := []int{ns[2].ID, ns[0].ID, ns[3].ID, ns[1].ID}
	for i, w := range want {
		require.Equal(t, w, res[i].ID, "关注列表应按 sort_order 升序")
	}
}

func Test_namespaceRepo_FavoriteSort(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{},
		DB:  entdb,
	}))
	ctx := context.TODO()
	var ns []*ent.Namespace
	for i := 0; i < 5; i++ {
		ns = append(ns, entdb.Namespace.Create().SetCreatorEmail("a").SetName(fmt.Sprintf("ns-%d", i)).SaveX(ctx))
	}
	email := "u@mars.com"
	for i, n := range ns {
		entdb.Favorite.Create().SetNamespaceID(n.ID).SetEmail(email).SetSortOrder(i).SaveX(ctx)
	}
	// 后→前：ns4 移到 ns1 位置，中间 ns1/ns2/ns3 整体后移一位。
	require.NoError(t, repo.FavoriteSort(ctx, email, ns[4].ID, ns[1].ID))

	var got []int
	all := entdb.Favorite.Query().Where(favorite.Email(email)).Order(ent.Asc(favorite.FieldSortOrder)).AllX(ctx)
	for _, f := range all {
		got = append(got, f.NamespaceID)
	}
	require.Equal(t, []int{ns[0].ID, ns[4].ID, ns[1].ID, ns[2].ID, ns[3].ID}, got, "ns4 应落到 ns1 原位置，中间元素顺移")
}

// Test_namespaceRepo_FavoriteSort_ForwardMove 前→后移动：ns0 移到 ns3 位置，中间元素整体前移一位。
func Test_namespaceRepo_FavoriteSort_ForwardMove(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))
	ctx := context.TODO()
	var ns []*ent.Namespace
	for i := 0; i < 5; i++ {
		ns = append(ns, entdb.Namespace.Create().SetCreatorEmail("a").SetName(fmt.Sprintf("ns-%d", i)).SaveX(ctx))
	}
	email := "u@mars.com"
	for i, n := range ns {
		entdb.Favorite.Create().SetNamespaceID(n.ID).SetEmail(email).SetSortOrder(i).SaveX(ctx)
	}

	require.NoError(t, repo.FavoriteSort(ctx, email, ns[0].ID, ns[3].ID))

	got := entdb.Favorite.Query().Where(favorite.Email(email)).Order(ent.Asc(favorite.FieldSortOrder)).AllX(ctx)
	require.Equal(t, []int{ns[1].ID, ns[2].ID, ns[3].ID, ns[0].ID, ns[4].ID}, []int{got[0].NamespaceID, got[1].NamespaceID, got[2].NamespaceID, got[3].NamespaceID, got[4].NamespaceID}, "ns0 应落到 ns3 原位置，中间元素顺移")
}

// Test_namespaceRepo_FavoriteSort_SameID first_id 与 second_id 相同必须被拒绝，否则半程移动静默破坏排序。
func Test_namespaceRepo_FavoriteSort_SameID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))
	ctx := context.TODO()
	ns0 := entdb.Namespace.Create().SetCreatorEmail("a").SetName("ns-0").SaveX(ctx)
	email := "u@mars.com"
	entdb.Favorite.Create().SetNamespaceID(ns0.ID).SetEmail(email).SetSortOrder(0).SaveX(ctx)

	err := repo.FavoriteSort(ctx, email, ns0.ID, ns0.ID)
	require.ErrorContains(t, err, "两个空间 id 不能相同", "相同 id 必须报错")
}

// Test_namespaceRepo_FavoriteSort_NotFavorite 传入非该用户关注的空间必须被拒绝，防止越权移动他人空间。
func Test_namespaceRepo_FavoriteSort_NotFavorite(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))
	ctx := context.TODO()
	ns0 := entdb.Namespace.Create().SetCreatorEmail("a").SetName("ns-0").SaveX(ctx)
	ns1 := entdb.Namespace.Create().SetCreatorEmail("a").SetName("ns-1").SaveX(ctx)
	ns2 := entdb.Namespace.Create().SetCreatorEmail("a").SetName("ns-2").SaveX(ctx)
	email := "u@mars.com"
	entdb.Favorite.Create().SetNamespaceID(ns0.ID).SetEmail(email).SetSortOrder(0).SaveX(ctx)
	entdb.Favorite.Create().SetNamespaceID(ns1.ID).SetEmail(email).SetSortOrder(1).SaveX(ctx)
	// ns2 未关注：查询只命中 1 行，必须报错。
	err := repo.FavoriteSort(ctx, email, ns2.ID, ns0.ID)
	require.ErrorContains(t, err, "必须同时包含这两个空间", "含非本人关注空间必须报错")
}

// Test_namespaceRepo_FavoriteSort_SameOrderDirtyData 历史脏数据下两个空间 sort_order 相同
// （迁移全 0 / 手工改库重复序）：区间顺移缺位置信息，须在事务内先按 (sort_order,id) 稳定序
// 重排为 0..N 再继续移动，保证"全部 sort_order 为 0 也能拖拽排序"，不随系统启动回填。
func Test_namespaceRepo_FavoriteSort_SameOrderDirtyData(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))
	ctx := context.TODO()
	ns0 := entdb.Namespace.Create().SetCreatorEmail("a").SetName("ns-0").SaveX(ctx)
	ns1 := entdb.Namespace.Create().SetCreatorEmail("a").SetName("ns-1").SaveX(ctx)
	email := "u@mars.com"
	// 脏数据：两个关注行 sort_order 相同（正常流程唯一，仅历史迁移/手工改库可达）。
	entdb.Favorite.Create().SetNamespaceID(ns0.ID).SetEmail(email).SetSortOrder(0).SaveX(ctx)
	entdb.Favorite.Create().SetNamespaceID(ns1.ID).SetEmail(email).SetSortOrder(0).SaveX(ctx)

	// ns0 移到 ns1 位置：先重排为 [ns0:0, ns1:1]，再区间顺移，ns1 前移到 0，ns0 落 1。
	require.NoError(t, repo.FavoriteSort(ctx, email, ns0.ID, ns1.ID), "全部 sort_order 为 0 也应能拖拽排序")
	all := entdb.Favorite.Query().Where(favorite.Email(email)).Order(ent.Asc(favorite.FieldSortOrder)).AllX(ctx)
	require.Equal(t, []int{ns1.ID, ns0.ID}, []int{all[0].NamespaceID, all[1].NamespaceID}, "ns0 应移到 ns1 位置，ns1 前移")
	require.Equal(t, []int{0, 1}, []int{all[0].SortOrder, all[1].SortOrder}, "sort_order 重排为 0,1")
}

func Test_namespaceRepo_FindByName(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: entdb,
	}))

	repo.Create(context.TODO(), &biz.CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
	})
	name, _ := repo.FindByName(context.TODO(), "aaa")
	assert.NotNil(t, name)
	name, _ = repo.FindByName(context.TODO(), "abc-aaa")
	assert.NotNil(t, name)
}

func TestNamespace_GetImagePullSecrets(t *testing.T) {
	ns := &biz.Namespace{}
	assert.NotNil(t, ns.GetImagePullSecrets())
	ns = &biz.Namespace{
		ImagePullSecrets: []string{"a", "b"},
	}
	assert.Len(t, ns.GetImagePullSecrets(), 2)
}

func Test_namespaceRepo_SyncMembers(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: entdb,
	}))
	ns := createNamespace(entdb)
	ns.Update().SetPrivate(true).SaveX(context.TODO())

	res, err := repo.SyncMembers(context.TODO(), ns.ID, []string{"a", "b"})
	assert.Nil(t, err)
	assert.Len(t, res.Members, 2)

	res, err = repo.SyncMembers(context.TODO(), ns.ID, []string{"c"})
	assert.Nil(t, err)
	assert.Len(t, res.Members, 1)
	assert.Equal(t, "c", res.Members[0].Email)

	res, err = repo.SyncMembers(context.TODO(), ns.ID, []string{})
	assert.Nil(t, err)
	assert.Len(t, res.Members, 0)
}

func Test_namespaceRepo_UpdatePrivate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: entdb,
	}))
	ns := createNamespace(entdb)
	ns.Update().SetPrivate(true).SaveX(context.TODO())

	res, _ := repo.SyncMembers(context.TODO(), ns.ID, []string{"a", "b"})
	assert.Len(t, res.Members, 2)

	private, err := repo.UpdatePrivate(context.TODO(), ns.ID, false)
	assert.Nil(t, err)
	assert.False(t, private.Private)

	x := entdb.Member.Query().AllX(context.TODO())
	assert.Len(t, x, 0)

	private, err = repo.UpdatePrivate(context.TODO(), ns.ID, true)
	assert.Nil(t, err)
	assert.True(t, private.Private)
}

func TestToNamespace(t *testing.T) {
	ns := toNamespace(&ent.Namespace{
		CreatorEmail: biz.SuperAdminEmail,
	})

	assert.Equal(t, "超级管理员", ns.CreatorEmail)
	ns = toNamespace(&ent.Namespace{
		CreatorEmail: "abc@qq.com",
	})

	assert.Equal(t, "abc@qq.com", ns.CreatorEmail)
}

func Test_namespaceRepo_Transfer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: entdb,
	}))
	ns := createNamespace(entdb)
	ns.Update().SetPrivate(true).SetCreatorEmail("bbb").SaveX(context.TODO())

	res, err := repo.Transfer(context.TODO(), ns.ID, "aaa")
	assert.Nil(t, err)
	assert.Equal(t, "aaa", res.CreatorEmail)

	_, err = repo.Transfer(context.TODO(), 9999999, "aaa")
	assert.Error(t, err)
}

// TestNamespaceRepo_ListAll 覆盖全量返回 namespace 的端口，cron 同步依赖。
// TestNamespaceRepo_List_NameMatchesProject 首页搜索框的 name 参数除匹配空间名外，
// 还须匹配该空间下的项目名（用户往往记得项目名而记不住空间名）。三条承重断言：
//  1. 项目名命中 → 返回其所属空间；
//  2. 同一空间多个项目命中 → 该空间只出现一次（HasProjectsWith 必须是子查询而非 JOIN，
//     否则行会被放大、分页与 count 双双失真）；
//  3. 项目名命中私有空间 → 非成员/非管理员仍不可见（新增的 OR 分支不得绕过可见性）。
func TestNamespaceRepo_List_NameMatchesProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ctx := context.TODO()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{},
		DB:  entdb,
	}))

	// 公开空间：两个项目名都含 "alpha"，用于验证不放大行数。
	pub := entdb.Namespace.Create().SetCreatorEmail("a@a.c").SetName("ns-public").SaveX(ctx)
	entdb.Project.Create().SetName("alpha-api").SetNamespaceID(pub.ID).SetCreator("").SaveX(ctx)
	entdb.Project.Create().SetName("inner-alpha").SetNamespaceID(pub.ID).SetCreator("").SaveX(ctx)
	// 私有空间：项目名含 "alpha"，但对非成员不可见。
	pri := entdb.Namespace.Create().SetCreatorEmail("b@b.c").SetName("ns-private").SetPrivate(true).SaveX(ctx)
	entdb.Project.Create().SetName("alpha-secret").SetNamespaceID(pri.ID).SetCreator("").SaveX(ctx)

	// ①+② 管理员视角：项目名命中所属空间，且多项目命中同一空间只返回一行。
	res, pag, err := repo.List(ctx, &biz.ListNamespaceInput{
		Page: 1, PageSize: 10, IsAdmin: true, Email: "a@a.c", Name: lo.ToPtr("alpha"),
	})
	assert.NoError(t, err)
	require.Len(t, res, 2, "两个空间各有项目命中；公开空间下两个项目命中不得令其重复出现")
	assert.Equal(t, int32(2), pag.Count, "count 必须与去重后的命中空间数一致")

	// ③ 非成员视角：私有空间的 name 命中同样不泄漏，只剩公开空间。
	res, pag, err = repo.List(ctx, &biz.ListNamespaceInput{
		Page: 1, PageSize: 10, IsAdmin: false, Email: "outsider@x.c", Name: lo.ToPtr("alpha"),
	})
	assert.NoError(t, err)
	require.Len(t, res, 1, "私有空间的项目名命中不得绕过可见性过滤")
	assert.Equal(t, int32(1), pag.Count)
	assert.Equal(t, pub.ID, res[0].ID)

	// 空间名匹配的原有语义不变（回归保护）。
	res, _, err = repo.List(ctx, &biz.ListNamespaceInput{
		Page: 1, PageSize: 10, IsAdmin: true, Email: "a@a.c", Name: lo.ToPtr("ns-public"),
	})
	assert.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, pub.ID, res[0].ID)

	// 无命中：项目名与空间名都不匹配时返回空。
	res, pag, err = repo.List(ctx, &biz.ListNamespaceInput{
		Page: 1, PageSize: 10, IsAdmin: true, Email: "a@a.c", Name: lo.ToPtr("no-such-thing"),
	})
	assert.NoError(t, err)
	assert.Empty(t, res)
	assert.Equal(t, int32(0), pag.Count)
}

func TestNamespaceRepo_ListAll(t *testing.T) {
	entdb, _ := NewSqliteDB()
	t.Cleanup(func() { entdb.Close() })
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{DB: entdb, Cfg: &config.Config{}}))

	entdb.Namespace.Create().SetName("ns-a").SetCreatorEmail("a@b.c").SaveX(context.TODO())
	entdb.Namespace.Create().SetName("ns-b").SetCreatorEmail("a@b.c").SaveX(context.TODO())

	list, err := repo.ListAll(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.ElementsMatch(t, []string{"ns-a", "ns-b"}, []string{list[0].Name, list[1].Name})
}

// TestNamespaceRepo_List_NameMatchesProject_SoftDeletedProject 守护 HasProjectsWith
// 子查询的软删过滤。SoftDeleteMixin 靠 ent Interceptor 注入 deleted_at IS NULL，而
// Interceptor 只遍历生成的查询图，**不会**进入 Has*With 产生的裸 sql.Selector 子查询——
// 谓词里漏写 project.DeletedAtIsNil() 时，已软删项目仍会命中其所属空间（回归曾实测复现）。
// 两条承重断言：① 唯一命中的项目被软删 → 空间不可再被搜到；② 同空间另有一存活项目命中
// → 空间仍可见（证明补的是过滤，不是把子查询整体误杀）。
func TestNamespaceRepo_List_NameMatchesProject_SoftDeletedProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ctx := context.TODO()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{},
		DB:  entdb,
	}))

	// ① 空间名与搜索词刻意不相交（"ns-holder-one" 不含 ghost），确保命中只能来自项目谓词；
	// 该唯一项目软删后，空间不得再被项目名搜到。
	ns := entdb.Namespace.Create().SetCreatorEmail("a@a.c").SetName("ns-holder-one").SaveX(ctx)
	ghost := entdb.Project.Create().SetName("ghost-project").SetNamespaceID(ns.ID).SetCreator("").SaveX(ctx)
	require.NoError(t, entdb.Project.DeleteOneID(ghost.ID).Exec(ctx))

	res, pag, err := repo.List(ctx, &biz.ListNamespaceInput{
		Page: 1, PageSize: 10, IsAdmin: true, Email: "a@a.c", Name: lo.ToPtr("ghost"),
	})
	assert.NoError(t, err)
	assert.Empty(t, res, "软删项目名不得再命中其所属空间")
	assert.Equal(t, int32(0), pag.Count)
	// 空间本身健在：按空间名仍搜得到，证明命中的消失源于项目软删而非空间被误过滤。
	res, _, err = repo.List(ctx, &biz.ListNamespaceInput{
		Page: 1, PageSize: 10, IsAdmin: true, Email: "a@a.c", Name: lo.ToPtr("ns-holder-one"),
	})
	assert.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, ns.ID, res[0].ID)

	// ② 另一空间：一个项目软删 + 一个存活，项目名命中时空间仍须出现一次。
	ns2 := entdb.Namespace.Create().SetCreatorEmail("a@a.c").SetName("ns-mixed").SaveX(ctx)
	dead := entdb.Project.Create().SetName("beta-dead").SetNamespaceID(ns2.ID).SetCreator("").SaveX(ctx)
	entdb.Project.Create().SetName("beta-alive").SetNamespaceID(ns2.ID).SetCreator("").SaveX(ctx)
	require.NoError(t, entdb.Project.DeleteOneID(dead.ID).Exec(ctx))

	res, pag, err = repo.List(ctx, &biz.ListNamespaceInput{
		Page: 1, PageSize: 10, IsAdmin: true, Email: "a@a.c", Name: lo.ToPtr("beta"),
	})
	assert.NoError(t, err)
	require.Len(t, res, 1, "存活项目命中即须返回其空间；软删项目不得放大行数")
	assert.Equal(t, ns2.ID, res[0].ID)
	assert.Equal(t, int32(1), pag.Count)
}

// TestNamespaceRepo_List_RemovedMemberSoftDeleted 守 members 谓词的软删过滤：成员被移出
// 私有空间后（Member.Delete() → 钩子转软删），该用户不得再在列表里看到这个空间。
// 谓词漏写 member.DeletedAtIsNil() 时本测试必须失败（越权回归，曾实测复现）。
func TestNamespaceRepo_List_RemovedMemberSoftDeleted(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ctx := context.TODO()
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))

	// 他人创建的私有空间 + X 是成员 + 另一个公开空间作对照（证明过滤只掐私有那条分支）。
	pri := entdb.Namespace.Create().SetCreatorEmail("owner@a.c").SetName("ns-pri").
		SetPrivate(true).SaveX(ctx)
	mem := entdb.Member.Create().SetNamespaceID(pri.ID).SetEmail("x@a.c").SaveX(ctx)
	pub := entdb.Namespace.Create().SetCreatorEmail("owner@a.c").SetName("ns-pub").SaveX(ctx)

	res, pag, err := repo.List(ctx, &biz.ListNamespaceInput{
		Page: 1, PageSize: 10, IsAdmin: false, Email: "x@a.c",
	})
	assert.NoError(t, err)
	require.Len(t, res, 2, "在册成员应能看到公开空间 + 所属私有空间")
	assert.Equal(t, int32(2), pag.Count)

	// 移出空间。
	require.NoError(t, entdb.Member.DeleteOneID(mem.ID).Exec(ctx))
	_, gerr := entdb.Member.Get(ctx, mem.ID)
	require.Error(t, gerr, "钩子须把成员删除转成软删，否则本测试不判别")

	res, pag, err = repo.List(ctx, &biz.ListNamespaceInput{
		Page: 1, PageSize: 10, IsAdmin: false, Email: "x@a.c",
	})
	assert.NoError(t, err)
	require.Len(t, res, 1, "被移出成员不得再看到私有空间")
	assert.Equal(t, pub.ID, res[0].ID)
	assert.Equal(t, int32(1), pag.Count, "count 也须排除私有空间")
}

// TestNamespaceRepo_Liveness_SoftDeletedProjectsNotActive 守活跃度谓词的软删过滤：
// 项目全部被软删的空间不得再被判为「活跃」。谓词漏写 project.DeletedAtIsNil() 时
// 本测试必须失败（后台统计失真回归，曾实测 Active 多算）。
func TestNamespaceRepo_Liveness_SoftDeletedProjectsNotActive(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ctx := context.TODO()

	// A：唯一项目刚创建（活跃）→ 软删后应回落为僵尸。
	nsA := entdb.Namespace.Create().SetName("ns-a").SetCreatorEmail("a@b.c").SaveX(ctx)
	pA := entdb.Project.Create().SetName("pa").SetNamespaceID(nsA.ID).SetCreator("").SaveX(ctx)
	// B：唯一项目刚创建且保持存活 → 仍为活跃，证明补的是过滤不是整体误杀。
	nsB := entdb.Namespace.Create().SetName("ns-b").SetCreatorEmail("b@b.c").SaveX(ctx)
	entdb.Project.Create().SetName("pb").SetNamespaceID(nsB.ID).SetCreator("").SaveX(ctx)

	// 前置：两个空间都算活跃。
	before, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Page: 1, PageSize: 10, Now: time.Now()})
	require.NoError(t, err)
	assert.Equal(t, biz.AdminLivenessStats{Total: 2, Active: 2}, before.Stats)

	// 软删 A 的项目。
	require.NoError(t, entdb.Project.DeleteOneID(pA.ID).Exec(ctx))

	got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Page: 1, PageSize: 10, Now: time.Now()})
	require.NoError(t, err)
	assert.Equal(t, biz.AdminLivenessStats{Total: 2, Active: 1, Zombie: 1}, got.Stats,
		"项目全被软删的空间不得再算活跃；存活项目所在空间仍须活跃")

	// 分类过滤同样生效：active 只剩 B，zombie 只剩 A。
	gotActive, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{
		Liveness: "active", Page: 1, PageSize: 10, Now: time.Now(),
	})
	require.NoError(t, err)
	require.Len(t, gotActive.Namespaces, 1)
	assert.Equal(t, nsB.ID, gotActive.Namespaces[0].ID)

	gotZombie, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{
		Liveness: "zombie", Page: 1, PageSize: 10, Now: time.Now(),
	})
	require.NoError(t, err)
	require.Len(t, gotZombie.Namespaces, 1)
	assert.Equal(t, nsA.ID, gotZombie.Namespaces[0].ID)
}

// timePtr 构造 time.Time 指针，供边界奇偶性测试的「无项目（nil）vs 有项目」种子区分。
func timePtr(t time.Time) *time.Time { return &t }

// TestNamespaceRepo_ListAdminPage 真 SQL 分页：search/私有过滤 DB 层生效 + 项目边装配
// （供活跃度聚合取项目 UpdatedAt 最大值）+ 分页（id 升序确定性）。分类过滤/统计由 kind 专项覆盖。
func TestNamespaceRepo_ListAdminPage(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ctx := context.TODO()
	ns := entdb.Namespace.Create().SetName("ns-a").SetCreatorEmail("a@b.c").SaveX(ctx)
	entdb.Project.Create().SetName("proj").SetNamespaceID(ns.ID).SetCreator("").SaveX(ctx)
	entdb.Namespace.Create().SetName("ns-b").SetCreatorEmail("b@b.c").SaveX(ctx)
	entdb.Namespace.Create().SetName("ns-c").SetCreatorEmail("c@c.c").SetPrivate(true).SaveX(ctx)

	t.Run("全量返回且计数正确", func(t *testing.T) {
		got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Page: 1, PageSize: 10, Now: time.Now()})
		assert.NoError(t, err)
		require.Len(t, got.Namespaces, 3)
		assert.Equal(t, 3, got.Count)
		// ns-a 有刚创建的项目（活跃），ns-b/ns-c 无项目（僵尸）。
		assert.Equal(t, biz.AdminLivenessStats{Total: 3, Active: 1, Zombie: 2}, got.Stats)
	})

	t.Run("search 过滤", func(t *testing.T) {
		got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Search: "ns-a", Page: 1, PageSize: 10, Now: time.Now()})
		assert.NoError(t, err)
		require.Len(t, got.Namespaces, 1)
		assert.Equal(t, "ns-a", got.Namespaces[0].Name)
		assert.Equal(t, 1, got.Count)
	})

	t.Run("privateOnly 过滤", func(t *testing.T) {
		got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{PrivateOnly: true, Page: 1, PageSize: 10, Now: time.Now()})
		assert.NoError(t, err)
		require.Len(t, got.Namespaces, 1)
		assert.True(t, got.Namespaces[0].Private)
	})

	t.Run("项目边装配", func(t *testing.T) {
		got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Search: "ns-a", Page: 1, PageSize: 10, Now: time.Now()})
		assert.NoError(t, err)
		require.Len(t, got.Namespaces, 1)
		require.Len(t, got.Namespaces[0].Projects, 1)
		assert.Equal(t, "proj", got.Namespaces[0].Projects[0].Name)
	})

	t.Run("分页 id 升序", func(t *testing.T) {
		// page=2 size=2：id 升序下第 [2:3) 行为 ns-c，count 保留全量 3。
		got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Page: 2, PageSize: 2, Now: time.Now()})
		assert.NoError(t, err)
		require.Len(t, got.Namespaces, 1)
		assert.Equal(t, "ns-c", got.Namespaces[0].Name)
		assert.Equal(t, 3, got.Count)
	})

	t.Run("越界页返回空但计数保留", func(t *testing.T) {
		got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Page: 99, PageSize: 10, Now: time.Now()})
		assert.NoError(t, err)
		assert.Empty(t, got.Namespaces)
		assert.Equal(t, 3, got.Count)
	})
}

// TestNamespaceRepo_ListAdminPage_KindFilter 分类过滤 + stats 全量口径：EXISTS 谓词命中正确
// 分类（活跃/低活跃/僵尸含无项目空间），非法分类恒空。
func TestNamespaceRepo_ListAdminPage_KindFilter(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ctx := context.TODO()
	now := time.Now()
	active := entdb.Namespace.Create().SetName("ns-active").SetCreatorEmail("a@b.c").SaveX(ctx)
	entdb.Project.Create().SetName("p1").SetUpdatedAt(now.Add(-10 * 24 * time.Hour)).SetNamespaceID(active.ID).SetCreator("").SaveX(ctx)
	dormant := entdb.Namespace.Create().SetName("ns-dormant").SetCreatorEmail("b@b.c").SaveX(ctx)
	entdb.Project.Create().SetName("p2").SetUpdatedAt(now.Add(-60 * 24 * time.Hour)).SetNamespaceID(dormant.ID).SetCreator("").SaveX(ctx)
	// 无项目空间：从未活跃 → 僵尸。
	entdb.Namespace.Create().SetName("ns-zombie").SetCreatorEmail("c@c.c").SaveX(ctx)

	t.Run("活跃分类命中且 stats 为全量", func(t *testing.T) {
		got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Liveness: "active", Page: 1, PageSize: 10, Now: now})
		assert.NoError(t, err)
		require.Len(t, got.Namespaces, 1)
		assert.Equal(t, "ns-active", got.Namespaces[0].Name)
		// 统计不随分类过滤裁剪（search 命中全量）。
		assert.Equal(t, biz.AdminLivenessStats{Total: 3, Active: 1, Dormant: 1, Zombie: 1}, got.Stats)
		assert.Equal(t, 1, got.Count)
	})

	t.Run("低活跃分类命中一行", func(t *testing.T) {
		got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Liveness: "dormant", Page: 1, PageSize: 10, Now: now})
		assert.NoError(t, err)
		require.Len(t, got.Namespaces, 1)
		assert.Equal(t, "ns-dormant", got.Namespaces[0].Name)
	})

	t.Run("僵尸分类含无项目空间", func(t *testing.T) {
		got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Liveness: "zombie", Page: 1, PageSize: 10, Now: now})
		assert.NoError(t, err)
		require.Len(t, got.Namespaces, 1)
		assert.Equal(t, "ns-zombie", got.Namespaces[0].Name)
	})

	t.Run("非法分类恒空", func(t *testing.T) {
		got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Liveness: "bogus", Page: 1, PageSize: 10, Now: now})
		assert.NoError(t, err)
		assert.Empty(t, got.Namespaces)
		assert.Equal(t, 0, got.Count)
	})
}

// TestNamespaceRepo_ListAdminPage_QueryErrors 注入第 N 次查询失败，逐一覆盖 ListAdminPage
// 各 COUNT/list 错误分支（total 已由 TestNamespaceRepo_ErrorBranches 用 closed DB 覆盖）：
// 无过滤查询序 total→active→dormant→zombie→listAll；有 liveness 过滤时第 5 次为 filtered，
// 第 6 次为 listAll。故 n=2/3/4 覆盖三分类 COUNT 错误，n=5+liveness 覆盖 filtered COUNT 错误，
// n=5 无过滤覆盖 listAll 错误。
func TestNamespaceRepo_ListAdminPage_QueryErrors(t *testing.T) {
	cases := []struct {
		name     string
		failAt   int
		liveness string
	}{
		{name: "active count error", failAt: 2},
		{name: "dormant count error", failAt: 3},
		{name: "zombie count error", failAt: 4},
		{name: "filtered count error", failAt: 5, liveness: "active"},
		{name: "list page error", failAt: 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var n int
			repo, entdb := newNsRepoWithIntercept(t, entgo.InterceptFunc(func(next entgo.Querier) entgo.Querier {
				return entgo.QuerierFunc(func(ctx context.Context, q entgo.Query) (entgo.Value, error) {
					n++
					if n == tc.failAt {
						return nil, errors.New("inject namespace query error")
					}
					return next.Query(ctx, q)
				})
			}))
			ns := entdb.Namespace.Create().SetName("ns-err").SetCreatorEmail("e@x.y").SaveX(context.TODO())
			entdb.Project.Create().SetName("p-err").SetNamespaceID(ns.ID).SetCreator("").SaveX(context.TODO())
			_, err := repo.ListAdminPage(context.TODO(), &biz.AdminListPageQuery{Liveness: tc.liveness, Page: 1, PageSize: 10, Now: time.Now()})
			require.Error(t, err)
			assert.ErrorContains(t, err, "inject namespace query error")
		})
	}
}

// TestNamespaceRepo_ListAdminPage_BoundaryParity 边界奇偶性守护：命名空间活跃度分类 SQL
// EXISTS 谓词 == Go ClassifyLiveness(MAX(project.updated_at))。种入跨边界命名空间
// （now-31d±1s、now-90d、无项目），断言 stats 四分类计数与各 kind 命中行数一致。
func TestNamespaceRepo_ListAdminPage_BoundaryParity(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ctx := context.TODO()
	now := time.Now()
	seeds := []struct {
		name string
		ts   *time.Time // nil = 无项目（零值时间 → 僵尸）
	}{
		{"s-active-1", timePtr(now.Add(-1 * 24 * time.Hour))},
		{"s-active-2", timePtr(now.Add(-31*24*time.Hour + time.Second))},
		{"s-dormant-1", timePtr(now.Add(-31 * 24 * time.Hour))},
		{"s-dormant-2", timePtr(now.Add(-60 * 24 * time.Hour))},
		{"s-zombie-1", timePtr(now.Add(-90 * 24 * time.Hour))},
		{"s-zombie-2", timePtr(now.Add(-120 * 24 * time.Hour))},
		{"s-noproj", nil},
	}
	want := map[string]int{"active": 0, "dormant": 0, "zombie": 0}
	for _, s := range seeds {
		ns := entdb.Namespace.Create().SetName(s.name).SetCreatorEmail("a@b.c").SaveX(ctx)
		var lastActive time.Time
		if s.ts != nil {
			entdb.Project.Create().SetName("proj").SetUpdatedAt(*s.ts).SetNamespaceID(ns.ID).SetCreator("").SaveX(ctx)
			lastActive = *s.ts
		}
		want[string(biz.ClassifyLiveness(lastActive, now))]++
	}

	page, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Page: 1, PageSize: 20, Now: now})
	assert.NoError(t, err)
	assert.Equal(t, biz.AdminLivenessStats{Total: len(seeds), Active: want["active"], Dormant: want["dormant"], Zombie: want["zombie"]}, page.Stats)
	for _, kind := range []string{"active", "dormant", "zombie"} {
		got, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Liveness: kind, Page: 1, PageSize: 20, Now: now})
		assert.NoError(t, err)
		assert.Len(t, got.Namespaces, want[kind], "kind=%s SQL 命中行数应等于 Go 分类计数", kind)
	}
}

// TestNamespaceRepo_UpdateImagePullSecrets 覆盖仅回写 imagePullSecrets 列表的端口。
func TestNamespaceRepo_UpdateImagePullSecrets(t *testing.T) {
	entdb, _ := NewSqliteDB()
	t.Cleanup(func() { entdb.Close() })
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{DB: entdb, Cfg: &config.Config{}}))

	ns := entdb.Namespace.Create().SetName("ns-a").SetCreatorEmail("a@b.c").SaveX(context.TODO())
	assert.NoError(t, repo.UpdateImagePullSecrets(context.TODO(), ns.ID, []string{"s1", "s2"}))

	got := entdb.Namespace.GetX(context.TODO(), ns.ID)
	assert.Equal(t, []string{"s1", "s2"}, got.ImagePullSecrets)
}

// newNsRepo 基于 sqlite 构造 namespaceRepo。
func newNsRepo(t *testing.T) (*namespaceRepo, *ent.Client) {
	entdb, err := NewSqliteDB()
	require.NoError(t, err)
	t.Cleanup(func() { entdb.Close() })
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{DB: entdb, Cfg: &config.Config{}}))
	return repo.(*namespaceRepo), entdb
}

// TestToMember_ToFavorite_Nil 覆盖 nil 安全转换分支。
func TestToMember_ToFavorite_Nil(t *testing.T) {
	assert.Nil(t, toMember(nil))
	assert.Nil(t, toFavorite(nil))
	assert.Nil(t, toNamespace(nil))
}

// TestNamespaceRepo_ErrorBranches 用 closed DB 触发各 repo 方法的查询错误分支。
func TestNamespaceRepo_ErrorBranches(t *testing.T) {
	closed := NewDataImpl(&NewDataParams{DB: mustClosedDB(t), Cfg: &config.Config{}})
	repo := NewNamespaceRepo(closed).(*namespaceRepo)
	ctx := context.TODO()

	t.Run("List query error", func(t *testing.T) {
		_, _, err := repo.List(ctx, &biz.ListNamespaceInput{Page: 1, PageSize: 10})
		assert.Error(t, err)
	})

	t.Run("Show query error", func(t *testing.T) {
		_, err := repo.Show(ctx, 1)
		assert.Error(t, err)
	})

	t.Run("Update query error", func(t *testing.T) {
		_, err := repo.Update(ctx, &biz.UpdateNamespaceInput{ID: 1})
		assert.Error(t, err)
	})

	t.Run("ListAll query error", func(t *testing.T) {
		_, err := repo.ListAll(ctx)
		assert.Error(t, err)
	})

	t.Run("ListAdminPage query error", func(t *testing.T) {
		_, err := repo.ListAdminPage(ctx, &biz.AdminListPageQuery{Page: 1, PageSize: 10})
		assert.Error(t, err)
	})

	t.Run("Delete query error", func(t *testing.T) {
		err := repo.Delete(ctx, 1)
		assert.Error(t, err)
	})

	t.Run("Favorite exist query error", func(t *testing.T) {
		err := repo.Favorite(ctx, &biz.FavoriteNamespaceInput{NamespaceID: 1, UserEmail: "e", Favorite: true})
		assert.Error(t, err)
	})

	t.Run("Favorite delete query error", func(t *testing.T) {
		err := repo.Favorite(ctx, &biz.FavoriteNamespaceInput{NamespaceID: 1, UserEmail: "e", Favorite: false})
		assert.Error(t, err)
	})

	t.Run("SyncMembers WithTx error", func(t *testing.T) {
		_, err := repo.SyncMembers(ctx, 1, []string{"a"})
		assert.Error(t, err)
	})

	t.Run("UpdatePrivate WithTx error", func(t *testing.T) {
		_, err := repo.UpdatePrivate(ctx, 1, false)
		assert.Error(t, err)
	})
}

// TestNamespaceRepo_SyncMembers_MissingNamespace 覆盖 tx 内查询不存在的 namespace 报错。
func TestNamespaceRepo_SyncMembers_MissingNamespace(t *testing.T) {
	repo, _ := newNsRepo(t)
	_, err := repo.SyncMembers(context.TODO(), 999999, []string{"a"})
	assert.Error(t, err)
}

// TestNamespaceRepo_UpdatePrivate_MissingNamespace 覆盖 tx 内 Get 不存在的 namespace 报错。
func TestNamespaceRepo_UpdatePrivate_MissingNamespace(t *testing.T) {
	repo, _ := newNsRepo(t)
	_, err := repo.UpdatePrivate(context.TODO(), 999999, true)
	assert.Error(t, err)
}

// TestNamespaceRepo_Transfer_SameEmailSkips 覆盖 CreatorEmail 相同 → 跳过更新的分支。
func TestNamespaceRepo_Transfer_SameEmailSkips(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ns := entdb.Namespace.Create().SetName("abc-x").SetCreatorEmail("me@x.y").SaveX(context.TODO())
	res, err := repo.Transfer(context.TODO(), ns.ID, "me@x.y")
	assert.NoError(t, err)
	assert.Equal(t, "me@x.y", res.CreatorEmail)
}

// newNsFault 基于故障注入驱动构造 namespaceRepo，返回 repo 与驱动句柄。
// 注入点在 setup 完成后由测试显式 fd.Arm() 激活，
// 计数从 0 开始，qAfter/eAfter 直接对应 repo 方法内的第 N 条查询/写入
// （见各测试注释中的 SQL 序列）。
func newNsFault(t *testing.T, qAfter, eAfter int32) (*namespaceRepo, *failDriver) {
	t.Helper()
	client, fd := newFailDB(t, qAfter, eAfter)
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{DB: client, Cfg: &config.Config{}})).(*namespaceRepo)
	return repo, fd
}

// TestNamespaceRepo_Transfer_SaveError 覆盖 Get 成功后 Update Save 失败的错误分支
// SQL 序列：Get(namespace)→UPDATE creator_email。
// 第 1 条写入（eAfter=0）即该 UPDATE，注入成功。
func TestNamespaceRepo_Transfer_SaveError(t *testing.T) {
	ctx := context.TODO()
	repo, fd := newNsFault(t, -1, 0)
	ns := repo.data.DB().Namespace.Create().SetName("tr-fail").SetCreatorEmail("o@x").SaveX(ctx)
	fd.Arm()
	_, err := repo.Transfer(ctx, ns.ID, "new@x")
	assert.Error(t, err)
}

// TestNamespaceRepo_SyncMembers_InternalErrors 覆盖事务内错误分支。
// SQL 序列（del=[a], add=[c]，两段都真实执行）：
//
//	Q1 First(namespace) → Q2 WithMembers → Q3 CreateBulk INSERT(RETURNING)
//	→ E1 DELETE members（软删，限定 namespace+email）
//
// qAfter=2 → 第 3 条查询（CreateBulk）；eAfter=0 → 第 1 条写入（成员删除）。
func TestNamespaceRepo_SyncMembers_InternalErrors(t *testing.T) {
	ctx := context.TODO()

	// setup 建 ns + 成员 a；SyncMembers 换成 c → del=[a], add=[c]。
	setup := func(t *testing.T, repo *namespaceRepo) int {
		ns := repo.data.DB().Namespace.Create().SetName("sm-fail").SetCreatorEmail("o@x").SaveX(ctx)
		repo.data.DB().Member.Create().SetEmail("a@x").SetNamespaceID(ns.ID).SaveX(ctx)
		return ns.ID
	}

	t.Run("CreateBulk query error", func(t *testing.T) {
		repo, fd := newNsFault(t, 2, -1)
		id := setup(t, repo)
		fd.Arm()
		_, err := repo.SyncMembers(ctx, id, []string{"c@x"})
		assert.Error(t, err)
	})

	t.Run("MemberDelete exec error", func(t *testing.T) {
		repo, fd := newNsFault(t, -1, 0)
		id := setup(t, repo)
		fd.Arm()
		_, err := repo.SyncMembers(ctx, id, []string{"c@x"})
		assert.Error(t, err)
	})
}

// TestNamespaceRepo_UpdatePrivate_MemberDeleteError 覆盖私有转公开时删除成员失败的错误分支
// SQL 序列：Get(namespace)→Delete(members)→Save(namespace)。
// 第 1 条写入（eAfter=0）即 Member.Delete，注入成功。
func TestNamespaceRepo_UpdatePrivate_MemberDeleteError(t *testing.T) {
	ctx := context.TODO()
	repo, fd := newNsFault(t, -1, 0)
	ns := repo.data.DB().Namespace.Create().SetName("up-fail").SetCreatorEmail("o@x").SaveX(ctx)
	repo.data.DB().Member.Create().SetEmail("a@x").SetNamespaceID(ns.ID).SaveX(ctx)
	fd.Arm()
	_, err := repo.UpdatePrivate(ctx, ns.ID, false)
	assert.Error(t, err)
}

// TestNamespaceRepo_UpdatePrivate_SaveError 覆盖 Member.Delete 成功后 up.Save 失败的错误分支
// SQL 序列：Get(namespace)→Delete(members)→Save(namespace)。
// eAfter=1 → 第 2 条写入即 up.Save 的 UPDATE namespaces，注入成功。
func TestNamespaceRepo_UpdatePrivate_SaveError(t *testing.T) {
	ctx := context.TODO()
	repo, fd := newNsFault(t, -1, 1)
	ns := repo.data.DB().Namespace.Create().SetName("upsave-fail").SetCreatorEmail("o@x").SaveX(ctx)
	repo.data.DB().Member.Create().SetEmail("a@x").SetNamespaceID(ns.ID).SaveX(ctx)
	fd.Arm()
	_, err := repo.UpdatePrivate(ctx, ns.ID, false)
	assert.Error(t, err)
}

// TestNamespaceRepo_Delete_ProjectDeleteError 覆盖级联删除项目失败的错误分支
// SQL 序列：First(namespace)+WithProjects→Delete(projects)→Delete(namespace)。
// 第 1 条写入（eAfter=0）即 Project.Delete，注入成功。
func TestNamespaceRepo_Delete_ProjectDeleteError(t *testing.T) {
	ctx := context.TODO()
	repo, fd := newNsFault(t, -1, 0)
	ns := repo.data.DB().Namespace.Create().SetName("del-fail").SetCreatorEmail("o@x").SaveX(ctx)
	repo.data.DB().Project.Create().SetName("p").SetNamespaceID(ns.ID).SetCreator("").SetGitProjectID(1).SaveX(ctx)
	fd.Arm()
	require.Error(t, repo.Delete(ctx, ns.ID))
}

// TestNamespaceRepo_List_CountError 覆盖 List 中 All 成功后 Count 失败的错误分支
// （namespace.go Count→errs.Wrap "count namespaces"）。SQL 序列：
//
//	Q1 SELECT namespaces(LIMIT/OFFSET) → Q2 SELECT favorites → Q3 SELECT members
//	→ Q4 SELECT projects → Q5 COUNT(*)
//
// qAfter=4 → 第 5 条查询即 Count，注入失败。
func TestNamespaceRepo_List_CountError(t *testing.T) {
	ctx := context.TODO()
	repo, fd := newNsFault(t, 4, -1)
	repo.data.DB().Namespace.Create().SetName("cnt-fail").SetCreatorEmail("o@x").SaveX(ctx)
	fd.Arm()
	_, _, err := repo.List(ctx, &biz.ListNamespaceInput{Page: 1, PageSize: 10, IsAdmin: true})
	assert.ErrorContains(t, err, "count namespaces")
}

// TestNamespaceRepo_List_ReturnsImagePullSecrets 覆盖 List 返回 imagePullSecrets
// （Select 必须含该列，避免 toNamespace 映射恒为空）。
func TestNamespaceRepo_List_ReturnsImagePullSecrets(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ns := entdb.Namespace.Create().SetName("ns-secrets").SetCreatorEmail("a@b.c").SetImagePullSecrets([]string{"s1", "s2"}).SaveX(context.TODO())

	res, _, err := repo.List(context.TODO(), &biz.ListNamespaceInput{Page: 1, PageSize: 10, IsAdmin: true})
	require.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, ns.ID, res[0].ID)
	assert.Equal(t, []string{"s1", "s2"}, res[0].ImagePullSecrets)
}

// Test_namespaceRepo_UpdateConfig_Combined 覆盖单事务内四项配置（描述/私有/成员/转让）一次提交全生效。
func Test_namespaceRepo_UpdateConfig_Combined(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ns := createNamespace(entdb)
	ns.Update().SetPrivate(true).SetDescription("old desc").SetCreatorEmail("old@x.y").SaveX(context.TODO())
	entdb.Member.Create().SetEmail("a@x.y").SetNamespaceID(ns.ID).SaveX(context.TODO())

	res, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{
		ID:            ns.ID,
		Description:   lo.ToPtr("new desc"),
		Private:       lo.ToPtr(true),
		Emails:        []string{"a@x.y", "b@x.y"},
		NewAdminEmail: "new@x.y",
	})
	require.NoError(t, err)
	assert.Equal(t, "new desc", res.Description)
	assert.True(t, res.Private)
	assert.Equal(t, "new@x.y", res.CreatorEmail)
	require.Len(t, res.Members, 2)
	assert.ElementsMatch(t, []string{"a@x.y", "b@x.y"}, []string{res.Members[0].Email, res.Members[1].Email})
}

// Test_namespaceRepo_UpdateConfig_Partial 覆盖仅更新部分字段：未传字段保持不变。
func Test_namespaceRepo_UpdateConfig_Partial(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ns := entdb.Namespace.Create().SetName("partial-ns").SetCreatorEmail("me@x.y").SetPrivate(true).SetDescription("keep").SaveX(context.TODO())
	entdb.Member.Create().SetEmail("keep@x.y").SetNamespaceID(ns.ID).SaveX(context.TODO())

	// 只传描述：私有/成员/管理员均不变
	res, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{
		ID:          ns.ID,
		Description: lo.ToPtr("changed"),
	})
	require.NoError(t, err)
	assert.Equal(t, "changed", res.Description)
	assert.True(t, res.Private)
	assert.Equal(t, "me@x.y", res.CreatorEmail)
	require.Len(t, res.Members, 1)
	assert.Equal(t, "keep@x.y", res.Members[0].Email)
}

// Test_namespaceRepo_UpdateConfig_PrivateFalseClears 覆盖 private=false 转公开清空成员（对齐 UpdatePrivate 规则）。
func Test_namespaceRepo_UpdateConfig_PrivateFalseClears(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ns := entdb.Namespace.Create().SetName("clear-ns").SetCreatorEmail("me@x.y").SetPrivate(true).SaveX(context.TODO())
	entdb.Member.Create().SetEmail("a@x.y").SetNamespaceID(ns.ID).SaveX(context.TODO())

	res, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{
		ID:      ns.ID,
		Private: lo.ToPtr(false),
	})
	require.NoError(t, err)
	assert.False(t, res.Private)
	assert.Len(t, res.Members, 0)
	got := entdb.Member.Query().CountX(context.TODO())
	assert.Zero(t, got)
}

// Test_namespaceRepo_UpdateConfig_PrivateFalseWithEmails 覆盖 private=false + 新名单：清空后按名单重建。
func Test_namespaceRepo_UpdateConfig_PrivateFalseWithEmails(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ns := entdb.Namespace.Create().SetName("rebuild-ns").SetCreatorEmail("me@x.y").SetPrivate(true).SaveX(context.TODO())
	entdb.Member.Create().SetEmail("old@x.y").SetNamespaceID(ns.ID).SaveX(context.TODO())

	res, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{
		ID:      ns.ID,
		Private: lo.ToPtr(false),
		Emails:  []string{"new@x.y"},
	})
	require.NoError(t, err)
	assert.False(t, res.Private)
	require.Len(t, res.Members, 1)
	assert.Equal(t, "new@x.y", res.Members[0].Email)
	got := entdb.Member.Query().AllX(context.TODO())
	require.Len(t, got, 1)
	assert.Equal(t, "new@x.y", got[0].Email)
}

// Test_namespaceRepo_UpdateConfig_EmailsEmptyClears 覆盖 Emails 非 nil 空切片 → 成员清空（对齐 SyncMembers 空名单）。
func Test_namespaceRepo_UpdateConfig_EmailsEmptyClears(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ns := entdb.Namespace.Create().SetName("empty-ns").SetCreatorEmail("me@x.y").SetPrivate(true).SaveX(context.TODO())
	entdb.Member.Create().SetEmail("a@x.y").SetNamespaceID(ns.ID).SaveX(context.TODO())

	res, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{
		ID:      ns.ID,
		Emails:  []string{},
		Private: lo.ToPtr(true),
	})
	require.NoError(t, err)
	assert.True(t, res.Private)
	assert.Len(t, res.Members, 0)
}

// Test_namespaceRepo_UpdateConfig_SameAdminSkips 覆盖 newAdminEmail 与当前 creator 相同 → 不转让。
func Test_namespaceRepo_UpdateConfig_SameAdminSkips(t *testing.T) {
	repo, entdb := newNsRepo(t)
	ns := entdb.Namespace.Create().SetName("same-ns").SetCreatorEmail("me@x.y").SaveX(context.TODO())

	res, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{
		ID:            ns.ID,
		NewAdminEmail: "me@x.y",
	})
	require.NoError(t, err)
	assert.Equal(t, "me@x.y", res.CreatorEmail)
}

// Test_namespaceRepo_UpdateConfig_MissingNamespace 覆盖 tx 内查询不存在的 namespace 报错。
func Test_namespaceRepo_UpdateConfig_MissingNamespace(t *testing.T) {
	repo, _ := newNsRepo(t)
	_, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{ID: 999999, Description: lo.ToPtr("x")})
	assert.Error(t, err)
}

// TestNamespaceRepo_UpdateConfig_ErrorBranch 用 closed DB 覆盖查询错误分支。
func TestNamespaceRepo_UpdateConfig_ErrorBranch(t *testing.T) {
	repo, _ := newNsRepo(t)
	_ = repo.data.DB().Close()
	_, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{ID: 1, Description: lo.ToPtr("x")})
	assert.Error(t, err)
}

// newNsRepoWithHook 构造一个挂了 mutation hook 的 namespaceRepo（真实 sqlite），
// 用于在 UpdateConfig 单事务中间注入确定性错误，补齐 closed DB 无法触达的中间错误分支。
func newNsRepoWithHook(t *testing.T, hook func(next entgo.Mutator) entgo.Mutator) (*namespaceRepo, *ent.Client) {
	t.Helper()
	entdb, err := NewSqliteDB()
	require.NoError(t, err)
	t.Cleanup(func() { entdb.Close() })
	if hook != nil {
		entdb.Use(hook)
	}
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{DB: entdb, Cfg: &config.Config{}}))
	return repo.(*namespaceRepo), entdb
}

// newNsRepoWithIntercept 构造一个挂了 query 拦截器的 namespaceRepo（真实 sqlite），
// 用于注入成员查询失败。
func newNsRepoWithIntercept(t *testing.T, intercept entgo.Interceptor) (*namespaceRepo, *ent.Client) {
	t.Helper()
	entdb, err := NewSqliteDB()
	require.NoError(t, err)
	t.Cleanup(func() { entdb.Close() })
	if intercept != nil {
		entdb.Intercept(intercept)
	}
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{DB: entdb, Cfg: &config.Config{}}))
	return repo.(*namespaceRepo), entdb
}

// Test_namespaceRepo_UpdateConfig_SaveError 注入 namespace 更新失败，覆盖 up.Save 错误分支。
func Test_namespaceRepo_UpdateConfig_SaveError(t *testing.T) {
	repo, entdb := newNsRepoWithHook(t, func(next entgo.Mutator) entgo.Mutator {
		return entgo.MutateFunc(func(ctx context.Context, m entgo.Mutation) (entgo.Value, error) {
			if mm, ok := m.(*ent.NamespaceMutation); ok && mm.Op() == entgo.OpUpdateOne {
				return nil, errors.New("inject save error")
			}
			return next.Mutate(ctx, m)
		})
	})
	ns := entdb.Namespace.Create().SetName("save-err").SetCreatorEmail("me@x.y").SaveX(context.TODO())
	_, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{ID: ns.ID, Description: lo.ToPtr("x")})
	require.Error(t, err)
	assert.ErrorContains(t, err, "inject save error")
}

// Test_namespaceRepo_UpdateConfig_DeleteMembersError 注入转公开清空成员失败，覆盖 Private=false 时 Delete 错误分支。
func Test_namespaceRepo_UpdateConfig_DeleteMembersError(t *testing.T) {
	repo, entdb := newNsRepoWithHook(t, func(next entgo.Mutator) entgo.Mutator {
		return entgo.MutateFunc(func(ctx context.Context, m entgo.Mutation) (entgo.Value, error) {
			if mm, ok := m.(*ent.MemberMutation); ok && mm.Op() == entgo.OpDelete {
				return nil, errors.New("inject member delete error")
			}
			return next.Mutate(ctx, m)
		})
	})
	ns := entdb.Namespace.Create().SetName("del-err").SetCreatorEmail("me@x.y").SetPrivate(true).SaveX(context.TODO())
	_, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{ID: ns.ID, Private: lo.ToPtr(false)})
	require.Error(t, err)
	assert.ErrorContains(t, err, "inject member delete error")
}

// Test_namespaceRepo_UpdateConfig_MembersQueryError 注入成员名单查询失败，覆盖差量同步 Query 错误分支。
func Test_namespaceRepo_UpdateConfig_MembersQueryError(t *testing.T) {
	repo, entdb := newNsRepoWithIntercept(t, entgo.InterceptFunc(func(next entgo.Querier) entgo.Querier {
		return entgo.QuerierFunc(func(ctx context.Context, q entgo.Query) (entgo.Value, error) {
			if _, ok := q.(*ent.MemberQuery); ok {
				return nil, errors.New("inject member query error")
			}
			return next.Query(ctx, q)
		})
	}))
	ns := entdb.Namespace.Create().SetName("q-err").SetCreatorEmail("me@x.y").SaveX(context.TODO())
	_, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{ID: ns.ID, Emails: []string{"a@x.y"}})
	require.Error(t, err)
	assert.ErrorContains(t, err, "inject member query error")
}

// Test_namespaceRepo_UpdateConfig_CreateBulkError 注入成员批量创建失败，覆盖 CreateBulk 错误分支。
func Test_namespaceRepo_UpdateConfig_CreateBulkError(t *testing.T) {
	repo, entdb := newNsRepoWithHook(t, func(next entgo.Mutator) entgo.Mutator {
		return entgo.MutateFunc(func(ctx context.Context, m entgo.Mutation) (entgo.Value, error) {
			if mm, ok := m.(*ent.MemberMutation); ok && mm.Op() == entgo.OpCreate {
				return nil, errors.New("inject member create error")
			}
			return next.Mutate(ctx, m)
		})
	})
	ns := entdb.Namespace.Create().SetName("create-err").SetCreatorEmail("me@x.y").SaveX(context.TODO())
	_, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{ID: ns.ID, Emails: []string{"a@x.y"}})
	require.Error(t, err)
	assert.ErrorContains(t, err, "inject member create error")
}

// Test_namespaceRepo_UpdateConfig_DeleteMembersDiffError 注入差量删除成员失败，覆盖 Emails 差量同步 Delete 错误分支。
func Test_namespaceRepo_UpdateConfig_DeleteMembersDiffError(t *testing.T) {
	repo, entdb := newNsRepoWithHook(t, func(next entgo.Mutator) entgo.Mutator {
		return entgo.MutateFunc(func(ctx context.Context, m entgo.Mutation) (entgo.Value, error) {
			if mm, ok := m.(*ent.MemberMutation); ok && mm.Op() == entgo.OpDelete {
				return nil, errors.New("inject member delete error")
			}
			return next.Mutate(ctx, m)
		})
	})
	ns := entdb.Namespace.Create().SetName("del-diff-err").SetCreatorEmail("me@x.y").SetPrivate(true).SaveX(context.TODO())
	entdb.Member.Create().SetEmail("a@x.y").SetNamespaceID(ns.ID).SaveX(context.TODO())
	_, err := repo.UpdateConfig(context.TODO(), &biz.UpdateConfigInput{ID: ns.ID, Emails: []string{}})
	require.Error(t, err)
	assert.ErrorContains(t, err, "inject member delete error")
}

// createNamedProject 建一个指定名字的项目（createProject 固定用 testProject，同名项目
// 无法区分，恢复相关用例需要按名定位）。
func createNamedProject(entdb *ent.Client, nsID int, name string) *ent.Project {
	return entdb.Project.Create().
		SetGitProjectID(1).
		SetConfig("").
		SetCreator("").
		SetName(name).
		SetNamespaceID(nsID).
		SaveX(context.TODO())
}

// TestNamespaceRepo_Delete_CascadeBatchFlag 删除空间时级联项目被软删并置批次标识
// deleted_with_namespace=true（恢复时判定"同批"的依据）；其他空间的项目不受影响，
// 标识保持 false。
func TestNamespaceRepo_Delete_CascadeBatchFlag(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	ctx := context.TODO()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{NsPrefix: "abc"},
		DB:  entdb,
	}))

	ns := createNamespace(entdb)
	other := createNamespace(entdb)
	p1 := createNamedProject(entdb, ns.ID, "p1")
	p2 := createNamedProject(entdb, ns.ID, "p2")
	p3 := createNamedProject(entdb, other.ID, "p3")

	require.NoError(t, repo.Delete(ctx, ns.ID))

	skip := mixin.SkipSoftDelete(ctx)
	deletedNs := entdb.Namespace.Query().Where(namespace.ID(ns.ID)).OnlyX(skip)
	require.NotNil(t, deletedNs.DeletedAt)

	for _, id := range []int{p1.ID, p2.ID} {
		got := entdb.Project.Query().Where(project.ID(id)).OnlyX(skip)
		require.NotNil(t, got.DeletedAt, "级联项目 %d 应被软删", id)
		assert.True(t, got.DeletedWithNamespace, "级联项目 %d 应带批次标识", id)
		// 同一时刻下线只是一层可读语义，批次判定已改用上面的显式标识。
		assert.Equal(t, *deletedNs.DeletedAt, *got.DeletedAt, "级联项目 %d 应与空间共享同一 deleted_at", id)
	}

	// 非本空间的项目不得被误删，也不得被标记。
	untouched := entdb.Project.Query().Where(project.ID(p3.ID)).OnlyX(ctx)
	assert.Nil(t, untouched.DeletedAt)
	assert.False(t, untouched.DeletedWithNamespace)
}

// TestNamespaceRepo_RestoreDeleted_BatchPrecision 恢复空间只还原"随空间一起被删"的项目：
// 用户早先单独删除的项目 deleted_at 更早，必须保持在软删态——否则恢复空间会把早已
// 主动删除的项目静默复活成无部署资源的幽灵记录。这是本功能最容易写错的一条不变量。
func TestNamespaceRepo_RestoreDeleted_BatchPrecision(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	ctx := context.TODO()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))

	ns := createNamespace(entdb)
	cascaded := createNamedProject(entdb, ns.ID, "cascaded")
	earlier := createNamedProject(entdb, ns.ID, "earlier-by-user")

	// 用户先单独删除 earlier（走软删钩子，deleted_at = 此刻），再删整个空间。
	require.NoError(t, entdb.Project.DeleteOneID(earlier.ID).Exec(ctx))
	require.NoError(t, repo.Delete(ctx, ns.ID))

	restoredNames, err := repo.RestoreDeleted(ctx, ns.ID)
	require.NoError(t, err)
	// 回传的恢复名单就是审计日志的口径：只含随空间级联删除的那批，早先被用户单独删除的
	// 项目（earlier-by-user）不得出现在里面。
	assert.Equal(t, []string{"cascaded"}, restoredNames, "恢复名单必须精确等于级联批次")

	skip := mixin.SkipSoftDelete(ctx)
	assert.Nil(t, entdb.Namespace.Query().Where(namespace.ID(ns.ID)).OnlyX(skip).DeletedAt, "空间应已恢复")
	assert.Nil(t, entdb.Project.Query().Where(project.ID(cascaded.ID)).OnlyX(skip).DeletedAt,
		"随空间一起删除的项目应被恢复")
	assert.NotNil(t, entdb.Project.Query().Where(project.ID(earlier.ID)).OnlyX(skip).DeletedAt,
		"用户早先单独删除的项目不得被恢复")

	// 不变式：标记为 true ⟺ 项目正处于随空间级联软删的状态。
	restored := entdb.Project.Query().Where(project.ID(cascaded.ID)).OnlyX(skip)
	assert.False(t, restored.DeletedWithNamespace, "恢复后批次标识必须清零")
	assert.False(t, entdb.Project.Query().Where(project.ID(earlier.ID)).OnlyX(skip).DeletedWithNamespace,
		"用户单独删除的项目从不带批次标识")
}

// TestNamespaceRepo_RestoreDeleted_SameSecondCollision 核心回归：批次归属不再由 deleted_at
// 的秒级相等推断。
//
// deleted_at 列是 MySQL `datetime`（= datetime(0)，秒级），若同一秒内先单独删了某项目、再删
// 它所属空间，两行落库后 deleted_at 完全相同——旧实现（谓词 deleted_at EQ 空间的 deleted_at）
// 会把用户早已主动删除的项目一并复活成 deploy_status 未知的幽灵记录。
//
// 本用例显式把单独删除的项目 deleted_at 覆写成与空间**逐字相等**的值来复现该碰撞（不依赖
// 测试库的时间精度，sqlite 保留亚秒、MySQL 截断到秒，两者行为一致地命中同一分支），断言它
// 在恢复空间时保持软删——旧实现在此必然失败。
func TestNamespaceRepo_RestoreDeleted_SameSecondCollision(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	ctx := context.TODO()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))

	ns := createNamespace(entdb)
	cascaded := createNamedProject(entdb, ns.ID, "cascaded")
	solo := createNamedProject(entdb, ns.ID, "solo-deleted-earlier")

	// 用户先单独删除 solo，再删整个空间。
	require.NoError(t, entdb.Project.DeleteOneID(solo.ID).Exec(ctx))
	require.NoError(t, repo.Delete(ctx, ns.ID))

	skip := mixin.SkipSoftDelete(ctx)
	nsDeletedAt := *entdb.Namespace.Query().Where(namespace.ID(ns.ID)).OnlyX(skip).DeletedAt
	// 复现"同秒"：把 solo 的删除时间强行对齐到与空间逐字相同。
	entdb.Project.UpdateOneID(solo.ID).
		SetDeletedAt(nsDeletedAt).
		SaveX(skip)
	require.Equal(t, nsDeletedAt,
		*entdb.Project.Query().Where(project.ID(solo.ID)).OnlyX(skip).DeletedAt,
		"前置：solo 与空间必须处于同一 deleted_at（碰撞已构造）")

	restored, err := repo.RestoreDeleted(ctx, ns.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"cascaded"}, restored,
		"恢复名单不得含 deleted_at 与空间同值的单独删除项目（名单与 UPDATE 同谓词）")

	assert.Nil(t, entdb.Project.Query().Where(project.ID(cascaded.ID)).OnlyX(skip).DeletedAt,
		"级联项目应被恢复")
	assert.NotNil(t, entdb.Project.Query().Where(project.ID(solo.ID)).OnlyX(skip).DeletedAt,
		"deleted_at 与空间同值的单独删除项目不得被恢复（批次判定不得依赖时间戳）")
}

// TestNamespaceRepo_RestoreDeleted_RepeatedCycles 反复"删除→恢复"多轮后批次标识不得跨轮
// 污染：每轮级联只标记**当轮仍存活**的项目，用户单独删除的项目跨轮次始终保持软删且无标记。
// 场景矩阵里最容易漏的一条——单轮通过不代表状态机收敛。
func TestNamespaceRepo_RestoreDeleted_RepeatedCycles(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	ctx := context.TODO()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))

	ns := createNamespace(entdb)
	alive := createNamedProject(entdb, ns.ID, "alive")
	// 用户单独删除的项目：此后无论空间删/恢复多少轮，都不得被复活或被打上批次标记。
	solo := createNamedProject(entdb, ns.ID, "solo")
	require.NoError(t, entdb.Project.DeleteOneID(solo.ID).Exec(ctx))

	skip := mixin.SkipSoftDelete(ctx)
	for round := 1; round <= 3; round++ {
		require.NoError(t, repo.Delete(ctx, ns.ID), "第 %d 轮：删除空间", round)
		restoredNames, err := repo.RestoreDeleted(ctx, ns.ID)
		require.NoError(t, err, "第 %d 轮：恢复空间", round)
		assert.Equal(t, []string{"alive"}, restoredNames,
			"第 %d 轮：恢复名单只含当轮存活的项目，跨轮不得污染", round)

		assert.Nil(t, entdb.Namespace.Query().Where(namespace.ID(ns.ID)).OnlyX(skip).DeletedAt,
			"第 %d 轮：空间应已恢复", round)

		restored := entdb.Project.Query().Where(project.ID(alive.ID)).OnlyX(skip)
		assert.Nil(t, restored.DeletedAt, "第 %d 轮：当轮存活的项目应被恢复", round)
		assert.False(t, restored.DeletedWithNamespace, "第 %d 轮：恢复后批次标识必须清零", round)

		untouched := entdb.Project.Query().Where(project.ID(solo.ID)).OnlyX(skip)
		assert.NotNil(t, untouched.DeletedAt, "第 %d 轮：单独删除的项目不得被复活", round)
		assert.False(t, untouched.DeletedWithNamespace, "第 %d 轮：单独删除的项目不得被打标记", round)
	}
}

// TestNamespaceRepo_RestoreDeleted_ResetsDeployStatus 恢复时被还原项目的部署状态一律
// 归为"未知"：删除空间时 helm release 已被物理卸载，残留的"已部署"是对不存在资源的
// 错误描述，会让前端显示一个实际没有负载的"已部署"项目。
func TestNamespaceRepo_RestoreDeleted_ResetsDeployStatus(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	ctx := context.TODO()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))

	ns := createNamespace(entdb)
	p := createNamedProject(entdb, ns.ID, "deployed")
	entdb.Project.UpdateOneID(p.ID).SetDeployStatus(types.Deploy_StatusDeployed).SaveX(ctx)

	require.NoError(t, repo.Delete(ctx, ns.ID))
	restored, err := repo.RestoreDeleted(ctx, ns.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"deployed"}, restored)

	got := entdb.Project.Query().Where(project.ID(p.ID)).OnlyX(ctx)
	assert.Equal(t, types.Deploy_StatusUnknown, got.DeployStatus)
}

// TestNamespaceRepo_RestoreDeleted_NotDeleted 未软删的空间调用恢复必须显式报错：
// 静默成功会让调用方把"什么都没做"当成恢复完成。
func TestNamespaceRepo_RestoreDeleted_NotDeleted(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))

	ns := createNamespace(entdb)
	restored, err := repo.RestoreDeleted(context.TODO(), ns.ID)
	require.Error(t, err)
	assert.Nil(t, restored, "报错时不得交出半份恢复名单")
	assert.ErrorContains(t, err, "未被删除")
}

// TestNamespaceRepo_RestoreDeleted_MissingNamespace 恢复一个不存在的 id：加载软删空间即失败，
// 错误必须按 NotFound 表达（而不是落成 500），调用方才能回「空间不存在」而不是「服务异常」。
func TestNamespaceRepo_RestoreDeleted_MissingNamespace(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))

	restored, err := repo.RestoreDeleted(context.TODO(), 999999)
	require.Error(t, err)
	assert.Nil(t, restored, "报错时不得交出半份恢复名单")
	assert.True(t, errs.IsNotFound(err), "不存在的空间应表达为 NotFound，实际: %v", err)
	assert.ErrorContains(t, err, "restore namespace")
}

// TestNamespaceRepo_RestoreDeleted_ProjectUpdateError 事务内还原项目的 UPDATE 失败：
// 错误上抛且整个事务回滚（空间不得停在"已恢复但项目仍是软删"的半成品状态）。
// SQL 序列：Q1 SELECT namespaces(Only) → E1 UPDATE projects（失败注入点）。
func TestNamespaceRepo_RestoreDeleted_ProjectUpdateError(t *testing.T) {
	ctx := context.TODO()
	repo, fd := newNsFault(t, -1, 0)
	ns := repo.data.DB().Namespace.Create().SetName("restore-fail").SetCreatorEmail("o@x").SaveX(ctx)
	createNamedProject(repo.data.DB(), ns.ID, "p1")
	require.NoError(t, repo.Delete(ctx, ns.ID))
	fd.Arm()

	restored, err := repo.RestoreDeleted(ctx, ns.ID)
	require.ErrorContains(t, err, "restore namespace")
	assert.Nil(t, restored, "事务回滚后恢复名单随之作废")

	// 事务回滚：空间仍处于软删态，未留下半恢复状态。
	skip := mixin.SkipSoftDelete(ctx)
	assert.NotNil(t, repo.data.DB().Namespace.Query().Where(namespace.ID(ns.ID)).OnlyX(skip).DeletedAt,
		"项目还原失败时空间必须保持软删（事务回滚）")
}

// TestNamespaceRepo_RestoreDeleted_ProjectNamesQueryError 事务内捞恢复名单的 SELECT 失败：
// 错误上抛且事务回滚。名单是审计日志的口径，查不出来就不该继续恢复——否则会落下一条
// "空间恢复了、但日志说恢复了 0 个项目"的假记录，比直接失败更难排查。
// SQL 序列：Q1 SELECT namespaces(Only) → Q2 SELECT projects(name)（失败注入点）。
func TestNamespaceRepo_RestoreDeleted_ProjectNamesQueryError(t *testing.T) {
	ctx := context.TODO()
	repo, fd := newNsFault(t, 1, -1)
	ns := repo.data.DB().Namespace.Create().SetName("restore-names-fail").SetCreatorEmail("o@x").SaveX(ctx)
	createNamedProject(repo.data.DB(), ns.ID, "p1")
	require.NoError(t, repo.Delete(ctx, ns.ID))
	fd.Arm()

	restored, err := repo.RestoreDeleted(ctx, ns.ID)
	require.Error(t, err)
	assert.Nil(t, restored, "查询失败不得交出恢复名单")

	// 注入是持续的（命中条件后每条同类语句都失败），回读断言前必须解除武装。
	fd.Disarm()

	// 事务回滚：空间与项目都必须保持软删（没有"恢复了一半"的中间态）。
	skip := mixin.SkipSoftDelete(ctx)
	assert.NotNil(t, repo.data.DB().Namespace.Query().Where(namespace.ID(ns.ID)).OnlyX(skip).DeletedAt,
		"名单查询失败时空间必须保持软删（事务回滚）")
	assert.NotNil(t, repo.data.DB().Project.Query().Where(project.NamespaceID(ns.ID)).OnlyX(skip).DeletedAt,
		"名单查询失败时级联项目必须保持软删（事务回滚）")
}

// TestNamespaceRepo_FindDeletedByName 按展示名定位软删空间：前缀幂等（带不带前缀均可命中）、
// 未删除时按 NotFound 表达、同名多行时取最近一次删除的那行。
func TestNamespaceRepo_FindDeletedByName(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	ctx := context.TODO()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{NsPrefix: "abc-"},
		DB:  entdb,
	}))

	ns := entdb.Namespace.Create().SetName("abc-demo").SetCreatorEmail("a@b.c").SaveX(ctx)

	// 未删除：软删查询不应命中（否则会把在册空间当误删空间恢复）。
	_, err := repo.FindDeletedByName(ctx, "demo")
	require.Error(t, err)
	assert.True(t, errs.IsNotFound(err), "在册空间对软删查询应表现为 NotFound")

	require.NoError(t, repo.Delete(ctx, ns.ID))

	for _, name := range []string{"demo", "abc-demo"} {
		got, err := repo.FindDeletedByName(ctx, name)
		require.NoError(t, err, "传 %q 应能命中", name)
		assert.Equal(t, ns.ID, got.ID)
	}

	// 同名空间被"删除→新建→再删除"：应取最近一次删除的那行（id 更大的新空间）。
	newer := entdb.Namespace.Create().SetName("abc-demo").SetCreatorEmail("a@b.c").SaveX(ctx)
	require.NoError(t, repo.Delete(ctx, newer.ID))

	got, err := repo.FindDeletedByName(ctx, "demo")
	require.NoError(t, err)
	assert.Equal(t, newer.ID, got.ID, "同名重复删除时应取最近一次删除的记录")
}

// TestNamespaceRepo_FindDeletedByName_SameSecondTie 同名空间在同一秒内被删两次时，必须取
// id 更大（= 后删）的那条。理由同项目级同名用例：deleted_at 是秒级精度，同秒碰撞下仅按时间戳
// 排序结果不确定，取错会让真正该恢复的那行被"同名在册"检查永久挡死。
func TestNamespaceRepo_FindDeletedByName_SameSecondTie(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	ctx := context.TODO()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{
		Cfg: &config.Config{NsPrefix: "abc-"},
		DB:  entdb,
	}))

	older := entdb.Namespace.Create().SetName("abc-demo").SetCreatorEmail("a@b.c").SaveX(ctx)
	require.NoError(t, repo.Delete(ctx, older.ID))
	newer := entdb.Namespace.Create().SetName("abc-demo").SetCreatorEmail("a@b.c").SaveX(ctx)
	require.NoError(t, repo.Delete(ctx, newer.ID))

	// 复现"同秒"：把后删那行的删除时间强行对齐到与前一行逐字相等。
	skip := mixin.SkipSoftDelete(ctx)
	olderDeletedAt := *entdb.Namespace.Query().Where(namespace.ID(older.ID)).OnlyX(skip).DeletedAt
	entdb.Namespace.UpdateOneID(newer.ID).SetDeletedAt(olderDeletedAt).SaveX(skip)
	require.Equal(t, olderDeletedAt,
		*entdb.Namespace.Query().Where(namespace.ID(newer.ID)).OnlyX(skip).DeletedAt,
		"前置：两行必须处于同一 deleted_at（碰撞已构造）")

	got, err := repo.FindDeletedByName(ctx, "demo")
	require.NoError(t, err)
	assert.Equal(t, newer.ID, got.ID, "deleted_at 相同时必须由主键决胜，取 id 更大的那行")
}

// TestNamespaceRepo_Delete_MissingNamespace 删除不存在的空间必须仍是 NotFound（404）。
// 这是去掉事务外预查询后最主要的回归风险：404 原先来自预查询 First，现在改由目标
// UPDATE 命中 0 行、经 sqlgraph ensureExists 带谓词复查给出。
func TestNamespaceRepo_Delete_MissingNamespace(t *testing.T) {
	repo, _ := newNsRepo(t)

	err := repo.Delete(context.TODO(), 999999)
	require.Error(t, err)
	assert.True(t, errs.IsNotFound(err), "不存在的空间应表达为 NotFound，实际: %v", err)
	assert.ErrorContains(t, err, "delete namespace")
}

// TestNamespaceRepo_Delete_RepeatDelete 重复删除已软删的空间：第二次必须 NotFound，
// 且先执行的那条级联项目 UPDATE 不得留下副作用（事务回滚）。
func TestNamespaceRepo_Delete_RepeatDelete(t *testing.T) {
	ctx := context.TODO()
	repo, entdb := newNsRepo(t)

	ns := createNamespace(entdb)
	p := createNamedProject(entdb, ns.ID, "p")

	require.NoError(t, repo.Delete(ctx, ns.ID))

	skip := mixin.SkipSoftDelete(ctx)
	require.NotNil(t, entdb.Namespace.Query().Where(namespace.ID(ns.ID)).OnlyX(skip).DeletedAt)

	err := repo.Delete(ctx, ns.ID)
	require.Error(t, err)
	assert.True(t, errs.IsNotFound(err), "重复删除应表达为 NotFound，实际: %v", err)

	got := entdb.Project.Query().Where(project.ID(p.ID)).OnlyX(skip)
	require.NotNil(t, got.DeletedAt)
	assert.True(t, got.DeletedWithNamespace, "事务回滚后应保持首次删除写入的批次标识")
}

// TestNamespaceRepo_Delete_NoPreQueryFence 锁定「级联项目 UPDATE 无条件发出」这一事实：
// 级联标记不再受「事务外预查询存活项目数 > 0」的守卫拦截。
//
// 判据用写入次数而非读取次数：ent 的 UpdateOneID(...).Exec() 内部走 Save() → sqlgraph.UpdateNode，
// 为返回更新后的实体会额外发一次读（实测一次 Delete 恒有 qCount=1）。因此「有没有 SELECT」区分不了
// 新旧实现，只有 Exec 数能：**空空间**下新实现仍会发出那条项目 UPDATE（0 行），eCount=2；
// 旧守卫在空空间时直接跳过，eCount=1。守卫一旦被加回来，本用例立刻转红。
func TestNamespaceRepo_Delete_NoPreQueryFence(t *testing.T) {
	ctx := context.TODO()
	repo, fd := newNsFault(t, -1, -1)
	ns := repo.data.DB().Namespace.Create().SetName("fence").SetCreatorEmail("o@x").SaveX(ctx)
	fd.Arm()

	require.NoError(t, repo.Delete(ctx, ns.ID))
	assert.EqualValues(t, 2, fd.eCount.Load(),
		"空空间也必须无条件发出级联项目 UPDATE（第 1 条 Exec），随后才是空间自身 UPDATE（第 2 条）")
}

// TestNamespaceRepo_UpdateImagePullSecrets_SoftDeletedRowForRestore 回归护栏：本方法**必须**
// 能写已软删的 namespace 行，故意不加 namespace.DeletedAtIsNil() 谓词。
//
// 原因：恢复被误删空间的链路（namespaceBiz.Restore）在清 deleted_at 之前，先用本方法回写
// 重建的 docker secret 名单——"先补 DB 骨架再清软删标记"是刻意的顺序（换序会在失败时留下
// 不可恢复的孤儿空间）。一旦给本方法加上软删谓词，此调用命中 0 行报 NotFound，恢复链路直接
// 断裂；本用例就是拦住那次"顺手对齐"的防线。
func TestNamespaceRepo_UpdateImagePullSecrets_SoftDeletedRowForRestore(t *testing.T) {
	ctx := context.TODO()
	repo, entdb := newNsRepo(t)

	ns := createNamespace(entdb)
	require.NoError(t, repo.Delete(ctx, ns.ID))

	skip := mixin.SkipSoftDelete(ctx)
	require.NotNil(t, entdb.Namespace.Query().Where(namespace.ID(ns.ID)).OnlyX(skip).DeletedAt,
		"前置条件：空间已处于软删态")

	require.NoError(t, repo.UpdateImagePullSecrets(ctx, ns.ID, []string{"restored-secret"}),
		"恢复链路依赖向软删空间回写 secret 名单，本方法不得过滤软删行")
	assert.Equal(t, []string{"restored-secret"},
		entdb.Namespace.Query().Where(namespace.ID(ns.ID)).OnlyX(skip).ImagePullSecrets)
}

// TestNamespaceRepo_ListAdminDeletedPage_FiltersAndProjects 列「可恢复空间」的三条核心口径。
// (1) 只软删行可见——在册空间与「删过又已恢复」的空间都不得出现（拦截器被绕过不等于不过滤）。
// (2) 行内 DeletedAt 真的被 SELECT 出来——列裁剪漏 FieldDeletedAt 会静默为 nil，UI 就显示不出
// 删除时间，这条断言专门守它。
// (3) 项目边只含 deleted_with_namespace=true 的级联批，即 RestoreDeleted 的实际恢复集合；
// 用户早先单独删除的项目不出现在计数里，否则前端会承诺恢复一个并不会被恢复的项目。
func TestNamespaceRepo_ListAdminDeletedPage_FiltersAndProjects(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	ctx := context.TODO()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))
	skip := mixin.SkipSoftDelete(ctx)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// 在册空间：从未删除，必须被 deleted_at IS NOT NULL 滤掉。
	live := createNamespace(entdb)
	// 恢复过的空间：deleted_at 已被清空，与在册空间等价，同样不得出现。
	restored := createNamespace(entdb)
	entdb.Namespace.UpdateOneID(restored.ID).SetDeletedAt(base).SaveX(skip)
	entdb.Namespace.UpdateOneID(restored.ID).ClearDeletedAt().SaveX(skip)

	target := createNamespace(entdb)
	cascaded1 := createNamedProject(entdb, target.ID, "cascaded-1")
	cascaded2 := createNamedProject(entdb, target.ID, "cascaded-2")
	solo := createNamedProject(entdb, target.ID, "solo-by-user")

	// 用户先单独删掉 solo（不带批次标识），随后整个空间下线（级联批打标）。
	entdb.Project.UpdateOneID(solo.ID).SetDeletedAt(base.Add(-time.Hour)).SaveX(skip)
	entdb.Project.Update().
		Where(project.NamespaceID(target.ID), project.IDIn(cascaded1.ID, cascaded2.ID)).
		SetDeletedAt(base).SetDeletedWithNamespace(true).ExecX(skip)
	entdb.Namespace.UpdateOneID(target.ID).SetDeletedAt(base).SaveX(skip)

	// 删过但无项目：必须出现（用户要能恢复空空间），且 Projects 为空切片而非 nil 造成前端崩溃。
	empty := createNamespace(entdb)
	entdb.Namespace.UpdateOneID(empty.ID).SetDeletedAt(base.Add(time.Minute)).SaveX(skip)

	page, err := repo.ListAdminDeletedPage(ctx, &biz.NamespaceDeletedListPageQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count, "只应命中 2 个待恢复空间")

	ids := lo.Map(page.Namespaces, func(ns *biz.Namespace, _ int) int { return ns.ID })
	assert.Equal(t, []int{empty.ID, target.ID}, ids, "按删除时间倒序：最近删除的 empty 在前")

	assert.NotContains(t, ids, live.ID, "在册空间不得出现在恢复列表")
	assert.NotContains(t, ids, restored.ID, "已恢复（deleted_at 为空）的空间不得出现在恢复列表")

	byID := lo.KeyBy(page.Namespaces, func(ns *biz.Namespace) int { return ns.ID })
	targetRow := byID[target.ID]
	require.NotNil(t, targetRow.DeletedAt, "DeletedAt 必须被列裁剪 SELECT 出来，否则 UI 无法展示删除时间")
	assert.Equal(t, base, *targetRow.DeletedAt)
	assert.Empty(t, byID[empty.ID].Projects, "无项目的已删除空间 Projects 应为空")

	projectIDs := lo.Map(targetRow.Projects, func(p *biz.Project, _ int) int { return p.ID })
	assert.ElementsMatch(t, []int{cascaded1.ID, cascaded2.ID}, projectIDs,
		"项目边只含随空间级联删除的那批（solo 不在恢复范围内，不得计入）")
	assert.NotContains(t, projectIDs, solo.ID,
		"用户早先单独删除的项目不得出现——它不会被 RestoreDeleted 恢复")
}

// TestNamespaceRepo_ListAdminDeletedPage_OrderAndPagination 顺序与分页：
// 主键 deleted_at 倒序（最近删除优先）——恢复页最想看到的就是刚删错的那个；
// 第二键 id 倒序——deleted_at 是秒级 datetime，同秒删除的多行只按它排序时
// LIMIT/OFFSET 翻页结果不确定（可能同一行出现两次或整行漏掉），必须由 id 兜底稳定序。
// Count 为搜索命中总数、与分页无关（前端分页器依赖它算总页数）。
func TestNamespaceRepo_ListAdminDeletedPage_OrderAndPagination(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	ctx := context.TODO()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))
	skip := mixin.SkipSoftDelete(ctx)
	base := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	// 三个空间同秒删除：顺序完全由 id 倒序决定（id 越大创建越晚）。
	same1 := createNamespace(entdb)
	same2 := createNamespace(entdb)
	same3 := createNamespace(entdb)
	older := createNamespace(entdb)
	for _, id := range []int{same1.ID, same2.ID, same3.ID} {
		entdb.Namespace.UpdateOneID(id).SetDeletedAt(base).SaveX(skip)
	}
	entdb.Namespace.UpdateOneID(older.ID).SetDeletedAt(base.Add(-time.Hour)).SaveX(skip)

	// 全量：同秒三行按 id 倒序，随后是更早删除的 older。
	all, err := repo.ListAdminDeletedPage(ctx, &biz.NamespaceDeletedListPageQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 4, all.Count, "Count 为命中总数，不随分页裁剪")
	assert.Equal(t, []int{same3.ID, same2.ID, same1.ID, older.ID},
		lo.Map(all.Namespaces, func(ns *biz.Namespace, _ int) int { return ns.ID }),
		"同秒按 id 倒序，再接更早删除的行")

	// 分页：page=2 size=3 应只落最后一行（older），且 Count 仍为 4。
	p2, err := repo.ListAdminDeletedPage(ctx, &biz.NamespaceDeletedListPageQuery{Page: 2, PageSize: 3})
	require.NoError(t, err)
	require.Len(t, p2.Namespaces, 1)
	assert.Equal(t, older.ID, p2.Namespaces[0].ID)
	assert.Equal(t, 4, p2.Count, "翻页不得改变 Count")

	// 越界页返回空行集而非报错（前端翻过头不该炸），Count 仍为 4。
	p9, err := repo.ListAdminDeletedPage(ctx, &biz.NamespaceDeletedListPageQuery{Page: 9, PageSize: 3})
	require.NoError(t, err)
	assert.Empty(t, p9.Namespaces)
	assert.Equal(t, 4, p9.Count)
}

// TestNamespaceRepo_ListAdminDeletedPage_Search 搜索与「只列软删」两个条件必须叠加生效：
// 关键词匹配空间名或创建者邮箱（复用 adminNamespaceBaseQuery 的 search 语义，与
// ListAdminPage 保持一致），但命中的在册空间仍不得出现——搜索放宽不了软删过滤。
func TestNamespaceRepo_ListAdminDeletedPage_Search(t *testing.T) {
	entdb, _ := NewSqliteDB()
	defer entdb.Close()
	ctx := context.TODO()
	repo := NewNamespaceRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: entdb}))
	skip := mixin.SkipSoftDelete(ctx)
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	deletedByName := entdb.Namespace.Create().SetName("alpha-team").SetCreatorEmail("zed@mars.com").SaveX(ctx)
	entdb.Namespace.UpdateOneID(deletedByName.ID).SetDeletedAt(base).SaveX(skip)
	deletedByMail := entdb.Namespace.Create().SetName("beta-team").SetCreatorEmail("alpha-owner@mars.com").SaveX(ctx)
	entdb.Namespace.UpdateOneID(deletedByMail.ID).SetDeletedAt(base).SaveX(skip)
	// 名字与邮箱都含 alpha，但在册：搜索命中也必须被软删过滤排除。
	liveAlpha := entdb.Namespace.Create().SetName("alpha-live").SetCreatorEmail("alpha@mars.com").SaveX(ctx)

	page, err := repo.ListAdminDeletedPage(ctx, &biz.NamespaceDeletedListPageQuery{Search: "alpha", Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count, "按名（alpha-team）与按邮箱（alpha-owner@）各命中一条")
	ids := lo.Map(page.Namespaces, func(ns *biz.Namespace, _ int) int { return ns.ID })
	assert.ElementsMatch(t, []int{deletedByName.ID, deletedByMail.ID}, ids)
	assert.NotContains(t, ids, liveAlpha.ID, "在册空间即使被关键词命中也不得出现在恢复列表")

	// 精确到邮箱前缀时只剩一条，证明邮箱侧确实参与了匹配（而非只匹配名字）。
	byMail, err := repo.ListAdminDeletedPage(ctx, &biz.NamespaceDeletedListPageQuery{Search: "alpha-owner@", Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, byMail.Namespaces, 1)
	assert.Equal(t, deletedByMail.ID, byMail.Namespaces[0].ID)

	// 无命中：Count=0 且行集为空（不是错误）。
	none, err := repo.ListAdminDeletedPage(ctx, &biz.NamespaceDeletedListPageQuery{Search: "no-such-space", Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 0, none.Count)
	assert.Empty(t, none.Namespaces)
}

// TestNamespaceRepo_ListAdminDeletedPage_QueryErrors 覆盖两条真实数据库错误分支：计数失败与
// 取页失败都必须带着上下文上抛，不能返回"计数为 0 的空页"——那会把 DB 故障伪装成"没有可恢复空间"，
// 让超管以为误删的空间凭空消失了。SQL 序列：Q1 COUNT(*) → Q2 SELECT namespaces(带项目边)。
//
//	qAfter=0 → 第 1 条查询（COUNT）失败，命中 "count deleted namespaces"；
//	qAfter=1 → COUNT 成功后第 2 条查询（SELECT namespaces）失败，命中 "list deleted namespaces page"。
func TestNamespaceRepo_ListAdminDeletedPage_QueryErrors(t *testing.T) {
	ctx := context.TODO()
	skip := mixin.SkipSoftDelete(ctx)

	// setup 建一个已软删空间：让两条查询都有真实结果集要处理，而非空表短路。
	setup := func(t *testing.T, repo *namespaceRepo) {
		t.Helper()
		ns := repo.data.DB().Namespace.Create().SetName("ldp-err").SetCreatorEmail("o@x").SaveX(ctx)
		repo.data.DB().Namespace.UpdateOneID(ns.ID).SetDeletedAt(time.Now()).SaveX(skip)
	}

	t.Run("Count query error", func(t *testing.T) {
		repo, fd := newNsFault(t, 0, -1)
		setup(t, repo)
		fd.Arm()
		_, err := repo.ListAdminDeletedPage(ctx, &biz.NamespaceDeletedListPageQuery{Page: 1, PageSize: 10})
		assert.ErrorContains(t, err, "count deleted namespaces")
	})

	t.Run("All query error", func(t *testing.T) {
		repo, fd := newNsFault(t, 1, -1)
		setup(t, repo)
		fd.Arm()
		_, err := repo.ListAdminDeletedPage(ctx, &biz.NamespaceDeletedListPageQuery{Page: 1, PageSize: 10})
		assert.ErrorContains(t, err, "list deleted namespaces page")
	})
}
