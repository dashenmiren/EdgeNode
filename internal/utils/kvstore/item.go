package kvstore

type Item[T any] struct {
	Key      string
	Value    T
	FieldKey []byte
}
