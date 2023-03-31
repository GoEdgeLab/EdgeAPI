// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package teaconst

import (
	"crypto/sha1"
	"fmt"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"os"
	"strings"
	"time"
)

var (
	IsPlus             = false
	Edition            = ""
	MaxNodes     int32 = 0
	NodeId       int64 = 0
	Debug              = false
	InstanceCode       = fmt.Sprintf("%x", sha1.Sum([]byte("INSTANCE"+types.String(time.Now().UnixNano())+"@"+types.String(rands.Int64()))))
	IsMain             = checkMain()
)

// 检查是否为主程序
func checkMain() bool {
	if len(os.Args) == 1 ||
		(len(os.Args) >= 2 && os.Args[1] == "pprof") {
		return true
	}
	exe, _ := os.Executable()
	return strings.HasSuffix(exe, ".test") ||
		strings.HasSuffix(exe, ".test.exe") ||
		strings.Contains(exe, "___")
}
