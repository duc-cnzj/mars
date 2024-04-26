package utils

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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
