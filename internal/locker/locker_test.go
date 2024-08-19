package locker

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/timer"
	"github.com/stretchr/testify/assert"
)

func TestNewLocker_WithDBDriver(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{DBDriver: "db", CacheDriver: "db"}
	logger := mlog.NewLogger(nil)
	data := data.NewData(nil, logger)
	timer := timer.NewRealTimer()

	locker, err := NewLocker(cfg, data, logger, timer)

	assert.NoError(t, err)
	assert.IsType(t, &databaseLock{}, locker)
}

func TestNewLocker_WithSQLiteDriver(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{DBDriver: "sqlite", CacheDriver: "db"}
	logger := mlog.NewLogger(nil)
	data := data.NewData(nil, logger)
	timer := timer.NewRealTimer()

	locker, err := NewLocker(cfg, data, logger, timer)

	assert.NoError(t, err)
	assert.IsType(t, &memoryLock{}, locker)
}

func TestNewLocker_WithMemoryDriver(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{DBDriver: "db", CacheDriver: "memory"}
	logger := mlog.NewLogger(nil)
	data := data.NewData(nil, logger)
	timer := timer.NewRealTimer()
	locker, err := NewLocker(cfg, data, logger, timer)

	assert.NoError(t, err)
	assert.IsType(t, &memoryLock{}, locker)
}

func TestNewLocker_WithUnknownDriver(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{DBDriver: "db", CacheDriver: "unknown"}
	logger := mlog.NewLogger(nil)
	data := data.NewData(nil, logger)
	timer := timer.NewRealTimer()
	locker, err := NewLocker(cfg, data, logger, timer)

	assert.NoError(t, err)
	assert.IsType(t, &memoryLock{}, locker)
}