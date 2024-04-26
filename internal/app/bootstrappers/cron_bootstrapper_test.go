package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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
