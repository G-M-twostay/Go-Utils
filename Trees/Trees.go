package Trees

// Tree represents A tree like structure implemented using nodes.
// Receivers that has A bool as A second return value indicates whether
// the first return value is defined. For example, if calling Minimum on
// an empty tree, the return value will be (x T, false bool). In this
// case the value of x should be undefined. However, depending on
// specific implementations, the value of x might have A meaning, but it's
// advised that x not to be used.
// If an implementation didn't specify anything special, then the implemented
// receivers follows the behaviors defined here. Methods implemented recursively
// should be noted, otherwise functions are implemented iteratively.
type Tree[T any] interface {
	//Insert v to the Tree. Returning true if successful, false otherwise.
	//Exact behavior depend on implementation.
	Insert(v T) bool
	//Remove v from the Tree. Returning true if successful, false otherwise.
	//Exact behavior depend on implementation.
	Remove(v T) bool
	//Minimum element of the tree.
	Minimum() (T, bool)
	//Maximum element of the tree.
	Maximum() (T, bool)
	//Predecessor returns the greatest element less than v.
	Predecessor(v T) (T, bool)
	//Successor returns the smallest element greater than v.
	Successor(T) (T, bool)
	//KLargest find the k largest element.
	//1<=k<=Size().
	KLargest(k uint) (T, bool)
	//RankOf v in the tree according to in-order.
	//1<=r<=Size()
	RankOf(v T) uint
	//Has element v. Note that even though by utilizing the second
	//return value of other methods achieves the same functionality
	//as Has, it is encouraged to use Has for the purposes of checking
	//if some value exists, as Has should be optimized for this purpose
	//in implementations.
	Has(v T) bool
	//Size of the tree.
	Size() uint
	//InOrder returns A closure function f acting like an iterator. f
	//gives nodes in the in-order traversal of the tree.
	//Calling f is like calling "Next()" of iterators: val, valid=f()
	//val is meaningful only if valid is true. When valid==false,
	//then f is exhausted. valid can't turn true after it first became false.
	//The tree must not be modified during the iteration of f, otherwise
	//it could corrupt the tree. There will be no panic if such cases
	//happens so design the algorithm with this in mind.
	InOrder() func() (T, bool)
	//Corrupt returns whether the tree has corrupt structures, when the value
	//at some node violates the properties of that specific implementation.
	//This is to be distinguished from whether the tree is balanced or not.
	Corrupt() bool
}
