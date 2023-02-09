package HopMap

import "math"

type bit byte

const (
	NAN16 int16 = math.MinInt16
)

type Element[K any, V any] struct {
	key            K
	val            V
	hashOS, linkOS int16
	used           bool
}

func (e *Element[K, V]) hashed() bool {
	return e.hashOS != NAN16
}

func (e *Element[K, V]) clrHash() {
	e.hashOS = NAN16
}

func (e *Element[K, V]) linked() bool {
	return e.linkOS != NAN16
}

func (e *Element[K, V]) clrLink() {
	e.linkOS = NAN16
}

func (e *Element[K, V]) init() {
	e.linkOS, e.hashOS = NAN16, NAN16
}

//func (e *Element[K, V]) String() string {
//	return fmt.Sprintf("key: %v, val: %v, ho: %v, lo: %v, info: %b", e.key, e.val, e.hashOS, e.linkOS, e.info)
//}
