package Go_Utils

import (
	"math/bits"
)

const LogUintSize int = 5 + (bits.UintSize >> 6) //2^5==32 or 2^6==64

func NewBitArray(size uint) BitArray {
	if size&(bits.UintSize-1) == 0 {
		return BitArray{bits: make([]uint, size>>LogUintSize)}
	}
	return BitArray{bits: make([]uint, 1+(size>>LogUintSize))}
}

type BitArray struct {
	bits []uint
}

func (u BitArray) Len() int {
	return len(u.bits) << LogUintSize
}

func (u BitArray) Get(i int) bool {
	//t := uint(1 << (i & (bits.UintSize - 1)))
	//or X&t==t
	return u.bits[i>>LogUintSize]>>(i&(bits.UintSize-1))&1 == 1
}

func (u BitArray) Set(i int) {
	u.bits[i>>LogUintSize] |= 1 << (i & (bits.UintSize - 1))
}

func (u BitArray) Clr(i int) {
	u.bits[i>>LogUintSize] &^= 1 << (i & (bits.UintSize - 1))
}

func (u BitArray) Invert(i int) bool {
	t := u.Get(i)
	u.bits[i>>LogUintSize] ^= 1 << (i & (bits.UintSize - 1))
	return t
}
