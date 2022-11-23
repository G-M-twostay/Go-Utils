package Maps

import "math"

const (
	MaxUintHash = math.MaxUint
	MinUintHash = 0
)

type Hashable interface {
	Hash() uint
	Equal(other Hashable) bool
}

type Map[K Hashable, V any] interface {
	Delete(K)
	Load(K) (V, bool)
	LoadAndDelete(K) (V, bool)
	LoadOrStore(K, V) (V, bool)
	Range(func(K, V) bool)
	Store(K, V)
}

type ExtendedMap[K Hashable, V any] interface {
	Map[K, V]
	HasKey(K) bool
	Size() uint
	Take() (K, V)
	LoadPtr(K) *V
}
