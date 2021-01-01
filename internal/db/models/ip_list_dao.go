package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ipconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	IPListStateEnabled  = 1 // 已启用
	IPListStateDisabled = 0 // 已禁用
)

type IPListDAO dbs.DAO

func NewIPListDAO() *IPListDAO {
	return dbs.NewDAO(&IPListDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeIPLists",
			Model:  new(IPList),
			PkName: "id",
		},
	}).(*IPListDAO)
}

var SharedIPListDAO *IPListDAO

func init() {
	dbs.OnReady(func() {
		SharedIPListDAO = NewIPListDAO()
	})
}

// 启用条目
func (this *IPListDAO) EnableIPList(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", IPListStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *IPListDAO) DisableIPList(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", IPListStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *IPListDAO) FindEnabledIPList(tx *dbs.Tx, id int64) (*IPList, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", IPListStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*IPList), err
}

// 根据主键查找名称
func (this *IPListDAO) FindIPListName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建名单
func (this *IPListDAO) CreateIPList(tx *dbs.Tx, listType ipconfigs.IPListType, name string, code string, timeoutJSON []byte) (int64, error) {
	op := NewIPListOperator()
	op.IsOn = true
	op.State = IPListStateEnabled
	op.Type = listType
	op.Name = name
	op.Code = code
	if len(timeoutJSON) > 0 {
		op.Timeout = timeoutJSON
	}
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改名单
func (this *IPListDAO) UpdateIPList(tx *dbs.Tx, listId int64, name string, code string, timeoutJSON []byte) error {
	if listId <= 0 {
		return errors.New("invalid listId")
	}
	op := NewIPListOperator()
	op.Id = listId
	op.Name = name
	op.Code = code
	if len(timeoutJSON) > 0 {
		op.Timeout = timeoutJSON
	} else {
		op.Timeout = "null"
	}
	err := this.Save(tx, op)
	return err
}

// 增加版本
func (this *IPListDAO) IncreaseVersion(tx *dbs.Tx) (int64, error) {
	valueJSON, err := SharedSysSettingDAO.ReadSetting(tx, SettingCodeIPListVersion)
	if err != nil {
		return 0, err
	}
	if len(valueJSON) == 0 {
		err = SharedSysSettingDAO.UpdateSetting(tx, SettingCodeIPListVersion, []byte("1"))
		if err != nil {
			return 0, err
		}
		return 1, nil
	}

	value := types.Int64(string(valueJSON)) + 1
	err = SharedSysSettingDAO.UpdateSetting(tx, SettingCodeIPListVersion, []byte(numberutils.FormatInt64(value)))
	return value, nil
}
