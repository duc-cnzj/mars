package bootstrappers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/frontend"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/socket"
	"github.com/duc-cnzj/mars/pkg/cluster"
	"github.com/duc-cnzj/mars/pkg/gitlab"
	"github.com/duc-cnzj/mars/pkg/mars"
	"github.com/duc-cnzj/mars/pkg/namespace"
	"github.com/duc-cnzj/mars/pkg/project"
	"github.com/duc-cnzj/mars/third_party/doc/data"
	swagger_ui "github.com/duc-cnzj/mars/third_party/doc/swagger-ui"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

type ApiGatewayBootstrapper struct{}

func (a *ApiGatewayBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.AddServer(&apiGateway{endpoint: grpcEndpoint})

	return nil
}

type apiGateway struct {
	endpoint string
	server   *http.Server
}

func (a *apiGateway) Run(ctx context.Context) error {
	mlog.Debug("[Runner]: start apiGateway runner.")
	router := mux.NewRouter()
	gmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
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
		if err := f(ctx, gmux, a.endpoint, opts); err != nil {
			return err
		}
	}

	serveWs(router)
	frontend.LoadFrontendRoutes(router)
	LoadSwaggerUI(router)

	router.PathPrefix("/").Handler(gmux)

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	s := &http.Server{
		Addr:    ":" + app.Config().AppPort,
		Handler: routeLogger(allowCORS(router)),
	}

	a.server = s

	go func() {
		mlog.Info("api-gateway start at ", s.Addr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mlog.Error(err)
		}
	}()

	return nil
}

func (a *apiGateway) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		mlog.Error(err)
	}

	mlog.Info("[Runner]: shutdown api-gateway runner.")

	return nil
}

func serveWs(mux *mux.Router) {
	ws := socket.NewWebsocketManager()
	mux.HandleFunc("/api/ws_info", ws.Info)
	mux.HandleFunc("/ws", ws.Ws)
}

func LoadSwaggerUI(mux *mux.Router) {
	mux.HandleFunc("/doc/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data.SwaggerJson)
	})

	mux.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.FS(swagger_ui.SwaggerUI))))
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
			mlog.Debugf("[Http]: method: %v, url: %v, use %v", r.Method, r.URL, time.Since(t))
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
