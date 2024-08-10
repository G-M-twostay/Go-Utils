/*
Package Trees implements Size Balanced Tree using arrays in a minimally recursive approach. Elements in the tree is unique. The
left most element is the smallest element.

# Size Balanced Tree
Size Balanced Tree(SBTree) is a binary tree that balances itself naturally using subtree sizes. It's flatter than other prevalent implementations
like RBTree. It's also faster than those implementations as indicated by the benchmark and th original paper. Here its superiority
over RBTree is also demonstrated.

# Using Arrays
SBTree is quite tricky to implement using pointers Therefore, like the original paper, we use arrays
to implement it. However, 1 very common problem of using arrays is indexing. A tree that's good for general use must utilize all
indexes in the array, it mustn't contain holes. The most intuitive "left=2^n, right=2^n+1" approach will leave lots of wasted indexes
should the tree not be perfect. Maintaining a top index count doesn't have this issue, but, like the above, are prone to creating and forgetting
hole indexes when deletion is involved. In this implementation, I employed a way to record the hole indexes without additional memory
and can be stored and retried in O(1) time. This enables this implementation to utilize the advantages of using arrays while
not worry about wasting indexes. In fact, this implementation is guaranteed to exhaust existing array before growing.

In Go, array are indexed using `int`, so we restrict the indexing type to `uint` at most. This means that you won't be able
to use `uint64` as indexes on 32bit machine. This is pretty reasonable.

# Modifications
This implementation is slightly different from the original implementation. First, maintaining is split into 2 orientations
to make it more intuitive without degrading performance. Second, unlike the original, deletion will sometimes balance the tree
afterward. See [base.maintainLeft], [Tree.Del] for details.

# Minimal Recursions
This implementation is minimally recursive. Mutating operations are mostly iterative while readonly operations are usually constant space
iterative. Balancing operations, namely [base.maintainLeft] and [base.maintainRight], are recursive because they are hard to express
iteratively, but they're amortized O(1). [Tree.Add] and [Tree.Del] requires backtracking, thus we've to use a something to store
the stack. Fortunately, they both involve very simple backtracking and can be implemented efficiently and easily by using a slice.
Also, using the height guarantee of SBTree, we can actually allocate this slice beforehand and reuse it as long as we know how big will the tree
be.

For traversing, I provided both a normal stack based traversal and a constant space Morris traversal. This is because Morris
traversal is a mutating operation, so only 1 traversal may happen at any time. Also, Morris traversal doesn't support early
termination by its nature, so in some cases, normal traversals may still be helpful.

# Concurrency
This isn't safe for concurrent usages. Writes shouldn't happen at the same time with other reads or writes. Section below
is for those who are interested in details. Generally, you shouldn't worry about it.

Due to the existence of Morris traversal, we can't classify operations into simple reads and writes. We define the followings.
Read operations:
  - R0: Reads indexing data, [info.l] and [info.r], in [base.ifsHead].
  - R1: Reads value data as stored in [base.vsHead].
  - R2: Reads balancing data, [info.sz], in [base.ifsHead].

Write Operations:
  - W0: Corresponds to R0.
  - W1: Corresponds to R1.
  - W2: Corresponds to R2.

Multiple reads of any types may happen at the same time. Only 1 write of a single type can happen at the same time. This means
that reads and write of different type can happen at the same type. We will classify all public functions using these definitions.

This implementation returns pointers instead of values. Generally, you shouldn't keep track of the pointer and should immediately
dereference it for safety. However, it's fine to use the pointers later under some occasions. A W1 operation can invalidate
the pointer, so pointers prior to the beginning of a W1 operation should be discarded. Other than this, just make sure you're not writing to
any pointers during R1 operations. Writing to the pointer can also break the structure of the tree, leading to undefined behaviors.
Unfortunately, Go doesn't have const pointers like C++, so it's up to the users to either not write to the pointer or ensure
that the write operation doesn't break its standing in the tree.

# Performance
Like any other balanced binary trees, most operations are all O(log(N)) in time while the memory usage is O(N). Section below
is for those interested in the constant factors.

Let A0 be the backing [info] array [base.ifsHead], A1 be the backing value array [base.vsHead]. Note that len(A1)=len(A0)-1. len(Ai)<=cap(Ai) and
cap(Ai) is linear to len(Ai).
Let B be the total number of unique insertions, calls to [Tree.Add] that returns true. Let C be the total number of elements
in the array. Note that C<=len(A1) Let D be the actual average height of the tree, D<=e(C) or e(B) where e is a function that gives the theoretically max height
of the tree given C. e(x)~=1.4404*log2(x) disregarding some small constant offset. In practice, this can be estimated easily by `bits.Len(x)*7/5`.
We will use these terms when giving a detailed description of performance for each function.

The non-constant term of memory usage for value type T and index type S is sizeof(T)*cap(A1)+sizeof(S)*3*cap(A1). As you may notice,
it's extremely compact in terms of memory. [info] has no padding itself, so putting T and [info] each into their own respective
array is the most compact format. Also, on 64-bit machine, using uint32 as indexing will mean that each node only takes an
additional 12 bytes space, or 1.5 words, this isn't achievable on any pointer based implementations assuming that pointers
take 1 full word.
*/
package Trees

import (
	"math/bits"
	"reflect"
	"unsafe"
)

// Indexable are types that can be used as indexes in the tree.
type Indexable interface {
	~byte | ~uint16 | ~uint32 | ~uint
} // Exclude uint64 from being used as indexes.

// A node in the Tree. The zero value is meaningful.
// info[S] is at most 3 words of memory, which the same as a slice, so it can be cheaply value copied.
type info[S Indexable] struct {
	l, r, sz S
}

type base[T any, S Indexable] struct {
	ifsHead            unsafe.Pointer // ifs[0] is zero value, which is a 0 size loopback. all index are based on ifs. len(ifs)=size+1
	vsHead             unsafe.Pointer // v[i] corresponds to ifs[i+1]. len(vs)=size
	caps               [2]int         // caps[0]=cap(ifs), caps[1]=cap(vs)
	root, free, ifsLen S              // free is the beginning of the linked list that contains all the free indexes; info[S].l represents next.
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

func (u *base[T, S]) addFree(a S) {
	u.getIf(a).l = u.free
	u.free = a
}

func (u *base[T, S]) popFree() S {
	b := u.free
	u.free = u.getIf(u.free).l
	return b
}

/*
Maintaining is split into 2 functions instead of the single one in original
paper so that we don't need the flag bool. maintainLeft is equivalent to the
!flag case in original implementation.
*/

func (u *base[T, S]) maintainLeft(curI *S) {
	cur := u.getIf(*curI)
	if rcsz, lc := u.getIf(cur.r).sz, u.getIf(cur.l); u.getIf(lc.l).sz > rcsz {
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
	if rc, lcsz := u.getIf(cur.r), u.getIf(cur.l).sz; u.getIf(rc.r).sz > lcsz {
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
// Time: O(n). Space: O(1) when using Morris Traversal, O(sizeof(S)*D) when using normal traversal.
// Type: W0 when Morris Traversal, R0 when normal traversal.
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

// InOrderR is the reverse in order traversal.
func (u *base[T, S]) InOrderR(f func(*T) bool, st []S) []S {
	if curI := u.root; st == nil { //use morris traversal
	iter1:
		for curI != 0 {
			if u.getIf(curI).r == 0 {
				if !f(u.getV(curI - 1)) {
					break
				}
				curI = u.getIf(curI).l
			} else {
				for next := u.getIf(u.getIf(curI).r); ; next = u.getIf(next.l) {
					if next.l == 0 {
						next.l = curI
						curI = u.getIf(curI).r
						break
					} else if next.l == curI {
						next.l = 0
						if !f(u.getV(curI - 1)) {
							break iter1
						}
						curI = u.getIf(curI).l
						break
					}
				}

			}
		}
		for curI != 0 { //deplete the remaining traversal.
			if u.getIf(curI).r == 0 {
				curI = u.getIf(curI).l
			} else {
				for next := u.getIf(u.getIf(curI).r); ; next = u.getIf(next.l) {
					if next.l == 0 {
						next.l = curI
						curI = u.getIf(curI).r
						break
					} else if next.l == curI {
						next.l = 0
						curI = u.getIf(curI).l
						break
					}
				}
			}
		}
	} else { //use normal traversal
		for st = st[:0]; curI != 0; curI = u.getIf(curI).r {
			st = append(st, curI)
		}
		for len(st) > 0 {
			curI, st = st[len(st)-1], st[:len(st)-1]
			if !f(u.getV(curI - 1)) {
				break
			}
			for curI = u.getIf(curI).l; curI != 0; curI = u.getIf(curI).r {
				st = append(st, curI)
			}
		}
	}
	return st
}

// Size of tree.
// Type: R2.
func (u *base[T, S]) Size() S {
	return u.getIf(u.root).sz
}

// Clear the tree, also zeroes the value array's values if zero is true. Doesn't allocate new arrays.
// Time: O(1) when !zero, O(len(A1)) when zero. Space: O(1).
// Type: W0, W1, W2.
func (u *base[T, S]) Clear(zero bool) {
	if zero {
		clear(unsafe.Slice((*T)(u.vsHead), u.ifsLen-1))
	}
	u.ifsLen, u.free, u.root = 1, 0, 0
}

// RankK element in tree, starting from 0.
// Time: O(D). Space: O(1).
// Type: R0, R2.
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

// mid is equivalent to (a+b)/2 but deals with overflow.
func mid[S Indexable](a, b S) S {
	low, high := bits.Add(uint(a), uint(b), 0)
	r, _ := bits.Div(high, low, 2)
	return S(r)
}

// buildIfs array of size vsLen to represent a complete binary tree.
func buildIfs[S Indexable](vsLen S, st [][3]S) (root S, ifs []info[S]) {
	ifs = make([]info[S], vsLen+1)
	{
		root = (1 + vsLen) >> 1
		st = append(st, [3]S{1, vsLen, root}) //[left,right,mid]
	}
	for len(st) > 0 {
		top := st[len(st)-1]
		st = st[:len(st)-1]
		ifs[top[2]].sz = top[1] - top[0] + 1
		if top[0] < top[2] {
			nr := top[2] - 1
			ifs[top[2]].l = mid(top[0], nr)
			st = append(st, [3]S{top[0], nr, ifs[top[2]].l})
		}
		if top[2] < top[1] {
			nl := top[2] + 1
			ifs[top[2]].r = mid(nl, top[1])
			st = append(st, [3]S{nl, top[1], ifs[top[2]].r})
		}
	}
	return
}

// Compact the tree by copying the content to a smaller array and filling the holes if necessary.
// Time: O(C). Space: sizeof(T)*C+sizeof(S)*3*(C+1).
// Type: W0, W1, W2.
func (u *base[T, S]) Compact() {
	if u.free == 0 {
		{
			a := make([]T, u.ifsLen-1)
			copy(a, unsafe.Slice((*T)(u.vsHead), u.ifsLen-1))
			u.vsHead, u.caps[1] = unsafe.Pointer(unsafe.SliceData(a)), cap(a)
		}
		a := make([]info[S], u.ifsLen)
		copy(a, unsafe.Slice((*info[S])(u.ifsHead), u.ifsLen))
		u.ifsHead, u.caps[0] = unsafe.Pointer(unsafe.SliceData(a)), cap(a)
	} else {
		u.free = 0
		{
			a := make([]T, 0, u.ifsLen-1)
			u.InOrder(func(vp *T) bool {
				a = append(a, *vp)
				return true
			}, nil)
			u.vsHead, u.caps[1] = unsafe.Pointer(unsafe.SliceData(a)), cap(a)
		}
		var a []info[S]
		u.root, a = buildIfs[S](u.Size(), *(*[][3]S)(unsafe.Pointer(&reflect.SliceHeader{uintptr(u.ifsHead), 0, u.caps[0]})))
		u.ifsHead, u.caps[0], u.ifsLen = unsafe.Pointer(unsafe.SliceData(a)), cap(a), S(len(a))
	}
}
