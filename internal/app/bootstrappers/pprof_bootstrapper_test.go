package bootstrappers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestPprofBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().AddServer(gomock.Any()).Times(1)
	(&PprofBootstrapper{}).Bootstrap(app)
}

func TestPprofRunner_Shutdown(t *testing.T) {
	s := &mockHttpServer{}
	assert.Nil(t, (&pprofRunner{server: s}).Shutdown(context.TODO()))
	assert.True(t, s.shutdownCalled)
}

func TestPprofBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"profile"}, (&PprofBootstrapper{}).Tags())
}

func Test_pprofMux(t *testing.T) {
	var tests = []struct {
		url  string
		code int
	}{
		{
			url:  "/debug/pprof/",
			code: 200,
		},
		{
			url:  "/debug/pprof/cmdline",
			code: 200,
		},
		{
			url:  "/debug/pprof/symbol",
			code: 200,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRecorder()
			request, _ := http.NewRequest("GET", tt.url, nil)
			pprofMux().ServeHTTP(r, request)
			assert.Equal(t, tt.code, r.Code)
		})
	}
}
