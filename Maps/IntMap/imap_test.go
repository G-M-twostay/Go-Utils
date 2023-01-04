package IntMap

import (
	"GMUtils/Maps/BucketMap"
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

func TestIntMap_All(t *testing.T) {
	M := New[int, int](1, 1, blockNum*blockSize-1, func(x int) uint { return uint(x) })
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
	for cur := (M.buckets.Load().Get(0)); cur != nil; cur = (*node[int])(cur.nx) {
		if !cur.isRelay() {
			t.Log("have", M.HasKey(cur.k))
		}
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
	//t.Log("added")
	//M.Delete(O(0))
	//t.Log("delted 0")
	//M.Delete(O(1))
	//t.Log("removed 0 and 1")
	//M.Delete(O(2))
	//t.Log("removed 0 and 1 and 2")
	//for cur := (*M.buckets.Load())[0]; cur != nil; cur = (*node[O])(cur.nx) {
	//	t.Log(cur.String(), "\n")
	//}
	//M.Load(O(0))
}

//https://github.com/cornelk/hashmap/ is incorrect in all 3 test cases, see https://github.com/cornelk/hashmap/issues/73.

func BenchmarkBucketMap_Case1(b *testing.B) {
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := BucketMap.New[int, int](0, 2, elementNum0*iter0-1, hasher, cmp)
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

//	func BenchmarkHashMap_Case1(b *testing.B) {
//		b.StopTimer()
//		wg := sync.WaitGroup{}
//		for i := 0; i < b.N; i++ {
//			M := hashmap.New[int, int]()
//			b.StartTimer()
//			for k := 0; k < iter0; k++ {
//				wg.Add(1)
//				go func(l, h int) {
//					for j := l; j < h; j++ {
//						M.Insert(j, j)
//					}
//					for j := l; j < h; j++ {
//						_, a := M.Get(j)
//						if !a {
//							b.Error("key doesn't exist", j)
//						}
//					}
//					for j := l; j < h; j++ {
//						x, _ := M.Get(j)
//						if x != j {
//							b.Error("incorrect value", j, x)
//						}
//					}
//					wg.Done()
//				}(k*elementNum0, (k+1)*elementNum0)
//			}
//			wg.Wait()
//			b.StopTimer()
//		}
//	}
func BenchmarkBucketMap_Case2(b *testing.B) {
	//runtime.GC()
	b.StopTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		M := BucketMap.New[int, int](0, 2, elementNum0*iter0-1, hasher, cmp)
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
	//runtime.GC()
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

//	func BenchmarkHashMap_Case2(b *testing.B) {
//		b.StopTimer()
//		wg := sync.WaitGroup{}
//		for i := 0; i < b.N; i++ {
//			M := hashmap.New[int, int]()
//			for j := 0; j < elementNum0*iter0; j++ {
//				M.Insert(j, j)
//			}
//			b.StartTimer()
//			for k := 0; k < iter0; k++ {
//				wg.Add(1)
//				go func(l, h int) {
//					for j := l; j < h; j++ {
//						x, _ := M.Get(j)
//						if x != j {
//							b.Error("incorrect value 1")
//						}
//					}
//					for j := l; j < h; j++ {
//						M.Set(j, j+1)
//					}
//					for j := l; j < h; j++ {
//						x, _ := M.Get(j)
//						if x != j+1 {
//							b.Error("incorrect value 2")
//						}
//					}
//					wg.Done()
//				}(k*elementNum0, (k+1)*elementNum0)
//			}
//			wg.Wait()
//			b.StopTimer()
//		}
//	}
func BenchmarkBucketMap_Case3(b *testing.B) {
	//runtime.GC()
	b.StopTimer()
	wg := &sync.WaitGroup{}
	for a := 0; a < b.N; a++ {
		M := BucketMap.New[int, int](2, 8, iter0*elementNum0-1, hasher, cmp)
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

//func BenchmarkHashMap_Case3(b *testing.B) {
//	b.StopTimer()
//	wg := &sync.WaitGroup{}
//	for a := 0; a < b.N; a++ {
//		M := hashmap.New[int, int]()
//		b.StartTimer()
//		for j := 0; j < iter0; j++ {
//			wg.Add(1)
//			go func(l, h int) {
//				defer wg.Done()
//				for i := l; i < h; i++ {
//					M.Insert(i, i)
//				}
//
//				for i := l; i < h; i++ {
//					_, x := M.Get(i)
//					if !x {
//						b.Errorf("not put: %v\n", O(i))
//					}
//				}
//				for i := l; i < h; i++ {
//					M.Del(i)
//
//				}
//				for i := l; i < h; i++ {
//					_, x := M.Get(i)
//					if x {
//						b.Errorf("not removed: %v\n", O(i))
//					}
//				}
//
//			}(j*elementNum0, (j+1)*elementNum0)
//		}
//		wg.Wait()
//		b.StopTimer()
//	}
//
//}
