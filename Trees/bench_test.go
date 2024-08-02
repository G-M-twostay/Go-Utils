package Trees

import (
	"cmp"
	impl1 "github.com/emirpasic/gods/trees/redblacktree"
	impl3 "github.com/google/btree"
	impl2 "github.com/petar/GoLLRB/llrb"
	"math/bits"
	"slices"
	"testing"
	"unsafe"
)

var (
	bAddN  uint32 = 1000000
	bRmvN  uint32 = bAddN
	bQryN  uint32 = bAddN / 2
	bBTDeg        = 4
)

func BenchmarkBT_ReplaceOrInsert(b *testing.B) {
	for range b.N {
		tree := impl3.NewOrderedG[int](bBTDeg)
		for range bAddN {
			tree.ReplaceOrInsert(rg.Int())
		}
	}
}
func BenchmarkLLRB_ReplaceOrInsertBulk(b *testing.B) {
	all := make([]impl2.Item, bAddN)
	b.ResetTimer()
	for range b.N {
		tree := impl2.New()
		for i := range len(all) {
			all[i] = impl2.Int(rg.Int())
		}
		tree.ReplaceOrInsertBulk(all...)
	}
}
func BenchmarkRBT_Put(b *testing.B) {
	for range b.N {
		tree := impl1.NewWithIntComparator()
		for range bAddN {
			tree.Put(rg.Int(), nil)
		}
	}
}
func BenchmarkSBT_Add0(b *testing.B) {
	for range b.N {
		tree := New[int](uint32(0))
		var buf []uintptr
		for range bAddN {
			_, buf = tree.Add(rg.Int(), buf[:0])
		}
	}
}
func BenchmarkSBT_Add1(b *testing.B) {
	for range b.N {
		tree := New[int](bAddN)
		buf := make([]uintptr, 0, bits.Len32(bAddN)*4/3)
		for range bAddN {
			tree.Add(rg.Int(), buf)
		}
	}
}
func createBT(b *testing.B, all []int) *impl3.BTreeG[int] {
	b.Helper()
	t := impl3.NewOrderedG[int](bBTDeg)
	for i := range all {
		all[i] = rg.Int()
		t.ReplaceOrInsert(all[i])
	}
	return t
}
func createLLRB(b *testing.B, all []impl2.Item) *impl2.LLRB {
	b.Helper()
	for i := range all {
		all[i] = impl2.Int(rg.Int())
	}
	t := impl2.New()
	t.ReplaceOrInsertBulk(all...)
	return t
}
func createSBT(b *testing.B) *Tree[int, uint32] {
	b.Helper()
	t := New[int, uint32](bAddN)
	buf := make([]uintptr, 0, bits.Len32(bAddN)*4/3)
	for range bAddN {
		t.Add(rg.Int(), buf)
	}
	return t
}
func createRBT(b *testing.B) *impl1.Tree {
	b.Helper()
	tree := impl1.NewWithIntComparator()
	for range bAddN {
		tree.Put(rg.Int(), nil)
	}
	return tree
}
func BenchmarkBT_Delete(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *createBT(b, all)
		b.StartTimer()
		for _, v := range all {
			tree.Delete(v)
		}
	}
}
func BenchmarkLLRB_Delete(b *testing.B) {
	all := make([]impl2.Item, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *createLLRB(b, all)
		b.StartTimer()
		for _, v := range all {
			tree.Delete(v)
		}
	}
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
			_, buf = tree.Del(v, buf[:0])
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
		buf := make([]uintptr, 0, bits.Len32(bAddN)*4/3)
		for _, v := range all {
			tree.Del(v, buf)
		}
	}
}

var sideEff0 uintptr
var sideEff1 bool

func BenchmarkRBT_Get(b *testing.B) {
	for range b.N {
		b.StopTimer()
		tree := *createRBT(b)
		all := tree.Keys()
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
			sideEff0 = uintptr(unsafe.Pointer(tree.Get(v)))
		}
		for range bAddN - bQryN {
			sideEff0 = uintptr(unsafe.Pointer(tree.Get(rg.Intn(m))))
		}
	}
}
func BenchmarkLLRB_Has(b *testing.B) {
	all := make([]impl2.Item, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *createLLRB(b, all)
		m := int(slices.MaxFunc(all[bQryN:], func(a, b impl2.Item) int {
			return cmp.Compare(a.(impl2.Int), b.(impl2.Int))
		}).(impl2.Int))
		b.StartTimer()
		for _, v := range all[:bQryN] {
			sideEff1 = tree.Has(v)
		}
		for range bAddN - bQryN {
			sideEff1 = tree.Has(impl2.Int(rg.Intn(m)))
		}
	}
}
func BenchmarkBT_Has(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *createBT(b, all)
		m := slices.Max(all[bQryN:])
		b.StartTimer()
		for _, v := range all[:bQryN] {
			sideEff1 = tree.Has(v)
		}
		for range bAddN - bQryN {
			sideEff1 = tree.Has(rg.Intn(m))
		}
	}
}
