package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"time"
)

// 安装状态
func (this *Node) DecodeInstallStatus() (*NodeInstallStatus, error) {
	if len(this.InstallStatus) == 0 || this.InstallStatus == "null" {
		return NewNodeInstallStatus(), nil
	}
	status := &NodeInstallStatus{}
	err := json.Unmarshal([]byte(this.InstallStatus), status)
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

// 节点状态
func (this *Node) DecodeStatus() (*nodeconfigs.NodeStatus, error) {
	if len(this.Status) == 0 || this.Status == "null" {
		return nil, nil
	}
	status := &nodeconfigs.NodeStatus{}
	err := json.Unmarshal([]byte(this.Status), status)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// 所有的DNS线路
func (this *Node) DNSRoutes() (map[int64]string, error) {
	routes := map[int64]string{} // domainId => route
	if len(this.DnsRoutes) == 0 || this.DnsRoutes == "null" {
		return routes, nil
	}
	err := json.Unmarshal([]byte(this.DnsRoutes), &routes)
	if err != nil {
		return map[int64]string{}, err
	}
	return routes, nil
}

// DNS线路
func (this *Node) DNSRoute(dnsDomainId int64) (string, error) {
	routes := map[int64]string{} // domainId => route
	if len(this.DnsRoutes) == 0 || this.DnsRoutes == "null" {
		return "", nil
	}
	err := json.Unmarshal([]byte(this.DnsRoutes), &routes)
	if err != nil {
		return "", err
	}
	route, _ := routes[dnsDomainId]
	return route, nil
}
