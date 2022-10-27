package Maps

import "sync/atomic"

type head[K Hashable, V any] struct {
	nx atomic.Pointer[node[K, V]]
}

type node[K Hashable, V any] struct {
	head[K, V]
	k   K
	v   V
	del bool
}

// given b=a->nx, c
// result a->nx->nx=c
func (u *head[K, V]) addAfter(n *node[K, V]) {
	added := false
	for !added {
		oldCur := u.nx.Load()
		oldNext := oldCur.nx.Load()
		if oldCur.del {
			u.nx.CompareAndSwap(oldCur, oldNext)
			continue
		}
		n.nx.Store(oldNext)
		added = u.nx.CompareAndSwap(oldCur, n)
	}
}
