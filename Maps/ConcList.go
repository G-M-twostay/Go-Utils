package Maps

import "sync/atomic"

type head[K Hashable, V any] struct {
	nx atomic.Pointer[node[K, V]]
}

type node[K Hashable, V any] struct {
	head[K, V]
	k   K
	v   V
	del atomic.Bool
}

// given *a, a->nx=nil
// result a->next=n; n->next=nil
func (u *head[K, V]) addAtEnd(n *node[K, V]) {
	added := false
	for !added {
		oldCur := u.next()
		added = u.nx.CompareAndSwap(oldCur, n)
	}
}

func (u *head[K, V]) next() *node[K, V] {
	for {
		oldCur := u.nx.Load()
		oldNext := oldCur.nx.Load()
		if oldCur.del.Load() {
			u.nx.CompareAndSwap(oldCur, oldNext)
		} else {
			return oldNext
		}
	}
}

func (u *node[K, V]) delete() {
	u.del.Store(false)
}
