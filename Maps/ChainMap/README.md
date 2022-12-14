ChainMap is a fast concurrent hash map implementation.

ChainMap is implemented using a concurrent(lock free) sorted singly linked list. This is the reason that this
implementation can be _completely_ lock free without complicated logics.
If `Go` allows double CAS or manual memory management like `C++`, this implementation can be even more efficient. Yes,
implementing ChainMap in `C++` would be way faster than in `Go`.

Details:

1. There is completely lock-free and non-blocking.
2. The linked list is implemented using https://www.cl.cam.ac.uk/research/srg/netos/papers/2001-caslists.pdf by Timothy L. Harris.
3. Read operations are completely non-blocking with no busy waiting.
4. Expanding the map is treated as a series of insertion operation.
5. Shrinking the map is treated as a series of deletion operation.

Here are the benchmarks I did to compare ChainMap to sync.Map and a normal map with RWMutex. These test cases are
inspired by https://github.com/dustinxie/lockfree/blob/master/map_test.go.

1. This is to mimic the first use case described for sync.Map. Some keys are inserted and read multiple times.
2. This is to mimic the second use case described for sync.Map. Disjoint set of keys are read and write.
3. This is a general case to mimic a concurrent environment where keys are inserted, read, and deleted.


BenchmarkChainMap_Case1-16 1528  748276 ns/op  918799 B/op 45113 allocs/op

BenchmarkSyncMap_Case1-16   370 3212790 ns/op 1438684 B/op 40750 allocs/op

BenchmarkMutexMap_Case1-16  337 3540829 ns/op  686400 B/op   293 allocs/op

BenchmarkChainMap_Case2-16 9228  113376 ns/op   66072 B/op  8208 allocs/op

BenchmarkSyncMap_Case2-16  1004 1147019 ns/op  258789 B/op 24083 allocs/op

BenchmarkMutexMap_Case2-16  450 2665258 ns/op     644 B/op    16 allocs/op

BenchmarkChainMap_Case3-16 1230  999950 ns/op 1016945 B/op 55616 allocs/op

BenchmarkSyncMap_Case3-16   364 3304403 ns/op 1461508 B/op 40750 allocs/op

BenchmarkMutexMap_Case3-16  188 6297385 ns/op  622392 B/op   306 allocs/op

Observations:

1. Sync.Map is faster than mutex map in the first 2 cases.
2. In all 3 cases, ChainMap is **significantly faster** than the rest.
3. ChainMap uses lots of memory **allocations**, which is true according to the implementation. However, many of these
   are **garbage collected** (many of these are temporary variables), so it can be potentially optimized by using some
   pooling mechanics.
4. On average, ChainMap uses **33% less memory** than Sync.Map. However, a normal map uses the least amount of memory.

