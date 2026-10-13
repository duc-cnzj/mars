package data

import (
	"context"
	"testing"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/stretchr/testify/assert"
)

// TestHibernateRepo_CRUD 覆盖按项目读写休眠记录：未配置返回 nil、Upsert 幂等、批量查询、
// 按开关/运行态检索与删除。
func TestHibernateRepo_CRUD(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	defer entdb.Close()
	repo := NewHibernateRepo(NewDataImpl(&NewDataParams{DB: entdb}))

	ns := createNamespace(entdb)
	p1 := createProject(entdb, ns.ID)
	p2 := createProject(entdb, ns.ID)

	// 未配置：nil, nil。
	got, err := repo.Get(context.TODO(), p1.ID)
	assert.NoError(t, err)
	assert.Nil(t, got)

	// 空集合批量查询返回空 map。
	all, err := repo.ListByProjectIDs(context.TODO())
	assert.NoError(t, err)
	assert.Empty(t, all)

	// Upsert 写入。
	saved, err := repo.Upsert(context.TODO(), &biz.ProjectHibernate{
		ProjectID: p1.ID,
		Enabled:   true,
		State:     biz.HibernateStateAwake,
		Snapshot:  []*websocket_pb.ExtraValue{{Path: "a", Value: "3"}},
	})
	assert.NoError(t, err)
	assert.Equal(t, p1.ID, saved.ProjectID)
	assert.Equal(t, biz.HibernateStateAwake, saved.State)

	// Upsert 幂等：同 project_id 再写为更新而非新增。
	saved2, err := repo.Upsert(context.TODO(), &biz.ProjectHibernate{
		ProjectID: p1.ID,
		Enabled:   true,
		State:     biz.HibernateStateSleeping,
	})
	assert.NoError(t, err)
	assert.Equal(t, p1.ID, saved2.ProjectID)
	assert.Equal(t, biz.HibernateStateSleeping, saved2.State)
	// 快照被本次 Upsert 覆盖（UpdateNewValues 语义）。
	assert.Empty(t, saved2.Snapshot)

	// 第二条记录，供批量/检索断言。
	_, err = repo.Upsert(context.TODO(), &biz.ProjectHibernate{ProjectID: p2.ID, Enabled: false, State: biz.HibernateStateAwake})
	assert.NoError(t, err)

	byIDs, err := repo.ListByProjectIDs(context.TODO(), p1.ID, p2.ID)
	assert.NoError(t, err)
	assert.Len(t, byIDs, 2)
	assert.Equal(t, biz.HibernateStateSleeping, byIDs[p1.ID].State)

	enabled, err := repo.ListEnabled(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, enabled, 1)
	assert.Equal(t, p1.ID, enabled[0].ProjectID)

	// 删除后查不到。
	assert.NoError(t, repo.Delete(context.TODO(), p1.ID))
	gone, err := repo.Get(context.TODO(), p1.ID)
	assert.NoError(t, err)
	assert.Nil(t, gone)
	assert.NoError(t, repo.Delete(context.TODO(), p1.ID))
}

// TestToProjectHibernate_Nil 覆盖 nil 安全转换。
func TestToProjectHibernate_Nil(t *testing.T) {
	assert.Nil(t, toProjectHibernate(nil))
}

// TestHibernateRepo_Errors 覆盖各方法在 DB 故障时的错误上抛（关闭连接触发）。
func TestHibernateRepo_Errors(t *testing.T) {
	entdb, err := NewSqliteDB()
	assert.NoError(t, err)
	repo := NewHibernateRepo(NewDataImpl(&NewDataParams{DB: entdb}))
	entdb.Close()

	_, err = repo.Get(context.TODO(), 1)
	assert.Error(t, err)
	_, err = repo.ListByProjectIDs(context.TODO(), 1)
	assert.Error(t, err)
	_, err = repo.ListEnabled(context.TODO())
	assert.Error(t, err)
	_, err = repo.Upsert(context.TODO(), &biz.ProjectHibernate{ProjectID: 1})
	assert.Error(t, err)
	assert.Error(t, repo.Delete(context.TODO(), 1))
}
