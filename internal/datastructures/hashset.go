package datastructures

type HashSet[T comparable] map[T]struct{}

func NewHashSetFromArr[T comparable](arr []T) HashSet[T] {
	hashSet := NewHashSet[T]()

	for _, t := range arr {
		hashSet.AddValue(t)
	}

	return hashSet
}

func NewHashSet[T comparable]() HashSet[T] {
	hashSet := make(HashSet[T])
	return hashSet
}

func (h HashSet[T]) AddValue(val T) {
	h[val] = struct{}{}
}

func (h HashSet[T]) Contains(key T) bool {
	_, ok := h[key]
	return ok
}
