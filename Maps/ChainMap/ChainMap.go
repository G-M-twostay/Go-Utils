package ChainMap

import (
	"GMUtils/Maps"
	"hash/maphash"
	"sync/atomic"
)

type ChainMap[K Maps.Hashable, V any] struct {
	buckets []head[K, V]
	chunk   byte
	seed    maphash.Seed
}

func MakeChainMap[K Maps.Hashable, V any](size uint) *ChainMap[K, V] {
	t := new(ChainMap[K, V])
	t.buckets = make([]head[K, V], size)
	t.seed = maphash.MakeSeed()
	t.chunk = 64
	return t
}

func (u ChainMap[K, V]) rehash(k K) uint64 {
	//b := make([]byte, 8)
	//binary.PutVarint(b, k.Hash())
	//return maphash.Bytes(u.seed, b)
	return uint64(k.Hash())
}

// [0,2^n) to [0,2^n-1),[2^n-1,2^n)
func (u ChainMap[K, V]) splitBucket() {
	newBuckets := make([]head[K, V], len(u.buckets)*2)
	newChunk := u.chunk - 1
	for i, b := range u.buckets {

	}
}

func (u *ChainMap[K, V]) putAt1(h *head[K, V], info *Hold[K, V]) {
	for pre, curPtr := h, h.nextPtr(); ; curPtr = pre.nextPtr() {
		if curPtr == nil {
			if atomic.CompareAndSwapPointer(&pre.nx, curPtr, info.makeNode(curPtr)) {
				break
			}
		} else {
			cur := (*chain[K, V])(curPtr)
			if info.isKey(cur.k) {
				cur.v = info.val
				break
			} else if info.hash <= u.rehash(cur.k) {
				if atomic.CompareAndSwapPointer(&pre.nx, curPtr, info.makeNode(curPtr)) {
					break
				}
			} else if info.hash > u.rehash(cur.k) {
				pre = &cur.head
			} else {
				panic("unexpected case")
			}
		}
	}

}

//func (u *ChainMap[K, V]) putAt2(prev *head[K, V], info *Hold[K, V]) bool {
//	for {
//		curPtr := prev.nextPtr()
//		cur := (*chain[K, V])(curPtr)
//		if curPtr == nil {
//			if atomic.CompareAndSwapPointer(&prev.nx, curPtr, info.makeNode(curPtr)) {
//				break
//			}
//		} else if info.isKey(cur.k) {
//			cur.v = info.val
//			break
//		} else if info.hash <= u.rehash(cur.k) {
//			if atomic.CompareAndSwapPointer(&prev.nx, curPtr, info.makeNode(curPtr)) {
//				break
//			}
//		} else if info.hash > u.rehash(cur.k) {
//			if u.putAt2(&cur.head, info) {
//				break
//			}
//		} else {
//			panic("unexpected case")
//		}
//	}
//
//	return true
//}

func (u ChainMap[K, V]) PrintAll() {
	for cur := u.buckets[0].next(); cur != nil; cur = cur.next() {
		println("k: ", cur.k, ". v: ", cur.v)
	}
}

func (u *ChainMap[K, V]) Put(key K, val V) {
	bucket := &u.buckets[u.rehash(key)>>u.chunk]
	u.putAt1(bucket, &Hold[K, V]{key, val, u.rehash(key)})
}

func (u *ChainMap[K, V]) Remove(key K) {
	for cur := u.buckets[u.rehash(key)>>u.chunk].next(); cur != nil && u.rehash(cur.k) <= u.rehash(key); cur = cur.next() {
		if cur.k.Equal(key) {
			cur.delete()
			break
		}
	}
}

func (u ChainMap[K, V]) Get(key K) (r V) {
	for cur := u.buckets[u.rehash(key)>>u.chunk].next(); cur != nil && u.rehash(cur.k) <= u.rehash(key); cur = cur.next() {
		if cur.k.Equal(key) {
			r = cur.v
			break
		}
	}
	return
}
