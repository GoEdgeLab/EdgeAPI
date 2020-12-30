package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
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

// 启用条目
func (this *UserAccessKeyDAO) EnableUserAccessKey(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", UserAccessKeyStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *UserAccessKeyDAO) DisableUserAccessKey(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", UserAccessKeyStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *UserAccessKeyDAO) FindEnabledUserAccessKey(id int64) (*UserAccessKey, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", UserAccessKeyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserAccessKey), err
}

// 创建Key
func (this *UserAccessKeyDAO) CreateAccessKey(userId int64, description string) (int64, error) {
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
	return this.SaveInt64(op)
}

// 查找用户所有的Key
func (this *UserAccessKeyDAO) FindAllEnabledAccessKeys(userId int64) (result []*UserAccessKey, err error) {
	_, err = this.Query().
		State(UserAccessKeyStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 检查用户的AccessKey
func (this *UserAccessKeyDAO) CheckUserAccessKey(userId int64, accessKeyId int64) (bool, error) {
	return this.Query().
		Pk(accessKeyId).
		State(UserAccessKeyStateEnabled).
		Attr("userId", userId).
		Exist()
}

// 设置是否启用
func (this *UserAccessKeyDAO) UpdateAccessKeyIsOn(accessKeyId int64, isOn bool) error {
	if accessKeyId <= 0 {
		return errors.New("invalid accessKeyId")
	}
	_, err := this.Query().
		Pk(accessKeyId).
		Set("isOn", isOn).
		Update()
	return err
}
