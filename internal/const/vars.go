// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package teaconst

import (
	"crypto/sha1"
	"fmt"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"time"
)

var (
	IsPlus             = false
	MaxNodes     int32 = 0
	NodeId       int64 = 0
	Debug              = false
	InstanceCode       = fmt.Sprintf("%x", sha1.Sum([]byte("INSTANCE"+types.String(time.Now().UnixNano())+"@"+types.String(rands.Int64()))))
)
