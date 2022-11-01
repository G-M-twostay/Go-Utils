package ChainMap

import (
	"GMUtils/Maps"
	"unsafe"
)

type hold[K Maps.Hashable, V any] interface {
	next() hold[K, V]
	nextPtr() unsafe.Pointer
	isHead() bool
}
