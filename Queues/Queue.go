package Queues

type Queue[T any] interface {
	Push(item T)
	Pop() (T, error)
	Peek() T
	Empty() bool
}

type ArrayQueue[T any] interface {
	Queue[T]
	Shrink()
	Clear()
	Size() uint
	resize(newLen uint)
}

type BlockingQueue[T any] interface {
	Queue[T]
	WaitAndPop() T
}

type EmptyQueueError struct {
}

func (e *EmptyQueueError) Error() string {
	return "Queue is Empty: cannot Pop."
}

type UnexpectedError struct {
	msg string
}

func (e *UnexpectedError) Error() string {
	return e.msg
}
