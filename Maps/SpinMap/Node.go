package SpinMap

import (
	"GMUtils/Maps"
	"fmt"
	"runtime"
	"sync/atomic"
	"unsafe"
)

const (
	locked uint32 = 0b00000000000000000000000000000001
	free   uint32 = 0b11111111111111111111111111111110
	del    uint   = ^Maps.MaxArrayLen
)

type node[K Maps.Hashable] struct {
	k     K
	pinfo uint //last bit: del
	v     unsafe.Pointer
	nx    *node[K]
	cinfo uint32 //1 bit: updating
}

func makeNode[K Maps.Hashable](k K, hash uint, v unsafe.Pointer) *node[K] {
	t := new(node[K])
	t.k, t.pinfo, t.v = k, hash, v
	return t
}

func (cur *node[K]) Lock() {
	for infoPtr := &cur.cinfo; ; {
		if curInfo := atomic.LoadUint32(infoPtr); atomic.CompareAndSwapUint32(infoPtr, curInfo&free, curInfo|locked) {
			break
		} else {
			runtime.Gosched()
		}
	}
}

func (cur *node[K]) hash() uint {
	return cur.pinfo & Maps.MaxArrayLen
}

func (cur *node[K]) Unlock() {
	atomic.StoreUint32(&cur.cinfo, cur.cinfo&free)
}

func (cur *node[K]) safeDelete() bool {
	cur.Lock()
	defer cur.Unlock()
	if cur.pinfo&del == del {
		return false
	} else {
		cur.pinfo |= del
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
		if nx := cur.nx; nx != nil {
			nx.Lock()
			if nx.pinfo&del == del {
				cur.nx = nx.nx
				nx.Unlock()
			} else {
				nx.Unlock()
				return nx
			}
		} else {
			return nil
		}
	}
}

// Unlock cur
func (cur *node[K]) addAndUnlock(newNx *node[K]) bool {
	defer cur.Unlock()
	if cur.pinfo&del != del {
		newNx.nx = cur.nx
		cur.nx = newNx
		return true
	} else {
		return false
	}
}

func (cur *node[K]) isRelay() bool {
	return cur.v == nil
}

// Lock left
func (cur *node[K]) searchHashAndAcquire(at uint) (*node[K], *node[K]) {
	for left := cur; ; {
		left.Lock()
		if right := left.dangerNext(); right == nil {
			return left, nil
		} else if at <= right.hash() {
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
		} else if at == right.hash() {
			if k.Equal(right.k) && !right.isRelay() {
				return left, right, true
			} else {
				left.Unlock()
				left = right
			}
		} else if at > right.hash() {
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
		} else if at == right.hash() {
			if k.Equal(right.k) && !right.isRelay() {
				return left, right, true
			} else {
				left = right
			}
		} else if at > right.hash() {
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
	return fmt.Sprintf("key: %#v; val: %#v; pinfo: %d; cinfo: %t", u.k, u.get(), u.hash(), u.pinfo&del == del)
}
