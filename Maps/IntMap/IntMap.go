package IntMap

import (
	"github.com/g-m-twostay/go-utils/Maps/internal"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

// IntMap is a specialized version of BucketMap for integers. It avoids all the interface operations.
type IntMap[K comparable, V any] struct {
	hash                           func(K) uint
	buckets                        atomic.Pointer[internal.HashList[*relay]]
	size                           atomic.Uintptr
	state                          atomic.Uint32
	minAvgLen, maxAvgLen, maxChunk byte
}

func New[K comparable, V any](minBucketLen, maxBucketLen byte, maxHash uint, hasher func(K) uint) *IntMap[K, V] {
	M := new(IntMap[K, V])

	M.minAvgLen, M.maxAvgLen = minBucketLen, maxBucketLen
	M.maxChunk = byte(bits.Len(internal.Mask(maxHash)))
	M.hash = hasher

	t := []*relay{{node: node{info: internal.Mark(0)}}}
	M.buckets.Store(&internal.HashList[*relay]{First: unsafe.SliceData(t), Chunk: M.maxChunk})

	return M
}

func (u *IntMap[K, V]) Size() uint {
	return uint(u.size.Load())
}

func (u *IntMap[K, V]) trySplit() {
	if u.state.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if lc := u.maxChunk - s.Chunk; u.Size()>>lc > uint(u.maxAvgLen) {

			newBuckets := make([]*relay, 1<<(lc+1))

			for i := uint(0); i < 1<<lc; i++ {
				v := s.Fetch(i)
				newBuckets[i<<1] = v

				hash := (1<<s.Chunk)*i + (1 << (s.Chunk - 1))
				newRelay := &relay{node: node{info: internal.Mark(hash)}}
				newBuckets[(i<<1)+1] = newRelay

				v.RLock()
				for left, newRelayPtr := &v.node, unsafe.Pointer(newRelay); ; {
					rightPtr := left.Next()
					if rightB := (*node)(rightPtr); rightB == nil || hash <= rightB.Hash() {
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

			u.buckets.Store(&internal.HashList[*relay]{First: unsafe.SliceData(newBuckets), Chunk: s.Chunk - 1})

		}
		u.state.Store(0)
	}
}

func (u *IntMap[K, V]) tryMerge() {
	if u.state.CompareAndSwap(0, 1) {
		s := *u.buckets.Load()
		if lc := u.maxChunk - s.Chunk; lc > 0 && u.Size()>>lc < uint(u.minAvgLen) {

			newBuckets := make([]*relay, 1<<(lc-1))

			for i := range newBuckets {
				newBuckets[i] = s.Fetch(uint(i) << 1)
			}

			u.buckets.Store(&internal.HashList[*relay]{First: unsafe.SliceData(newBuckets), Chunk: s.Chunk + 1})

			for _, v := range newBuckets {
				v.RLock()
				for left := &v.node; ; {
					rightPtr := left.Next()
					if rightB := (*node)(rightPtr); rightB.isRelay() {
						if left.unlinkRelay((*relay)(rightPtr), rightPtr) {
							break
						}
					} else {
						left = rightB
					}
				}
				v.RUnlock()
			}
		}
		u.state.Store(0)
	}
}

func (u *IntMap[K, V]) Store(key K, val V) {
	hash, vPtr := internal.Mask(u.hash(key)), unsafe.Pointer(&val)

	prevLock := u.buckets.Load().Get(hash)

	if !prevLock.safeRLock() {
		prevLock.RUnlock()
		prevLock = u.buckets.Load().Get(hash)
		prevLock.RLock()
	}

	for left := &prevLock.node; ; {
		rightPtr := left.Next()
		if rightB := (*node)(rightPtr); rightB == nil || hash < rightB.Hash() {
			if left.dangerLink(rightPtr, unsafe.Pointer(&value[K]{node: node{info: hash, nx: rightPtr}, v: vPtr, k: key})) {
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
				if prevLock = (*relay)(rightPtr); !prevLock.safeRLock() {
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
	hash, vPtr := internal.Mask(u.hash(key)), unsafe.Pointer(&val)

	prevLock := u.buckets.Load().Get(hash)
	if !prevLock.safeRLock() {
		prevLock.RUnlock()
		prevLock = u.buckets.Load().Get(hash)
		prevLock.RLock()
	}

	for left := &prevLock.node; ; {
		rightPtr := left.Next()
		if rightB := (*node)(rightPtr); rightB == nil || hash < rightB.Hash() {
			if left.dangerLink(rightPtr, unsafe.Pointer(&value[K]{node: node{info: hash, nx: rightPtr}, v: vPtr, k: key})) {
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
				if prevLock = (*relay)(rightPtr); !prevLock.safeRLock() {
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
	hash := internal.Mask(u.hash(key))
	if r := search(u.buckets.Load().Get(hash), key, hash); r == nil {
		return nil
	} else {
		return (*V)(r.get())
	}
}

func (u *IntMap[K, V]) Load(key K) (V, bool) {
	hash := internal.Mask(u.hash(key))
	if r := search(u.buckets.Load().Get(hash), key, hash); r == nil {
		return *new(V), false
	} else {
		return *(*V)(r.get()), true
	}
}

func (u *IntMap[K, V]) HasKey(key K) bool {
	hash := internal.Mask(u.hash(key))
	return search(u.buckets.Load().Get(hash), key, hash) != nil
}

func (u *IntMap[K, V]) LoadPtrAndDelete(key K) (v *V, loaded bool) {
	hash := internal.Mask(u.hash(key))
	prevLock := u.buckets.Load().Get(hash)

	if !prevLock.safeLock() {
		prevLock.Unlock()
		prevLock = u.buckets.Load().Get(hash)
		prevLock.Lock()
	}

	for left := &prevLock.node; ; {
		rightPtr := left.nx
		if rightB := (*node)(rightPtr); rightB == nil || hash < rightB.Hash() {
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
				if prevLock = (*relay)(rightPtr); !prevLock.safeLock() {
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
	for cur := (*node)(u.buckets.Load().Get(0).Next()); cur != nil; cur = (*node)(cur.Next()) {
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
	hash := internal.Mask(u.hash(key))
	if r := search(u.buckets.Load().Get(hash), key, hash); r != nil {
		v = (*V)(r.swap(unsafe.Pointer(&val)))
	}
	return
}
