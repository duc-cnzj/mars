package services

import (
	"errors"

	"github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/event/events"
	eventpb "github.com/duc-cnzj/mars/pkg/event"
)

var ErrorPermissionDenied = errors.New("没有权限执行该操作")

var MustGetUser = auth.MustGetUser

var AuditLog = func(username string, action eventpb.ActionType, msg string) {
	events.AuditLog(username, action, msg, nil, nil)
}
var FileAuditLog = func(username string, msg string, fileID int) {
	events.FileAuditLog(username, msg, fileID)
}
var AuditLogWithChange = events.AuditLog
