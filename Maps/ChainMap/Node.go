package ChainMap

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type state[K any] struct {
	del bool
	nx  *node[K]
}

func (u *state[K]) changeNext(nx *node[K]) *state[K] {
	return &state[K]{u.del, nx}
}

type node[K any] struct {
	k    K
	hash uint
	v    unsafe.Pointer
	s    unsafe.Pointer
}

func (u *node[K]) addAfter(oldSt *state[K], oldStPtr unsafe.Pointer, newRight *node[K]) bool {
	newRight.s = unsafe.Pointer(&state[K]{false, oldSt.nx})
	return atomic.CompareAndSwapPointer(&u.s, oldStPtr, unsafe.Pointer(oldSt.changeNext(newRight)))
}

func (cur *node[K]) next() (*node[K], *state[K], unsafe.Pointer) {
	for {
		curStPtr := atomic.LoadPointer(&cur.s)
		curSt := (*state[K])(curStPtr)
		if nx := curSt.nx; nx == nil {
			return nil, curSt, curStPtr
		} else if nxSt := (*state[K])(atomic.LoadPointer(&nx.s)); nxSt.del {
			atomic.CompareAndSwapPointer(&cur.s, curStPtr, unsafe.Pointer(curSt.changeNext(nxSt.nx)))
		} else {
			return nx, curSt, curStPtr
		}
	}
}

func (u *node[K]) searchHash(at uint) (*node[K], *state[K], unsafe.Pointer, *node[K]) {
	for left := u; ; {
		if right, leftSt, leftStPtr := left.next(); right == nil {
			return left, leftSt, leftStPtr, nil
		} else if at <= right.hash { //put at the first possible position: 1, x, 2; x=2
			return left, leftSt, leftStPtr, right
		} else {
			left = right
		}
	}

}

func (u *node[K]) searchKey(k K, at uint, cmp func(K, K) bool) (*node[K], *state[K], unsafe.Pointer, *node[K], bool) {
	for left := u; ; {
		if right, leftSt, leftStPtr := left.next(); right == nil {
			return left, leftSt, leftStPtr, nil, false
		} else if at == right.hash {
			if cmp(k, right.k) && !right.isRelay() { //found
				return left, leftSt, leftStPtr, right, true
			} else {
				left = right
			}
		} else if at > right.hash {
			left = right
		} else { //put at the last possible position: 1, x, 2; x=1
			return left, leftSt, leftStPtr, right, false
		}
	}

}

func (u *node[K]) delete() bool {
	for curStPtr := atomic.LoadPointer(&u.s); ; curStPtr = atomic.LoadPointer(&u.s) {
		if curSt := (*state[K])(curStPtr); curSt.del {
			return false
		} else if atomic.CompareAndSwapPointer(&u.s, curStPtr, unsafe.Pointer(&state[K]{true, curSt.nx})) {
			return true
		}
	}
}

func (u *node[K]) get() unsafe.Pointer {
	return atomic.LoadPointer(&u.v)
}

func (u *node[K]) set(newPtr unsafe.Pointer) {
	atomic.StorePointer(&u.v, newPtr)
}

func (cur *node[K]) swap(v unsafe.Pointer) unsafe.Pointer {
	return atomic.SwapPointer(&cur.v, v)
}

func (u *node[K]) isRelay() bool {
	return u.v == nil //this is technically dirty, but since non-relay node will never have u.v be nil and relay nodes will always have u.v=nil this is fine.
}

func (u *node[K]) String() string {
	t := (*state[K])(atomic.LoadPointer(&u.s))
	return fmt.Sprintf("key: %#v; val: %#v; hash: %d; del: %t", u.k, u.get(), u.hash, t.del)
}
