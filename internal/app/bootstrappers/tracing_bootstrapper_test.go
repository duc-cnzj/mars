package bootstrappers

import (
	"sort"
	"testing"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestTracingBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := testutil.MockApp(controller)
	app.EXPECT().Config().Return(&config.Config{JaegerAgentHostPort: ""})
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).Times(0)
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
