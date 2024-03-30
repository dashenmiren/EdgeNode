package caches_test

import (
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/caches"
	"github.com/iwind/TeaGo/assert"
)

func TestItemKVEncoder_EncodeField(t *testing.T) {
	var a = assert.NewAssertion(t)

	var encoder = caches.NewItemKVEncoder[*caches.Item]()
	{
		key, err := encoder.EncodeField(&caches.Item{
			Key: "https://example.com/index.html",
		}, "key")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("key:", string(key))
		a.IsTrue(string(key) == "https://example.com/index.html")
	}

	{
		key, err := encoder.EncodeField(&caches.Item{
			Key: "",
		}, "wildKey")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("key:", string(key))
		a.IsTrue(string(key) == "")
	}

	{
		key, err := encoder.EncodeField(&caches.Item{
			Key: "example.com/index.html",
		}, "wildKey")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("key:", string(key))
		a.IsTrue(string(key) == "example.com/index.html")
	}

	{
		key, err := encoder.EncodeField(&caches.Item{
			Key: "https://example.com/index.html",
		}, "wildKey")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("key:", string(key))
		a.IsTrue(string(key) == "https://*.com/index.html")
	}

	{
		key, err := encoder.EncodeField(&caches.Item{
			Key: "https://www.example.com/index.html",
		}, "wildKey")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("key:", string(key))
		a.IsTrue(string(key) == "https://*.example.com/index.html")
	}
}
