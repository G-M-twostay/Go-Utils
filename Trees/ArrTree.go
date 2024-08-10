package Trees

import (
	"cmp"
	Go_Utils "github.com/g-m-twostay/go-utils"
	"math/bits"
	"reflect"
	"unsafe"
)

// Tree is a variant that supports only cmp.Ordered as keys.
type Tree[T cmp.Ordered, S Indexable] struct {
	base[T, S]
}

// New Tree that can hold hint number of elements without growing.
func New[T cmp.Ordered, S Indexable](hint S) *Tree[T, S] {
	ifs := make([]info[S], 1, hint+1)
	vs := make([]T, 0, hint)
	return &Tree[T, S]{base[T, S]{ifsHead: unsafe.Pointer(unsafe.SliceData(ifs)), ifsLen: S(len(ifs)), vsHead: unsafe.Pointer(unsafe.SliceData(vs)), caps: [2]int{cap(ifs), cap(vs)}}}
}

// From a given value array, directly build a tree. The array is handled to the tree, and it mustn't be modified by the caller later. vs must be
// sorted in ascending order for the tree to not be corrupt.
// Time: O(C).
func From[T cmp.Ordered, S Indexable](vs []T) *Tree[T, S] {
	root, ifs := buildIfs(S(len(vs)), make([][3]S, 0, bits.Len(uint(len(vs)))))
	return &Tree[T, S]{base[T, S]{root: root, ifsHead: unsafe.Pointer(unsafe.SliceData(ifs)), ifsLen: S(len(ifs)), vsHead: unsafe.Pointer(unsafe.SliceData(vs)), caps: [2]int{cap(ifs), cap(vs)}}}
}

// Add an element to the tree. Add guarantees that holes are filled first before appending to the underlying arrays. st is
// the slice used as the recursion stack. Returns whether the element is added and the grown recursion stack. To reuse the
// slice:
//
//	var st []uintptr
//	for ... {
//		_, st = tree.Add(..., st)
//	}
//
// Reassigning the value is unnecessary if st is allocated to be enough.
// Time: O(D). Space: O(H).
// Type: W0, W1, W2.
func (u *Tree[T, S]) Add(v T, st []uintptr) (bool, []uintptr) {
	// st stores the address offset from ifs[0] to either ifs[i].l or ifs[i].r
	for curI := u.root; curI != 0; {
		if v < *u.getV(curI - 1) {
			l := &u.getIf(curI).l
			st = append(st, uintptr(unsafe.Pointer(l))-uintptr(u.ifsHead))
			curI = *l
		} else if v > *u.getV(curI - 1) {
			r := &u.getIf(curI).r
			st = append(st, uintptr(unsafe.Pointer(r))-uintptr(u.ifsHead))
			curI = *r
		} else {
			return false, st
		}
	}
	if u.root = u.popFree(); u.root == 0 {
		u.root = u.ifsLen
		//use reflect.SliceHeader to directly set both cap and len.
		a := append(*(*[]info[S])(unsafe.Pointer(&reflect.SliceHeader{uintptr(u.ifsHead), int(u.root), u.caps[0]})), info[S]{0, 0, 1})
		u.ifsHead, u.ifsLen, u.caps[0] = unsafe.Pointer(unsafe.SliceData(a)), S(len(a)), cap(a)
		b := append(*(*[]T)(unsafe.Pointer(&reflect.SliceHeader{uintptr(u.vsHead), int(u.root - 1), u.caps[1]})), v)
		u.vsHead, u.caps[1] = unsafe.Pointer(unsafe.SliceData(b)), cap(b)
	} else {
		*u.getIf(u.root) = info[S]{0, 0, 1}
		*u.getV(u.root - 1) = v
	}

	for i := len(st) - 1; i > -1; i-- {
		*(*S)(unsafe.Add(u.ifsHead, st[i])) = u.root //ptr to u.ifs[index].l or u.ifs[index].r
		u.root = S(st[i] / unsafe.Sizeof(info[S]{}))
		u.getIf(u.root).sz++ //index of st[i] in ifs
		if v >= *u.getV(u.root - 1) {
			u.maintainRight(&u.root)
		} else {
			u.maintainLeft(&u.root)
		}
	}
	return true, st
}

/*
In the original implementation, deletion don't trigger maintain to balance the tree. The reasoning is that because deletion don't
add new elements, even if the tree's structure is broken D<=e(B). Then, on the next insertion, the tree is fixed.
This is fine if we're only deleting things from the tree,
but it can get problematic if lots of read operations would happen after some deletions. However, balancing after every
deletion is unnecessary, as deletion don't tend to create chains that would significant degrade performance like insertion, so
how do we determine when to balance?

Consider a perfect binary tree, in order to reduce D by 1, one would need to delete half of its elements. Just think of removing
the entire last layer, which has C/2 elements. Intuitively, the larger the tree, the less likely is a single deletion to
be able to reduce the height of the tree significantly. Therefore, we perform a die roll after each deletion with the chance
of maintaining inversely proportionally to C/2. This also proved to be the best overall balance when doing reading-after-delete
benchmarks.
*/

// Del an element from the tree. Del sometimes balances the tree; the chance is inversely proportional to tree's size. Returns
// the grown stack and whether the element is deleted.
// Time: O(D). Space: O(H).
// Type: W0, W1, W2.
func (u *Tree[T, S]) Del(v T, st []uintptr) (bool, []uintptr) {
	//st stores &ifs[i]
	for curI := &u.root; *curI != 0; {
		if cvp := u.getV(*curI - 1); v < *cvp {
			st = append(st, uintptr(unsafe.Pointer(curI)))
			curI = &u.getIf(*curI).l
		} else if v > *cvp {
			st = append(st, uintptr(unsafe.Pointer(curI)))
			curI = &u.getIf(*curI).r
		} else {
			if cur := u.getIf(*curI); cur.l == 0 {
				u.addFree(*curI)
				*curI = cur.r
			} else if cur.r == 0 {
				a := *curI
				*curI = cur.l
				u.addFree(a)
			} else {
				si := &cur.r
				for cur.sz--; u.getIf(*si).l != 0; si = &u.getIf(*si).l {
					u.getIf(*si).sz--
				}
				*cvp = *u.getV(*si - 1)
				u.addFree(*si)
				*si = u.getIf(*si).r
			}
			for _, a := range st {
				u.getIf(*(*S)(unsafe.Pointer(a))).sz--
			}
			if Go_Utils.CheapRandN(uint32((u.getIf(u.root).sz+1)>>1)) == 2 { //when sz is 0-2 balancing is unnecessary
				for i := len(st) - 1; i > -1; i-- {
					if v <= *u.getV(*(*S)(unsafe.Pointer(st[i])) - 1) {
						u.maintainRight((*S)(unsafe.Pointer(st[i])))
					} else {
						u.maintainLeft((*S)(unsafe.Pointer(st[i])))
					}
				}
			}
			return true, st
		}
	}
	return false, st
}

// Get the pointer to the element that's equal to v in the tree.
// Time: O(D). Space: O(1).
// Type: R0, R1..
func (u *Tree[T, S]) Get(v T) *T {
	for curI := u.root; curI != 0; {
		if cvp := u.getV(curI - 1); v < *cvp {
			curI = u.getIf(curI).l
		} else if v > *cvp {
			curI = u.getIf(curI).r
		} else {
			return cvp
		}
	}
	return nil
}

// Predecessor of v. If strict is true, result<v if found; otherwise, result<=v.
// Time: O(D). Space: O(1).
// Type: R0, R1.
func (u *Tree[T, S]) Predecessor(v T, strict bool) (p *T) {
	if curI := u.root; strict {
		for curI != 0 {
			if v <= *u.getV(curI - 1) {
				curI = u.getIf(curI).l
			} else {
				p = u.getV(curI - 1)
				curI = u.getIf(curI).r
			}
		}
	} else {
		for curI != 0 {
			if v < *u.getV(curI - 1) {
				curI = u.getIf(curI).l
			} else {
				p = u.getV(curI - 1)
				curI = u.getIf(curI).r
			}
		}
	}
	return
}

// Successor of v. If strict is true, result>v if found; otherwise, result>=v.
// Time: O(D). Space: O(1).
// Type: R0, R1.
func (u *Tree[T, S]) Successor(v T, strict bool) (p *T) {
	if curI := u.root; strict {
		for curI != 0 {
			if v < *u.getV(curI - 1) {
				p = u.getV(curI - 1)
				curI = u.getIf(curI).l
			} else {
				curI = u.getIf(curI).r
			}
		}
	} else {
		for curI != 0 {
			if v > *u.getV(curI - 1) {
				curI = u.getIf(curI).r
			} else {
				p = u.getV(curI - 1)
				curI = u.getIf(curI).l
			}
		}
	}
	return
}

// RankOf v, starting from 0. If v isn't found, returns the rank as if v is added to the tree.
// Time: O(D). Space: O(1).
// Type: R0, R1, R2.
func (u *Tree[T, S]) RankOf(v T) (S, bool) {
	var ra S = 0
	for curI := u.root; curI != 0; {
		if cur := *u.getIf(curI); v < *u.getV(curI - 1) {
			curI = cur.l
		} else if v > *u.getV(curI - 1) {
			ra += u.getIf(cur.l).sz + 1
			curI = cur.r
		} else {
			return ra + u.getIf(cur.l).sz, true
		}
	}
	return ra, false
}

// Clone the tree, making an almost exact copy (up to len(A0) and len(A1)).
// Time: O(C). Space: O(1) disregarding the new tree.
// Type: R0, R1, R2.
func (u *Tree[T, S]) Clone() *Tree[T, S] {
	newIfs := make([]info[S], u.ifsLen, u.caps[0])
	copy(newIfs, unsafe.Slice((*info[S])(u.ifsHead), u.ifsLen))
	newVs := make([]T, u.ifsLen-1, u.caps[1])
	copy(newVs, unsafe.Slice((*T)(u.vsHead), u.ifsLen-1))
	return &Tree[T, S]{base[T, S]{unsafe.Pointer(unsafe.SliceData(newIfs)), unsafe.Pointer(unsafe.SliceData(newVs)), u.caps, u.root, u.free, u.ifsLen}}
}
