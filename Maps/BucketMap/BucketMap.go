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

				prevLock := noOpLock
				for left := v; ; {
					if left.isRelay() {
						prevLock.Unlock()
						prevLock = left.RLocker()
						prevLock.Lock()
					}
					if rightPtr := left.Next(); rightPtr == nil {
						left.tryLink(rightPtr, unsafe.Pointer(newRelay))
					}
				}
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

			newBuckets := make([]bucket[K], ul>>1)

			for i := range newBuckets {
				newBuckets[i] = s[i<<1]
			}

			u.buckets.Store(&newBuckets)

			for i := uint(0); i < ul; i += 2 {
				s[i+1].Lock()
				s[i].rmvNode(s[i+1].node)
				s[i+1].Unlock()
			}

		}
		u.resizing.Store(false)
	}
}

func (u *BucketMap[K, V]) findHash(hash uint) bucket[K] {
	s := *u.buckets.Load()
	return s[hash/((u.maxHash+1)/uint(len(s)))]
}

func (u *BucketMap[K, V]) Store(key K, val V) {
	hash := u.rehash(key)
	b := u.findHash(hash)
	if b.setOrAdd(key, hash, unsafe.Pointer(&val)) {
		u.size.Add(1)
		u.trySplit()
	}
}

func (u *BucketMap[K, V]) Load(key K) (val V, ok bool) {
	hash := u.rehash(key)
	vp := u.findHash(hash).get(key, hash)
	if ok = vp != nil; ok {
		val = *(*V)(vp)
	}
	return
}

func (u *BucketMap[K, V]) HasKey(key K) bool {
	hash := u.rehash(key)
	return u.findHash(hash).get(key, hash) != nil
}

func (u *BucketMap[K, V]) LoadAndDelete(key K) (val V, loaded bool) {
	hash := u.rehash(key)
	vp := u.findHash(hash).rmv(key, hash)
	if loaded = vp != nil; loaded {
		val = *(*V)(vp)
		u.size.Add(^uint64(1 - 1))
		//u.tryMerge()
	}
	return
}

func (u *BucketMap[K, V]) Delete(key K) {
	u.LoadAndDelete(key)
}
