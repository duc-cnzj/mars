package repo

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPictureRepo_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPluginManager := application.NewMockPluginManger(ctrl)
	mockPicture := application.NewMockPicture(ctrl)

	mockPluginManager.EXPECT().Picture().Return(mockPicture).Times(2)

	mockPicture.EXPECT().Get(context.TODO(), true).Return(&application.PictureItem{}, nil).Times(1)
	mockPicture.EXPECT().Get(context.TODO(), false).Return(&application.PictureItem{}, nil).Times(1)

	repo := NewPictureRepo(mlog.NewForConfig(nil), mockPluginManager)

	_, err := repo.Get(context.TODO(), true)
	assert.Nil(t, err)

	_, err = repo.Get(context.TODO(), false)
	assert.Nil(t, err)
}

func TestPictureRepo_Get_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPluginManager := application.NewMockPluginManger(ctrl)
	mockPicture := application.NewMockPicture(ctrl)

	mockPluginManager.EXPECT().Picture().Return(mockPicture).Times(1)

	mockPicture.EXPECT().Get(context.TODO(), true).Return(nil, assert.AnError).Times(1)

	repo := NewPictureRepo(mlog.NewForConfig(nil), mockPluginManager)

	_, err := repo.Get(context.TODO(), true)
	assert.NotNil(t, err)
}
