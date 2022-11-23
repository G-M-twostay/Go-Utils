ChainMap is a fast concurrent hash map implementation.

ChainMap is implemented using a concurrent(lock free) sorted singly linked list. This is the reason that this
implementation can be fast, memory-efficient, and _completely_ lock free.
If `Go` allows double CAS or manual memory management like `C++`, this implementation can be even more efficient. Yes,
implementing ChainMap in `C++` would be way faster than in `Go`.

Here are the benchmarks I did to compare ChainMap to sync.Map and a normal map with RWMutex.

1. This is to mimic the first use case described for sync.Map. Some keys are inserted and read multiple times.
2. This is to mimic the second use case described for sync.Map. Disjoint set of keys are read and write.
3. This is a general case to mimic a concurrent environment where keys are inserted, read, and deleted.

BenchmarkChainMap_Case1-16 1410 768560 ns/op 1045891 B/op 60988 allocs/op
BenchmarkSyncMap_Case1-16 368 3172170 ns/op 1439990 B/op 40749 allocs/op
BenchmarkMutexMap_Case1-16 332 3457266 ns/op 686377 B/op 293 allocs/op
BenchmarkChainMap_Case2-16 7620 194309 ns/op 256552 B/op 32016 allocs/op
BenchmarkSyncMap_Case2-16 1081 1128869 ns/op 258839 B/op 24084 allocs/op
BenchmarkMutexMap_Case2-16 470 2657474 ns/op 659 B/op 16 allocs/op
BenchmarkChainMap_Case3-16 1704 734508 ns/op 1062006 B/op 68149 allocs/op
BenchmarkSyncMap_Case3-16 368 3253068 ns/op 1472779 B/op 40751 allocs/op
BenchmarkMutexMap_Case3-16 195 5962092 ns/op 623006 B/op 306 allocs/op

Observations:

1. Sync.Map is faster than mutex map in the first 2 cases.
2. In all 3 cases, ChainMap is **significantly faster** than the rest.
3. ChainMap uses lots of memory **allocations**, which is true according to the implementation. However, many of these
   are **garbage collected** (many of these are temporary variables), so it can be potentially optimized by using some
   pooling mechanics.
4. On average, ChainMap uses **33% less memory** than Sync.Map. However, a normal map uses the least amount of memory.

