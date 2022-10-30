package Maps

import (
	"encoding/binary"
	"hash/maphash"
	"sync"
	"sync/atomic"
	"unsafe"
)

type ConcArrMap1[K Hashable, V any] struct {
	buckets   []head[K, V]
	sz        atomic.Uint64
	low, high float32
	seed      maphash.Seed
	l0, l2    sync.RWMutex //l0: operations can't be conducted while resizing. l2: can't access buckets while it is being changed in resize.
	l1        sync.Mutex   //only one thread will need to call resize at a time. indicates whether some thread has called resize.
}

func (u ConcArrMap1[K, V]) toIndex(hash int64, bLen uint64) uint64 {
	b := make([]byte, 8)
	binary.PutVarint(b, hash)
	return maphash.Bytes(u.seed, b) & (bLen - 1)
}

func MakeConcArrMap[K Hashable, V any](sz uint) *ConcArrMap1[K, V] {
	t := new(ConcArrMap1[K, V])
	t.buckets = make([]head[K, V], sz)
	t.high = 0.9
	t.seed = maphash.MakeSeed()
	return t
}

func (u *ConcArrMap1[K, V]) resize(newSize uint64) {
	u.l0.Lock()
	defer u.l0.Unlock()
	nb := make([]head[K, V], newSize)
	for _, h := range u.buckets {
		for cur := h.next(); cur != nil; cur = cur.next() {
			nh := &nb[u.toIndex(cur.k.Hash(), newSize)]
			t := *cur
			t.nx = nh.nx
			nh.nx = unsafe.Pointer(&t)
		}
	}
	u.l2.Lock()
	defer u.l2.Unlock()
	u.buckets = nb
}

func (u *ConcArrMap1[K, V]) Put(key K, val V) (oldVal V) {
	u.l0.RLock()
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
	} else {
		oldVal = cur.v
		cur.v = val
	}
	u.l0.RUnlock()
	if float32(u.sz.Load())/float32(len(u.buckets)) > u.high {
		if u.l1.TryLock() {
			//no thread has performed this, we will perform it
			u.resize(uint64(len(u.buckets)) << 1) //len*2
			u.l1.Unlock()
		} //else: some thread has already performed it.
	}
	return
}

func (u ConcArrMap1[K, V]) Get(key K) (r V) {
	u.l2.RLock()
	defer u.l2.RUnlock()
	for cur := u.buckets[u.toIndex(key.Hash(), uint64(len(u.buckets)))].next(); cur != nil; cur = cur.next() {
		if cur.k.Equal(key) {
			r = cur.v
			break
		}
	}
	return
}

func (u ConcArrMap1[K, V]) HasKey(key K) bool {
	u.l2.RLock()
	defer u.l2.RUnlock()
	for cur := u.buckets[u.toIndex(key.Hash(), uint64(len(u.buckets)))].next(); cur != nil; cur = cur.next() {
		if cur.k.Equal(key) {
			return true
		}
	}
	return false
}

func (u *ConcArrMap1[K, V]) Remove(key K) bool {
	u.l0.RLock()
	cur := u.buckets[u.toIndex(key.Hash(), uint64(len(u.buckets)))].next()
	for ; cur != nil && !cur.k.Equal(key); cur = cur.next() {
	}
	if cur != nil {
		cur.delete()
		u.l0.RUnlock()
		u.sz.Add(^uint64(0))
		if float32(u.sz.Load())/float32(len(u.buckets)) < u.low {
			if u.l1.TryLock() {
				u.resize(uint64(len(u.buckets)) >> 1) //len/2
				u.l1.Unlock()
			}
		}
		return true
	}
	u.l0.RUnlock()
	return false
}

func (u ConcArrMap1[K, V]) Take() (k K, v V) {
	u.l2.RLock()
	defer u.l2.RUnlock()
	for _, h := range u.buckets {
		if t := h.next(); t != nil {
			k, v = t.k, t.v
			break
		}
	}
	return
}

func (u ConcArrMap1[K, V]) Size() uint {
	return uint(u.sz.Load())
}

func (u ConcArrMap1[K, V]) Pairs() func() (K, V, bool) {
	var i uint = 0
	var c *chain[K, V] = nil
	u.l2.RLock()
	defer u.l2.RUnlock()
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
