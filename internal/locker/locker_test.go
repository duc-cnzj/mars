package locker

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/stretchr/testify/assert"
)

func TestNewLocker_WithDBDriver(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{DBDriver: "db", CacheDriver: "db"}
	logger := mlog.NewForConfig(nil)
	data := data.NewData(nil, logger)
	timer := timer.NewReal()

	locker, err := NewLocker(cfg, data, logger, timer)

	assert.NoError(t, err)
	assert.IsType(t, &databaseLock{}, locker)
}

func TestNewLocker_WithSQLiteDriver(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{DBDriver: "sqlite", CacheDriver: "db"}
	logger := mlog.NewForConfig(nil)
	data := data.NewData(nil, logger)
	timer := timer.NewReal()

	locker, err := NewLocker(cfg, data, logger, timer)

	assert.NoError(t, err)
	assert.IsType(t, &memoryLock{}, locker)
}

func TestNewLocker_WithMemoryDriver(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{DBDriver: "db", CacheDriver: "memory"}
	logger := mlog.NewForConfig(nil)
	data := data.NewData(nil, logger)
	timer := timer.NewReal()
	locker, err := NewLocker(cfg, data, logger, timer)

	assert.NoError(t, err)
	assert.IsType(t, &memoryLock{}, locker)
}

func TestNewLocker_WithUnknownDriver(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{DBDriver: "db", CacheDriver: "unknown"}
	logger := mlog.NewForConfig(nil)
	data := data.NewData(nil, logger)
	timer := timer.NewReal()
	locker, err := NewLocker(cfg, data, logger, timer)

	assert.NoError(t, err)
	assert.IsType(t, &memoryLock{}, locker)
}
