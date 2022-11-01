package ChainMap

import (
	"GMUtils/Maps"
	"GMUtils/Maps/ArrMap"
	"sync/atomic"
	"unsafe"
)

type ChainMap[K Maps.Hashable, V any] struct {
	buckets []ArrMap.head[K, V]
	chunk   uint8
}

func (u *ChainMap[K, V]) searchHash(h *ArrMap.head[K, V], hash uint64) *ArrMap.chain[K, V] {
	var pre *ArrMap.chain[K, V] = nil
	cur := h.next()
	for ; cur != nil && uint64(cur.k.Hash()) < hash; cur = cur.next() {
		pre = cur
	}
	return pre
}

func (u *ChainMap[K, V]) putAt(h *ArrMap.head[K, V], k K, v V) {
begin:
	var preh *ArrMap.head[K, V] = h
	cur := h.next()
	for ; cur != nil && uint64(cur.k.Hash()) <= uint64(k.Hash()) && !k.Equal(cur.k); cur = cur.next() {
		preh = &cur.head
	}
	if cur != nil && cur.k.Equal(k) {
		cur.v = v
		return
	} else {
		added := false
		for !added {
			newCurPtr := preh.nextPtr()
			if newCurPtr == nil || uint64(k.Hash()) <= uint64((*ArrMap.chain[K, V])(newCurPtr).k.Hash()) {
				t := ArrMap.chain[K, V]{ArrMap.head[K, V]{newCurPtr}, k, v, false}
				added = atomic.CompareAndSwapPointer(&preh.nx, newCurPtr, unsafe.Pointer(&t))
			} else {
				goto begin
			}
		}

	}
}
