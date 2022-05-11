package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisteredEndpoints(t *testing.T) {
	assert.Len(t, RegisteredEndpoints(), 14)
}

func TestRegisteredServers(t *testing.T) {
	assert.Len(t, RegisteredServers(), 14)
}
