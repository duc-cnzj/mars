package middlewares

import (
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestRouteLogger(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	l := mock.NewMockLoggerInterface(controller)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	m := &mockHandler{}
	rw := &mockResponseWriter{h: map[string][]string{}}
	l.EXPECT().Debugf("[Http]: method: %v, url: %v, use %v", gomock.Any()).Times(1)
	RouteLogger(m).ServeHTTP(rw, &http.Request{})
}
