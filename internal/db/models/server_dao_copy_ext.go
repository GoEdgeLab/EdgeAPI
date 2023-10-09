// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import (
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/dbs"
)

// CopyServerConfigToServers 拷贝服务配置到一组服务
func (this *ServerDAO) CopyServerConfigToServers(tx *dbs.Tx, fromServerId int64, toServerIds []int64, configCode serverconfigs.ConfigCode, wafCopyRegions bool) error {
	return errors.New("not implemented")
}

// CopyServerConfigToGroups 拷贝服务配置到分组
func (this *ServerDAO) CopyServerConfigToGroups(tx *dbs.Tx, fromServerId int64, groupIds []int64, configCode string, wafCopyRegions bool) error {
	return errors.New("not implemented")
}

// CopyServerConfigToCluster 拷贝服务配置到集群
func (this *ServerDAO) CopyServerConfigToCluster(tx *dbs.Tx, fromServerId int64, clusterId int64, configCode string, wafCopyRegions bool) error {
	return errors.New("not implemented")
}

// CopyServerConfigToUser 拷贝服务配置到用户
func (this *ServerDAO) CopyServerConfigToUser(tx *dbs.Tx, fromServerId int64, userId int64, configCode string, wafCopyRegions bool) error {
	return errors.New("not implemented")
}

// CopyServerUAMConfigs 复制UAM设置
func (this *ServerDAO) CopyServerUAMConfigs(tx *dbs.Tx, fromServerId int64, toServerIds []int64, wafCopyRegions bool) error {
	return errors.New("not implemented")
}
