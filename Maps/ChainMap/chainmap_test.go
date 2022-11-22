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
	blockSize                  = 64
	blockNum                   = 64
	iter0                      = 1 << 3
	iter1                      = 3
	elementNum0                = 1 << 10
)

type O int

func (u O) Equal(o Maps.Hashable) bool {
	return u == o.(O)
}

func (u O) Hash() uint {
	return uint(u)
}

func TestChainMap_All(t *testing.T) {
	M := MakeChainMap[O, int](2, 4, blockNum*blockSize-1)
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
	for cur := (*M.buckets.Load())[0]; cur != nil; cur = (*state[O])(cur.s).nx {
		t.Log(cur.String(), "\n")
	}
	//for i := 0; i < 8; i++ {
	//	M.Put(O(i), i+1)
	//}
	//for i := 0; i < 8; i++ {
	//	t.Log(i, M.Get(O(i)))
	//}
	//for i := 0; i < 8; i++ {
	//	M.Remove(O(i))
	//}
	//for i := 0; i < 8; i++ {
	//	t.Log(i, M.Get(O(i)))
	//}
	//for cur := M.buckets[0]; cur != nil; cur = (*state[O])(cur.s).nx {
	//	t.Log(cur.String(), "\n")
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
	//for cur := M.bucketsPtr[0]; cur != nil; cur = (*state[O])(cur.s).nx {
	//	t.Log(cur.String(), "\n")
	//}
	//M.Get(O(0))
}

func BenchmarkChainMap_Case1(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := MakeChainMap[O, int](0, 2, elementNum0*iter0-1)
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					M.Put(O(j), j)
				}
				for j := l; j < h; j++ {
					if !M.HasKey(O(j)) {
						b.Error("key doesn't exist")
					}
				}
				for j := l; j < h; j++ {
					if M.Get(O(j)) != j {
						b.Error("incorrect value")
					}
				}
				wg.Done()
			}(k*elementNum0, (k+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}
}

func BenchmarkSyncMap_Case1(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := sync.Map{}
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					M.Store(O(j), j)
				}
				for j := l; j < h; j++ {
					_, x := M.Load(O(j))
					if !x {
						b.Error("key doesn't exist")
					}
				}
				for j := l; j < h; j++ {
					x, _ := M.Load(O(j))
					if x != j {
						b.Error("incorrect value")
					}
				}
				wg.Done()
			}(k*elementNum0, (k+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}
}

func BenchmarkMutexMap_Case1(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	lc := sync.RWMutex{}
	for i := 0; i < b.N; i++ {
		M := make(map[O]int)
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					lc.Lock()
					M[O(j)] = j
					lc.Unlock()
				}
				for j := l; j < h; j++ {
					lc.RLock()
					_, x := M[O(j)]
					lc.RUnlock()
					if !x {
						b.Error("key doesn't exist")
					}
				}
				for j := l; j < h; j++ {
					lc.RLock()
					x, _ := M[O(j)]
					lc.RUnlock()
					if x != j {
						b.Error("incorrect value")
					}
				}
				wg.Done()
			}(k*elementNum0, (k+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}
}
