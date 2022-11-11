package setup

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/acme"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
)

type upgradeVersion struct {
	version string
	f       func(db *dbs.DB) error
}

var upgradeFuncs = []*upgradeVersion{
	{
		"0.0.3", upgradeV0_0_3,
	},
	{
		"0.0.5", upgradeV0_0_5,
	},
	{
		"0.0.6", upgradeV0_0_6,
	},
	{
		"0.0.9", upgradeV0_0_9,
	},
	{
		"0.0.10", upgradeV0_0_10,
	},
	{
		"0.2.5", upgradeV0_2_5,
	},
	{
		"0.2.8.1", upgradeV0_2_8_1,
	},
	{
		"0.3.0", upgradeV0_3_0,
	},
	{
		"0.3.1", upgradeV0_3_1,
	},
	{
		"0.3.2", upgradeV0_3_2,
	},
	{
		"0.3.3", upgradeV0_3_3,
	},
	{
		"0.3.7", upgradeV0_3_7,
	},
	{
		"0.4.0", upgradeV0_4_0,
	},
	{
		"0.4.1", upgradeV0_4_1,
	},
	{
		"0.4.5", upgradeV0_4_5,
	},
	{
		"0.4.7", upgradeV0_4_7,
	},
	{
		"0.4.8", upgradeV0_4_8,
	},
	{
		"0.4.9", upgradeV0_4_9,
	},
	{
		"0.4.11", upgradeV0_4_11,
	},
	{
		"0.5.3", upgradeV0_5_3,
	},
	{
		"0.5.6", upgradeV0_5_6,
	},
	{
		"0.5.8", upgradeV0_5_8,
	},
}

// UpgradeSQLData 升级SQL数据
func UpgradeSQLData(db *dbs.DB) error {
	version, err := db.FindCol(0, "SELECT version FROM edgeVersions")
	if err != nil {
		return err
	}
	versionString := types.String(version)
	if len(versionString) > 0 {
		for _, f := range upgradeFuncs {
			if stringutil.VersionCompare(versionString, f.version) >= 0 {
				continue
			}
			err = f.f(db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// v0.0.3
func upgradeV0_0_3(db *dbs.DB) error {
	// 获取第一个管理员
	adminIdCol, err := db.FindCol(0, "SELECT id FROM edgeAdmins ORDER BY id ASC LIMIT 1")
	if err != nil {
		return err
	}
	adminId := types.Int64(adminIdCol)
	if adminId <= 0 {
		return errors.New("'edgeAdmins' table should not be empty")
	}

	// 升级edgeDNSProviders
	_, err = db.Exec("UPDATE edgeDNSProviders SET adminId=? WHERE adminId=0 AND userId=0", adminId)
	if err != nil {
		return err
	}

	// 升级edgeDNSDomains
	_, err = db.Exec("UPDATE edgeDNSDomains SET adminId=? WHERE adminId=0 AND userId=0", adminId)
	if err != nil {
		return err
	}

	// 升级edgeSSLCerts
	_, err = db.Exec("UPDATE edgeSSLCerts SET adminId=? WHERE adminId=0 AND userId=0", adminId)
	if err != nil {
		return err
	}

	// 升级edgeNodeClusters
	_, err = db.Exec("UPDATE edgeNodeClusters SET adminId=? WHERE adminId=0 AND userId=0", adminId)
	if err != nil {
		return err
	}

	// 升级edgeNodes
	_, err = db.Exec("UPDATE edgeNodes SET adminId=? WHERE adminId=0 AND userId=0", adminId)
	if err != nil {
		return err
	}

	// 升级edgeNodeGrants
	_, err = db.Exec("UPDATE edgeNodeGrants SET adminId=? WHERE adminId=0", adminId)
	if err != nil {
		return err
	}

	return nil
}

// v0.0.5
func upgradeV0_0_5(db *dbs.DB) error {
	// 升级edgeACMETasks
	_, err := db.Exec("UPDATE edgeACMETasks SET authType=? WHERE authType IS NULL OR LENGTH(authType)=0", acme.AuthTypeDNS)
	if err != nil {
		return err
	}

	return nil
}

// v0.0.6
func upgradeV0_0_6(db *dbs.DB) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeAPITokens WHERE role='user'")
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()
	col, err := stmt.FindCol(0)
	if err != nil {
		return err
	}
	count := types.Int(col)
	if count > 0 {
		return nil
	}

	nodeId := rands.HexString(32)
	secret := rands.String(32)
	_, err = db.Exec("INSERT INTO edgeAPITokens (nodeId, secret, role) VALUES (?, ?, ?)", nodeId, secret, "user")
	if err != nil {
		return err
	}

	return nil
}

// v0.0.9
func upgradeV0_0_9(db *dbs.DB) error {
	// firewall policies
	var tx *dbs.Tx
	dbs.NotifyReady()
	policies, err := models.NewHTTPFirewallPolicyDAO().FindAllEnabledFirewallPolicies(tx)
	if err != nil {
		return err
	}
	for _, policy := range policies {
		if policy.ServerId > 0 {
			continue
		}
		policyId := int64(policy.Id)
		webIds, err := models.NewHTTPWebDAO().FindAllWebIdsWithHTTPFirewallPolicyId(tx, policyId)
		if err != nil {
			return err
		}
		serverIds := []int64{}
		for _, webId := range webIds {
			serverId, err := models.NewServerDAO().FindEnabledServerIdWithWebId(tx, webId)
			if err != nil {
				return err
			}
			if serverId > 0 && !lists.ContainsInt64(serverIds, serverId) {
				serverIds = append(serverIds, serverId)
			}
		}
		if len(serverIds) == 1 {
			err = models.NewHTTPFirewallPolicyDAO().UpdateFirewallPolicyServerId(tx, policyId, serverIds[0])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// v0.0.10
func upgradeV0_0_10(db *dbs.DB) error {
	// IP Item列表转换
	ones, _, err := db.FindOnes("SELECT * FROM edgeIPItems ORDER BY id ASC")
	if err != nil {
		return err
	}
	for _, one := range ones {
		ipFromLong := utils.IP2Long(one.GetString("ipFrom"))
		ipToLong := utils.IP2Long(one.GetString("ipTo"))
		_, err = db.Exec("UPDATE edgeIPItems SET ipFromLong=?, ipToLong=? WHERE id=?", ipFromLong, ipToLong, one.GetInt64("id"))
		if err != nil {
			return err
		}
	}

	return nil
}

// v0.2.5
func upgradeV0_2_5(db *dbs.DB) error {
	// 更新用户
	_, err := db.Exec("UPDATE edgeUsers SET day=FROM_UNIXTIME(createdAt,'%Y%m%d') WHERE day IS NULL OR LENGTH(day)=0")
	if err != nil {
		return err
	}

	// 更新防火墙规则
	ones, _, err := db.FindOnes("SELECT id, actions, action, actionOptions FROM edgeHTTPFirewallRuleSets WHERE actions IS NULL OR LENGTH(actions)=0")
	if err != nil {
		return err
	}
	for _, one := range ones {
		oneId := one.GetInt64("id")
		action := one.GetString("action")
		options := one.GetString("actionOptions")
		var optionsMap = maps.Map{}
		if len(options) > 0 {
			_ = json.Unmarshal([]byte(options), &optionsMap)
		}
		var actions = []*firewallconfigs.HTTPFirewallActionConfig{
			{
				Code:    action,
				Options: optionsMap,
			},
		}
		actionsJSON, err := json.Marshal(actions)
		if err != nil {
			return err
		}
		_, err = db.Exec("UPDATE edgeHTTPFirewallRuleSets SET actions=? WHERE id=?", string(actionsJSON), oneId)
		if err != nil {
			return err
		}
	}

	return nil
}

// v0.3.0
func upgradeV0_3_0(db *dbs.DB) error {
	// 升级健康检查
	ones, _, err := db.FindOnes("SELECT id,healthCheck FROM edgeNodeClusters WHERE state=1")
	if err != nil {
		return err
	}
	for _, one := range ones {
		var clusterId = one.GetInt64("id")
		var healthCheck = one.GetString("healthCheck")
		if len(healthCheck) == 0 {
			continue
		}
		var config = &serverconfigs.HealthCheckConfig{}
		err = json.Unmarshal([]byte(healthCheck), config)
		if err != nil {
			continue
		}
		if config.CountDown <= 1 {
			config.CountDown = 3
			configJSON, err := json.Marshal(config)
			if err != nil {
				continue
			}
			_, err = db.Exec("UPDATE edgeNodeClusters SET healthCheck=? WHERE id=?", string(configJSON), clusterId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// v0.3.1
func upgradeV0_3_1(db *dbs.DB) error {
	// 清空域名统计，已使用分表代替
	// 因为可能有权限问题，所以我们忽略错误
	_, _ = db.Exec("TRUNCATE table edgeServerDomainHourlyStats")

	// 升级APIToken
	ones, _, err := db.FindOnes("SELECT uniqueId,secret FROM edgeNodeClusters")
	if err != nil {
		return err
	}
	for _, one := range ones {
		var uniqueId = one.GetString("uniqueId")
		var secret = one.GetString("secret")
		tokenOne, err := db.FindOne("SELECT id FROM edgeAPITokens WHERE nodeId=? LIMIT 1", uniqueId)
		if err != nil {
			return err
		}
		if len(tokenOne) == 0 {
			_, err = db.Exec("INSERT INTO edgeAPITokens (nodeId, secret, role, state) VALUES (?, ?, 'cluster', 1)", uniqueId, secret)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// v0.3.2
func upgradeV0_3_2(db *dbs.DB) error {
	// gzip => compression

	type HTTPGzipRef struct {
		IsPrior bool  `yaml:"isPrior" json:"isPrior"` // 是否覆盖
		IsOn    bool  `yaml:"isOn" json:"isOn"`       // 是否开启
		GzipId  int64 `yaml:"gzipId" json:"gzipId"`   // 使用的配置ID
	}

	webOnes, _, err := db.FindOnes("SELECT id, gzip FROM edgeHTTPWebs WHERE gzip IS NOT NULL AND compression IS NULL")
	if err != nil {
		return err
	}
	for _, webOne := range webOnes {
		var gzipRef = &HTTPGzipRef{}
		err = json.Unmarshal([]byte(webOne.GetString("gzip")), gzipRef)
		if err != nil {
			continue
		}
		if gzipRef == nil || gzipRef.GzipId <= 0 {
			continue
		}
		var webId = webOne.GetInt("id")

		var compressionConfig = &serverconfigs.HTTPCompressionConfig{
			UseDefaultTypes: true,
		}
		compressionConfig.IsPrior = gzipRef.IsPrior
		compressionConfig.IsOn = gzipRef.IsOn

		gzipOne, err := db.FindOne("SELECT * FROM edgeHTTPGzips WHERE id=?", gzipRef.GzipId)
		if err != nil {
			return err
		}
		if len(gzipOne) == 0 {
			continue
		}

		level := gzipOne.GetInt("level")
		if level <= 0 {
			continue
		}
		if level > 0 && level <= 10 {
			compressionConfig.Level = types.Int8(level)
		} else if level > 10 {
			compressionConfig.Level = 10
		}

		var minLengthBytes = []byte(gzipOne.GetString("minLength"))
		if len(minLengthBytes) > 0 {
			var sizeCapacity = &shared.SizeCapacity{}
			err = json.Unmarshal(minLengthBytes, sizeCapacity)
			if err != nil {
				continue
			}
			if sizeCapacity != nil {
				compressionConfig.MinLength = sizeCapacity
			}
		}

		var maxLengthBytes = []byte(gzipOne.GetString("maxLength"))
		if len(maxLengthBytes) > 0 {
			var sizeCapacity = &shared.SizeCapacity{}
			err = json.Unmarshal(maxLengthBytes, sizeCapacity)
			if err != nil {
				continue
			}
			if sizeCapacity != nil {
				compressionConfig.MaxLength = sizeCapacity
			}
		}

		var condsBytes = []byte(gzipOne.GetString("conds"))
		if len(condsBytes) > 0 {
			var conds = &shared.HTTPRequestCondsConfig{}
			err = json.Unmarshal(condsBytes, conds)
			if err != nil {
				continue
			}
			if conds != nil {
				compressionConfig.Conds = conds
			}
		}

		configJSON, err := json.Marshal(compressionConfig)
		if err != nil {
			return err
		}
		_, err = db.Exec("UPDATE edgeHTTPWebs SET compression=? WHERE id=?", string(configJSON), webId)
		if err != nil {
			return err
		}
	}

	// 更新服务端口
	var serverDAO = models.NewServerDAO()
	ones, err := serverDAO.Query(nil).
		ResultPk().
		FindAll()
	if err != nil {
		return err
	}
	for _, one := range ones {
		var serverId = int64(one.(*models.Server).Id)
		err = serverDAO.NotifyServerPortsUpdate(nil, serverId)
		if err != nil {
			return err
		}
	}

	return nil
}

// v0.3.3
func upgradeV0_3_3(db *dbs.DB) error {
	// 升级CC请求数Code
	_, err := db.Exec("UPDATE edgeHTTPFirewallRuleSets SET code='8002' WHERE name='CC请求数' AND code='8001'")
	if err != nil {
		return err
	}

	// 清除节点
	// 删除7天以前的info日志
	err = models.NewNodeLogDAO().DeleteExpiredLogsWithLevel(nil, "info", 7)
	if err != nil {
		return err
	}

	return nil
}

// v0.3.7
func upgradeV0_3_7(db *dbs.DB) error {
	// 修改所有edgeNodeGrants中的su为0
	_, err := db.Exec("UPDATE edgeNodeGrants SET su=0 WHERE su=1")
	if err != nil {
		return err
	}

	// WAF预置分组
	_, err = db.Exec("UPDATE edgeHTTPFirewallRuleGroups SET isTemplate=1 WHERE LENGTH(code)>0")
	if err != nil {
		return err
	}

	return nil
}

// v0.4.0
func upgradeV0_4_0(db *dbs.DB) error {
	// 升级SYN Flood配置
	synFloodJSON, err := json.Marshal(firewallconfigs.DefaultSYNFloodConfig())
	if err == nil {
		_, err := db.Exec("UPDATE edgeHTTPFirewallPolicies SET synFlood=? WHERE synFlood IS NULL AND state=1", string(synFloodJSON))
		if err != nil {
			return err
		}
	}

	return nil
}

// v0.4.1
func upgradeV0_4_1(db *dbs.DB) error {
	// 升级 servers.lastUserPlanId
	_, err := db.Exec("UPDATE edgeServers SET lastUserPlanId=userPlanId WHERE userPlanId>0")
	if err != nil {
		return err
	}

	// 执行域名统计清理
	err = stats.NewServerDomainHourlyStatDAO().Clean(nil, 7)
	if err != nil {
		return err
	}

	return nil
}

// v0.4.5
func upgradeV0_4_5(db *dbs.DB) error {
	// 升级访问日志自动分表
	{
		var dao = models.NewSysSettingDAO()
		valueJSON, err := dao.ReadSetting(nil, systemconfigs.SettingCodeAccessLogQueue)
		if err != nil {
			return err
		}
		if len(valueJSON) > 0 {
			var config = &serverconfigs.AccessLogQueueConfig{}
			err = json.Unmarshal(valueJSON, config)
			if err == nil && config.RowsPerTable == 0 {
				config.EnableAutoPartial = true
				config.RowsPerTable = 500_000
				configJSON, err := json.Marshal(config)
				if err == nil {
					err = dao.UpdateSetting(nil, systemconfigs.SettingCodeAccessLogQueue, configJSON)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// 升级一个防SQL注入规则
	{
		ones, _, err := db.FindOnes(`SELECT id FROM edgeHTTPFirewallRules WHERE value=?`, "(updatexml|extractvalue|ascii|ord|char|chr|count|concat|rand|floor|substr|length|len|user|database|benchmark|analyse)\\s*\\(")
		if err != nil {
			return err
		}
		for _, one := range ones {
			var ruleId = one.GetInt64("id")
			_, err = db.Exec(`UPDATE edgeHTTPFirewallRules SET value=? WHERE id=? LIMIT 1`, `\b(updatexml|extractvalue|ascii|ord|char|chr|count|concat|rand|floor|substr|length|len|user|database|benchmark|analyse)\s*\(.*\)`, ruleId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// v0.4.7
func upgradeV0_4_7(db *dbs.DB) error {
	// 升级 edgeServers 中的 plainServerNames
	{
		ones, _, err := db.FindOnes("SELECT id,serverNames FROM edgeServers WHERE state=1")
		if err != nil {
			return err
		}
		for _, one := range ones {
			var serverId = one.GetInt64("id")
			var serverNamesJSON = one.GetBytes("serverNames")
			if len(serverNamesJSON) > 0 {
				var serverNames = []*serverconfigs.ServerNameConfig{}
				err = json.Unmarshal(serverNamesJSON, &serverNames)
				if err != nil {
					return err
				}
				plainServerNamesJSON, err := json.Marshal(serverconfigs.PlainServerNames(serverNames))
				if err != nil {
					return err
				}
				_, err = db.Exec("UPDATE edgeServers SET plainServerNames=? WHERE id=?", plainServerNamesJSON, serverId)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// v0.4.8
func upgradeV0_4_8(db *dbs.DB) error {
	// 设置edgeIPLists中的serverId
	{
		firewallPolicyOnes, _, err := db.FindOnes("SELECT inbound,serverId FROM edgeHTTPFirewallPolicies WHERE serverId>0")
		if err != nil {
			return err
		}
		for _, one := range firewallPolicyOnes {
			var inboundBytes = one.GetBytes("inbound")
			var serverId = one.GetInt64("serverId")

			var listIds = []int64{}

			if len(inboundBytes) > 0 {
				var inbound = &firewallconfigs.HTTPFirewallInboundConfig{}
				err = json.Unmarshal(inboundBytes, inbound)
				if err == nil { // we ignore errors
					if inbound.AllowListRef != nil && inbound.AllowListRef.ListId > 0 {
						listIds = append(listIds, inbound.AllowListRef.ListId)
					}
					if inbound.DenyListRef != nil && inbound.DenyListRef.ListId > 0 {
						listIds = append(listIds, inbound.DenyListRef.ListId)
					}
					if inbound.GreyListRef != nil && inbound.GreyListRef.ListId > 0 {
						listIds = append(listIds, inbound.GreyListRef.ListId)
					}
				}
			}

			if len(listIds) == 0 {
				continue
			}
			for _, listId := range listIds {
				isPublicCol, err := db.FindCol(0, "SELECT isPublic FROM edgeIPLists WHERE id=? LIMIT 1", listId)
				if err != nil {
					return err
				}
				var isPublic = types.Bool(isPublicCol)
				if !isPublic {
					_, err = db.Exec("UPDATE edgeIPLists SET serverId=? WHERE id=?", serverId, listId)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// v0.4.11
func upgradeV0_4_11(db *dbs.DB) error {
	// 升级ns端口
	{
		// TCP
		{
			var config = &serverconfigs.TCPProtocolConfig{}
			config.IsOn = true
			config.Listen = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  serverconfigs.ProtocolTCP,
					PortRange: "53",
				},
			}
			configJSON, err := json.Marshal(config)
			if err != nil {
				return err
			}
			_, err = db.Exec("UPDATE edgeNSClusters SET tcp=? WHERE tcp IS NULL", configJSON)
			if err != nil {
				return err
			}
		}

		// UDP
		{
			var config = &serverconfigs.UDPProtocolConfig{}
			config.IsOn = true
			config.Listen = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  serverconfigs.ProtocolUDP,
					PortRange: "53",
				},
			}
			configJSON, err := json.Marshal(config)
			if err != nil {
				return err
			}
			_, err = db.Exec("UPDATE edgeNSClusters SET udp=? WHERE udp IS NULL", configJSON)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
