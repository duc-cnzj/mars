package metrics

import (
	"github.com/google/wire"
)

// WireMetrics 提供 metrics 注册表的装配集。
var WireMetrics = wire.NewSet(NewRegistry)
