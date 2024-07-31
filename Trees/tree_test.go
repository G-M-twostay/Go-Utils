package Trees

import (
	"math/bits"
	"math/rand"
	"slices"
	"testing"
	"unsafe"
)

var rg = *rand.New(rand.NewSource(0))
var cache [4]uint

func (u *Tree[T, S]) _depth(curI S, d byte) {
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
func (u *Tree[T, S]) depth() float32 {
	cache[0], cache[1] = 0, 0
	u._depth(u.root, 1)
	return float32(cache[1]) / float32(cache[0])
}

const (
	tAddN        uint16 = 40000
	tAddValRange        = 80000
)

func TestTree_Add(t *testing.T) {
	tree := *New[int, uint16](1)
	content := make(map[int]struct{})
	{
		a := make([]int, tAddN)
		for i := range a {
			a[i] = rg.Intn(tAddValRange)
		}
		buf := make([]uintptr, bits.Len16(tAddN))
		for _, b := range a {
			_, in := content[b]
			if c, _ := tree.Add(b, buf); !in && c == false {
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
		if tree.Get(k) == nil {
			t.Errorf("tree does not have key %v", k)
		}
	}
	for _, v := range unsafe.Slice((*int)(tree.vsHead), tree.ifsLen-1) {
		if _, in := content[v]; !in {
			t.Errorf("tree has non existent key %v", v)
		}
	}
}
func TestTree_Del(t *testing.T) {
	tree := *New[int, uint16](1)
	content := make(map[int]struct{})
	if a, _ := tree.Del(0, nil); a != false {
		t.Errorf("empty tree has non existent key %v", 0)
	}
	{
		a := make([]int, tAddN)
		for i := range a {
			a[i] = rg.Intn(tAddValRange)
		}
		buf := make([]uintptr, bits.Len(tAddValRange))
		for _, b := range a {
			_, buf = tree.Add(b, buf)
			content[b] = struct{}{}
		}
		for i := range rg.Intn(len(a)) {
			_, in := content[a[i]]
			if b, _ := tree.Del(a[i], buf); b != in {
				t.Errorf("failed to delete key %v", a[i])
			}
			if b, _ := tree.Del(a[i], buf); b == true {
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
		if tree.Get(k) == nil {
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
func TestTree_AddDel(t *testing.T) {
	tree := *New[int, uint16](1)
	content := make(map[int]struct{})
	{
		a := make([]int, tAddN)
		for i := range a {
			a[i] = rg.Intn(tAddValRange)
		}
		buf := make([]uintptr, bits.Len16(tAddN))
		for _, b := range a {
			_, buf = tree.Add(b, buf)
			content[b] = struct{}{}
		}
		for i := range rg.Intn(len(a)) {
			_, buf = tree.Del(a[i], buf)
			delete(content, a[i])
		}
	}
	{
		a := make([]int, rg.Intn(int(tAddN)))
		for i := range a {
			a[i] = rg.Intn(tAddValRange)
		}
		for _, b := range a {
			_, in := content[b]
			if c, _ := tree.Add(b, nil); !in && c == false {
				t.Errorf("failed to insert key %v", b)
			}
			content[b] = struct{}{}
		}
		for i := range rg.Intn(len(a)) {
			_, in := content[a[i]]
			if b, _ := tree.Del(a[i], nil); b != in {
				t.Errorf("failed to delete key %v", a[i])
			}
			if b, _ := tree.Del(a[i], nil); b == true {
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
		if tree.Get(k) == nil {
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
func TestTree_InOrder0(t *testing.T) {
	tree := *New[int](uint16(1))
	content := make(map[int]struct{})
	{
		a := make([]int, tAddN)
		for i := range a {
			a[i] = rg.Intn(tAddValRange)
		}
		buf := make([]uintptr, tAddN)
		for _, b := range a {
			_, buf = tree.Add(b, buf)
			content[b] = struct{}{}
		}
	}
	for range 10 {
		var s []int
		tree.InOrder(func(v *int) bool {
			s = append(s, *v)
			return rg.Intn(int(tree.Size()/2)) == 0
		}, nil)
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
func TestTree_InOrder1(t *testing.T) {
	tree := *New[int](uint16(1))
	content := make(map[int]struct{})
	{
		a := make([]int, tAddN)
		for i := range a {
			a[i] = rg.Intn(tAddValRange)
		}
		buf := make([]uintptr, tAddN)
		for _, b := range a {
			_, buf = tree.Add(b, buf)
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
func TestTree_InOrderR0(t *testing.T) {
	tree := *New[int](uint16(1))
	content := make(map[int]struct{})
	{
		a := make([]int, tAddN)
		for i := range a {
			a[i] = rg.Intn(tAddValRange)
		}
		buf := make([]uintptr, tAddN)
		for _, b := range a {
			_, buf = tree.Add(b, buf)
			content[b] = struct{}{}
		}
	}
	for range 10 {
		var s []int
		tree.InOrder(func(v *int) bool {
			s = append(s, *v)
			return rg.Intn(int(tree.Size()/2)) == 0
		}, nil)
		for _, v := range s {
			if _, in := content[v]; !in {
				t.Errorf("sorted has non existent key %v", v)
			}
		}
		if slices.Reverse(s); !slices.IsSorted(s) {
			t.Log(s)
			t.Errorf("sorted is not sorted")
		}
	}
	var s []int
	tree.InOrderR(func(v *int) bool {
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
	if slices.Reverse(s); !slices.IsSorted(s) {
		t.Log(s)
		t.Errorf("sorted is not sorted")
	}
}
func TestTree_InOrderR1(t *testing.T) {
	tree := *New[int](uint16(1))
	content := make(map[int]struct{})
	{
		a := make([]int, tAddN)
		for i := range a {
			a[i] = rg.Intn(tAddValRange)
		}
		buf := make([]uintptr, tAddN)
		for _, b := range a {
			_, buf = tree.Add(b, buf)
			content[b] = struct{}{}
		}
	}
	var s []int
	tree.InOrderR(func(v *int) bool {
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
	if slices.Reverse(s); !slices.IsSorted(s) {
		t.Log(s)
		t.Errorf("sorted is not sorted")
	}
}
func TestTree_RankK(t *testing.T) {
	tree := *New[int](uint16(1))
	sorted := make([]int, 0, tAddN)
	{
		content := make(map[int]struct{})
		a := make([]int, tAddN)
		for i := range a {
			a[i] = rg.Intn(tAddValRange)
		}
		buf := make([]uintptr, bits.Len(tAddValRange))
		for _, b := range a {
			_, buf = tree.Add(b, buf)
			content[b] = struct{}{}
		}
		for k := range content {
			sorted = append(sorted, k)
		}
	}
	slices.Sort(sorted)
	for i, v := range sorted {
		a := tree.RankK(uint16(i))
		if a == nil {
			t.Fatalf("nil at rank k %d\n", i)
		}
		if *a != v {
			t.Fatalf("wrong rank k %d, want %d has %d\n", i, v, a)
		}
	}
}

func TestTree_buildIfs(t *testing.T) {
	count := uint16(tAddN)
	root, ifs := buildIfs(count, make([][3]uint16, 0, bits.Len16(count)))
	if ifs[root].sz != count {
		t.Fatalf("wrong size of ifs %d, want %d", ifs[root].sz, count)
	}
	if ifs[0].sz != 0 {
		t.Fatalf("wrong size at 0 %d", ifs[0].sz)
	}
	for i, v := range ifs[1:] {
		if v.sz != ifs[v.l].sz+ifs[v.r].sz+1 {
			t.Fatalf("wrong size at %d", i+1)
		}
	}
	st := make([]uint16, count/2)
	st[0] = root
	all := make(map[uint16]struct{}, count)
	for st = st[:1]; len(st) > 0; {
		top := st[len(st)-1]
		st = st[:len(st)-1]
		all[top] = struct{}{}
		if ifs[top].l != 0 {
			st = append(st, ifs[top].l)
		}
		if ifs[top].r != 0 {
			st = append(st, ifs[top].r)
		}
	}
	if uint16(len(all)) != count {
		t.Fatalf("unvisited %d %d", len(all), count)
	}
	for i := range count {
		if _, in := all[i+1]; !in {
			t.Fatalf("missing index %d", i)
		}
	}
}
func TestTree_From(t *testing.T) {
	content := make([]int, tAddN)
	{
		all := make(map[int]struct{}, len(content))
		for i := 0; i < len(content); {
			a := rg.Intn(tAddValRange)
			if _, in := all[a]; !in {
				all[a] = struct{}{}
				content[i] = a
				i++
			}
		}
	}
	slices.Sort(content)
	tree := *From[int, uint16](content)
	if tree.Size() != uint16(len(content)) {
		t.Fatalf("tree size is %d, want %d", tree.Size(), len(content))
	}
	{
		s := make([]int, 0, len(content))
		tree.InOrder(func(v *int) bool {
			s = append(s, *v)
			return true
		}, make([]uint16, 16))
		for i := range content {
			if s[i] != content[i] {
				t.Fatalf("wrong value at index %d", i)
			}
		}
	}
	t.Logf("depth: %f, size: %d.\n", tree.depth(), tree.Size())
}
func TestTree_RankOf(t *testing.T) {
	var tree *Tree[int, uint16]
	content := make([]int, tAddN)
	{
		for i := range int(tAddN) {
			content[i] = i * 2
		}
		a := make([]int, len(content))
		copy(a, content)
		tree = From[int, uint16](a)
	}
	for i, v := range content {
		a, b := tree.RankOf(v)
		if !b {
			t.Fatalf("should have %d %d", i, v)
		}
		if a != uint16(i) {
			t.Fatalf("1wrong rank %d %d", a, i)
		}
		a, b = tree.RankOf(v + 1)
		if b {
			t.Fatalf("shouldn't have %d %d", i, v)
		}
		if a != uint16(i)+1 {
			t.Fatalf("2wrong rank %d %d", a, i)
		}
	}
	a, b := tree.RankOf(-1)
	if b {
		t.Fatalf("shouldn't have %d", -1)
	}
	if a != 0 {
		t.Fatalf("wrong rank %d", a)
	}
	a, b = tree.RankOf(tAddValRange + 1)
	if b {
		t.Fatalf("shouldn't have %d", tAddValRange+1)
	}
	if a != tree.Size() {
		t.Fatalf("wrong rank %d", a)
	}
}

func TestTree_PreSucc(t *testing.T) {
	var tree *Tree[int, uint16]
	content := make([]int, tAddN+2)
	{
		content[0] = -1
		content[tAddN+1] = int(tAddN) * 3
		for i := uint16(1); i <= tAddN; i++ {
			content[i] = int(i) * 2
		}
		a := make([]int, len(content))
		copy(a, content)
		tree = From[int, uint16](a)
	}
	for i := uint16(1); i <= tAddN; i++ {
		a := *tree.Predecessor(content[i], true)
		if a != content[i-1] {
			t.Fatalf("wrong predecessor %d %d", a, content[i-1])
		}
		a = *tree.Successor(content[i], true)
		if a != content[i+1] {
			t.Fatalf("wrong successor %d %d", a, content[i+1])
		}
	}
	for i := uint16(1); i <= tAddN; i++ {
		a := *tree.Predecessor(content[i]-1, true)
		if a != content[i-1] {
			t.Fatalf("wrong predecessor %d %d", a, content[i-1])
		}
		a = *tree.Successor(content[i]+1, true)
		if a != content[i+1] {
			t.Fatalf("wrong successor %d %d", a, content[i+1])
		}
	}
	for i := uint16(1); i <= tAddN; i++ {
		a := *tree.Predecessor(content[i], false)
		if a != content[i] {
			t.Fatalf("wrong predecessor %d %d", a, content[i])
		}
		a = *tree.Successor(content[i], false)
		if a != content[i] {
			t.Fatalf("wrong successor %d %d", a, content[i])
		}
	}
	if tree.Predecessor(content[0], true) != nil {
		t.Fatal("shouldn't have predecessor")
	}
	if tree.Successor(content[len(content)-1], true) != nil {
		t.Fatal("shouldn't have successor")
	}
}

func TestTree_Compact(t *testing.T) {
	content := make([]int, 0, tAddN)
	tree := *New[int](uint32(tAddN))
	{
		all := make(map[int]struct{}, tAddN)
		buf := make([]uintptr, bits.Len16(tAddN))
		for range cap(content) / 2 {
			a := rg.Intn(tAddValRange)
			_, buf = tree.Add(a, buf)
			all[a] = struct{}{}
		}
		for range cap(content) {
			if rg.Uint32()&1 == 0 {
				i := len(all)
				for k := range all {
					if rg.Intn(i) == 0 {
						delete(all, k)
						_, buf = tree.Del(k, buf)
						break
					}
					i--
				}
			} else {
				a := rg.Intn(tAddValRange)
				_, buf = tree.Add(a, buf)
				all[a] = struct{}{}
			}
		}
	}
	t.Logf("depth: %f, size: %d.\n", tree.depth(), tree.Size())
	tree.InOrder(func(vp *int) bool {
		content = append(content, *vp)
		return true
	}, make([]uint32, 0))
	tree.Compact()
	tc := make([]int, 0, tree.Size())
	tree.InOrder(func(vp *int) bool {
		tc = append(tc, *vp)
		return true
	}, nil)
	if !slices.Equal(tc, content) {
		t.Log(content)
		t.Log(tc)
		t.Fail()
	}
	if tree.caps[0] != int(tree.ifsLen) {
		t.Fatal("not compact")
	}
}
