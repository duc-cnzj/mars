package services

import (
	"context"
	"errors"

	"github.com/duc-cnzj/mars/v4/internal/auth"
)

var ErrorPermissionDenied = errors.New("没有权限执行该操作")

var MustGetUser = auth.MustGetUser

type guest struct{}

func (c guest) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	//mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}
