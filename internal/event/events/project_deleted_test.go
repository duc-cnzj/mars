package events

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"

	"go.uber.org/mock/gomock"
)

func TestHandleProjectDeleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	pubsub := mock.NewMockPubSub(ctrl)
	sender := testutil.MockWsServer(ctrl, app)
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
