package mlog

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewLoggerWithZapChannel(t *testing.T) {
	cfg := &config.Config{LogChannel: "zap", Debug: true}
	logger := NewLogger(cfg)
	_, ok := logger.(*zapLogger)
	assert.True(t, ok)
}

func TestNewLoggerWithLogrusChannel(t *testing.T) {
	cfg := &config.Config{LogChannel: "logrus", Debug: true}
	logger := NewLogger(cfg)
	_, ok := logger.(*logrusLogger)
	assert.True(t, ok)
}

func TestNewLoggerWithDefaultChannel(t *testing.T) {
	cfg := &config.Config{LogChannel: "unknown", Debug: true}
	logger := NewLogger(cfg)
	_, ok := logger.(*logrusLogger)
	assert.True(t, ok)
}

func TestNewLoggerWithNilConfig(t *testing.T) {
	logger := NewLogger(nil)
	_, ok := logger.(*logrusLogger)
	assert.True(t, ok)
}
