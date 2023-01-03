package IntMap

import (
	"GMUtils/Maps"
	"fmt"
	"golang.org/x/exp/constraints"
	"sync/atomic"
	"unsafe"
)

type node[K constraints.Integer] struct {
	k    K
	info uint
	v    unsafe.Pointer
	nx   unsafe.Pointer
}

func (cur *node[K]) Hash() uint {
	return Maps.Mask(cur.info)
}

func (cur *node[K]) isRelay() bool {
	return cur.info > Maps.MaxArrayLen
}

func (cur *node[K]) lock() *Maps.FlagLock {
	return (*Maps.FlagLock)(cur.v)
}

func (cur *node[K]) Next() unsafe.Pointer {
	return atomic.LoadPointer(&cur.nx)
}

func (cur *node[K]) dangerLink(oldRight, newRight unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&cur.nx, oldRight, newRight)
}

func (cur *node[K]) unlinkRelay(next *node[K], nextPtr unsafe.Pointer) bool {
	t := next.lock()
	t.Lock()
	defer t.Unlock()
	t.Del = atomic.CompareAndSwapPointer(&cur.nx, nextPtr, next.nx)
	return t.Del
}

func (cur *node[K]) dangerUnlink(next *node[K]) {
	atomic.StorePointer(&cur.nx, next.nx)
}

func (cur *node[K]) search(k K, at uint) *node[K] {
	for left := cur; ; {
		if right := (*node[K])(left.Next()); right == nil || at < right.Hash() {
			return nil
		} else if at == right.info && k == right.k {
			return right
		} else {
			left = right
		}
	}
}

func (cur *node[K]) set(v unsafe.Pointer) {
	atomic.StorePointer(&cur.v, v)
}

func (cur *node[K]) get() unsafe.Pointer {
	return atomic.LoadPointer(&cur.v)
}
func (cur *node[K]) swap(v unsafe.Pointer) unsafe.Pointer {
	return atomic.SwapPointer(&cur.v, v)
}

func (cur *node[K]) String() string {
	return fmt.Sprintf("key: %#v; val: %#v; info: %d; relay: %t", cur.k, cur.get(), cur.Hash(), cur.isRelay())

}
