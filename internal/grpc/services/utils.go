package services

import (
	"context"
	"errors"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/event/events"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

var ErrorPermissionDenied = errors.New("没有权限执行该操作")

var MustGetUser = auth.MustGetUser

var AuditLog = func(username string, action types.EventActionType, msg string) {
	events.AuditLog(username, action, msg, nil, nil)
}
var FileAuditLog = func(username string, msg string, fileID int) {
	events.FileAuditLog(username, msg, fileID)
}
var AuditLogWithChange = events.AuditLog

type guest struct{}

func (c guest) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}
