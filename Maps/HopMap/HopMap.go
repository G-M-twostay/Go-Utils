package HopMap

import (
	"golang.org/x/exp/constraints"
	"hash/maphash"
	"reflect"
	"unsafe"
)

func New[K constraints.Integer, V any](h byte) *HopMap[K, V] {
	return &HopMap[K, V]{make([]Element[K, V], 8), h, maphash.MakeSeed()}
}

type HopMap[K constraints.Integer, V any] struct {
	bkt  []Element[K, V]
	H    byte
	seed maphash.Seed
}

func (u *HopMap[K, V]) hash(key K) uint {
	l := int(unsafe.Sizeof(key))
	s := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  l,
		Cap:  l,
	}
	return uint(maphash.Bytes(u.seed, *(*[]byte)(unsafe.Pointer(&s))))
	//return uint(key)
}

func (u *HopMap[K, V]) modGet(index int) (int, *Element[K, V]) {
	t := index & (len(u.bkt) - 1)
	return t, &u.bkt[t]
}

func (u *HopMap[K, V]) expand() {
	M := HopMap[K, V]{buckets: make([]Element[K, V], len(u.bkt)*2), H: u.H, seed: u.seed}
	for _, e := range u.bkt {
		if e.get(used) {
			M.Put(e.key, e.val)
		}
	}
	u.bkt = M.bkt
}

func (u *HopMap[K, V]) Get(key K) (V, bool) {
	if i0, e0 := u.modGet(int(u.hash(key))); e0.get(hashed) {
		for i1, e1 := u.modGet(i0 + int(e0.hashOS)); ; i1, e1 = u.modGet(i1 + int(e1.linkOS)) {
			if e1.get(used) && e1.key == key {
				return e1.val, true
			}
			if !e1.get(linked) {
				break
			}
		}
	}
	return *new(V), false
}

func (u *HopMap[K, V]) within(v, high int) bool {
	if low, _ := u.modGet(high - int(u.H)); high >= low {
		return high > v && v > low
	} else {
		return v > low || v < high
	}
}

func (u *HopMap[K, V]) fillEmpty(i_hash int, i_empty int, k *K, v *V) {
	u.bkt[i_empty].key, u.bkt[i_empty].val = *k, *v
	u.bkt[i_empty].set(used)
	if u.bkt[i_hash].get(hashed) { //something else already hashed to i_hash, chain it to linked list
		i0, e0 := u.modGet(i_hash + int(u.bkt[i_hash].hashOS))
		for ; e0.get(linked); i0, e0 = u.modGet(i0 + int(e0.linkOS)) {
			//find the end of the linked list
		}
		e0.linkOS = int16(i_empty - i0) //link e_empty after e0.
		e0.set(linked)
	} else { //nothing hashed to i_hash
		u.bkt[i_hash].hashOS = int16(i_empty - i_hash) //fillEmpty e_empty to be hashed to e_hash
		u.bkt[i_hash].set(hashed)
	}
	//if an empty spot within H is found, an insertion will always be made immediately.
}

func (u *HopMap[K, V]) moveBack(i_hash int, i_empty int) int {
	for i := i_hash + 1; i != i_empty; i++ { //iterate from i_hash to i_empty
		if i0, e0 := u.modGet(i); e0.get(hashed) { //there is some value hashed to i0(i). i0 refers to the prev in the linked iteration.
			//find the start of the chain and iterate in the chain.
			prev := &e0.hashOS //initially e0 pointed to e1 by hash
			for i1, e1 := u.modGet(i0 + int(e0.hashOS)); ; i1, e1 = u.modGet(i1 + int(e1.linkOS)) {
				if u.within(i1, i_empty) { //a value e1 with hash i is located in [i_empty-H,i_empty); so we swap e1 with e_empty
					//make everything that pointed to e1 from e0 point to e_empty
					*prev = int16(i_empty - i0)

					u.bkt[i_empty].key, u.bkt[i_empty].val = e1.key, e1.val      //copies e1 to e_empty
					u.bkt[i_empty].info |= e1.info &^ hashed                     //we don't want to change hash information
					u.bkt[i_empty].linkOS = int16(int(e1.linkOS) + i1 - i_empty) //e_empty links to the original next of e1
					//now e1 is copied to e_empty, and all references to e1 is now to e_empty, we can change i_empty to i1
					e1.clear(used | linked) //e1 is now empty, but it may still hashes to something.
					return i1
				}
				if !e1.get(linked) { //reached the end without finding one.
					break
				}
				i0, e0 = i1, e1   //store the previous in the chain
				prev = &e0.linkOS //now the previous one point to e1 by link
			}
		}
	}
	return i_empty
}

func (u *HopMap[K, V]) Put(key K, val V) {
	i_hash, e_hash := u.modGet(int(u.hash(key)))
	if e_hash.get(hashed) { //there exists some elements with hash i_hash; check if key already exists.
		for i0, e0 := u.modGet(i_hash + int(e_hash.hashOS)); ; i0, e0 = u.modGet(i0 + int(e0.linkOS)) { //find i_hash+hashOS: start of the chain
			if e0.key == key {
				e0.val = val
				return
			}
			if !e0.get(linked) {
				break
			}
		}
	}
insert:
	//now since i_hash is either free or belongs to some other hash, we need to find an open spot
	for step := 1; step < len(u.bkt); step++ {
		if i_empty, e_empty := u.modGet(step + i_hash); !e_empty.get(used) { //found an empty spot
			if step < int(u.H) { //within H. we insert it here
				u.fillEmpty(i_hash, i_empty, &key, &val)
				return //if an empty spot within H is found, an insertion will always be made immediately.
			} else { //j+step>=H. so we find open spot and move it back
			move:
				for i := i_hash; i != i_empty; i++ { //iterate from i_hash to i_empty
					if i0, e0 := u.modGet(i); e0.get(hashed) { //there is some value hashed to i0(i). i0 refers to the prev in the linked iteration.
						//find the start of the chain and iterate in the chain.
						prev := &e0.hashOS //initially e0 pointed to e1 by hash
						for i1, e1 := u.modGet(i0 + int(e0.hashOS)); ; i1, e1 = u.modGet(i1 + int(e1.linkOS)) {
							if u.within(i1, i_empty) { //a value e1 with hash i is located in [i_empty-H,i_empty); so we swap e1 with e_empty
								//make everything that pointed to e1 from e0 point to e_empty
								*prev = int16(i_empty - i0)

								e_empty.key, e_empty.val = e1.key, e1.val             //copies e1 to e_empty
								e_empty.info |= e1.info &^ hashed                     //we don't want to change hash information
								e_empty.linkOS = int16(int(e1.linkOS) + i1 - i_empty) //e_empty links to the original next of e1
								//now e1 is copied to e_empty, and all references to e1 is now to e_empty, we can change i_empty to i1
								i_empty, e_empty = i1, e1
								e1.clear(used | linked) //e1 is now empty, but it may still hashes to something.

								if t, _ := u.modGet(i_empty - i_hash); t < int(u.H) { //i_empty is now within i_hash+H
									u.fillEmpty(i_hash, i_empty, &key, &val)
									return
								} else { //need move more empty spots
									continue move
								}
							}
							if !e1.get(linked) { //reached the end without finding one.
								break
							}
							i0, e0 = i1, e1   //store the previous in the chain
							prev = &e0.linkOS //now the previous one point to e1 by link
						}
					}
				}
				goto expand
			}
		}
	}
expand:
	u.expand()                                  //if the loop ended without returning, that means no empty spot is found
	i_hash, e_hash = u.modGet(int(u.hash(key))) //since bucket size changed, recalculate i_hash, e_hash
	goto insert
}
