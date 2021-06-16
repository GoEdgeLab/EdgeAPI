package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"time"
)

const (
	UserAccessKeyStateEnabled  = 1 // 已启用
	UserAccessKeyStateDisabled = 0 // 已禁用
)

type UserAccessKeyDAO dbs.DAO

func NewUserAccessKeyDAO() *UserAccessKeyDAO {
	return dbs.NewDAO(&UserAccessKeyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserAccessKeys",
			Model:  new(UserAccessKey),
			PkName: "id",
		},
	}).(*UserAccessKeyDAO)
}

var SharedUserAccessKeyDAO *UserAccessKeyDAO

func init() {
	dbs.OnReady(func() {
		SharedUserAccessKeyDAO = NewUserAccessKeyDAO()
	})
}

// EnableUserAccessKey 启用条目
func (this *UserAccessKeyDAO) EnableUserAccessKey(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserAccessKeyStateEnabled).
		Update()
	return err
}

// DisableUserAccessKey 禁用条目
func (this *UserAccessKeyDAO) DisableUserAccessKey(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserAccessKeyStateDisabled).
		Update()
	return err
}

// FindEnabledUserAccessKey 查找启用中的条目
func (this *UserAccessKeyDAO) FindEnabledUserAccessKey(tx *dbs.Tx, id int64) (*UserAccessKey, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", UserAccessKeyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserAccessKey), err
}

// CreateAccessKey 创建Key
func (this *UserAccessKeyDAO) CreateAccessKey(tx *dbs.Tx, userId int64, description string) (int64, error) {
	if userId <= 0 {
		return 0, errors.New("invalid userId")
	}
	op := NewUserAccessKeyOperator()
	op.UserId = userId
	op.Description = description
	op.UniqueId = rands.String(16)
	op.Secret = rands.String(32)
	op.IsOn = true
	op.State = UserAccessKeyStateEnabled
	return this.SaveInt64(tx, op)
}

// FindAllEnabledAccessKeys 查找用户所有的Key
func (this *UserAccessKeyDAO) FindAllEnabledAccessKeys(tx *dbs.Tx, userId int64) (result []*UserAccessKey, err error) {
	_, err = this.Query(tx).
		Attr("userId", userId).
		State(UserAccessKeyStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CheckUserAccessKey 检查用户的AccessKey
func (this *UserAccessKeyDAO) CheckUserAccessKey(tx *dbs.Tx, userId int64, accessKeyId int64) (bool, error) {
	return this.Query(tx).
		Pk(accessKeyId).
		State(UserAccessKeyStateEnabled).
		Attr("userId", userId).
		Exist()
}

// UpdateAccessKeyIsOn 设置是否启用
func (this *UserAccessKeyDAO) UpdateAccessKeyIsOn(tx *dbs.Tx, accessKeyId int64, isOn bool) error {
	if accessKeyId <= 0 {
		return errors.New("invalid accessKeyId")
	}
	_, err := this.Query(tx).
		Pk(accessKeyId).
		Set("isOn", isOn).
		Update()
	return err
}

// FindAccessKeyWithUniqueId 根据UniqueId查找AccessKey
func (this *UserAccessKeyDAO) FindAccessKeyWithUniqueId(tx *dbs.Tx, uniqueId string) (*UserAccessKey, error) {
	one, err := this.Query(tx).
		Attr("uniqueId", uniqueId).
		Attr("isOn", true).
		State(UserAccessKeyStateEnabled).
		Find()
	if one == nil || err != nil {
		return nil, err
	}

	return one.(*UserAccessKey), nil
}

// UpdateAccessKeyAccessedAt 更新AccessKey访问时间
func (this *UserAccessKeyDAO) UpdateAccessKeyAccessedAt(tx *dbs.Tx, accessKeyId int64) error {
	return this.Query(tx).
		Pk(accessKeyId).
		Set("accessedAt", time.Now().Unix()).
		UpdateQuickly()
}
