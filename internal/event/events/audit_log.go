package events

import (
	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/models"
)

const EventAuditLog contracts.Event = "audit_log"

func init() {
	Register(EventAuditLog, HandleAuditLog)
}

func HandleAuditLog(data any, e contracts.Event) error {
	logData := data.(contracts.AuditLogImpl)
	var fid *int
	if logData.GetFileID() != 0 {
		ffid := logData.GetFileID()
		fid = &ffid
	}
	app.DB().Create(&models.Event{
		Action:   uint8(logData.GetAction()),
		Username: logData.GetUsername(),
		Message:  logData.GetMsg(),
		Old:      logData.GetOldStr(),
		New:      logData.GetNewStr(),
		FileID:   fid,
	})

	return nil
}

func AuditLog(username string, action types.EventActionType, msg string, oldS, newS YamlPrettier) {
	if oldS == nil {
		oldS = &emptyYamlPrettier{}
	}
	if newS == nil {
		newS = &emptyYamlPrettier{}
	}
	app.Event().Dispatch(EventAuditLog, &eventAuditLog{
		Username: username,
		Action:   action,
		Msg:      msg,
		OldS:     oldS.PrettyYaml(),
		NewS:     newS.PrettyYaml(),
	})
}

func FileAuditLog(username string, msg string, fileId int) {
	app.Event().Dispatch(EventAuditLog, &eventAuditLog{
		Username: username,
		Action:   types.EventActionType_Upload,
		Msg:      msg,
		FileId:   fileId,
	})
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

var _ contracts.AuditLogImpl = (*eventAuditLog)(nil)

type eventAuditLog struct {
	Username        string
	Action          types.EventActionType
	Msg, OldS, NewS string
	FileId          int
}

type AuditOption func(*eventAuditLog)

func AuditWithOldNewStr(o, n string) AuditOption {
	return func(e *eventAuditLog) {
		e.OldS = o
		e.NewS = n
	}
}

func AuditWithFileID(id int) AuditOption {
	return func(e *eventAuditLog) {
		e.FileId = id
	}
}

func NewEventAuditLog(username string, action types.EventActionType, msg string, opts ...AuditOption) contracts.AuditLogImpl {
	e := &eventAuditLog{Username: username, Action: action, Msg: msg}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *eventAuditLog) GetUsername() string {
	return e.Username
}

func (e *eventAuditLog) GetAction() types.EventActionType {
	return e.Action
}

func (e *eventAuditLog) GetMsg() string {
	return e.Msg
}

func (e *eventAuditLog) GetOldStr() string {
	return e.OldS
}

func (e *eventAuditLog) GetNewStr() string {
	return e.NewS
}

func (e *eventAuditLog) GetFileID() int {
	return e.FileId
}
