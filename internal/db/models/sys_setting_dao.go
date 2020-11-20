package models

import (
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

type SysSettingDAO dbs.DAO

type SettingCode = string

const (
	SettingCodeServerGlobalConfig  SettingCode = "serverGlobalConfig"  // 服务相关全局设置
	SettingCodeNodeMonitor         SettingCode = "nodeMonitor"         // 监控节点状态
	SettingCodeClusterHealthCheck  SettingCode = "clusterHealthCheck"  // 集群健康检查
	SettingCodeIPListVersion       SettingCode = "ipListVersion"       // IP名单的版本号
	SettingCodeAdminSecurityConfig SettingCode = "adminSecurityConfig" // 管理员安全设置
)

func NewSysSettingDAO() *SysSettingDAO {
	return dbs.NewDAO(&SysSettingDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeSysSettings",
			Model:  new(SysSetting),
			PkName: "id",
		},
	}).(*SysSettingDAO)
}

var SharedSysSettingDAO *SysSettingDAO

func init() {
	dbs.OnReady(func() {
		SharedSysSettingDAO = NewSysSettingDAO()
	})
}

// 设置配置
func (this *SysSettingDAO) UpdateSetting(codeFormat string, valueJSON []byte, codeFormatArgs ...interface{}) error {
	if len(codeFormatArgs) > 0 {
		codeFormat = fmt.Sprintf(codeFormat, codeFormatArgs...)
	}

	countRetries := 3
	var lastErr error
	for i := 0; i < countRetries; i++ {
		settingId, err := this.Query().
			Attr("code", codeFormat).
			ResultPk().
			FindInt64Col(0)
		if err != nil {
			return err
		}

		if settingId == 0 {
			// 新建
			op := NewSysSettingOperator()
			op.Code = codeFormat
			op.Value = valueJSON
			_, err = this.Save(op)
			if err != nil {
				lastErr = err

				// 因为错误的原因可能是因为code冲突，所以这里我们继续执行
				continue
			}
			return nil
		}

		// 修改
		op := NewSysSettingOperator()
		op.Id = settingId
		op.Value = valueJSON
		_, err = this.Save(op)
		if err != nil {
			return err
		}
	}

	return lastErr
}

// 读取配置
func (this *SysSettingDAO) ReadSetting(code string, codeFormatArgs ...interface{}) (valueJSON []byte, err error) {
	if len(codeFormatArgs) > 0 {
		code = fmt.Sprintf(code, codeFormatArgs...)
	}
	col, err := this.Query().
		Attr("code", code).
		Result("value").
		FindStringCol("")
	return []byte(col), err
}

// 对比配置中的数字大小
func (this *SysSettingDAO) CompareInt64Setting(code string, anotherValue int64) (int8, error) {
	valueJSON, err := this.ReadSetting(code)
	if err != nil {
		return 0, err
	}
	value := types.Int64(string(valueJSON))
	if value > anotherValue {
		return 1, nil
	}
	if value < anotherValue {
		return -1, nil
	}
	return 0, nil
}

// 读取全局配置
func (this *SysSettingDAO) ReadGlobalConfig() (*serverconfigs.GlobalConfig, error) {
	globalConfigData, err := this.ReadSetting(SettingCodeServerGlobalConfig)
	if err != nil {
		return nil, err
	}
	if len(globalConfigData) == 0 {
		return &serverconfigs.GlobalConfig{}, nil
	}
	config := &serverconfigs.GlobalConfig{}
	err = json.Unmarshal(globalConfigData, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
