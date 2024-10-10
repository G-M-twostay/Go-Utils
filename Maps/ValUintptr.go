package Maps

import (
	"math/bits"
	"sync/atomic"
	"unsafe"
)

type ValUintptr[K comparable, V ~uintptr | ~uint | ~int] struct {
	base[K]
}

func NewValUintptr[K comparable, V ~uintptr | ~uint | ~int](minBucketSize, maxBucketSize byte, maxHash uint, hashF func(K) uint) *ValUintptr[K, V] {
	vp := ValUintptr[K, V]{
		base[K]{minAvgBucketSize: minBucketSize,
			maxAvgBucketSize: maxBucketSize,
			maxLogChunkSize:  byte(bits.Len(maxHash)),
			HashF:            hashF},
	}
	vp.buckets = newChunkArr(vp.maxLogChunkSize, vp.maxLogChunkSize)
	vp.buckets.set(0, &vp.firstRelay)
	return &vp
}

func (vv *ValUintptr[K, V]) LoadAndDelete(key K) (V, bool) {
	hash := vv.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vv.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return 0, false
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*valNode[K, uintptr])(curAddr).key == key {
			if (*relay)(curAddr).mark() {
				vv.size.Add(^uintptr(resizingMask<<1 - 1))
				vv.tryMerge()
				return V(atomic.LoadUintptr(&(*valNode[K, uintptr])(curAddr).val)), true
			}
			return 0, false
		}
	}
}
func (vv *ValUintptr[K, V]) Load(key K) (V, bool) {
	hash := vv.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vv.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return 0, false
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*valNode[K, uintptr])(curAddr).key == key {
			return V(atomic.LoadUintptr(&(*valNode[K, uintptr])(curAddr).val)), true
		}
	}
}
func (vv *ValUintptr[K, V]) LoadPtr(key K) *V {
	hash := vv.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vv.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return nil
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*valNode[K, uintptr])(curAddr).key == key {
			return (*V)(unsafe.Pointer(&(*valNode[K, uintptr])(curAddr).val))
		}
	}
}
func (vv *ValUintptr[K, V]) Store(key K, val V) bool {
	hash := vv.HashF(key)
	var new *valNode[K, uintptr]
	fb, path := func() *relay {
		return (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vv.buckets)))).Get(hash)
	}, evictStack{}
	for left, right := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vv.buckets)))).Get(hash).crawl(&path, fb); ; left, right = left.crawl(&path, fb) {
		if rightAddr := addr(right); right == nil || hash < (*relay)(rightAddr).hash {
			if new == nil {
				new = &valNode[K, uintptr]{relay{hash: hash}, key, uintptr /*typeCast*/ (val)}
			}
			if new.next = right; left.tryLink(right, unsafe.Pointer(new)) {
				vv.size.Add(resizingMask << 1)
				vv.trySplit()
				return true
			}
		} else if (*relay)(rightAddr).hash == hash && !isRelay(right) && (*valNode[K, uintptr])(rightAddr).key == key {
			atomic.StoreUintptr(&(*valNode[K, uintptr])(rightAddr).val, uintptr /*typeCast*/ (val))
			return false
		} else {
			path.Push(rightAddr)
			left = (*relay)(rightAddr)
		}
	}
}
func (vv *ValUintptr[K, V]) LoadOrStore(key K, val V) (V, bool) {
	hash := vv.HashF(key)
	var new *valNode[K, uintptr]
	fb, path := func() *relay {
		return (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vv.buckets)))).Get(hash)
	}, evictStack{}
	for left, right := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vv.buckets)))).Get(hash).crawl(&path, fb); ; left, right = left.crawl(&path, fb) {
		if rightAddr := addr(right); right == nil || hash < (*relay)(rightAddr).hash {
			if new == nil {
				new = &valNode[K, uintptr]{relay{hash: hash}, key, uintptr /*typeCast*/ (val)}
			}
			if new.next = right; left.tryLink(right, unsafe.Pointer(new)) {
				vv.size.Add(resizingMask << 1)
				vv.trySplit()
				return 0, false
			}
		} else if (*relay)(rightAddr).hash == hash && !isRelay(right) && (*valNode[K, uintptr])(rightAddr).key == key {
			return V(atomic.LoadUintptr(&(*valNode[K, uintptr])(rightAddr).val)), true
		} else {
			path.Push(rightAddr)
			left = (*relay)(rightAddr)
		}
	}
}
func (vv *ValUintptr[K, V]) Swap(key K, val V) (V, bool) {
	hash := vv.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vv.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return 0, false
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*valNode[K, uintptr])(curAddr).key == key {
			return V(atomic.SwapUintptr(&(*valNode[K, uintptr])(curAddr).val, uintptr /*typeCast*/ (val))), true
		}
	}
}
func (vv *ValUintptr[K, V]) CompareAndSwap(key K, old, new V) CASResult {
	hash := vv.HashF(key)
	for cur, curAddr := (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vv.buckets)))).Get(hash).walk(), unsafe.Pointer(nil); ; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); cur == nil || hash < (*relay)(curAddr).hash {
			return NULL
		} else if (*relay)(curAddr).hash == hash && !isRelay(cur) && (*valNode[K, uintptr])(curAddr).key == key {
			a := atomic.CompareAndSwapUintptr(&(*valNode[K, uintptr])(curAddr).val, uintptr /*typeCast*/ (old), uintptr /*typeCast*/ (new))
			return *(*CASResult)(unsafe.Pointer(&a))
		}
	}
}

func (vv *ValUintptr[K, V]) Take() (*K, V) {
	cur := vv.firstRelay.walk()
	for ; isRelay(cur); cur = (*relay)(addr(cur)).walk() {
	}
	if cur == nil {
		return nil, 0
	}
	a := (*valNode[K, uintptr])(cur)
	return &a.key, V(atomic.LoadUintptr(&a.val))
}
func (vv *ValUintptr[K, V]) Range(yield func(K, V) bool) {
	for cur, curAddr := vv.firstRelay.walk(), (unsafe.Pointer)(nil); cur != nil; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); !isRelay(cur) {
			if a := (*valNode[K, uintptr])(curAddr); !yield(a.key, V(atomic.LoadUintptr(&a.val))) {
				break
			}
		}
	}
}
func (vv *ValUintptr[K, V]) Copy() *ValUintptr[K, V] {
	copied := ValUintptr[K, V]{base[K]{minAvgBucketSize: vv.minAvgBucketSize, maxAvgBucketSize: vv.maxAvgBucketSize, maxLogChunkSize: vv.maxLogChunkSize, buckets: newChunkArr(vv.maxLogChunkSize, (*chunkArr)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&vv.buckets)))).logChunkSize), HashF: vv.HashF}}
	tail := &copied.firstRelay
	tailIndex := uint(0)
	copied.buckets.set(tailIndex, tail)
	for cur, curAddr := vv.firstRelay.walk(), (unsafe.Pointer)(nil); cur != nil; cur = (*relay)(curAddr).walk() {
		if curAddr = addr(cur); !isRelay(cur) {
			a := (*valNode[K, uintptr])(curAddr)
			if index := copied.buckets.Index(a.hash); index != tailIndex {
				new := &relay{hash: index * (1 << copied.buckets.logChunkSize)}
				tail.next = unsafe.Pointer(uintptr(unsafe.Pointer(new)) | relayMask)
				tail = new
				copied.buckets.set(index, new)
				tailIndex = index
			}
			tail.next = unsafe.Pointer(&valNode[K, uintptr]{relay{hash: a.hash}, a.key, atomic.LoadUintptr(&(*valNode[K, uintptr])(curAddr).val)})
			tail = (*relay)(tail.next)
			copied.size.Add(resizingMask << 1)
		}
	}
	return &copied
}
