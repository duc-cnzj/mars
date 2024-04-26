package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	mevent "github.com/duc-cnzj/mars/v4/internal/event/events"
	"github.com/duc-cnzj/mars/v4/internal/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestEventBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	mevent.Register("xx", func(a any, event contracts.Event) error {
		return nil
	})
	app := mock.NewMockApplicationInterface(controller)
	d := mock.NewMockDispatcherInterface(controller)
	app.EXPECT().EventDispatcher().Return(d).AnyTimes()
	assert.Greater(t, len(mevent.RegisteredEvents()), 0)
	d.EXPECT().Listen(gomock.Any(), gomock.Any()).Times(len(mevent.RegisteredEvents()))
	(&EventBootstrapper{}).Bootstrap(app)
}

func TestEventBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&EventBootstrapper{}).Tags())
}
