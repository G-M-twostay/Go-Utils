package HopMap

import "math"

type Bucket[K comparable, V any] struct {
	key          K
	val          V
	dHash, dLink byte //0 indicates not valid, otherwise value is offset by min value of signed version
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
	return int(e.dLink) + math.MinInt8
}

func (e *Bucket[K, V]) deltaHash() int {
	return int(e.dHash) + math.MinInt8
}

func (e *Bucket[K, V]) useDeltaHash(d int) {
	e.dHash = offset(d)
}

func (e *Bucket[K, V]) useDeltaLink(d int) {
	e.dLink = offset(d)
}

func offset(x int) byte {
	return byte(x - math.MinInt8)
}

//func (e *Bucket[K, V]) String() string {
//	return fmt.Sprintf("key: %v, val: %v, ho: %v, lo: %v, info: %b", e.key, e.val, e.dHash, e.dLink, e.info)
//}
