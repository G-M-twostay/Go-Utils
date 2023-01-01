package benchmarks

import (
	"GMUtils/Maps"
	"GMUtils/Maps/BucketMap"
	"github.com/cornelk/hashmap"
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

type O uintptr

func (u O) Equal(o Maps.Hashable) bool {
	return u == o.(O)
}

func (u O) Hash() uint {
	return uint(u)
}

func setupHashMap(b *testing.B) *hashmap.Map[uintptr, uintptr] {
	b.Helper()

	m := hashmap.New[uintptr, uintptr]()
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Set(i, i)
	}
	return m
}

func setupBMap(b *testing.B, a byte, c byte) *BucketMap.BucketMap[O, uintptr] {
	b.Helper()
	m := BucketMap.MakeBucketMap[O, uintptr](a, c, benchmarkItemCount)
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Store(O(i), i)
	}
	return m
}

func setupIntMap(b *testing.B, a, c byte) *BucketMap.IntMap[uintptr, uintptr] {
	b.Helper()
	m := BucketMap.MakeIntMap[uintptr, uintptr](a, c, benchmarkItemCount, func(x uintptr) uint { return uint(x) })
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Store(i, i)
	}
	return m
}

func BenchmarkReadHashMapUint(b *testing.B) {
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

func BenchmarkReadBMapUint(b *testing.B) {
	m := setupBMap(b, minLen1, maxLen1)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				j, _ := m.Load(O(i))
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadIntMapUint(b *testing.B) {
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

func BenchmarkReadHashMapWithWritesUint(b *testing.B) {
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

func BenchmarkReadBMapWithWritesUint(b *testing.B) {
	m := setupBMap(b, minLen2, maxLen2)
	var writer uintptr
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					m.Store(O(i), i)
				}
			}
		} else {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					j, _ := m.Load(O(i))
					if j != i {
						b.Fail()
					}
				}
			}
		}
	})
}

func BenchmarkReadIntMapWithWritesUint(b *testing.B) {
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

func BenchmarkWriteHashMapUint(b *testing.B) {
	m := hashmap.New[uintptr, uintptr]()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Set(i, i)
		}
	}
}

func BenchmarkWriteBMapUint(b *testing.B) {
	m := BucketMap.MakeBucketMap[O, uintptr](minLen3, maxLen3, benchmarkItemCount)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Store(O(i), i)
		}
	}
}

func BenchmarkWriteIntMapUint(b *testing.B) {
	m := BucketMap.MakeIntMap[uintptr, uintptr](minLen3, maxLen3, benchmarkItemCount, func(x uintptr) uint { return uint(x) })
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Store(uintptr(O(i)), i)
		}
	}
}
