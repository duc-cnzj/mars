package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cast"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/dustin/go-humanize"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

// fileHandler 负责文件相关 HTTP 路由（上传/下载/从 pod 拷贝）：独立持有文件边界
// 依赖（authBiz/uploader/fileBiz/eventBiz/timer/k8sBiz/containerBiz/accessBiz），与
// websocket/swagger 关注点解耦，避免把全部 HTTP 依赖缝进单一 httpHandlerImpl。
type fileHandler struct {
	logger       mlog.Logger
	authBiz      biz.AuthBiz
	uploader     uploader.Uploader
	fileBiz      biz.FileBiz
	eventBiz     biz.EventBiz
	timer        timer.Timer
	k8sBiz       biz.K8sBiz
	containerBiz biz.ContainerBiz
	accessBiz    biz.AccessBiz
}

// newFileHandler 从 HttpHandlerDeps 提取文件边界所需依赖，由 NewHttpHandler 内部调用。
func newFileHandler(deps HttpHandlerDeps) *fileHandler {
	return &fileHandler{
		logger:       deps.Logger.WithModule("services/fileHandler"),
		authBiz:      deps.AuthBiz,
		uploader:     deps.Uploader,
		fileBiz:      deps.FileBiz,
		eventBiz:     deps.EventBiz,
		timer:        deps.Timer,
		k8sBiz:       deps.K8sBiz,
		containerBiz: deps.ContainerBiz,
		accessBiz:    deps.AccessBiz,
	}
}

// RegisterFileRoute 注册文件相关的 HTTP 路由（非 gRPC-gateway 标准代理，手工 handler）：
// POST /api/files 上传、GET /api/download_file/{id} 下载、POST /api/copy_from_pod 从 pod 拷文件。
// 三条路由统一经 authHandler 套 LoginHTTP 中间件鉴权，handler 内部不再各自校验 token。
func (f *fileHandler) RegisterFileRoute(mux *runtime.ServeMux) {
	mux.HandlePath("POST", "/api/files", f.authHandler(func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		f.handleBinaryFileUpload(w, r)
	}))
	mux.HandlePath("GET", "/api/download_file/{id}", f.authHandler(f.httpDownload))
	mux.HandlePath("POST", "/api/copy_from_pod", f.authHandler(f.copyFromPod))
}

// authHandler 把 file handler 套上 LoginHTTP 鉴权中间件：校验 Authorization header
// 中的 token（经 biz.Authenticate 统一"校验+注入用户+应用用户表角色
// 接管"到 ctx）后进入真实 handler，失败统一 401。作为 HTTP 文件路由的唯一鉴权入口，
// 替代原先各 handler 手写的 authenticated() 副本——鉴权核心与 gRPC 拦截器共用
// biz.Authenticate，杜绝双实现漂移。
func (f *fileHandler) authHandler(handler func(w http.ResponseWriter, r *http.Request, pathParams map[string]string)) runtime.HandlerFunc {
	authMW := middlewares.LoginHTTP(func(ctx context.Context, token string) (context.Context, error) {
		return biz.Authenticate(ctx, f.authBiz, token)
	}, f.logger)
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		authMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler(w, r, pathParams)
		})).ServeHTTP(w, r)
	}
}

// httpDownload 处理 GET /api/download_file/{id}：先校验 id 合法性，再按
// biz.RequireFileAccess 判定所有者/admin 权限，落审计日志后把文件流写回响应。
// 下载本体（读存储 + 回写帧）下沉 handleDownload，本方法只做编排与授权判定。
// token 校验由路由层 authHandler 完成，ctx 已注入用户，这里直接 MustGetUser。
func (f *fileHandler) httpDownload(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	id := cast.ToInt(pathParams["id"])
	if id < 1 {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	fil, err := f.fileBiz.GetByID(ctx, id)
	if err != nil {
		f.logger.ErrorCtx(ctx, fmt.Sprintf("Error getting file id %d: ", id), err)
		toHttpError(w, err)
		return
	}
	user := biz.MustGetUser(ctx)
	// 文件可能含部署配置/执行记录等敏感内容，只允许所有者或 admin 下载，
	// 防止任意登录用户枚举文件 ID 拖库。判定收进 accessBiz.RequireFileAccess。
	if err := f.accessBiz.RequireFileAccess(ctx, fil); err != nil {
		f.logger.ErrorCtx(ctx, err)
		http.Error(w, "没有权限执行该操作", http.StatusForbidden)
		return
	}
	f.eventBiz.FileAuditLog(
		types.EventActionType_Download,
		user.Name,
		user.Email,
		fmt.Sprintf("下载文件 '%s', 大小 %s",
			fil.Path,
			fil.HumanizeSize,
		),
		fil.ID,
	)
	f.handleDownload(w, fil)
}

// copyFromPod 处理 POST /api/copy_from_pod：解析 body 后先做命名空间级访问控制，
// 再从 pod 拷贝文件到本地存储，落审计日志后复用 handleDownload 把文件流写回。
// 与 gRPC container.CopyToPod 的访问模型对齐，防止任意登录用户枚举 pod 路径拖敏感文件。
// token 校验由路由层 authHandler 完成，ctx 已注入用户，这里直接 MustGetUser。
func (f *fileHandler) copyFromPod(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	var input copyFromPodRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		f.logger.Debug("bad request body: ", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if input.Namespace == "" || input.Pod == "" || input.FilePath == "" {
		http.Error(w, "missing namespace, pod or filepath", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	info := biz.MustGetUser(ctx)
	// 与 gRPC container.CopyToPod 对齐：从 pod 拷文件必须先做命名空间级访问控制，
	// 否则任意登录用户可枚举 namespace/pod/container/filepath 把私有命名空间的
	// 容器文件（含 .env、密钥、配置）拷走。
	// 命名空间解析失败或无权访问：toHttpError 会把 ErrorPermissionDenied(403)
	// 映射为同样的 403 响应，与显式 http.Error 分支等价，无需区分两种错误。
	if _, nserr := f.accessBiz.RequireNamespaceAccessByName(ctx, input.Namespace); nserr != nil {
		f.logger.ErrorCtx(ctx, nserr)
		toHttpError(w, nserr)
		return
	}
	// 未指定容器时回落默认容器，与 gRPC CopyToPod/StreamCopyToPod/Exec 共用
	// biz.ResolveContainer 的"空则找默认"语义，闭合 HTTP 边界与 gRPC 的默认容器分歧。
	resolved, err := f.containerBiz.ResolveContainer(ctx, input.Namespace, input.Pod, input.Container)
	if err != nil {
		f.logger.ErrorCtx(ctx, err)
		toHttpError(w, err)
		return
	}
	input.Container = resolved
	fromPod, err := f.k8sBiz.CopyFromPod(ctx, &biz.CopyFromPodInput{
		Namespace: input.Namespace,
		Pod:       input.Pod,
		Container: input.Container,
		FilePath:  input.FilePath,
		UserName:  info.Name,
	})
	if err != nil {
		f.logger.Error("Error copying file from pod: ", err)
		toHttpError(w, err)
		return
	}
	f.eventBiz.FileAuditLog(
		types.EventActionType_Download,
		info.Name,
		info.Email,
		fmt.Sprintf("从 Pod '%s' 复制文件 '%s' 到本地", input.Pod, input.FilePath),
		fromPod.ID,
	)
	f.handleDownload(w, fromPod)
}

// toHttpError 把 gRPC status 错误映射为 HTTP 状态码写回：有 code 的走
// runtime.HTTPStatusFromCode（如 403/404），其余按 500 处理。
func toHttpError(w http.ResponseWriter, err error) {
	s, ok := status.FromError(err)
	if ok {
		http.Error(w, s.Message(), runtime.HTTPStatusFromCode(s.Code()))
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// copyFromPodRequest 是 POST /api/copy_from_pod 的 JSON 请求体。
type copyFromPodRequest struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
	FilePath  string `json:"filepath"`
}

// handleDownload 把文件对象从存储流回 HTTP 响应：设置下载相关响应头后按 2MB 缓冲
// io.Copy 传输。文件名用 filepath.Base 防目录穿越，非 ASCII 走 RFC 2231 编码。
func (f *fileHandler) handleDownload(w http.ResponseWriter, fil *biz.File) {
	fileName := filepath.Base(fil.Path)
	read, err := f.uploader.Read(fil.Path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		f.logger.Error("Error reading file: ", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer read.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	// mime.FormatMediaType 按 RFC 2231/5987 编码：ASCII 文件名（含空格）保留在
	// filename="..." 内；非 ASCII 走 filename*=utf-8''percent-encoded，浏览器才能
	// 正确显示中文/特殊字符文件名。url.QueryEscape 会把空格转成 +、中文直接乱码。
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": fileName}))
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Access-Control-Expose-Headers", "*")

	// 调用 Write 之后就会写入 200 code
	if _, err := io.Copy(w, bufio.NewReaderSize(read, 1024*1024*2)); err != nil {
		f.logger.Error("Error writing file to response: ", err)
	}
}

// handleBinaryFileUpload 处理 POST /api/files 的 multipart 上传：解析表单、净化
// 文件名/用户名防目录穿越、写入 users/ 存储盘、落元数据（失败回滚清理孤儿对象）、
// 落审计日志后返回新文件 ID。上传编排（含存储盘路径约定）保留在 HTTP 边界，因为
// 上传适配器与计时器本就是 transport 输出端口，与 container/grpc 的边界聚合一致。
func (f *fileHandler) handleBinaryFileUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(int64(f.fileBiz.MaxUploadSize())); err != nil {
		f.logger.Debug("failed to parse form: ", err)
		http.Error(w, fmt.Sprintf("failed to parse form: %s", err.Error()), http.StatusBadRequest)
		return
	}

	file, fh, err := r.FormFile("file")
	if err != nil {
		f.logger.Debug("failed to get file 'attachment': ", err)
		http.Error(w, fmt.Sprintf("failed to get file 'attachment': %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 确保上传的文件名不会导致路径遍历攻击
	filename := filepath.Base(fh.Filename)
	// 注意：Go 的 multipart 会把空 filename 的部分当普通表单值而非文件，故 fh.Filename 不会为空；
	// 但 "."、".."、"/" 这类纯路径哨兵仍会原样通过 Base 落到路径里（".." 会逃出 rand 子目录），必须拒绝。
	if filename == "" || filename == "." || filename == ".." || filename == "/" {
		http.Error(w, "invalid file name", http.StatusBadRequest)
		return
	}

	info := biz.MustGetUser(r.Context())

	// 用户名可能来自 OIDC 签发/管理员创建，未保证不含路径分隔符；
	// 与文件名一致用 filepath.Base 净化，避免目录穿越逃出 users/ 根目录。
	name := filepath.Base(info.Name)
	// 用户名缺失/纯路径哨兵同样拒绝：name 是第一段路径，若为 ".." 会直接逃出 users/ 根目录。
	if name == "" || name == "." || name == ".." || name == "/" {
		http.Error(w, "invalid username", http.StatusBadRequest)
		return
	}

	// 某个用户/那天/时间/文件名称
	put, err := f.uploader.Disk("users").Put(
		fmt.Sprintf("%s/%s/%s/%s",
			name,
			f.timer.Now().Format("2006-01-02"),
			fmt.Sprintf("%s-%s", f.timer.Now().Format("15-04-05"), rand.String(20)),
			filename), file)
	if err != nil {
		f.logger.Error("Error uploading file: ", err)
		http.Error(w, fmt.Sprintf("failed to upload file %s", err.Error()), http.StatusInternalServerError)
		return
	}

	createdFile, err := f.fileBiz.Create(r.Context(), &biz.CreateFileInput{
		Path:       put.Path(),
		Username:   info.Name,
		Size:       put.Size(),
		UploadType: f.uploader.Type(),
	})
	if err != nil {
		f.logger.Error("Error saving file metadata: ", err)
		// 元数据写入失败，清理已上传的对象，避免孤儿文件
		if derr := f.uploader.Delete(put.Path()); derr != nil {
			f.logger.Error("Error cleaning up orphaned upload: ", derr)
		}
		toHttpError(w, err)
		return
	}

	f.eventBiz.FileAuditLog(
		types.EventActionType_Upload,
		info.Name,
		info.Email,
		fmt.Sprintf("上传文件 '%s', 大小 %s", createdFile.Path, humanize.Bytes(createdFile.Size)),
		createdFile.ID,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// json.Marshal 对一个仅含 int 字段的固定结构体不可能失败，直接序列化。
	var res = struct {
		ID int `json:"id"`
	}{
		ID: createdFile.ID,
	}
	marshal, _ := json.Marshal(&res)
	if _, err := w.Write(marshal); err != nil {
		f.logger.Debug("write upload response: ", err)
	}
}
