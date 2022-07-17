package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

const (
	UserIdentityStateEnabled  = 1 // 已启用
	UserIdentityStateDisabled = 0 // 已禁用
)

type UserIdentityDAO dbs.DAO

func NewUserIdentityDAO() *UserIdentityDAO {
	return dbs.NewDAO(&UserIdentityDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserIdentities",
			Model:  new(UserIdentity),
			PkName: "id",
		},
	}).(*UserIdentityDAO)
}

var SharedUserIdentityDAO *UserIdentityDAO

func init() {
	dbs.OnReady(func() {
		SharedUserIdentityDAO = NewUserIdentityDAO()
	})
}

// EnableUserIdentity 启用条目
func (this *UserIdentityDAO) EnableUserIdentity(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserIdentityStateEnabled).
		Update()
	return err
}

// DisableUserIdentity 禁用条目
func (this *UserIdentityDAO) DisableUserIdentity(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserIdentityStateDisabled).
		Update()
	return err
}

// FindEnabledUserIdentity 查找启用中的条目
func (this *UserIdentityDAO) FindEnabledUserIdentity(tx *dbs.Tx, id int64) (*UserIdentity, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", UserIdentityStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserIdentity), err
}

// CreateUserIdentity 创建
func (this *UserIdentityDAO) CreateUserIdentity(tx *dbs.Tx, userId int64, idType userconfigs.UserIdentityType, realName string, number string, fileIds []int64) (int64, error) {
	var op = NewUserIdentityOperator()
	op.UserId = userId
	op.Type = idType
	op.RealName = realName
	op.Number = number

	if fileIds == nil {
		fileIds = []int64{}
	}
	fileIdsJSON, err := json.Marshal(fileIds)
	if err != nil {
		return 0, err
	}
	op.FileIds = fileIdsJSON

	op.Status = userconfigs.UserIdentityStatusNone
	op.State = UserIdentityStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateUserIdentity 修改
func (this *UserIdentityDAO) UpdateUserIdentity(tx *dbs.Tx, identityId int64, idType userconfigs.UserIdentityType, realName string, number string, fileIds []int64) error {
	if identityId <= 0 {
		return nil
	}

	var op = NewUserIdentityOperator()
	op.Id = identityId
	op.Type = idType
	op.Number = number

	if fileIds == nil {
		fileIds = []int64{}
	}
	fileIdsJSON, err := json.Marshal(fileIds)
	if err != nil {
		return err
	}
	op.FileIds = fileIdsJSON

	return this.Save(tx, op)
}

// SubmitUserIdentity 提交审核
func (this *UserIdentityDAO) SubmitUserIdentity(tx *dbs.Tx, identityId int64) error {
	return this.Query(tx).
		Pk(identityId).
		Set("status", userconfigs.UserIdentityStatusSubmitted).
		Set("submittedAt", time.Now().Unix()).
		UpdateQuickly()
}

// CancelUserIdentity 取消提交审核
func (this *UserIdentityDAO) CancelUserIdentity(tx *dbs.Tx, identityId int64) error {
	return this.Query(tx).
		Pk(identityId).
		Set("status", userconfigs.UserIdentityStatusNone).
		Set("updatedAt", time.Now().Unix()).
		UpdateQuickly()
}

// RejectUserIdentity 拒绝
func (this *UserIdentityDAO) RejectUserIdentity(tx *dbs.Tx, identityId int64) error {
	return this.Query(tx).
		Pk(identityId).
		Set("status", userconfigs.UserIdentityStatusRejected).
		Set("rejectedAt", time.Now().Unix()).
		UpdateQuickly()
}

// VerifyUserIdentity 通过
func (this *UserIdentityDAO) VerifyUserIdentity(tx *dbs.Tx, identityId int64) error {
	return this.Query(tx).
		Pk(identityId).
		Set("status", userconfigs.UserIdentityStatusVerified).
		Set("verifiedAt", time.Now().Unix()).
		UpdateQuickly()
}

// CheckUserIdentity 检查用户认证
func (this *UserIdentityDAO) CheckUserIdentity(tx *dbs.Tx, userId int64, identityId int64) error {
	b, err := this.Query(tx).
		Pk(identityId).
		Attr("userId", userId).
		State(UserIdentityStateEnabled).
		Exist()
	if err != nil {
		return err
	}
	if !b {
		return ErrNotFound
	}
	return nil
}

// FindUserIdentityStatus 查找认证信息当前状态
func (this *UserIdentityDAO) FindUserIdentityStatus(tx *dbs.Tx, identityId int64) (userconfigs.UserIdentityStatus, error) {
	return this.Query(tx).
		Pk(identityId).
		Result("status").
		FindStringCol("")
}

// FindEnabledUserIdentityWithType 查找某个类型的认证信息
func (this *UserIdentityDAO) FindEnabledUserIdentityWithType(tx *dbs.Tx, userId int64, idType userconfigs.UserIdentityType) (*UserIdentity, error) {
	one, err := this.Query(tx).
		Attr("userId", userId).
		Attr("type", idType).
		State(UserIdentityStateEnabled).
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*UserIdentity), nil
}
