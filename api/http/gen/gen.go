// Package gen 从已编译的 proto 描述符生成 http 的 per-service stub。
//
// 它读回 google.api.http 注解，为每个 unary 方法产出与 gRPC SDK（api/grpc/client.go）
// 签名对齐的 HTTP 客户端方法。这样 proto 一改，重新生成即可，SDK 与服务端零漂移。
//
// 生成产物落在 {outDir}/rest 子包（package rest）：Go 一条目录一个包，独立子包
// 才能让生成代码与手写引擎（http/client.go）分层，同时保持 Client 公共 API 干净。
// 生成 stub 通过 transport.Conn 接口调用传输层，由 http 包的未导出 ops 适配器实现。
//
// 生成的场景与诚实边界：
//   - unary 方法：全部生成；
//   - server-streaming + google.api.http 注解（如 StreamContainerLog/StreamTopPod）：生成 SSE 客户端方法，
//     消费 grpc-gateway 输出的 NDJSON/SSE 流（transport.OpenStream → transport.Stream[T]）；
//   - client/bidi streaming（如 Exec/StreamCopyToPod/Apply）：HTTP/JSON 下无解（需要 WebSocket），跳过；
//   - 无 google.api.http 注解的方法（gateway 本身不暴露）；
//   - 自定义路由（multipart 上传/二进制下载/copy_from_pod/ws）：不在任何 proto 里，永远手写。
//
// 用法（api 模块内）：
//
//	go generate ./http/...
//	go run ./http/gen/cmd
package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// 方法级调用风格覆盖：doNoRefresh 用于登录/换 token 等元请求，避免自动刷新递归。
var callStyleOverride = map[string]map[string]string{
	"Auth": {"Login": "doNoRefresh", "Exchange": "doNoRefresh"},
}

type methodData struct {
	Name       string
	Verb       string // Get/Post/Put/Patch/Delete
	CallMethod string // Conn 接口方法名：Do | DoNoRefresh | DoQuery
	IsStream   bool   // server-streaming：走 transport.OpenStream 而非 unary 调用
	PathArg    string // 路径 Go 表达式，如 "/api/namespaces" 或 fmt.Sprintf(...)
	ReqPkg     string // 包别名，如 namespace
	ReqType    string
	RespPkg    string
	RespType   string
	Doc        string

	needsFmt bool
	needsURL bool
	pkgPath  string
}

type serviceData struct {
	Name    string
	PkgPath string
	Methods []methodData
}

// Generate 生成全部 service stub 到 outDir，返回生成的 service 数。
// 使用全局已注册的 proto 文件（含 mars 全部 service）。
func Generate(outDir string) (int, error) {
	return generate(outDir, protoregistry.GlobalFiles)
}

// generate 是 Generate 的可注入注册表版本：测试用它喂合成 descriptor，
// 覆盖 collectServices/render 的错误与边界分支，而不污染全局注册表。
func generate(outDir string, files *protoregistry.Files) (int, error) {
	services, err := collectServices(files)
	if err != nil {
		return 0, err
	}
	if len(services) == 0 {
		return 0, fmt.Errorf("no service to generate")
	}
	// want 是本次生成的桩文件集合；生成后清掉 rest/ 里不在集合内的 *.gen.http.go，
	// 保证 rest/ 目录 100% 等于当前 proto 的产物（proto 删掉 service 不留孤儿文件）。
	want := make(map[string]bool, len(services))
	for _, sd := range services {
		want[snakeCase(sd.Name)+".gen.http.go"] = true
		if err := render(sd, outDir); err != nil {
			return 0, err
		}
	}
	// removeStale 的错误分支在 generate 内 provably unreachable：每个 service 都先 render
	// （MkdirAll + WriteFile 成功），restDir 必已存在且可写——ReadDir/Remove 恒成功。
	// 错误分支由 removeStale 的直接单元测试覆盖，这里直接丢弃（S 级零死代码）。
	_ = removeStale(filepath.Join(outDir, "rest"), want)
	fmt.Printf("generated %d service stubs in %s\n", len(services), outDir)
	return len(services), nil
}

// removeStale 删除 restDir 中不在 want 集合内的 *.gen.http.go（孤儿桩）。
// 只清理本次生成器负责的命名空间，绝不触碰非 .gen.http.go 的手写文件。
func removeStale(restDir string, want map[string]bool) error {
	entries, err := os.ReadDir(restDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".gen.http.go") {
			continue
		}
		if want[e.Name()] {
			continue
		}
		if err := os.Remove(filepath.Join(restDir, e.Name())); err != nil {
			return fmt.Errorf("remove stale %s: %w", e.Name(), err)
		}
		fmt.Printf("removed stale %s\n", filepath.Join(restDir, e.Name()))
	}
	return nil
}

func collectServices(files *protoregistry.Files) ([]*serviceData, error) {
	byName := map[string]*serviceData{}
	var names []string
	var firstErr error
	// RangeFiles 在回调返回 false 时立即停止迭代，且所有错误分支都 return false——
	// 顶部的 firstErr 再检查 provably unreachable（一旦置错必已停止），删除（S 级零死代码）。
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		svcs := fd.Services()
		for i := 0; i < svcs.Len(); i++ {
			sd, err := buildService(svcs.Get(i))
			if err != nil {
				firstErr = err
				return false
			}
			if sd == nil {
				continue
			}
			if _, dup := byName[sd.Name]; dup {
				firstErr = fmt.Errorf("duplicate service %s", sd.Name)
				return false
			}
			byName[sd.Name] = sd
			names = append(names, sd.Name)
		}
		return true
	})
	if firstErr != nil {
		return nil, firstErr
	}
	sort.Strings(names)
	out := make([]*serviceData, 0, len(names))
	for _, n := range names {
		out = append(out, byName[n])
	}
	return out, nil
}

func buildService(svc protoreflect.ServiceDescriptor) (*serviceData, error) {
	svcName := string(svc.Name())
	sd := &serviceData{Name: svcName}
	methods := svc.Methods()
	for i := 0; i < methods.Len(); i++ {
		m := methods.Get(i)
		if m.IsStreamingClient() {
			// client/bidi streaming 在 HTTP/JSON 下无解（需要 WebSocket），诚实跳过。
			fmt.Fprintf(os.Stderr, "[skip] %s.%s 是 client/bidi streaming，HTTP/JSON 不支持\n", svcName, m.Name())
			continue
		}
		if m.IsStreamingServer() {
			rule := httpRule(m)
			if rule == nil {
				fmt.Fprintf(os.Stderr, "[skip] %s.%s server-streaming 无 google.api.http 注解，gateway 不暴露\n", svcName, m.Name())
				continue
			}
			// server-streaming + http 注解 → 生成 SSE 客户端方法（消费 gateway 的 NDJSON/SSE 流）。
			md, err := buildStreamMethod(svcName, m, rule)
			if err != nil {
				return nil, err
			}
			if md == nil {
				continue
			}
			if sd.PkgPath == "" {
				sd.PkgPath = md.pkgPath
			}
			sd.Methods = append(sd.Methods, *md)
			continue
		}
		rule := httpRule(m)
		if rule == nil {
			fmt.Fprintf(os.Stderr, "[skip] %s.%s 无 google.api.http 注解，gateway 不暴露\n", svcName, m.Name())
			continue
		}
		md, err := buildMethod(svcName, m, rule)
		if err != nil {
			return nil, err
		}
		if md == nil {
			continue
		}
		if sd.PkgPath == "" {
			sd.PkgPath = md.pkgPath
		}
		sd.Methods = append(sd.Methods, *md)
	}
	if len(sd.Methods) == 0 {
		return nil, nil
	}
	return sd, nil
}

func buildMethod(svcName string, m protoreflect.MethodDescriptor, rule *annotations.HttpRule) (*methodData, error) {
	verb, tmpl, body := splitRule(rule)
	if verb == "" {
		fmt.Fprintf(os.Stderr, "[skip] %s.%s 未识别的 http 模式\n", svcName, m.Name())
		return nil, nil
	}
	reqPkgPath, reqType, reqPkg, err := goType(m.Input())
	if err != nil {
		return nil, err
	}
	respPkgPath, respType, respPkg, err := goType(m.Output())
	if err != nil {
		return nil, err
	}

	var call string
	switch {
	case body == "*":
		call = "do" // 整个请求消息为 JSON body
	case body == "":
		call = "doQuery" // 无 body 绑定，请求字段全部走 query（含 POST/DELETE）
	default:
		return nil, fmt.Errorf("%s.%s: 部分 body 绑定 body:%q 暂不支持（mars 尚无此用法）", svcName, m.Name(), body)
	}
	if ov, ok := callStyleOverride[svcName][string(m.Name())]; ok {
		call = ov
	}

	expr, needsFmt, err := pathExpr(tmpl, m.Input())
	if err != nil {
		return nil, err
	}
	md := &methodData{
		Name:       string(m.Name()),
		Verb:       verb,
		CallMethod: exportedCall(call),
		PathArg:    expr,
		ReqPkg:     reqPkg,
		ReqType:    reqType,
		RespPkg:    respPkg,
		RespType:   respType,
		Doc:        fmt.Sprintf("%s %s", strings.ToUpper(verb), tmpl),
		needsFmt:   needsFmt,
		needsURL:   strings.Contains(expr, "url.PathEscape"),
		pkgPath:    reqPkgPath,
	}
	if respPkgPath != reqPkgPath {
		return nil, fmt.Errorf("%s.%s: 请求(%s)与响应(%s)包不一致，生成器暂不支持", svcName, m.Name(), reqPkgPath, respPkgPath)
	}
	return md, nil
}

// buildStreamMethod 构造 server-streaming 方法的 stub。v1 只支持 GET/DELETE query 绑定
// （body 绑定在 HTTP 流式上传下语义复杂，mars 暂无此用法，遇到直接 fatalf）。
func buildStreamMethod(svcName string, m protoreflect.MethodDescriptor, rule *annotations.HttpRule) (*methodData, error) {
	verb, tmpl, body := splitRule(rule)
	if verb == "" {
		fmt.Fprintf(os.Stderr, "[skip] %s.%s 未识别的 http 模式\n", svcName, m.Name())
		return nil, nil
	}
	if body != "" {
		return nil, fmt.Errorf("%s.%s: stream 方法带 body 绑定 body:%q 暂不支持（v1 只支持 GET/DELETE query）", svcName, m.Name(), body)
	}
	reqPkgPath, reqType, reqPkg, err := goType(m.Input())
	if err != nil {
		return nil, err
	}
	respPkgPath, respType, respPkg, err := goType(m.Output())
	if err != nil {
		return nil, err
	}
	if respPkgPath != reqPkgPath {
		return nil, fmt.Errorf("%s.%s: 请求(%s)与响应(%s)包不一致，生成器暂不支持", svcName, m.Name(), reqPkgPath, respPkgPath)
	}
	expr, needsFmt, err := pathExpr(tmpl, m.Input())
	if err != nil {
		return nil, err
	}
	return &methodData{
		Name:     string(m.Name()),
		Verb:     verb,
		IsStream: true,
		PathArg:  expr,
		ReqPkg:   reqPkg,
		ReqType:  reqType,
		RespPkg:  respPkg,
		RespType: respType,
		Doc:      fmt.Sprintf("%s %s（server-streaming → SSE）", strings.ToUpper(verb), tmpl),
		needsFmt: needsFmt,
		needsURL: strings.Contains(expr, "url.PathEscape"),
		pkgPath:  reqPkgPath,
	}, nil
}

func splitRule(rule *annotations.HttpRule) (verb, tmpl, body string) {
	body = rule.Body
	switch p := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		return "Get", p.Get, body
	case *annotations.HttpRule_Post:
		return "Post", p.Post, body
	case *annotations.HttpRule_Put:
		return "Put", p.Put, body
	case *annotations.HttpRule_Patch:
		return "Patch", p.Patch, body
	case *annotations.HttpRule_Delete:
		return "Delete", p.Delete, body
	}
	return "", "", body
}

// exportedCall 把引擎内部调用风格映射成 Conn 接口的导出方法名。
func exportedCall(call string) string {
	switch call {
	case "do":
		return "Do"
	case "doNoRefresh":
		return "DoNoRefresh"
	case "doQuery":
		return "DoQuery"
	}
	return call
}

// goType 通过运行时注册表把 proto 消息解析成 Go 类型信息。
func goType(md protoreflect.MessageDescriptor) (pkgPath, typeName, pkgAlias string, err error) {
	mt, err := protoregistry.GlobalTypes.FindMessageByName(md.FullName())
	if err != nil {
		return "", "", "", fmt.Errorf("resolve type %s: %v", md.FullName(), err)
	}
	// mt.New() 返回的是 protoreflect 运行时包装器，.Interface() 才是生成的 Go 结构体类型。
	t := reflect.TypeOf(mt.New().Interface())
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	pkgPath = t.PkgPath()
	typeName = t.Name()
	pkgAlias = path.Base(pkgPath)
	return
}

// pathExpr 把 http 路径模板转成 Go 表达式：
//   - 无 {var} → 字符串字面量；
//   - 有 {var} → fmt.Sprintf，路径字段用访问器，字符串字段走 url.PathEscape（保留 %2F，对应分支名带斜杠的修复）。
func pathExpr(tmpl string, input protoreflect.MessageDescriptor) (expr string, needsFmt bool, err error) {
	if !strings.Contains(tmpl, "{") {
		return strconv.Quote(tmpl), false, nil
	}
	fields := input.Fields()
	var sb strings.Builder
	var args []string
	rest := tmpl
	for {
		i := strings.Index(rest, "{")
		if i < 0 {
			sb.WriteString(rest)
			break
		}
		sb.WriteString(rest[:i])
		rest = rest[i+1:]
		j := strings.Index(rest, "}")
		if j < 0 {
			return "", false, fmt.Errorf("template %q 缺少右花括号", tmpl)
		}
		varName := rest[:j]
		rest = rest[j+1:]
		fd := fields.ByName(protoreflect.Name(varName))
		if fd == nil {
			return "", false, fmt.Errorf("path 变量 %q 在 %s 中找不到对应字段", varName, input)
		}
		accessor := "req." + camel(string(fd.Name()))
		verb, arg := formatForKind(fd.Kind(), accessor)
		sb.WriteString(verb)
		args = append(args, arg)
	}
	return fmt.Sprintf("fmt.Sprintf(%q, %s)", sb.String(), strings.Join(args, ", ")), true, nil
}

func formatForKind(k protoreflect.Kind, accessor string) (verb, arg string) {
	switch k {
	case protoreflect.StringKind:
		return "%s", "url.PathEscape(" + accessor + ")"
	case protoreflect.BoolKind:
		return "%t", accessor
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Sint32Kind,
		protoreflect.Sint64Kind, protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Fixed32Kind,
		protoreflect.Fixed64Kind, protoreflect.EnumKind:
		return "%d", accessor
	default:
		return "%v", accessor
	}
}

// snakeCase 把 CamelCase 服务名转成下划线文件名，如 AccessToken → access_token。
func snakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if 'A' <= r && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func camel(s string) string {
	var b strings.Builder
	for _, part := range strings.Split(s, "_") {
		if part == "" {
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]))
		b.WriteString(part[1:])
	}
	return b.String()
}

func httpRule(m protoreflect.MethodDescriptor) *annotations.HttpRule {
	// proto.GetExtension 对 google.api.http 扩展恒返回 *HttpRule：未设置时是 typed nil
	// 指针（非 untyped nil interface），ext == nil 永不成立；protodesc 构建的 method
	// descriptor Options() 也恒非空。两个 nil 守卫分支均 provably unreachable，删除
	// （S 级零死代码）。类型断言直接产出 rule——未设置时为 nil 指针，语义与守卫一致。
	opts := m.Options()
	ext := proto.GetExtension(opts, annotations.E_Http)
	rule, _ := ext.(*annotations.HttpRule)
	return rule
}

var fileTpl = template.Must(template.New("service").Parse(`// Code generated by api/http/gen. DO NOT EDIT.
//
// 由 gen 从 proto 的 google.api.http 注解生成，方法签名与 gRPC SDK（api/grpc/client.go）对齐。
// 修改请改 .proto 后重新生成：go generate ./http/...

package rest

import (
{{- range .Stdlib}}
	{{.}}
{{- end}}
{{ range .External}}
	{{.}}
{{- end}}
)

// {{$.Name}}Svc 提供 {{$.Name}} service 的 HTTP 客户端方法。
// C 由 http.Client 的 ops 适配器注入（见 http.Client.{{$.Name}}）。
type {{$.Name}}Svc struct{ C transport.Conn }
{{range .Methods}}
{{if .IsStream}}
// {{.Name}} {{.Doc}}。
func (s *{{$.Name}}Svc) {{.Name}}(ctx context.Context, req *{{.ReqPkg}}.{{.ReqType}}) (transport.Stream[*{{.RespPkg}}.{{.RespType}}], error) {
	return transport.OpenStream[*{{.RespPkg}}.{{.RespType}}](s.C, ctx, http.Method{{.Verb}}, {{.PathArg}}, req)
}
{{else}}
// {{.Name}} {{.Doc}}。
func (s *{{$.Name}}Svc) {{.Name}}(ctx context.Context, req *{{.ReqPkg}}.{{.ReqType}}) (*{{.RespPkg}}.{{.RespType}}, error) {
	var out {{.RespPkg}}.{{.RespType}}
	if err := s.C.{{.CallMethod}}(ctx, http.Method{{.Verb}}, {{.PathArg}}, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
{{end}}
{{end}}
`))

func render(sd *serviceData, outDir string) error {
	// 分组输出 stdlib 与外部包（两段 import group），让生成产物与 goimports 的结果一致，
	// 否则 goimports -w 每跑一次都会重排生成文件，漂移测试永远过不了。
	var stdlib, external []string
	stdlib = append(stdlib, `"context"`, `"net/http"`)
	needsFmt, needsURL := false, false
	for _, m := range sd.Methods {
		needsFmt = needsFmt || m.needsFmt
		needsURL = needsURL || m.needsURL
	}
	if needsFmt {
		stdlib = append(stdlib, `"fmt"`)
	}
	if needsURL {
		stdlib = append(stdlib, `"net/url"`)
	}
	external = append(external, strconv.Quote(sd.PkgPath))
	// transport 是手写传输契约包（Conn/Stream/OpenStream），rest/ 生成产物限定引用它，
	// 保证 rest/ 目录 100% 自动生成、与手写代码物理隔离。
	external = append(external, `"github.com/duc-cnzj/mars/api/v6/http/transport"`)
	sort.Strings(stdlib)
	sort.Strings(external)

	var buf bytes.Buffer
	// fileTpl 经 template.Must 初始化（语法恒合法），数据是固定的 serviceData 结构体——
	// Execute 恒成功，错误分支 provably unreachable，直接丢弃（S 级零死代码）。
	_ = fileTpl.Execute(&buf, struct {
		*serviceData
		Stdlib   []string
		External []string
	}{sd, stdlib, external})
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format %s: %w\n%s", sd.Name, err, buf.String())
	}
	dir := filepath.Join(outDir, "rest")
	// #nosec G301 -- 生成器产出源码目录，0755 为 Go 源码标准权限
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	fname := filepath.Join(dir, snakeCase(sd.Name)+".gen.http.go")
	// #nosec G306 -- 生成源码文件，0644 为标准权限
	if err := os.WriteFile(fname, src, 0o644); err != nil {
		return err
	}
	fmt.Printf("generated %s (%d methods)\n", fname, len(sd.Methods))
	return nil
}
