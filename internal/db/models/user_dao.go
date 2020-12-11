package models

import (
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

// 启用条目
func (this *UserDAO) EnableUser(id int64) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", UserStateEnabled).
		Update()
}

// 禁用条目
func (this *UserDAO) DisableUser(id int64) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", UserStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *UserDAO) FindEnabledUser(id int64) (*User, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", UserStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*User), err
}

// 查找用户基本信息
func (this *UserDAO) FindEnabledBasicUser(id int64) (*User, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", UserStateEnabled).
		Result("id", "fullname", "username").
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*User), err
}

// 获取管理员名称
func (this *UserDAO) FindUserFullname(userId int64) (string, error) {
	return this.Query().
		Pk(userId).
		Result("fullname").
		FindStringCol("")
}

// 创建用户
func (this *UserDAO) CreateUser(username string, password string, fullname string, mobile string, tel string, email string, remark string, source string) (int64, error) {
	op := NewUserOperator()
	op.Username = username
	op.Password = stringutil.Md5(password)
	op.Fullname = fullname
	op.Mobile = mobile
	op.Tel = tel
	op.Email = email
	op.Remark = remark
	op.Source = source

	op.IsOn = true
	op.State = UserStateEnabled
	err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改用户
func (this *UserDAO) UpdateUser(userId int64, username string, password string, fullname string, mobile string, tel string, email string, remark string, isOn bool) error {
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
	op.IsOn = isOn
	err := this.Save(op)
	return err
}

// 计算用户数量
func (this *UserDAO) CountAllEnabledUsers(keyword string) (int64, error) {
	query := this.Query()
	query.State(UserStateEnabled)
	if len(keyword) > 0 {
		query.Where("(username LIKE :keyword OR fullname LIKE :keyword OR mobile LIKE :keyword OR email LIKE :keyword OR tel LIKE :keyword OR remark LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	return query.Count()
}

// 列出单页用户
func (this *UserDAO) ListEnabledUsers(keyword string) (result []*User, err error) {
	query := this.Query()
	query.State(UserStateEnabled)
	if len(keyword) > 0 {
		query.Where("(username LIKE :keyword OR fullname LIKE :keyword OR mobile LIKE :keyword OR email LIKE :keyword OR tel LIKE :keyword OR remark LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	_, err = query.
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 检查用户名是否存在
func (this *UserDAO) ExistUser(userId int64, username string) (bool, error) {
	return this.Query().
		State(UserStateEnabled).
		Attr("username", username).
		Neq("id", userId).
		Exist()
}

// 列出单页的用户ID
func (this *UserDAO) ListEnabledUserIds(offset, size int64) ([]int64, error) {
	ones, _, err := this.Query().
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
