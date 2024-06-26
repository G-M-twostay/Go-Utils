package Maps

// These types defines what receivers each map should offer.

// Map is designed for compatibility with sync.Map. All the below functions have the exact same usage/behavior as documented in sync.Map.
type Map[K any, V any] interface {
	Delete(K)
	Load(K) (V, bool)
	LoadAndDelete(K) (V, bool)
	LoadOrStore(K, V) (V, bool)
	Range(func(K, V) bool)
	Store(K, V)
	Swap(K, V) (V, bool)
	CompareAndSwap(K, V, V) bool
}

// ExtendedMap is the additional operation that my implementation support. Note that these operations aren't explicit implemented, meaning that they're merely taking advantage of the implementation. For example, values are internally stored as pointers in all implementations, so why not just provide a method to access the pointers directly?
type ExtendedMap[K any, V any] interface {
	Map[K, V]
	//HasKey is a convenient alias for `_,x:=M.Load(K)`
	HasKey(K) bool
	Size() uint
	//Take an arbitrary key value pair from the Map.
	Take() (K, V)
	//Set is equivalent to Store(K,V) on a existing key, it won't do anything on a key that's not in the Map. In the prior case, it should be designed to be faster than Store.
	Set(K, V) bool
	SetPtr(K, *V) bool
}

type PtrMap[K any, V any] interface {
	Map[K, V]
	//TakePtr is the pointer variant of Take.
	TakePtr() (K, *V)
	//LoadPtr is the pointer variant of Map.Load.
	LoadPtr(K) *V
	//LoadPtrAndDelete is the pointer variant of Map.LoadAndDelete.
	LoadPtrAndDelete(K) (*V, bool)
	//LoadPtrOrStore is the pointer variant of Map.LoadOrStore.
	LoadPtrOrStore(K, V) (*V, bool)
	//RangePtr is the pointer variant of Map.Range.
	RangePtr(func(K, *V) bool)
	SwapPtr(K, *V) *V
	CompareAndSwapPtr(K, *V, *V) bool
}
