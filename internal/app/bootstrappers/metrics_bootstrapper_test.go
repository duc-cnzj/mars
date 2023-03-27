package bootstrappers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/mock"
)

func TestMetricsBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().AddServer(gomock.Any()).Times(1)
	app.EXPECT().Config().Return(&config.Config{MetricsPort: "9091"}).Times(1)
	(&MetricsBootstrapper{}).Bootstrap(app)
}

func TestMetricsBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"metrics"}, (&MetricsBootstrapper{}).Tags())
}

func Test_metricsRunner_Shutdown(t *testing.T) {
	assert.Nil(t, (&metricsRunner{s: &http.Server{}}).Shutdown(context.TODO()))
}

func Test_metricsRunner_Run(t *testing.T) {
	port, _ := config.GetFreePort()
	mr := &metricsRunner{port: strconv.Itoa(port)}
	mr.Run(context.TODO())
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Nil(t, mr.Shutdown(context.TODO()))
	_, err = http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
	assert.Error(t, err)
}
