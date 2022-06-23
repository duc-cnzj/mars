package bootstrappers

import (
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
	assert.Nil(t, (&TracingBootstrapper{}).Bootstrap(app))
}
