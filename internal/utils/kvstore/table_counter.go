package kvstore

type CounterTable[T int64 | uint64] struct {
	*Table[T]
}

func NewCounterTable[T int64 | uint64](name string) (*CounterTable[T], error) {
	table, err := NewTable[T](name, NewIntValueEncoder[T]())
	if err != nil {
		return nil, err
	}

	return &CounterTable[T]{
		Table: table,
	}, nil
}

func (this *CounterTable[T]) Increase(key string, delta T) (newValue T, err error) {
	err = this.Table.WriteTx(func(tx *Tx[T]) error {
		value, getErr := tx.Get(key)
		if getErr != nil {
			if !IsKeyNotFound(getErr) {
				return getErr
			}
		}

		newValue = value + delta
		return tx.Set(key, newValue)
	})
	return
}
