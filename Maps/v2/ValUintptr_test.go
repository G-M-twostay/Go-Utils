package v2

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

type testVUintptrT uintptr

func TestValUintptr_LoadOrStore2(t *testing.T) {
	mq := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	for i := range testVUintptrT(testThrdsN) {
		go func() {
			for j := i * testAddNEach; j < (i+1)*testAddNEach; j++ {
				mq.LoadOrStore(testVPT(j), j)
				if a, b := mq.LoadOrStore(testVPT(j), j); !b || a != j {
					t.Fail()
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for i := range testVUintptrT(testThrdsN * testAddNEach) {
		av, l := mq.LoadOrStore(testVPT(i), i)
		if !l || av != i {
			t.Fail()
		}
	}
	if mq.Size() != testThrdsN*testAddNEach {
		t.Fail()
	}
}
func TestValUintptr_LoadOrStore3(t *testing.T) {
	mq := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	counts := make([]atomic.Uint32, testThrdsN*testAddNEach)
	for range testThrdsN {
		go func() {
			for i := range testVUintptrT(testThrdsN * testAddNEach) {
				if _, b := mq.LoadOrStore(testVPT(i), i); !b {
					counts[i].Add(1)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for i := range counts {
		if counts[i].Load() != 1 {
			t.Fail()
		}
	}
	if mq.Size() != testThrdsN*testAddNEach {
		t.Fail()
	}
}
func TestValUintptr_LoadOrStore1(t *testing.T) {
	std := make(map[testVPT]testVUintptrT, testAddN/2)
	mq := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i := range testVUintptrT(rand.Intn(testAddN)) {
		k := testVPT(i)
		if _, a := mq.LoadOrStore(k, i); a {
			t.Fail()
		}
		if mq.Size() != uint(i)+1 {
			t.Fail()
		}
		std[k] = i
	}
	for k, ev := range std {
		av, b := mq.LoadOrStore(k, 0)
		if !b || av != ev {
			t.Fatal(av, ev)
		}
	}
	if mq.Size() != uint(len(std)) {
		t.Fail()
	}
}
func TestValUintptr_Load_Store1(t *testing.T) {
	mq := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i := range testVUintptrT(testAddN) {
		if !mq.Store(testVPT(i), i) {
			t.Fail()
		}
		if mq.Size() != uint(i)+1 {
			t.Fail()
		}
	}
	for k := range testVUintptrT(testAddN) {
		if a, b := mq.Load(testVPT(k)); a != k || !b {
			t.Fail()
		}
	}
}
func TestValUintptr_Load_Store2(t *testing.T) {
	mq := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	for i := range testThrdsN {
		go func() {
			for j := i * testAddNEach; j < (i+1)*testAddNEach; j++ {
				if !mq.Store(testVPT(j), testVUintptrT(j)) {
					t.Fail()
				}
				if a, b := mq.Load(testVPT(j)); !b || a != testVUintptrT(j) {
					t.Fail()
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if mq.Size() != testThrdsN*testAddNEach {
		t.Fail()
	}
	for i := range testVUintptrT(testThrdsN * testAddNEach) {
		if a, b := mq.Load(testVPT(i)); a != i || !b {
			t.Fail()
		}
	}
}
func TestValUintptr_Load_Store_Delete(t *testing.T) {
	mq := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	for i := range testThrdsN {
		go func() {
			for j := i * testAddNEach; j < (i+1)*testAddNEach; j++ {
				mq.Store(testVPT(j), testVUintptrT(j))
				if a, b := mq.Load(testVPT(j)); a != testVUintptrT(j) || !b {
					t.Error("didn't store", j, a)
				}
				mq.LoadAndDelete(testVPT(j))
				if _, b := mq.Load(testVPT(j)); b {
					t.Error("didn't delete", j)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if mq.Size() != 0 {
		t.Fail()
	}
}
func TestValUintptr_LoadAndDelete1(t *testing.T) {
	mq := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i := range testAddN {
		mq.Store(testVPT(i), testVUintptrT(i))
	}
	for i := range testVUintptrT(testAddN) {
		if a, b := mq.LoadAndDelete(testVPT(i)); a != i || !b {
			t.Fatal("wrong delete", a, i)
		}
		if _, b := mq.LoadAndDelete(testVPT(i)); b {
			t.Fatal("can't delete")
		}
		if mq.Size() != uint(testAddN-i)-1 {
			t.Fail()
		}
	}
}
func TestValUintptr_LoadPtrAndDelete2(t *testing.T) {
	mq := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i := range testThrdsN * testAddNEach {
		mq.Store(testVPT(i), testVUintptrT(i))
	}
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	for i := range testThrdsN {
		go func() {
			for j := i * testAddNEach; j < (i+1)*testAddNEach; j++ {
				if a, b := mq.LoadAndDelete(testVPT(j)); a != testVUintptrT(j) || !b {
					t.Error("wrong delete", a, j)
				}
				if _, b := mq.LoadAndDelete(testVPT(j)); b {
					t.Error("can't delete")
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
func TestValUintptr_LoadPtrAndDelete3(t *testing.T) {
	mq := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i := range testThrdsN * testAddNEach {
		mq.Store(testVPT(i), testVUintptrT(i))
	}
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	count := make([]atomic.Uint32, testThrdsN*testAddNEach)
	for range testThrdsN {
		go func() {
			for i := range testVPT(testThrdsN * testAddNEach) {
				if _, a := mq.LoadAndDelete(i); a {
					count[i].Add(1)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for i := range count {
		if count[i].Load() != 1 {
			t.Fail()
		}
	}
	if mq.Size() != 0 {
		t.Fail()
	}
}
func TestValUintptr_Swap(t *testing.T) {
	vp := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, 16, testHashF)
	if _, b := vp.Swap(0, 0); b {
		t.Fail()
	}
	v1, v2 := testVUintptrT(0), testVUintptrT(1)
	vp.Store(0, v1)
	if a, b := vp.Swap(0, v2); !b || a != v1 {
		t.Fail()
	}
	if a, b := vp.Load(0); !b || a != v2 {
		t.Fail()
	}
}
func TestValUintptr_LoadOrStore_Delete(t *testing.T) {
	mq := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	for i := range testThrdsN {
		go func() {
			for j := i * testAddNEach; j < (i+1)*testAddNEach; j++ {
				mq.LoadOrStore(testVPT(j), testVUintptrT(j))
				if a, b := mq.LoadOrStore(testVPT(j), testVUintptrT(j)); !b || a != testVUintptrT(j) {
					t.Error("can't store", j)
				}
				if a, b := mq.LoadAndDelete(testVPT(j)); a != testVUintptrT(j) || !b {
					t.Error("wrong delete", a, j)
				}
				if _, b := mq.LoadAndDelete(testVPT(j)); b {
					t.Error("can't delete")
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
func TestValUintptr_CompareAndSwap(t *testing.T) {
	vp := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, 16, testHashF)
	if vp.CompareAndSwap(0, 0, 0) != NULL {
		t.Fail()
	}
	vp.Store(0, 0)
	results := make([]bool, 4)
	for range rand.Intn(testAddN) {
		wg := sync.WaitGroup{}
		wg.Add(4)
		go func() {
			if a := vp.CompareAndSwap(0, 0, 1); a == NULL {
				t.Fail()
			} else {
				results[0] = a == SUCCESS
			}
			wg.Done()
		}()
		go func() {
			if a := vp.CompareAndSwap(0, 0, 4); a == NULL {
				t.Fail()
			} else {
				results[3] = a == SUCCESS
			}
			wg.Done()
		}()
		go func() {
			if a := vp.CompareAndSwap(0, 1, 2); a == NULL {
				t.Fail()
			} else {
				results[1] = a == SUCCESS
			}
			wg.Done()
		}()
		go func() {
			if a := vp.CompareAndSwap(0, 1, 3); a == NULL {
				t.Fail()
			} else {
				results[2] = a == SUCCESS
			}
			wg.Done()
		}()
		wg.Wait()
		vp.Store(0, 0)
		if results[1] && results[2] {
			t.Fatal("1 2 are exclusive")
		}
		if (results[1] || results[2]) && !results[0] {
			t.Fatal("1 2 depends on 0")
		}
		if results[0] == results[3] {
			t.Fatal("0 3 are exclusive")
		}
	}
}
func TestValUintptr_Take(t *testing.T) {
	vp := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, 16, testHashF)
	if kp, _ := vp.Take(); kp != nil {
		t.Fail()
	}
	a := testVUintptrT(15)
	vp.Store(15, a)
	if kp, v := vp.Take(); v != a || *kp != 15 {
		t.Fail()
	}
	b := testVUintptrT(0)
	vp.Store(0, b)
	if kp, v := vp.Take(); v != b || *kp != 0 {
		t.Fail()
	}
}
func TestValUintptr_Range(t *testing.T) {
	mq := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i := range testAddN {
		mq.Store(testVPT(i), testVUintptrT(i))
	}
	count := 0
	for k, v := range mq.Range {
		if k != testVPT(count) {
			t.Fail()
		}
		if v != testVUintptrT(count) {
			t.Fail()
		}
		count++
	}
}
func TestValUintptr_Copy(t *testing.T) {
	vp0 := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for range rand.Intn(testAddN) {
		vp0.Store(testVPT(rand.Uint32()%testMaxHash), testVUintptrT(rand.Intn(testAddN)))
	}
	vp1 := vp0.Copy()
	if vp0.Size() != vp1.Size() {
		t.Fail()
	}
	for k, v := range vp0.Range {
		if a, _ := vp1.Load(k); a != v {
			t.Fail()
		}
	}
	for k, v := range vp1.Range {
		if a, _ := vp0.Load(k); a != v {
			t.Fail()
		}
	}
}
func TestValUintptr_LoadPtr(t *testing.T) {
	vu := NewValUintptr[testVPT, testVUintptrT](testMinBSz, testMaxBSz, 16, testHashF)
	if vu.LoadPtr(0) != nil {
		t.Fail()
	}
	vu.Store(0, 0)
	*vu.LoadPtr(0)++
	if *vu.LoadPtr(0) != 1 {
		t.Fail()
	}
}
