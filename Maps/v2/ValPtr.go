package v2

import (
	"github.com/g-m-twostay/go-utils/Maps/internal"
	"math"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

/*
Linearizability: the effect of all calls can be squashed down to a point.
Sequential consistency: all calls will see the results of all calls that finished before it started. This is a weaker version of linearizability.
*/

type ValPtr[K comparable, V any] struct {
	minAvgBucketSize, maxAvgBucketSize, maxLogChunkSize byte
	firstRelay                                          relay
	size                                                atomic.Uintptr //LS bit is used to indicate whether a resize is happening. Therefore, this should be changed by 2 each time.
	buckets                                             *chunkArr      //bucket referring to ordered linked list as table.
	HashF                                               func(K) uint
}

func NewValPtr[K comparable, V any](minBucketSize, maxBucketSize byte, maxHash uint, hashF func(K) uint) *ValPtr[K, V] {
	vp := ValPtr[K, V]{
		minAvgBucketSize: minBucketSize,
		maxAvgBucketSize: maxBucketSize,
		maxLogChunkSize:  byte(bits.Len(maxHash)),
		HashF:            hashF,
	}
	vp.buckets = newChunkArr(vp.maxLogChunkSize, vp.maxLogChunkSize)
	vp.buckets.first = uintptr(unsafe.Pointer(&vp.firstRelay))
	return &vp
}

type CASResult byte

const (
	resizingMask uintptr = 1
	MaxSize      uint    = math.MaxUint
)
const (
	FAILED  CASResult = iota //indicates that CAS is performed but failed. failure can be caused by 1. the value changed, or 2. the node is deleted within the call's duration.
	SUCCESS                  //indicates that CAS is performed and successes.
	NULL                     //indicates that the key doesn't exist and the CAS isn't performed. A CaD operation on the same key that SUCCESS before the current operation started guarantees NULL while happening at the same time may result in either FAILED or NULL.
)

func (vp *ValPtr[K, V]) trySplit() {
	if size := vp.size.Or(resizingMask); size&resizingMask == 0 { //we acquired the exclusive right to change vp.buckets, so we can read it non-atomically since we know no one else will change it.
		size >>= 1
		if logChunks := vp.maxLogChunkSize - vp.buckets.logChunkSize; vp.buckets.logChunkSize > 0 && byte(size>>logChunks) >= vp.maxAvgBucketSize {
			nb := newChunkArr(vp.maxLogChunkSize, vp.buckets.logChunkSize-1)
			for i := range uint(1) << logChunks {
				left := vp.buckets.Fetch(i)
				nb.set(i<<1, left)
				nr := &relay{hash: i*(1<<vp.buckets.logChunkSize) | 1<<nb.logChunkSize}
				nb.set(i<<1|1, nr)
				path, fb := internal.EvictStack{}, func() *relay {
					return vp.buckets.Fetch(i)
				}
				left, right := left.crawl(&path, fb)
				for nrAddr := unsafe.Pointer(uintptr(unsafe.Pointer(nr)) | relayMask); ; left, right = left.crawl(&path, fb) {
					if rightAddr := addr(right); right == nil || nr.hash <= (*relay)(rightAddr).hash {
						if nr.next = right; left.tryLink(right, nrAddr) {
							break
						}
					} else {
						path.Push(rightAddr)
						left = (*relay)(rightAddr)
					}
				}
			}
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)), unsafe.Pointer(nb))
		}
		vp.size.And(^resizingMask)
	}
}
func (vp *ValPtr[K, V]) tryMerge() {
	if size := vp.size.Or(resizingMask); size&resizingMask == 0 {
		size >>= 1
		b := vp.buckets
		if logChunks := vp.maxLogChunkSize - b.logChunkSize; logChunks > 0 && byte(vp.size.Load()>>logChunks) < vp.minAvgBucketSize {
			nb := newChunkArr(vp.maxLogChunkSize, b.logChunkSize+1)
			for i := range uint(1) << (logChunks - 1) {
				nb.set(i, b.Fetch(i<<1))
			}
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)), unsafe.Pointer(nb))
			for i := uint(1); i < 1<<logChunks; i += 2 {
				b.Fetch(i).mark()
			}
		}
		vp.size.And(^resizingMask)
	}
}

// Size isn't linearizable but is sequential consistent. Calling Size during any Store and Delete calls can result in it returning intermediate values. This is ok when the size of the map is >0 but can cause underflow when the size of map is 0.
func (vp *ValPtr[K, V]) Size() uint {
	return uint(vp.size.Load()) >> 1
}
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
func (vp *ValPtr[K, V]) StorePtr(key K, val /*not nil*/ *V) bool {
	hash := vp.HashF(key)
	var new *ptrNode[K]
	fb, path := func() *relay {
		return (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash)
	}, internal.EvictStack{}
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
func (vp *ValPtr[K, V]) LoadOrStorePtr(key K, val /*not nil*/ *V) *V {
	hash := vp.HashF(key)
	var new *ptrNode[K]
	fb, path := func() *relay {
		return (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash)
	}, internal.EvictStack{}
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
func (vp *ValPtr[K, V]) SwapPtr(key K, val /*not nil*/ *V) *V {
	hash := vp.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return nil
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*ptrNode[K])(curAddr).key == key {
			return (*V)(atomic.SwapPointer(&(*ptrNode[K])(curAddr).val, unsafe.Pointer(val)))
		}
	}
}
func (vp *ValPtr[K, V]) CompareAndSwapPtr(key K, old /*not nil*/, new /*not nil*/ *V) CASResult {
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
func (vp *ValPtr[K, V]) CompareAndSwap(key K, new /*not nil*/ *V, eq func( /*not nil*/ *V) bool) CASResult {
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

func (vp *ValPtr[K, V]) TakePtr() (K, *V) {
	cur := vp.firstRelay.walk()
	for ; isRelay(cur); cur = (*relay)(addr(cur)).walk() {
	}
	if cur == nil {
		return *new(K), nil
	}
	a := (*ptrNode[K])(cur)
	return a.key, (*V)(atomic.LoadPointer(&a.val))
}
func (vp *ValPtr[K, V]) Range(yield func(K /*not nil*/, *V) bool) {
	for cur, curAddr := vp.firstRelay.walk(), (unsafe.Pointer)(nil); cur != nil; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); !isRelay(cur) {
			if a := (*ptrNode[K])(curAddr); !yield(a.key, (*V)(atomic.LoadPointer(&a.val))) {
				break
			}
		}
	}
}
func (vp *ValPtr[K, V]) Copy() *ValPtr[K, V] {
	copied := ValPtr[K, V]{minAvgBucketSize: vp.minAvgBucketSize, maxAvgBucketSize: vp.maxAvgBucketSize, maxLogChunkSize: vp.maxLogChunkSize, buckets: newChunkArr(vp.maxLogChunkSize, (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)))).logChunkSize), HashF: vp.HashF}
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
