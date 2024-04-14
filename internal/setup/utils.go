// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package setup

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"strings"
)

// ComposeSQLVersion 组合SQL的版本号
func ComposeSQLVersion() string {
	return teaconst.Version
}

// CompareVersion 对比版本
func CompareVersion(version1 string, version2 string) int8 {
	if len(version1) == 0 || len(version2) == 0 {
		return 0
	}

	return stringutil.VersionCompare(fixVersion(version1), fixVersion(version2))
}

func fixVersion(version string) string {
	var pieces = strings.Split(version, ".")
	var lastPiece = types.Int(pieces[len(pieces)-1])
	if lastPiece > 10 {
		// 这个是以前使用的SQL版本号，我们给去掉
		version = strings.Join(pieces[:len(pieces)-1], ".")
	}
	return version
}
