/*
Package Maps implements linearizable, wait-free general-purpose concurrent hashmap. Multiple variants are offered to take advantage of all supported types in the atomic package.

# Linearizability
The effect of all calls can be squashed down to a point, so there is no intermediate states during operations. Implementations here are nearly 100% linearizable except for size, range, and copy.

# Sequential Consistency
All calls will see the results of all calls that finished before it started. This is a weaker version of linearizability. In go terminology, it's basically the synchronize before thing, so any write operation synchronize before any read operation. All implementations here are sequentially consistent.

# Wait Free
A stronger version of Lock-Free. All operations, regardless of the number of threads calling them, must finish within bounded time and procedures. Wait-free means lock-free. All implementations here are wait-free. A typical example that violates wait-free is spin lock. Basically, wait-free means no busy waiting.

# Usage
It's recommended to use your own hash function whenever possible instead of just using the general hash function offered by go. A good hash function with its lower maxHash bound can increase performance by up to 50%.
Public fields of structs aren't thread-safe to modify and should only be used readonly or modify in synchronized manner.
Unlike usual hashmap implementation where the underlying backing array only expands and never shrinks, these implementations also shrink them when they become sparse enough. Shrinking and expanding happens seamlessly with delete and add operations.
*/
package Maps

// Generates all other ValVal map variants using ValUintptr.go and ValUintptr_test.go as templates.
//go:generate go run gen.go -implTmpl "ValUintptr.go" -testTmpl "ValUintptr_test.go" -- int64 uint64 int32 uint32
import (
	"math"
	"sync/atomic"
	"unsafe"
)

type CASResult byte

const (
	resizingMask uintptr = 1
	MaxSize      uint    = math.MaxInt
)
const (
	FAILED  CASResult = iota //indicates that CAS is performed but failed. failure can be caused by 1. the value changed, or 2. the node is deleted within the call's duration.
	SUCCESS                  //indicates that CAS is performed and successes.
	NULL                     //indicates that the key doesn't exist and the CAS isn't performed. A CaD operation on the same key that SUCCESS before the current operation started guarantees NULL while happening at the same time may result in either FAILED or NULL.
)

// base is the shared parts of ValPtr and all other ValVal maps.
type base[K comparable] struct {
	MinAvgBucketSize, MaxAvgBucketSize, maxLogChunkSize byte
	firstRelay                                          relay
	size                                                atomic.Uintptr //LS bit is used to indicate whether a resize is happening. Therefore, this should be changed by 2 each time.
	buckets                                             *chunkArr      //bucket referring to ordered linked list as table.
	HashF                                               func(K) uint
}

func (vp *base[K]) trySplit() {
	if size := vp.size.Or(resizingMask); size&resizingMask == 0 { //we acquired the exclusive right to change vp.buckets, so we can read it non-atomically since we know no one else will change it.
		size >>= 1
		if logChunks := vp.maxLogChunkSize - vp.buckets.logChunkSize; vp.buckets.logChunkSize > 0 && byte(size>>logChunks) >= vp.MaxAvgBucketSize {
			newBuckets, newRelays := newChunkArr(vp.maxLogChunkSize, vp.buckets.logChunkSize-1), make([]relay, 1<<logChunks)
			for i := range uint(len(newRelays)) {
				left := vp.buckets.Fetch(i)
				newBuckets.set(i<<1, left)
				newRelays[i].hash = i*(1<<vp.buckets.logChunkSize) | 1<<newBuckets.logChunkSize
				newBuckets.set(i<<1|1, &newRelays[i])
				path, fb := evictStack{}, func() *relay {
					return vp.buckets.Fetch(i)
				}
				left, right := left.crawl(&path, fb)
				for nrAddr := unsafe.Pointer(uintptr(unsafe.Pointer(&newRelays[i])) | relayMask); ; left, right = left.crawl(&path, fb) {
					if rightAddr := addr(right); right == nil || newRelays[i].hash <= (*relay)(rightAddr).hash {
						if newRelays[i].next = right; left.tryLink(right, nrAddr) {
							break
						}
					} else {
						path.Push(rightAddr)
						left = (*relay)(rightAddr)
					}
				}
			}
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)), unsafe.Pointer(newBuckets))
		}
		vp.size.And(^resizingMask)
	}
}
func (vp *base[K]) tryMerge() {
	if size := vp.size.Or(resizingMask); size&resizingMask == 0 {
		size >>= 1
		b := vp.buckets
		if logChunks := vp.maxLogChunkSize - b.logChunkSize; logChunks > 0 && byte(size>>logChunks) < vp.MinAvgBucketSize {
			newBuckets := newChunkArr(vp.maxLogChunkSize, b.logChunkSize+1)
			for i := range uint(1) << (logChunks - 1) {
				newBuckets.set(i, b.Fetch(i<<1))
			}
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&vp.buckets)), unsafe.Pointer(newBuckets))
			for i := uint(1); i < 1<<logChunks; i += 2 {
				b.Fetch(i).mark()
			}
		}
		vp.size.And(^resizingMask)
	}
}

// Size isn't linearizable. Calling Size during any Store and Delete calls can result in it returning intermediate values. This isn't a big deal when the size of the map is >0 but can cause underflow when the size of map is 0. As a result, be careful when calling size on a map whose initial size is 0 and while a Store and Delete operation are happening simultaneously.
func (vp *base[K]) Size() uint {
	return uint(vp.size.Load()) >> 1 //LS bit is resizingMask bit.
}
