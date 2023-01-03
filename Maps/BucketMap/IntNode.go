package BucketMap

import (
	"GMUtils/Maps"
	"fmt"
	"golang.org/x/exp/constraints"
	"sync/atomic"
	"unsafe"
)

type intNode[K constraints.Integer] struct {
	k    K
	info uint
	v    unsafe.Pointer
	nx   unsafe.Pointer
}

func (cur *intNode[K]) Hash() uint {
	return Maps.Mask(cur.info)
}

func (cur *intNode[K]) isRelay() bool {
	return cur.info > Maps.MaxArrayLen
}

func (cur *intNode[K]) lock() *relayLock {
	return (*relayLock)(cur.v)
}

func (cur *intNode[K]) Next() unsafe.Pointer {
	return atomic.LoadPointer(&cur.nx)
}

func (cur *intNode[K]) dangerLink(oldRight, newRight unsafe.Pointer) bool {
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

func (cur *intNode[K]) searchKey(k K, at uint) (*intNode[K], bool) {
	for left := cur; ; {
		if right := (*intNode[K])(left.Next()); right == nil || at < right.Hash() {
			return right, false
		} else if at == right.info && k == right.k {
			return right, true
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
	return fmt.Sprintf("key: %#v; val: %#v; info: %d; relay: %t", cur.k, cur.get(), cur.Hash(), cur.isRelay())

}
