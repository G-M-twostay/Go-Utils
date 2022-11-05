package ChainMap

import (
	"GMUtils/Maps"
	"sync/atomic"
	"unsafe"
)

const (
	relayState   uint32 = 0
	normalState  uint32 = 1
	deletedState uint32 = 2
)

type node[K Maps.Hashable] struct {
	nx    unsafe.Pointer
	k     K
	v     unsafe.Pointer
	hash  uint
	state uint32 //0:not deleted relay. 1: not deleted normal. 2: deleted
}

func (u *node[K]) next() *node[K] {
	return (*node[K])(u.nextPtr())
}

func (u *node[K]) nextPtr() unsafe.Pointer {
	for oldNext := atomic.LoadPointer(&u.nx); oldNext != nil; oldNext = atomic.LoadPointer(&u.nx) { //find the next node if there exists one
		if t := (*node[K])(oldNext); t.deleted() {
			atomic.CompareAndSwapPointer(&u.nx, oldNext, atomic.LoadPointer(&t.nx)) //current node is marked, try to delete it
		} else {
			return oldNext
		}
	}
	return nil
}

func (u node[K]) deleted() bool {
	return atomic.LoadUint32(&u.state) == deletedState
}

func (u *node[K]) delete() {
	atomic.StoreUint32(&u.state, deletedState)
}

func (u node[K]) valuePtr() unsafe.Pointer {
	return atomic.LoadPointer(&u.v)
}

func (u *node[K]) setValuePtr(newPtr unsafe.Pointer) {
	atomic.StorePointer(&u.v, newPtr)
}

func (u node[K]) isRelay() bool {
	//println("is relay:", atomic.LoadUint32(&u.state) == relayState)
	return atomic.LoadUint32(&u.state) == relayState
}
