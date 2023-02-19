package HashSet

import "math"

type bucket[E comparable] struct {
	element      E
	dHash, dLink byte //0 indicates not valid, otherwise value is offset by min value of signed version
}

func (e *bucket[K]) hashed() bool {
	return e.dHash != 0
}

func (e *bucket[K]) clrHash() {
	e.dHash = 0
}

func (e *bucket[K]) linked() bool {
	return e.dLink != 0
}

func (e *bucket[K]) clrLink() {
	e.dLink = 0
}

func (e *bucket[K]) deltaLink() int {
	return int(e.dLink) + math.MinInt8
}

func (e *bucket[K]) deltaHash() int {
	return int(e.dHash) + math.MinInt8
}

func (e *bucket[K]) useDeltaHash(d int) {
	e.dHash = offset(d)
}

func (e *bucket[K]) useDeltaLink(d int) {
	e.dLink = offset(d)
}

func offset(x int) byte {
	return byte(x - math.MinInt8)
}
