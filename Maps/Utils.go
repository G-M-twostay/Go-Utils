package Maps

import "sync"

//These are all internal helper structs/functions, these will eventually all be sealed.

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

func Mark(hash uint) uint {
	return hash | ^MaxArrayLen
}

func Mask(hash uint) uint {
	return hash & MaxArrayLen
}

type FlagLock struct {
	sync.RWMutex
	Del bool
}

func (l *FlagLock) SafeLock() bool {
	l.Lock()
	return !l.Del
}

func (l *FlagLock) SafeRLock() bool {
	l.RLock()
	return !l.Del
}
