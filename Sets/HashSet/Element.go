package HashSet

import "math"

type Bucket[E comparable] struct {
	element      E
	dHash, dLink byte //0 indicates not valid, otherwise value is offset by min value of signed version
}

func (e *Bucket[K]) hashed() bool {
	return e.dHash != 0
}

func (e *Bucket[K]) clrHash() {
	e.dHash = 0
}

func (e *Bucket[K]) linked() bool {
	return e.dLink != 0
}

func (e *Bucket[K]) clrLink() {
	e.dLink = 0
}

func (e *Bucket[K]) deltaLink() int {
	return int(e.dLink) + math.MinInt8
}

func (e *Bucket[K]) deltaHash() int {
	return int(e.dHash) + math.MinInt8
}

func (e *Bucket[K]) useDeltaHash(d int) {
	e.dHash = offset(d)
}

func (e *Bucket[K]) useDeltaLink(d int) {
	e.dLink = offset(d)
}

func offset(x int) byte {
	return byte(x - math.MinInt8)
}
