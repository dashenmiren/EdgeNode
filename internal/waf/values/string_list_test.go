package values_test

import (
	"testing"

	"github.com/TeaOSLab/EdgeNode/internal/waf/values"
	"github.com/iwind/TeaGo/assert"
)

func TestParseStringList(t *testing.T) {
	var a = assert.NewAssertion(t)

	{
		var list = values.ParseStringList("", false)
		a.IsFalse(list.Contains("hello"))
	}

	{
		var list = values.ParseStringList(`hello

world
hi

people`, false)
		a.IsTrue(list.Contains("hello"))
		a.IsFalse(list.Contains("hello1"))
		a.IsFalse(list.Contains("Hello"))
		a.IsTrue(list.Contains("hi"))
	}
	{
		var list = values.ParseStringList(`Hello

world
hi

people`, true)
		a.IsTrue(list.Contains("hello"))
		a.IsTrue(list.Contains("Hello"))
		a.IsTrue(list.Contains("HELLO"))
		a.IsFalse(list.Contains("How"))
	}
}
