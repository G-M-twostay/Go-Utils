package Maps

import (
	"encoding/binary"
	"hash/maphash"
	"sync"
)

type headNode[K Hashable, V any] struct {
	nx *linkNode[K, V]
}

type ConcHMap[K Hashable, V any] struct {
	buckets                []headNode[K, V]
	sz                     uint
	low, high              float32
	seed                   maphash.Seed
	resizing, doneResizing sync.Mutex
}

func (u ConcHMap[K, V]) toIndex(hash int64, bLen uint) uint {
	b := make([]byte, 8)
	binary.PutVarint(b, hash)
	return uint(maphash.Bytes(u.seed, b)) & (bLen - 1)
}

func (u *ConcHMap[K, V]) resize(newSize uint) {
	nb := make([]headNode[K, V], newSize)
	for _, h := range u.buckets {
		for cur := h.nx; cur != nil; {
			t := cur.nx
			nh := nb[u.toIndex(cur.k.Hash(), newSize)]
			cur.nx = nh.nx
			nh.nx = cur
			cur = t
		}
	}
	u.buckets = nb
}

func (u *ConcHMap[K, V]) Put(key K, val V) (oldVal V) {
	pre := u.buckets[u.toIndex(key.Hash(), uint(len(u.buckets)))]
	cur := pre.nx
	for ; cur != nil && !cur.k.Equal(key); cur = cur.nx {
		pre = cur.headNode
	}
	if cur == nil {
		pre.nx = &linkNode[K, V]{headNode[K, V]{}, key, val}
	} else {
		oldVal = cur.v
		cur.v = val
	}
	return
}

func (u *ConcHMap[K, V]) Remove(key K) bool {
	pre := u.buckets[u.toIndex(key.Hash())]
	cur := pre.nx
	for ; cur != nil && !cur.k.Equal(key); cur = cur.nx {
		pre = cur.headNode
	}
	if cur != nil {
		pre.nx = cur.nx
		return true
	}
	return false
}
