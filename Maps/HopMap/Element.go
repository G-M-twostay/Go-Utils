package HopMap

import (
	"math"
)

type bucket[K comparable, V any] struct {
	key          K
	val          V
	dHash, dLink byte //0 indicates not valid, otherwise value is offset by min value of signed version
}

func (e *bucket[K, V]) hashed() bool {
	return e.dHash != 0
}

func (e *bucket[K, V]) clrHash() {
	e.dHash = 0
}

func (e *bucket[K, V]) linked() bool {
	return e.dLink != 0
}

func (e *bucket[K, V]) clrLink() {
	e.dLink = 0
}

func (e *bucket[K, V]) deltaLink() int {
	return int(e.dLink) + math.MinInt8
}

func (e *bucket[K, V]) deltaHash() int {
	return int(e.dHash) + math.MinInt8
}

func (e *bucket[K, V]) useDeltaHash(d int) {
	e.dHash = offset(d)
}

func (e *bucket[K, V]) useDeltaLink(d int) {
	e.dLink = offset(d)
}

func offset(x int) byte {
	return byte(x - math.MinInt8)
}

const step = 8

type buffer[K comparable, V any] []struct {
	key  K
	val  V
	hash uint
}

func (u buffer[K, V]) full() bool {
	return len(u) == cap(u)
}

func (u buffer[K, V]) empty() bool {
	return len(u) == 0
}

func (u *buffer[K, V]) put(k *K, v *V, hash uint) bool {
	for i, c := range *u {
		if c.hash == hash && c.key == *k {
			(*u)[i].val = *v
			return false
		}
	} //didn't find this value, append

	//*u = append(*u, struct {
	//	key  K
	//	val  V
	//	hash uint
	//}{*k, *v, hash})

	l := len(*u)
	if u.full() {
		n := make(buffer[K, V], l+step)
		copy(n, *u)
		*u = n
	}
	*u = (*u)[:l+1]
	(*u)[l].key, (*u)[l].val, (*u)[l].hash = *k, *v, hash
	return true
}

func (u buffer[K, V]) set(k *K, v *V, hash uint) bool {
	for i, c := range u {
		if c.hash == hash && c.key == *k {
			u[i].val = *v
			return true
		}
	}
	return false
}

func (u *buffer[K, V]) pop(k *K, hash uint) *V {
	for i, c := range *u {
		if c.hash == hash && c.key == *k {
			r := &c.val               //get the desired value
			(*u)[i] = (*u)[len(*u)-1] //copy the last value to i
			*u = (*u)[:len(*u)-1]     //truncate the last value
			return r
		}
	}
	return nil
}

func (u buffer[K, V]) get(k *K, hash uint) (V, bool) {
	for _, c := range u {
		if c.hash == hash && c.key == *k {
			return c.val, true
		}
	}
	return *new(V), false
}
