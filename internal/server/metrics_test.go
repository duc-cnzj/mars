package server

import (
	"context"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMetricsRunnerRunAndShutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockLogger := mlog.NewMockLogger(m)
	mockRegistry := prometheus.NewRegistry()

	mockLogger.EXPECT().WithModule("server/metricsRunner").Return(mockLogger).Times(1)
	mockLogger.EXPECT().Infof("[Server]: metrics running at :%s/metrics", gomock.Any()).Times(1)

	metricsRunner := NewMetricsRunner("8080", mockLogger, mockRegistry)
	err := metricsRunner.Run(context.TODO())
	time.Sleep(1 * time.Second)
	metricsRunner.Shutdown(context.TODO())

	assert.NoError(t, err)
}

func TestMetricsRunnerShutdownError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	server := NewMockHttpServer(m)
	server.EXPECT().Shutdown(gomock.Any()).Return(assert.AnError).Times(1)
	metricsRunner := &metricsRunner{s: server}
	err := metricsRunner.Shutdown(context.TODO())
	assert.Error(t, err)
}
