package adapter

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestNsqLoggerAdapter_Output(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	l.EXPECT().Error(gomock.Any()).Times(1)
	nsql := &NsqLoggerAdapter{}
	nsql.Output(1, "")
}
