package Trees

import (
	"golang.org/x/exp/constraints"
)

// A node in the SBTree
// The zero value is meaningless.
type info[T constraints.Unsigned] struct {
	l, r, sz T
}

type base[T any, S constraints.Unsigned] struct {
	root, free S         //free is the beginning of the linked list that contains all the free indexes, in which case we use l as next.
	ifs        []info[S] //0 is loopback nil
	vs         []T       //vs[0] is temporary store for recursion
}

func (u *base[T, S]) rotateLeft(ni *S) {
	n := &u.ifs[*ni]
	rci := n.r

	n.r = u.ifs[rci].l
	u.ifs[rci].l = *ni
	u.ifs[rci].sz = n.sz
	n.sz = u.ifs[n.l].sz + u.ifs[n.r].sz + 1
	*ni = rci
}

func (u *base[T, S]) rotateRight(ni *S) {
	n := &u.ifs[*ni]
	lci := n.l

	n.l = u.ifs[lci].r
	u.ifs[lci].r = *ni
	u.ifs[lci].sz = n.sz
	n.sz = u.ifs[n.l].sz + u.ifs[n.r].sz + 1
	*ni = lci
}

// adds a free index
func (u *base[T, S]) addFree(a S) {
	u.ifs[a].l = u.free
	u.ifs[a].sz = 0
	u.free = a
}

// gets a free index
func (u *base[T, S]) popFree() S {
	b := u.free
	u.free = u.ifs[u.free].l
	return b
}
func (u *base[T, S]) maintainLeft(curI *S) {
	cur := &u.ifs[*curI]
	if rc, lc := u.ifs[cur.r], u.ifs[cur.l]; u.ifs[lc.l].sz > rc.sz {
		u.rotateRight(curI)
	} else if u.ifs[lc.r].sz > rc.sz {
		u.rotateLeft(&cur.l)
		u.rotateRight(curI)
	} else {
		return
	}
	u.maintainLeft(&cur.l)
	u.maintainRight(&cur.r)
	u.maintainLeft(curI)
	u.maintainRight(curI)
}

func (u *base[T, S]) maintainRight(curI *S) {
	cur := &u.ifs[*curI]
	if rc, lc := u.ifs[cur.r], u.ifs[cur.l]; u.ifs[rc.r].sz > lc.sz {
		u.rotateLeft(curI)
	} else if u.ifs[rc.l].sz > lc.sz {
		u.rotateRight(&cur.r)
		u.rotateLeft(curI)
	} else {
		return
	}
	u.maintainLeft(&cur.l)
	u.maintainRight(&cur.r)
	u.maintainLeft(curI)
	u.maintainRight(curI)
}

//	func (u *base[T, S]) maintain(curI S, rightBigger bool) {
//		if rightBigger {
//			u.maintainRight(curI)
//		} else {
//			u.maintainLeft(curI)
//		}
//	}
func (u *base[T, S]) inOrder(curI S, f func(v T) bool) {
	cur := u.ifs[curI]
	if cur.l != 0 {
		u.inOrder(cur.l, f)
	}
	if curI != 0 {
		if !f(u.vs[curI]) {
			return
		}
	}
	if cur.r != 0 {
		u.inOrder(cur.r, f)
	}
}
func (u *base[T, S]) InOrder(f func(v T) bool) {
	u.inOrder(u.root, f)
}
func (u *base[T, S]) KLargest(k S) *T {
	for curI := u.root; curI != 0; {
		if lc := u.ifs[u.ifs[curI].l]; k <= lc.sz {
			curI = lc.l
		} else if k == lc.sz+1 {
			return &u.vs[curI]
		} else {
			k -= lc.sz + 1
			curI = u.ifs[curI].r
		}
	}
	return nil
}
func (u *base[T, S]) Size() uint {
	return uint(u.ifs[u.root].sz)
}
