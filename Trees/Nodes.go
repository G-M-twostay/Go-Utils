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
