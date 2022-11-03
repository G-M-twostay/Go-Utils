package ChainMap

import (
	"GMUtils/Maps"
	"sync/atomic"
	"unsafe"
)

type head[K Maps.Hashable, V any] struct {
	nx unsafe.Pointer
}

type chain[K Maps.Hashable, V any] struct {
	head[K, V]
	k   K
	v   V
	del uint32
}

func (u *head[K, V]) next() *chain[K, V] {
	return (*chain[K, V])(u.nextPtr())
}

func (u *head[K, V]) nextPtr() unsafe.Pointer {
	for oldNext := atomic.LoadPointer(&u.nx); oldNext != nil; oldNext = atomic.LoadPointer(&u.nx) { //find the next node if there exists one
		if t := (*chain[K, V])(oldNext); atomic.LoadUint32(&t.del) != 0 {
			atomic.CompareAndSwapPointer(&u.nx, oldNext, atomic.LoadPointer(&t.nx)) //current node is marked, try to delete it
		} else {
			return oldNext
		}
	}
	return nil
}

func (u *chain[K, V]) delete() {
	atomic.StoreUint32(&u.del, 1)
}

type Hold[K Maps.Hashable, V any] struct {
	key  K
	val  V
	hash uint64
}

func (u Hold[K, V]) makeNode(nx unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(&chain[K, V]{head[K, V]{nx}, u.key, u.val, 0})
}
func (u Hold[K, V]) isKey(k Maps.Hashable) bool {
	return u.key.Equal(k)
}
