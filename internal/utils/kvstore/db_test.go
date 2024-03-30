package kvstore_test

import (
	"testing"

	"github.com/TeaOSLab/EdgeNode/internal/utils/kvstore"
	"github.com/cockroachdb/pebble"
)

func TestNewDB(t *testing.T) {
	store, err := kvstore.OpenStore("test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = store.Close()
	}()

	_, err = store.NewDB("TEST_DB")
	if err != nil {
		t.Fatal(err)
	}

	testingStore = store
	testInspectDB(t)
}

func testInspectDB(t *testing.T) {
	if testingStore == nil {
		return
	}
	it, err := testingStore.RawDB().NewIter(&pebble.IterOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = it.Close()
	}()

	for it.First(); it.Valid(); it.Next() {
		valueBytes, valueErr := it.ValueAndErr()
		if valueErr != nil {
			t.Fatal(valueErr)
		}
		t.Log(string(it.Key()), "=>", string(valueBytes))
	}
}
