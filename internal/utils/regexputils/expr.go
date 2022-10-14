// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package regexputils

import "regexp"

var (
	YYYYMMDD = regexp.MustCompile(`^\d{8}$`)
	YYYYMM   = regexp.MustCompile(`^\d{6}$`)
)
