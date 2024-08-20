package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/uploader"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewApiGateway(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := application.NewMockApp(m)

	app.EXPECT().Logger().Return(mlog.NewLogger(nil)).Times(1)
	app.EXPECT().Auth().Return(nil).Times(1)
	app.EXPECT().GrpcRegistry().Return(nil).Times(1)
	app.EXPECT().Config().Return(&config.Config{}).Times(2)
	app.EXPECT().WsServer().Return(nil).Times(1)
	app.EXPECT().Data().Return(nil)
	app.EXPECT().Uploader().Return(nil)
	app.EXPECT().Dispatcher().Return(nil)

	gw := NewApiGateway("test-endpoint", app)
	assert.NotNil(t, gw)
}

func Test_apiGateway_Run(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	server := NewMockHttpServer(m)
	logger := mlog.NewMockLogger(m)

	gw := &apiGateway{
		newServerFunc: func(ctx context.Context, a *apiGateway) (HttpServer, error) {
			return server, nil
		},
		logger: logger,
		port:   "111",
	}

	logger.EXPECT().Infof("[Server]: start apiGateway runner at :%s.", "111").Times(1)
	server.EXPECT().ListenAndServe().Return(assert.AnError).Times(1)
	logger.EXPECT().Error(gomock.Any()).Times(1)
	err := gw.Run(context.TODO())
	time.Sleep(1 * time.Second)
	assert.NoError(t, err)
}

func Test_apiGateway_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	server := NewMockHttpServer(m)
	logger := mlog.NewMockLogger(m)

	gw := &apiGateway{
		server: server,
		logger: logger,
	}
	logger.EXPECT().Info("[Server]: shutdown api-gateway runner.").Times(1)
	server.EXPECT().Shutdown(gomock.Any()).Return(nil).Times(1)
	assert.Nil(t, gw.Shutdown(context.TODO()))
}

func TestMiddlewareList_Wrap(t *testing.T) {
	logger := mlog.NewLogger(nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	middleware := func(logger mlog.Logger, handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			handler.ServeHTTP(w, r)
		})
	}

	middlewareList := middlewareList{middleware}

	wrappedHandler := middlewareList.Wrap(logger, handler)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, "middleware", rr.Header().Get("X-Test"))
	assert.Equal(t, "test", rr.Body.String())
}

func TestMiddlewareList_Wrap_Empty(t *testing.T) {
	logger := mlog.NewLogger(nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	middlewareList := middlewareList{}

	wrappedHandler := middlewareList.Wrap(logger, handler)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, "", rr.Header().Get("X-Test"))
	assert.Equal(t, "test", rr.Body.String())
}

func TestHandler_Authenticated(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authMock := auth.NewMockAuth(m)
	logger := mlog.NewMockLogger(m)

	h := &handler{
		logger: logger,
		auth:   authMock,
	}

	// Test case: valid token
	authMock.EXPECT().VerifyToken("valid-token").Return(&auth.JwtClaims{UserInfo: &auth.UserInfo{Name: "test-user"}}, true).Times(1)
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "valid-token")
	newReq, ok := h.authenticated(req)
	assert.True(t, ok)
	assert.Equal(t, "test-user", auth.MustGetUser(newReq.Context()).Name)

	// Test case: invalid token
	authMock.EXPECT().VerifyToken("invalid-token").Return(nil, false).Times(1)
	req2, _ := http.NewRequest("GET", "/", nil)
	req2.Header.Set("Authorization", "invalid-token")
	_, ok = h.authenticated(req2)
	assert.False(t, ok)
}

func TestHeaderMatcher(t *testing.T) {
	// Test case: tracestate key
	key, ok := headerMatcher("tracestate")
	assert.True(t, ok)
	assert.Equal(t, "tracestate", key)

	// Test case: traceparent key
	key, ok = headerMatcher("traceparent")
	assert.True(t, ok)
	assert.Equal(t, "traceparent", key)

	// Test case: other key
	key, ok = headerMatcher("other")
	assert.False(t, ok)
	assert.Equal(t, "", key)

	// Test case: empty key
	key, ok = headerMatcher("")
	assert.False(t, ok)
	assert.Equal(t, "", key)
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
	h := &handler{
		logger:   mlog.NewLogger(nil),
		auth:     auth.NewMockAuth(m),
		uploader: up,
		data:     data.NewDataImpl(&data.NewDataParams{DB: db}),
	}

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

	up.EXPECT().Type().Return(schematype.Local)
	up.EXPECT().Disk("users").Return(up)
	finfo := uploader.NewMockFileInfo(m)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(finfo, nil)
	finfo.EXPECT().Path().Return("/app.txt")
	finfo.EXPECT().Size().Return(uint64(1000))
	h.handleBinaryFileUpload(rr2, req2)
	assert.Equal(t, 201, rr2.Code)
	f, _ := db.File.Query().First(context.TODO())
	assert.Equal(t, "application/json", rr2.Header().Get("Content-Type"))
	assert.Equal(t, "/app.txt", f.Path)
	assert.Equal(t, uint64(1000), f.Size)
	assert.Equal(t, "duc", f.Username)
}

func Test_download(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	h := &handler{
		logger: mlog.NewLogger(nil),
	}

	recorder := httptest.NewRecorder()
	h.download(recorder, "f.txt", strings.NewReader("aaa"))
	assert.Equal(t, "application/octet-stream", recorder.Header().Get("Content-Type"))
	assert.Equal(t, fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape("f.txt")), recorder.Header().Get("Content-Disposition"))
	assert.Equal(t, "0", recorder.Header().Get("Expires"))
	assert.Equal(t, "binary", recorder.Header().Get("Content-Transfer-Encoding"))
	assert.Equal(t, "*", recorder.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "aaa", recorder.Body.String())
	assert.Equal(t, http.StatusOK, recorder.Code)
}
