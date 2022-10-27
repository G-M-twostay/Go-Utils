package Maps

import (
	"encoding/binary"
	"hash/maphash"
	"sync"
)

type ConcHMap[K Hashable, V any] struct {
	buckets                []head[K, V]
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
	nb := make([]head[K, V], newSize)
	for _, h := range u.buckets {
		for cur := h.nx.Load(); cur != nil; cur = cur.nx.Load() {
			nh := &nb[u.toIndex(cur.k.Hash(), newSize)].nx
			t := new(node[K, V])
			t.nx.Store(nh.Load())
			t.k, t.v = cur.k, cur.v
			nh.Store(t)
		}
	}
	u.buckets = nb
}

func (u *ConcHMap[K, V]) Put(key K, val V) (oldVal V) {
	pre := u.buckets[u.toIndex(key.Hash(), uint(len(u.buckets)))]
	cur := pre.next()
	for ; cur != nil && !cur.k.Equal(key); cur = cur.next() {
		pre = cur.head
	}
	if cur == nil {
		t := new(node[K, V])
		t.k, t.v = key, val
		pre.addAtEnd(t)
	} else {
		oldVal = cur.v
		cur.v = val
	}
	return
}

func (u *ConcHMap[K, V]) Remove(key K) bool {
	pre := u.buckets[u.toIndex(key.Hash(), uint(len(u.buckets)))]
	cur := pre.next()
	for ; cur != nil && !cur.k.Equal(key); cur = cur.next() {
		pre = cur.head
	}
	if cur != nil {
		cur.delete()
		return true
	}
	return false
}
