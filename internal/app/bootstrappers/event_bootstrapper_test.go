package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	mevent "github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mock"

	"github.com/dustin/go-humanize"
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

func TestEventBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&EventBootstrapper{}).Tags())
}

func TestMaxRecvSize(t *testing.T) {
	assert.Equal(t, 20*1024*1024, MaxRecvMsgSize)
	bytes, _ := humanize.ParseBytes("20Mib")
	assert.Equal(t, int(bytes), MaxRecvMsgSize)
}
