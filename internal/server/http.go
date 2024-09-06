package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/v5/frontend"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/server/middlewares"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const maxRecvMsgSize = 1 << 20 * 100 // 100 MiB

var defaultMiddlewares = middlewareList{
	middlewares.Recovery,
	middlewares.RouteLogger,
	middlewares.AllowCORS,
}

type apiGateway struct {
	endpoint      string
	port          string
	server        HttpServer
	logger        mlog.Logger
	grpcRegistry  *application.GrpcRegistry
	handler       application.HttpHandler
	newServerFunc func(ctx context.Context, a *apiGateway) (HttpServer, error)
}

func NewApiGateway(endpoint string, app application.App) application.Server {
	return &apiGateway{
		endpoint:      endpoint,
		port:          app.Config().AppPort,
		logger:        app.Logger().WithModule("server/apiGateway"),
		grpcRegistry:  app.GrpcRegistry(),
		handler:       app.HttpHandler(),
		newServerFunc: initServer,
	}
}

func (a *apiGateway) Run(ctx context.Context) error {
	s, err := a.newServerFunc(ctx, a)
	if err != nil {
		return err
	}

	a.server = s

	go a.handler.TickClusterHealth(ctx.Done())

	go func(s HttpServer) {
		a.logger.Infof("[Server]: start apiGateway runner at :%s.", a.port)
		if err = s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error(err)
		}
	}(s)

	return nil
}

func (a *apiGateway) Shutdown(ctx context.Context) error {
	a.logger.Info("[Server]: shutdown api-gateway runner.")
	a.handler.Shutdown(ctx)
	return a.server.Shutdown(ctx)
}

func initServer(ctx context.Context, a *apiGateway) (HttpServer, error) {
	router := mux.NewRouter()

	gmux := runtime.NewServeMux(
		runtime.WithOutgoingHeaderMatcher(headerMatcher),
		runtime.WithIncomingHeaderMatcher(headerMatcher),
		runtime.WithForwardResponseOption(func(ctx context.Context, writer http.ResponseWriter, message proto.Message) error {
			writer.Header().Set("X-Content-Type-Options", "nosniff")
			return nil
		}),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseEnumNumbers:  false,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	opts := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxRecvMsgSize)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	for _, f := range a.grpcRegistry.EndpointFuncs {
		if err := f(ctx, gmux, a.endpoint, opts); err != nil {
			return nil, err
		}
	}

	router.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.Write([]byte("pong"))
	})

	a.handler.RegisterFileRoute(gmux)
	a.handler.RegisterWsRoute(router)
	frontend.LoadFrontendRoutes(router)
	a.handler.RegisterSwaggerUIRoute(router)
	router.PathPrefix("/").Handler(gmux)

	s := &http.Server{
		Addr: ":" + a.port,
		Handler: defaultMiddlewares.Wrap(
			a.logger,
			otelhttp.NewHandler(
				router,
				"grpc-gateway",
				otelhttp.WithFilter(func(request *http.Request) bool {
					return strings.HasPrefix(request.URL.Path, "/api")
				}),
				otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
					return fmt.Sprintf("grpc-gateway [%s] %s", r.Method, r.URL.Path)
				}),
			),
		),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s, nil
}

type middlewareList []func(logger mlog.Logger, handler http.Handler) http.Handler

func (m middlewareList) Wrap(logger mlog.Logger, r http.Handler) (h http.Handler) {
	if len(m) == 0 {
		return r
	}
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](logger, r)
		r = h
	}
	return
}

func headerMatcher(key string) (string, bool) {
	key = strings.ToLower(key)
	switch key {
	case "tracestate":
		fallthrough
	case "traceparent":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
