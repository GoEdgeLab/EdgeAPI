// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils

import (
	"strings"
)

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

// Similar 计算相似度
// between 0-1
func Similar(s1 string, s2 string) float32 {
	var r1s = []rune(s1)
	var r2s = []rune(s2)
	var l1 = len(r1s)
	var l2 = len(r2s)

	if l1 > l2 {
		r1s, r2s = r2s, r1s
	}

	if len(r1s) == 0 {
		return 0
	}

	var count = 0
	for _, r := range r1s {
		for index, r2 := range r2s {
			if r == r2 {
				count++
				r2s = r2s[index+1:]
				break
			}
		}
	}

	return (float32(count)/float32(l1) + float32(count)/float32(l2)) / 2
}
