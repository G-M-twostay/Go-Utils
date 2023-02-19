Fast hash set implementation based on HopMap. For details on implementation and benchmarks see the Maps/HopMap/README.md

Usage:
```
import hashset "github.com/g-m-twostay/go-utils/Sets/HashSet"

h:=16//the neighborhood size parameter H in hopscotch hashing.
seed:=uint(0)//seed for hash function
isz:=uint(7)//the amount of elements that the initial table should be able to handle.

S:=hashset.New[int](h, isz, seed)

println(S.Put(1))//prints true if the insertion is successful, i.e. when this element is new.
println(S.Has(1))//true
println(S.Has(2))//false
println(S.Remove(1))//true, 1 is succesfully removed, i.e. 1 exists in the set
pritnln(S.Remove(2))//false

for i:=0;i<10;i++{
    println(S.Put(i))//true
}

S.Range(func(i int)bool{
    println(i)
    return true
})//prints 1-10
```


