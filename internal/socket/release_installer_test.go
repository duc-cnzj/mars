package socket

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
)

func Test_newReleaseInstaller(t *testing.T) {
	chart := &chart.Chart{}
	installer := newReleaseInstaller(nil, "app", "dev", chart, nil, true, 10, true)
	assert.Equal(t, "app", installer.releaseName)
	assert.Equal(t, "dev", installer.namespace)
	assert.Equal(t, chart, installer.chart)
	assert.Equal(t, true, installer.wait)
	assert.Equal(t, int64(10), installer.timeoutSeconds)
	assert.Equal(t, true, installer.dryRun)
}

func Test_releaseInstaller_Chart(t *testing.T) {
	chart := &chart.Chart{}
	installer := newReleaseInstaller(nil, "app", "dev", chart, nil, true, 10, true)
	assert.Same(t, chart, installer.Chart())
}

func Test_releaseInstaller_Logs(t *testing.T) {
	installer := newReleaseInstaller(nil, "app", "dev", nil, nil, true, 10, true)
	installer.Logs()
}

func Test_releaseInstaller_Run(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	h := mock.NewMockHelmer(m)
	h.EXPECT().UpgradeOrInstall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "xxx").Return(nil, nil).Times(1)
	_, err := (&releaseInstaller{helmer: h, dryRun: false}).Run(context.TODO(), nil, nil, false, "xxx")
	assert.Nil(t, err)
}

func Test_releaseInstaller_Run_Rollback(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	h := mock.NewMockHelmer(m)
	h.EXPECT().UpgradeOrInstall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "xxx").Return(nil, errors.New("xxx")).Times(1)
	h.EXPECT().Rollback(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("aaa")).Times(1)
	_, err := (&releaseInstaller{helmer: h, dryRun: false}).Run(context.TODO(), nil, nil, false, "xxx")
	assert.Equal(t, "xxx", err.Error())
}

func Test_releaseInstaller_Run_Uninstall(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	h := mock.NewMockHelmer(m)
	h.EXPECT().UpgradeOrInstall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "xxx").Return(nil, errors.New("xxx")).Times(1)
	h.EXPECT().Uninstall(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("bbb")).Times(1)
	_, err := (&releaseInstaller{helmer: h, dryRun: false}).Run(context.TODO(), nil, nil, true, "xxx")
	assert.Equal(t, "xxx", err.Error())
}

func Test_releaseInstaller_logger(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := mock.NewMockDeployMsger(m)
	msger.EXPECT().SendProcessPercent(int64(1)).Times(1)
	installer := newReleaseInstaller(nil, "app", "dev", nil, nil, true, 10, true)
	installer.messageCh = &safeWriteMessageCh{ch: make(chan contracts.MessageItem, 100)}
	installer.startTime = time.Now().Add(-5 * time.Minute)
	installer.percenter = newProcessPercent(msger, &fakeSleeper{})
	installer.logger()(nil, "test: %s", "aaa")
	assert.Equal(t, int64(1), installer.percenter.Current())
	msg := <-installer.messageCh.Chan()
	assert.Equal(t, "test: aaa", msg.Msg)
	assert.Equal(t, contracts.MessageText, msg.Type)
}
