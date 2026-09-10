// Package flight 提供 singleflight 语义（对同一 key 的并发调用只执行一次、其余共享结果）。
//
// 这是 golang.org/x/sync/singleflight 的裁剪副本，只保留 grpc 与 http 实际消费的 Do：
// http 刻意不 import grpc 包（会拖进整个 gRPC SDK），因此把这份零依赖实现
// 抽成独立叶子子包，两个消费方各取所需。仅依赖标准库 sync。
package flight

import "sync"

// call 是一次在途（或已完成）的 Do 调用。
type call struct {
	wg sync.WaitGroup

	// val/err 在 WaitGroup 完成前写入一次，完成后只读。
	val interface{}
	err error

	// dups 在单飞锁保护下自增，用于判定结果是否被多个调用方共享。
	dups int
}

// Group 代表一类工作集合，构成一个命名空间：同 key 的并发工作被去重抑制（只执行一次）。
type Group struct {
	mu sync.Mutex       // 保护 m
	m  map[string]*call // 惰性初始化
}

// Do 执行并返回给定函数的结果：同一 key 在任意时刻只允许一个执行在途。
// 若 key 已存在重复调用，则等待首次执行完成后共享同一份结果。
// 返回值 shared 表示结果是否被多个调用方共享。
func (g *Group) Do(key string, fn func() (interface{}, error)) (v interface{}, err error, shared bool) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		c.dups++
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err, true
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	g.doCall(c, key, fn)
	return c.val, c.err, c.dups > 0
}

// doCall 执行 key 对应的唯一一次函数调用，完成后把该 key 从映射中移除。
func (g *Group) doCall(c *call, key string, fn func() (interface{}, error)) {
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
}
