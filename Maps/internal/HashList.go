package internal

import (
	"math"
	"unsafe"
)

const (
	MaxArrayLen uint    = math.MaxInt
	ptrSz       uintptr = unsafe.Sizeof(uintptr(0))
)

//These are all internal helper structs/functions, these will eventually all be sealed.

// HashList is a array with length 2^n
type HashList[V any] struct {
	First *V   //pointer to first element of array
	Chunk byte //HashAny range of the first segment is [0,2^chunk)
}

func (u HashList[V]) Get(hash uint) V {
	//return u.Array[hash>>u.Chunk]
	return u.Fetch(u.Index(hash))
}
func (u HashList[V]) Fetch(i uint) V {
	return *(*V)(unsafe.Add(unsafe.Pointer(u.First), uintptr(i)*ptrSz))
}
func (u HashList[V]) Index(hash uint) uint {
	return hash >> u.Chunk
}

func (u HashList[V]) Intv() uint {
	return 1 << u.Chunk
}
func (u HashList[V]) Len() uint {
	return MaxArrayLen >> u.Chunk
}

// Mark the first bit with 1.
func Mark(hash uint) uint {
	return hash | ^MaxArrayLen
}

// Mask hash to ignore the first bit.
func Mask(hash uint) uint {
	return hash & MaxArrayLen
}
