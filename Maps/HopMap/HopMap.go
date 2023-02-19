package HopMap

import (
	"github.com/g-m-twostay/go-utils"
	"math/bits"
	"unsafe"
)

func New[K comparable, V any](h byte, size, seed uint) *HopMap[K, V] {
	bktLen := 1<<bits.Len(size) + uint(h)
	return &HopMap[K, V]{bkt: make([]Bucket[K, V], bktLen), usedBkt: Go_Utils.NewBitArray(bktLen), h: h, hashes: make([]uint, bktLen), Seed: Go_Utils.Hasher(seed)}
}

type HopMap[K comparable, V any] struct {
	bkt     []Bucket[K, V]
	usedBkt Go_Utils.BitArray
	hashes  []uint
	Seed    Go_Utils.Hasher
	sz      uint
	h       byte
}

func (u *HopMap[K, V]) hash(k *K) uint {
	return u.Seed.HashMem(unsafe.Pointer(k), unsafe.Sizeof(*k))
}

func (u *HopMap[K, V]) mod(hash uint) int {
	return int(hash) & (len(u.bkt) - int(u.h) - 1)
}

func (u *HopMap[K, V]) expand() {
	newSize := uint((len(u.bkt)-int(u.h))<<1) + uint(u.h)
	M := HopMap[K, V]{bkt: make([]Bucket[K, V], newSize), h: u.h, usedBkt: Go_Utils.NewBitArray(newSize), hashes: make([]uint, newSize), Seed: u.Seed}
	for i, e := range u.bkt {
		if u.usedBkt.Get(i) {
			if !M.trySet(&e.key, &e.val, u.hashes[i]) {
				M.expand()
				M.trySet(&e.key, &e.val, u.hashes[i])
			}
		}
	}

	u.bkt = M.bkt
	u.usedBkt = M.usedBkt
	u.hashes = M.hashes
}

func (u *HopMap[K, V]) Size() uint {
	return u.sz
}

func (u *HopMap[K, V]) LoadAndDelete(key K) (V, bool) {
	if i0 := u.mod(u.hash(&key)); u.bkt[i0].hashed() {
		prev := &u.bkt[i0].dHash
		for i1 := i0 + u.bkt[i0].deltaHash(); ; i1 = i1 + u.bkt[i1].deltaLink() {
			if u.usedBkt.Get(i1) && u.bkt[i1].key == key {
				u.usedBkt.Clr(i1)
				u.sz--
				if u.bkt[i1].linked() {
					*prev = offset(u.bkt[i1].deltaLink() + i1 - i0)
				} else {
					*prev = 0
				}
				return u.bkt[i1].val, true
			}
			if !u.bkt[i1].linked() {
				break
			}
			i0 = i1
			prev = &u.bkt[i0].dLink
		}
	}
	return *new(V), false
}

func (u *HopMap[K, V]) Delete(key K) {
	u.LoadAndDelete(key)
}

func (u *HopMap[K, V]) Load(key K) (V, bool) {
	if i0 := u.mod(u.hash(&key)); u.bkt[i0].hashed() {
		for i1 := i0 + u.bkt[i0].deltaHash(); ; i1 = i1 + u.bkt[i1].deltaLink() {
			if u.usedBkt.Get(i1) && u.bkt[i1].key == key {
				return u.bkt[i1].val, true
			}
			if !u.bkt[i1].linked() {
				break
			}
		}
	}
	return *new(V), false
}

// this doesn't mark i_free as usedBkt
func (u *HopMap[K, V]) fillEmpty(i_hash int, i_free int, k *K, v *V) {
	u.bkt[i_free].key, u.bkt[i_free].val = *k, *v
	u.sz++
	if u.bkt[i_hash].hashed() { //something else already hashed to i_hash, chain it to after i_free
		u.bkt[i_free].useDeltaLink(i_hash + u.bkt[i_hash].deltaHash() - i_free)
	}
	u.bkt[i_hash].useDeltaHash(i_free - i_hash) //i_hash now hashes to i_free
	//if an empty spot within h is found, an insertion will always be made immediately.
}

func (u *HopMap[K, V]) trySet(k *K, v *V, hash uint) bool {
	i_hash := u.mod(hash)
	if u.bkt[i_hash].hashed() { //there exists some elements with hash i_hash; check if key already exists.
		for i0 := i_hash + u.bkt[i_hash].deltaHash(); ; i0 = i0 + u.bkt[i0].deltaLink() { //find i_hash+dHash: start of the chain
			if u.bkt[i0].key == *k {
				u.bkt[i0].val = *v
				return true
			}
			if !u.bkt[i0].linked() {
				break
			}
		}
	}
	//now since i_hash is either usedBkt or belongs to some other hash, we need to find an open spot
	for i_free := i_hash; i_free < len(u.bkt); i_free++ {
		if !u.usedBkt.Get(i_free) { //found an empty spot
			if i_free-i_hash < int(u.h) { //within h. we insert it here
				u.usedBkt.Set(i_free)
				u.fillEmpty(i_hash, i_free, k, v)
				u.hashes[i_free] = hash
				return true
			} else { //j+step>=h. so we find open spot and move it back
			search:
				for i := i_free - int(u.h) + 1; i < i_free; i++ { //iterate in (i_free-H, i_free). if i_free-H<0, then i_free must be in [i_hash, i_hash+H). i_free-H>=0.
					if i0 := i; u.bkt[i0].hashed() { //there is some value hashed to i0(i). i0 refers to the prev in the linked iteration.
						//find the start of the chain and iterate in the chain.
						prev := &u.bkt[i0].dHash //initially e0 pointed to e1 by hash
						for i1 := i0 + u.bkt[i0].deltaHash(); ; i1 = i1 + u.bkt[i1].deltaLink() {
							if i_free-int(u.h) < i1 && i1 < i_free { //a value e1 with hash i is located in [i_empty-h,i_empty); so we swap e1 with i_free
								//make everything that pointed to e1 from e0 point to i_free
								*prev = offset(i_free - i0)

								u.bkt[i_free].key, u.bkt[i_free].val = u.bkt[i1].key, u.bkt[i1].val //copies e1 to i_free
								u.hashes[i_free] = u.hashes[i1]
								u.usedBkt.Set(i_free)

								if u.bkt[i1].linked() { //i_free links to the original next of i1 if i1 has one
									u.bkt[i_free].useDeltaLink(u.bkt[i1].deltaLink() + i1 - i_free)
								}

								//now e1 is copied to i_free, and all references to e1 is now to i_free, we can change i_empty to i1
								u.bkt[i1].clrLink() //e1 is now empty, but it may still hashes to something.

								if i1 < i_hash+int(u.h) {
									u.fillEmpty(i_hash, i1, k, v) //i1 is already usedBkt so we don't have to explicitly set it.
									u.hashes[i1] = hash
									return true
								} else {
									u.usedBkt.Clr(i1) //set it to usedBkt only when we need more swaps
									i_free = i1
									continue search
								}
							}
							if !u.bkt[i1].linked() { //reached the end without finding one.
								break
							}
							i0 = i1                 //store the previous in the chain
							prev = &u.bkt[i0].dLink //now the previous one point to e1 by link
						}
					}
				}
				return false //unable to move usedBkt buckets near i_hash
			}
		}
	}
	return false //no usedBkt buckets are found
}

func (u *HopMap[K, V]) tryLoad(k *K, v *V, hash uint) (*V, bool) {
	i_hash := u.mod(hash)
	if u.bkt[i_hash].hashed() {
		for i0 := i_hash + u.bkt[i_hash].deltaHash(); ; i0 = i0 + u.bkt[i0].deltaLink() {
			if u.bkt[i0].key == *k {
				return &u.bkt[i0].val, true
			}
			if !u.bkt[i0].linked() {
				break
			}
		}
	}
	for i_free := i_hash; i_free < len(u.bkt); i_free++ {
		if !u.usedBkt.Get(i_free) {
			if i_free-i_hash < int(u.h) {
				u.usedBkt.Set(i_free)
				u.fillEmpty(i_hash, i_free, k, v)
				u.hashes[i_free] = hash
				return nil, true
			} else {
			search:
				for i := i_free - int(u.h) + 1; i < i_free; i++ {
					if i0 := i; u.bkt[i0].hashed() {
						prev := &u.bkt[i0].dHash
						for i1 := i0 + u.bkt[i0].deltaHash(); ; i1 = i1 + u.bkt[i1].deltaLink() {
							if i_free-int(u.h) < i1 && i1 < i_free {

								*prev = offset(i_free - i0)

								u.bkt[i_free].key, u.bkt[i_free].val = u.bkt[i1].key, u.bkt[i1].val //copies e1 to i_free
								u.hashes[i_free] = u.hashes[i1]
								u.usedBkt.Set(i_free)

								if u.bkt[i1].linked() {
									u.bkt[i_free].useDeltaLink(u.bkt[i1].deltaLink() + i1 - i_free)
								}

								u.bkt[i1].clrLink()

								if i1 < i_hash+int(u.h) {
									u.fillEmpty(i_hash, i1, k, v)
									u.hashes[i1] = hash
									return nil, true
								} else {
									u.usedBkt.Clr(i1)
									i_free = i1
									continue search
								}
							}
							if !u.bkt[i1].linked() {
								break
							}
							i0 = i1
							prev = &u.bkt[i0].dLink
						}
					}
				}
				return nil, false
			}
		}
	}
	return nil, false
}

func (u *HopMap[K, V]) LoadOrStore(key K, val V) (v V, loaded bool) {
	var r *V
	for hash, suc := u.hash(&key), false; !suc; r, suc = u.tryLoad(&key, &val, hash) {
		u.expand()
	}
	if loaded = r != nil; loaded {
		v = *r
	}
	return
}

func (u *HopMap[K, V]) Store(key K, val V) {
	for hash := u.hash(&key); !u.trySet(&key, &val, hash); {
		u.expand()
	}
}

func (u *HopMap[K, V]) HasKey(key K) bool {
	_, r := u.Load(key)
	return r
}

func (u *HopMap[K, V]) Take() (key K, val V) {
	if i := u.usedBkt.First(); i > -1 {
		key, val = u.bkt[i].key, u.bkt[i].val
	}
	return
}

func (u *HopMap[K, V]) Range(f func(K, V) bool) {
	for i, b := range u.bkt {
		if u.usedBkt.Get(i) {
			if !f(b.key, b.val) {
				return
			}
		}
	}
}
