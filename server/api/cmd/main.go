package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/duc-cnzj/mars/internal/app"
	"github.com/duc-cnzj/mars/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	_ "github.com/duc-cnzj/mars/internal/plugins/docker"
	_ "github.com/duc-cnzj/mars/internal/plugins/domain_resolver"
	_ "github.com/duc-cnzj/mars/internal/plugins/wssender"
	"github.com/duc-cnzj/mars/pkg/cluster"
	"github.com/duc-cnzj/mars/pkg/gitlab"
	"github.com/duc-cnzj/mars/pkg/mars"
	"github.com/duc-cnzj/mars/pkg/namespace"
	"github.com/duc-cnzj/mars/pkg/project"
	"github.com/duc-cnzj/mars/server/api/services"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
)

func main() {
	app.DefaultBootstrappers = []contracts.Bootstrapper{
		&bootstrappers.PluginsBootstrapper{},
		&bootstrappers.K8sClientBootstrapper{},
		&bootstrappers.GitlabBootstrapper{},
		&bootstrappers.I18nBootstrapper{},
		&bootstrappers.DBBootstrapper{},
	}
	a := app.NewApplication(config.Init("/Users/duc/goMod/mars/config.yaml"))
	if err := a.Bootstrap(); err != nil {
		mlog.Fatal(err)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	listen, _ := net.Listen("tcp", ":9999")
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
				mlog.Error(err)
				return nil
			})),
		),
	)

	cluster.RegisterClusterServer(server, new(services.Cluster))
	gitlab.RegisterGitlabServer(server, new(services.Gitlab))
	mars.RegisterMarsServer(server, new(services.Mars))
	namespace.RegisterNamespaceServer(server, new(services.Namespace))
	project.RegisterProjectServer(server, new(services.Project))

	go func() {
		if err := server.Serve(listen); err != nil {
			mlog.Error(err)
		}
	}()
	<-sig
	server.GracefulStop()
	a.Shutdown()
}
