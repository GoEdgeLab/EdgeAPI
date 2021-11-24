// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils

import "strings"

// SplitStrings 分隔字符串
// 忽略其中为空的片段
func SplitStrings(s string, glue string) []string {
	var result = []string{}

	if len(s) > 0 {
		for _, p := range strings.Split(s, glue) {
			p = strings.TrimSpace(p)
			if len(p) > 0 {
				result = append(result, p)
			}
		}
	}
	return result
}
