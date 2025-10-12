package ioutil

type Closer[T any] struct {
	Value     T
	CloseFunc CloserFunc
}

func (c *Closer[T]) Close() error {
	if c.CloseFunc != nil {
		return c.CloseFunc()
	}
	return nil
}

type CloserFunc func() error

func NewCloser[T any](value T, closeFunc CloserFunc) *Closer[T] {
	return &Closer[T]{
		Value:     value,
		CloseFunc: closeFunc,
	}
}
