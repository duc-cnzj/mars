package services

import (
	"context"
	"errors"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/event/events"
	eventpb "github.com/duc-cnzj/mars/pkg/event"
)

var ErrorPermissionDenied = errors.New("没有权限执行该操作")

type CtxTokenInfo struct{}

func SetUser(ctx context.Context, info *contracts.UserInfo) context.Context {
	return context.WithValue(ctx, &CtxTokenInfo{}, info)
}

func GetUser(ctx context.Context) (*contracts.UserInfo, error) {
	if info, ok := ctx.Value(&CtxTokenInfo{}).(*contracts.UserInfo); ok {
		return info, nil
	}

	return nil, errors.New("user not found")
}
func MustGetUser(ctx context.Context) *contracts.UserInfo {
	info, _ := ctx.Value(&CtxTokenInfo{}).(*contracts.UserInfo)
	return info
}

var AuditLog = func(username string, action eventpb.ActionType, msg string) {
	events.AuditLog(username, action, msg, nil, nil)
}
var AuditLogWithChange = events.AuditLog
