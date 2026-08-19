package data

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	entevent "github.com/duc-cnzj/mars/v6/internal/data/ent/event"
	"github.com/duc-cnzj/mars/v6/internal/event"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// failingYAMLMarshaler 实现 goccy/go-yaml 的 Marshaler 接口但恒返回错误，
// 用于触发 AuditLogWithRequest 的序列化失败降级分支。
type failingYAMLMarshaler struct {
	Foo string
}

// MarshalYAML 恒返回错误，模拟不可序列化的请求对象。
func (failingYAMLMarshaler) MarshalYAML() (any, error) {
	return nil, errors.New("boom")
}

// newEventRepo 构造 eventRepo：entdb 落库 + MockDispatcher 断言事件分发。
// 审计监听注册已上收 eventhandler，此处不再断言 Listen。
func newEventRepo(t *testing.T) (*eventRepo, *ent.Client, *event.MockDispatcher) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	entdb, err := NewSqliteDB()
	require.NoError(t, err)
	t.Cleanup(func() { entdb.Close() })
	eventer := event.NewMockDispatcher(m)
	repo := NewEventRepo(mlog.NewForConfig(nil), NewDataImpl(&NewDataParams{DB: entdb, Cfg: &config.Config{}}), eventer).(*eventRepo)
	return repo, entdb, eventer
}

// TestEventRepo_HandleAuditLog 覆盖审计日志落库：有/无 fileID 两种分支 + DB 错误透传。
func TestEventRepo_HandleAuditLog(t *testing.T) {
	repo, entdb, _ := newEventRepo(t)
	ctx := context.TODO()

	t.Run("without fileID", func(t *testing.T) {
		err := repo.HandleAuditLog(NewEventAuditLog("u1", types.EventActionType_Create, "create ns"), biz.AuditLogEvent)
		assert.NoError(t, err)
		got := entdb.Event.Query().Where(entevent.Message("create ns")).OnlyX(ctx)
		assert.Equal(t, "u1", got.Username)
		assert.Zero(t, got.FileID)
		assert.False(t, got.HasDiff)
	})

	t.Run("with fileID and old/new diff", func(t *testing.T) {
		f := entdb.File.Create().SetPath("a.txt").SetSize(1).SaveX(ctx)
		payload := NewEventAuditLog("u2", types.EventActionType_Delete, "delete file",
			AuditWithOldNewStr("old", "new"), AuditWithFileID(f.ID))
		assert.NoError(t, repo.HandleAuditLog(payload, biz.AuditLogEvent))
		got := entdb.Event.Query().Where(entevent.Message("delete file")).OnlyX(ctx)
		assert.Equal(t, f.ID, *got.FileID)
		assert.True(t, got.HasDiff)
	})

	t.Run("DB error propagated", func(t *testing.T) {
		closed := NewDataImpl(&NewDataParams{DB: mustClosedDB(t), Cfg: &config.Config{}})
		bad := &eventRepo{logger: mlog.NewForConfig(nil), d: closed}
		err := bad.HandleAuditLog(NewEventAuditLog("x", types.EventActionType_Create, "boom"), biz.AuditLogEvent)
		assert.Error(t, err)
	})
}

// TestEventRepo_Dispatch 覆盖事件原样转发到 dispatcher。
func TestEventRepo_Dispatch(t *testing.T) {
	repo, _, eventer := newEventRepo(t)
	eventer.EXPECT().Dispatch(event.Event("my-event"), "payload").Times(1)
	repo.Dispatch("my-event", "payload")
}

// TestEventRepo_List 覆盖 action 过滤/search 过滤/排序三种路径。
func TestEventRepo_List(t *testing.T) {
	repo, entdb, _ := newEventRepo(t)
	ctx := context.TODO()

	entdb.Event.Create().SetAction(types.EventActionType_Create).SetUsername("alice").SetMessage("create a").SetDuration("1s").SetHasDiff(false).SaveX(ctx)
	entdb.Event.Create().SetAction(types.EventActionType_Delete).SetUsername("bob").SetMessage("delete b").SetDuration("2s").SetHasDiff(true).SaveX(ctx)

	t.Run("no filters returns all", func(t *testing.T) {
		items, pag, err := repo.List(ctx, &biz.ListEventInput{Page: 1, PageSize: 10})
		assert.NoError(t, err)
		assert.Len(t, items, 2)
		assert.Equal(t, int32(1), pag.Page)
		assert.Equal(t, int32(10), pag.PageSize)
	})

	t.Run("action filter single", func(t *testing.T) {
		items, _, _ := repo.List(ctx, &biz.ListEventInput{Page: 1, PageSize: 10, ActionTypes: []types.EventActionType{types.EventActionType_Delete}})
		require.Len(t, items, 1)
		assert.Equal(t, "delete b", items[0].Message)
	})

	t.Run("action filter multi (IN)", func(t *testing.T) {
		items, _, _ := repo.List(ctx, &biz.ListEventInput{Page: 1, PageSize: 10, ActionTypes: []types.EventActionType{types.EventActionType_Create, types.EventActionType_Delete}})
		require.Len(t, items, 2)
		assert.ElementsMatch(t, []string{"create a", "delete b"}, []string{items[0].Message, items[1].Message})
	})

	t.Run("action filter no match", func(t *testing.T) {
		items, _, _ := repo.List(ctx, &biz.ListEventInput{Page: 1, PageSize: 10, ActionTypes: []types.EventActionType{types.EventActionType_Shell}})
		require.Len(t, items, 0)
	})

	t.Run("search by message", func(t *testing.T) {
		items, _, _ := repo.List(ctx, &biz.ListEventInput{Page: 1, PageSize: 10, Search: "delete"})
		require.Len(t, items, 1)
		assert.Equal(t, "bob", items[0].Username)
	})

	t.Run("search by username", func(t *testing.T) {
		items, _, _ := repo.List(ctx, &biz.ListEventInput{Page: 1, PageSize: 10, Search: "alice"})
		require.Len(t, items, 1)
		assert.Equal(t, "create a", items[0].Message)
	})

	t.Run("id desc ordering", func(t *testing.T) {
		items, _, _ := repo.List(ctx, &biz.ListEventInput{Page: 1, PageSize: 10, OrderIDDesc: lo.ToPtr(true)})
		require.Len(t, items, 2)
		assert.Greater(t, items[0].ID, items[1].ID)
	})
}

// TestEventRepo_Show 覆盖存在与不存在两种结果。
func TestEventRepo_Show(t *testing.T) {
	repo, entdb, _ := newEventRepo(t)
	ctx := context.TODO()

	row := entdb.Event.Create().SetAction(types.EventActionType_Create).SetUsername("u").SetMessage("m").SaveX(ctx)
	got, err := repo.Show(ctx, row.ID)
	assert.NoError(t, err)
	assert.Equal(t, "m", got.Message)

	_, err = repo.Show(ctx, 999999)
	assert.Error(t, err)
}

// TestEventRepo_AuditMethods 覆盖全部 Dispatch 型审计方法的载荷构造。
func TestEventRepo_AuditMethods(t *testing.T) {
	repo, _, eventer := newEventRepo(t)
	want := event.Event(biz.AuditLogEvent)

	t.Run("AuditLog no options", func(t *testing.T) {
		eventer.EXPECT().Dispatch(want, gomock.Any()).Do(func(_ event.Event, p any) {
			al := p.(AuditLog)
			assert.Equal(t, "u", al.GetUsername())
			assert.Equal(t, types.EventActionType_Create, al.GetAction())
			assert.Equal(t, "msg", al.GetMsg())
			assert.Equal(t, "", al.GetOldStr())
			assert.Equal(t, "", al.GetNewStr())
		})
		repo.AuditLog(types.EventActionType_Create, "u", "msg")
	})

	t.Run("FileAuditLog", func(t *testing.T) {
		eventer.EXPECT().Dispatch(want, gomock.Any()).Do(func(_ event.Event, p any) {
			assert.Equal(t, 7, p.(AuditLog).GetFileID())
		})
		repo.FileAuditLog(types.EventActionType_Delete, "u", "m", 7)
	})

	t.Run("FileAuditLogWithDuration", func(t *testing.T) {
		eventer.EXPECT().Dispatch(want, gomock.Any()).Do(func(_ event.Event, p any) {
			al := p.(AuditLog)
			assert.Equal(t, 7, al.GetFileID())
			assert.NotEmpty(t, al.GetDuration())
		})
		repo.FileAuditLogWithDuration(types.EventActionType_Delete, "u", "m", 7, time.Second)
	})

	t.Run("AuditLogWithRequest", func(t *testing.T) {
		eventer.EXPECT().Dispatch(want, gomock.Any()).Do(func(_ event.Event, p any) {
			al := p.(AuditLog)
			assert.Contains(t, al.GetNewStr(), "name: x")
		})
		repo.AuditLogWithRequest(types.EventActionType_Update, "u", "m", struct {
			Name string `yaml:"name"`
		}{Name: "x"})
	})

	t.Run("AuditLogWithRequest marshal error fallback", func(t *testing.T) {
		// PrettyMarshal 失败时降级为 %+v，审计记录不为空
		eventer.EXPECT().Dispatch(want, gomock.Any()).Do(func(_ event.Event, p any) {
			al := p.(AuditLog)
			assert.Equal(t, "{Foo:x}", al.GetNewStr())
		})
		repo.AuditLogWithRequest(types.EventActionType_Update, "u", "m", failingYAMLMarshaler{Foo: "x"})
	})

	t.Run("AuditLogWithChange", func(t *testing.T) {
		eventer.EXPECT().Dispatch(want, gomock.Any()).Do(func(_ event.Event, p any) {
			al := p.(AuditLog)
			assert.Equal(t, "old-yaml", al.GetOldStr())
			assert.Equal(t, "new-yaml", al.GetNewStr())
		})
		repo.AuditLogWithChange(types.EventActionType_Update, "u", "m",
			yamlPrettierFn(func() string { return "old-yaml" }),
			yamlPrettierFn(func() string { return "new-yaml" }),
		)
	})

	t.Run("AuditLogWithChange nil options", func(t *testing.T) {
		eventer.EXPECT().Dispatch(want, gomock.Any()).Do(func(_ event.Event, p any) {
			al := p.(AuditLog)
			assert.Equal(t, "", al.GetOldStr())
			assert.Equal(t, "", al.GetNewStr())
		})
		repo.AuditLogWithChange(types.EventActionType_Update, "u", "m", nil, nil)
	})
}

// TestToEvent 覆盖 nil 与实体两种转换。
func TestToEvent(t *testing.T) {
	assert.Nil(t, toEvent(nil))
	got := toEvent(&ent.Event{ID: 1, Message: "m", HasDiff: true})
	assert.Equal(t, "m", got.Message)
	assert.True(t, got.HasDiff)
}

// TestAuditOptions 覆盖各 AuditOption 构造器与审计日志载荷 getter。
func TestAuditOptions(t *testing.T) {
	e := &auditLogImpl{}

	AuditWithOldNewStr("o", "n")(e)
	assert.Equal(t, "o", e.OldS)
	assert.Equal(t, "n", e.NewS)

	AuditWithOldNew(
		yamlPrettierFn(func() string { return "A" }),
		yamlPrettierFn(func() string { return "B" }),
	)(e)
	assert.Equal(t, "A", e.OldS)
	assert.Equal(t, "B", e.NewS)

	// nil 选项不覆盖已设置字段（if o != nil 守卫）
	AuditWithOldNew(nil, nil)(e)
	assert.Equal(t, "A", e.OldS)

	AuditWithFileID(9)(e)
	assert.Equal(t, 9, e.FileId)

	AuditWithDuration("3s")(e)
	assert.Equal(t, "3s", e.Duration)

	al := NewEventAuditLog("u", types.EventActionType_Create, "m", AuditWithFileID(5), AuditWithDuration("1s"))
	assert.Equal(t, "u", al.GetUsername())
	assert.Equal(t, types.EventActionType_Create, al.GetAction())
	assert.Equal(t, "m", al.GetMsg())
	assert.Equal(t, 5, al.GetFileID())
	assert.Equal(t, "1s", al.GetDuration())
	assert.Equal(t, "", al.GetOldStr())
	assert.Equal(t, "", al.GetNewStr())
	assert.Equal(t, "", (&emptyYamlPrettier{}).PrettyYaml())
}

// yamlPrettierFn 适配 biz.YamlPrettier 接口的测试替身。
type yamlPrettierFn func() string

func (f yamlPrettierFn) PrettyYaml() string { return f() }

// mustClosedDB 返回已关闭的 ent client 作为 dataStore 的 DB，用于触发落库错误。
func mustClosedDB(t *testing.T) *ent.Client {
	entdb, err := NewSqliteDB()
	require.NoError(t, err)
	entdb.Close()
	return entdb
}
