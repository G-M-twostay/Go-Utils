package ChainMap

import (
	"GMUtils/Maps"
	"encoding/binary"
	"hash/fnv"
	"hash/maphash"
	"sync"
	"testing"
)

const (
	FNV_PRIME_32        uint   = 16777619
	FNV_PRIME_64        uint64 = 1099511628211
	FNV_OFFSET_BASIS_32 uint   = 2166136261
	FNV_OFFSET_BASIS_64 uint64 = 14695981039346656037
	times                      = 1024
	mapSize                    = 2048
)

var seed maphash.Seed = maphash.MakeSeed()

type KeyObj struct {
	v    int
	hash int
}

func (u KeyObj) Equal(o Maps.Hashable) bool {
	return u.v == o.(KeyObj).v
}

func (u KeyObj) Hash() int {
	return u.hash
}

var hasher = fnv.New64a()

func makeKey(a int) KeyObj {
	b := make([]byte, 8)
	binary.PutVarint(b, int64(a))
	t := KeyObj{}
	t.v = a
	t.hash = int(maphash.Bytes(seed, b))
	return t
}

func BenchmarkChainMap_All(b *testing.B) {
	keys := make([]KeyObj, mapSize*5)
	for i, _ := range keys {
		keys[i] = makeKey(i)
	}

	wg := sync.WaitGroup{}
	M := MakeChainMap[KeyObj, int](1, 4)
	var put func(int, int) = func(low, high int) {
		for ; low < high; low++ {
			M.Put(keys[low], low)
		}
	}
	var remove func(int, int) = func(low, high int) {
		for ; low < high; low++ {
			M.Remove(keys[low])
		}
	}

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		b.StopTimer()

		b.StartTimer()
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(low, high int) {
				put(low, high)
				remove(low, high)
				wg.Done()
			}((i%5)*mapSize, ((i+1)%5)*mapSize)
		}

		wg.Wait()
	}
}

func TestChainMap_All(t *testing.T) {
	keys := make([]KeyObj, mapSize*5)
	for i, _ := range keys {
		keys[i] = makeKey(i)
	}

	wg := sync.WaitGroup{}
	M := MakeChainMap[KeyObj, int](1, 4)
	var put func(int, int) = func(low, high int) {
		for ; low < high; low++ {
			M.Put(keys[low], low)
		}
	}
	var remove func(int, int) = func(low, high int) {
		for ; low < high; low++ {
			M.Remove(keys[low])
		}
	}
	var check func(int, int) = func(low, high int) {
		for ; low < high; low++ {
			if M.HasKey(keys[low]) {
				t.Log("nKeys: ", low)
			}
		}
	}

	for j := 0; j < 1; j++ {
		for k := 0; k < 1; k++ {
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func(low, high int) {
					put(low, high)
					remove(low, high)
					check(low, high)
					wg.Done()
				}((i)*70, (i+1)*70)
			}
		}

		wg.Wait()
		f := M.Pairs()
		t.Log("size: ", M.Size())

		for a, b, ok := f(); ok; {
			t.Log("key: ", a.v, "val: ", b, "hash: ", uint(a.hash))
			a1, b1, ok1 := f()
			//if ok {
			//	t.Log(uint(a1.hash) > uint(a.hash))
			//}
			a, b, ok = a1, b1, ok1
		}
		M.PrintAll()
	}
}

func BenchmarkSyncMap_All(b *testing.B) {
	keys := make([]KeyObj, mapSize*5)
	for i, _ := range keys {
		keys[i] = makeKey(i)
	}
	wg := sync.WaitGroup{}
	M := sync.Map{}
	var put func(int, int) = func(low, high int) {
		for ; low < high; low++ {
			M.Store(keys[low], low)
		}
	}
	var remove func(int, int) = func(low, high int) {
		for ; low < high; low++ {
			M.Delete(keys[low])
		}
	}
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		b.StopTimer()

		b.StartTimer()
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(low, high int) {
				put(low, high)
				remove(low, high)
				wg.Done()
			}((i%5)*mapSize, ((i+1)%5)*mapSize)
		}
		wg.Wait()

	}
}

func BenchmarkLockedMap_All(b *testing.B) {
	keys := make([]KeyObj, mapSize*5)
	for i, _ := range keys {
		keys[i] = makeKey(i)
	}
	wg := sync.WaitGroup{}
	lock := sync.RWMutex{}
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		b.StopTimer()
		M := make(map[KeyObj]int)
		var put func(int, int) = func(low, high int) {
			for ; low < high; low++ {
				lock.Lock()
				M[keys[low]] = low
				lock.Unlock()
			}
		}
		var remove func(int, int) = func(low, high int) {
			for ; low < high; low++ {
				lock.Lock()
				delete(M, keys[low])
				lock.Unlock()
			}
		}
		b.StartTimer()
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(low, high int) {
				put(low, high)
				remove(low, high)
				wg.Done()
			}((i%5)*mapSize, ((i+1)%5)*mapSize)
		}
		wg.Wait()

	}
}
