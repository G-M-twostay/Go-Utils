package HopMap

import (
	"github.com/g-m-twostay/go-utils"
	"math/bits"
	"unsafe"
)

func New[K comparable, V any](h byte, size, seed uint) *HopMap[K, V] {
	bktLen := 1<<bits.Len(size) + uint(h)
	return &HopMap[K, V]{bkt: make([]bucket[K, V], bktLen), h: h, hashes: make([]uint, bktLen), Seed: Go_Utils.Hasher(seed)}
}

type HopMap[K comparable, V any] struct {
	bkt    []bucket[K, V]
	hashes []uint
	Seed   Go_Utils.Hasher
	sz     uint
	bufs   *omap[K, V]
	h      byte
}

func (u *HopMap[K, V]) hash(k *K) uint {
	return u.Seed.HashMem(unsafe.Pointer(k), unsafe.Sizeof(*k))
	//return *(*uint)(unsafe.Pointer(k))
}

func (u *HopMap[K, V]) mod(hash uint) int {
	return int(hash) & (len(u.bkt) - int(u.h) - 1)
}

func (u *HopMap[K, V]) putOverflow(k *K, v *V, hash uint, i_hash int) bool {
	if u.bufs == nil {
		u.bufs = newOmap[K, V](uint(bits.Len(uint(len(u.bkt)))))
	}
	if u.bufs.avgLen() > 10 {
		return false
	}
	if u.bufs.put(k, v, hash) {
		u.sz++
		u.bkt[i_hash].incCount()
	}
	return true
}

func (u *HopMap[K, V]) expand() {
	newSize := uint((len(u.bkt)-int(u.h))<<1) + uint(u.h)
	M := HopMap[K, V]{bkt: make([]bucket[K, V], newSize), h: u.h, hashes: make([]uint, newSize), Seed: u.Seed}
	for i, e := range u.bkt {
		if e.get(used) {
			if !M.trySet(&e.key, &e.val, u.hashes[i]) {
				M.expand()
				M.trySet(&e.key, &e.val, u.hashes[i])
			}
		}
	}
	for _, b := range u.bufs.bkts_() {
		for _, c := range b {
			M.trySet(&c.key, &c.val, c.hash)
		}
	}

	u.bkt = M.bkt
	u.hashes = M.hashes

	u.bufs = M.bufs
}

func (u *HopMap[K, V]) Size() uint {
	return u.sz
}

func (u *HopMap[K, V]) LoadAndDelete(key K) (V, bool) {
	hash := u.hash(&key)
	i_hash := u.mod(hash)
	if i0 := i_hash; u.bkt[i0].get(hashed) {
		prev := &u.bkt[i0].dHash
		for i1 := i0 + int(u.bkt[i0].dHash); ; i1 = i1 + int(u.bkt[i1].dLink) {
			if u.bkt[i1].get(used) && u.bkt[i1].key == key {
				u.bkt[i1].clr(used) //i1 is no longer used
				u.sz--

				*prev = int8(int(u.bkt[i1].dLink) + i1 - i0) //set i0 to link to what i1 linked to
				u.bkt[i0].set(u.bkt[i1].getRaw(linked))      //if i1 has something linked to

				return u.bkt[i1].val, true
			}
			if !u.bkt[i1].get(linked) {
				break
			}
			i0 = i1
			prev = &u.bkt[i0].dLink
		}
	}
	if u.bkt[i_hash].count() > 0 {
		if v := u.bufs.pop(&key, hash); v != nil {
			u.sz--
			u.bkt[i_hash].decCount()
			return *v, true
		}
	}
	return *new(V), false
}

func (u *HopMap[K, V]) Delete(key K) {
	u.LoadAndDelete(key)
}

func (u *HopMap[K, V]) Load(key K) (V, bool) {
	hash := u.hash(&key)
	i_hash := u.mod(hash)
	if i0 := i_hash; u.bkt[i0].get(hashed) {
		for i1 := i0 + int(u.bkt[i0].dHash); ; i1 = i1 + int(u.bkt[i1].dLink) {
			if u.bkt[i1].get(used) && u.bkt[i1].key == key {
				return u.bkt[i1].val, true
			}
			if !u.bkt[i1].get(linked) {
				break
			}
		}
	}
	if u.bkt[i_hash].count() > 0 {
		return u.bufs.get(&key, hash)
	}
	return *new(V), false
}

// this doesn't mark i_free as usedBkt
func (u *HopMap[K, V]) fillEmpty(i_hash int, i_free int, k *K, v *V) {
	u.bkt[i_free].key, u.bkt[i_free].val = *k, *v
	u.sz++

	u.bkt[i_free].dLink = int8(i_hash + int(u.bkt[i_hash].dHash) - i_free)    //link the next of i_hash after i_free
	u.bkt[i_free].set((u.bkt[i_hash].info >> hashedIndex & 1) << linkedIndex) //if i_hash is hashed, then we set i_free to be linked.
	u.bkt[i_hash].dHash = int8(i_free - i_hash)                               //i_hash now hashes to i_free
	u.bkt[i_hash].set(hashed)
	//if an empty spot within h is found, an insertion will always be made immediately.
}

func (u *HopMap[K, V]) trySet(k *K, v *V, hash uint) bool {
	i_hash := u.mod(hash)
	if u.bkt[i_hash].get(hashed) { //there exists some elements with hash i_hash; check if key already exists.
		for i0 := i_hash + int(u.bkt[i_hash].dHash); ; i0 = i0 + int(u.bkt[i0].dLink) { //find i_hash+dHash: start of the chain
			if u.bkt[i0].key == *k {
				u.bkt[i0].val = *v
				return true
			}
			if !u.bkt[i0].get(linked) {
				break
			}
		}
	}
	if u.bkt[i_hash].count() > 0 && u.bufs.set(k, v, hash) { //check the buffer
		return true
	}
search:
	//now since i_hash is either usedBkt or belongs to some other hash, we need to find an open spot
	for i_free := i_hash; i_free < len(u.bkt); i_free++ {
		if !u.bkt[i_free].get(used) { //found an empty spot
			if i_free-i_hash < int(u.h) { //within h. we insert it here
				u.bkt[i_free].set(used)
				u.fillEmpty(i_hash, i_free, k, v)
				u.hashes[i_free] = hash
				return true
			} else { //j+step>=h. so we find open spot and move it back
			move:
				for i := i_free - int(u.h) + 1; i < i_free; i++ { //iterate in (i_free-H, i_free). if i_free-H<0, then i_free must be in [i_hash, i_hash+H). i_free-H>=0.
					if i0 := i; u.bkt[i0].get(hashed) { //there is some value hashed to i0(i). i0 refers to the prev in the linked iteration.
						//find the start of the chain and iterate in the chain.
						prev := &u.bkt[i0].dHash //initially e0 pointed to e1 by hash
						for i1 := i0 + int(u.bkt[i0].dHash); ; i1 = i1 + int(u.bkt[i1].dLink) {
							if i_free-int(u.h) < i1 && i1 < i_free { //a value e1 with hash i is located in [i_empty-h,i_empty); so we swap e1 with i_free
								//make everything that pointed to e1 from e0 point to i_free
								*prev = int8(i_free - i0)

								u.bkt[i_free].key, u.bkt[i_free].val = u.bkt[i1].key, u.bkt[i1].val //copies e1 to i_free
								u.hashes[i_free] = u.hashes[i1]

								u.bkt[i_free].set(used | u.bkt[i1].getRaw(linked)) //i_free is used; depending on whether i1 is linked, i_free might be linked

								u.bkt[i_free].dLink = int8(int(u.bkt[i1].dLink) + i1 - i_free) //i_free links to the original next of i1

								//now e1 is copied to i_free, and all references to e1 is now to i_free, we can change i_empty to i1
								u.bkt[i1].clr(linked) //i1 is now empty, but it may still hashes to something.

								if i1 < i_hash+int(u.h) {
									u.fillEmpty(i_hash, i1, k, v) //i1 is already used so we don't have to explicitly set it.
									u.hashes[i1] = hash
									return true
								} else {
									u.bkt[i1].clr(used) //set it to usedBkt only when we need more swaps
									i_free = i1
									continue move
								}
							}
							if !u.bkt[i1].get(linked) { //reached the end without finding one.
								break
							}
							i0 = i1                 //store the previous in the chain
							prev = &u.bkt[i0].dLink //now the previous one point to e1 by link
						}
					}
				}
				break search //unable to move usedBkt buckets near i_hash
			}
		}
	}
	return u.putOverflow(k, v, hash, i_hash) //no usedBkt buckets are found
}

//func (u *HopMap[K, V]) tryLoad(k *K, v *V, hash uint) (*V, bool) {
//	i_hash := u.mod(hash)
//	if u.bkt[i_hash].get(hashed) {
//		for i0 := i_hash + int(u.bkt[i_hash].dHash); ; i0 = i0 + int(u.bkt[i0].dLink) {
//			if u.bkt[i0].key == *k {
//				return &u.bkt[i0].val, true
//			}
//			if !u.bkt[i0].get(linked) {
//				break
//			}
//		}
//	}
//	for i_free := i_hash; i_free < len(u.bkt); i_free++ {
//		if !u.usedBkt.Get(i_free) {
//			if i_free-i_hash < int(u.h) {
//				u.usedBkt.Set(i_free)
//				u.fillEmpty(i_hash, i_free, k, v)
//				u.hashes[i_free] = hash
//				return nil, true
//			} else {
//			search:
//				for i := i_free - int(u.h) + 1; i < i_free; i++ {
//					if i0 := i; u.bkt[i0].get(hashed) {
//						prev := &u.bkt[i0].dHash
//						for i1 := i0 + int(u.bkt[i0].dHash); ; i1 = i1 + int(u.bkt[i1].dLink) {
//							if i_free-int(u.h) < i1 && i1 < i_free {
//
//								*prev = offset(i_free - i0)
//
//								u.bkt[i_free].key, u.bkt[i_free].val = u.bkt[i1].key, u.bkt[i1].val //copies e1 to i_free
//								u.hashes[i_free] = u.hashes[i1]
//								u.usedBkt.Set(i_free)
//
//								if u.bkt[i1].get(linked) {
//									u.bkt[i_free].useDeltaLink(int(u.bkt[i1].dLink) + i1 - i_free)
//								}
//
//								u.bkt[i1].clrLink()
//
//								if i1 < i_hash+int(u.h) {
//									u.fillEmpty(i_hash, i1, k, v)
//									u.hashes[i1] = hash
//									return nil, true
//								} else {
//									u.usedBkt.Clr(i1)
//									i_free = i1
//									continue search
//								}
//							}
//							if !u.bkt[i1].get(linked) {
//								break
//							}
//							i0 = i1
//							prev = &u.bkt[i0].dLink
//						}
//					}
//				}
//				return nil, false
//			}
//		}
//	}
//	return nil, false
//}
//
//func (u *HopMap[K, V]) LoadOrStore(key K, val V) (v V, loaded bool) {
//	var r *V
//	for hash, suc := u.hash(&key), false; !suc; r, suc = u.tryLoad(&key, &val, hash) {
//		u.expand()
//	}
//	if loaded = r != nil; loaded {
//		v = *r
//	}
//	return
//}

func (u *HopMap[K, V]) Store(key K, val V) {
	for hash := u.hash(&key); !u.trySet(&key, &val, hash); {
		u.expand()
	}
}

func (u *HopMap[K, V]) HasKey(key K) bool {
	_, r := u.Load(key)
	return r
}

//func (u *HopMap[K, V]) Take() (key K, val V) {
//	if i := u.usedBkt.First(); i > -1 {
//		key, val = u.bkt[i].key, u.bkt[i].val
//	}
//	return
//}
//
//func (u *HopMap[K, V]) Range(f func(K, V) bool) {
//	for i, b := range u.bkt {
//		if u.usedBkt.Get(i) {
//			if !f(b.key, b.val) {
//				return
//			}
//		}
//	}
//}
