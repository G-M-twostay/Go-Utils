package Maps

import (
	"encoding/binary"
	"hash/maphash"
	"sync/atomic"
)

type baseMap[K Hashable, V any] struct {
	buckets   []head[K, V]
	sz        atomic.Uint64
	low, high float32
	seed      maphash.Seed
}

func (u baseMap[K, V]) toIndex(hash int64, bLen uint64) uint64 {
	b := make([]byte, 8)
	binary.PutVarint(b, hash)
	return maphash.Bytes(u.seed, b) & (bLen - 1)
}

func (u *baseMap[K, V]) put(key K, val V) (added bool) {
	pre := &u.buckets[u.toIndex(key.Hash(), uint64(len(u.buckets)))]
	cur := pre.next()
	for ; cur != nil && !cur.k.Equal(key); cur = cur.next() {
		pre = &cur.head
	}
	if cur == nil {
		t := new(chain[K, V])
		t.k, t.v = key, val
		pre.addAfter(t)
		u.sz.Add(1)
		added = true
	} else {
		cur.v = val
	}
	return
}

func (u *baseMap[K, V]) getOrPut(k K, v V) (old V, putted bool) {
	pre := &u.buckets[u.toIndex(k.Hash(), uint64(len(u.buckets)))]
	cur := pre.next()
	for ; cur != nil && !cur.k.Equal(k); cur = cur.next() {
		pre = &cur.head
	}
	if cur == nil {
		t := new(chain[K, V])
		t.k, t.v = k, v
		pre.addAfter(t)
		u.sz.Add(1)
		putted = true
	} else {
		old = cur.v
	}
	return
}

func (u baseMap[K, V]) get(key K) (r V) {
	for cur := u.buckets[u.toIndex(key.Hash(), uint64(len(u.buckets)))].next(); cur != nil; cur = cur.next() {
		if cur.k.Equal(key) {
			r = cur.v
			break
		}
	}
	return
}

func (u baseMap[K, V]) hasKey(key K) bool {
	for cur := u.buckets[u.toIndex(key.Hash(), uint64(len(u.buckets)))].next(); cur != nil; cur = cur.next() {
		if cur.k.Equal(key) {
			return true
		}
	}
	return false
}

func (u *baseMap[K, V]) findRemove(key K) *chain[K, V] {
	cur := u.buckets[u.toIndex(key.Hash(), uint64(len(u.buckets)))].next()
	for ; cur != nil && !cur.k.Equal(key); cur = cur.next() {
	}
	return cur
}

func (u baseMap[K, V]) take() (k K, v V) {
	for _, h := range u.buckets {
		if t := h.next(); t != nil {
			k, v = t.k, t.v
			break
		}
	}
	return
}

func (u baseMap[K, V]) Size() uint {
	return uint(u.sz.Load())
}

func (u baseMap[K, V]) pairs() func() (K, V, bool) {
	var i uint = 0
	var c *chain[K, V] = nil
	return func() (k K, v V, b bool) {

		for {
			if c == nil {
				if i < uint(len(u.buckets)) {
					c = u.buckets[i].next()
					i++
				} else {
					break
				}
			} else {
				k, v, b = c.k, c.v, true
				c = c.next()
				break
			}
		}
		return
	}
}
