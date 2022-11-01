package ArrMap

import (
	"GMUtils/Maps"
	"sync/atomic"
	"unsafe"
)

// didn't use atomic.Pointer to allow easy copy and unsafe set(when there is a lock protecting it).
type head[K Maps.Hashable, V any] struct {
	nx unsafe.Pointer
}

type chain[K Maps.Hashable, V any] struct {
	head[K, V]
	k   K
	v   V
	del bool
}

type hold[K Maps.Hashable, V any] interface {
	next() *chain[K, V]
	nextPtr() unsafe.Pointer
	addAfter(n *chain[K, V])
}

// given *a, a->nx=b
// result a->next=n; n->next=b
func (u *head[K, V]) addAfter(n *chain[K, V]) {
	for added, t := false, unsafe.Pointer(n); !added; {
		oldNext := u.nextPtr()
		n.nx = oldNext                                          //set n->next=b
		added = atomic.CompareAndSwapPointer(&u.nx, oldNext, t) //try to make a->next=n
	}
}

func (u *head[K, V]) next() *chain[K, V] {
	return (*chain[K, V])(u.nextPtr())
}

func (u *head[K, V]) nextPtr() unsafe.Pointer {
	for oldNext := atomic.LoadPointer(&u.nx); oldNext != nil; oldNext = atomic.LoadPointer(&u.nx) { //find the next node if there exists one
		if t := (*chain[K, V])(oldNext); t.del {
			atomic.CompareAndSwapPointer(&u.nx, oldNext, atomic.LoadPointer(&t.nx)) //current node is marked, try to delete it
		} else {
			return oldNext
		}
	}
	return nil
}

func (u *chain[K, V]) delete() {
	//atomic.StoreUint32(&u.del, 1)
	u.del = true
}
