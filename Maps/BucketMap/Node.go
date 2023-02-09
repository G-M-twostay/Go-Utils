package BucketMap

import (
	"fmt"
	"github.com/g-m-twostay/go-utils/Maps"
	"sync"
	"sync/atomic"
	"unsafe"
)

// node doesn't have to be generic, but removing the type parameter here somehow makes the code run very slow.
type node[K any] struct {
	info uint
	nx   unsafe.Pointer
}

func (cur *node[K]) isRelay() bool {
	return cur.info > Maps.MaxArrayLen
}

func (cur *node[K]) Hash() uint {
	return Maps.Mask(cur.info)
}

func (cur *node[K]) Next() unsafe.Pointer {
	return atomic.LoadPointer(&cur.nx)
}

func (cur *node[K]) dangerLink(oldRight, newRight unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&cur.nx, oldRight, newRight)
}

func (cur *node[K]) unlinkRelay(next *relay[K], nextPtr unsafe.Pointer) bool {
	next.Lock()
	defer next.Unlock()
	next.del = atomic.CompareAndSwapPointer(&cur.nx, nextPtr, next.nx)
	return next.del
}

func (cur *node[K]) dangerUnlink(next *node[K]) {
	atomic.StorePointer(&cur.nx, next.nx)
}

type value[K any] struct {
	node[K]
	v unsafe.Pointer
	k K
}

func (cur *value[K]) String() string {
	return fmt.Sprintf("key: %#v; val: %#v; hash: %d; relay: %t", cur.k, cur.get(), cur.info, cur.isRelay())
}

func (cur *value[K]) set(v unsafe.Pointer) {
	atomic.StorePointer(&cur.v, v)
}

func (cur *value[K]) get() unsafe.Pointer {
	return atomic.LoadPointer(&cur.v)
}
func (cur *value[K]) swap(v unsafe.Pointer) unsafe.Pointer {
	return atomic.SwapPointer(&cur.v, v)
}

// node doesn't have to be generic, but removing the type parameter here somehow makes the code run very slow.
type relay[K any] struct {
	node[K]
	sync.RWMutex
	del bool
}

func (cur *relay[K]) safeLock() bool {
	cur.Lock()
	return !cur.del
}

func (cur *relay[K]) safeRLock() bool {
	cur.RLock()
	return !cur.del
}

func (cur *relay[K]) search(k K, at uint, cmp func(K, K) bool) *value[K] {
	for left := &cur.node; ; {
		if rightB := (*node[K])(left.Next()); rightB == nil || at < rightB.Hash() {
			return nil
		} else if right := (*value[K])(unsafe.Pointer(rightB)); at == rightB.info && cmp(k, right.k) {
			return right
		} else {
			left = rightB
		}
	}
}
