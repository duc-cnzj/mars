package bootstrappers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestApiGatewayBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().Config().Return(&config.Config{GrpcPort: "50000"})
	app.EXPECT().AddServer(&apiGateway{endpoint: fmt.Sprintf("localhost:%s", "50000")}).Times(1)
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).Times(1)
	assert.Nil(t, (&ApiGatewayBootstrapper{}).Bootstrap(app))
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

func Test_apiGateway_Run(t *testing.T) {}

func Test_apiGateway_Shutdown(t *testing.T) {}

func Test_authenticated(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	auth := mock.NewMockAuthInterface(m)
	auth.EXPECT().VerifyToken("aaa").Return(nil, false)
	c := &contracts.JwtClaims{
		UserInfo: contracts.UserInfo{OpenIDClaims: contracts.OpenIDClaims{Name: "duc"}},
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
	value := ctx.Context().Value(authCtx{})
	assert.Equal(t, "duc", value.(*contracts.UserInfo).Name)
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
	req2 = req2.WithContext(context.WithValue(context.TODO(), authCtx{}, &contracts.UserInfo{OpenIDClaims: contracts.OpenIDClaims{Name: "duc"}}))
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

func Test_handleDownload(t *testing.T) {}

func Test_handleDownloadConfig(t *testing.T) {}

func Test_serveWs(t *testing.T) {}

func TestApiGatewayBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"api", "gateway"}, (&ApiGatewayBootstrapper{}).Tags())
}
