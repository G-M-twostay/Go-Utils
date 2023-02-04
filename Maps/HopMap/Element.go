package HopMap

import "fmt"

type bit byte

const (
	used bit = 1 << iota
	hashed
	linked
)

type Element[K any, V any] struct {
	key            K
	val            V
	hashOS, linkOS int8
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

func (e *Element[K, V]) String() string {
	return fmt.Sprintf("key: %v, val: %v, ho: %v, lo: %v, info: %b", e.key, e.val, e.hashOS, e.linkOS, e.info)
}

func (e *Element[K, V]) offsetHash(i uint) uint {
	return uint(e.hashOS) + i
}

func (e *Element[K, V]) offsetLink(i uint) uint {
	return uint(e.linkOS) + i
}
