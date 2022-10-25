package Trees

import "golang.org/x/exp/constraints"

// A node in the SBTree
// The zero value is meaningless.
type node[T any, S constraints.Unsigned] struct {
	v    T
	l, r nodePtr[T, S]
	sz   S
}

// Pointer to a node
// nil Pointer is meaningless. A nodePtr is considered to be nil if the
// pointer is equal to the nilPtr in SBTree. The value of this node has
// both node.l, node.r = itself, and sz=0. v is the zero value of T
type nodePtr[T any, S constraints.Unsigned] *node[T, S]

// rotateLeft performs a left rotation on nodePtr n. n is passed by reference in order
// to modify its content.
// Time: O(1); Space: O(1)
func rotateLeft[T any, S constraints.Unsigned](n *nodePtr[T, S]) {
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
func rotateRight[T any, S constraints.Unsigned](n *nodePtr[T, S]) {
	r := *n
	lc := r.l
	r.l = lc.r
	lc.r = r
	lc.sz = r.sz
	r.sz = r.l.sz + r.r.sz + 1
	*n = lc
}

// maintain the subtree rooting at cur recursively to satisfy the CSBTree properties
// using rotateLeft and rotateRight.
// right Bigger indicates whether the right subtree is larger than the left,
// this is for removing redundant size comparisons.
// curPtr is passed by reference.
// Time: amortized O(1)
func (u *CSBTree[T, S]) maintain(curPtr *nodePtr[T, S], rightBigger bool) {
	cur := *curPtr
	if rc, lc := cur.r, cur.l; rightBigger {
		if rc.r.sz > lc.sz {
			rotateLeft(curPtr)
		} else if rc.l.sz > lc.sz {
			rotateRight(&cur.r)
			rotateLeft(curPtr)
		} else {
			return
		}
	} else {
		if lc.l.sz > rc.sz {
			rotateRight(curPtr)
		} else if lc.r.sz > rc.sz {
			rotateLeft(&cur.l)
			rotateRight(curPtr)
		} else {
			return
		}
	}
	u.maintain(&cur.l, false)
	u.maintain(&cur.r, true)
	u.maintain(curPtr, false)
	u.maintain(curPtr, true)

}
