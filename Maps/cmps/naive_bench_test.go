package cmps

import (
	"sync"
	"sync/atomic"
	"testing"
)

func fillNaiveMap(b *testing.B, keyRange uint) map[uint]uint {
	b.Helper()
	m := make(map[uint]uint, keyRange)
	for i := range keyRange {
		m[i] = i
	}
	return m
}
func loadOrStore(m map[uint]uint, key, val uint) (uint, bool) {
	if v, ok := m[key]; ok {
		return v, true
	} else {
		m[key] = val
		return 0, false
	}
}
func BenchmarkNaiveMap_Load_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	vp := fillNaiveMap(b, hits)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, sideEff = vp[uint(count.Add(1)-1)%(hits+misses)]
		}
	})
}
func BenchmarkNaiveMap_Delete_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	vp := fillNaiveMap(b, hits)
	var count uint
	lock := sync.Mutex{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lock.Lock()
			delete(vp, count%(hits+misses))
			count++
			lock.Unlock()
		}
	})
}
func BenchmarkNaiveMap_Delete_Adversarial(b *testing.B) {
	const mapSize = 2048
	vp := fillNaiveMap(b, mapSize)
	var count atomic.Uintptr
	lock := sync.RWMutex{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			lock.RLock()
			_, sideEff = vp[a]
			lock.RUnlock()
			if a%mapSize == 0 {
				lock.Lock()
				for k := range vp {
					delete(vp, k)
					break
				}
				vp[a] = a
				lock.Unlock()
			}
		}
	})
}
func BenchmarkNaiveMap_LoadOrStore_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	m := fillNaiveMap(b, hits)
	var count uint
	lock := sync.Mutex{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lock.Lock()
			loadOrStore(m, count%(hits+misses), count)
			count++
			lock.Unlock()
		}
	})
}
func BenchmarkNaiveMap_LoadOrStore_Adversarial(b *testing.B) {
	m := make(map[uint]uint)
	lock := sync.RWMutex{}
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for stores, loadSinceStores := uint(0), uint(0); pb.Next(); {
			a := uint(count.Add(1)) - 1
			lock.RLock()
			_, sideEff = m[a]
			lock.RUnlock()
			if loadSinceStores++; loadSinceStores > stores {
				lock.Lock()
				loadOrStore(m, a, a)
				lock.Unlock()
				loadSinceStores = 0
				stores++
			}
		}
	})
}
func BenchmarkNaiveMap_Case1(b *testing.B) {
	const readRatio = 4
	m := make(map[uint]uint)
	var count atomic.Uintptr
	var loaded uint
	lock := sync.RWMutex{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			if a%readRatio == 0 {
				lock.Lock()
				m[loaded] = a
				loaded++
				lock.Unlock()
			} else {
				lock.RLock()
				_, sideEff = m[a%loaded]
				lock.RUnlock()
			}
		}
	})
}
func BenchmarkNaiveMap_Case2(b *testing.B) {
	const actions = 3
	m := make(map[uint]uint)
	lock := sync.RWMutex{}
	var count, vals atomic.Uintptr
	var loaded uint
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			switch a := uint(count.Add(1) - 1); a % actions {
			case 0:
				lock.Lock()
				m[loaded] = a
				loaded++
				lock.Unlock()
			case 1:
				lock.Lock()
				m[uint(vals.Add(1)-1)%loaded] = a
				lock.Unlock()
			default:
				lock.RLock()
				_, sideEff = m[uint(vals.Add(1)-1)%loaded]
				lock.RUnlock()
			}
		}
	})
}
