package flight

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo_DeduplicatesConcurrent(t *testing.T) {
	g := new(Group)
	var calls int32
	fn := func() (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(50 * time.Millisecond)
		return "result", nil
	}

	const n = 10
	var wg sync.WaitGroup
	results := make([]interface{}, n)
	shares := make([]bool, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], _, shares[i] = g.Do("key", fn)
		}(i)
	}
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("fn 应只执行 1 次，实际 %d 次", got)
	}
	for i := 0; i < n; i++ {
		if results[i] != "result" {
			t.Errorf("caller %d 结果错误: %v", i, results[i])
		}
		// 首个发起者可能 shared=false（无 dup），其余都是 true
	}
}

func TestDo_DistinctKeysNoDedup(t *testing.T) {
	g := new(Group)
	var calls int32
	for i := 0; i < 3; i++ {
		v, err, shared := g.Do(fmt.Sprintf("key-%d", i), func() (interface{}, error) {
			atomic.AddInt32(&calls, 1)
			return i, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if shared {
			t.Errorf("独立 key 不应 shared")
		}
		if v != i {
			t.Errorf("key-%d 结果 = %v, want %d", i, v, i)
		}
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("fn 应执行 3 次，实际 %d", atomic.LoadInt32(&calls))
	}
}

func TestDo_PropagatesError(t *testing.T) {
	wantErr := errors.New("boom")
	g := new(Group)
	_, err, _ := g.Do("key", func() (interface{}, error) {
		return nil, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestDo_ExecutesAgainAfterCompletion(t *testing.T) {
	g := new(Group)
	var calls int32
	for i := 0; i < 2; i++ {
		v, err, _ := g.Do("key", func() (interface{}, error) {
			atomic.AddInt32(&calls, 1)
			return "v", nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if v != "v" {
			t.Fatalf("结果 = %v", v)
		}
	}
	// Do 完成后 key 已从 map 删除（不再缓存结果），二次调用必须重新执行 fn。
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("fn 应执行 2 次（完成后不缓存），实际 %d", atomic.LoadInt32(&calls))
	}
}

func TestDoChan(t *testing.T) {
	g := new(Group)
	ch, ok := g.DoChan("key", func() (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return "val", nil
	})
	if !ok {
		t.Fatalf("首个 DoChan 应返回 ok=true")
	}
	select {
	case r := <-ch:
		if r.Val != "val" || r.Err != nil {
			t.Fatalf("Result = %+v", r)
		}
	case <-time.After(time.Second):
		t.Fatal("DoChan 结果超时")
	}
}

func TestDoChan_DuplicateReceivesSharedResult(t *testing.T) {
	g := new(Group)
	var calls int32
	ch, first := g.DoChan("key", func() (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(20 * time.Millisecond)
		return "val", nil
	})
	if !first {
		t.Fatal("首个应 first=true")
	}
	ch2, first2 := g.DoChan("key", func() (interface{}, error) {
		return "should-not-run", nil
	})
	if first2 {
		t.Fatal("重复调用应 first=false")
	}
	for i, c := range []<-chan Result{ch, ch2} {
		r := <-c
		if r.Val != "val" {
			t.Errorf("ch%d 结果 = %v", i, r.Val)
		}
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("fn 应执行 1 次，实际 %d", atomic.LoadInt32(&calls))
	}
}

func TestForgetUnshared(t *testing.T) {
	g := new(Group)
	// 未知 key → 视为已 forgotten
	if !g.ForgetUnshared("nonexistent") {
		t.Fatal("未知 key 应返回 true")
	}

	g.mu.Lock()
	g.m = map[string]*call{"k": {dups: 0}}
	g.mu.Unlock()
	if !g.ForgetUnshared("k") {
		t.Fatal("无 dup 的 key 应被 forget")
	}
	if _, ok := g.m["k"]; ok {
		t.Fatal("key 应已删除")
	}

	g.mu.Lock()
	g.m = map[string]*call{"k": {dups: 2}}
	g.mu.Unlock()
	if g.ForgetUnshared("k") {
		t.Fatal("有 dup 的 key 不应被 forget")
	}
}
