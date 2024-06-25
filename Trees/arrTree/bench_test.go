package Trees

import (
	"math/bits"
	"testing"
)

const (
	bAddN uint32 = 1000000
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
			_, buf1 = tree.BufferedInsert(_R.Int(), buf1[:0])
		}
	}
}
func create(b *testing.B) *SBTree[int, uint32] {
	b.Helper()
	tree := New[int, uint32](bAddN)
	buf := make([]uintptr, bits.Len32(bAddN))
	for range bAddN {
		_, buf = tree.BufferedInsert(_R.Int(), buf[:0])
	}
	return tree
}
func BenchmarkDel0(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *create(b)
		copy(all, tree.vs)
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
		copy(all, tree.vs)
		b.StartTimer()
		var buf []uint32
		for _, v := range all {
			_, buf = tree.BufferedRemove(v, buf[:0])
		}
	}
}
