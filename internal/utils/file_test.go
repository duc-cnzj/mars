package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	getwd, _ := os.Getwd()
	assert.False(t, FileExists("xxx"))
	assert.True(t, FileExists(filepath.Join(getwd, "file_test.go")))
}
