package SpinMap

import (
	"GMUtils/Maps"
	"GMUtils/Maps/ChainMap"
	"sync"
	"testing"
)

const (
	blockSize   = 64
	blockNum    = 64
	iter0       = 1 << 4
	elementNum0 = 1 << 13
)

type O int

func (u O) Equal(o Maps.Hashable) bool {
	return u == o.(O)
}

func (u O) Hash() uint {
	return uint(u)
}

func TestChainMap_All(t *testing.T) {
	M := MakeSpinMap[O, int](2, 4, blockNum*blockSize-1)
	wg := &sync.WaitGroup{}
	wg.Add(blockNum)
	for j := 0; j < blockNum; j++ {
		go func(l, h int) {
			defer wg.Done()
			for i := l; i < h; i++ {
				M.Store(O(i), i)
			}

			for i := l; i < h; i++ {
				if !M.HasKey(O(i)) {
					t.Errorf("not put: %v\n", O(i))
					//return
				}
			}
			for i := l; i < h; i++ {
				M.Delete(O(i))

			}
			for i := l; i < h; i++ {
				if M.HasKey(O(i)) {
					t.Errorf("not removed: %v\n", O(i))
				}
			}

		}(j*blockSize, (j+1)*blockSize)
	}
	wg.Wait()
	for cur := (*M.buckets.Load())[0]; cur != nil; cur = cur.nx {
		t.Log(cur.String(), "\n")
	}
	//for i := 0; i < 8; i++ {
	//	M.Store(O(i), i+1)
	//}
	//for i := 0; i < 8; i++ {
	//	t.Log(i, M.Load(O(i)))
	//}
	//for i := 0; i < 8; i++ {
	//	M.Delete(O(i))
	//}
	//for i := 0; i < 8; i++ {
	//	t.Log(i, M.Load(O(i)))
	//}
	//for cur := M.buckets[0]; cur != nil; cur = (*state[O])(cur.s).nx {
	//	t.Log(cur.String(), "\n")
	//}
	//t.Log(M.HasKey(O(0)))
	//M.Store(O(0), 1)
	//M.Store(O(1), 2)
	//M.Store(O(2), 3)
	//M.Delete(O(0))
	//M.Delete(O(1))
	//t.Log("removed 0 and 1")
	//M.Delete(O(2))
	//t.Log("removed 0 and 1 and 2")
	//for cur := M.bucketsPtr[0]; cur != nil; cur = (*state[O])(cur.s).nx {
	//	t.Log(cur.String(), "\n")
	//}
	//M.Load(O(0))
}

func BenchmarkSpinMap_Case1(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		M := MakeSpinMap[O, int](0, 2, elementNum0*iter0-1)
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					M.Store(O(j), j)
				}
				for j := l; j < h; j++ {
					if !M.HasKey(O(j)) {
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
func BenchmarkChainMap_Case1(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		M := ChainMap.MakeChainMap[O, int](0, 2, elementNum0*iter0-1)
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					M.Store(O(j), j)
				}
				for j := l; j < h; j++ {
					if !M.HasKey(O(j)) {
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
func BenchmarkSpinMap_Case2(b *testing.B) {
	//runtime.GC()
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		M := MakeSpinMap[O, int](0, 2, elementNum0*iter0-1)
		for j := 0; j < elementNum0*iter0; j++ {
			M.Store(O(j), j)
		}
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					x, _ := M.Load(O(j))
					if x != j {
						b.Error("incorrect value")
					}
				}
				for j := l; j < h; j++ {
					M.Store(O(j), j+1)
				}
				for j := l; j < h; j++ {
					x, _ := M.Load(O(j))
					if x != j+1 {
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
func BenchmarkChainMap_Case2(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		M := ChainMap.MakeChainMap[O, int](0, 2, elementNum0*iter0-1)
		for j := 0; j < elementNum0*iter0; j++ {
			M.Store(O(j), j)
		}
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					x, _ := M.Load(O(j))
					if x != j {
						b.Error("incorrect value")
					}
				}
				for j := l; j < h; j++ {
					M.Store(O(j), j+1)
				}
				for j := l; j < h; j++ {
					x, _ := M.Load(O(j))
					if x != j+1 {
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
func BenchmarkSpinMap_Case3(b *testing.B) {
	//runtime.GC()
	b.StopTimer()
	wg := &sync.WaitGroup{}
	for a := 0; a < 10; a++ {
		M := MakeSpinMap[O, int](2, 8, iter0*elementNum0-1)
		b.StartTimer()
		for j := 0; j < iter0; j++ {
			wg.Add(1)
			go func(l, h int) {
				defer wg.Done()
				for i := l; i < h; i++ {
					M.Store(O(i), i)
				}

				for i := l; i < h; i++ {
					if !M.HasKey(O(i)) {
						b.Errorf("not put: %v\n", O(i))
					}
				}
				for i := l; i < h; i++ {
					M.Delete(O(i))

				}
				for i := l; i < h; i++ {
					if M.HasKey(O(i)) {
						b.Errorf("not removed: %v\n", O(i))
					}
				}

			}(j*elementNum0, (j+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}

}
func BenchmarkChainMap_Case3(b *testing.B) {
	b.StopTimer()
	wg := &sync.WaitGroup{}
	for a := 0; a < 10; a++ {
		M := ChainMap.MakeChainMap[O, int](2, 8, iter0*elementNum0-1)
		b.StartTimer()
		for j := 0; j < iter0; j++ {
			wg.Add(1)
			go func(l, h int) {
				defer wg.Done()
				for i := l; i < h; i++ {
					M.Store(O(i), i)
				}

				for i := l; i < h; i++ {
					if !M.HasKey(O(i)) {
						b.Errorf("not put: %v\n", O(i))
					}
				}
				for i := l; i < h; i++ {
					M.Delete(O(i))

				}
				for i := l; i < h; i++ {
					if M.HasKey(O(i)) {
						b.Errorf("not removed: %v\n", O(i))
					}
				}

			}(j*elementNum0, (j+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}

}
