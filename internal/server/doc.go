// Package server 提供 api-gateway 的传输层：grpcRunner 承载 gRPC、apiGateway 用
// grpc-gateway 承载 HTTP/前端/websocket、metricsRunner 暴露 /metrics、pprofRunner
// 暴露本机剖析端点；四个 runner 均实现 application.Server 生命周期（Run/Shutdown）。
//
// HttpServer 是 HTTP 传输层的最小抽象，定义于此供 apiGateway/metricsRunner/
// pprofRunner 三个 runner 共享（测试替身缝，见 *_test.go 的 fake 实现）。
package server

//go:generate go tool mockgen -destination ./mock_server_test.go -package server github.com/duc-cnzj/mars/v6/internal/server HttpServer,GrpcServerImp

import "context"

// HttpServer 是 *http.Server 的最小抽象：仅暴露启动（ListenAndServe）与优雅停止
// （Shutdown），供 apiGateway/metricsRunner/pprofRunner 持有时可注入测试替身。
type HttpServer interface {
	Shutdown(ctx context.Context) error
	ListenAndServe() error
}
