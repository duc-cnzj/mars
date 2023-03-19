package adapter

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/mock"
)

func TestNsqLoggerAdapter_Output(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Error(gomock.Any()).Times(1)
	l.EXPECT().Debug(gomock.Any()).Times(1)
	nsql := &NsqLoggerAdapter{}
	nsql.Output(1, "")
	nsql.Output(1, "TOPIC_NOT_FOUND")
}
