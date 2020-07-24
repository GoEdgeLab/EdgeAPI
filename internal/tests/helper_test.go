package tests

import (
	"github.com/iwind/TeaGo/rands"
	"math"
	"net/url"
	"testing"
)

func TestRandString(t *testing.T) {
	t.Log(rands.HexString(32))
}

func TestCharset(t *testing.T) {
	t.Log(url.QueryEscape("中文"))
}

func TestInt(t *testing.T) {
	t.Log(math.MaxInt64)
}