package HopMap

import "golang.org/x/exp/constraints"

func New[K constraints.Integer, V any]() *HopMap[K, V] {
	return &HopMap[K, V]{make([]Element[K, V], 8), 3}
}

type HopMap[K constraints.Integer, V any] struct {
	buckets []Element[K, V]
	H       byte
}

func (u *HopMap[K, V]) hash(key K) uint {
	return uint(key)
}

func (u *HopMap[K, V]) modGet(index uint) (uint, *Element[K, V]) {
	t := index & uint(len(u.buckets)-1)
	return t, &u.buckets[t]
}

func (u *HopMap[K, V]) expand() {
	M := HopMap[K, V]{buckets: make([]Element[K, V], len(u.buckets)*2), H: u.H}
	for i, e := range u.buckets {
		if e.get(hashed) {
			for i0, e0 := u.modGet(uint(i) + uint(e.hashOS)); ; i0, e0 = u.modGet(i0 + uint(e0.linkOS)) {
				M.Put(e0.key, e0.val)
				if !e0.get(linked) {
					break
				}
			}
		}
	}
	u.buckets = M.buckets
}

func (u *HopMap[K, V]) Get(key K) (V, bool) {
	hash := u.hash(key)
	if i0, e0 := u.modGet(hash); e0.get(hashed) {
		for i1, e1 := u.modGet(i0 + uint(e0.hashOS)); ; i1, e1 = u.modGet(i1 + uint(e1.linkOS)) {
			if e1.key == key {
				return e1.val, true
			}
			if !e1.get(linked) {
				break
			}
		}
	}
	return *new(V), false
}

func (u *HopMap[K, V]) Put(key K, val V) {
begin:
	i_hash, e_hash := u.modGet(u.hash(key))
	if e_hash.get(hashed) { //there exists some elements with hash i_hash; check if key already exists.
		for nx := i_hash + uint(e_hash.hashOS); ; { //find i_hash+hashOS: start of the chain
			if i1, e1 := u.modGet(nx); e1.key == key {
				e1.val = val
				return
			} else if e1.get(linked) {
				nx = i1 + uint(e1.linkOS) //find the next in the chain: i1+linkOS
			} else {
				break //since this key doesn't exist, continue to find open spot.
			}
		}
	}

	//now since i_hash is either free or belongs to some other hash, we need to find an open spot
	for step := uint(0); step < uint(len(u.buckets)); step++ {
		if i_empty, e_empty := u.modGet(step + i_hash); !e_empty.get(used) { //found an empty spot
			if step < uint(u.H) { //within H.
				e_empty.key, e_empty.val = key, val
				e_empty.set(used)
				if e_hash.get(hashed) { //something else already hashed to i_hash, chain it to linked list
					i0, e0 := u.modGet(i_hash + uint(e_hash.hashOS))
					for ; e0.get(linked); i0, e0 = u.modGet(i0 + uint(e0.linkOS)) {
						//find the end of the linked list
					}
					e0.linkOS = int8(i_empty - i0) //link e_empty after e0.
					e0.set(linked)
				} else { //nothing hashed to i_hash
					e_hash.hashOS = int8(step) //set e_empty to be hashed to e_hash
					e_hash.set(hashed)
				}
				return //if an empty spot within H is found, an insertion will always be made immediately.
			} else { //j+step>=H.
				for i := i_hash; i < i_empty; i++ { //iterate from i_t to i_f
					if i0, e0 := u.modGet(i); e0.get(hashed) { //there is some value hashed to i0(i)
						//find the start of the chain and iterate in the chain.
						for i1, e1 := u.modGet(i0 + uint(e0.hashOS)); ; i1, e1 = u.modGet(i1 + uint(e1.linkOS)) {
							if i1 < i_empty && i1 > i_empty-uint(u.H) { //a value e1 with hash i is located in [i_f-H,i_f); so we swap e1 with e_empty
								if t, _ := u.modGet(i0); t == i { //i1 is the beginning of the list, so i0 hashes to i1; i0 is the initial i0
									e0.hashOS = int8(i_empty - i0) //e0 now hashes to e_empty
								} else { //if e0 doesn't hash to e1, then e0 link to e1
									e0.linkOS = int8(i_empty - i0) //e0 now links to e_empty
								}
								e_empty.key, e_empty.val, e_empty.info = e1.key, e1.val, e1.info //copies e1 to e_empty
								e_empty.linkOS = int8(uint(e1.linkOS) + i1 - i_empty)            //e_empty links to the original next of e1
								e_empty.hashOS = int8(uint(e1.hashOS) + i1 - i_empty)
								//now e1 is copied to e_empty, and all references to e1 is now to e_empty, we can change i_empty to i1
								i_empty, e_empty = i1, e1
								e1.clear(used | linked) //e1 is now empty
								goto begin              //go back to retry insert or move back
							}
							if !e1.get(linked) { //reached the end without finding one.
								break
							}
							i0, e0 = i1, e1 //store the previous in the chain
						}
					}
				}
				u.expand()
				goto begin
			}
		}
	}
	u.expand() //if the loop ended without returning, that means no empty spot is found
	goto begin
}
