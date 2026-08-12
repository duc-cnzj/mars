package config

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 说明：这两个测试是包内白盒测试，唯一目的是通过替换 listenTCP 探针
// 确定性覆盖 net.Listen 失败的错误分支——真实网络层无法稳定触发该错误，
// 黑盒（config_test）无从构造。包内其它测试保持黑盒，不因此破坏约定。

// TestGetFreePortListenError 覆盖 GetFreePort 中 net.Listen 失败的返回分支。
func TestGetFreePortListenError(t *testing.T) {
	orig := listenTCP
	listenTCP = func(network, address string) (net.Listener, error) {
		return nil, errors.New("listen failed")
	}
	defer func() { listenTCP = orig }()

	_, err := GetFreePort()
	assert.Error(t, err)
}

// TestInitPanicsWhenPortAllocationFails 覆盖 Init 自动分配端口失败时的 panic 分支。
func TestInitPanicsWhenPortAllocationFails(t *testing.T) {
	orig := listenTCP
	listenTCP = func(network, address string) (net.Listener, error) {
		return nil, errors.New("listen failed")
	}
	defer func() { listenTCP = orig }()

	assert.Panics(t, func() { Init("testdata/config_minimal.yaml") })
}
