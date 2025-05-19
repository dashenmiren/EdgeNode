package kvstore

import (
	"errors"

	"github.com/cockroachdb/pebble"
)

var ErrTableNotFound = errors.New("table not found")
var ErrKeyTooLong = errors.New("too long key")

func IsKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, pebble.ErrNotFound)
}
