package HopMap

import "math/bits"

func newOmap[K comparable, V any](size uint) *omap[K, V] {
	x := bits.Len(size) + 1
	return &omap[K, V]{logLen: byte(bits.UintSize - x), bkt: make([]buffer[K, V], 1<<x)}
}

type omap[K comparable, V any] struct {
	bkt    []buffer[K, V]
	size   uint
	logLen byte
}

func (u *omap[K, V]) avgLen() float32 {
	return float32(u.size) / float32(len(u.bkt))
}

// nil friendly
func (u *omap[K, V]) bkts_() []buffer[K, V] {
	if u == nil {
		return nil
	}
	return u.bkt
}

func (u *omap[K, V]) mod(hash uint) int {
	return int(hash >> u.logLen)
	//return int(hash) & (len(u.bkt) - 1)
}
func (u *omap[K, V]) set(key *K, val *V, hash uint) bool {
	return u.bkt[u.mod(hash)].set(key, val, hash)
}
func (u *omap[K, V]) get(key *K, hash uint) (V, bool) {
	return u.bkt[u.mod(hash)].get(key, hash)
}
func (u *omap[K, V]) put(key *K, val *V, hash uint) (added bool) {
	if added = u.bkt[u.mod(hash)].put(key, val, hash); added {
		u.size++
	}
	return
}

func (u *omap[K, V]) pop(key *K, hash uint) (val *V) {
	i_hash := u.mod(hash)
	if val = u.bkt[i_hash].pop(key, hash); val != nil {
		u.size--
		if u.bkt[i_hash].empty() { //a buffer is empty, free it
			u.bkt[i_hash] = nil
		}
	}
	return
}

// nil friendly
func (u *omap[K, V]) Size_() uint {
	if u == nil {
		return 0
	}
	return u.size
}
