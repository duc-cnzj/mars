package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCronBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().CronManager().Times(1)
	app.EXPECT().AddServer(gomock.Any()).Times(1)
	(&CronBootstrapper{}).Bootstrap(app)
}

func TestCronBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"cron"}, (&CronBootstrapper{}).Tags())
}
