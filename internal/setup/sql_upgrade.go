package setup

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/acme"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
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
	"regexp"
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

// v0.2.8.1
func upgradeV0_2_8_1(db *dbs.DB) error {
	// 访问日志设置
	{
		one, err := db.FindOne("SELECT id FROM edgeSysSettings WHERE code=? LIMIT 1", systemconfigs.SettingCodeNSAccessLogSetting)
		if err != nil {
			return err
		}
		if len(one) == 0 {
			ref := &dnsconfigs.NSAccessLogRef{
				IsPrior:           false,
				IsOn:              true,
				LogMissingDomains: false,
			}
			refJSON, err := json.Marshal(ref)
			if err != nil {
				return err
			}
			_, err = db.Exec("INSERT edgeSysSettings (code, value) VALUES (?, ?)", systemconfigs.SettingCodeNSAccessLogSetting, refJSON)
			if err != nil {
				return err
			}
		}
	}

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

	return nil
}
