package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	mevent "github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
