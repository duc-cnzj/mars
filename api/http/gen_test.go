package http

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/http/gen"
)

// 防漂移红线：proto 一改，生成器输出必须与已提交的 rest/*.gen.http.go 一致，否则 CI 直接红。
// 改完 .proto 忘了 go generate 会在这里现形。
func TestGeneratedStubsUpToDate(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	pkgDir := filepath.Dir(thisFile)

	tmp := t.TempDir()
	if _, err := gen.Generate(tmp); err != nil {
		t.Fatalf("generate: %v", err)
	}

	// 生成产物落在 {tmp}/rest 子包，与已提交的 rest/ 目录比对。
	genDir := filepath.Join(tmp, "rest")
	committedDir := filepath.Join(pkgDir, "rest")
	entries, err := os.ReadDir(genDir)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".go" {
			continue
		}
		seen[e.Name()] = true
		fresh, err := os.ReadFile(filepath.Join(genDir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		committed, err := os.ReadFile(filepath.Join(committedDir, e.Name()))
		if err != nil {
			t.Fatalf("生成器多产出 %s 但仓库没有——请 go generate ./http/... 后提交", e.Name())
		}
		if string(fresh) != string(committed) {
			t.Errorf("%s 已过期（proto 与 stub 不一致）——请运行 go generate ./http/...", e.Name())
		}
	}

	// 反向防漂移：提交的 rest/ 里不允许存在生成器已不产出的 *.gen.http.go（孤儿）。
	// 例如 proto 删掉了某个 service，旧桩必须删除而非留在仓库里。
	committed, err := os.ReadDir(committedDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range committed {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".gen.http.go") {
			continue
		}
		if !seen[e.Name()] {
			t.Errorf("仓库有孤儿桩 %s 但生成器已不产出——请删除它（proto 里该 service 已不存在）", e.Name())
		}
	}
}

// 生成器必须清掉 rest/ 里不在本次生成集合内的 *.gen.http.go，保持目录干净。
func TestGeneratedStubsRemoveStale(t *testing.T) {
	tmp := t.TempDir()
	restDir := filepath.Join(tmp, "rest")
	if err := os.MkdirAll(restDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// 预置一个孤儿桩（模拟 proto 删 service 后遗留的文件），以及一个非生成文件（不得被删）。
	stale := filepath.Join(restDir, "deleted_service.gen.http.go")
	if err := os.WriteFile(stale, []byte("// orphan"), 0o644); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(restDir, "handwritten.go")
	if err := os.WriteFile(keep, []byte("package rest"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := gen.Generate(tmp); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("孤儿桩 %s 应被删除", stale)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Errorf("非生成文件 %s 不应被删除: %v", keep, err)
	}
	if _, err := os.Stat(filepath.Join(restDir, "namespace.gen.http.go")); err != nil {
		t.Errorf("正常桩 namespace.gen.http.go 应存在: %v", err)
	}
}
