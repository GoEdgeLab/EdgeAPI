package configs

import (
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/Tea"
	"io/ioutil"
)

var sharedAPIConfig *APIConfig = nil

// API节点配置
type APIConfig struct {
	NodeId string `yaml:"nodeId" json:"nodeId"`
	Secret string `yaml:"secret" json:"secret"`

	numberId int64 // 数字ID
}

// 获取共享配置
func SharedAPIConfig() (*APIConfig, error) {
	sharedLocker.Lock()
	defer sharedLocker.Unlock()

	if sharedAPIConfig != nil {
		return sharedAPIConfig, nil
	}

	data, err := ioutil.ReadFile(Tea.ConfigFile("api.yaml"))
	if err != nil {
		return nil, err
	}

	config := &APIConfig{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	sharedAPIConfig = config
	return config, nil
}

// 设置数字ID
func (this *APIConfig) SetNumberId(numberId int64) {
	this.numberId = numberId
}

// 获取数字ID
func (this *APIConfig) NumberId() int64 {
	return this.numberId
}
