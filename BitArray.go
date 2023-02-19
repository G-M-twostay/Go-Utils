package Go_Utils

import (
	"math/bits"
)

const LogUintSize int = 5 + (bits.UintSize >> 6) //2^5==32 or 2^6==64

func NewBitArray(size uint) BitArray {
	if size&(bits.UintSize-1) == 0 {
		return make([]uint, size>>LogUintSize)
	}
	return make([]uint, 1+(size>>LogUintSize))
}

type BitArray []uint

func (u BitArray) Len() int {
	return len(u) << LogUintSize
}

func (u BitArray) Get(i int) bool {
	//t := uint(1 << (i & (bits.UintSize - 1)))
	//or X&t==t
	return u[i>>LogUintSize]>>(i&(bits.UintSize-1))&1 == 1
}

func (u BitArray) Set(i int) {
	u[i>>LogUintSize] |= 1 << (i & (bits.UintSize - 1))
}

func (u BitArray) Clr(i int) {
	u[i>>LogUintSize] &^= 1 << (i & (bits.UintSize - 1))
}

func (u BitArray) Invert(i int) bool {
	t := u.Get(i)
	u[i>>LogUintSize] ^= 1 << (i & (bits.UintSize - 1))
	return t
}

func (u BitArray) First() int {
	for i, c := range u {
		if c > 0 {
			return i<<LogUintSize + bits.LeadingZeros(c)
		}
	}
	return -1
}

func (u BitArray) Empty() bool {
	for _, c := range u {
		if c != 0 {
			return false
		}
	}
	return true
}

func (u BitArray) Append(unitCount uint) BitArray {
	newArr := make([]uint, uint(len(u))+unitCount)
	copy(newArr, u)
	return newArr
}
