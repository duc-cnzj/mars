package locker

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
)

func TestResolveDriver_DB(t *testing.T) {
	assert.Equal(t, DriverDB, ResolveDriver("mysql", "db", mlog.NewForConfig(nil)))
}

func TestResolveDriver_Memory(t *testing.T) {
	assert.Equal(t, DriverMemory, ResolveDriver("mysql", "memory", mlog.NewForConfig(nil)))
}

func TestResolveDriver_Unknown(t *testing.T) {
	assert.Equal(t, Driver("unknown"), ResolveDriver("mysql", "unknown", mlog.NewForConfig(nil)))
}

// TestResolveDriver_SQLiteFallback 覆盖 sqlite + db 组合强制回退内存锁的规则。
func TestResolveDriver_SQLiteFallback(t *testing.T) {
	assert.Equal(t, DriverMemory, ResolveDriver("sqlite", "db", mlog.NewForConfig(nil)))
}

func TestNewLocker_WithDBDriver(t *testing.T) {
	t.Parallel()

	locker := NewLocker(DriverDB, func() *ent.Client { return nil }, mlog.NewForConfig(nil), timer.NewReal())

	assert.IsType(t, &databaseLock{}, locker)
}

func TestNewLocker_WithMemoryDriver(t *testing.T) {
	t.Parallel()

	locker := NewLocker(DriverMemory, func() *ent.Client { return nil }, mlog.NewForConfig(nil), timer.NewReal())

	assert.IsType(t, &memoryLock{}, locker)
}

func TestNewLocker_WithUnknownDriver(t *testing.T) {
	t.Parallel()

	locker := NewLocker(Driver("unknown"), func() *ent.Client { return nil }, mlog.NewForConfig(nil), timer.NewReal())

	assert.IsType(t, &memoryLock{}, locker)
}
