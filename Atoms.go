package Go_Utils

import "sync/atomic"

// AtomicUint backed by uintptr
type AtomicUint struct {
	v uintptr
}

func (u *AtomicUint) Load() uint {
	return uint(atomic.LoadUintptr(&u.v))
}
func (u *AtomicUint) Store(v uint) {
	atomic.StoreUintptr(&u.v, uintptr(v))
}
func (u *AtomicUint) Add(d uint) uint {
	return uint(atomic.AddUintptr(&u.v, uintptr(d)))
}
func (u *AtomicUint) Swap(v uint) uint {
	return uint(atomic.SwapUintptr(&u.v, uintptr(v)))
}
func (u *AtomicUint) CompareAndSwap(exp, v uint) bool {
	return atomic.CompareAndSwapUintptr(&u.v, uintptr(exp), uintptr(v))
}

// AtomicInt backed by uintptr
type AtomicInt struct {
	v uintptr
}

func (u *AtomicInt) Load() int {
	return int(atomic.LoadUintptr(&u.v))
}
func (u *AtomicInt) Store(v int) {
	atomic.StoreUintptr(&u.v, uintptr(v))
}
func (u *AtomicInt) Add(d int) int {
	return int(atomic.AddUintptr(&u.v, uintptr(d)))
}
func (u *AtomicInt) Swap(v int) int {
	return int(atomic.SwapUintptr(&u.v, uintptr(v)))
}
func (u *AtomicInt) CompareAndSwap(exp, v int) bool {
	return atomic.CompareAndSwapUintptr(&u.v, uintptr(exp), uintptr(v))
}
