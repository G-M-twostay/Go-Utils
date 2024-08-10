Here are the [benchmark](bench_test.go) results ran on my computer with `-benchmem -benchtime=15s`. The benchmark compares SBTree with
some common binary tree implementation. Most of the targets are some kind of RBTree except a degree 4 B-Tree. Other common Binary
tree implementations such as AVL are already compared in the original paper.

It's not a perfectly fair benchmark because only SBTree and B-Tree supports generics while the rest all uses `any`, which incurs a non-trivial
cost. However, gcgo compiler has optimizations for using word sized values in interfaces. Also, I tried to produce the fastest possible result
for each of the implementation (some contains optimizations for `int`). In conclusion, I don't consider the performance penalties of using `any`
to be very significant. On a personal note, I believe it's those's implementations fault to not use the new generic features in Go.

Some SBTree operations have 2 types. "0" is when everything, backing arrays and recursion stack, are all grown on demand using `append`. "1"
is when everything is allocated priorly. Because the recursion stack is at most the height of the tree, we can calculate the size 
beforehand as SBTree has a very strict height guarantee.
### Results
```console
goos: windows
goarch: amd64
pkg: github.com/g-m-twostay/go-utils/Trees     
BenchmarkBT_ReplaceOrInsert
BenchmarkBT_ReplaceOrInsert-16          	      31	 496722019 ns/op	43276068 B/op	  820500 allocs/op
BenchmarkLLRB_ReplaceOrInsertBulk
BenchmarkLLRB_ReplaceOrInsertBulk-16    	      18	 895253983 ns/op	56000006 B/op	 2000000 allocs/op
BenchmarkRBT_Put
BenchmarkRBT_Put-16                     	      22	 738142045 ns/op	72000069 B/op	 2000000 allocs/op
BenchmarkSBT_Add0
BenchmarkSBT_Add0-16                    	      32	 547291903 ns/op	111947523 B/op	      82 allocs/op
BenchmarkSBT_Add1
BenchmarkSBT_Add1-16                    	      34	 483302932 ns/op	20005074 B/op	       3 allocs/op
BenchmarkBT_Delete
BenchmarkBT_Delete-16                   	      30	 592289490 ns/op	 5877787 B/op	   74419 allocs/op
BenchmarkLLRB_Delete
BenchmarkLLRB_Delete-16                 	      16	1082013212 ns/op	       0 B/op	       0 allocs/op
BenchmarkRBT_Remove
BenchmarkRBT_Remove-16                  	      21	 773756505 ns/op	 8381124 B/op	 1000000 allocs/op
BenchmarkSBT_Del0
BenchmarkSBT_Del0-16                    	      44	 386036024 ns/op	     512 B/op	       6 allocs/op
BenchmarkSBT_Del1
BenchmarkSBT_Del1-16                    	      45	 375220932 ns/op	     208 B/op	       1 allocs/op
BenchmarkRBT_Get
BenchmarkRBT_Get-16                     	      22	 771337600 ns/op	 8363806 B/op	 1000000 allocs/op
BenchmarkSBT_Get
BenchmarkSBT_Get-16                     	      50	 360062384 ns/op	       0 B/op	       0 allocs/op
BenchmarkLLRB_Has
BenchmarkLLRB_Has-16                    	      22	 769680527 ns/op	 4000005 B/op	  500000 allocs/op
BenchmarkBT_Has
BenchmarkBT_Has-16                      	      31	 561787677 ns/op	       0 B/op	       0 allocs/op
PASS
```

### Observations
1. In terms of insertion, SBTree is on par with degree 4 B-tree and **52% faster** than the fastest RBtree implementation.
2. In terms of deletion, SBTree is **57% faster** than degree 4 B-Tree, which is 30% faster than the fastest RBTree implementation.
3. In terms of searching, SBTree is **56% faster** than degree 4 B-Tree, which is 37% faster than the fastest RBTree implementation. However, the penalties of using `any` are also the most significant here as witnessed by the positive allocations.
4. In almost all cases, SBTree makes significantly fewer allocations than all other implementations. The only exception is deletion, where SBTree allocates a small recursion stack on heap.
5. In almost all cases, SBTree allocates less memory. The major exception is Add0, where the `append` frequently copies the backing arrays.

### Insights
1. B-Tree with degree <4 is significantly slower than SBTree, but I expect a B-Tree with a reasonable higher degree will be faster than SBTree, at least when concerning `int`. However, the speed that B-Tree offers is based on the fast speed of `int` comparisons. When comparisons are slow, such as `string`, B-Tree's performance degrade linearly with its degree.
2. The 3 allocations of Add1 are the backing array containing node information (children and size), the backing array of values, and the recursion stack.
3. The 1 allocation of Del1 is the recursion stack.
4. By using 2 arrays instead of packing nodes into structs, SBTree avoids padding and is way more GC friendly.