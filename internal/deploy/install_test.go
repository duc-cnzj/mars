package deploy

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	timer2 "github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewReleaseInstaller(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewForConfig(nil)
	helmer := data.NewMockHelmerRepo(m)
	timer := timer2.NewReal()

	installer := NewReleaseInstaller(logger, helmer, &config.Config{
		InstallTimeout: 100 * time.Second,
	}, timer)

	assert.NotNil(t, installer)
	assert.Equal(t, logger, installer.(*releaseInstaller).logger)
	assert.Equal(t, helmer, installer.(*releaseInstaller).helmer)
	assert.Equal(t, int64(100), installer.(*releaseInstaller).timeoutSeconds)
	assert.Equal(t, timer, installer.(*releaseInstaller).timer)
}

func TestTimeOrderedSetString(t *testing.T) {
	tos := newTimeOrderedSetString(timer2.NewReal())

	tos.add("test1")
	assert.True(t, tos.has("test1"))
	assert.False(t, tos.has("test2"))

	tos.add("test2")
	assert.True(t, tos.has("test2"))

	// 去重：重复 add 不增加集合大小。
	tos.add("test1")
	assert.Equal(t, 2, len(tos.items))
}

func TestTimeOrderedSetString_Concurrency(t *testing.T) {
	tos := newTimeOrderedSetString(timer2.NewReal())
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
	// 并发去重后集合大小为 100 个并发字符串 + 1 个 "duc"。
	assert.Equal(t, 101, len(tos.items))
}

func TestLoggerWrapFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	messageChan := newSafeWriteMessageCh(mlog.NewForConfig(nil), 100)
	percenter := NewMockPercentable(m)
	logs := newTimeOrderedSetString(timer2.NewReal())

	// Mock expectations
	percenter.EXPECT().Current().Return(int64(98)).Times(1)
	percenter.EXPECT().Add().Times(1)

	// Call the function under test
	loggerWrap := NewReleaseInstaller(nil, nil, &config.Config{
		InstallTimeout: 100 * time.Second,
	}, nil).(*releaseInstaller).
		loggerWrap(messageChan, percenter, logs)
	loggerWrap(nil, "test message %d", 1)

	// Assert that the message was added to logs
	assert.True(t, logs.has("test message 1"))
}

func TestLoggerWrapEdgeCase(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	messageChan := newSafeWriteMessageCh(mlog.NewForConfig(nil), 100)
	percenter := NewMockPercentable(m)
	logs := newTimeOrderedSetString(timer2.NewReal())

	// Mock expectations
	percenter.EXPECT().Current().Return(int64(99)).Times(1)

	loggerWrap := NewReleaseInstaller(nil, nil, &config.Config{
		InstallTimeout: 100 * time.Second,
	}, nil).(*releaseInstaller).
		loggerWrap(messageChan, percenter, logs)
	loggerWrap(nil, "test message %d", 1)

	// Assert that the message was added to logs
	assert.True(t, logs.has("test message 1"))
}

func Test_releaseInstaller_Run_Dry(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	helmer := data.NewMockHelmerRepo(m)
	ri := &releaseInstaller{
		timer:          timer2.NewReal(),
		helmer:         helmer,
		timeoutSeconds: 10,
		logger:         mlog.NewForConfig(nil),
	}

	ctx := context.TODO()
	helmer.EXPECT().UpgradeOrInstall(ctx,
		"name", "ns",
		gomock.Not(nil), gomock.Not(nil),
		gomock.Not(nil), false,
		int64(10), true, "desc",
	).Return(nil, errors.New("x"))
	percentable := NewMockPercentable(m)
	percentable.EXPECT().Add().AnyTimes()
	percentable.EXPECT().Current().AnyTimes()
	messageChan := NewMockSafeWriteMessageChan(m)
	messageChan.EXPECT().Send(MessageItem{
		Msg:  "部署出现问题: x",
		Type: MessageText,
	})
	_, err := ri.Run(ctx, &InstallInput{
		DryRun:      true,
		Namespace:   "ns",
		ReleaseName: "name",
		Description: "desc",
		messageChan: messageChan,
		percenter:   percentable,
	})
	assert.Error(t, err)
}

func Test_releaseInstaller_Run_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	helmer := data.NewMockHelmerRepo(m)
	ri := &releaseInstaller{
		timer:          timer2.NewReal(),
		helmer:         helmer,
		timeoutSeconds: 10,
		logger:         mlog.NewForConfig(nil),
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
	helmer := data.NewMockHelmerRepo(m)
	ri := &releaseInstaller{
		timer:          timer2.NewReal(),
		helmer:         helmer,
		timeoutSeconds: 10,
		logger:         mlog.NewForConfig(nil),
	}

	ctx := context.TODO()
	helmer.EXPECT().UpgradeOrInstall(ctx, "name", "ns", gomock.Any(), gomock.Any(), gomock.Any(), false, int64(10), false, "desc").Return(nil, errors.New("x"))

	helmer.EXPECT().Uninstall("name", "ns", gomock.Any()).Return(errors.New("y"))
	percentable := NewMockPercentable(m)
	percentable.EXPECT().Add().AnyTimes()
	percentable.EXPECT().Current().AnyTimes()
	messageChan := NewMockSafeWriteMessageChan(m)
	messageChan.EXPECT().Send(gomock.Any()).AnyTimes()
	_, err := ri.Run(ctx, &InstallInput{
		IsNew:       true,
		DryRun:      false,
		Namespace:   "ns",
		ReleaseName: "name",
		Description: "desc",
		percenter:   percentable,
		messageChan: messageChan,
	})
	assert.Error(t, err)
}

func Test_releaseInstaller_Run_2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	helmer := data.NewMockHelmerRepo(m)
	ri := &releaseInstaller{
		timer:          timer2.NewReal(),
		helmer:         helmer,
		timeoutSeconds: 10,
		logger:         mlog.NewForConfig(nil),
	}

	ctx := context.TODO()
	helmer.EXPECT().UpgradeOrInstall(ctx, "name", "ns", gomock.Any(), gomock.Any(), gomock.Any(), false, int64(10), false, "desc").Return(nil, errors.New("x"))

	helmer.EXPECT().Rollback("name", "ns", false, gomock.Any(), false).Return(errors.New("y"))
	percentable := NewMockPercentable(m)
	percentable.EXPECT().Add().AnyTimes()
	percentable.EXPECT().Current().AnyTimes()
	messageChan := NewMockSafeWriteMessageChan(m)
	messageChan.EXPECT().Send(gomock.Any()).AnyTimes()
	_, err := ri.Run(ctx, &InstallInput{
		IsNew:       false,
		DryRun:      false,
		Namespace:   "ns",
		ReleaseName: "name",
		Description: "desc",
		percenter:   percentable,
		messageChan: messageChan,
	})
	assert.Error(t, err)
}

func Test_releaseInstaller_Run_TimeoutOverride(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	helmer := data.NewMockHelmerRepo(m)
	ri := &releaseInstaller{
		timer:          timer2.NewReal(),
		helmer:         helmer,
		timeoutSeconds: 10,
		logger:         mlog.NewForConfig(nil),
	}

	ctx := context.TODO()
	// 调用方按请求传入 TimeoutSeconds 时，必须覆盖构造时的默认超时。
	helmer.EXPECT().UpgradeOrInstall(ctx, "name", "ns", gomock.Any(), gomock.Any(), gomock.Any(), false, int64(30), true, "desc").Return(nil, nil)
	_, err := ri.Run(ctx, &InstallInput{
		DryRun:         true,
		Namespace:      "ns",
		ReleaseName:    "name",
		Description:    "desc",
		TimeoutSeconds: 30,
	})
	assert.Nil(t, err)
}

func TestSafeWriteMessageChSendWhenNotClosed(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	ch := newSafeWriteMessageCh(logger, 1)

	ch.Send(MessageItem{Msg: "test", Type: MessageSuccess})

	select {
	case msg := <-ch.Chan():
		assert.Equal(t, "test", msg.Msg)
		assert.Equal(t, MessageSuccess, msg.Type)
	default:
		t.Fatal("Expected message to be sent")
	}
}

func TestSafeWriteMessageChSendWhenClosed(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	ch := newSafeWriteMessageCh(logger, 1)

	ch.Close()
	ch.Send(MessageItem{Msg: "test", Type: MessageText})

	select {
	case _, ok := <-ch.Chan():
		if ok {
			t.Fatal("Expected no message to be sent")
		}
	default:
	}
}

func TestSafeWriteMessageChSendWhenFull(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	ch := newSafeWriteMessageCh(logger, 1)

	ch.Send(MessageItem{Msg: "test1", Type: MessageError})
	ch.Send(MessageItem{Msg: "test2", Type: MessageError})

	select {
	case msg := <-ch.Chan():
		assert.Equal(t, "test1", msg.Msg)
		assert.Equal(t, MessageError, msg.Type)
	default:
		t.Fatal("Expected message to be sent")
	}

	select {
	case <-ch.Chan():
		t.Fatal("Expected no message to be sent")
	default:
	}
}
