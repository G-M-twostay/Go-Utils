This implementation is extremely fast, I compared it wo some other implementations:


BenchmarkReadHashMapUint-16                      1496050               729.5 ns/ op             0 B/op          0 allocs/op

BenchmarkReadBMapUint-16                         1352514               886.7 ns/ op             0 B/op          0 allocs/op

BenchmarkReadIntMapUint-16                       1939606               620.1 ns/ op             0 B/op          0 allocs/op

BenchmarkReadHaxMapUint-16                       1693029               702.5 ns/ op             0 B/op          0 allocs/op

BenchmarkReadHashMapWithWritesUint-16            1547737               784.0 ns/ op            50 B/op          6 allocs/op

BenchmarkReadBMapWithWritesUint-16               1295151               899.2 ns/ op            50 B/op          6 allocs/op

BenchmarkReadIntMapWithWritesUint-16             1804292               649.5 ns/ op            54 B/op          6 allocs/op

BenchmarkReadHaxMapWithWritesUint-16             1530337               784.0 ns/ op            48 B/op          6 allocs/op

BenchmarkWriteHashMapUint-16                       48426             24823 ns/op             8193 B/op       1024 allocs/op

BenchmarkWriteBMapUint-16                          49976             24101 ns/op             8193 B/op       1024 allocs/op

BenchmarkWriteIntMapUint-16                        52161             23081 ns/op             8193 B/op       1024 allocs/op

BenchmarkWriteHaxMapUint-16                        43310             27194 ns/op             8193 B/op       1024 allocs/op

See comparisons/cmp1_test.go for detailed information.

This implementation has the same underlying sorted linked list structure like ChainMap. However, it has a RWLock on
every relay nodes. This lock is to make sure deletion doesn't happen simultaneously with insertion in a hope to simplify
and speed things up(all the complex logics in ChainMap are caused by deletion).

Details:
1. Each bucket(segment) has its own RWMutex. No global lock.
2. Read operations are totally non-blocking, lock-free, with no busy waiting.
3. Insert operations hold read lock on its bucket. Then, it traverses the bucket at most once and tries to use CAS to insert the node.
4. Deletion operations hold write lock on its bucket. Then, it deletes the node in one attempt.
5. Expanding the map is achieved by splitting each bucket into 2. This is treated as a series(order doesn't matter) of Insert operations.
6. Shrinking the map is achieved by merging consecutive paris({(0,1),(2,3),...}) of buckets into 1. For each pair, it's treated as a insert on the first bucket and delete on the second bucket. Order of pairs doesn't matter.

Another goal when implementing this is to minimize the amounts of atomic operations by taking advantages of the
additional locks used. We can also search the linked list more efficiently since we know deletion won't happen with
insertion, so we can assume all nodes are valid nodes. With this in mind, I used the most significant bit of hash value to store the type of node, which made it takes a lot less memory. The reason behind this is that I designed hash value to be uint, but array length is int, which leaves the first bit for me to use. In addition, since the type of nodes don't change, this together with hash value is constant and don't need other synchronization.

Below are the benchmark results:


BenchmarkChainMap_Case1-16          1496            744744 ns/op          918901 B/op      45117 allocs/op

BenchmarkBucketMap_Case1-16         2552            444687 ns/op          656343 B/op      24618 allocs/op

BenchmarkChainMap_Case2-16          9998            114738 ns/op           66071 B/op       8208 allocs/op

BenchmarkBucketMap_Case2-16        15654             76757 ns/op           66065 B/op       8208 allocs/op

BenchmarkChainMap_Case3-16          1236            977952 ns/op         1016226 B/op      55603 allocs/op

BenchmarkBucketMap_Case3-16         1498            781988 ns/op          418960 B/op      18488 allocs/op

Observations:
1. ChainMap makes double the allocations in case1 because each node also need a state.
2. Access is faster because all nodes are valid and no need to consider deleted nodes.
3. Makes significantly less allocations in case3 as deletions are done by a CAS on immutable states in ChainMap.
4. Faster in all cases.