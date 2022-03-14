// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package setup

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"strings"
)

func ComposeSQLVersion() string {
	var version = teaconst.Version
	if len(teaconst.SQLVersion) == 0 {
		return version
	}

	if strings.Count(version, ".") <= 2 {
		return version + "." + teaconst.SQLVersion
	}
	return version
}
