package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
)

// DecodeAddr 解析地址
func (this *Origin) DecodeAddr() (*serverconfigs.NetworkAddressConfig, error) {
	if len(this.Addr) == 0 {
		return nil, errors.New("addr is empty")
	}
	addr := &serverconfigs.NetworkAddressConfig{}
	err := json.Unmarshal(this.Addr, addr)
	return addr, err
}

func (this *Origin) DecodeDomains() []string {
	var result = []string{}
	if len(this.Domains) > 0 {
		err := json.Unmarshal(this.Domains, &result)
		if err != nil {
			remotelogs.Error("Origin.DecodeDomains", err.Error())
		}
	}
	return result
}
