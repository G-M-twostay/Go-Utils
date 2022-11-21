package ChainMap

import (
	"GMUtils/Maps"
	"math/bits"
	"sync"
	"sync/atomic"
	"unsafe"
)

const ( //both inclusive
	maxArrayLen  uint = ^uint(0) >> 1
	maxHashRange uint = ^uint(0)
)

type ChainMap[K Maps.Hashable, V any] struct {
	buckets                               []*node[K]
	chunk, maxChunk, minAvgLen, maxAvgLen byte //1<<chunk=len(buckets)
	size                                  atomic.Uint64
	bucketsLock                           sync.RWMutex
	resizing                              atomic.Bool
	maxHash                               uint
}

func MakeChainMap[K Maps.Hashable, V any](minBucketLen, maxBucketLen byte, minHash, maxHash int) *ChainMap[K, V] {
	M := new(ChainMap[K, V])

	M.minAvgLen, M.maxAvgLen = minBucketLen, maxBucketLen
	t := bits.LeadingZeros(uint(maxHash - minHash))
	M.maxHash = maxHashRange >> t
	M.maxChunk = byte(bits.UintSize - t)

	M.buckets = []*node[K]{makeRelay[K](0, nil)}

	return M
}

func (u *ChainMap[K, V]) rehash(k K) uint {
	return uint(k.Hash())
}

// chunk=n
//
//	[0,2^n -1] to [0,2^n-1 -1],[2^n-1,2^n]
//
// or [0,2^n) to [0,2^n-1),[2^n-1,2^n)
func (u *ChainMap[K, V]) trySplit() {
	if u.Size()>>uint(u.chunk) > uint(u.maxAvgLen) {
		if u.resizing.CompareAndSwap(false, true) {

			newBuckets := make([]*node[K], 1<<(u.chunk+1))

			for i, v := range u.buckets {

				newBuckets[i<<1] = v
				newRelay := makeRelay[K](((u.maxHash>>uint(u.chunk))+1)*uint(i)+(u.maxHash>>(u.chunk+1)), nil)
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

			u.bucketsLock.Lock()
			u.buckets = newBuckets
			u.chunk++
			u.bucketsLock.Unlock()

			u.resizing.Store(false)
		}
	}
}

func (u *ChainMap[K, V]) tryMerge() {
	if u.Size()>>uint(u.chunk) < uint(u.minAvgLen) && u.chunk > 0 {
		if u.resizing.CompareAndSwap(false, true) {

			newBuckets := make([]*node[K], 1<<(u.chunk-1))

			for i, v := range u.buckets {
				if i&1 == 0 {
					newBuckets[i>>1] = v
				} else {
					defer v.delete()
				}
			}

			u.bucketsLock.Lock()
			u.buckets = newBuckets
			u.chunk--
			u.bucketsLock.Unlock()

			u.resizing.Store(false)
		}
	}
}

func (u *ChainMap[K, V]) findHash(hash uint) *node[K] {
	u.bucketsLock.RLock()
	defer u.bucketsLock.RUnlock()
	return u.buckets[hash>>uint(u.maxChunk-u.chunk)]
}

func (u *ChainMap[K, V]) Put(key K, val V) {
	for hash, vPtr := u.rehash(key), unsafe.Pointer(&val); ; {
		if l, ls, lsp, r, f := u.findHash(hash).searchKey(key, hash); f {
			r.setVPtr(vPtr)
			return
		} else if l.addAfter(ls, lsp, &node[K]{key, vPtr, hash, nil}) {
			u.size.Add(1)
			u.trySplit()
			return
		}
	}
}

func (u *ChainMap[K, V]) Remove(key K) {
	u.GetAndRmv(key)
}

func (u *ChainMap[K, V]) Get(key K) (val V) {
	hash := u.rehash(key)
	if _, _, _, r, f := u.findHash(hash).searchKey(key, hash); f {
		val = *(*V)(r.getVPtr())
	}
	return
}

func (u *ChainMap[K, V]) HasKey(key K) bool {
	hash := u.rehash(key)
	_, _, _, _, f := u.findHash(hash).searchKey(key, hash)
	return f
}

func (u *ChainMap[K, V]) GetOrPut(key K, val V) (V, bool) {
	for hash, vPtr := u.rehash(key), unsafe.Pointer(&val); ; {
		if l, ls, lsp, r, f := u.findHash(hash).searchKey(key, hash); f {
			return *(*V)(r.getVPtr()), true
		} else if l.addAfter(ls, lsp, &node[K]{key, vPtr, hash, nil}) {
			return *new(V), false
		}
	}
}

func (u *ChainMap[K, V]) GetAndRmv(key K) (val V, removed bool) {
	hash := u.rehash(key)
	if _, _, _, r, f := u.findHash(hash).searchKey(key, hash); f {
		removed = r.delete()
		if removed {
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

func (u *ChainMap[K, V]) Pairs() func() (K, V, bool) {
	cur := u.findHash(0)
	return func() (k K, v V, b bool) {
		for {
			if cur == nil {
				return
			} else if cur.isRelay() {
				cur, _, _ = cur.next()
			} else {
				k, v, b = cur.k, *(*V)(cur.getVPtr()), true
				cur, _, _ = cur.next()
				return
			}
		}
	}
}
