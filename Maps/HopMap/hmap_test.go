package HopMap

import (
	"testing"
)

const COUNT int = 8192

func TestHopMap_All(t *testing.T) {
	M := New[int, int](2, 2)
	for i := 0; i < 8; i++ {
		M.Store(1+8*i, 1+8*i)
		t.Logf("putted %v\n", 1+8*i)
		t.Log(M.bkt)
	}
	for i := 0; i < 8; i++ {
		x, y := M.Load(1 + 8*i)
		t.Log(x, y)
	}
	t.Log(M.Load(0))
	//for i := 0; i < 128; i++ {
	//	M.Store(i, i)
	//}
	//for i := 0; i < 128; i++ {
	//	fmt.Println(M.Load(i))
	//}

}

func BenchmarkHopMap_Put(b *testing.B) {
	for _t := 0; _t < b.N; _t++ {
		M := New[int, int](uint(COUNT)*2, 16)
		for i := 0; i < COUNT; i++ {
			M.Store(i, i)
		}
	}
}

func BenchmarkMap_Put(b *testing.B) {
	for _t := 0; _t < b.N; _t++ {
		M := make(map[int]int, COUNT)
		for i := 0; i < COUNT; i++ {
			M[i] = i
		}
	}
}

func BenchmarkMap_Get(b *testing.B) {
	for _t := 0; _t < b.N; _t++ {
		b.StopTimer()
		M := make(map[int]int, COUNT)
		for i := 0; i < COUNT; i++ {
			M[i] = i
		}
		b.StartTimer()
		for i := 0; i < COUNT; i++ {
			x := M[i]
			if x != i {
				b.Error("wrong")
			}
		}
	}
}

func BenchmarkHopMap_Get(b *testing.B) {
	var M *HopMap[int, int]
	for _t := 0; _t < b.N; _t++ {
		b.StopTimer()
		M = New[int, int](uint(COUNT), 16)
		for i := 0; i < COUNT; i++ {
			M.Store(i, i)
		}
		b.StartTimer()
		for i := 0; i < COUNT; i++ {
			x, y := M.Load(i)
			if !y || x != i {
				b.Error("wrong value", i, x)
			}
		}
	}
}

func BenchmarkHopMap_Del(b *testing.B) {
	var M *HopMap[int, int]
	for _t := 0; _t < b.N; _t++ {
		b.StopTimer()
		M = New[int, int](uint(COUNT), 16)
		for i := 0; i < COUNT; i++ {
			M.Store(i, i)
		}
		b.StartTimer()
		for i := 0; i < COUNT; i++ {
			M.LoadAndDelete(i)
		}
	}
}

func BenchmarkMap_Del(b *testing.B) {
	for _t := 0; _t < b.N; _t++ {
		b.StopTimer()
		M := make(map[int]int, COUNT)
		for i := 0; i < COUNT; i++ {
			M[i] = i
		}
		b.StartTimer()
		for i := 0; i < COUNT; i++ {
			delete(M, i)
		}
	}
}
