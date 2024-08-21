package socket

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	timer2 "github.com/duc-cnzj/mars/v4/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewReleaseInstaller(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewLogger(nil)
	helmer := repo.NewMockHelmerRepo(m)
	data := data.NewMockData(m)
	timer := timer2.NewRealTimer()
	data.EXPECT().Config().Return(&config.Config{
		InstallTimeout: 100 * time.Second,
	})

	installer := NewReleaseInstaller(logger, helmer, data, timer)

	assert.NotNil(t, installer)
	assert.Equal(t, logger, installer.(*releaseInstaller).logger)
	assert.Equal(t, helmer, installer.(*releaseInstaller).helmer)
	assert.Equal(t, int64(100), installer.(*releaseInstaller).timeoutSeconds)
	assert.Equal(t, timer, installer.(*releaseInstaller).timer)
}

func TestTimeOrderedSetString(t *testing.T) {
	tos := newTimeOrderedSetString(timer2.NewRealTimer())

	tos.add("test1")
	assert.True(t, tos.has("test1"))
	assert.False(t, tos.has("test2"))

	tos.add("test2")
	assert.True(t, tos.has("test2"))

	items := tos.sortedItems()
	assert.Equal(t, 2, len(items))
	assert.Equal(t, "test1", items[0])
	assert.Equal(t, "test2", items[1])
}

func TestTimeOrderedSetString_Concurrency(t *testing.T) {
	tos := newTimeOrderedSetString(timer2.NewRealTimer())
	var wg sync.WaitGroup

	tos.add("duc")
	tos.add("duc")
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tos.add(fmt.Sprintf("%v", i))
		}(i)
	}

	wg.Wait()

	for i := 0; i < 100; i++ {
		assert.True(t, tos.has(fmt.Sprintf("%v", i)))
	}

	items := tos.sortedItems()
	assert.Equal(t, 101, len(items))
}

func TestLoggerWrapFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	messageChan := NewSafeWriteMessageCh(mlog.NewLogger(nil), 100)
	percenter := NewMockPercentable(m)
	logs := newTimeOrderedSetString(timer2.NewRealTimer())

	// Mock expectations
	percenter.EXPECT().Current().Return(int64(98)).Times(1)
	percenter.EXPECT().Add().Times(1)

	data := data.NewMockData(m)
	data.EXPECT().Config().Return(&config.Config{
		InstallTimeout: 100 * time.Second,
	})

	// Call the function under test
	loggerWrap := NewReleaseInstaller(nil, nil, data, nil).(*releaseInstaller).
		loggerWrap(messageChan, percenter, logs)
	loggerWrap(nil, "test message %d", 1)

	// Assert that the message was added to logs
	assert.True(t, logs.has("test message 1"))
}

func TestLoggerWrapEdgeCase(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	messageChan := NewSafeWriteMessageCh(mlog.NewLogger(nil), 100)
	percenter := NewMockPercentable(m)
	logs := newTimeOrderedSetString(timer2.NewRealTimer())

	// Mock expectations
	percenter.EXPECT().Current().Return(int64(99)).Times(1)

	data := data.NewMockData(m)
	data.EXPECT().Config().Return(&config.Config{
		InstallTimeout: 100 * time.Second,
	})
	loggerWrap := NewReleaseInstaller(nil, nil, data, nil).(*releaseInstaller).
		loggerWrap(messageChan, percenter, logs)
	loggerWrap(nil, "test message %d", 1)

	// Assert that the message was added to logs
	assert.True(t, logs.has("test message 1"))
}

func Test_releaseInstaller_Run_Dry(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	helmer := repo.NewMockHelmerRepo(m)
	ri := &releaseInstaller{
		timer:          timer2.NewRealTimer(),
		helmer:         helmer,
		timeoutSeconds: 10,
		logger:         mlog.NewLogger(nil),
	}

	ctx := context.TODO()
	helmer.EXPECT().UpgradeOrInstall(ctx, "name", "ns", gomock.Any(), gomock.Any(), gomock.Any(), false, int64(10), true, "desc").Return(nil, errors.New("x"))
	_, err := ri.Run(ctx, &InstallInput{
		DryRun:      true,
		Namespace:   "ns",
		ReleaseName: "name",
		Description: "desc",
	})
	assert.Error(t, err)
}

func Test_releaseInstaller_Run_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	helmer := repo.NewMockHelmerRepo(m)
	ri := &releaseInstaller{
		timer:          timer2.NewRealTimer(),
		helmer:         helmer,
		timeoutSeconds: 10,
		logger:         mlog.NewLogger(nil),
	}

	ctx := context.TODO()
	helmer.EXPECT().UpgradeOrInstall(ctx, "name", "ns", gomock.Any(), gomock.Any(), gomock.Any(), false, int64(10), true, "desc").Return(nil, nil)
	_, err := ri.Run(ctx, &InstallInput{
		DryRun:      true,
		Namespace:   "ns",
		ReleaseName: "name",
		Description: "desc",
	})
	assert.Nil(t, err)
}

func Test_releaseInstaller_Run(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	helmer := repo.NewMockHelmerRepo(m)
	ri := &releaseInstaller{
		timer:          timer2.NewRealTimer(),
		helmer:         helmer,
		timeoutSeconds: 10,
		logger:         mlog.NewLogger(nil),
	}

	ctx := context.TODO()
	helmer.EXPECT().UpgradeOrInstall(ctx, "name", "ns", gomock.Any(), gomock.Any(), gomock.Any(), false, int64(10), false, "desc").Return(nil, errors.New("x"))

	helmer.EXPECT().Uninstall("name", "ns", gomock.Any()).Return(errors.New("y"))
	_, err := ri.Run(ctx, &InstallInput{
		IsNew:       true,
		DryRun:      false,
		Namespace:   "ns",
		ReleaseName: "name",
		Description: "desc",
	})
	assert.Error(t, err)
}

func Test_releaseInstaller_Run_2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	helmer := repo.NewMockHelmerRepo(m)
	ri := &releaseInstaller{
		timer:          timer2.NewRealTimer(),
		helmer:         helmer,
		timeoutSeconds: 10,
		logger:         mlog.NewLogger(nil),
	}

	ctx := context.TODO()
	helmer.EXPECT().UpgradeOrInstall(ctx, "name", "ns", gomock.Any(), gomock.Any(), gomock.Any(), false, int64(10), false, "desc").Return(nil, errors.New("x"))

	helmer.EXPECT().Rollback("name", "ns", false, gomock.Any(), false).Return(errors.New("y"))
	_, err := ri.Run(ctx, &InstallInput{
		IsNew:       false,
		DryRun:      false,
		Namespace:   "ns",
		ReleaseName: "name",
		Description: "desc",
	})
	assert.Error(t, err)
}
