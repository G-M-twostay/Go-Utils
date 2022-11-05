package ChainMap

import (
	"GMUtils/Maps"
	"sync"
	"sync/atomic"
	"unsafe"
)

const ( //both inclusive
	maxArrayLen  uint = ^uint(0) >> 1
	maxHashRange uint = ^uint(0)
	maxChunk     byte = 64
)

type ChainMap[K Maps.Hashable, V any] struct {
	buckets             []*node[K]
	chunk, lower, upper byte //1<<chunk=len(buckets)
	size                atomic.Uint64
	l0                  sync.RWMutex
	resizing            atomic.Bool
}

func MakeChainMap[K Maps.Hashable, V any](size uint) *ChainMap[K, V] {
	t := new(ChainMap[K, V])
	t.buckets = make([]*node[K], size)
	for i := uint(0); i < size; i++ {
		t.buckets[i] = new(node[K])
		a := new(node[K])
		a.hash = maxHashRange
		t.buckets[i].nx = (unsafe.Pointer)(a)
	}
	t.upper = 8
	t.chunk = 2
	return t
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
	if u.Size()>>uint(u.chunk) > uint(u.upper) {
		if u.resizing.CompareAndSwap(false, true) {

			newBuckets := make([]*node[K], 1<<(u.chunk+1))

			for i := 0; i < len(u.buckets); i++ {

				newBuckets[i<<1] = u.buckets[i]
				newRelay := new(node[K])
				newBuckets[(i<<1)+1] = newRelay
				newRelay.hash = ((maxHashRange>>uint(u.chunk))+1)*uint(i) + (maxHashRange >> (u.chunk + 1))

				for pre, curPtr := u.buckets[i], u.buckets[i].nextPtr(); ; curPtr = pre.nextPtr() {
					if cur := (*node[K])(curPtr); cur.isRelay() || newRelay.hash <= cur.hash { //put at the first possible position
						newRelay.nx = curPtr
						if atomic.CompareAndSwapPointer(&pre.nx, curPtr, unsafe.Pointer(newRelay)) {
							break
						}
					} else if newRelay.hash > cur.hash {
						pre = cur
					} else {
						panic("unexpected case")
					}
				}
			}

			u.l0.Lock()
			u.buckets = newBuckets
			u.chunk++
			u.l0.Unlock()

			u.resizing.Store(false)
		}
	}

}

func (u *ChainMap[K, V]) tryMerge() {
	if u.Size()>>uint(u.chunk) < uint(u.lower) {
		if u.resizing.CompareAndSwap(false, true) {

			newBuckets := make([]*node[K], 1<<(u.chunk-1))

			for i, v := range u.buckets {
				if i&1 == 0 {
					newBuckets[i>>1] = v
				} else {
					v.delete()
				}
			}

			u.l0.Lock()
			u.buckets = newBuckets
			u.chunk--
			u.l0.Unlock()

			u.resizing.Store(false)
		}
	}
}

func (u *ChainMap[K, V]) findKey(k K) *node[K] {
	return u.findHash(u.rehash(k) >> uint(maxChunk-u.chunk))
}

func (u *ChainMap[K, V]) findHash(hash uint) *node[K] {
	u.l0.RLock()
	defer u.l0.RUnlock()
	return u.buckets[hash]
}

func (u *ChainMap[K, V]) PrintAll() {
	for cur := u.buckets[0]; cur != nil; cur = cur.next() {
		if cur.v == nil {
			println("k: ", cur.k, "; h: ", cur.hash, "; v: relay")
		} else {
			println("k: ", cur.k, "; h: ", cur.hash, "; v: ", *(*V)(cur.v))
		}

	}
}

func (u *ChainMap[K, V]) Put(key K, val V) {
	pre := u.findKey(key)
	for curPtr, hash, vPtr := pre.nextPtr(), u.rehash(key), unsafe.Pointer(&val); ; curPtr = pre.nextPtr() {
		if cur := (*node[K])(curPtr); cur.isRelay() || hash < cur.hash { //put at the last possible position.
			if atomic.CompareAndSwapPointer(&pre.nx, curPtr, unsafe.Pointer(&node[K]{curPtr, key, vPtr, hash, normalState})) {
				u.size.Add(1)
				u.trySplit()
				break
			}
		} else {
			if key.Equal(cur.k) {
				cur.setValuePtr(vPtr)
				break
			} else if hash > cur.hash {
				pre = cur
			} else {
				panic("unexpected case")
			}
		}
	}
}

func (u *ChainMap[K, V]) Remove(key K) {
	u.GetAndRmv(key)
}

func (u *ChainMap[K, V]) Get(key K) (val V) {
	for cur := u.findKey(key).next(); !cur.isRelay() && u.rehash(cur.k) <= u.rehash(key); cur = cur.next() {
		if key.Equal(cur.k) {
			val = *(*V)(cur.valuePtr())
			break
		}
	}
	return
}

func (u *ChainMap[K, V]) HasKey(key K) bool {
	for cur := u.findKey(key).next(); !cur.isRelay() && u.rehash(cur.k) <= u.rehash(key); cur = cur.next() {
		if key.Equal(cur.k) {
			return true
		}
	}
	return false
}

func (u *ChainMap[K, V]) GetOrPut(key K, val V) (oldVal V, putted bool) {
	pre := u.findKey(key)
	for curPtr, hash, vPtr := pre.nextPtr(), u.rehash(key), unsafe.Pointer(&val); ; curPtr = pre.nextPtr() {
		if cur := (*node[K])(curPtr); cur.isRelay() || hash < cur.hash { //put at the last possible position.
			if atomic.CompareAndSwapPointer(&pre.nx, curPtr, unsafe.Pointer(&node[K]{curPtr, key, vPtr, hash, normalState})) {
				u.size.Add(1)
				u.trySplit()
				return oldVal, true
			}
		} else {
			if key.Equal(cur.k) {
				return *(*V)(cur.valuePtr()), false
			} else if hash > cur.hash {
				pre = cur
			} else {
				panic("unexpected case")
			}
		}
	}
}

func (u *ChainMap[K, V]) GetAndRmv(key K) (val V, removed bool) {
	for cur := u.findKey(key).next(); !cur.isRelay() && u.rehash(cur.k) <= u.rehash(key); cur = cur.next() {
		if key.Equal(cur.k) {
			val, removed = *(*V)(cur.valuePtr()), true
			cur.delete()
			u.size.Add(^uint64(0))
			u.tryMerge()
			break
		}
	}
	return
}

func (u *ChainMap[K, V]) Size() uint {
	return uint(u.size.Load())
}

func (u *ChainMap[K, V]) Take() (K, *V) {
	t := u.findHash(0).next()
	return t.k, (*V)(t.valuePtr())
}

func (u *ChainMap[K, V]) Pairs() func() (K, V, bool) {
	cur := u.findHash(0)
	return func() (k K, v V, b bool) {
		for {
			if cur == nil {
				return
			} else if cur.isRelay() {
				cur = cur.next()
			} else {
				k, v, b = cur.k, *(*V)(cur.valuePtr()), true
				cur = cur.next()
				return
			}
		}
	}
}
