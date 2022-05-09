package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBCache_TableName(t *testing.T) {
	assert.Equal(t, "db_cache", (DBCache{}).TableName())
}
