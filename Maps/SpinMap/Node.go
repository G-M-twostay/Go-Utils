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
	del    uint32 = 0b00000000000000000000000000000010
)

type node[K Maps.Hashable] struct {
	k    K
	hash uint
	v    unsafe.Pointer
	nx   *node[K]
	//info uint32 //1 bit: updating; 2 bit: delete
	updating atomic.Bool
	del      bool
}

func makeNode[K Maps.Hashable](k K, hash uint, v unsafe.Pointer) *node[K] {
	t := new(node[K])
	t.k, t.hash, t.v = k, hash, v
	return t
}

func (cur *node[K]) acquire() {
	//for infoPtr := &cur.info; ; {
	//	curInfo := atomic.LoadUint32(infoPtr)
	//	if atomic.CompareAndSwapUint32(infoPtr, curInfo&free, curInfo|locked) {
	//		break
	//	}
	//}
	for !cur.updating.CompareAndSwap(false, true) {
		runtime.Gosched()
	}
}

func (cur *node[K]) release() {
	//if info := atomic.LoadUint32(&cur.info); info&locked == locked {
	//	atomic.StoreUint32(&cur.info, info&free)
	//} else {
	//	panic("not locked, can't release")
	//}
	if !cur.updating.Swap(false) {
		panic("not locked, can't release")
	}
}

func (cur *node[K]) safeDelete() bool {
	cur.acquire()
	defer cur.release()
	//if info := atomic.LoadUint32(&cur.info); info&del == del {
	//	return false
	//} else {
	//	atomic.StoreUint32(&cur.info, info|del)
	//	return true
	//}
	if cur.del {
		return false
	} else {
		cur.del = true
		return true
	}
}

func (cur *node[K]) safeNext() *node[K] {
	cur.acquire()
	defer cur.release()
	return cur.dangerNext()
}

func (cur *node[K]) dangerNext() *node[K] {
	for {
		if oldNx := cur.nx; oldNx != nil {
			oldNx.acquire()
			if oldNx.del {
				cur.nx = oldNx.nx
				oldNx.release()
			} else {
				oldNx.release()
				return oldNx
			}
		} else {
			return nil
		}
	}
}

// release cur
func (cur *node[K]) addAndRelease(newNx *node[K]) bool {
	defer cur.release()
	if !cur.del {
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

// acquire left
func (cur *node[K]) searchHashAndAcquire(at uint) (*node[K], *node[K]) {
	for left := cur; ; {
		left.acquire()
		if right := left.dangerNext(); right == nil {
			return left, nil
		} else if at <= right.hash {
			return left, right
		} else {
			left.release()
			left = right
		}
	}
}

// acquire left
func (cur *node[K]) searchKeyAndAcquire(k K, at uint) (*node[K], *node[K], bool) {
	for left := cur; ; {
		left.acquire()
		if right := left.dangerNext(); right == nil {
			return left, nil, false
		} else if at == right.hash {
			if k.Equal(right.k) && !right.isRelay() {
				return left, right, true
			} else {
				left.release()
				left = right
			}
		} else if at > right.hash {
			left.release()
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
			if k.Equal(right.k) && !right.isRelay() {
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
