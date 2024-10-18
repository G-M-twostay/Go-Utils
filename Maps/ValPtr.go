package Maps

import (
	"math/bits"
	"sync/atomic"
	"unsafe"
)

// ValPtr is a map that stores keys by value and values by pointer. Pointers to values can be nil, but isn't suggested.
type ValPtr[K comparable, V any] struct {
	base[K]
}

// NewValPtr is the constructor for ValPtr. maxHash is max{for all a in K | hashF(a)}. Using a tightly bounded maxHash makes the distribution of keys more even and thus speeds up the map. Using a general hash function would require setting maxHash to the appropriate upper bound, likely things like math.MaxUint.
func NewValPtr[K comparable, V any](minBucketSize, maxBucketSize byte, maxHash uint, hashF func(K) uint) *ValPtr[K, V] {
	vp := ValPtr[K, V]{
		base[K]{MinAvgBucketSize: minBucketSize,
			MaxAvgBucketSize: maxBucketSize,
			maxLogChunkSize:  byte(bits.Len(maxHash)),
			HashF:            hashF},
	}
	vp.buckets = newChunkArr(vp.maxLogChunkSize, vp.maxLogChunkSize)
	vp.buckets.first = uintptr(unsafe.Pointer(&vp.firstRelay))
	return &vp
}

// Has reports whether a key is present, regardless of the value.
func (vp *ValPtr[K, V]) Has(key K) bool {
	hash := vp.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return false
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*ptrNode[K])(curAddr).key == key {
			return true
		}
	}
}

// Delete a key from the map, reporting whether it's successful. Delete is successful when the key was present.
func (vp *ValPtr[K, V]) Delete(key K) bool {
	hash := vp.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return false
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*ptrNode[K])(curAddr).key == key {
			if (*relay)(curAddr).mark() {
				vp.size.Add(^uintptr(resizingMask<<1 - 1))
				vp.tryMerge()
				return true
			}
			return false
		}
	}
}

// LoadPtrAndDelete returns the pointer to the value of the deleted key. Returns nil when key isn't found. You can recycle the deleted pointers via a sync.Pool.
func (vp *ValPtr[K, V]) LoadPtrAndDelete(key K) *V {
	hash := vp.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return nil
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*ptrNode[K])(curAddr).key == key {
			if (*relay)(curAddr).mark() {
				vp.size.Add(^uintptr(resizingMask<<1 - 1))
				vp.tryMerge()
				return (*V)(atomic.LoadPointer(&(*ptrNode[K])(curAddr).val)) //val==nil is the same as node not exist to the caller.
			}
			return nil
		}
	}
}

// LoadPtr returns the pointer to the value of a key. Returns nil when key isn't found.
func (vp *ValPtr[K, V]) LoadPtr(key K) *V {
	hash := vp.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return nil
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*ptrNode[K])(curAddr).key == key {
			return (*V)(atomic.LoadPointer(&(*ptrNode[K])(curAddr).val))
		}
	}
}

// StorePtr of the value of a given key.
func (vp *ValPtr[K, V]) StorePtr(key K, val *V) bool {
	hash := vp.HashF(key)
	var new *ptrNode[K]
	fb, path := func() *relay {
		return (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash)
	}, evictStack{}
	for left, right := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).crawl(&path, fb); ; left, right = left.crawl(&path, fb) {
		if rightAddr := addr(right); right == nil || hash < (*relay)(rightAddr).hash {
			/*
				In the extreme case when lots of consecutive unique insertions are to happen, according to benchmarks, up to half of the tryLink calls will fail. This means if we were to allocate on demand, half of the allocated objects are wasted.
				Also according to benchmarks, the cases where an object is allocated but ultimately unused is never encountered, meaning it's extremely rare.
			*/
			if new == nil {
				new = &ptrNode[K]{relay{hash: hash}, unsafe.Pointer(val), key}
			}
			if new.next = right; left.tryLink(right, unsafe.Pointer(new)) {
				vp.size.Add(resizingMask << 1)
				vp.trySplit()
				return true
			}
		} else if (*relay)(rightAddr).hash == hash && !isRelay(right) && (*ptrNode[K])(rightAddr).key == key {
			atomic.StorePointer(&(*ptrNode[K])(rightAddr).val, unsafe.Pointer(val))
			return false
		} else {
			path.Push(rightAddr)
			left = (*relay)(rightAddr)
		}
	}
}

// LoadOrStorePtr stores val to key when key wasn't present and returns nil or returns the pointer to the value corresponding to key.
func (vp *ValPtr[K, V]) LoadOrStorePtr(key K, val *V) *V {
	hash := vp.HashF(key)
	var new *ptrNode[K]
	fb, path := func() *relay {
		return (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash)
	}, evictStack{}
	for left, right := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).crawl(&path, fb); ; left, right = left.crawl(&path, fb) {
		if rightAddr := addr(right); right == nil || hash < (*relay)(rightAddr).hash {
			if new == nil {
				new = &ptrNode[K]{relay{hash: hash}, unsafe.Pointer(val), key}
			}
			if new.next = right; left.tryLink(right, unsafe.Pointer(new)) {
				vp.size.Add(resizingMask << 1)
				vp.trySplit()
				return nil
			}
		} else if (*relay)(rightAddr).hash == hash && !isRelay(right) && (*ptrNode[K])(rightAddr).key == key {
			return (*V)(atomic.LoadPointer(&(*ptrNode[K])(rightAddr).val))
		} else {
			path.Push(rightAddr)
			left = (*relay)(rightAddr)
		}
	}
}

// SwapPtr of a given key. Returns the old value or nil if key wasn't present.
func (vp *ValPtr[K, V]) SwapPtr(key K, val *V) *V {
	hash := vp.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return nil
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*ptrNode[K])(curAddr).key == key {
			return (*V)(atomic.SwapPointer(&(*ptrNode[K])(curAddr).val, unsafe.Pointer(val)))
		}
	}
}

// CompareAndSwapPtr of a given key.
func (vp *ValPtr[K, V]) CompareAndSwapPtr(key K, old, new *V) CASResult {
	hash := vp.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return NULL
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*ptrNode[K])(curAddr).key == key {
			a := atomic.CompareAndSwapPointer(&(*ptrNode[K])(curAddr).val, unsafe.Pointer(old), unsafe.Pointer(new))
			return *(*CASResult)(unsafe.Pointer(&a))
		}
	}
}

// CompareAndSwap value of a given key. That is, set the value to new only when when eq(old)==true.
func (vp *ValPtr[K, V]) CompareAndSwap(key K, new *V, eq func(*V) bool) CASResult {
	hash := vp.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return NULL
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*ptrNode[K])(curAddr).key == key {
			if old := atomic.LoadPointer(&(*ptrNode[K])(curAddr).val); eq((*V)(old)) {
				a := atomic.CompareAndSwapPointer(&(*ptrNode[K])(curAddr).val, old, unsafe.Pointer(new))
				return *(*CASResult)(unsafe.Pointer(&a))
			}
			return FAILED
		}
	}
}

/*
CaD ops implemented by itself like this isn't linerizable. To make them linerizable. First, one would need to change StorePtr to check for nil val. Second, one would need to change CaS to check for nil val. These changes will slightly degrade the performance of both, but the reason that I decided to not make CaD available is because I don't see it as a useful operation.
*/
/*
func (vp *ValPtr[K, V]) ComparePtrAndDelete(key K, old  *V) CASResult {
	hash := vp.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return NULL
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*ptrNode[K])(curAddr).key == key {
			if atomic.CompareAndSwapPointer(&(*ptrNode[K])(curAddr).val, unsafe.Pointer(old), nil) { //because old!=nil, so this will fail if another CaD changed val to nil; only 1 CaD may success on the same key without Store.
				if (*relay)(curAddr).mark() {
					vp.size.And(^uintptr(1 - 1))
					vp.tryMerge()
					return SUCCESS
				}
				return NULL
			}
			return FAILED
		}
	}
}
func (vp *ValPtr[K, V]) CompareAndDelete(key K, eq func(  *V) bool) CASResult {
	hash := vp.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return NULL
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*ptrNode[K])(curAddr).key == key {
			if old := atomic.LoadPointer(&(*ptrNode[K])(curAddr).val); old == nil {
				return NULL //another CaD operation deleted it; only 1 CaD may success on the same key without Store.
			} else if eq((*V)(old)) && atomic.CompareAndSwapPointer(&(*ptrNode[K])(curAddr).val, old, nil) {
				if (*relay)(curAddr).mark() {
					vp.size.And(^uintptr(1 - 1))
					vp.tryMerge()
					return SUCCESS
				}
				return NULL
			}
			return FAILED
		}
	}
}
*/

// TakePtr returns a key value pair from the map that has the smallest hash value for the key. This is designed to replace the patterns
//
//	  m.Range(func(k K, v V) bool {
//						...
//						return false
//					})
//
// and
//
//	  for k,v := range m.Range {
//						...
//						break
//					}
func (vp *ValPtr[K, V]) TakePtr() (*K, *V) {
	cur := vp.firstRelay.walk()
	for ; isRelay(cur); cur = (*relay)(addr(cur)).walk() {
	}
	if cur == nil {
		return nil, nil
	}
	a := (*ptrNode[K])(cur)
	return &a.key, (*V)(atomic.LoadPointer(&a.val))
}

// Range over the key value pairs in the map, stopping when yield returns false. Range isn't linearizable.
func (vp *ValPtr[K, V]) Range(yield func(K, *V) bool) {
	for cur, curAddr := vp.firstRelay.walk(), (unsafe.Pointer)(nil); cur != nil; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); !isRelay(cur) {
			if a := (*ptrNode[K])(curAddr); !yield(a.key, (*V)(atomic.LoadPointer(&a.val))) {
				break
			}
		}
	}
}

// Copy the map. This is faster than adding the keys one by one. Copy isn't linearizable.
func (vp *ValPtr[K, V]) Copy() *ValPtr[K, V] {
	copied := ValPtr[K, V]{base[K]{MinAvgBucketSize: vp.MinAvgBucketSize, MaxAvgBucketSize: vp.MaxAvgBucketSize, maxLogChunkSize: vp.maxLogChunkSize, buckets: newChunkArr(vp.maxLogChunkSize, (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).logChunkSize), HashF: vp.HashF}}
	tail := &copied.firstRelay
	copied.buckets.first = uintptr(unsafe.Pointer(tail))
	tailIndex := uint(0)
	for cur, curAddr := vp.firstRelay.walk(), (unsafe.Pointer)(nil); cur != nil; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); !isRelay(cur) {
			a := (*ptrNode[K])(curAddr)
			if index := copied.buckets.Index(a.hash); index != tailIndex {
				new := &relay{hash: index * (1 << copied.buckets.logChunkSize)}
				tail.next = unsafe.Pointer(uintptr(unsafe.Pointer(new)) | relayMask)
				tail = new
				copied.buckets.set(index, new)
				tailIndex = index
			}
			tail.next = unsafe.Pointer(&ptrNode[K]{relay{hash: a.hash}, atomic.LoadPointer(&(*ptrNode[K])(curAddr).val), a.key})
			tail = (*relay)(tail.next)
			copied.size.Add(resizingMask << 1)
		}
	}
	return &copied
}
