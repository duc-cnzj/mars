//go:generate mockgen -destination ./mock_server.go -package server github.com/duc-cnzj/mars/v5/internal/server HttpServer,GrpcServerImp
package server
