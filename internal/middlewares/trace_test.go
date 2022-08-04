package middlewares

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracingIgnoreFn(t *testing.T) {
	assert.False(t, TracingIgnoreFn(context.TODO(), "/api/xxx"))
	assert.True(t, TracingIgnoreFn(context.TODO(), "/xxx"))
	assert.True(t, TracingIgnoreFn(context.TODO(), "/ws"))
	assert.True(t, TracingIgnoreFn(context.TODO(), "/api/containers/namespaces/{namespace}/pods/{pod}/containers/{container}/stream_logs"))
	assert.True(t, TracingIgnoreFn(context.TODO(), "/api/metrics/namespace/{namespace}/pods/{pod}/stream"))
}

func TestTracingWrapper(t *testing.T) {
}
