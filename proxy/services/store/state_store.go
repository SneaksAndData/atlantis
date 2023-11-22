package store

import "github.com/pkg/errors"

type StateStore[T any] interface {
	Insert(key string, value *T, out chan<- *int)
	Remove(key string) error
	Get(key string) (*T, error)
	Exists(key string) (bool, error)
	Size() int
}

func AtomicInsert[A any, B any](keyA string, valueA *A, storeA *StateStore[A], keyB string, valueB *B, storeB *StateStore[B]) error {
	atomicChannel := make(chan *int, 2)
	completed := 0

	go (*storeA).Insert(keyA, valueA, atomicChannel)
	go (*storeB).Insert(keyB, valueB, atomicChannel)

	for v := range atomicChannel {
		if v == nil {
			return errors.New("Failed to insert a record to one of the stores")
		}

		completed = completed + *v

		if completed == 2 {
			close(atomicChannel)
			return nil
		}
	}

	return errors.New("Failed to insert a record atomically")
}
