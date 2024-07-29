package Trees

import (
	"golang.org/x/exp/constraints"
	"math/bits"
	"unsafe"
)

// A node in the Tree
// The zero value is meaningful.
type info[S constraints.Unsigned] struct {
	l, r, sz S
}

type base[T any, S constraints.Unsigned] struct {
	root, free, ifsLen S              // free is the beginning of the linked list that contains all the free indexes; info[S]::l represents next.
	ifsHead            unsafe.Pointer // ifs[0] is zero value, which is a 0 size loopback. all index are based on ifs. len(ifs)=size+1
	vsHead             unsafe.Pointer // v[i] corresponds to ifs[i+1]. len(vs)=size
}

func (u *base[T, S]) getIf(i S) *info[S] {
	return (*info[S])(unsafe.Add(u.ifsHead, unsafe.Sizeof(info[S]{})*uintptr(i)))
}
func (u *base[T, S]) rotateLeft(ni *S) {
	n := u.getIf(*ni)
	rci := n.r

	n.r = u.getIf(rci).l
	u.getIf(rci).l, u.getIf(rci).sz = *ni, n.sz
	n.sz = u.getIf(n.l).sz + u.getIf(n.r).sz + 1
	*ni = rci
}

func (u *base[T, S]) rotateRight(ni *S) {
	n := u.getIf(*ni)
	lci := n.l

	n.l = u.getIf(lci).r
	u.getIf(lci).r, u.getIf(lci).sz = *ni, n.sz
	n.sz = u.getIf(n.l).sz + u.getIf(n.r).sz + 1
	*ni = lci
}

// addFree index once.
func (u *base[T, S]) addFree(a S) {
	u.getIf(a).l = u.free
	u.free = a
}

// popFree index once. Returns 0 when there's no free index(when u.free==0).
func (u *base[T, S]) popFree() S {
	b := u.free
	u.free = u.getIf(u.free).l
	return b
}
func (u *base[T, S]) maintainLeft(curI *S) {
	cur := u.getIf(*curI)
	if rcsz, lc := u.getIf(cur.r).sz, *u.getIf(cur.l); u.getIf(lc.l).sz > rcsz {
		u.rotateRight(curI)
	} else if u.getIf(lc.r).sz > rcsz {
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
	cur := u.getIf(*curI)
	if rc, lcsz := *u.getIf(cur.r), u.getIf(cur.l).sz; u.getIf(rc.r).sz > lcsz {
		u.rotateLeft(curI)
	} else if u.getIf(rc.l).sz > lcsz {
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

func (u *base[T, S]) getV(i S) *T {
	return (*T)(unsafe.Add(u.vsHead, unsafe.Sizeof(*new(T))*uintptr(i)))
}

// InOrder traversal of teh tree. When st==nil, uses morris traversal; otherwise, use normal stack based iterative traversal.
func (u *base[T, S]) InOrder(f func(*T) bool, st []S) []S {
	if curI := u.root; st == nil { //use morris traversal
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
		for st = st[:0]; curI != 0; curI = u.getIf(curI).l {
			st = append(st, curI)
		}
		for len(st) > 0 {
			curI, st = st[len(st)-1], st[:len(st)-1]
			if !f(u.getV(curI - 1)) {
				break
			}
			for curI = u.getIf(curI).r; curI != 0; curI = u.getIf(curI).l {
				st = append(st, curI)
			}
		}
	}
	return st
}

func (u *base[T, S]) Size() S {
	return u.getIf(u.root).sz
}

// Clear the tree, also resets memory of underlying value array to size if reset is true. O(1) if reset==false. O(size) if reset==true. Doesn't allocate new arrays.
func (u *base[T, S]) Clear(reset bool) {
	if reset {
		vs := unsafe.Slice((*T)(u.vsHead), u.ifsLen-1)
		for i := range vs {
			vs[i] = *new(T)
		}
	}
	u.ifsLen, u.free, u.root = 1, 0, 0
}

// RankK element in tree, starting from 0.
func (u *base[T, S]) RankK(k S) *T {
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

// overflowMid is equivalent to (a+b)/2 but deals with overflow.
func overflowMid[S constraints.Unsigned](a, b S) S {
	d := a + b
	overflowed := d < a
	c := (^S(0)) >> S(*(*byte)(unsafe.Pointer(&overflowed)))
	return d>>1 | ^c
}

// buildIfs array of size vsLen to represent a complete binary tree.
func buildIfs[S constraints.Unsigned](vsLen S) (root S, ifs []info[S]) {
	ifs = make([]info[S], vsLen+1)
	st := make([][3]S, 0, bits.Len64(uint64(vsLen))) //[left,right,mid]
	{
		root = (1 + vsLen) >> 1
		st = append(st, [3]S{1, vsLen, root})
	}
	for len(st) > 0 {
		top := st[len(st)-1]
		st = st[:len(st)-1]
		ifs[top[2]].sz = top[1] - top[0] + 1
		if top[0] < top[2] {
			nr := top[2] - 1
			ifs[top[2]].l = overflowMid(top[0], nr)
			st = append(st, [3]S{top[0], nr, ifs[top[2]].l})
		}
		if top[2] < top[1] {
			nl := top[2] + 1
			ifs[top[2]].r = overflowMid(nl, top[1])
			st = append(st, [3]S{nl, top[1], ifs[top[2]].r})
		}
	}
	return
}
