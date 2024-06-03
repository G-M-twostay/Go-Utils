package IntMap

import (
	"github.com/g-m-twostay/go-utils/Maps/internal"
	"unsafe"
)

func (u *IntMap[K, V]) LoadPtrOrStore(key K, val V) (v *V, loaded bool) {
	hash, vPtr := internal.Mask(u.hash(key)), unsafe.Pointer(&val)

	prevLock := u.buckets.Load().Get(hash)
	if !prevLock.safeRLock() {
		prevLock.RUnlock()
		prevLock = u.buckets.Load().Get(hash)
		prevLock.RLock()
	}

	for left := &prevLock.node; ; {
		rightPtr := left.Next()
		if rightB := (*node)(rightPtr); rightB == nil || hash < rightB.Hash() {
			if left.dangerLink(rightPtr, unsafe.Pointer(&value[K]{node: node{info: hash, nx: rightPtr}, v: vPtr, k: key})) {
				prevLock.RUnlock()
				u.size.Add(1)
				u.trySplit()
				return
			}
		} else if right := (*value[K])(rightPtr); hash == rightB.info && key == right.k {
			prevLock.RUnlock()
			return (*V)(right.get()), true
		} else {
			if left = rightB; rightB.isRelay() {
				prevLock.RUnlock()
				if prevLock = (*relay)(rightPtr); !prevLock.safeRLock() {
					prevLock.RUnlock()
					prevLock = u.buckets.Load().Get(hash)
					prevLock.RLock()
					left = &prevLock.node
				}
			}
		}
	}
}
func (u *IntMap[K, V]) LoadPtr(key K) *V {
	hash := internal.Mask(u.hash(key))
	if r := search(u.buckets.Load().Get(hash), key, hash); r == nil {
		return nil
	} else {
		return (*V)(r.get())
	}
}
func (u *IntMap[K, V]) LoadPtrAndDelete(key K) (v *V, loaded bool) {
	hash := internal.Mask(u.hash(key))
	prevLock := u.buckets.Load().Get(hash)

	if !prevLock.safeLock() {
		prevLock.Unlock()
		prevLock = u.buckets.Load().Get(hash)
		prevLock.Lock()
	}

	for left := &prevLock.node; ; {
		rightPtr := left.nx
		if rightB := (*node)(rightPtr); rightB == nil || hash < rightB.Hash() {
			prevLock.Unlock()
			return
		} else if right := (*value[K])(rightPtr); hash == rightB.info && key == right.k {
			left.dangerUnlink(rightB)
			prevLock.Unlock()
			u.size.Add(^uintptr(1 - 1))
			u.tryMerge()
			return
		} else {
			if left = rightB; rightB.isRelay() {
				prevLock.Unlock()
				if prevLock = (*relay)(rightPtr); !prevLock.safeLock() {
					prevLock.Unlock()
					prevLock = u.buckets.Load().Get(hash)
					prevLock.Lock()
					left = &prevLock.node
				}
			}
		}
	}

}

func (u *IntMap[K, V]) RangePtr(f func(K, *V) bool) {
	for cur := (*node)(u.buckets.Load().Get(0).Next()); cur != nil; cur = (*node)(cur.Next()) {
		if !cur.isRelay() {
			if t := (*value[K])(unsafe.Pointer(cur)); !f(t.k, (*V)(t.get())) {
				break
			}
		}
	}
}
func (u *IntMap[K, V]) TakePtr() (key K, val *V) {
	if firstPtr := u.buckets.Load().Fetch(0).Next(); firstPtr != nil {
		first := (*value[K])(firstPtr)
		key, val = first.k, (*V)(first.get())
	}
	return
}

func (u *IntMap[K, V]) SetPtr(key K, val *V) (set bool) {
	hash := internal.Mask(u.hash(key))
	r := search(u.buckets.Load().Get(hash), key, hash)
	set = r != nil
	if set {
		r.set(unsafe.Pointer(&val))
	}
	return
}
func (u *IntMap[K, V]) CompareAndSwapPtr(key K, old *V, new *V) (success bool) {
	hash := internal.Mask(u.hash(key))
	if r := search(u.buckets.Load().Get(hash), key, hash); r != nil {
		success = r.cas(unsafe.Pointer(old), unsafe.Pointer(new))
	}
	return
}

func (u *IntMap[K, V]) SwapPtr(key K, val *V) (old *V) {
	hash := internal.Mask(u.hash(key))
	if r := search(u.buckets.Load().Get(hash), key, hash); r != nil {
		old = (*V)(r.swap(unsafe.Pointer(val)))
	}
	return
}
