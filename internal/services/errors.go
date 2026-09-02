package services

import (
	"context"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

// logError 记录错误日志并原样返回错误，收敛各 service 中重复的
// logger.ErrorCtx(ctx, err); return nil, err 三段式样板。行为与原文完全等价：
// 先落日志再返回原始错误，调用方无需区分二者。
//
// 服务层错误日志是唯一出口（gRPC 错误日志拦截器已移除）：每个 service 的 logger
// 自带 WithModule("services/xxx") 模块标签，日志按服务归属，而非中间件的统一 "grpc"。
func logError(ctx context.Context, logger mlog.Logger, err error) error {
	// logError 自身引入一帧调用栈，必须经 CallerSkipAdjuster 补偿一帧，
	// 否则 file 字段指向本文件（services/errors.go）而非真实调用方。
	if a, ok := logger.(mlog.CallerSkipAdjuster); ok {
		logger = a.WithCallerSkip(1)
	}
	logger.ErrorCtx(ctx, err)
	return err
}
