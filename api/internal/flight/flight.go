// Package flight 提供 singleflight 语义（对同一 key 的并发调用只执行一次、其余共享结果）。
//
// 这是 golang.org/x/sync/singleflight 的完整副本，供 grpc 包与 http 共同使用：
// http 刻意不 import grpc 包（会拖进整个 gRPC SDK），因此把这份零依赖实现
// 抽成独立叶子子包，两个消费方各取所需。仅依赖标准库 sync。
package flight

import "sync"

// call is an in-flight or completed singleflight.Do call
type call struct {
	wg sync.WaitGroup

	// These fields are written once before the WaitGroup is done
	// and are only read after the WaitGroup is done.
	val interface{}
	err error

	// These fields are read and written with the singleflight
	// mutex held before the WaitGroup is done, and are read but
	// not written after the WaitGroup is done.
	dups  int
	chans []chan<- Result
}

// Group 代表一类工作集合，构成一个命名空间：同 key 的并发工作被去重抑制（只执行一次）。
type Group struct {
	mu sync.Mutex       // 保护 m
	m  map[string]*call // 惰性初始化
}

// Result 保存 Do 的返回结果，供通过 channel 异步接收。
type Result struct {
	Val    interface{}
	Err    error
	Shared bool
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

// DoChan 类似 Do，但通过返回的 channel 异步接收结果。
// 第二个返回值 true 表示本调用将真正执行函数；false 表示已存在同 key 的挂起请求，仅共享其结果。
func (g *Group) DoChan(key string, fn func() (interface{}, error)) (<-chan Result, bool) {
	ch := make(chan Result, 1)
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		c.dups++
		c.chans = append(c.chans, ch)
		g.mu.Unlock()
		return ch, false
	}
	c := &call{chans: []chan<- Result{ch}}
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	go g.doCall(c, key, fn)

	return ch, true
}

// doCall handles the single call for a key.
func (g *Group) doCall(c *call, key string, fn func() (interface{}, error)) {
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	for _, ch := range c.chans {
		ch <- Result{c.val, c.err, c.dups > 0}
	}
	g.mu.Unlock()
}

// ForgetUnshared 若 key 当前未被其他 goroutine 共享则将其从映射中遗忘：
// 之后对该 key 的 Do 调用会重新执行函数，而非等待早前的调用完成。
// 返回值表示该 key 是否被遗忘（或本就未知），即是否没有其他 goroutine 在等待其结果。
func (g *Group) ForgetUnshared(key string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	c, ok := g.m[key]
	if !ok {
		return true
	}
	if c.dups == 0 {
		delete(g.m, key)
		return true
	}
	return false
}
