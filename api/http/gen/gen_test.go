package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// 本文件覆盖 gen.go 的全部错误/边界分支。generate 的可注入注册表（files 参数）喂合成
// descriptor，避免污染 protoregistry.GlobalFiles；合成消息以动态类型注册进
// GlobalTypes，goType() 解析它们为 dynamicpb 包。每个测试用独立 proto package，
// 避免 GlobalTypes 重名冲突（同一测试二进制内全局注册表共享）。

func sp(s string) *string { return &s }
func bp(b bool) *bool     { return &b }
func ip(i int32) *int32   { return &i }

// newFile 构造一个 proto3 的 FileDescriptorProto，可带外部依赖（按文件路径）。
func newFile(name, pkg string, deps ...string) *descriptorpb.FileDescriptorProto {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    sp(name),
		Package: sp(pkg),
		Syntax:  sp("proto3"),
	}
	fd.Dependency = append(fd.Dependency, deps...)
	return fd
}

// msg 构造一个消息（scalar 字段走 descriptorpb 的 field 列表）。
func msg(name string, fields ...*descriptorpb.FieldDescriptorProto) *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{Name: sp(name), Field: fields}
}

// fld 构造一个 scalar 字段，JsonName 取 proto 默认 camelCase（protodesc 校验一致性用）。
func fld(name string, num int32, typ descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:     sp(name),
		JsonName: sp(defaultJSONName(name)),
		Number:   ip(num),
		Type:     typ.Enum(),
	}
}

// defaultJSONName 复刻 protobuf 默认 json_name（下划线转 camelCase），仅供测试辅助使用。
func defaultJSONName(s string) string {
	var b strings.Builder
	up := false
	for _, r := range s {
		if r == '_' {
			up = true
			continue
		}
		if up {
			b.WriteRune(r - ('a' - 'A'))
			up = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// method 构造一个方法，带可选的 http 注解与 streaming 标记。
func method(name, in, out string, clientStream, serverStream bool, rule *annotations.HttpRule) *descriptorpb.MethodDescriptorProto {
	m := &descriptorpb.MethodDescriptorProto{
		Name:       sp(name),
		InputType:  sp(in),
		OutputType: sp(out),
	}
	if clientStream {
		m.ClientStreaming = bp(true)
	}
	if serverStream {
		m.ServerStreaming = bp(true)
	}
	if rule != nil {
		opts := &descriptorpb.MethodOptions{}
		// E_Http 的扩展体是 *HttpRule，SetExtension 恒成功。
		proto.SetExtension(opts, annotations.E_Http, rule)
		m.Options = opts
	}
	return m
}

// svc 把 service（含若干方法）与一个 Req/Resp 消息对挂到文件上。
// 同一文件挂多个 service 时消息对只加一次（protodesc 不允许重名消息）。
func svc(fd *descriptorpb.FileDescriptorProto, svcName string, methods ...*descriptorpb.MethodDescriptorProto) {
	has := false
	for _, m := range fd.MessageType {
		if m.GetName() == "Req" {
			has = true
			break
		}
	}
	if !has {
		fd.MessageType = append(fd.MessageType,
			msg("Req", fld("name", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)),
			msg("Resp", fld("data", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)),
		)
	}
	fd.Service = append(fd.Service, &descriptorpb.ServiceDescriptorProto{
		Name:   sp(svcName),
		Method: methods,
	})
}

// serviceFile 一步构造：文件 + Req/Resp 消息 + 一个 service。方法内部引用
// ".pkg.Req"/".pkg.Resp"（self-contained，protodesc 自解析）。
func serviceFile(path, pkg, svcName string, methods ...*descriptorpb.MethodDescriptorProto) *descriptorpb.FileDescriptorProto {
	fd := newFile(path, pkg)
	svc(fd, svcName, methods...)
	return fd
}

// goodMethod 生成一个标准可生成方法：GET "/v1/foos/{name}"，body 不绑定 → doQuery。
func goodMethod(pkg, name string) *descriptorpb.MethodDescriptorProto {
	return method(name, "."+pkg+".Req", "."+pkg+".Resp", false, false,
		&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/foos/{name}"}})
}

// buildFiles 把合成 FileDescriptorProto 转成 *protoregistry.Files，并把文件内消息注册成
// 动态类型（GlobalTypes），让 goType() 能解析它们。protodesc.NewFile 的 resolver 用
// GlobalFiles，外部依赖（如 proto/namespace/namespace.proto）可正常解析。
func buildFiles(t *testing.T, fds ...*descriptorpb.FileDescriptorProto) *protoregistry.Files {
	t.Helper()
	reg := &protoregistry.Files{}
	for _, fd := range fds {
		f, err := protodesc.NewFile(fd, protoregistry.GlobalFiles)
		if err != nil {
			t.Fatalf("protodesc.NewFile(%s): %v", fd.GetName(), err)
		}
		if err := reg.RegisterFile(f); err != nil {
			t.Fatalf("RegisterFile(%s): %v", fd.GetName(), err)
		}
		registerDyn(t, f.Messages())
	}
	return reg
}

func registerDyn(t *testing.T, msgs protoreflect.MessageDescriptors) {
	t.Helper()
	for i := 0; i < msgs.Len(); i++ {
		md := msgs.Get(i)
		if err := protoregistry.GlobalTypes.RegisterMessage(dynamicpb.NewMessageType(md)); err != nil {
			t.Fatalf("RegisterMessage(%s): %v", md.FullName(), err)
		}
		registerDyn(t, md.Messages())
	}
}

// buildFile 只构建并注册文件（不注册消息类型），用于 goType 解析失败类测试——
// 测试选择性把某个消息注册进 GlobalTypes，制造「输入可解析/输出不可解析」的错位。
func buildFile(t *testing.T, fd *descriptorpb.FileDescriptorProto) *protoregistry.Files {
	t.Helper()
	f, err := protodesc.NewFile(fd, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatalf("protodesc.NewFile(%s): %v", fd.GetName(), err)
	}
	reg := &protoregistry.Files{}
	if err := reg.RegisterFile(f); err != nil {
		t.Fatalf("RegisterFile(%s): %v", fd.GetName(), err)
	}
	return reg
}

// generate 空注册表 → "no service to generate"。
func TestGenerate_NoServices(t *testing.T) {
	_, err := generate(t.TempDir(), &protoregistry.Files{})
	if err == nil || !strings.Contains(err.Error(), "no service to generate") {
		t.Fatalf("want no-service error, got %v", err)
	}
}

// 两个不同 package 的 service 同名 → collectServices 报 duplicate。
func TestGenerate_DuplicateService(t *testing.T) {
	reg := buildFiles(t,
		serviceFile("gen_dup_a.proto", "gen.dupa", "Foo", goodMethod("gen.dupa", "GetA")),
		serviceFile("gen_dup_b.proto", "gen.dupb", "Foo", goodMethod("gen.dupb", "GetB")),
	)
	_, err := generate(t.TempDir(), reg)
	if err == nil || !strings.Contains(err.Error(), "duplicate service Foo") {
		t.Fatalf("want duplicate-service error, got %v", err)
	}
}

// unary 方法部分 body 绑定 → 生成器显式报错（mars 尚无此用法）。
func TestGenerate_PartialBody(t *testing.T) {
	fd := serviceFile("gen_body.proto", "gen.body", "Foo",
		method("Create", ".gen.body.Req", ".gen.body.Resp", false, false,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Post{Post: "/v1/foos"}, Body: "name"}))
	_, err := generate(t.TempDir(), buildFiles(t, fd))
	if err == nil || !strings.Contains(err.Error(), "部分 body 绑定") {
		t.Fatalf("want partial-body error, got %v", err)
	}
}

// server-streaming 方法带 body 绑定 → 显式报错（v1 只支持 GET/DELETE query）。
func TestGenerate_StreamBody(t *testing.T) {
	fd := serviceFile("gen_sbody.proto", "gen.sbody", "Foo",
		method("Watch", ".gen.sbody.Req", ".gen.sbody.Resp", false, true,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/foos"}, Body: "name"}))
	_, err := generate(t.TempDir(), buildFiles(t, fd))
	if err == nil || !strings.Contains(err.Error(), "stream 方法带 body 绑定") {
		t.Fatalf("want stream-body error, got %v", err)
	}
}

// 请求是动态类型（dynamicpb 包）、响应是真实 mars 消息（namespace 包）→ 包不一致。
func TestGenerate_PkgMismatch(t *testing.T) {
	fd := newFile("gen_mm.proto", "gen.mm", "proto/namespace/namespace.proto")
	fd.MessageType = append(fd.MessageType, msg("Req", fld("name", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)))
	fd.Service = append(fd.Service, &descriptorpb.ServiceDescriptorProto{
		Name: sp("Foo"),
		Method: []*descriptorpb.MethodDescriptorProto{
			method("Get", ".gen.mm.Req", ".namespace.ListRequest", false, false,
				&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/x"}}),
		},
	})
	_, err := generate(t.TempDir(), buildFiles(t, fd))
	if err == nil || !strings.Contains(err.Error(), "包不一致") {
		t.Fatalf("want pkg-mismatch error, got %v", err)
	}
}

// server-streaming 的包不一致分支。
func TestGenerate_StreamPkgMismatch(t *testing.T) {
	fd := newFile("gen_smm.proto", "gen.smm", "proto/namespace/namespace.proto")
	fd.MessageType = append(fd.MessageType, msg("Req", fld("name", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)))
	fd.Service = append(fd.Service, &descriptorpb.ServiceDescriptorProto{
		Name: sp("Foo"),
		Method: []*descriptorpb.MethodDescriptorProto{
			method("Watch", ".gen.smm.Req", ".namespace.ListRequest", false, true,
				&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/x"}}),
		},
	})
	_, err := generate(t.TempDir(), buildFiles(t, fd))
	if err == nil || !strings.Contains(err.Error(), "包不一致") {
		t.Fatalf("want stream pkg-mismatch error, got %v", err)
	}
}

// 方法名 "func" 是合法 proto 标识符但非法 Go 方法名 → format.Source 报错。
func TestGenerate_FormatError(t *testing.T) {
	fd := serviceFile("gen_fn.proto", "gen.fn", "Foo",
		method("func", ".gen.fn.Req", ".gen.fn.Resp", false, false,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/x"}}))
	_, err := generate(t.TempDir(), buildFiles(t, fd))
	if err == nil || !strings.Contains(err.Error(), "format Foo") {
		t.Fatalf("want format error, got %v", err)
	}
}

// outDir/rest 已存在同名文件 → MkdirAll 失败。
func TestGenerate_RenderMkdirErr(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "rest"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fd := serviceFile("gen_mk.proto", "gen.mk", "Foo", goodMethod("gen.mk", "Get"))
	_, err := generate(tmp, buildFiles(t, fd))
	if err == nil {
		t.Fatal("want mkdir error")
	}
}

// rest 目录只读 → WriteFile 失败。
func TestGenerate_RenderWriteErr(t *testing.T) {
	tmp := t.TempDir()
	restDir := filepath.Join(tmp, "rest")
	if err := os.MkdirAll(restDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(restDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(restDir, 0o755) })

	fd := serviceFile("gen_wr.proto", "gen.wr", "Foo", goodMethod("gen.wr", "Get"))
	_, err := generate(tmp, buildFiles(t, fd))
	if err == nil {
		t.Fatal("want write error")
	}
}

// 端到端快乐路径：unary(doQuery + do) + SSE(OpenStream) + 各类 skip 分支。
func TestGenerate_EndToEnd(t *testing.T) {
	fd := serviceFile("gen_e2e.proto", "gen.e2e", "FooService",
		// GET + body 空 → doQuery；路径含 string 字段 → needsFmt + needsURL。
		method("Get", ".gen.e2e.Req", ".gen.e2e.Resp", false, false,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/foos/{name}"}}),
		// POST + body "*" → do。
		method("Create", ".gen.e2e.Req", ".gen.e2e.Resp", false, false,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Post{Post: "/v1/foos"}, Body: "*"}),
		// server-streaming + http 注解 → SSE（OpenStream）。
		method("Watch", ".gen.e2e.Req", ".gen.e2e.Resp", false, true,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/foos/{name}"}}),
		// client-streaming → 诚实跳过。
		method("ClientPut", ".gen.e2e.Req", ".gen.e2e.Resp", true, false, nil),
		// unary 无 http 注解 → httpRule ext==nil → 跳过。
		method("NoRule", ".gen.e2e.Req", ".gen.e2e.Resp", false, false, nil),
		// options 非空但无 http 扩展 → httpRule ext==nil → 跳过。
		&descriptorpb.MethodDescriptorProto{Name: sp("NoExt"), InputType: sp(".gen.e2e.Req"), OutputType: sp(".gen.e2e.Resp"), Options: &descriptorpb.MethodOptions{}},
		// unary 空 rule（无 pattern）→ splitRule 默认分支 → 跳过。
		method("BadRule", ".gen.e2e.Req", ".gen.e2e.Resp", false, false, &annotations.HttpRule{}),
		// server-streaming 无 http 注解 → buildService rule==nil → 跳过。
		method("WatchNoRule", ".gen.e2e.Req", ".gen.e2e.Resp", false, true, nil),
		// server-streaming 空 rule → buildStreamMethod verb=="" → 跳过。
		method("WatchBadRule", ".gen.e2e.Req", ".gen.e2e.Resp", false, true, &annotations.HttpRule{}),
	)
	// 第二个 service 只有被跳过的方法 → buildService 返回 nil，collectServices continue。
	svc(fd, "Skipped", method("OnlySkip", ".gen.e2e.Req", ".gen.e2e.Resp", false, false, nil))

	tmp := t.TempDir()
	n, err := generate(tmp, buildFiles(t, fd))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1 service generated, got %d", n)
	}
	// FooService → foo_service.gen.http.go（snakeCase 覆盖中间大写加下划线分支）。
	out, err := os.ReadFile(filepath.Join(tmp, "rest", "foo_service.gen.http.go"))
	if err != nil {
		t.Fatalf("read stub: %v", err)
	}
	src := string(out)
	for _, want := range []string{"DoQuery", "OpenStream", "Do(",
		"net/url", "dynamicpb.Message", "transport.Conn"} {
		if !strings.Contains(src, want) {
			t.Errorf("stub 缺少 %s\n---\n%s", want, src)
		}
	}
	for _, banned := range []string{"ClientPut", "NoRule", "NoExt", "BadRule", "WatchNoRule", "WatchBadRule", "Skipped"} {
		if strings.Contains(src, banned) {
			t.Errorf("stub 不应包含被跳过的方法 %s", banned)
		}
	}
	if _, err := os.Stat(filepath.Join(tmp, "rest", "skipped.gen.http.go")); !os.IsNotExist(err) {
		t.Error("Skipped service 不应产出桩文件")
	}
}

// callStyleOverride：Auth.Login → doNoRefresh。
func TestGenerate_CallStyleOverride(t *testing.T) {
	fd := serviceFile("gen_ov.proto", "gen.ov", "Auth",
		method("Login", ".gen.ov.Req", ".gen.ov.Resp", false, false,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Post{Post: "/v1/login"}, Body: "*"}))
	tmp := t.TempDir()
	if _, err := generate(tmp, buildFiles(t, fd)); err != nil {
		t.Fatalf("generate: %v", err)
	}
	src, err := os.ReadFile(filepath.Join(tmp, "rest", "auth.gen.http.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "DoNoRefresh") {
		t.Errorf("Auth.Login 应生成 DoNoRefresh 调用\n%s", src)
	}
}

// removeStale 三态：目录不存在→nil；路径是文件→ReadDir 错误；只读目录→Remove 错误。
func TestRemoveStale_NotExist(t *testing.T) {
	if err := removeStale(filepath.Join(t.TempDir(), "nope"), map[string]bool{}); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func TestRemoveStale_ReadError(t *testing.T) {
	f := filepath.Join(t.TempDir(), "notdir")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := removeStale(f, map[string]bool{}); err == nil {
		t.Fatal("want read error")
	}
}

func TestRemoveStale_RemoveError(t *testing.T) {
	restDir := filepath.Join(t.TempDir(), "rest")
	if err := os.MkdirAll(restDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(restDir, "orphan.gen.http.go"), []byte("// x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(restDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(restDir, 0o755) })
	err := removeStale(restDir, map[string]bool{})
	if err == nil || !strings.Contains(err.Error(), "remove stale") {
		t.Fatalf("want remove-stale error, got %v", err)
	}
}

// removeStale 快乐路径：孤儿被删、want 内保留、手写文件不受影响。
func TestRemoveStale_RemoveSuccess(t *testing.T) {
	restDir := filepath.Join(t.TempDir(), "rest")
	if err := os.MkdirAll(restDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"orphan.gen.http.go", "keep.gen.http.go", "handwritten.go"} {
		if err := os.WriteFile(filepath.Join(restDir, name), []byte("// x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// 目录条目不应被当成孤儿删除。
	if err := os.Mkdir(filepath.Join(restDir, "subdir.gen.http.go"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := removeStale(restDir, map[string]bool{"keep.gen.http.go": true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(restDir, "orphan.gen.http.go")); !os.IsNotExist(err) {
		t.Error("orphan 应被删除")
	}
	if _, err := os.Stat(filepath.Join(restDir, "keep.gen.http.go")); err != nil {
		t.Error("keep 应保留")
	}
	if _, err := os.Stat(filepath.Join(restDir, "handwritten.go")); err != nil {
		t.Error("手写文件应保留")
	}
	if _, err := os.Stat(filepath.Join(restDir, "subdir.gen.http.go")); err != nil {
		t.Error("目录条目应被跳过")
	}
}

// exportedCall 未识别风格 → 原样透传。
func TestExportedCall_Unknown(t *testing.T) {
	if got := exportedCall("custom"); got != "custom" {
		t.Fatalf("want passthrough, got %s", got)
	}
}

// formatForKind 的 %t 与默认 %v 分支（%s/%d 由快乐路径覆盖）。
func TestFormatForKind(t *testing.T) {
	cases := []struct {
		kind protoreflect.Kind
		want string
	}{
		{protoreflect.BoolKind, "%t"},
		{protoreflect.Int32Kind, "%d"},
		{protoreflect.EnumKind, "%d"},
		{protoreflect.DoubleKind, "%v"},
	}
	for _, c := range cases {
		if verb, _ := formatForKind(c.kind, "req.F"); verb != c.want {
			t.Errorf("formatForKind(%v) = %q, want %q", c.kind, verb, c.want)
		}
	}
}

// camel 对连续下划线字段名跳过空 part（foo__bar → FooBar）。
func TestPathExpr_CamelEmptyPart(t *testing.T) {
	fd := newFile("gen_camel.proto", "gen.camel")
	fd.MessageType = append(fd.MessageType,
		msg("Req", fld("foo__bar", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)))
	md := buildFiles(t, fd)
	d, err := md.FindDescriptorByName("gen.camel.Req")
	if err != nil {
		t.Fatal(err)
	}
	req := d.(protoreflect.MessageDescriptor)

	expr, needsFmt, err := pathExpr("/v1/{foo__bar}", req)
	if err != nil {
		t.Fatal(err)
	}
	if !needsFmt || !strings.Contains(expr, "req.FooBar") {
		t.Fatalf("want camel empty-part 跳过, got %q (needsFmt=%v)", expr, needsFmt)
	}
}

// pathExpr 两个错误分支：缺右花括号 + 字段不存在。用真实 mars 消息避免合成开销。
func TestPathExpr_Errors(t *testing.T) {
	req := (&namespace.ListRequest{}).ProtoReflect().Descriptor()
	if _, _, err := pathExpr("/v1/{name", req); err == nil || !strings.Contains(err.Error(), "缺少右花括号") {
		t.Fatalf("want missing-brace error, got %v", err)
	}
	if _, _, err := pathExpr("/v1/{nonexistent}", req); err == nil || !strings.Contains(err.Error(), "找不到对应字段") {
		t.Fatalf("want field-not-found error, got %v", err)
	}
}

// goType 解析不到消息类型（不注册进 GlobalTypes）→ resolve error。
func TestGoType_ResolveError(t *testing.T) {
	fd := serviceFile("gen_gt.proto", "gen.gt", "Foo", goodMethod("gen.gt", "Get"))
	f, err := protodesc.NewFile(fd, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	md := f.Messages().Get(0) // Req，故意不注册进 GlobalTypes
	if _, _, _, err := goType(md); err == nil || !strings.Contains(err.Error(), "resolve type") {
		t.Fatalf("want resolve-type error, got %v", err)
	}
}

// Generate（导出包装）走全局注册表：真实 mars 全部 service 都能生成，覆盖 90-91 行。
func TestGenerate_GlobalWrapper(t *testing.T) {
	n, err := Generate(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("want >0 services from global registry")
	}
}

// 只含 server-streaming 方法的 service：首次生成即走 streaming 分支的 PkgPath 赋值。
func TestGenerate_StreamOnly(t *testing.T) {
	fd := serviceFile("gen_so.proto", "gen.so", "WatchService",
		method("Watch", ".gen.so.Req", ".gen.so.Resp", false, true,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/foos/{name}"}}))
	tmp := t.TempDir()
	n, err := generate(tmp, buildFiles(t, fd))
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("want 1 service, got %d", n)
	}
	src, err := os.ReadFile(filepath.Join(tmp, "rest", "watch_service.gen.http.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "OpenStream") {
		t.Errorf("SSE stub 应含 OpenStream\n%s", src)
	}
}

// buildMethod 的 goType(input) 解析失败（消息未注册进 GlobalTypes）。
func TestGenerate_GoTypeInputErr(t *testing.T) {
	fd := serviceFile("gen_gin.proto", "gen.gin", "Foo",
		method("Get", ".gen.gin.Req", ".gen.gin.Resp", false, false,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/x"}}))
	_, err := generate(t.TempDir(), buildFile(t, fd))
	if err == nil || !strings.Contains(err.Error(), "resolve type") {
		t.Fatalf("want resolve-type error, got %v", err)
	}
}

// buildMethod 的 goType(output) 解析失败：只注册 Req，Resp 未注册。
func TestGenerate_GoTypeOutputErr(t *testing.T) {
	fd := serviceFile("gen_gout.proto", "gen.gout", "Foo",
		method("Get", ".gen.gout.Req", ".gen.gout.Resp", false, false,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/x"}}))
	reg := buildFile(t, fd)
	f, err := protodesc.NewFile(fd, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	// 只注册 Req（Messages().Get(0)），Resp 不注册 → 输入可解析、输出不可解析。
	if err := protoregistry.GlobalTypes.RegisterMessage(dynamicpb.NewMessageType(f.Messages().Get(0))); err != nil {
		t.Fatal(err)
	}
	_, err = generate(t.TempDir(), reg)
	if err == nil || !strings.Contains(err.Error(), "resolve type") {
		t.Fatalf("want resolve-type error, got %v", err)
	}
}

// buildMethod 的 pathExpr 失败：类型可解析但路径模板引用不存在的字段。
func TestGenerate_PathExprErr(t *testing.T) {
	fd := serviceFile("gen_pe.proto", "gen.pe", "Foo",
		method("Get", ".gen.pe.Req", ".gen.pe.Resp", false, false,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/{nofield}"}}))
	_, err := generate(t.TempDir(), buildFiles(t, fd))
	if err == nil || !strings.Contains(err.Error(), "找不到对应字段") {
		t.Fatalf("want field-not-found error, got %v", err)
	}
}

// buildStreamMethod 的 goType(input) 解析失败。
func TestGenerate_StreamGoTypeInputErr(t *testing.T) {
	fd := serviceFile("gen_sgin.proto", "gen.sgin", "Foo",
		method("Watch", ".gen.sgin.Req", ".gen.sgin.Resp", false, true,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/x"}}))
	_, err := generate(t.TempDir(), buildFile(t, fd))
	if err == nil || !strings.Contains(err.Error(), "resolve type") {
		t.Fatalf("want resolve-type error, got %v", err)
	}
}

// buildStreamMethod 的 goType(output) 解析失败。
func TestGenerate_StreamGoTypeOutputErr(t *testing.T) {
	fd := serviceFile("gen_sgout.proto", "gen.sgout", "Foo",
		method("Watch", ".gen.sgout.Req", ".gen.sgout.Resp", false, true,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/x"}}))
	reg := buildFile(t, fd)
	f, err := protodesc.NewFile(fd, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	if err := protoregistry.GlobalTypes.RegisterMessage(dynamicpb.NewMessageType(f.Messages().Get(0))); err != nil {
		t.Fatal(err)
	}
	_, err = generate(t.TempDir(), reg)
	if err == nil || !strings.Contains(err.Error(), "resolve type") {
		t.Fatalf("want resolve-type error, got %v", err)
	}
}

// buildStreamMethod 的 pathExpr 失败。
func TestGenerate_StreamPathExprErr(t *testing.T) {
	fd := serviceFile("gen_spe.proto", "gen.spe", "Foo",
		method("Watch", ".gen.spe.Req", ".gen.spe.Resp", false, true,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/{nofield}"}}))
	_, err := generate(t.TempDir(), buildFiles(t, fd))
	if err == nil || !strings.Contains(err.Error(), "找不到对应字段") {
		t.Fatalf("want field-not-found error, got %v", err)
	}
}

// splitRule 全部动词分支（Get/Post/Put/Patch/Delete），默认分支由 EndToEnd 的空 rule 覆盖。
func TestSplitRule_AllVerbs(t *testing.T) {
	cases := []struct {
		rule *annotations.HttpRule
		want string
	}{
		{&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/g"}}, "Get"},
		{&annotations.HttpRule{Pattern: &annotations.HttpRule_Post{Post: "/p"}}, "Post"},
		{&annotations.HttpRule{Pattern: &annotations.HttpRule_Put{Put: "/u"}}, "Put"},
		{&annotations.HttpRule{Pattern: &annotations.HttpRule_Patch{Patch: "/h"}}, "Patch"},
		{&annotations.HttpRule{Pattern: &annotations.HttpRule_Delete{Delete: "/d"}}, "Delete"},
	}
	for _, c := range cases {
		if verb, _, _ := splitRule(c.rule); verb != c.want {
			t.Errorf("splitRule(%v) = %q, want %q", c.rule, verb, c.want)
		}
	}
}

// httpRule 有扩展 → 返回 rule；无扩展（空 options）→ nil。
func TestHttpRule_HasExt(t *testing.T) {
	fd := serviceFile("gen_hr.proto", "gen.hr", "Foo",
		method("Get", ".gen.hr.Req", ".gen.hr.Resp", false, false,
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/x"}}))
	f, err := protodesc.NewFile(fd, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	m := f.Services().Get(0).Methods().Get(0)
	if rule := httpRule(m); rule == nil {
		t.Fatal("want non-nil rule")
	}
}

func TestHttpRule_NilExt(t *testing.T) {
	fd := serviceFile("gen_hrn.proto", "gen.hrn", "Foo",
		&descriptorpb.MethodDescriptorProto{Name: sp("NoExt"), InputType: sp(".gen.hrn.Req"), OutputType: sp(".gen.hrn.Resp"), Options: &descriptorpb.MethodOptions{}})
	f, err := protodesc.NewFile(fd, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	m := f.Services().Get(0).Methods().Get(0)
	if rule := httpRule(m); rule != nil {
		t.Fatal("want nil rule")
	}
}
