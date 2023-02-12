package Trees

import (
	"golang.org/x/exp/constraints"
)

// A node in the SBTree
// The zero value is meaningless.
type node[T any, S constraints.Unsigned] struct {
	v    T
	sz   S
	l, r nodePtr[T, S]
}

// Pointer to a node
// nil Pointer is meaningless. A nodePtr is considered to be nil if the
// pointer is equal to the nilPtr in SBTree. The value of this node has
// both node.l, node.r = itself, and sz=0. v is the zero value of T
type nodePtr[T any, S constraints.Unsigned] *node[T, S]

// rotateLeft performs a left rotation on nodePtr n. n is passed by reference in order
// to modify its content.
// Time: O(1); Space: O(1)
func (u *base[T, S]) rotateLeft(n *nodePtr[T, S]) {
	r := *n
	rc := r.r
	r.r = rc.l
	rc.l = r
	rc.sz = r.sz
	r.sz = r.l.sz + r.r.sz + 1
	*n = rc
}

// rotateRight performs a left rotation on nodePtr n. n is passed by reference in order
// to modify its content.
// Time: O(1); Space: O(1)
func (u *base[T, S]) rotateRight(n *nodePtr[T, S]) {
	r := *n
	lc := r.l
	r.l = lc.r
	lc.r = r
	lc.sz = r.sz
	r.sz = r.l.sz + r.r.sz + 1
	*n = lc
}

// SBTree is a binary search tree with no repeated values. It maintains
// balance through rotations by checking the sizes of subtrees.
// T is the type of values it will hold, S is the type of the variables
// used for storing the sizes of different subtrees.
// This struct holds a root pointer and a corresponding nilPtr used
// as nil described in nodePtr.
// This tree needs to keep track of the sizes of each subtree, so the additional
// memory cost is size(S)*n.
// The worst case height of the tree is less than f(n)=1.44*log2(n+1.5)-1.33. So the height D
// of the tree is of O(log n). However, on average, D=log2(N).
// Note that due to the way uint works in Go, and that the Tree interface
// defines the return value of some functions to be uint. S shouldn't be
// any type that will cause overflow when converted to uint. For example,
// uint on 32 bit machine is uint32, if S=uint64, then calling Size() can
// potentially result in undefined values as uint64 would cause overflow
// if converted to uint32. Generally, you should let S be a wide upperbound
// for the size of the tree.
type base[T any, S constraints.Unsigned] struct {
	//the root of the tree. It should be nilPtr initially. nilPtr is the pointer used instead of nil here, it follows the description in nodePtr
	root, nilPtr nodePtr[T, S]
}

// Size returns the size of the tree.
// Time: O(1); Space: O(1)
func (u *base[T, S]) Size() uint {
	return uint(u.root.sz)
}

func (u *base[T, S]) maintainLeft(curPtr *nodePtr[T, S]) {
	cur := *curPtr
	if rc, lc := cur.r, cur.l; lc.l.sz > rc.sz {
		u.rotateRight(curPtr)
	} else if lc.r.sz > rc.sz {
		u.rotateLeft(&cur.l)
		u.rotateRight(curPtr)
	} else {
		return
	}
	u.maintainLeft(&cur.l)
	u.maintainRight(&cur.r)
	u.maintainLeft(curPtr)
	u.maintainRight(curPtr)

}

func (u *base[T, S]) maintainRight(curPtr *nodePtr[T, S]) {
	cur := *curPtr
	if rc, lc := cur.r, cur.l; rc.r.sz > lc.sz {
		u.rotateLeft(curPtr)
	} else if rc.l.sz > lc.sz {
		u.rotateRight(&cur.r)
		u.rotateLeft(curPtr)
	} else {
		return
	}
	u.maintainLeft(&cur.l)
	u.maintainRight(&cur.r)
	u.maintainLeft(curPtr)
	u.maintainRight(curPtr)

}

// maintain the subtree rooting at cur recursively to satisfy the base properties
// using rotateLeft and rotateRight.
// right Bigger indicates whether the right subtree is larger than the left,
// this is for removing redundant size comparisons.
// curPtr is passed by reference.
// Time: amortized O(1)
//func (u *base[T, S]) maintain(curPtr *nodePtr[T, S], rightBigger bool) {
//	cur := *curPtr
//	if rc, lc := cur.r, cur.l; rightBigger {
//		if rc.r.sz > lc.sz {
//			u.rotateLeft(curPtr)
//		} else if rc.l.sz > lc.sz {
//			u.rotateRight(&cur.r)
//			u.rotateLeft(curPtr)
//		} else {
//			return
//		}
//	} else {
//		if lc.l.sz > rc.sz {
//			u.rotateRight(curPtr)
//		} else if lc.r.sz > rc.sz {
//			u.rotateLeft(&cur.l)
//			u.rotateRight(curPtr)
//		} else {
//			return
//		}
//	}
//	u.maintain(&cur.l, false)
//	u.maintain(&cur.r, true)
//	u.maintain(curPtr, false)
//	u.maintain(curPtr, true)
//
//}

// this is just splitting the above maintain into 2 pieces
func (u *base[T, S]) maintain(curPtr *nodePtr[T, S], rightBigger bool) {
	if rightBigger {
		u.maintainRight(curPtr)
	} else {
		u.maintainLeft(curPtr)
	}
}

// Minimum [Tree.Minimum]
// Time: O(D); Space: O(1)
func (u *base[T, S]) Minimum() (T, bool) {
	if cur := u.root; cur == u.nilPtr {
		return cur.v, false
	} else {
		for cur.l != u.nilPtr {
			cur = cur.l
		}
		return cur.v, true
	}
}

// Maximum [Tree.Maximum]
// Time: O(D); Space: O(1)
func (u *base[T, S]) Maximum() (T, bool) {
	if cur := u.root; cur == u.nilPtr {
		return cur.v, false
	} else {
		for cur.r != u.nilPtr {
			cur = cur.r
		}
		return cur.v, true
	}
}

func (u *base[T, S]) avgDepth(cur nodePtr[T, S], h uint) uint {
	a := h
	if cur.l != u.nilPtr {
		a += u.avgDepth(cur.l, h+1)
	}
	if cur.r != u.nilPtr {
		a += u.avgDepth(cur.r, h+1)
	}
	return a
}

func (u *base[T, S]) averageDepth() uint {
	return u.avgDepth(u.root, 0) / u.Size()
}

func (u *base[T, S]) _Print(c nodePtr[T, S], d uint) {
	if c == u.nilPtr {
		return
	} else {
		println("node", c.v, "depth", d)
		u._Print(c.l, d+1)
		u._Print(c.r, d+1)
	}
}

func (u *base[T, S]) print() {
	u._Print(u.root, 0)
}

// InOrder [Tree.InOrder]
// Time: f(): amortized O(1) at each call to the returned function. Space: O(1)
func (u *base[T, S]) InOrder() func() (T, bool) {
	cur := u.root
	return func() (r T, has bool) {
		if cur == u.nilPtr {
			return
		} else {
			has = true
			for cur != u.nilPtr {
				if cur.l == u.nilPtr {
					r = cur.v
					cur = cur.r
					break
				} else {
					p := cur.l
					for p.r != u.nilPtr && p.r != cur {
						p = p.r
					}
					if p.r == u.nilPtr {
						p.r = cur
						cur = cur.l
					} else {
						p.r = u.nilPtr
						r = cur.v
						cur = cur.r
						break
					}
				}
			}
			return
		}

	}
}

// KLargest [Tree.KLargest]
// Returns (x,true) if k<= Size(), otherwise (0,false).
// This function utilizes the fact that base balances according to the
// sizes of each subtree to provide O(D) performance with very small constant.
// Time: O(D); Space: O(1)
func (u *base[T, S]) KLargest(k uint) (T, bool) {
	if cur, t := u.root, S(k); t <= cur.sz {
		for cur != u.nilPtr {
			if t < cur.l.sz+1 {
				cur = cur.l
			} else if t == cur.l.sz+1 {
				break
			} else {
				t -= cur.l.sz + 1
				cur = cur.r
			}
		}
		return cur.v, true
	} else {
		return *new(T), false
	}

}
