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
	Put(K, V)
	HasKey(K) bool
	Get(K) V
	GetOrPut(K, V) (V, bool)
	GetAndRmv(K) (V, bool)
	Remove(K)
	Take() (K, V)
	Pairs() func() (K, V, bool)
	Size() uint
}
