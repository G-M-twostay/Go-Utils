package IntMap

import (
	"fmt"
	"github.com/g-m-twostay/go-utils/Maps"
	"sync"
	"sync/atomic"
	"unsafe"
)

type base[K comparable] struct {
	info uint
	nx   unsafe.Pointer
}

func (cur *base[K]) isRelay() bool {
	return cur.info > Maps.MaxArrayLen
}

func (cur *base[K]) Hash() uint {
	return Maps.Mask(cur.info)
}

func (cur *base[K]) Next() unsafe.Pointer {
	return atomic.LoadPointer(&cur.nx)
}

func (cur *base[K]) dangerLink(oldRight, newRight unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&cur.nx, oldRight, newRight)
}

func (cur *base[K]) unlinkRelay(next *relay[K], nextPtr unsafe.Pointer) bool {
	next.Lock()
	defer next.Unlock()
	next.del = atomic.CompareAndSwapPointer(&cur.nx, nextPtr, next.nx)
	return next.del
}

func (cur *base[K]) dangerUnlink(next *base[K]) {
	atomic.StorePointer(&cur.nx, next.nx)
}

type node[K comparable] struct {
	base[K]
	v unsafe.Pointer
	k K
}

func (cur *node[K]) String() string {
	return fmt.Sprintf("key: %#v; val: %#v; hash: %d; relay: %t", cur.k, cur.get(), cur.info, cur.isRelay())
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

type relay[K comparable] struct {
	base[K]
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

func (cur *relay[K]) search(k K, at uint) *node[K] {
	for left := &cur.base; ; {
		if rightB := (*base[K])(left.Next()); rightB == nil || at < rightB.Hash() {
			return nil
		} else if right := (*node[K])(unsafe.Pointer(rightB)); at == rightB.info && k == right.k {
			return right
		} else {
			left = rightB
		}
	}
}
