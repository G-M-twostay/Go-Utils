HopMap is a fast, memory-efficient, and cache-friendly general purpose HashMap for single threaded uses. It is based on
the open-addressing collision resolution algorithm hopscotch hashing.

Keys and values are stored in an array of buckets. Each bucket contains only one key value pair and 2 bytes for
hopscotch hashing's parameters. Hash values of keys are lazily kept in an separate array for faster resizing operation.
In general, hopscotch hashing doesn't need hash values to be stored in order to be fast.

It's memory-efficient because everything is kept in several arrays, there is no overflow buckets like the native
implementation. It's cache-friendly because in Hopscotch hashing, all keys with the same hash value are guaranteed to be
close to each other in the bucket array. The speed improvements are likely result of these 2. Resizing operations
will copy all the values into a new array of buckets twice the size, same as native map. 

The usage is similar to native map and exactly the same as sync.Map(except for the concurrent features). There is no
need to provide an external hash function. I exported the native hash functions used by native map, see Map.go.

Below are the benchmark results:

BenchmarkHopMap_Put-16 8257  141115 ns/op 542977 B/op 3 allocs/op
BenchmarkMap_Put-16    6853  177834 ns/op 320539 B/op 9 allocs/op
BenchmarkHopMap_Get-16 14912  70332 ns/op      0 B/op 0 allocs/op
BenchmarkMap_Get-16    9776  126093 ns/op      0 B/op 0 allocs/op
BenchmarkHopMap_Del-16 10000 136786 ns/op      0 B/op 0 allocs/op
BenchmarkMap_Del-16    5731  194555 ns/op      0 B/op 0 allocs/op

Observations:

1. HopMap is faster than native map in all cases.
2. HopMap makes fewer allocations.
3. In the put test, the reason why HopMap took more spaces is that I declared the initial bucket size to be 2 times the
   needed to avoid resizing operations. Native map use chaining therefore won't incur resizing in this case. If the
   initial array size is declared to be similar to native map, then it would take 279680 B/op, which is less than native
   map's size. 
