package models

import (
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/zero"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

type SysSettingDAO dbs.DAO

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

// UpdateSetting 设置配置
func (this *SysSettingDAO) UpdateSetting(tx *dbs.Tx, codeFormat string, valueJSON []byte, codeFormatArgs ...interface{}) error {
	if len(codeFormatArgs) > 0 {
		codeFormat = fmt.Sprintf(codeFormat, codeFormatArgs...)
	}

	countRetries := 3
	var lastErr error

	defer func() {
		if lastErr == nil {
			err := this.NotifyUpdate(tx, codeFormat)
			if err != nil {
				remotelogs.Error("SysSettingDAO", "notify update failed: "+err.Error())
			}
		}
	}()

	for i := 0; i < countRetries; i++ {
		settingId, err := this.Query(tx).
			Attr("code", codeFormat).
			ResultPk().
			FindInt64Col(0)
		if err != nil {
			return err
		}

		if settingId == 0 {
			// 新建
			var op = NewSysSettingOperator()
			op.Code = codeFormat
			op.Value = valueJSON
			err = this.Save(tx, op)
			if err != nil {
				lastErr = err

				// 因为错误的原因可能是因为code冲突，所以这里我们继续执行
				continue
			}

			lastErr = nil
			return nil
		}

		// 修改
		var op = NewSysSettingOperator()
		op.Id = settingId
		op.Value = valueJSON
		err = this.Save(tx, op)
		if err != nil {
			return err
		}
		lastErr = nil
		break
	}

	return lastErr
}

// ReadSetting 读取配置
func (this *SysSettingDAO) ReadSetting(tx *dbs.Tx, code string, codeFormatArgs ...interface{}) (valueJSON []byte, err error) {
	if len(codeFormatArgs) > 0 {
		code = fmt.Sprintf(code, codeFormatArgs...)
	}
	col, err := this.Query(tx).
		Attr("code", code).
		Result("value").
		FindStringCol("")
	return []byte(col), err
}

// CompareInt64Setting 对比配置中的数字大小
func (this *SysSettingDAO) CompareInt64Setting(tx *dbs.Tx, code string, anotherValue int64) (int8, error) {
	valueJSON, err := this.ReadSetting(tx, code)
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

// ReadGlobalConfig 读取全局配置
func (this *SysSettingDAO) ReadGlobalConfig(tx *dbs.Tx) (*serverconfigs.GlobalConfig, error) {
	globalConfigData, err := this.ReadSetting(tx, systemconfigs.SettingCodeServerGlobalConfig)
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

// ReadUserServerConfig 读取用户服务配置
func (this *SysSettingDAO) ReadUserServerConfig(tx *dbs.Tx) (*userconfigs.UserServerConfig, error) {
	valueJSON, err := this.ReadSetting(tx, systemconfigs.SettingCodeUserServerConfig)
	if err != nil {
		return nil, err
	}
	if len(valueJSON) == 0 {
		return userconfigs.DefaultUserServerConfig(), nil
	}

	var config = userconfigs.DefaultUserServerConfig()
	err = json.Unmarshal(valueJSON, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// ReadUserFinanceConfig 读取用户服务配置
func (this *SysSettingDAO) ReadUserFinanceConfig(tx *dbs.Tx) (*userconfigs.UserFinanceConfig, error) {
	valueJSON, err := this.ReadSetting(tx, systemconfigs.SettingCodeUserFinanceConfig)
	if err != nil {
		return nil, err
	}
	if len(valueJSON) == 0 {
		return userconfigs.DefaultUserFinanceConfig(), nil
	}

	var config = userconfigs.DefaultUserFinanceConfig()
	err = json.Unmarshal(valueJSON, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// ReadAdminUIConfig 读取管理员界面配置
func (this *SysSettingDAO) ReadAdminUIConfig(tx *dbs.Tx, cacheMap *utils.CacheMap) (*systemconfigs.AdminUIConfig, error) {
	var cacheKey = this.Table + ":ReadAdminUIConfig"
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok && cache != nil {
			return cache.(*systemconfigs.AdminUIConfig), nil
		}
	}

	valueJSON, err := this.ReadSetting(tx, systemconfigs.SettingCodeAdminUIConfig)
	if err != nil {
		return nil, err
	}
	if len(valueJSON) > 0 {
		var config = &systemconfigs.AdminUIConfig{}
		err = json.Unmarshal(valueJSON, config)
		if err != nil {
			return nil, err
		}

		if cacheMap != nil {
			cacheMap.Put(cacheKey, config)
		}

		return config, nil
	}
	return &systemconfigs.AdminUIConfig{}, nil
}

// NotifyUpdate 通知更改
func (this *SysSettingDAO) NotifyUpdate(tx *dbs.Tx, code string) error {
	switch code {
	case systemconfigs.SettingCodeAccessLogQueue:
		accessLogQueueChanged <- zero.New()
	}
	return nil
}
