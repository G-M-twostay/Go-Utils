package ChainMap

import (
	"GMUtils/Maps"
	"math"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

const ( //both inclusive
	maxArrayLen uint = math.MaxInt //so maxArrayLen+1 won't overflow
)

type ChainMap[K Maps.Hashable, V any] struct {
	minAvgLen, maxAvgLen byte
	resizing             atomic.Bool
	maxHash              uint
	size                 atomic.Uint64
	buckets              atomic.Pointer[[]*node[K]]
}

func MakeChainMap[K Maps.Hashable, V any](minBucketLen, maxBucketLen byte, maxHash uint) *ChainMap[K, V] {
	M := new(ChainMap[K, V])

	M.minAvgLen, M.maxAvgLen = minBucketLen, maxBucketLen
	M.maxHash = (Maps.MaxUintHash >> bits.LeadingZeros(maxHash)) & maxArrayLen

	M.buckets.Store(&[]*node[K]{makeRelay[K](0, nil)})

	return M
}

func (u *ChainMap[K, V]) rehash(k K) uint {
	return k.Hash() & maxArrayLen
}

// chunk=n
//
//	[0,2^n -1] to [0,2^n-1 -1],[2^n-1,2^n]
//
// or [0,2^n) to [0,2^n-1),[2^n-1,2^n)
func (u *ChainMap[K, V]) trySplit() {
	if u.resizing.CompareAndSwap(false, true) {
		s := *u.buckets.Load()
		if ul := uint(len(s)); u.Size()/ul > uint(u.maxAvgLen) {

			newBuckets := make([]*node[K], ul<<1)

			for i, v := range s {

				newBuckets[i<<1] = v
				newRelay := makeRelay[K]((u.maxHash/ul+1)*uint(i)+u.maxHash/(ul<<1), nil)
				newBuckets[(i<<1)+1] = newRelay

				for tempState := (*state[K])(newRelay.s); ; {
					l, ls, lsp, _ := v.searchHash(newRelay.hash)
					tempState.nx = ls.nx
					//didn't use l.addAfter(ls, lsp, newRelay) because this would repeatedly allocate new states is unnecessary as we only need to change nx since we know we will always add this new relay
					if atomic.CompareAndSwapPointer(&l.s, lsp, unsafe.Pointer(ls.changeNext(newRelay))) {
						break
					}
				}
			}

			u.buckets.Store(&newBuckets)

		}
		u.resizing.Store(false)
	}
}

func (u *ChainMap[K, V]) tryMerge() {
	if u.resizing.CompareAndSwap(false, true) {
		s := *u.buckets.Load()
		if ul := uint(len(s)); u.Size()/ul < uint(u.minAvgLen) && ul > 1 {

			newBuckets := make([]*node[K], ul>>1)

			for i, v := range s {
				if i&1 == 0 {
					newBuckets[i>>1] = v
				} else {
					defer v.delete()
				}
			}

			u.buckets.Store(&newBuckets)

		}
		u.resizing.Store(false)
	}
}

func (u *ChainMap[K, V]) findHash(hash uint) *node[K] {
	s := *u.buckets.Load()
	return s[hash/((u.maxHash+1)/uint(len(s)))]
}

func (u *ChainMap[K, V]) Store(key K, val V) {
	for hash, vPtr := u.rehash(key), unsafe.Pointer(&val); ; {
		if l, ls, lsp, r, f := u.findHash(hash).searchKey(key, hash); f {
			r.setVPtr(vPtr)
			return
		} else if l.addAfter(ls, lsp, &node[K]{key, hash, vPtr, nil}) {
			u.size.Add(1)
			u.trySplit()
			return
		}
	}
}

func (u *ChainMap[K, V]) Delete(key K) {
	u.LoadAndDelete(key)
}

func (u *ChainMap[K, V]) Load(key K) (val V, ok bool) {
	hash := u.rehash(key)
	if _, _, _, r, f := u.findHash(hash).searchKey(key, hash); f {
		val, ok = *(*V)(r.getVPtr()), true
	}
	return
}

func (u *ChainMap[K, V]) HasKey(key K) bool {
	hash := u.rehash(key)
	_, _, _, _, f := u.findHash(hash).searchKey(key, hash)
	return f
}

func (u *ChainMap[K, V]) LoadOrStore(key K, val V) (V, bool) {
	for hash, vPtr := u.rehash(key), unsafe.Pointer(&val); ; {
		if l, ls, lsp, r, f := u.findHash(hash).searchKey(key, hash); f {
			return *(*V)(r.getVPtr()), true
		} else if l.addAfter(ls, lsp, &node[K]{key, hash, vPtr, nil}) {
			return *new(V), false
		}
	}
}

func (u *ChainMap[K, V]) LoadAndDelete(key K) (val V, loaded bool) {
	hash := u.rehash(key)
	if _, _, _, r, f := u.findHash(hash).searchKey(key, hash); f {
		loaded = !r.delete()
		if loaded {
			u.size.Add(^uint64(1 - 1))
			u.tryMerge()
			val = *(*V)(r.getVPtr())
		}
	}
	return
}

func (u *ChainMap[K, V]) Size() uint {
	return uint(u.size.Load())
}

func (u *ChainMap[K, V]) Take() (K, V) {
	t, _, _ := u.findHash(0).next()
	return t.k, *(*V)(t.getVPtr())
}

func (u *ChainMap[K, V]) Range(f func(K, V) bool) {
	for cur := u.findHash(0); ; cur, _, _ = cur.next() {
		if !cur.isRelay() {
			if !f(cur.k, *(*V)(cur.getVPtr())) {
				break
			}
		}
	}
}

func (u *ChainMap[K, V]) LoadPtr(key K) (vp *V) {
	hash := u.rehash(key)
	if _, _, _, r, f := u.findHash(hash).searchKey(key, hash); f {
		vp = (*V)(r.getVPtr())
	}
	return
}
