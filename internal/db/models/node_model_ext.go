package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ddosconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"sort"
	"time"
)

// DecodeInstallStatus 安装状态
func (this *Node) DecodeInstallStatus() (*NodeInstallStatus, error) {
	if len(this.InstallStatus) == 0 {
		return NewNodeInstallStatus(), nil
	}
	var status = &NodeInstallStatus{}
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
	var status = &nodeconfigs.NodeStatus{}
	err := json.Unmarshal(this.Status, status)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// DNSRouteCodes 所有的DNS线路
func (this *Node) DNSRouteCodes() map[int64][]string {
	var routes = map[int64][]string{} // domainId => routes
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
	var routes = map[int64][]string{} // domainId => routes
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
	var apiNodeIds = []int64{}
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

// AllClusterIds 获取所属集群IDs
func (this *Node) AllClusterIds() []int64 {
	var result = []int64{}

	if this.ClusterId > 0 {
		result = append(result, int64(this.ClusterId))
	}

	result = append(result, this.DecodeSecondaryClusterIds()...)

	return result
}

// DecodeDDoSProtection 解析DDoS Protection设置
func (this *Node) DecodeDDoSProtection() *ddosconfigs.ProtectionConfig {
	if IsNull(this.DdosProtection) {
		return nil
	}

	var result = &ddosconfigs.ProtectionConfig{}
	err := json.Unmarshal(this.DdosProtection, &result)
	if err != nil {
		// ignore err
	}
	return result
}

// HasDDoSProtection 检查是否有DDOS设置
func (this *Node) HasDDoSProtection() bool {
	var config = this.DecodeDDoSProtection()
	if config != nil {
		return !config.IsPriorEmpty()
	}
	return false
}

// DecodeMaxCacheDiskCapacity 解析硬盘容量
func (this *Node) DecodeMaxCacheDiskCapacity() *shared.SizeCapacity {
	if this.MaxCacheDiskCapacity.IsNull() {
		return nil
	}

	// ignore error
	capacity, _ := shared.DecodeSizeCapacityJSON(this.MaxCacheDiskCapacity)
	return capacity
}

// DecodeMaxCacheMemoryCapacity 解析内存容量
func (this *Node) DecodeMaxCacheMemoryCapacity() *shared.SizeCapacity {
	if this.MaxCacheMemoryCapacity.IsNull() {
		return nil
	}

	// ignore error
	capacity, _ := shared.DecodeSizeCapacityJSON(this.MaxCacheMemoryCapacity)
	return capacity
}

// DecodeDNSResolver 解析DNS解析主机配置
func (this *Node) DecodeDNSResolver() *nodeconfigs.DNSResolverConfig {
	if this.DnsResolver.IsNull() {
		return nil
	}

	var resolverConfig = nodeconfigs.DefaultDNSResolverConfig()
	err := json.Unmarshal(this.DnsResolver, resolverConfig)
	if err != nil {
		// ignore error
	}
	return resolverConfig
}

// DecodeLnAddrs 解析Ln地址
func (this *Node) DecodeLnAddrs() []string {
	if IsNull(this.LnAddrs) {
		return nil
	}

	var result = []string{}
	err := json.Unmarshal(this.LnAddrs, &result)
	if err != nil {
		// ignore error
	}
	return result
}

// DecodeCacheDiskSubDirs 解析缓存目录
func (this *Node) DecodeCacheDiskSubDirs() []*serverconfigs.CacheDir {
	if IsNull(this.CacheDiskSubDirs) {
		return nil
	}

	var result = []*serverconfigs.CacheDir{}
	err := json.Unmarshal(this.CacheDiskSubDirs, &result)
	if err != nil {
		remotelogs.Error("Node.DecodeCacheDiskSubDirs", err.Error())
	}
	return result
}

// DecodeAPINodeAddrs 解析API节点地址
func (this *Node) DecodeAPINodeAddrs() []*serverconfigs.NetworkAddressConfig {
	var result = []*serverconfigs.NetworkAddressConfig{}
	if IsNull(this.ApiNodeAddrs) {
		return result
	}

	err := json.Unmarshal(this.ApiNodeAddrs, &result)
	if err != nil {
		remotelogs.Error("Node.DecodeAPINodeAddrs", err.Error())
	}
	return result
}

// CheckIsOffline 检查是否已经离线
func (this *Node) CheckIsOffline() bool {
	return len(this.OfflineDay) > 0 && this.OfflineDay < timeutil.Format("Ymd")
}
