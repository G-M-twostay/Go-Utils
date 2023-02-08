package HopMap2

import (
	"fmt"
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

func (u *HopMap[K, V]) expand() {
	M := HopMap[K, V]{bkt: make([]Element[K, V], (len(u.bkt)-int(u.H))*2+int(u.H)), H: u.H, seed: u.seed}
	for _, e := range u.bkt {
		if !e.isFree() {
			if !M.tryPut(&e.key, &e.val, e.Hash()) {
				M.expand()
				M.tryPut(&e.key, &e.val, e.Hash())
			}
		}
	}
	u.bkt = M.bkt
}

func (u *HopMap[K, V]) hash(key K) uint {
	l := int(unsafe.Sizeof(key))
	s := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  l,
		Cap:  l,
	}
	//return Maps.Mask(uint(maphash.Bytes(u.seed, *(*[]byte)(unsafe.Pointer(&s)))))
	//return int(key) & (len(u.bkt) - int(u.H) - 1)
	return uint(xxhash.Sum64(*(*[]byte)(unsafe.Pointer(&s))))
	//return uint(Maps.RTHash(unsafe.Pointer(&key), 0, unsafe.Sizeof(key)))
}

func (u *HopMap[K, V]) mod(x uint) int {
	return int(x) & (len(u.bkt) - int(u.H) - 1)
}

func (u *HopMap[K, V]) Put(key K, val V) {
	for !u.tryPut(&key, &val, u.hash(key)) {
		u.expand()
	}
}

func (u *HopMap[K, V]) Get(key K) (V, bool) {
	i_hash := u.mod(u.hash(key))
	for i := i_hash; i < i_hash+int(u.H); i++ {
		if !u.bkt[i].isFree() && u.bkt[i].key == key {
			return u.bkt[i].val, true
		}
	}
	return *new(V), false
}

func (u *HopMap[K, V]) tryPut(key *K, val *V, hash uint) bool {
	//check if this key already exists in its neighbor
	i_hash := u.mod(hash)
	for i := i_hash; i < i_hash+int(u.H); i++ {
		if !u.bkt[i].isFree() && u.bkt[i].key == *key {
			u.bkt[i].val = *val
			return true
		}
	}
	//no same key exists, try to insert by first finding a free spot using linear probe: i_free
	for i := i_hash; i < len(u.bkt); i++ {
		if u.bkt[i].isFree() { //found a potential free spot
			if i < i_hash+int(u.H) { //within neighbor of i_hash
				u.bkt[i].Use(hash)
				u.bkt[i].key, u.bkt[i].val = *key, *val
				return true
			} else { //outside of neighbor, we now perform swap
				i_free := i
				for i_act := Maps.Max(0, i_free-int(u.H)+1); i_act < i_free; i_act++ { //iterate in (i_free-H,i_free)
					if u.mod(u.bkt[i_act].Hash()) > i_free-int(u.H) { //we know that everything here is used and that the actual index>=desired index.
						//found such an element where i_free is near its desired index
						u.bkt[i_free] = u.bkt[i_act] //regardless i_act will be copied to i_free
						if i_act < i_hash+int(u.H) { //the actual index is within neighbor of i_hash
							//no need to swap, just set i_des to the new value
							u.bkt[i_act].key, u.bkt[i_act].val = *key, *val
							u.bkt[i_act].Use(hash)
							return true
						}
						//need more swaps
						u.bkt[i_act].free() //free i_des
						i_free = i_act
						i_act = Maps.Max(0, i_free-int(u.H)+1)
					}
				}
				//we're unable to move free spots near H
				return false
			}
		}
	}
	//if no potential free spot are found
	return false
}

func (u *HopMap[K, V]) String() string {
	var s string = "["
	for i, e := range u.bkt {
		s += fmt.Sprintf("{k: %v; v: %v; des: %d; act: %d; f: %t }", e.key, e.val, u.mod(e.Hash()), i, e.isFree())
	}
	return s + "]"
}
