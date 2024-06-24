package Trees

import (
	"cmp"
	"golang.org/x/exp/constraints"
)

type SBTree[T cmp.Ordered, S constraints.Unsigned] struct {
	base[T, S]
}

func New[T cmp.Ordered, S constraints.Unsigned](hint S) *SBTree[T, S] {
	if hint < 1 {
		hint = 1
	}
	return &SBTree[T, S]{base[T, S]{ifs: make([]info[S], hint), vs: make([]T, hint)}}
}

//func Build[T cmp.Ordered, S constraints.Unsigned](sli []T) *SBTree[T, S] {
//	z := new(node[T, S])
//	z.l, z.r = z, z
//	var build func([]T) nodePtr[T, S]
//	build = func(s []T) nodePtr[T, S] {
//		if len(s) > 0 {
//			mid := len(s) >> 1
//			return &node[T, S]{s[mid], S(len(s)), build(s[0:mid]), build(s[mid+1:])}
//		} else {
//			return z
//		}
//	}
//	return &SBTree[T, S]{base[T, S]{build(sli), z}}
//}

// target value stored in v[0]
func (u *SBTree[T, S]) insert(curI S) (S, bool) {
	if curI == 0 {
		if u.free == 0 {
			curI = S(len(u.ifs))
			u.ifs = append(u.ifs, info[S]{0, 0, 1})
			u.vs = append(u.vs, u.vs[0])
		} else {
			curI = u.popFree()
			u.ifs[curI], u.vs[curI] = info[S]{0, 0, 1}, u.vs[0]
		}
		return curI, true
	} else {
		inserted := false
		if u.vs[0] < u.vs[curI] {
			u.ifs[curI].l, inserted = u.insert(u.ifs[curI].l)
		} else if u.vs[0] == u.vs[curI] {
			return curI, false
		} else {
			u.ifs[curI].r, inserted = u.insert(u.ifs[curI].r)
		}
		if inserted {
			u.ifs[curI].sz++
			if u.vs[0] >= u.vs[curI] {
				u.maintainRight(&curI)
			} else {
				u.maintainLeft(&curI)
			}
		}
		return curI, inserted
	}
}

func (u *SBTree[T, S]) Insert(v T) (r bool) {
	var st []uint //go array fits in int, so we use uint and take first bit as l or r
	for curI := u.root; curI != 0; {
		if v < u.vs[curI] {
			st = append(st, uint(curI))
			curI = u.ifs[curI].l
		} else if v > u.vs[curI] {
			st = append(st, uint(curI)|1<<63)
			curI = u.ifs[curI].r
		} else {
			return false
		}
	}
	last := (len(u.ifs))
	u.ifs = append(u.ifs, info[S]{0, 0, 1})
	u.vs = append(u.vs, u.vs[0])
	for i := last; i > 0; i-- {
		st[i-1] = uint(last)

	}
}

func (u *SBTree[T, S]) Remove(v T) bool {
	var st []S
	for curI := &u.root; *curI != 0; {
		if v < u.vs[*curI] {
			st = append(st, *curI)
			curI = &(u.ifs[*curI].l)
		} else if v > u.vs[*curI] {
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
				u.vs[*curI] = u.vs[*si]
				u.addFree(*si)
				*si = u.ifs[*si].r
			}
			return true
		}
	}
	return false
}

func (u *SBTree[T, S]) Has(v T) bool {
	for curI := u.root; curI != 0; {
		if v < u.vs[curI] {
			curI = u.ifs[curI].l
		} else if v == u.vs[curI] {
			return true
		} else {
			curI = u.ifs[curI].r
		}
	}
	return false
}

func (u *SBTree[T, S]) Predecessor(v T) (p *T) {
	for curI := u.root; curI != 0; {
		if v <= u.vs[curI] {
			curI = u.ifs[curI].l
		} else {
			p = &u.vs[curI]
			curI = u.ifs[curI].r
		}
	}
	return
}

func (u *SBTree[T, S]) Successor(v T) (p *T) {
	for curI := u.root; curI != 0; {
		if v < u.vs[curI] {
			p = &u.vs[curI]
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
		if lci := u.ifs[curI].l; v < u.vs[curI] {
			curI = lci
		} else if v == u.vs[curI] {
			return ra + u.ifs[lci].sz + 1
		} else {
			ra += u.ifs[lci].sz + 1
			curI = u.ifs[curI].r
		}
	}
	return 0
}
