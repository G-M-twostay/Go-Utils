package v2

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

const (
	testAddN     = 1 << 16
	testAddNEach = 1 << 11
	testThrdsN   = 12
	testMinBSz   = 4
	testMaxBSz   = 8
	testMaxHash  = max(testAddN, testAddNEach*testThrdsN)
)

type testVPT uint32

func testHashF(a testVPT) uint {
	return uint(a)
}
func TestValPtr_LoadOrStore2(t *testing.T) {
	all := make([]testVPT, testAddNEach*testThrdsN)
	for i := range all {
		all[i] = testVPT(i)
	}
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	for i := range testThrdsN {
		go func() {
			for j := i * testAddNEach; j < (i+1)*testAddNEach; j++ {
				mq.LoadOrStorePtr(all[j], &all[j])
				if &all[j] != mq.LoadOrStorePtr(all[j], nil) {
					t.Fail()
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for i := range all {
		av := mq.LoadOrStorePtr(all[i], nil)
		if av != &all[i] {
			t.Fail()
		}
	}
	if mq.Size() != uint(len(all)) {
		t.Fail()
	}
}
func TestValPtr_LoadOrStore3(t *testing.T) {
	all := make([]testVPT, testAddNEach*testThrdsN)
	for i := range all {
		all[i] = testVPT(i)
	}
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	counts := make([]atomic.Uint32, len(all))
	for range testThrdsN {
		go func() {
			for i := range all {
				if mq.LoadOrStorePtr(all[i], &all[i]) == nil {
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
	if mq.Size() != uint(len(all)) {
		t.Fail()
	}
}
func TestValPtr_LoadOrStore1(t *testing.T) {
	std := make(map[testVPT]*testVPT, testAddN/2)
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i := range uint(rand.Intn(testAddN)) {
		k := testVPT(i)
		if mq.LoadOrStorePtr(k, &k) != nil {
			t.Fail()
		}
		if mq.Size() != i+1 {
			t.Fail()
		}
		std[k] = &k
	}
	for k, ev := range std {
		av := mq.LoadOrStorePtr(k, nil)
		if av != ev {
			t.Fatal(av, ev, *ev)
		}
	}
	if mq.Size() != uint(len(std)) {
		t.Fail()
	}
}
func TestValPtr_Load_Store1(t *testing.T) {
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	std := make([]*testVPT, testAddN)
	for i := range std {
		k := testVPT(i)
		std[i] = &k
		if !mq.StorePtr(k, &k) {
			t.Fail()
		}
		if mq.Size() != uint(i)+1 {
			t.Fail()
		}
	}
	for k, ev := range std {
		av := mq.LoadPtr(testVPT(k))
		if av != ev {
			t.Fatal(av, ev, *ev)
		}
	}
	if mq.Size() != uint(len(std)) {
		t.Fail()
	}
}
func TestValPtr_Load_Store2(t *testing.T) {
	all := make([]testVPT, testAddNEach*testThrdsN)
	for i := range all {
		all[i] = testVPT(i)
	}
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	for i := range testThrdsN {
		go func() {
			for j := i * testAddNEach; j < (i+1)*testAddNEach; j++ {
				if !mq.StorePtr(all[j], &all[j]) {
					t.Fail()
				}
				if mq.LoadPtr(all[j]) != &all[j] {
					t.Fail()
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if mq.Size() != uint(len(all)) {
		t.Fail()
	}
	for i := range all {
		av := mq.LoadPtr(all[i])
		if av != &all[i] {
			t.Fail()
		}
	}
}
func TestValPtr_Load_Delete1(t *testing.T) {
	all := make([]testVPT, testAddN)
	for i := range all {
		all[i] = testVPT(i)
	}
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i, k := range all {
		mq.StorePtr(k, &all[i])
	}
	for i := range all {
		mq.LoadPtrAndDelete(all[i])
		if nil != mq.LoadPtr(all[i]) {
			t.Fatal("can't delete")
		}
		if mq.Size() != uint(len(all)-i)-1 {
			t.Fail()
		}
	}
}
func TestValPtr_Load_Delete2(t *testing.T) {
	all := make([]testVPT, testAddNEach*testThrdsN)
	for i := range all {
		all[i] = testVPT(i)
	}
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i, k := range all {
		mq.StorePtr(k, &all[i])
	}
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	for i := range testThrdsN {
		go func() {
			for j := i * testAddNEach; j < (i+1)*testAddNEach; j++ {
				if v := mq.LoadPtr(all[j]); &all[j] != v {
					t.Error("should have", all[j], v)
				}
				mq.LoadPtrAndDelete(all[j])
				if v := mq.LoadPtr(all[j]); v != nil {
					t.Error("can't delete", all[j], v)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
func TestValPtr_Load_Store_Delete(t *testing.T) {
	all := make([]testVPT, testAddNEach*testThrdsN)
	for i := range all {
		all[i] = testVPT(i)
	}
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	for i := range testThrdsN {
		go func() {
			for j := i * testAddNEach; j < (i+1)*testAddNEach; j++ {
				mq.StorePtr(all[j], &all[j])
				if a := mq.LoadPtr(all[j]); &all[j] != a {
					t.Error("didn't store", all[j], a)
				}
				mq.LoadPtrAndDelete(all[j])
				if mq.LoadPtr(all[j]) != nil {
					t.Error("didn't delete", all[j])
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
func TestValPtr_LoadPtrAndDelete1(t *testing.T) {
	all := make([]testVPT, testAddNEach*testThrdsN)
	for i := range all {
		all[i] = testVPT(i)
	}
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i, k := range all {
		mq.StorePtr(k, &all[i])
	}
	for i := range all {
		if &all[i] != mq.LoadPtrAndDelete(all[i]) {
			t.Fatal("wrong delete", mq.LoadPtrAndDelete(all[i]), i)
		}
		if mq.LoadPtrAndDelete(all[i]) != nil {
			t.Fatal("can't delete")
		}
		if mq.Size() != uint(len(all)-i)-1 {
			t.Fail()
		}
	}
}
func TestValPtr_LoadPtrAndDelete2(t *testing.T) {
	all := make([]testVPT, testAddNEach*testThrdsN)
	for i := range all {
		all[i] = testVPT(i)
	}
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i, k := range all {
		mq.StorePtr(k, &all[i])
	}
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	for i := range testThrdsN {
		go func() {
			for j := i * testAddNEach; j < (i+1)*testAddNEach; j++ {
				if v := mq.LoadPtrAndDelete(all[j]); &all[j] != v {
					t.Error("wrong delete", all[j], v)
				}
				if mq.LoadPtrAndDelete(all[j]) != nil {
					t.Error("can't delete")
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
func TestValPtr_LoadPtrAndDelete3(t *testing.T) {
	all := make([]testVPT, testAddNEach*testThrdsN)
	for i := range all {
		all[i] = testVPT(i)
	}
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i, k := range all {
		mq.StorePtr(k, &all[i])
	}
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	count := make([]atomic.Uint32, len(all))
	for range testThrdsN {
		go func() {
			for i := range all {
				if mq.LoadPtrAndDelete(all[i]) != nil {
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
func TestValPtr_SwapPtr(t *testing.T) {
	vp := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, 16, testHashF)
	if vp.SwapPtr(0, nil) != nil {
		t.Fail()
	}
	v1, v2 := testVPT(0), testVPT(1)
	vp.StorePtr(0, &v1)
	if *vp.SwapPtr(0, &v2) != v1 {
		t.Fail()
	}
	if *vp.LoadPtr(0) != v2 {
		t.Fail()
	}
}
func TestValPtr_LoadOrStore_Delete(t *testing.T) {
	all := make([]testVPT, testAddNEach*testThrdsN)
	for i := range all {
		all[i] = testVPT(i)
	}
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	wg := sync.WaitGroup{}
	wg.Add(testThrdsN)
	for i := range testThrdsN {
		go func() {
			for j := i * testAddNEach; j < (i+1)*testAddNEach; j++ {
				mq.LoadOrStorePtr(all[j], &all[j])
				if &all[j] != mq.LoadOrStorePtr(all[j], nil) {
					t.Error("can't store", all[j])
				}
				if &all[j] != mq.LoadPtrAndDelete(all[j]) {
					t.Fail()
				}
				if mq.LoadPtrAndDelete(all[j]) != nil {
					t.Fail()
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
func TestValPtr_CompareAndSwapPtr(t *testing.T) {
	vp := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, 16, testHashF)
	if vp.CompareAndSwapPtr(0, nil, nil) != NULL {
		t.Fail()
	}
	vs := make([]testVPT, 5)
	vp.StorePtr(0, &vs[0])
	results := make([]bool, 4)
	for range rand.Intn(testAddN) {
		wg := sync.WaitGroup{}
		wg.Add(4)
		go func() {
			if a := vp.CompareAndSwapPtr(0, &vs[0], &vs[1]); a == NULL {
				t.Fail()
			} else {
				results[0] = a == SUCCESS
			}
			wg.Done()
		}()
		go func() {
			if a := vp.CompareAndSwapPtr(0, &vs[0], &vs[4]); a == NULL {
				t.Fail()
			} else {
				results[3] = a == SUCCESS
			}
			wg.Done()
		}()
		go func() {
			if a := vp.CompareAndSwapPtr(0, &vs[1], &vs[2]); a == NULL {
				t.Fail()
			} else {
				results[1] = a == SUCCESS
			}
			wg.Done()
		}()
		go func() {
			if a := vp.CompareAndSwapPtr(0, &vs[1], &vs[3]); a == NULL {
				t.Fail()
			} else {
				results[2] = a == SUCCESS
			}
			wg.Done()
		}()
		wg.Wait()
		vp.StorePtr(0, &vs[0])
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
func TestValPtr_CompareAndSwap(t *testing.T) {
	vp := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, 16, testHashF)
	if vp.CompareAndSwap(0, nil, func(*testVPT) bool { return true }) != NULL {
		t.Fail()
	}
	vs := make([]testVPT, 5)
	for i := range vs {
		vs[i] = testVPT(i)
	}
	vp.StorePtr(0, &vs[0])
	results := make([]bool, 4)
	for range rand.Intn(testAddN) {
		wg := sync.WaitGroup{}
		wg.Add(4)
		go func() {
			if a := vp.CompareAndSwap(0, &vs[1], func(v *testVPT) bool {
				return *v == vs[0]
			}); a == NULL {
				t.Fail()
			} else {
				results[0] = a == SUCCESS
			}
			wg.Done()
		}()
		go func() {
			if a := vp.CompareAndSwap(0, &vs[4], func(v *testVPT) bool {
				return *v == vs[0]
			}); a == NULL {
				t.Fail()
			} else {
				results[3] = a == SUCCESS
			}
			wg.Done()
		}()
		go func() {
			if a := vp.CompareAndSwap(0, &vs[2], func(v *testVPT) bool {
				return *v == vs[1]
			}); a == NULL {
				t.Fail()
			} else {
				results[1] = a == SUCCESS
			}
			wg.Done()
		}()
		go func() {
			if a := vp.CompareAndSwap(0, &vs[3], func(v *testVPT) bool {
				return *v == vs[1]
			}); a == NULL {
				t.Fail()
			} else {
				results[2] = a == SUCCESS
			}
			wg.Done()
		}()
		wg.Wait()
		vp.StorePtr(0, &vs[0])
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

/*
	func TestValPtr_ComparePtrAndDelete(t *testing.T) {
		vp := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, 16, testHashF)
		if vp.ComparePtrAndDelete(0, nil) != NULL {
			t.Fail()
		}
		vs := make([]testVPT, 5)
		results := make([]CASResult, 4)
		for range rand.Intn(testAddN) {
			vp.StorePtr(0, &vs[0])
			wg := sync.WaitGroup{}
			wg.Add(3)
			go func() {
				results[0] = vp.ComparePtrAndDelete(0, &vs[0])
				wg.Done()
			}()
			go func() {
				results[1] = vp.ComparePtrAndDelete(0, &vs[0])
				wg.Done()
			}()
			go func() {
				if vp.StorePtr(0, &vs[1]) {
					results[2] = NULL
				} else {
					results[2] = SUCCESS
				}
				wg.Done()
			}()
			wg.Wait()
			if results[2] == NULL {
				if (results[0]+results[1])&1 == 0 {
					t.Fatal("0 and 1 should contain 1 NULL/FAILED and 1 SUCCESS", results)
				}
			} else {
				if results[0] == SUCCESS && results[1] == SUCCESS {
					t.Fatal("0 and 1 mustn't all SUCCESS if node is changed.", results)
				}
			}
		}
	}

	func TestValPtr_CompareAndDelete(t *testing.T) {
		vp := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, 16, testHashF)
		if vp.CompareAndDelete(0, func(*testVPT) bool { return true }) != NULL {
			t.Fail()
		}
		vs := make([]testVPT, 5)
		for i := range vs {
			vs[i] = testVPT(i)
		}
		results := make([]CASResult, 4)
		for range rand.Intn(testAddN) {
			wg := sync.WaitGroup{}
			wg.Add(3)
			vp.StorePtr(0, &vs[0])
			go func() {
				results[0] = vp.CompareAndDelete(0, func(val *testVPT) bool { return *val == vs[0] })
				wg.Done()
			}()
			go func() {
				results[1] = vp.CompareAndDelete(0, func(val *testVPT) bool { return *val == vs[0] })
				wg.Done()
			}()
			go func() {
				if vp.StorePtr(0, &vs[1]) {
					results[2] = NULL
				} else {
					results[2] = SUCCESS
				}
				wg.Done()
			}()
			wg.Wait()
			if results[2] == NULL {
				if (results[0]+results[1])&1 == 0 {
					t.Fatal("0 and 1 should contain 1 NULL/FAILED and 1 SUCCESS", results)
				}
			} else {
				if results[0] == SUCCESS && results[1] == SUCCESS {
					t.Fatal("0 and 1 mustn't all SUCCESS if node is changed.", results)
				}
			}
		}
	}
*/
func TestValPtr_TakePtr(t *testing.T) {
	vp := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, 16, testHashF)
	if _, v := vp.TakePtr(); v != nil {
		t.Fail()
	}
	a := testVPT(15)
	vp.StorePtr(15, &a)
	if _, v := vp.TakePtr(); v != &a {
		t.Fail()
	}
	b := testVPT(0)
	vp.StorePtr(0, &b)
	if _, v := vp.TakePtr(); v != &b {
		t.Fail()
	}
}
func TestValPtr_Range(t *testing.T) {
	all := make([]testVPT, testAddN)
	for i := range all {
		all[i] = testVPT(i)
	}
	mq := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for i, k := range all {
		mq.StorePtr(k, &all[i])
	}
	count := 0
	for k, v := range mq.Range {
		if k != all[count] {
			t.Fail()
		}
		if v != &all[count] {
			t.Fail()
		}
		count++
	}
}
func TestValPtr_Copy(t *testing.T) {
	vp0 := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	for range rand.Intn(testAddN) {
		vp0.StorePtr(testVPT(rand.Uint32()%testMaxHash), new(testVPT))
	}
	vp1 := vp0.Copy()
	if vp0.Size() != vp1.Size() {
		t.Fail()
	}
	for k, v := range vp0.Range {
		if vp1.LoadPtr(k) != v {
			t.Fail()
		}
	}
	for k, v := range vp1.Range {
		if vp0.LoadPtr(k) != v {
			t.Fail()
		}
	}
}
func TestValPtr_Delete(t *testing.T) {
	vp := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	if vp.Delete(testVPT(rand.Intn(testMaxHash))) {
		t.Fail()
	}
	a := testVPT(rand.Uint32() % (testMaxHash - 1))
	vp.StorePtr(a, new(testVPT))
	vp.StorePtr(a+1, nil)
	if !vp.Delete(a) {
		t.Fail()
	}
	if !vp.Delete(a + 1) {
		t.Fail()
	}
	if vp.Delete(a) {
		t.Fail()
	}
	if vp.Delete(a + 1) {
		t.Fail()
	}
}
func TestValPtr_Has(t *testing.T) {
	vp := NewValPtr[testVPT, testVPT](testMinBSz, testMaxBSz, testMaxHash, testHashF)
	if vp.Has(testVPT(rand.Intn(testMaxHash))) {
		t.Fail()
	}
	a := testVPT(rand.Uint32() % (testMaxHash - 1))
	vp.StorePtr(a, new(testVPT))
	vp.StorePtr(a+1, nil)
	if !vp.Has(a) {
		t.Fail()
	}
	if !vp.Has(a + 1) {
		t.Fail()
	}
}
func TestValPtr_invalidSplit(t *testing.T) { //don't split into chunks more than maxHash allowed.
	maxHash := uint(4)
	vp := NewValPtr[testVPT, testVPT](testMinBSz, 0, maxHash, func(v testVPT) uint {
		return uint(v) % maxHash
	})
	for i := range testVPT(maxHash) * 10 {
		vp.StorePtr(i, &i)
	}
	for i := range testVPT(maxHash) * 10 {
		vp.Delete(i)
	}
}
