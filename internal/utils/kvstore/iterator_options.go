package kvstore

import "github.com/cockroachdb/pebble"

type IteratorOptions struct {
	LowerBound []byte
	UpperBound []byte
}

func (this *IteratorOptions) RawOptions() *pebble.IterOptions {
	return &pebble.IterOptions{
		LowerBound: this.LowerBound,
		UpperBound: this.UpperBound,
	}
}
