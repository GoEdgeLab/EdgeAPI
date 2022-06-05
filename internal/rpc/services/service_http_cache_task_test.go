// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"github.com/iwind/TeaGo/assert"
	"testing"
)

func TestHTTPCacheTaskService_CountHTTPCacheTasks(t *testing.T) {
	var a = assert.NewAssertion(t)

	var service = &HTTPCacheTaskService{}
	a.IsTrue(service.parseDomain("aaa") == "aaa")
	a.IsTrue(service.parseDomain("AAA") == "aaa")
	a.IsTrue(service.parseDomain("a.b-c.com") == "a.b-c.com")
	a.IsTrue(service.parseDomain("a.b-c.com/hello/world") == "a.b-c.com")
	a.IsTrue(service.parseDomain("https://a.b-c.com") == "a.b-c.com")
	a.IsTrue(service.parseDomain("http://a.b-c.com/hello/world") == "a.b-c.com")
	a.IsTrue(service.parseDomain("http://a.B-c.com/hello/world") == "a.b-c.com")
	a.IsTrue(service.parseDomain("http:/aaaa.com") == "http")
	a.IsTrue(service.parseDomain("北京") == "")
}
