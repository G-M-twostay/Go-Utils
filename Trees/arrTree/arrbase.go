package Trees

import (
	"golang.org/x/exp/constraints"
)

// A node in the SBTree
// The zero value is meaningless.
type info[S constraints.Unsigned] struct {
	l, r, sz S
}

type base[S constraints.Unsigned] struct {
	root, free S         //free is the beginning of the linked list that contains all the free indexes, in which case we use l as next.
	ifs        []info[S] //0 is loopback nil. all index is based on ifs
}

func (u *base[S]) rotateLeft(ni *S) {
	n := &u.ifs[*ni]
	rci := n.r

	n.r = u.ifs[rci].l
	u.ifs[rci].l, u.ifs[rci].sz = *ni, n.sz
	n.sz = u.ifs[n.l].sz + u.ifs[n.r].sz + 1
	*ni = rci
}

func (u *base[S]) rotateRight(ni *S) {
	n := &u.ifs[*ni]
	lci := n.l

	n.l = u.ifs[lci].r
	u.ifs[lci].r, u.ifs[lci].sz = *ni, n.sz
	n.sz = u.ifs[n.l].sz + u.ifs[n.r].sz + 1
	*ni = lci
}

// adds a free index
func (u *base[S]) addFree(a S) {
	u.ifs[a].l = u.free
	u.free = a
}

// gets a free index
func (u *base[S]) popFree() S {
	b := u.free
	u.free = u.ifs[u.free].l
	return b
}
func (u *base[S]) maintainLeft(curI *S) {
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
func (u *base[S]) mtLe(curI *S) {

}
func (u *base[S]) maintainRight(curI *S) {
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
func (u *base[S]) inOrder(curI S, f func(S) bool) {
	cur := u.ifs[curI]
	if cur.l != 0 {
		u.inOrder(cur.l, f)
	}
	if curI != 0 {
		if !f(curI) {
			return
		}
	}
	if cur.r != 0 {
		u.inOrder(cur.r, f)
	}
}
func (u *base[S]) InOrder(f func(S) bool) {
	u.inOrder(u.root, f)
}

func (u *base[S]) Size() S {
	return u.ifs[u.root].sz
}
func (u *base[S]) clrIfs() {
	u.ifs = u.ifs[:1]
	u.root, u.free = 0, 0
}
