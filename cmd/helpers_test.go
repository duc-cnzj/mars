package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFileExists 验证 fileExists 对存在/不存在路径的判定。
func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(existing, []byte("x"), 0o600))

	assert.True(t, fileExists(existing))
	assert.False(t, fileExists(filepath.Join(dir, "missing.yaml")))
}

// TestGetFunctionName 验证 GetFunctionName 返回可辨识的运行时函数名。
func TestGetFunctionName(t *testing.T) {
	name := GetFunctionName(fileExists)
	assert.NotEmpty(t, name)
	assert.Contains(t, name, "fileExists")
}
