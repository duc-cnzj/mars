package eventhandler

import "github.com/google/wire"

// WireEventHandler 是 eventhandler 包的 wire provider 集合。
// 惰性插件闭包（getCerts/toAll/PodEventPublisher）由组合根（cmd）注入。
var WireEventHandler = wire.NewSet(NewEventCoordinator, NewPodEventListener)
