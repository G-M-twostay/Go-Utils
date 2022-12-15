package SpinMap

import (
	"GMUtils/Maps"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

type node[K Maps.Hashable] struct {
	k    K
	hash uint
	v    unsafe.Pointer
	nx   *node[K]
	sync.Mutex
	del bool
}

func makeNode[K Maps.Hashable](k K, hash uint, v unsafe.Pointer) *node[K] {
	t := new(node[K])
	t.k, t.hash, t.v = k, hash, v
	return t
}

func (cur *node[K]) safeDelete() (r bool) {
	cur.Lock()
	defer cur.Unlock()
	if cur.del {
		return false
	} else {
		cur.del = true
		return true
	}
}

func (cur *node[K]) safeNext() *node[K] {
	cur.Lock()
	defer cur.Unlock()
	return cur.dangerNext()
}

func (cur *node[K]) dangerNext() *node[K] {
	for {
		if oldNx := cur.nx; oldNx != nil {
			oldNx.Lock()
			if oldNx.del {
				cur.nx = oldNx.nx
				oldNx.Unlock()
			} else {
				oldNx.Unlock()
				return oldNx
			}
		} else {
			return nil
		}
	}
}

// Unlock cur
func (cur *node[K]) addAndRelease(newNx *node[K]) bool {
	defer cur.Unlock()
	if !cur.del {
		newNx.nx = cur.nx
		cur.nx = newNx
		return true
	} else {
		return false
	}
}

// Lock left
func (cur *node[K]) searchHashAndAcquire(at uint) (*node[K], *node[K]) {
	for left := cur; ; {
		left.Lock()
		if right := left.dangerNext(); right == nil {
			return left, nil
		} else if at <= right.hash {
			return left, right
		} else {
			left.Unlock()
			left = right
		}
	}
}

// Lock left
func (cur *node[K]) searchKeyAndAcquire(k K, at uint) (*node[K], *node[K], bool) {
	for left := cur; ; {
		left.Lock()
		if right := left.dangerNext(); right == nil {
			return left, nil, false
		} else if at == right.hash {
			if k.Equal(right.k) && right.v != nil {
				return left, right, true
			} else {
				left.Unlock()
				left = right
			}
		} else if at > right.hash {
			left.Unlock()
			left = right
		} else {
			return left, right, false
		}
	}
}

func (cur *node[K]) searchKey(k K, at uint) (*node[K], *node[K], bool) {
	for left := cur; ; {
		if right := left.safeNext(); right == nil {
			return left, nil, false
		} else if at == right.hash {
			if k.Equal(right.k) && right.v != nil {
				return left, right, true
			} else {
				left = right
			}
		} else if at > right.hash {
			left = right
		} else {
			return left, right, false
		}
	}
}

func (cur *node[K]) get() unsafe.Pointer {
	return atomic.LoadPointer(&cur.v)
}

func (cur *node[K]) set(v unsafe.Pointer) {
	atomic.StorePointer(&cur.v, v)
}

func (u *node[K]) String() string {
	return fmt.Sprintf("key: %#v; val: %#v; hash: %d; info: %t", u.k, u.get(), u.hash, u.del)
}
