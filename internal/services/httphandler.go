package services

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
	"time"

	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/doc"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/server/middlewares"
	"github.com/duc-cnzj/mars/v5/internal/uploader"
	"github.com/duc-cnzj/mars/v5/internal/util/rand"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	swagger_ui "github.com/duc-cnzj/mars/v5/third_party/swagger-ui"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

var _ application.HttpHandler = (*httpHandlerImpl)(nil)

type httpHandlerImpl struct {
	application.WsHttpServer
	logger    mlog.Logger
	authRepo  repo.AuthRepo
	uploader  uploader.Uploader
	fileRepo  repo.FileRepo
	eventRepo repo.EventRepo
	timer     timer.Timer
	k8sRepo   repo.K8sRepo
}

func NewHttpHandler(
	wsHttpServer application.WsHttpServer,
	logger mlog.Logger,
	uploader uploader.Uploader,
	authRepo repo.AuthRepo,
	eventRepo repo.EventRepo,
	fileRepo repo.FileRepo,
	timer timer.Timer,
	k8sRepo repo.K8sRepo,
) application.HttpHandler {
	return &httpHandlerImpl{
		WsHttpServer: wsHttpServer,
		logger:       logger.WithModule("services/httpHandler"),
		authRepo:     authRepo,
		uploader:     uploader,
		fileRepo:     fileRepo,
		eventRepo:    eventRepo,
		timer:        timer,
		k8sRepo:      k8sRepo,
	}
}

func (h *httpHandlerImpl) RegisterWsRoute(mux *mux.Router) {
	mux.HandleFunc("/api/ws_info", h.Info).Name("ws_info")
	mux.HandleFunc("/ws", h.Serve).Name("ws")
}

func (h *httpHandlerImpl) RegisterSwaggerUIRoute(mux *mux.Router) {
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

func (h *httpHandlerImpl) RegisterFileRoute(mux *runtime.ServeMux) {
	mux.HandlePath("POST", "/api/files", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		if req, ok := h.authenticated(r); ok {
			h.handleBinaryFileUpload(w, req)
			return
		}
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
	})
	mux.HandlePath("GET", "/api/download_file/{id}", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
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
	mux.HandlePath("POST", "/api/copy_from_pod", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		var input CopyFromPodRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if input.Namespace == "" || input.Pod == "" || input.Container == "" || input.FilePath == "" {
			http.Error(w, "missing namespace, pod or container", http.StatusBadRequest)
			return
		}

		if req, ok := h.authenticated(r); ok {
			ctx := req.Context()
			info := auth.MustGetUser(ctx)
			fromPod, err := h.k8sRepo.CopyFromPod(ctx, &repo.CopyFromPodInput{
				Namespace: input.Namespace,
				Pod:       input.Pod,
				Container: input.Container,
				FilePath:  input.FilePath,
				UserName:  info.Name,
			})
			if err != nil {
				h.logger.Error("Error copying file from pod: ", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			h.eventRepo.FileAuditLog(
				types.EventActionType_Download,
				info.Name,
				fmt.Sprintf("从 Pod '%s' 复制文件 '%s' 到本地", input.Pod, input.FilePath),
				fromPod.ID,
			)
			h.handleDownload(w, req, fromPod.ID)
			return
		}
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
	})
}

type CopyFromPodRequest struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
	FilePath  string `json:"filepath"`
}

func (h *httpHandlerImpl) Shutdown(ctx context.Context) error {
	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()
	err := h.WsHttpServer.Shutdown(ctx)
	if err != nil {
		h.logger.Warning("shutdown ws server error: ", err.Error())
	}
	return err
}

func (h *httpHandlerImpl) authenticated(r *http.Request) (*http.Request, bool) {
	verifyToken, _ := h.authRepo.VerifyToken(r.Context(), r.Header.Get("Authorization"))
	if verifyToken != nil {
		return r.WithContext(auth.SetUser(r.Context(), verifyToken)), true
	}

	return nil, false
}

func (h *httpHandlerImpl) handleDownload(w http.ResponseWriter, r *http.Request, fid int) {
	fil, err := h.fileRepo.GetByID(r.Context(), fid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fileName := filepath.Base(fil.Path)

	user := auth.MustGetUser(r.Context())
	h.eventRepo.FileAuditLog(
		types.EventActionType_Download,
		user.Name,
		fmt.Sprintf("下载文件 '%s', 大小 %s",
			fil.Path,
			fil.HumanizeSize,
		),
		fil.ID,
	)
	read, err := h.uploader.Read(fil.Path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Error reading file: ", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer read.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape(fileName)))
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Access-Control-Expose-Headers", "*")

	// 调用 Write 之后就会写入 200 code
	if _, err := io.Copy(w, bufio.NewReaderSize(read, 1024*1024*2)); err != nil {
		h.logger.Error("Error writing file to response: ", err)
	}
}

func (h *httpHandlerImpl) handleBinaryFileUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(int64(h.fileRepo.MaxUploadSize())); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %s", err.Error()), http.StatusBadRequest)
		return
	}

	f, fh, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get file 'attachment': %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer f.Close()

	// 确保上传的文件名不会导致路径遍历攻击
	filename := filepath.Base(fh.Filename)

	info := auth.MustGetUser(r.Context())

	var uploader uploader.Uploader = h.uploader
	// 某个用户/那天/时间/文件名称
	put, err := uploader.Disk("users").Put(
		fmt.Sprintf("%s/%s/%s/%s",
			info.Name,
			h.timer.Now().Format("2006-01-02"),
			fmt.Sprintf("%s-%s", h.timer.Now().Format("15-04-05"), rand.String(20)),
			filename), f)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to upload file %s", err.Error()), http.StatusInternalServerError)
		return
	}

	createdFile, err := h.fileRepo.Create(r.Context(), &repo.CreateFileInput{
		Path:       put.Path(),
		Username:   info.Name,
		Size:       put.Size(),
		UploadType: uploader.Type(),
	})
	if err != nil {
		h.logger.Error("Error saving file metadata: ", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.eventRepo.FileAuditLog(
		types.EventActionType_Upload,
		info.Name,
		fmt.Sprintf("上传文件 '%s', 大小 %s", createdFile.Path, createdFile.HumanizeSize),
		createdFile.ID,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	var res = struct {
		ID int `json:"id"`
	}{
		ID: createdFile.ID,
	}
	marshal, err := json.Marshal(&res)
	if err != nil {
		h.logger.Error("Error marshaling response: ", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Write(marshal)
}
