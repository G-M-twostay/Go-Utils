package Trees

import (
	"fmt"
	"math/bits"
	"slices"
	"testing"
)

var (
	bAddN uint32 = 1000000
	bRmvN uint32 = bAddN
	bQryN uint32 = bRmvN
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
func BenchmarkDelQry(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *create(b)
		copy(all, tree.vs)
		m := slices.Max(all[bRmvN:])
		b.StartTimer()
		var buf []uint32
		for _, v := range all[bRmvN:] {
			_, buf = tree.BufferedRemove(v, buf[:0])
		}
		tree.maintainLeft(&tree.root)
		tree.maintainRight(&tree.root)
		for _, v := range all[:bRmvN] {
			tree.Has(v)
		}
		for range bQryN {
			tree.Has(_R.Intn(m))
		}
	}
}

var bNumSteps uint32 = 25

func BenchmarkDelQry2(b *testing.B) {
	for i := uint32(1); i < bNumSteps; i++ {
		bRmvN = bAddN / bNumSteps * i
		bQryN = bRmvN
		b.Run(fmt.Sprintf("%d", i), BenchmarkDelQry)
	}
}
