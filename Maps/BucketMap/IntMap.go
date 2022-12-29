package BucketMap

import (
	"GMUtils/Maps"
	"golang.org/x/exp/constraints"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

type IntMap[K constraints.Integer, V any] struct {
	rehash               func(K) uint
	buckets              atomic.Pointer[[]*intNode[K]]
	size                 atomic.Uint64
	maxHash              uint
	resizing             atomic.Uint32
	minAvgLen, maxAvgLen byte
}

func MakeIntMap[K constraints.Integer, V any](minBucketLen, maxBucketLen byte, maxHash uint, hasher func(K) uint) *IntMap[K, V] {
	M := new(IntMap[K, V])

	M.minAvgLen, M.maxAvgLen = minBucketLen, maxBucketLen
	M.maxHash = Maps.MaxUintHash >> bits.LeadingZeros(maxHash)
	M.rehash = hasher

	M.buckets.Store(&([]*intNode[K]{&intNode[K]{0, unsafe.Pointer(new(relayLock)), nil, 0, true}}))

	return M
}

func (u *IntMap[K, V]) Size() uint {
	return uint(u.size.Load())
}

func (u *IntMap[K, V]) trySplit() {
	if u.resizing.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if ul := uint(len(s)); u.Size()/ul > uint(u.maxAvgLen) {

			newBuckets := make([]*intNode[K], ul<<1)

			for i, v := range s {

				newBuckets[i<<1] = v

				newRelay := &intNode[K]{(u.maxHash/ul+1)*uint(i) + u.maxHash/(ul<<1), unsafe.Pointer(new(relayLock)), nil, 0, true}
				newBuckets[(i<<1)+1] = newRelay

				t := v.lock()
				t.RLock()
				for left, newRelayPtr := v, unsafe.Pointer(newRelay); ; {
					if rightPtr := left.Next(); rightPtr == nil {
						if left.tryLink(nil, newRelay, newRelayPtr) {
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

			u.buckets.Store(&newBuckets)

		}
		u.resizing.Store(0)
	}
}

func (u *IntMap[K, V]) tryMerge() {
	if u.resizing.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if ul := uint(len(s)); u.Size()/ul < uint(u.minAvgLen) && ul > 1 {

			newBuckets := make([]*intNode[K], ul>>1)

			for i := range newBuckets {
				newBuckets[i] = s[i<<1]
			}

			u.buckets.Store(&newBuckets)

			for i := uint(0); i < ul; i += 2 {
				t := s[i].lock()
				t.RLock()
				for left := s[i]; ; {
					rightPtr := left.Next()
					if right := (*intNode[K])(rightPtr); right == s[i+1] {
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
		u.resizing.Store(0)
	}
}

func (u *IntMap[K, V]) findHash(hash uint) *intNode[K] {
	s := *u.buckets.Load()
	return s[hash/((u.maxHash+1)/uint(len(s)))]
}

func (u *IntMap[K, V]) Store(key K, val V) {
	hash, vPtr := u.rehash(key), unsafe.Pointer(&val)
	var prevLock *relayLock = nil
	for left := u.findHash(hash); ; {
		if left.flag {
			if prevLock != nil {
				prevLock.RUnlock()
			}
			prevLock = left.lock()
			if !prevLock.safeRLock() {
				left = u.findHash(hash)
				continue
			}
		}
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
	}
}

func (u *IntMap[K, V]) LoadPtrOrStore(key K, val V) (v *V, loaded bool) {
	hash, vPtr := u.rehash(key), unsafe.Pointer(&val)
	var prevLock *relayLock = nil
	for left := u.findHash(hash); ; {
		if left.flag {
			if prevLock != nil {
				prevLock.RUnlock()
			}
			prevLock = left.lock()
			if !prevLock.safeRLock() {
				left = u.findHash(hash)
				continue
			}
		}
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
	_, r, _ := u.findHash(hash).searchKey(key, hash)
	return (*V)(r.get())
}

func (u *IntMap[K, V]) Load(key K) (V, bool) {
	hash := u.rehash(key)
	_, r, f := u.findHash(hash).searchKey(key, hash)
	var v V
	if f {
		v = *(*V)(r.get())
	}
	return v, f
}

func (u *IntMap[K, V]) HasKey(key K) bool {
	hash := u.rehash(key)
	_, _, f := u.findHash(hash).searchKey(key, hash)
	return f
}

func (u *IntMap[K, V]) LoadPtrAndDelete(key K) (v *V, loaded bool) {
	hash := u.rehash(key)
	var prevLock *relayLock = nil
	for left := u.findHash(hash); ; {
		if left.flag {
			if prevLock != nil {
				prevLock.Unlock()
			}
			prevLock = left.lock()
			if !prevLock.safeLock() {
				left = u.findHash(hash)
				continue
			}
		}
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
	for cur := u.findHash(0); cur != nil; cur = (*intNode[K])(cur.Next()) {
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
	if firstPtr := u.findHash(0).Next(); firstPtr != nil {
		first := (*intNode[K])(firstPtr)
		key, val = first.k, (*V)(first.get())
	}
	return
}
