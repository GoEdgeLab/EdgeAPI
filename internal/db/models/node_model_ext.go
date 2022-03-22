package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"sort"
	"time"
)

// DecodeInstallStatus 安装状态
func (this *Node) DecodeInstallStatus() (*NodeInstallStatus, error) {
	if len(this.InstallStatus) == 0 {
		return NewNodeInstallStatus(), nil
	}
	status := &NodeInstallStatus{}
	err := json.Unmarshal(this.InstallStatus, status)
	if err != nil {
		return NewNodeInstallStatus(), err
	}

	// 如果N秒钟没有更新状态，则认为不在运行
	if status.IsRunning && status.UpdatedAt < time.Now().Unix()-10 {
		status.IsRunning = false
		status.IsFinished = true
		status.Error = "timeout"
	}

	return status, nil
}

// DecodeStatus 节点状态
func (this *Node) DecodeStatus() (*nodeconfigs.NodeStatus, error) {
	if len(this.Status) == 0 {
		return nil, nil
	}
	status := &nodeconfigs.NodeStatus{}
	err := json.Unmarshal(this.Status, status)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// DNSRouteCodes 所有的DNS线路
func (this *Node) DNSRouteCodes() map[int64][]string {
	routes := map[int64][]string{} // domainId => routes
	if len(this.DnsRoutes) == 0 {
		return routes
	}
	err := json.Unmarshal(this.DnsRoutes, &routes)
	if err != nil {
		// 忽略错误
		return routes
	}
	return routes
}

// DNSRouteCodesForDomainId DNS线路
func (this *Node) DNSRouteCodesForDomainId(dnsDomainId int64) ([]string, error) {
	routes := map[int64][]string{} // domainId => routes
	if len(this.DnsRoutes) == 0 {
		return nil, nil
	}
	err := json.Unmarshal(this.DnsRoutes, &routes)
	if err != nil {
		return nil, err
	}
	domainRoutes, _ := routes[dnsDomainId]

	if len(domainRoutes) > 0 {
		sort.Strings(domainRoutes)
	}

	return domainRoutes, nil
}

// DecodeConnectedAPINodeIds 连接的API
func (this *Node) DecodeConnectedAPINodeIds() ([]int64, error) {
	apiNodeIds := []int64{}
	if IsNotNull(this.ConnectedAPINodes) {
		err := json.Unmarshal(this.ConnectedAPINodes, &apiNodeIds)
		if err != nil {
			return nil, err
		}
	}
	return apiNodeIds, nil
}

// DecodeSecondaryClusterIds 从集群IDs
func (this *Node) DecodeSecondaryClusterIds() []int64 {
	if len(this.SecondaryClusterIds) == 0 {
		return []int64{}
	}
	var result = []int64{}
	// 不需要处理错误
	_ = json.Unmarshal(this.SecondaryClusterIds, &result)
	return result
}
