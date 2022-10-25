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
	root   nodePtr[T, S] //the root of the tree. It should be nilPtr initially.
	nilPtr nodePtr[T, S] // nilPtr is the pointer used instead of nil here, it follows the description in nodePtr
}

// MakeSBTree returns a SBTree satisfying the above definitions for nilPtr, root, and types.
// SBTree shouldn't be created directly using struct literal.
func MakeSBTree[T constraints.Ordered, S constraints.Unsigned]() *SBTree[T, S] {
	z := new(node[T, S])
	z.l, z.r = z, z
	return &SBTree[T, S]{z, z}
}

// BuildSBTree builds a SBTree using the given sorted slice recursively. This is faster than
// repeatedly calling Insert. The word "set" is used to show that there shouldn't be any repeated
// element.
// The given slice must be sorted
// in ascending order and mustn't contain duplicate elements(satisfying SBTree conditions).
// If safe==true, this function will check if the conditions are met and panic with InvalidSliceError
// if the conditions are broken. Otherwise, this function won't perform the check, and it is
// up to the user to ensure the conditions are met(otherwise the tree will be corrupt). It's
// suggested to set safe to false if the conditions are met as this can reduce some redundant
// checks and associated memory costs.
// Time: O(n).
func BuildSBTree[T constraints.Ordered, S constraints.Unsigned](sli []T, safe bool) *SBTree[T, S] {
	z := new(node[T, S])
	z.l, z.r = z, z
	var build func([]T) nodePtr[T, S]
	if safe {
		build = func(s []T) nodePtr[T, S] {
			if len(s) > 0 {
				mid := len(s) >> 1
				l, r := build(s[0:mid]), build(s[mid+1:])
				if (l == z || l.v < s[mid]) && (r == z || s[mid] < r.v) {
					return &node[T, S]{s[mid], l, r, S(len(s))}
				} else {
					panic(InvalidSliceError{l.v, s[mid], s[mid], r.v})
				}
			} else {
				return z
			}
		}
	} else {
		build = func(s []T) nodePtr[T, S] {
			if len(s) > 0 {
				mid := len(s) >> 1
				return &node[T, S]{s[mid], build(s[0:mid]), build(s[mid+1:]), S(len(s))}
			} else {
				return z
			}
		}
	}
	return &SBTree[T, S]{build(sli), z}
}

// Size returns the size of the tree.
// Time: O(1); Space: O(1)
func (u *SBTree[T, S]) Size() uint {
	return uint(u.root.sz)
}

// maintain the subtree rooting at cur recursively to satisfy the SBTree properties
// using rotateLeft and rotateRight.
// right Bigger indicates whether the right subtree is larger than the left,
// this is for removing redundant size comparisons.
// curPtr is passed by reference.
// Time: amortized O(1)
func (u *SBTree[T, S]) maintain(curPtr *nodePtr[T, S], rightBigger bool) {
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

// insert the value v to the subtree rooting at cur recursively. cur is
// passed by reference. A successful insertion returns true. A failed insertion
// happens when the value is already in u, in which case it returns false.
func (u *SBTree[T, S]) insert(curPtr *nodePtr[T, S], v T) bool {
	if cur := *curPtr; cur == u.nilPtr {
		*curPtr = &node[T, S]{v, u.nilPtr, u.nilPtr, 1}
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

// Minimum [Tree.Minimum]
// Time: O(D); Space: O(1)
func (u SBTree[T, S]) Minimum() (T, bool) {
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
func (u SBTree[T, S]) Maximum() (T, bool) {
	if cur := u.root; cur == u.nilPtr {
		return cur.v, false
	} else {
		for cur.r != u.nilPtr {
			cur = cur.r
		}
		return cur.v, true
	}
}

func (u SBTree[T, S]) minDepth(c nodePtr[T, S], cd uint) uint {
	if c == u.nilPtr {
		return cd - 1
	}
	return Min(u.minDepth(c.l, cd+1), u.minDepth(c.r, cd+1))
}

func (u SBTree[T, S]) MinDepth() uint {
	return u.minDepth(u.root, 0)
}

func (u SBTree[T, S]) maxDepth(c nodePtr[T, S], cd uint) uint {
	if c == u.nilPtr {
		return cd - 1
	}
	return Max(u.maxDepth(c.l, cd+1), u.maxDepth(c.r, cd+1))
}

func (u SBTree[T, S]) MaxDepth() uint {
	return u.maxDepth(u.root, 0)
}

func (u SBTree[T, S]) _Print(c nodePtr[T, S], d uint) {
	if c == u.nilPtr {
		return
	} else {
		println("node", c.v, "depth", d)
		u._Print(c.l, d+1)
		u._Print(c.r, d+1)
	}
}

func (u SBTree[T, S]) Print() {
	u._Print(u.root, 0)
}

// InOrder [Tree.InOrder]
// Time: f(): amortized O(1) at each call to the returned function. Space: O(1)
func (u SBTree[T, S]) InOrder() func() (T, bool) {
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
					if p.r != cur {
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

// KLargest [Tree.KLargest]
// Returns (x,true) if k<= Size(), otherwise (0,false).
// This function utilizes the fact that SBTree balances according to the
// sizes of each subtree to provide O(D) performance with very small constant.
// Time: O(D); Space: O(1)
func (u SBTree[T, S]) KLargest(k uint) (T, bool) {
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
