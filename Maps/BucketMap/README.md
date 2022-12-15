This implementation has the same underlying sorted linked list structure like ChainMap. However, it has a RWLock on
every relay nodes. This lock is to make sure deletion doesn't happen simultaneously with insertion in a hope to simplify
and speed things up(all the complex logics in ChainMap are caused by deletion).

To be clear, for each bucket(relay), insert operations hold the read lock, delete operations hold to write lock, read
operations doesn't hold the lock. Resizing(rehashing, or shrinking and expanding) are treated as normal insert/deletion
operations, so they don't need extra locks(an important objective when writing this), which is one part why this
implementation is very fast.

Another goal when implementing this is to minimize the amounts of atomic operations by taking advantages of the
additional locks used. We can also search the linked list more efficiently since we know deletion won't happen with
insertion, so we can assume all nodes are valid nodes.