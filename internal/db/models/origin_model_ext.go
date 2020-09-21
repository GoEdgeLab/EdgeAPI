package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
)

// 解析地址
func (this *Origin) DecodeAddr() (*serverconfigs.NetworkAddressConfig, error) {
	if len(this.Addr) == 0 || this.Addr == "null" {
		return nil, errors.New("addr is empty")
	}
	addr := &serverconfigs.NetworkAddressConfig{}
	err := json.Unmarshal([]byte(this.Addr), addr)
	return addr, err
}
