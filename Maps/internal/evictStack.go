package internal

import "unsafe"

const (
	deqSize = 1 << 2
)

type EvictStack struct {
	vs         [deqSize]unsafe.Pointer
	head, tail byte
}

func (es *EvictStack) Push(v unsafe.Pointer) {
	es.vs[es.head&(deqSize-1)] = v
	es.head++
	if es.head-es.tail > deqSize {
		es.tail++
	}
}
func (es *EvictStack) Pop() unsafe.Pointer {
	if es.tail == es.head {
		return nil
	}
	es.head--
	a := es.vs[es.head&(deqSize-1)]
	return a
}
func (es *EvictStack) Empty() bool {
	return es.head == es.tail
}
