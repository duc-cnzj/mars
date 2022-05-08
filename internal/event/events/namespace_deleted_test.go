package events

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestHandleNamespaceDeleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	sender := mock.NewMockWsSender(ctrl)
	sender.EXPECT().Initialize(gomock.Any()).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).AnyTimes()
	pubsub := mock.NewMockPubSub(ctrl)
	app.EXPECT().GetPluginByName("test_wssender_driver").Return(sender)
	app.EXPECT().Config().Return(&config.Config{WsSenderPlugin: config.Plugin{
		Name: "test_wssender_driver",
		Args: nil,
	}})
	sender.EXPECT().New("", "").Return(pubsub)
	pubsub.EXPECT().ToAll(gomock.Any()).Times(1)
	HandleNamespaceDeleted(nil, EventNamespaceDeleted)
}
