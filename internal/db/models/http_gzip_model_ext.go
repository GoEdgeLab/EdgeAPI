package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
)

// DecodeMinLength 解析最小长度
func (this *HTTPGzip) DecodeMinLength() (*shared.SizeCapacity, error) {
	if len(this.MinLength) == 0 {
		return nil, nil
	}
	capacity := &shared.SizeCapacity{}
	err := json.Unmarshal(this.MinLength, capacity)
	return capacity, err
}

// DecodeMaxLength 解析最大长度
func (this *HTTPGzip) DecodeMaxLength() (*shared.SizeCapacity, error) {
	if len(this.MaxLength) == 0 {
		return nil, nil
	}
	capacity := &shared.SizeCapacity{}
	err := json.Unmarshal(this.MaxLength, capacity)
	return capacity, err
}
