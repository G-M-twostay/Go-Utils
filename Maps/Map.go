package Maps

const (
	MaxUintHash = ^uint(0)
	MinUintHash = 0
	MaxIntHash  = int(MaxUintHash >> 1)
	MinIntHash  = -MaxIntHash - 1
)

type Hashable interface {
	Hash() int
	Equal(other Hashable) bool
}

type Map1[K Hashable, V any] interface {
	Put(K, V)
	HasKey(K) bool
	Get(K) V
	GetOrPut(K, V) (V, bool)
	GetAndRmv(V, bool)
	Remove(K)
	Take() (K, V)
	Pairs() func() (K, V, bool)
	Size() uint
}
