package Go_Utils

import (
	_ "runtime"
	"unsafe"
)

//go:linkname RTHash runtime.memhash
//go:noescape
func RTHash(ptr unsafe.Pointer, seed uint, len uintptr) uint

//go:linkname RTHash64 runtime.memhash64
//go:noescape
func RTHash64(ptr unsafe.Pointer, seed uint) uint

//go:linkname RTHash32 runtime.memhash32
//go:noescape
func RTHash32(ptr unsafe.Pointer, seed uint) uint

//go:linkname RTStrHash runtime.strhash
//go:noescape
func RTStrHash(ptr unsafe.Pointer, seed uint) uint

type hold struct {
	rtype *uintptr
	ptr   unsafe.Pointer
}

// Hasher is an ailas for maphash.Seed, create it using Hasher(maphash.MakeSeed()). The receivers are thread-safe, but the memory contents aren't read in a thread-safe way, so only use it on synchronized memory.
type Hasher uint

// HashAny hashes an interface value based on memory content of v. It uses internal struct's memory layout, which is unsafe practice. Avoid using it.
func (u Hasher) HashAny(v any) uint {
	h := (*hold)(unsafe.Pointer(&v))
	return u.HashMem(h.ptr, *h.rtype)
}

// HashMem hashes the memory contents in the range [addr, addr+length) as bytes.
func (u Hasher) HashMem(addr unsafe.Pointer, size uintptr) uint {
	if size == 4 {
		return RTHash32(addr, uint(u))
	} else if size == 8 {
		return RTHash64(addr, uint(u))
	}
	return RTHash(addr, uint(u), size)
}

// HashBytes hashes the given byte slice.
func (u Hasher) HashBytes(b []byte) uint {
	return u.HashMem(unsafe.Pointer(&b[0]), uintptr(uint(len(b))))
}

// HashInt hashes v.
func (u Hasher) HashInt(v int) uint {
	if unsafe.Sizeof(v) == 4 {
		return RTHash32(unsafe.Pointer(&v), uint(u))
	}
	return RTHash64(unsafe.Pointer(&v), uint(u))
}

// HashString directly hashes a string, it's faster than HashAny(string).
func (u Hasher) HashString(v string) uint {
	return RTStrHash(unsafe.Pointer(&v), uint(u))
}
