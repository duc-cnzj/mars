package server

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
)

func TestPprofRunnerRunAndShutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockLogger := mlog.NewMockLogger(m)

	mockLogger.EXPECT().WithModule("server/pprofRunner").Return(mockLogger).Times(1)
	mockLogger.EXPECT().Info("[Server]: start pprofRunner runner.").Times(1)
	mockLogger.EXPECT().Info("Starting pprof server on localhost:6060.").Times(1)
	mockLogger.EXPECT().Info("[Server]: shutdown pprofRunner runner.").Times(1)

	runner := NewPprofRunner(mockLogger)
	err := runner.Run(context.TODO())
	time.Sleep(1 * time.Second)
	runner.Shutdown(context.TODO())

	assert.NoError(t, err)
}

func TestPprofRunnerRunError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockLogger := mlog.NewMockLogger(m)
	server := NewMockHttpServer(m)
	runner := &pprofRunner{logger: mockLogger, server: server}
	mockLogger.EXPECT().Info("[Server]: start pprofRunner runner.").Times(1)
	mockLogger.EXPECT().Info("Starting pprof server on localhost:6060.").Times(1)
	mockLogger.EXPECT().Error(gomock.Any()).Times(1)
	server.EXPECT().ListenAndServe().Return(assert.AnError).Times(1)
	runner.Run(context.TODO())

	time.Sleep(1 * time.Second)
}

func Test_pprofRunner_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockLogger := mlog.NewMockLogger(m)
	server := NewMockHttpServer(m)
	runner := &pprofRunner{logger: mockLogger, server: server}
	mockLogger.EXPECT().Info("[Server]: shutdown pprofRunner runner.").Times(1)
	server.EXPECT().Shutdown(gomock.Any()).Return(nil).Times(1)
	assert.Nil(t, runner.Shutdown(context.TODO()))
}
