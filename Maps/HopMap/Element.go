package HopMap

type Element[K comparable, V any] struct {
	key            K
	val            V
	hashOS, linkOS int16 //lowest bit indicates if it's valid
}

func (e *Element[K, V]) hashed() bool {
	return e.hashOS&1 == 1
}

func (e *Element[K, V]) clrHash() {
	e.hashOS = 0
}

func (e *Element[K, V]) linked() bool {
	return e.linkOS&1 == 1
}

func (e *Element[K, V]) clrLink() {
	e.linkOS = 0
}

func (e *Element[K, V]) linkOffSet() int {
	return int(e.linkOS) >> 1
}

func (e *Element[K, V]) hashOffSet() int {
	return int(e.hashOS) >> 1
}

func (e *Element[K, V]) UseHashOffSet(d int) {
	e.hashOS = markLowestBit16(d, 1)
}

func (e *Element[K, V]) UseLinkOffSet(d int) {
	e.linkOS = markLowestBit16(d, 1)
}

func markLowestBit16(x, low int) int16 {
	return int16((x << 1) | low)
}

//func (e *Element[K, V]) String() string {
//	return fmt.Sprintf("key: %v, val: %v, ho: %v, lo: %v, info: %b", e.key, e.val, e.hashOS, e.linkOS, e.info)
//}
