package cmd

import (
	"context"
	"github.com/duc-cnzj/mars/internal/app"
	"github.com/duc-cnzj/mars/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/controllers"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/pkg/cluster"
	"github.com/duc-cnzj/mars/pkg/gitlab"
	"github.com/duc-cnzj/mars/pkg/mars"
	"github.com/duc-cnzj/mars/pkg/namespace"
	"github.com/duc-cnzj/mars/pkg/project"
	"github.com/duc-cnzj/mars/server/api/services"
	"github.com/duc-cnzj/mars/third_party/doc/data"
	swagger_ui "github.com/duc-cnzj/mars/third_party/doc/swagger-ui"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var apiGatewayCmd = &cobra.Command{
	Use:   "grpc",
	Short: "start mars server use grpc.",
	Run: func(cmd *cobra.Command, args []string) {
		app.DefaultBootstrappers = []contracts.Bootstrapper{
			&bootstrappers.PluginsBootstrapper{},
			&bootstrappers.K8sClientBootstrapper{},
			&bootstrappers.GitlabBootstrapper{},
			&bootstrappers.I18nBootstrapper{},
			&bootstrappers.DBBootstrapper{},
		}
		a := app.NewApplication(config.Init(cfgFile))
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
		done := RunApiGateway()
		<-sig
		server.GracefulStop()
		a.Shutdown()
		done()
	},
}


var endpoint = "localhost:9999"

func RunApiGateway() func() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	mux := http.NewServeMux()

	gmux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseEnumNumbers:  true,
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var serviceList = []func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error){
		namespace.RegisterNamespaceHandlerFromEndpoint,
		cluster.RegisterClusterHandlerFromEndpoint,
		gitlab.RegisterGitlabHandlerFromEndpoint,
		mars.RegisterMarsHandlerFromEndpoint,
		project.RegisterProjectHandlerFromEndpoint,
	}

	for _, f := range serviceList {
		fatalError(f(ctx, gmux, endpoint, opts))
	}

	serveWs(mux)
	mux.Handle("/", gmux)

	runSwaggerUI()

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	s := &http.Server{
		Addr:    ":4000",
		Handler: routeLogger(allowCORS(mux)),
	}

	go func() {
		mlog.Info("api-gateway start at: ", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			mlog.Error("error: ", err)
		}
	}()

	return func() {
		timeout, cancelFunc := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()
		defer cancelFunc()
		s.Shutdown(timeout)
		mlog.Info("api-gateway shutdown")
	}
}

func serveWs(mux *http.ServeMux) {
	//e.GET("/ws", wsC.Ws)
	//api.GET("/ws_info", wsC.Info)
	//response.Success(ctx, 200, utils.ClusterInfo())
	//mux.HandleFunc("/ws_info", )
	ws := &controllers.WebsocketController{}
	mux.HandleFunc("/ws", ws.Ws)
}

func runSwaggerUI() {
	http.HandleFunc("/doc/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data.SwaggerJson)
	})

	http.Handle("/", http.FileServer(http.FS(swagger_ui.SwaggerUI)))

	mlog.Info("swagger ui running at: 8888")
	go func() {
		http.ListenAndServe(":8888", nil)
	}()
}

func fatalError(err error) {
	if err != nil {
		mlog.Fatal(err)
	}
}
func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "X-Requested-With"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	return
}

func routeLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(t time.Time) {
			mlog.Infof("method: %v, url: %v, use %v", r.Method, r.URL, time.Since(t))
		}(time.Now())
		h.ServeHTTP(w, r)
	})
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
