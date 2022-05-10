package middlewares

import (
	"net/http"
	"testing"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestRouteLogger(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	l := mock.NewMockLoggerInterface(controller)
	mlog.SetLogger(l)
	m := &mockHandler{}
	rw := &mockResponseWriter{h: map[string][]string{}}
	l.EXPECT().Debugf("[Http]: method: %v, url: %v, use %v", gomock.Any()).Times(1)
	RouteLogger(m).ServeHTTP(rw, &http.Request{})
}
