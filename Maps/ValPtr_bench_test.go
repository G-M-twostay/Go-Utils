package Maps

//these are mainly used for measuring performances to find the optimal implementation.
import (
	Go_Utils "github.com/g-m-twostay/go-utils"
	"math/rand/v2"
	"sync/atomic"
	"testing"
)

const (
	benchMinBucketSize = 2
	benchMaxBucketSize = 16
)

var hasher = Go_Utils.Hasher(rand.Uint())

func customStat(b *testing.B) {
	b.Helper()
}
func benchHashF(i uint) uint {
	return i
}
func makeWithKeys(b *testing.B, maxKey, maxHash uint) *ValPtr[uint, uint] {
	b.Helper()
	vp := NewValPtr[uint, uint](benchMinBucketSize, benchMaxBucketSize, maxHash, benchHashF)
	vs := make([]uint, maxKey)
	for i := range maxKey {
		vp.StorePtr(i, &vs[i])
	}
	return vp
}
func BenchmarkLoadPtr_MostlyHits(b *testing.B) {
	const hits, misses = 1023, 1
	vp := makeWithKeys(b, hits, hits+misses-1)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vp.LoadPtr(uint(count.Add(1)-1) % (hits + misses))
		}
	})
}
func BenchmarkLoadPtr_MostlyMisses(b *testing.B) {
	const hits, misses = 1, 1023
	vp := makeWithKeys(b, hits, hits+misses-1)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vp.LoadPtr(uint(count.Add(1)-1) % (hits + misses))
		}
	})
}
func BenchmarkDelete_Balanced(b *testing.B) {
	const hits, misses = 512, 512
	vp := makeWithKeys(b, hits, hits+misses-1)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vp.Delete(uint(count.Add(1)-1) % (hits + misses))
		}
	})
}
func BenchmarkDelete_NearlyAll(b *testing.B) {
	const hits, misses = 1023, 1
	vp := makeWithKeys(b, hits, hits+misses-1)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vp.Delete(uint(count.Add(1)-1) % (hits + misses))
		}
	})
}
func BenchmarkDelete_Unique(b *testing.B) {
	const maxHash = 1024
	vp := NewValPtr[uint, uint](benchMinBucketSize, benchMaxBucketSize, maxHash-1, benchHashF)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vp.Delete(uint(count.Add(1)-1) % maxHash)
		}
	})
}
func BenchmarkDelete_Collision(b *testing.B) {
	const maxHash = 1024
	vp := makeWithKeys(b, 1, maxHash-1)
	a := vp.LoadPtr(0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if vp.Delete(0) {
				vp.StorePtr(0, a)
			}
		}
	})
}
func BenchmarkStoreAndDelete_ScatteredEmpty(b *testing.B) {
	const maxHash, existing uint = 1024, 0
	vp := NewValPtr[uint, uint](benchMinBucketSize, benchMaxBucketSize, maxHash-1, benchHashF)
	all := make([]uint, maxHash)
	for i := range existing {
		vp.StorePtr(i, &all[i])
	}
	var count atomic.Uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if a := uint(count.Add(1)-1) % maxHash; a&1 == 1 {
				vp.StorePtr(a, &all[a])
			} else {
				vp.Delete(a)
			}
		}
	})
	customStat(b)
}
func BenchmarkStoreAndDelete_ScatteredFull(b *testing.B) {
	const maxHash, existing uint = 1024, 1024
	vp := NewValPtr[uint, uint](benchMinBucketSize, benchMaxBucketSize, maxHash-1, benchHashF)
	all := make([]uint, maxHash)
	for i := range existing {
		vp.StorePtr(i, &all[i])
	}
	var count atomic.Uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if a := uint(count.Add(1)-1) % maxHash; a&1 == 1 {
				vp.StorePtr(a, &all[a])
			} else {
				vp.Delete(a)
			}
		}
	})
	customStat(b)
}
func BenchmarkStoreAndDelete_ScatteredHalf(b *testing.B) {
	const maxHash, existing uint = 1024, 512
	vp := NewValPtr[uint, uint](benchMinBucketSize, benchMaxBucketSize, maxHash-1, benchHashF)
	all := make([]uint, maxHash)
	for i := range existing {
		vp.StorePtr(i, &all[i])
	}
	var count atomic.Uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if a := uint(count.Add(1)-1) % maxHash; a&1 == 1 {
				vp.StorePtr(a, &all[a])
			} else {
				vp.Delete(a)
			}
		}
	})
	customStat(b)
}
func BenchmarkStoreAndDelete_GroupedEmpty(b *testing.B) {
	const maxHash, existing uint = 1024, 0
	vp := NewValPtr[uint, uint](benchMinBucketSize, benchMaxBucketSize, maxHash-1, benchHashF)
	all := make([]uint, maxHash)
	for i := range existing {
		vp.StorePtr(i, &all[i])
	}
	var count atomic.Uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			c := a % maxHash
			if a&15 < 8 {
				vp.StorePtr(c, &all[c])
			} else {
				vp.Delete(c)
			}
		}
	})
	customStat(b)
}
func BenchmarkStoreAndDelete_GroupedFull(b *testing.B) {
	const maxHash, existing uint = 1024, 1024
	vp := NewValPtr[uint, uint](benchMinBucketSize, benchMaxBucketSize, maxHash-1, benchHashF)
	all := make([]uint, maxHash)
	for i := range existing {
		vp.StorePtr(i, &all[i])
	}
	var count atomic.Uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			c := a % maxHash
			if a&15 < 8 {
				vp.StorePtr(c, &all[c])
			} else {
				vp.Delete(c)
			}
		}
	})
	customStat(b)
}
func BenchmarkStoreAndDelete_GroupedHalf(b *testing.B) {
	const maxHash, existing uint = 1024, 512
	vp := NewValPtr[uint, uint](benchMinBucketSize, benchMaxBucketSize, maxHash-1, benchHashF)
	all := make([]uint, maxHash)
	for i := range existing {
		vp.StorePtr(i, &all[i])
	}
	var count atomic.Uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1) - 1)
			c := a % maxHash
			if a&15 < 8 {
				vp.StorePtr(c, &all[c])
			} else {
				vp.Delete(c)
			}
		}
	})
	customStat(b)
}
func BenchmarkDelete_Adversarial(b *testing.B) {
	const maxHash = 1024
	vp := makeWithKeys(b, maxHash, maxHash-1)
	var count atomic.Uintptr
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := (count.Add(1) - 1) % maxHash
			vp.LoadPtr(uint(a))
			if a == 0 {
				d, _ := vp.TakePtr()
				/*Take is designed to replace
				m.Range(func(k, _ any) bool {
					m.Delete(k)
					return false
				})
				*/
				vp.StorePtr(uint(a), vp.LoadPtrAndDelete(*d))
			}
		}
	})
	customStat(b)
}
func BenchmarkLoadOrStorePtr_Balanced(b *testing.B) {
	const hits, misses uint = 512, 512
	vp := NewValPtr[uint, uint](benchMinBucketSize, benchMaxBucketSize, hits+misses-1, benchHashF)
	all := make([]uint, hits+misses)
	for i := range hits {
		vp.StorePtr(i, &all[i])
	}
	var count atomic.Uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := uint(count.Add(1)-1) % (hits + misses)
			vp.LoadOrStorePtr(a, &all[a])
		}
	})
	customStat(b)
}
func BenchmarkLoadOrStorePtr_Collision(b *testing.B) {
	const maxHash = 1024
	vp := makeWithKeys(b, 1, maxHash-1)
	a := vp.LoadPtr(0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vp.LoadOrStorePtr(0, a)
		}
	})
	customStat(b)
}
func BenchmarkLoadOrStorePtr_Unique(b *testing.B) {
	const maxHash uint = 1 << 20
	vp := NewValPtr[uint, uint](benchMinBucketSize, benchMaxBucketSize, maxHash-1, benchHashF)
	all := make([]uint, maxHash)
	var count atomic.Uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := (uint(count.Add(1)) - 1) % maxHash
			vp.LoadOrStorePtr(a, &all[a])
		}
	})
	customStat(b)
}
func BenchmarkLoadOrStorePtr_Adversarial(b *testing.B) {
	const maxHash uint = 1 << 20
	vp := NewValPtr[uint, uint](benchMinBucketSize, benchMaxBucketSize, maxHash-1, benchHashF)
	all := make([]uint, maxHash)
	var count atomic.Uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for stores, loadSinceStores := uint(0), uint(0); pb.Next(); {
			a := (uint(count.Add(1)) - 1) % maxHash
			vp.LoadPtr(a)
			if loadSinceStores++; loadSinceStores > stores {
				vp.LoadOrStorePtr(a, &all[stores%maxHash])
				loadSinceStores = 0
				stores++
			}
		}
	})
	customStat(b)
}
