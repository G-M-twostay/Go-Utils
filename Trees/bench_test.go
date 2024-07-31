package Trees

import (
	"math/bits"
	"slices"
	"testing"
	"unsafe"
)

var (
	bAddN uint32 = 1000000
	bRmvN uint32 = bAddN
	bQryN uint32 = bAddN / 2
)

func BenchmarkAdd0(b *testing.B) {
	for range b.N {
		tree := *New[int](uint32(0))
		var buf []uintptr
		for range bAddN {
			_, buf = tree.Add(rg.Int(), buf)
		}
	}
}
func BenchmarkAdd1(b *testing.B) {
	for range b.N {
		tree := *New[int](bAddN)
		buf := make([]uintptr, bits.Len32(bAddN))
		for range bAddN {
			_, buf = tree.Add(rg.Int(), buf)
		}
	}
}
func create(b *testing.B) *Tree[int, uint32] {
	b.Helper()
	all := make([]int, bAddN)
	rg.Read(unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(all))), uintptr(bAddN)*unsafe.Sizeof(0)))
	slices.Sort(all)
	return From[int, uint32](all)
}
func BenchmarkDel0(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *create(b)
		copy(all, unsafe.Slice((*int)(tree.vsHead), tree.ifsLen-1))
		b.StartTimer()
		for _, v := range all {
			tree.Del(v, nil)
		}
	}
}

func BenchmarkDel1(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *create(b)
		copy(all, unsafe.Slice((*int)(tree.vsHead), tree.ifsLen-1))
		b.StartTimer()
		buf := make([]uintptr, bits.Len32(bAddN))
		for _, v := range all {
			_, buf = tree.Del(v, buf)
		}
	}
}

var sideEff *int

func BenchmarkQry(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *create(b)
		copy(all, unsafe.Slice((*int)(tree.vsHead), tree.ifsLen-1))
		rg.Shuffle(int(bQryN), func(i, j int) {
			all[i], all[j] = all[j], all[i]
		})
		m := slices.Max(all[bQryN:])
		b.StartTimer()
		for _, v := range all[:bQryN] {
			sideEff = tree.Get(v)
		}
		for range bAddN - bQryN {
			sideEff = tree.Get(rg.Intn(m))
		}
	}
}
