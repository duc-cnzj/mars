package timer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRealTimer(t *testing.T) {
	assert.Implements(t, (*Timer)(nil), NewRealTimer())
}

func Test_realTimer_Now(t *testing.T) {
	format := "200601021504"
	assert.Equal(t, NewRealTimer().Now().Format(format), time.Now().Format(format))
}
