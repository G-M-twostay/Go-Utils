package Trees

import (
	"math/rand"
	"slices"
	"testing"
	"unsafe"
)

var _R rand.Rand = *rand.New(rand.NewSource(0))
var cache [4]uint

func (u *SBTree[T, S]) _depth(curI S, d byte) {
	cur := u.getIf(curI)
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
func (u *SBTree[T, S]) depth() float32 {
	cache[0], cache[1] = 0, 0
	u._depth(u.root, 1)
	return float32(cache[1]) / float32(cache[0])
}

const (
	tAddN        uint16 = 40000
	tAddValRange        = 20000
)

func Test_Insert(t *testing.T) {
	tree := *New[int, uint16](1)
	content := make(map[int]struct{})
	{
		a := make([]int, tAddN)
		for i := range a {
			a[i] = _R.Intn(tAddValRange)
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
	for _, v := range unsafe.Slice((*int)(tree.vsHead), tree.ifsLen-1) {
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
		a := make([]int, tAddN)
		for i := range a {
			a[i] = _R.Intn(tAddValRange)
		}
		for _, b := range a {
			tree.Insert(b)
			content[b] = struct{}{}
		}
		for i := range _R.Intn(len(a)) {
			_, in := content[a[i]]
			if tree.Remove(a[i]) != in {
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
		for i, v := range unsafe.Slice((*int)(tree.vsHead), tree.ifsLen-1) {
			_, in1 := content[v]
			_, in2 := empties[i+1]
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
		a := make([]int, tAddN)
		for i := range a {
			a[i] = _R.Intn(tAddValRange)
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
		a := make([]int, _R.Intn(int(tAddN)))
		for i := range a {
			a[i] = _R.Intn(tAddValRange)
		}
		for _, b := range a {
			_, in := content[b]
			if !in && tree.Insert(b) == false {
				t.Errorf("failed to insert key %v", b)
			}
			content[b] = struct{}{}
		}
		for i := range _R.Intn(len(a)) {
			_, in := content[a[i]]
			if tree.Remove(a[i]) != in {
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
		for i, v := range unsafe.Slice((*int)(tree.vsHead), tree.ifsLen-1) {
			_, in1 := content[v]
			_, in2 := empties[i+1]
			if !in2 {
				if !in1 {
					t.Errorf("tree has non existent key %v at %d", v, i)
				}
			}
		}
	}
}
func TestInOrder0(t *testing.T) {
	tree := *New[int](uint16(1))
	content := make(map[int]struct{})
	{
		a := make([]int, tAddN)
		for i := range a {
			a[i] = _R.Intn(tAddValRange)
		}
		for _, b := range a {
			tree.Insert(b)
			content[b] = struct{}{}
		}
	}
	var s []int
	tree.InOrder(func(v *int) bool {
		s = append(s, *v)
		return true
	}, nil)
	if int(tree.Size()) != len(s) {
		t.Errorf("sorted size is %d, want %d", tree.Size(), len(content))
	}
	t.Logf("depth: %f, size: %d.\n", tree.depth(), tree.Size())
	for k := range content {
		if !slices.Contains(s, k) {
			t.Errorf("sorted does not have key %v", k)
		}
	}
	for _, v := range s {
		if _, in := content[v]; !in {
			t.Errorf("sorted has non existent key %v", v)
		}
	}
	if !slices.IsSorted(s) {
		t.Log(s)
		t.Errorf("sorted is not sorted")
	}
}
func TestInOrder1(t *testing.T) {
	tree := *New[int](uint16(1))
	content := make(map[int]struct{})
	{
		a := make([]int, tAddN)
		for i := range a {
			a[i] = _R.Intn(tAddValRange)
		}
		for _, b := range a {
			tree.Insert(b)
			content[b] = struct{}{}
		}
	}
	var s []int
	tree.InOrder(func(v *int) bool {
		s = append(s, *v)
		return true
	}, make([]uint16, 0))
	if int(tree.Size()) != len(s) {
		t.Errorf("sorted size is %d, want %d", tree.Size(), len(content))
	}
	t.Logf("depth: %f, size: %d.\n", tree.depth(), tree.Size())
	for k := range content {
		if !slices.Contains(s, k) {
			t.Errorf("sorted does not have key %v", k)
		}
	}
	for _, v := range s {
		if _, in := content[v]; !in {
			t.Errorf("sorted has non existent key %v", v)
		}
	}
	if !slices.IsSorted(s) {
		t.Log(s)
		t.Errorf("sorted is not sorted")
	}
}
