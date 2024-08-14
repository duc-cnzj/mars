package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/file"
	"github.com/duc-cnzj/mars/v4/internal/event"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/dustin/go-humanize"

	"github.com/duc-cnzj/mars/v4/doc"
	"github.com/duc-cnzj/mars/v4/frontend"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/server/middlewares"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util/rand"
	swagger_ui "github.com/duc-cnzj/mars/v4/third_party/swagger-ui"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const maxRecvMsgSize = 1 << 20 * 20 // 20 MiB

var defaultMiddlewares = middlewareList{
	middlewares.Recovery,
	middlewares.RouteLogger,
	middlewares.AllowCORS,
}

type apiGateway struct {
	ws            application.WsServer
	endpoint      string
	port          string
	server        httpServer
	logger        mlog.Logger
	grpcRegistry  *application.GrpcRegistry
	newServerFunc func(ctx context.Context, a *apiGateway) (httpServer, error)
	auth          auth.Auth
	maxUploadSize uint64
	data          data.Data
	uploader      uploader.Uploader
	event         event.Dispatcher
}

func NewApiGateway(endpoint string, app application.App) application.Server {
	return &apiGateway{
		ws:            app.WsServer(),
		endpoint:      endpoint,
		port:          app.Config().AppPort,
		logger:        app.Logger().WithModule("server/apiGateway"),
		grpcRegistry:  app.GrpcRegistry(),
		newServerFunc: initServer,
		auth:          app.Auth(),
		maxUploadSize: app.Config().MaxUploadSize(),
		data:          app.Data(),
		uploader:      app.Uploader(),
		event:         app.Dispatcher(),
	}
}

func (a *apiGateway) Run(ctx context.Context) error {
	s, err := a.newServerFunc(ctx, a)
	if err != nil {
		return err
	}

	a.server = s

	go func(s httpServer) {
		a.logger.Infof("[Server]: start apiGateway runner at :%s.", a.port)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error(err)
		}
	}(s)

	return nil
}

func (a *apiGateway) Shutdown(ctx context.Context) error {
	defer a.logger.Info("[Server]: shutdown api-gateway runner.")
	if a.server == nil {
		return nil
	}

	return a.server.Shutdown(ctx)
}

func initServer(ctx context.Context, a *apiGateway) (httpServer, error) {
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
		}))

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

	h := &handler{
		ws:            a.ws,
		logger:        a.logger,
		auth:          a.auth,
		maxUploadSize: a.maxUploadSize,
		uploader:      a.uploader,
		data:          a.data,
		event:         a.event,
		ServeMux:      gmux,
	}
	h.handFile()
	//h.handleDownloadConfig()
	h.serveWs(router)
	frontend.LoadFrontendRoutes(router)
	h.loadSwaggerUI(router)
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
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](logger, r)
		r = h
	}
	return
}

// handler is a http.Handler that handles all incoming requests.
type handler struct {
	ws            application.WsServer
	logger        mlog.Logger
	auth          auth.Auth
	maxUploadSize uint64
	uploader      uploader.Uploader
	data          data.Data
	event         event.Dispatcher

	*runtime.ServeMux
}

func (h *handler) serveWs(mux *mux.Router) {
	mux.HandleFunc("/api/ws_info", h.ws.Info).Name("ws_info")
	mux.HandleFunc("/ws", h.ws.Serve).Name("ws")
}

func (h *handler) handleDownload(w http.ResponseWriter, r *http.Request, fid int) {
	fil, err := h.data.DB().File.Query().Where(file.ID(fid)).Only(context.TODO())
	if err != nil {
		if ent.IsNotFound(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	fileName := filepath.Base(fil.Path)

	user := auth.MustGetUser(r.Context())
	h.event.Dispatch(repo.AuditLogEvent, repo.NewEventAuditLog(
		user.Name,
		types.EventActionType_Download,
		fmt.Sprintf("下载文件 '%s', 大小 %s",
			fil.Path, humanize.Bytes(fil.Size)),
	))
	read, err := h.uploader.Read(fil.Path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		h.logger.Error(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer read.Close()

	h.download(w, fileName, read)
}

func (h *handler) download(w http.ResponseWriter, filename string, reader io.Reader) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape(filename)))
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Access-Control-Expose-Headers", "*")

	// 调用 Write 之后就会写入 200 code
	if _, err := io.Copy(w, bufio.NewReaderSize(reader, 1024*1024*5)); err != nil {
		h.logger.Error(err)
	}
}

func (h *handler) authenticated(r *http.Request) (*http.Request, bool) {
	if verifyToken, b := h.auth.VerifyToken(r.Header.Get("Authorization")); b {
		return r.WithContext(auth.SetUser(r.Context(), verifyToken.UserInfo)), true
	}

	return nil, false
}

func (h *handler) handleBinaryFileUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(int64(h.maxUploadSize)); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %s", err.Error()), http.StatusBadRequest)
		return
	}

	f, fh, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get file 'attachment': %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer f.Close()

	info := auth.MustGetUser(r.Context())

	var uploader uploader.Uploader = h.uploader
	// 某个用户/那天/时间/文件名称
	put, err := uploader.Disk("users").Put(
		fmt.Sprintf("%s/%s/%s/%s",
			info.Name,
			time.Now().Format("2006-01-02"),
			fmt.Sprintf("%s-%s", time.Now().Format("15-04-05"), rand.String(20)),
			fh.Filename), f)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to upload file %s", err.Error()), http.StatusInternalServerError)
		return
	}

	createdFile, _ := h.data.DB().File.Create().
		SetPath(put.Path()).
		SetSize(put.Size()).
		SetUsername(info.Name).
		SetUploadType(uploader.Type()).
		Save(context.TODO())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	var res = struct {
		ID int `json:"id"`
	}{
		ID: createdFile.ID,
	}
	marshal, _ := json.Marshal(&res)
	w.Write(marshal)
}

func (h *handler) handFile() {
	h.HandlePath("POST", "/api/files", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		if req, ok := h.authenticated(r); ok {
			h.handleBinaryFileUpload(w, req)
			return
		}
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
	})
	h.HandlePath("GET", "/api/download_file/{id}", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		idstr, ok := pathParams["id"]
		if !ok {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(idstr)
		if err != nil {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}
		if req, ok := h.authenticated(r); ok {
			h.handleDownload(w, req, id)
			return
		}
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
	})
}

func (h *handler) loadSwaggerUI(mux *mux.Router) {
	subrouter := mux.PathPrefix("").Subrouter()
	subrouter.Use(middlewares.HttpCache)

	subrouter.Handle("/doc/swagger.json",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(doc.SwaggerJson)
		}),
	)

	subrouter.PathPrefix("/docs/").Handler(
		http.StripPrefix("/docs/", http.FileServer(http.FS(swagger_ui.SwaggerUI))),
	)
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
