package Trees

import (
	"golang.org/x/exp/constraints"
)

// CSBTree is the version of SBTree for user-defined struct satisfying Ordered interface.
// All methods are implemented exactly as SBTree except for using Ordered.LessThan and
// Ordered.Equals for comparisons.
type CSBTree[T Ordered, S constraints.Unsigned] struct {
	root   nodePtr[T, S]
	nilPtr nodePtr[T, S]
}

// MakeCSBTree is the CSBTree equivalence of MakeSBTree
func MakeCSBTree[T Ordered, S constraints.Unsigned]() *CSBTree[T, S] {
	z := new(node[T, S])
	z.l, z.r = z, z
	return &CSBTree[T, S]{z, z}
}

// BuildCSBTree is the CSBTree equivalence of BuildSBTree
func BuildCSBTree[T Ordered, S constraints.Unsigned](sli []T, safe bool) *CSBTree[T, S] {
	z := new(node[T, S])
	z.l, z.r = z, z
	var build func([]T) nodePtr[T, S]
	if safe {
		build = func(s []T) nodePtr[T, S] {
			if len(s) > 0 {
				mid := len(s) >> 1
				l, r := build(s[0:mid]), build(s[mid+1:])
				if (l == z || l.v.LessThan(s[mid])) && (r == z || s[mid].LessThan(r.v)) {
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
	return &CSBTree[T, S]{build(sli), z}
}

func (u *CSBTree[T, S]) Size() uint {
	return uint(u.root.sz)
}

func (u *CSBTree[T, S]) insert(curPtr *nodePtr[T, S], v T) bool {
	if cur := *curPtr; cur == u.nilPtr {
		*curPtr = &node[T, S]{v, u.nilPtr, u.nilPtr, 1}
		return true
	} else {
		inserted := false
		if v.LessThan(cur.v) {
			inserted = u.insert(&cur.l, v)
		} else if v.Equals(cur.v) {
			return false
		} else {
			inserted = u.insert(&cur.r, v)
		}
		if inserted {
			cur.sz++
			u.maintain(curPtr, !v.LessThan(cur.v))
		}
		return inserted
	}

}

func (u *CSBTree[T, S]) Insert(v T) bool {
	return u.insert(&u.root, v)
}

func (u *CSBTree[T, S]) remove(curPtr *nodePtr[T, S], v T) bool {
	if cur := *curPtr; cur == u.nilPtr {
		return false
	} else {
		deleted := false
		if v.LessThan(cur.v) {
			deleted = u.remove(&cur.l, v)
		} else if v.Equals(cur.v) {
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

func (u *CSBTree[T, S]) Remove(v T) bool {
	return u.remove(&u.root, v)
}

func (u CSBTree[T, S]) Has(v T) bool {
	for cur := u.root; cur != u.nilPtr; {
		if v.LessThan(cur.v) {
			cur = cur.l
		} else if v.Equals(cur.v) {
			return true
		} else {
			cur = cur.r
		}
	}
	return false
}

func (u CSBTree[T, S]) Minimum() (T, bool) {
	if cur := u.root; cur == u.nilPtr {
		return cur.v, false
	} else {
		for cur.l != u.nilPtr {
			cur = cur.l
		}
		return cur.v, true
	}
}

func (u CSBTree[T, S]) Maximum() (T, bool) {
	if cur := u.root; cur == u.nilPtr {
		return cur.v, false
	} else {
		for cur.r != u.nilPtr {
			cur = cur.r
		}
		return cur.v, true
	}
}

func (u CSBTree[T, S]) minDepth(c nodePtr[T, S], cd uint) uint {
	if c == u.nilPtr {
		return cd - 1
	}
	return Min(u.minDepth(c.l, cd+1), u.minDepth(c.r, cd+1))
}

func (u CSBTree[T, S]) MinDepth() uint {
	return u.minDepth(u.root, 0)
}

func (u CSBTree[T, S]) maxDepth(c nodePtr[T, S], cd uint) uint {
	if c == u.nilPtr {
		return cd - 1
	}
	return Max(u.maxDepth(c.l, cd+1), u.maxDepth(c.r, cd+1))
}

func (u CSBTree[T, S]) MaxDepth() uint {
	return u.maxDepth(u.root, 0)
}

func (u CSBTree[T, S]) _Print(c nodePtr[T, S], d uint) {
	if c == u.nilPtr {
		return
	} else {
		println("node", c.v, "depth", d)
		u._Print(c.l, d+1)
		u._Print(c.r, d+1)
	}
}

func (u CSBTree[T, S]) Print() {
	u._Print(u.root, 0)
}

func (u CSBTree[T, S]) InOrder() func() (T, bool) {
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

func (u CSBTree[T, S]) Predecessor(v T) (T, bool) {
	cur, p := u.root, u.nilPtr
	for cur != u.nilPtr {
		if v.LessThan(cur.v) || v.Equals(cur.v) {
			cur = cur.l
		} else {
			p = cur
			cur = cur.r
		}
	}
	return p.v, p != u.nilPtr
}

func (u CSBTree[T, S]) Successor(v T) (T, bool) {
	cur, p := u.root, u.nilPtr
	for cur != u.nilPtr {
		if v.LessThan(cur.v) {
			p = cur
			cur = cur.l
		} else {
			cur = cur.r
		}
	}
	return p.v, p != u.nilPtr
}

func (u CSBTree[T, S]) KLargest(k uint) (T, bool) {
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

func (u CSBTree[T, S]) RankOf(v T) uint {
	cur := u.root
	var ra S = 0
	for cur != u.nilPtr {
		if v.LessThan(cur.v) {
			cur = cur.l
		} else if v.Equals(cur.v) {
			return uint(ra + cur.l.sz + 1)
		} else {
			ra += cur.l.sz + 1
			cur = cur.r
		}
	}
	return 0
}
