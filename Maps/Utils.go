package Maps

type HashList[V any] struct {
	Array []V
	Chunk byte //hash range of the first segment is [0,2^chunk)
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

func Mark(hash uint) uint {
	return hash | ^MaxArrayLen
}

func Mask(hash uint) uint {
	return hash & MaxArrayLen
}
