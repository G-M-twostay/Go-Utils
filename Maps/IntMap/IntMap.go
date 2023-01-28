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
	buckets                        atomic.Pointer[Maps.HashList[*relay[K]]]
	size                           atomic.Uintptr
	state                          atomic.Uint32
	minAvgLen, maxAvgLen, maxChunk byte
}

func New[K comparable, V any](minBucketLen, maxBucketLen byte, maxHash uint, hasher func(K) uint) *IntMap[K, V] {
	M := new(IntMap[K, V])

	M.minAvgLen, M.maxAvgLen = minBucketLen, maxBucketLen
	M.maxChunk = byte(bits.Len(Maps.Mask(maxHash)))
	M.rehash = hasher

	t := []*relay[K]{{node: node[K]{info: Maps.Mark(0)}}}
	M.buckets.Store(&Maps.HashList[*relay[K]]{Array: t, Chunk: M.maxChunk})

	return M
}

func (u *IntMap[K, V]) Size() uint {
	return uint(u.size.Load())
}

func (u *IntMap[K, V]) trySplit() {
	if u.state.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if u.Size()>>(u.maxChunk-s.Chunk) > uint(u.maxAvgLen) {

			newBuckets := make([]*relay[K], len(s.Array)<<1)

			for i, v := range s.Array {

				newBuckets[i<<1] = v

				hash := (1<<s.Chunk)*uint(i) + (1 << (s.Chunk - 1))
				newRelay := &relay[K]{node: node[K]{info: Maps.Mark(hash)}}
				newBuckets[(i<<1)+1] = newRelay

				v.RLock()
				for left, newRelayPtr := &v.node, unsafe.Pointer(newRelay); ; {
					rightPtr := left.Next()
					if rightB := (*node[K])(rightPtr); rightB == nil || hash <= rightB.Hash() {
						newRelay.nx = rightPtr
						if left.dangerLink(rightPtr, newRelayPtr) {
							break
						}
					} else {
						left = rightB
					}
				}
				v.RUnlock()
			}

			u.buckets.Store(&Maps.HashList[*relay[K]]{Array: newBuckets, Chunk: s.Chunk - 1})

		}
		u.state.Store(0)
	}
}

func (u *IntMap[K, V]) tryMerge() {
	if u.state.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if u.Size()>>(u.maxChunk-s.Chunk) < uint(u.minAvgLen) && len(s.Array) > 1 {

			newBuckets := make([]*relay[K], len(s.Array)>>1)

			for i := range newBuckets {
				newBuckets[i] = s.Array[i<<1]
			}

			u.buckets.Store(&Maps.HashList[*relay[K]]{Array: newBuckets, Chunk: s.Chunk + 1})

			for i := 0; i < len(s.Array); i += 2 {
				s.Array[i].RLock()
				for left := &s.Array[i].node; ; {
					rightPtr := left.Next()
					if rightB := (*node[K])(rightPtr); rightB.isRelay() {
						if left.unlinkRelay((*relay[K])(rightPtr), rightPtr) {
							break
						}
					} else {
						left = rightB
					}
				}
				s.Array[i].RUnlock()
			}
		}
		u.state.Store(0)
	}
}

func (u *IntMap[K, V]) Store(key K, val V) {
	hash, vPtr := Maps.Mask(u.rehash(key)), unsafe.Pointer(&val)

	prevLock := u.buckets.Load().Get(hash)

	if !prevLock.safeRLock() {
		prevLock.RUnlock()
		prevLock = u.buckets.Load().Get(hash)
		prevLock.RLock()
	}

	for left := &prevLock.node; ; {
		rightPtr := left.Next()
		if rightB := (*node[K])(rightPtr); rightB == nil || hash < rightB.Hash() {
			if left.dangerLink(rightPtr, unsafe.Pointer(&value[K]{node: node[K]{info: hash, nx: rightPtr}, v: vPtr, k: key})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if right := (*value[K])(rightPtr); hash == rightB.info && key == right.k {
			prevLock.RUnlock()
			right.set(vPtr)
			return
		} else {
			if left = rightB; rightB.isRelay() {
				prevLock.RUnlock()
				if prevLock = (*relay[K])(rightPtr); !prevLock.safeRLock() {
					prevLock.RUnlock()
					prevLock = u.buckets.Load().Get(hash)
					prevLock.RLock()
					left = &prevLock.node
				}
			}
		}
	}
}

func (u *IntMap[K, V]) LoadPtrOrStore(key K, val V) (v *V, loaded bool) {
	hash, vPtr := Maps.Mask(u.rehash(key)), unsafe.Pointer(&val)

	prevLock := u.buckets.Load().Get(hash)
	if !prevLock.safeRLock() {
		prevLock.RUnlock()
		prevLock = u.buckets.Load().Get(hash)
		prevLock.RLock()
	}

	for left := &prevLock.node; ; {
		rightPtr := left.Next()
		if rightB := (*node[K])(rightPtr); rightB == nil || hash < rightB.Hash() {
			if left.dangerLink(rightPtr, unsafe.Pointer(&value[K]{node: node[K]{info: hash, nx: rightPtr}, v: vPtr, k: key})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if right := (*value[K])(rightPtr); hash == rightB.info && key == right.k {
			prevLock.RUnlock()
			return (*V)(right.get()), true
		} else {
			if left = rightB; rightB.isRelay() {
				prevLock.RUnlock()
				if prevLock = (*relay[K])(rightPtr); !prevLock.safeRLock() {
					prevLock.RUnlock()
					prevLock = u.buckets.Load().Get(hash)
					prevLock.RLock()
					left = &prevLock.node
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
	prevLock := u.buckets.Load().Get(hash)

	if !prevLock.safeLock() {
		prevLock.Unlock()
		prevLock = u.buckets.Load().Get(hash)
		prevLock.Lock()
	}

	for left := &prevLock.node; ; {
		rightPtr := left.nx
		if rightB := (*node[K])(rightPtr); rightB == nil || hash < rightB.Hash() {
			prevLock.Unlock()
			return
		} else if right := (*value[K])(rightPtr); hash == rightB.info && key == right.k {
			left.dangerUnlink(rightB)
			prevLock.Unlock()
			u.size.Add(^uintptr(1 - 1))
			u.tryMerge()
			return
		} else {
			if left = rightB; rightB.isRelay() {
				prevLock.Unlock()
				if prevLock = (*relay[K])(rightPtr); !prevLock.safeLock() {
					prevLock.Unlock()
					prevLock = u.buckets.Load().Get(hash)
					prevLock.Lock()
					left = &prevLock.node
				}
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
	for cur := (*node[K])(u.buckets.Load().Get(0).Next()); cur != nil; cur = (*node[K])(cur.Next()) {
		if !cur.isRelay() {
			if t := (*value[K])(unsafe.Pointer(cur)); !f(t.k, (*V)(t.get())) {
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
		first := (*value[K])(firstPtr)
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
