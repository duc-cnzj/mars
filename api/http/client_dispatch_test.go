package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/metrics"
	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
	"github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// encodeQuery 全字段分派：bool/enum/int/uint/float/double/bytes/嵌套/重复 bool，
// 以及 proto3 optional 零值跳过。用真实 proto 消息逐 kind 验证。
func Test_encodeQuery_AllKinds(t *testing.T) {
	t.Run("bool+enum+int32+optional zero-skip", func(t *testing.T) {
		// atomic=true 走 scalarString bool true；type 枚举走 enum 分支；namespace_id 走 int32。
		got := encodeQuery(&websocket.CreateProjectInput{
			Type:        websocket.Type_CreateProject,
			NamespaceId: 5,
			Atomic:      ptr(true),
		})
		for _, want := range []string{"type=4", "namespaceId=5", "atomic=true"} {
			if !strings.Contains(got, want) {
				t.Errorf("got %q, want contains %q", got, want)
			}
		}
	})

	t.Run("optional零值仍被Range枚举并跳过", func(t *testing.T) {
		// proto3 optional 显式置零会出现在 Range 里 → appendQuery 的 isZeroScalar return true。
		// 结果应不含 atomic（零值跳过）。
		got := encodeQuery(&websocket.CreateProjectInput{Atomic: ptr(false)})
		if got != "" {
			t.Errorf("got %q, want empty（零值跳过）", got)
		}
	})

	t.Run("uint32+bytes", func(t *testing.T) {
		got := encodeQuery(&websocket.TerminalMessage{
			Op:     "resize",
			Height: 24,
			Data:   []byte{1, 2}, // bytes 无标量分支 → isZeroScalar default → scalarString default
		})
		for _, want := range []string{"op=resize", "height=24", "data="} {
			if !strings.Contains(got, want) {
				t.Errorf("got %q, want contains %q", got, want)
			}
		}
	})

	t.Run("double", func(t *testing.T) {
		got := encodeQuery(&metrics.TopPodResponse{Cpu: 1.5, Memory: 2.5})
		for _, want := range []string{"cpu=1.5", "memory=2.5"} {
			if !strings.Contains(got, want) {
				t.Errorf("got %q, want contains %q", got, want)
			}
		}
	})

	t.Run("嵌套消息带prefix递归", func(t *testing.T) {
		got := encodeQuery(&websocket.WsMetadataResponse{
			Metadata: &websocket.Metadata{Id: "x", Type: websocket.Type_SetUid},
		})
		for _, want := range []string{"metadata.id=x", "metadata.type=1"} {
			if !strings.Contains(got, want) {
				t.Errorf("got %q, want contains %q", got, want)
			}
		}
	})
}

// encodeQuery(nil) → 空串。
func Test_encodeQuery_NilMessage(t *testing.T) {
	if got := encodeQuery(nil); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

// 动态消息合成 float32 / repeated bool 字段（全仓 proto 无此类字段，用 descriptor 合成直测）。
func Test_encodeQuery_DynamicKinds(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Syntax:  strPtr("proto3"),
		Name:    strPtr("query_kinds.proto"),
		Package: strPtr("querykinds"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: strPtr("Q"),
			Field: []*descriptorpb.FieldDescriptorProto{
				{Name: strPtr("fl"), Number: int32Ptr(1), Label: labelOpt(), Type: typeFloat()},
				{Name: strPtr("rb"), Number: int32Ptr(2), Label: labelRep(), Type: typeBool()},
			},
		}},
	}
	file, err := protodesc.NewFile(fd, nil)
	if err != nil {
		t.Fatal(err)
	}
	md := file.Messages().ByName("Q")
	flFD, rbFD := md.Fields().ByName("fl"), md.Fields().ByName("rb")

	// float32 非零 → scalarString Float 分支。
	// repeated bool 含 false 元素 → scalarString Bool false 分支。
	m := dynamicpb.NewMessage(md)
	m.Set(flFD, protoreflect.ValueOfFloat32(1.5))
	rb := m.Mutable(rbFD).List()
	rb.Append(protoreflect.ValueOfBool(false))
	rb.Append(protoreflect.ValueOfBool(true))

	got := encodeQuery(m)
	for _, want := range []string{"fl=1.5", "rb=false", "rb=true"} {
		if !strings.Contains(got, want) {
			t.Errorf("got %q, want contains %q", got, want)
		}
	}
}

// isZeroScalar 各 kind 零值判定（纯函数直测，覆盖 proto3 下 Range 枚举不到的分支）。
func Test_isZeroScalar(t *testing.T) {
	field := func(m protoreflect.Message, name string) protoreflect.FieldDescriptor {
		return m.Descriptor().Fields().ByName(protoreflect.Name(name))
	}
	create := (&websocket.CreateProjectInput{}).ProtoReflect()
	term := (&websocket.TerminalMessage{}).ProtoReflect()
	top := (&metrics.TopPodResponse{}).ProtoReflect()

	tests := []struct {
		name string
		fd   protoreflect.FieldDescriptor
		v    protoreflect.Value
		want bool
	}{
		{"bool zero → true", field(create, "atomic"), protoreflect.ValueOfBool(false), true},
		{"bool nonzero → false", field(create, "atomic"), protoreflect.ValueOfBool(true), false},
		{"enum zero → true", field(create, "type"), protoreflect.ValueOfEnum(0), true},
		{"enum nonzero → false", field(create, "type"), protoreflect.ValueOfEnum(1), false},
		{"uint zero → true", field(term, "height"), protoreflect.ValueOfUint32(0), true},
		{"uint nonzero → false", field(term, "height"), protoreflect.ValueOfUint32(1), false},
		{"double zero → true", field(top, "cpu"), protoreflect.ValueOfFloat64(0), true},
		{"double nonzero → false", field(top, "cpu"), protoreflect.ValueOfFloat64(1), false},
		{"bytes 无标量分支 → default false", field(term, "data"), protoreflect.ValueOfBytes([]byte{1}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isZeroScalar(tt.fd, tt.v); got != tt.want {
				t.Errorf("isZeroScalar = %v, want %v", got, tt.want)
			}
		})
	}
}

// scalarString 各 kind 序列化（纯函数直测，覆盖 encodeQuery 枚举不到的分支）。
func Test_scalarString(t *testing.T) {
	field := func(m protoreflect.Message, name string) protoreflect.FieldDescriptor {
		return m.Descriptor().Fields().ByName(protoreflect.Name(name))
	}
	create := (&websocket.CreateProjectInput{}).ProtoReflect()
	term := (&websocket.TerminalMessage{}).ProtoReflect()
	top := (&metrics.TopPodResponse{}).ProtoReflect()

	tests := []struct {
		name string
		fd   protoreflect.FieldDescriptor
		v    protoreflect.Value
		want string
	}{
		{"string url转义", field(create, "name"), protoreflect.ValueOfString("a b"), "a+b"},
		{"bool true", field(create, "atomic"), protoreflect.ValueOfBool(true), "true"},
		{"bool false", field(create, "atomic"), protoreflect.ValueOfBool(false), "false"},
		{"enum", field(create, "type"), protoreflect.ValueOfEnum(4), "4"},
		{"int", field(create, "namespace_id"), protoreflect.ValueOfInt32(5), "5"},
		{"uint", field(term, "height"), protoreflect.ValueOfUint32(24), "24"},
		{"double", field(top, "cpu"), protoreflect.ValueOfFloat64(1.5), "1.5"},
		{"bytes → default", field(term, "data"), protoreflect.ValueOfBytes([]byte{1}), ""}, // 只要不 panic
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scalarString(tt.fd, tt.v)
			if tt.want != "" && got != tt.want {
				t.Errorf("scalarString = %q, want %q", got, tt.want)
			}
		})
	}
}

// doReq 错误分支 -------------------------------------------------

// protojson.Marshal 失败：proto3 string 字段含非法 UTF-8 → Marshal 返回 ErrInvalidUTF8。
func TestDoReq_MarshalError(t *testing.T) {
	cli, err := NewHTTPClient("http://example.com")
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	req := &websocket.CreateProjectInput{Name: ptr(string([]byte{0xff}))} // 非法 UTF-8
	var resp websocket.CreateProjectInput
	if err := cli.do(context.Background(), http.MethodPost, "/api/x", req, &resp); err == nil {
		t.Fatal("want Marshal error")
	}
}

// 2xx 且 resp 为 nil：doReq 直接返回 nil（无响应绑定）。
func TestDoReq_NilResponse(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	req := &namespace.ListRequest{PageSize: ptr(int32(10))}
	if err := cli.do(context.Background(), http.MethodGet, "/api/x", req, nil); err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
}

// 2xx 但读 body 失败：io.ReadAll error。
func TestDoReq_ReadBodyError(t *testing.T) {
	cli, err := NewHTTPClient("http://example.com", WithHTTPClient(&http.Client{Transport: bodyErrTransport{}}))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	req := &namespace.ListRequest{}
	var resp namespace.ListResponse
	if err := cli.do(context.Background(), http.MethodGet, "/api/x", req, &resp); err == nil {
		t.Fatal("want ReadAll error")
	}
}

// bodyErrTransport 返回 2xx 但 body 读取即失败。
type bodyErrTransport struct{}

func (bodyErrTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(errReader{err: errors.New("body boom")}),
	}, nil
}

// 辅助：descriptorpb 指针构造。
func strPtr(s string) *string { return &s }

func int32Ptr(i int32) *int32 { return &i }

func labelOpt() *descriptorpb.FieldDescriptorProto_Label {
	v := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	return &v
}

func labelRep() *descriptorpb.FieldDescriptorProto_Label {
	v := descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	return &v
}

func typeFloat() *descriptorpb.FieldDescriptorProto_Type {
	v := descriptorpb.FieldDescriptorProto_TYPE_FLOAT
	return &v
}

func typeBool() *descriptorpb.FieldDescriptorProto_Type {
	v := descriptorpb.FieldDescriptorProto_TYPE_BOOL
	return &v
}
