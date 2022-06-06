package socket

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
)

func Test_newReleaseInstaller(t *testing.T) {
	chart := &chart.Chart{}
	installer := newReleaseInstaller("app", "dev", chart, nil, true, 10, true)
	assert.Equal(t, "app", installer.releaseName)
	assert.Equal(t, "dev", installer.namespace)
	assert.Equal(t, chart, installer.chart)
	assert.Equal(t, true, installer.wait)
	assert.Equal(t, int64(10), installer.timeoutSeconds)
	assert.Equal(t, true, installer.dryRun)
}

func Test_releaseInstaller_Chart(t *testing.T) {
	chart := &chart.Chart{}
	installer := newReleaseInstaller("app", "dev", chart, nil, true, 10, true)
	assert.Same(t, chart, installer.Chart())
}

func Test_releaseInstaller_Logs(t *testing.T) {
	installer := newReleaseInstaller("app", "dev", nil, nil, true, 10, true)
	installer.Logs()
}

func Test_releaseInstaller_Run(t *testing.T) {

}

func Test_releaseInstaller_logger(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := mock.NewMockDeployMsger(m)
	msger.EXPECT().SendProcessPercent("1").Times(1)
	installer := newReleaseInstaller("app", "dev", nil, nil, true, 10, true)
	installer.messageCh = &SafeWriteMessageCh{ch: make(chan contracts.MessageItem, 100)}
	installer.startTime = time.Now().Add(-5 * time.Minute)
	installer.percenter = newProcessPercent(msger, &fakeSleeper{})
	installer.logger()("test: %s", "aaa")
	assert.Equal(t, int64(1), installer.percenter.Current())
	msg := <-installer.messageCh.Chan()
	assert.Equal(t, "[如果长时间未成功，请试试 debug 模式]: test: aaa", msg.Msg)
	assert.Equal(t, contracts.MessageText, msg.Type)
}
