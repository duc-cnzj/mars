package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/logger"
)

func TestGormLoggerAdapter(t *testing.T) {
	assert.Implements(t, (*logger.Interface)(nil), &GormLoggerAdapter{})
}
