package Trees

import (
	"cmp"
	"golang.org/x/exp/constraints"
	"unsafe"
)

type SBTree[T cmp.Ordered, S constraints.Unsigned] struct {
	base[S]
	vs []T //v[i]corresponds to ifs[i+1]
}

func New[T cmp.Ordered, S constraints.Unsigned](hint S) *SBTree[T, S] {
	return &SBTree[T, S]{base[S]{ifs: make([]info[S], 1, hint+1)}, make([]T, 0, hint)}
}

//	func Build[T cmp.Ordered, S constraints.Unsigned](sli []T) *SBTree[T, S] {
//		z := new(node[T, S])
//		z.l, z.r = z, z
//		var build func([]T) nodePtr[T, S]
//		build = func(s []T) nodePtr[T, S] {
//			if len(s) > 0 {
//				mid := len(s) >> 1
//				return &node[T, S]{s[mid], S(len(s)), build(s[0:mid]), build(s[mid+1:])}
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
		if v < u.vs[curI-1] {
			st = append(st, uintptr(unsafe.Pointer(&u.ifs[curI].l))-uintptr(unsafe.Pointer(&u.ifs[0])))
			curI = u.ifs[curI].l
		} else if v > u.vs[curI-1] {
			st = append(st, uintptr(unsafe.Pointer(&u.ifs[curI].r))-uintptr(unsafe.Pointer(&u.ifs[0])))
			curI = u.ifs[curI].r
		} else {
			return false, st
		}
	}
	prev := S(len(u.ifs))
	u.ifs, u.vs = append(u.ifs, info[S]{0, 0, 1}), append(u.vs, v)
	for i := len(st) - 1; i > -1; i-- {
		*(*S)(unsafe.Add(unsafe.Pointer(&u.ifs[0]), st[i])) = prev //ptr to u.ifs[index].l or u.ifs[index].r
		index := S(st[i] / unsafe.Sizeof(u.ifs[0]))
		u.ifs[index].sz++ //index of st[i] in ifs
		if v >= u.vs[index-1] {
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
		if v < u.vs[*curI-1] {
			st = append(st, *curI)
			curI = &(u.ifs[*curI].l)
		} else if v > u.vs[*curI-1] {
			st = append(st, *curI)
			curI = &(u.ifs[*curI].r)
		} else {
			for _, v := range st {
				u.ifs[v].sz--
			}
			if u.ifs[*curI].l == 0 {
				u.addFree(*curI)
				*curI = u.ifs[*curI].r
			} else if u.ifs[*curI].r == 0 {
				a := *curI
				*curI = u.ifs[*curI].l
				u.addFree(a)
			} else {
				si := &u.ifs[*curI].r
				for u.ifs[*curI].sz--; u.ifs[*si].l != 0; si = &u.ifs[*si].l {
					u.ifs[*si].sz--
				}
				u.vs[*curI-1] = u.vs[*si-1]
				u.addFree(*si)
				*si = u.ifs[*si].r
			}
			return true, st
		}
	}
	return false, st
}

func (u *SBTree[T, S]) Has(v T) bool {
	for curI := u.root; curI != 0; {
		if v < u.vs[curI-1] {
			curI = u.ifs[curI].l
		} else if v == u.vs[curI-1] {
			return true
		} else {
			curI = u.ifs[curI].r
		}
	}
	return false
}

func (u *SBTree[T, S]) Predecessor(v T) (p *T) {
	for curI := u.root; curI != 0; {
		if v <= u.vs[curI-1] {
			curI = u.ifs[curI].l
		} else {
			p = &u.vs[curI-1]
			curI = u.ifs[curI].r
		}
	}
	return
}

func (u *SBTree[T, S]) Successor(v T) (p *T) {
	for curI := u.root; curI != 0; {
		if v < u.vs[curI-1] {
			p = &u.vs[curI-1]
			curI = u.ifs[curI].l
		} else {
			curI = u.ifs[curI].r
		}
	}
	return
}

func (u *SBTree[T, S]) RankOf(v T) S {
	var ra S = 0
	for curI := u.root; curI != 0; {
		if lci := u.ifs[curI].l; v < u.vs[curI-1] {
			curI = lci
		} else if v == u.vs[curI-1] {
			return ra + u.ifs[lci].sz + 1
		} else {
			ra += u.ifs[lci].sz + 1
			curI = u.ifs[curI].r
		}
	}
	return 0
}
func (u *SBTree[T, S]) KLargest(k S) *T {
	for curI := u.root; curI != 0; {
		if lc := u.ifs[u.ifs[curI].l]; k <= lc.sz {
			curI = lc.l
		} else if k == lc.sz+1 {
			return &u.vs[curI-1]
		} else {
			k -= lc.sz + 1
			curI = u.ifs[curI].r
		}
	}
	return nil
}
