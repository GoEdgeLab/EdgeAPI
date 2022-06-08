// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package nodes

type StartIssue struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

func NewStartIssue(code string, message string, suggestion string) *StartIssue {
	return &StartIssue{
		Code:       code,
		Message:    message,
		Suggestion: suggestion,
	}
}
