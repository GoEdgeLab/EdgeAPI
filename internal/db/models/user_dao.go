package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
)

const (
	UserStateEnabled  = 1 // 已启用
	UserStateDisabled = 0 // 已禁用
)

type UserDAO dbs.DAO

func NewUserDAO() *UserDAO {
	return dbs.NewDAO(&UserDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUsers",
			Model:  new(User),
			PkName: "id",
		},
	}).(*UserDAO)
}

var SharedUserDAO *UserDAO

func init() {
	dbs.OnReady(func() {
		SharedUserDAO = NewUserDAO()
	})
}

// EnableUser 启用条目
func (this *UserDAO) EnableUser(tx *dbs.Tx, id int64) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(id).
		Set("state", UserStateEnabled).
		Update()
}

// DisableUser 禁用条目
func (this *UserDAO) DisableUser(tx *dbs.Tx, id int64) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(id).
		Set("state", UserStateDisabled).
		Update()
}

// FindEnabledUser 查找启用中的条目
func (this *UserDAO) FindEnabledUser(tx *dbs.Tx, id int64) (*User, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", UserStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*User), err
}

// FindEnabledBasicUser 查找用户基本信息
func (this *UserDAO) FindEnabledBasicUser(tx *dbs.Tx, id int64) (*User, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", UserStateEnabled).
		Result("id", "fullname", "username").
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*User), err
}

// FindUserFullname 获取管理员名称
func (this *UserDAO) FindUserFullname(tx *dbs.Tx, userId int64) (string, error) {
	return this.Query(tx).
		Pk(userId).
		Result("fullname").
		FindStringCol("")
}

// CreateUser 创建用户
func (this *UserDAO) CreateUser(tx *dbs.Tx, username string, password string, fullname string, mobile string, tel string, email string, remark string, source string, clusterId int64) (int64, error) {
	op := NewUserOperator()
	op.Username = username
	op.Password = stringutil.Md5(password)
	op.Fullname = fullname
	op.Mobile = mobile
	op.Tel = tel
	op.Email = email
	op.Remark = remark
	op.Source = source
	op.ClusterId = clusterId

	op.IsOn = true
	op.State = UserStateEnabled
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateUser 修改用户
func (this *UserDAO) UpdateUser(tx *dbs.Tx, userId int64, username string, password string, fullname string, mobile string, tel string, email string, remark string, isOn bool, nodeClusterId int64) error {
	if userId <= 0 {
		return errors.New("invalid userId")
	}
	op := NewUserOperator()
	op.Id = userId
	op.Username = username
	if len(password) > 0 {
		op.Password = stringutil.Md5(password)
	}
	op.Fullname = fullname
	op.Mobile = mobile
	op.Tel = tel
	op.Email = email
	op.Remark = remark
	op.ClusterId = nodeClusterId
	op.IsOn = isOn
	err := this.Save(tx, op)
	return err
}

// UpdateUserInfo 修改用户基本信息
func (this *UserDAO) UpdateUserInfo(tx *dbs.Tx, userId int64, fullname string) error {
	if userId <= 0 {
		return errors.New("invalid userId")
	}
	op := NewUserOperator()
	op.Id = userId
	op.Fullname = fullname
	return this.Save(tx, op)
}

// UpdateUserLogin 修改用户登录信息
func (this *UserDAO) UpdateUserLogin(tx *dbs.Tx, userId int64, username string, password string) error {
	if userId <= 0 {
		return errors.New("invalid userId")
	}
	op := NewUserOperator()
	op.Id = userId
	op.Username = username
	if len(password) > 0 {
		op.Password = stringutil.Md5(password)
	}
	err := this.Save(tx, op)
	return err
}

// CountAllEnabledUsers 计算用户数量
func (this *UserDAO) CountAllEnabledUsers(tx *dbs.Tx, clusterId int64, keyword string) (int64, error) {
	query := this.Query(tx)
	query.State(UserStateEnabled)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if len(keyword) > 0 {
		query.Where("(username LIKE :keyword OR fullname LIKE :keyword OR mobile LIKE :keyword OR email LIKE :keyword OR tel LIKE :keyword OR remark LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	return query.Count()
}

// ListEnabledUsers 列出单页用户
func (this *UserDAO) ListEnabledUsers(tx *dbs.Tx, clusterId int64, keyword string, offset int64, size int64) (result []*User, err error) {
	query := this.Query(tx)
	query.State(UserStateEnabled)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if len(keyword) > 0 {
		query.Where("(username LIKE :keyword OR fullname LIKE :keyword OR mobile LIKE :keyword OR email LIKE :keyword OR tel LIKE :keyword OR remark LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	_, err = query.
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// ExistUser 检查用户名是否存在
func (this *UserDAO) ExistUser(tx *dbs.Tx, userId int64, username string) (bool, error) {
	return this.Query(tx).
		State(UserStateEnabled).
		Attr("username", username).
		Neq("id", userId).
		Exist()
}

// ListEnabledUserIds 列出单页的用户ID
func (this *UserDAO) ListEnabledUserIds(tx *dbs.Tx, offset, size int64) ([]int64, error) {
	ones, _, err := this.Query(tx).
		ResultPk().
		State(UserStateEnabled).
		Offset(offset).
		Limit(size).
		AscPk().
		FindOnes()
	if err != nil {
		return nil, err
	}
	result := []int64{}
	for _, one := range ones {
		result = append(result, one.GetInt64("id"))
	}
	return result, nil
}

// CheckUserPassword 检查用户名、密码
func (this *UserDAO) CheckUserPassword(tx *dbs.Tx, username string, encryptedPassword string) (int64, error) {
	if len(username) == 0 || len(encryptedPassword) == 0 {
		return 0, nil
	}
	return this.Query(tx).
		Attr("username", username).
		Attr("password", encryptedPassword).
		Attr("state", UserStateEnabled).
		Attr("isOn", true).
		ResultPk().
		FindInt64Col(0)
}

// FindUserClusterId 查找用户所在集群
func (this *UserDAO) FindUserClusterId(tx *dbs.Tx, userId int64) (int64, error) {
	return this.Query(tx).
		Pk(userId).
		Result("clusterId").
		FindInt64Col(0)
}

// UpdateUserFeatures 更新用户Features
func (this *UserDAO) UpdateUserFeatures(tx *dbs.Tx, userId int64, featuresJSON []byte) error {
	if userId <= 0 {
		return errors.New("invalid userId")
	}
	if len(featuresJSON) == 0 {
		featuresJSON = []byte("[]")
	}
	_, err := this.Query(tx).
		Pk(userId).
		Set("features", featuresJSON).
		Update()
	if err != nil {
		return err
	}
	return nil
}

// FindUserFeatures 查找用户Features
func (this *UserDAO) FindUserFeatures(tx *dbs.Tx, userId int64) ([]*UserFeature, error) {
	featuresJSON, err := this.Query(tx).
		Pk(userId).
		Result("features").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if len(featuresJSON) == 0 {
		return nil, nil
	}

	featureCodes := []string{}
	err = json.Unmarshal([]byte(featuresJSON), &featureCodes)
	if err != nil {
		return nil, err
	}

	// 检查是否还存在以及设置名称
	result := []*UserFeature{}
	if len(featureCodes) > 0 {
		for _, featureCode := range featureCodes {
			f := FindUserFeature(featureCode)
			if f != nil {
				result = append(result, &UserFeature{Name: f.Name, Code: f.Code, Description: f.Description})
			}
		}
	}

	return result, nil
}
