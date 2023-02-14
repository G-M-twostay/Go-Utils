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
