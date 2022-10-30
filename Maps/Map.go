package Maps

type Hashable interface {
	Hash() int64
	Equal(other Hashable) bool
}

type Map[K Hashable, V any] interface {
	Put(K, V) V
	HasKey(K) bool
	Get(K) V
	Remove(K) bool
	Take() (K, V)
	Pairs() func() (K, V, bool)
	Size() uint
}
