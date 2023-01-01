package BucketMap

import (
	"GMUtils/Maps"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

type BucketMap[K Maps.Hashable, V any] struct {
	buckets                        atomic.Pointer[Maps.HashList[*node[K]]]
	size                           atomic.Uint64
	resizing                       atomic.Uint32
	minAvgLen, maxAvgLen, maxChunk byte
}

func MakeBucketMap[K Maps.Hashable, V any](minBucketLen, maxBucketLen byte, maxHash uint) *BucketMap[K, V] {
	M := new(BucketMap[K, V])

	M.minAvgLen, M.maxAvgLen = minBucketLen, maxBucketLen
	M.maxChunk = byte(bits.Len(maxHash))

	t := []*node[K]{&node[K]{hash: 0, relayLock: new(relayLock)}}
	M.buckets.Store(&Maps.HashList[*node[K]]{Array: t, Chunk: M.maxChunk})

	return M
}

func (u *BucketMap[K, V]) rehash(k K) uint {
	return k.Hash()
}

func (u *BucketMap[K, V]) Size() uint64 {
	return u.size.Load()
}

func (u *BucketMap[K, V]) trySplit() {
	if u.resizing.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if u.Size()>>(u.maxChunk-s.Chunk) > uint64(u.maxAvgLen) {

			newBuckets := make([]*node[K], len(s.Array)<<1)

			for i, v := range s.Array {

				newBuckets[i<<1] = v

				newRelay := &node[K]{hash: (1<<s.Chunk)*uint(i) + (1 << (s.Chunk - 1)), relayLock: new(relayLock)}
				newBuckets[(i<<1)+1] = newRelay

				v.RLock()
				for left, newRelayPtr := v, unsafe.Pointer(newRelay); ; {
					if rightPtr := left.Next(); rightPtr == nil {
						if left.tryLink(nil, newRelay, newRelayPtr) {
							break
						}
					} else if right := (*node[K])(rightPtr); newRelay.hash <= right.hash {
						if left.tryLink(rightPtr, newRelay, newRelayPtr) {
							break
						}
					} else {
						left = right
					}
				}
				v.RUnlock()
			}

			u.buckets.Store(&Maps.HashList[*node[K]]{Array: newBuckets, Chunk: s.Chunk - 1})

		}
		u.resizing.Store(0)
	}
}

func (u *BucketMap[K, V]) tryMerge() {
	if u.resizing.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if u.Size()>>(u.maxChunk-s.Chunk) < uint64(u.minAvgLen) && len(s.Array) > 1 {

			newBuckets := make([]*node[K], len(s.Array)>>1)

			for i := range newBuckets {
				newBuckets[i] = s.Array[i<<1]
			}

			u.buckets.Store(&Maps.HashList[*node[K]]{Array: newBuckets, Chunk: s.Chunk + 1})

			for i := 0; i < len(s.Array); i += 2 {
				s.Array[i].RLock()
				for left := s.Array[i]; ; {
					rightPtr := left.Next()
					if right := (*node[K])(rightPtr); right.isRelay() {
						if left.unlinkRelay(right, rightPtr) {
							break
						}
					} else {
						left = right
					}
				}
				s.Array[i].RUnlock()
			}
		}
		u.resizing.Store(0)
	}
}

func (u *BucketMap[K, V]) Store(key K, val V) {
	hash, vPtr := u.rehash(key), unsafe.Pointer(&val)
	var prevLock *relayLock = nil
	for left := u.buckets.Load().Get(hash); ; {
		if left.isRelay() {
			if prevLock != nil {
				prevLock.RUnlock()
			}
			prevLock = left.relayLock
			if !prevLock.safeRLock() {
				left = u.buckets.Load().Get(hash)
				continue
			}
		}
		if rightPtr := left.Next(); rightPtr == nil {
			if left.tryLazyLink(nil, unsafe.Pointer(&node[K]{key, hash, vPtr, nil, nil})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if right := (*node[K])(rightPtr); hash < right.hash {
			if left.tryLazyLink(rightPtr, unsafe.Pointer(&node[K]{key, hash, vPtr, rightPtr, nil})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if hash == right.hash && right.cmpKey(key) {
			prevLock.RUnlock()
			right.set(vPtr)
			return
		} else {
			left = right
		}
	}
}

func (u *BucketMap[K, V]) LoadPtrOrStore(key K, val V) (v *V, loaded bool) {
	hash, vPtr := u.rehash(key), unsafe.Pointer(&val)
	var prevLock *relayLock = nil
	for left := u.buckets.Load().Get(hash); ; {
		if left.isRelay() {
			if prevLock != nil {
				prevLock.RUnlock()
			}
			prevLock = left.relayLock
			if !prevLock.safeRLock() {
				left = u.buckets.Load().Get(hash)
				continue
			}
		}
		if rightPtr := left.Next(); rightPtr == nil {
			if left.tryLazyLink(nil, unsafe.Pointer(&node[K]{key, hash, vPtr, nil, nil})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if right := (*node[K])(rightPtr); hash < right.hash {
			if left.tryLazyLink(rightPtr, unsafe.Pointer(&node[K]{key, hash, vPtr, rightPtr, nil})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if hash == right.hash && right.cmpKey(key) {
			prevLock.RUnlock()
			return (*V)(right.get()), true
		} else {
			left = right
		}
	}
}

func (u *BucketMap[K, V]) LoadOrStore(key K, val V) (v V, loaded bool) {
	a, b := u.LoadPtrOrStore(key, val)
	if b {
		v = *a
	}
	return v, b
}

func (u *BucketMap[K, V]) LoadPtr(key K) *V {
	hash := u.rehash(key)
	_, r, _ := u.buckets.Load().Get(hash).searchKey(key, hash)
	return (*V)(r.get())
}

func (u *BucketMap[K, V]) Load(key K) (V, bool) {
	hash := u.rehash(key)
	_, r, f := u.buckets.Load().Get(hash).searchKey(key, hash)
	var v V
	if f {
		v = *(*V)(r.get())
	}
	return v, f
}

func (u *BucketMap[K, V]) HasKey(key K) bool {
	hash := u.rehash(key)
	_, _, f := u.buckets.Load().Get(hash).searchKey(key, hash)
	return f
}

func (u *BucketMap[K, V]) LoadPtrAndDelete(key K) (v *V, loaded bool) {
	hash := u.rehash(key)
	var prevLock *relayLock = nil
	for left := u.buckets.Load().Get(hash); ; {
		if left.isRelay() {
			if prevLock != nil {
				prevLock.Unlock()
			}
			prevLock = left.relayLock
			if !prevLock.safeLock() {
				left = u.buckets.Load().Get(hash)
				continue
			}
		}
		if rightPtr := left.nx; rightPtr == nil {
			prevLock.Unlock()
			return
		} else if right := (*node[K])(rightPtr); hash < right.hash {
			prevLock.Unlock()
			return
		} else if hash == right.hash && right.cmpKey(key) {
			left.dangerUnlink(right)
			prevLock.Unlock()
			u.size.Add(^uint64(1 - 1))
			u.tryMerge()
			return (*V)(right.get()), true
		} else {
			left = right
		}
	}
}

func (u *BucketMap[K, V]) LoadAndDelete(key K) (v V, loaded bool) {
	a, b := u.LoadPtrAndDelete(key)
	if b {
		v = *a
	}
	return v, b
}

func (u *BucketMap[K, V]) Delete(key K) {
	u.LoadPtrAndDelete(key)
}

func (u *BucketMap[K, V]) RangePtr(f func(K, *V) bool) {
	for cur := u.buckets.Load().Get(0); cur != nil; cur = (*node[K])(cur.Next()) {
		if !cur.isRelay() {
			if !f(cur.k, (*V)(cur.get())) {
				break
			}
		}
	}
}

func (u *BucketMap[K, V]) Range(f func(K, V) bool) {
	u.RangePtr(func(k K, v *V) bool {
		return f(k, *v)
	})
}

func (u *BucketMap[K, V]) Take() (key K, val V) {
	a, b := u.TakePtr()
	return a, *b
}

func (u *BucketMap[K, V]) TakePtr() (key K, val *V) {
	if firstPtr := u.buckets.Load().Get(0).Next(); firstPtr != nil {
		first := (*node[K])(firstPtr)
		key, val = first.k, (*V)(first.get())
	}
	return
}
