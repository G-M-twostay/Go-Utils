package HopMap

import (
	"fmt"
	"testing"
)

const COUNT int = 4096

func TestHopMap_All(t *testing.T) {
	M := New[int, int](4)
	for i := 0; i < 8; i++ {
		M.Put(1+8*i, 1+8*i)
	}
	for i := 0; i < 8; i++ {
		x, y := M.Get(1 + 8*i)
		fmt.Println(x, y)
	}
	//for i := 0; i < 128; i++ {
	//	M.Put(i, i)
	//}
	//for i := 0; i < 128; i++ {
	//	fmt.Println(M.Get(i))
	//}
	fmt.Println(M.bkt)
}

func BenchmarkHopMap_Put(b *testing.B) {
	M := New[int, int](128)
	for _t := 0; _t < b.N; _t++ {
		for i := 0; i < COUNT; i++ {
			M.Put(i, i)
		}
	}
}

func BenchmarkMap_Get(b *testing.B) {
	for _t := 0; _t < b.N; _t++ {
		M := make(map[int]int)
		for i := 0; i < COUNT; i++ {
			M[i] = i
		}
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
out:
	for _t := 0; _t < b.N; _t++ {
		M = New[int, int](128)
		for i := 0; i < 4096; i++ {
			M.Put(i, i)
		}
		for i := 0; i < 4096; i++ {
			x, y := M.Get(i)
			if !y || x != i {
				b.Error("wrong value", i, x)
				a, _ := M.modGet(int(M.hash(i)))
				b.Logf("%v\n", a)
				fmt.Printf("%v\n", M.bkt)
				M.Put(i, -i)
				x, _ = M.Get(i)
				a, _ = M.modGet(int(M.hash(i)))
				b.Logf("%v\n", a)
				fmt.Printf("%v\n", M.bkt)
				break out
			}
		}
	}

}
