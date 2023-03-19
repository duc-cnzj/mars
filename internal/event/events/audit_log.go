package events

import (
	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/models"
)

const EventAuditLog contracts.Event = "audit_log"

type EventAuditLogData struct {
	Username        string
	Action          types.EventActionType
	Msg, OldS, NewS string
	FileId          int
}

func init() {
	Register(EventAuditLog, HandleAuditLog)
}

func HandleAuditLog(data any, e contracts.Event) error {
	logData := data.(EventAuditLogData)
	var fid *int
	if logData.FileId != 0 {
		fid = &logData.FileId
	}
	app.DB().Create(&models.Event{
		Action:   uint8(logData.Action),
		Username: logData.Username,
		Message:  logData.Msg,
		Old:      logData.OldS,
		New:      logData.NewS,
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
	app.Event().Dispatch(EventAuditLog, EventAuditLogData{
		Username: username,
		Action:   action,
		Msg:      msg,
		OldS:     oldS.PrettyYaml(),
		NewS:     newS.PrettyYaml(),
	})
}

func FileAuditLog(username string, msg string, fileId int) {
	app.Event().Dispatch(EventAuditLog, EventAuditLogData{
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
