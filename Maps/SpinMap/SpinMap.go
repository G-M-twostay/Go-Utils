package SpinMap

import (
	"GMUtils/Maps"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

type SpinMap[K Maps.Hashable, V any] struct {
	minAvgLen, maxAvgLen byte
	resizing             atomic.Bool
	maxHash              uint
	size                 atomic.Uint64
	buckets              atomic.Pointer[[]*node[K]]
}

func MakeSpinMap[K Maps.Hashable, V any](minBucketLen, maxBucketLen byte, maxHash uint) *SpinMap[K, V] {
	M := new(SpinMap[K, V])

	M.minAvgLen, M.maxAvgLen = minBucketLen, maxBucketLen
	M.maxHash = (Maps.MaxUintHash >> bits.LeadingZeros(maxHash)) & Maps.MaxArrayLen

	M.buckets.Store(&[]*node[K]{new(node[K])})

	return M
}

func (u *SpinMap[K, V]) rehash(k K) uint {
	return k.Hash() & Maps.MaxArrayLen
}

func (u *SpinMap[K, V]) Size() uint {
	return uint(u.size.Load())
}

func (u *SpinMap[K, V]) trySplit() {
	if u.resizing.CompareAndSwap(false, true) {
		s := *u.buckets.Load()
		if ul := uint(len(s)); u.Size()/ul > uint(u.maxAvgLen) {

			newBuckets := make([]*node[K], ul<<1)

			for i, v := range s {

				newBuckets[i<<1] = v
				newRelay := new(node[K])
				newRelay.hash = (u.maxHash/ul+1)*uint(i) + u.maxHash/(ul<<1)
				newBuckets[(i<<1)+1] = newRelay

				for {
					l, _ := v.searchHashAndAcquire(newRelay.hash)
					if l.addAndRelease(newRelay) {
						break
					}
				}
			}

			u.buckets.Store(&newBuckets)

		}
		u.resizing.Store(false)
	}
}

func (u *SpinMap[K, V]) tryMerge() {
	if u.resizing.CompareAndSwap(false, true) {
		s := *u.buckets.Load()
		if ul := uint(len(s)); u.Size()/ul < uint(u.minAvgLen) && ul > 1 {

			newBuckets := make([]*node[K], ul>>1)

			for i, v := range s {
				if i&1 == 0 {
					newBuckets[i>>1] = v
				} else {
					defer v.safeDelete()
				}
			}

			u.buckets.Store(&newBuckets)

		}
		u.resizing.Store(false)
	}
}

func (u *SpinMap[K, V]) findHash(hash uint) *node[K] {
	s := *u.buckets.Load()
	return s[hash/((u.maxHash+1)/uint(len(s)))]
}

func (u *SpinMap[K, V]) Store(key K, val V) {
	for hash, vPtr := u.rehash(key), unsafe.Pointer(&val); ; {
		if l, r, f := u.findHash(hash).searchKeyAndAcquire(key, hash); f {
			l.release()
			r.set(vPtr)
			return
		} else if l.addAndRelease(makeNode[K](key, hash, vPtr)) {
			u.size.Add(1)
			u.trySplit()
			return
		}
	}
}

func (u *SpinMap[K, V]) Load(key K) (val V, ok bool) {
	hash := u.rehash(key)
	if _, r, f := u.findHash(hash).searchKey(key, hash); f {
		val, ok = *(*V)(r.get()), true
	}
	return
}

func (u *SpinMap[K, V]) LoadAndDelete(key K) (val V, loaded bool) {
	hash := u.rehash(key)
	if _, r, f := u.findHash(hash).searchKey(key, hash); f {
		loaded = r.safeDelete()
		if loaded {
			u.size.Add(^uint64(1 - 1))
			u.tryMerge()
			val = *(*V)(r.get())
		}
	}

	return
}

func (u *SpinMap[K, V]) Delete(key K) {
	u.LoadAndDelete(key)
}

func (u *SpinMap[K, V]) HasKey(key K) bool {
	hash := u.rehash(key)
	_, _, f := u.findHash(hash).searchKey(key, hash)
	return f
}
