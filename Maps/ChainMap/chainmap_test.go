package ChainMap

import (
	"GMUtils/Maps"
	"sync"
	"testing"
)

const (
	FNV_PRIME_32        uint   = 16777619
	FNV_PRIME_64        uint64 = 1099511628211
	FNV_OFFSET_BASIS_32 uint   = 2166136261
	FNV_OFFSET_BASIS_64 uint64 = 14695981039346656037
	times                      = 1024
	mapSize                    = 2048
)

type O int

func (u O) Equal(o Maps.Hashable) bool {
	return u == o.(O)
}

func (u O) Hash() int {
	return int(u)
}

func TestChainMap_All(t *testing.T) {
	M := MakeChainMap[O, int](0, 100, 0, 63)
	wg := &sync.WaitGroup{}
	wg.Add(8)
	for j := 0; j < 8; j++ {
		go func(l, h int) {
			for i := l; i < h; i++ {
				M.Put(O(i), 1)
			}
			for i := l; i < h; i++ {
				if !M.HasKey(O(i)) {
					t.Errorf("not added: %v\n", O(i))
				}
			}
			for i := l; i < h; i++ {
				M.Remove(O(i))
			}
			for i := l; i < h; i++ {
				if M.HasKey(O(i)) {
					t.Errorf("not removed: %v\n", O(i))
				}
			}
			wg.Done()
		}(j*8, (j+1)*8)
	}
	wg.Wait()
	M.PrintAll()
}
