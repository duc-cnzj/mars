package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestK8sBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := application.NewMockApp(m)
	app.EXPECT().Config().Return(&config.Config{KubeConfig: "xxx"})
	mockData := data.NewMockData(m)
	app.EXPECT().Data().Return(mockData)
	app.EXPECT().Done()
	mockData.EXPECT().InitK8s(gomock.Any())
	assert.Nil(t, (&K8sBootstrapper{}).Bootstrap(app))

	app.EXPECT().Config().Return(&config.Config{KubeConfig: ""})
	assert.Nil(t, (&K8sBootstrapper{}).Bootstrap(app))
}

func TestK8sBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&K8sBootstrapper{}).Tags())
}
