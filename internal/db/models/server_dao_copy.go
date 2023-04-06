// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/dbs"
)

// 服务基本信息
type clusterServerList struct {
	clusterId int64
	serverIds []int64
}

// CopyServerConfigToServers 拷贝服务配置到一组服务
func (this *ServerDAO) CopyServerConfigToServers(tx *dbs.Tx, fromServerId int64, toServerIds []int64, configCode serverconfigs.ConfigCode) error {
	if fromServerId <= 0 {
		return nil
	}
	if len(toServerIds) == 0 {
		return nil
	}

	webId, err := SharedServerDAO.FindServerWebId(tx, fromServerId)
	if err != nil {
		return err
	}

	clusterServers, toWebIds, err := this.findServerClusterIdsAndWebIds(tx, toServerIds)
	if err != nil {
		return err
	}
	if len(clusterServers) == 0 {
		return nil
	}

	switch configCode {
	case serverconfigs.ConfigCodeStat: // 统计
		if webId <= 0 {
			return nil
		}

		err = SharedHTTPWebDAO.CopyWebStats(tx, webId, toWebIds)
		if err != nil {
			return err
		}
	}

	// 通知更新
	for _, serverList := range clusterServers {
		err = SharedUpdatingServerListDAO.CreateList(tx, serverList.clusterId, serverList.serverIds)
		if err != nil {
			return err
		}

		err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, serverList.clusterId, 0, 0,  NodeTaskTypeUpdatingServers)
		if err != nil {
			return err
		}
	}

	return nil
}

// 查找一组服务的集群和WebId信息
func (this *ServerDAO) findServerClusterIdsAndWebIds(tx *dbs.Tx, serverIds []int64) (clusterServers []*clusterServerList, webIds []int64, err error) {
	if len(serverIds) == 0 {
		return
	}

	ones, err := this.Query(tx).
		Result("id", "webId", "clusterId").
		Pk(serverIds).
		Reuse(false).
		FindAll()
	if err != nil {
		return nil, nil, err
	}

	var clusterMap = map[int64]*clusterServerList{} // clusterId => servers

	for _, one := range ones {
		var server = one.(*Server)
		var clusterId = int64(server.ClusterId)
		if clusterId <= 0 {
			continue
		}

		serverList, ok := clusterMap[clusterId]
		if ok {
			serverList.serverIds = append(serverList.serverIds, int64(server.Id))
		} else {
			clusterMap[clusterId] = &clusterServerList{
				clusterId: clusterId,
				serverIds: []int64{int64(server.Id)},
			}
		}

		var webId = int64(server.WebId)
		if webId > 0 {
			webIds = append(webIds, webId)
		}
	}

	for _, serverList := range clusterMap {
		clusterServers = append(clusterServers, serverList)
	}

	return
}

// CopyServerConfigToGroups 拷贝服务配置到分组
func (this *ServerDAO) CopyServerConfigToGroups(tx *dbs.Tx, fromServerId int64, groupIds []int64, configCode string) error {
	if len(groupIds) == 0 {
		return nil
	}

	var serverIds = []int64{}
	for _, groupId := range groupIds {
		ones, err := this.Query(tx).
			ResultPk().
			State(ServerStateEnabled).
			Where("JSON_CONTAINS(groupIds, :groupId)").
			Param("groupId", groupId).
			FindAll()
		if err != nil {
			return err
		}
		for _, one := range ones {
			serverIds = append(serverIds, int64(one.(*Server).Id))
		}
	}

	return this.CopyServerConfigToServers(tx, fromServerId, serverIds, configCode)
}

// CopyServerConfigToCluster 拷贝服务配置到集群
func (this *ServerDAO) CopyServerConfigToCluster(tx *dbs.Tx, fromServerId int64, clusterId int64, configCode string) error {
	ones, err := this.Query(tx).
		ResultPk().
		State(ServerStateEnabled).
		Attr("clusterId", clusterId).
		UseIndex("clusterId").
		FindAll()
	if err != nil {
		return err
	}
	var serverIds = []int64{}
	for _, one := range ones {
		serverIds = append(serverIds, int64(one.(*Server).Id))
	}
	return this.CopyServerConfigToServers(tx, fromServerId, serverIds, configCode)
}

// CopyServerConfigToUser 拷贝服务配置到用户
func (this *ServerDAO) CopyServerConfigToUser(tx *dbs.Tx, fromServerId int64, userId int64, configCode string) error {
	ones, err := this.Query(tx).
		ResultPk().
		State(ServerStateEnabled).
		Attr("userId", userId).
		UseIndex("userId").
		FindAll()
	if err != nil {
		return err
	}
	var serverIds = []int64{}
	for _, one := range ones {
		serverIds = append(serverIds, int64(one.(*Server).Id))
	}
	return this.CopyServerConfigToServers(tx, fromServerId, serverIds, configCode)
}
