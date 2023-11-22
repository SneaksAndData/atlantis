package store

type StateStore[T any] interface {
	Insert(key string, value *T) error
	Remove(key string) error
	Get(key string) (*T, error)
	Exists(key string) (bool, error)
	Size() int
}
