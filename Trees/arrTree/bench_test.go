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
		for range bAddN {
			tree.Insert(_R.Int())
		}
	}
}
func BenchmarkAdd1(b *testing.B) {
	for range b.N {
		tree := *New[int](bAddN)
		var buf1 []uintptr
		for range bAddN {
			_, buf1 = tree.BufferedInsert(_R.Int(), buf1)
		}
	}
}
func create(b *testing.B) *SBTree[int, uint32] {
	b.Helper()
	tree := New[int, uint32](bAddN)
	buf := make([]uintptr, bits.Len32(bAddN))
	for range bAddN {
		_, buf = tree.BufferedInsert(_R.Int(), buf)
	}
	return tree
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
			tree.Remove(v)
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
		var buf []uintptr
		for _, v := range all {
			_, buf = tree.BufferedRemove(v, buf)
		}
	}
}
func BenchmarkQry(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *create(b)
		copy(all, unsafe.Slice((*int)(tree.vsHead), tree.ifsLen-1))
		m := slices.Max(all[bQryN:])
		b.StartTimer()
		for _, v := range all[:bQryN] {
			tree.Has(v)
		}
		for range bAddN - bQryN {
			tree.Has(_R.Intn(m))
		}
	}
}
