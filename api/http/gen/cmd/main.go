// Command cmd 生成 http 的 per-service stub（命令行入口）。
//
// 用法（api 模块内）：
//
//	go run ./http/gen/cmd
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/duc-cnzj/mars/api/v6/http/gen"
)

func main() {
	// 默认输出到本文件所在目录的上两层（http 包），不依赖 CWD，无论从哪里调用都稳。
	_, thisFile, _, _ := runtime.Caller(0)
	defOut := filepath.Join(filepath.Dir(thisFile), "..", "..")
	outDir := flag.String("out", defOut, "输出目录")
	flag.Parse()

	if _, err := gen.Generate(*outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
