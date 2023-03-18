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

// LimitString 限制字符串长度
func LimitString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= 0 {
		return ""
	}

	var runes = []rune(s)
	var rs = len(runes)
	for i := 0; i < rs; i++ {
		if len(string(runes[:i+1])) > maxLength {
			return string(runes[:i])
		}
	}
	return s
}

// SplitKeywordArgs 分隔关键词参数
// 支持：hello, "hello", name:hello, name:"hello", name:\"hello\"
func SplitKeywordArgs(s string) (args []splitArg) {
	var value []rune
	var beginQuote = false
	var runes = []rune(s)
	for index, r := range runes {
		if r == '"' && (index == 0 || runes[index-1] != '\\') {
			beginQuote = !beginQuote
			continue
		}
		if !beginQuote && (r == ' ' || r == '\t' || r == '\n' || r == '\r') {
			if len(value) > 0 {
				args = append(args, parseKeywordValue(string(value)))
				value = nil
			}
		} else {
			value = append(value, r)
		}
	}

	if len(value) > 0 {
		args = append(args, parseKeywordValue(string(value)))
	}

	return
}

type splitArg struct {
	Key   string
	Value string
}

func (this *splitArg) String() string {
	if len(this.Key) > 0 {
		return this.Key + ":" + this.Value
	}
	return this.Value
}

func parseKeywordValue(value string) (arg splitArg) {
	var colonIndex = strings.Index(value, ":")
	if colonIndex > 0 {
		arg.Key = value[:colonIndex]
		arg.Value = value[colonIndex+1:]
	} else {
		arg.Value = value
	}
	return
}
