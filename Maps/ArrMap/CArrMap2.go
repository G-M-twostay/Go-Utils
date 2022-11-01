package ArrMap

import (
	"GMUtils/Maps"
	"hash/maphash"
	"sync"
)

// There are 1 lock obseravable: rw1.
// Put: rw1.read
// remove: rw1.read
// Get: rw1.read
// HasKey: rw1.read
// Take: rw1.read
// Pairs: rw1.read
// resize: rw1.write.
type ConcArrMap2[K Maps.Hashable, V any] struct {
	baseMap[K, V]
	l0 sync.RWMutex //l0: operations can't be conducted while resizing.
	l1 sync.Mutex   //only one thread will need to call resize at a time. indicates whether some thread has called resize.
}

func MakeConcArrMap2[K Maps.Hashable, V any](sz uint) *ConcArrMap1[K, V] {
	t := new(ConcArrMap1[K, V])
	t.buckets = make([]head[K, V], sz)
	t.high = 0.9
	t.seed = maphash.MakeSeed()
	return t
}

func (u *ConcArrMap2[K, V]) resize(newSize uint64) {
	nb := make([]head[K, V], newSize)
	u.l0.Lock()
	defer u.l0.Unlock()
	for _, h := range u.buckets {
		for curPtr := h.nextPtr(); curPtr != nil; {
			cur := (*chain[K, V])(curPtr)
			nh := &nb[u.toIndex(cur.k.Hash(), newSize)]
			t := cur.nextPtr()
			cur.nx = nh.nx
			nh.nx = curPtr
			curPtr = t
		}
	}
	u.buckets = nb
}

func (u *ConcArrMap2[K, V]) tryGrow() {
	if float32(u.Size())/float32(len(u.buckets)) > u.high {
		if u.l1.TryLock() {
			//no thread has performed this, we will perform it
			u.resize(uint64(len(u.buckets)) << 1) //len*2
			u.l1.Unlock()
		} //else: some thread has already performed it.
	}
}

func (u *ConcArrMap2[K, V]) tryShrink() {
	if float32(u.Size())/float32(len(u.buckets)) < u.low {
		if u.l1.TryLock() {
			u.resize(uint64(len(u.buckets)) >> 1) //len/2
			u.l1.Unlock()
		}
	}
}

func (u *ConcArrMap2[K, V]) Put(key K, val V) {
	u.l0.RLock()
	added := u.put(key, val)
	u.l0.RUnlock()
	if added {
		u.tryGrow()
	}
}

func (u ConcArrMap2[K, V]) Get(key K) V {
	u.l0.RLock()
	defer u.l0.RUnlock()
	return u.get(key)
}

func (u *ConcArrMap2[K, V]) GetOrPut(key K, val V) (V, bool) {
	u.l0.RLock()
	a, b := u.getOrPut(key, val)
	u.l0.RUnlock()
	if b {
		u.tryGrow()
	}
	return a, b
}

func (u ConcArrMap2[K, V]) HasKey(key K) bool {
	u.l0.RLock()
	defer u.l0.RUnlock()
	return u.hasKey(key)
}

func (u *ConcArrMap2[K, V]) Remove(key K) {
	u.GetAndRmv(key)
}

func (u *ConcArrMap2[K, V]) GetAndRmv(key K) (old V, removed bool) {
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

func (u ConcArrMap2[K, V]) Take() (K, V) {
	u.l0.RLock()
	defer u.l0.RUnlock()
	return u.take()
}

func (u ConcArrMap2[K, V]) Pairs() func() (K, V, bool) {
	u.l0.RLock()
	defer u.l0.RUnlock()
	return u.pairs()
}
