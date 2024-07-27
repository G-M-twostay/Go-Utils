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

type indexer[S constraints.Unsigned, T any] struct {
	root, free, ifsLen S              //free is the beginning of the linked list that contains all the free indexes, in which case we use l as next.
	ifsHead            unsafe.Pointer //0 is loopback nil. all index is based on ifs. length is size+1
	vsHead             unsafe.Pointer //v[i]corresponds to ifs[i+1]
}

func (u *indexer[S, T]) getIf(i S) *info[S] {
	return (*info[S])(unsafe.Add(u.ifsHead, unsafe.Sizeof(info[S]{})*uintptr(i)))
}
func (u *indexer[S, T]) rotateLeft(ni *S) {
	n := u.getIf(*ni)
	rci := n.r

	n.r = u.getIf(rci).l
	u.getIf(rci).l, u.getIf(rci).sz = *ni, n.sz
	n.sz = u.getIf(n.l).sz + u.getIf(n.r).sz + 1
	*ni = rci
}

func (u *indexer[S, T]) rotateRight(ni *S) {
	n := u.getIf(*ni)
	lci := n.l

	n.l = u.getIf(lci).r
	u.getIf(lci).r, u.getIf(lci).sz = *ni, n.sz
	n.sz = u.getIf(n.l).sz + u.getIf(n.r).sz + 1
	*ni = lci
}

// adds a free index
func (u *indexer[S, T]) addFree(a S) {
	u.getIf(a).l = u.free
	u.free = a
}

// gets a free index
func (u *indexer[S, T]) popFree() S {
	b := u.free
	u.free = u.getIf(u.free).l
	return b
}
func (u *indexer[S, T]) maintainLeft(curI *S) {
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
func (u *indexer[S, T]) maintainRight(curI *S) {
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

func (u *indexer[S, T]) getV(i S) *T {
	return (*T)(unsafe.Add(u.vsHead, unsafe.Sizeof(*new(T))*uintptr(i)))
}

func (u *indexer[S, T]) InOrder(f func(*T) bool, buf []S) []S {
	if curI := u.root; buf == nil { //use morris traversal
	iter1:
		for curI != 0 {
			if u.getIf(curI).l == 0 {
				if !f(u.getV(curI - 1)) {
					break
				}
				curI = u.getIf(curI).r
			} else {
				for next := u.getIf(u.getIf(curI).l); ; next = u.getIf(next.r) {
					if next.r == 0 {
						next.r = curI
						curI = u.getIf(curI).l
						break
					} else if next.r == curI {
						next.r = 0
						if !f(u.getV(curI - 1)) {
							break iter1
						}
						curI = u.getIf(curI).r
						break
					}
				}

			}
		}
		for curI != 0 { //deplete the remaining traversal.
			if u.getIf(curI).l == 0 {
				curI = u.getIf(curI).r
			} else {
				for next := u.getIf(u.getIf(curI).l); ; next = u.getIf(next.r) {
					if next.r == 0 {
						next.r = curI
						curI = u.getIf(curI).l
						break
					} else if next.r == curI {
						next.r = 0
						curI = u.getIf(curI).r
						break
					}
				}
			}
		}
	} else { //use normal traversal
		for buf = buf[:0]; curI != 0; curI = u.getIf(curI).l {
			buf = append(buf, curI)
		}
		for len(buf) > 0 {
			curI, buf = buf[len(buf)-1], buf[:len(buf)-1]
			if !f(u.getV(curI - 1)) {
				break
			}
			for curI = u.getIf(curI).r; curI != 0; curI = u.getIf(curI).l {
				buf = append(buf, curI)
			}
		}
	}
	return buf
}

func (u *indexer[S, T]) Size() S {
	return u.getIf(u.root).sz
}
func (u *indexer[S, T]) Clear() {
	u.ifsLen = 1
}
func (u *indexer[S, T]) RankK(k S) *T {
	for curI := u.root; curI != 0; {
		if li := u.getIf(curI).l; k < u.getIf(li).sz {
			curI = li
		} else if k > u.getIf(li).sz {
			k -= u.getIf(li).sz + 1
			curI = u.getIf(curI).r
		} else {
			return u.getV(curI - 1)
		}
	}
	return nil
}
