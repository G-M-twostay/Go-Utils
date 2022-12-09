package BucketMap

import (
	"GMUtils/Maps"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

type BucketMap[K Maps.Hashable, V any] struct {
	minAvgLen, maxAvgLen byte
	resizing             atomic.Bool
	maxHash              uint
	size                 atomic.Uint64
	buckets              atomic.Pointer[[]*node[K]]
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
	if u.resizing.CompareAndSwap(false, true) {
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
		u.resizing.Store(false)
	}
}

func (u *BucketMap[K, V]) tryMerge() {
	if u.resizing.CompareAndSwap(false, true) {
		s := *u.buckets.Load()
		if ul := uint(len(s)); u.Size()/ul < uint(u.minAvgLen) && ul > 1 {

			newBuckets := make([]*node[K], ul>>1)

			for i := range newBuckets {
				newBuckets[i] = s[i<<1]
			}

			u.buckets.Store(&newBuckets)

			for i := uint(0); i < ul; i += 2 {
				s[i].Lock()
				for left := s[i]; ; {
					if rightPtr := left.Next(); rightPtr == nil {
						panic("unexpected")
					} else if right := (*node[K])(rightPtr); right == s[i+1] {
						left.dangerUnlinkNext(right, rightPtr)
						break
					} else {
						left = right
					}
				}
				s[i].Unlock()
			}
		}
		u.resizing.Store(false)
	}
}

func (u *BucketMap[K, V]) findHash(hash uint) *node[K] {
	s := *u.buckets.Load()
	return s[hash/((u.maxHash+1)/uint(len(s)))]
}

func (u *BucketMap[K, V]) Store(key K, val V) {
	hash, vPtr := u.rehash(key), unsafe.Pointer(&val)
	prevLock := noOpLock
	for left := u.findHash(hash); ; {
		if left.isRelay() {
			prevLock.Unlock()
			prevLock = left.RLocker()
			prevLock.Lock()
		}
		if rightPtr := left.Next(); rightPtr == nil {
			if left.tryLazyLink(nil, unsafe.Pointer(&node[K]{key, hash, vPtr, nil, nil})) {
				prevLock.Unlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if right := (*node[K])(rightPtr); hash == right.hash {
			if key.Equal(right.k) && !right.isRelay() {
				prevLock.Unlock()
				right.set(vPtr)
				return
			} else {
				left = right
			}
		} else if hash > right.hash {
			left = right
		} else {
			if left.tryLazyLink(rightPtr, unsafe.Pointer(&node[K]{key, hash, vPtr, rightPtr, nil})) {
				prevLock.Unlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		}
	}
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

func (u *BucketMap[K, V]) LoadAndDelete(key K) (V, bool) {
	hash := u.rehash(key)
	prevLock := noOpLock
	for left := u.findHash(hash); ; {
		if left.isRelay() {
			prevLock.Unlock()
			prevLock = left
			prevLock.Lock()
		}
		if rightPtr := left.Next(); rightPtr == nil {
			prevLock.Unlock()
			return *new(V), false
		} else if right := (*node[K])(rightPtr); hash == right.hash {
			if key.Equal(right.k) && !right.isRelay() {
				left.dangerUnlinkNext(right, rightPtr)
				prevLock.Unlock()
				u.size.Add(^uint64(1 - 1))
				u.tryMerge()
				return *(*V)(right.get()), true
			} else {
				left = right
			}
		} else if hash > right.hash {
			left = right
		} else {
			prevLock.Unlock()
			return *new(V), false
		}
	}
}

func (u *BucketMap[K, V]) Delete(key K) {
	u.LoadAndDelete(key)
}
