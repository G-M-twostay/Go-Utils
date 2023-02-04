package HopMap

import (
	"fmt"
	"testing"
)

func TestHopMap_All(t *testing.T) {
	M := New[int, int]()
	for i := 0; i < 7; i++ {
		M.Put(1+8*i, 1+8*i)
		x, _ := M.modGet(uint(1 + 8*i))
		fmt.Println(x)
	}
	//M.Put(1, 1)
	//
	//M.Put(9, 9)
	//M.Put(17, 17)
	M.Put(2, 2)
	fmt.Println(M.Get(2))
	fmt.Println(M.buckets)
}
