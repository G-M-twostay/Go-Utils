package BucketMap

import (
	"GMUtils/Maps"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

type relayLock struct {
	sync.RWMutex
	del bool
}

func (l *relayLock) safeLock() bool {
	l.Lock()
	if l.del {
		l.Unlock()
		return false
	}
	return true
}

func (l *relayLock) safeRLock() bool {
	l.RLock()
	if l.del {
		l.RUnlock()
		return false
	}
	return true
}

type node[K Maps.Hashable] struct {
	k    K
	hash uint
	v    unsafe.Pointer
	nx   unsafe.Pointer
	*relayLock
}

func makeRelay[K Maps.Hashable](hash uint) *node[K] {
	t := new(node[K])
	t.relayLock = new(relayLock)
	t.hash = hash
	return t
}

func (cur *node[K]) Next() unsafe.Pointer {
	return atomic.LoadPointer(&cur.nx)
}

func (cur *node[K]) tryLazyLink(oldRight, newRight unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&cur.nx, oldRight, newRight)
}

func (cur *node[K]) tryLink(oldRight unsafe.Pointer, newRight *node[K]) bool {
	newRight.nx = oldRight
	return atomic.CompareAndSwapPointer(&cur.nx, oldRight, unsafe.Pointer(newRight))
}

func (cur *node[K]) unlinkRelay(next *node[K], nextPtr unsafe.Pointer) bool {
	next.Lock()
	defer next.Unlock()
	if atomic.CompareAndSwapPointer(&cur.nx, nextPtr, next.Next()) {
		next.del = true
	}
	return next.del
}

func (cur *node[K]) dangerUnlink(next *node[K]) {
	atomic.StorePointer(&cur.nx, next.Next())
}

func (cur *node[K]) isRelay() bool {
	return cur.relayLock != nil
}

func (cur *node[K]) searchKey(k K, at uint) (*node[K], *node[K], unsafe.Pointer, bool) {
	for left := cur; ; {
		if rightPtr := left.Next(); rightPtr == nil {
			return left, nil, nil, false
		} else if right := (*node[K])(rightPtr); at == right.hash {
			if k.Equal(right.k) && !right.isRelay() {
				return left, right, rightPtr, true
			} else {
				left = right
			}
		} else if at > right.hash {
			left = right
		} else {
			return left, right, rightPtr, false
		}
	}
}

func (cur *node[K]) set(v unsafe.Pointer) {
	atomic.StorePointer(&cur.v, v)
}

func (cur *node[K]) get() unsafe.Pointer {
	return atomic.LoadPointer(&cur.v)
}

func (cur *node[K]) String() string {
	return fmt.Sprintf("key: %#v; val: %#v; hash: %d", cur.k, cur.get(), cur.hash)

}
