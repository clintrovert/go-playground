package collections

type Streamable interface{}

type CollectionStream[T any] struct {
	collection []T
}

func Stream[T any](collection []T) *CollectionStream[T] {
	return &CollectionStream[T]{
		collection: collection,
	}
}

func (stream *CollectionStream[T]) Map(func(E) T) CollectionStream[T] {

}
