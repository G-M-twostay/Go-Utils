package Trees

import (
	"cmp"
	Go_Utils "github.com/g-m-twostay/go-utils"
	"golang.org/x/exp/constraints"
	"math/bits"
	"reflect"
	"unsafe"
)

// Tree is a variant that supports only cmp.Ordered as keys.
type Tree[T cmp.Ordered, S constraints.Unsigned] struct {
	base[T, S]
}

func New[T cmp.Ordered, S constraints.Unsigned](hint S) *Tree[T, S] {
	ifs := make([]info[S], 1, hint+1)
	vs := make([]T, 0, hint)
	return &Tree[T, S]{base[T, S]{ifsHead: unsafe.Pointer(unsafe.SliceData(ifs)), ifsLen: S(len(ifs)), vsHead: unsafe.Pointer(unsafe.SliceData(vs)), caps: [2]int{cap(ifs), cap(vs)}}}
}

// From a given value array, directly build a tree. The array is handled to the tree and it mustn't be modified by the caller later.
func From[T cmp.Ordered, S constraints.Unsigned](vs []T) *Tree[T, S] {
	root, ifs := buildIfs(S(len(vs)), make([][3]S, 0, bits.Len(uint(len(vs)))))
	return &Tree[T, S]{base[T, S]{root: root, ifsHead: unsafe.Pointer(unsafe.SliceData(ifs)), ifsLen: S(len(ifs)), vsHead: unsafe.Pointer(unsafe.SliceData(vs)), caps: [2]int{cap(ifs), cap(vs)}}}
}

// Add an element to the tree. Add guarantees that holes are filled first before appending to the underlying arrays.
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
		index := S(st[i] / unsafe.Sizeof(info[S]{}))
		u.getIf(index).sz++ //index of st[i] in ifs
		if v >= *u.getV(index - 1) {
			u.maintainRight(&index)
		} else {
			u.maintainLeft(&index)
		}
		u.root = index
	}
	return true, st
}

// Del an element from the tree. Del sometimes balances the tree; the chance is inversely proportional to tree's size.
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
