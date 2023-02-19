package HopMap

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

const COUNT int = 8192
const defaultSize = 7 //native map can store 8 elements by default.
const size = 10

func TestHopMap_All(t *testing.T) {
	M := New[int, int](2, 2, 0)
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
		M := New[int, int](16, uint(COUNT), 0)
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

func BenchmarkHopMap_Get(b *testing.B) {
	var M *HopMap[int, int]
	for _t := 0; _t < b.N; _t++ {
		b.StopTimer()
		M = New[int, int](16, uint(COUNT), 0)
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

func BenchmarkHopMap_Del(b *testing.B) {
	var M *HopMap[int, int]
	for _t := 0; _t < b.N; _t++ {
		b.StopTimer()
		M = New[int, int](16, uint(COUNT), 0)
		for i := 0; i < COUNT; i++ {
			M.Store(i, i)
		}
		b.StartTimer()
		for i := 0; i < COUNT; i++ {
			M.LoadAndDelete(i)
		}
		for i := 0; i < COUNT; i++ {
			if M.HasKey(i) {
				b.Error("key exists", i)
			}
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
		for i := 0; i < COUNT; i++ {
			if _, ok := M[i]; ok {
				b.Error("key exists", i)
			}
		}
	}
}

func BenchmarkHopMapPopulate(b *testing.B) {
	for size := 1; size < 1000000; size *= 10 {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m := New[int, bool](16, defaultSize, 0)
				for j := 0; j < size; j++ {
					m.Store(j, true)
				}
			}
		})
	}
}

func BenchmarkMapPopulate(b *testing.B) {
	for size := 1; size < 1000000; size *= 10 {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m := make(map[int]bool)
				for j := 0; j < size; j++ {
					m[j] = true
				}
			}
		})
	}
}

func BenchmarkHopHashStringSpeed(b *testing.B) {
	strings := make([]string, size)
	for i := 0; i < size; i++ {
		strings[i] = fmt.Sprintf("string#%d", i)
	}
	sum := 0
	m := New[string, int](16, size, 0)
	for i := 0; i < size; i++ {
		m.Store(strings[i], 0)
	}
	idx := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t, _ := m.Load(strings[idx])
		sum += t
		idx++
		if idx == size {
			idx = 0
		}
	}
}

func BenchmarkHashStringSpeed(b *testing.B) {
	strings := make([]string, size)
	for i := 0; i < size; i++ {
		strings[i] = fmt.Sprintf("string#%d", i)
	}
	sum := 0
	m := make(map[string]int, size)
	for i := 0; i < size; i++ {
		m[strings[i]] = 0
	}
	idx := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum += m[strings[idx]]
		idx++
		if idx == size {
			idx = 0
		}
	}
}

func BenchmarkHopHashBytesSpeed(b *testing.B) {
	// a bunch of chunks, each with a different alignment mod 16
	type chunk [17]byte
	var chunks [size]chunk
	// initialize each to a different value
	for i := 0; i < size; i++ {
		chunks[i][0] = byte(i)
	}
	// put into a map
	m := New[chunk, int](16, size, 0)
	for i, c := range chunks {
		m.Store(c, i)
	}
	idx := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if t, _ := m.Load(chunks[idx]); t != idx {
			b.Error("bad map entry for chunk")
		}
		idx++
		if idx == size {
			idx = 0
		}
	}
}

func BenchmarkHashBytesSpeed(b *testing.B) {
	// a bunch of chunks, each with a different alignment mod 16
	type chunk [17]byte
	var chunks [size]chunk
	// initialize each to a different value
	for i := 0; i < size; i++ {
		chunks[i][0] = byte(i)
	}
	// put into a map
	m := make(map[chunk]int, size)
	for i, c := range chunks {
		m[c] = i
	}
	idx := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if m[chunks[idx]] != idx {
			b.Error("bad map entry for chunk")
		}
		idx++
		if idx == size {
			idx = 0
		}
	}
}

func BenchmarkHopHashInt32Speed(b *testing.B) {
	ints := make([]int32, size)
	for i := 0; i < size; i++ {
		ints[i] = int32(i)
	}
	sum := 0
	m := New[int32, int](16, size, 0)
	for i := 0; i < size; i++ {
		m.Store(ints[i], 0)
	}
	idx := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t, _ := m.Load(ints[idx])
		sum += t
		idx++
		if idx == size {
			idx = 0
		}
	}
}
func BenchmarkHashInt32Speed(b *testing.B) {
	ints := make([]int32, size)
	for i := 0; i < size; i++ {
		ints[i] = int32(i)
	}
	sum := 0
	m := make(map[int32]int, size)
	for i := 0; i < size; i++ {
		m[ints[i]] = 0
	}
	idx := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum += m[ints[idx]]
		idx++
		if idx == size {
			idx = 0
		}
	}
}

func BenchmarkHopHashInt64Speed(b *testing.B) {
	ints := make([]int64, size)
	for i := 0; i < size; i++ {
		ints[i] = int64(i)
	}
	sum := 0
	m := New[int64, int](16, size, 0)
	for i := 0; i < size; i++ {
		m.Store(ints[i], 0)
	}
	idx := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t, _ := m.Load(ints[idx])
		sum += t
		idx++
		if idx == size {
			idx = 0
		}
	}
}

func BenchmarkHashInt64Speed(b *testing.B) {
	ints := make([]int64, size)
	for i := 0; i < size; i++ {
		ints[i] = int64(i)
	}
	sum := 0
	m := make(map[int64]int, size)
	for i := 0; i < size; i++ {
		m[ints[i]] = 0
	}
	idx := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum += m[ints[idx]]
		idx++
		if idx == size {
			idx = 0
		}
	}
}
func BenchmarkHopHashStringArraySpeed(b *testing.B) {
	stringpairs := make([][2]string, size)
	for i := 0; i < size; i++ {
		for j := 0; j < 2; j++ {
			stringpairs[i][j] = fmt.Sprintf("string#%d/%d", i, j)
		}
	}
	sum := 0
	m := New[[2]string, int](16, size, 0)
	for i := 0; i < size; i++ {
		m.Store(stringpairs[i], 0)
	}
	idx := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t, _ := m.Load(stringpairs[idx])
		sum += t
		idx++
		if idx == size {
			idx = 0
		}
	}
}

func BenchmarkHashStringArraySpeed(b *testing.B) {
	stringpairs := make([][2]string, size)
	for i := 0; i < size; i++ {
		for j := 0; j < 2; j++ {
			stringpairs[i][j] = fmt.Sprintf("string#%d/%d", i, j)
		}
	}
	sum := 0
	m := make(map[[2]string]int, size)
	for i := 0; i < size; i++ {
		m[stringpairs[i]] = 0
	}
	idx := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum += m[stringpairs[idx]]
		idx++
		if idx == size {
			idx = 0
		}
	}
}

// Accessing the same keys in a row.
func benchmarkHopRepeatedLookup(b *testing.B, lookupKeySize int) {
	m := New[string, bool](16, 64, 0)
	// At least bigger than a single bucket:
	for i := 0; i < 64; i++ {
		m.Store(fmt.Sprintf("some key %d", i), true)
	}
	base := strings.Repeat("x", lookupKeySize-1)
	key1 := base + "1"
	key2 := base + "2"
	b.ResetTimer()
	for i := 0; i < b.N/4; i++ {
		_, _ = m.Load(key1)
		_, _ = m.Load(key1)
		_, _ = m.Load(key2)
		_, _ = m.Load(key2)
	}
}

func BenchmarkHopRepeatedLookupStrMapKey32(b *testing.B) { benchmarkHopRepeatedLookup(b, 32) }
func BenchmarkRepeatedLookupStrMapKey32(b *testing.B)    { benchmarkRepeatedLookup(b, 32) }

func BenchmarkHopRepeatedLookupStrMapKey1M(b *testing.B) { benchmarkHopRepeatedLookup(b, 1<<20) }

// Accessing the same keys in a row.
func benchmarkRepeatedLookup(b *testing.B, lookupKeySize int) {
	m := make(map[string]bool)
	// At least bigger than a single bucket:
	for i := 0; i < 64; i++ {
		m[fmt.Sprintf("some key %d", i)] = true
	}
	base := strings.Repeat("x", lookupKeySize-1)
	key1 := base + "1"
	key2 := base + "2"
	b.ResetTimer()
	for i := 0; i < b.N/4; i++ {
		_ = m[key1]
		_ = m[key1]
		_ = m[key2]
		_ = m[key2]
	}
}

func BenchmarkRepeatedLookupStrMapKey1M(b *testing.B) { benchmarkRepeatedLookup(b, 1<<20) }
