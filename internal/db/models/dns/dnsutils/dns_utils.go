// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsutils

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
)

// CheckClusterDNS 检查集群的DNS问题
// 藏这么深是避免package循环引用的问题
func CheckClusterDNS(tx *dbs.Tx, cluster *models.NodeCluster, checkNodeIssues bool) (issues []*pb.DNSIssue, err error) {
	var clusterId = int64(cluster.Id)
	var domainId = int64(cluster.DnsDomainId)

	// 集群DNS设置
	var clusterDNSConfig, _ = cluster.DecodeDNSConfig()

	// 检查域名
	domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, domainId, nil)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		issues = append(issues, &pb.DNSIssue{
			Target:      cluster.Name,
			TargetId:    clusterId,
			Type:        "cluster",
			Description: "域名选择错误，需要重新选择",
			Params:      nil,
			MustFix:     true,
		})
		return
	}

	// Provider
	provider, err := dns.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, int64(domain.ProviderId))
	if err != nil {
		return nil, err
	}
	if provider == nil {
		issues = append(issues, &pb.DNSIssue{
			Target:      cluster.Name,
			TargetId:    clusterId,
			Type:        "cluster",
			Description: "域名服务商不可用，需要重新选择",
			Params:      nil,
			MustFix:     true,
		})
		return
	}
	paramsMap, err := provider.DecodeAPIParams()
	if err != nil {
		issues = append(issues, &pb.DNSIssue{
			Target:      cluster.Name,
			TargetId:    clusterId,
			Type:        "cluster",
			Description: "域名服务商参数配置错误，需要重新配置",
			Params:      nil,
			MustFix:     true,
		})
		return
	}
	var dnsProvider = dnsclients.FindProvider(provider.Type, int64(provider.Id))
	if dnsProvider == nil {
		issues = append(issues, &pb.DNSIssue{
			Target:      cluster.Name,
			TargetId:    clusterId,
			Type:        "cluster",
			Description: "目前不支持\"" + provider.Type + "\"服务商，需要重新配置",
			Params:      nil,
			MustFix:     true,
		})
		return
	}
	err = dnsProvider.Auth(paramsMap)
	if err != nil {
		return
	}
	var defaultRoute = dnsProvider.DefaultRoute()
	var hasDefaultRoute = len(defaultRoute) > 0

	// 检查二级域名
	if len(cluster.DnsName) == 0 {
		issues = append(issues, &pb.DNSIssue{
			Target:      cluster.Name,
			TargetId:    clusterId,
			Type:        "cluster",
			Description: "没有设置二级域名",
			Params:      nil,
			MustFix:     true,
		})
		return
	}

	// TODO 检查域名格式

	// TODO 检查域名是否已解析

	// 检查节点
	if checkNodeIssues {
		nodes, err := models.SharedNodeDAO.FindAllEnabledNodesDNSWithClusterId(tx, clusterId, true, clusterDNSConfig != nil && clusterDNSConfig.IncludingLnNodes, true)
		if err != nil {
			return nil, err
		}

		// TODO 检查节点数量不能为0

		for _, node := range nodes {
			var nodeId = int64(node.Id)

			routeCodes, err := node.DNSRouteCodesForDomainId(domainId)
			if err != nil {
				return nil, err
			}
			if len(routeCodes) == 0 && !hasDefaultRoute {
				issues = append(issues, &pb.DNSIssue{
					Target:      node.Name,
					TargetId:    nodeId,
					Type:        "node",
					Description: "没有选择节点所属线路",
					Params: map[string]string{
						"clusterName": cluster.Name,
						"clusterId":   numberutils.FormatInt64(clusterId),
					},
					MustFix: true,
				})
				continue
			}

			// 检查线路是否在已有线路中
			for _, routeCode := range routeCodes {
				routeOk, err := domain.ContainsRouteCode(routeCode)
				if err != nil {
					return nil, err
				}
				if !routeOk {
					issues = append(issues, &pb.DNSIssue{
						Target:      node.Name,
						TargetId:    nodeId,
						Type:        "node",
						Description: "线路已经失效，请重新选择",
						Params: map[string]string{
							"clusterName": cluster.Name,
							"clusterId":   numberutils.FormatInt64(clusterId),
						},
						MustFix: true,
					})
					continue
				}
			}

			// 检查IP地址
			ipAddr, _, err := models.SharedNodeIPAddressDAO.FindFirstNodeAccessIPAddress(tx, nodeId, true, nodeconfigs.NodeRoleNode)
			if err != nil {
				return nil, err
			}
			if len(ipAddr) == 0 {
				// 检查是否有离线
				anyIPAddr, _, err := models.SharedNodeIPAddressDAO.FindFirstNodeAccessIPAddress(tx, nodeId, false, nodeconfigs.NodeRoleNode)
				if err != nil {
					return nil, err
				}
				if len(anyIPAddr) > 0 {
					issues = append(issues, &pb.DNSIssue{
						Target:      node.Name,
						TargetId:    nodeId,
						Type:        "node",
						Description: "节点所有IP地址处于离线状态",
						Params: map[string]string{
							"clusterName": cluster.Name,
							"clusterId":   numberutils.FormatInt64(clusterId),
						},
						MustFix: true,
					})
				} else {
					issues = append(issues, &pb.DNSIssue{
						Target:      node.Name,
						TargetId:    nodeId,
						Type:        "node",
						Description: "没有设置可用的IP地址",
						Params: map[string]string{
							"clusterName": cluster.Name,
							"clusterId":   numberutils.FormatInt64(clusterId),
						},
						MustFix: true,
					})
				}
				continue
			}

			// TODO 检查是否有解析记录
		}
	}

	return
}

// FindDefaultDomainRoute 获取域名默认的线路
func FindDefaultDomainRoute(tx *dbs.Tx, domain *dns.DNSDomain) (string, error) {
	if domain == nil {
		return "", errors.New("can not find domain")
	}

	provider, err := dns.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, int64(domain.ProviderId))
	if err != nil {
		return "", err
	}
	if provider == nil {
		return "", errors.New("provider not found")
	}
	paramsMap, err := provider.DecodeAPIParams()
	if err != nil {
		return "", errors.New("decode provider params failed: " + err.Error())
	}
	var dnsProvider = dnsclients.FindProvider(provider.Type, int64(provider.Id))
	if dnsProvider == nil {
		return "", errors.New("not supported provider type '" + provider.Type + "'")
	}
	err = dnsProvider.Auth(paramsMap)
	if err != nil {
		return "", err
	}
	return dnsProvider.DefaultRoute(), nil
}
