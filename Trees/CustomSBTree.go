package Trees

import (
	"golang.org/x/exp/constraints"
)

// CSBTree is the version of SBTree for user-defined struct satisfying Ordered interface.
// All methods are implemented exactly as SBTree except for using Ordered.LessThan and
// Ordered.Equals for comparisons. Argument passed to Ordered.LessThan and Ordered.Equals
// will always be type T so no type checks are needed.
type CSBTree[T Ordered, S constraints.Unsigned] struct {
	base[T, S]
}

// MakeCSBTree is the CSBTree equivalence of MakeSBTree
func MakeCSBTree[T Ordered, S constraints.Unsigned]() *CSBTree[T, S] {
	z := new(node[T, S])
	z.l, z.r = z, z
	return &CSBTree[T, S]{base[T, S]{z, z}}
}

// BuildCSBTree is the CSBTree equivalence of BuildSBTree
func BuildCSBTree[T Ordered, S constraints.Unsigned](sli []T) *CSBTree[T, S] {
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
	return &CSBTree[T, S]{base[T, S]{build(sli), z}}
}

func (u *CSBTree[T, S]) insert(curPtr *nodePtr[T, S], v T) bool {
	if cur := *curPtr; cur == u.nilPtr {
		*curPtr = &node[T, S]{v, 1, u.nilPtr, u.nilPtr}
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

func (u CSBTree[T, S]) corrupt(cur nodePtr[T, S]) bool {
	if cur.l != u.nilPtr {
		if !cur.l.v.LessThan(cur.v) || u.corrupt(cur.l) {
			return true
		}
	}
	if cur.r != u.nilPtr {
		if !cur.v.LessThan(cur.r.v) || u.corrupt(cur.r) {
			return true
		}
	}
	return false
}

func (u CSBTree[T, S]) Corrupt() bool {
	return u.corrupt(u.root)
}
