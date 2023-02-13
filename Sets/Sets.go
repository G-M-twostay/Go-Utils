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

func putAll[E any](to, from Set[E]) {
	from.Range(func(e E) bool {
		to.Put(e)
		return true
	})
}

func removeAll[E any](to, from Set[E]) {
	from.Range(func(e E) bool {
		to.Remove(e)
		return true
	})
}

func filter[E any](src, dest Set[E], f func(E) bool) {
	src.Range(func(e E) bool {
		if f(e) {
			dest.Put(e)
		}
		return true
	})
}
