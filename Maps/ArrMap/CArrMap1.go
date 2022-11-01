package ArrMap

import (
	"GMUtils/Maps"
	"hash/maphash"
	"sync"
	"unsafe"
)

// There are 2 locks obseravable: rw1,rw2.
// Put: rw1.read
// remove: rw1.read
// Get: rw2.read
// HasKey: rw2.read
// Take: rw2.read
// Pairs: rw2.read
// resize: first part: rw1.write, second part: rw2.write; unlocks both after second part is done.
type ConcArrMap1[K Maps.Hashable, V any] struct {
	baseMap[K, V]
	l0, l2 sync.RWMutex //l0: operations can't be conducted while resizing. l2: can't access buckets while it is being changed in resize.
	l1     sync.Mutex   //only one thread will need to call resize at a time. indicates whether some thread has called resize.
}

func MakeConcArrMap[K Maps.Hashable, V any](sz uint) *ConcArrMap1[K, V] {
	t := new(ConcArrMap1[K, V])
	t.buckets = make([]head[K, V], sz)
	t.high = 0.9
	t.seed = maphash.MakeSeed()
	return t
}

func (u *ConcArrMap1[K, V]) resize(newSize uint64) {
	nb := make([]head[K, V], newSize)
	u.l0.Lock()
	defer u.l0.Unlock()
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

func (u *ConcArrMap1[K, V]) tryGrow() {
	if float32(u.Size())/float32(len(u.buckets)) > u.high {
		if u.l1.TryLock() {
			//no thread has performed this, we will perform it
			u.resize(uint64(len(u.buckets)) << 1) //len*2
			u.l1.Unlock()
		} //else: some thread has already performed it.
	}
}

func (u *ConcArrMap1[K, V]) tryShrink() {
	if float32(u.Size())/float32(len(u.buckets)) < u.low {
		if u.l1.TryLock() {
			u.resize(uint64(len(u.buckets)) >> 1) //len/2
			u.l1.Unlock()
		}
	}
}

func (u *ConcArrMap1[K, V]) Put(key K, val V) {
	u.l0.RLock()
	added := u.put(key, val)
	u.l0.RUnlock()
	if added {
		u.tryGrow()
	}
}

func (u ConcArrMap1[K, V]) Get(key K) V {
	u.l2.RLock()
	defer u.l2.RUnlock()
	return u.get(key)
}

func (u *ConcArrMap1[K, V]) GetOrPut(key K, val V) (V, bool) {
	u.l0.RLock()
	a, b := u.getOrPut(key, val)
	u.l0.RUnlock()
	if b {
		u.tryGrow()
	}
	return a, b
}

func (u ConcArrMap1[K, V]) HasKey(key K) bool {
	u.l2.RLock()
	defer u.l2.RUnlock()
	return u.hasKey(key)
}

func (u *ConcArrMap1[K, V]) Remove(key K) {
	u.GetAndRmv(key)
}

func (u *ConcArrMap1[K, V]) GetAndRmv(key K) (old V, removed bool) {
	u.l0.RLock()
	n := u.findRemove(key)
	removed = n != nil
	if removed {
		n.delete()
		old = n.v
		u.sz.Add(^uint64(0))
	}
	u.l0.RUnlock()
	if removed {
		u.tryShrink()
	}
	return
}

func (u ConcArrMap1[K, V]) Take() (K, V) {
	u.l2.RLock()
	defer u.l2.RUnlock()
	return u.take()
}

func (u ConcArrMap1[K, V]) Pairs() func() (K, V, bool) {
	u.l2.RLock()
	defer u.l2.RUnlock()
	return u.pairs()
}
