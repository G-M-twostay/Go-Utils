package cmps

import (
	"sync"
	"sync/atomic"
	"testing"
)

func fillSyncMap(b *testing.B, keyRange uint) *sync.Map {
	b.Helper()
	m := sync.Map{}
	for i := range keyRange {
		m.Store(i, i)
	}
	return &m
}
func BenchmarkSyncMap_Load_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	vp := fillSyncMap(b, hits)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, sideEff = vp.Load(uint(count.Add(1)-1) % (hits + misses))
		}
	})
}
func BenchmarkSyncMap_Delete_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	vp := fillSyncMap(b, hits)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vp.Delete(uint(count.Add(1)-1) % (hits + misses))
		}
	})
}
func BenchmarkSyncMap_Delete_Adversarial(b *testing.B) {
	const mapSize = 2048
	vp := fillSyncMap(b, mapSize)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			vp.Load(a)
			if a%mapSize == 0 {
				vp.Range(func(k, _ any) bool {
					vp.Delete(k)
					return false
				})
				vp.Store(a, a)
			}
		}
	})
}
func BenchmarkSyncMap_LoadOrStore_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	m := fillSyncMap(b, hits)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			c := a % (hits + misses)
			m.LoadOrStore(c, a)
		}
	})
}
func BenchmarkSyncMap_LoadOrStorePtr_Adversarial(b *testing.B) {
	vp := sync.Map{}
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for stores, loadSinceStores := uint(0), uint(0); pb.Next(); {
			a := uint(count.Add(1)) - 1
			vp.Load(a)
			if loadSinceStores++; loadSinceStores > stores {
				vp.LoadOrStore(a, a)
				loadSinceStores = 0
				stores++
			}
		}
	})
}
func BenchmarkSyncMap_Case1(b *testing.B) {
	const readRatio = 4
	m := sync.Map{}
	var loaded, count atomic.Uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			if a%readRatio == 0 {
				m.Store(uint(loaded.Add(1)-1), a)
			} else {
				_, sideEff = m.Load(a % uint(loaded.Load()))
			}
		}
	})
}
func BenchmarkSyncMap_Case2(b *testing.B) {
	const actions = 3
	m := sync.Map{}
	var loaded, count, vals atomic.Uintptr
	m.Store(0, 0)
	loaded.Store(1)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			switch a := uint(count.Add(1) - 1); a % actions {
			case 0:
				m.Store(uint(loaded.Add(1)-1), a)
			case 1:
				m.Store(uint(vals.Add(1)-1)%uint(loaded.Load()), a)
			default:
				_, sideEff = m.Load(uint(vals.Add(1)-1) % uint(loaded.Load()))
			}
		}
	})
}
func BenchmarkSyncMap_Case3(b *testing.B) {
	const actions, initSize = 5, 2048
	m := sync.Map{}
	var left, right, count, read atomic.Uintptr
	for i := range uint(initSize) {
		m.Store(i, i)
	}
	right.Store(initSize)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			switch a := uint(count.Add(1) - 1); a % actions {
			case 0:
				m.Store(uint(right.Add(1)-1), a)
			case 1:
				if r, l := right.Load(), left.Load(); r > l {
					m.LoadAndDelete(uint(left.Add(1) - 1))
				}
			default:
				if r, l := right.Load(), left.Load(); r > l {
					_, sideEff = m.Load(uint((read.Add(1)-1)%(r-l) + l))
				} else {
					_, sideEff = m.Load(uint(read.Add(1) - 1))
				}
			}
		}
	})
}
func BenchmarkSyncMap_Case4(b *testing.B) {
	const batchSize, ranges = 2048, 32
	m := sync.Map{}
	var count atomic.Uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1)-1) % ranges
			for i := range uint(batchSize) {
				i += a * batchSize
				m.LoadOrStore(i, i)
				_, sideEff = m.Load(i)
				m.LoadAndDelete(i)
			}
		}
	})
}
