package Trees

import (
	"golang.org/x/exp/constraints"
	"unsafe"
)

// A node in the SBTree
// The zero value is meaningless.
type info[S constraints.Unsigned] struct {
	l, r, sz S
}

type base[S constraints.Unsigned] struct {
	root, free, ifsLen S              //free is the beginning of the linked list that contains all the free indexes, in which case we use l as next.
	ifsHead            unsafe.Pointer //0 is loopback nil. all index is based on ifs. length is size+1
}

func (u *base[S]) getIf(i S) *info[S] {
	return (*info[S])(unsafe.Add(u.ifsHead, unsafe.Sizeof(info[S]{})*uintptr(i)))
}
func (u *base[S]) rotateLeft(ni *S) {
	n := u.getIf(*ni)
	rci := n.r

	n.r = u.getIf(rci).l
	u.getIf(rci).l, u.getIf(rci).sz = *ni, n.sz
	n.sz = u.getIf(n.l).sz + u.getIf(n.r).sz + 1
	*ni = rci
}

func (u *base[S]) rotateRight(ni *S) {
	n := u.getIf(*ni)
	lci := n.l

	n.l = u.getIf(lci).r
	u.getIf(lci).r, u.getIf(lci).sz = *ni, n.sz
	n.sz = u.getIf(n.l).sz + u.getIf(n.r).sz + 1
	*ni = lci
}

// adds a free index
func (u *base[S]) addFree(a S) {
	u.getIf(a).l = u.free
	u.free = a
}

// gets a free index
func (u *base[S]) popFree() S {
	b := u.free
	u.free = u.getIf(u.free).l
	return b
}
func (u *base[S]) maintainLeft(curI *S) {
	cur := u.getIf(*curI)
	if rc, lc := *u.getIf(cur.r), *u.getIf(cur.l); u.getIf(lc.l).sz > rc.sz {
		u.rotateRight(curI)
	} else if u.getIf(lc.r).sz > rc.sz {
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
func (u *base[S]) maintainRight(curI *S) {
	cur := u.getIf(*curI)
	if rc, lc := *u.getIf(cur.r), *u.getIf(cur.l); u.getIf(rc.r).sz > lc.sz {
		u.rotateLeft(curI)
	} else if u.getIf(rc.l).sz > lc.sz {
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

func (u *base[S]) inOrder(curI S, f func(S) bool) {
	cur := *u.getIf(curI)
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
func (u base[S]) InOrder(f func(S) bool) {
	u.inOrder(u.root, f)
}

func (u base[S]) Size() S {
	return u.getIf(u.root).sz
}
func (u *base[S]) Clear() {
	u.ifsLen = 1
}
