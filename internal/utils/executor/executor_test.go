package executor

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultRemoteExecutor(t *testing.T) {
	assert.Implements(t, (*contracts.RemoteExecutor)(nil), NewDefaultRemoteExecutor())
}

func Test_defaultRemoteExecutor_Execute(t *testing.T) {}

func Test_defaultRemoteExecutor_WithCommand(t *testing.T) {
	cmd := []string{"sh", "-c", "ls"}
	assert.Equal(t, cmd, NewDefaultRemoteExecutor().WithCommand(cmd).(*defaultRemoteExecutor).cmd)
}

func Test_defaultRemoteExecutor_WithContainer(t *testing.T) {
	d := NewDefaultRemoteExecutor().WithContainer("ns", "app", "c").(*defaultRemoteExecutor)
	assert.Equal(t, "ns", d.namespace)
	assert.Equal(t, "app", d.pod)
	assert.Equal(t, "c", d.container)
}

func Test_defaultRemoteExecutor_WithMethod(t *testing.T) {
	assert.Equal(t, "POST", NewDefaultRemoteExecutor().WithMethod("POST").(*defaultRemoteExecutor).method)
}
