package Maps

import (
	"hash/maphash"
	"math"
	"unsafe"
)

const (
	MaxUintHash      = math.MaxUint
	MaxArrayLen uint = math.MaxInt
)

// Map is designed for compatibility with sync.Map. All the below functions have the exact same usage/behavior as documented in sync.Map.
type Map[K any, V any] interface {
	Delete(K)
	Load(K) (V, bool)
	LoadAndDelete(K) (V, bool)
	LoadOrStore(K, V) (V, bool)
	Range(func(K, V) bool)
	Store(K, V)
}

// ExtendedMap is the additional operation that my implementation support. Note that these operations aren't explicit implemented, meaning that they're merely taking advantage of the implementation. For example, values are internally stored as pointers in all implementations, so why not just provide a method to access the pointers directly?
type ExtendedMap[K any, V any] interface {
	Map[K, V]
	HasKey(K) bool
	Size() uint64
	//Take an arbitrary key value pair from the Map.
	Take() (K, V)
	TakePtr(K, *V)
	LoadPtr(K) *V
	LoadPtrAndDelete(K) (*V, bool)
	LoadPtrOrStore(K, V) (*V, bool)
	RangePtr(func(K, *V) bool)
	Set(K, V) *V
}

type Hasher maphash.Seed

func (u Hasher) HashAny(v any) uint {
	b := (*[unsafe.Sizeof(v)]byte)(unsafe.Pointer(&v))
	return uint(maphash.Bytes(maphash.Seed(u), (*b)[:]))
}

func (u Hasher) HashString(v string) uint {
	return uint(maphash.String(maphash.Seed(u), v))
}
