package repo

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/util/yaml"

	"entgo.io/ent/dialect/sql"
	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	entevent "github.com/duc-cnzj/mars/v4/internal/ent/event"
	"github.com/duc-cnzj/mars/v4/internal/event"
	"github.com/duc-cnzj/mars/v4/internal/filters"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
	"github.com/duc-cnzj/mars/v4/internal/util/pagination"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
	corev1 "k8s.io/api/core/v1"
)

const (
	AuditLogEvent         event.Event = "audit_log"
	EventNamespaceCreated event.Event = "namespace_created"
	EventNamespaceDeleted event.Event = "namespace_deleted"
	EventProjectChanged   event.Event = "project_changed"
	EventProjectDeleted   event.Event = "project_deleted"
)

type Event struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Action    types.EventActionType
	Username  string
	Message   string
	Old       string
	New       string
	Duration  string
	FileID    *int
	HasDiff   bool

	File *File
}

type EventRepo interface {
	List(ctx context.Context, input *ListEventInput) (events []*Event, pag *pagination.Pagination, err error)
	Show(ctx context.Context, id int) (*Event, error)
	Dispatch(created event.Event, createdData any)

	// AuditLogWithChange 记录审计日志
	AuditLogWithChange(action types.EventActionType, username string, msg string, oldS, newS YamlPrettier)

	// AuditLog 记录审计日志
	AuditLog(action types.EventActionType, username string, msg string)

	// AuditLogWithRequest grpc审计日志
	AuditLogWithRequest(action types.EventActionType, username string, msg string, req any)

	// FileAuditLog 记录文件审计日志
	FileAuditLog(action types.EventActionType, username string, msg string, fileId int)

	// FileAuditLogWithDuration 记录文件审计日志
	FileAuditLogWithDuration(action types.EventActionType, username string, msg string, fileId int, duration time.Duration)

	HandleAuditLog(data any, e event.Event) error
}

var _ EventRepo = (*eventRepo)(nil)

type eventRepo struct {
	logger      mlog.Logger
	eventer     event.Dispatcher
	pl          application.PluginManger
	clRepo      ChangelogRepo
	data        data.Data
	k8sRepo     K8sRepo
	projectRepo ProjectRepo
}

func NewEventRepo(projectRepo ProjectRepo, k8sRepo K8sRepo, pl application.PluginManger, clRepo ChangelogRepo, logger mlog.Logger, data data.Data, eventer event.Dispatcher) EventRepo {
	r := &eventRepo{projectRepo: projectRepo, k8sRepo: k8sRepo, clRepo: clRepo, logger: logger.WithModule("repo/event"), eventer: eventer, data: data, pl: pl}
	eventer.Listen(AuditLogEvent, r.HandleAuditLog)
	eventer.Listen(EventNamespaceCreated, r.HandleInjectTlsSecret)
	eventer.Listen(EventNamespaceDeleted, r.HandleNamespaceDeleted)
	eventer.Listen(EventProjectChanged, r.HandleProjectChanged)
	eventer.Listen(EventProjectDeleted, r.HandleProjectDeleted)
	return r
}

func (repo *eventRepo) HandleAuditLog(data any, e event.Event) error {
	logData := data.(AuditLog)
	var fid *int
	if logData.GetFileID() != 0 {
		ffid := logData.GetFileID()
		fid = &ffid
	}
	var db = repo.data.DB()
	db.Event.Create().SetAction(logData.GetAction()).
		SetUsername(logData.GetUsername()).
		SetMessage(logData.GetMsg()).
		SetOld(logData.GetOldStr()).
		SetNew(logData.GetNewStr()).
		SetDuration(logData.GetDuration()).
		SetHasDiff(logData.GetOldStr() != logData.GetNewStr()).
		SetNillableFileID(fid).
		Save(context.TODO())

	return nil
}

func (repo *eventRepo) Dispatch(created event.Event, createdData any) {
	repo.eventer.Dispatch(created, createdData)
}

type ListEventInput struct {
	Page, PageSize int32
	ActionType     types.EventActionType
	Search         string
	OrderIDDesc    *bool
}

func (repo *eventRepo) List(ctx context.Context, input *ListEventInput) (events []*Event, pag *pagination.Pagination, err error) {
	var db = repo.data.DB()
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

	return serialize.Serialize(items, ToEvent), pagination.NewPagination(input.Page, input.PageSize, 0), nil
}

func (repo *eventRepo) Show(ctx context.Context, id int) (*Event, error) {
	var db = repo.data.DB()
	first, err := db.Event.Query().WithFile().Where(entevent.ID(id)).First(ctx)
	return ToEvent(first), err
}

func (repo *eventRepo) AuditLog(action types.EventActionType, username string, msg string) {
	repo.AuditLogWithChange(action, username, msg, nil, nil)
}

func (repo *eventRepo) FileAuditLog(action types.EventActionType, username string, msg string, fileId int) {
	repo.eventer.Dispatch(AuditLogEvent, NewEventAuditLog(username, action, msg, AuditWithFileID(fileId)))
}

func (repo *eventRepo) FileAuditLogWithDuration(action types.EventActionType, username string, msg string, fileId int, d time.Duration) {
	repo.eventer.Dispatch(AuditLogEvent, NewEventAuditLog(username, action, msg, AuditWithFileID(fileId), AuditWithDuration(date.HumanDuration(d))))
}

func (repo *eventRepo) AuditLogWithRequest(action types.EventActionType, username string, msg string, req any) {
	marshal, _ := yaml.PrettyMarshal(req)
	repo.eventer.Dispatch(AuditLogEvent, NewEventAuditLog(username, action, msg, AuditWithOldNewStr("", string(marshal))))
}

func (repo *eventRepo) AuditLogWithChange(action types.EventActionType, username string, msg string, oldS, newS YamlPrettier) {
	if oldS == nil {
		oldS = &emptyYamlPrettier{}
	}
	if newS == nil {
		newS = &emptyYamlPrettier{}
	}
	repo.eventer.Dispatch(AuditLogEvent, NewEventAuditLog(username, action, msg, AuditWithOldNew(oldS, newS)))
}

type NamespaceCreatedData struct {
	NsModel  *Namespace
	NsK8sObj *corev1.Namespace
}

func (repo *eventRepo) HandleInjectTlsSecret(data any, e event.Event) error {
	if createdData, ok := data.(NamespaceCreatedData); ok {
		name, key, crt := repo.pl.Domain().GetCerts()
		if name != "" && key != "" && crt != "" {
			ns := createdData.NsK8sObj.Name
			_, err := repo.k8sRepo.AddTlsSecret(ns, name, key, crt)
			if err != nil {
				repo.logger.Error(err)
			}
		}
	}
	return nil
}

type NamespaceDeletedData struct {
	ID int
}

func (repo *eventRepo) HandleNamespaceDeleted(data any, e event.Event) error {
	var (
		ws     = repo.pl.Ws()
		logger = repo.logger
	)
	sub := ws.New("", "")
	defer sub.Close()
	sub.ToAll(&websocket_pb.WsReloadProjectsResponse{
		Metadata:    &websocket_pb.Metadata{Type: websocket_pb.Type_ReloadProjects},
		NamespaceId: int32(data.(NamespaceDeletedData).ID),
	})
	logger.Debug("event handled: ", e.String())
	return nil
}

type ProjectChangedData struct {
	ID int

	Username string
}

func (repo *eventRepo) HandleProjectChanged(data any, e event.Event) error {
	if changedData, ok := data.(*ProjectChangedData); ok {
		proj, err := repo.projectRepo.Show(context.TODO(), changedData.ID)
		if err != nil {
			repo.logger.Error("[HandleProjectChanged]: ", err)
			return err
		}
		var configChanged bool
		if lastChange, err := repo.clRepo.FindLastChangeByProjectID(context.TODO(), changedData.ID); err == nil {
			repo.logger.Debug(lastChange, "lastChange")
			configChanged = lastChange.Config != proj.Config || lastChange.GitCommit != proj.GitCommit
		}
		repo.clRepo.Create(context.TODO(), &CreateChangeLogInput{
			Version:          proj.Version,
			Username:         changedData.Username,
			Config:           proj.Config,
			GitBranch:        proj.GitBranch,
			GitCommit:        proj.GitCommit,
			DockerImage:      proj.DockerImage,
			EnvValues:        proj.EnvValues,
			ExtraValues:      proj.ExtraValues,
			FinalExtraValues: proj.FinalExtraValues,
			GitCommitWebURL:  proj.GitCommitWebURL,
			GitCommitTitle:   proj.GitCommitTitle,
			GitCommitAuthor:  proj.GitCommitAuthor,
			GitCommitDate:    proj.GitCommitDate,
			ConfigChanged:    configChanged,
			ProjectID:        changedData.ID,
		})
	}
	return nil
}

type ProjectDeletedPayload struct {
	NamespaceID int
	ProjectID   int
}

func (repo *eventRepo) HandleProjectDeleted(data any, e event.Event) error {
	var (
		ws     = repo.pl.Ws()
		logger = repo.logger
	)
	input := data.(*ProjectDeletedPayload)
	sub := ws.New("", "")
	defer sub.Close()
	sub.ToAll(&websocket_pb.WsReloadProjectsResponse{
		Metadata:    &websocket_pb.Metadata{Type: websocket_pb.Type_ReloadProjects},
		NamespaceId: int32(input.NamespaceID),
	})
	logger.Debug("event handled: ", e.String(), data)
	return nil
}

type AuditLog interface {
	// GetUsername 获取用户
	GetUsername() string
	// GetAction 行为
	GetAction() types.EventActionType
	// GetMsg desc
	GetMsg() string
	// GetOldStr old config str
	GetOldStr() string
	// GetNewStr new config str
	GetNewStr() string
	// GetFileID file id
	GetFileID() int
	// GetDuration duration
	GetDuration() string
}

var _ AuditLog = (*auditLogImpl)(nil)

type auditLogImpl struct {
	Username        string
	Action          types.EventActionType
	Msg, OldS, NewS string
	FileId          int
	Duration        string
}

func (e *auditLogImpl) GetDuration() string {
	return e.Duration
}

type AuditOption func(*auditLogImpl)

func AuditWithOldNewStr(o, n string) AuditOption {
	return func(e *auditLogImpl) {
		e.OldS = o
		e.NewS = n
	}
}
func AuditWithOldNew(o, n YamlPrettier) AuditOption {
	return func(e *auditLogImpl) {
		if o != nil {
			e.OldS = o.PrettyYaml()
		}
		if n != nil {
			e.NewS = n.PrettyYaml()
		}
	}
}

func AuditWithFileID(id int) AuditOption {
	return func(e *auditLogImpl) {
		e.FileId = id
	}
}
func AuditWithDuration(d string) AuditOption {
	return func(e *auditLogImpl) {
		e.Duration = d
	}
}

func NewEventAuditLog(username string, action types.EventActionType, msg string, opts ...AuditOption) AuditLog {
	e := &auditLogImpl{Username: username, Action: action, Msg: msg}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *auditLogImpl) GetUsername() string {
	return e.Username
}

func (e *auditLogImpl) GetAction() types.EventActionType {
	return e.Action
}

func (e *auditLogImpl) GetMsg() string {
	return e.Msg
}

func (e *auditLogImpl) GetOldStr() string {
	return e.OldS
}

func (e *auditLogImpl) GetNewStr() string {
	return e.NewS
}

func (e *auditLogImpl) GetFileID() int {
	return e.FileId
}

type YamlPrettier interface {
	PrettyYaml() string
}

type emptyYamlPrettier struct{}

func (e *emptyYamlPrettier) PrettyYaml() string {
	return ""
}

type StringYamlPrettier struct {
	Str string
}

func (s *StringYamlPrettier) PrettyYaml() string {
	return s.Str
}

type AnyYamlPrettier map[string]any

func (s *AnyYamlPrettier) PrettyYaml() string {
	marshal, _ := yaml.PrettyMarshal(s)
	return string(marshal)
}

func ToEvent(data *ent.Event) *Event {
	if data == nil {
		return nil
	}
	return &Event{
		ID:        data.ID,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
		DeletedAt: data.DeletedAt,
		Action:    data.Action,
		Username:  data.Username,
		Message:   data.Message,
		Old:       data.Old,
		New:       data.New,
		Duration:  data.Duration,
		FileID:    data.FileID,
		HasDiff:   data.HasDiff,
		File:      ToFile(data.Edges.File),
	}
}
