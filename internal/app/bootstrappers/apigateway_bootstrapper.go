package bootstrappers

import (
	"bufio"
	"bytes"
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
	"strings"
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/v4/frontend"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	e "github.com/duc-cnzj/mars/v4/internal/event/events"
	"github.com/duc-cnzj/mars/v4/internal/grpc/services"
	"github.com/duc-cnzj/mars/v4/internal/middlewares"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/socket"
	"github.com/duc-cnzj/mars/v4/internal/utils"
	"github.com/duc-cnzj/mars/v4/third_party/doc/data"
	swagger_ui "github.com/duc-cnzj/mars/v4/third_party/doc/swagger-ui"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const maxRecvMsgSize = 1 << 20 * 20 // 20 MiB

var defaultMiddlewares = middlewareList{
	middlewares.Recovery,
	middlewares.DeletePatternHeader,
	middlewares.ResponseMetrics,
	middlewares.TracingWrapper,
	middlewares.RouteLogger,
	middlewares.AllowCORS,
}

type ApiGatewayBootstrapper struct{}

func (a *ApiGatewayBootstrapper) Tags() []string {
	return []string{"api", "gateway"}
}

func (a *ApiGatewayBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.AddServer(&apiGateway{endpoint: fmt.Sprintf("localhost:%s", app.Config().GrpcPort), newServerFunc: initServer})
	app.RegisterAfterShutdownFunc(func(app contracts.ApplicationInterface) {
		t := time.NewTimer(5 * time.Second)
		defer t.Stop()
		ch := make(chan struct{})
		go func() {
			socket.Wait.Wait()
			close(ch)
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

type httpServer interface {
	Shutdown(ctx context.Context) error
	ListenAndServe() error
}

type apiGateway struct {
	endpoint string
	server   httpServer

	newServerFunc func(ctx context.Context, a *apiGateway) (httpServer, error)
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

func (a *apiGateway) Run(ctx context.Context) error {
	s, err := a.newServerFunc(ctx, a)
	if err != nil {
		return err
	}

	a.server = s

	go func(s httpServer) {
		mlog.Infof("[Server]: start apiGateway runner at :%s.", app.Config().AppPort)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mlog.Error(err)
		}
	}(s)

	return nil
}

func initServer(ctx context.Context, a *apiGateway) (httpServer, error) {
	router := mux.NewRouter()

	gmux := runtime.NewServeMux(
		runtime.WithOutgoingHeaderMatcher(headerMatcher),
		runtime.WithIncomingHeaderMatcher(headerMatcher),
		runtime.WithForwardResponseOption(func(ctx context.Context, writer http.ResponseWriter, message proto.Message) error {
			writer.Header().Set("X-Content-Type-Options", "nosniff")
			pattern, ok := runtime.HTTPPathPattern(ctx)
			if ok {
				middlewares.SetPatternHeader(writer, pattern)
			}

			return nil
		}),
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
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxRecvMsgSize)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(middlewares.TraceUnaryClientInterceptor),
	}

	for _, f := range services.RegisteredEndpoints() {
		if err := f(ctx, gmux, a.endpoint, opts); err != nil {
			return nil, err
		}
	}

	handFile(gmux)
	handleDownloadConfig(gmux)
	router.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.Write([]byte("pong"))
	})
	serveWs(router)
	frontend.LoadFrontendRoutes(router)
	loadSwaggerUI(router)
	router.PathPrefix("/").Handler(gmux)

	s := &http.Server{
		Addr:    ":" + app.Config().AppPort,
		Handler: defaultMiddlewares.Wrap(router),
	}

	return s, nil
}

type middlewareList []func(handler http.Handler) http.Handler

func (m middlewareList) Wrap(r http.Handler) (h http.Handler) {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](r)
	}
	return
}

func (a *apiGateway) Shutdown(ctx context.Context) error {
	defer mlog.Info("[Server]: shutdown api-gateway runner.")
	if a.server == nil {
		return nil
	}

	return a.server.Shutdown(ctx)
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
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(idstr)
		if err != nil {
			http.Error(w, "bad id", http.StatusBadRequest)
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

	user := auth.MustGetUser(r.Context())
	e.AuditLog(user.Name,
		types.EventActionType_Download,
		fmt.Sprintf("下载文件 '%s', 大小 %s",
			fil.Path, humanize.Bytes(fil.Size)), nil, nil)
	read, err := fil.Uploader().Read(fil.Path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		mlog.Error(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer read.Close()

	download(w, fileName, read)
}

func download(w http.ResponseWriter, filename string, reader io.Reader) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape(filename)))
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Access-Control-Expose-Headers", "*")

	// 调用 Write 之后就会写入 200 code
	if _, err := io.Copy(w, bufio.NewReaderSize(reader, 1024*1024*5)); err != nil {
		mlog.Error(err)
	}
}

func authenticated(r *http.Request) (*http.Request, bool) {
	if verifyToken, b := app.Auth().VerifyToken(r.Header.Get("Authorization")); b {
		return r.WithContext(auth.SetUser(r.Context(), &verifyToken.UserInfo)), true
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

	info := auth.MustGetUser(r.Context())

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

	file := models.File{Path: put.Path(), Username: info.Name, Size: put.Size(), UploadType: uploader.Type()}
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

func serveWs(mux *mux.Router) {
	ws := socket.NewWebsocketManager(15 * time.Second)
	app.App().BeforeServerRunHooks(func(contracts.ApplicationInterface) {
		ws.TickClusterHealth()
	})
	mux.HandleFunc("/api/ws_info", ws.Info).Name("ws_info")
	mux.HandleFunc("/ws", ws.Ws).Name("ws")
}

type ExportProject struct {
	DefaultBranch string `json:"default_branch"`
	Name          string `json:"name"`
	GitProjectId  int    `json:"git_project_id"`
	Enabled       bool   `json:"enabled"`
	GlobalEnabled bool   `json:"global_enabled"`
	GlobalConfig  string `json:"global_config"`
}

func handleDownloadConfig(gmux *runtime.ServeMux) {
	gmux.HandlePath("GET", "/api/config/export/{git_project_id}", exportMarsConfig)
	gmux.HandlePath("GET", "/api/config/export", exportMarsConfig)
	gmux.HandlePath("POST", "/api/config/import", importMarsConfig)
}

func importMarsConfig(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	req, ok := authenticated(r)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	user := auth.MustGetUser(req.Context())
	if !user.IsAdmin() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseMultipartForm(int64(app.Config().MaxUploadSize())); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %s", err.Error()), http.StatusBadRequest)
		return
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get file 'attachment': %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer f.Close()
	all, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	e.AuditLog(user.Name,
		types.EventActionType_Upload,
		"导入配置文件", nil, &e.StringYamlPrettier{Str: string(all)})
	var data []ExportProject
	err = json.Unmarshal(all, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, item := range data {
		var p models.GitProject
		if err := app.DB().Where("`git_project_id` = ?", item.GitProjectId).First(&p).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			app.DB().Create(&models.GitProject{
				DefaultBranch: item.DefaultBranch,
				Name:          item.Name,
				GitProjectId:  item.GitProjectId,
				Enabled:       item.Enabled,
				GlobalEnabled: item.GlobalEnabled,
				GlobalConfig:  item.GlobalConfig,
			})
		} else {
			app.DB().Model(&p).Updates(map[string]any{
				"default_branch": item.DefaultBranch,
				"name":           item.Name,
				"git_project_id": item.GitProjectId,
				"enabled":        item.Enabled,
				"global_enabled": item.GlobalEnabled,
				"global_config":  item.GlobalConfig,
			})
		}
	}
	w.WriteHeader(204)
	w.Write([]byte(""))
}

func exportMarsConfig(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	req, ok := authenticated(r)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	user := auth.MustGetUser(req.Context())
	if !user.IsAdmin() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var db = app.DB()
	var pname []string
	pid := pathParams["git_project_id"]
	if pid != "" {
		db = db.Where("git_project_id = ?", pid)
	}
	var projects []models.GitProject
	db.Find(&projects)
	data := make([]ExportProject, 0, len(projects))
	for _, gitProject := range projects {
		pname = append(pname, gitProject.Name)
		data = append(data, ExportProject{
			DefaultBranch: gitProject.DefaultBranch,
			Name:          gitProject.Name,
			GitProjectId:  gitProject.GitProjectId,
			Enabled:       gitProject.Enabled,
			GlobalEnabled: gitProject.GlobalEnabled,
			GlobalConfig:  gitProject.GlobalConfig,
		})
	}
	marshal, _ := json.MarshalIndent(&data, "", "\t")
	if pid == "" {
		pname = []string{"全部"}
	}
	e.AuditLog(user.Name,
		types.EventActionType_Download,
		fmt.Sprintf("下载配置文件: %v", strings.Join(pname, ",")), nil, &e.StringYamlPrettier{Str: string(marshal)})
	download(w, "mars-config.json", bytes.NewReader(marshal))
}

func loadSwaggerUI(mux *mux.Router) {
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
