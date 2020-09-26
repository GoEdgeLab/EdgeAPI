package models

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type SysSettingDAO dbs.DAO

type SettingCode = string

const (
	SettingCodeGlobalConfig SettingCode = "globalConfig"
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

var SharedSysSettingDAO = NewSysSettingDAO()

// 设置配置
func (this *SysSettingDAO) UpdateSetting(code string, valueJSON []byte, args ...interface{}) error {
	if len(args) > 0 {
		code = fmt.Sprintf(code, args...)
	}

	countRetries := 3
	var lastErr error
	for i := 0; i < countRetries; i++ {
		settingId, err := this.Query().
			Attr("code", code).
			ResultPk().
			FindInt64Col(0)
		if err != nil {
			return err
		}

		if settingId == 0 {
			// 新建
			op := NewSysSettingOperator()
			op.Code = code
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
func (this *SysSettingDAO) ReadSetting(code string, args ...interface{}) (valueJSON []byte, err error) {
	if len(args) > 0 {
		code = fmt.Sprintf(code, args...)
	}
	col, err := this.Query().
		Attr("code", code).
		Result("value").
		FindStringCol("")
	return []byte(col), err
}
