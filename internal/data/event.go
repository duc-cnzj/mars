package data

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	entevent "github.com/duc-cnzj/mars/v6/internal/data/ent/event"
	"github.com/duc-cnzj/mars/v6/internal/data/filters"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/event"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/duc-cnzj/mars/v6/internal/util/yaml"
)

// AuditLog 是审计日志事件负载的抽象，由 event.Dispatcher 回调消费方（HandleAuditLog）读取。
type AuditLog interface {
	// GetUsername 返回操作人。
	GetUsername() string
	// GetAction 返回操作动作类型。
	GetAction() types.EventActionType
	// GetMsg 返回操作描述消息。
	GetMsg() string
	// GetOldStr 返回变更前的原始值。
	GetOldStr() string
	// GetNewStr 返回变更后的新值。
	GetNewStr() string
	// GetFileID 返回关联的文件 ID（0 表示无）。
	GetFileID() int
	// GetDuration 返回操作耗时描述。
	GetDuration() string
}

var _ biz.EventRepo = (*eventRepo)(nil)

// eventRepo 是 biz.EventRepo 的实现：把审计日志写入 DB，并透传事件分发。
type eventRepo struct {
	logger  mlog.Logger
	eventer event.Dispatcher
	d       dataStore
}

// NewEventRepo 构造事件 repo 实现，注入事件分发器与 dataStore。
func NewEventRepo(logger mlog.Logger, d dataStore, eventer event.Dispatcher) biz.EventRepo {
	return &eventRepo{logger: logger.WithModule("repo/event"), eventer: eventer, d: d}
}

// HandleAuditLog 消费事件负载：把 AuditLog 落库为一条事件记录（含变更对比与可选文件 ID）。
func (repo *eventRepo) HandleAuditLog(d any, e biz.EventKey) error {
	logData := d.(AuditLog)
	var fid *int
	if logData.GetFileID() != 0 {
		ffid := logData.GetFileID()
		fid = &ffid
	}
	var db = repo.d.DB()
	if _, err := db.Event.Create().SetAction(logData.GetAction()).
		SetUsername(logData.GetUsername()).
		SetMessage(logData.GetMsg()).
		SetOld(logData.GetOldStr()).
		SetNew(logData.GetNewStr()).
		SetDuration(logData.GetDuration()).
		SetHasDiff(logData.GetOldStr() != logData.GetNewStr()).
		SetNillableFileID(fid).
		Save(context.TODO()); err != nil {
		return errs.Wrap(err, "handle audit log")
	}

	return nil
}

// Dispatch 把业务事件与负载透传给内部事件分发器。
func (repo *eventRepo) Dispatch(created biz.EventKey, createdData any) {
	repo.eventer.Dispatch(event.Event(created), createdData)
}

// List 分页查询事件列表，支持按动作类型、搜索关键词与倒序过滤。
func (repo *eventRepo) List(ctx context.Context, input *biz.ListEventInput) (events []*biz.Event, pag *pagination.Pagination, err error) {
	var db = repo.d.DB()
	items := db.Event.Query().Where(
		filters.IfIntEQ[types.EventActionType](entevent.FieldAction)(input.ActionType),
		filters.IfOrderByDesc("id")(input.OrderIDDesc),
		filters.If(func(t string) bool {
			return t != ""
		}, func(t string) func(*sql.Selector) {
			return entevent.Or(
				entevent.MessageContains(t),
				entevent.UsernameContains(t),
			)
		})(input.Search),
	).
		Select(
			entevent.FieldID,
			entevent.FieldAction,
			entevent.FieldUsername,
			entevent.FieldMessage,
			entevent.FieldDuration,
			entevent.FieldFileID,
			entevent.FieldHasDiff,
			entevent.FieldCreatedAt,
			entevent.FieldUpdatedAt,
		).
		WithFile().
		Offset(pagination.GetPageOffset(input.Page, input.PageSize)).
		Limit(int(input.PageSize)).
		AllX(ctx)

	return slice.Map(items, toEvent), pagination.NewPagination(input.Page, input.PageSize, 0), nil
}

// toEvent 把 ent.Event 转换为 biz.Event（nil 安全），并顺带转换关联的 File。
func toEvent(e *ent.Event) *biz.Event {
	if e == nil {
		return nil
	}
	return &biz.Event{
		ID:        e.ID,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		DeletedAt: e.DeletedAt,
		Action:    e.Action,
		Username:  e.Username,
		Message:   e.Message,
		Old:       e.Old,
		New:       e.New,
		Duration:  e.Duration,
		FileID:    e.FileID,
		HasDiff:   e.HasDiff,
		File:      toFile(e.Edges.File),
	}
}

// Show 查询单条事件记录并预加载关联文件。
func (repo *eventRepo) Show(ctx context.Context, id int) (*biz.Event, error) {
	var db = repo.d.DB()
	first, err := db.Event.Query().WithFile().Where(entevent.ID(id)).First(ctx)
	return toEvent(first), errs.Wrap(err, "show event")
}

// AuditLog 记录一条无变更对比的审计日志（空旧/新值）。
func (repo *eventRepo) AuditLog(action types.EventActionType, username string, msg string) {
	repo.AuditLogWithChange(action, username, msg, nil, nil)
}

// FileAuditLog 记录一条关联文件 ID 的审计日志。
func (repo *eventRepo) FileAuditLog(action types.EventActionType, username string, msg string, fileId int) {
	repo.eventer.Dispatch(event.Event(biz.AuditLogEvent), NewEventAuditLog(username, action, msg, AuditWithFileID(fileId)))
}

// FileAuditLogWithDuration 记录关联文件 ID 并带耗时的审计日志。
func (repo *eventRepo) FileAuditLogWithDuration(action types.EventActionType, username string, msg string, fileId int, d time.Duration) {
	repo.eventer.Dispatch(event.Event(biz.AuditLogEvent), NewEventAuditLog(username, action, msg, AuditWithFileID(fileId), AuditWithDuration(date.HumanDuration(d))))
}

// AuditLogWithRequest 把请求对象 YAML 序列化后作为"变更后"值记录审计日志。
// 本方法无 error 返回值（审计失败不阻断主流程），属"无法向调用方返回错误"的边缘
// 路径：序列化失败时降级为 %+v，避免审计记录静默为空；错误不在此处打印日志
// （数据层不打印错误日志），留待链路最上层消费。
func (repo *eventRepo) AuditLogWithRequest(action types.EventActionType, username string, msg string, req any) {
	marshal, err := yaml.PrettyMarshal(req)
	after := string(marshal)
	if err != nil {
		after = fmt.Sprintf("%+v", req)
	}
	repo.eventer.Dispatch(event.Event(biz.AuditLogEvent), NewEventAuditLog(username, action, msg, AuditWithOldNewStr("", after)))
}

// AuditLogWithChange 记录带旧/新 YAML 变更对比的审计日志（nil 用空值兜底）。
func (repo *eventRepo) AuditLogWithChange(action types.EventActionType, username string, msg string, oldS, newS biz.YamlPrettier) {
	if oldS == nil {
		oldS = &emptyYamlPrettier{}
	}
	if newS == nil {
		newS = &emptyYamlPrettier{}
	}
	repo.eventer.Dispatch(event.Event(biz.AuditLogEvent), NewEventAuditLog(username, action, msg, AuditWithOldNew(oldS, newS)))
}

// auditLogImpl 是 AuditLog 接口的默认实现：保存审计字段并序列化变更对比。
type auditLogImpl struct {
	Username        string
	Action          types.EventActionType
	Msg, OldS, NewS string
	FileId          int
	Duration        string
}

// GetDuration 返回操作耗时描述。
func (e *auditLogImpl) GetDuration() string {
	return e.Duration
}

// AuditOption 是 NewEventAuditLog 的可选参数：以函数式设置 auditLogImpl 的字段。
type AuditOption func(*auditLogImpl)

// AuditWithOldNewStr 直接设置旧/新值的字符串。
func AuditWithOldNewStr(o, n string) AuditOption {
	return func(e *auditLogImpl) {
		e.OldS = o
		e.NewS = n
	}
}

// AuditWithOldNew 把旧/新 YamlPrettier 序列化后设置到负载（nil 跳过）。
func AuditWithOldNew(o, n biz.YamlPrettier) AuditOption {
	return func(e *auditLogImpl) {
		if o != nil {
			e.OldS = o.PrettyYaml()
		}
		if n != nil {
			e.NewS = n.PrettyYaml()
		}
	}
}

// AuditWithFileID 设置关联的文件 ID。
func AuditWithFileID(id int) AuditOption {
	return func(e *auditLogImpl) {
		e.FileId = id
	}
}

// AuditWithDuration 设置操作耗时描述。
func AuditWithDuration(d string) AuditOption {
	return func(e *auditLogImpl) {
		e.Duration = d
	}
}

// NewEventAuditLog 构造审计日志负载，依次应用可选参数。
func NewEventAuditLog(username string, action types.EventActionType, msg string, opts ...AuditOption) AuditLog {
	e := &auditLogImpl{Username: username, Action: action, Msg: msg}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// GetUsername 返回操作人。
func (e *auditLogImpl) GetUsername() string {
	return e.Username
}

// GetAction 返回操作动作类型。
func (e *auditLogImpl) GetAction() types.EventActionType {
	return e.Action
}

// GetMsg 返回操作描述消息。
func (e *auditLogImpl) GetMsg() string {
	return e.Msg
}

// GetOldStr 返回变更前的原始值。
func (e *auditLogImpl) GetOldStr() string {
	return e.OldS
}

// GetNewStr 返回变更后的新值。
func (e *auditLogImpl) GetNewStr() string {
	return e.NewS
}

// GetFileID 返回关联的文件 ID（0 表示无）。
func (e *auditLogImpl) GetFileID() int {
	return e.FileId
}

// emptyYamlPrettier 是 YamlPrettier 的空实现，序列化为空字符串，用于审计变更兜底。
type emptyYamlPrettier struct{}

// PrettyYaml 返回空字符串。
func (e *emptyYamlPrettier) PrettyYaml() string { return "" }
