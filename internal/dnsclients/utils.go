// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsclients

import (
	"encoding/json"
	"github.com/iwind/TeaGo/maps"
	"strings"
)

// MaskString 对字符串进行掩码
func MaskString(s string) string {
	var l = len(s)
	if l == 0 {
		return ""
	}
	if l < 8 {
		return strings.Repeat("*", l)
	}
	return s[:4] + strings.Repeat("*", l-4)
}

// IsMasked 判断字符串是否被掩码
func IsMasked(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s == strings.Repeat("*", len(s)) || strings.HasSuffix(s, "**")
}

// UnmaskAPIParams 恢复API参数
func UnmaskAPIParams(oldParamsJSON []byte, newParamsJSON []byte) (resultJSON []byte, err error) {
	var oldParams maps.Map
	var newParams maps.Map

	if len(oldParamsJSON) == 0 || len(newParamsJSON) == 0 {
		return newParamsJSON, nil
	}
	err = json.Unmarshal(oldParamsJSON, &oldParams)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(newParamsJSON, &newParams)
	if err != nil {
		return nil, err
	}

	if oldParams == nil || newParams == nil {
		return newParamsJSON, nil
	}

	for k, v := range newParams {
		if v != nil {
			s, ok := v.(string)
			if ok && IsMasked(s) {
				var oldV = oldParams.GetString(k)
				if len(oldV) > 0 {
					newParams[k] = oldV
				}
			}
		}
	}

	resultJSON, err = json.Marshal(newParams)
	return
}
