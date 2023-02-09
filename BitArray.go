package Go_Utils

import (
	"math/bits"
)

const LOG_UINT_SIZE int = 5 + (bits.UintSize >> 6) //2^5==32 or 2^6==64

func New(size int) BitArray {
	return BitArray{bits: make([]uint, 1+(size>>LOG_UINT_SIZE))}
}

type BitArray struct {
	bits []uint
}

func (u BitArray) Len() int {
	return len(u.bits) << LOG_UINT_SIZE
}

func (u BitArray) Get(i int) bool {
	t := uint(1 << (i & (bits.UintSize - 1)))
	return u.bits[i>>LOG_UINT_SIZE]&t == t
}

func (u BitArray) Up(i int) {
	u.bits[i>>LOG_UINT_SIZE] |= 1 << (i & (bits.UintSize - 1))
}

func (u BitArray) Down(i int) {
	u.bits[i>>LOG_UINT_SIZE] &^= 1 << (i & (bits.UintSize - 1))
}
