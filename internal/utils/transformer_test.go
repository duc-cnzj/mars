package utils

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetMarsNamespace(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	app.EXPECT().Config().Return(&config.Config{NsPrefix: "duc-"}).AnyTimes()
	assert.Equal(t, "duc-aa", GetMarsNamespace("aa"))
	assert.Equal(t, "duc-", GetMarsNamespace("duc-"))
}
