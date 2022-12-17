package BucketMap

import (
	"GMUtils/Maps"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

type BucketMap[K Maps.Hashable, V any] struct {
	buckets              atomic.Pointer[[]*node[K]]
	size                 atomic.Uint64
	maxHash              uint
	resizing             atomic.Uint32
	minAvgLen, maxAvgLen byte
}

func MakeBucketMap[K Maps.Hashable, V any](minBucketLen, maxBucketLen byte, maxHash uint) *BucketMap[K, V] {
	M := new(BucketMap[K, V])

	M.minAvgLen, M.maxAvgLen = minBucketLen, maxBucketLen
	M.maxHash = (Maps.MaxUintHash >> bits.LeadingZeros(maxHash)) & Maps.MaxArrayLen

	M.buckets.Store(&[]*node[K]{makeRelay[K](0)})

	return M
}

func (u *BucketMap[K, V]) rehash(k K) uint {
	return k.Hash() & Maps.MaxArrayLen
}

func (u *BucketMap[K, V]) Size() uint {
	return uint(u.size.Load())
}

func (u *BucketMap[K, V]) trySplit() {
	if u.resizing.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if ul := uint(len(s)); u.Size()/ul > uint(u.maxAvgLen) {

			newBuckets := make([]*node[K], ul<<1)

			for i, v := range s {

				newBuckets[i<<1] = v

				newRelay := makeRelay[K]((u.maxHash/ul+1)*uint(i) + u.maxHash/(ul<<1))
				newBuckets[(i<<1)+1] = newRelay

				v.RLock()
				for left := v; ; {
					if rightPtr := left.Next(); rightPtr == nil {
						if left.tryLink(nil, newRelay) {
							break
						}
					} else if right := (*node[K])(rightPtr); newRelay.hash <= right.hash {
						if left.tryLink(rightPtr, newRelay) {
							break
						}
					} else {
						left = right
					}
				}
				v.RUnlock()
			}

			u.buckets.Store(&newBuckets)

		}
		u.resizing.Store(0)
	}
}

func (u *BucketMap[K, V]) tryMerge() {
	if u.resizing.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if ul := uint(len(s)); u.Size()/ul < uint(u.minAvgLen) && ul > 1 {

			newBuckets := make([]*node[K], ul>>1)

			for i := range newBuckets {
				newBuckets[i] = s[i<<1]
			}

			u.buckets.Store(&newBuckets)

			for i := uint(0); i < ul; i += 2 {
				s[i].RLock()
				for left := s[i]; ; {
					rightPtr := left.Next()
					if right := (*node[K])(rightPtr); right == s[i+1] {
						if left.unlinkRelay(right, rightPtr) {
							break
						}
					} else {
						left = right
					}
				}
				s[i].RUnlock()
			}
		}
		u.resizing.Store(0)
	}
}

func (u *BucketMap[K, V]) findHash(hash uint) *node[K] {
	s := *u.buckets.Load()
	return s[hash/((u.maxHash+1)/uint(len(s)))]
}

func (u *BucketMap[K, V]) Store(key K, val V) {
	hash, vPtr := u.rehash(key), unsafe.Pointer(&val)
	var prevLock *relayLock = nil
	for left := u.findHash(hash); ; {
		if left.isRelay() {
			if prevLock != nil {
				prevLock.RUnlock()
			}
			prevLock = left.relayLock
			if !prevLock.safeRLock() {
				left = u.findHash(hash)
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
		} else if right := (*node[K])(rightPtr); hash == right.hash {
			if key.Equal(right.k) && !right.isRelay() {
				prevLock.RUnlock()
				right.set(vPtr)
				return
			} else {
				left = right
			}
		} else if hash > right.hash {
			left = right
		} else {
			if left.tryLazyLink(rightPtr, unsafe.Pointer(&node[K]{key, hash, vPtr, rightPtr, nil})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		}
	}
}

func (u *BucketMap[K, V]) LoadPtrOrStore(key K, val V) (v *V, loaded bool) {
	hash, vPtr := u.rehash(key), unsafe.Pointer(&val)
	var prevLock *relayLock = nil
	for left := u.findHash(hash); ; {
		if left.isRelay() {
			if prevLock != nil {
				prevLock.RUnlock()
			}
			prevLock = left.relayLock
			if !prevLock.safeRLock() {
				left = u.findHash(hash)
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
		} else if right := (*node[K])(rightPtr); hash == right.hash {
			if key.Equal(right.k) && !right.isRelay() {
				prevLock.RUnlock()
				return (*V)(right.get()), true
			} else {
				left = right
			}
		} else if hash > right.hash {
			left = right
		} else {
			if left.tryLazyLink(rightPtr, unsafe.Pointer(&node[K]{key, hash, vPtr, rightPtr, nil})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
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
	_, r, _, _ := u.findHash(hash).searchKey(key, hash)
	return (*V)(r.get())
}

func (u *BucketMap[K, V]) Load(key K) (V, bool) {
	hash := u.rehash(key)
	_, r, _, f := u.findHash(hash).searchKey(key, hash)
	var v V
	if f {
		v = *(*V)(r.get())
	}
	return v, f
}

func (u *BucketMap[K, V]) HasKey(key K) bool {
	hash := u.rehash(key)
	_, _, _, f := u.findHash(hash).searchKey(key, hash)
	return f
}

func (u *BucketMap[K, V]) LoadPtrAndDelete(key K) (v *V, loaded bool) {
	hash := u.rehash(key)
	var prevLock *relayLock = nil
	for left := u.findHash(hash); ; {
		if left.isRelay() {
			if prevLock != nil {
				prevLock.Unlock()
			}
			prevLock = left.relayLock
			if !prevLock.safeLock() {
				left = u.findHash(hash)
				continue
			}
		}
		if rightPtr := left.nx; rightPtr == nil {
			prevLock.Unlock()
			return
		} else if right := (*node[K])(rightPtr); hash == right.hash {
			if key.Equal(right.k) && !right.isRelay() {
				left.dangerUnlink(right)
				prevLock.Unlock()
				u.size.Add(^uint64(1 - 1))
				u.tryMerge()
				return (*V)(right.get()), true
			} else {
				left = right
			}
		} else if hash > right.hash {
			left = right
		} else {
			prevLock.Unlock()
			return
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
	u.LoadAndDelete(key)
}

func (u *BucketMap[K, V]) RangePtr(f func(K, *V) bool) {
	for cur := u.findHash(0); cur != nil; cur = (*node[K])(cur.Next()) {
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
	if firstPtr := u.findHash(0).Next(); firstPtr != nil {
		first := (*node[K])(firstPtr)
		key, val = first.k, (*V)(first.get())
	}
	return
}
