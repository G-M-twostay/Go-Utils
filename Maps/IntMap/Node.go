package IntMap

import (
	"fmt"
	"github.com/g-m-twostay/go-utils/Maps/internal"
	"sync"
	"sync/atomic"
	"unsafe"
)

// node doesn't have to be generic, but removing the type parameter here somehow makes the code run very slow.
type node struct {
	info uint
	nx   unsafe.Pointer
}

func (cur node) isRelay() bool {
	return cur.info > internal.MaxArrayLen
}

func (cur node) Hash() uint {
	return internal.Mask(cur.info)
}

func (cur *node) Next() unsafe.Pointer {
	return atomic.LoadPointer(&cur.nx)
}

func (cur *node) dangerLink(oldRight, newRight unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&cur.nx, oldRight, newRight)
}

func (cur *node) unlinkRelay(next *relay, nextPtr unsafe.Pointer) bool {
	next.Lock()
	defer next.Unlock()
	next.del = atomic.CompareAndSwapPointer(&cur.nx, nextPtr, next.nx)
	return next.del
}

func (cur *node) dangerUnlink(next *node) {
	atomic.StorePointer(&cur.nx, next.nx)
}

type value[K comparable] struct {
	node
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

func (cur *value[K]) cas(old, new unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&cur.v, new, old)
}

// node doesn't have to be generic, but removing the type parameter here somehow makes the code run very slow.
type relay struct {
	node
	sync.RWMutex
	del bool
}

func (cur *relay) safeLock() bool {
	cur.Lock()
	return !cur.del
}

func (cur *relay) safeRLock() bool {
	cur.RLock()
	return !cur.del
}

func search[K comparable](cur *relay, k K, at uint) *value[K] {
	for left := &cur.node; ; {
		if rightB := (*node)(left.Next()); rightB == nil || at < rightB.Hash() {
			return nil
		} else if right := (*value[K])(unsafe.Pointer(rightB)); at == rightB.info && k == right.k {
			return right
		} else {
			left = rightB
		}
	}
}
