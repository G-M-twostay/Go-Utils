package HopMap

import "math"

type Bucket[K comparable, V any] struct {
	key          K
	val          V
	dHash, dLink uint16 //lowest bit indicates if it's valid
}

func (e *Bucket[K, V]) hashed() bool {
	return e.dHash != 0
}

func (e *Bucket[K, V]) clrHash() {
	e.dHash = 0
}

func (e *Bucket[K, V]) linked() bool {
	return e.dLink != 0
}

func (e *Bucket[K, V]) clrLink() {
	e.dLink = 0
}

func (e *Bucket[K, V]) deltaLink() int {
	return int(e.dLink) + math.MinInt16
}

func (e *Bucket[K, V]) deltaHash() int {
	return int(e.dHash) + math.MinInt16
}

func (e *Bucket[K, V]) useDeltaHash(d int) {
	e.dHash = offset(d)
}

func (e *Bucket[K, V]) useDeltaLink(d int) {
	e.dLink = offset(d)
}

func offset(x int) uint16 {
	return uint16(x - math.MinInt16)
}

//func (e *Bucket[K, V]) String() string {
//	return fmt.Sprintf("key: %v, val: %v, ho: %v, lo: %v, info: %b", e.key, e.val, e.dHash, e.dLink, e.info)
//}
