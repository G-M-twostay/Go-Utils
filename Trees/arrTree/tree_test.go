package Trees

import (
	"math/rand"
	"testing"
)

var _R rand.Rand = *rand.New(rand.NewSource(0))
var cache [4]uint

func (u *base[T, S]) _depth(curI S, d byte) {
	cur := u.ifs[curI]
	if cur.l != 0 {
		u._depth(cur.l, d+1)
	}
	if cur.r != 0 {
		u._depth(cur.r, d+1)
	}
	if cur.l == 0 && cur.r == 0 {
		cache[0]++
		cache[1] += uint(d)
	}
}
func (u *base[T, S]) depth() float32 {
	cache[0], cache[1] = 0, 0
	u._depth(u.root, 1)
	return float32(cache[1]) / float32(cache[0])
}

const (
	insertNum      uint16 = 1000
	insertValRange        = 20000
)

func Test_Insert(t *testing.T) {
	tree := *New[int, uint16](1)
	content := make(map[int]struct{})
	{
		a := make([]int, insertNum)
		for i := range a {
			a[i] = _R.Intn(insertValRange)
		}
		for _, b := range a {
			_, in := content[b]
			if !in && tree.Insert(b) == false {
				t.Errorf("failed to insert key %v", b)
			}
			content[b] = struct{}{}
		}
	}
	if int(tree.Size()) != len(content) {
		t.Errorf("tree size is %d, want %d", tree.Size(), len(content))
	}
	t.Logf("depth: %f, size: %d.\n", tree.depth(), tree.Size())
	for k := range content {
		if !tree.Has(k) {
			t.Errorf("tree does not have key %v", k)
		}
	}
	for _, v := range tree.vs[1:] {
		if _, in := content[v]; !in {
			t.Errorf("tree has non existent key %v", v)
		}
	}
}
func TestDelete(t *testing.T) {
	tree := *New[int, uint16](1)
	content := make(map[int]struct{})
	if tree.Remove(0) != false {
		t.Errorf("empty tree has non existent key %v", 0)
	}
	{
		a := make([]int, insertNum)
		for i := range a {
			a[i] = _R.Intn(insertValRange)
		}
		for _, b := range a {
			tree.Insert(b)
			content[b] = struct{}{}
		}
		for i := range len(a) {
			if tree.Remove(a[i]) == false {
				for b := tree.free; b != 0; b = tree.ifs[b].l {
					if tree.vs[b] == a[i] {
						t.Errorf("delete key %v", b)
					}
				}
				for _, v := range tree.vs {
					if v == a[i] {
						t.Errorf("in %v", a[i])
					}
				}
			}
			if tree.Remove(a[i]) == true {
				t.Errorf("can delete a second time key %v", a[i])
			}
			delete(content, a[i])
		}
	}
	if int(tree.Size()) != len(content) {
		t.Errorf("tree size is %d, want %d", tree.Size(), len(content))
	}
	t.Logf("depth: %f, size: %d.\n", tree.depth(), tree.Size())
	for k := range content {
		if !tree.Has(k) {
			t.Errorf("tree does not have key %v", k)
		}
	}
	{
		empties := make(map[int]struct{})
		empties[0] = struct{}{}
		for a := tree.popFree(); a != 0; a = tree.popFree() {
			empties[int(a)] = struct{}{}
		}
		for i, v := range tree.vs[:] {
			_, in1 := content[v]
			_, in2 := empties[i]
			if !in2 {
				if !in1 {
					t.Errorf("tree has non existent key %v at %d", v, i)
				}
			}
		}
	}
}
func TestInsertDel(t *testing.T) {
	tree := *New[int, uint16](1)
	content := make(map[int]struct{})
	{
		a := make([]int, insertNum)
		for i := range a {
			a[i] = _R.Intn(insertValRange)
		}
		for _, b := range a {
			tree.Insert(b)
			content[b] = struct{}{}
		}
		for i := range _R.Intn(len(a)) {
			tree.Remove(a[i])
			delete(content, a[i])
		}
	}
	{
		a := make([]int, _R.Intn(int(insertNum)))
		for i := range a {
			a[i] = _R.Intn(insertValRange)
		}
		for _, b := range a {
			_, in := content[b]
			if !in && tree.Insert(b) == false {
				t.Errorf("failed to insert key %v", b)
			}
			content[b] = struct{}{}
		}
		for i := range _R.Intn(len(a)) {
			if tree.Remove(a[i]) == false {
				t.Errorf("failed to delete key %v", a[i])
			}
			if tree.Remove(a[i]) == true {
				t.Errorf("can delete a second time key %v", a[i])
			}
			delete(content, a[i])
		}
	}
	if int(tree.Size()) != len(content) {
		t.Errorf("tree size is %d, want %d", tree.Size(), len(content))
	}
	t.Logf("depth: %f, size: %d.\n", tree.depth(), tree.Size())
	for k := range content {
		if !tree.Has(k) {
			t.Errorf("tree does not have key %v", k)
		}
	}
	{
		empties := make(map[int]struct{})
		empties[0] = struct{}{}
		for a := tree.popFree(); a != 0; a = tree.popFree() {
			empties[int(a)] = struct{}{}
		}
		for i, v := range tree.vs[:] {
			_, in1 := content[v]
			_, in2 := empties[i]
			if !in2 {
				if !in1 {
					t.Errorf("tree has non existent key %v at %d", v, i)
				}
			}
		}
	}
}
