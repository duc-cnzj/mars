package bootstrappers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	auth2 "github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/dustin/go-humanize"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
)

type apiGWMatcher struct {
	gw *apiGateway
}

func (a *apiGWMatcher) Matches(x any) bool {
	gateway, ok := x.(*apiGateway)
	if !ok {
		return false
	}
	if gateway.newServerFunc == nil {
		return false
	}
	return gateway.endpoint == a.gw.endpoint
}

func (a *apiGWMatcher) String() string {
	return ""
}

func TestApiGatewayBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().Config().Return(&config.Config{GrpcPort: "50000"})
	app.EXPECT().AddServer(&apiGWMatcher{gw: &apiGateway{endpoint: fmt.Sprintf("localhost:%s", "50000")}}).Times(1)
	ex := &testutil.ValueMatcher{}
	app.EXPECT().RegisterAfterShutdownFunc(ex).Times(1)
	assert.Nil(t, (&ApiGatewayBootstrapper{}).Bootstrap(app))
	assert.NotPanics(t, func() {
		fn, ok := ex.Value.(contracts.Callback)
		assert.True(t, ok)
		fn(app)
	})
}

func TestLoadSwaggerUI(t *testing.T) {
	r := mux.NewRouter()
	LoadSwaggerUI(r)

	req, err := http.NewRequest("GET", "/doc/swagger.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}

func Test_apiGateway_Run(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	port, _ := config.GetFreePort()
	app.EXPECT().Config().Return(&config.Config{AppPort: fmt.Sprintf("%v", port)}).AnyTimes()
	s := &mockHttpServer{}
	s.wg.Add(1)
	assert.Nil(t, (&apiGateway{newServerFunc: func(ctx context.Context, a *apiGateway) (httpServer, error) {
		return s, nil
	}}).Run(context.TODO()))
	s.wg.Wait()
}

type mockHttpServer struct {
	wg sync.WaitGroup
}

func (m *mockHttpServer) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockHttpServer) ListenAndServe() error {
	defer m.wg.Done()
	return nil
}

func Test_apiGateway_Shutdown(t *testing.T) {
	err := (&apiGateway{server: &mockHttpServer{}}).Shutdown(context.TODO())
	assert.Nil(t, err)
	err = (&apiGateway{}).Shutdown(context.TODO())
	assert.Nil(t, err)
}

func Test_authenticated(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	auth := mock.NewMockAuthInterface(m)
	auth.EXPECT().VerifyToken("aaa").Return(nil, false)
	c := &contracts.JwtClaims{
		UserInfo: contracts.UserInfo{Name: "duc"},
	}
	auth.EXPECT().VerifyToken("aaa").Return(c, true)
	app.EXPECT().Auth().Return(auth).AnyTimes()

	r := &http.Request{
		Header: map[string][]string{"Authorization": {"aaa"}},
	}
	_, b := authenticated(r)
	assert.False(t, b)
	ctx, b := authenticated(r)
	assert.True(t, b)
	assert.Equal(t, "duc", auth2.MustGetUser(ctx.Context()).Name)
}

type closedReader struct{}

func (c closedReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("xxx")
}

func Test_download(t *testing.T) {
	recorder := httptest.NewRecorder()
	download(recorder, "f.txt", strings.NewReader("aaa"))
	assert.Equal(t, "application/octet-stream", recorder.Header().Get("Content-Type"))
	assert.Equal(t, fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape("f.txt")), recorder.Header().Get("Content-Disposition"))
	assert.Equal(t, "0", recorder.Header().Get("Expires"))
	assert.Equal(t, "binary", recorder.Header().Get("Content-Transfer-Encoding"))
	assert.Equal(t, "*", recorder.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "aaa", recorder.Body.String())
	assert.Equal(t, http.StatusOK, recorder.Code)

	recorder2 := httptest.NewRecorder()
	download(recorder2, "f.txt", closedReader{})
	assert.Equal(t, http.StatusOK, recorder2.Code)
}

func Test_handFile(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	auth := mock.NewMockAuthInterface(m)
	auth.EXPECT().VerifyToken(gomock.Any()).Return(nil, false).Times(2)
	app.EXPECT().Auth().Return(auth).AnyTimes()
	r := runtime.NewServeMux()
	handFile(r)

	req, err := http.NewRequest("POST", "/api/files", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, 401, rr.Code)

	req2, _ := http.NewRequest("GET", "/api/download_file/1", nil)

	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	assert.Equal(t, 401, rr2.Code)
}

type filepathEqual struct {
	t   *testing.T
	reg *regexp.Regexp
}

func (f *filepathEqual) Matches(x any) bool {
	return f.reg.Match([]byte(x.(string)))
}

func (f *filepathEqual) String() string {
	return ""
}

func Test_handleBinaryFileUpload(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{UploadMaxSize: "100m"}).AnyTimes()
	req := &http.Request{
		Form: map[string][]string{},
	}

	rr := httptest.NewRecorder()
	handleBinaryFileUpload(rr, req)
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
	req2 = req2.WithContext(auth2.SetUser(req2.Context(), &contracts.UserInfo{Name: "duc"}))
	rr2 := httptest.NewRecorder()

	up := mock.NewMockUploader(m)
	up.EXPECT().Type().Return(contracts.Local)
	app.EXPECT().Uploader().Return(up)
	up.EXPECT().Disk("users").Return(up)
	finfo := mock.NewMockFileInfo(m)
	up.EXPECT().Put(&filepathEqual{t: t, reg: regexp.MustCompile(`duc/\d{4}-\d{2}-\d{2}/\d{2}-\d{2}-\d{2}-\w{20}/a.txt`)}, gomock.Any()).Return(finfo, nil)
	finfo.EXPECT().Path().Return("/app.txt")
	finfo.EXPECT().Size().Return(uint64(1000))
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	assert.Nil(t, db.AutoMigrate(&models.File{}))
	handleBinaryFileUpload(rr2, req2)
	assert.Equal(t, 201, rr2.Code)
	f := &models.File{}
	db.First(f)
	assert.Equal(t, "application/json", rr2.Header().Get("Content-Type"))
	assert.Equal(t, "/app.txt", f.Path)
	assert.Equal(t, uint64(1000), f.Size)
	assert.Equal(t, "duc", f.Username)
}

func Test_handleDownload(t *testing.T) {
	r := &http.Request{}
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	r1 := httptest.NewRecorder()
	handleDownload(r1, r, 1)
	assert.Equal(t, 500, r1.Code)

	db.AutoMigrate(&models.File{})
	r2 := httptest.NewRecorder()
	handleDownload(r2, r, 1)
	assert.Equal(t, 404, r2.Code)

	r3 := httptest.NewRecorder()
	aCtx := auth2.SetUser(r.Context(), &contracts.UserInfo{
		ID:    "1",
		Email: "admin@qq.com",
		Name:  "duc",
		Roles: []string{"admin"},
	})
	db.Create(&models.File{
		UploadType:    contracts.Local,
		Path:          "/tmp/a.txt",
		Size:          100,
		Username:      "duc",
		Namespace:     "xx",
		Pod:           "xx",
		Container:     "xx",
		ContainerPath: "/tmp/xx",
	})
	testutil.AssertAuditLogFired(m, app)
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up).AnyTimes()
	up.EXPECT().Type().Return(contracts.Local)
	up.EXPECT().Read("/tmp/a.txt").Return(io.NopCloser(strings.NewReader("xxx")), nil)
	handleDownload(r3, r.WithContext(aCtx), 1)
	assert.Equal(t, "xxx", r3.Body.String())
}
func Test_handleDownloadFileNotExists(t *testing.T) {
	r := &http.Request{}
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.File{})
	r3 := httptest.NewRecorder()
	aCtx := auth2.SetUser(r.Context(), &contracts.UserInfo{
		ID:    "1",
		Email: "admin@qq.com",
		Name:  "duc",
		Roles: []string{"admin"},
	})
	db.Create(&models.File{
		UploadType:    contracts.Local,
		Path:          "/tmp/a.txt",
		Size:          100,
		Username:      "duc",
		Namespace:     "xx",
		Pod:           "xx",
		Container:     "xx",
		ContainerPath: "/tmp/xx",
	})
	fired := testutil.AssertAuditLogFired(m, app)
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up).AnyTimes()
	up.EXPECT().Type().Return(contracts.Local).AnyTimes()
	up.EXPECT().Read("/tmp/a.txt").Return(nil, errors.New("xxx"))
	handleDownload(r3, r.WithContext(aCtx), 1)
	assert.Equal(t, 500, r3.Code)
	r4 := httptest.NewRecorder()
	up.EXPECT().Read("/tmp/a.txt").Return(nil, os.ErrNotExist)
	fired.EXPECT().Dispatch(contracts.Event("audit_log"), gomock.Any()).Times(1)
	handleDownload(r4, r.WithContext(aCtx), 1)
	assert.Equal(t, 404, r4.Code)
}

func Test_handleDownloadConfig(t *testing.T) {
	//handleDownloadConfig()
}

func TestApiGatewayBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"api", "gateway"}, (&ApiGatewayBootstrapper{}).Tags())
}

func Test_serveWs(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().BeforeServerRunHooks(gomock.Any()).Times(1)
	r := mux.NewRouter()
	serveWs(r)
	assert.NotNil(t, r.Get("ws"))
	assert.NotNil(t, r.Get("ws_info"))
	assert.Nil(t, r.Get("xxx"))
}

func TestHeaderMatcher(t *testing.T) {
	var tests = []struct {
		input string
		want  func() string
	}{
		{
			input: "a",
			want: func() string {
				matcher, _ := runtime.DefaultHeaderMatcher("a")
				return matcher
			},
		},
		{
			input: "tracestate",
			want: func() string {
				return "tracestate"
			},
		},
		{
			input: "traceparent",
			want: func() string {
				return "traceparent"
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.input, func(t *testing.T) {
			matcher, _ := HeaderMatcher(tt.input)
			assert.Equal(t, tt.want(), matcher)
		})
	}
}

func TestMaxRecvSize(t *testing.T) {
	assert.Equal(t, 20*1024*1024, MaxRecvMsgSize)
	bytes, _ := humanize.ParseBytes("20Mib")
	assert.Equal(t, int(bytes), MaxRecvMsgSize)
}

func Test_exportMarsConfig(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	auth := mock.NewMockAuthInterface(m)
	app.EXPECT().Auth().Return(auth).AnyTimes()

	r := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config/export", nil)
	auth.EXPECT().VerifyToken(gomock.Any()).Return(nil, false).Times(1)
	exportMarsConfig(r, req, map[string]string{})
	assert.Equal(t, 401, r.Code)
	c := &contracts.JwtClaims{
		UserInfo: contracts.UserInfo{Name: "duc", Roles: []string{}},
	}
	auth.EXPECT().VerifyToken(gomock.Any()).Return(c, true).Times(1)
	r2 := httptest.NewRecorder()
	exportMarsConfig(r2, req, map[string]string{})
	assert.Equal(t, 403, r2.Code)

	db, f := testutil.SetGormDB(m, app)
	defer f()

	db.AutoMigrate(&models.GitProject{})
	db.Create(&models.GitProject{
		DefaultBranch: "dev",
		Name:          "app",
		GitProjectId:  1,
		Enabled:       true,
		GlobalEnabled: false,
		GlobalConfig:  "xxx",
	})
	admin := &contracts.JwtClaims{
		UserInfo: contracts.UserInfo{Name: "duc", Roles: []string{"admin"}},
	}
	auth.EXPECT().VerifyToken(gomock.Any()).Return(admin, true).Times(1)
	r3 := httptest.NewRecorder()
	testutil.AssertAuditLogFired(m, app)
	exportMarsConfig(r3, req, map[string]string{})
	assert.Equal(t, 200, r3.Code)
	assert.Equal(t, dedent.Dedent(`
			[
				{
					"default_branch": "dev",
					"name": "app",
					"git_project_id": 1,
					"enabled": true,
					"global_enabled": false,
					"global_config": "xxx"
				}
			]
`), fmt.Sprintf("\n%s\n", r3.Body.String()))
}

func Test_exportMarsConfigWithPid(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	auth := mock.NewMockAuthInterface(m)
	app.EXPECT().Auth().Return(auth).AnyTimes()

	db, f := testutil.SetGormDB(m, app)
	defer f()

	db.AutoMigrate(&models.GitProject{})
	gp1 := &models.GitProject{
		DefaultBranch: "dev",
		Name:          "app",
		GitProjectId:  1,
		Enabled:       true,
		GlobalEnabled: false,
		GlobalConfig:  "xxx",
	}
	db.Create(gp1)
	gp2 := &models.GitProject{
		DefaultBranch: "dev2",
		Name:          "app2",
		GitProjectId:  2,
		Enabled:       true,
		GlobalEnabled: false,
		GlobalConfig:  "yyy",
	}
	db.Create(gp2)
	admin := &contracts.JwtClaims{
		UserInfo: contracts.UserInfo{Name: "duc", Roles: []string{"admin"}},
	}
	auth.EXPECT().VerifyToken(gomock.Any()).Return(admin, true).Times(1)
	r3 := httptest.NewRecorder()

	e := mock.NewMockDispatcherInterface(m)
	app.EXPECT().EventDispatcher().Return(e).AnyTimes()
	e.EXPECT().Dispatch(contracts.Event("audit_log"), gomock.Any()).Times(1)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/config/export/%v", gp2.GitProjectId), nil)
	exportMarsConfig(r3, req, map[string]string{"pid": fmt.Sprintf("%d", gp2.GitProjectId)})
	assert.Equal(t, 200, r3.Code)
	assert.Equal(t, dedent.Dedent(`
			[
				{
					"default_branch": "dev2",
					"name": "app2",
					"git_project_id": 2,
					"enabled": true,
					"global_enabled": false,
					"global_config": "yyy"
				}
			]
`), fmt.Sprintf("\n%s\n", r3.Body.String()))

	req2, _ := http.NewRequest("GET", fmt.Sprintf("/api/config/export/%v", gp1.GitProjectId), nil)
	r4 := httptest.NewRecorder()
	auth.EXPECT().VerifyToken(gomock.Any()).Return(admin, true).Times(1)
	e.EXPECT().Dispatch(contracts.Event("audit_log"), gomock.Any()).Times(1)
	exportMarsConfig(r4, req2, map[string]string{"pid": fmt.Sprintf("%d", gp1.GitProjectId)})
	assert.Equal(t, dedent.Dedent(`
			[
				{
					"default_branch": "dev",
					"name": "app",
					"git_project_id": 1,
					"enabled": true,
					"global_enabled": false,
					"global_config": "xxx"
				}
			]
`), fmt.Sprintf("\n%s\n", r4.Body.String()))

	req3, _ := http.NewRequest("GET", fmt.Sprintf("/api/config/export/%v", "999"), nil)
	r5 := httptest.NewRecorder()
	auth.EXPECT().VerifyToken(gomock.Any()).Return(admin, true).Times(1)
	e.EXPECT().Dispatch(contracts.Event("audit_log"), gomock.Any()).Times(1)
	exportMarsConfig(r5, req3, map[string]string{"pid": fmt.Sprintf("%v", "999")})
	assert.Equal(t, "[]", r5.Body.String())
}

func Test_importMarsConfig(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	auth := mock.NewMockAuthInterface(m)
	app.EXPECT().Auth().Return(auth).AnyTimes()

	r := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config/import", nil)
	auth.EXPECT().VerifyToken(gomock.Any()).Return(nil, false).Times(1)
	importMarsConfig(r, req, map[string]string{})
	assert.Equal(t, 401, r.Code)
	c := &contracts.JwtClaims{
		UserInfo: contracts.UserInfo{Name: "duc", Roles: []string{}},
	}
	auth.EXPECT().VerifyToken(gomock.Any()).Return(c, true).Times(1)
	r2 := httptest.NewRecorder()
	importMarsConfig(r2, req, map[string]string{})
	assert.Equal(t, 403, r2.Code)

	admin := &contracts.JwtClaims{
		UserInfo: contracts.UserInfo{Name: "duc", Roles: []string{"admin"}},
	}
	auth.EXPECT().VerifyToken(gomock.Any()).Return(admin, true).Times(1)
	r3 := httptest.NewRecorder()
	app.EXPECT().Config().Return(&config.Config{
		UploadMaxSize: "5Mib",
	})
	importMarsConfig(r3, req, map[string]string{})
	assert.Equal(t, 400, r3.Code)

	r4 := httptest.NewRecorder()
	req2 := newTestMultipartRequest(t)
	app.EXPECT().Config().Return(&config.Config{
		UploadMaxSize: "5Mib",
	})
	auth.EXPECT().VerifyToken(gomock.Any()).Return(admin, true).Times(1)
	testutil.AssertAuditLogFired(m, app)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.GitProject{})
	db.Create(&models.GitProject{
		GitProjectId: 2,
	})
	importMarsConfig(r4, req2, map[string]string{})
	assert.Equal(t, 204, r4.Code)
	var count int64
	db.Model(&models.GitProject{}).Count(&count)
	assert.Equal(t, int64(2), count)
	var gp models.GitProject
	db.Model(&models.GitProject{}).Where("`git_project_id` = ?", 2).First(&gp)
	assert.Equal(t, "app2", gp.Name)
	assert.Equal(t, "master", gp.DefaultBranch)
	assert.Equal(t, "xxx", gp.GlobalConfig)
	assert.Equal(t, true, gp.Enabled)
	assert.Equal(t, false, gp.GlobalEnabled)
}

func newTestMultipartRequest(t *testing.T) *http.Request {
	postData :=
		`
--xxx
Content-Disposition: form-data; name="file"; filename="file"
Content-Type: application/octet-stream
Content-Transfer-Encoding: binary

[
	{
		"default_branch": "dev",
		"name": "app",
		"git_project_id": 1,
		"enabled": true,
		"global_enabled": false,
		"global_config": "xxx"
	},
	{
		"default_branch": "master",
		"name": "app2",
		"git_project_id": 2,
		"enabled": true,
		"global_enabled": false,
		"global_config": "xxx"
	}
]
--xxx--
`
	req := &http.Request{
		Method: "POST",
		Header: http.Header{"Content-Type": {`multipart/form-data; boundary=xxx`}},
		Body:   io.NopCloser(strings.NewReader(postData)),
	}

	req.Form = make(url.Values)

	return req
}

func Test_initServer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().BeforeServerRunHooks(gomock.Any()).Times(1)
	app.EXPECT().Config().Return(&config.Config{}).Times(1)
	server, err := initServer(context.TODO(), &apiGateway{})
	assert.Nil(t, err)
	assert.NotNil(t, server)
}
