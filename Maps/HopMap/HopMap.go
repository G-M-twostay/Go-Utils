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
	return &HopMap[K, V]{make([]Element[K, V], bktl), Go_Utils.NewBitArray(bktl), h, maphash.MakeSeed()}
}

type HopMap[K constraints.Integer, V any] struct {
	bkt  []Element[K, V]
	used Go_Utils.BitArray
	H    byte
	seed maphash.Seed
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
	M := HopMap[K, V]{bkt: make([]Element[K, V], nl), H: u.H, seed: u.seed, used: Go_Utils.NewBitArray(nl)}
	for i, e := range u.bkt {
		if u.used.Get(i) {
			if !M.tryPut(&e.key, &e.val, 0) {
				M.expand()
				M.tryPut(&e.key, &e.val, 0)
			}

		}
	}

	u.bkt = M.bkt
	u.used = M.used
}

func (u *HopMap[K, V]) LoadAndDelete(key K) (V, bool) {
	if i0 := u.mod(u.hash(&key)); u.bkt[i0].hashed() {
		prev := &u.bkt[i0].hashOS
		for i1 := i0 + u.bkt[i0].hashOffSet(); ; i1 = i1 + u.bkt[i1].linkOffSet() {
			if u.used.Get(i1) && u.bkt[i1].key == key {
				u.used.Down(i1)
				*prev = markLowestBit16(u.bkt[i1].linkOffSet()+i1-i0, int(u.bkt[i1].linkOS&1))
				return u.bkt[i1].val, true
			}
			if !u.bkt[i1].linked() {
				break
			}
			i0 = i1
			prev = &u.bkt[i0].linkOS
		}
	}
	return *new(V), false
}

func (u *HopMap[K, V]) Get(key K) (V, bool) {
	if i0 := u.mod(u.hash(&key)); u.bkt[i0].hashed() {
		for i1 := i0 + u.bkt[i0].hashOffSet(); ; i1 = i1 + u.bkt[i1].linkOffSet() {
			if u.used.Get(i1) && u.bkt[i1].key == key {
				return u.bkt[i1].val, true
			}
			if !u.bkt[i1].linked() {
				break
			}
		}
	}
	return *new(V), false
}

// this doesn't mark i_free as used
func (u *HopMap[K, V]) fillEmpty(i_hash int, i_free int, k *K, v *V) {
	u.bkt[i_free].key, u.bkt[i_free].val = *k, *v
	if u.bkt[i_hash].hashed() { //something else already hashed to i_hash, chain it to linked list
		i0 := i_hash + u.bkt[i_hash].hashOffSet()
		//find the end of the linked list
		for ; u.bkt[i0].linked(); i0 = i0 + u.bkt[i0].linkOffSet() {
		}
		u.bkt[i0].UseLinkOffSet(i_free - i0) //link i_free after e0.
	} else { //nothing hashed to i_hash
		u.bkt[i_hash].UseHashOffSet(i_free - i_hash) //fillEmpty i_free to be hashed to e_hash
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
		for i0 := i_hash + u.bkt[i_hash].hashOffSet(); ; i0 = i0 + u.bkt[i0].linkOffSet() { //find i_hash+hashOS: start of the chain
			if u.bkt[i0].key == *k {
				u.bkt[i0].val = *v
				return true
			}
			if !u.bkt[i0].linked() {
				break
			}
		}
	}
	//now since i_hash is either used or belongs to some other hash, we need to find an open spot
	for i_free := i_hash; i_free < len(u.bkt); i_free++ {
		if !u.used.Get(i_free) { //found an empty spot
			if i_free-i_hash < int(u.H) { //within H. we insert it here
				u.used.Up(i_free)
				u.fillEmpty(i_hash, i_free, k, v)
				return true
			} else { //j+step>=H. so we find open spot and move it back
			search:
				for i := Maps.Max(i_free-int(u.H)+1, 0); i < i_free; i++ { //iterate from i_hash to i_empty
					if i0 := i; u.bkt[i0].hashed() { //there is some value hashed to i0(i). i0 refers to the prev in the linked iteration.
						//find the start of the chain and iterate in the chain.
						prev := &u.bkt[i0].hashOS //initially e0 pointed to e1 by hash
						for i1 := i0 + u.bkt[i0].hashOffSet(); ; i1 = i1 + u.bkt[i1].linkOffSet() {
							if i_free-int(u.H) < i1 && i1 < i_free { //a value e1 with hash i is located in [i_empty-H,i_empty); so we swap e1 with i_free
								//make everything that pointed to e1 from e0 point to i_free
								*prev = markLowestBit16(i_free-i0, 1)

								u.bkt[i_free].key, u.bkt[i_free].val = u.bkt[i1].key, u.bkt[i1].val //copies e1 to i_free
								u.used.Up(i_free)

								u.bkt[i_free].linkOS = markLowestBit16(u.bkt[i1].linkOffSet()+i1-i_free, int(u.bkt[i1].linkOS&1)) //i_free links to the original next of i1 if i1 has one
								//now e1 is copied to i_free, and all references to e1 is now to i_free, we can change i_empty to i1
								u.bkt[i1].clrLink() //e1 is now empty, but it may still hashes to something.

								if i1 < i_hash+int(u.H) {
									u.fillEmpty(i_hash, i1, k, v) //i1 is already used so we don't have to explicitly set it.
									return true
								} else {
									u.used.Down(i1) //set it to used only when we need more swaps
									i_free = i1
									continue search
								}
							}
							if !u.bkt[i1].linked() { //reached the end without finding one.
								break
							}
							i0 = i1                  //store the previous in the chain
							prev = &u.bkt[i0].linkOS //now the previous one point to e1 by link
						}
					}
				}
				return false //unable to move used buckets near i_hash
			}
		}
	}
	return false //no used buckets are found
}
