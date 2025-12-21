package ioutil

type Closer[T any] struct {
	Value     T
	CloseFunc CloseFunc
}

func (c *Closer[T]) Close() error {
	if c.CloseFunc != nil {
		return c.CloseFunc()
	}
	return nil
}

type CloseFunc func() error

func NewCloser[T any](value T, closeFunc CloseFunc) *Closer[T] {
	return &Closer[T]{
		Value:     value,
		CloseFunc: closeFunc,
	}
}

func NewNoopCloser[T any](value T) *Closer[T] {
	return &Closer[T]{
		Value:     value,
		CloseFunc: func() error { return nil },
	}
}
