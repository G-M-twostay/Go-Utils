package BucketMap

import (
	"GMUtils/Maps"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

type lockLevel uint8

const (
	noLock lockLevel = iota
	rLock
	wLock
)

var noOpLock sync.Locker = NoLock{}

type NoLock struct{}

func (NoLock) Lock()   {}
func (NoLock) Unlock() {}

type node[K Maps.Hashable] struct {
	k    K
	hash uint
	v    unsafe.Pointer
	nx   unsafe.Pointer
	*sync.RWMutex
}

func makeRelay[K Maps.Hashable](hash uint) *node[K] {
	t := new(node[K])
	t.RWMutex = new(sync.RWMutex)
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

func (cur *node[K]) dangerUnlinkNext(next *node[K], nextPtr unsafe.Pointer) {
	atomic.StorePointer(&cur.nx, next.Next())
}

func (cur *node[K]) isRelay() bool {
	return cur.RWMutex != nil
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

func (cur *node[K]) searchKeyW(k K, at uint) (*node[K], *node[K], unsafe.Pointer, bool, sync.Locker) {
	prevLock := noOpLock
	for left := cur; ; {

		if left.isRelay() {
			prevLock.Unlock()
			prevLock = left
			prevLock.Lock()
		}

		if rightPtr := left.Next(); rightPtr == nil {
			return left, nil, nil, false, prevLock
		} else if right := (*node[K])(rightPtr); at == right.hash {
			if k.Equal(right.k) && !right.isRelay() {
				return left, right, rightPtr, true, prevLock
			} else {
				left = right
			}
		} else if at > right.hash {
			left = right
		} else {
			return left, right, rightPtr, false, prevLock
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
