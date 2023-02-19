package Trees

import "testing"
import "math/rand"

const (
	size = 1 << 15
	iter = 10
)

func BenchmarkSBTree_Insert(b *testing.B) {
	var t *SBTree[int, uint]
	for i := 0; i < b.N; i++ {
		t = New[int, uint]()
		for j, _ := range rand.Perm(size) {
			t.Insert(j)
		}
	}
	b.Log(t.averageDepth())
}

func BenchmarkSBTree_Delete(b *testing.B) {
	var t Tree[int]
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		t = New[int, uint]()
		for j, _ := range rand.Perm(size) {
			t.Insert(j)
		}
		b.StartTimer()
		for j := 0; j < size; j++ {
			t.Remove(j)
		}
	}
}

func BenchmarkSBTree_All(b *testing.B) {
	var t *SBTree[int, uint]
	for i := 0; i < b.N; i++ {
		t = New[int, uint]()
		for j, _ := range rand.Perm(size / 2) {
			t.Insert(j)
		}
		for j, k := range rand.Perm(size / 2) {
			if k&1 == 1 {
				t.Remove(j)
			}
		}
		for j, _ := range rand.Perm(size / 2) {
			t.Insert(j + size)
		}
		for j, k := range rand.Perm(size / 2) {
			if k&1 == 1 {
				t.Insert(j)
			}
		}
	}
	b.Log(t.averageDepth())
}
