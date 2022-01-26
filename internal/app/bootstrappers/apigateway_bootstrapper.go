package bootstrappers

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	e "github.com/duc-cnzj/mars/internal/event/events"
	"github.com/dustin/go-humanize"

	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/pkg/file"

	"github.com/duc-cnzj/mars/frontend"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/middlewares"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/socket"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/auth"
	"github.com/duc-cnzj/mars/pkg/changelog"
	"github.com/duc-cnzj/mars/pkg/cluster"
	"github.com/duc-cnzj/mars/pkg/container"
	"github.com/duc-cnzj/mars/pkg/event"
	"github.com/duc-cnzj/mars/pkg/gitserver"
	"github.com/duc-cnzj/mars/pkg/mars"
	rpcmetrics "github.com/duc-cnzj/mars/pkg/metrics"
	"github.com/duc-cnzj/mars/pkg/namespace"
	"github.com/duc-cnzj/mars/pkg/picture"
	"github.com/duc-cnzj/mars/pkg/project"
	"github.com/duc-cnzj/mars/pkg/version"
	"github.com/duc-cnzj/mars/third_party/doc/data"

	swagger_ui "github.com/duc-cnzj/mars/third_party/doc/swagger-ui"
	"github.com/gorilla/mux"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

type ApiGatewayBootstrapper struct{}

func (a *ApiGatewayBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.AddServer(&apiGateway{endpoint: fmt.Sprintf("localhost:%s", app.Config().GrpcPort)})
	app.RegisterAfterShutdownFunc(func(app contracts.ApplicationInterface) {
		t := time.NewTimer(5 * time.Second)
		defer t.Stop()
		ch := make(chan struct{}, 1)
		go func() {
			socket.Wait.Wait()
			ch <- struct{}{}
		}()
		select {
		case <-ch:
			mlog.Info("[Websocket]: all socket connection closed")
		case <-t.C:
			mlog.Warningf("[Websocket]: 等待超时, 未等待所有 socket 连接退出，当前剩余连接 %v 个。", socket.Wait.Count())
		}
	})

	return nil
}

type apiGateway struct {
	endpoint string
	server   *http.Server
}

func (a *apiGateway) Run(ctx context.Context) error {
	mlog.Infof("[Server]: start apiGateway runner at %s.", a.endpoint)

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

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithFilterFunc(middlewares.TracingIgnoreFn), grpc_opentracing.WithTracer(opentracing.GlobalTracer()))),
	}
	var serviceList = []func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error){
		namespace.RegisterNamespaceHandlerFromEndpoint,
		cluster.RegisterClusterHandlerFromEndpoint,
		gitserver.RegisterGitServerHandlerFromEndpoint,
		mars.RegisterMarsHandlerFromEndpoint,
		project.RegisterProjectHandlerFromEndpoint,
		picture.RegisterPictureHandlerFromEndpoint,
		auth.RegisterAuthHandlerFromEndpoint,
		container.RegisterContainerSvcHandlerFromEndpoint,
		rpcmetrics.RegisterMetricsHandlerFromEndpoint,
		version.RegisterVersionHandlerFromEndpoint,
		changelog.RegisterChangelogHandlerFromEndpoint,
		event.RegisterEventHandlerFromEndpoint,
		file.RegisterFileSvcHandlerFromEndpoint,
	}

	for _, f := range serviceList {
		if err := f(ctx, gmux, a.endpoint, opts); err != nil {
			return err
		}
	}

	handFile(gmux)
	router.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.Write([]byte("pong"))
	})
	serveWs(router)
	frontend.LoadFrontendRoutes(router)
	LoadSwaggerUI(router)
	router.PathPrefix("/").Handler(gmux)

	s := &http.Server{
		Addr: ":" + app.Config().AppPort,
		Handler: middlewares.TracingWrapper(
			middlewares.RouteLogger(
				middlewares.AllowCORS(
					router,
				),
			),
		),
	}

	a.server = s

	go func(s *http.Server) {
		mlog.Info("api-gateway start at ", s.Addr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mlog.Error(err)
		}
	}(s)

	return nil
}

func handFile(gmux *runtime.ServeMux) {
	gmux.HandlePath("POST", "/api/files", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		if req, ok := authenticated(r); ok {
			handleBinaryFileUpload(w, req)
			return
		}
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
	})
	gmux.HandlePath("GET", "/api/download_file/{id}", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		idstr, ok := pathParams["id"]
		if !ok {
			http.Error(w, fmt.Sprintf("missing id"), http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(idstr)
		if err != nil {
			http.Error(w, fmt.Sprintf("bad id"), http.StatusBadRequest)
			return
		}
		if req, ok := authenticated(r); ok {
			handleDownload(w, req, id)
			return
		}
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
	})
}

func handleDownload(w http.ResponseWriter, r *http.Request, fid int) {
	var fil = &models.File{ID: fid}
	if err := app.DB().First(&fil).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	fileName := filepath.Base(fil.Path)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape(fileName)))
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Access-Control-Expose-Headers", "*")

	user := r.Context().Value(authCtx{}).(*contracts.UserInfo)
	e.AuditLog(user.Name,
		event.ActionType_Download,
		fmt.Sprintf("下载文件 '%s', 大小 %s",
			fil.Path, humanize.Bytes(fil.Size)), nil, nil)
	open, err := os.Open(fil.Path)
	if err != nil {
		mlog.Error(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer open.Close()

	if _, err := io.Copy(w, bufio.NewReaderSize(open, 1024*1024*5)); err != nil {
		mlog.Error(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

type authCtx struct{}

func authenticated(r *http.Request) (*http.Request, bool) {
	if verifyToken, b := app.Auth().VerifyToken(r.Header.Get("Authorization")); b {
		return r.WithContext(context.WithValue(r.Context(), authCtx{}, &verifyToken.UserInfo)), true
	}

	return nil, false
}

func handleBinaryFileUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(int64(app.Config().MaxUploadSize())); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %s", err.Error()), http.StatusBadRequest)
		return
	}

	f, h, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get file 'attachment': %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer f.Close()

	info := r.Context().Value(authCtx{}).(*contracts.UserInfo)

	var uploader contracts.Uploader = app.Uploader()
	// 某个用户/那天/时间/文件名称
	put, err := uploader.Disk("users").Put(
		fmt.Sprintf("%s/%s/%s/%s",
			info.Name,
			time.Now().Format("2006-01-02"),
			fmt.Sprintf("%s-%s", time.Now().Format("15-04-05"), utils.RandomString(20)),
			h.Filename), f)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to upload file %s", err.Error()), http.StatusInternalServerError)
		return
	}

	file := models.File{Path: put.Path(), Username: info.Name, Size: put.Size()}
	app.DB().Create(&file)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	var res = struct {
		ID int `json:"id"`
	}{
		ID: file.ID,
	}
	marshal, _ := json.Marshal(&res)
	w.Write(marshal)
}

func (a *apiGateway) Shutdown(ctx context.Context) error {
	defer mlog.Info("[Server]: shutdown api-gateway runner.")
	if a.server == nil {
		return nil
	}

	return a.server.Shutdown(ctx)
}

func serveWs(mux *mux.Router) {
	ws := socket.NewWebsocketManager()
	ws.TickClusterHealth()
	mux.HandleFunc("/api/ws_info", ws.Info)
	mux.HandleFunc("/ws", ws.Ws)
}

func LoadSwaggerUI(mux *mux.Router) {
	subrouter := mux.PathPrefix("").Subrouter()
	subrouter.Use(middlewares.HttpCache)

	subrouter.Handle("/doc/swagger.json",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(data.SwaggerJson)
		}),
	)

	subrouter.PathPrefix("/docs/").Handler(
		http.StripPrefix("/docs/", http.FileServer(http.FS(swagger_ui.SwaggerUI))),
	)
}
