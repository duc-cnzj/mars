package services

import (
	"bufio"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzToValidUTF8String 属性：任意字节输入，输出必须是合法 UTF-8（gRPC 序列化前提）。
func FuzzToValidUTF8String(f *testing.F) {
	f.Add([]byte("hello"))
	f.Add([]byte{0xff, 0xfe, 0x00})
	f.Add([]byte{})
	f.Add([]byte("中文\xff测试"))
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	f.Fuzz(func(t *testing.T, b []byte) {
		result := toValidUTF8String(b)
		if !utf8.ValidString(result) {
			t.Fatalf("toValidUTF8String(%q) 输出非法 UTF-8: %q", b, result)
		}
	})
}

// FuzzScannerText 属性：任意输入不 panic，且返回错误只能是 nil 或 bufio.ErrTooLong
// （单行超过 8MB 缓冲上限时合法返回 ErrTooLong，不得 panic）。
func FuzzScannerText(f *testing.F) {
	f.Add("")
	f.Add("a\nb\nc")
	f.Add("no-newline")
	f.Add(strings.Repeat("x", 10*1024*1024))
	f.Fuzz(func(t *testing.T, text string) {
		got := 0
		err := scannerText(text, func(string) { got++ })
		if err != nil && !errors.Is(err, bufio.ErrTooLong) {
			t.Fatalf("scannerText 返回意外错误: %v", err)
		}
		if got == 0 && text != "" && len(text) <= 8*1024*1024 {
			// 输入非空且行未超上限，至少应产出 1 行
			t.Fatalf("scannerText(%q) 未产出任何行", text)
		}
	})
}

// FuzzMaskToken 属性：len<=8 恒为 ******；len>8 首尾 4 位必须保留。
func FuzzMaskToken(f *testing.F) {
	f.Add("")
	f.Add("12345678")
	f.Add("123456789012")
	f.Add("abcdefghijklmnopqrstuvwxyz")
	f.Fuzz(func(t *testing.T, token string) {
		masked := maskToken(token)
		if len(token) <= 8 {
			if masked != "******" {
				t.Fatalf("maskToken(%q) = %q, want ******", token, masked)
			}
			return
		}
		if masked[:4] != token[:4] || masked[len(masked)-4:] != token[len(token)-4:] {
			t.Fatalf("maskToken(%q) = %q, 首尾 4 位未保留", token, masked)
		}
	})
}
