package bootstrappers

import (
	"errors"
	"sort"
	"testing"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/sirupsen/logrus"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestTracingBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := testutil.MockApp(controller)
	app.EXPECT().Config().Return(&config.Config{JaegerAgentHostPort: "xxxxxxxxxx"})
	app.EXPECT().IsDebug().Return(false)
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).Times(1)
	app.EXPECT().SetTracer(gomock.Any()).Times(1)
	assert.Nil(t, (&TracingBootstrapper{}).Bootstrap(app))
}

func Test_newResource(t *testing.T) {
	resource := newResource()
	var keys []string
	for _, a := range resource.Attributes() {
		keys = append(keys, string(a.Key))
	}
	sort.Strings(keys)
	assert.Equal(t, []string{
		"service.name",
		"service.version",
		"system.build_date",
		"system.git_branch",
		"system.git_commit",
		"system.go_version",
		"system.platform",
	}, keys)
}

func Test_newJaegerExporter(t *testing.T) {
	_, err := newJaegerExporter("", "", "")
	assert.Nil(t, err)
}

func TestErrorHandler_Handle(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())

	e := errors.New("xxx")
	l.EXPECT().Warning(e)
	eh := &errorHandler{}
	eh.Handle(e)
}
