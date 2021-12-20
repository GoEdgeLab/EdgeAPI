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

// ContainsStringInsensitive 检查是否包含某个字符串，并且不区分大小写
func ContainsStringInsensitive(list []string, search string) bool {
	search = strings.ToLower(search)
	for _, s := range list {
		if strings.ToLower(s) == search {
			return true
		}
	}
	return false
}
