package Maps

import (
	"encoding/binary"
	"hash/maphash"
	"sync"
	"sync/atomic"
	"unsafe"
)

type ConcHMap[K Hashable, V any] struct {
	buckets   []head[K, V]
	sz        atomic.Uint32
	low, high float32
	seed      maphash.Seed
	l0        sync.RWMutex //operations can't be conducted while resizing
	l1        sync.Mutex   //only one thread will need to call resize at a time
}

func (u ConcHMap[K, V]) toIndex(hash int64, bLen uint) uint {
	b := make([]byte, 8)
	binary.PutVarint(b, hash)
	return uint(maphash.Bytes(u.seed, b)) & (bLen - 1)
}

func MakeConcArrMap[K Hashable, V any](sz uint) *ConcHMap[K, V] {
	t := new(ConcHMap[K, V])
	t.buckets = make([]head[K, V], sz)
	t.high = 0.9
	t.seed = maphash.MakeSeed()
	return t
}

func (u *ConcHMap[K, V]) resize(newSize uint) {
	u.l0.Lock()
	defer u.l0.Unlock()
	nb := make([]head[K, V], newSize)
	for _, h := range u.buckets {
		for cur := h.next(); cur != nil; cur = cur.next() {
			nh := &nb[u.toIndex(cur.k.Hash(), newSize)]
			t := chain[K, V]{&head[K, V]{nh.nx}, cur.k, cur.v, false}
			nh.nx = unsafe.Pointer(&t)
		}
	}
	u.buckets = nb
}

func (u *ConcHMap[K, V]) Put(key K, val V) (oldVal V) {
	u.l0.RLock()
	pre := &u.buckets[u.toIndex(key.Hash(), uint(len(u.buckets)))]
	cur := pre.next()
	for ; cur != nil && !cur.k.Equal(key); cur = cur.next() {
		pre = cur.head
	}
	if cur == nil {
		pre.addAfter(&chain[K, V]{new(head[K, V]), key, val, false})
	} else {
		oldVal = cur.v
		cur.v = val
	}
	u.l0.RUnlock()
	u.sz.Add(1)
	if float32(u.sz.Load())/float32(len(u.buckets)) > u.high {
		if u.l1.TryLock() {
			u.resize(uint(len(u.buckets)) * 2)
			u.l1.Unlock()
		}
	}
	return
}

func (u ConcHMap[K, V]) Get(key K) (r V) {
	for cur := u.buckets[u.toIndex(key.Hash(), uint(len(u.buckets)))].next(); cur != nil; cur = cur.next() {
		if cur.k.Equal(key) {
			r = cur.v
			break
		}
	}
	return

}

func (u *ConcHMap[K, V]) Remove(key K) bool {
	u.l0.RLock()
	cur := u.buckets[u.toIndex(key.Hash(), uint(len(u.buckets)))].next()
	for ; cur != nil && !cur.k.Equal(key); cur = cur.next() {
	}
	u.l0.RUnlock()
	if cur != nil {
		cur.delete()
		u.sz.Add(^uint32(0))
		if float32(u.sz.Load())/float32(len(u.buckets)) < u.low {
			if u.l1.TryLock() {
				u.resize(uint(len(u.buckets)) / 2)
				u.l1.Unlock()
			}
		}
		return true
	}
	return false
}
