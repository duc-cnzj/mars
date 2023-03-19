package executor

import (
	"bufio"
	"io"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
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

func Test_defaultRemoteExecutor_newOption(t *testing.T) {
	var (
		reader = strings.NewReader("")
		writer = bufio.NewWriter(nil)
	)
	var tests = []struct {
		wants     *v1.PodExecOptions
		in        io.Reader
		out, err  io.Writer
		tty       bool
		cmd       []string
		container string
	}{
		{
			wants: &v1.PodExecOptions{
				Stdin:     true,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
				Container: "c",
				Command:   []string{"ls"},
			},
			in:        reader,
			out:       writer,
			err:       writer,
			tty:       true,
			cmd:       []string{"ls"},
			container: "c",
		},
		{
			wants: &v1.PodExecOptions{
				Stdin:     false,
				Stdout:    false,
				Stderr:    false,
				TTY:       true,
				Container: "",
				Command:   []string{"ls"},
			},
			in:        nil,
			out:       nil,
			err:       nil,
			tty:       true,
			cmd:       []string{"ls"},
			container: "",
		},
	}

	for _, test := range tests {
		tt := test
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.wants, NewDefaultRemoteExecutor().WithCommand(tt.cmd).WithContainer("", "", tt.container).(*defaultRemoteExecutor).newOption(tt.in, tt.out, tt.err, tt.tty))
		})
	}
}
