package IntMap

import (
	"github.com/g-m-twostay/go-utils"
	"github.com/g-m-twostay/go-utils/Maps"
	"github.com/g-m-twostay/go-utils/Maps/BucketMap"
	"math"
	"sync"
	"testing"
)

const (
	blockSize   = 64
	blockNum    = 64
	iter0       = 1 << 3
	elementNum0 = 1 << 10
)

func hasher(x int) uint {
	return uint(x)
}

func cmp(x, y int) bool {
	return x == y
}

var (
	_ Maps.Map[int, int]         = new(IntMap[int, int])
	_ Maps.ExtendedMap[int, int] = new(IntMap[int, int])
	_ Maps.PtrMap[int, int]      = new(IntMap[int, int])
)

func TestIntMap_Basic(t *testing.T) {
	h := Go_Utils.Hasher(0)
	M := New[int, int](1, 6, math.MaxUint, h.HashInt)
	wg := &sync.WaitGroup{}
	wg.Add(blockNum)
	for j := 0; j < blockNum; j++ {
		go func(l, h int) {
			defer wg.Done()
			for i := l; i < h; i++ {
				M.Store(i, i)
			}

			for i := l; i < h; i++ {
				if !M.HasKey(i) {
					t.Errorf("not put: %v\n", i)
					return
				}
			}
			for i := l; i < h; i++ {
				M.Delete(i)
			}
			for i := l; i < h; i++ {
				if M.HasKey(i) {
					t.Errorf("not removed: %v\n", i)
					return
				}
			}

		}(j*blockSize, (j+1)*blockSize)
	}
	wg.Wait()
	//ct := 0
	//for cur := (*node)(M.buckets.Load().Load(0).Next()); cur != nil; cur = (*node)(cur.Next()) {
	//	if cur.isRelay() {
	//		t.Log(ct)
	//		ct = 0
	//	} else {
	//		ct++
	//	}
	//}
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
	//t.Log("added")
	//M.Delete(O(0))
	//t.Log("delted 0")
	//M.Delete(O(1))
	//t.Log("removed 0 and 1")
	//M.Delete(O(2))
	//t.Log("removed 0 and 1 and 2")
	//for cur := (*M.buckets.Load())[0]; cur != nil; cur = (*value[O])(cur.nx) {
	//	t.Log(cur.String(), "\n")
	//}
	//M.Load(O(0))
}
func TestIntMap_Take(t *testing.T) {
	m := New[int, int](0, 8, elementNum0*iter0-1, hasher)
	{
		_, b := m.TakePtr()
		if b != nil {
			t.Fail()
		}
	}
	{
		for i := range 100 {
			m.Store(i, i)
		}
		a, b := m.TakePtr()
		if b == nil {
			t.Fail()
		}
		if a != 0 || *b != 0 {
			t.Fail()
		}
	}
	{
		for i := range 100 {
			m.Delete(i)
		}
		_, b := m.TakePtr()
		if b != nil {
			t.Fail()
		}
	}

}
func BenchmarkBucketMap_Case1(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := BucketMap.New[int, int](0, 2, elementNum0*iter0-1, hasher, cmp, cmp)
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					M.Store(j, j)
				}
				for j := l; j < h; j++ {
					if !M.HasKey(j) {
						b.Error("getKey doesn't exist")
					}
				}
				for j := l; j < h; j++ {
					x, _ := M.Load(j)
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
func BenchmarkIntMap_Case1(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := New[int, int](0, 2, elementNum0*iter0-1, hasher)
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					M.Store(j, j)
				}
				for j := l; j < h; j++ {
					if !M.HasKey(j) {
						b.Error("getKey doesn't exist")
					}
				}
				for j := l; j < h; j++ {
					x, _ := M.Load(j)
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

func BenchmarkBucketMap_Case2(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := BucketMap.New[int, int](0, 2, elementNum0*iter0-1, hasher, cmp, cmp)
		for j := 0; j < elementNum0*iter0; j++ {
			M.Store(j, j)
		}
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					x, _ := M.Load(j)
					if x != j {
						b.Error("incorrect value")
					}
				}
				for j := l; j < h; j++ {
					M.Store(j, j+1)
				}
				for j := l; j < h; j++ {
					x, _ := M.Load(j)
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
func BenchmarkIntMap_Case2(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := New[int, int](0, 2, elementNum0*iter0-1, hasher)
		for j := 0; j < elementNum0*iter0; j++ {
			M.Store(j, j)
		}
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					x, _ := M.Load(j)
					if x != j {
						b.Error("incorrect value")
					}
				}
				for j := l; j < h; j++ {
					M.Store(j, j+1)
				}
				for j := l; j < h; j++ {
					x, _ := M.Load(j)
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

func BenchmarkBucketMap_Case3(b *testing.B) {
	//runtime.GC()
	b.StopTimer()
	wg := &sync.WaitGroup{}
	for a := 0; a < b.N; a++ {
		M := BucketMap.New[int, int](2, 8, iter0*elementNum0-1, hasher, cmp, cmp)
		b.StartTimer()
		for j := 0; j < iter0; j++ {
			wg.Add(1)
			go func(l, h int) {
				defer wg.Done()
				for i := l; i < h; i++ {
					M.Store(i, i)
				}

				for i := l; i < h; i++ {
					if !M.HasKey(i) {
						b.Errorf("not put: %v\n", i)
					}
				}
				for i := l; i < h; i++ {
					M.Delete(i)

				}
				for i := l; i < h; i++ {
					if M.HasKey(i) {
						b.Errorf("not removed: %v\n", i)
					}
				}

			}(j*elementNum0, (j+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}

}
func BenchmarkIntMap_Case3(b *testing.B) {
	//runtime.GC()
	b.StopTimer()
	wg := &sync.WaitGroup{}
	for a := 0; a < b.N; a++ {
		M := New[int, int](2, 8, iter0*elementNum0-1, hasher)
		b.StartTimer()
		for j := 0; j < iter0; j++ {
			wg.Add(1)
			go func(l, h int) {
				defer wg.Done()
				for i := l; i < h; i++ {
					M.Store(i, i)
				}

				for i := l; i < h; i++ {
					if !M.HasKey(i) {
						b.Errorf("not put: %v\n", i)
					}
				}
				for i := l; i < h; i++ {
					M.Delete(i)

				}
				for i := l; i < h; i++ {
					if M.HasKey(i) {
						b.Errorf("not removed: %v\n", i)
					}
				}

			}(j*elementNum0, (j+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}

}
