package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/event"
	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func TestEventSvc_Authorize(t *testing.T) {
	e := new(EventSvc)
	ctx := context.TODO()
	ctx = auth.SetUser(ctx, &contracts.UserInfo{})
	_, err := e.Authorize(ctx, "")
	assert.ErrorIs(t, err, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error()))
	ctx = auth.SetUser(ctx, &contracts.UserInfo{
		Roles: []string{"admin"},
	})
	_, err = e.Authorize(ctx, "")
	assert.Nil(t, err)
}

func TestEventSvc_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	db, c := testutil.SetGormDB(ctrl, app)
	defer c()
	e := new(EventSvc)
	_, err := e.List(context.TODO(), &event.ListRequest{})
	assert.Error(t, err)
	db.AutoMigrate(&models.Event{}, &models.File{})
	f := seedEvents(db)
	list, _ := e.List(context.TODO(), &event.ListRequest{Page: 1, PageSize: 1})
	assert.Len(t, list.Items, 1)
	assert.Equal(t, int64(3), list.Count)
	list, _ = e.List(context.TODO(), &event.ListRequest{
		Page:       1,
		PageSize:   2,
		ActionType: types.EventActionType_Upload,
	})
	assert.Len(t, list.Items, 1)
	assert.Equal(t, int64(1), list.Count)
	list, _ = e.List(context.TODO(), &event.ListRequest{
		Page:       1,
		PageSize:   2,
		ActionType: types.EventActionType_Delete,
	})
	assert.Equal(t, int64(f.ID), list.Items[0].FileId)
	assert.Equal(t, int64(f.ID), list.Items[0].File.Id)
	db.Delete(&f)
	list, _ = e.List(context.TODO(), &event.ListRequest{
		Page:       1,
		PageSize:   2,
		ActionType: types.EventActionType_Delete,
	})
	assert.Equal(t, int64(f.ID), list.Items[0].FileId)
	assert.Nil(t, list.Items[0].File)

	list, _ = e.List(context.TODO(), &event.ListRequest{
		Page:     1,
		PageSize: 100,
		Search:   "by duc",
	})
	assert.Len(t, list.Items, 1)
	assert.Equal(t, list.Items[0].Message, "message by duc at now")
	list, _ = e.List(context.TODO(), &event.ListRequest{
		Page:     1,
		PageSize: 100,
		Search:   "Hi.Du",
	})
	assert.Len(t, list.Items, 1)
	assert.Equal(t, list.Items[0].Username, "Hi.Duc")
}

func seedEvents(db *gorm.DB) *models.File {
	f := &models.File{
		Path:          "a.txt",
		Size:          100,
		Username:      "duc",
		Namespace:     "ns",
		Pod:           "pod",
		Container:     "c",
		ContainerPath: "/cpath",
	}
	db.Create(f)
	var testData = []models.Event{
		{
			Action:   uint8(types.EventActionType_Shell),
			Username: "Hi.Duc",
			Message:  "message by duc at now",
			Old:      "old",
			New:      "new",
			Duration: "10s",
		},
		{
			Action:   uint8(types.EventActionType_Upload),
			Username: "duc1",
			Message:  "message1",
			Old:      "old1",
			New:      "new1",
			Duration: "101s",
		},
		{
			Action:   uint8(types.EventActionType_Delete),
			Username: "duc1",
			Message:  "message1",
			Old:      "old1",
			New:      "new1",
			Duration: "101s",
			FileID:   &f.ID,
		},
	}
	for _, datum := range testData {
		db.Create(&datum)
	}
	return f
}
