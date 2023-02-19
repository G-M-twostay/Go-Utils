package HashSet

import "testing"

func TestHashSet_All(t *testing.T) {
	S := New[int](16, 7, 0)
	for i := 0; i < 10; i++ {
		if !S.Put(i) {
			t.Error("wrong put 1")
		}
		if S.Put(i) {
			t.Error("wrong put 2")
		}
	}
	for i := 0; i < 10; i++ {
		if !S.Has(i) {
			t.Error("wrong has 1")
		}
	}
	for i := 0; i < 5; i++ {
		if !S.Remove(i) {
			t.Error("wrong remove 1")
		}
		if S.Remove(i) {
			t.Error("wrong remove 2")
		}
	}
	for i := 0; i < 5; i++ {
		if S.Has(i) {
			t.Error("wrong has 2")
		}
	}
}

func TestHashSet_Range(t *testing.T) {
	x := New[int](16, 8, 0)
	x.Put(1)
	x.Put(2)
	x.Put(3)
	count := 0
	x.Range(func(v int) bool {
		count++
		x.Put(4)
		x.Range(func(v int) bool {
			t.Log("in", count, "val:", v)
			x.Remove(2)
			return true
		})
		x.Put(2)

		x.Remove(1)
		x.Remove(2)
		x.Remove(3)
		x.Remove(4)
		x.Put(1)
		x.Put(2)
		x.Put(3)
		x.Put(4)

		t.Log("out", count, "val:", v)
		return true
	})
}
