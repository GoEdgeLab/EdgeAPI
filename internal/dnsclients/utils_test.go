// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsclients_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/iwind/TeaGo/assert"
	"testing"
)

func TestIsMasked(t *testing.T) {
	var a = assert.NewAssertion(t)
	a.IsFalse(dnsclients.IsMasked(""))
	a.IsFalse(dnsclients.IsMasked("abc"))
	a.IsFalse(dnsclients.IsMasked("abc*"))
	a.IsTrue(dnsclients.IsMasked("*"))
	a.IsTrue(dnsclients.IsMasked("**"))
	a.IsTrue(dnsclients.IsMasked("***"))
	a.IsTrue(dnsclients.IsMasked("*******"))
	a.IsTrue(dnsclients.IsMasked("abc**"))
	a.IsTrue(dnsclients.IsMasked("abcd*********"))
}

func TestUnmaskAPIParams(t *testing.T) {
	data, err := dnsclients.UnmaskAPIParams([]byte(`{
	"key": "a",
	"secret": "abc12"
}`), []byte(`{
	"secret": "abc**"
}`))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}
