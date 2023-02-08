package HopMap2

import (
	"testing"
)

const COUNT int = 8192
const H_ byte = 128

func TestHopMap_All(t *testing.T) {
	M := New[int, int](4, 4)
	for i := 0; i < 8; i++ {
		M.Put(1+8*i, 1+8*i)
		t.Logf("putted %v\n", 1+8*i)
		//fmt.Println(M.bkt)
	}
	for i := 0; i < 8; i++ {
		x, y := M.Get(1 + 8*i)
		t.Log(x, y)
	}
	M.Put(0, -2)
	t.Log(M.Get(0))
	//for i := 0; i < 128; i++ {
	//	M.Put(i, i)
	//}
	//for i := 0; i < 128; i++ {
	//	fmt.Println(M.Get(i))
	//}
	t.Log(M.String())
}

func BenchmarkHopMap_Put(b *testing.B) {
	M := New[int, int](COUNT, 64)
	for _t := 0; _t < b.N; _t++ {
		for i := 0; i < COUNT; i++ {
			M.Put(i, i)
		}
	}
}

func BenchmarkMap_Get(b *testing.B) {
	for _t := 0; _t < b.N; _t++ {
		//b.StopTimer()
		M := make(map[int]int)
		for i := 0; i < COUNT; i++ {
			M[i] = i
		}
		//b.StartTimer()
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
		//b.StopTimer()
		M = New[int, int](COUNT, H_)
		for i := 0; i < COUNT; i++ {
			M.Put(i, i)
		}
		//b.StartTimer()
		for i := 0; i < COUNT; i++ {
			x, y := M.Get(i)
			if !y || x != i {
				b.Error("wrong value", i, x)
				b.Log(M.String())
				return
			}
		}
	}

}
