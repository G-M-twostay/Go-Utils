package Go_Utils

import (
	"math/bits"
)

func New(size int) BitArray {
	return BitArray{bits: make([]uint, size/bits.UintSize)}
}

type BitArray struct {
	bits []uint
}

func (u BitArray) Len() int {
	return len(u.bits) * bits.UintSize
}

func (u BitArray) Get(i int) bool {
	return (u.bits[i/bits.UintSize]>>(i%bits.UintSize))&1 == 1
}

func (u BitArray) Up(i int) {
	u.bits[i/bits.UintSize] |= 1 << (i % bits.UintSize)
}

func (u BitArray) Down(i int) {
	u.bits[i/bits.UintSize] &^= 1 << (i % bits.UintSize)
}
