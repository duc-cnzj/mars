package events

import (
	"testing"

	"github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"

	"github.com/golang/mock/gomock"
)

func TestHandleProjectDeleted(t *testing.T) {
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
	pubsub.EXPECT().ToAll(&EventNamespaceDeletedMatcher{
		nsID: 1,
	}).Times(1)
	HandleProjectDeleted(&models.Project{Name: "app", NamespaceId: 1}, EventNamespaceDeleted)
}

type EventNamespaceDeletedMatcher struct {
	nsID int64
}

func (e *EventNamespaceDeletedMatcher) Matches(x any) bool {
	response, ok := x.(*websocket.WsReloadProjectsResponse)
	if !ok {
		return false
	}
	return response.NamespaceId == e.nsID
}

func (e *EventNamespaceDeletedMatcher) String() string {
	return ""
}
