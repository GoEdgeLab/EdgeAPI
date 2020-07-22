package tests

import (
	"github.com/iwind/TeaGo/rands"
	"testing"
)

func TestRandString(t *testing.T) {
	t.Log(rands.HexString(32))
}
