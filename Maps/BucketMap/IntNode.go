package BucketMap

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"sync/atomic"
	"unsafe"
)

type intNode[K constraints.Integer] struct {
	hash uint
	v    unsafe.Pointer
	nx   unsafe.Pointer
	k    K
	flag bool //true if relay
}

func (cur *intNode[K]) lock() *relayLock {
	return (*relayLock)(cur.v)
}

func (cur *intNode[K]) Next() unsafe.Pointer {
	return atomic.LoadPointer(&cur.nx)
}

func (cur *intNode[K]) tryLazyLink(oldRight, newRight unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&cur.nx, oldRight, newRight)
}

func (cur *intNode[K]) unlinkRelay(next *intNode[K], nextPtr unsafe.Pointer) bool {
	t := next.lock()
	t.Lock()
	defer t.Unlock()
	t.del = atomic.CompareAndSwapPointer(&cur.nx, nextPtr, next.nx)
	return t.del
}

func (cur *intNode[K]) dangerUnlink(next *intNode[K]) {
	atomic.StorePointer(&cur.nx, next.nx)
}

func (cur *intNode[K]) cmpKey(k K) bool {
	return k == cur.k && !cur.flag
}

func (cur *intNode[K]) searchKey(k K, at uint) (*intNode[K], *intNode[K], bool) {
	for left := cur; ; {
		if rightPtr := left.Next(); rightPtr == nil {
			return left, nil, false
		} else if right := (*intNode[K])(rightPtr); at < right.hash {
			return left, right, false
		} else if right.cmpKey(k) {
			return left, right, true
		} else {
			left = right
		}
	}
}

func (cur *intNode[K]) set(v unsafe.Pointer) {
	atomic.StorePointer(&cur.v, v)
}

func (cur *intNode[K]) get() unsafe.Pointer {
	return atomic.LoadPointer(&cur.v)
}

func (cur *intNode[K]) String() string {
	return fmt.Sprintf("key: %#v; val: %#v; hash: %d; relay: %t", cur.k, cur.get(), cur.hash, cur.flag)

}
