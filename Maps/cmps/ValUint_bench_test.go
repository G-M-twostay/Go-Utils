package cmps

import (
	Go_Utils "github.com/g-m-twostay/go-utils"
	"github.com/g-m-twostay/go-utils/Maps"
	"math"
	"math/rand/v2"
	"sync/atomic"
	"testing"
)

var (
	hasher  = Go_Utils.Hasher(rand.Uint())
	sideEff bool
)

func HashUint(v uint) uint {
	return v
}
func fillValUint(b *testing.B, keyRange, maxHash uint) *Maps.ValUintptr[uint, uint] {
	b.Helper()
	m := Maps.NewValUintptr[uint, uint](2, 8, maxHash, HashUint)
	for i := range keyRange {
		m.Store(i, i)
	}
	return m
}
func BenchmarkValUintptr_Load_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	vp := fillValUint(b, hits, hits+misses-1)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, sideEff = vp.Load(uint(count.Add(1)-1) % (hits + misses))
		}
	})
}
func BenchmarkValUintptr_LoadAndDelete_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	vp := fillValUint(b, hits, hits+misses-1)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vp.LoadAndDelete(uint(count.Add(1)-1) % (hits + misses))
		}
	})
}
func BenchmarkValUintptr_LoadAndDelete_Adversarial(b *testing.B) {
	const mapSize = 2048
	vp := fillValUint(b, mapSize, mapSize-1)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			c := a % mapSize
			vp.Load(c)
			if c == 0 {
				d, _ := vp.Take()
				vp.LoadAndDelete(*d)
				vp.Store(c, a)
			}
		}
	})
}
func BenchmarkValUintptr_LoadOrStore_Balanced(b *testing.B) {
	const hits, misses = 1024, 1024
	m := fillValUint(b, hits, hits+misses-1)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			c := a % (hits + misses)
			m.LoadOrStore(c, a)
		}
	})
}
func BenchmarkValUintptr_LoadOrStorePtr_Adversarial(b *testing.B) {
	vp := Maps.NewValUintptr[uint, uint](2, 8, math.MaxUint, hasher.HashUint)
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

// Case 1: when the entry for a given key is only ever written once but read many times, as in caches that only grow.

func BenchmarkValUintptr_Case1(b *testing.B) {
	const readRatio = 4
	m := Maps.NewValUintptr[uint, uint](2, 8, math.MaxUint, hasher.HashUint)
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

// Case 2: when multiple goroutines read, write, and overwrite entries for disjoint sets of keys.

func BenchmarkValUintptr_Case2(b *testing.B) {
	const actions = 3
	m := Maps.NewValUintptr[uint, uint](2, 8, math.MaxUint, hasher.HashUint)
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
func BenchmarkValUintptr_Case3(b *testing.B) {
	const actions, initSize = 5, 2048
	m := Maps.NewValUintptr[uint, uint](2, 8, math.MaxUint, hasher.HashUint)
	var left, right, count, read atomic.Uintptr
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
func BenchmarkValUintptr_Case4(b *testing.B) {
	const batchSize, ranges = 2048, 32
	m := Maps.NewValUintptr[uint, uint](2, 8, batchSize*ranges-1, HashUint)
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
