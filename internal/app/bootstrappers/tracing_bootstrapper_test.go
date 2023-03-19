package bootstrappers

import (
	"errors"
	"sort"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type trueExtractValueMatcher struct {
	v any
}

func (t *trueExtractValueMatcher) Matches(x any) bool {
	t.v = x
	return true
}

func (t *trueExtractValueMatcher) String() string {
	return ""
}

func TestTracingBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())

	controller := gomock.NewController(t)
	defer controller.Finish()
	app := testutil.MockApp(controller)
	app.EXPECT().Config().Return(&config.Config{JaegerAgentHostPort: "127.0.0.1:6831"})
	app.EXPECT().IsDebug().Return(false)
	tm := &trueExtractValueMatcher{}
	app.EXPECT().RegisterAfterShutdownFunc(tm).Times(1)
	app.EXPECT().SetTracer(gomock.Any()).Times(1)
	assert.Nil(t, (&TracingBootstrapper{}).Bootstrap(app))

	l.EXPECT().Info("shutdown tracer").Times(1)
	tm.v.(contracts.Callback)(app)
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
	_, err := newJaegerExporter("", "")
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

func TestTracingBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"trace"}, (&TracingBootstrapper{}).Tags())
}
