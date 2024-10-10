package Maps

import (
	"reflect"
	"unsafe"
)

type chunkArr struct {
	logChunkSize byte
	// followed by 1<<logChunkSize-1 num of uintptrs.
	first uintptr
}

var chunkArrType unsafe.Pointer

func init() {
	t := reflect.TypeOf(chunkArr{})
	chunkArrType = (*struct {
		_     uintptr
		Value unsafe.Pointer
	})(unsafe.Pointer(&t)).Value
}

//go:linkname malloc runtime.mallocgc
func malloc(size uintptr, typ unsafe.Pointer, zero bool) unsafe.Pointer
func newChunkArr(maxLogChunkSize, logChunkSize byte) *chunkArr {
	temp := chunkArr{}
	mem := (*chunkArr)(malloc(unsafe.Offsetof(temp.first)+(1<<(maxLogChunkSize-logChunkSize))*unsafe.Sizeof(temp.first), chunkArrType, false))
	mem.logChunkSize = logChunkSize
	return mem
}

func (ca *chunkArr) Fetch(i uint) *relay {
	return *(**relay)(unsafe.Add(unsafe.Pointer(&ca.first), unsafe.Sizeof(ca.first)*uintptr(i)))
}
func (ca *chunkArr) Index(hash uint) uint {
	return hash >> ca.logChunkSize
}
func (ca *chunkArr) Get(hash uint) *relay {
	return ca.Fetch(ca.Index(hash))
}
func (ca *chunkArr) set(i uint, v *relay) {
	*(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&ca.first)) + unsafe.Sizeof(ca.first)*uintptr(i))) = uintptr(unsafe.Pointer(v))
	//*(**relay)(unsafe.Add(unsafe.Pointer(&ca.first), unsafe.Sizeof(ca.first)*uintptr(i))) = v// this is incorrect because we chunkArr is uninitialized, so casting to pointer will make it point to junk temporarily.
}
