package HopMap

import (
	"github.com/cespare/xxhash"
	"github.com/g-m-twostay/go-utils"
	"github.com/g-m-twostay/go-utils/Maps"
	"golang.org/x/exp/constraints"
	"hash/maphash"
	"reflect"
	"unsafe"
)

func New[K constraints.Integer, V any](dl uint, h byte) *HopMap[K, V] {
	bktl := dl + uint(h)
	return &HopMap[K, V]{make([]Bucket[K, V], bktl), Go_Utils.NewBitArray(bktl), h, maphash.MakeSeed()}
}

type HopMap[K constraints.Integer, V any] struct {
	bkt     []Bucket[K, V]
	usedBkt Go_Utils.BitArray
	H       byte
	seed    maphash.Seed
}

func (u *HopMap[K, V]) hash(key *K) uint {
	l := int(unsafe.Sizeof(*key))
	s := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(key)),
		Len:  l,
		Cap:  l,
	}
	//return int(key) & (len(u.bkt) - int(u.H) - 1)
	return uint(xxhash.Sum64(*(*[]byte)(unsafe.Pointer(&s))))
	//return uint(Maps.RTHash(unsafe.Pointer(key), 0, unsafe.Sizeof(*key)))
}

func (u *HopMap[K, V]) mod(hash uint) int {
	return int(hash) & (len(u.bkt) - int(u.H) - 1)
}

func (u *HopMap[K, V]) expand() {
	nl := uint((len(u.bkt)-int(u.H))<<1) + uint(u.H)
	M := HopMap[K, V]{bkt: make([]Bucket[K, V], nl), H: u.H, seed: u.seed, usedBkt: Go_Utils.NewBitArray(nl)}
	for i, e := range u.bkt {
		if u.usedBkt.Get(i) {
			if !M.tryPut(&e.key, &e.val, 0) {
				M.expand()
				M.tryPut(&e.key, &e.val, 0)
			}

		}
	}

	u.bkt = M.bkt
	u.usedBkt = M.usedBkt
}

func (u *HopMap[K, V]) LoadAndDelete(key K) (V, bool) {
	if i0 := u.mod(u.hash(&key)); u.bkt[i0].hashed() {
		prev := &u.bkt[i0].dHash
		for i1 := i0 + u.bkt[i0].deltaHash(); ; i1 = i1 + u.bkt[i1].deltaLink() {
			if u.usedBkt.Get(i1) && u.bkt[i1].key == key {
				u.usedBkt.Down(i1)
				*prev = markLowBit16(u.bkt[i1].deltaLink()+i1-i0, int(u.bkt[i1].dLink&1))
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

func (u *HopMap[K, V]) Get(key K) (V, bool) {
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
	if u.bkt[i_hash].hashed() { //something else already hashed to i_hash, chain it to linked list
		i0 := i_hash + u.bkt[i_hash].deltaHash()
		//find the end of the linked list
		for ; u.bkt[i0].linked(); i0 = i0 + u.bkt[i0].deltaLink() {
		}
		u.bkt[i0].useDeltaLink(i_free - i0) //link i_free after e0.
	} else { //nothing hashed to i_hash
		u.bkt[i_hash].useDeltaHash(i_free - i_hash) //fillEmpty i_free to be hashed to e_hash
	}
	//if an empty spot within H is found, an insertion will always be made immediately.
}

func (u *HopMap[K, V]) Put(key K, val V) {
	for !u.tryPut(&key, &val, 0) {
		u.expand()
	}
}

func (u *HopMap[K, V]) tryPut(k *K, v *V, hash uint) bool {
	i_hash := u.mod(u.hash(k))
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
			if i_free-i_hash < int(u.H) { //within H. we insert it here
				u.usedBkt.Up(i_free)
				u.fillEmpty(i_hash, i_free, k, v)
				return true
			} else { //j+step>=H. so we find open spot and move it back
			search:
				for i := Maps.Max(i_free-int(u.H)+1, 0); i < i_free; i++ { //iterate from i_hash to i_empty
					if i0 := i; u.bkt[i0].hashed() { //there is some value hashed to i0(i). i0 refers to the prev in the linked iteration.
						//find the start of the chain and iterate in the chain.
						prev := &u.bkt[i0].dHash //initially e0 pointed to e1 by hash
						for i1 := i0 + u.bkt[i0].deltaHash(); ; i1 = i1 + u.bkt[i1].deltaLink() {
							if i_free-int(u.H) < i1 && i1 < i_free { //a value e1 with hash i is located in [i_empty-H,i_empty); so we swap e1 with i_free
								//make everything that pointed to e1 from e0 point to i_free
								*prev = markLowBit16(i_free-i0, 1)

								u.bkt[i_free].key, u.bkt[i_free].val = u.bkt[i1].key, u.bkt[i1].val //copies e1 to i_free
								u.usedBkt.Up(i_free)

								u.bkt[i_free].dLink = markLowBit16(u.bkt[i1].deltaLink()+i1-i_free, int(u.bkt[i1].dLink&1)) //i_free links to the original next of i1 if i1 has one
								//now e1 is copied to i_free, and all references to e1 is now to i_free, we can change i_empty to i1
								u.bkt[i1].clrLink() //e1 is now empty, but it may still hashes to something.

								if i1 < i_hash+int(u.H) {
									u.fillEmpty(i_hash, i1, k, v) //i1 is already usedBkt so we don't have to explicitly set it.
									return true
								} else {
									u.usedBkt.Down(i1) //set it to usedBkt only when we need more swaps
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
