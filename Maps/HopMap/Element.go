package HopMap

const (
	used, _ byte = 1 << iota, iota
	hashed, hashedIndex
	linked, linkedIndex
	_, topIndex
)

type OffsetType int8
type extra struct {
	dHash, dLink OffsetType
	info         byte
}

func (e extra) getRaw(pos byte) byte {
	return e.info & pos
}

func (e extra) get(pos byte) bool {
	return e.info&pos == pos
}

func (e *extra) set(pos byte) {
	e.info |= pos
}

func (e *extra) clr(pos byte) {
	e.info &^= pos
}

func (e extra) count() byte {
	return e.info >> topIndex
}

func (e *extra) incCount() {
	e.info += 1 << topIndex
}

func (e *extra) decCount() {
	e.info -= 1 << topIndex
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
		n := make(buffer[K, V], l+step) //manually manage the allocation of more memory
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
