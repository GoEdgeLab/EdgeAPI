// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package utils

import (
	"crypto/sha1"
	"fmt"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"sync/atomic"
	"time"
)

const sha1RandomPrefix = "SHA1_RANDOM"

var sha1Id int64 = 0

func Sha1RandomString() string {
	var s = sha1RandomPrefix + types.String(time.Now().UnixNano()) + "@" + types.String(rands.Int64()) + "@" + types.String(atomic.AddInt64(&sha1Id, 1))
	return fmt.Sprintf("%x", sha1.Sum([]byte(s)))
}
