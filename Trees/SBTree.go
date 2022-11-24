package Trees

import (
	"golang.org/x/exp/constraints"
)

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
type SBTree[T constraints.Ordered, S constraints.Unsigned] struct {
	base[T, S]
}

// MakeSBTree returns a SBTree satisfying the above definitions for nilPtr, root, and types.
// SBTree shouldn't be created directly using struct literal.
func MakeSBTree[T constraints.Ordered, S constraints.Unsigned]() *SBTree[T, S] {
	z := new(node[T, S])
	z.l, z.r = z, z
	return &SBTree[T, S]{base[T, S]{z, z}}
}

// BuildSBTree builds a SBTree using the given sorted slice recursively. This is faster than
// repeatedly calling Insert. The word "set" is used to show that there shouldn't be any repeated
// element.
// The given slice must be sorted
// in ascending order and mustn't contain duplicate elements(satisfying SBTree conditions), otherwise
// there will be corrupt structures.
// Time: O(n).
func BuildSBTree[T constraints.Ordered, S constraints.Unsigned](sli []T) *SBTree[T, S] {
	z := new(node[T, S])
	z.l, z.r = z, z
	var build func([]T) nodePtr[T, S]
	build = func(s []T) nodePtr[T, S] {
		if len(s) > 0 {
			mid := len(s) >> 1
			return &node[T, S]{s[mid], S(len(s)), build(s[0:mid]), build(s[mid+1:])}
		} else {
			return z
		}
	}
	return &SBTree[T, S]{base[T, S]{build(sli), z}}
}

// insert the value v to the subtree rooting at cur recursively. cur is
// passed by reference. A successful insertion returns true. A failed insertion
// happens when the value is already in u, in which case it returns false.
func (u *SBTree[T, S]) insert(curPtr *nodePtr[T, S], v T) bool {
	if cur := *curPtr; cur == u.nilPtr {
		*curPtr = &node[T, S]{v, 1, u.nilPtr, u.nilPtr}
		return true
	} else {
		inserted := false
		if v < cur.v {
			inserted = u.insert(&cur.l, v)
		} else if v == cur.v {
			return false
		} else {
			inserted = u.insert(&cur.r, v)
		}
		if inserted {
			cur.sz++
			u.maintain(curPtr, v >= cur.v)
		}
		return inserted
	}

}

// Insert [Tree.Insert]. Recursive.
// It is a wrapper for insert.
// Time: O(D)
func (u *SBTree[T, S]) Insert(v T) bool {
	return u.insert(&u.root, v)
}

// remove an element v from the subtree rooting at cur recursively. cur is
// passed by reference. Returns false if the removal failed(v
// doesn't exist in u), otherwise true. Note that remove doesn't call maintain,
// this means that D isn't O(f(n)) after calling removal, but instead O(f(i))
// where i is the number of elements before this sequence of calls to remove.
// After calling insert once D will again be O(log n).
// Time: O(D)
func (u *SBTree[T, S]) remove(curPtr *nodePtr[T, S], v T) bool {
	if cur := *curPtr; cur == u.nilPtr {
		return false
	} else {
		deleted := false
		if v < cur.v {
			deleted = u.remove(&cur.l, v)
		} else if v == cur.v {
			deleted = true
			if cur.l == u.nilPtr {
				*curPtr = cur.r
			} else if cur.r == u.nilPtr {
				*curPtr = cur.l
			} else {
				t := &cur.r
				for (*t).l != u.nilPtr {
					(*t).sz--
					t = &(*t).l
				}
				cur.v = (*t).v
				*t = u.nilPtr
			}
		} else {
			deleted = u.remove(&cur.r, v)
		}
		if deleted {
			cur.sz--
		}
		return deleted
	}

}

// Remove [Tree.Remove]. Recursive.
// It is a wrapper for remove.
// Time: O(D)
func (u *SBTree[T, S]) Remove(v T) bool {
	return u.remove(&u.root, v)
}

// Has [Tree.Has]
// Time: O(D); Space: O(1)
func (u SBTree[T, S]) Has(v T) bool {
	for cur := u.root; cur != u.nilPtr; {
		if v < cur.v {
			cur = cur.l
		} else if v == cur.v {
			return true
		} else {
			cur = cur.r
		}
	}
	return false
}

// Predecessor [Tree.Predecessor]
// Time: O(D); Space: O(1)
func (u SBTree[T, S]) Predecessor(v T) (T, bool) {
	cur, p := u.root, u.nilPtr
	for cur != u.nilPtr {
		if v <= cur.v {
			cur = cur.l
		} else {
			p = cur
			cur = cur.r
		}
	}
	return p.v, p != u.nilPtr
}

// Successor [Tree.Successor]
// Time: O(D); Space: O(1)
func (u SBTree[T, S]) Successor(v T) (T, bool) {
	cur, p := u.root, u.nilPtr
	for cur != u.nilPtr {
		if v < cur.v {
			p = cur
			cur = cur.l
		} else {
			cur = cur.r
		}
	}
	return p.v, p != u.nilPtr
}

// RankOf [Tree.RankOf]
// This function utilizes the fact that SBTree balances according to the
// sizes of each subtree to provide O(D) performance with very small constant.
// Time: O(D); Space: O(1)
func (u SBTree[T, S]) RankOf(v T) uint {
	cur := u.root
	var ra S = 0
	for cur != u.nilPtr {
		if v < cur.v {
			cur = cur.l
		} else if v == cur.v {
			return uint(ra + cur.l.sz + 1)
		} else {
			ra += cur.l.sz + 1
			cur = cur.r
		}
	}
	return 0
}

func (u SBTree[T, S]) corrupt(cur nodePtr[T, S]) bool {
	if cur.l != u.nilPtr {
		if cur.l.v >= cur.v || u.corrupt(cur.l) {
			return true
		}
	}
	if cur.r != u.nilPtr {
		if cur.v >= cur.r.v || u.corrupt(cur.r) {
			return true
		}
	}
	return false
}

func (u SBTree[T, S]) Corrupt() bool {
	return u.corrupt(u.root)
}
