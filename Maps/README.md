Thread-safe, high-performance, and simple concurrent hash map implementations. Most importantly, they're correct. Contains a blocking and a non-blocking version.

However, for single threaded case, you are probably better-off using default map as this wasn't the intended use case
for these implementations. In the worst case, these implementations are all very simple and easy to understand(a sorted
linked list with array as indexes), and you can easily modify them.

Changes in the Map are reflected immediately at any time(even during resizing). Unlike some other implementations, if
operation B started at any time after operation A completed(slightly before function call returns), then operation A is guaranteed to be visible to operation
B. In other words, operation A happens-before/synchronizes-before all operations starting after A completes.

ChainMap is more of a demonstration of how a simple completely lock-free hash map is possible in Go. It's slower than
the not completely lock-free implementation Bucketmap and puts a bigger strain on GC. ChainMap is better implemented in
C++ because of manual memory management and the ability to use pointer tagging. In conclusion, use BucketMap; ChainMap
is just a demonstration.

BucketMap is a version of ChainMap that instead uses a lock to ensure deletion and insertion don't
happen simultaneously. This proved to be a very good strategy as more assumptions about the list structure can be made,
which made this map the fastest implementation. It's faster than ChainMap in all cases.

IntMap is a specialized version of BucketMap for all types satisfying comparable. It mainly reduces the comparator
function overhead because we can use "==".

Map.go also includes a general purpose thread-safe hasher for any struct written using hash/maphash(thus it's not
secure). It's a hash function based on memory content, so you should make sure that the memory accessed aren't modified
concurrently. However, I do recommend designing your own hash function if possible as all map implementations are highly
flexible with hash values. It's designed for cases where you are lazy to write the hash function. A optimal hash function should be evenly distributed in [0,2^n).

All these implementations support concurrent expanding/shrinking(rehashing) without complicated logics, as this was one
of the design goal. Previously, I had a Map.Hashable interfacce to handle the hashing and comparing; however, interface
are very slow, so turned to a functional approach. All these maps don't have an internal hash function for rehashing,
only a simple bitmask. So it's important to use a
good hash function yourself. In fact, the performance of all these implementations depends heavily on the hash function.
Also, BucketMap doesn't use the most significant bit of the hash value, so don't use it.

The Map[K,V] interface is for compatibility with sync.Map(so you can switch to mine by changing the name). All my
implementations also implement this interface. ExtendedMap interface is for some additional operations that my
implementations support.

See detailed information under each implementations' own directory. 