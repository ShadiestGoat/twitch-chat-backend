package db

func getZero[T any]() T {
    var result T
    return result
}

type Ring[T any] []T

func NewRing[T any](size int) Ring[T] {
	return make(Ring[T], size)
}

// n must always be >= 0
func (r Ring[T]) Shift(n int) {
	if n <= 0 {
		panic("n < 0")
	}

	for i := len(r) - n; i >= n; i-- {
		r[i] = r[i - n]
	}
	
	for i := 0; i < n; i++ {
		r[i] = getZero[T]()
	}
}

func (r Ring[T]) Add(v T) {
	r.Shift(1)
	r[0] = v
}


func (r Ring[T]) Delete(deleteIndex int) {
	if deleteIndex < 0 {
		panic("deleteIndex < 0")
	}
	for i := deleteIndex + 1; i < len(r); i++ {
		r[i-1] = r[i]
	}
}
