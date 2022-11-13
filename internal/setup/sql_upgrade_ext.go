// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package setup

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
)

// v0.2.8.1
func upgradeV0_2_8_1(db *dbs.DB) error {
	// 升级EdgeDNS线路
	ones, _, err := db.FindOnes("SELECT id, dnsRoutes FROM edgeNodes WHERE dnsRoutes IS NOT NULL")
	if err != nil {
		return err
	}
	for _, one := range ones {
		var nodeId = one.GetInt64("id")
		var dnsRoutes = one.GetString("dnsRoutes")
		if len(dnsRoutes) == 0 {
			continue
		}
		var m = map[string][]string{}
		err = json.Unmarshal([]byte(dnsRoutes), &m)
		if err != nil {
			continue
		}
		var isChanged = false
		var reg = regexp.MustCompile(`^\d+$`)
		for k, routes := range m {
			for index, route := range routes {
				if reg.MatchString(route) {
					route = "id:" + route
					isChanged = true
				}
				routes[index] = route
			}
			m[k] = routes
		}

		if isChanged {
			mJSON, err := json.Marshal(m)
			if err != nil {
				return err
			}
			_, err = db.Exec("UPDATE edgeNodes SET dnsRoutes=? WHERE id=? LIMIT 1", string(mJSON), nodeId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// v0.4.9
func upgradeV0_4_9(db *dbs.DB) error {
	// 升级管理配置
	{
		one, err := db.FindOne("SELECT value FROM edgeSysSettings WHERE code=?", systemconfigs.SettingCodeAdminSecurityConfig)
		if err != nil {
			return err
		}
		if one != nil {
			var valueJSON = one.GetBytes("value")
			if len(valueJSON) > 0 {
				var config = &systemconfigs.SecurityConfig{}
				err = json.Unmarshal(valueJSON, config)
				if err == nil {
					config.DenySearchEngines = true
					config.DenySpiders = true
					configJSON, err := json.Marshal(config)
					if err != nil {
						return errors.New("encode SecurityConfig failed: " + err.Error())
					} else {
						_, err := db.Exec("UPDATE edgeSysSettings SET value=? WHERE code=?", configJSON, systemconfigs.SettingCodeAdminSecurityConfig)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

// v0.5.3
func upgradeV0_5_3(db *dbs.DB) error {
	// 升级集群服务配置
	{
		type oldGlobalConfig struct {
			// HTTP & HTTPS相关配置
			HTTPAll struct {
				MatchDomainStrictly  bool                                `yaml:"matchDomainStrictly" json:"matchDomainStrictly"`   // 是否严格匹配域名
				AllowMismatchDomains []string                            `yaml:"allowMismatchDomains" json:"allowMismatchDomains"` // 允许的不匹配的域名
				DefaultDomain        string                              `yaml:"defaultDomain" json:"defaultDomain"`               // 默认的域名
				DomainMismatchAction *serverconfigs.DomainMismatchAction `yaml:"domainMismatchAction" json:"domainMismatchAction"` // 不匹配时采取的动作
			} `yaml:"httpAll" json:"httpAll"`
		}

		value, err := db.FindCol(0, "SELECT value FROM edgeSysSettings WHERE code='serverGlobalConfig'")
		if err != nil {
			return err
		}
		if value != nil {
			var valueJSON = []byte(types.String(value))
			var oldConfig = &oldGlobalConfig{}
			err = json.Unmarshal(valueJSON, oldConfig)
			if err == nil {
				var newConfig = &serverconfigs.GlobalServerConfig{}
				newConfig.HTTPAll.MatchDomainStrictly = oldConfig.HTTPAll.MatchDomainStrictly
				newConfig.HTTPAll.AllowMismatchDomains = oldConfig.HTTPAll.AllowMismatchDomains
				newConfig.HTTPAll.DefaultDomain = oldConfig.HTTPAll.DefaultDomain
				if oldConfig.HTTPAll.DomainMismatchAction != nil {
					newConfig.HTTPAll.DomainMismatchAction = oldConfig.HTTPAll.DomainMismatchAction
				}
				newConfig.HTTPAll.AllowNodeIP = true

				newConfig.Log.RecordServerError = false
				newConfigJSON, err := json.Marshal(newConfig)
				if err == nil {
					_, err = db.Exec("UPDATE edgeNodeClusters SET globalServerConfig=?", newConfigJSON)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// v0.5.6
func upgradeV0_5_6(db *dbs.DB) error {
	// 修复默认集群的DNS设置
	{
		var id = 1
		clusterMap, err := db.FindOne("SELECT dns FROM edgeNodeClusters WHERE id=? AND state=1", id)
		if err != nil {
			return err
		}
		if len(clusterMap) > 0 {
			var dnsString = clusterMap.GetString("dns")
			if len(dnsString) > 0 && dnsString != "null" {
				var dnsData = []byte(dnsString)
				var dnsConfig = &dnsconfigs.ClusterDNSConfig{
					CNAMEAsDomain:    true,
					IncludingLnNodes: true,
				}
				err = json.Unmarshal(dnsData, dnsConfig)
				if err == nil && !dnsConfig.NodesAutoSync && !dnsConfig.ServersAutoSync {
					dnsConfig.NodesAutoSync = true
					dnsConfig.ServersAutoSync = true
					dnsConfigJSON, err := json.Marshal(dnsConfig)
					if err != nil {
						return err
					}
					_, err = db.Exec("UPDATE edgeNodeClusters SET dns=? WHERE id=?", dnsConfigJSON, id)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// v0.5.7
func upgradeV0_5_8(db *dbs.DB) error {
	// node task versions
	{
		_, err := db.Exec("UPDATE edgeNodeTasks SET version=0 WHERE LENGTH(version)=19")
		if err != nil {
			return err
		}
	}

	// 删除操作系统和浏览器相关统计
	// 只删除当前月，避免因为数据过多阻塞
	{
		_, err := db.Exec("DELETE FROM edgeServerClientSystemMonthlyStats WHERE month=?", timeutil.Format("Ym"))
		if err != nil {
			return err
		}
	}
	{
		_, err := db.Exec("DELETE FROM edgeServerClientBrowserMonthlyStats WHERE month=?", timeutil.Format("Ym"))
		if err != nil {
			return err
		}
	}

	// 修复默认黑白名单不是全局的问题
	{
		_, err := db.Exec("UPDATE edgeIPLists SET isGlobal=1 WHERE id IN (1, 2)")
		if err != nil {
			return err
		}
	}

	return nil
}
