package repo

import (
	"context"

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
	"github.com/duc-cnzj/mars/v4/internal/utils/pagination"
	"github.com/duc-cnzj/mars/v4/plugins/domainmanager"
	corev1 "k8s.io/api/core/v1"
)

const (
	AuditLogEvent         event.Event = "audit_log"
	EventNamespaceCreated event.Event = "namespace_created"
	EventNamespaceDeleted event.Event = "namespace_deleted"
	EventProjectChanged   event.Event = "project_changed"
	EventProjectDeleted   event.Event = "project_deleted"
)

type EventRepo interface {
	List(ctx context.Context, input *ListEventInput) (events []*ent.Event, pag *pagination.Pagination, err error)
	Show(ctx context.Context, id int) (*ent.Event, error)
	Dispatch(created event.Event, createdData any)

	// AuditLogWithChange 记录审计日志
	AuditLogWithChange(action types.EventActionType, username string, msg string, oldS, newS YamlPrettier)

	// AuditLog 记录审计日志
	AuditLog(action types.EventActionType, username string, msg string)

	// FileAuditLog 记录文件审计日志
	FileAuditLog(action types.EventActionType, username string, msg string, fileId int)

	HandleAuditLog(data any, e event.Event) error
}

var _ EventRepo = (*eventRepo)(nil)

type eventRepo struct {
	logger         mlog.Logger
	db             *ent.Client
	eventer        event.Dispatcher
	pl             application.PluginManger
	k8sCli         *data.K8sClient
	clRepo         ChangelogRepo
	gitprojectRepo GitProjectRepo
}

func NewEventRepo(gitprojectRepo GitProjectRepo, clRepo ChangelogRepo, logger mlog.Logger, data *data.Data, eventer event.Dispatcher) EventRepo {
	repo := &eventRepo{gitprojectRepo: gitprojectRepo, clRepo: clRepo, logger: logger, db: data.DB, eventer: eventer, k8sCli: data.K8sClient}
	eventer.Listen(AuditLogEvent, repo.HandleAuditLog)
	eventer.Listen(EventNamespaceCreated, repo.HandleInjectTlsSecret)
	eventer.Listen(EventNamespaceDeleted, repo.HandleNamespaceDeleted)
	eventer.Listen(EventProjectChanged, repo.HandleProjectChanged)
	eventer.Listen(EventProjectDeleted, repo.HandleProjectDeleted)
	return repo
}

func (repo *eventRepo) HandleAuditLog(data any, e event.Event) error {
	logData := data.(AuditLog)
	var fid *int
	if logData.GetFileID() != 0 {
		ffid := logData.GetFileID()
		fid = &ffid
	}
	var db = repo.db
	db.Event.Create().SetAction(logData.GetAction()).
		SetUsername(logData.GetUsername()).
		SetMessage(logData.GetMsg()).
		SetOld(logData.GetOldStr()).
		SetNew(logData.GetNewStr()).
		SetNillableFileID(fid).
		Save(context.TODO())

	return nil
}

func (repo *eventRepo) Dispatch(created event.Event, createdData any) {
	repo.eventer.Dispatch(created, createdData)
}

type ListEventInput struct {
	Page, PageSize int64
	ActionType     types.EventActionType
	Search         string
	OrderIDDesc    *bool
}

func (repo *eventRepo) List(ctx context.Context, input *ListEventInput) (events []*ent.Event, pag *pagination.Pagination, err error) {
	query := repo.db.Event.Query().Where(
		filters.IfInt64EQ(entevent.FieldAction)(int64(input.ActionType)),
		filters.IfOrderByDesc("id")(input.OrderIDDesc),
		filters.If(func(t string) bool {
			return t != ""
		}, func(t string) func(*sql.Selector) {
			return entevent.Or(
				entevent.MessageContains(t),
				entevent.UsernameContains(t),
			)
		})(input.Search),
	)
	items := query.Clone().
		Offset(pagination.GetPageOffset(int(input.Page), int(input.PageSize))).
		Limit(int(input.PageSize)).
		AllX(ctx)

	return items, &pagination.Pagination{
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}

func (repo *eventRepo) Show(ctx context.Context, id int) (*ent.Event, error) {
	return repo.db.Event.Query().WithFile().Where(entevent.ID(id)).First(ctx)
}

func (repo *eventRepo) AuditLog(action types.EventActionType, username string, msg string) {
	repo.AuditLogWithChange(action, username, msg, nil, nil)
}

func (repo *eventRepo) FileAuditLog(action types.EventActionType, username string, msg string, fileId int) {
	repo.eventer.Dispatch(AuditLogEvent, NewEventAuditLog(username, action, msg, AuditWithFileID(fileId)))
}

func (repo *eventRepo) AuditLogWithChange(action types.EventActionType, username string, msg string, oldS, newS YamlPrettier) {
	if oldS == nil {
		oldS = &emptyYamlPrettier{}
	}
	if newS == nil {
		newS = &emptyYamlPrettier{}
	}
	repo.eventer.Dispatch(AuditLogEvent, NewEventAuditLog(username, action, msg, AuditWithOldNewStr(oldS.PrettyYaml(), newS.PrettyYaml())))
}

type NamespaceCreatedData struct {
	NsModel  *ent.Namespace
	NsK8sObj *corev1.Namespace
}

func (repo *eventRepo) HandleInjectTlsSecret(data any, e event.Event) error {
	if createdData, ok := data.(NamespaceCreatedData); ok {
		name, key, crt := repo.pl.Domain().GetCerts()
		if name != "" && key != "" && crt != "" {
			ns := createdData.NsK8sObj.Name
			err := domainmanager.AddTlsSecret(repo.k8sCli, ns, name, key, crt)
			if err != nil {
				repo.logger.Error(err)
			}
		}
	}
	return nil
}

type NamespaceDeletedData struct {
	NsModel *ent.Namespace
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
		NamespaceId: int64(data.(NamespaceDeletedData).NsModel.ID),
	})
	logger.Debug("event handled: ", e.String())
	return nil
}

type ProjectChangedData struct {
	Project *ent.Project

	Username string
}

func (repo *eventRepo) HandleProjectChanged(data any, e event.Event) error {
	if changedData, ok := data.(*ProjectChangedData); ok {
		last, _ := repo.clRepo.FindLastChangeByProjectID(context.TODO(), changedData.Project.ID)
		gp, _ := repo.gitprojectRepo.GetByProjectID(context.TODO(), changedData.Project.ID)
		var (
			configChanged bool
			version       int = 1
		)
		if last != nil {
			if last.Config != changedData.Project.Config || last.GitCommit != changedData.Project.GitCommit {
				configChanged = true
			}
			version = last.Version + 1
		}
		repo.clRepo.Create(context.TODO(), &CreateChangeLogInput{
			Version:          version,
			Username:         changedData.Username,
			Manifest:         changedData.Project.Manifest,
			Config:           changedData.Project.Config,
			ConfigType:       changedData.Project.ConfigType,
			GitBranch:        changedData.Project.GitBranch,
			GitCommit:        changedData.Project.GitCommit,
			DockerImage:      changedData.Project.DockerImage,
			EnvValues:        changedData.Project.EnvValues,
			ExtraValues:      changedData.Project.ExtraValues,
			FinalExtraValues: changedData.Project.FinalExtraValues,
			GitCommitWebURL:  changedData.Project.GitCommitWebURL,
			GitCommitTitle:   changedData.Project.GitCommitTitle,
			GitCommitAuthor:  changedData.Project.GitCommitAuthor,
			GitCommitDate:    changedData.Project.GitCommitDate,
			ConfigChanged:    configChanged,
			ProjectID:        changedData.Project.ID,
			GitProjectID:     gp.ID,
		})
	}
	return nil
}

func (repo *eventRepo) HandleProjectDeleted(data any, e event.Event) error {
	var (
		ws     = repo.pl.Ws()
		logger = repo.logger
	)
	project := data.(*ent.Project)
	sub := ws.New("", "")
	defer sub.Close()
	ws.New("", "").ToAll(&websocket_pb.WsReloadProjectsResponse{
		Metadata:    &websocket_pb.Metadata{Type: websocket_pb.Type_ReloadProjects},
		NamespaceId: int64(project.NamespaceID),
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
}

var _ AuditLog = (*auditLogImpl)(nil)

type auditLogImpl struct {
	Username        string
	Action          types.EventActionType
	Msg, OldS, NewS string
	FileId          int
}

type AuditOption func(*auditLogImpl)

func AuditWithOldNewStr(o, n string) AuditOption {
	return func(e *auditLogImpl) {
		e.OldS = o
		e.NewS = n
	}
}

func AuditWithFileID(id int) AuditOption {
	return func(e *auditLogImpl) {
		e.FileId = id
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
