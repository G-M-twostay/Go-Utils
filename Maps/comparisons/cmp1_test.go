package comparisons

import (
	"github.com/alphadose/haxmap"
	"github.com/cornelk/hashmap"
	"github.com/g-m-twostay/go-utils/Maps/BucketMap"
	"github.com/g-m-twostay/go-utils/Maps/IntMap"
	"sync/atomic"
	"testing"
)

const (
	benchmarkItemCount      = 1024
	minLen1            byte = 1
	maxLen1            byte = 1
	minLen2            byte = 2
	maxLen2            byte = 1
	minLen3            byte = 1
	maxLen3            byte = 5
)

func hashUintptr(x uintptr) uint {
	return uint(x)
}

func cmp(x, y uintptr) bool {
	return x == y
}

// compares with https://github.com/cornelk/hashmap using https://github.com/cornelk/hashmap/blob/main/benchmarks/benchmark_test.go.
// compares with https://github.com/alphadose/haxmap using the above benchmarks.
// Note that these 2 implementations are both incorrect, see cmp2_test.go
func setupHashMap(b *testing.B) *hashmap.Map[uintptr, uintptr] {
	b.Helper()

	m := hashmap.New[uintptr, uintptr]()
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Set(i, i)
	}
	return m
}

func setupBMap(b *testing.B, a byte, c byte) *BucketMap.BucketMap[uintptr, uintptr] {
	b.Helper()
	m := BucketMap.New[uintptr, uintptr](a, c, benchmarkItemCount, hashUintptr, cmp)
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Store(i, i)
	}
	return m
}

func setupIntMap(b *testing.B, a, c byte) *IntMap.IntMap[uintptr, uintptr] {
	b.Helper()
	m := IntMap.New[uintptr, uintptr](a, c, benchmarkItemCount, hashUintptr)
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Store(i, i)
	}
	return m
}

func setupHaxMap(b *testing.B) *haxmap.Map[uintptr, uintptr] {
	b.Helper()

	m := haxmap.New[uintptr, uintptr]()
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Set(i, i)
	}
	return m
}

func Benchmark1ReadHashMapUint(b *testing.B) {
	m := setupHashMap(b)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				j, _ := m.Get(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func Benchmark1ReadBMapUint(b *testing.B) {
	m := setupBMap(b, minLen1, maxLen1)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				j, _ := m.Load(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func Benchmark1ReadIntMapUint(b *testing.B) {
	m := setupIntMap(b, minLen1, maxLen1)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				j, _ := m.Load(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func Benchmark1ReadHaxMapUint(b *testing.B) {
	m := setupHaxMap(b)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				j, _ := m.Get(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func Benchmark1ReadHashMapWithWritesUint(b *testing.B) {
	m := setupHashMap(b)
	var writer uintptr
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					m.Set(i, i)
				}
			}
		} else {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					j, _ := m.Get(i)
					if j != i {
						b.Fail()
					}
				}
			}
		}
	})
}

func Benchmark1ReadBMapWithWritesUint(b *testing.B) {
	m := setupBMap(b, minLen2, maxLen2)
	var writer uintptr
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					m.Store(i, i)
				}
			}
		} else {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					j, _ := m.Load(i)
					if j != i {
						b.Fail()
					}
				}
			}
		}
	})
}

func Benchmark1ReadIntMapWithWritesUint(b *testing.B) {
	m := setupIntMap(b, minLen2, maxLen2)
	var writer uintptr
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					m.Store(i, i)
				}
			}
		} else {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					j, _ := m.Load(i)
					if j != i {
						b.Fail()
					}
				}
			}
		}
	})
}

func Benchmark1ReadHaxMapWithWritesUint(b *testing.B) {
	m := setupHaxMap(b)
	var writer uintptr
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					m.Set(i, i)
				}
			}
		} else {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					j, _ := m.Get(i)
					if j != i {
						b.Fail()
					}
				}
			}
		}
	})
}

func Benchmark1WriteHashMapUint(b *testing.B) {
	m := hashmap.New[uintptr, uintptr]()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Set(i, i)
		}
	}
}

func Benchmark1WriteBMapUint(b *testing.B) {
	m := BucketMap.New[uintptr, uintptr](minLen3, maxLen3, benchmarkItemCount, hashUintptr, cmp)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Store(i, i)
		}
	}
}

func Benchmark1WriteIntMapUint(b *testing.B) {
	m := IntMap.New[uintptr, uintptr](minLen3, maxLen3, benchmarkItemCount, func(x uintptr) uint { return uint(x) })
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Store(i, i)
		}
	}
}

func Benchmark1WriteHaxMapUint(b *testing.B) {
	m := haxmap.New[uintptr, uintptr]()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Set(i, i)
		}
	}
}
