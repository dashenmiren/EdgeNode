package kvstore_test

import (
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/utils/kvstore"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/assert"
	_ "github.com/iwind/TeaGo/bootstrap"
)

func TestRemoveDB(t *testing.T) {
	err := kvstore.RemoveStore(Tea.Root + "/data/stores/test.store")
	if err != nil {
		t.Fatal(err)
	}
}

func TestIsValidName(t *testing.T) {
	var a = assert.NewAssertion(t)

	a.IsTrue(kvstore.IsValidName("a"))
	a.IsTrue(kvstore.IsValidName("aB"))
	a.IsTrue(kvstore.IsValidName("aBC1"))
	a.IsTrue(kvstore.IsValidName("aBC1._-"))
	a.IsFalse(kvstore.IsValidName(" aBC1._-"))
	a.IsFalse(kvstore.IsValidName(""))
}
