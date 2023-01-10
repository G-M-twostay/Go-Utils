package BucketMap

import (
	"fmt"
	"github.com/g-m-twostay/go-utils/Maps"
	"sync/atomic"
	"unsafe"
)

type node[K any] struct {
	k    K
	info uint //first bit: relay or not; other bits: hash value
	v    unsafe.Pointer
	nx   unsafe.Pointer
}

func (cur *node[K]) Hash() uint {
	return Maps.Mask(cur.info)
}

func (cur *node[K]) isRelay() bool {
	return cur.info > Maps.MaxArrayLen
}

// lock returns the pointer to the lock by node.v; this will panic if cur isn't a relay
func (cur *node[K]) lock() *Maps.FlagLock {
	return (*Maps.FlagLock)(cur.v)
}

func (cur *node[K]) Next() unsafe.Pointer {
	return atomic.LoadPointer(&cur.nx)
}

func (cur *node[K]) dangerLink(oldRight, newRight unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&cur.nx, oldRight, newRight)
}

// unlinkRelay performs CAS on cur.nx with next and sets next.Del to true if success. next is a relay.
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

// search the given key with hash value at from cur. Return nil if not found.
func (cur *node[K]) search(k K, at uint, cmp func(K, K) bool) *node[K] {
	for left := cur; ; {
		if right := (*node[K])(left.Next()); right == nil || at < right.Hash() {
			return nil
		} else if at == right.info && cmp(k, right.k) {
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
