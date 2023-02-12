package Maps

import (
	"math/bits"
	_ "runtime"
	"unsafe"
)
import _ "hash/maphash"

//These are all internal helper structs/functions, these will eventually all be sealed.

// HashList is a array with length 2^n
type HashList[V any] struct {
	Array []V
	Chunk byte //HashAny range of the first segment is [0,2^chunk)
}

func (u HashList[V]) Get(hash uint) V {
	return u.Array[hash>>u.Chunk]
}

func (u HashList[V]) Index(hash uint) uint {
	return hash >> u.Chunk
}

func (u HashList[V]) Intv() uint {
	return 1 << u.Chunk
}

// Mark the first bit with 1.
func Mark(hash uint) uint {
	return hash | ^MaxArrayLen
}

// Mask hash to ignore the first bit.
func Mask(hash uint) uint {
	return hash & MaxArrayLen
}

type hold struct {
	rtype *uintptr
	ptr   unsafe.Pointer
}

func Max(x, y int) int {
	d := x - y
	return x - (d & (d >> (bits.UintSize - 1)))
}

//go:linkname RTHash runtime.memhash
//go:noescape
func RTHash(ptr unsafe.Pointer, seed uint, len uintptr) uint

//go:linkname RTHash64 runtime.memhash64
//go:noescape
func RTHash64(ptr unsafe.Pointer, seed uint) uint

//go:linkname RTHash32 runtime.memhash32
//go:noescape
func RTHash32(ptr unsafe.Pointer, seed uint) uint

//go:linkname RTStrHash runtime.strhash
//go:noescape
func RTStrHash(ptr unsafe.Pointer, seed uint) uint
