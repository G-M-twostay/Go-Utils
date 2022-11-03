package ChainMap

import (
	"GMUtils/Maps"
	"sync/atomic"
	"unsafe"
)

type chain[K Maps.Hashable, V any] struct {
	nx    unsafe.Pointer
	k     K
	v     unsafe.Pointer
	state uint64 //first bit indicates deleted or not. last 63 bits indicate the hash value.
}

func (u *chain[K, V]) next() *chain[K, V] {
	return (*chain[K, V])(u.nextPtr())
}

func (u *chain[K, V]) nextPtr() unsafe.Pointer {
	for oldNext := atomic.LoadPointer(&u.nx); oldNext != nil; oldNext = atomic.LoadPointer(&u.nx) { //find the next node if there exists one
		if t := (*chain[K, V])(oldNext); t.deleted() {
			atomic.CompareAndSwapPointer(&u.nx, oldNext, atomic.LoadPointer(&t.nx)) //current node is marked, try to delete it
		} else {
			return oldNext
		}
	}
	return nil
}

func (u chain[K, V]) deleted() bool {
	return atomic.LoadUint64(&u.state)>>63 == 1
}

func (u *chain[K, V]) delete() {
	cur := atomic.LoadUint64(&u.state)
	atomic.StoreUint64(&u.state, cur|(1<<63))
}

func (u chain[K, V]) hash() uint64 {
	return atomic.LoadUint64(&u.state) &^ (1 << 63)
}

func (u chain[K, V]) compareKey(o K) bool {
	if u.k == nil {
		return false
	} else {
		return u.k.Equal(o)
	}
}

func (u chain[K, V]) valuePtr() *V {
	return (*V)(u.v)
}

func (u *chain[K, V]) setValuePtr(newPtr *V) {
	atomic.StorePointer(&u.v, unsafe.Pointer(newPtr))
}
