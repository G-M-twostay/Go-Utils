package HopMap

import (
	"github.com/cespare/xxhash"
	"github.com/g-m-twostay/go-utils/Maps"
	"golang.org/x/exp/constraints"
	"hash/maphash"
	"reflect"
	"unsafe"
)

func New[K constraints.Integer, V any](dl int, h byte) *HopMap[K, V] {
	return &HopMap[K, V]{make([]Element[K, V], dl+int(h)), h, maphash.MakeSeed()}
}

type HopMap[K constraints.Integer, V any] struct {
	bkt  []Element[K, V]
	H    byte
	seed maphash.Seed
}

func (u *HopMap[K, V]) hash(key K) int {
	l := int(unsafe.Sizeof(key))
	s := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  l,
		Cap:  l,
	}
	//return int(maphash.Bytes(u.seed, *(*[]byte)(unsafe.Pointer(&s)))) & (len(u.bkt) - int(u.H) - 1)
	//return int(key) & (len(u.bkt) - int(u.H) - 1)
	return int(xxhash.Sum64(*(*[]byte)(unsafe.Pointer(&s)))) & (len(u.bkt) - int(u.H) - 1)
	//return int(Maps.RTHash(unsafe.Pointer(&key), 0, unsafe.Sizeof(key))) & (len(u.bkt) - int(u.H) - 1)
}

func (u *HopMap[K, V]) expand() {

	M := HopMap[K, V]{bkt: make([]Element[K, V], (len(u.bkt)-int(u.H))*2+int(u.H)), H: u.H, seed: u.seed}
	for _, e := range u.bkt {
		if e.get(used) {
			M.Put(e.key, e.val)
		}
	}
	u.bkt = M.bkt

}

func (u *HopMap[K, V]) Get(key K) (V, bool) {
	if i0 := u.hash(key); u.bkt[i0].get(hashed) {
		for i1 := i0 + int(u.bkt[i0].hashOS); ; i1 = i1 + int(u.bkt[i1].linkOS) {
			if u.bkt[i1].get(used) && u.bkt[i1].key == key {
				return u.bkt[i1].val, true
			}
			if !u.bkt[i1].get(linked) {
				break
			}
		}
	}
	return *new(V), false
}

func (u *HopMap[K, V]) fillEmpty(i_hash int, i_empty int, k *K, v *V) {
	u.bkt[i_empty].key, u.bkt[i_empty].val = *k, *v
	u.bkt[i_empty].set(used)
	if u.bkt[i_hash].get(hashed) { //something else already hashed to i_hash, chain it to linked list
		i0 := i_hash + int(u.bkt[i_hash].hashOS)
		for ; u.bkt[i0].get(linked); i0 = i0 + int(u.bkt[i0].linkOS) {
			//find the end of the linked list
		}
		u.bkt[i0].linkOS = int16(i_empty - i0) //link e_empty after e0.
		u.bkt[i0].set(linked)
	} else { //nothing hashed to i_hash
		u.bkt[i_hash].hashOS = int16(i_empty - i_hash) //fillEmpty e_empty to be hashed to e_hash
		u.bkt[i_hash].set(hashed)
	}
	//if an empty spot within H is found, an insertion will always be made immediately.
}

func (u *HopMap[K, V]) moveAndPut(i_empty int) int {
	for i := Maps.Max(i_empty-int(u.H)+1, 0); i < i_empty; i++ { //iterate from i_hash to i_empty
		if i0 := i; u.bkt[i0].get(hashed) { //there is some value hashed to i0(i). i0 refers to the prev in the linked iteration.
			//find the start of the chain and iterate in the chain.
			prev := &u.bkt[i0].hashOS //initially e0 pointed to e1 by hash
			for i1 := i0 + int(u.bkt[i0].hashOS); ; i1 = i1 + int(u.bkt[i1].linkOS) {
				if i_empty-int(u.H) < i1 && i1 < i_empty { //a value e1 with hash i is located in [i_empty-H,i_empty); so we swap e1 with e_empty
					//make everything that pointed to e1 from e0 point to e_empty
					*prev = int16(i_empty - i0)

					u.bkt[i_empty].key, u.bkt[i_empty].val = u.bkt[i1].key, u.bkt[i1].val //copies e1 to e_empty
					u.bkt[i_empty].info |= u.bkt[i1].info &^ hashed                       //we don't want to change hash information
					u.bkt[i_empty].linkOS = int16(int(u.bkt[i1].linkOS) + i1 - i_empty)   //e_empty links to the original next of e1
					//now e1 is copied to e_empty, and all references to e1 is now to e_empty, we can change i_empty to i1
					u.bkt[i1].clear(used | linked) //e1 is now empty, but it may still hashes to something.
					return i1
				}
				if !u.bkt[i1].get(linked) { //reached the end without finding one.
					break
				}
				i0 = i1                  //store the previous in the chain
				prev = &u.bkt[i0].linkOS //now the previous one point to e1 by link
			}
		}
	}
	return i_empty
}

func (u *HopMap[K, V]) Put(key K, val V) {
	i_hash := u.hash(key)
	if u.bkt[i_hash].get(hashed) { //there exists some elements with hash i_hash; check if key already exists.
		for i0 := i_hash + int(u.bkt[i_hash].hashOS); ; i0 = i0 + int(u.bkt[i0].linkOS) { //find i_hash+hashOS: start of the chain
			if u.bkt[i0].key == key {
				u.bkt[i0].val = val
				return
			}
			if !u.bkt[i0].get(linked) {
				break
			}
		}
	}
	for ; ; i_hash = u.hash(key) {
	search:
		//now since i_hash is either free or belongs to some other hash, we need to find an open spot
		for step := i_hash; step < len(u.bkt); step++ {
			if i_empty := step; !u.bkt[i_empty].get(used) { //found an empty spot
				if i_empty-i_hash < int(u.H) { //within H. we insert it here
					u.fillEmpty(i_hash, i_empty, &key, &val)
					return //if an empty spot within H is found, an insertion will always be made immediately.
				} else { //j+step>=H. so we find open spot and move it back
					for {
						if i_new := u.moveAndPut(i_empty); i_new == i_empty {
							break search
						} else if i_new < i_hash+int(u.H) {
							u.fillEmpty(i_hash, i_new, &key, &val)
							return
						} else {
							i_empty = i_new
						}
					}
				}
			}
		}
		//fmt.Printf("prev: %d; %v\n", len(u.bkt)-int(u.H), u.bkt)
		u.expand()
		//fmt.Printf("post: %d; %v\n", len(u.bkt)-int(u.H), u.bkt)
	}

}
