package IntMap

import (
	"github.com/g-m-twostay/go-utils/Maps"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

// IntMap is a specialized version of BucketMap for integers. It avoids all the interface operations.
type IntMap[K comparable, V any] struct {
	rehash                         func(K) uint
	buckets                        atomic.Pointer[Maps.HashList[*node[K]]]
	size                           atomic.Uintptr
	state                          atomic.Uint32
	minAvgLen, maxAvgLen, maxChunk byte
}

func New[K comparable, V any](minBucketLen, maxBucketLen byte, maxHash uint, hasher func(K) uint) *IntMap[K, V] {
	M := new(IntMap[K, V])

	M.minAvgLen, M.maxAvgLen = minBucketLen, maxBucketLen
	M.maxChunk = byte(bits.Len(Maps.Mask(maxHash)))
	M.rehash = hasher

	t := []*node[K]{{info: Maps.Mark(0), v: unsafe.Pointer(new(Maps.FlagLock))}}
	M.buckets.Store(&Maps.HashList[*node[K]]{Array: t, Chunk: M.maxChunk})

	return M
}

func (u *IntMap[K, V]) Size() uint {
	return uint(u.size.Load())
}

func (u *IntMap[K, V]) trySplit() {
	if u.state.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if u.Size()>>(u.maxChunk-s.Chunk) > uint(u.maxAvgLen) {

			newBuckets := make([]*node[K], len(s.Array)<<1)

			for i, v := range s.Array {

				newBuckets[i<<1] = v

				hash := (1<<s.Chunk)*uint(i) + (1 << (s.Chunk - 1))
				newRelay := &node[K]{info: Maps.Mark(hash), v: unsafe.Pointer(new(Maps.FlagLock))}
				newBuckets[(i<<1)+1] = newRelay

				t := v.lock()
				t.RLock()
				for left, newRelayPtr := v, unsafe.Pointer(newRelay); ; {
					rightPtr := left.Next()
					if right := (*node[K])(rightPtr); right == nil || hash <= right.Hash() {
						newRelay.nx = rightPtr
						if left.dangerLink(rightPtr, newRelayPtr) {
							break
						}
					} else {
						left = right
					}
				}
				t.RUnlock()
			}

			u.buckets.Store(&Maps.HashList[*node[K]]{Array: newBuckets, Chunk: s.Chunk - 1})

		}
		u.state.Store(0)
	}
}

func (u *IntMap[K, V]) tryMerge() {
	if u.state.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if u.Size()>>(u.maxChunk-s.Chunk) < uint(u.minAvgLen) && len(s.Array) > 1 {

			newBuckets := make([]*node[K], len(s.Array)>>1)

			for i := range newBuckets {
				newBuckets[i] = s.Array[i<<1]
			}

			u.buckets.Store(&Maps.HashList[*node[K]]{Array: newBuckets, Chunk: s.Chunk + 1})

			for i := 0; i < len(s.Array); i += 2 {
				t := s.Array[i].lock()
				t.RLock()
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
				t.RUnlock()
			}
		}
		u.state.Store(0)
	}
}

func (u *IntMap[K, V]) Store(key K, val V) {
	hash, vPtr := Maps.Mask(u.rehash(key)), unsafe.Pointer(&val)

	left := u.buckets.Load().Get(hash)
	prevLock := left.lock()
	if !prevLock.SafeRLock() {
		prevLock.RUnlock()
		left = u.buckets.Load().Get(hash)
		prevLock = left.lock()
		prevLock.RLock()
	}

	for {
		rightPtr := left.Next()
		if right := (*node[K])(rightPtr); right == nil || hash < right.Hash() {
			if left.dangerLink(rightPtr, unsafe.Pointer(&node[K]{info: hash, v: vPtr, nx: rightPtr, k: key})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if hash == right.info && key == right.k {
			prevLock.RUnlock()
			right.set(vPtr)
			return
		} else {
			left = right
		}

		if left.isRelay() {
			if cl := left.lock(); cl != prevLock {
				prevLock.RUnlock()
				if prevLock = cl; !prevLock.SafeRLock() {
					prevLock.RUnlock()
					left = u.buckets.Load().Get(hash)
					prevLock = left.lock()
					prevLock.RLock()
				}
			}
		}
	}
}

func (u *IntMap[K, V]) LoadPtrOrStore(key K, val V) (v *V, loaded bool) {
	hash, vPtr := Maps.Mask(u.rehash(key)), unsafe.Pointer(&val)

	left := u.buckets.Load().Get(hash)
	prevLock := left.lock()
	if !prevLock.SafeRLock() {
		prevLock.RUnlock()
		left = u.buckets.Load().Get(hash)
		prevLock = left.lock()
		prevLock.RLock()
	}

	for {
		rightPtr := left.Next()
		if right := (*node[K])(rightPtr); right == nil || hash < right.Hash() {
			if left.dangerLink(rightPtr, unsafe.Pointer(&node[K]{info: hash, v: vPtr, nx: rightPtr, k: key})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if hash == right.info && key == right.k {
			prevLock.RUnlock()
			return (*V)(right.get()), true
		} else {
			left = right
		}

		if left.isRelay() {
			if cl := left.lock(); cl != prevLock {
				prevLock.RUnlock()
				if prevLock = cl; !prevLock.SafeRLock() {
					prevLock.RUnlock()
					left = u.buckets.Load().Get(hash)
					prevLock = left.lock()
					prevLock.RLock()
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
	hash := Maps.Mask(u.rehash(key))
	if r := u.buckets.Load().Get(hash).search(key, hash); r == nil {
		return nil
	} else {
		return (*V)(r.get())
	}
}

func (u *IntMap[K, V]) Load(key K) (V, bool) {
	hash := Maps.Mask(u.rehash(key))
	if r := u.buckets.Load().Get(hash).search(key, hash); r == nil {
		return *new(V), false
	} else {
		return *(*V)(r.get()), true
	}
}

func (u *IntMap[K, V]) HasKey(key K) bool {
	hash := Maps.Mask(u.rehash(key))
	return u.buckets.Load().Get(hash).search(key, hash) != nil
}

func (u *IntMap[K, V]) LoadPtrAndDelete(key K) (v *V, loaded bool) {
	hash := Maps.Mask(u.rehash(key))

	left := u.buckets.Load().Get(hash)
	prevLock := left.lock()
	if !prevLock.SafeLock() {
		prevLock.Unlock()
		left = u.buckets.Load().Get(hash)
		prevLock = left.lock()
		prevLock.Lock()
	}

	for {
		rightPtr := left.nx
		if right := (*node[K])(rightPtr); right == nil || hash < right.Hash() {
			prevLock.Unlock()
			return
		} else if hash == right.info && key == right.k {
			left.dangerUnlink(right)
			prevLock.Unlock()
			u.size.Add(^uintptr(1 - 1))
			u.tryMerge()
			return (*V)(right.get()), true
		} else {
			left = right
		}

		if left.isRelay() {
			prevLock.Unlock()
			if prevLock = left.lock(); !prevLock.SafeLock() {
				prevLock.Unlock()
				left = u.buckets.Load().Get(hash)
				prevLock = left.lock()
				prevLock.Lock()
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
	for cur := u.buckets.Load().Get(0); cur != nil; cur = (*node[K])(cur.Next()) {
		if !cur.isRelay() {
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
		first := (*node[K])(firstPtr)
		key, val = first.k, (*V)(first.get())
	}
	return
}

func (u *IntMap[K, V]) Set(key K, val V) (v *V) {
	hash := Maps.Mask(u.rehash(key))
	if r := u.buckets.Load().Get(hash).search(key, hash); r != nil {
		v = (*V)(r.swap(unsafe.Pointer(&val)))
	}
	return
}
