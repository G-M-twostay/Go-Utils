package HopMap

type bit byte

const (
	used bit = 1 << iota
	hashed
	linked
	overflow
)

type Element[K any, V any] struct {
	key            K
	val            V
	hashOS, linkOS int16
	info           bit
}

func (e *Element[K, V]) get(pos bit) bool {
	return e.info&pos == pos
}

func (e *Element[K, V]) set(pos bit) {
	e.info |= pos
}

func (e *Element[K, V]) clear(pos bit) {
	e.info &^= pos
}

//func (e *Element[K, V]) String() string {
//	return fmt.Sprintf("key: %v, val: %v, ho: %v, lo: %v, info: %b", e.key, e.val, e.hashOS, e.linkOS, e.info)
//}
