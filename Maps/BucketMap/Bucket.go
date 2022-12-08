package BucketMap

import (
	"GMUtils/Maps"
	"sync"
	"unsafe"
)

type bucket[K Maps.Hashable] struct {
	*sync.RWMutex
	*node[K]
}

func (b *bucket[K]) init(n *node[K]) {
	b.RWMutex = new(sync.RWMutex)
	b.node = n
}

func (b bucket[K]) rmv(k K, at uint) (v unsafe.Pointer) {
	b.Lock()
	defer b.Unlock()
	if l, r, rp, f := b.searchKey(k, at); f {
		l.dangerUnlinkNext(r, rp)
		v = r.get()
	}
	return
}

func (b bucket[K]) rmvNode(target *node[K]) bool {
	b.Lock()
	defer b.Unlock()
	for left := b.node; ; {
		if rightPtr := left.Next(); rightPtr == nil {
			return false
		} else if right := (*node[K])(rightPtr); target.hash > right.hash {
			left = right
		} else if target.hash == right.hash {
			if target == right {
				left.dangerUnlinkNext(right, rightPtr)
				return true
			} else {
				left = right
			}
		} else {
			return false
		}
	}
}

func (b bucket[K]) setOrAdd(k K, at uint, v unsafe.Pointer) bool {
	b.RLock()
	defer b.RUnlock()
	var l, r *node[K] = b.node, nil
	var rp unsafe.Pointer
	var f bool
	for {
		l, r, rp, f = l.searchKey(k, at)
		if f {
			r.set(v)
			return false
		} else if l.tryLink(rp, unsafe.Pointer(&node[K]{k, at, v, rp})) {
			return true
		}
	}
}

func (b bucket[K]) getOrAdd(k K, at uint, v unsafe.Pointer) unsafe.Pointer {
	b.RLock()
	defer b.RUnlock()
	var l, r *node[K] = b.node, nil
	var rp unsafe.Pointer
	var f bool
	for {
		l, r, rp, f = l.searchKey(k, at)
		if f {
			return r.get()
		} else if l.tryLink(rp, unsafe.Pointer(&node[K]{k, at, v, rp})) {
			return nil
		}
	}
}

func (b bucket[K]) addNode(newRelay *node[K]) {
	b.RLock()
	defer b.RUnlock()
	var l *node[K] = b.node
	var rp unsafe.Pointer
	for {
		l, _, rp = l.searchHash(newRelay.hash)
		newRelay.nx = rp
		if l.tryLink(rp, unsafe.Pointer(newRelay)) {
			break
		}
	}
}

func (b bucket[K]) get(k K, at uint) (v unsafe.Pointer) {
	b.RLock()
	defer b.RUnlock()
	if _, r, _, f := b.searchKey(k, at); f {
		v = r.get()
	}
	return
}
