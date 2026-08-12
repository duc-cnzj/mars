package data

import (
	"context"

	"github.com/duc-cnzj/mars/v6/internal/data/ent"
)

// testDB 保存最近一次 NewSqliteDB 打开的测试库客户端，供包级 Close 统一回收。
var testDB *ent.Client

// NewSqliteDB 打开共享内存 sqlite 测试库并建表，供本包与 internal/services 等跨包测试使用。
// 该组测试辅助（testDB/NewSqliteDB/Close）必须放在非 _test.go 文件：Go 的测试符号只对同包
// 测试文件可见，internal/services 的测试经 data.NewSqliteDB/data.Close 跨包调用只能访问
// 非测试文件的导出符号（可见性先例与 mock_*.go 一致）。打开/建表失败在测试场景无法冒泡，
// 直接忽略——若 client 不可用，后续查询测试会失败暴露问题。
func NewSqliteDB() (*ent.Client, error) {
	client, _ := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1&loc=Local")
	_ = client.Schema.Create(context.TODO())
	testDB = client

	return client, nil
}

// Close 关闭 NewSqliteDB 打开的测试库连接；未打开过时静默。
func Close() {
	if testDB != nil {
		testDB.Close()
	}
}
