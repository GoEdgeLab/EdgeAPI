package models

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	ACMEUserStateEnabled  = 1 // 已启用
	ACMEUserStateDisabled = 0 // 已禁用
)

type ACMEUserDAO dbs.DAO

func NewACMEUserDAO() *ACMEUserDAO {
	return dbs.NewDAO(&ACMEUserDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeACMEUsers",
			Model:  new(ACMEUser),
			PkName: "id",
		},
	}).(*ACMEUserDAO)
}

var SharedACMEUserDAO *ACMEUserDAO

func init() {
	dbs.OnReady(func() {
		SharedACMEUserDAO = NewACMEUserDAO()
	})
}

// 启用条目
func (this *ACMEUserDAO) EnableACMEUser(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", ACMEUserStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *ACMEUserDAO) DisableACMEUser(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", ACMEUserStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *ACMEUserDAO) FindEnabledACMEUser(id int64) (*ACMEUser, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", ACMEUserStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ACMEUser), err
}

// 创建用户
func (this *ACMEUserDAO) CreateACMEUser(adminId int64, userId int64, email string, description string) (int64, error) {
	// 生成私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return 0, err
	}

	privateKeyData, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return 0, err
	}
	privateKeyText := base64.StdEncoding.EncodeToString(privateKeyData)

	op := NewACMEUserOperator()
	op.AdminId = adminId
	op.UserId = userId
	op.Email = email
	op.Description = description
	op.PrivateKey = privateKeyText
	op.State = ACMEUserStateEnabled
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 查找用户列表
func (this *ACMEUserDAO) UpdateACMEUser(acmeUserId int64, description string) error {
	if acmeUserId <= 0 {
		return errors.New("invalid acmeUserId")
	}
	op := NewACMEUserOperator()
	op.Id = acmeUserId
	op.Description = description
	_, err := this.Save(op)
	return err
}

// 计算用户数量
func (this *ACMEUserDAO) CountACMEUsersWithAdminId(adminId int64, userId int64) (int64, error) {
	query := this.Query()
	if adminId > 0 {
		query.Attr("adminId", adminId)
	}
	if userId > 0 {
		query.Attr("userId", userId)
	}

	return query.
		State(ACMEUserStateEnabled).
		Count()
}

// 列出当前管理员的用户
func (this *ACMEUserDAO) ListACMEUsers(adminId int64, userId int64, offset int64, size int64) (result []*ACMEUser, err error) {
	query := this.Query()
	if adminId > 0 {
		query.Attr("adminId", adminId)
	}
	if userId > 0 {
		query.Attr("userId", userId)
	}

	_, err = query.
		State(ACMEUserStateEnabled).
		Offset(offset).
		Limit(size).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// 查找所有用户
func (this *ACMEUserDAO) FindAllACMEUsers(adminId int64, userId int64) (result []*ACMEUser, err error) {
	// 防止没有传入条件导致返回的数据过多
	if adminId <= 0 && userId <= 0 {
		return nil, errors.New("'adminId' or 'userId' should not be empty")
	}

	query := this.Query()
	if adminId > 0 {
		query.Attr("adminId", adminId)
	}
	if userId > 0 {
		query.Attr("userId", userId)
	}
	_, err = query.
		State(ACMEUserStateEnabled).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// 检查用户权限
func (this *ACMEUserDAO) CheckACMEUser(acmeUserId int64, adminId int64, userId int64) (bool, error) {
	if acmeUserId <= 0 {
		return false, nil
	}

	query := this.Query()
	if adminId > 0 {
		query.Attr("adminId", adminId)
	} else if userId > 0 {
		query.Attr("userId", userId)
	} else {
		return false, nil
	}

	return query.
		State(ACMEUserStateEnabled).
		Exist()
}
