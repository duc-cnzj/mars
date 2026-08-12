package transport

import (
	"context"
	"errors"
	"io"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// fakeRawStream 是 RawStream 的最小实现：Recv 用 proto.Unmarshal 把 payload 灌进 out。
type fakeRawStream struct {
	payloads [][]byte
	closeErr error
	closed   bool
}

func (f *fakeRawStream) Recv(out proto.Message) error {
	if len(f.payloads) == 0 {
		return io.EOF
	}
	b := f.payloads[0]
	f.payloads = f.payloads[1:]
	return proto.Unmarshal(b, out)
}

func (f *fakeRawStream) Close() error {
	f.closed = true
	return f.closeErr
}

// fakeConn 是最小 Conn，只实现 OpenStream（其余方法本测试不调用）。
type fakeConn struct {
	raw RawStream
	err error
}

func (f *fakeConn) OpenStream(ctx context.Context, method, path string, req proto.Message) (RawStream, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.raw, nil
}

func (f *fakeConn) Do(ctx context.Context, method, path string, req, resp proto.Message) error {
	return errors.New("unimplemented")
}

func (f *fakeConn) DoNoRefresh(ctx context.Context, method, path string, req, resp proto.Message) error {
	return errors.New("unimplemented")
}

func (f *fakeConn) DoQuery(ctx context.Context, method, path string, req, resp proto.Message) error {
	return errors.New("unimplemented")
}

func TestOpenStream_ErrorPropagates(t *testing.T) {
	want := errors.New("open failed")
	conn := &fakeConn{err: want}
	_, err := OpenStream[*emptypb.Empty](conn, context.TODO(), "GET", "/x", nil)
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}

func TestOpenStream_WrapsTyped(t *testing.T) {
	payload, err := proto.Marshal(&emptypb.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	conn := &fakeConn{raw: &fakeRawStream{payloads: [][]byte{payload}}}
	s, err := OpenStream[*emptypb.Empty](conn, context.TODO(), "GET", "/x", nil)
	if err != nil {
		t.Fatal(err)
	}

	msg, err := s.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("Recv 返回了 nil")
	}

	if _, err := s.Recv(); err != io.EOF {
		t.Fatalf("流结束应返回 io.EOF，实际 %v", err)
	}
}

func TestTypedStream_CloseDelegates(t *testing.T) {
	raw := &fakeRawStream{}
	s := &typedStream[*emptypb.Empty]{raw: raw}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if !raw.closed {
		t.Fatal("Close 应委托给底层 RawStream")
	}
}

func TestTypedStream_RecvErrorPropagates(t *testing.T) {
	raw := &fakeRawStream{payloads: [][]byte{{1, 2, 3}}} // 非法 proto 字节 → 解码错误
	s := &typedStream[*emptypb.Empty]{raw: raw}
	if _, err := s.Recv(); err == nil {
		t.Fatal("非法字节应返回解码错误")
	}
}
