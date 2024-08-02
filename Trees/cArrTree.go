package Trees

import (
	Go_Utils "github.com/g-m-twostay/go-utils"
	"golang.org/x/exp/constraints"
	"math/bits"
	"reflect"
	"unsafe"
)

type CTree[T any, S constraints.Unsigned] struct {
	base[T, S]
	//returns negative number if first < second, 0 if first==second, positive number if first>second. see cmp.Compare for an example.
	Cmp func(T, T) int
}

func NewC[T any, S constraints.Unsigned](hint S, cmp func(T, T) int) *CTree[T, S] {
	ifs := make([]info[S], 1, hint+1)
	vs := make([]T, 0, hint)
	return &CTree[T, S]{base[T, S]{ifsHead: unsafe.Pointer(unsafe.SliceData(ifs)), ifsLen: S(len(ifs)), vsHead: unsafe.Pointer(unsafe.SliceData(vs)), caps: [2]int{cap(ifs), cap(vs)}}, cmp}
}

func FromC[T any, S constraints.Unsigned](vs []T, cmp func(T, T) int) *CTree[T, S] {
	root, ifs := buildIfs(S(len(vs)), make([][3]S, 0, bits.Len(uint(len(vs)))))
	return &CTree[T, S]{base[T, S]{root: root, ifsHead: unsafe.Pointer(unsafe.SliceData(ifs)), ifsLen: S(len(ifs)), vsHead: unsafe.Pointer(unsafe.SliceData(vs)), caps: [2]int{cap(ifs), cap(vs)}}, cmp}
}

func (u *CTree[T, S]) Add(v T, st []uintptr) (bool, []uintptr) {
	for curI := u.root; curI != 0; {
		if order := u.Cmp(v, *u.getV(curI - 1)); order < 0 {
			l := &u.getIf(curI).l
			st = append(st, uintptr(unsafe.Pointer(l))-uintptr(u.ifsHead))
			curI = *l
		} else if order > 0 {
			r := &u.getIf(curI).r
			st = append(st, uintptr(unsafe.Pointer(r))-uintptr(u.ifsHead))
			curI = *r
		} else {
			return false, st
		}
	}
	if u.root = u.popFree(); u.root == 0 {
		u.root = u.ifsLen
		a := append(*(*[]info[S])(unsafe.Pointer(&reflect.SliceHeader{uintptr(u.ifsHead), int(u.root), u.caps[0]})), info[S]{0, 0, 1})
		u.ifsHead, u.ifsLen, u.caps[0] = unsafe.Pointer(unsafe.SliceData(a)), S(len(a)), cap(a)
		b := append(*(*[]T)(unsafe.Pointer(&reflect.SliceHeader{uintptr(u.vsHead), int(u.root - 1), u.caps[1]})), v)
		u.vsHead, u.caps[1] = unsafe.Pointer(unsafe.SliceData(b)), cap(b)
	} else {
		*u.getIf(u.root) = info[S]{0, 0, 1}
		*u.getV(u.root - 1) = v
	}

	for i := len(st) - 1; i > -1; i-- {
		*(*S)(unsafe.Add(u.ifsHead, st[i])) = u.root
		index := S(st[i] / unsafe.Sizeof(info[S]{}))
		u.getIf(index).sz++
		if u.Cmp(v, *u.getV(index - 1)) >= 0 {
			u.maintainRight(&index)
		} else {
			u.maintainLeft(&index)
		}
		u.root = index
	}
	return true, st
}

func (u *CTree[T, S]) Del(v T, st []uintptr) (bool, []uintptr) {
	for curI := &u.root; *curI != 0; {
		if order := u.Cmp(v, *u.getV(*curI - 1)); order < 0 {
			st = append(st, uintptr(unsafe.Pointer(curI)))
			curI = &u.getIf(*curI).l
		} else if order > 0 {
			st = append(st, uintptr(unsafe.Pointer(curI)))
			curI = &u.getIf(*curI).r
		} else {
			if cur := u.getIf(*curI); cur.l == 0 {
				u.addFree(*curI)
				*curI = cur.r
			} else if cur.r == 0 {
				a := *curI
				*curI = cur.l
				u.addFree(a)
			} else {
				si := &cur.r
				for cur.sz--; u.getIf(*si).l != 0; si = &u.getIf(*si).l {
					u.getIf(*si).sz--
				}
				*u.getV(*curI - 1) = *u.getV(*si - 1)
				u.addFree(*si)
				*si = u.getIf(*si).r
			}
			for _, a := range st {
				u.getIf(*(*S)(unsafe.Pointer(a))).sz--
			}
			if Go_Utils.CheapRandN(uint32((u.getIf(u.root).sz+1)>>1)) == 2 {
				for i := len(st) - 1; i > -1; i-- {
					if u.Cmp(v, *u.getV(*(*S)(unsafe.Pointer(st[i])) - 1)) <= 0 {
						u.maintainRight((*S)(unsafe.Pointer(st[i])))
					} else {
						u.maintainLeft((*S)(unsafe.Pointer(st[i])))
					}
				}
			}
			return true, st
		}
	}
	return false, st
}

func (u *CTree[T, S]) Get(v T) *T {
	for curI := u.root; curI != 0; {
		if order := u.Cmp(v, *u.getV(curI - 1)); order < 0 {
			curI = u.getIf(curI).l
		} else if order > 0 {
			curI = u.getIf(curI).r
		} else {
			return u.getV(curI - 1)
		}
	}
	return nil
}

func (u *CTree[T, S]) Predecessor(v T, strict bool) (p *T) {
	if curI := u.root; strict {
		for curI != 0 {
			if u.Cmp(v, *u.getV(curI - 1)) <= 0 {
				curI = u.getIf(curI).l
			} else {
				p = u.getV(curI - 1)
				curI = u.getIf(curI).r
			}
		}
	} else {
		for curI != 0 {
			if u.Cmp(v, *u.getV(curI - 1)) < 0 {
				curI = u.getIf(curI).l
			} else {
				p = u.getV(curI - 1)
				curI = u.getIf(curI).r
			}
		}
	}
	return
}

func (u *CTree[T, S]) Successor(v T, strict bool) (p *T) {
	if curI := u.root; strict {
		for curI != 0 {
			if u.Cmp(v, *u.getV(curI - 1)) < 0 {
				p = u.getV(curI - 1)
				curI = u.getIf(curI).l
			} else {
				curI = u.getIf(curI).r
			}
		}
	} else {
		for curI != 0 {
			if u.Cmp(v, *u.getV(curI - 1)) > 0 {
				curI = u.getIf(curI).r
			} else {
				p = u.getV(curI - 1)
				curI = u.getIf(curI).l
			}
		}
	}
	return
}

func (u *CTree[T, S]) RankOf(v T) (S, bool) {
	var ra S = 0
	for curI := u.root; curI != 0; {
		if order, cur := u.Cmp(v, *u.getV(curI - 1)), *u.getIf(curI); order < 0 {
			curI = cur.l
		} else if order > 0 {
			ra += u.getIf(cur.l).sz + 1
			curI = cur.r
		} else {
			return ra + u.getIf(cur.l).sz, true
		}
	}
	return ra, false
}

// Zero all the removed elements(holes) in the value array.
func (u *CTree[T, S]) Zero() (count S) {
	for curI := u.free; curI != 0; curI = u.getIf(curI).l {
		*u.getV(curI - 1) = *new(T)
		count++
	}
	return count
}
