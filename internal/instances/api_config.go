// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package instances

import "gopkg.in/yaml.v3"

type APIConfig struct {
	RPCEndpoints     []string `yaml:"rpc.endpoints,flow,omitempty" json:"rpc.endpoints"`
	RPCDisableUpdate bool     `yaml:"rpc.disableUpdate,omitempty" json:"rpc.disableUpdate"`
	NodeId           string   `yaml:"nodeId" json:"nodeId"`
	Secret           string   `yaml:"secret" json:"secret"`
}

func (this *APIConfig) AsYAML() ([]byte, error) {
	return yaml.Marshal(this)
}
