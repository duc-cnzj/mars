package sort

import (
	"sort"
	"sync"
)

type PrioritySort[T any] struct {
	sync.RWMutex
	list []callback[T]
}

func NewPriority[T any]() *PrioritySort[T] {
	return &PrioritySort[T]{}
}

func (m *PrioritySort[T]) Add(priority int, t T) {
	m.Lock()
	defer m.Unlock()
	m.list = append(m.list, callback[T]{
		priority: priority,
		fn:       t,
	})
}

func (m *PrioritySort[T]) Sort() []T {
	m.RLock()
	defer m.RUnlock()
	var l = make(sortImpl[T], len(m.list))
	copy(l, m.list)
	sort.Sort(l)
	var res = make([]T, 0, len(l))
	for _, c := range l {
		res = append(res, c.fn)
	}
	return res
}

type sortImpl[T any] []callback[T]

func (s sortImpl[T]) Len() int {
	return len(s)
}

func (s sortImpl[T]) Less(i, j int) bool {
	return s[i].priority > s[j].priority
}

func (s sortImpl[T]) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type callback[T any] struct {
	priority int
	fn       T
}
