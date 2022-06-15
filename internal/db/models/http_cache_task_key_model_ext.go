package models

import "encoding/json"

// DecodeNodes 解析已完成节点信息
func (this *HTTPCacheTaskKey) DecodeNodes() map[string]bool {
	var result = map[string]bool{}
	var nodesJSON = this.Nodes
	if IsNull(nodesJSON) {
		return result
	}

	err := json.Unmarshal(nodesJSON, &result)
	if err != nil {
		// ignore error
		return result
	}

	return result
}
