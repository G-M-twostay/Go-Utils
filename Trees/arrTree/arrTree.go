package Trees

import (
	"cmp"
	Go_Utils "github.com/g-m-twostay/go-utils"
	"golang.org/x/exp/constraints"
	"reflect"
	"unsafe"
)

type SBTree[T cmp.Ordered, S constraints.Unsigned] struct {
	indexer[S, T]
	caps [2]int //caps[0]=cap(ifs), caps[1]=cap(vs)
}

func New[T cmp.Ordered, S constraints.Unsigned](hint S) *SBTree[T, S] {
	ifs := make([]info[S], 1, hint+1)
	vs := make([]T, 0, hint)
	return &SBTree[T, S]{indexer[S, T]{ifsHead: unsafe.Pointer(unsafe.SliceData(ifs)), ifsLen: S(len(ifs)), vsHead: unsafe.Pointer(unsafe.SliceData(vs))}, [2]int{cap(ifs), cap(vs)}}
}

func (u *SBTree[T, S]) Insert(v T) bool {
	a, _ := u.BufferedInsert(v, nil)
	return a
}
func (u *SBTree[T, S]) BufferedInsert(v T, st []uintptr) (bool, []uintptr) {
	st = st[:0] //offset from ifs[0] to either ifs[i].l or ifs[i].r
	for curI := u.root; curI != 0; {
		if v < *u.getV(curI - 1) {
			st = append(st, uintptr(unsafe.Pointer(&u.getIf(curI).l))-uintptr(u.ifsHead))
			curI = u.getIf(curI).l
		} else if v > *u.getV(curI - 1) {
			st = append(st, uintptr(unsafe.Pointer(&u.getIf(curI).r))-uintptr(u.ifsHead))
			curI = u.getIf(curI).r
		} else {
			return false, st
		}
	}
	prev := u.ifsLen
	{ //use reflect.SliceHeader to directly set both cap and len.
		a := append(*(*[]info[S])(unsafe.Pointer(&reflect.SliceHeader{uintptr(u.ifsHead), int(prev), u.caps[0]})), info[S]{0, 0, 1})
		u.ifsHead, u.ifsLen, u.caps[0] = unsafe.Pointer(unsafe.SliceData(a)), S(len(a)), cap(a)
		b := append(*(*[]T)(unsafe.Pointer(&reflect.SliceHeader{uintptr(u.vsHead), int(prev - 1), u.caps[1]})), v)
		u.vsHead, u.caps[1] = unsafe.Pointer(unsafe.SliceData(b)), cap(b)
	}
	for i := len(st) - 1; i > -1; i-- {
		*(*S)(unsafe.Add(u.ifsHead, st[i])) = prev //ptr to u.ifs[index].l or u.ifs[index].r
		index := S(st[i] / unsafe.Sizeof(info[S]{}))
		u.getIf(index).sz++ //index of st[i] in ifs
		if v >= *u.getV(index - 1) {
			u.maintainRight(&index)
		} else {
			u.maintainLeft(&index)
		}
		prev = index
	}
	u.root = prev
	return true, st
}
func (u *SBTree[T, S]) Remove(v T) bool {
	a, _ := u.BufferedRemove(v, nil)
	return a
}
func (u *SBTree[T, S]) BufferedRemove(v T, st []uintptr) (bool, []uintptr) {
	st = st[:0] //stores &ifs[i]
	for curI := &u.root; *curI != 0; {
		if v < *u.getV(*curI - 1) {
			st = append(st, uintptr(unsafe.Pointer(curI)))
			curI = &u.getIf(*curI).l
		} else if v > *u.getV(*curI - 1) {
			st = append(st, uintptr(unsafe.Pointer(curI)))
			curI = &u.getIf(*curI).r
		} else {
			if u.getIf(*curI).l == 0 {
				u.addFree(*curI)
				*curI = u.getIf(*curI).r
			} else if u.getIf(*curI).r == 0 {
				a := *curI
				*curI = u.getIf(*curI).l
				u.addFree(a)
			} else {
				si := &u.getIf(*curI).r
				for u.getIf(*curI).sz--; u.getIf(*si).l != 0; si = &u.getIf(*si).l {
					u.getIf(*si).sz--
				}
				*u.getV(*curI - 1) = *u.getV(*si - 1)
				u.addFree(*si)
				*si = u.getIf(*si).r
			}
			for _, a := range st {
				u.getIf(*(*S)(unsafe.Pointer(a))).sz--
			}
			if Go_Utils.CheapRandN(uint32(u.getIf(u.root).sz+1)/2) == 2 { //when sz is 0-2 balancing is unnecessary
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

func (u *SBTree[T, S]) Has(v T) bool {
	for curI := u.root; curI != 0; {
		if v < *u.getV(curI - 1) {
			curI = u.getIf(curI).l
		} else if v == *u.getV(curI - 1) {
			return true
		} else {
			curI = u.getIf(curI).r
		}
	}
	return false
}

func (u *SBTree[T, S]) Predecessor(v T) (p *T) {
	for curI := u.root; curI != 0; {
		if v <= *u.getV(curI - 1) {
			curI = u.getIf(curI).l
		} else {
			p = u.getV(curI - 1)
			curI = u.getIf(curI).r
		}
	}
	return
}

func (u *SBTree[T, S]) Successor(v T) (p *T) {
	for curI := u.root; curI != 0; {
		if v < *u.getV(curI - 1) {
			p = u.getV(curI - 1)
			curI = u.getIf(curI).l
		} else {
			curI = u.getIf(curI).r
		}
	}
	return
}

//func (u *SBTree[T, S]) RankOf(v T) S {
//	var ra S = 0
//	for curI := u.root; curI != 0; {
//		if lci := u.ifs[curI].l; v < u.vs[curI-1] {
//			curI = lci
//		} else if v == u.vs[curI-1] {
//			return ra + u.ifs[lci].sz + 1
//		} else {
//			ra += u.ifs[lci].sz + 1
//			curI = u.ifs[curI].r
//		}
//	}
//	return 0
//}
//func (u *SBTree[T, S]) KLargest(k S) *T {
//	for curI := u.root; curI != 0; {
//		if lc := u.ifs[u.ifs[curI].l]; k <= lc.sz {
//			curI = lc.l
//		} else if k == lc.sz+1 {
//			return &u.vs[curI-1]
//		} else {
//			k -= lc.sz + 1
//			curI = u.ifs[curI].r
//		}
//	}
//	return nil
//}
