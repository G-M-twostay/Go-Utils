package Maps

import "math"

const (
	MaxUintHash      = math.MaxUint
	MinUintHash      = 0
	MaxArrayLen uint = math.MaxInt
)

type Hashable interface {
	//Hash of the object
	Hash() uint
	//Equal with other
	Equal(other Hashable) bool
}

// Map is designed for compatibility with sync.Map. All the below functions have the exact same usage/behavior as documented in sync.Map.
type Map[K Hashable, V any] interface {
	Delete(K)
	Load(K) (V, bool)
	LoadAndDelete(K) (V, bool)
	LoadOrStore(K, V) (V, bool)
	Range(func(K, V) bool)
	Store(K, V)
}

// ExtendedMap is the additional operation that my implementation support. Note that these operations aren't explicit implemented, meaning that they're merely taking advantage of the implementation. For example, values are internally stored as pointers in all implementations, so why not just provide a method to access the pointers directly?
type ExtendedMap[K Hashable, V any] interface {
	Map[K, V]
	HasKey(K) bool
	Size() uint
	//Take an arbitrary key value pair from the Map.
	Take() (K, V)
	TakePtr(K, *V)
	LoadPtr(K) *V
	LoadPtrAndDelete(K) (*V, bool)
	LoadPtrOrStore(K, V) (*V, bool)
	RangePtr(func(K, *V) bool)
}
