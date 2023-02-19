package comparisons

import (
	"github.com/alphadose/haxmap"
	"github.com/cornelk/hashmap"
	"github.com/g-m-twostay/go-utils/Maps/IntMap"
	"sync"
	"testing"
)

// https://github.com/cornelk/hashmap/ is incorrect in all 3 test cases, see https://github.com/cornelk/hashmap/issues/73.
// https://github.com/alphadose/haxmap/ suffers the same fate, see https://github.com/alphadose/haxmap/issues/32.
const (
	blockSize   = 64
	blockNum    = 64
	iter0       = 1 << 3
	elementNum0 = 1 << 10
)

func hashInt(x int) uint {
	return uint(x)
}

func Benchmark2IntMap_Case1(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := IntMap.New[int, int](0, 2, elementNum0*iter0-1, hashInt)
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					M.Store(j, j)
				}
				for j := l; j < h; j++ {
					if !M.HasKey(j) {
						b.Error("key doesn't exist")
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

func Benchmark2HashMap_Case1(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := hashmap.New[int, int]()
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					M.Insert(j, j)
				}
				for j := l; j < h; j++ {
					_, a := M.Get(j)
					if !a {
						b.Error("key doesn't exist", j)
					}
				}
				for j := l; j < h; j++ {
					x, _ := M.Get(j)
					if x != j {
						b.Error("incorrect value", j, x)
					}
				}
				wg.Done()
			}(k*elementNum0, (k+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}
}

func Benchmark2HaxMap_Case1(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := haxmap.New[int, int]()
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					M.Set(j, j)
				}
				for j := l; j < h; j++ {
					_, a := M.Get(j)
					if !a {
						b.Error("key doesn't exist", j)
					}
				}
				for j := l; j < h; j++ {
					x, _ := M.Get(j)
					if x != j {
						b.Error("incorrect value", j, x)
					}
				}
				wg.Done()
			}(k*elementNum0, (k+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}
}

func Benchmark2IntMap_Case2(b *testing.B) {
	//runtime.GC()
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := IntMap.New[int, int](0, 2, elementNum0*iter0-1, hashInt)
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

func Benchmark2HashMap_Case2(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := hashmap.New[int, int]()
		for j := 0; j < elementNum0*iter0; j++ {
			M.Insert(j, j)
		}
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					x, _ := M.Get(j)
					if x != j {
						b.Error("incorrect value 1")
					}
				}
				for j := l; j < h; j++ {
					M.Set(j, j+1)
				}
				for j := l; j < h; j++ {
					x, _ := M.Get(j)
					if x != j+1 {
						b.Error("incorrect value 2")
					}
				}
				wg.Done()
			}(k*elementNum0, (k+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}
}

func Benchmark2HaxMap_Case2(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := haxmap.New[int, int]()
		for j := 0; j < elementNum0*iter0; j++ {
			M.Set(j, j)
		}
		b.StartTimer()
		for k := 0; k < iter0; k++ {
			wg.Add(1)
			go func(l, h int) {
				for j := l; j < h; j++ {
					x, _ := M.Get(j)
					if x != j {
						b.Error("incorrect value 1")
					}
				}
				for j := l; j < h; j++ {
					M.Set(j, j+1)
				}
				for j := l; j < h; j++ {
					x, _ := M.Get(j)
					if x != j+1 {
						b.Error("incorrect value 2")
					}
				}
				wg.Done()
			}(k*elementNum0, (k+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}
}

func Benchmark2IntMap_Case3(b *testing.B) {
	//runtime.GC()
	b.StopTimer()
	wg := &sync.WaitGroup{}
	for a := 0; a < b.N; a++ {
		M := IntMap.New[int, int](2, 8, iter0*elementNum0-1, hashInt)
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

func Benchmark2HashMap_Case3(b *testing.B) {
	b.StopTimer()
	wg := &sync.WaitGroup{}
	for a := 0; a < b.N; a++ {
		M := hashmap.New[int, int]()
		b.StartTimer()
		for j := 0; j < iter0; j++ {
			wg.Add(1)
			go func(l, h int) {
				defer wg.Done()
				for i := l; i < h; i++ {
					M.Insert(i, i)
				}

				for i := l; i < h; i++ {
					_, x := M.Get(i)
					if !x {
						b.Errorf("not put: %v\n", i)
					}
				}
				for i := l; i < h; i++ {
					M.Del(i)

				}
				for i := l; i < h; i++ {
					_, x := M.Get(i)
					if x {
						b.Errorf("not removed: %v\n", i)
					}
				}

			}(j*elementNum0, (j+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}

}

func Benchmark2HaxMap_Case3(b *testing.B) {
	b.StopTimer()
	wg := &sync.WaitGroup{}
	for a := 0; a < b.N; a++ {
		M := haxmap.New[int, int]()
		b.StartTimer()
		for j := 0; j < iter0; j++ {
			wg.Add(1)
			go func(l, h int) {
				defer wg.Done()
				for i := l; i < h; i++ {
					M.Set(i, i)
				}

				for i := l; i < h; i++ {
					_, x := M.Get(i)
					if !x {
						b.Errorf("not put: %v\n", i)
					}
				}
				for i := l; i < h; i++ {
					M.Del(i)

				}
				for i := l; i < h; i++ {
					_, x := M.Get(i)
					if x {
						b.Errorf("not removed: %v\n", i)
					}
				}

			}(j*elementNum0, (j+1)*elementNum0)
		}
		wg.Wait()
		b.StopTimer()
	}

}
