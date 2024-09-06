package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/uploader"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_Authenticated(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	logger := mlog.NewMockLogger(m)
	authRepo := repo.NewMockAuthRepo(m)
	h := &httpHandlerImpl{
		logger:   logger,
		authRepo: authRepo,
	}

	// Test case: valid token
	authRepo.EXPECT().VerifyToken(gomock.Any(), "valid-token").Return(&auth.UserInfo{Name: "test-user"}, nil).Times(1)
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "valid-token")
	newReq, ok := h.authenticated(req)
	assert.True(t, ok)
	assert.Equal(t, "test-user", auth.MustGetUser(newReq.Context()).Name)

	// Test case: invalid token
	authRepo.EXPECT().VerifyToken(gomock.Any(), "invalid-token").Return(nil, errors.New("x")).Times(1)
	req2, _ := http.NewRequest("GET", "/", nil)
	req2.Header.Set("Authorization", "invalid-token")
	_, ok = h.authenticated(req2)
	assert.False(t, ok)
}

func Test_handleBinaryFileUpload(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	req := &http.Request{
		Form: map[string][]string{},
	}
	up := uploader.NewMockUploader(m)
	fileRepo := repo.NewMockFileRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	logger := mlog.NewMockLogger(m)
	authRepo := repo.NewMockAuthRepo(m)
	h := &httpHandlerImpl{
		fileRepo:  fileRepo,
		logger:    logger,
		uploader:  up,
		eventRepo: eventRepo,
		authRepo:  authRepo,
		timer:     timer.NewRealTimer(),
	}

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
	req2 = req2.WithContext(auth.SetUser(req2.Context(), &auth.UserInfo{Name: "duc"}))
	rr2 := httptest.NewRecorder()

	fileRepo.EXPECT().Create(gomock.Any(), &repo.CreateFileInput{
		Path:       "/app.txt",
		Username:   "duc",
		Size:       1000,
		UploadType: schematype.Local,
	}).Return(&repo.File{
		ID:       1,
		Path:     "/app.txt",
		Size:     1000,
		Username: "duc",
	}, nil)

	eventRepo.EXPECT().FileAuditLog(
		types.EventActionType_Upload,
		"duc",
		gomock.Any(),
		1,
	)
	up.EXPECT().Type().Return(schematype.Local)
	up.EXPECT().Disk("users").Return(up)
	finfo := uploader.NewMockFileInfo(m)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(finfo, nil)
	finfo.EXPECT().Path().Return("/app.txt")
	finfo.EXPECT().Size().Return(uint64(1000))
	h.handleBinaryFileUpload(rr2, req2)
	assert.Equal(t, 201, rr2.Code)
	assert.Equal(t, "application/json", rr2.Header().Get("Content-Type"))
}

func Test_httpHandlerImpl_handleDownload(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	logger := mlog.NewMockLogger(m)
	authRepo := repo.NewMockAuthRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	mockUploader := uploader.NewMockUploader(m)
	h := &httpHandlerImpl{
		logger:    logger,
		eventRepo: eventRepo,
		authRepo:  authRepo,
		fileRepo:  fileRepo,
		uploader:  mockUploader,
	}

	req := &http.Request{
		Form: map[string][]string{},
	}
	req.Form = make(url.Values)
	rr := httptest.NewRecorder()
	req = req.WithContext(auth.SetUser(req.Context(), &auth.UserInfo{Name: "duc"}))

	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&repo.File{
		ID:   1,
		Path: "/aaa/b.txt",
	}, nil)
	eventRepo.EXPECT().FileAuditLog(
		types.EventActionType_Download,
		"duc",
		gomock.Any(),
		1,
	)
	mockUploader.EXPECT().Read("/aaa/b.txt").Return(io.NopCloser(strings.NewReader("aaa")), nil)

	h.handleDownload(rr, req, 1)

	assert.Equal(t, "application/octet-stream", rr.Header().Get("Content-Type"))
	assert.Equal(t, fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape("b.txt")), rr.Header().Get("Content-Disposition"))
	assert.Equal(t, "0", rr.Header().Get("Expires"))
	assert.Equal(t, "binary", rr.Header().Get("Content-Transfer-Encoding"))
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "aaa", rr.Body.String())
	assert.Equal(t, http.StatusOK, rr.Code)
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

func TestNewHttpHandler(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	httpHandler := application.NewMockHttpHandler(m)
	handler := NewHttpHandler(
		httpHandler,
		mlog.NewForConfig(nil),
		uploader.NewMockUploader(m),
		repo.NewMockAuthRepo(m),
		repo.NewMockEventRepo(m),
		repo.NewMockFileRepo(m),
		timer.NewRealTimer(),
		repo.NewMockK8sRepo(m),
	).(*httpHandlerImpl)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.logger)
	assert.NotNil(t, handler.uploader)
	assert.NotNil(t, handler.authRepo)
	assert.NotNil(t, handler.eventRepo)
	assert.NotNil(t, handler.fileRepo)
	assert.NotNil(t, handler.timer)
	assert.NotNil(t, handler.k8sRepo)

	router := mux.NewRouter()
	handler.RegisterWsRoute(router)
	handler.RegisterSwaggerUIRoute(router)
	handler.RegisterFileRoute(runtime.NewServeMux())

	httpHandler.EXPECT().Shutdown(gomock.Not(nil)).Times(1)
	handler.Shutdown(context.TODO())
}
