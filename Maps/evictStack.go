package Maps

import "unsafe"

const (
	deqSize = 1 << 2
)

// evictStack is a stack that drops the oldest elements when its pushed full. It's used to record the path traversed in the linked list.
type evictStack struct {
	vs         [deqSize]unsafe.Pointer
	head, tail byte
}

func (es *evictStack) Push(v unsafe.Pointer) {
	es.vs[es.head&(deqSize-1)] = v
	es.head++
	if es.head-es.tail > deqSize {
		es.tail++
	}
}
func (es *evictStack) Pop() unsafe.Pointer {
	if es.tail == es.head {
		return nil
	}
	es.head--
	a := es.vs[es.head&(deqSize-1)]
	return a
}
func (es *evictStack) Empty() bool {
	return es.head == es.tail
}
