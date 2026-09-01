package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_fileHandler_authHandler(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	authBiz := mocks.authBiz

	// 有效 token：VerifyToken 通过 → 注入用户并进入 handler。
	authBiz.EXPECT().VerifyToken(gomock.Any(), "valid-token").Return(&biz.UserInfo{Name: "test-user"}, nil).Times(1)
	// 有效 token 路径经 authBiz.EffectiveRoles 解析生效角色（对齐 gRPC 鉴权入口）。
	// mock UserInfo 未设 Email，EffectiveRoles 实收空串，matcher 用 Any 兜住。
	authBiz.EXPECT().EffectiveRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil).Times(1)
	var gotUser string
	handler := h.authHandler(func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		gotUser = biz.MustGetUser(r.Context()).Name
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "valid-token")
	handler(httptest.NewRecorder(), req, nil)
	assert.Equal(t, "test-user", gotUser)

	// 无效 token：VerifyToken 失败 → 中间件统一 401，handler 不进入。
	authBiz.EXPECT().VerifyToken(gomock.Any(), "bad-token").Return(nil, errors.New("invalid")).Times(1)
	called := false
	h2 := h.authHandler(func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		called = true
	})
	rec := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "bad-token")
	h2(rec, req2, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, called)
}

func Test_fileHandler_handleBinaryFileUpload(t *testing.T) {
	_, _ = data.NewSqliteDB()
	defer data.Close()
	req := &http.Request{
		Form: map[string][]string{},
	}
	h, mocks := newFileHandlerWithMocks(t)
	up := mocks.uploader
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo

	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000)).AnyTimes()
	rr := httptest.NewRecorder()
	h.handleBinaryFileUpload(rr, req)
	assert.Equal(t, 400, rr.Code)

	postData :=
		`value2
--xxx
Content-Disposition: form-data; name="file"; filename="a.txt"
Content-Type: application/octet-stream
Content-Transfer-Encoding: binary

binary data
--xxx--
		`
	req2 := &http.Request{
		Method: "POST",
		Header: http.Header{"Content-Type": {`multipart/form-data; boundary=xxx`}},
		Body:   io.NopCloser(strings.NewReader(postData)),
	}

	req2.Form = make(url.Values)
	req2 = req2.WithContext(biz.SetUser(req2.Context(), &biz.UserInfo{Name: "duc"}))
	rr2 := httptest.NewRecorder()

	fileRepo.EXPECT().Create(gomock.Any(), &biz.CreateFileInput{
		Path:       "/app.txt",
		Username:   "duc",
		Size:       1000,
		UploadType: schematype.Local,
	}).Return(&biz.File{
		ID:       1,
		Path:     "/app.txt",
		Size:     1000,
		Username: "duc",
	}, nil)

	// 上传审计日志的"大小"必须真实渲染，不能是空值（回归：曾用 createdFile.HumanizeSize，
	// 该字段在 ToFile 从未填充，导致 "大小 " 后无内容）。
	eventRepo.EXPECT().FileAuditLog(
		types.EventActionType_Upload,
		"duc",
		gomock.Any(),
		"上传文件 '/app.txt', 大小 1.0 kB",
		1,
	)
	up.EXPECT().Type().Return(schematype.Local)
	up.EXPECT().Disk("users").Return(up)
	finfo := uploader.NewMockFileInfo(mocks.ctrl)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(finfo, nil)
	finfo.EXPECT().Path().Return("/app.txt")
	finfo.EXPECT().Size().Return(uint64(1000))
	h.handleBinaryFileUpload(rr2, req2)
	assert.Equal(t, 201, rr2.Code)
	assert.Equal(t, "application/json", rr2.Header().Get("Content-Type"))
}

// 上传成功后 w.Write(marshal) 失败（客户端断连）：只打日志，不 panic。
func Test_fileHandler_handleBinaryFileUpload_WriteError(t *testing.T) {
	_, _ = data.NewSqliteDB()
	defer data.Close()
	postData :=
		`value2
--xxx
Content-Disposition: form-data; name="file"; filename="a.txt"
Content-Type: application/octet-stream
Content-Transfer-Encoding: binary

binary data
--xxx--
		`
	req := &http.Request{
		Method: "POST",
		Header: http.Header{"Content-Type": {`multipart/form-data; boundary=xxx`}},
		Body:   io.NopCloser(strings.NewReader(postData)),
	}
	req.Form = make(url.Values)
	req = req.WithContext(biz.SetUser(req.Context(), &biz.UserInfo{Name: "duc"}))

	h, mocks := newFileHandlerWithMocks(t)
	up := mocks.uploader
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo
	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000)).AnyTimes()
	fileRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&biz.File{
		ID: 1, Path: "/app.txt", Size: 1000, Username: "duc",
	}, nil)
	eventRepo.EXPECT().FileAuditLog(types.EventActionType_Upload, "duc", gomock.Any(), gomock.Any(), 1)
	up.EXPECT().Type().Return(schematype.Local)
	up.EXPECT().Disk("users").Return(up)
	finfo := uploader.NewMockFileInfo(mocks.ctrl)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(finfo, nil)
	finfo.EXPECT().Path().Return("/app.txt")
	finfo.EXPECT().Size().Return(uint64(1000))

	fw := &failWriter{}
	h.handleBinaryFileUpload(fw, req)
	assert.Equal(t, http.StatusCreated, fw.status)
}

// 用户名含路径分隔符时，存储路径第一段必须被净化（filepath.Base），
// 防止目录穿越逃出 users/ 根目录；审计记录的 Username 保留原始值。
func Test_fileHandler_handleBinaryFileUpload_PathTraversal(t *testing.T) {
	_, _ = data.NewSqliteDB()
	defer data.Close()
	postData :=
		`value2
--xxx
Content-Disposition: form-data; name="file"; filename="a.txt"
Content-Type: application/octet-stream
Content-Transfer-Encoding: binary

binary data
--xxx--
			`
	req := &http.Request{
		Method: "POST",
		Header: http.Header{"Content-Type": {`multipart/form-data; boundary=xxx`}},
		Body:   io.NopCloser(strings.NewReader(postData)),
	}
	req.Form = make(url.Values)
	req = req.WithContext(biz.SetUser(req.Context(), &biz.UserInfo{Name: "../../evil"}))

	h, mocks := newFileHandlerWithMocks(t)
	up := mocks.uploader
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo
	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000)).AnyTimes()
	fileRepo.EXPECT().Create(gomock.Any(), &biz.CreateFileInput{
		Path:       "/app.txt",
		Username:   "../../evil",
		Size:       1000,
		UploadType: schematype.Local,
	}).Return(&biz.File{ID: 1, Path: "/app.txt", Size: 1000, Username: "../../evil"}, nil)
	eventRepo.EXPECT().FileAuditLog(types.EventActionType_Upload, "../../evil", gomock.Any(), gomock.Any(), 1)
	up.EXPECT().Type().Return(schematype.Local)
	up.EXPECT().Disk("users").Return(up)
	finfo := uploader.NewMockFileInfo(mocks.ctrl)
	up.EXPECT().Put(gomock.Cond(func(x any) bool {
		p, ok := x.(string)
		if !ok {
			return false
		}
		// 第一段必须是净化后的 "evil"，绝不能是 ".."。
		return strings.HasPrefix(p, "evil/")
	}), gomock.Any()).Return(finfo, nil)
	finfo.EXPECT().Path().Return("/app.txt")
	finfo.EXPECT().Size().Return(uint64(1000))

	rr := httptest.NewRecorder()
	h.handleBinaryFileUpload(rr, req)
	assert.Equal(t, 201, rr.Code)
}

func Test_fileHandler_handleDownload(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	mockUploader := mocks.uploader

	rr := httptest.NewRecorder()

	mockUploader.EXPECT().Read("/aaa/b.txt").Return(io.NopCloser(strings.NewReader("aaa")), nil)

	h.handleDownload(rr, &biz.File{
		ID:   1,
		Path: "/aaa/b.txt",
	})

	assert.Equal(t, "application/octet-stream", rr.Header().Get("Content-Type"))
	assert.Equal(t, mime.FormatMediaType("attachment", map[string]string{"filename": "b.txt"}), rr.Header().Get("Content-Disposition"))
	assert.Equal(t, "0", rr.Header().Get("Expires"))
	assert.Equal(t, "binary", rr.Header().Get("Content-Transfer-Encoding"))
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "aaa", rr.Body.String())
	assert.Equal(t, http.StatusOK, rr.Code)
}

// 回归防护：Content-Disposition 必须走 mime.FormatMediaType（RFC 2231/5987），
// 而不是 url.QueryEscape —— 后者会把空格转成 +（浏览器显示字面 +）、非 ASCII 直接乱码。
func Test_fileHandler_handleDownload_ContentDisposition(t *testing.T) {
	for _, tt := range []struct {
		path string
		want string
	}{
		{path: "/aaa/foo bar.txt", want: `attachment; filename="foo bar.txt"`},
		{path: "/aaa/中文.txt", want: `attachment; filename*=utf-8''%E4%B8%AD%E6%96%87.txt`},
	} {
		h, mocks := newFileHandlerWithMocks(t)
		mockUploader := mocks.uploader
		mockUploader.EXPECT().Read(tt.path).Return(io.NopCloser(strings.NewReader("data")), nil)

		rr := httptest.NewRecorder()
		h.handleDownload(rr, &biz.File{Path: tt.path})

		assert.Equal(t, tt.want, rr.Header().Get("Content-Disposition"))
	}
}

func Test_toHttpError(t *testing.T) {
	recorder := httptest.NewRecorder()
	err := status.Error(codes.InvalidArgument, "invalid argument")
	toHttpError(recorder, err)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "invalid argument\n", recorder.Body.String())
}

func Test_toHttpError2(t *testing.T) {
	recorder := httptest.NewRecorder()
	err := errors.New("x")
	toHttpError(recorder, err)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "x\n", recorder.Body.String())
}

func Test_fileHandler_copyFromPod_Fail(t *testing.T) {
	h, _ := newFileHandlerWithMocks(t)
	w := httptest.NewRecorder()
	r := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}
	h.copyFromPod(w, r, map[string]string{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_fileHandler_copyFromPod_Fail2(t *testing.T) {
	h, _ := newFileHandlerWithMocks(t)
	w := httptest.NewRecorder()
	marshal, _ := json.Marshal(&copyFromPodRequest{
		Namespace: "",
		Pod:       "",
		Container: "",
		FilePath:  "",
	})
	r := &http.Request{
		Body: io.NopCloser(bytes.NewReader(marshal)),
	}
	h.copyFromPod(w, r, map[string]string{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_fileHandler_copyFromPod_Success(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo
	nsRepo := mocks.nsRepo
	mockUploader := mocks.uploader
	w := httptest.NewRecorder()
	marshal, _ := json.Marshal(&copyFromPodRequest{
		Namespace: "a",
		Pod:       "b",
		Container: "c",
		FilePath:  "d",
	})
	// token 校验已收敛到路由层 authHandler，此处直接以注入用户 ctx 的请求测 handler。
	r := (&http.Request{
		Body: io.NopCloser(bytes.NewReader(marshal)),
	}).WithContext(biz.SetUser(context.Background(), &biz.UserInfo{Name: "duc"}))
	mockUploader.EXPECT().Read("/aaa/b.txt").Return(io.NopCloser(strings.NewReader("aaa")), nil)

	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Name: "a"}, nil)
	k8sRepo.EXPECT().CopyFromPod(gomock.Any(), &biz.CopyFromPodInput{
		Namespace: "a",
		Pod:       "b",
		Container: "c",
		FilePath:  "d",
		UserName:  "duc",
	}).Return(&biz.File{ID: 1, Path: "/aaa/b.txt"}, nil)
	eventRepo.EXPECT().FileAuditLog(types.EventActionType_Download, "duc", gomock.Any(), gomock.Any(), 1)
	h.copyFromPod(w, r, map[string]string{})
	assert.Equal(t, http.StatusOK, w.Code)
}

// 回归防护：未指定 container 时回落默认容器（与 gRPC CopyToPod/StreamCopyToPod/Exec
// 共用 biz.ResolveContainer 的"空则找默认"语义），CopyFromPod 拿到的是解析后的容器名。
// 去掉 ResolveContainer 调用，本测试会因 FindDefaultContainer 未被调用而失败。
func Test_fileHandler_copyFromPod_DefaultContainer(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	k8sRepo := mocks.k8sRepo
	eventRepo := mocks.eventRepo
	nsRepo := mocks.nsRepo
	mockUploader := mocks.uploader
	w := httptest.NewRecorder()
	// Container 故意留空：应解析为默认容器 "c" 后再拷文件。
	marshal, _ := json.Marshal(&copyFromPodRequest{
		Namespace: "a",
		Pod:       "b",
		Container: "",
		FilePath:  "d",
	})
	r := (&http.Request{
		Body: io.NopCloser(bytes.NewReader(marshal)),
	}).WithContext(biz.SetUser(context.Background(), &biz.UserInfo{Name: "duc"}))
	mockUploader.EXPECT().Read("/aaa/b.txt").Return(io.NopCloser(strings.NewReader("aaa")), nil)

	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Name: "a"}, nil)
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("c", nil)
	k8sRepo.EXPECT().CopyFromPod(gomock.Any(), &biz.CopyFromPodInput{
		Namespace: "a",
		Pod:       "b",
		Container: "c",
		FilePath:  "d",
		UserName:  "duc",
	}).Return(&biz.File{ID: 1, Path: "/aaa/b.txt"}, nil)
	eventRepo.EXPECT().FileAuditLog(types.EventActionType_Download, "duc", gomock.Any(), gomock.Any(), 1)
	h.copyFromPod(w, r, map[string]string{})
	assert.Equal(t, http.StatusOK, w.Code)
}

// 回归防护：FindDefaultContainer 失败（如 pod 无默认容器）→ 404，不应继续走到
// CopyFromPod 或下载路径。
func Test_fileHandler_copyFromPod_ResolveContainerError(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	w := httptest.NewRecorder()
	marshal, _ := json.Marshal(&copyFromPodRequest{
		Namespace: "a",
		Pod:       "b",
		Container: "",
		FilePath:  "d",
	})
	r := (&http.Request{
		Body: io.NopCloser(bytes.NewReader(marshal)),
	}).WithContext(biz.SetUser(context.Background(), &biz.UserInfo{Name: "duc"}))
	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Name: "a"}, nil)
	k8sRepo.EXPECT().FindDefaultContainer(gomock.Any(), "a", "b").Return("", status.Error(codes.NotFound, "no default container"))
	h.copyFromPod(w, r, map[string]string{})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// 回归防护：POST /api/copy_from_pod 从 pod 拷文件必须先做命名空间级访问控制。
// 去掉 CopyFromPod 里的 CanAccess 检查，本测试必须失败。
func Test_fileHandler_copyFromPod_AccessDenied(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo
	eventRepo := mocks.eventRepo
	mockUploader := mocks.uploader

	// 当前用户 "other" 非命名空间成员，对私有空间无访问权限 → 403
	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Private: true, CreatorEmail: "user@mars.com"}, nil)
	// 若访问控制被删除，请求会继续走到 CopyFromPod + 下载路径 → 200，而非 403
	k8sRepo.EXPECT().CopyFromPod(gomock.Any(), gomock.Any()).Return(&biz.File{ID: 1, Path: "/aaa/b.txt"}, nil).AnyTimes()
	eventRepo.EXPECT().FileAuditLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockUploader.EXPECT().Read("/aaa/b.txt").Return(io.NopCloser(strings.NewReader("aaa")), nil).AnyTimes()

	marshal, _ := json.Marshal(&copyFromPodRequest{Namespace: "a", Pod: "b", Container: "c", FilePath: "d"})
	w := httptest.NewRecorder()
	r := (&http.Request{Body: io.NopCloser(bytes.NewReader(marshal))}).
		WithContext(biz.SetUser(context.Background(), &biz.UserInfo{Name: "other"}))
	h.copyFromPod(w, r, map[string]string{})
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_fileHandler_copyFromPod_NamespaceError(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(nil, errors.New("boom"))

	marshal, _ := json.Marshal(&copyFromPodRequest{Namespace: "a", Pod: "b", Container: "c", FilePath: "d"})
	w := httptest.NewRecorder()
	r := (&http.Request{Body: io.NopCloser(bytes.NewReader(marshal))}).
		WithContext(biz.SetUser(context.Background(), &biz.UserInfo{Name: "duc"}))
	h.copyFromPod(w, r, map[string]string{})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_fileHandler_httpDownload_Fail(t *testing.T) {
	h, _ := newFileHandlerWithMocks(t)
	w := httptest.NewRecorder()
	r := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}
	h.httpDownload(w, r, map[string]string{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_fileHandler_httpDownload_Ok(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	eventRepo := mocks.eventRepo
	fileRepo := mocks.fileRepo
	mockUploader := mocks.uploader
	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{ID: 1, Path: "/aaa/b.txt", Username: "duc"}, nil)
	eventRepo.EXPECT().FileAuditLog(types.EventActionType_Download, "duc", gomock.Any(), gomock.Any(), 1)
	mockUploader.EXPECT().Read("/aaa/b.txt").Return(io.NopCloser(strings.NewReader("aaa")), nil)

	w := httptest.NewRecorder()
	r := (&http.Request{}).WithContext(biz.SetUser(context.Background(), &biz.UserInfo{Name: "duc"}))
	h.httpDownload(w, r, map[string]string{
		"id": "1",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_fileHandler_httpDownload_GetByIDError(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(nil, errors.New("boom"))

	w := httptest.NewRecorder()
	r := (&http.Request{}).WithContext(biz.SetUser(context.Background(), &biz.UserInfo{Name: "duc"}))
	h.httpDownload(w, r, map[string]string{"id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// 回归防护：GET /api/download_file/{id} 文件可能含部署配置/执行记录等敏感内容，
// 只允许所有者或 admin 下载。去掉 httpDownload 里的所有权检查，本测试必须失败。
func Test_fileHandler_httpDownload_AccessDenied(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo
	mockUploader := mocks.uploader

	// 文件属于其他人，当前用户非 admin → 403
	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{ID: 1, Path: "/aaa/b.txt", Username: "someone-else"}, nil)
	h.eventBiz = biz.NewEventBiz(eventRepo)
	h.uploader = mockUploader
	// 若所有权检查被删除，会走到下载路径 → 200，而非 403
	eventRepo.EXPECT().FileAuditLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockUploader.EXPECT().Read("/aaa/b.txt").Return(io.NopCloser(strings.NewReader("aaa")), nil).AnyTimes()

	w := httptest.NewRecorder()
	r := (&http.Request{}).WithContext(biz.SetUser(context.Background(), &biz.UserInfo{Name: "duc"}))
	h.httpDownload(w, r, map[string]string{"id": "1"})
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_fileHandler_copyFromPod_K8sError(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	k8sRepo := mocks.k8sRepo
	nsRepo := mocks.nsRepo

	nsRepo.EXPECT().FindByName(gomock.Any(), "a").Return(&biz.Namespace{Name: "a"}, nil)
	k8sRepo.EXPECT().CopyFromPod(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

	marshal, _ := json.Marshal(&copyFromPodRequest{Namespace: "a", Pod: "b", Container: "c", FilePath: "d"})
	w := httptest.NewRecorder()
	r := (&http.Request{Body: io.NopCloser(bytes.NewReader(marshal))}).
		WithContext(biz.SetUser(context.Background(), &biz.UserInfo{Name: "duc"}))
	h.copyFromPod(w, r, map[string]string{})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_fileHandler_handleDownload_NotFound(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	up := mocks.uploader

	up.EXPECT().Read("/missing.txt").Return(nil, &os.PathError{Op: "open", Path: "/missing.txt", Err: os.ErrNotExist})

	w := httptest.NewRecorder()
	h.handleDownload(w, &biz.File{Path: "/missing.txt"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_fileHandler_handleDownload_ReadError(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	up := mocks.uploader

	up.EXPECT().Read("/x.txt").Return(nil, errors.New("boom"))

	w := httptest.NewRecorder()
	h.handleDownload(w, &biz.File{Path: "/x.txt"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read boom") }

func Test_fileHandler_handleDownload_CopyError(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	up := mocks.uploader

	up.EXPECT().Read("/x.txt").Return(io.NopCloser(errorReader{}), nil)

	w := httptest.NewRecorder()
	h.handleDownload(w, &biz.File{Path: "/x.txt"})
	// 响应头在 io.Copy 前已 flush，所以状态码保持 200
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_fileHandler_handleBinaryFileUpload_MissingFile(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000))

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.Close() // 表单里没有 file 字段

	req := httptest.NewRequest("POST", "/api/files", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = req.WithContext(biz.SetUser(req.Context(), &biz.UserInfo{Name: "duc"}))

	w := httptest.NewRecorder()
	h.handleBinaryFileUpload(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_fileHandler_handleBinaryFileUpload_TraversalFilename(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000))

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	// 路径哨兵文件名：filepath.Base(".") 返回 "."，若不拒绝会落到 users/./ 路径；
	// ".." 则会逃出 rand 子目录。注意空 filename 的部分会被 Go 当普通表单值而非文件，到不了这里。
	fw, _ := mw.CreateFormFile("file", ".")
	_, _ = fw.Write([]byte("data"))
	mw.Close()

	req := httptest.NewRequest("POST", "/api/files", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = req.WithContext(biz.SetUser(req.Context(), &biz.UserInfo{Name: "duc"}))

	w := httptest.NewRecorder()
	h.handleBinaryFileUpload(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_fileHandler_handleBinaryFileUpload_EmptyUsername(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000))

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "a.txt")
	_, _ = fw.Write([]byte("data"))
	mw.Close()

	req := httptest.NewRequest("POST", "/api/files", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = req.WithContext(biz.SetUser(req.Context(), &biz.UserInfo{Name: ""}))

	w := httptest.NewRecorder()
	h.handleBinaryFileUpload(w, req)
	// 用户名缺失 → 拒绝，防止 users/./ 桶根目录。
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_fileHandler_handleBinaryFileUpload_PutError(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	fileRepo := mocks.fileRepo
	up := mocks.uploader

	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000))

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "a.txt")
	_, _ = fw.Write([]byte("data"))
	mw.Close()

	req := httptest.NewRequest("POST", "/api/files", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = req.WithContext(biz.SetUser(req.Context(), &biz.UserInfo{Name: "duc"}))

	up.EXPECT().Disk("users").Return(up)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

	w := httptest.NewRecorder()
	h.handleBinaryFileUpload(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_fileHandler_handleBinaryFileUpload_CreateError(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	fileRepo := mocks.fileRepo
	up := mocks.uploader

	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000))

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "a.txt")
	_, _ = fw.Write([]byte("data"))
	mw.Close()

	req := httptest.NewRequest("POST", "/api/files", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = req.WithContext(biz.SetUser(req.Context(), &biz.UserInfo{Name: "duc"}))

	up.EXPECT().Disk("users").Return(up)
	finfo := uploader.NewMockFileInfo(mocks.ctrl)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(finfo, nil)
	finfo.EXPECT().Path().Return("/app.txt").AnyTimes()
	finfo.EXPECT().Size().Return(uint64(5))
	up.EXPECT().Type().Return(schematype.Local)
	fileRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))
	up.EXPECT().Delete("/app.txt").Return(nil)

	w := httptest.NewRecorder()
	h.handleBinaryFileUpload(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_fileHandler_handleBinaryFileUpload_CreateDeleteError(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	fileRepo := mocks.fileRepo
	up := mocks.uploader

	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000))

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "a.txt")
	_, _ = fw.Write([]byte("data"))
	mw.Close()

	req := httptest.NewRequest("POST", "/api/files", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = req.WithContext(biz.SetUser(req.Context(), &biz.UserInfo{Name: "duc"}))

	up.EXPECT().Disk("users").Return(up)
	finfo := uploader.NewMockFileInfo(mocks.ctrl)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(finfo, nil)
	finfo.EXPECT().Path().Return("/app.txt").AnyTimes()
	finfo.EXPECT().Size().Return(uint64(5))
	up.EXPECT().Type().Return(schematype.Local)
	fileRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))
	// 孤儿清理也失败 → 覆盖 if derr != nil 的 Error 日志分支
	up.EXPECT().Delete("/app.txt").Return(errors.New("del boom"))

	w := httptest.NewRecorder()
	h.handleBinaryFileUpload(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_fileHandler_RegisterFileRoute_Closures(t *testing.T) {
	h, mocks := newFileHandlerWithMocks(t)
	authBiz := mocks.authBiz
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo
	up := mocks.uploader

	runtimeMux := runtime.NewServeMux()
	h.RegisterFileRoute(runtimeMux)
	srv := httptest.NewServer(runtimeMux)
	defer srv.Close()

	// 未认证 POST /api/files -> 401
	authBiz.EXPECT().VerifyToken(gomock.Any(), "").Return(nil, errors.New("bad"))
	resp, err := http.Post(srv.URL+"/api/files", "application/octet-stream", strings.NewReader("x"))
	assert.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// 已认证 POST /api/files -> 201
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "a.txt")
	_, _ = fw.Write([]byte("data"))
	mw.Close()

	// 三条路由统一经 authHandler 鉴权：download/copy_from_pod 也先过 token 校验，
	// 故 "tok" 期望 AnyTimes（POST files / download / copy_from_pod 共 3 处）。
	authBiz.EXPECT().VerifyToken(gomock.Any(), "tok").Return(&biz.UserInfo{Name: "duc"}, nil).AnyTimes()
	// 有效 token 路径经 authBiz.EffectiveRoles 解析生效角色，同样 AnyTimes（3 处路由鉴权）。
	// mock UserInfo 未设 Email，EffectiveRoles 实收空串，email matcher 用 Any。
	authBiz.EXPECT().EffectiveRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil).AnyTimes()
	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000))
	up.EXPECT().Type().Return(schematype.Local)
	up.EXPECT().Disk("users").Return(up)
	finfo := uploader.NewMockFileInfo(mocks.ctrl)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(finfo, nil)
	finfo.EXPECT().Path().Return("/app.txt")
	finfo.EXPECT().Size().Return(uint64(4))
	fileRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&biz.File{ID: 1, Path: "/app.txt", Size: 4, Username: "duc"}, nil)
	eventRepo.EXPECT().FileAuditLog(types.EventActionType_Upload, "duc", gomock.Any(), gomock.Any(), 1)

	req2, _ := http.NewRequest("POST", srv.URL+"/api/files", &buf)
	req2.Header.Set("Content-Type", mw.FormDataContentType())
	req2.Header.Set("Authorization", "tok")
	resp2, err := http.DefaultClient.Do(req2)
	assert.NoError(t, err)
	resp2.Body.Close()
	assert.Equal(t, http.StatusCreated, resp2.StatusCode)

	// GET /api/download_file/0 缺 id -> 400（鉴权通过后进入 id 校验）
	req3, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/download_file/0", nil)
	req3.Header.Set("Authorization", "tok")
	resp3, err := http.DefaultClient.Do(req3)
	assert.NoError(t, err)
	resp3.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp3.StatusCode)

	// POST /api/copy_from_pod 非法 body -> 400（鉴权通过后进入 body 校验）
	req4, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/copy_from_pod", strings.NewReader(""))
	req4.Header.Set("Authorization", "tok")
	resp4, err := http.DefaultClient.Do(req4)
	assert.NoError(t, err)
	resp4.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp4.StatusCode)
}
