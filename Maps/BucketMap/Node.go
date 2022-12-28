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
	return !l.del
}

func (l *relayLock) safeRLock() bool {
	l.RLock()
	return !l.del
}

type node[K Maps.Hashable] struct {
	k    K
	hash uint
	v    unsafe.Pointer
	nx   unsafe.Pointer
	flag bool //true if relay
}

func makeRelay[K Maps.Hashable](hash uint) *node[K] {
	t := new(node[K])
	t.v = unsafe.Pointer(new(relayLock))
	t.hash = hash
	t.flag = true
	return t
}

func (cur *node[K]) lock() *relayLock {
	return (*relayLock)(cur.v)
}

func (cur *node[K]) Next() unsafe.Pointer {
	return atomic.LoadPointer(&cur.nx)
}

func (cur *node[K]) tryLazyLink(oldRight, newRight unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&cur.nx, oldRight, newRight)
}

func (cur *node[K]) tryLink(oldRight unsafe.Pointer, newRight *node[K], newRightPtr unsafe.Pointer) bool {
	newRight.nx = oldRight
	return atomic.CompareAndSwapPointer(&cur.nx, oldRight, newRightPtr)
}

func (cur *node[K]) unlinkRelay(next *node[K], nextPtr unsafe.Pointer) bool {
	t := next.lock()
	t.Lock()
	defer t.Unlock()
	t.del = atomic.CompareAndSwapPointer(&cur.nx, nextPtr, next.nx)
	return t.del
}

func (cur *node[K]) dangerUnlink(next *node[K]) {
	atomic.StorePointer(&cur.nx, next.nx)
}

func (cur *node[K]) cmpKey(k K) bool {
	return k.Equal(cur.k) && !cur.flag
}

func (cur *node[K]) searchKey(k K, at uint) (*node[K], *node[K], bool) {
	for left := cur; ; {
		if rightPtr := left.Next(); rightPtr == nil {
			return left, nil, false
		} else if right := (*node[K])(rightPtr); at < right.hash {
			return left, right, false
		} else if at == right.hash && right.cmpKey(k) {
			return left, right, true
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

func (cur *node[K]) String() string {
	return fmt.Sprintf("key: %#v; val: %#v; hash: %d", cur.k, cur.get(), cur.hash)

}
