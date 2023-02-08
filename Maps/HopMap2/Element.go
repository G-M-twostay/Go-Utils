package HopMap2

import (
	"fmt"
	"github.com/g-m-twostay/go-utils/Maps"
)

type Element[K any, V any] struct {
	key  K
	val  V
	info uint
}

func (e *Element[K, V]) Hash() uint {
	return Maps.Mask(e.info)
}

func (e *Element[K, V]) Use(hash uint) {
	e.info = Maps.Mark(hash)
}

func (e *Element[K, V]) isFree() bool {
	return e.info <= Maps.MaxArrayLen
}

func (e *Element[K, V]) free() {
	e.info &= Maps.MaxArrayLen
}

func (e *Element[K, V]) swap(o *Element[K, V]) {
	t := *o
	*o = *e
	*e = t
}

func (e *Element[K, V]) String() string {
	return fmt.Sprintf("key: %v; val: %v; hash: %d; free: %t", e.key, e.val, e.Hash(), e.isFree())
}
