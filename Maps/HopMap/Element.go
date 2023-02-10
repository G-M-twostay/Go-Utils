package HopMap

type Bucket[K comparable, V any] struct {
	key          K
	val          V
	dHash, dLink int16 //lowest bit indicates if it's valid
}

func (e *Bucket[K, V]) hashed() bool {
	return e.dHash&1 == 1
}

func (e *Bucket[K, V]) clrHash() {
	e.dHash = 0
}

func (e *Bucket[K, V]) linked() bool {
	return e.dLink&1 == 1
}

func (e *Bucket[K, V]) clrLink() {
	e.dLink = 0
}

func (e *Bucket[K, V]) deltaLink() int {
	return int(e.dLink) >> 1
}

func (e *Bucket[K, V]) deltaHash() int {
	return int(e.dHash) >> 1
}

func (e *Bucket[K, V]) useDeltaHash(d int) {
	e.dHash = markLowBit16(d, 1)
}

func (e *Bucket[K, V]) useDeltaLink(d int) {
	e.dLink = markLowBit16(d, 1)
}

func markLowBit16(x, low int) int16 {
	return int16((x << 1) | low)
}

//func (e *Bucket[K, V]) String() string {
//	return fmt.Sprintf("key: %v, val: %v, ho: %v, lo: %v, info: %b", e.key, e.val, e.dHash, e.dLink, e.info)
//}
