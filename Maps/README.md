All the Map implementations are intended for concurrent uses, they shouldn't be used in single-threaded cases. ChainMap and BucketMap 
are completed and ready to use. 

PoolMap is a version of ChainMap using sync.Pool for states. It's not completed
and only a test. 

SpinMap is ChainMap but instead of using states uses a lock on all nodes. It's not completed but the
written functions are ready to use. It's faster than ChainMap in case 1 and 3, a bit slower in case 2. The significant
part is the less memory usage and fewer allocations. SpinMap is only good for high-end multicore CPU, otherwise it's
slower than ChainMap. 

BucketMap is a version of ChainMap that instead uses a lock to ensure deletion and insertion don't
happen simultaneously. This proved to be a very good strategy as more assumptions about the list structure can be made,
which made this map the fastest implementation. It's faster than ChainMap in all cases. Takes less memory(slightly more
than SpinMap) and makes fewer allocations(slightly more than SpinMap).

All these implementations support concurrent expanding/shrinking(rehashing) without complicated logics, as this was one
of the design goal. However, I used interface for handling the hash part, which is probably not a very good practice,
which also means that the performance can be further improved(by quite a decent amount) by using a non-interface
approach.

All these maps don't have an internal hash function for rehashing, only a simply bitmask. So it's important to use a
good hash function yourself. In fact, the performance of all these implementations depends heavily on the hash function.

The Map[K,V] interface is for compatibility with sync.Map(so you can switch to mine by changing the name). All my implementations also implement this interface. ExtendedMap interface is for some additional operations that my implementations support.