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

BenchmarkHopMap_Put-16 8892 135787 ns/op
BenchmarkMap_Put-16 5977 177923 ns/op
BenchmarkHopMap_Get-16 15405 74715 ns/op
BenchmarkMap_Get-16 10000 133596 ns/op
BenchmarkHopMap_Del-16 10000 141037 ns/op
BenchmarkMap_Del-16 7238 191821 ns/op
BenchmarkHopMapPopulate/1-16 11022957 106.2 ns/op 584 B/op 3 allocs/op
BenchmarkHopMapPopulate/10-16 5982703 198.2 ns/op 584 B/op 3 allocs/op
BenchmarkHopMapPopulate/100-16 299104 3857 ns/op 7936 B/op 15 allocs/op
BenchmarkHopMapPopulate/1000-16 34078 35089 ns/op 117296 B/op 27 allocs/op
BenchmarkHopMapPopulate/10000-16 1993 593601 ns/op 1666293 B/op 39 allocs/op
BenchmarkHopMapPopulate/100000-16 195 5818099 ns/op 12794369 B/op 48 allocs/op
BenchmarkMapPopulate/1-16 132243280 9.188 ns/op 0 B/op 0 allocs/op
BenchmarkMapPopulate/10-16 4254435 278.8 ns/op 179 B/op 1 allocs/op
BenchmarkMapPopulate/100-16 281372 4249 ns/op 3348 B/op 17 allocs/op
BenchmarkMapPopulate/1000-16 22122 53811 ns/op 53313 B/op 73 allocs/op
BenchmarkMapPopulate/10000-16 2480 469339 ns/op 427563 B/op 319 allocs/op
BenchmarkMapPopulate/100000-16 229 5214349 ns/op 3617236 B/op 3998 allocs/op
BenchmarkHopHashStringSpeed-16 79769731 15.99 ns/op
BenchmarkHashStringSpeed-16 183372990 6.797 ns/op
BenchmarkHopHashBytesSpeed-16 39801918 25.79 ns/op
BenchmarkHashBytesSpeed-16 100000000 10.47 ns/op
BenchmarkHopHashInt32Speed-16 246852882 4.906 ns/op
BenchmarkHashInt32Speed-16 228438063 5.353 ns/op
BenchmarkHopHashInt64Speed-16 243141238 5.006 ns/op
BenchmarkHashInt64Speed-16 243429778 5.252 ns/op
BenchmarkHopHashStringArraySpeed-16 60006000 17.60 ns/op
BenchmarkHashStringArraySpeed-16 100000000 15.87 ns/op
BenchmarkHopRepeatedLookupStrMapKey32-16 96466232 12.44 ns/op
BenchmarkRepeatedLookupStrMapKey32-16 135220377 8.856 ns/op
BenchmarkHopRepeatedLookupStrMapKey1M-16 100000000 11.24 ns/op
BenchmarkRepeatedLookupStrMapKey1M-16 62787 18616 ns/op

Observations:

1. HopMap has a slow hash function. The native map has a very good and fast hashing strategy, but I don't know how(and
   can't) use it. I blame the `hash/maphash` package for this. I'm using same hashing strategy as
   `hash/maphash`, and this strategy is especially bad at hashing strings or other structs with arbitrary length.
   However,
   when the hashing strategy are the same(4, 8 bytes), than HopMap is faster than native implementation. So, you should
   customize it to your own hash function if you want real performance.
2. HopMap makes fewer allocations.
3. HopMap takes more constant memory than native one, but less memory when the size becomes bigger. Each bucket in
   HopMap is very small, it has the signature of `{K, V, byte, byte}`.
4. In general, because HopMap don't use chaining like native map, under the same data HopMap will use one more resize
   operations, so under dense and clustered data, it will take twice the memory of native map, though the memory is
   highly continuous. However, if the size is same, HopMap takes less memory than native map. In the case of Put test,
   if the initial array size is declared to be similar(both 8192) to native map, then it would take 279680 B/op, which
   is less than native map(321598 B/op).
5. Despite using 1 more resize operation, HopMap still exceeds native map in write speed until when the table size
   becomes to big and the cost of resizing is too big. You can, however, easily customize the expanding strategy of this
   implementation. For example, you can control by how much should the array size increase of if the neighborhood size H
   is increased before expansion is needed. You can also modify HopMap to use chaining when a resizing is determined to
   be needed. This is probably a good strategy, maybe I'll implement it.
6. HopMap stores both key and value by value, meaning it stores a copy of them in the underlying array. It only makes 2
   copies of them: 1 during the call to Store, the second when copying into the array. Values are passed by pointers
   internally to avoid additional copies.
7. Deletion is fast because deletion is done lazily by marking a bucket as empty. This also means that when using
   pointers as values the pointer won't be freed until that bucket is used for other purposes. However, given the
   displacement strategy of Hopscotch hashing, if the table is dense and frequent insertion are happening, that bucket
   is likely to be filled pretty soon.
