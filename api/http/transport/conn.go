// Package transport 提供生成 stub 依赖的手写传输契约，与自动生成的
// rest 子包（api/http/rest）物理隔离：rest/ 只放 *.gen.http.go，
// 本包只放手写类型。
package transport

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// Conn 是生成 stub（*.gen.http.go）依赖的最小传输接口，由 http.Client
// 的未导出 ops 适配器实现。
//
// 拆成独立子包是 Go 的 one-package-per-directory 约束决定的：生成代码必须落在
// 单独目录（rest 子包），而一个目录一个包，所以生成的 stub 无法裸引用本包符号，
// 只能通过 transport.Conn 等限定名访问。这样生成代码与手写引擎
// （http/client.go）分层，同时 http.Client 的公共 API 零膨胀。
// File 的 multipart 上传/二进制下载等自定义路由不在任何 proto 里，由生成器覆盖
// 不到，永远手写在 http 包（file_custom.go），挂 *Client 上，不进本包。
type Conn interface {
	Do(ctx context.Context, method, path string, req, resp proto.Message) error
	DoNoRefresh(ctx context.Context, method, path string, req, resp proto.Message) error
	DoQuery(ctx context.Context, method, path string, req, resp proto.Message) error
	// OpenStream 打开一条 server-streaming 流（消费 grpc-gateway 的 NDJSON/SSE 输出），
	// 由 OpenStream[T] 泛型工厂包装成类型安全的 Stream[T]。
	OpenStream(ctx context.Context, method, path string, req proto.Message) (RawStream, error)
}
