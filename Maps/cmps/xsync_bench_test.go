package cmps

import (
	"github.com/puzpuzpuz/xsync/v3"
	"sync/atomic"
	"testing"
)

func xsyncHashUint(v uint, _ uint64) uint64 {
	return uint64(v)
}
func fillXSyncMap(b *testing.B, keyRange uint) *xsync.MapOf[uint, uint] {
	b.Helper()
	m := xsync.NewMapOfWithHasher[uint, uint](xsyncHashUint)
	for i := range keyRange {
		m.Store(i, i)
	}
	return m
}
func BenchmarkXSyncMap_Load_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	vp := fillXSyncMap(b, hits)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, sideEff = vp.Load(uint(count.Add(1)-1) % (hits + misses))
		}
	})
}
func BenchmarkXSyncMap_LoadAndDelete_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	vp := fillXSyncMap(b, hits)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vp.LoadAndDelete(uint(count.Add(1)-1) % (hits + misses))
		}
	})
}
func BenchmarkXSyncMap_LoadAndDelete_Adversarial(b *testing.B) {
	const mapSize = 2048
	vp := fillXSyncMap(b, mapSize)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			c := a % mapSize
			vp.Load(c)
			if c == 0 {
				vp.Range(func(k uint, _ uint) bool {
					vp.Delete(k)
					return false
				})
				vp.Store(c, a)
			}
		}
	})
}
func BenchmarkXSyncMap_LoadOrStore_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	m := fillXSyncMap(b, hits)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			c := a % (hits + misses)
			m.LoadOrStore(c, a)
		}
	})
}
func BenchmarkXSyncMap_LoadOrStorePtr_Adversarial(b *testing.B) {
	vp := xsync.NewMapOf[uint, uint]()
	var count atomic.Uintptr
	b.ResetTimer()
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
func BenchmarkXSyncMap_Case1(b *testing.B) {
	const readRatio = 4
	m := xsync.NewMapOf[uint, uint]()
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
func BenchmarkXSyncMap_Case2(b *testing.B) {
	const actions = 3
	m := xsync.NewMapOf[uint, uint]()
	var loaded, count, vals atomic.Uintptr
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
