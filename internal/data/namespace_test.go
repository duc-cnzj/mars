package data

import (
	"context"
	"errors"
	"testing"

	entgo "entgo.io/ent"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/favorite"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/namespace"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/project"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
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
