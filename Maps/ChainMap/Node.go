package ChainMap

import (
	"GMUtils/Maps"
	"fmt"
	"sync/atomic"
	"unsafe"
)

type state[K Maps.Hashable] struct {
	del bool
	nx  *node[K]
}

func (u *state[K]) changeDel() *state[K] {
	return &state[K]{true, u.nx}
}

type node[K Maps.Hashable] struct {
	k    K
	v    unsafe.Pointer
	hash uint
	s    unsafe.Pointer
}

func (u *node[K]) next() (*node[K], unsafe.Pointer) {
	for {
		curStPtr := atomic.LoadPointer(&u.s)
		if nxNode := (*state[K])(curStPtr).nx; nxNode != nil {
			if nxSt := (*state[K])(atomic.LoadPointer(&nxNode.s)); nxSt.del {
				atomic.CompareAndSwapPointer(&u.s, curStPtr, unsafe.Pointer(&state[K]{false, nxSt.nx}))
			} else {
				if nxNode.hash < u.hash {
					println("error")
				}
				return nxNode, curStPtr
			}
		} else {
			return nil, nil
		}
	}
}

func (u *node[K]) deleted() bool {
	return (*state[K])(atomic.LoadPointer(&u.s)).del
}

func (u *node[K]) delete() bool {
	for curStPtr := atomic.LoadPointer(&u.s); ; curStPtr = atomic.LoadPointer(&u.s) {
		if curSt := (*state[K])(curStPtr); curSt.del {
			return false
		} else {
			if atomic.CompareAndSwapPointer(&u.s, curStPtr, unsafe.Pointer(curSt.changeDel())) {
				return true
			}
		}
	}
}

func (u *node[K]) getVPtr() unsafe.Pointer {
	return atomic.LoadPointer(&u.v)
}

func (u *node[K]) setVPtr(newPtr unsafe.Pointer) {
	atomic.StorePointer(&u.v, newPtr)
}

func (u *node[K]) isRelay() bool {
	return u.v == nil
}

func (u *node[K]) String() string {
	t := (*state[K])(atomic.LoadPointer(&u.s))
	return fmt.Sprintf("key: %#v; val: %#v; hash: %d; del: %t; next: %s", u.k, u.getVPtr(), u.hash, t.del, t.nx)
}
