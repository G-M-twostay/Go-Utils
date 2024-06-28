package main

import (
	"fmt"
	"github.com/g-m-twostay/go-utils/Trees/arrTree"
	"math"
	"math/bits"
	"math/rand"
	"slices"
	"testing"
)

var (
	bAddN uint32 = 1000000
	bRmvN uint32 = bAddN
	bQryN uint32 = bRmvN
)
var _R rand.Rand = *rand.New(rand.NewSource(0))

func create(b *testing.B, all []int) *Trees.SBTree[int, uint32] {
	b.Helper()
	tree := Trees.New[int, uint32](bAddN)
	buf := make([]uintptr, bits.Len32(bAddN))
	for range bAddN {
		a := _R.Int()
		_, buf = tree.BufferedInsert(a, buf[:0])
		all = append(all, a)
	}
	return tree
}

var __r1 bool

func BenchmarkDelQry(b *testing.B) {
	all := make([]int, bAddN)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		tree := *create(b, all[:0])
		m := slices.Max(all[bRmvN:])
		b.StartTimer()
		var buf []uint32
		for _, v := range all[bRmvN:] {
			_, buf = tree.BufferedRemove(v, buf[:0])
		}
		for _, v := range all[:bRmvN] {
			__r1 = tree.Has(v)
		}
		for range bQryN {
			__r1 = tree.Has(_R.Intn(m))
		}
	}
}

const bNumSteps uint32 = 50

func main() {
	testing.Init()
	var cs []float64
	var N uint16
	for i := uint32(1); i < bNumSteps; i++ {
		bRmvN = bAddN / bNumSteps * i
		bQryN = bRmvN
		br := testing.Benchmark(BenchmarkDelQry)
		cs = append(cs, float64(br.T.Milliseconds()))
		N += uint16(br.N)
		fmt.Println(i)
	}
	var sum float64 = 0
	for _, v := range cs {
		sum += v
	}
	avg := sum / float64(N)
	fmt.Printf("average: %fms/op\n", avg)
	sum = 0
	for _, v := range cs {
		a := v - avg
		sum += a * a
	}
	fmt.Printf("stddev: %fms/op\n", math.Sqrt(sum/float64(N)))
}
