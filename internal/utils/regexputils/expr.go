// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package regexputils

import "regexp"

var (
	YYYYMMDDHH = regexp.MustCompile(`^\d{10}$`)
	YYYYMMDD   = regexp.MustCompile(`^\d{8}$`)
	YYYYMM     = regexp.MustCompile(`^\d{6}$`)
)
