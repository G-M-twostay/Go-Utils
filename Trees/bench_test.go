package Trees

import (
	"cmp"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
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

func BenchmarkRBT_Put(b *testing.B) {
	for range b.N {
		tree := *rbt.NewWithIntComparator()
		for range bAddN {
			tree.Put(rg.Int(), nil)
		}
	}
}
func BenchmarkSBT_Add0(b *testing.B) {
	for range b.N {
		tree := *New[int](uint32(0))
		var buf []uintptr
		for range bAddN {
			_, buf = tree.Add(rg.Int(), buf)
		}
	}
}
func BenchmarkSBT_Add1(b *testing.B) {
	for range b.N {
		tree := *New[int](bAddN)
		buf := make([]uintptr, bits.Len32(bAddN)*4/3)
		for range bAddN {
			_, buf = tree.Add(rg.Int(), buf)
		}
	}
}
func createSBT(b *testing.B) *Tree[int, uint32] {
	b.Helper()
	all := make([]int, bAddN)
	rg.Read(unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(all))), uintptr(bAddN)*unsafe.Sizeof(0)))
	slices.Sort(all)
	return From[int, uint32](all)
}
func createRBT(b *testing.B) *rbt.Tree {
	b.Helper()
	tree := rbt.NewWithIntComparator()
	for range bAddN {
		tree.Put(rg.Int(), nil)
	}
	return tree
}
func BenchmarkRBT_Remove(b *testing.B) {
	for range b.N {
		b.StopTimer()
		tree := *createRBT(b)
		all := tree.Keys()
		b.StartTimer()
		for _, v := range all {
			tree.Remove(v)
		}
	}
}
func BenchmarkSBT_Del0(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *createSBT(b)
		copy(all, unsafe.Slice((*int)(tree.vsHead), tree.ifsLen-1))
		b.StartTimer()
		var buf []uintptr
		for _, v := range all {
			_, buf = tree.Del(v, buf)
		}
	}
}

func BenchmarkSBT_Del1(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *createSBT(b)
		copy(all, unsafe.Slice((*int)(tree.vsHead), tree.ifsLen-1))
		b.StartTimer()
		buf := make([]uintptr, bits.Len32(bAddN))
		for _, v := range all {
			_, buf = tree.Del(v, buf)
		}
	}
}

var sideEff0 *int
var sideEff1 bool

func BenchmarkRBT_Get(b *testing.B) {
	for range b.N {
		b.StopTimer()
		tree := *createRBT(b)
		all := tree.Keys()
		rg.Shuffle(int(bQryN), func(i, j int) {
			all[i], all[j] = all[j], all[i]
		})
		m := slices.MaxFunc(all[bQryN:], func(a, b any) int {
			return cmp.Compare(a.(int), b.(int))
		}).(int)
		b.StartTimer()
		for _, v := range all[:bQryN] {
			_, sideEff1 = tree.Get(v)
		}
		for range bAddN - bQryN {
			_, sideEff1 = tree.Get(rg.Intn(m))
		}
	}
}
func BenchmarkSBT_Get(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *createSBT(b)
		copy(all, unsafe.Slice((*int)(tree.vsHead), tree.ifsLen-1))
		rg.Shuffle(int(bQryN), func(i, j int) {
			all[i], all[j] = all[j], all[i]
		})
		m := slices.Max(all[bQryN:])
		b.StartTimer()
		for _, v := range all[:bQryN] {
			sideEff0 = tree.Get(v)
		}
		for range bAddN - bQryN {
			sideEff0 = tree.Get(rg.Intn(m))
		}
	}
}
