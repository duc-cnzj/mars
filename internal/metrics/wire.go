package metrics

import (
	"github.com/google/wire"
)

var WireMetrics = wire.NewSet(NewRegistry)
