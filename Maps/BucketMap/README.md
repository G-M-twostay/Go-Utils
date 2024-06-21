This implementation is extremely fast, I compared it wo some other implementations:
Benchmark1ReadHashMapUint
Benchmark1ReadHashMapUint-8             	  507007	      2090 ns/op	       0 B/op	       0 allocs/op
Benchmark1ReadBMapUint
Benchmark1ReadBMapUint-8                	  666699	      1960 ns/op	       0 B/op	       0 allocs/op
Benchmark1ReadIntMapUint
Benchmark1ReadIntMapUint-8              	  863122	      1345 ns/op	       0 B/op	       0 allocs/op
Benchmark1ReadHaxMapUint
Benchmark1ReadHaxMapUint-8              	  571471	      2156 ns/op	       0 B/op	       0 allocs/op
Benchmark1ReadHashMapWithWritesUint
Benchmark1ReadHashMapWithWritesUint-8   	  484990	      2581 ns/op	     255 B/op	      31 allocs/op
Benchmark1ReadBMapWithWritesUint
Benchmark1ReadBMapWithWritesUint-8      	  550442	      2243 ns/op	     249 B/op	      31 allocs/op
Benchmark1ReadIntMapWithWritesUint
Benchmark1ReadIntMapWithWritesUint-8    	  753247	      1661 ns/op	     206 B/op	      25 allocs/op
Benchmark1ReadHaxMapWithWritesUint
Benchmark1ReadHaxMapWithWritesUint-8    	  444886	      2454 ns/op	     227 B/op	      28 allocs/op
Benchmark1WriteHashMapUint
Benchmark1WriteHashMapUint-8            	   30560	     38657 ns/op	    8194 B/op	    1024 allocs/op
Benchmark1WriteBMapUint
Benchmark1WriteBMapUint-8               	   28940	     40726 ns/op	    8193 B/op	    1024 allocs/op
Benchmark1WriteIntMapUint
Benchmark1WriteIntMapUint-8             	   30994	     38416 ns/op	    8193 B/op	    1024 allocs/op
Benchmark1WriteHaxMapUint
Benchmark1WriteHaxMapUint-8             	   28213	     42924 ns/op	    8195 B/op	    1024 allocs/op

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