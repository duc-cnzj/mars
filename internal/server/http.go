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

	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/doc"
	"github.com/duc-cnzj/mars/v4/frontend"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/file"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/server/middlewares"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util/rand"
	swagger_ui "github.com/duc-cnzj/mars/v4/third_party/swagger-ui"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

type apiGateway struct {
	endpoint string
	server   httpServer
	app      application.App

	newServerFunc func(ctx context.Context, a *apiGateway) (httpServer, error)
}

func NewApiGateway(endpoint string, app application.App) application.Server {
	return &apiGateway{endpoint: endpoint, app: app, newServerFunc: initServer}
}

func (a *apiGateway) Run(ctx context.Context) error {
	s, err := a.newServerFunc(ctx, a)
	if err != nil {
		return err
	}

	a.server = s

	go func(s httpServer) {
		a.app.Logger().Infof("[Server]: start apiGateway runner at :%s.", a.app.Config().AppPort)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.app.Logger().Error(err)
		}
	}(s)

	return nil
}

func (a *apiGateway) Shutdown(ctx context.Context) error {
	defer a.app.Logger().Info("[Server]: shutdown api-gateway runner.")
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
			pattern, ok := runtime.HTTPPathPattern(ctx)
			if ok {
				middlewares.SetPatternHeader(writer, pattern)
			}

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
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxRecvMsgSize)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(middlewares.TraceUnaryClientInterceptor),
	}

	for _, f := range a.app.GrpcRegistry().EndpointFuncs {
		if err := f(ctx, gmux, a.endpoint, opts); err != nil {
			return nil, err
		}
	}

	router.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.Write([]byte("pong"))
	})

	h := &handler{
		app:      a.app,
		logger:   a.app.Logger(),
		ServeMux: gmux,
	}
	h.handFile()
	h.handleDownloadConfig()
	h.serveWs(router)
	frontend.LoadFrontendRoutes(router)
	h.loadSwaggerUI(router)
	router.PathPrefix("/").Handler(gmux)

	s := &http.Server{
		Addr:              ":" + a.app.Config().AppPort,
		Handler:           defaultMiddlewares.Wrap(a.app.Logger(), router),
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
	app    application.App
	logger mlog.Logger

	*runtime.ServeMux
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

func (h *handler) handleDownload(w http.ResponseWriter, r *http.Request, fid int) {
	fil, err := h.app.DB().File.Query().Where(file.ID(fid)).Only(context.TODO())
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
	h.app.Dispatcher().Dispatch(repo.AuditLogEvent, repo.NewEventAuditLog(
		user.Name,
		types.EventActionType_Download,
		fmt.Sprintf("下载文件 '%s', 大小 %s",
			fil.Path, humanize.Bytes(fil.Size)),
	))
	read, err := h.app.Uploader().Read(fil.Path)
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
	if verifyToken, b := h.app.Auth().VerifyToken(r.Header.Get("Authorization")); b {
		return r.WithContext(auth.SetUser(r.Context(), verifyToken.UserInfo)), true
	}

	return nil, false
}

func (h *handler) handleBinaryFileUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(int64(h.app.Config().MaxUploadSize())); err != nil {
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

	var uploader uploader.Uploader = h.app.Uploader()
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

	createdFile, _ := h.app.DB().File.Create().
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

func (h *handler) serveWs(mux *mux.Router) {
	ws := h.app.WsServer()
	h.app.BeforeServerRunHooks(func(app application.App) {
		go ws.TickClusterHealth(app.Done())
	})
	mux.HandleFunc("/api/ws_info", ws.Info).Name("ws_info")
	mux.HandleFunc("/ws", ws.Serve).Name("ws")
}

type ExportProject struct {
	DefaultBranch string       `json:"default_branch"`
	Name          string       `json:"name"`
	GitProjectId  int          `json:"git_project_id"`
	Enabled       bool         `json:"enabled"`
	GlobalEnabled bool         `json:"global_enabled"`
	GlobalConfig  *mars.Config `json:"global_config"`
}

func (h *handler) handleDownloadConfig() {
	//h.HandlePath("GET", "/api/config/export/{git_project_id}", h.exportMarsConfig)
	//h.HandlePath("GET", "/api/config/export", h.exportMarsConfig)
	//h.HandlePath("POST", "/api/config/import", h.importMarsConfig)
}

//func (h *handler) importMarsConfig(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
//	req, ok := h.authenticated(r)
//	if !ok {
//		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
//		return
//	}
//	user := auth.MustGetUser(req.Context())
//	if !user.IsAdmin() {
//		http.Error(w, "Forbidden", http.StatusForbidden)
//		return
//	}
//	if err := r.ParseMultipartForm(int64(h.app.Config().MaxUploadSize())); err != nil {
//		http.Error(w, fmt.Sprintf("failed to parse form: %s", err.Error()), http.StatusBadRequest)
//		return
//	}
//
//	f, _, err := r.FormFile("file")
//	if err != nil {
//		http.Error(w, fmt.Sprintf("failed to get file 'attachment': %s", err.Error()), http.StatusBadRequest)
//		return
//	}
//	defer f.Close()
//	all, err := io.ReadAll(f)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//	var data []ExportProject
//	err = json.Unmarshal(all, &data)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//	var oldExportProjects []*ent.GitProject
//	for _, item := range data {
//		var p *ent.GitProject
//		first, err := h.app.DB().GitProject.Query().Where(gitproject.GitProjectID(item.GitProjectId)).First(req.Context())
//		if ent.IsNotFound(err) {
//			h.app.DB().GitProject.Create().SetDefaultBranch(item.DefaultBranch).
//				SetName(item.Name).
//				SetGitProjectID(item.GitProjectId).
//				SetEnabled(item.Enabled).
//				SetGlobalEnabled(item.GlobalEnabled).
//				SetGlobalConfig(item.GlobalConfig).
//				Save(context.TODO())
//		} else {
//			p, _ = first.Update().
//				SetDefaultBranch(item.DefaultBranch).
//				SetName(item.Name).
//				SetEnabled(item.Enabled).
//				SetGlobalEnabled(item.GlobalEnabled).
//				SetGlobalConfig(item.GlobalConfig).
//				Save(context.TODO())
//			oldExportProjects = append(oldExportProjects, p)
//		}
//	}
//
//	h.app.Dispatcher().Dispatch(repo.AuditLogEvent, repo.NewEventAuditLog(
//		user.Name,
//		types.EventActionType_Upload,
//		"导入配置文件",
//		repo.AuditWithOldNewStr(
//			gitProjectList(oldExportProjects).ExportJsonString(),
//			string(all),
//		),
//	))
//
//	w.WriteHeader(204)
//	w.Write([]byte(""))
//}

//func (h *handler) exportMarsConfig(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
//	db := h.app.DB()
//	req, ok := h.authenticated(r)
//	if !ok {
//		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
//		return
//	}
//	user := auth.MustGetUser(req.Context())
//	if !user.IsAdmin() {
//		http.Error(w, "Forbidden", http.StatusForbidden)
//		return
//	}
//	query := db.GitProject.Query()
//	pid := pathParams["git_project_id"]
//	if pid != "" {
//		query.Where(gitproject.GitProjectID(cast.ToInt(pid)))
//	}
//	projects, _ := query.All(context.TODO())
//	var pname []string = gitProjectList(projects).ExportNames()
//	jsonString := gitProjectList(projects).ExportJsonString()
//	if pid == "" {
//		pname = []string{"全部"}
//	}
//	h.app.Dispatcher().Dispatch(repo.AuditLogEvent, repo.NewEventAuditLog(
//		user.Name,
//		types.EventActionType_Download,
//		fmt.Sprintf("下载配置文件: %v", strings.Join(pname, ",")),
//		repo.AuditWithOldNewStr("", jsonString),
//	))
//	fileName := "mars-config.json"
//	if pid != "" && len(pname) == 1 {
//		fileName = pname[0] + ".json"
//	}
//	h.download(w, fileName, strings.NewReader(jsonString))
//}

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

//type gitProjectList []*ent.GitProject
//
//func (projects gitProjectList) ExportNames() (res []string) {
//	for _, p := range projects {
//		res = append(res, p.Name)
//	}
//	return
//}
//
//func (projects gitProjectList) ExportJsonString() string {
//	exportData := make([]ExportProject, 0, len(projects))
//	for _, gitProject := range projects {
//		exportData = append(exportData, ExportProject{
//			DefaultBranch: gitProject.DefaultBranch,
//			Name:          gitProject.Name,
//			GitProjectId:  gitProject.GitProjectID,
//			Enabled:       gitProject.Enabled,
//			GlobalEnabled: gitProject.GlobalEnabled,
//			GlobalConfig:  gitProject.GlobalConfig,
//		})
//	}
//	marshal, _ := json.MarshalIndent(&exportData, "", "\t")
//	return string(marshal)
//}

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
