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
	blockSize                  = 256
	blockNum                   = 256
)

type O int

func (u O) Equal(o Maps.Hashable) bool {
	return u == o.(O)
}

func (u O) Hash() int {
	return int(u)
}

func TestChainMap_All(t *testing.T) {
	M := MakeChainMap[O, int](2, 2, 0, blockSize*blockNum-1)
	wg := &sync.WaitGroup{}
	wg.Add(blockNum)
	for j := 0; j < blockNum; j++ {
		go func(l, h int) {
			defer wg.Done()
			for i := l; i < h; i++ {
				M.Put(O(i), i)
			}

			for i := l; i < h; i++ {
				if !M.HasKey(O(i)) {
					t.Errorf("not put: %v\n", O(i))
					//return
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

		}(j*blockSize, (j+1)*blockSize)
	}
	wg.Wait()
	for cur := M.buckets[0]; cur != nil; cur = (*state[O])(cur.s).nx {
		t.Log(cur.String(), "\n")
	}
	//for i := 0; i < 10; i++ {
	//	M.Put(O(i), i+1)
	//}
	//for i := 0; i < 10; i++ {
	//	t.Log(i, M.Get(O(i)))
	//}
	//for i := 0; i < 10; i++ {
	//	M.Remove(O(i))
	//}
	//for i := 0; i < 10; i++ {
	//	t.Log(i, M.Get(O(i)))
	//}
	//t.Log(M.HasKey(O(0)))
	//M.Put(O(0), 1)
	//M.Put(O(1), 2)
	//M.Put(O(2), 3)
	//M.Remove(O(0))
	//M.Remove(O(1))
	//t.Log("removed 0 and 1")
	//M.Remove(O(2))
	//t.Log("removed 0 and 1 and 2")
	//for cur := M.buckets[0]; cur != nil; cur = (*state[O])(cur.s).nx {
	//	t.Log(cur.String(), "\n")
	//}
	//M.Get(O(0))
}
