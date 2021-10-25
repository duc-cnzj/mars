package services

import (
	"context"
	"errors"
)

var ErrorPermissionDenied = errors.New("没有权限执行该操作")

type CtxTokenInfo struct{}

func SetUser(ctx context.Context, info *UserInfo) context.Context {
	return context.WithValue(ctx, &CtxTokenInfo{}, info)
}

func GetUser(ctx context.Context) (*UserInfo, error) {
	if info, ok := ctx.Value(&CtxTokenInfo{}).(*UserInfo); ok {
		return info, nil
	}

	return nil, errors.New("user not found")
}
func MustGetUser(ctx context.Context) *UserInfo {
	info, _ := ctx.Value(&CtxTokenInfo{}).(*UserInfo)
	return info
}
