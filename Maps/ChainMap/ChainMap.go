package ChainMap

import (
	"GMUtils/Maps"
	"fmt"
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

	var f = func(hash uint, nx *node[K]) *node[K] {
		n, s := new(node[K]), new(state[K])
		s.nx = nx
		n.hash, n.s = hash, unsafe.Pointer(s)
		return n
	}

	M.buckets = make([]*node[K], 1)
	lastRelay := f(M.maxHash, nil)
	M.buckets[0] = f(0, lastRelay)

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
				newRelay := new(node[K])
				newBuckets[(i<<1)+1] = newRelay
				newRelay.hash = ((u.maxHash>>uint(u.chunk))+1)*uint(i) + (u.maxHash >> (u.chunk + 1))
				tempState := new(state[K])
				newRelay.s = (unsafe.Pointer)(tempState)

			search:
				for left := v; ; left = v {
					for {
						right, leftStatePtr := left.next()
						if right.isRelay() || newRelay.hash < right.hash {
							tempState.nx = right

							if atomic.CompareAndSwapPointer(&left.s, leftStatePtr, unsafe.Pointer(&state[K]{false, newRelay})) {
								break search
							} else {
								continue search
							}
						} else {
							left = right
						}
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
			//println("merging")
			newBuckets := make([]*node[K], 1<<(u.chunk-1))

			for i, v := range u.buckets {
				if i&1 == 0 {
					newBuckets[i>>1] = v
				} else {
					v.delete()
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

func (u *ChainMap[K, V]) findKey(k K) *node[K] {
	u.bucketsLock.RLock()
	defer u.bucketsLock.RUnlock()
	return u.buckets[u.rehash(k)>>uint(u.maxChunk-u.chunk)]
}

func (u *ChainMap[K, V]) findHash(hash uint) *node[K] {
	u.bucketsLock.RLock()
	defer u.bucketsLock.RUnlock()
	return u.buckets[hash]
}

func (u *ChainMap[K, V]) Put(key K, val V) {
	hash, vPtr := u.rehash(key), unsafe.Pointer(&val)
search:
	for left := u.findKey(key); ; left = u.findKey(key) {
		for {
			right, leftStatePtr := left.next()
			if hash <= right.hash {
				newNode := &node[K]{key, vPtr, hash, unsafe.Pointer(&state[K]{false, right})}
				if atomic.CompareAndSwapPointer(&left.s, leftStatePtr, unsafe.Pointer(&state[K]{false, newNode})) {
					u.size.Add(1)
					//println("added", newNode.hash)
					u.trySplit()
					return
				} else {
					continue search
				}
			} else if key.Equal(right.k) {
				right.setVPtr(vPtr)
				return
			} else {
				left = right
			}
		}

	}
}

func (u *ChainMap[K, V]) Remove(key K) {
	u.GetAndRmv(key)
}

func (u *ChainMap[K, V]) Get(key K) (val V) {
	for cur, _ := u.findKey(key).next(); cur != nil && cur.hash <= u.rehash(key); cur, _ = cur.next() {
		if key.Equal(cur.k) {
			val = *(*V)(cur.getVPtr())
			break
		}
	}
	return
}

func (u *ChainMap[K, V]) HasKey(key K) bool {
	for cur, _ := u.findKey(key).next(); cur != nil && cur.hash <= u.rehash(key); cur, _ = cur.next() {
		if key.Equal(cur.k) {
			return true
		}
	}
	return false
}

func (u *ChainMap[K, V]) GetOrPut(key K, val V) (V, bool) {
	hash, vPtr := u.rehash(key), unsafe.Pointer(&val)
search:
	for left := u.findKey(key); ; left = u.findKey(key) {
		for {
			right, leftStatePtr := left.next()
			if hash <= right.hash {
				newNode := &node[K]{key, vPtr, hash, unsafe.Pointer(&state[K]{false, right})}
				if atomic.CompareAndSwapPointer(&left.s, leftStatePtr, unsafe.Pointer(&state[K]{false, newNode})) {
					u.size.Add(1)
					u.trySplit()
					return *new(V), false
				} else {
					continue search
				}
			} else if key.Equal(right.k) {
				return *(*V)(right.getVPtr()), true
			} else {
				left = right
			}
		}

	}
}

func (u *ChainMap[K, V]) GetAndRmv(key K) (val V, removed bool) {
	for cur, _ := u.findKey(key).next(); cur != nil && cur.hash <= u.rehash(key); cur, _ = cur.next() {
		if key.Equal(cur.k) {
			if cur.delete() {
				u.size.Add(^uint64(0))
				u.tryMerge()
				val, removed = *(*V)(cur.getVPtr()), true
			}
			break
		}
	}
	return
}

func (u *ChainMap[K, V]) Size() uint {
	//println(u.chunk, u.maxChunk, u.maxHash)
	return uint(u.size.Load())
}

func (u *ChainMap[K, V]) Take() (K, V) {
	t, _ := u.findHash(0).next()
	return t.k, *(*V)(t.getVPtr())
}

func (u *ChainMap[K, V]) PrintAll() {
	cur := u.findHash(0)
	//oldh := uint(0)
	for ; cur != nil; cur, _ = cur.next() {
		//fmt.Printf("oldh: %d; curh: %d; ordered: %v\n", oldh, cur.hash, cur.hash >= oldh)
		//oldh = cur.hash
		if cur.isRelay() {
			fmt.Printf("relay: %#v. Hash: %d\n", cur.k, cur.hash)
		} else {
			fmt.Printf("key: %#v. Hash: %d\n", cur.k, cur.hash)
		}
	}
}

func (u *ChainMap[K, V]) Pairs() func() (K, V, bool) {
	cur := u.findHash(0)
	return func() (k K, v V, b bool) {
		for {
			if cur == nil {
				return
			} else if cur.isRelay() {
				cur, _ = cur.next()
			} else {
				k, v, b = cur.k, *(*V)(cur.getVPtr()), true
				cur, _ = cur.next()
				return
			}
		}
	}
}
