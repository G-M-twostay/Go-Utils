package Maps

import (
	"hash/maphash"
	"math"
	"reflect"
	"unsafe"
)

const (
	MaxArrayLen uint = math.MaxInt
)

// Map is designed for compatibility with sync.Map. All the below functions have the exact same usage/behavior as documented in sync.Map.
type Map[K any, V any] interface {
	Delete(K)
	Load(K) (V, bool)
	LoadAndDelete(K) (V, bool)
	LoadOrStore(K, V) (V, bool)
	Range(func(K, V) bool)
	Store(K, V)
}

// ExtendedMap is the additional operation that my implementation support. Note that these operations aren't explicit implemented, meaning that they're merely taking advantage of the implementation. For example, values are internally stored as pointers in all implementations, so why not just provide a method to access the pointers directly?
type ExtendedMap[K any, V any] interface {
	Map[K, V]
	//HasKey is a convenient alias for `_,x:=M.Load(K)`
	HasKey(K) bool
	Size() uint
	//Take an arbitrary key value pair from the Map.
	Take() (K, V)
	//TakePtr is the pointer variant of Take.
	TakePtr(K, *V)
	//LoadPtr is the pointer variant of Map.Load.
	LoadPtr(K) *V
	//LoadPtrAndDelete is the pointer variant of Map.LoadAndDelete.
	LoadPtrAndDelete(K) (*V, bool)
	//LoadPtrOrStore is the pointer variant of Map.LoadOrStore.
	LoadPtrOrStore(K, V) (*V, bool)
	//RangePtr is the pointer variant of Map.Range.
	RangePtr(func(K, *V) bool)
	//Set is equivalent to Store(K,V) on a existing key, it won't do anything on a key that's not in the Map. In the prior case, it should be designed to be faster than Store.
	Set(K, V) *V
}

// Hasher is an ailas for maphash.Seed, create it using Hasher(maphash.MakeSeed()). The receivers are thread-safe, but the memory contents aren't read in a thread-safe way, so only use it on synchronized memory.
type Hasher maphash.Seed

// HashAny hashes v based on memory content of v. It uses internal struct's memory layout, which is unsafe practice. Avoid using it.
func (u Hasher) HashAny(v any) uint {
	h := (*hold)(unsafe.Pointer(&v))
	return u.HashMem(h.ptr, *(*int)(h.rtype))
}

// HashMem hashes the memory contents in the range [addr, addr+length) as bytes.
func (u Hasher) HashMem(addr uintptr, length int) uint {
	s := reflect.SliceHeader{addr, length, length}
	return uint(maphash.Bytes(maphash.Seed(u), *(*[]byte)(unsafe.Pointer(&s))))
}

// HashBytes hashes the given byte slice.
func (u Hasher) HashBytes(b []byte) uint {
	return uint(maphash.Bytes(maphash.Seed(u), b))
}

// HashInt hashes v.
func (u Hasher) HashInt(v int) uint {
	b := (*[unsafe.Sizeof(v)]byte)(unsafe.Pointer(&v))
	return uint(maphash.Bytes(maphash.Seed(u), b[:]))
}

// HashString directly hashes a string, it's faster than HashAny(string).
func (u Hasher) HashString(v string) uint {
	return uint(maphash.String(maphash.Seed(u), v))
}
