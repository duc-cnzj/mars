package deploy

import (
	"sort"
	"sync"
)

// priority 按 priority 从高到低对加入的元素排序；零值即可直接使用（list 为 nil 时 Add/Sort 均正常）。
// 互斥锁放在内部字段，避免向调用方暴露 Lock/Unlock，也杜绝值拷贝时连带拷贝锁。
// 仅 deploy 内部使用（唯一消费方是 jobRunner 的钩子回调），故保持未导出。
type priority[T any] struct {
	mu   sync.RWMutex
	list []priorityCallback[T]
}

// Add 按 priority 追加一个元素。
func (m *priority[T]) Add(priority int, t T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.list = append(m.list, priorityCallback[T]{
		priority: priority,
		fn:       t,
	})
}

// Sort 返回按 priority 降序排列的元素副本，不改动内部列表。
func (m *priority[T]) Sort() []T {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var l = make(prioritySort[T], len(m.list))
	copy(l, m.list)
	sort.Sort(l)
	var res = make([]T, 0, len(l))
	for _, c := range l {
		res = append(res, c.fn)
	}
	return res
}

// prioritySort 是 sort.Interface 的实现，按 priority 降序比较。
type prioritySort[T any] []priorityCallback[T]

// Len 返回列表长度。
func (s prioritySort[T]) Len() int {
	return len(s)
}

// Less 返回第 i 个元素优先级是否高于第 j 个（降序）。
func (s prioritySort[T]) Less(i, j int) bool {
	return s[i].priority > s[j].priority
}

// Swap 交换第 i、j 个元素。
func (s prioritySort[T]) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// priorityCallback 保存一个元素及其排序优先级。
type priorityCallback[T any] struct {
	priority int
	fn       T
}
