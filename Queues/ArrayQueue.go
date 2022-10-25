package Queues

type circArrQ[T any] struct {
	sz, head, tail uint
	content        []T
}

func MakeArrayQueue[T any](initCap uint) ArrayQueue[T] {
	return &circArrQ[T]{0, 0, 0, make([]T, initCap)}
}

func (this circArrQ[T]) Empty() bool {
	return this.sz == 0
}

func (this *circArrQ[T]) resize(newLen uint) {
	nc := make([]T, newLen)
	if this.head < this.tail {
		copy(nc, this.content[this.head:this.tail])
	} else {
		copy(nc, this.content[this.head:])
		copy(nc[uint(len(this.content))-this.head:], this.content[:this.tail])
		this.head, this.tail = 0, this.sz
	}
	this.content = nc
}

func (this *circArrQ[T]) Shrink() {
	this.resize(this.sz | 1)
}

func (this *circArrQ[T]) Clear() {
	this.tail, this.head, this.sz = 0, 0, 0

}

func (this circArrQ[T]) Size() uint {
	return this.sz
}

func (this *circArrQ[T]) Push(item T) {
	if this.sz == uint(len(this.content)) {
		this.resize(this.sz * 3 / 2)
	}
	this.content[this.tail] = item
	this.tail = (this.tail + 1) % uint(len(this.content))
	this.sz++
}

func (this *circArrQ[T]) Pop() (item T, e error) {
	if this.Empty() {
		return *new(T), &EmptyQueueError{}
	} else {
		t := this.content[this.head]
		this.content[this.head] = *new(T)
		this.head = (this.head + 1) % uint(len(this.content))
		this.sz--
		return t, nil
	}
}

func (this circArrQ[T]) Peek() (item T) {
	if this.Empty() {
		return *new(T)
	} else {
		return this.content[this.head]
	}
}
