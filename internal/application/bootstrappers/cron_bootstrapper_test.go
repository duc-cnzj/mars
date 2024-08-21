package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCronBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := application.NewMockApp(m)
	app.EXPECT().AddServer(gomock.Any())
	app.EXPECT().CronManager()
	a := &CronBootstrapper{}
	assert.Nil(t, a.Bootstrap(app))
}

func TestCronBootstrapper_Tags(t *testing.T) {
	a := &CronBootstrapper{}
	got := a.Tags()
	want := []string{"cron"}
	assert.Equal(t, got, want)
}
