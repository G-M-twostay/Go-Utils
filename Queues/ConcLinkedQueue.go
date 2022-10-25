package Queues

import (
	"sync/atomic"
)

type node[T any] struct {
	v  T
	nx atomic.Pointer[node[T]]
}

type syncLinkedQ[T any] struct {
	headPtr, tail atomic.Pointer[node[T]]
}

func MakeConcurrentLinkedQueue[T any]() Queue[T] {
	t := syncLinkedQ[T]{}
	a := new(node[T])
	t.headPtr.Store(a)
	t.tail.Store(a)
	return &t
}

func (c *syncLinkedQ[T]) Push(item T) {
	newNode := &node[T]{item, atomic.Pointer[node[T]]{}}
	var oldTail *node[T]
	for added := false; !added; {
		oldTail = c.tail.Load()
		oldTailNext := oldTail.nx.Load()
		if oldTailNext != nil {
			c.tail.CompareAndSwap(oldTail, oldTailNext)
		} else {
			added = c.tail.Load().nx.CompareAndSwap(oldTailNext, newNode)
		}
	}
	c.tail.CompareAndSwap(oldTail, newNode)
}

func (c *syncLinkedQ[T]) Pop() (T, error) {
	var oldHead *node[T]
	for removed := false; !removed; {
		oldHeadPtr, oldTail := c.headPtr.Load(), c.tail.Load()
		oldHead = oldHeadPtr.nx.Load()
		if oldTail == oldHeadPtr {
			if oldHead == nil {
				return *new(T), &EmptyQueueError{}
			}
			c.tail.CompareAndSwap(oldTail, oldHead)
		} else {
			removed = c.headPtr.CompareAndSwap(oldHeadPtr, oldHead)
		}
	}
	return oldHead.v, nil
}

func (c syncLinkedQ[T]) Peek() T {
	return c.headPtr.Load().nx.Load().v
}

func (c syncLinkedQ[T]) Empty() bool {
	return c.headPtr.Load().nx.Load() == nil
}
