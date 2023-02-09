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
	t := HopMap[K, V]{make([]Element[K, V], dl+int(h)), h, maphash.MakeSeed()}
	for i := range t.bkt {
		t.bkt[i].init()
	}
	return &t
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

func (u *HopMap[K, V]) mod(hash uint) int {
	return int(hash) & (len(u.bkt) - int(u.H) - 1)
}

func (u *HopMap[K, V]) expand() {

	M := HopMap[K, V]{bkt: make([]Element[K, V], (len(u.bkt)-int(u.H))*2+int(u.H)), H: u.H, seed: u.seed}
	for i := range M.bkt {
		M.bkt[i].init()
	}
	for _, e := range u.bkt {
		if e.used {
			M.Put(e.key, e.val)
		}
	}
	u.bkt = M.bkt

}

func (u *HopMap[K, V]) Get(key K) (V, bool) {
	if i0 := u.hash(key); u.bkt[i0].hashed() {
		for i1 := i0 + int(u.bkt[i0].hashOS); ; i1 = i1 + int(u.bkt[i1].linkOS) {
			if u.bkt[i1].used && u.bkt[i1].key == key {
				return u.bkt[i1].val, true
			}
			if !u.bkt[i1].linked() {
				break
			}
		}
	}
	return *new(V), false
}

func (u *HopMap[K, V]) fillEmpty(i_hash int, i_free int, k *K, v *V) {
	u.bkt[i_free].key, u.bkt[i_free].val = *k, *v
	u.bkt[i_free].used = true
	if u.bkt[i_hash].hashed() { //something else already hashed to i_hash, chain it to linked list
		i0 := i_hash + int(u.bkt[i_hash].hashOS)
		for ; u.bkt[i0].linked(); i0 = i0 + int(u.bkt[i0].linkOS) {
			//find the end of the linked list
		}
		u.bkt[i0].linkOS = int16(i_free - i0) //link i_free after e0.
	} else { //nothing hashed to i_hash
		u.bkt[i_hash].hashOS = int16(i_free - i_hash) //fillEmpty i_free to be hashed to e_hash
	}
	//if an empty spot within H is found, an insertion will always be made immediately.
}

func (u *HopMap[K, V]) Put(key K, val V) {
	for !u.tryPut(&key, &val, 0) {
		u.expand()
	}
}

func (u *HopMap[K, V]) tryPut(k *K, v *V, hash uint) bool {
	i_hash := u.hash(*k)
	if u.bkt[i_hash].hashed() { //there exists some elements with hash i_hash; check if key already exists.
		for i0 := i_hash + int(u.bkt[i_hash].hashOS); ; i0 = i0 + int(u.bkt[i0].linkOS) { //find i_hash+hashOS: start of the chain
			if u.bkt[i0].key == *k {
				u.bkt[i0].val = *v
				return true
			}
			if !u.bkt[i0].linked() {
				break
			}
		}
	}
	//now since i_hash is either free or belongs to some other hash, we need to find an open spot
	for step := i_hash; step < len(u.bkt); step++ {
		if i_free := step; !u.bkt[i_free].used { //found an empty spot
			if i_free-i_hash < int(u.H) { //within H. we insert it here
				u.fillEmpty(i_hash, i_free, k, v)
				return true
			} else { //j+step>=H. so we find open spot and move it back
			search:
				for i := Maps.Max(i_free-int(u.H)+1, 0); i < i_free; i++ { //iterate from i_hash to i_empty
					if i0 := i; u.bkt[i0].hashed() { //there is some value hashed to i0(i). i0 refers to the prev in the linked iteration.
						//find the start of the chain and iterate in the chain.
						prev := &u.bkt[i0].hashOS //initially e0 pointed to e1 by hash
						for i1 := i0 + int(u.bkt[i0].hashOS); ; i1 = i1 + int(u.bkt[i1].linkOS) {
							if i_free-int(u.H) < i1 && i1 < i_free { //a value e1 with hash i is located in [i_empty-H,i_empty); so we swap e1 with i_free
								//make everything that pointed to e1 from e0 point to i_free
								*prev = int16(i_free - i0)

								u.bkt[i_free].key, u.bkt[i_free].val = u.bkt[i1].key, u.bkt[i1].val //copies e1 to i_free
								u.bkt[i_free].used = true
								if u.bkt[i1].linked() {
									u.bkt[i_free].linkOS = int16(int(u.bkt[i1].linkOS) + i1 - i_free) //i_free links to the original next of e1
								}
								//now e1 is copied to i_free, and all references to e1 is now to i_free, we can change i_empty to i1
								u.bkt[i1].clrLink() //e1 is now empty, but it may still hashes to something.

								if i1 < i_hash+int(u.H) {
									u.fillEmpty(i_hash, i1, k, v)
									return true
								} else {
									u.bkt[i1].used = false //set it to free only when we need more swaps
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
				return false //unable to move free buckets near i_hash
			}
		}
	}
	return false //no free buckets are found
}
