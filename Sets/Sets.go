package Sets

type Set[E any] interface {
	Put(E) bool
	Has(E) bool
	Remove(E) bool
	Size() uint
	Take() E
	Range(func(E) bool)
}

type ExtendedSet[E any] interface {
	PutAll(Set[E]) uint
	RemoveAll(Set[E]) uint
	Eq(Set[E]) bool
	Union(Set[E])
	Intersect(Set[E])
	Filter(func(E) bool) ExtendedSet[E]
}
