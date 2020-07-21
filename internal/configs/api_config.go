package configs

import (
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/Tea"
	"io/ioutil"
)

var sharedAPIConfig *APIConfig = nil

type APIConfig struct {
	RPC struct {
		Listen string `yaml:"listen"`
	} `yaml:"rpc"`
}

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
