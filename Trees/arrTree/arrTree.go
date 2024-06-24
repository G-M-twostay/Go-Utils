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
	u.vs[0] = v
	u.root, r = u.insert(u.root)
	return
}

// vs[0] is target value
func (u *SBTree[T, S]) remove(curI *S) (deleted byte) {
	if *curI == 0 {
		return
	} else {
		cur := &u.ifs[*curI]
		if u.vs[0] < u.vs[*curI] {
			deleted = u.remove(&cur.l)
		} else if u.vs[0] == u.vs[*curI] {
			deleted = 1
			if cur.l == 0 {
				u.addFree(*curI)
				*curI = cur.r
			} else if cur.r == 0 {
				a := *curI
				*curI = cur.l
				u.addFree(a) //free is linked using .l
			} else {
				ni := &cur.r
				for ; u.ifs[*ni].l != 0; ni = &u.ifs[*ni].l {
					u.ifs[*ni].sz--
				}
				u.vs[*curI] = u.vs[*ni]
				u.addFree(*ni)
				*ni = u.ifs[*ni].r
			}
		} else {
			deleted = u.remove(&cur.r)
		}
		cur.sz -= S(deleted)
		return
	}
}

func (u *SBTree[T, S]) Remove(v T) bool {
	u.vs[0] = v
	return u.remove(&u.root) != 0
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
