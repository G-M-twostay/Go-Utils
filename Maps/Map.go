package Maps

type Hashable interface {
	Hash() int64
	Equal(other Hashable) bool
}

type Map[K Hashable, V any] interface {
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
