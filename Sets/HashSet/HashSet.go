package HashSet

import (
	Go_Utils "github.com/g-m-twostay/go-utils"
	"math/bits"
	"unsafe"
)

const (
	fail byte = iota
	added
	exist
)

// New HashSet of type E.
// h is the neighborhood size parameter in Hopscotch hashing, 16 is a good value.
// size is used to calculate the initial table size that should handle size elements without resizing.
func New[E comparable](h byte, size, seed uint) *HashSet[E] {
	bktLen := 1<<bits.Len(size) + uint(h)
	return &HashSet[E]{bkt: make([]bucket[E], bktLen), usedBkt: Go_Utils.NewBitArray(bktLen), h: h, hashes: make([]uint, bktLen), Seed: Go_Utils.Hasher(seed)}
}

type HashSet[E comparable] struct {
	bkt     []bucket[E]
	usedBkt Go_Utils.BitArray
	hashes  []uint
	Seed    Go_Utils.Hasher
	sz      uint
	h       byte
}

func (u *HashSet[E]) hash(e *E) uint {
	return u.Seed.HashMem(unsafe.Pointer(e), unsafe.Sizeof(*e))
}

func (u *HashSet[E]) mod(hash uint) int {
	return int(hash) & (len(u.bkt) - int(u.h) - 1)
}

func (u *HashSet[E]) expand() {
	newSize := uint((len(u.bkt)-int(u.h))<<1) + uint(u.h)
	M := HashSet[E]{bkt: make([]bucket[E], newSize), h: u.h, usedBkt: Go_Utils.NewBitArray(newSize), hashes: make([]uint, newSize), Seed: u.Seed}
	for i, e := range u.bkt {
		if u.usedBkt.Get(i) {
			if M.tryPut(&e.element, u.hashes[i]) == fail {
				M.expand()
				M.tryPut(&e.element, u.hashes[i])
			}
		}
	}

	u.bkt = M.bkt
	u.usedBkt = M.usedBkt
	u.hashes = M.hashes
}

// Size of the set.
func (u *HashSet[E]) Size() uint {
	return u.sz
}

// Remove e from the set. Returns true if the removal is successful.
func (u *HashSet[E]) Remove(e E) bool {
	if i0 := u.mod(u.hash(&e)); u.bkt[i0].hashed() {
		prev := &u.bkt[i0].dHash
		for i1 := i0 + u.bkt[i0].deltaHash(); ; i1 = i1 + u.bkt[i1].deltaLink() {
			if u.usedBkt.Get(i1) && u.bkt[i1].element == e {
				u.usedBkt.Clr(i1)
				u.sz--
				if u.bkt[i1].linked() {
					*prev = offset(u.bkt[i1].deltaLink() + i1 - i0)
				} else {
					*prev = 0
				}
				return true
			}
			if !u.bkt[i1].linked() {
				break
			}
			i0 = i1
			prev = &u.bkt[i0].dLink
		}
	}
	return false
}

// Has e in the set. Returns true if e is present in the set.
func (u *HashSet[E]) Has(e E) bool {
	if i0 := u.mod(u.hash(&e)); u.bkt[i0].hashed() {
		for i1 := i0 + u.bkt[i0].deltaHash(); ; i1 = i1 + u.bkt[i1].deltaLink() {
			if u.usedBkt.Get(i1) && u.bkt[i1].element == e {
				return true
			}
			if !u.bkt[i1].linked() {
				break
			}
		}
	}
	return false
}

func (u *HashSet[E]) fillEmpty(i_hash int, i_free int, e *E) {
	u.bkt[i_free].element = *e
	u.sz++
	if u.bkt[i_hash].hashed() {
		u.bkt[i_free].useDeltaLink(i_hash + u.bkt[i_hash].deltaHash() - i_free)
	}
	u.bkt[i_hash].useDeltaHash(i_free - i_hash)
}

func (u *HashSet[E]) tryPut(e *E, hash uint) byte {
	i_hash := u.mod(hash)
	if u.bkt[i_hash].hashed() {
		for i0 := i_hash + u.bkt[i_hash].deltaHash(); ; i0 = i0 + u.bkt[i0].deltaLink() {
			if u.bkt[i0].element == *e {
				return exist
			}
			if !u.bkt[i0].linked() {
				break
			}
		}
	}
	for i_free := i_hash; i_free < len(u.bkt); i_free++ {
		if !u.usedBkt.Get(i_free) {
			if i_free-i_hash < int(u.h) {
				u.usedBkt.Set(i_free)
				u.fillEmpty(i_hash, i_free, e)
				u.hashes[i_free] = hash
				return added
			} else {
			search:
				for i := i_free - int(u.h) + 1; i < i_free; i++ {
					if i0 := i; u.bkt[i0].hashed() {
						prev := &u.bkt[i0].dHash
						for i1 := i0 + u.bkt[i0].deltaHash(); ; i1 = i1 + u.bkt[i1].deltaLink() {
							if i_free-int(u.h) < i1 && i1 < i_free {
								*prev = offset(i_free - i0)

								u.bkt[i_free].element = u.bkt[i1].element
								u.hashes[i_free] = u.hashes[i1]
								u.usedBkt.Set(i_free)

								if u.bkt[i1].linked() {
									u.bkt[i_free].useDeltaLink(u.bkt[i1].deltaLink() + i1 - i_free)
								}

								u.bkt[i1].clrLink()

								if i1 < i_hash+int(u.h) {
									u.fillEmpty(i_hash, i1, e)
									u.hashes[i1] = hash
									return added
								} else {
									u.usedBkt.Clr(i1)
									i_free = i1
									continue search
								}
							}
							if !u.bkt[i1].linked() {
								break
							}
							i0 = i1
							prev = &u.bkt[i0].dLink
						}
					}
				}
				return fail
			}
		}
	}
	return fail
}

// Put e into the set. Returns true if successful.
func (u *HashSet[E]) Put(e E) bool {
	var t byte = 4
	for hash := u.hash(&e); ; {
		if t = u.tryPut(&e, hash); t == fail {
			u.expand()
		} else {
			break
		}
	}
	return t == added
}

// Take an arbitrary element from the set. Returns zero value if the set is empty.
// Doesn't guarantee which element it will return.
// Faster than iterating with Range.
func (u *HashSet[E]) Take() (e E) {
	if i := u.usedBkt.First(); i > -1 {
		e = u.bkt[i].element
	}
	return
}

// Range over elements in a snapshot of the set at the time of the call to Range and call f on the elements.
// Stops when f returns false.
// It uses range to iterate through the bucket array, so concurrent modification during iteration won't be visible to f.
func (u *HashSet[E]) Range(f func(E) bool) {
	for i, b := range u.bkt {
		if u.usedBkt.Get(i) {
			if !f(b.element) {
				return
			}
		}
	}
}
