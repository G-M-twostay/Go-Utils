This implementation has the same underlying sorted linked list structure like ChainMap. However, it has a RWLock on
every relay nodes. This lock is to make sure deletion doesn't happen simultaneously with insertion in a hope to simplify
and speed things up(all the complex logics in ChainMap are caused by deletion).

To be clear, for each bucket(relay), insert operations hold the read lock, delete operations hold to write lock, read
operations doesn't hold the lock. Resizing(rehashing, or shrinking and expanding) are treated as normal insert/deletion
operations, so they don't need extra locks(an important objective when writing this), which is one part why this
implementation is very fast.

Another goal when implementing this is to minimize the amounts of atomic operations by taking advantages of the
additional locks used. We can also search the linked list more efficiently since we know deletion won't happen with
insertion, so we can assume all nodes are valid nodes. With this in mind, I used the most significant bit of hash value to store the type of node, which made it takes a lot less memory. The reason behind this is that I designed hash value to be uint, but array length is int, which leaves the first bit for me to use. In addition, since the type of nodes don't change, this together with hash value is constant and don't need other synchronization.

Below are the benchmark results:

BenchmarkChainMap_Case1
BenchmarkChainMap_Case1-16          1496            744744 ns/op          918901 B/op      45117 allocs/op
BenchmarkBucketMap_Case1
BenchmarkBucketMap_Case1-16         2552            444687 ns/op          656343 B/op      24618 allocs/op
BenchmarkChainMap_Case2
BenchmarkChainMap_Case2-16          9998            114738 ns/op           66071 B/op       8208 allocs/op
BenchmarkBucketMap_Case2
BenchmarkBucketMap_Case2-16        15654             76757 ns/op           66065 B/op       8208 allocs/op
BenchmarkChainMap_Case3
BenchmarkChainMap_Case3-16          1236            977952 ns/op         1016226 B/op      55603 allocs/op
BenchmarkBucketMap_Case3
BenchmarkBucketMap_Case3-16         1498            781988 ns/op          418960 B/op      18488 allocs/op
PASS

Observations:
1. ChainMap makes double the allocations in case1 because each node also need a state.
2. Access is faster because all nodes are valid and no need to consider deleted nodes.
3. Makes significantly less allocations in case3 as deletions are done by a CAS on immutable states in ChainMap.
4. Faster in all cases.