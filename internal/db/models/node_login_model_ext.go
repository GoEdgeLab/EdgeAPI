package models

import (
	"encoding/json"
	"errors"
)

// DecodeSSHParams 解析SSH登录参数
func (this *NodeLogin) DecodeSSHParams() (*NodeLoginSSHParams, error) {
	if this.Type != NodeLoginTypeSSH {
		return nil, errors.New("invalid login type '" + this.Type + "'")
	}

	if len(this.Params) == 0 {
		return nil, errors.New("'params' should not be empty")
	}

	params := &NodeLoginSSHParams{}
	err := json.Unmarshal(this.Params, params)
	if err != nil {
		return nil, err
	}

	return params, nil
}
