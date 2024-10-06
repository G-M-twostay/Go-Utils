package v2

import (
	"github.com/g-m-twostay/go-utils/Maps/internal"
	"sync/atomic"
	"unsafe"
)

const (
	deletedMask = 1 << iota
	relayMask   //1 is relay, 0 is normal
	_           = uint(unsafe.Alignof(relay{}.next) - 4)
)

func addr(tagged unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(tagged) &^ (deletedMask | relayMask))
}
func isRelay(tagged unsafe.Pointer) bool {
	return uintptr(tagged)&relayMask != 0
}

type relay struct {
	hash uint
	next unsafe.Pointer //0 bit indicate whether this node is marked. 1 bit indicate whether next node is relay.
}

//go:nosplit
func (r *relay) mark() bool {
	return atomic.OrUintptr((*uintptr)(unsafe.Pointer(&r.next)), deletedMask)&deletedMask == 0
}

func (r *relay) tryLink(old, new unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&r.next, old, new)
}

// crawl gives 2 consecutive valid nodes. when it encounters logically deleted nodes, it tries to remove it. if removing failed because left node is deleted, it backtracks using first using path and then using fb. otherwise it reloads right and see whether a remove is still necessary. this is used for inserting nodes.
func (left *relay) crawl(path *internal.EvictStack, fb func() *relay) (*relay, unsafe.Pointer) {
retry:
	right := atomic.LoadPointer(&left.next)
	if uintptr(right)&deletedMask != 0 { //check if left is logically deleted.
		if left = (*relay)(path.Pop()); left == nil { //try to backtrack using path.
			left = fb() //path is empty, use fallback.
		}
		goto retry //reload right using new left.
	}
	if right == nil {
		return left, nil
	}
	if right2 := atomic.LoadUintptr((*uintptr)(unsafe.Pointer(&(*relay)(addr(right)).next))); right2&deletedMask == 0 {
		return left, right //right points to valid right2, return.
	} else { //right2 is logically deleted, try to remove it.
		for right2 &^= deletedMask; right2 != 0; { //find right2 whose right(right3) is valid.
			right3 := atomic.LoadUintptr((*uintptr)(unsafe.Pointer(&(*relay)(unsafe.Pointer(right2 &^ relayMask)).next)))
			if right3&deletedMask == 0 {
				break
			}
			right2 = right3 &^ deletedMask
		}
		if atomic.CompareAndSwapUintptr((*uintptr)(unsafe.Pointer(&left.next)), uintptr(right), right2) {
			return left, unsafe.Pointer(right2)
		}
		goto retry //failed, reload right for new right2.
	}
}

// walk gives the next valid node. walk is the weaker and faster version of crawl. when it encounters a logically deleted node, it attempts to remove it. regardless of whether removal is successful, it always moves on.
func (r *relay) walk() unsafe.Pointer {
	right := atomic.LoadPointer(&r.next)
	if uintptr(right) <= deletedMask {
		return nil
	}
	if right2 := atomic.LoadUintptr((*uintptr)(unsafe.Pointer(&(*relay)(addr(right)).next))); right2&deletedMask == 0 {
		return right //right2 is valid, return.
	} else { //right2 is logically deleted, try to remove it.
		for right2 &^= deletedMask; right2 != 0; { //find right2 whose right(right3) is valid.
			right3 := atomic.LoadUintptr((*uintptr)(unsafe.Pointer(&(*relay)(unsafe.Pointer(right2 &^ relayMask)).next)))
			if right3&deletedMask == 0 {
				break
			}
			right2 = right3 &^ deletedMask
		}
		atomic.CompareAndSwapUintptr((*uintptr)(unsafe.Pointer(&r.next)), uintptr(right)&^deletedMask, right2) //attempt to remove right2 when right is valid.
		return unsafe.Pointer(right2)                                                                          //always return.
	}
}

type ptrNode[K any] struct {
	relay
	val unsafe.Pointer
	key K
}

type valNode[K any, V ~uint32 | ~int32 | ~uint64 | ~int64 | ~uintptr] struct {
	relay
	key     K
	val     V
	version atomic.Uint32
}
