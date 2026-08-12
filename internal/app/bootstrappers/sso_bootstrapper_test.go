package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSSOBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := app.NewMockApp(m)
	mockData := data.NewMockData(m)
	app.EXPECT().Data().Return(mockData)
	mockData.EXPECT().InitOidcProvider()

	assert.Nil(t, (&SSOBootstrapper{}).Bootstrap(app))
}

func TestSSOBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&SSOBootstrapper{}).Tags())
}
