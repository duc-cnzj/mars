package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/event"
	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

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
	list, _ = e.List(context.TODO(), &event.ListRequest{
		Page:       1,
		PageSize:   2,
		ActionType: types.EventActionType_Upload,
	})
	assert.Len(t, list.Items, 1)
	list, _ = e.List(context.TODO(), &event.ListRequest{
		Page:       1,
		PageSize:   2,
		ActionType: types.EventActionType_Delete,
	})
	assert.Equal(t, int64(f.ID), list.Items[0].FileId)
	assert.Equal(t, int64(f.ID), list.Items[0].File.Id)
	assert.Equal(t, int64(100), list.Items[0].File.Size)
	assert.True(t, list.Items[0].File.HumanizeSize != "")
	db.Delete(&f)
	list, _ = e.List(context.TODO(), &event.ListRequest{
		Page:       1,
		PageSize:   2,
		ActionType: types.EventActionType_Delete,
	})
	assert.Equal(t, int64(f.ID), list.Items[0].FileId)
	assert.Nil(t, list.Items[0].File)
	assert.True(t, list.Items[0].HasDiff)
	assert.Empty(t, list.Items[0].Old)
	assert.Empty(t, list.Items[0].New)
	assert.NotEmpty(t, list.Items[0].Id)
	assert.NotEmpty(t, list.Items[0].Action)
	assert.NotEmpty(t, list.Items[0].Username)
	assert.NotEmpty(t, list.Items[0].Message)
	assert.NotEmpty(t, list.Items[0].Duration)
	assert.NotEmpty(t, list.Items[0].CreatedAt)
	assert.NotEmpty(t, list.Items[0].UpdatedAt)
	assert.NotEmpty(t, list.Items[0].DeletedAt)

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

func TestEventSvc_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	app.EXPECT().IsDebug().Return(false).AnyTimes()
	_, err := new(EventSvc).Show(adminCtx(), &event.ShowRequest{
		Id: 1,
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, fromError.Code())
	assert.Nil(t, db.AutoMigrate(&models.Event{}, &models.File{}))
	_, err = new(EventSvc).Show(adminCtx(), &event.ShowRequest{
		Id: 1,
	})
	fromError, _ = status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())

	f := &models.File{
		UploadType: "local",
		Path:       "/app/a.txt",
		Size:       100,
		Username:   "duc",
	}
	db.Create(f)
	ev := &models.Event{
		Action:   1,
		Username: "duc",
		Message:  "aaa",
		Old:      "old",
		New:      "new",
		Duration: "10s",
		FileID:   &f.ID,
	}
	db.Create(ev)
	r, err := new(EventSvc).Show(adminCtx(), &event.ShowRequest{
		Id: int64(ev.ID),
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, r.Event.Old)
	assert.NotEmpty(t, r.Event.New)
	assert.NotEmpty(t, r.Event.Duration)
	assert.NotEmpty(t, r.Event.Action)
	assert.NotEmpty(t, r.Event.File)
	assert.NotEmpty(t, r.Event.Id)
	assert.NotEmpty(t, r.Event.Message)
	assert.NotEmpty(t, r.Event.Username)
	assert.NotEmpty(t, r.Event.EventAt)
	assert.NotEmpty(t, r.Event.CreatedAt)
	assert.NotEmpty(t, r.Event.DeletedAt)
	assert.NotEmpty(t, r.Event.UpdatedAt)
	assert.NotEmpty(t, r.Event.File.Id)
	assert.Equal(t, r.Event.File.Id, r.Event.FileId)
}
