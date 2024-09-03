package socket

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
)

func TestTaskManagerRegister(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	taskManager := NewTaskManager(logger)

	err := taskManager.Register("task1", func(err error) {})
	assert.NoError(t, err)

	err = taskManager.Register("task1", func(err error) {})
	assert.Error(t, err)
}

func TestTaskManagerHas(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	taskManager := NewTaskManager(logger)

	taskManager.Register("task1", func(err error) {})
	assert.True(t, taskManager.Has("task1"))
	assert.False(t, taskManager.Has("task2"))
}

func TestTaskManagerRemove(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	taskManager := NewTaskManager(logger)

	taskManager.Register("task1", func(err error) {})
	taskManager.Remove("task1")
	assert.False(t, taskManager.Has("task1"))
}

func TestTaskManagerStop(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	taskManager := NewTaskManager(logger)

	stopped := false
	taskManager.Register("task1", func(err error) {
		if err == errCancel {
			stopped = true
		}
	})

	taskManager.Stop("task1")
	assert.True(t, stopped)
}

func TestTaskManagerStopAll(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	taskManager := NewTaskManager(logger)

	stopped1 := false
	stopped2 := false
	taskManager.Register("task1", func(err error) {
		if err == errCancel {
			stopped1 = true
		}
	})
	taskManager.Register("task2", func(err error) {
		if err == errCancel {
			stopped2 = true
		}
	})

	taskManager.StopAll()
	assert.True(t, stopped1)
	assert.True(t, stopped2)
}
