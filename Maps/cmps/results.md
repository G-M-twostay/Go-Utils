The benchmarks are based on https://github.com/golang/go/blob/master/src/sync/map_bench_test.go and the cases described at https://pkg.go.dev/sync#Map . I tried to be as fair as possible while utilizing each implementation to the most.
Below are the benchmark results ran on my computer.
```console
BenchmarkValUintptr_Load_Balanced
BenchmarkValUintptr_Load_Balanced-16                 	84103110	        13.97 ns/op	       0 B/op	       0 allocs/op
BenchmarkValUintptr_LoadAndDelete_Balanced
BenchmarkValUintptr_LoadAndDelete_Balanced-16        	86523901	        13.46 ns/op	       0 B/op	       0 allocs/op
BenchmarkValUintptr_LoadAndDelete_Adversarial
BenchmarkValUintptr_LoadAndDelete_Adversarial-16     	88566768	        13.47 ns/op	       0 B/op	       0 allocs/op
BenchmarkValUintptr_LoadOrStore_Balanced
BenchmarkValUintptr_LoadOrStore_Balanced-16          	86322240	        13.78 ns/op	       0 B/op	       0 allocs/op
BenchmarkValUintptr_LoadOrStorePtr_Adversarial
BenchmarkValUintptr_LoadOrStorePtr_Adversarial-16    	87505651	        13.61 ns/op	       0 B/op	       0 allocs/op
BenchmarkValUintptr_Case1
BenchmarkValUintptr_Case1-16                         	21700838	        72.44 ns/op	       9 B/op	       0 allocs/op
BenchmarkValUintptr_Case2
BenchmarkValUintptr_Case2-16                         	27355290	        95.48 ns/op	      20 B/op	       0 allocs/op
BenchmarkNaiveMap_Load_Balanced
BenchmarkNaiveMap_Load_Balanced-16                   	56696305	        21.68 ns/op	       0 B/op	       0 allocs/op
BenchmarkNaiveMap_Delete_Balanced
BenchmarkNaiveMap_Delete_Balanced-16                 	30169173	        50.45 ns/op	       0 B/op	       0 allocs/op
BenchmarkNaiveMap_Delete_Adversarial
BenchmarkNaiveMap_Delete_Adversarial-16              	29961723	        41.37 ns/op	       0 B/op	       0 allocs/op
BenchmarkNaiveMap_LoadOrStore_Balanced
BenchmarkNaiveMap_LoadOrStore_Balanced-16            	21999013	        72.48 ns/op	       0 B/op	       0 allocs/op
BenchmarkNaiveMap_LoadOrStore_Adversarial
BenchmarkNaiveMap_LoadOrStore_Adversarial-16         	26537866	        45.06 ns/op	       0 B/op	       0 allocs/op
BenchmarkNaiveMap_Case1
BenchmarkNaiveMap_Case1-16                           	 6588852	       197.3 ns/op	      14 B/op	       0 allocs/op
BenchmarkNaiveMap_Case2
BenchmarkNaiveMap_Case2-16                           	15457940	       104.3 ns/op	      22 B/op	       0 allocs/op
BenchmarkSyncMap_Load_Balanced
BenchmarkSyncMap_Load_Balanced-16                    	52363788	        23.13 ns/op	       0 B/op	       0 allocs/op
BenchmarkSyncMap_Delete_Balanced
BenchmarkSyncMap_Delete_Balanced-16                  	87588043	        13.34 ns/op	       0 B/op	       0 allocs/op
BenchmarkSyncMap_Delete_Adversarial
BenchmarkSyncMap_Delete_Adversarial-16               	15780254	        74.15 ns/op	      25 B/op	       0 allocs/op
BenchmarkSyncMap_LoadOrStore_Balanced
BenchmarkSyncMap_LoadOrStore_Balanced-16             	62924742	        18.73 ns/op	      15 B/op	       1 allocs/op
BenchmarkSyncMap_LoadOrStorePtr_Adversarial
BenchmarkSyncMap_LoadOrStorePtr_Adversarial-16       	 6201582	       204.8 ns/op	      44 B/op	       0 allocs/op
BenchmarkSyncMap_Case1
BenchmarkSyncMap_Case1-16                            	 3405900	       435.7 ns/op	      65 B/op	       1 allocs/op
BenchmarkSyncMap_Case2
BenchmarkSyncMap_Case2-16                            	 4281628	       309.3 ns/op	      58 B/op	       2 allocs/op
BenchmarkXSyncMap_Load_Balanced
BenchmarkXSyncMap_Load_Balanced-16                   	89714259	        13.38 ns/op	       0 B/op	       0 allocs/op
BenchmarkXSyncMap_LoadAndDelete_Balanced
BenchmarkXSyncMap_LoadAndDelete_Balanced-16          	13168117	        92.86 ns/op	       0 B/op	       0 allocs/op
BenchmarkXSyncMap_LoadAndDelete_Adversarial
BenchmarkXSyncMap_LoadAndDelete_Adversarial-16       	87106748	        13.54 ns/op	       0 B/op	       0 allocs/op
BenchmarkXSyncMap_LoadOrStore_Balanced
BenchmarkXSyncMap_LoadOrStore_Balanced-16            	86548863	        13.55 ns/op	       0 B/op	       0 allocs/op
BenchmarkXSyncMap_LoadOrStorePtr_Adversarial
BenchmarkXSyncMap_LoadOrStorePtr_Adversarial-16      	82441929	        14.24 ns/op	       0 B/op	       0 allocs/op
BenchmarkXSyncMap_Case1
BenchmarkXSyncMap_Case1-16                           	16892842	        93.75 ns/op	      21 B/op	       0 allocs/op
BenchmarkXSyncMap_Case2
BenchmarkXSyncMap_Case2-16                           	14647471	       106.5 ns/op	      30 B/op	       0 allocs/op
```

Observations:
1. `sync.map` is just bad.
2. The time difference in the read-only benchmark is caused by hash functions. If you use the exposed runtime hash function you'll find that they actually run at same speed.
3. My implementation is especially good at dealing with deletions.
4. If you have a well-designed concurrency pattern, using the normal map with locks can actually be extremely fast and beneficial. 
5. I don't know why the benchmarks didn't record allocations properly. I'm guessing it rounded the average down.

Conclusions:
1. Looks like my implementation is the best here.
2. With good locking and synchronization logic, normal maps is the best. This might not always be possible though.