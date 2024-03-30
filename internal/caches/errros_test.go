package caches_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/caches"
	"github.com/iwind/TeaGo/assert"
)

func TestCanIgnoreErr(t *testing.T) {
	var a = assert.NewAssertion(t)

	a.IsTrue(caches.CanIgnoreErr(caches.ErrFileIsWriting))
	a.IsTrue(caches.CanIgnoreErr(fmt.Errorf("error: %w", caches.ErrFileIsWriting)))
	a.IsTrue(errors.Is(fmt.Errorf("error: %w", caches.ErrFileIsWriting), caches.ErrFileIsWriting))
	a.IsTrue(errors.Is(caches.ErrFileIsWriting, caches.ErrFileIsWriting))
	a.IsTrue(caches.CanIgnoreErr(caches.NewCapacityError("over capacity")))
	a.IsTrue(caches.CanIgnoreErr(fmt.Errorf("error: %w", caches.NewCapacityError("over capacity"))))
	a.IsFalse(caches.CanIgnoreErr(caches.ErrNotFound))
	a.IsFalse(caches.CanIgnoreErr(errors.New("test error")))
}
