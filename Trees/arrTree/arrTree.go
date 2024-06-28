package Trees

import (
	"cmp"
	"golang.org/x/exp/constraints"
	"reflect"
	"unsafe"
)

type SBTree[T cmp.Ordered, S constraints.Unsigned] struct {
	base[S]
	vsHead *T     //v[i]corresponds to ifs[i+1]
	caps   [2]int //caps[0]=cap(ifs), caps[1]=cap(vs)
}

func (u *SBTree[T, S]) getV(i S) *T {
	return (*T)(unsafe.Add(unsafe.Pointer(u.vsHead), unsafe.Sizeof(*new(T))*uintptr(i)))
}
func New[T cmp.Ordered, S constraints.Unsigned](hint S) *SBTree[T, S] {
	ifs := make([]info[S], 1, hint+1)
	vs := make([]T, 0, hint)
	return &SBTree[T, S]{base[S]{ifsHead: unsafe.Pointer(unsafe.SliceData(ifs)), ifsLen: S(len(ifs))}, unsafe.SliceData(vs), [2]int{cap(ifs), cap(vs)}}
}

//	func Build[T cmp.Ordered, S constraints.Unsigned](sli []T) *SBTree[T, S] {
//		z := new(node[T, S])
//		z.l, z.r = z, z
//		var build func([]T) nodePtr[T, S]
//		build = func(s []T) nodePtr[T, S] {
//			if ifsLen(s) > 0 {
//				mid := ifsLen(s) >> 1
//				return &node[T, S]{s[mid], S(ifsLen(s)), build(s[0:mid]), build(s[mid+1:])}
//			} else {
//				return z
//			}
//		}
//		return &SBTree[T, S]{base[T, S]{build(sli), z}}
//	}
func (u *SBTree[T, S]) Insert(v T) bool {
	a, _ := u.BufferedInsert(v, nil)
	return a
}
func (u *SBTree[T, S]) BufferedInsert(v T, st []uintptr) (bool, []uintptr) { //offset from ifs[0] to either ifs[i].l or ifs[i].r
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
	{
		a := append(*(*[]info[S])(unsafe.Pointer(&reflect.SliceHeader{uintptr(u.ifsHead), int(prev), u.caps[0]})), info[S]{0, 0, 1})
		u.ifsHead, u.ifsLen, u.caps[0] = unsafe.Pointer(unsafe.SliceData(a)), S(len(a)), cap(a)
		b := append(*(*[]T)(unsafe.Pointer(&reflect.SliceHeader{uintptr(unsafe.Pointer(u.vsHead)), int(prev - 1), u.caps[1]})), v)
		u.vsHead, u.caps[1] = unsafe.SliceData(b), cap(b)
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
func (u *SBTree[T, S]) BufferedRemove(v T, st []S) (bool, []S) {
	for curI := &u.root; *curI != 0; {
		if v < *u.getV(*curI - 1) {
			st = append(st, *curI)
			curI = &u.getIf(*curI).l
		} else if v > *u.getV(*curI - 1) {
			st = append(st, *curI)
			curI = &u.getIf(*curI).r
		} else {
			for _, i := range st {
				u.getIf(i).sz--
			}
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
			//if Go_Utils.CheapRandN(uint32(bits.Len(uint(u.ifs[u.root].sz)))) == 2 { //when sz is 0-2 balancing is unnecessary
			//	for i := ifsLen(st) - 1; i > -1; i-- {
			//		u.ifs[*st[i]].sz--
			//		if v >= u.vs[*st[i]-1] {
			//			u.maintainRight(st[i])
			//		} else {
			//			u.maintainLeft(st[i])
			//		}
			//	}
			//}
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
