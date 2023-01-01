package BucketMap

import (
	"GMUtils/Maps"
	"golang.org/x/exp/constraints"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

// IntMap is a specialized version of BucketMap for integers. It avoids all the interface operations.
type IntMap[K constraints.Integer, V any] struct {
	rehash                         func(K) uint
	buckets                        atomic.Pointer[Maps.HashList[*intNode[K]]]
	size                           atomic.Uint64
	state                          atomic.Uint32
	minAvgLen, maxAvgLen, maxChunk byte
}

func MakeIntMap[K constraints.Integer, V any](minBucketLen, maxBucketLen byte, maxHash uint, hasher func(K) uint) *IntMap[K, V] {
	M := new(IntMap[K, V])

	M.minAvgLen, M.maxAvgLen = minBucketLen, maxBucketLen
	M.maxChunk = byte(bits.Len(maxHash))
	M.rehash = hasher

	t := []*intNode[K]{{hash: 0, v: unsafe.Pointer(new(relayLock)), flag: true}}
	M.buckets.Store(&Maps.HashList[*intNode[K]]{Array: t, Chunk: M.maxChunk})

	return M
}

func (u *IntMap[K, V]) Size() uint64 {
	return u.size.Load()
}

func (u *IntMap[K, V]) trySplit() {
	if u.state.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if u.Size()>>(u.maxChunk-s.Chunk) > uint64(u.maxAvgLen) {

			newBuckets := make([]*intNode[K], len(s.Array)<<1)

			for i, v := range s.Array {

				newBuckets[i<<1] = v

				newRelay := &intNode[K]{hash: (1<<s.Chunk)*uint(i) + (1 << (s.Chunk - 1)), v: unsafe.Pointer(new(relayLock)), flag: true}
				newBuckets[(i<<1)+1] = newRelay

				t := v.lock()
				t.RLock()
				for left, newRelayPtr := v, unsafe.Pointer(newRelay); ; {
					if rightPtr := left.Next(); rightPtr == nil {
						if left.tryLink(rightPtr, newRelay, newRelayPtr) {
							break
						}
					} else if right := (*intNode[K])(rightPtr); newRelay.hash <= right.hash {
						if left.tryLink(rightPtr, newRelay, newRelayPtr) {
							break
						}
					} else {
						left = right
					}
				}
				t.RUnlock()
			}

			u.buckets.Store(&Maps.HashList[*intNode[K]]{Array: newBuckets, Chunk: s.Chunk - 1})

		}
		u.state.Store(0)
	}
}

func (u *IntMap[K, V]) tryMerge() {
	if u.state.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if u.Size()>>(u.maxChunk-s.Chunk) < uint64(u.minAvgLen) && len(s.Array) > 1 {

			newBuckets := make([]*intNode[K], len(s.Array)>>1)

			for i := range newBuckets {
				newBuckets[i] = s.Array[i<<1]
			}

			u.buckets.Store(&Maps.HashList[*intNode[K]]{Array: newBuckets, Chunk: s.Chunk + 1})

			for i := 0; i < len(s.Array); i += 2 {
				t := s.Array[i].lock()
				t.RLock()
				for left := s.Array[i]; ; {
					rightPtr := left.Next()
					if right := (*intNode[K])(rightPtr); right.flag {
						if left.unlinkRelay(right, rightPtr) {
							break
						}
					} else {
						left = right
					}
				}
				t.RUnlock()
			}
		}
		u.state.Store(0)
	}
}

func (u *IntMap[K, V]) Store(key K, val V) {
	hash, vPtr := u.rehash(key), unsafe.Pointer(&val)

	left := u.buckets.Load().Get(hash)
	prevLock := left.lock()
	for ; !prevLock.safeRLock(); prevLock = left.lock() {
		prevLock.RUnlock()
		left = u.buckets.Load().Get(hash)
	}

	for {
		if rightPtr := left.Next(); rightPtr == nil {
			if left.tryLazyLink(nil, unsafe.Pointer(&intNode[K]{hash, vPtr, rightPtr, key, false})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if right := (*intNode[K])(rightPtr); hash < right.hash {
			if left.tryLazyLink(rightPtr, unsafe.Pointer(&intNode[K]{hash, vPtr, rightPtr, key, false})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if right.cmpKey(key) {
			prevLock.RUnlock()
			right.set(vPtr)
			return
		} else {
			left = right
		}

		if left.flag {
			if cl := left.lock(); cl != prevLock {
				prevLock.RUnlock()
				for prevLock = cl; !prevLock.safeRLock(); prevLock = left.lock() {
					prevLock.RUnlock()
					left = u.buckets.Load().Get(hash)
				}
			}
		}
	}
}

func (u *IntMap[K, V]) LoadPtrOrStore(key K, val V) (v *V, loaded bool) {
	hash, vPtr := u.rehash(key), unsafe.Pointer(&val)

	left := u.buckets.Load().Get(hash)
	prevLock := left.lock()
	for ; !prevLock.safeRLock(); prevLock = left.lock() {
		prevLock.RUnlock()
		left = u.buckets.Load().Get(hash)
	}

	for {
		if rightPtr := left.Next(); rightPtr == nil {
			if left.tryLazyLink(nil, unsafe.Pointer(&intNode[K]{hash, vPtr, rightPtr, key, false})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if right := (*intNode[K])(rightPtr); hash < right.hash {
			if left.tryLazyLink(rightPtr, unsafe.Pointer(&intNode[K]{hash, vPtr, rightPtr, key, false})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if right.cmpKey(key) {
			prevLock.RUnlock()
			return (*V)(right.get()), true
		} else {
			left = right
		}

		if left.flag {
			if cl := left.lock(); cl != prevLock {
				prevLock.RUnlock()
				for prevLock = cl; !prevLock.safeRLock(); prevLock = left.lock() {
					prevLock.RUnlock()
					left = u.buckets.Load().Get(hash)
				}
			}
		}
	}
}

func (u *IntMap[K, V]) LoadOrStore(key K, val V) (v V, loaded bool) {
	a, b := u.LoadPtrOrStore(key, val)
	if b {
		v = *a
	}
	return v, b
}

func (u *IntMap[K, V]) LoadPtr(key K) *V {
	hash := u.rehash(key)
	_, r, _ := u.buckets.Load().Get(hash).searchKey(key, hash)
	return (*V)(r.get())
}

func (u *IntMap[K, V]) Load(key K) (V, bool) {
	hash := u.rehash(key)
	_, r, f := u.buckets.Load().Get(hash).searchKey(key, hash)
	var v V
	if f {
		v = *(*V)(r.get())
	}
	return v, f
}

func (u *IntMap[K, V]) HasKey(key K) bool {
	hash := u.rehash(key)
	_, _, f := u.buckets.Load().Get(hash).searchKey(key, hash)
	return f
}

func (u *IntMap[K, V]) LoadPtrAndDelete(key K) (v *V, loaded bool) {
	hash := u.rehash(key)

	left := u.buckets.Load().Get(hash)
	prevLock := left.lock()
	for ; !prevLock.safeLock(); prevLock = left.lock() {
		prevLock.Unlock()
		left = u.buckets.Load().Get(hash)
	}

	for {
		if rightPtr := left.nx; rightPtr == nil {
			prevLock.Unlock()
			return
		} else if right := (*intNode[K])(rightPtr); hash < right.hash {
			prevLock.Unlock()
			return
		} else if right.cmpKey(key) {
			left.dangerUnlink(right)
			prevLock.Unlock()
			u.size.Add(^uint64(1 - 1))
			u.tryMerge()
			return (*V)(right.get()), true
		} else {
			left = right
		}

		if left.flag {
			prevLock.Unlock()
			for prevLock = left.lock(); !prevLock.safeLock(); prevLock = left.lock() {
				prevLock.Unlock()
				left = u.buckets.Load().Get(hash)
			}
		}
	}
}

func (u *IntMap[K, V]) LoadAndDelete(key K) (v V, loaded bool) {
	a, b := u.LoadPtrAndDelete(key)
	if b {
		v = *a
	}
	return v, b
}

func (u *IntMap[K, V]) Delete(key K) {
	u.LoadPtrAndDelete(key)
}

func (u *IntMap[K, V]) RangePtr(f func(K, *V) bool) {
	for cur := u.buckets.Load().Get(0); cur != nil; cur = (*intNode[K])(cur.Next()) {
		if !cur.flag {
			if !f(cur.k, (*V)(cur.get())) {
				break
			}
		}
	}
}

func (u *IntMap[K, V]) Range(f func(K, V) bool) {
	u.RangePtr(func(k K, v *V) bool {
		return f(k, *v)
	})
}

func (u *IntMap[K, V]) Take() (key K, val V) {
	a, b := u.TakePtr()
	return a, *b
}

func (u *IntMap[K, V]) TakePtr() (key K, val *V) {
	if firstPtr := u.buckets.Load().Get(0).Next(); firstPtr != nil {
		first := (*intNode[K])(firstPtr)
		key, val = first.k, (*V)(first.get())
	}
	return
}

func (u *IntMap[K, V]) Set(key K, val V) (v *V) {
	hash := u.rehash(key)
	if _, r, f := u.buckets.Load().Get(hash).searchKey(key, hash); f {
		v = (*V)(r.get())
		r.set(unsafe.Pointer(&val))
	}
	return
}
